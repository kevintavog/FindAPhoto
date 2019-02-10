package api

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/lars"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
	"github.com/kevintavog/findaphoto/findaphotoserver/controllers/files"
	"github.com/kevintavog/findaphoto/findaphotoserver/search"
)

func ConfigureRouting(l *lars.LARS) {
	l.Use(handleErrors)
	l.Use(Cors)

	api := l.Group("/api")
	api.Get("/search", Search)
	api.Get("/nearby", Nearby)
	api.Get("/by-day", ByDay)
	api.Get("/media/:id", MediaById)

	index := api.Group("/index")
	index.Get("/fieldvalues", IndexFieldValues)
	index.Get("/duplicates", DuplicateMedia)
	index.Get("/info", Index)
	index.Post("/reindex", Reindex)
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
			panic(&InvalidRequest{message: fmt.Sprintf("'%s' is not a bool: %s", name, s)})
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
		// listItem["locations"] = aggregateLocations(group.Items)
	}
	return list
}

func aggregateLocations(items []*search.MediaHit) interface{} {
	list := make([]map[string]interface{}, 0)
	for _, mh := range items {
		if mh.Media.Location != nil {
			fmt.Printf("site: %s, city: %s, state: %s, country: %s\n",
				mh.Media.LocationSiteName, mh.Media.LocationCityName,
				mh.Media.LocationStateName, mh.Media.LocationCountryName)
		}
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
	case "country":
		return common.ConvertToCountryName(mh.Media.LocationCountryCode, mh.Media.LocationCountryName)
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
	case "exposuretimestring":
		return mh.Media.ExposureTimeString
	case "flash":
		return mh.Media.Flash
	case "fnumber":
		return mh.Media.FNumber
	case "focallength":
		return mh.Media.FocalLengthMm
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
	case "locationplacename":
		if mh.Media.Location == nil {
			return nil
		}
		return mh.Media.LocationPlaceName
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
	case "tags":
		return mh.Media.Tags
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
			case "tags":
				categoryOptions.TagCount = 10
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
