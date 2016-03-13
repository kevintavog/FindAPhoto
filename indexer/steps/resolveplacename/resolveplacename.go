package resolveplacename

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/indexer/steps/indexmedia"

	"github.com/Jeffail/gabs"
	"github.com/ian-kent/go-log/log"
)

var PlacenameLookups int64
var FailedLookups int64
var Failures int64
var ServerErrors int64

var OpenStreetMapUrl = ""
var OpenStreetMapKey = ""

const numConsumers = 8

var queue = make(chan *common.Media, numConsumers)
var waitGroup sync.WaitGroup

func Start() {
	indexmedia.Start()

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
	indexmedia.Done()
	indexmedia.Wait()
}

func Enqueue(media *common.Media) {
	queue <- media
}

func dequeue() {
	for media := range queue {
		lookupInCache(media)
		//		resolvePlacename(media)
		indexmedia.Enqueue(media)
	}
}

func addWarning(media *common.Media, warning string) {
	log.Warn("%q - %q; %f, %f", media.Path, warning, media.Location.Latitude, media.Location.Longitude)
	media.Warnings = append(media.Warnings, warning)
}

func lookupInCache(media *common.Media) bool {
	if media.Location.Latitude == 0 && media.Location.Longitude == 0 {
		return false
	}

	json, err := placenameFromLocalCache(media.Location.Latitude, media.Location.Longitude)
	if err != nil {
		addWarning(media, fmt.Sprintf("Failed looking up via cache: %v", err.Error()))
		return false
	}
	if len(json) == 0 {
		return false
	}

	placenameFromText(media, bytes.NewBufferString(json).Bytes())
	return true
}

func resolvePlacename(media *common.Media) {
	if media.Location.Latitude == 0 && media.Location.Longitude == 0 {
		return
	}

	atomic.AddInt64(&PlacenameLookups, 1)
	url := fmt.Sprintf("%s/nominatim/v1/reverse?key=%s&format=json&lat=%f&lon=%f&addressdetails=1&zoom=18&accept-language=en-us",
		OpenStreetMapUrl, OpenStreetMapKey, media.Location.Latitude, media.Location.Longitude)

	response, err := http.Get(url)
	if err != nil {
		atomic.AddInt64(&FailedLookups, 1)
		addWarning(media, fmt.Sprintf("Failed getting placename (%f, %f): %s", media.Location.Latitude, media.Location.Longitude, err.Error()))
		return
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		atomic.AddInt64(&Failures, 1)
		addWarning(media, fmt.Sprintf("Failed reading body of response (%f, %f): %s", media.Location.Latitude, media.Location.Longitude, err.Error()))
		return
	}

	placenameFromText(media, body)
}

func placenameFromText(media *common.Media, blob []byte) {
	json, err := gabs.ParseJSON(blob)
	if err != nil {
		atomic.AddInt64(&Failures, 1)
		addWarning(media, fmt.Sprintf(
			"Failed deserializing json: %s: ('%s', %f, %f in %s)", err.Error(), blob, media.Location.Latitude, media.Location.Longitude, media.Path))
		return
	}

	if json.Exists("error") {
		atomic.AddInt64(&ServerErrors, 1)
		addWarning(media, fmt.Sprintf("Reverse lookup returned an error: %s", json.Path("error").Data().(string)))
		return
	}

	if !json.Exists("address") {
		atomic.AddInt64(&ServerErrors, 1)
		addWarning(media, fmt.Sprintf("Reverse lookup didn't return an address: %v", json))
		return
	}

	address := json.Path("address")
	generatePlacename(media, address)
}
