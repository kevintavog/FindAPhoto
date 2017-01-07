package main

import (
	"github.com/kevintavog/findaphoto/common/clarifai"

	"github.com/garyburd/redigo/redis"
	"github.com/ian-kent/go-log/log"
	"github.com/jawher/mow.cli"
)

func dumpCommand(cmd *cli.Cmd) {
	cmd.Spec = "-s"
	redisServer := cmd.String(cli.StringOpt{Name: "s", Value: "", Desc: "The URL for the redis server (redis://localhost:6379)"})

	cmd.Action = func() {
		log.Info(appIntro)
		log.Info("  Redis server: %s", *redisServer)

		redisConnection, err := redis.DialURL(*redisServer)
		if err != nil {
			log.Fatalf("Unable to connect to Redis server: %s", err)
		}
		defer redisConnection.Close()

		scanKeys(redisConnection, "", dumpKeyContents, dumpKeyContentsComplete)
	}
}

func dumpKeyContents(c redis.Conn, key string, value string) {

	tagsAndProbs, _, err := clarifaifp.TagsAndProbabilitiesFromJson(value, 0)
	if err != nil {
		log.Error("Failed getting tags and probabilities from %s: %s", key, err)
	} else {
		for _, cc := range tagsAndProbs {
			minProbability = minInt8(minProbability, cc.Probability)
			maxProbability = maxInt8(maxProbability, cc.Probability)
		}

		log.Info("%s -- %s", key, tagsAndProbs)
	}
}

func dumpKeyContentsComplete() {
	log.Info("Min prob: %d, max prob: %d", minProbability, maxProbability)
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
