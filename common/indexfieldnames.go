package common

import (
	"strings"
)

var fieldsOverride = map[string]string{
	"cameramake":          "cameramake.value",
	"cameramodel":         "cameramodel.value",
	"cityname":            "cityname.value",
	"countrycode":         "countrycode.value",
	"countryname":         "countryname.value",
	"dayname":             "dayname.value",
	"displayname":         "displayname.value",
	"exposureprogram":     "exposureprogram.value",
	"exposuretimestring":  "exposuretimestring.value",
	"filename":            "filename.value",
	"flash":               "flash.value",
	"hierarchicalname":    "hierarchicalname.value",
	"keywords":            "keywords.value",
	"lensinfo":            "lensinfo.value",
	"lensmodel":           "lensmodel.value",
	"mimetype":            "mimetype.value",
	"monthname":           "monthname.value",
	"originalcameramake":  "originalcameramake.value",
	"originalcameramodel": "originalcameramodel.value",
	"path":                "path.value",
	"placename":           "placename.value",
	"sitename":            "sitename.value",
	"statename":           "statename.value",
	"tags":                "tags.value",
	"warnings":            "warnings.value",
	"whitebalance":        "whitebalance.value",
}

func GetIndexFieldName(fieldName string) (string, bool) {
	lowerFieldName := strings.ToLower(fieldName)
	override, ok := fieldsOverride[lowerFieldName]
	if ok {
		return override, true
	}
	return lowerFieldName, false
}
