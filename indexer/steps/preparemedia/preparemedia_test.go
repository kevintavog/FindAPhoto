package preparemedia

import (
	"testing"

	"github.com/kevintavog/findaphoto/common"
)

var EPSILON float64 = 0.000001

func TestPopulateLocation(t *testing.T) {
	media := &common.Media{}
	candidate := &common.CandidateFile{}

	candidate.Exif.Composite.GPSPosition = "47 deg 35' 50.66\" N, 122 deg 19' 59.50\" W"

	populateLocation(media, candidate)
	if !floatEquals(media.Location.Latitude, 47.597405) || !floatEquals(media.Location.Longitude, -122.333194) {
		t.Fatalf("Wrong lat/long: %v, %v", media.Location.Latitude, media.Location.Longitude)
	} else {
		t.Logf("Ended up with lat/long: %v, %v", media.Location.Latitude, media.Location.Longitude)
	}
}

func floatEquals(a, b float64) bool {
	if (a-b) < EPSILON && (b-a) < EPSILON {
		return true
	}
	return false
}
