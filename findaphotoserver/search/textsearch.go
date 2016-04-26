package search

import (
	"gopkg.in/olivere/elastic.v3"

	"github.com/kevintavog/findaphoto/common"
)

type SearchOptions struct {
	Query string
	Index int
	Count int
}

//-------------------------------------------------------------------------------------------------
func NewSearchOptions(query string) *SearchOptions {
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
	return invokeSearch(search, GroupByPath, nil)
}
