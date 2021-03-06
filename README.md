# db2m4atag
A small Go application to tag m4a files from an iPhone, iPad or  SQLite DB

## Build
This app use `github.com/mattn/go-sqlite3` to open the SQLite file. It's a CGO library (for now), so you may need to build it first with GCC. Please refer to [the doc here](https://github.com/mattn/go-sqlite3#installation).

then a simple `go build` should create the `db2m4atag` binary.

## Usage

This program use `mp4tags` command from [mp4v2](https://code.google.com/archive/p/mp4v2/) application to edit the M4A tags. 
It's easy to install on a Mac using Brew:

```bash
brew install mp4v2
```

To make this work you need the `MediaLibrary.sqlitedb` file and the folders containing the song files (mp3 and m4a). 
So you need to access your iPhone content. 

This can be done on linux with free tools, as described [here](https://www.dedoimedo.com/computers/linux-iphone-6s-ios-11.html).

I'm copying the commands here:

```bash
sudo apt-get install ideviceinstaller python-imobiledevice libimobiledevice-utils libimobiledevice6 libplist3 python-plist ifuse usbmuxd
usbmuxd -f -v
idevicepair pair

SUCCESS: Paired with device 0000XXXXXXXXXXXXX

sudo mkdir /media/iPhone
sudo chown <your user>:<your group> /media/iPhone
ifuse /media/iPhone/

ls -l /media/iPhone

# it things goes wrong:
sudo umount /media/iPhone
ifuse /media/iPhone/
ls -l /media/iPhone
...
```

You should now find the sqlite database in `/media/iPhone/iTunes_Control/iTunes/MediaLibrary.sqlitedb` and the audio files in `/media/iPhone/iTunes_Control/Music`

Here's the default usage:
```bash
./db2m4atag -help
Usage of ./db2m4atag:
  -DBpath string
    	path to the sqlite DB file (default "MediaLibrary.sqlitedb")
  -musicDir string
    	folder to scan for mp3 and m4a (default "F00")
  -tagger string
    	programme to use for managing tags (default "mp4tags")
  -version
    	Show version and quit
```

If you're following the steps above, I strongly suggest copying the iphone files in a local folder before running the command:

```bash
mkdir -p /data/iphone-backup
cp /media/iPhone/iTunes_Control/iTunes/MediaLibrary.sqlitedb* /data/iphone-backup/
rsync -r --progress --stats  /media/iPhone/iTunes_Control/Music /data/iphone-backup/
cd /data/iphone-backup
```

Now you can start using the tool to tag all the files:

```bash
for i in F00  F01  F02  F03  F04  F05  F06  F07  F08  F09  F10  F11  F12  F13  F14  F15  F16  F17  F18  F19  F20  F21  F22  F23  F24  F25  F26  F27  F28  F29  F30  F31  F32  F33  F34  F35  F36  F37  F38  F39  F40  F41  F42  F43  F44  F45  F46  F47  F48  F49 ; do ./db2m4atag -musicDir /data/iphone-backup/Music/${i} -DBpath /data/iphone-backup/MediaLibrary.sqlitedb ; done
```
