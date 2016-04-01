package getexif

import (
	"bytes"
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"sync"
	"sync/atomic"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/indexer/steps/preparemedia"

	"github.com/ian-kent/go-log/log"
)

var ExifToolInvocations int64
var ExifToolFailed int64

type ExifForDirectory struct {
	Directory string
	Files     *list.List
	lock      sync.Mutex
	dequeued  bool
	Exif      []*common.ExifOutput
}

const numConsumers = 8

var queue = make(chan *ExifForDirectory, numConsumers)
var waitGroup sync.WaitGroup
var allDirectories map[string]*ExifForDirectory
var allLock sync.Mutex

func Start() {
	preparemedia.Start()

	allDirectories = make(map[string]*ExifForDirectory)

	waitGroup.Add(numConsumers)
	for idx := 0; idx < numConsumers; idx++ {
		go func() {
			dequeue()
			waitGroup.Done()
		}()
	}
}

func Done() {
	close(queue)
}

func Wait() {
	waitGroup.Wait()
	preparemedia.Done()
	preparemedia.Wait()
}

func Enqueue(candidate *common.CandidateFile) {
	directory := path.Dir(candidate.FullPath)

	var inQueue bool
	var exifForDirectory *ExifForDirectory

	// Lock only to check if this has been queued & added to allDirectories
	{
		allLock.Lock()
		defer allLock.Unlock()

		if exifForDirectory, inQueue = allDirectories[directory]; !inQueue {
			exifForDirectory = &ExifForDirectory{
				Directory: directory,
				Files:     list.New(),
				dequeued:  false,
			}

			allDirectories[directory] = exifForDirectory
		}
	}

	// Add this file to the exif for the directory
	{
		exifForDirectory.lock.Lock()
		defer exifForDirectory.lock.Unlock()

		if !exifForDirectory.dequeued {
			exifForDirectory.Files.PushBack(candidate)
		} else {
			ex, err := exifForFile(exifForDirectory, path.Base(candidate.FullPath))
			if err != nil {
				candidate.Exif = *ex
			}
			preparemedia.Enqueue(candidate)
		}
	}

	if !inQueue {
		queue <- exifForDirectory
	}
}

func dequeue() {
	for exifForDirectory := range queue {

		atomic.AddInt64(&ExifToolInvocations, 1)

		exifOutput, err := getDirectoryExif(exifForDirectory.Directory)
		if err != nil {
			atomic.AddInt64(&ExifToolFailed, 1)
			continue
		}

		exifForDirectory.Exif = exifOutput

		exifForDirectory.lock.Lock()
		defer exifForDirectory.lock.Unlock()

		exifForDirectory.dequeued = true

		for ele := exifForDirectory.Files.Front(); ele != nil; ele = ele.Next() {
			candidate := ele.Value.(*common.CandidateFile)
			ex, err := exifForFile(exifForDirectory, path.Base(candidate.FullPath))
			if err == nil {
				candidate.Exif = *ex
			} else {
				candidate.AddWarning(err.Error())
			}
			preparemedia.Enqueue(candidate)
		}
	}
}

func exifForFile(exifForDirectory *ExifForDirectory, filename string) (*common.ExifOutput, error) {
	if !exifForDirectory.dequeued {
		return nil, errors.New("No exif data available")
	}

	for idx, ex := range exifForDirectory.Exif {
		if filename == path.Base(ex.SourceFile) {

			exifForDirectory.Exif = append(exifForDirectory.Exif[:idx], exifForDirectory.Exif[idx+1:]...)
			//			directory := path.Dir(filename)

			// Locking here is likely a poor idea - there's a lock on 'exifForDirectory', so a deadlock is quite possible
			//			allLock.Lock()
			//			defer allLock.Unlock()
			//			log.Warn("Deleting directory fomr allDirectories: %v", directory)
			//			delete(allDirectories, directory)

			return ex, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("No exif for %s", filename))
}

func getDirectoryExif(directory string) ([]*common.ExifOutput, error) {
	out, err := exec.Command(common.ExifToolPath, "-a", "-j", "-g", "-x", "Directory", "-x", "FileAccessDate", "-x", "FileInodeChangeDate", directory).Output()
	if err != nil {
		log.Fatal("Failed executing exiftool for '%s': %s", directory, err.Error())
	}

	var response []*common.ExifOutput
	decoder := json.NewDecoder(bytes.NewReader(out))
	err = decoder.Decode(&response)
	if err != nil {
		log.Error("json decoder error for '%s': %s", directory, err.Error())
		return nil, err
	}

	return response, nil
}
