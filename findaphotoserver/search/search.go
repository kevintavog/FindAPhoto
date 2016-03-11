package search

import (
	"strings"

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
	Groups       []*SearchGroup
	ResultCount  int
}

type SearchGroup struct {
	Name  string
	Items []*common.Media
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

	sr := &SearchResult{TotalMatches: result.TotalHits()}
	if sr.TotalMatches > 0 {
		var lastGroup = ""
		sr.Groups = []*SearchGroup{}
		var group *SearchGroup

		for _, hit := range result.Hits.Hits {
			media := &common.Media{}
			err := json.Unmarshal(*hit.Source, media)
			if err != nil {
				return nil, err
			}

			groupName := groupName(media)

			if lastGroup == "" || lastGroup != groupName {
				group = &SearchGroup{Name: groupName, Items: []*common.Media{}}
				lastGroup = groupName
				sr.Groups = append(sr.Groups, group)
			}

			group.Items = append(group.Items, media)
			sr.ResultCount += 1
		}
	}

	return sr, nil
}

func groupName(media *common.Media) string {
	// Skip to first '\', take up to last '\'
	components := strings.Split(media.Path, "\\")

	if len(components) > 2 {
		return strings.Join(components[1:len(components)-1], "\\")
	}
	return ""
}
