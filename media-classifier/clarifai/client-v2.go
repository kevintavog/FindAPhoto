// Clarifai V2 API
package clarifaiv2

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func NewClarifaiError(statusCode int, textCode string, message string) error {
	return &ClarifaiError{StatusCode: statusCode, TextCode: textCode, Message: message}
}

type ImagePredictRequest struct {
	Inputs []ImagePredictInputRequest `json:"inputs"`
}

type ImagePredictInputRequest struct {
	Data struct {
		Image struct {
			Base64 string `json:"base64"`
		} `json:"image"`
	} `json:"data"`
}

type VideoPredictRequest struct {
	Inputs []VideoPredictInputRequest `json:"inputs"`
}

type VideoPredictInputRequest struct {
	Data struct {
		Video struct {
			Base64 string `json:"base64"`
		} `json:"video"`
	} `json:"data"`
}

func createPredictRequest(isImage bool, base64 string) ([]byte, error) {
	if isImage {
		req := &ImagePredictRequest{Inputs: make([]ImagePredictInputRequest, 0)}
		input := ImagePredictInputRequest{}
		input.Data.Image.Base64 = base64
		req.Inputs = append(req.Inputs, input)
		return json.Marshal(req)
	} else {
		req := &VideoPredictRequest{Inputs: make([]VideoPredictInputRequest, 0)}
		input := VideoPredictInputRequest{}
		input.Data.Video.Base64 = base64
		req.Inputs = append(req.Inputs, input)
		return json.Marshal(req)
	}
}

type ClarifaiError struct {
	StatusCode int
	TextCode   string
	Message    string
}

func (e *ClarifaiError) Error() string {
	return fmt.Sprintf("%d: %s; %s", e.StatusCode, e.TextCode, e.Message)
}

// Configurations
const (
	version2 = "v2"
	rootURL  = "https://api.clarifai.com"
)

// Client contains scoped variables for individual clients
type Client struct {
	APIKey    string
	APIRoot   string
	Throttled bool
}

// NewClient initializes a new Clarifai client
func NewClient(apiKey string) *Client {
	return &Client{apiKey, rootURL, false}
}

// Predict on a single photo/video
func (client *Client) Predict(isImage bool, filePath string) ([]byte, error) {
	// Use the generic model (aaa03c23b3724a16a56b629203edc62c)
	return client.fileHTTPRequest(isImage, filePath, "models/aaa03c23b3724a16a56b629203edc62c/outputs")
}

func (client *Client) fileHTTPRequest(isImage bool, filePath string, endpoint string) ([]byte, error) {
	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	jsonRequest, err := createPredictRequest(isImage, base64.StdEncoding.EncodeToString(fileContents))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", client.buildV2URL(endpoint), bytes.NewBuffer(jsonRequest))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Key "+client.APIKey)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	res, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	responseBody, bodyReadErr := ioutil.ReadAll(res.Body)

	switch res.StatusCode {
	case 200, 201:
		if client.Throttled {
			client.setThrottle(false)
		}
		return responseBody, bodyReadErr
	case 401:
		return nil, NewClarifaiError(res.StatusCode, "TOKEN_INVALID", "")
	case 429:
		client.setThrottle(true)
		return nil, NewClarifaiError(res.StatusCode, "THROTTLED", string(responseBody))
	case 400:
		return nil, NewClarifaiError(res.StatusCode, "ALL_ERROR", string(responseBody))
	case 402:
		return nil, NewClarifaiError(res.StatusCode, "EXCEEDED_ACCOUNT_LIMITS", string(responseBody))
	case 500:
		return nil, NewClarifaiError(res.StatusCode, "CLARIFAI_ERROR", string(responseBody))
	default:
		return nil, NewClarifaiError(res.StatusCode, "UNEXPECTED_STATUS_CODE", string(responseBody))
	}
}

func (client *Client) buildV2URL(endpoint string) string {
	parts := []string{client.APIRoot, version2, endpoint}
	return strings.Join(parts, "/")
}

func (client *Client) setAPIRoot(root string) {
	client.APIRoot = root
}

func (client *Client) setThrottle(throttle bool) {
	client.Throttled = throttle
}
