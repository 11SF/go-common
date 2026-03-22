package httpclient

import (
	"encoding/json"
	"fmt"
)

// HTTPError represents an HTTP error with status code and response data
type HTTPError struct {
	StatusCode   int    // HTTP status code
	ResponseBody []byte // Raw response body
	Message      string // Error message
	Err          error  // Underlying error (if any)
}

// Error implements the error interface
func (e *HTTPError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, string(e.ResponseBody))
}

// Unwrap returns the underlying error for errors.Is and errors.As
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// GetResponseBody returns the raw response body
func (e *HTTPError) GetResponseBody() []byte {
	return e.ResponseBody
}

// UnmarshalResponse unmarshals the response body into the provided struct
func (e *HTTPError) UnmarshalResponse(v any) error {
	if len(e.ResponseBody) == 0 {
		return fmt.Errorf("empty response body")
	}
	return json.Unmarshal(e.ResponseBody, v)
}

// NewHTTPError creates a new HTTPError
func NewHTTPError(statusCode int, responseBody []byte, message string) *HTTPError {
	return &HTTPError{
		StatusCode:   statusCode,
		ResponseBody: responseBody,
		Message:      message,
	}
}

// NewHTTPErrorWithErr creates a new HTTPError with an underlying error
func NewHTTPErrorWithErr(statusCode int, responseBody []byte, message string, err error) *HTTPError {
	return &HTTPError{
		StatusCode:   statusCode,
		ResponseBody: responseBody,
		Message:      message,
		Err:          err,
	}
}

// IsHTTPError checks if the error is an HTTPError
func IsHTTPError(err error) bool {
	_, ok := err.(*HTTPError)
	return ok
}

// AsHTTPError tries to convert an error to *HTTPError
func AsHTTPError(err error) (*HTTPError, bool) {
	httpErr, ok := err.(*HTTPError)
	return httpErr, ok
}
