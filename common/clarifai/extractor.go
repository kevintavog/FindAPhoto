package clarifaifp

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	//	"github.com/ian-kent/go-log/log"
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
			Concepts []ClarifaiConcept `json:"concepts,omitempty"`
			Frames   []struct {
				FrameInfo ClarifaiFrameInfo `json:"frame_info"`
				Data      struct {
					Concepts []ClarifaiConcept `json:"concepts"`
				} `json:"data"`
			} `json:"frames,omitempty"`
		} `json:"data"`
	} `json:"outputs"`
}

type ClarifaiFrameInfo struct {
	Index int `json:"index"`
	Time  int `json:"time"`
}

type ClarifaiConcept struct {
	Id    string  `json:"id"`
	Name  string  `json:"name"`
	Appid string  `json:"app_id,omitempty"`
	Value float32 `json:"value"`
}

type ClarifaiStatus struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func TagsAndProbabilitiesFromJson(clarifaiJson string, minProbability int8) ([]ClarifaiTag, int, error) {
	return tagsFromV2Json(clarifaiJson, minProbability)
}

func tagsFromV2Json(clarifaiJson string, minProbability int8) ([]ClarifaiTag, int, error) {

	// We skipped this item, there's nothing to do
	if len(clarifaiJson) == 0 {
		return nil, 0, nil
	}

	predictResponse := &PredictResponse{}
	err := json.Unmarshal([]byte(clarifaiJson), predictResponse)
	if err != nil {
		return nil, 0, err
	}

	if predictResponse.Status.Code == 0 {
		return nil, 0, fmt.Errorf("Mismatched JSON, status is 0")
	}

	unitCount := 0

	// Retain the highest rated probability for each tag - for images, there's only one but for
	// video, there is a set for every second of video.
	uniqueTags := make(map[string]int8)
	for _, outputs := range predictResponse.Outputs {
		unitCount++
		for _, concept := range outputs.Data.Concepts {
			prob := int8(concept.Value * 100)
			uniqueTags[concept.Name] = prob
		}
		for _, frame := range outputs.Data.Frames {
			for _, concept := range frame.Data.Concepts {
				prob := int8(concept.Value * 100)
				if value, exists := uniqueTags[concept.Name]; !exists || value < prob {
					uniqueTags[concept.Name] = prob
				}
			}
		}
	}

	tags := make([]ClarifaiTag, 0)
	for name, prob := range uniqueTags {
		tags = append(tags, ClarifaiTag{Name: name, Probability: prob})
	}

	sort.Sort(sort.Reverse(ClarifaiTagSort(tags)))
	return tags, unitCount, nil
}
