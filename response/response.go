package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

type response struct {
	Code    ResponseCode `json:"code"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data"`
}

type responseError struct {
	Code    ResponseCode `json:"code"`
	Message string       `json:"message"`
}

func (r *responseError) Error() string {
	return fmt.Sprintf("status %v: err %v", r.Code, r.Message)
}

func NewError(code ResponseCode, err string) error {
	return &responseError{
		Code:    code,
		Message: err,
	}
}

func NewGinResponse(c *gin.Context, httpStatusCode int, data interface{}) {
	c.JSON(httpStatusCode, response{
		Code:    "00000",
		Message: "",
		Data:    data,
	})
}

func NewGinResponseError(c *gin.Context, httpStatusCode int, err error) {
	c.JSON(httpStatusCode, err)
}

func NewEchoResponse(c echo.Context, statusCode ResponseCode, data interface{}) error {
	return c.JSON(http.StatusOK, response{
		Code:    "00000",
		Message: "",
		Data:    data,
	})
}

func NewEchoResponseError(c echo.Context, httpStatusCode int, err error) error {
	errRes := &responseError{
		Code:    GenericError,
		Message: "Internal server error",
	}

	if err, ok := err.(*responseError); ok {
		errRes = err
	}

	return c.JSON(httpStatusCode, errRes)
}
