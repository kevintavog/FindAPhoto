package preparemedia

import (
	"math"
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

func TestExposureTIme(t *testing.T) {
	media := &common.Media{}
	candidate := &common.CandidateFile{}

	candidate.Exif.EXIF.ExposureTime = "3"
	populateExposureTime(media, candidate)
	if media.ExposureTime != 3 {
		t.Fatalf("Expected 3, got %f", media.ExposureTime)
	}
	if media.ExposureTimeString != "3" {
		t.Fatalf("Expected 3, got '%s'", media.ExposureTimeString)
	}

	candidate.Exif.EXIF.ExposureTime = "1/640"
	populateExposureTime(media, candidate)
	expected := 0.001563
	if math.Abs(float64(media.ExposureTime)-expected) > 0.000001 {
		t.Fatalf("Expected %f, got %f", expected, media.ExposureTime)
	}
	if media.ExposureTimeString != "1/640" {
		t.Fatalf("Expected 1/640, got '%s'", media.ExposureTimeString)
	}
}

func TestCameraMakeAndModel(t *testing.T) {
	media := &common.Media{}
	candidate := &common.CandidateFile{}

	testCases := [][]string{
		[]string{"Canon", "Canon EOS REBEL T3i", "Canon", "EOS Rebel T3i"},
		[]string{"CASIO COMPUTER CO.,LTD", "EX-Z55", "Casio", "EX-Z55"},
		[]string{"EASTMAN KODAK COMPANY", "KODAK V550 ZOOM DIGITAL CAMERA", "Kodak", "V550 Zoom"},
		[]string{"EASTMAN KODAK COMPANY", "KODAK EASYSHARE M1093 IS DIGITAL CAMERA", "Kodak", "Easyshare M1093 IS"},
		[]string{"Minolta Co., Ltd.", "DiMAGE A1", "Minolta", "DiMAGE A1"},
		[]string{"NIKON CORPORATION", "NIKON D70s", "Nikon", "D70s"},
		[]string{"OLYMPUS IMAGING CORP.", "u760,S760", "Olympus", "u760,S760"},
		[]string{"OLYMPUS OPTICAL CO.,LTD", "C740UZ", "Olympus", "C740UZ"},
		[]string{"SONY", "DSC-W230", "Sony", "DSC-W230"},
	}

	for _, test := range testCases {
		candidate.Exif.EXIF.Make = test[0]
		candidate.Exif.EXIF.Model = test[1]
		populateCameraMakeAndModel(media, candidate)

		t.Logf("From '%s':'%s' to '%s':'%s'", test[0], test[1], media.CameraMake, media.CameraModel)

		if media.CameraMake != test[2] {
			t.Fatalf("Expected make: '%s', got '%s'", test[2], media.CameraMake)
		}

		if media.CameraModel != test[3] {
			t.Fatalf("Expected model: '%s', got '%s'", test[3], media.CameraModel)
		}
	}

}

func TestCameraMakeAndModelForVideo(t *testing.T) {
	media := &common.Media{}
	candidate := &common.CandidateFile{}

	candidate.Exif.XMP.Make = "CASIO COMPUTER CO.,LTD"
	candidate.Exif.XMP.Model = "EX-Z55"
	populateCameraMakeAndModel(media, candidate)

	if media.CameraMake != "Casio" {
		t.Fatalf("Wrong make: '%s'", media.CameraMake)
	}
	if media.OriginalCameraMake != "CASIO COMPUTER CO.,LTD" {
		t.Fatalf("Wrong original make: '%s'", media.OriginalCameraMake)
	}

	if media.CameraModel != "EX-Z55" {
		t.Fatalf("Wrong model: '%s'", media.CameraModel)
	}
	if media.OriginalCameraModel != "EX-Z55" {
		t.Fatalf("Wrong original model: '%s'", media.OriginalCameraModel)
	}

}

func TestVideoDuration(t *testing.T) {
	media := &common.Media{}
	candidate := &common.CandidateFile{}

	candidate.Exif.Quicktime.Duration = "0.35 s"
	populateDimensions(media, candidate)
	t.Logf("'0.35 s' became %f", media.DurationSeconds)
	if media.DurationSeconds != 0.35 {
		t.Fatalf("Sub-second duration failed: %f", media.DurationSeconds)
	}

	candidate.Exif.Quicktime.Duration = "0:00:42"
	populateDimensions(media, candidate)
	t.Logf("'0:00:42' became %f", media.DurationSeconds)
	if media.DurationSeconds != 42 {
		t.Fatalf("Sub-second duration failed: %f", media.DurationSeconds)
	}
}

func TestFocalLength(t *testing.T) {
	media := &common.Media{}
	candidate := &common.CandidateFile{}

	candidate.Exif.EXIF.FocalLength = "23.7 mm"
	populateFocalLength(media, candidate)
	if media.FocalLengthMm != 23.7 {
		t.Fatalf("FocalLength failed: %f", media.FocalLengthMm)
	}
}

func floatEquals(a, b float64) bool {
	if (a-b) < EPSILON && (b-a) < EPSILON {
		return true
	}
	return false
}
