package checkindex

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"sync"
	"sync/atomic"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/indexer/steps/checkthumbnail"
	"github.com/kevintavog/findaphoto/indexer/steps/generatethumbnail"
	"github.com/kevintavog/findaphoto/indexer/steps/getexif"

	"github.com/ian-kent/go-log/log"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
)

var BadJson int64
var CheckFailed int64
var SignatureGenerationFailed int64
var ChecksMade int64

var ForceIndex bool

const numConsumers = 8
const numBytesForSignature = 20 * 1024

var queue = make(chan *common.CandidateFile, numConsumers)
var waitGroup sync.WaitGroup

func Start() {
	getexif.Start()
	checkthumbnail.Start()

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
	getexif.Done()
	getexif.Wait()
	checkthumbnail.Done()
	checkthumbnail.Wait()
}

func Enqueue(fullFilename, aliasedFilename string, lengthInBytes int64) {
	candidateFile := &common.CandidateFile{
		FullPath:      fullFilename,
		AliasedPath:   aliasedFilename,
		LengthInBytes: lengthInBytes,
	}
	queue <- candidateFile
}

func dequeue() {
	client := common.CreateClient()

	for candidateFile := range queue {

		// We need the signature & length for further validation, below - and another step needs it
		// if the file is added to the index. Essentially, we need it most of the time - calculate it now
		signature, err := generateSignature(candidateFile.FullPath)
		if err != nil {
			atomic.AddInt64(&SignatureGenerationFailed, 1)
			continue
		}
		candidateFile.Signature = signature

		atomic.AddInt64(&ChecksMade, 1)
		if ChecksMade%1000 == 0 {
			log.Info("Checking [%d] for %s", ChecksMade, candidateFile.AliasedPath)
		}

		termQuery := elastic.NewTermQuery("_id", candidateFile.AliasedPath)
		searchResult, err := client.Search().
			Index(common.MediaIndexName).
			Type(common.MediaTypeName).
			Query(termQuery).
			Pretty(true).
			Do(context.TODO())
		if err != nil {
			atomic.AddInt64(&CheckFailed, 1)
			log.Error("Error checking document existence for '%s': %s", candidateFile.AliasedPath, err.Error())
			continue
		}

		if ForceIndex || searchResult.TotalHits() != 1 {
			getexif.Enqueue(candidateFile)
		} else {
			hit := searchResult.Hits.Hits[0]
			var media common.Media
			err := json.Unmarshal(*hit.Source, &media)
			if err != nil {
				log.Error("Failed deserializing search result: %s", err.Error())
				atomic.AddInt64(&BadJson, 1)
			} else {
				if media.Signature != candidateFile.Signature || media.LengthInBytes != candidateFile.LengthInBytes {
					getexif.Enqueue(candidateFile)

					// Because it's an update, ask to generate the thumbnail rather than check if it exists
					generatethumbnail.Enqueue(candidateFile.FullPath, candidateFile.AliasedPath, media.MimeType)
				} else {
					checkthumbnail.Enqueue(candidateFile.FullPath, candidateFile.AliasedPath, media.MimeType)
				}
			}
		}
	}
}

func generateSignature(filename string) (string, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		log.Error("Failed opening '%s': %s", filename, err.Error())
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, numBytesForSignature)
	bytesRead, err := file.Read(buffer)
	if err != nil {
		log.Error("Failed reading '%s': %s", filename, err.Error())
		return "", err
	}

	sha := sha256.New()
	sha.Write(buffer[0:bytesRead])
	return hex.EncodeToString(sha.Sum(nil)), nil
}
