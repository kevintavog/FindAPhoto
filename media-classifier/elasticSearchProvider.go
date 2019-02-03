package main

import (
	"encoding/json"
	"fmt"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/common/clarifai"

	"github.com/ian-kent/go-log/log"

	"golang.org/x/net/context"

	"gopkg.in/olivere/elastic.v5"
)

func doesTagsFieldExistInElasticSearch(esClient *elastic.Client, alias string) (bool, error) {
	media, err := getMedia(esClient, alias)
	if err != nil {
		return false, err
	}

	if media == nil {
		return false, nil
	}

	return media.Tags != nil, nil
}

func addToElasticSearch(esClient *elastic.Client, alias string, tags []clarifaifp.ClarifaiTag) error {
	tagsOnly := make([]string, len(tags))
	for c, ct := range tags {
		tagsOnly[c] = ct.Name
	}

	update := make(map[string]interface{})
	update["tags"] = tagsOnly

	_, err := esClient.Update().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Id(alias).
		Doc(update).
		Do(context.TODO())

	return err
}

func getMedia(esClient *elastic.Client, alias string) (*common.Media, error) {
	pathSearchResult, err := esClient.Search().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Query(elastic.NewTermQuery("_id", alias)).
		Pretty(true).
		Do(context.TODO())
	if err != nil {
		return nil, err
	}

	if pathSearchResult.TotalHits() < 1 {
		return nil, nil
	}

	hit := pathSearchResult.Hits.Hits[0]
	var media common.Media
	err = json.Unmarshal(*hit.Source, &media)
	if err != nil {
		log.Error("Failed deserializing search result: %s", err.Error())
		return nil, err
	}

	return &media, nil
}

func getCachedClassification(esClient *elastic.Client, key string) (string, error) {
	response, err := esClient.Search().
		Index(common.ClarifaiCacheIndexName).
		Type(common.ClarifaiTypeName).
		Query(elastic.NewTermQuery("_id", key)).
		Do(context.TODO())
	if err != nil {
		return "", err
	}

	if response.TotalHits() < 1 {
		return "", nil
	}

	hit := response.Hits.Hits[0]
	// There has to be a better way to convert json.RawMarshall to a string...
	return fmt.Sprintf("%s", *hit.Source), nil
}
