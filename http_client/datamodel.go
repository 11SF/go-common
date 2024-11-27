package httpclient

type CommonResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var ContentTypeJSON = "application/json"
