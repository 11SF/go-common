package httpclient

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/lo"
	"github.com/valyala/fasthttp"
)

type Client struct {
	Client               *fasthttp.Client
	Timeout              time.Duration
	DefaultHeaderOptions []HeaderOption
	Logger               *slog.Logger // optional custom logger; falls back to slog.Default() when EnableLogging is true
	EnableLogging        bool
}

func (c Client) logger() *slog.Logger {
	if !c.EnableLogging {
		return nil
	}
	if c.Logger != nil {
		return c.Logger
	}
	return slog.Default()
}

func (c Client) logRequest(method, url string, body []byte) {
	l := c.logger()
	if l == nil {
		return
	}
	args := []any{"method", method, "url", url}
	if len(body) > 0 {
		args = append(args, "body", string(body))
	}
	l.Debug("http request", args...)
}

func (c Client) logResponse(method, url string, statusCode int, body []byte) {
	l := c.logger()
	if l == nil {
		return
	}
	l.Debug("http response", "method", method, "url", url, "status", statusCode, "body", string(body))
}

type Response[T any] struct {
	Code     int
	Response T
}

// Get sends a GET request to the specified endpoint with optional query parameters.
func Get[RES any](client Client, url string, queryParams map[string]string, headers ...HeaderOption) (Response[RES], error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// Set URL and query params
	req.SetRequestURI(url)
	for key, value := range queryParams {
		req.URI().QueryArgs().Add(key, value)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.Header.SetMethod(fasthttp.MethodGet)
	for _, headerFunc := range lo.Concat(client.DefaultHeaderOptions, headers) {
		k, v := headerFunc()
		req.Header.Set(k, v)
	}

	client.logRequest(fasthttp.MethodGet, url, nil)

	err := client.Client.DoTimeout(req, resp, client.Timeout)
	if err != nil {
		return Response[RES]{}, fmt.Errorf("failed to make GET request: %w", err)
	}

	statusCode := resp.StatusCode()
	client.logResponse(fasthttp.MethodGet, url, statusCode, resp.Body())

	var res RES
	err = json.Unmarshal(resp.Body(), &res)
	if err != nil {
		return Response[RES]{}, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return Response[RES]{
		Code:     statusCode,
		Response: res,
	}, nil
}

// Post sends a POST request with a JSON payload.
func Post[REQ, RES any](client Client, url string, reqBody REQ, headers ...HeaderOption) (Response[RES], error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	bodyData, err := json.Marshal(reqBody)
	if err != nil {
		return Response[RES]{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodPost)
	for _, headerFunc := range lo.Concat(client.DefaultHeaderOptions, headers) {
		k, v := headerFunc()
		req.Header.Set(k, v)
	}
	req.SetBody(bodyData)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	client.logRequest(fasthttp.MethodPost, url, bodyData)

	err = client.Client.DoTimeout(req, resp, client.Timeout)
	if err != nil {
		return Response[RES]{}, fmt.Errorf("failed to make POST request: %w", err)
	}

	statusCode := resp.StatusCode()
	client.logResponse(fasthttp.MethodPost, url, statusCode, resp.Body())

	var res RES
	err = json.Unmarshal(resp.Body(), &res)
	if err != nil {
		return Response[RES]{}, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return Response[RES]{
		Code:     statusCode,
		Response: res,
	}, nil
}

// Delete sends a DELETE request to the specified endpoint.
func Delete[RES any](client Client, url string, headers ...HeaderOption) (Response[RES], error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodDelete)
	for _, headerFunc := range lo.Concat(client.DefaultHeaderOptions, headers) {
		k, v := headerFunc()
		req.Header.Set(k, v)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	client.logRequest(fasthttp.MethodDelete, url, nil)

	err := client.Client.DoTimeout(req, resp, client.Timeout)
	if err != nil {
		return Response[RES]{}, fmt.Errorf("failed to make DELETE request: %w", err)
	}

	statusCode := resp.StatusCode()
	client.logResponse(fasthttp.MethodDelete, url, statusCode, resp.Body())

	var res RES
	err = json.Unmarshal(resp.Body(), &res)
	if err != nil {
		return Response[RES]{}, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return Response[RES]{
		Code:     statusCode,
		Response: res,
	}, nil
}
