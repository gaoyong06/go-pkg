// Package error 提供错误处理中间件
package error

import (
	"net/http"

	"github.com/gaoyong06/go-pkg/errors"
	"github.com/gin-gonic/gin"
)

// ErrorResponse 是错误响应的标准格式
type ErrorResponse struct {
	Error ErrorData `json:"error"`
}

// ErrorData 包含错误的详细信息
type ErrorData struct {
	Code    string              `json:"code"`              // 错误代码
	Message string              `json:"message"`           // 错误消息
	Details []errors.ErrorDetail `json:"details,omitempty"` // 错误详情
}

// ErrorHandlerMiddleware 是一个 Gin 中间件，用于统一处理错误
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 处理请求前的逻辑
		c.Next()

		// 如果有错误，处理它们
		if len(c.Errors) > 0 {
			handleError(c, c.Errors.Last().Err)
		}
	}
}

// handleError 处理错误并返回适当的响应
func handleError(c *gin.Context, err error) {
	var statusCode int
	var errorResponse ErrorResponse

	// 检查是否为 APIError
	var apiErr *errors.APIError
	if errors.As(err, &apiErr) {
		statusCode = apiErr.StatusCode()
		errorResponse = ErrorResponse{
			Error: ErrorData{
				Code:    apiErr.Code,
				Message: apiErr.Message,
				Details: apiErr.Details,
			},
		}
	} else {
		// 未知错误
		statusCode = http.StatusInternalServerError
		errorResponse = ErrorResponse{
			Error: ErrorData{
				Code:    "INTERNAL_ERROR",
				Message: "服务器内部错误",
			},
		}
	}

	c.JSON(statusCode, errorResponse)
	c.Abort()
}

// HandleError 手动处理错误并返回响应
func HandleError(c *gin.Context, err error) {
	handleError(c, err)
}
