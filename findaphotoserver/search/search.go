package search

import (
	"fmt"
	"reflect"
	"strings"

	"encoding/json"

	"gopkg.in/olivere/elastic.v3"

	"github.com/kevintavog/findaphoto/common"
)

type SearchOptions struct {
	Query string
	Index int
	Count int
}

type SearchResult struct {
	TotalMatches int64
	Groups       []*SearchGroup
	ResultCount  int
}

type SearchGroup struct {
	Name  string
	Items []*MediaHit
}

type MediaHit struct {
	Media      *common.Media
	DistanceKm *float64
}

type NearbyOptions struct {
	Latitude, Longitude float64
	Distance            string
	MaxCount            int
	Index               int
	Count               int
}

type ByDayOptions struct {
	Month      int
	DayOfMonth int
	Index      int
	Count      int
}

const (
	GroupByAll = iota
	GroupByPath
	GroupByDate
)

//-------------------------------------------------------------------------------------------------

// Each search may return specific fields
type mappingAction func(searchHit *elastic.SearchHit, mediaHit *MediaHit)

//-------------------------------------------------------------------------------------------------
func NewSearchOptions(query string) *SearchOptions {
	return &SearchOptions{
		Query: query,
		Index: 0,
		Count: 20,
	}
}

func (so *SearchOptions) Search() (*SearchResult, error) {
	client := common.CreateClient()
	search := client.Search().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Pretty(true)

	if so.Query == "" {
		search.Query(elastic.NewMatchAllQuery())
	} else {
		search.Query(elastic.NewQueryStringQuery(so.Query).
			Field("path"). // Folder name
			Field("monthname").
			Field("dayname").
			Field("keywords").
			Field("placename")) // Full reverse location lookup
	}

	search.From(so.Index).Size(so.Count).Sort("datetime", false)
	return invokeSearch(search, GroupByPath, nil)
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
		Pretty(true).
		Size(no.MaxCount)

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

//-------------------------------------------------------------------------------------------------
func NewByDayOptions(month, dayOfMonth int) *ByDayOptions {
	return &ByDayOptions{
		Month:      month,
		DayOfMonth: dayOfMonth,
		Index:      0,
		Count:      20,
	}
}

func (bdo *ByDayOptions) Search() (*SearchResult, error) {
	client := common.CreateClient() // consider using elastic.NewSimpleClient
	search := client.Search().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Pretty(true).
		Size(bdo.Count)

	search.Query(elastic.NewWildcardQuery("date", fmt.Sprintf("*%02d%02d", bdo.Month, bdo.DayOfMonth)))
	search.From(bdo.Index).Size(bdo.Count).Sort("datetime", false)

	return invokeSearch(search, GroupByDate, nil)
}

//-------------------------------------------------------------------------------------------------
func invokeSearch(search *elastic.SearchService, groupBy int, extraMapping mappingAction) (*SearchResult, error) {
	result, err := search.Do()
	if err != nil {
		return nil, err
	}

	sr := &SearchResult{TotalMatches: result.TotalHits()}
	if sr.TotalMatches > 0 {
		var lastGroup = ""
		sr.Groups = []*SearchGroup{}
		var group *SearchGroup

		for _, hit := range result.Hits.Hits {
			mh := &MediaHit{Media: &common.Media{}}
			err := json.Unmarshal(*hit.Source, mh.Media)
			if err != nil {
				return nil, err
			}

			if extraMapping != nil {
				extraMapping(hit, mh)
			}

			groupName := groupName(mh.Media, groupBy)

			if lastGroup == "" || lastGroup != groupName {
				group = &SearchGroup{Name: groupName, Items: []*MediaHit{}}
				lastGroup = groupName
				sr.Groups = append(sr.Groups, group)
			}

			group.Items = append(group.Items, mh)
			sr.ResultCount += 1
		}
	}

	return sr, nil
}

func groupName(media *common.Media, groupBy int) string {
	switch groupBy {
	case GroupByAll:
		return "all"

	case GroupByPath:
		// Skip to first '\', take up to last '\'
		components := strings.Split(media.Path, "\\")

		if len(components) > 2 {
			return strings.Join(components[1:len(components)-1], "\\")
		}

	case GroupByDate:
		return media.DateTime.Format("2006-01-02")
	}

	return ""
}
