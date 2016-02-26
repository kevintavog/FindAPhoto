package api

import (
	"fmt"
	"net/http"
	//	"runtime"
	"strconv"
	"strings"

	"github.com/go-playground/lars"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
	"github.com/kevintavog/findaphoto/findaphotoserver/search"
)

func Search(c *lars.Context) {

	searchOptions := populateSearchOptions(c)
	propertiesFilter := getPropertiesFilter(c.Request.Form.Get("properties"))

	app := c.AppContext.(*applicationglobals.ApplicationGlobals)

	var searchResult *search.SearchResult
	var err error
	app.FieldLogger.Time("search", func() { searchResult, err = searchOptions.Search() })

	if err != nil {
		app.Error(http.StatusInternalServerError, "SearchFailed", "", err)
		return
	}

	app.FieldLogger.Add("totalMatches", fmt.Sprintf("%d", searchResult.TotalMatches))
	app.FieldLogger.Add("itemCount", strconv.Itoa(len(searchResult.Items)))
	// q=<query>; count=<per page>; first=<index of first, 1-based>; sort=<ReverseDate is all that's used currently>
	// group=<yes or no - group by folder>; properties=<id,formattedCreatedDate,keywords,city,thumbUrl,imageName,mediaType>
	// categories=<keywords,placename,date>; max=<max category count>
	// drilldown=

	// response:
	// int totalMatches
	// string oldestDateOnPage
	// string newestDateOnPage
	// GroupResult groups[]
	// Category categories[]

	// GroupResult:
	// string name
	// Image images[]

	// Category:
	// string field
	// bool isHierarchical
	// CategoryDetail details[]

	// CategoryDetail:
	// CategoryDetail children[]
	// string value
	// int count

	response := filterResults(searchResult, propertiesFilter)

	app.WriteResponse(response)
}

func filterResults(searchResult *search.SearchResult, propertiesFilter []string) map[string]interface{} {

	filtered := make(map[string]interface{})
	filtered["totalMatches"] = searchResult.TotalMatches
	filtered["items"] = filteredItems(searchResult.Items, propertiesFilter)

	return filtered
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
	case "bad":
		panic("bad")
	}

	panic(&InvalidRequest{message: fmt.Sprintf("Unknown property: '%s'", name)})
}

func populateSearchOptions(c *lars.Context) *search.SearchOptions {
	app := c.AppContext.(*applicationglobals.ApplicationGlobals)

	// TODO: Is this a LARS bug? The examples don't show a call to ParseForm being required to get query parameters
	// Even with this, the query param example isn't working for me
	err := c.ParseForm()
	if err != nil {
		app.FieldLogger.Add("parseFormError", err.Error())
	}

	// defaults: query all, return 20 results, sort by reverse date, return image id's only
	q := c.Request.Form.Get("q") // Grumble grumble - should be 'query := c.Param("q")'
	searchOptions := search.New(q)

	count := c.Request.Form.Get("count")
	if count != "" {
		searchOptions.Count, err = strconv.Atoi(count)
		if err != nil {
			app.Error(http.StatusBadRequest, "InvalidRequest", "count is not an int", err)
			return nil
		}
	}

	// TODO: Enforced here or in search? Likely better in search...
	if searchOptions.Count < 1 || searchOptions.Count > 100 {
		app.Error(http.StatusBadRequest, "InvalidRequest", "count must be between 1 and 100, inclusive", nil)
		return nil
	}

	index := c.Request.Form.Get("first")
	if index != "" {
		v, err := strconv.Atoi(index)
		if err != nil {
			app.Error(http.StatusBadRequest, "InvalidRequest", "first is not an int", err)
			return nil
		}
		if v < 1 {
			app.Error(http.StatusBadRequest, "InvalidRequest", "first must be 1 or greater", nil)
		}
		searchOptions.Index = v - 1
	}

	return searchOptions
}

func getPropertiesFilter(propertiesFilter string) []string {
	if propertiesFilter == "" {
		return []string{"id"}
	}
	return strings.Split(propertiesFilter, ",")
}
