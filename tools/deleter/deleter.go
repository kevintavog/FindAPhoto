package main

import (
	_ "encoding/json"
	_ "fmt"

	"github.com/kevintavog/findaphoto/common"

	"github.com/ian-kent/go-log/log"
	"gopkg.in/olivere/elastic.v3"
)

func main() {
	common.ElasticSearchServer = "http://elasticsearch.local:9200"
	client := common.CreateClient()

	query := elastic.NewWildcardQuery("path", "2*")
	deleteResult, err := client.DeleteByQuery().
		Index("dev-" + common.MediaIndexName).
		Type(common.MediaTypeName).
		Query(query).
		Pretty(true).
		Do()
	if err != nil {
		log.Error("Error deleting documents: %s", err.Error())
	}

	for _, failure := range deleteResult.Failures {
		log.Error("Failure: %q", failure.Status)
	}
}
