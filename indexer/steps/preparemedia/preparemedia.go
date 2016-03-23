package preparemedia

import (
	"errors"
	"fmt"
	"math"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/indexer/steps/checkthumbnail"
	"github.com/kevintavog/findaphoto/indexer/steps/resolveplacename"
)

const numConsumers = 8

var queue = make(chan *common.CandidateFile, numConsumers)
var waitGroup sync.WaitGroup

func Start() {
	resolveplacename.Start()

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
	resolveplacename.Done()
	resolveplacename.Wait()
}

func Enqueue(candidate *common.CandidateFile) {
	queue <- candidate
}

func dequeue() {
	for candidate := range queue {
		media := populate(candidate)
		resolveplacename.Enqueue(media)
		checkthumbnail.Enqueue(candidate.FullPath, candidate.AliasedPath, media.MimeType)
	}
}

func populate(candidate *common.CandidateFile) *common.Media {
	media := &common.Media{
		Filename:      path.Base(candidate.FullPath),
		Path:          candidate.AliasedPath,
		Signature:     candidate.Signature,
		LengthInBytes: candidate.LengthInBytes,

		MimeType: candidate.Exif.File.MIMEType,

		ApertureValue:   strconv.FormatFloat(float64(candidate.Exif.EXIF.ApertureValue), 'f', -1, 32),
		ExposureProgram: candidate.Exif.EXIF.ExposureProgram,
		Flash:           candidate.Exif.EXIF.Flash,
		FNumber:         strconv.FormatFloat(float64(candidate.Exif.EXIF.FNumber), 'f', -1, 32),
		FocalLength:     candidate.Exif.EXIF.FocalLength,
		WhiteBalance:    candidate.Exif.EXIF.WhiteBalance,
		LensInfo:        candidate.Exif.EXIF.LensInfo,
		LensModel:       candidate.Exif.EXIF.LensModel,
		CameraMake:      candidate.Exif.EXIF.Make,
		CameraModel:     candidate.Exif.EXIF.Model,
	}

	populateIso(media, candidate)
	populateExposureTime(media, candidate)
	populateKeywords(media, candidate)
	populateDateTime(media, candidate)
	populateLocation(media, candidate)
	populateDimensions(media, candidate)

	media.Warnings = candidate.Warnings

	return media
}

func populateDimensions(media *common.Media, candidate *common.CandidateFile) {
	if candidate.Exif.File.ImageWidth != 0 && candidate.Exif.File.ImageHeight != 0 {
		media.Width = candidate.Exif.File.ImageWidth
		media.Height = candidate.Exif.File.ImageHeight
	} else if candidate.Exif.Quicktime.ImageWidth != 0 && candidate.Exif.Quicktime.ImageHeight != 0 {
		media.Width = candidate.Exif.Quicktime.ImageWidth
		media.Height = candidate.Exif.Quicktime.ImageHeight
	}

	if candidate.Exif.Quicktime.Duration != "" {
		// 10.15 s
		tokens := strings.Split(candidate.Exif.Quicktime.Duration, " ")
		if len(tokens) >= 1 {
			v, err := strconv.ParseFloat(tokens[0], 32)
			if err == nil {
				media.DurationSeconds = float32(v)
			}
		}
	}
}

func populateKeywords(media *common.Media, candidate *common.CandidateFile) {
	switch keyType := candidate.Exif.IPTC.Keywords.(type) {
	default:
		candidate.AddWarning(fmt.Sprintf("Unexpected keyword type %T (%q)", keyType, candidate.Exif.IPTC.Keywords))
	case []interface{}:
		for _, s := range candidate.Exif.IPTC.Keywords.([]interface{}) {
			media.Keywords = append(media.Keywords, s.(string))
		}
	case interface{}:
		media.Keywords = []string{candidate.Exif.IPTC.Keywords.(string)}
	case nil:
		// Nothing to do, keywords not present
	}
}

func populateIso(media *common.Media, candidate *common.CandidateFile) {
	switch isoType := candidate.Exif.EXIF.ISO.(type) {
	default:
		candidate.AddWarning(fmt.Sprintf("Unexpected ISO type: %T (%q)", isoType, candidate.Exif.EXIF.ISO))
	case int:
		media.Iso = strconv.FormatInt(int64(candidate.Exif.EXIF.ISO.(int)), 10)
	case float64:
		media.Iso = strconv.FormatFloat(candidate.Exif.EXIF.ISO.(float64), 'f', -1, 64)
	case string:
		s := candidate.Exif.EXIF.ISO.(string)
		re := regexp.MustCompile("[0-9]+")
		media.Iso = re.FindString(s)
	case nil:
		// Nothing to do, no value present
	}
}

func populateExposureTime(media *common.Media, candidate *common.CandidateFile) {
	switch etType := candidate.Exif.EXIF.ExposureTime.(type) {
	default:
		candidate.AddWarning(fmt.Sprintf("Unexpected ExposureTime type: %T", etType))
	case float64:
		media.ExposureTime = strconv.FormatFloat(candidate.Exif.EXIF.ExposureTime.(float64), 'f', -1, 64)
	case string:
		media.ExposureTime = candidate.Exif.EXIF.ExposureTime.(string)
	case nil:
		// Nothing to do, no value present (videos, for instance)
	}
}

func populateDateTime(media *common.Media, candidate *common.CandidateFile) {
	var dateTime time.Time
	var err error

	if candidate.Exif.Quicktime.ContentCreateDate != "" {
		dateTime, err = time.Parse("2006:01:02 15:04:05-07:00", candidate.Exif.Quicktime.ContentCreateDate)
		if err != nil {
			candidate.AddWarning(fmt.Sprintf(
				"Failed parsing ContentCreateDate '%s': %s (in %s)", candidate.Exif.Quicktime.ContentCreateDate, err.Error(), candidate.FullPath))
		}
	}
	if dateTime.IsZero() && len(candidate.Exif.Quicktime.CreateDate) > 0 {
		// UTC according to spec - no timezone like there is for 'ContentCreateDate'
		dateTime, err = time.Parse("2006:01:02 15:04:05", candidate.Exif.Quicktime.CreateDate)
		if err != nil {
			candidate.AddWarning(fmt.Sprintf(
				"Failed parsing CreateDate '%s': %s (in %s)", candidate.Exif.Quicktime.CreateDate, err.Error(), candidate.FullPath))
		}
	}

	if dateTime.IsZero() {
		exifDateTime := candidate.Exif.EXIF.CreateDate
		if exifDateTime == "" {
			exifDateTime = candidate.Exif.EXIF.DateTimeOriginal
		}
		if exifDateTime == "" {
			exifDateTime = candidate.Exif.EXIF.ModifyDate
		}
		if exifDateTime != "" {
			// No timezone - and the spec doesn't specify.
			dateTime, err = time.ParseInLocation("2006:01:02 15:04:05", exifDateTime, time.Local)
			if err != nil {
				candidate.AddWarning(fmt.Sprintf(
					"Failed parsing '%s': %s (in %s)", exifDateTime, err.Error(), candidate.FullPath))
			}
		}
	}

	if dateTime.IsZero() {
		candidate.AddWarning("No usable date in EXIF, using file timestamp")
		dateTime, err = time.Parse("2006:01:02 15:04:05-07:00", candidate.Exif.File.FileModifyDate)
		if err != nil {
			candidate.AddWarning(fmt.Sprintf(
				"Failed parsing File.FileModifyDate '%s': %s (in %s)", candidate.Exif.File.FileModifyDate, err.Error(), candidate.FullPath))
		}
	}

	if candidate.Exif.File.FileModifyDate != "" {
		fileModifyDateTime, err := time.Parse("2006:01:02 15:04:05-07:00", candidate.Exif.File.FileModifyDate)
		if err != nil {
			candidate.AddWarning(fmt.Sprintf(
				"Failed parsing File.FileModifyDate '%s': %s (in %s)", candidate.Exif.File.FileModifyDate, err.Error(), candidate.FullPath))
		} else {
			// Allow a small amount of difference to account for somefile systems (FAT) that have poor timestamp granularity
			if math.Abs(fileModifyDateTime.Sub(dateTime).Seconds()) > 2 {
				candidate.AddWarning(fmt.Sprintf(
					"File modify date does not match media date (%q - %q)", fileModifyDateTime, dateTime))
			}
		}
	}

	media.Date = dateTime.Format("20060102")
	media.DateTime = dateTime
	media.MonthName = dateTime.Month().String() + " " + dateTime.Month().String()[:3]
	media.DayName = dateTime.Weekday().String() + " " + dateTime.Weekday().String()[:3]
}

func populateLocation(media *common.Media, candidate *common.CandidateFile) {
	if candidate.Exif.Composite.GPSPosition != "" {
		if populateWithGpsPosition(media, candidate, candidate.Exif.Composite.GPSPosition) {
			return
		}
	}

	populateWithGpsAndRef(media, candidate, candidate.Exif.EXIF.GPSLatitude, candidate.Exif.EXIF.GPSLatitudeRef, candidate.Exif.EXIF.GPSLongitude, candidate.Exif.EXIF.GPSLongitudeRef)
}

func populateWithGpsPosition(media *common.Media, candidate *common.CandidateFile, gpsPosition string) bool {
	// 47 deg 37' 23.06" N, 122 deg 20' 59.08" W
	// 47 deg 35' 50.66" N, 122 deg 19' 59.50" W == 47.597389 -122.333194
	latAndLongTokens := strings.Split(gpsPosition, ",")
	if len(latAndLongTokens) != 2 {
		candidate.AddWarning(fmt.Sprintf("Unsupported GPSPosition: '%s'", gpsPosition))
		return false
	}

	latitudeValue := strings.Trim(latAndLongTokens[0], " ")
	latitudeTokens := strings.Split(latitudeValue, " ")
	if len(latitudeTokens) != 5 {
		candidate.AddWarning(fmt.Sprintf("Unsupported GPSPosition (latitude): '%s'", gpsPosition))
		return false
	}

	longitudeValue := strings.Trim(latAndLongTokens[1], " ")
	longitudeTokens := strings.Split(longitudeValue, " ")
	if len(longitudeTokens) != 5 {
		candidate.AddWarning(fmt.Sprintf(
			"Unsupported GPSPosition (longitude): '%s' - %s - %s", gpsPosition, latAndLongTokens[1], strings.Join(longitudeTokens, ", ")))
		return false
	}

	var latRef string
	switch latitudeTokens[4] {
	case "N":
		latRef = "North"
	case "S":
		latRef = "South"
	default:
		candidate.AddWarning(fmt.Sprintf("Unsupported GPSPosition (latitude ref): '%s'", gpsPosition))
		return false
	}
	var lonRef string
	switch longitudeTokens[4] {
	case "W":
		lonRef = "West"
	case "E":
		lonRef = "East"
	default:
		candidate.AddWarning(fmt.Sprintf("Unsupported GPSPosition (longitude ref): '%s'", gpsPosition))
		return false
	}

	return populateWithGpsAndRef(media, candidate, strings.TrimRight(latitudeValue, "NSEW "), latRef, strings.TrimRight(longitudeValue, "NSEW "), lonRef)
}

func populateWithGpsAndRef(media *common.Media, candidate *common.CandidateFile, gpsLatitude, gpsLatitudeRef, gpsLongitude, gpsLongitudeRef string) bool {
	// all or nothing for location
	if gpsLatitude == "" && gpsLatitudeRef == "" && gpsLongitude == "" && gpsLongitudeRef == "" {
		return false
	}

	location := fmt.Sprintf("%s %s, %s %s", gpsLatitude, gpsLatitudeRef, gpsLongitude, gpsLongitudeRef)
	if gpsLatitude == "" || gpsLatitudeRef == "" || gpsLongitude == "" || gpsLongitudeRef == "" {
		candidate.AddWarning(fmt.Sprintf("Ignoring poorly formed location: %s", location))
		return false
	}
	if (gpsLatitudeRef != "North" && gpsLatitudeRef != "South") || (gpsLongitudeRef != "West" && gpsLongitudeRef != "East") {
		candidate.AddWarning(fmt.Sprintf("Ignoring poorly formed location - invalid reference: '%s', '%s' (%s)", gpsLatitudeRef, gpsLongitudeRef, location))
		return false
	}

	latFloat, laErr := dmsToFloat(gpsLatitude)
	lonFloat, loErr := dmsToFloat(gpsLongitude)
	if laErr != nil || loErr != nil {
		candidate.AddWarning(fmt.Sprintf("Ignoring location, unable to parse lat/lon %q, %q (%s)", laErr, loErr, location))
		return false
	}

	if gpsLatitudeRef == "South" {
		latFloat = latFloat * -1.0
	}
	if gpsLongitudeRef == "West" {
		lonFloat = lonFloat * -1.0
	}

	media.Location.Latitude = latFloat
	media.Location.Longitude = lonFloat
	return true
}

func dmsToFloat(dms string) (float64, error) {
	// 47 deg 37' 23.06"
	// 122 deg 20' 59.08"

	tokens := strings.Split(dms, " ")
	if len(tokens) == 4 {
		strMinutes := tokens[2][:len(tokens[2])-1]
		strSeconds := tokens[3][:len(tokens[3])-1]

		degrees, dErr := strconv.Atoi(tokens[0])
		minutes, mErr := strconv.Atoi(strMinutes)
		seconds, sErr := strconv.ParseFloat(strSeconds, 64)

		if dErr == nil && mErr == nil && sErr == nil {
			return float64(degrees) + (float64(minutes) / 60.0) + (seconds / 3600.0), nil
		} else {
			return math.NaN(), errors.New(fmt.Sprintf("Unable to convert: %q, %q, %q", dErr, mErr, sErr))
		}
	}
	return math.NaN(), errors.New(fmt.Sprintf("Invalid DMS (wrong number of tokens): %s", dms))
}
