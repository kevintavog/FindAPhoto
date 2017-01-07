package scanner

import (
	"io/ioutil"
	"path"
	"strings"
	"sync"

	"github.com/kevintavog/findaphoto/indexer/steps/checkindex"

	"github.com/ian-kent/go-log/log"
)

var FilesScanned int64
var SupportedFilesFound int64
var DirectoriesScanned int64

var supportedFileExtensions = map[string]bool{
	".BMP":  true,
	".GIF":  true,
	".JPEG": true,
	".JPG":  true,
	".M4V":  true,
	".MP4":  true,
	".PNG":  true,
	".TIF":  true,
	".TIFF": true,
}

func Scan(scanPath, alias string) {
	checkindex.Start()

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		RemoveFiles()
		waitGroup.Done()
	}()

	go func() {
		scan(scanPath, alias, scanPath)
		log.Debug("scan completed")
		checkindex.Done()
	}()

	waitGroup.Wait()
	checkindex.Wait()
}

func scan(basePath, alias, scanPath string) {
	dirReader, err := ioutil.ReadDir(scanPath)
	if err != nil {
		log.Warn("Failed reading files in '%s': %s", scanPath, err.Error())
		return
	}

	subDirectories := []string{}

	baseLength := len(basePath)
	for _, fileInfo := range dirReader {
		if fileInfo.IsDir() {
			DirectoriesScanned += 1
			subDirectories = append(subDirectories, fileInfo.Name())
		} else {
			FilesScanned += 1
			var ok bool
			if _, ok = supportedFileExtensions[strings.ToUpper(path.Ext(fileInfo.Name()))]; ok {
				SupportedFilesFound += 1
				fullPath := path.Join(scanPath, fileInfo.Name())

				relativePath := fullPath[baseLength:len(fullPath)]
				if len(relativePath) > 0 && relativePath[0] == '/' {
					relativePath = relativePath[1:len(relativePath)]
				}
				aliasedPath := strings.Replace(path.Join(alias, relativePath), "/", "\\", -1)

				checkindex.Enqueue(fullPath, aliasedPath, fileInfo.Size())
			}
		}
	}

	for _, directory := range subDirectories {
		scan(basePath, alias, path.Join(scanPath, directory))
	}
}
