package search

import (
	"fmt"
	"strings"
	"time"

	"encoding/json"

	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
)

type CategoryOptions struct {
	PlacenameCount int
	KeywordCount   int
	DateCount      int
	YearCount      int
}

type DrilldownOptions struct {
	Drilldown map[string][]string
}

type SearchResult struct {
	TotalMatches int64
	Groups       []*SearchGroup
	ResultCount  int

	// Returned only for byday searches
	NextAvailableByDay     *ByDayResult
	PreviousAvailableByDay *ByDayResult

	Categories []*CategoryResult
}

type ByDayResult struct {
	Month int
	Day   int
}

type SearchGroup struct {
	Name  string
	Items []*MediaHit
}

type MediaHit struct {
	Media      *common.Media
	DistanceKm *float64
}

type CategoryResult struct {
	Field   string
	Details []*CategoryDetailResult
}

type CategoryDetailResult struct {
	Field    *string
	Value    string
	Count    int
	Children []*CategoryDetailResult
}

const (
	GroupByAll = iota
	GroupByPath
	GroupByDate
)

//-------------------------------------------------------------------------------------------------
func NewCategoryOptions() *CategoryOptions {
	return &CategoryOptions{
		PlacenameCount: 0,
		KeywordCount:   0,
		DateCount:      0,
		YearCount:      0,
	}
}

//-------------------------------------------------------------------------------------------------
func NewDrilldownOptions() *DrilldownOptions {
	return &DrilldownOptions{
		Drilldown: make(map[string][]string),
	}
}

//-------------------------------------------------------------------------------------------------

// Each search may return specific fields
type mappingAction func(searchHit *elastic.SearchHit, mediaHit *MediaHit)

//-------------------------------------------------------------------------------------------------
func invokeSearch(search *elastic.SearchService, query *elastic.Query, groupBy int, categoryOptions *CategoryOptions, drilldownOptions *DrilldownOptions, extraMapping mappingAction) (*SearchResult, error) {

	//	addAggregations(search, categoryOptions)
	addDrilldown(search, query, drilldownOptions)

	result, err := search.Do(context.TODO())
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

	sr.Categories = processAggregations(&result.Aggregations)
	return sr, nil
}

//-------------------------------------------------------------------------------------------------
func clearScroll(scrollId string) {
	_, err := common.CreateClient().ClearScroll().ScrollId(scrollId).Do(context.TODO())
	if err != nil {
		fmt.Printf("Clearing a scroll id failed: ")
		fmt.Print(err)
		fmt.Println()
	}
}

//-------------------------------------------------------------------------------------------------
func invokeScrollSearch(scroll *elastic.ScrollService, query *elastic.Query, first int, count int, groupBy int, categoryOptions *CategoryOptions, drilldownOptions *DrilldownOptions, extraMapping mappingAction) (*SearchResult, error) {

	startTime := time.Now()

	last := first + count - 1

	scroll.Size(first)

	results, err := scroll.Do(context.TODO())
	if err != nil {
		fmt.Print("First scroll search failed: ")
		fmt.Println(err)
		return nil, err
	}

	d := time.Now().Sub(startTime).Seconds() * 1000
	fmt.Printf("First search took %d (%01.3f)", results.TookInMillis, d)
	fmt.Println()

	scrollId := results.ScrollId
	defer clearScroll(scrollId)

	documentsVisited := len(results.Hits.Hits)
	pagesVisited := 0

	fmt.Printf("Taking from %d to %d, of %d (%d)", first, last, results.TotalHits(), documentsVisited)
	fmt.Println()
	var sr = &SearchResult{TotalMatches: results.TotalHits()}
	if sr.TotalMatches > 0 {
		var lastGroup = ""
		var group *SearchGroup
		sr.Groups = []*SearchGroup{}

		client := common.CreateClient()
		takeScroll := client.Scroll(common.MediaIndexName).
			Type(common.MediaTypeName).
			Sort("datetime", false).
			ScrollId(scrollId).
			Size(count)

		for documentsVisited <= last {
			results, err := takeScroll.Do((context.TODO()))

			if err != nil {
				fmt.Printf("scroll test search failed at %d of %d: ", documentsVisited, last)
				fmt.Println(err)
				return nil, err
			} else {
				dur := time.Now().Sub(startTime).Seconds() * 1000
				fmt.Printf("next search took %d, returned %d hits, %01.3f msecs", results.TookInMillis, len(results.Hits.Hits), dur)
				fmt.Println()
				pagesVisited++

				numHitDocuments := len(results.Hits.Hits)
				if documentsVisited+numHitDocuments < first {
					documentsVisited += numHitDocuments
				} else {
					for _, hit := range results.Hits.Hits {

						if documentsVisited > last {
							break
						}

						if documentsVisited >= first && documentsVisited <= last {
							mh := &MediaHit{Media: &common.Media{}}
							err := json.Unmarshal(*hit.Source, mh.Media)
							if err != nil {
								return nil, err
							}

							if extraMapping != nil {
								extraMapping(hit, mh)
							}

							groupName := groupName(mh.Media, GroupByPath)

							if lastGroup == "" || lastGroup != groupName {
								group = &SearchGroup{Name: groupName, Items: []*MediaHit{}}
								lastGroup = groupName
								sr.Groups = append(sr.Groups, group)
							}

							group.Items = append(group.Items, mh)
							sr.ResultCount += 1
						}

						documentsVisited++
					}
				}
			}
		}

		durationMsecs := time.Now().Sub(startTime).Seconds() * 1000
		fmt.Printf("scroll test search was OK, visited %d pages and %d documents in %01.3f msecs", pagesVisited, documentsVisited, durationMsecs)
		fmt.Println()
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

func returnFirstMatch(search *elastic.SearchService) (*MediaHit, error) {
	result, err := search.Do(context.TODO())
	if err != nil {
		return nil, err
	}

	if result.TotalHits() > 0 {
		hit := result.Hits.Hits[0]
		mh := &MediaHit{Media: &common.Media{}}
		err := json.Unmarshal(*hit.Source, mh.Media)
		if err != nil {
			return nil, err
		}
		return mh, nil
	}

	return nil, nil
}

func addAggregations(search *elastic.SearchService, categoryOptions *CategoryOptions) {
	if categoryOptions.KeywordCount > 0 {
		search.Aggregation("keywords", elastic.NewTermsAggregation().Field("keywords").Size(categoryOptions.KeywordCount))
	}

	if categoryOptions.DateCount > 0 {
		search.Aggregation("date", elastic.NewTermsAggregation().Field("date").Size(categoryOptions.DateCount))
	}

	if categoryOptions.PlacenameCount > 0 {
		search.Aggregation("countryName", elastic.NewTermsAggregation().Field("countryname.value").Size(categoryOptions.PlacenameCount).
			SubAggregation("stateName", elastic.NewTermsAggregation().Field("statename.value").Size(categoryOptions.PlacenameCount).
				SubAggregation("cityName", elastic.NewTermsAggregation().Field("cityname.value").Size(categoryOptions.PlacenameCount).
					SubAggregation("siteName", elastic.NewTermsAggregation().Field("sitename.value").Size(categoryOptions.PlacenameCount)))))
	}
}

func processAggregations(aggregations *elastic.Aggregations) []*CategoryResult {
	if aggregations == nil || len(*aggregations) < 1 {
		return nil
	}

	result := make([]*CategoryResult, 0)

	for key, _ := range *aggregations {
		terms, ok := aggregations.Terms(key)
		if ok {
			topCategory := &CategoryResult{}
			detailsList := make([]*CategoryDetailResult, 0)

			for _, bucket := range terms.Buckets {
				if bucket.DocCount == 0 {
					continue
				}

				v := ""
				switch bucket.Key.(type) {
				case string:
					v = bucket.Key.(string)
				case float64:
					fmt.Printf("Handling a float64 conversion for aggregation\n")
					// Assume it's a time, specifically, milliseconds since the epoch
					msec := int64(bucket.Key.(float64))
					v = fmt.Sprintf("%s", time.Unix(msec/1000, 0))
				}

				detail := &CategoryDetailResult{}
				detail.Value = v
				detail.Count = int(bucket.DocCount)
				detailsList = append(detailsList, detail)

				// ElasticSearch doesn't return aggregations in buckets EXCEPT for sub-aggregates. But the conversion to a Go
				// response always includes aggregates. The comparison of length is to filter out the expected 'key' & 'doc_count'
				// fields, allowing the code to get the actual values of interest.
				if bucket.Aggregations != nil && len(bucket.Aggregations) > 2 {
					detail.Field, detail.Children = processDetailAggregations(&bucket.Aggregations)
				}
			}

			topCategory.Field = key
			if len(detailsList) > 0 {
				topCategory.Details = detailsList
				result = append(result, topCategory)
			}
		}
	}

	return result
}

func processDetailAggregations(aggregations *elastic.Aggregations) (*string, []*CategoryDetailResult) {
	if aggregations == nil || len(*aggregations) < 1 {
		return nil, nil
	}

	result := make([]*CategoryDetailResult, 0)

	var key string
	for k, _ := range *aggregations {
		terms, ok := aggregations.Terms(k)
		if ok {
			key = k
			for _, bucket := range terms.Buckets {
				if bucket.DocCount == 0 {
					continue
				}

				v := ""
				switch bucket.Key.(type) {
				case string:
					v = bucket.Key.(string)
				case float64:
					fmt.Printf("Handling a float64 conversion for aggregation\n")
					// Assume it's a time, specifically, milliseconds since the epoch
					msec := int64(bucket.Key.(float64))
					v = fmt.Sprintf("%s", time.Unix(msec/1000, 0))
				default:
					v = "error"
					fmt.Printf("Unhandled type '%v'\n", bucket.Key)
				}

				detail := &CategoryDetailResult{}
				result = append(result, detail)
				detail.Value = v
				detail.Count = int(bucket.DocCount)

				// ElasticSearch doesn't return aggregations in buckets EXCEPT for sub-aggregates. But the conversion to a Go
				// response always includes aggregates. The comparison of length is to filter out the expected 'key' & 'doc_count'
				// fields, allowing the code to get the actual values of interest.
				if bucket.Aggregations != nil && len(bucket.Aggregations) > 2 {
					detail.Field, detail.Children = processDetailAggregations(&bucket.Aggregations)
				}
			}
		}
	}

	if len(result) > 0 {
		return &key, result
	}

	return nil, nil
}

func addDrilldown(search *elastic.SearchService, searchQuery *elastic.Query, drilldownOptions *DrilldownOptions) {
	if searchQuery == nil {
		return
	}

	if len(drilldownOptions.Drilldown) < 1 {
		return
	}

	// locations (site, city, state, country) are OR - also, OR between each type/level
	// dates are OR
	// keywords are OR
	// ANDed together ((locations) AND (dates) AND (keywords))

	// countryName:Canada;stateName:Washington,Ile-de-France;keywords:trip,flower
	// (countryName=Canada OR stateName=Washington OR stateName=Ile-de-France) AND (keywords=trip OR keywords=flower)

	drilldownQuery := elastic.NewBoolQuery()
	drilldownQuery.Must(*searchQuery)
	var locationQuery *elastic.BoolQuery

	for key, value := range drilldownOptions.Drilldown {
		isLocationField := false
		var fieldQuery *elastic.BoolQuery

		switch strings.ToLower(key) {
		// Re-use location query to have OR semantics for all location related fields
		case "countryname", "statename", "cityname", "sitename":
			if locationQuery == nil {
				locationQuery = elastic.NewBoolQuery()
			}
			fieldQuery = locationQuery
			isLocationField = true
		default:
			fieldQuery = elastic.NewBoolQuery()
		}

		for _, v := range value {
			fieldQuery.Should(elastic.NewTermQuery(strings.ToLower(key), strings.ToLower(v)))
		}

		if !isLocationField {
			drilldownQuery.Must(fieldQuery)
		}
	}

	if locationQuery != nil {
		drilldownQuery.Must(locationQuery)
	}

	search.Query(drilldownQuery)
}
