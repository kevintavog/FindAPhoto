package api

import (
	"strconv"

	"github.com/go-playground/lars"

	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
	"github.com/kevintavog/findaphoto/findaphotoserver/search"
)

func Nearby(c lars.Context) {
	fc := c.(*applicationglobals.FpContext)
	nearbyOptions := populateNearbyOptions(fc)
	propertiesFilter := getPropertiesFilter(fc.Ctx.Request().Form.Get("properties"))

	fc.AppContext.FieldLogger.Time("nearby", func() {
		searchResult, err := nearbyOptions.Search()
		if err != nil {
			panic(&InternalError{message: "SearchFailed", err: err})
		}

		fc.AppContext.FieldLogger.Add("itemCount", strconv.Itoa(searchResult.ResultCount))
		fc.WriteResponse(filterResults(searchResult, propertiesFilter))
	})
}

func populateNearbyOptions(fc *applicationglobals.FpContext) *search.NearbyOptions {

	// TODO: Is this a LARS bug? The examples don't show a call to ParseForm being required to get query parameters
	// Even with this, the query param example isn't working for me
	err := fc.Ctx.ParseForm()
	if err != nil {
		panic(&InvalidRequest{message: "parseFormError", err: err})
	}

	lat := float64FromQuery(fc.Ctx, "lat")
	lon := float64FromQuery(fc.Ctx, "lon")
	// The intent of this api is to return the top few closest items - even if they're on the other side of the world
	nearbyOptions := search.NewNearbyOptions(lat, lon, "13000km")
	nearbyOptions.MaxCount = intFromQuery(fc.Ctx, "count", 5)

	nearbyOptions.Count = intFromQuery(fc.Ctx, "count", nearbyOptions.Count)
	if nearbyOptions.Count < 1 || nearbyOptions.Count > 100 {
		panic(&InvalidRequest{message: "count must be between 1 and 100, inclusive"})
	}

	nearbyOptions.Index = intFromQuery(fc.Ctx, "first", 1) - 1
	populateCategoryOptions(fc, nearbyOptions.CategoryOptions)

	return nearbyOptions
}
