package search

import (
	"strings"

	"encoding/json"

	"gopkg.in/olivere/elastic.v3"

	"github.com/kevintavog/findaphoto/common"
)

type SearchResult struct {
	TotalMatches int64
	Groups       []*SearchGroup
	ResultCount  int

	// Returned only for byday searches
	NextAvailableByDay     *ByDayResult
	PreviousAvailableByDay *ByDayResult
}

type ByDayResult struct {
	Month int
	Day   int
}

type SearchGroup struct {
	Name  string
	Items []*MediaHit
}

type MediaHit struct {
	Media      *common.Media
	DistanceKm *float64
}

const (
	GroupByAll = iota
	GroupByPath
	GroupByDate
)

//-------------------------------------------------------------------------------------------------

// Each search may return specific fields
type mappingAction func(searchHit *elastic.SearchHit, mediaHit *MediaHit)

//-------------------------------------------------------------------------------------------------
func invokeSearch(search *elastic.SearchService, groupBy int, extraMapping mappingAction) (*SearchResult, error) {
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
			mh := &MediaHit{Media: &common.Media{}}
			err := json.Unmarshal(*hit.Source, mh.Media)
			if err != nil {
				return nil, err
			}

			if extraMapping != nil {
				extraMapping(hit, mh)
			}

			groupName := groupName(mh.Media, groupBy)

			if lastGroup == "" || lastGroup != groupName {
				group = &SearchGroup{Name: groupName, Items: []*MediaHit{}}
				lastGroup = groupName
				sr.Groups = append(sr.Groups, group)
			}

			group.Items = append(group.Items, mh)
			sr.ResultCount += 1
		}
	}

	return sr, nil
}

func groupName(media *common.Media, groupBy int) string {
	switch groupBy {
	case GroupByAll:
		return "all"

	case GroupByPath:
		// Skip to first '\', take up to last '\'
		components := strings.Split(media.Path, "\\")

		if len(components) > 2 {
			return strings.Join(components[1:len(components)-1], "\\")
		}

	case GroupByDate:
		return media.DateTime.Format("2006-01-02")
	}

	return ""
}

func returnFirstMatch(search *elastic.SearchService) (*MediaHit, error) {
	result, err := search.Do()
	if err != nil {
		return nil, err
	}

	if result.TotalHits() > 0 {
		hit := result.Hits.Hits[0]
		mh := &MediaHit{Media: &common.Media{}}
		err := json.Unmarshal(*hit.Source, mh.Media)
		if err != nil {
			return nil, err
		}
		return mh, nil
	}

	return nil, nil
}
