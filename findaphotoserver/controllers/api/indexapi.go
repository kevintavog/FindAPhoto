package api

import (
	"fmt"
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

func Index(c lars.Context) {
	fc := c.(*applicationglobals.FpContext)
	err := fc.Ctx.ParseForm()
	if err != nil {
		panic(&InvalidRequest{message: "parseFormError", err: err})
	}
	propertiesFilter := getStatsPropertiesFilter(fc.Ctx.Request().Form.Get("properties"))

	fc.AppContext.FieldLogger.Time("index", func() {
		props := make(map[string]interface{})

		for _, name := range propertiesFilter {
			v := getValue(name)
			if v != nil {
				props[name] = v
			}
		}

		fc.WriteResponse(props)
	})
}

func getValue(name string) interface{} {
	switch strings.ToLower(name) {

	case "paths":
		return getAliasedPaths()

	case "imagecount":
		return getCountsSearch("mimetype:image*")

	case "versionnumber":
		return FindAPhotoVersionNumber

	case "videocount":
		return getCountsSearch("mimetype:video*")

	case "warningcount":
		return getCountsSearch("warnings:*")
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

func getCountsSearch(query string) int64 {
	client := common.CreateClient()
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

func getStatsPropertiesFilter(propertiesFilter string) []string {
	if propertiesFilter == "" {
		return []string{"versionNumber"}
	}
	return strings.Split(propertiesFilter, ",")
}
