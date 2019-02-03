package main

import (
	"encoding/json"
	"os"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/common/clarifai"

	"github.com/garyburd/redigo/redis"
	"github.com/ian-kent/go-log/log"
	"github.com/jawher/mow.cli"

	"golang.org/x/net/context"

	"gopkg.in/olivere/elastic.v5"
)

// Given an image or video, use the Clarafai.com tag API to come up with a set of terms.
// The raw response from Clarafai is stored in Redis, which has a persistent store - calls are
// made to Clarafai only if the media does NOT exist in Redis

var esClient *elastic.Client

func main() {
	common.InitDirectories("FindAPhoto")
	common.ConfigureLogging(common.LogDirectory, "findaphoto-media-classifier")

	app := cli.App("media-classifier", "The FindAPhoto media classifier")
	app.Spec = "-r -e [-i] [-v] -a"
	indexPrefix := app.StringOpt("i", "", "The prefix for the index (for development) (optional)")
	elasticSearchServer := app.StringOpt("e", "", "The URL for the ElasticSearch server")
	redisServer := app.StringOpt("r", "", "The URL for the redis server")
	clarifaiAPIKeyArg := app.StringOpt("a", "", "The clarifai.com API key")
	app.Version("v", "Show the version and exit")
	app.Action = func() {

		common.MediaIndexName = *indexPrefix + common.MediaIndexName
		common.ElasticSearchServer = *elasticSearchServer
		ClarifaiAPIKey = *clarifaiAPIKeyArg

		log.Info("FindAPhoto media classifier")
		log.Info("  Redis server: %s; ElasticSearch server: %s, using index %s; ", *redisServer, common.ElasticSearchServer, common.MediaIndexName)

		c, err := redis.DialURL(*redisServer)
		if err != nil {
			log.Fatalf("Unable to connect to Redis server: %s", err)
		}
		defer c.Close()

		esClient, err = elastic.NewSimpleClient(
			elastic.SetURL(common.ElasticSearchServer),
			elastic.SetSniff(false))
		if err != nil {
			log.Fatalf("Unable to connect to '%s': %s", common.ElasticSearchServer, err.Error())
		}

		exists, err := esClient.IndexExists(common.MediaIndexName).Do(context.TODO())
		if err != nil {
			log.Fatalf("Failed querying index: %s", err.Error())
		}
		if !exists {
			log.Fatalf("The index does not exist: %s", common.MediaIndexName)
		}

		// data := `{
		// 	"file": "/mnt/mediafiles/2018/2018-01-01 Seattle/IMG_0475.jpg",
		// 	"alias": "1\\2018\\2018-01-01 Seattle\\IMG_0475.jpg"
		// }`
		// err = handleRedisMessage([]byte(data))
		// if err != nil {
		// 	log.Fatalf("error: %s", err)
		// }

		getClassifyMessages(*redisServer, handleRedisMessage)
	}

	app.Run(os.Args)
}

// Dequeue from Redis; if the document hasn't been tagged, first try to retrieve
// the classification from the cache, otherwise call the Clarifai API.
// Add the result to the queue and update the media document.
func handleRedisMessage(rawMessage []byte) error {
	var classifyMessage common.ClassifyMessage
	err := json.Unmarshal(rawMessage, &classifyMessage)
	if err != nil {
		log.Error("Failed converting classify message: %s (%s)", err, string(rawMessage[:]))
		return nil
	}

	// Does the ElasticSearch document exist? And have the proper field already?
	exists, err := doesTagsFieldExistInElasticSearch(esClient, classifyMessage.Alias)
	if err != nil {
		log.Error("Failed checking ElasticSearch for %s: %s", classifyMessage.Alias, err)
		return nil
	}
	if exists {
		return nil
	}

	if _, err := os.Stat(classifyMessage.File); err != nil {
		if os.IsNotExist(err) {
			log.Error("The file '%s' does not exist", classifyMessage.File)
		} else {
			log.Error("Unable to access file '%s': %s", classifyMessage.File, err)
		}
		return nil
	}

	// Is the Clarifai response in the cache?
	json, err := getCachedClassification(esClient, classifyMessage.File)
	if err != nil || len(json) < 1 {
		json, err := classifyV2(classifyMessage.File)
		if err != nil {
			log.Error("Classification error for %s: %s (%s)\n", classifyMessage.File, err, json)
			return err
		}

		if len(json) == 0 {
			return nil
		}

		// Add the response to the cache.
		_, err = esClient.Index().
			Index(common.ClarifaiCacheIndexName).
			Type(common.ClarifaiTypeName).
			Id(classifyMessage.File).
			BodyJson(json).
			Do(context.TODO())
		if err != nil {
			return err
		}
	}

	tags, _, err := clarifaifp.TagsAndProbabilitiesFromJSON(json, 0)
	if err != nil {
		log.Error("Failed getting tags from JSON [%s] (%s): %s", classifyMessage.File, json, err)
		return nil
	}

	// Add it to the ElasticSearch document
	err = addToElasticSearch(esClient, classifyMessage.Alias, tags)
	if err != nil {
		log.Error("Failed updating ElasticSearch: %s -- %s", classifyMessage.Alias, err)
	}

	return nil
}
