package api

import (
	"fmt"
	"net/http"

	"github.com/kevintavog/findaphoto/findaphotoserver/search"
	"github.com/kevintavog/findaphoto/findaphotoserver/util"
	"github.com/labstack/echo"
)

func nearbyAPI(c echo.Context) error {
	fc := c.(*util.FpContext)
	nearbyOptions := populateNearbyOptions(fc)
	propertiesFilter := getPropertiesFilter(c.QueryParam("properties"))

	return fc.Time("nearby", func() error {
		searchResult, err := nearbyOptions.Search()
		if err != nil {
			panic(&util.InvalidRequest{Message: "SearchFailed", Err: err})
		}

		fc.LogInt("itemCount", searchResult.ResultCount)
		return c.JSON(http.StatusOK, filterResults(searchResult, propertiesFilter))
	})
}

func populateNearbyOptions(fc *util.FpContext) *search.NearbyOptions {

	lat := fc.Float64FromQuery("lat")
	lon := fc.Float64FromQuery("lon")
	// The intent of this api is to return the top few closest items - even if they're on the other side of the world
	nearbyOptions := search.NewNearbyOptions(lat, lon, "13000km")
	nearbyOptions.MaxCount = fc.IntFromQuery("count", 5)

	maxKilometers := fc.OptionalFloat64FromQuery("maxKilometers", 13000)
	if maxKilometers < 1 || maxKilometers > 20000 {
		panic(&util.InvalidRequest{Message: "maxKilometers must be between 1 and 20,000, inclusive"})
	}
	nearbyOptions.Distance = fmt.Sprintf("%fkm", maxKilometers)

	nearbyOptions.Count = fc.IntFromQuery("count", nearbyOptions.Count)
	if nearbyOptions.Count < 1 || nearbyOptions.Count > 100 {
		panic(&util.InvalidRequest{Message: "count must be between 1 and 100, inclusive"})
	}

	nearbyOptions.Index = fc.IntFromQuery("first", 1) - 1
	populateCategoryOptions(fc, nearbyOptions.CategoryOptions)
	populateDrilldownOptions(fc, nearbyOptions.DrilldownOptions)

	return nearbyOptions
}
