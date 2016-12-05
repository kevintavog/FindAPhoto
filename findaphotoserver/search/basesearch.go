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

	addAggregations(search, categoryOptions)
	AddDrilldown(search, query, drilldownOptions)

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
		search.Aggregation("keywords", elastic.NewTermsAggregation().Field("keywords.value").Size(categoryOptions.KeywordCount))
	}

	// Dates are in a Year, Month, Day hierarchy - years & days are limited by the requested limit, while all 12 months are returned (if they exist)
	if categoryOptions.DateCount > 0 {
		scriptYear := elastic.NewScriptInline("doc['datetime'].date.toString('YYYY')").Lang("painless")
		scriptMonth := elastic.NewScriptInline("doc['datetime'].date.toString('MMMM')").Lang("painless")
		scriptDay := elastic.NewScriptInline("doc['datetime'].date.toString('dd')").Lang("painless")
		search.Aggregation("dateYear", elastic.NewTermsAggregation().Script(scriptYear).Size(categoryOptions.DateCount).
			SubAggregation("dateMonth", elastic.NewTermsAggregation().Script(scriptMonth).Size(12).
				SubAggregation("dateDay", elastic.NewTermsAggregation().Script(scriptDay).Size(categoryOptions.DateCount))))
	}

	// Location name is returned as a Country, State, City, Site hierarchy
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

func AddDrilldown(search *elastic.SearchService, searchQuery *elastic.Query, drilldownOptions *DrilldownOptions) {
	if searchQuery == nil {
		return
	}

	if len(drilldownOptions.Drilldown) < 1 {
		return
	}

	// locations (site, city, state, country) are OR - also, OR between each type/level
	// dates (year, month, day) are OR - also OR between each value
	// keywords are OR
	// ANDed together ((locations) AND (dates) AND (keywords))

	// countryName:Canada;stateName:Washington,Ile-de-France;keywords:trip,flower
	// (countryName=Canada OR stateName=Washington OR stateName=Ile-de-France) AND (keywords=trip OR keywords=flower)

	drilldownQuery := elastic.NewBoolQuery()
	drilldownQuery.Must(*searchQuery)

	locationQueryList := make([]*elastic.TermQuery, 0)
	dateQueryList := make([]*elastic.ScriptQuery, 0)

	for key, value := range drilldownOptions.Drilldown {
		if isLocationField(key) {
			for _, v := range value {
				locationQueryList = append(locationQueryList, elastic.NewTermQuery(getLocationFieldName(key), fmt.Sprintf("%s", v)))
			}

		} else if isDateField(key) {
			for _, v := range value {
				script := elastic.NewScriptInline(getDateFieldQuery(key)).Lang("painless").Param("dateValue", v)
				dateQueryList = append(dateQueryList, elastic.NewScriptQuery(script))
			}

		} else {
			fieldQuery := elastic.NewBoolQuery()
			for _, v := range value {
				fieldName, overridden := common.GetIndexFieldName(key)
				fieldValue := v
				if !overridden {
					fieldValue = strings.ToLower(v)
				}
				fieldQuery.Should(elastic.NewTermQuery(fieldName, fieldValue))
			}
			drilldownQuery.Must(fieldQuery)
		}
	}

	if len(locationQueryList) > 0 {
		locationQuery := elastic.NewBoolQuery()
		for _, q := range locationQueryList {
			locationQuery.Should(q)
		}
		drilldownQuery.Must(locationQuery)
	}

	if len(dateQueryList) > 0 {
		dateQuery := elastic.NewBoolQuery()
		for _, q := range dateQueryList {
			dateQuery.Should(q)
		}
		drilldownQuery.Must(dateQuery)
	}

	//	src, _ := drilldownQuery.Source()
	//	dataMap, _ := json.Marshal(src)
	//	jsonString := string(dataMap)
	//	fmt.Printf("drilldown: '%s'\n", jsonString)

	search.Query(drilldownQuery)
}

func isLocationField(name string) bool {
	switch strings.ToLower(name) {
	case "countryname", "statename", "cityname", "sitename":
		return true
	default:
		return false

	}
}

func isDateField(name string) bool {
	switch strings.ToLower(name) {
	case "dateyear", "datemonth", "dateday":
		return true
	default:
		return false

	}
}

func getLocationFieldName(name string) string {
	return strings.ToLower(name + ".value")
}

func getDateFieldQuery(name string) string {
	switch strings.ToLower(name) {
	case "dateyear":
		return "doc['datetime'].date.toString('YYYY') == params.dateValue"
	case "datemonth":
		return "doc['datetime'].date.toString('MMMM') == params.dateValue"
	case "dateday":
		return "doc['datetime'].date.toString('dd') == params.dateValue"
	default:
		fmt.Errorf("Unhandled date field name: '%s'\n", name)
		return ""
	}
}
