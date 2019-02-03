package main

import (
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/ian-kent/go-log/log"
	"github.com/jawher/mow.cli"
)

var makeChanges bool
var sourcePrefix string
var destinationPrefix string
var countMappedKeys int
var destinationConnection redis.Conn
var keysAdded int

func mapCommand(cmd *cli.Cmd) {
	cmd.Spec = "[-n] -s [-d] --spre --dpre"
	noChanges := cmd.Bool(cli.BoolOpt{Name: "n", Value: false, Desc: "Make no changes, emit information on what will happen"})
	sourceServer := cmd.String(cli.StringOpt{Name: "s", Value: "", Desc: "The URL for the source redis server (redis://localhost:6379)"})
	destServerOption := cmd.String(cli.StringOpt{Name: "d", Value: "", Desc: "The URL for the destination redis server (redis://localhost:6379)"})
	sourcePrefixOption := cmd.String(cli.StringOpt{Name: "spre", Value: "", Desc: "The prefix for the source keys"})
	destinationPrefixOption := cmd.String(cli.StringOpt{Name: "dpre", Value: "", Desc: "The prefix for the destination keys"})

	cmd.Action = func() {
		destinationServer := *destServerOption
		if len(destinationServer) < 1 {
			destinationServer = *sourceServer
			log.Info("NOTE: Source and destination server are the same: %s", destinationServer)
		}

		makeChanges = !*noChanges
		sourcePrefix = *sourcePrefixOption
		destinationPrefix = *destinationPrefixOption

		log.Info(appIntro)
		log.Info("  Redis source server: %s; destination server: %s", *sourceServer, destinationServer)
		log.Info("  Mapping from '%s' to '%s'", sourcePrefix, destinationPrefix)
		if !makeChanges {
			log.Info("NO CHANGES WILL BE MADE")
		}

		redisConnection, err := redis.DialURL(*sourceServer)
		if err != nil {
			log.Fatalf("Unable to connect to source Redis server '%s': %s", *sourceServer, err)
		}
		defer redisConnection.Close()

		destinationConnection, err = redis.DialURL(destinationServer)
		if err != nil {
			log.Fatalf("Unable to connect to destination Redis server '%s': %s", destinationServer, err)
		}

		scanKeys(redisConnection, sourcePrefix+"*", mapKey, mapKeysComplete)
	}
}

func mapKey(c redis.Conn, key string, value string) bool {
	if !strings.HasPrefix(key, sourcePrefix) {
		log.Warn("Key doesn't match the prefix? %s -- %s", key, sourcePrefix)
		return false
	}

	node := key[len(sourcePrefix):]
	destinationKey := destinationPrefix + node // Need to handle extra & missing '/' between the two strings

	destValue, err := redis.String(destinationConnection.Do("GET", destinationKey))
	addDestionationKey := false
	if err != redis.ErrNil && err != nil {
		addDestionationKey = true
	} else {
		addDestionationKey = value != destValue
	}

	if addDestionationKey {
		keysAdded++
		if makeChanges {
			err = destinationConnection.Send("SET", destinationKey, value)
			if err != nil {
				log.Error("Failed setting %s (%s)", destinationKey, err)
			} else {
				err = destinationConnection.Flush()
				if err != nil {
					log.Error("Failed flusing %s (%s)", destinationKey, err)
				}
			}
		}
	}

	return true
}

func mapKeysComplete() {
	log.Info("Mapping completed, added %d keys", keysAdded)
}
