package common

import (
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/ian-kent/go-log/log"
)

var HomeDirectory string
var LogDirectory string
var ConfigDirectory string
var ThumbnailDirectory string
var LocationCacheDirectory string
var ExifToolPath string
var FfmpegPath string
var VipsThumbnailPath string
var ExecutingDirectory string
var IndexerPath string

func InitDirectories(appName string) {
	if HomeDirectory == "" {
		exeDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal("Unable to get executing directory: %s", err)
		}

		ExecutingDirectory = exeDir
		IndexerPath = ExecutingDirectory + "/indexer"

		if runtime.GOOS == "darwin" {
			HomeDirectory = os.Getenv("HOME")
			LogDirectory = path.Join(HomeDirectory, "Library", "Logs", appName)
			ConfigDirectory = path.Join(HomeDirectory, "Library", "Preferences")
			ThumbnailDirectory = path.Join(HomeDirectory, "Library", "Application Support", "FindAPhoto", "thumbnails")
			LocationCacheDirectory = path.Join(HomeDirectory, "Library", "Application Support", "FindAPhoto")
			FfmpegPath = "/usr/local/bin/ffmpeg"
			ExifToolPath = "/usr/local/bin/exiftool"
			VipsThumbnailPath = "/usr/local/bin/vipsthumbnail"
		} else if runtime.GOOS == "linux" {
			HomeDirectory = os.Getenv("HOME")
			ThumbnailDirectory = path.Join(HomeDirectory, ".findaphoto", "thumbnails")
			LogDirectory = path.Join(HomeDirectory, ".findaphoto", "logs")
			ConfigDirectory = path.Join(HomeDirectory, ".findaphoto")
			LocationCacheDirectory = path.Join(HomeDirectory, ".findaphoto")
			FfmpegPath = "/usr/bin/ffmpeg"
			ExifToolPath = "/usr/bin/exiftool"
			VipsThumbnailPath = "/usr/bin/vipsthumbnail"
		} else {
			log.Fatal("Come up with directories for: %v", runtime.GOOS)
		}
	}
}
