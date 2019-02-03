package main

import (
	"context"
	"fmt"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/common/clarifai"

	"github.com/garyburd/redigo/redis"
	"github.com/ian-kent/go-log/log"
	"github.com/jawher/mow.cli"

	"gopkg.in/olivere/elastic.v5"
)

var esClient *elastic.Client
var v1Count = 0

func dumpCommand(cmd *cli.Cmd) {
	cmd.Spec = "-s"
	redisServer := cmd.String(cli.StringOpt{Name: "s", Value: "", Desc: "The URL for the redis server (redis://localhost:6379)"})

	var err error
	esClient, err = elastic.NewSimpleClient(
		elastic.SetURL("http://jupiter/elasticsearch"),
		elastic.SetSniff(false))
	if err != nil {
		log.Fatalf("Unable to connect to '%s': %s", common.ElasticSearchServer, err.Error())
	}

	cmd.Action = func() {
		log.Info(appIntro)
		log.Info("  Redis server: %s", *redisServer)

		redisConnection, err := redis.DialURL(*redisServer)
		if err != nil {
			log.Fatalf("Unable to connect to Redis server: %s", err)
		}
		defer redisConnection.Close()

		scanKeys(redisConnection, "", dumpPrintKeyContents, dumpKeyContentsComplete)
	}
}

func dumpPrintKeyContents(c redis.Conn, key string, value string) bool {
	tagsAndProbs, _, err := clarifaifp.TagsAndProbabilitiesFromJSON(value, 0)
	if err != nil {
		log.Error("Failed getting tags and probabilities from %s: %s", key, err)
		return false
	}

	for _, cc := range tagsAndProbs {
		minProbability = minInt8(minProbability, cc.Probability)
		maxProbability = maxInt8(maxProbability, cc.Probability)
	}

	log.Info("%s -- %s", key, tagsAndProbs)
	return true
}

func dumpKeyContentsToElastic(c redis.Conn, key string, value string) bool {
	// Ignore V1 tags, no reason to include them
	_, _, err := clarifaifp.TagsAndProbabilitiesFromJSON(value, 0)
	if err != nil {
		if 0 == (v1Count % 1000) {
			fmt.Printf("v1: %s - %s\n", key, value)
		}
		v1Count++
		return true
	}
	_, err = esClient.Index().
		Index("clarifai_cache").
		Type("document").
		Id(key).
		BodyJson(value).
		Do(context.TODO())
	if err != nil {
		log.Error("FAILED writing to ElasticSearch: %s; %s (%s)", err, key, value)
		return false
	}
	return true
}

func dumpKeyContentsComplete() {
	log.Info("Min prob: %d, max prob: %d, v1 count: %d", minProbability, maxProbability, v1Count)
}

var minProbability int8 = 100
var maxProbability int8 = 0

func minInt8(a int8, b int8) int8 {
	if a < b {
		return a
	}
	return b
}

func maxInt8(a int8, b int8) int8 {
	if a > b {
		return a
	}
	return b
}
