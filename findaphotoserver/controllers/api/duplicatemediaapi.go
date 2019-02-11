package api

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/util"
	"github.com/labstack/echo"
)

type DuplicateSearchResult struct {
	TotalMatches int64
	ResultCount  int64
	Duplicates   []*common.DuplicateItem
}

func duplicateMediaAPI(c echo.Context) error {
	fc := c.(*util.FpContext)
	count := fc.IntFromQuery("count", 20)
	if count < 1 || count > 100 {
		panic(&util.InvalidRequest{Message: "count must be between 1 and 100, inclusive"})
	}

	index := fc.IntFromQuery("first", 1) - 1

	return fc.Time("duplicates", func() error {
		client := common.CreateClient()
		result, err := client.Search().
			Index(common.MediaIndexName).
			Type(common.DuplicateTypeName).
			Query(elastic.NewMatchAllQuery()).
			From(index).
			Size(count).
			Pretty(true).
			Do(context.TODO())
		if err != nil {
			panic(&util.InvalidRequest{Message: "Failed searching for duplicates", Err: err})
		}

		return c.JSON(http.StatusOK, processSearchResults(result))
	})
}

func processSearchResults(result *elastic.SearchResult) map[string]interface{} {
	dsr := make(map[string]interface{})
	dsr["totalMatches"] = result.TotalHits()

	duplicates := make([]map[string]interface{}, 0)

	resultCount := 0
	for _, hit := range result.Hits.Hits {
		di := &common.DuplicateItem{}
		err := json.Unmarshal(*hit.Source, di)
		if err != nil {
			panic(&util.InvalidRequest{Message: "JSON unmarshalling failed", Err: err})
		}

		item := make(map[string]interface{})
		item["ignoredPath"] = di.IgnoredPath
		item["existingPath"] = di.ExistingPath
		duplicates = append(duplicates, item)
		resultCount++
	}

	dsr["resultCount"] = resultCount
	dsr["duplicates"] = duplicates
	return dsr
}
