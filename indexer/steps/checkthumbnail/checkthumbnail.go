package checkthumbnail

import (
	"sync"
	"sync/atomic"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/indexer/steps/generatethumbnail"

	"github.com/ian-kent/go-log/log"
)

var FailedChecks int64

const numConsumers = 8

var queue = make(chan *generatethumbnail.ThumbnailInfo, numConsumers)
var waitGroup sync.WaitGroup

func Start() {
	err := common.CreateDirectory(common.ThumbnailDirectory)
	if err != nil {
		log.Fatal("Unable to create thumbnail directory (%s): %s", common.ThumbnailDirectory, err.Error())
	}

	generatethumbnail.Start()

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
	generatethumbnail.Done()
	generatethumbnail.Wait()
}

func Enqueue(fullPath, aliasedPath, mimeType string) {
	thumbnailInfo := &generatethumbnail.ThumbnailInfo{
		FullPath:    fullPath,
		AliasedPath: aliasedPath,
		MimeType:    mimeType,
	}
	queue <- thumbnailInfo
}

func dequeue() {

	for thumbnailInfo := range queue {
		thumbPath := common.ToThumbPath(thumbnailInfo.AliasedPath)
		exists, err := common.PathExists(thumbPath)
		if err != nil {
			log.Warn("Error checking thumbnail existence of %s: %s", thumbPath, err.Error())
			atomic.AddInt64(&FailedChecks, 1)
			continue
		}

		if !exists {
			generatethumbnail.Enqueue(thumbnailInfo.FullPath, thumbnailInfo.AliasedPath, thumbnailInfo.MimeType)
		}
	}
}
