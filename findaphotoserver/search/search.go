package search

import (
	"encoding/json"

	"gopkg.in/olivere/elastic.v3"

	"github.com/kevintavog/findaphoto/common"
)

type SearchOptions struct {
	Query string
	Index int
	Count int
}

type SearchResult struct {
	TotalMatches int64
	Items        []*common.Media
	Count        int
}

func New(query string) *SearchOptions {
	return &SearchOptions{
		Query: query,
		Index: 0,
		Count: 20,
	}
}

func (so *SearchOptions) Search() (*SearchResult, error) {

	client := common.CreateClient()
	search := client.Search().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Pretty(true)

	if so.Query == "" {
		search.Query(elastic.NewMatchAllQuery())
	} else {
		search.Query(elastic.NewQueryStringQuery(so.Query).
			Field("path"). // Folder name
			Field("monthname").
			Field("dayname").
			Field("keywords").
			Field("placename")) // Full reverse location lookup
	}

	search.From(so.Index).Size(so.Count).Sort("datetime", false)

	result, err := search.Do()
	if err != nil {
		return nil, err
	}

	sr := &SearchResult{
		TotalMatches: result.TotalHits(),
	}

	if sr.TotalMatches > 0 {
		sr.Items = make([]*common.Media, len(result.Hits.Hits))
		for index, hit := range result.Hits.Hits {
			media := &common.Media{}
			err := json.Unmarshal(*hit.Source, media)
			if err != nil {
				return nil, err
			} else {
				sr.Items[index] = media
				sr.Count += 1
			}
		}

	}

	return sr, nil
}
