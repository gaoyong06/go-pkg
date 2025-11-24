// Package i18n 提供国际化（i18n）中间件和工具函数
package i18n

import (
	"context"
)

// langKey 是 context 中存储语言信息的键
type langKey struct{}

// LanguageKey 导出语言键，供外部使用
var LanguageKey = langKey{}

// Language 从 context 中获取语言
// 如果 context 中没有语言信息，返回默认语言 "zh-CN"
func Language(ctx context.Context) string {
	if lang, ok := ctx.Value(LanguageKey).(string); ok && lang != "" {
		return lang
	}
	return "zh-CN" // 默认语言
}

// WithLanguage 将语言存入 context
func WithLanguage(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, LanguageKey, lang)
}

