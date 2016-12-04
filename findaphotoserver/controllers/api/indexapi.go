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

var FindAPhotoVersionNumber string

type PathAndDate struct {
	Path        string     `json:"path,omitempty"`
	LastIndexed *time.Time `json:"lastIndexed,omitempty"`
}

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

func Index(c lars.Context) {
	fc := c.(*applicationglobals.FpContext)
	err := fc.Ctx.ParseForm()
	if err != nil {
		panic(&InvalidRequest{message: "parseFormError", err: err})
	}
	propertiesFilter := getStatsPropertiesFilter(fc.Ctx.Request().Form.Get("properties"))

	fc.AppContext.FieldLogger.Time("index", func() {
		props := make(map[string]interface{})
		client := common.CreateClient()

		for _, name := range propertiesFilter {
			v := getValue(name, client)
			if v != nil {
				props[name] = v
			}
		}

		fc.WriteResponse(props)
	})
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

func getValue(name string, client *elastic.Client) interface{} {
	switch strings.ToLower(name) {

	case "duplicatecount":
		return getDuplicateCount(client)

	case "imagecount":
		return getCountsSearch(client, "mimetype:image*")

	case "paths":
		return getAliasedPaths()

	case "versionnumber":
		return FindAPhotoVersionNumber

	case "videocount":
		return getCountsSearch(client, "mimetype:video*")

	case "warningcount":
		return getCountsSearch(client, "warnings:*")
	}

	panic(&InvalidRequest{message: fmt.Sprintf("Unknown property: '%s'", name)})
}

func getAliasedPaths() []PathAndDate {
	allPaths := make([]PathAndDate, 0)

	common.VisitAllPaths(func(alias common.AliasDocument) {
		pd := &PathAndDate{
			Path: alias.Path,
		}
		if !alias.DateLastIndexed.IsZero() {
			pd.LastIndexed = &alias.DateLastIndexed
		}
		allPaths = append(allPaths, *pd)
	})
	return allPaths
}

func getCountsSearch(client *elastic.Client, query string) int64 {
	search := client.Search().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Query(elastic.NewQueryStringQuery(query)).
		From(0).
		Size(1).
		Pretty(true)

	result, err := search.Do(context.TODO())
	if err != nil {
		panic(&InvalidRequest{message: fmt.Sprintf("Failed searching for count (%s)", query), err: err})
	}
	return result.TotalHits()
}

func getDuplicateCount(client *elastic.Client) int64 {
	count, err := client.Count().
		Index(common.MediaIndexName).
		Type(common.DuplicateTypeName).
		Do(context.TODO())
	if err != nil {
		panic(&InvalidRequest{message: "Failed searching for duplicate count", err: err})
	}
	return count
}

func getStatsPropertiesFilter(propertiesFilter string) []string {
	if propertiesFilter == "" {
		return []string{"versionNumber"}
	}
	return strings.Split(propertiesFilter, ",")
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
	internalFieldName, _ := common.GetIndexFieldName(fieldName)
	if _, notSupported := fieldsAggregateDisallowed[internalFieldName]; notSupported {
		return make([]string, 0)
	}

	indexFieldName, _ := common.GetIndexFieldName(fieldName)

	client := common.CreateClient()
	result, err := client.Search().
		Pretty(true).
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Query(elastic.NewMatchAllQuery()).
		Aggregation("field", elastic.NewTermsAggregation().Field(indexFieldName).Size(maxCount).OrderByCountDesc()).
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
