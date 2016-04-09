package search

import (
	"fmt"
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
	Items []*common.Media
}

type NearbyOptions struct {
	Latitude, Longitude float64
	Distance            string
}

type ByDayOptions struct {
	Month      int
	DayOfMonth int
	Count      int
}

const (
	GroupByAll = iota
	GroupByPath
	GroupByDate
)

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
	return invokeSearch(search, GroupByPath)
}

//-------------------------------------------------------------------------------------------------
func NewNearbyOptions(lat, lon float64, distance string) *NearbyOptions {
	return &NearbyOptions{
		Latitude:  lat,
		Longitude: lon,
		Distance:  distance,
	}
}

func (no *NearbyOptions) Search() (*SearchResult, error) {
	client := common.CreateClient() // consider using elastic.NewSimpleClient
	search := client.Search().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Pretty(true).
		Size(10)

	search.Query(elastic.NewGeoDistanceQuery("location").Lat(no.Latitude).Lon(no.Longitude).Distance(no.Distance))
	search.SortBy(elastic.NewGeoDistanceSort("location").Point(no.Latitude, no.Longitude).Order(true).Unit("km"))

	return invokeSearch(search, GroupByAll)
}

//-------------------------------------------------------------------------------------------------
func NewByDayOptions(month, dayOfMonth int) *ByDayOptions {
	return &ByDayOptions{
		Month:      month,
		DayOfMonth: dayOfMonth,
		Count:      25,
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
	search.Sort("datetime", false)

	return invokeSearch(search, GroupByDate)
}

//-------------------------------------------------------------------------------------------------
func invokeSearch(search *elastic.SearchService, groupBy int) (*SearchResult, error) {
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
			media := &common.Media{}
			err := json.Unmarshal(*hit.Source, media)
			if err != nil {
				return nil, err
			}

			groupName := groupName(media, groupBy)

			if lastGroup == "" || lastGroup != groupName {
				group = &SearchGroup{Name: groupName, Items: []*common.Media{}}
				lastGroup = groupName
				sr.Groups = append(sr.Groups, group)
			}

			group.Items = append(group.Items, media)
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
