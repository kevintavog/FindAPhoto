package search

import (
	"reflect"

	"gopkg.in/olivere/elastic.v3"

	"github.com/kevintavog/findaphoto/common"
)

type NearbyOptions struct {
	Latitude, Longitude float64
	Distance            string
	MaxCount            int
	Index               int
	Count               int
}

//-------------------------------------------------------------------------------------------------
func NewNearbyOptions(lat, lon float64, distance string) *NearbyOptions {
	return &NearbyOptions{
		Latitude:  lat,
		Longitude: lon,
		Distance:  distance,
		Index:     0,
		Count:     20,
	}
}

func (no *NearbyOptions) Search() (*SearchResult, error) {
	client := common.CreateClient() // consider using elastic.NewSimpleClient
	search := client.Search().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Pretty(true)

	search.Query(elastic.NewGeoDistanceQuery("location").Lat(no.Latitude).Lon(no.Longitude).Distance(no.Distance))
	search.SortBy(elastic.NewGeoDistanceSort("location").Point(no.Latitude, no.Longitude).Order(true).Unit("km"))
	search.From(no.Index).Size(no.Count)

	return invokeSearch(search, GroupByAll, func(searchHit *elastic.SearchHit, mediaHit *MediaHit) {

		// For the geo sort, the returned sort value is the distance from the given point, in kilometers
		if len(searchHit.Sort) > 0 {
			first := searchHit.Sort[0]
			if reflect.TypeOf(first).Name() == "float64" {
				v := first.(float64)
				mediaHit.DistanceKm = &v
			}
		}
	})
}
