package common

import (
	"strings"
	"time"
)

var MediaIndexName = "media-index"

const MediaTypeName = "media"
const (
	MediaTypeImage   = "image"
	MediaTypeVideo   = "video"
	MediaTypeUnknown = "unknown"
)

func (m *Media) MediaType() string {
	var mediaType = strings.Split(m.MimeType, "/")
	if len(mediaType) < 1 {
		return MediaTypeUnknown
	}

	switch strings.ToLower(mediaType[0]) {
	case "video":
		return MediaTypeVideo
	case "image":
		return MediaTypeImage
	default:
		return MediaTypeUnknown
	}
}

type Media struct {
	Signature     string `json:"signature"`
	Filename      string `json:"filename"`
	Path          string `json:"path"`
	LengthInBytes int64  `json:"lengthinbytes"`

	MimeType        string  `json:"mimetype,omitempty"`
	Width           int     `json:"width,omitempty"`
	Height          int     `json:"height,omitempty"`
	DurationSeconds float32 `json:"durationseconds,omitempty"`

	// EXIF info
	ApertureValue       float32 `json:"aperture,omitempty"`
	ExposureProgram     string  `json:"exposureprogram,omitempty"`
	ExposureTime        float32 `json:"exposuretime,omitempty"`
	ExposureTimeString  string  `json:"exposuretimestring,omitempty"`
	Flash               string  `json:"flash,omitempty"`
	FNumber             float32 `json:"fnumber,omitempty"`
	FocalLengthMm       float32 `json:"focallengthmm,omitempty"`
	Iso                 int     `json:"iso,omitempty"`
	WhiteBalance        string  `json:"whitebalance,omitempty"`
	LensInfo            string  `json:"lensinfo,omitempty"`
	LensModel           string  `json:"lensmodel,omitempty"`
	CameraMake          string  `json:"cameramake,omitempty"`
	CameraModel         string  `json:"cameramodel,omitempty"`
	OriginalCameraMake  string  `json:"originalcameramake,omitempty"`
	OriginalCameraModel string  `json:"originalcameramodel,omitempty"`

	// For arrays - see here for mappings & searching: http://stackoverflow.com/questions/26258292/querystring-search-on-array-elements-in-elastic-search
	Keywords []string `json:"keywords,omitempty"`

	// Auto-classified
	Tags *[]string `json:"tags,omitempty"`

	// Location
	Location *GeoPoint `json:"location,omitempty"`

	// Placename, from the reverse coding of the location
	LocationCountryName          string `json:"countryname,omitempty"`
	LocationCountryCode          string `json:"countrycode,omitempty"`
	LocationStateName            string `json:"statename,omitempty"`
	LocationCityName             string `json:"cityname,omitempty"`
	LocationSiteName             string `json:"sitename,omitempty"`
	LocationPlaceName            string `json:"placename,omitempty"`
	LocationHierarchicalName     string `json:"hierarchicalname,omitempty"`
	LocationDisplayName          string `json:"displayname,omitempty"`
	CachedLocationDistanceMeters int    `json:"cachedlocationdistancemeters,omitempty"` // # of meters away from stored location the placename came from (due to using caching server)

	// Date related fields
	DateTime  time.Time `json:"datetime"`  // 2009-06-15T13:45:30.0000000-07:00 'round trip pattern'
	Date      string    `json:"date"`      // yyyyMMdd - for aggregating by date
	DayName   string    `json:"dayname"`   // (Wed, Wednesday)
	MonthName string    `json:"monthname"` // (Apr, April)
	DayOfYear int       `json:"dayofyear"` // Index of the day in the year, to help with byday searches (1-366; Jan/1 = 1, Feb/29 =60, Mar/1 = 61)

	Warnings []string `json:"warnings,omitempty"`
}

type GeoPoint struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

type CandidateFile struct {
	FullPath      string
	AliasedPath   string
	Signature     string
	LengthInBytes int64
	Exif          ExifOutput
	Warnings      []string
}

type ExifOutput struct {
	SourceFile string
	File       ExifOutputFile
	EXIF       ExifOutputExif
	IPTC       ExifOutputIptc
	Quicktime  ExifOutputQuicktime
	XMP        ExifOutputXmp
	Composite  ExifOutputComposite
}

type ExifOutputFile struct {
	MIMEType       string
	ImageHeight    int
	ImageWidth     int
	FileModifyDate string
}

type ExifOutputExif struct {
	ApertureValue    float32
	CreateDate       string
	DateTimeOriginal string
	ModifyDate       string
	ExposureProgram  string
	ExposureTime     interface{} // Sigh - sometimes a number, sometimes a string - 1 is a number, while "1/200" is a string. Probably an exiftool'ism
	Flash            string
	FNumber          float32
	FocalLength      string
	GPSLatitudeRef   string
	GPSLatitude      string
	GPSLongitudeRef  string
	GPSLongitude     string
	ISO              interface{} // Most cameras use an int, some a string (!)
	LensInfo         string
	LensModel        string
	Make             string
	Model            string
	WhiteBalance     string
}

type ExifOutputQuicktime struct {
	ContentCreateDate string
	CreateDate        string
	ModifyDate        string
	ImageWidth        int
	ImageHeight       int
	Duration          string
}

type ExifOutputXmp struct {
	Subject interface{} // Some are []string - others are string. Exiftool seems to be the source
	Make    string
	Model   string
}

type ExifOutputComposite struct {
	GPSPosition string
}

type ExifOutputIptc struct {
	Keywords interface{} // Some are []string - others are string. Exiftool seems to be the source
}

func (cf *CandidateFile) AddWarning(warning string) {
	cf.Warnings = append(cf.Warnings, warning)
}
