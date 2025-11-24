// Package errors 提供通用错误码定义
package errors

// 通用错误码 (1000-1099)
// 这些错误码适用于所有项目，定义在公共库中
const (
	// ErrCodeInvalidArgument 无效参数错误
	ErrCodeInvalidArgument = 1001
	// ErrCodeInternalError 内部错误
	ErrCodeInternalError = 1002
	// ErrCodeUnauthorized 未授权错误
	ErrCodeUnauthorized = 1003
	// ErrCodeForbidden 禁止访问错误
	ErrCodeForbidden = 1004
)

