package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/labstack/echo/v4"
)

type Type[T any] struct {
	Code    ResponseCode `json:"code"`
	Message string       `json:"message"`
	Data    T            `json:"data"`
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

func NewGinResponse[T any](c *gin.Context, httpStatusCode int, data T) {
	c.JSON(httpStatusCode, Type[T]{
		Code:    "00000",
		Message: "",
		Data:    data,
	})
}

func NewGinResponseError(c *gin.Context, httpStatusCode int, err error) {
	c.JSON(httpStatusCode, err)
}

func NewEchoResponse[T any](c echo.Context, statusCode ResponseCode, data T) error {
	return c.JSON(http.StatusOK, Type[T]{
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

func NewFiberResponse[T any](c *fiber.Ctx, data T) error {
	return c.Status(http.StatusOK).JSON(Type[T]{
		Code:    "00000",
		Message: "success",
		Data:    data,
	})
}

func NewFiberResponseError(c *fiber.Ctx, httpStatusCode int, err error) error {
	errRes := &responseError{
		Code:    GenericError,
		Message: "Internal server error",
	}

	if err != nil {
		errRes.Message = err.Error()
	}

	if err, ok := err.(*responseError); ok {
		errRes = err
	}

	return c.Status(httpStatusCode).JSON(errRes)
}
