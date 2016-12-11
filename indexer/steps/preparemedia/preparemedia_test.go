package preparemedia

import (
	"strings"
	"testing"
	"time"

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

func TestVideoCreateDate(t *testing.T) {
	media := &common.Media{}
	candidate := &common.CandidateFile{}

	candidate.Exif.Quicktime.CreateDate = "2016:12:10 01:25:54"

	populateDateTime(media, candidate)
	t.Logf("Ended up with %v and %v", media.Date, media.DateTime)
	if strings.Compare(media.Date, "20161209") != 0 {
		t.Fatalf("Wrong date: %v", media.Date)
	}

	expectedTime, err := time.Parse("2006-01-02 15:04:05", "2016-12-10 01:25:54")
	if err != nil {
		t.Fatalf("Failed parsing test date: %v", err)
	}

	expectedTime = expectedTime.In(time.Local)

	if !media.DateTime.Equal(expectedTime) {
		t.Fatalf("Wrong date/time: %v (expected %v)", media.DateTime, expectedTime)
	}
}

func floatEquals(a, b float64) bool {
	if (a-b) < EPSILON && (b-a) < EPSILON {
		return true
	}
	return false
}
