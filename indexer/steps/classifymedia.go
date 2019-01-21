package classifymedia

import (
	"encoding/json"

	"github.com/kevintavog/findaphoto/common"

	"github.com/garyburd/redigo/redis"
	"github.com/ian-kent/go-log/log"
)

var redisPool *redis.Pool

func Start() {
	if !common.IndexMakeNoChanges {
		redisPool = &redis.Pool{Dial: func() (redis.Conn, error) { return redis.DialURL(common.RedisServer) }}
	}
}

func Enqueue(fullPath string, alias string, tags *[]string) {
	if tags != nil {
		return
	}

	if !common.IndexMakeNoChanges {
		conn := redisPool.Get()
		defer conn.Close()

		classifyMessage := new(common.ClassifyMessage)
		classifyMessage.File = fullPath
		classifyMessage.Alias = alias

		details, err := json.Marshal(classifyMessage)
		if err != nil {
			log.Error("Failed converting to JSON for publishing to redis: %s", err)
		} else {
			_, err = conn.Do("PUBLISH", "classify", string(details))
			if err != nil {
				log.Error("Failed publishing to redis: %s", err)
			}
		}
	}
}
