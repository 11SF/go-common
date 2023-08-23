package ginresponse

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type response struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type responseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r *responseError) Error() string {
	return fmt.Sprintf("status %v: err %v", r.Code, r.Message)
}

func NewError(code string, err string) error {
	return &responseError{
		Code:    code,
		Message: err,
	}
}

func NewResponse(c *gin.Context, httpStatusCode int, data interface{}) {
	c.JSON(httpStatusCode, response{
		Code:    "00000",
		Message: "",
		Data:    data,
	})
}

func NewResponseError(c *gin.Context, httpStatusCode int, err error) {
	c.JSON(httpStatusCode, err)
}
