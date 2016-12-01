package scanner

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/kevintavog/findaphoto/common"

	"github.com/ian-kent/go-log/log"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
)

var MediaScanned int64
var MediaRemoved int64

// Walk through the index, removing any items no longer on the file system
func RemoveFiles() {
	client := common.CreateClient()

	scrollService := client.Scroll(common.MediaIndexName).Type(common.MediaTypeName).Size(100)

	_, err := scrollService.Do(context.TODO())
	if err != nil {
		log.Error("Failed starting scan: %s", err.Error())
		return
	}

	checked := 0
	removed := 0
	for {
		results, err := scrollService.Do(context.TODO())
		if err == io.EOF {
			break
		}
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
			checked += 1

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
				removed += 1
				MediaRemoved += 1
				deleteResponse, err := client.Delete().
					Index(common.MediaIndexName).
					Type(common.MediaTypeName).
					Id(media.Path).
					Do(context.TODO())
				if err != nil {
					log.Error("Failed removing document '%s' from index: %s", media.Path, err.Error())
				} else if deleteResponse.Found != true {
					log.Error("Delete of document '%s' failed", media.Path)
				}
			}
		}
	}

	log.Info("Remover checked %d files and removed %d of them", checked, removed)
}
