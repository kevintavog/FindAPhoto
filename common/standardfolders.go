package common

import (
	"os"
	"path"
	"runtime"

	"github.com/ian-kent/go-log/log"
)

var HomeDirectory string
var LogDirectory string
var ConfigDirectory string
var ThumbnailDirectory string
var ExifToolPath string
var FfmpegPath string

func InitDirectories(appName string) {
	if HomeDirectory == "" {
		if runtime.GOOS == "darwin" {
			HomeDirectory = os.Getenv("HOME")
			LogDirectory = path.Join(HomeDirectory, "Library", "Logs", appName)
			ConfigDirectory = path.Join(HomeDirectory, "Library", "Preferences")
			ThumbnailDirectory = path.Join(HomeDirectory, "Library", "Application Support", "FindAPhoto", "thumbnails")
			FfmpegPath = "/usr/local/bin/ffmpeg"
			ExifToolPath = "/usr/local/bin/exiftool"
		} else if runtime.GOOS == "linux" {
			log.Fatal("Come up with directories for Linux!")
			FfmpegPath = "/usr/bin/ffmpeg"
			ExifToolPath = "/usr/bin/exiftool"
		} else {
			log.Fatal("Come up with directories for: %v", runtime.GOOS)
		}
	}
}
