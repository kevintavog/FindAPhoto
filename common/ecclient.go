package common

import (
	"log"
	"os"

	"gopkg.in/olivere/elastic.v3"
)

import klog "github.com/ian-kent/go-log/log"

var ElasticSearchServer = "http://localhost:9200"

func CreateClient() *elastic.Client {
	client, err := elastic.NewClient(
		elastic.SetURL(ElasticSearchServer),
		elastic.SetSniff(false),
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC-error ", log.LstdFlags)),
		//		elastic.SetTraceLog(log.New(os.Stderr, "ELASTIC-trace ", log.LstdFlags)),
		elastic.SetHealthcheck(false))

	if err != nil {
		klog.Fatal("Unable to connect to '%s': %s", ElasticSearchServer, err.Error())
	}

	return client
}
