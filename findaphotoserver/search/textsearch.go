package search

import (
	"gopkg.in/olivere/elastic.v5"

	"github.com/kevintavog/findaphoto/common"
)

type SearchOptions struct {
	Query            string
	Index            int
	Count            int
	CategoryOptions  *CategoryOptions
	DrilldownOptions *DrilldownOptions
}

//-------------------------------------------------------------------------------------------------
func NewSearchOptions(query string) *SearchOptions {
	return &SearchOptions{
		Query:            query,
		Index:            0,
		Count:            20,
		CategoryOptions:  NewCategoryOptions(),
		DrilldownOptions: NewDrilldownOptions(),
	}
}

func (so *SearchOptions) Search() (*SearchResult, error) {
	client := common.CreateClient()
	search := client.Search().
		Index(common.MediaIndexName).
		Type(common.MediaTypeName).
		Pretty(true)

	var query elastic.Query
	if so.Query == "" {
		query = elastic.NewMatchAllQuery()
	} else {
		query = elastic.NewQueryStringQuery(so.Query).
			Field("path"). // Folder name
			Field("monthname").
			Field("dayname").
			Field("keywords").
			Field("placename"). // Full reverse location lookup
			Field("tags")
	}

	search.Query(query)

	search.From(so.Index).Size(so.Count).Sort("datetime", false)
	return invokeSearch(search, &query, GroupByDate, so.CategoryOptions, so.DrilldownOptions, nil)
}
