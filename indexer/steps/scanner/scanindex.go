package scanner

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/kevintavog/findaphoto/common"

	"github.com/ian-kent/go-log/log"
	"gopkg.in/olivere/elastic.v3"
)

var MediaScanned int64
var MediaRemoved int64

// Walk through the index, removing any items no longer on the file system
func RemoveFiles() {
	client := common.CreateClient()

	scrollResponse, err := client.Scroll(common.MediaIndexName).
		Size(100).
		Type(common.MediaTypeName).
		Do()
	if err != nil {
		log.Error("Failed starting scan: %s", err.Error())
		return
	}

	scrollId := scrollResponse.ScrollId
	for {
		results, err := client.Scroll(common.MediaIndexName).
			Size(100).
			ScrollId(scrollId).
			Do()
		if err != nil {
			if el, ok := err.(*elastic.Error); ok {
				if el.Status == http.StatusNotFound {
					break
				}
			}

			log.Error("Failed scanning index: %s", err.Error())
			break
		}

		for _, hit := range results.Hits.Hits {
			var media common.Media
			err := json.Unmarshal(*hit.Source, &media)
			if err != nil {
				log.Error("Failed deserializing search result: %s", err.Error())
				continue
			}

			MediaScanned += 1
			removeDocument := false
			if !common.IsValidAliasedPath(media.Path) {
				removeDocument = true
			} else {
				fullPath, err := common.FullPathForAliasedPath(media.Path)
				if err != nil {
					log.Error("Unable to convert %s to a path: %s", media.Path, err.Error())
					continue
				}

				if _, err = os.Stat(fullPath); os.IsNotExist(err) {
					removeDocument = true
					log.Info("File doesn't exist: %s", fullPath)
				}
			}

			if removeDocument {
				MediaRemoved += 1
				deleteResponse, err := client.Delete().
					Index(common.MediaIndexName).
					Type(common.MediaTypeName).
					Id(media.Path).
					Do()
				if err != nil {
					log.Error("Failed removing document '%s' from index: %s", media.Path, err.Error())
				} else if deleteResponse.Found != true {
					log.Error("Delete of document '%s' failed", media.Path)
				}
			}
		}
	}
}
