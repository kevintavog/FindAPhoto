package resolveplacename

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
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

var LocationLookupUrl = ""

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

func addWarning(media *common.Media, warning string) {
	log.Warn("%q - %q; %f, %f", media.Path, warning, media.Location.Latitude, media.Location.Longitude)
	media.Warnings = append(media.Warnings, warning)
}

func resolvePlacename(media *common.Media) {
	if media.Location == nil {
		return
	}
	if media.Location.Latitude == 0 && media.Location.Longitude == 0 {
		return
	}

	atomic.AddInt64(&PlacenameLookups, 1)
	url := fmt.Sprintf("%s/api/v1/name?lat=%f&lon=%f", LocationLookupUrl, media.Location.Latitude, media.Location.Longitude)
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

	if !json.Exists("fullDescription") {
		atomic.AddInt64(&ServerErrors, 1)
		addWarning(media, fmt.Sprintf("Reverse lookup didn't return a description: %v", json))
		return
	}

	atomic.AddInt64(&PlacenameLookups, 1)
	media.CachedLocationDistanceMeters = 0

	val, ok := json.Path("countryCode").Data().(string)
	if ok {
		media.LocationCountryCode = val
	}
	val, ok = json.Path("countryName").Data().(string)
	if ok {
		media.LocationCountryName = val
	}
	val, ok = json.Path("state").Data().(string)
	if ok {
		media.LocationStateName = val
	}
	val, ok = json.Path("city").Data().(string)
	if ok {
		media.LocationCityName = val
	}
	sites := []string{}
	jsonSites, err := json.Search("sites").Children()
	for _, s := range jsonSites {
		sites = append(sites, s.Data().(string))
	}
	media.LocationSiteName = strings.Join(sites, ", ")

	media.LocationHierarchicalName = joinSkipEmpty(",", media.LocationSiteName, media.LocationCityName, media.LocationStateName, media.LocationCountryName)
	media.LocationPlaceName = media.LocationHierarchicalName
	media.LocationDisplayName = media.LocationHierarchicalName
}

func joinSkipEmpty(separator string, items ...string) string {
	list := []string{}
	for _, s := range items {
		if s != "" {
			list = append(list, s)
		}
	}
	return strings.Join(list, ", ")
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
