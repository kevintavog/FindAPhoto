package helpers

import (
	"encoding/json"

	"github.com/ian-kent/go-log/log"
	"github.com/kevintavog/findaphoto/common"

	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
)

func InitializeDuplicates() {
	client := common.CreateClient()

	_, err := client.DeleteByQuery().
		Index(common.MediaIndexName).
		Type(common.DuplicateTypeName).
		Query(elastic.NewMatchAllQuery()).
		Do(context.TODO())

	if err != nil {
		log.Error("Failed removing all duplicate types: %s", err.Error())
	}
}

func AddDuplicateToIndex(client *elastic.Client, ignoredPath string, existingPath string) {
	dupItem := &common.DuplicateItem{
		IgnoredPath:  ignoredPath,
		ExistingPath: existingPath,
	}

	_, err := client.Index().
		Index(common.MediaIndexName).
		Type(common.DuplicateTypeName).
		Id(ignoredPath).
		BodyJson(dupItem).
		Do(context.TODO())

	if err != nil {
		log.Error("Failed adding DupliateItem %s: %s", ignoredPath, err.Error())
	}
}

func returnFirstMatch(result *elastic.SearchResult) *common.Media {
	if result.TotalHits() > 0 {
		hit := result.Hits.Hits[0]
		media := &common.Media{}
		err := json.Unmarshal(*hit.Source, media)
		if err != nil {
			return nil
		}
		return media
	}

	return nil
}
