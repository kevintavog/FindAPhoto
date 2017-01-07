package clarifaifp

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ian-kent/go-log/log"
)

type ClarifaiTag struct {
	Name        string
	Probability int8
}

func (cc ClarifaiTag) String() string {
	return fmt.Sprintf("%s:%d", cc.Name, cc.Probability)
}

type ClarifaiTagSort []ClarifaiTag

func (s ClarifaiTagSort) Len() int      { return len(s) }
func (s ClarifaiTagSort) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ClarifaiTagSort) Less(i, j int) bool {
	return s[i].Probability < s[j].Probability
}

type PredictResponse struct {
	Status  ClarifaiStatus `json:"status"`
	Outputs []struct {
		Id        string         `json:"id"`
		Status    ClarifaiStatus `json:"status"`
		CreatedAt time.Time      `json:"created_at"`
		Model     struct {
			Name       string    `json:"name"`
			Id         string    `json:"id"`
			CreatedAt  time.Time `json:"created_at"`
			AppId      string    `json:"app_id,omitempty"`
			OutputInfo struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"output_info"`
			ModelVersion struct {
				Id        string         `json:"id"`
				CreatedAt time.Time      `json:"created_at"`
				Status    ClarifaiStatus `json:"status"`
			} `json:"model_version"`
		} `json:"model"`
		Input struct {
			Id   string `json:"id"`
			Data struct {
				Image struct {
					Url string `json:"url"`
				} `json:"image"`
			} `json:"data"`
		} `json:"input"`
		Data struct {
			Concepts []struct {
				Id    string  `json:"id"`
				Name  string  `json:"name"`
				Appid string  `json:"app_id,omitempty"`
				Value float32 `json:"value"`
			} `json:"concepts"`
		} `json:"data"`
	} `json:"outputs"`
}

type ClarifaiStatus struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

// Hackily included here to better handle the different way Classes & Probs are represented in
// JSON for images versus videos.
// TagResp represents the expected JSON response from /tag/
type ClarifAiTagResp struct {
	StatusCode    string `json:"status_code"`
	StatusMessage string `json:"status_msg"`
	Meta          struct {
		Tag struct {
			Timestamp json.Number `json:"timestamp"`
			Model     string      `json:"model"`
			Config    string      `json:"config"`
		}
	}
	Results []ClarifAiTagResult
}

// TagResult represents the expected data for a single tag result
type ClarifAiTagResult struct {
	DocID         *big.Int `json:"docid"`
	URL           string   `json:"url"`
	StatusCode    string   `json:"status_code"`
	StatusMessage string   `json:"status_msg"`
	LocalID       string   `json:"local_id"`
	Result        struct {
		Tag struct {
			Classes []interface{} `json:"classes"` // []string for images and [][]string for videos
			CatIDs  []string      `json:"catids"`
			Probs   []interface{} `json:"probs"` // []float64 for images and [][]float64 for videos
		}
	}
	DocIDString string `json:"docid_str"`
}

func TagsAndProbabilitiesFromJson(clarifaiJson string, minProbability int8) ([]ClarifaiTag, int, error) {
	tags, count, err := tagsFromV2Json(clarifaiJson, minProbability)
	if err != nil {
		tags, count, err = tagsFromV1Json(clarifaiJson, minProbability)
	}
	return tags, count, err
}

func tagsFromV2Json(clarifaiJson string, minProbability int8) ([]ClarifaiTag, int, error) {
	predictResponse := &PredictResponse{}
	err := json.Unmarshal([]byte(clarifaiJson), predictResponse)
	if err != nil {
		return nil, 0, err
	}

	// Mismatched JSON - perhaps v1?
	if predictResponse.Status.Code == 0 {
		return nil, 0, fmt.Errorf("Mismatched JSON, status is 0")
	}

	unitCount := 0
	tags := make([]ClarifaiTag, 0)

	for _, outputs := range predictResponse.Outputs {
		unitCount++
		tags = make([]ClarifaiTag, len(outputs.Data.Concepts))
		for idx, concept := range outputs.Data.Concepts {
			tags[idx] = ClarifaiTag{Name: concept.Name, Probability: int8(concept.Value * 100)}
		}
	}

	sort.Sort(sort.Reverse(ClarifaiTagSort(tags)))
	return tags, unitCount, nil
}

func tagsFromV1Json(clarifaiJson string, minProbability int8) ([]ClarifaiTag, int, error) {
	tagResponse := new(ClarifAiTagResp)
	err := json.Unmarshal([]byte(clarifaiJson), tagResponse)
	if err != nil {
		return nil, 0, err
	}

	uniqueTags := make(map[string]int8)
	unitCount := 0

	for _, doc := range tagResponse.Results {

		switch doc.Result.Tag.Classes[0].(type) {
		case string:
			unitCount += 1
		case []interface{}:
			unitCount += len(doc.Result.Tag.Classes)
		}

		for imageIndex, c := range doc.Result.Tag.Classes {

			switch c.(type) {
			case string:
				prob := int8(doc.Result.Tag.Probs[imageIndex].(float64) * 100)
				uniqueTags[c.(string)] = prob
			case []interface{}:
				probs := doc.Result.Tag.Probs[imageIndex].([]interface{})
				for videoIndex, i := range c.([]interface{}) {
					p := int8(probs[videoIndex].(float64) * 100)
					n := i.(string)
					if value, exists := uniqueTags[n]; !exists || value < p {
						uniqueTags[n] = p
					}
				}
			default:
				log.Error("  Unexpected tag class type for %#v", c)
			}
		}
	}

	tags := make([]ClarifaiTag, len(uniqueTags))
	idx := 0
	for n, p := range uniqueTags {

		if p >= minProbability {
			tags[idx] = ClarifaiTag{Name: n, Probability: p}
			idx++
		}
	}
	tags = tags[0:idx]
	sort.Sort(sort.Reverse(ClarifaiTagSort(tags)))
	return tags, unitCount, nil
}
