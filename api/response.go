package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/api/types"
)

type Response[T any] struct {
	Success    bool                      `json:"success"`
	Error      string                    `json:"error"`
	Result     []T                       `json:"result"`
	Pagination *types.PaginationResponse `json:"pagination,omitempty"`
}

func SuccessResponse[T any](c *gin.Context, result ...T) {
	if result == nil {
		result = make([]T, 0)
	}

	c.JSON(http.StatusOK, Response[T]{
		Success: true,
		Error:   "",
		Result:  result,
	})
}

func SuccessResponsePaginated[T any](c *gin.Context, pagination *types.PaginationResponse, result ...T) {
	if result == nil {
		result = make([]T, 0)
	}

	c.JSON(http.StatusOK, Response[T]{
		Success:    true,
		Error:      "",
		Result:     result,
		Pagination: pagination,
	})
}

func ErrorResponse(c *gin.Context, err error, result ...any) {
	if result == nil {
		result = make([]any, 0)
	}
	c.JSON(http.StatusBadRequest, Response[any]{
		Success: false,
		Error:   err.Error(),
		Result:  result,
	})

	logrus.
		WithField("error response code:", http.StatusBadRequest).
		WithField("request url:", c.Request.URL).
		Error(err)

	c.Abort()
}

func RateLimitErrorResponse(c *gin.Context, err error) {
	c.JSON(http.StatusTooManyRequests, Response[int]{
		Success: false,
		Error:   err.Error(),
		Result:  make([]int, 0),
	})

	logrus.
		WithField("error response code:", http.StatusTooManyRequests).
		WithField("request url:", c.Request.URL).
		Error(err)

	c.Abort()
}
