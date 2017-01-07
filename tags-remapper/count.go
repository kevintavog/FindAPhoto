package main

import (
	"github.com/garyburd/redigo/redis"
	"github.com/ian-kent/go-log/log"
	"github.com/jawher/mow.cli"
)

var matchingKeyCount int

func countMatchesCommand(cmd *cli.Cmd) {
	cmd.Spec = "-s -k"
	redisServer := cmd.String(cli.StringOpt{Name: "s", Value: "", Desc: "The URL for the redis server (redis://localhost:6379)"})
	keyPrefix := cmd.String(cli.StringOpt{Name: "k", Value: "", Desc: "The prefix to count ('/Users/name/Pictures' or /mnt/mediafiles'"})

	cmd.Action = func() {
		log.Info(appIntro)
		log.Info("  Redis server: %s; counting matches for '%s'", *redisServer, *keyPrefix)

		redisConnection, err := redis.DialURL(*redisServer)
		if err != nil {
			log.Fatalf("Unable to connect to Redis server: %s", err)
		}
		defer redisConnection.Close()

		scanKeys(redisConnection, *keyPrefix+"*", countKeyMatches, countKeyMatchesComplete)
	}
}

func countKeyMatches(c redis.Conn, key string, value string) {
	matchingKeyCount += 1
}

func countKeyMatchesComplete() {
	log.Info("Matching key count: %d", matchingKeyCount)
}
