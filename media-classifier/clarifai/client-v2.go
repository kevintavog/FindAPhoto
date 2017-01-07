// Clarifai V2 API
package clarifaiv2

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func NewClarifaiError(statusCode int, textCode string, message string) error {
	return &ClarifaiError{StatusCode: statusCode, TextCode: textCode, Message: message}
}

type PredictRequest struct {
	Inputs []PredictInputRequest `json:"inputs"`
}

type PredictInputRequest struct {
	Data struct {
		Image struct {
			Base64 string `json:"base64"`
		} `json:"image"`
	} `json:"data"`
}

func createPredictRequest() *PredictRequest {

	p := make([]PredictInputRequest, 0)

	return &PredictRequest{Inputs: p}
}

func (pr *PredictRequest) addFileInput(base64 string) {
	input := PredictInputRequest{}
	input.Data.Image.Base64 = base64

	pr.Inputs = append(pr.Inputs, input)
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
	version1 = "v1"
	version2 = "v2"
	rootURL  = "https://api.clarifai.com"
)

// Client contains scoped variables forindividual clients
type Client struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	APIRoot      string
	Throttled    bool
}

// TokenResp is the expected response from /token/
type TokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

var cachedToken *TokenResp = nil
var cachedTokenExpireTimestamp time.Time

// NewClient initializes a new Clarifai client
func NewClient(clientID, clientSecret string) *Client {
	return &Client{clientID, clientSecret, "", rootURL, false}
}

// Predict on a single photo/video
func (client *Client) Predict(filePath string) ([]byte, error) {
	// Use the generic model (aaa03c23b3724a16a56b629203edc62c)
	return client.fileHTTPRequest(filePath, "models/aaa03c23b3724a16a56b629203edc62c/outputs")
}

func (client *Client) fileHTTPRequest(filePath string, endpoint string) ([]byte, error) {
	err := client.requestAccessToken()
	if err != nil {
		return nil, err
	}

	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	request := createPredictRequest()
	request.addFileInput(base64.StdEncoding.EncodeToString(fileContents))

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", client.buildV2URL(endpoint), bytes.NewBuffer(jsonRequest))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+client.AccessToken)
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
	case 500:
		return nil, NewClarifaiError(res.StatusCode, "CLARIFAI_ERROR", string(responseBody))
	default:
		return nil, NewClarifaiError(res.StatusCode, "UNEXPECTED_STATUS_CODE", string(responseBody))
	}
}

func (client *Client) requestAccessToken() error {
	if cachedToken != nil && time.Now().Before(cachedTokenExpireTimestamp) {
		client.setAccessToken(cachedToken.AccessToken)
		return nil
	}

	fmt.Printf("%s Getting access token\n", time.Now())
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", client.ClientID)
	form.Set("client_secret", client.ClientSecret)
	formData := strings.NewReader(form.Encode())

	req, err := http.NewRequest("POST", client.buildV1URL("token"), formData)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+client.AccessToken)
	req.Header.Set("Content-Length", strconv.Itoa(len(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := &http.Client{}
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	token := new(TokenResp)
	err = json.Unmarshal(body, token)
	if err != nil {
		return err
	}

	cachedToken = token
	cachedTokenExpireTimestamp = time.Now().Add(time.Duration(cachedToken.ExpiresIn) * time.Second)
	client.setAccessToken(token.AccessToken)

	fmt.Printf("Token expires at %s\n", cachedTokenExpireTimestamp)
	return nil
}

func (client *Client) buildV1URL(endpoint string) string {
	parts := []string{client.APIRoot, version1, endpoint}
	return strings.Join(parts, "/")
}

func (client *Client) buildV2URL(endpoint string) string {
	parts := []string{client.APIRoot, version2, endpoint}
	return strings.Join(parts, "/")
}

// SetAccessToken will set accessToken to a new value
func (client *Client) setAccessToken(token string) {
	client.AccessToken = token
}

func (client *Client) setAPIRoot(root string) {
	client.APIRoot = root
}

func (client *Client) setThrottle(throttle bool) {
	client.Throttled = throttle
}
