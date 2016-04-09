package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kevintavog/findaphoto/common"

	"github.com/ian-kent/go-log/log"
	"gopkg.in/olivere/elastic.v3"
)

func addDateRangeQuery(search *elastic.SearchService) *elastic.BoolQuery {
	query := elastic.NewBoolQuery().
		Should(elastic.NewRangeQuery("datetime").Gte("03-12").Lte("03-28").Format("MM-dd").TimeZone("-07:00")) //.
	//		Should(elastic.NewRangeQuery("datetime").Gte("2013-03-12").Lte("2013-03-31").Format("yyyy-MM-dd").TimeZone("-07:00"))

	return query
}

func addQuery(search *elastic.SearchService) {
	hasSort := false
	// Good, useful queries
	//	query := elastic.NewTermQuery("_id", "1\\2001\\PDRM0300.JPG")
	//	query := elastic.NewTermQuery("path.raw", "1\\2001\\PDRM0300.JPG") // Search for a given path (can use _id instead...)
	//	query := elastic.NewGeoDistanceQuery("location").Lat(47.60863888888889).Lon(-122.43516666666666).Distance("7km")

	//	sort := elastic.NewGeoDistanceSort("location").Point(47.608638, -122.43516).Unit("km")
	sort := elastic.NewGeoDistanceSort("location").Point(55.804175, -43.584820).Unit("km")
	//	hasSort = true

	//	query := elastic.NewTermQuery("path", "2005") // 'path' is analyzed, so partial matches work
	//	query := elastic.NewTermQuery("path", "2")
	//	query := elastic.NewQueryStringQuery("path:2005 AND filename:100*") // A direct query, as typed in - good for adhoc/from or from search box in UI
	//	query := elastic.NewRangeQuery("datetime").Gte("2014/01/01").Lte("2014/12/31").Format("yyyy/MM/dd")
	//	query := elastic.NewWildcardQuery("date", "20140*")
	//	query := elastic.NewWildcardQuery("date", "*0904") // All matches on a given month/day, across all years
	//	query := elastic.NewQueryStringQuery("monthname:august") // string query is analyzed, so can find 'Sep' and 'sep'
	//	query := elastic.NewQueryStringQuery("filename:DSCN3380")
	//	query := elastic.NewQueryStringQuery("DSCN3380").Field("path").Field("monthname").Field("dayname").Field("keyword").Field("placename")
	//	query := elastic.NewQueryStringQuery("mural").Field("path").Field("monthname").Field("dayname").Field("keywords").Field("placename")
	//	query := elastic.NewQueryStringQuery("keywords:mount rainier")

	//	query := addDateRangeQuery(search)

	//	query := elastic.NewQueryStringQuery("cityname:Seattle")
	query := elastic.NewMatchAllQuery()
	//	query = nil

	//	query := elastic.NewQueryStringQuery("keywords:mount rainier")

	// Useful queries, but with problems (likely case-sensitive issues)
	//	query := elastic.NewWildcardQuery("filename", "dscn3*") // only lower case matches - as if query isn't being analyzed prior to search
	//	query := elastic.NewTermQuery("filename", "100_0304")

	// Work in progress
	//	query := elastic.NewTermQuery("filename", "DSCN3380")
	//	query := elastic.NewTermQuery("filename.raw", "DSCN*")
	//	query := elastic.NewTermQuery("path", "100_0304")

	if query != nil {
		search.Query(query).Size(10)
		if hasSort {
			search.SortBy(sort)
		} else {
			search.Sort("datetime", false)
		}
	} else {
		search.Size(0)
	}
}

func addAggregate(search *elastic.SearchService) {
	search.Aggregation("countryName", elastic.NewTermsAggregation().Field("countryname.value").Size(30).
		SubAggregation("stateName", elastic.NewTermsAggregation().Field("statename.value").Size(10).
		SubAggregation("cityName", elastic.NewTermsAggregation().Field("cityname.value").Size(10).
		SubAggregation("siteName", elastic.NewTermsAggregation().Field("sitename.value").Size(10)))))

	//	search.Aggregation("siteName", elastic.NewTermsAggregation().Field("sitename.value").Size(30))
	//	search.Aggregation("cityName", elastic.NewTermsAggregation().Field("cityname.value").Size(30))

	//	search.Aggregation("datetime", elastic.NewDateHistogramAggregation().
	//		Field("datetime").
	//		Interval("month").
	//		Format("YYYY-MM").
	//		TimeZone("UTC").
	//		Offset("-8h"))

	//	maxResults := 10
	//	field := "date"

	//	search.Aggregation("test-"+field, elastic.NewTermsAggregation().Field(field).Size(maxResults))
	//	search.Aggregation("test-values-"+field, elastic.NewTermsAggregation().Field(field+".value").Size(maxResults))
}

func main() {
	common.ElasticSearchServer = "http://elasticsearch.local:9200"
	client := common.CreateClient()

	search := client.Search().
		//		Index("dev-" + common.MediaIndexName).
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Pretty(true)

	addQuery(search)
	//	addAggregate(search)

	searchStartTime := time.Now()

	searchResult, err := search.Do()
	if err != nil {
		log.Fatal("Error searching: %s", err.Error())
	}
	searchDuration := time.Now().Sub(searchStartTime).Seconds() * 1000

	if searchResult.TotalHits() == 0 {
		log.Info("No matches were found, search took %d msecs (%01.3f round-trip)", searchResult.TookInMillis, searchDuration)
	} else {
		log.Info("Found %d matches; search took %d msecs (%01.3f round-trip):", searchResult.TotalHits(), searchResult.TookInMillis, searchDuration)
		for index, hit := range searchResult.Hits.Hits {
			var media common.Media
			err := json.Unmarshal(*hit.Source, &media)
			if err != nil {
				log.Error("Failed deserializing search result: %s", err.Error())
			} else {
				log.Info("  %d: %#v", index+1, media)

				log.Info("Sort info: %s; score: %d", sortToString(hit.Sort), hit.Score)
			}
		}

		if searchResult.Aggregations != nil {
			log.Info("Aggregation results:")
			emitAggregations(&searchResult.Aggregations, "")
		}
	}
}

func sortToString(sortArray []interface{}) string {
	values := []string{}
	for _, item := range sortArray {
		switch item.(type) {
		case float64:
			n := item.(float64)
			values = append(values, strconv.FormatFloat(n, 'f', -1, 64))
		}

	}

	return strings.Join(values, ", ")
	//	return fmt.Sprintf("%#v", sortArray)
}

func emitAggregations(aggs *elastic.Aggregations, prefix string) {
	for key, _ := range *aggs {
		terms, ok := aggs.Terms(key)
		if ok {
			log.Info("%s %s", prefix, key)
			for index, bucket := range terms.Buckets {

				v := ""
				switch bucket.Key.(type) {
				case string:
					v = bucket.Key.(string)
				case float64:
					// Assume it's a time, specifically, milliseconds since the epoch
					msec := int64(bucket.Key.(float64))
					v = fmt.Sprintf("%s", time.Unix(msec/1000, 0))

					if bucket.DocCount == 0 {
						continue
					}
				}
				log.Info("%s   %d: %s - %d", prefix, index+1, v, bucket.DocCount)

				if bucket.Aggregations != nil {
					emitAggregations(&bucket.Aggregations, prefix+"  ")
				}
			}
		}
	}

}
