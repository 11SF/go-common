package httpclient

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

// ClientConfig holds the configuration for the HTTP client.
type ClientConfig struct {
	BaseURL     string
	Timeout     time.Duration
	ContentType string
}

// HTTPClient wraps fasthttp.Client with custom configuration.
type HTTPClient struct {
	client *fasthttp.Client
	config ClientConfig
}

// NewHTTPClient initializes and returns a new HTTPClient.
func NewHTTPClient(config ClientConfig) *HTTPClient {
	return &HTTPClient{
		client: &fasthttp.Client{},
		config: config,
	}
}

// Get sends a GET request to the specified endpoint with optional query parameters.
func (hc *HTTPClient) Get(endpoint string, queryParams map[string]string, headers map[string]string) ([]byte, error) {
	url := hc.config.BaseURL + endpoint
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// Set URL and query params
	req.SetRequestURI(url)
	for key, value := range queryParams {
		req.URI().QueryArgs().Add(key, value)
	}

	req.Header.SetMethod(fasthttp.MethodGet)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Set timeout
	err := hc.client.DoTimeout(req, resp, hc.config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to make GET request: %w", err)
	}

	return resp.Body(), nil
}

// Post sends a POST request with a JSON payload.
func (hc *HTTPClient) Post(endpoint string, body interface{}, headers map[string]string) ([]byte, error) {
	url := hc.config.BaseURL + endpoint
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// Marshal body to JSON
	bodyData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType(hc.config.ContentType)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	req.SetBody(bodyData)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Set timeout
	err = hc.client.DoTimeout(req, resp, hc.config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to make POST request: %w", err)
	}

	return resp.Body(), nil
}

// Delete sends a DELETE request to the specified endpoint.
func (hc *HTTPClient) Delete(endpoint string, headers map[string]string) ([]byte, error) {
	url := hc.config.BaseURL + endpoint
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodDelete)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Set timeout
	err := hc.client.DoTimeout(req, resp, hc.config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to make DELETE request: %w", err)
	}

	return resp.Body(), nil
}
