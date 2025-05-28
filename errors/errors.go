// Package errors 提供统一的错误处理机制
package errors

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// 重新导出 github.com/pkg/errors 的基本功能
var (
	New    = errors.New
	Wrap   = errors.Wrap
	Wrapf  = errors.Wrapf
	Cause  = errors.Cause
	WithMessage = errors.WithMessage
	WithMessagef = errors.WithMessagef
	WithStack = errors.WithStack
	As     = errors.As
)

// ErrorType 定义错误类型
type ErrorType uint

const (
	// ErrorTypeUnknown 未知错误
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeValidation 验证错误
	ErrorTypeValidation
	// ErrorTypeDatabase 数据库错误
	ErrorTypeDatabase
	// ErrorTypeNotFound 资源不存在错误
	ErrorTypeNotFound
	// ErrorTypePermission 权限错误
	ErrorTypePermission
	// ErrorTypeConflict 冲突错误
	ErrorTypeConflict
	// ErrorTypeRateLimit 速率限制错误
	ErrorTypeRateLimit
)

// APIError 表示 API 错误
type APIError struct {
	Type    ErrorType // 错误类型
	Code    string    // 错误代码
	Message string    // 错误消息
	Err     error     // 原始错误
	Details []ErrorDetail // 错误详情
}

// ErrorDetail 表示错误的详细信息
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`   // 字段名
	Message string `json:"message,omitempty"` // 错误消息
}

// Error 实现 error 接口
func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap 返回原始错误
func (e *APIError) Unwrap() error {
	return e.Err
}

// StatusCode 返回对应的 HTTP 状态码
func (e *APIError) StatusCode() int {
	switch e.Type {
	case ErrorTypeValidation:
		return http.StatusUnprocessableEntity
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypePermission:
		return http.StatusForbidden
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// AddDetail 添加错误详情
func (e *APIError) AddDetail(field, message string) *APIError {
	e.Details = append(e.Details, ErrorDetail{Field: field, Message: message})
	return e
}

// NewValidationError 创建验证错误
func NewValidationError(message string, err error) *APIError {
	return &APIError{
		Type:    ErrorTypeValidation,
		Code:    "VALIDATION_FAILED",
		Message: message,
		Err:     err,
	}
}

// NewDatabaseError 创建数据库错误
func NewDatabaseError(message string, err error) *APIError {
	return &APIError{
		Type:    ErrorTypeDatabase,
		Code:    "DATABASE_ERROR",
		Message: message,
		Err:     err,
	}
}

// NewNotFoundError 创建资源不存在错误
func NewNotFoundError(message string, err error) *APIError {
	return &APIError{
		Type:    ErrorTypeNotFound,
		Code:    "RESOURCE_NOT_FOUND",
		Message: message,
		Err:     err,
	}
}

// NewPermissionError 创建权限错误
func NewPermissionError(message string, err error) *APIError {
	return &APIError{
		Type:    ErrorTypePermission,
		Code:    "PERMISSION_DENIED",
		Message: message,
		Err:     err,
	}
}

// NewConflictError 创建冲突错误
func NewConflictError(message string, err error) *APIError {
	return &APIError{
		Type:    ErrorTypeConflict,
		Code:    "RESOURCE_CONFLICT",
		Message: message,
		Err:     err,
	}
}

// NewRateLimitError 创建速率限制错误
func NewRateLimitError(message string, err error) *APIError {
	return &APIError{
		Type:    ErrorTypeRateLimit,
		Code:    "RATE_LIMIT_EXCEEDED",
		Message: message,
		Err:     err,
	}
}

// IsValidationError 检查是否为验证错误
func IsValidationError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Type == ErrorTypeValidation
	}
	return false
}

// IsNotFoundError 检查是否为资源不存在错误
func IsNotFoundError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Type == ErrorTypeNotFound
	}
	return false
}
