package main

import (
	"encoding/json"
	"fmt"
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
	// Good, useful queries
	//	query := elastic.NewTermQuery("_id", "1\\2001\\PDRM0300.JPG")
	//	query := elastic.NewTermQuery("path.raw", "1\\2001\\PDRM0300.JPG") // Search for a given path (can use _id instead...)
	//	query := elastic.NewGeoDistanceQuery("location").Lat(47.60863888888889).Lon(-122.43516666666666).Distance("20km")
	//	query := elastic.NewTermQuery("path", "2005") // 'path' is analyzed, so partial matches work
	//	query := elastic.NewMatchAllQuery()
	//	query := elastic.NewQueryStringQuery("path:2005 AND filename:100*") // A direct query, as typed in - good for adhoc/from or from search box in UI
	//	query := elastic.NewRangeQuery("datetime").Gte("2014/01/01").Lte("2014/12/31").Format("yyyy/MM/dd")
	//	query := elastic.NewWildcardQuery("date", "20140*")
	//	query := elastic.NewQueryStringQuery("monthname:September")	// string query is analyzed, so can find 'Sep' and 'sep'
	//	query := elastic.NewQueryStringQuery("filename:DSCN3380")
	//	query := elastic.NewQueryStringQuery("DSCN3380").Field("path").Field("monthname").Field("dayname").Field("keyword").Field("placename")
	//	query := elastic.NewQueryStringQuery("canada").Field("path").Field("monthname").Field("dayname").Field("keywords").Field("placename")
	//	query := elastic.NewQueryStringQuery("keywords:mount rainier")
	query := elastic.NewWildcardQuery("date", "*0220") // All matches for a given day

	//	query := addDateRangeQuery(search)

	//	query := elastic.NewQueryStringQuery("cityname:Seattle")
	//	query := elastic.NewMatchAllQuery()
	//	query = nil

	//	query := elastic.NewQueryStringQuery("keywords:mount rainier")

	// Useful queries, but with problems (likely case-sensitive issues)
	//	query := elastic.NewTermQuery("monthname", "September") // use string query instead: only lower case matches - as if query isn't being analyzed prior to search
	//	query := elastic.NewTermQuery("dayname", "saturday")	// use string query instead: only lower case matches - as if query isn't being analyzed prior to search
	//	query := elastic.NewWildcardQuery("filename", "dscn3*") // only lower case matches - as if query isn't being analyzed prior to search
	//	query := elastic.NewTermQuery("filename", "100_0304")

	// Work in progress
	//	query := elastic.NewTermQuery("filename", "DSCN3380")
	//	query := elastic.NewTermQuery("filename.raw", "DSCN*")
	//	query := elastic.NewTermQuery("path", "100_0304")

	if query != nil {
		search.Query(query).
			Size(10).
			Sort("datetime", false)
	} else {
		search.Size(0)
	}
}

func addAggregate(search *elastic.SearchService) {
	//	search.Aggregation("countryName", elastic.NewTermsAggregation().Field("countryname.value").Size(30))
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
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Pretty(true)

	addQuery(search)
	addAggregate(search)

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
				log.Info("  %d: %q", index+1, media)
			}
		}

		if searchResult.Aggregations != nil {
			log.Info("Aggregation results:")
			for key, _ := range searchResult.Aggregations {
				terms, ok := searchResult.Aggregations.Terms(key)
				if ok {
					log.Info(" %s", key)
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
						log.Info("   %d: %s - %d", index+1, v, bucket.DocCount)
					}
				}
			}
		}
	}
}
