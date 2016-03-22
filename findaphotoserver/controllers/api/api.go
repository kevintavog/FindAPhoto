package api

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-playground/lars"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
	"github.com/kevintavog/findaphoto/findaphotoserver/controllers/files"
	"github.com/kevintavog/findaphoto/findaphotoserver/search"
)

type InternalError struct {
	message string
	err     error
}

type InvalidRequest struct {
	message string
	err     error
}

func (ir *InvalidRequest) Error() string {
	return ir.message
}

func (ie *InternalError) Error() string {
	return ie.message
}

func ConfigureRouting(l *lars.LARS) {
	l.Use(handleErrors)

	api := l.Group("/api")
	api.Get("/search", Search)
	api.Get("/nearby", Nearby)
	api.Get("/by-day", ByDay)
}

func handleErrors(c lars.Context) {
	fc := c.(*applicationglobals.FpContext)
	defer func() {
		if r := recover(); r != nil {
			if ie, ok := r.(*InternalError); ok {
				fc.Error(http.StatusInternalServerError, "InternalError", ie.Error(), ie.err)
			} else if ir, ok := r.(*InvalidRequest); ok {
				fc.Error(http.StatusBadRequest, "InvalidRequest", ir.Error(), ir.err)
			} else if e, ok := r.(runtime.Error); ok {
				fc.Error(http.StatusInternalServerError, "UnhandledError", "", e)
			} else if s, ok := r.(string); ok {
				fc.Error(http.StatusInternalServerError, "UnhandledError", s, nil)
			} else {
				fc.Error(http.StatusInternalServerError, "UnhandledError", fmt.Sprintf("%v", r), nil)
			}
		}
	}()

	fc.Ctx.Next()
}

func float64FromQuery(ctx *lars.Ctx, name string) float64 {
	s := ctx.Request().Form.Get(name)
	if s != "" {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic(&InvalidRequest{message: fmt.Sprintf("'%s' is not a float: %s", name, s)})
		}
		return v
	}

	panic(&InvalidRequest{message: fmt.Sprintf("'%s' is missing from the query parameter", name)})
}

func intFromQuery(ctx *lars.Ctx, name string, defaultValue int) int {
	s := ctx.Request().Form.Get(name)
	if s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			panic(&InvalidRequest{message: fmt.Sprintf("'%s' is not an int: %s", name, s)})
		}
		return v
	}
	return defaultValue
}

func filterResults(searchResult *search.SearchResult, propertiesFilter []string) map[string]interface{} {

	filtered := make(map[string]interface{})
	filtered["totalMatches"] = searchResult.TotalMatches
	filtered["resultCount"] = searchResult.ResultCount
	filtered["groups"] = filteredGroups(searchResult.Groups, propertiesFilter)

	return filtered
}

func filteredGroups(groups []*search.SearchGroup, propertiesFilter []string) interface{} {

	list := make([]map[string]interface{}, len(groups))
	for index, group := range groups {
		listItem := make(map[string]interface{})
		list[index] = listItem
		listItem["name"] = group.Name
		listItem["items"] = filteredItems(group.Items, propertiesFilter)
	}
	return list
}

func filteredItems(items []*common.Media, propertiesFilter []string) interface{} {

	list := make([]map[string]interface{}, len(items))

	for mediaIndex, media := range items {
		listItem := make(map[string]interface{})
		list[mediaIndex] = listItem
		for _, prop := range propertiesFilter {
			listItem[prop] = property(prop, media)
		}
	}

	return list
}

func property(name string, media *common.Media) interface{} {
	switch strings.ToLower(name) {
	case "city":
		return media.LocationCityName
	case "createddate":
		return media.DateTime
	case "id":
		return media.Path
	case "imagename":
		return media.Filename
	case "keywords":
		return media.Keywords
	case "latitude":
		return media.Location.Latitude
	case "locationdetailedname":
		return media.LocationPlaceName
	case "locationname":
		return media.LocationHierarchicalName
	case "longitude":
		return media.Location.Longitude
	case "mediatype":
		return media.MediaType()
	case "mediaurl":
		return files.ToMediaUrl(media.Path)
	case "mimetype":
		return media.MimeType
	case "path":
		return media.Path
	case "slideurl":
		return files.ToSlideUrl(media.Path)
	case "thumburl":
		return files.ToThumbUrl(media.Path)
	case "warnings":
		return media.Warnings
	}

	panic(&InvalidRequest{message: fmt.Sprintf("Unknown property: '%s'", name)})
}
