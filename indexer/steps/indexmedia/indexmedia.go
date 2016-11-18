package indexmedia

import (
	"sync"
	"sync/atomic"

	"github.com/kevintavog/findaphoto/common"

	"github.com/ian-kent/go-log/log"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
)

var IndexedFiles int64
var ChangedFiles int64 // Already in the repository, a file change was detected
var FailedIndexAttempts int64

const numConsumers = 8

var queue = make(chan *common.Media, numConsumers)
var waitGroup sync.WaitGroup

func Start() {
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
}

func Enqueue(media *common.Media) {
	queue <- media
}

func dequeue() {
	var client = common.CreateClient()
	for media := range queue {

		response, err := client.Index().
			Index(common.MediaIndexName).
			Type(common.MediaTypeName).
			Id(media.Path).
			BodyJson(media).
			Do(context.TODO())

		if err != nil {
			atomic.AddInt64(&FailedIndexAttempts, 1)
			if elasticErr, ok := err.(*elastic.Error); ok {
				log.Error("Failed indexing %s: status=%d; %s", media.Path, elasticErr.Status, elasticErr.Details)
			} else {
				log.Error("Failed indexing %s: %q", media.Path, err.Error())
			}
			continue
		}

		atomic.AddInt64(&IndexedFiles, 1)
		if !response.Created {
			atomic.AddInt64(&ChangedFiles, 1)
		}

		if IndexedFiles%1000 == 0 {
			log.Info("Indexed [%d] for %s", IndexedFiles, media.Path)
		}
	}
}
