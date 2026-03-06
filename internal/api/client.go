package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spritz-finance/spritz-cli/internal/auth"
	"github.com/spritz-finance/spritz-cli/internal/config"
)

var Version = "dev"

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func NewFromEnv() (*Client, error) {
	apiKey, err := auth.GetAPIKey()
	if err != nil {
		return nil, err
	}
	return New(config.APIURL(), apiKey), nil
}

func (c *Client) do(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "spritz-cli/"+Version)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, parseAPIError(resp)
	}

	return resp, nil
}

func parseAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	apiErr := &APIError{StatusCode: resp.StatusCode}

	if err := json.Unmarshal(body, apiErr); err != nil || apiErr.Code == "" {
		apiErr.Code = http.StatusText(resp.StatusCode)
		if len(body) > 0 {
			apiErr.Message = string(body)
		}
	}

	return apiErr
}

func decodeJSON[T any](resp *http.Response) (T, error) {
	var v T
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("failed to decode response: %w", err)
	}
	return v, nil
}
