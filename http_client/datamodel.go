package httpclient

import "errors"

type CommonResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var (
	ErrorHTTPClient      = errors.New("http client error")
	ErrorHTTPStatusNotOk = errors.New("http status not ok")
)
