// Package errors 提供错误包装和创建功能
package errors

import (
	"context"
	"sync"

	kratosErrors "github.com/go-kratos/kratos/v2/errors"
)

// ErrorManager 错误管理器，提供错误创建和包装功能
type ErrorManager struct {
	messageLoader ErrorMessageLoader
	langGetter    func(context.Context) string // 从 context 获取语言的函数
}

// NewErrorManager 创建错误管理器
// messageLoader: 错误消息加载器
// langGetter: 从 context 获取语言的函数，如果为 nil，使用默认实现
func NewErrorManager(messageLoader ErrorMessageLoader, langGetter func(context.Context) string) *ErrorManager {
	if langGetter == nil {
		langGetter = defaultLangGetter
	}
	return &ErrorManager{
		messageLoader: messageLoader,
		langGetter:    langGetter,
	}
}

// defaultLangGetter 默认语言获取函数（返回 zh-CN）
func defaultLangGetter(ctx context.Context) string {
	return "zh-CN"
}

// NewBizError 创建业务错误，支持错误码和语言
// code: 错误码
// lang: 语言，如果为空，使用默认语言 "zh-CN"
func (m *ErrorManager) NewBizError(code int32, lang string) *kratosErrors.Error {
	if lang == "" {
		lang = "zh-CN"
	}
	message := m.messageLoader.GetMessage(lang, code)
	return kratosErrors.New(int(code), "BIZ_ERROR", message)
}

// NewBizErrorWithLang 从 context 中获取语言并创建业务错误
func (m *ErrorManager) NewBizErrorWithLang(ctx context.Context, code int32) *kratosErrors.Error {
	lang := m.langGetter(ctx)
	return m.NewBizError(code, lang)
}

// WrapError 包装错误为业务错误
// err: 原始错误
// code: 错误码
// lang: 语言，如果为空，使用默认语言 "zh-CN"
func (m *ErrorManager) WrapError(err error, code int32, lang string) *kratosErrors.Error {
	if err == nil {
		return nil
	}
	if lang == "" {
		lang = "zh-CN"
	}
	message := m.messageLoader.GetMessage(lang, code)
	return kratosErrors.New(int(code), "BIZ_ERROR", message)
}

// WrapErrorWithLang 从 context 中获取语言并包装错误
func (m *ErrorManager) WrapErrorWithLang(ctx context.Context, err error, code int32) *kratosErrors.Error {
	if err == nil {
		return nil
	}
	lang := m.langGetter(ctx)
	return m.WrapError(err, code, lang)
}

// GetErrorMessage 获取错误消息（便捷方法）
func (m *ErrorManager) GetErrorMessage(lang string, code int32) string {
	return m.messageLoader.GetMessage(lang, code)
}

// 全局错误管理器（用于便捷函数）
var (
	globalErrorManager     *ErrorManager
	globalErrorManagerOnce sync.Once
)

// InitGlobalErrorManager 初始化全局错误管理器
// configDir: i18n 配置目录，例如 "i18n"
// langGetter: 从 context 获取语言的函数，如果为 nil，使用默认实现
func InitGlobalErrorManager(configDir string, langGetter func(context.Context) string) {
	globalErrorManagerOnce.Do(func() {
		globalErrorManager = NewErrorManager(
			NewJSONErrorMessageLoader(configDir),
			langGetter,
		)
	})
}

// NewBizError 创建业务错误（使用全局错误管理器）
// 需要先调用 InitGlobalErrorManager 初始化
func NewBizError(code int32, lang string) *kratosErrors.Error {
	if globalErrorManager == nil {
		panic("global error manager not initialized, call InitGlobalErrorManager first")
	}
	return globalErrorManager.NewBizError(code, lang)
}

// NewBizErrorWithLang 从 context 中获取语言并创建业务错误（使用全局错误管理器）
func NewBizErrorWithLang(ctx context.Context, code int32) *kratosErrors.Error {
	if globalErrorManager == nil {
		panic("global error manager not initialized, call InitGlobalErrorManager first")
	}
	return globalErrorManager.NewBizErrorWithLang(ctx, code)
}

// WrapError 包装错误为业务错误（使用全局错误管理器）
func WrapError(err error, code int32, lang string) *kratosErrors.Error {
	if globalErrorManager == nil {
		panic("global error manager not initialized, call InitGlobalErrorManager first")
	}
	return globalErrorManager.WrapError(err, code, lang)
}

// WrapErrorWithLang 从 context 中获取语言并包装错误（使用全局错误管理器）
func WrapErrorWithLang(ctx context.Context, err error, code int32) *kratosErrors.Error {
	if globalErrorManager == nil {
		panic("global error manager not initialized, call InitGlobalErrorManager first")
	}
	return globalErrorManager.WrapErrorWithLang(ctx, err, code)
}

// GetErrorMessage 获取错误消息（使用全局错误管理器）
func GetErrorMessage(lang string, code int32) string {
	if globalErrorManager == nil {
		panic("global error manager not initialized, call InitGlobalErrorManager first")
	}
	return globalErrorManager.GetErrorMessage(lang, code)
}
