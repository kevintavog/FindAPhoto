package api

import (
	"net/http"

	"github.com/kevintavog/findaphoto/findaphotoserver/search"
	"github.com/kevintavog/findaphoto/findaphotoserver/util"
	"github.com/labstack/echo"
)

func byDayAPI(c echo.Context) error {
	fc := c.(*util.FpContext)
	bydayOptions := populateByDayOptions(fc)
	propertiesFilter := getPropertiesFilter(c.QueryParam("properties"))

	return fc.Time("byday", func() error {
		searchResult, err := bydayOptions.Search()
		if err != nil {
			panic(&util.InvalidRequest{Message: "SearchFailed", Err: err})
		}

		fc.LogInt("itemCount", searchResult.ResultCount)
		return c.JSON(http.StatusOK, filterResults(searchResult, propertiesFilter))
	})
}

func populateByDayOptions(fc *util.FpContext) *search.ByDayOptions {

	month := fc.IntFromQuery("month", -1)
	if month < 1 || month > 12 {
		panic(&util.InvalidRequest{Message: "'month' must be between 1 and 12, inclusive"})
	}
	// If the dayOfMonth is invalid for a month (31 for February), we continue on
	dayOfMonth := fc.IntFromQuery("day", -1)
	if dayOfMonth < 1 || dayOfMonth > 31 {
		panic(&util.InvalidRequest{Message: "'day' must be between 1 and 31, inclusive"})
	}

	byDayOptions := search.NewByDayOptions(month, dayOfMonth)
	byDayOptions.Count = fc.IntFromQuery("count", byDayOptions.Count)
	if byDayOptions.Count < 1 || byDayOptions.Count > 100 {
		panic(&util.InvalidRequest{Message: "'count' must be between 1 and 100, inclusive"})
	}

	byDayOptions.Random = fc.BoolFromQuery("random", byDayOptions.Random)
	byDayOptions.Index = fc.IntFromQuery("first", 1) - 1

	populateCategoryOptions(fc, byDayOptions.CategoryOptions)
	populateDrilldownOptions(fc, byDayOptions.DrilldownOptions)

	return byDayOptions
}
