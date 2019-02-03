package main

import (
	"github.com/garyburd/redigo/redis"
	"github.com/ian-kent/go-log/log"
)

type fnRedisMessageHandler func(rawMessage []byte) error

func getClassifyMessages(redisServer string, handler fnRedisMessageHandler) {
	c, err := redis.DialURL(redisServer)
	if err != nil {
		log.Fatalf("Unable to connect to Redis server: %s", err)
	}
	defer c.Close()

	psc := redis.PubSubConn{Conn: c}
	psc.Subscribe("classify")
	defer psc.Close()

	redisConnection, err := redis.DialURL(redisServer)
	if err != nil {
		log.Fatalf("Unable to connect to Redis server: %s", err)
	}
	defer redisConnection.Close()

	skipClassifying := false

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			if skipClassifying {
				log.Warn("Skipping classification of %s due to earlier error", string(v.Data[:]))
				continue
			}
			err = handler(v.Data)
			if err != nil {
				skipClassifying = true
				log.Warn("Skipping the rest of the classifications due to %v", err)
			}
		case redis.Subscription:
			// Do nothing
		case error:
			log.Error("Error in PubSub: %s", v)
			return
		}
	}
}
