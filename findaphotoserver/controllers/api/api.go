package api

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-playground/lars"
	"gopkg.in/olivere/elastic.v5"

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
	return ir.message + " -- " + getDetailedErrorMessage(ir.err)
}

func (ie *InternalError) Error() string {
	return ie.message + " -- " + getDetailedErrorMessage(ie.err)
}

func ConfigureRouting(l *lars.LARS) {
	l.Use(handleErrors)

	api := l.Group("/api")
	api.Get("/search", Search)
	api.Get("/nearby", Nearby)
	api.Get("/by-day", ByDay)

	index := api.Group("/index")
	index.Get("/fields/:field", IndexFieldValues)
	index.Get("/fields", IndexFields)
	index.Get("/duplicates", DuplicateMedia)
	index.Get("/", Index)
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

func getDetailedErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	if ee, ok := err.(*elastic.Error); ok {
		if ee.Details != nil && ee.Details.CausedBy != nil {
			return fmt.Sprintf("%s: %s", ee.Error(), ee.Details.CausedBy["reason"])
		}
	}
	return err.Error()
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

func optionalFloat64FromQuery(ctx *lars.Ctx, name string, defaultValue float64) float64 {
	s := ctx.Request().Form.Get(name)
	if s != "" {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic(&InvalidRequest{message: fmt.Sprintf("'%s' is not a float: %s", name, s)})
		}
		return v
	}

	return defaultValue
}

func intFromQuery(ctx *lars.Ctx, name string, defaultValue int) int {
	s := ctx.Request().Form.Get(name)
	if s != "" {
		return intFromString(name, s)
	}
	return defaultValue
}

func intFromString(name string, contents string) int {
	v, err := strconv.Atoi(contents)
	if err != nil {
		panic(&InvalidRequest{message: fmt.Sprintf("'%s' is not an int: %s", name, contents)})
	}
	return v
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

	if searchResult.NextAvailableByDay != nil {
		filtered["nextAvailableByDay"] = filterByDay(searchResult.NextAvailableByDay)
	}
	if searchResult.PreviousAvailableByDay != nil {
		filtered["previousAvailableByDay"] = filterByDay(searchResult.PreviousAvailableByDay)
	}

	if searchResult.Categories != nil {
		filtered["categories"] = convertCategories(searchResult.Categories)
	}

	return filtered
}

func filterByDay(byday *search.ByDayResult) map[string]interface{} {
	result := make(map[string]interface{})
	result["month"] = byday.Month
	result["day"] = byday.Day
	return result
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
	case "aperture":
		return mh.Media.ApertureValue
	case "cameramake":
		return mh.Media.CameraMake
	case "cameramodel":
		return mh.Media.CameraModel
	case "city":
		return mh.Media.LocationCityName
	case "createddate":
		return mh.Media.DateTime
	case "distancekm":
		if mh.DistanceKm != nil {
			return mh.DistanceKm
		}
		return nil
	case "durationseconds":
		return mh.Media.DurationSeconds
	case "exposeureprogram":
		return mh.Media.ExposureProgram
	case "exposuretime":
		return mh.Media.ExposureTime
	case "flash":
		return mh.Media.Flash
	case "fnumber":
		return mh.Media.FNumber
	case "focallength":
		return mh.Media.FocalLength
	case "height":
		return mh.Media.Height
	case "id":
		return mh.Media.Path
	case "iso":
		return mh.Media.Iso
	case "imagename":
		return mh.Media.Filename
	case "keywords":
		return mh.Media.Keywords
	case "latitude":
		if mh.Media.Location == nil {
			return nil
		}
		return mh.Media.Location.Latitude
	case "lensinfo":
		return mh.Media.LensInfo
	case "lensmodel":
		return mh.Media.LensModel
	case "locationdisplayname":
		if mh.Media.Location == nil {
			return nil
		}
		return mh.Media.LocationDisplayName
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
	case "signature":
		return mh.Media.Signature
	case "slideurl":
		return files.ToSlideUrl(mh.Media.Path)
	case "thumburl":
		return files.ToThumbUrl(mh.Media.Path)
	case "warnings":
		return mh.Media.Warnings
	case "width":
		return mh.Media.Width
	}

	panic(&InvalidRequest{message: fmt.Sprintf("Unknown property: '%s'", name)})
}

func populateCategoryOptions(fc *applicationglobals.FpContext, categoryOptions *search.CategoryOptions) {

	// TODO: Is this a LARS bug? The examples don't show a call to ParseForm being required to get query parameters
	// Even with this, the query param example isn't working for me
	err := fc.Ctx.ParseForm()
	if err != nil {
		panic(&InvalidRequest{message: "parseFormError", err: err})
	}

	categories := fc.Ctx.Request().Form.Get("categories") // Grumble grumble - should be 'query := c.Param("categories")'
	if len(categories) > 0 {
		for _, c := range strings.Split(categories, ",") {
			switch strings.ToLower(c) {
			case "keywords":
				categoryOptions.KeywordCount = 10
			case "placename":
				categoryOptions.PlacenameCount = 10
			case "date":
				categoryOptions.DateCount = 10
			case "year":
				categoryOptions.YearCount = 10
			default:
				panic(&InvalidRequest{message: fmt.Sprintf("Unknown category: '%s'", c)})
			}
		}
	}
}

func convertCategories(categories []*search.CategoryResult) interface{} {
	list := make([]map[string]interface{}, len(categories))

	for index, category := range categories {
		listItem := make(map[string]interface{})
		list[index] = listItem
		listItem["field"] = category.Field
		listItem["details"] = convertCategoryDetails(category.Details)
	}

	return list
}

func convertCategoryDetails(details []*search.CategoryDetailResult) interface{} {
	list := make([]map[string]interface{}, len(details))

	for index, detail := range details {
		listItem := make(map[string]interface{})
		list[index] = listItem
		listItem["value"] = detail.Value
		listItem["count"] = detail.Count
		if detail.Field != nil {
			listItem["field"] = detail.Field
		}

		if len(detail.Children) > 0 {
			listItem["details"] = convertCategoryDetails(detail.Children)
		}
	}

	return list
}

func populateDrilldownOptions(fc *applicationglobals.FpContext, drilldownOptions *search.DrilldownOptions) {

	// TODO: Is this a LARS bug? The examples don't show a call to ParseForm being required to get query parameters
	// Even with this, the query param example isn't working for me
	err := fc.Ctx.ParseForm()
	if err != nil {
		panic(&InvalidRequest{message: "parseFormError", err: err})
	}

	// drilldown=dateYear~dateMonth:2003~May_dateYear~dateMonth:2004~June

	// drilldown=dateYear:2016+dateMonth:December_dateYear:2016+dateMonth:November_cityname:Seattle,Berlin
	// Drilldown is provided as 'field1:val1-1,val1-2_field2:val2-1' - each field/value set is seperated by '_',
	// the field & values are separated by ':' and the values are separated by ','
	// Example: "countryName:Canada_stateName:Washington,Ile-de-France_keywords:trip,flower"
	drilldown := fc.Ctx.Request().Form.Get("drilldown") // Grumble grumble - should be 'query := c.Param("drilldown")'
	if len(drilldown) > 0 {
		for _, c := range strings.Split(drilldown, "_") {
			fieldAndValues := strings.SplitN(c, ":", 2)
			if len(fieldAndValues) != 2 {
				panic(&InvalidRequest{message: fmt.Sprintf("Poorly formed drilldown (missing ':'): '%s'", c)})
			}

			field := fieldAndValues[0]
			values := strings.Split(fieldAndValues[1], ",")
			drilldownOptions.Drilldown[field] = append(drilldownOptions.Drilldown[field], values...)
		}
	}
}
