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

func main() {
	common.InitDirectories("FindAPhoto")
	common.ConfigureLogging(common.LogDirectory, "findaphoto-media-classifier")

	app := cli.App("media-classifier", "The FindAPhoto media classifier")
	app.Spec = "-r -e [-i] [-v] -c -s"
	indexPrefix := app.StringOpt("i", "", "The prefix for the index (for development) (optional)")
	elasticSearchServer := app.StringOpt("e", "", "The URL for the ElasticSearch server")
	redisServer := app.StringOpt("r", "", "The URL for the redis server")
	clarisClientId := app.StringOpt("c", "", "The clarifai.com client id")
	clarisSecret := app.StringOpt("s", "", "The clarifai.com secret")
	app.Version("v", "Show the version and exit")
	app.Action = func() {

		common.MediaIndexName = *indexPrefix + common.MediaIndexName
		common.ElasticSearchServer = *elasticSearchServer
		ClientId = *clarisClientId
		ClientSecret = *clarisSecret

		log.Info("FindAPhoto media classifier")
		log.Info("  Redis server: %s; ElasticSearch server: %s, using index %s; ", *redisServer, common.ElasticSearchServer, common.MediaIndexName)

		c, err := redis.DialURL(*redisServer)
		if err != nil {
			log.Fatalf("Unable to connect to Redis server: %s", err)
		}
		defer c.Close()

		esClient, err := elastic.NewSimpleClient(
			elastic.SetURL(common.ElasticSearchServer),
			elastic.SetSniff(false))
		if err != nil {
			log.Fatal("Unable to connect to '%s': %s", common.ElasticSearchServer, err.Error())
		}

		exists, err := esClient.IndexExists(common.MediaIndexName).Do(context.TODO())
		if err != nil {
			log.Fatal("Failed querying index: %s", err.Error())
		}
		if !exists {
			log.Fatal("The index does not exist: %s", common.MediaIndexName)
		}

		err = checkClarifai()
		if err != nil {
			log.Fatal("Failed checking clarifai.com: %s", err)
		}

		//		classifyFile("/Users/goatboy/Pictures/Master/1999/1999-07 Crested Butte/Image-12.JPG")

		getClassifyMessages(*redisServer, esClient)
	}

	app.Run(os.Args)
}

func asJson(object interface{}) string {
	json, _ := json.MarshalIndent(object, "", "    ")
	return string(json)
}

func classifyFile(filePath string) {

	v2, _ := fakeClassifyV2(filePath)
	//	log.Info("V2: '%s'", v2)
	tags, unitCount, err := clarifaifp.TagsAndProbabilitiesFromJson(v2, 0)
	if err != nil {
		log.Fatalf("Getting v2 tags failed: %s", err)
	}

	log.Info("V2 tags, %d (unit count=%d)", len(tags), unitCount)
	for _, t := range tags {
		log.Info("  %s : %d", t.Name, t.Probability)
	}

	v1, _ := fakeClassify(filePath)
	//	log.Info("V1: '%s'", v1)
	tags, unitCount, err = clarifaifp.TagsAndProbabilitiesFromJson(v1, 0)
	if err != nil {
		log.Fatalf("Getting v1 tags failed: %s", err)
	}

	log.Info("V1 tags, %d (unit count=%d)", len(tags), unitCount)
	for _, t := range tags {
		log.Info("  %s : %d", t.Name, t.Probability)
	}

	//	response, err := classifyV2(filePath)
	//	if err != nil {
	//		log.Fatalf("Predict failed: %s", err)
	//	}

	//	log.Info("Response:")
	//	log.Info(string(response))
}
