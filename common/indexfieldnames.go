package common

import (
	"strings"
)

var fieldsOverride = map[string]string{
	"aperture":         "aperture.value",
	"cameramake":       "cameramake.value",
	"cameramodel":      "cameramodel.value",
	"cityname":         "cityname.value",
	"countrycode":      "countrycode.value",
	"countryname":      "countryname.value",
	"dayname":          "dayname.value",
	"displayname":      "displayname.value",
	"exposureprogram":  "exposureprogram.value",
	"exposuretime":     "exposuretime.value",
	"filename":         "filename.value",
	"flash":            "flash.value",
	"fnumber":          "fnumber.value",
	"focallength":      "focallength.value",
	"hierarchicalname": "hierarchicalname.value",
	"iso":              "iso.value",
	"keywords":         "keywords.value",
	"lensinfo":         "lensinfo.value",
	"lensmodel":        "lensmodel.value",
	"mimetype":         "mimetype.value",
	"monthname":        "monthname.value",
	"path":             "path.value",
	"placename":        "placename.value",
	"sitename":         "sitename.value",
	"statename":        "statename.value",
	"warnings":         "warnings.value",
	"whitebalance":     "whitebalance.value",
}

func GetIndexFieldName(fieldName string) (string, bool) {
	lowerFieldName := strings.ToLower(fieldName)
	override, ok := fieldsOverride[lowerFieldName]
	if ok {
		return override, true
	}
	return lowerFieldName, false
}
