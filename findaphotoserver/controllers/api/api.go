package api

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-playground/lars"

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
			logStack := true
			if ie, ok := r.(*InternalError); ok {
				fc.Error(http.StatusInternalServerError, "InternalError", ie.Error(), ie.err)
			} else if ir, ok := r.(*InvalidRequest); ok {
				logStack = false
				fc.Error(http.StatusBadRequest, "InvalidRequest", ir.Error(), ir.err)
			} else if e, ok := r.(runtime.Error); ok {
				fc.Error(http.StatusInternalServerError, "UnhandledError", "", e)
			} else if s, ok := r.(string); ok {
				fc.Error(http.StatusInternalServerError, "UnhandledError", s, nil)
			} else {
				fc.Error(http.StatusInternalServerError, "UnhandledError", fmt.Sprintf("%v", r), nil)
			}

			if logStack {
				buf := make([]byte, 1<<16)
				stackSize := runtime.Stack(buf, false)
				fc.AppContext.FieldLogger.Add("stack", string(buf[0:stackSize]))
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

func boolFromQuery(ctx *lars.Ctx, name string, defaultValue bool) bool {
	s := ctx.Request().Form.Get(name)
	if s != "" {
		v, err := strconv.ParseBool(s)
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

func filteredItems(items []*search.MediaHit, propertiesFilter []string) interface{} {

	list := make([]map[string]interface{}, len(items))

	for mediaIndex, mh := range items {
		listItem := make(map[string]interface{})
		list[mediaIndex] = listItem
		for _, prop := range propertiesFilter {
			v := property(prop, mh)
			if v != nil {
				listItem[prop] = v
			}
		}
	}

	return list
}

func property(name string, mh *search.MediaHit) interface{} {
	switch strings.ToLower(name) {
	case "city":
		return mh.Media.LocationCityName
	case "createddate":
		return mh.Media.DateTime
	case "distancekm":
		if mh.DistanceKm != nil {
			return mh.DistanceKm
		}
		return nil
	case "id":
		return mh.Media.Path
	case "imagename":
		return mh.Media.Filename
	case "keywords":
		return mh.Media.Keywords
	case "latitude":
		if mh.Media.Location == nil {
			return nil
		}
		return mh.Media.Location.Latitude
	case "locationdetailedname":
		if mh.Media.Location == nil {
			return nil
		}
		return mh.Media.LocationPlaceName
	case "locationname":
		if mh.Media.Location == nil {
			return nil
		}
		return mh.Media.LocationHierarchicalName
	case "longitude":
		if mh.Media.Location == nil {
			return nil
		}
		return mh.Media.Location.Longitude
	case "mediatype":
		return mh.Media.MediaType()
	case "mediaurl":
		return files.ToMediaUrl(mh.Media.Path)
	case "mimetype":
		return mh.Media.MimeType
	case "path":
		return mh.Media.Path
	case "slideurl":
		return files.ToSlideUrl(mh.Media.Path)
	case "thumburl":
		return files.ToThumbUrl(mh.Media.Path)
	case "warnings":
		return mh.Media.Warnings
	}

	panic(&InvalidRequest{message: fmt.Sprintf("Unknown property: '%s'", name)})
}
