// Package response 提供统一响应格式中间件
package response

// ErrorHandler 错误处理接口
// 项目需要实现此接口来处理业务特定的错误逻辑
type ErrorHandler interface {
	// GetHTTPStatusCode 根据错误获取 HTTP 状态码
	GetHTTPStatusCode(err error) int

	// GetErrorMessage 获取错误消息
	// includeDetailed: 是否包含详细错误信息
	GetErrorMessage(err error, includeDetailed bool) string

	// GetErrorShowType 获取错误显示类型
	GetErrorShowType(err error) int

	// GetErrorCode 获取错误代码（字符串格式）
	GetErrorCode(err error) string
}

