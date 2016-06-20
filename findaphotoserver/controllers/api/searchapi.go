package api

import (
	"strconv"
	"strings"

	"github.com/go-playground/lars"

	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
	"github.com/kevintavog/findaphoto/findaphotoserver/search"
)

func Search(c lars.Context) {
	fc := c.(*applicationglobals.FpContext)
	searchOptions := populateSearchOptions(fc)
	propertiesFilter := getPropertiesFilter(fc.Ctx.Request().Form.Get("properties"))

	fc.AppContext.FieldLogger.Time("search", func() {
		searchResult, err := searchOptions.Search()
		if err != nil {
			panic(&InternalError{message: "SearchFailed", err: err})
		}

		fc.AppContext.FieldLogger.Add("totalMatches", strconv.FormatInt(searchResult.TotalMatches, 10))
		fc.AppContext.FieldLogger.Add("itemCount", strconv.Itoa(searchResult.ResultCount))

		fc.WriteResponse(filterResults(searchResult, propertiesFilter))
	})
}

func populateSearchOptions(fc *applicationglobals.FpContext) *search.SearchOptions {

	// TODO: Is this a LARS bug? The examples don't show a call to ParseForm being required to get query parameters
	// Even with this, the query param example isn't working for me
	err := fc.Ctx.ParseForm()
	if err != nil {
		panic(&InvalidRequest{message: "parseFormError", err: err})
	}

	// defaults: query all, return 20 results, sort by reverse date, return image id's only
	q := fc.Ctx.Request().Form.Get("q") // Grumble grumble - should be 'query := c.Param("q")'
	searchOptions := search.NewSearchOptions(q)

	searchOptions.Count = intFromQuery(fc.Ctx, "count", searchOptions.Count)
	if searchOptions.Count < 1 || searchOptions.Count > 100 {
		panic(&InvalidRequest{message: "count must be between 1 and 100, inclusive"})
	}

	searchOptions.Index = intFromQuery(fc.Ctx, "first", 1) - 1
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
