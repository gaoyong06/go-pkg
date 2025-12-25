package response

import (
	"errors"
	"fmt"

	kratosErrors "github.com/go-kratos/kratos/v2/errors"
)

// DefaultErrorHandler 实现 ErrorHandler 接口，提供通用的错误处理逻辑
type DefaultErrorHandler struct {
	statusMapping   map[int]int
	showTypeMapping map[int]int
}

// HandlerOption 配置 DefaultErrorHandler 的可选参数
type HandlerOption func(*DefaultErrorHandler)

// WithStatusMapping 设置错误码到 HTTP 状态码的映射
func WithStatusMapping(mapping map[int]int) HandlerOption {
	return func(h *DefaultErrorHandler) {
		for code, status := range mapping {
			h.statusMapping[code] = status
		}
	}
}

// WithShowTypeMapping 设置错误码到 ShowType 的映射
func WithShowTypeMapping(mapping map[int]int) HandlerOption {
	return func(h *DefaultErrorHandler) {
		for code, showType := range mapping {
			h.showTypeMapping[code] = showType
		}
	}
}

// NewDefaultErrorHandler 创建默认的错误处理器
func NewDefaultErrorHandler(opts ...HandlerOption) ErrorHandler {
	handler := &DefaultErrorHandler{
		statusMapping:   make(map[int]int),
		showTypeMapping: make(map[int]int),
	}

	for _, opt := range opts {
		opt(handler)
	}

	return handler
}

// GetHTTPStatusCode 根据错误获取 HTTP 状态码
func (h *DefaultErrorHandler) GetHTTPStatusCode(err error) int {
	code, ok := extractErrorCode(err)
	if !ok {
		return 500
	}

	if status, exists := h.statusMapping[code]; exists {
		return status
	}

	// 如果错误码本身就是合法的 HTTP 状态码，则直接返回
	if code >= 100 && code <= 599 {
		return code
	}

	// 默认返回 400 Bad Request
	return 400
}

// GetErrorMessage 获取错误消息
func (h *DefaultErrorHandler) GetErrorMessage(err error, includeDetailed bool) string {
	var kratosErr *kratosErrors.Error
	if errors.As(err, &kratosErr) {
		// 由于 WrapError 已经在包装时提取并合并了 gRPC 错误信息到 message 中
		// 这里直接返回 kratos error 的 message 即可
		// 对于需要详细信息的场景，message 中已经包含了 gRPC 错误信息
		return kratosErr.Message
	}

	if includeDetailed {
		return err.Error()
	}

	return "操作失败，请重试或联系管理员"
}

// GetErrorShowType 获取错误显示类型
func (h *DefaultErrorHandler) GetErrorShowType(err error) int {
	code, ok := extractErrorCode(err)
	if !ok {
		return ShowTypeErrorMessage
	}

	if showType, exists := h.showTypeMapping[code]; exists {
		return showType
	}

	status := h.GetHTTPStatusCode(err)
	switch status {
	case 400:
		return ShowTypeWarnMessage
	case 401, 403:
		return ShowTypeNotification
	case 404:
		return ShowTypeWarnMessage
	case 500:
		return ShowTypeErrorMessage
	default:
		return ShowTypeErrorMessage
	}
}

// GetErrorCode 获取错误代码
func (h *DefaultErrorHandler) GetErrorCode(err error) string {
	var kratosErr *kratosErrors.Error
	if errors.As(err, &kratosErr) {
		return fmt.Sprintf("%d", kratosErr.Code)
	}
	return "UNKNOWN_ERROR"
}

// extractErrorCode 从错误中提取错误码
func extractErrorCode(err error) (int, bool) {
	var kratosErr *kratosErrors.Error
	if errors.As(err, &kratosErr) {
		return int(kratosErr.Code), true
	}
	return 0, false
}
