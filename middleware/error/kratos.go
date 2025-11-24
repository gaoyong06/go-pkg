// Package error 提供错误处理中间件
package error

import (
	"context"

	"github.com/gaoyong06/go-pkg/errors"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
)

// KratosErrorResponse 是 Kratos 错误响应的标准格式
type KratosErrorResponse struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Details []errors.ErrorDetail `json:"details,omitempty"`
}

// KratosErrorHandlerMiddleware 是一个 Kratos 中间件，用于统一处理错误
func KratosErrorHandlerMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 调用下一个处理器
			resp, err := handler(ctx, req)
			if err == nil {
				return resp, nil
			}

			// 处理错误
			return handleKratosError(ctx, err)
		}
	}
}

// handleKratosError 处理错误并返回适当的响应
func handleKratosError(ctx context.Context, err error) (interface{}, error) {
	// 检查是否为 APIError
	var apiErr *errors.APIError
	if errors.As(err, &apiErr) {
		// 创建 Kratos 错误
		kratosErr := kerrors.New(
			apiErr.StatusCode(),
			apiErr.Code,
			apiErr.Message,
		)

		// 添加详细信息
		if len(apiErr.Details) > 0 {
			metadata := make(map[string]string)
			for _, detail := range apiErr.Details {
				metadata[detail.Field] = detail.Message
			}
			
			// 注意：Kratos v2 中没有 WithMetadata 方法
			// 这里我们只能在日志中记录这些详细信息
			// 实际应用中可能需要使用自定义错误类型
		}

		return nil, kratosErr
	}

	// 未知错误
	return nil, kerrors.InternalServer("INTERNAL_ERROR", "服务器内部错误")
}

// HandleKratosError 手动处理错误并返回响应
func HandleKratosError(ctx context.Context, err error) (interface{}, error) {
	return handleKratosError(ctx, err)
}
