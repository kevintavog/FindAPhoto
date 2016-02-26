package api

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/lars"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
	"github.com/kevintavog/findaphoto/findaphotoserver/search"
)

/*
 q=<query>; count=<per page>; first=<index of first, 1-based>; sort=<ReverseDate is all that's used currently>
 group=<yes or no - group by folder>; properties=<id,formattedCreatedDate,keywords,city,thumbUrl,imageName,mediaType>
 categories=<keywords,placename,date>; max=<max category count>
 drilldown=

 response:
 int totalMatches
 string oldestDateOnPage
 string newestDateOnPage
 GroupResult groups[]
 Category categories[]

 GroupResult:
 string name
 Image images[]

 Category:
 string field
 bool isHierarchical
 CategoryDetail details[]

 CategoryDetail:
 CategoryDetail children[]
 string value
 int count

*/

func Search(c *lars.Context) {

	searchOptions := populateSearchOptions(c)
	propertiesFilter := getPropertiesFilter(c.Request.Form.Get("properties"))

	app := c.AppContext.(*applicationglobals.ApplicationGlobals)

	app.FieldLogger.Time("search", func() {
		searchResult, err := searchOptions.Search()

		if err != nil {
			panic(&InternalError{message: "SearchFailed", err: err})
		}

		app.FieldLogger.Add("totalMatches", fmt.Sprintf("%d", searchResult.TotalMatches))
		app.FieldLogger.Add("itemCount", strconv.Itoa(len(searchResult.Items)))

		app.WriteResponse(filterResults(searchResult, propertiesFilter))
	})

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
	}

	panic(&InvalidRequest{message: fmt.Sprintf("Unknown property: '%s'", name)})
}

func populateSearchOptions(c *lars.Context) *search.SearchOptions {

	// TODO: Is this a LARS bug? The examples don't show a call to ParseForm being required to get query parameters
	// Even with this, the query param example isn't working for me
	err := c.ParseForm()
	if err != nil {
		panic(&InvalidRequest{message: "parseFormError", err: err})
	}

	// defaults: query all, return 20 results, sort by reverse date, return image id's only
	q := c.Request.Form.Get("q") // Grumble grumble - should be 'query := c.Param("q")'
	searchOptions := search.New(q)

	count := c.Request.Form.Get("count")
	if count != "" {
		searchOptions.Count, err = strconv.Atoi(count)
		if err != nil {
			panic(&InvalidRequest{message: "count is not an int"})
		}
	}

	if searchOptions.Count < 1 || searchOptions.Count > 100 {
		panic(&InvalidRequest{message: "count must be between 1 and 100, inclusive"})
	}

	index := c.Request.Form.Get("first")
	if index != "" {
		v, err := strconv.Atoi(index)
		if err != nil {
			panic(&InvalidRequest{message: "first is not an int", err: err})
		}
		if v < 1 {
			panic(&InvalidRequest{message: "first must be 1 or greater"})
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
