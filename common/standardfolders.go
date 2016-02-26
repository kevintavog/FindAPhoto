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

func InitDirectories(appName string) {
	if HomeDirectory == "" {
		if runtime.GOOS == "darwin" {
			HomeDirectory = os.Getenv("HOME")
			LogDirectory = path.Join(HomeDirectory, "Library", "Logs", appName)
			ConfigDirectory = path.Join(HomeDirectory, "Library", "Preferences")
		} else if runtime.GOOS == "linux" {
			log.Fatal("Come up with directories for Linux!")
		} else {
			log.Fatal("Come up with directories for: %v", runtime.GOOS)
		}
	}
}
