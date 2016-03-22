package common

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"gopkg.in/olivere/elastic.v3"
)

import klog "github.com/ian-kent/go-log/log"

var ElasticSearchServer = "http://localhost:9200"

type warningUpdate struct {
	Warnings []string `json:"warnings,omitempty"`
}

func CreateClient() *elastic.Client {
	client, err := elastic.NewSimpleClient(
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

func AddWarning(client *elastic.Client, id string, newWarnings []string) {

	joinedWarnings := strings.Join(newWarnings, "; ")

	// Load existing document
	searchResult, err := client.Search().
		Index(MediaIndexName).
		Type(MediaTypeName).
		Query(elastic.NewTermQuery("_id", id)).
		Do()
	if err != nil {
		klog.Error("Unable to find document for warning: %q (on %q): %q", joinedWarnings, id, err.Error())
		return
	}

	if searchResult.TotalHits() == 0 {
		klog.Error("Unable to find document for warning: %q (on %q)", joinedWarnings, id)
		return
	} else {
		hit := searchResult.Hits.Hits[0]
		var media Media
		err := json.Unmarshal(*hit.Source, &media)
		if err != nil {
			klog.Error("Failed deserializing search result for %q: %q", id, err.Error())
			return
		}

		var updatedDoc warningUpdate
		updatedDoc.Warnings = append(media.Warnings, joinedWarnings)

		// Update document with full warning array
		_, err = client.Update().
			Fields().
			Index(MediaIndexName).
			Type(MediaTypeName).
			Id(id).
			Doc(updatedDoc).
			Do()
		if err != nil {
			klog.Error("Failed updating document with warning %q (on %q): %q", joinedWarnings, id, err.Error())
			return
		}
	}
}
