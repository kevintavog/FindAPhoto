package search

import (
	"fmt"
	"strings"
	"time"

	"encoding/json"

	"gopkg.in/olivere/elastic.v3"

	"github.com/kevintavog/findaphoto/common"
)

type CategoryOptions struct {
	PlacenameCount int
	KeywordCount   int
	DateCount      int
	YearCount      int
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
	Name   string
	Values []*CategoryValue
}

type CategoryValue struct {
	Value      string
	Count      int
	Categories []*CategoryResult
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

// Each search may return specific fields
type mappingAction func(searchHit *elastic.SearchHit, mediaHit *MediaHit)

//-------------------------------------------------------------------------------------------------
func invokeSearch(search *elastic.SearchService, groupBy int, categoryOptions *CategoryOptions, extraMapping mappingAction) (*SearchResult, error) {

	addAggregations(search, categoryOptions)

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

	sr.Categories = processAggregations(&result.Aggregations)
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
	result, err := search.Do()
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
			category := &CategoryResult{}
			values := make([]*CategoryValue, 0)

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

				categoryValue := &CategoryValue{}
				values = append(values, categoryValue)
				categoryValue.Value = v
				categoryValue.Count = int(bucket.DocCount)

				// ElasticSearch doesn't return aggregations in buckets EXCEPT for sub-aggregates. But the conversion to a Go
				// response always includes aggregates. The comparison of length is to filter out the expected 'key' & 'doc_count'
				// fields, allowing the code to get the actual values of interest.
				if bucket.Aggregations != nil && len(bucket.Aggregations) > 2 {
					categoryValue.Categories = processAggregations(&bucket.Aggregations)
				}
			}

			if len(values) > 0 {
				category.Name = key
				category.Values = values
				result = append(result, category)
			}
		}
	}

	return result
}
