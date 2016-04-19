package api

import (
	"strconv"

	"github.com/go-playground/lars"

	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
	"github.com/kevintavog/findaphoto/findaphotoserver/search"
)

func ByDay(c lars.Context) {
	fc := c.(*applicationglobals.FpContext)
	nearbyOptions := populateByDayOptions(fc)
	propertiesFilter := getPropertiesFilter(fc.Ctx.Request().Form.Get("properties"))

	fc.AppContext.FieldLogger.Time("byday", func() {
		searchResult, err := nearbyOptions.Search()
		if err != nil {
			panic(&InternalError{message: "SearchFailed", err: err})
		}

		fc.AppContext.FieldLogger.Add("itemCount", strconv.Itoa(searchResult.ResultCount))
		fc.WriteResponse(filterResults(searchResult, propertiesFilter))
	})
}

func populateByDayOptions(fc *applicationglobals.FpContext) *search.ByDayOptions {

	// TODO: Is this a LARS bug? The examples don't show a call to ParseForm being required to get query parameters
	// Even with this, the query param example isn't working for me
	err := fc.Ctx.ParseForm()
	if err != nil {
		panic(&InvalidRequest{message: "parseFormError", err: err})
	}

	month := intFromQuery(fc.Ctx, "month", -1)
	if month < 1 || month > 12 {
		panic(&InvalidRequest{message: "month must be between 1 and 12, inclusive"})
	}
	// If the dayOfMonth is invalid (31 for February), we continue on
	dayOfMonth := intFromQuery(fc.Ctx, "day", -1)
	if dayOfMonth < 1 || month > 31 {
		panic(&InvalidRequest{message: "day must be between 1 and 31, inclusive"})
	}

	byDayOptions := search.NewByDayOptions(month, dayOfMonth)
	byDayOptions.Count = intFromQuery(fc.Ctx, "count", byDayOptions.Count)
	if byDayOptions.Count < 1 || byDayOptions.Count > 100 {
		panic(&InvalidRequest{message: "count must be between 1 and 100, inclusive"})
	}

	byDayOptions.Random = boolFromQuery(fc.Ctx, "random", byDayOptions.Random)
	byDayOptions.Index = intFromQuery(fc.Ctx, "first", 1) - 1

	return byDayOptions
}
