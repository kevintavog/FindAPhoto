package api

import (
	"encoding/json"

	"github.com/go-playground/lars"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/applicationglobals"
)

type DuplicateSearchResult struct {
	TotalMatches int64
	ResultCount  int64
	Duplicates   []*common.DuplicateItem
}

func DuplicateMedia(c lars.Context) {
	fc := c.(*applicationglobals.FpContext)
	err := fc.Ctx.ParseForm()
	if err != nil {
		panic(&InvalidRequest{message: "parseFormError", err: err})
	}

	count := intFromQuery(fc.Ctx, "count", 20)
	if count < 1 || count > 100 {
		panic(&InvalidRequest{message: "count must be between 1 and 100, inclusive"})
	}

	index := intFromQuery(fc.Ctx, "first", 1) - 1

	fc.AppContext.FieldLogger.Time("duplicates", func() {
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
			panic(&InvalidRequest{message: "Failed searching for duplicates", err: err})
		}

		fc.WriteResponse(processSearchResults(result))
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
			panic(&InternalError{message: "JSON unmarshalling failed", err: err})
		}

		item := make(map[string]interface{})
		item["ignoredPath"] = di.IgnoredPath
		item["existingPath"] = di.ExistingPath
		duplicates = append(duplicates, item)
		resultCount += 1
	}

	dsr["resultCount"] = resultCount
	dsr["duplicates"] = duplicates
	return dsr
}
