package api

import (
	"net/http"
	"strings"

	"github.com/kevintavog/findaphoto/findaphotoserver/search"
	"github.com/kevintavog/findaphoto/findaphotoserver/util"
	"github.com/labstack/echo"
)

func searchAPI(c echo.Context) error {
	fc := c.(*util.FpContext)
	searchOptions := populateSearchOptions(fc)
	propertiesFilter := getPropertiesFilter(c.QueryParam("properties"))

	return fc.Time("search", func() error {
		searchResult, err := searchOptions.Search()
		util.PropogateError(err, "SearchFailed")

		fc.LogInt64("totalMatches", searchResult.TotalMatches)
		fc.LogInt("itemCount", searchResult.ResultCount)

		return c.JSON(http.StatusOK, filterResults(searchResult, propertiesFilter))
	})
}

func populateSearchOptions(fc *util.FpContext) *search.SearchOptions {

	// defaults: query all, return 20 results, sort by reverse date, return image id's only
	q := fc.Context.QueryParam("q")
	searchOptions := search.NewSearchOptions(q)

	searchOptions.Count = fc.IntFromQuery("count", searchOptions.Count)
	if searchOptions.Count < 1 || searchOptions.Count > 100 {
		panic(&util.InvalidRequest{Message: "count must be between 1 and 100, inclusive"})
	}

	searchOptions.Index = fc.IntFromQuery("first", 1) - 1
	populateCategoryOptions(fc, searchOptions.CategoryOptions)
	populateDrilldownOptions(fc, searchOptions.DrilldownOptions)
	return searchOptions
}

func getPropertiesFilter(propertiesFilter string) []string {
	if propertiesFilter == "" {
		return []string{"id"}
	}
	return strings.Split(propertiesFilter, ",")
}
