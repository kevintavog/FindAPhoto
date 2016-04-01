package resolveplacename

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
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
var CachedLocationsUrl = ""

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
		if !lookupInCache(media) {
			resolvePlacename(media)
		}
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

	if CachedLocationsUrl == "" {
		return false
	}

	url := fmt.Sprintf("%s/cache/find-nearest?lat=%f&lon=%f", CachedLocationsUrl, media.Location.Latitude, media.Location.Longitude)
	response, err := http.Get(url)
	if err != nil {
		addWarning(media, fmt.Sprintf("Failed getting cached placename (%f, %f): %s", media.Location.Latitude, media.Location.Longitude, err.Error()))
		return false
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		addWarning(media, fmt.Sprintf("Failed reading body of cached location response (%f, %f): %s", media.Location.Latitude, media.Location.Longitude, err.Error()))
		return false
	}

	json, err := gabs.ParseJSON(body)
	if err != nil {
		addWarning(media, fmt.Sprintf(
			"Failed deserializing cached location json: %s: ('%s', %f, %f in %s)",
			err.Error(), body, media.Location.Latitude, media.Location.Longitude, media.Path))
		return false
	}

	if !json.Exists("matchedLocation") || !json.Exists("placename") {
		return false
	}

	jsonPlacename, err := gabs.ParseJSON(bytes.NewBufferString(json.Path("placename").Data().(string)).Bytes())
	if err != nil {
		addWarning(media, fmt.Sprintf(
			"Failed deserializing cached location placename: %s: ('%s', %f, %f in %s)",
			err.Error(), json.Path("placename").Data().(string), media.Location.Latitude, media.Location.Longitude, media.Path))
		return false
	}
	if !jsonPlacename.Exists(("address")) {
		log.Warn("Unable to find address in '%q'", jsonPlacename)
		return false
	}

	lat, ok := json.Path("matchedLocation.latitude").Data().(float64)
	if !ok {
		log.Warn("Can't find latitude in '%q'", json)
		return false
	}
	lon, ok := json.Path("matchedLocation.longitude").Data().(float64)
	if !ok {
		log.Warn("Can't find longitude")
		return false
	}

	atomic.AddInt64(&PlacenameLookups, 1)
	media.CachedLocationDistanceMeters = int(calcDistance(media.Location.Latitude, media.Location.Longitude, lat, lon))
	generatePlacename(media, jsonPlacename.Path("address"))
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

//// FROM https://gist.github.com/cdipaolo/d3f8db3848278b49db68

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// Distance function returns the distance (in meters) between two points of
//     a given longitude and latitude relatively accurately (using a spherical
//     approximation of the Earth) through the Haversin Distance Formula for
//     great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula
func calcDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}
