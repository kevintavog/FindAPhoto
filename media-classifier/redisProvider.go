package main

import (
	"encoding/json"
	"os"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/common/clarifai"

	"github.com/garyburd/redigo/redis"
	"github.com/ian-kent/go-log/log"

	"gopkg.in/olivere/elastic.v5"
)

func getClassifyMessages(redisServer string, esClient *elastic.Client) {
	c, err := redis.DialURL(redisServer)
	if err != nil {
		log.Fatalf("Unable to connect to Redis server: %s", err)
	}
	defer c.Close()

	psc := redis.PubSubConn{c}
	psc.Subscribe("classify")
	defer psc.Close()

	redisConnection, err := redis.DialURL(redisServer)
	if err != nil {
		log.Fatalf("Unable to connect to Redis server: %s", err)
	}
	defer redisConnection.Close()

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			handleClassifyMessage(redisConnection, esClient, v.Data)
		case redis.Subscription:
			// Do nothing
		case error:
			log.Error("Error in PubSub: %s", v)
			return
		}
	}
}

func handleClassifyMessage(c redis.Conn, esClient *elastic.Client, rawMessage []byte) {
	classifyMessage := new(common.ClassifyMessage)
	err := json.Unmarshal(rawMessage, classifyMessage)
	if err != nil {
		log.Error("Failed converting classify message: %s (%s)", err, string(rawMessage[:]))
		return
	}

	// Does the ElasticSearch document exist? And have the proper field already?
	exists, err := doesTagsFieldExistInElasticSearch(esClient, classifyMessage.Alias)
	if err != nil {
		log.Error("Failed checking ElasticSearch for %s: %s", classifyMessage.Alias, err)
		return
	}
	if exists {
		return
	}

	// Get media details from Clarifai
	json, err := addOrGet(c, classifyMessage.File)
	if err == nil && len(json) < 1 {
		//		log.Error("No error, but no JSON? %s", classifyMessage.File)
		return
	}

	if err != nil {
		log.Error("Failed getting classification details: %s", err)
		return
	}

	tags, _, err := clarifaifp.TagsAndProbabilitiesFromJson(json, 0)
	if err != nil {
		log.Error("Failed getting tags from JSON (%s): %s", json, err)
		return
	}

	// Add it to the ElasticSearch document
	err = addToElasticSearch(esClient, classifyMessage.Alias, tags)
	if err != nil {
		log.Error("Failed updating ElasticSearch: %s -- %s", classifyMessage.Alias, err)
	}
}

func addOrGet(c redis.Conn, fullPath string) (string, error) {
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			log.Error("The file '%s' does not exist", fullPath)
		} else {
			log.Error("Unable to access file '%s': %s", fullPath, err)
		}

		return "", err
	}

	existingValue, err := redis.String(c.Do("GET", fullPath))
	if err != redis.ErrNil && err != nil {
		log.Error("Unable to get key in redis: %s (%s)", fullPath, err)
		return "", err
	}

	// Has the item been classified? If so, there's nothing to do
	if len(existingValue) > 0 {
		return existingValue, nil
	}

	json, err := classifyV2(fullPath)
	if err != nil {
		log.Error("Classification error: %s (%s)\n", err, json)
		return "", err
	}

	if len(json) == 0 {
		return "", nil
	}

	err = c.Send("SET", fullPath, json)
	if err != nil {
		log.Error("Failed setting %s (%s)", fullPath, err)
		return "", err
	}
	err = c.Flush()
	if err != nil {
		log.Error("Failed flusing %s (%s)", fullPath, err)
		return "", err
	}

	return json, nil
}
