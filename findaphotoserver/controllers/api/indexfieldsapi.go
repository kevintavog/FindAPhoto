package api

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-playground/lars"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
)

var fieldsAggregateToStringFormat = map[string]string{
	"cachedlocationdistancemeters": "%1.1f",
	"dayofyear":                    "%1.f",
	"durationseconds":              "%1.3f",
	"lengthinbytes":                "%1.f",
	"height":                       "%1.f",
	"width":                        "%1.f",
}

var fieldsAggregateDisallowed = map[string]bool{
	"location": true,
}

var fieldsAggregateOverride = map[string]string{
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

func IndexFields(c lars.Context) {
	fc := c.(*applicationglobals.FpContext)
	err := fc.Ctx.ParseForm()
	if err != nil {
		panic(&InvalidRequest{message: "parseFormError", err: err})
	}

	response := make(map[string]interface{})
	response["fields"] = getMappedFields()
	fc.WriteResponse(response)
}

func IndexAField(c lars.Context) {
	fc := c.(*applicationglobals.FpContext)
	err := fc.Ctx.ParseForm()
	if err != nil {
		panic(&InvalidRequest{message: "parseFormError", err: err})
	}

	fieldName := fc.Ctx.Request().Form.Get("field")

	values := make(map[string]interface{})
	values["name"] = fieldName
	values["values"] = getTopFieldValues(fieldName, 20)

	response := make(map[string]interface{})
	response["values"] = values

	fc.WriteResponse(response)
}

func getMappedFields() []string {
	client := common.CreateClient()
	results, err := client.GetMapping().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Do(context.TODO())
	if err != nil {
		panic(&InvalidRequest{message: "Failed searching for mappings", err: err})
	}

	allFields := make([]string, 0)

	// We expect a single index...
	index := results[common.MediaIndexName].(map[string]interface{})
	mappings := index["mappings"].(map[string]interface{})
	mediaType := mappings[common.MediaTypeName].(map[string]interface{})
	properties := mediaType["properties"].(map[string]interface{})
	for k, _ := range properties {
		allFields = append(allFields, k)
	}

	sort.Strings(allFields)
	return allFields
}

func getTopFieldValues(fieldName string, maxCount int) []string {
	internalFieldName := getFieldAggregateName(fieldName)
	if _, notSupported := fieldsAggregateDisallowed[internalFieldName]; notSupported {
		return make([]string, 0)
	}

	client := common.CreateClient()
	result, err := client.Search().
		Pretty(true).
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Query(elastic.NewMatchAllQuery()).
		Aggregation("field", elastic.NewTermsAggregation().Field(getFieldAggregateName(fieldName)).Size(maxCount).OrderByCountDesc()).
		Do(context.TODO())

	if err != nil {
		panic(&InvalidRequest{message: "Failed searching for field values", err: err})
	}

	values := make([]string, 0)
	fieldValues, found := result.Aggregations.Terms("field")
	if !found {
		return values
	}

	for _, bucket := range fieldValues.Buckets {

		// datetime needs to be converted to a Date
		if internalFieldName == "datetime" {
			msec := int64(bucket.Key.(float64))
			values = append(values, fmt.Sprintf("%s", time.Unix(msec/1000, 0)))
		} else {
			format, isSet := fieldsAggregateToStringFormat[internalFieldName]
			if !isSet {
				format = "%s"
			}

			values = append(values, fmt.Sprintf(format, bucket.Key))
		}
	}

	return values
}

func getFieldAggregateName(fieldName string) string {
	lowerFieldName := strings.ToLower(fieldName)
	override, ok := fieldsAggregateOverride[lowerFieldName]
	if ok {
		return override
	}
	return lowerFieldName
}
