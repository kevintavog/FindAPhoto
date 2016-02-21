package resolveplacename

import (
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
		resolvePlacename(media)
		indexmedia.Enqueue(media)
	}
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
		log.Error("Failed getting placename (%f, %f): %s", media.Location.Latitude, media.Location.Longitude, err.Error())
		return
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		atomic.AddInt64(&Failures, 1)
		log.Error("Failed reading body of response (%f, %f): %s", media.Location.Latitude, media.Location.Longitude, err.Error())
		return
	}

	json, err := gabs.ParseJSON(body)
	if err != nil {
		atomic.AddInt64(&Failures, 1)
		log.Error("Failed deserializing json: %s: ('%s', %f, %f in %s)", err.Error(), body, media.Location.Latitude, media.Location.Longitude, media.Path)
		return
	}

	if json.Exists("error") {
		atomic.AddInt64(&ServerErrors, 1)
		log.Error("Reverse lookup returned an error: %s", json.Path("error").Data().(string))
		return
	}

	if !json.Exists("address") {
		atomic.AddInt64(&ServerErrors, 1)
		log.Error("Reverse lookup didn't return an address: %s", body)
		return
	}

	address := json.Path("address")
	generatePlacename(media, address)
}
