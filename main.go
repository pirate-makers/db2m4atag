package main

import (
	"database/sql"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/dhowden/tag"
	_ "github.com/mattn/go-sqlite3"
)

type Song struct {
	Artist      sql.NullString `json:"artist"`
	Album       sql.NullString `json:"album"`
	Title       sql.NullString `json:"title"`
	Rating      sql.NullString `json:"rating"`
	Disc        sql.NullString `json:"disc"`
	DiscCount   sql.NullString `json:"disccount"`
	Track       sql.NullString `json:"track"`
	TrackCount  sql.NullString `json:"trackcount"`
	Genre       sql.NullString `json:"genre"`
	Grouping    sql.NullString `json:"grouping"`
	Year        sql.NullString `json:"year"`
	Lyrics      sql.NullString `json:"lyrics"`
	AlbumYear   sql.NullString `json:"albumyear"`
	AlbumArtist sql.NullString `json:"albumartist"`
	Location    sql.NullString `json:"location"`
	Artwork     sql.NullString `json:"artwork"`
	Composer    sql.NullString `json:"composer"`
}

const req = `
SELECT
	ar.item_artist,
	al.album,
	ex.title,
	st.user_rating AS "rating [integer]",
	i.disc_number,
	ex.disc_count,
	i.track_number,
	ex.track_count,
	g.genre,
	ex.grouping,
	ex.year,
	ly.lyrics,
	al.album_year AS "album_year",
	aar.album_artist,
	bl.path || "/" || ex.location AS "path",
	art.relative_path AS "artwork",
	c.composer
FROM item i
	LEFT JOIN item_extra      ex ON ex.item_pid = i.item_pid
	LEFT JOIN item_artist     ar ON ar.item_artist_pid = i.item_artist_pid
	LEFT JOIN album           al ON al.album_pid = i.album_pid
	LEFT JOIN item_stats      st ON st.item_pid = i.item_pid
	LEFT JOIN base_location   bl ON bl.base_location_id = i.base_location_id
	LEFT JOIN genre           g ON g.genre_id = i.genre_id
	LEFT JOIN album_artist    aar ON aar.album_artist_pid = al.album_artist_pid
	LEFT JOIN composer        c ON c.composer_pid = i.composer_pid
	LEFT JOIN lyrics          ly ON ly.item_pid = i.item_pid
	LEFT JOIN artwork_token   artt ON artt.entity_pid = i.item_pid
	LEFT JOIN artwork art ON  art.artwork_token = artt.artwork_token
WHERE
	i.item_pid IS NOT NULL
AND
	ex.location LIKE ?
`

var (
	version        = "no version set"
	dryrun         = flag.Bool("dryrun", false, "don't update the audio file, just display what's going to be done")
	displayVersion = flag.Bool("version", false, "Show version and quit")
	musicDir       = flag.String("musicDir", "F01", "folder to scan for mp3 and m4a")
	tagger         = flag.String("tagger", "mp4tags", "programme to use for managing tags")
	DBpath         = flag.String("DBpath", "MediaLibrary-bkp.sqlitedb", "path to the sqlite DB file")
)

func printVersion() {
	log.Printf("Go Version: %s\n", runtime.Version())
	log.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	log.Printf("Version: %v\n", version)
}

func main() {
	// parse flags
	flag.Parse()
	if *displayVersion {
		printVersion()
		os.Exit(0)
	}

	// open the DB
	db, err := sql.Open("sqlite3", *DBpath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// get file list
	files, err := ioutil.ReadDir(*musicDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		log.Printf("working on %s/%s", *musicDir, file.Name())

		if file.IsDir() || (filepath.Ext(file.Name()) != ".mp3" && filepath.Ext(file.Name()) != ".m4a") {
			continue
		}

		// check file tags
		f, err := os.Open(*musicDir + "/" + file.Name())
		if err != nil {
			log.Printf("error loading file, skipping it: %v", err)
			continue
		}
		defer f.Close()

		// read existing tags
		m, err := tag.ReadFrom(f)
		if err != nil {
			log.Printf("error reading tags in file: %v\n", err)
		}

		if err == nil {
			// if we already have the title tag, do nothing and go to next audio file
			if m.Title() != "" {
				log.Printf("skipping %s(%s): %s/%s/%s\n", f.Name(), m.Format(), m.Album(), m.Artist(), m.Title())
				// if we have tags, continue with the next item
				continue
			}
		}

		// select tags from the DB
		stmt, err := db.Prepare(req)
		if err != nil {
			log.Println(file.Name(), err)
			continue
		}
		defer stmt.Close()

		var s Song
		err = stmt.QueryRow(file.Name()).Scan(&s.Artist, &s.Album, &s.Title, &s.Rating, &s.Disc, &s.DiscCount, &s.Track, &s.TrackCount, &s.Genre, &s.Grouping, &s.Year, &s.Lyrics, &s.AlbumYear, &s.AlbumArtist, &s.Location, &s.Artwork, &s.Composer)
		if err != nil {
			log.Println(file.Name(), err)
			continue
		}

		log.Printf("updating %s: %s/%s/%s\n", file.Name(), s.Album.String, s.Artist.String, s.Title.String)

		if !*dryrun {
			// apply the tags from the DB into the M4A file
			if filepath.Ext(file.Name()) == ".m4a" {
				// use mp4tags to update
				cmd := exec.Command(
					"mp4tags",
					"-A", s.Album.String,
					"-a", s.Artist.String,
					"-c", "set by prune",
					"-d", s.Disc.String,
					"-D", s.DiscCount.String,
					"-g", s.Genre.String,
					"-L", s.Lyrics.String,
					"-G", s.Grouping.String,
					"-P", s.Artwork.String,
					"-R", s.AlbumArtist.String,
					"-s", s.Title.String,
					"-t", s.Track.String,
					"-T", s.TrackCount.String,
					"-X", s.Rating.String,
					"-w", s.Composer.String,
					"-y", s.Year.String,
					*musicDir+"/"+file.Name())

				// "-X", strconv.Itoa(int(s.Rating.Int64)),
				err = cmd.Run()
				if err != nil {
					log.Printf("Command finished with error: %v", err)
				}
				log.Printf("new tags applied on %s", file.Name())
				continue
			}

			// apply the tags from the DB into the MP3 file
			if filepath.Ext(file.Name()) == ".mp3" {
				// nothing to do for now
				// mp3 alredy have tags and artwork embeded
			}
		}
	}
}
