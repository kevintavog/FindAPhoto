package common

var RedisServer = "http://localhost:6379"

type ClassifyMessage struct {
	File  string `json:"file"`
	Alias string `json:"alias"`
}
