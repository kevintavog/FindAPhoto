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
	app.Spec = "-r -e [-i] [-v] -a"
	indexPrefix := app.StringOpt("i", "", "The prefix for the index (for development) (optional)")
	elasticSearchServer := app.StringOpt("e", "", "The URL for the ElasticSearch server")
	redisServer := app.StringOpt("r", "", "The URL for the redis server")
	clarifaiApiKeyArg := app.StringOpt("a", "", "The clarifai.com API key")
	app.Version("v", "Show the version and exit")
	app.Action = func() {

		common.MediaIndexName = *indexPrefix + common.MediaIndexName
		common.ElasticSearchServer = *elasticSearchServer
		ClarifaiApiKey = *clarifaiApiKeyArg

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

		//		fakeClassifyFile("/Users/goatboy/Pictures/master/2017/2017-01-16 Troy and Lucas/Image.JPG")
		//		fakeClassifyFile("/Users/goatboy/Pictures/master/2017/2017-01-16 Troy and Lucas/IMG_8866_V.MP4")

		//		classifyFile("/Users/goatboy/Pictures/Master/1999/1999-07 Crested Butte/Image-12.JPG")
		//		classifyFile("/Users/goatboy/Pictures/Master/2018/2018-01-01 Seattle/IMG_0475.jpg")
		//		classifyFile("/Users/goatboy/Pictures/master/2017/2017-01-16 Troy and Lucas/IMG_8866_V.MP4")
		//		classifyFile("/Users/goatboy/Pictures/master/2017/2017-01-16 Troy and Lucas/IMG_8865_V.MP4")

		getClassifyMessages(*redisServer, esClient)
	}

	app.Run(os.Args)
}

func asJson(object interface{}) string {
	json, _ := json.MarshalIndent(object, "", "    ")
	return string(json)
}

func classifyFile(filePath string) {
	json, err := classifyV2(filePath)
	if err != nil {
		log.Fatal("Failed: %s", err)
	}

	tags, unitCount, err := clarifaifp.TagsAndProbabilitiesFromJson(json, 0)
	if err != nil {
		log.Fatalf("Getting v2 tags failed: %s", err)
	}

	log.Info("V2 tags, %d (unit count=%d): %s", len(tags), unitCount, filePath)
	for _, t := range tags {
		log.Info("  %s : %d", t.Name, t.Probability)
	}
}

func fakeClassifyFile(filePath string) {

	v2, _ := fakeClassifyV2(filePath)
	tags, unitCount, err := clarifaifp.TagsAndProbabilitiesFromJson(v2, 0)
	if err != nil {
		log.Fatalf("Getting v2 tags failed: %s", err)
	}

	log.Info("Fake V2 image tags, %d (unit count=%d): %s", len(tags), unitCount, filePath)
	for _, t := range tags {
		log.Info("  %s : %d", t.Name, t.Probability)
	}
}
