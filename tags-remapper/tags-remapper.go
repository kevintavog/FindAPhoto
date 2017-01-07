package main

import (
	"encoding/json"
	"os"

	"github.com/kevintavog/findaphoto/common"

	"github.com/garyburd/redigo/redis"
	"github.com/ian-kent/go-log/log"
	"github.com/jawher/mow.cli"
)

const appIntro = "FindAPhoto tags-remapper"

type visitKey func(redis.Conn, string, string)
type visitComplete func()

func main() {
	common.InitDirectories("FindAPhoto")
	common.ConfigureLogging(common.LogDirectory, "findaphoto-tags-remapper")

	app := cli.App("tags-remapper", "The FindAPhoto tags remapper")
	app.Command("dump", "Dump all keys and tags/probabilities", dumpCommand)
	app.Command("map", "Map all keys matching a base to a new base", mapCommand)
	app.Command("count", "Count all keys matching a base", countMatchesCommand)

	app.Run(os.Args)
}

func scanKeys(c redis.Conn, matchPrefix string, visitFn visitKey, visitCompleteFn visitComplete) {

	totalKeysVisited := 0
	scanCursor := 0

	var keys []string
	for {
		var result []interface{}
		var err error

		if len(matchPrefix) > 0 {
			result, err = redis.Values(c.Do("SCAN", scanCursor, "MATCH", matchPrefix))
		} else {
			result, err = redis.Values(c.Do("SCAN", scanCursor))
		}

		if err != nil {
			log.Fatalf("Scan error for %d: %s", scanCursor, err)
		}

		scanCursor, err = redis.Int(result[0], nil)
		if err != nil {
			log.Error("Failed getting scan cursor: %s", err)
			continue
		}
		keys, err = redis.Strings(result[1], nil)
		if err != nil {
			log.Error("Failed getting keys from result: %s", err)
			continue
		}

		totalKeysVisited += len(keys)
		for _, k := range keys {
			value, err := redis.String(c.Do("GET", k))
			if err != redis.ErrNil && err != nil {
				log.Error("Unable to get value for key: %s (%s)", k, err)
			} else {
				visitFn(c, k, value)
			}
		}

		if scanCursor == 0 {
			break
		}
	}

	log.Info("Visited %d keys", totalKeysVisited)
	visitCompleteFn()
}

func asJson(object interface{}) string {
	json, _ := json.MarshalIndent(object, "", "    ")
	return string(json)
}
