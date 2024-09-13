package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

var DefaultHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

type HTTPMethod string

// Supported HTTP methods.
const (
	GET HTTPMethod = "GET"
)

// RestClient is a custom HTTP client that can be extended with additional features.
type RestClient struct{}

func NewRestClient() *RestClient {
	return &RestClient{}
}

type RequestConfig struct {
	Method      HTTPMethod
	URL         string
	Headers     map[string]string
	QueryParams map[string]string
	Body        []byte
}

type HTTPResponse struct {
	StatusCode int
	Body       string
	Headers    map[string][]string
}

var ErrNilRequest = errors.New("nil http.Request received")

// executeRequest sends the HTTP request using DefaultHTTPClient and returns the HTTP response.
func executeRequest(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	start := time.Now()

	resp, err := DefaultHTTPClient.Do(req)
	if err != nil {
		log.Printf("Request failed; URL: %s, Method: %s, Error: %v", req.URL, req.Method, err)

		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("request timeout: %w", err)
		}

		return nil, err
	}

	log.Printf("URL: %s, Method: %s, Response Time: %s, Status Code: %d", req.URL, req.Method, time.Since(start), resp.StatusCode)

	if 200 <= resp.StatusCode && resp.StatusCode <= 299 {
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return resp, err
		}

		// Restore the response body with the read content.
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	return resp, err
}

func (c *RestClient) Get(urlPath string, args ...interface{}) (*HTTPResponse, error) {
	var queryParams map[string]string
	var headers map[string]string

	if len(args) > 0 {
		if qp, ok := args[0].(map[string]string); ok {
			queryParams = qp
		}
	}

	if len(args) > 1 {
		if hdrs, ok := args[1].(map[string]string); ok {
			headers = hdrs
		}
	}

	requestConfig := RequestConfig{
		Method:      GET,
		URL:         urlPath,
		Headers:     headers,
		QueryParams: queryParams,
	}

	req, err := createHTTPRequest(requestConfig)
	if err != nil {
		return nil, err
	}

	resp, err := executeRequest(req)
	if err != nil {
		return nil, err
	}

	return parseHTTPResponse(resp)
}

// createHTTPRequest constructs an HTTP request from the given configuration.
func createHTTPRequest(config RequestConfig) (*http.Request, error) {
	if len(config.QueryParams) > 0 {
		config.URL = appendQueryParams(config.URL, config.QueryParams)
	}

	req, err := http.NewRequest(string(config.Method), config.URL, bytes.NewBuffer(config.Body))
	if err != nil {
		return nil, err
	}

	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	if len(config.Body) > 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// parseHTTPResponse reads and parses the HTTP response.
func parseHTTPResponse(resp *http.Response) (*HTTPResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	resp.Body.Close()

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       string(body),
		Headers:    resp.Header,
	}, nil
}

// appendQueryParams adds query parameters to the URL.
func appendQueryParams(baseURL string, queryParams map[string]string) string {
	urlParts := &url.URL{Path: baseURL}
	query := url.Values{}
	for key, value := range queryParams {
		query.Add(key, value)
	}
	urlParts.RawQuery = query.Encode()
	return urlParts.String()
}
