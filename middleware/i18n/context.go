// Package i18n 提供国际化（i18n）中间件和工具函数
package i18n

import (
	"context"
)

// langKey 是 context 中存储语言信息的键
type langKey struct{}

// LanguageKey 导出语言键，供外部使用
var LanguageKey = langKey{}

// Language 从 context 中获取规范化语言标签。
// 未注入请求语言时保留旧默认值，保证尚未迁移的服务行为不变。
func Language(ctx context.Context) string {
	if ctx == nil {
		return defaultResolver.DefaultLanguage()
	}
	if lang, ok := ctx.Value(LanguageKey).(string); ok && lang != "" {
		if normalized := Normalize(lang); normalized != "" {
			return normalized
		}
	}
	return defaultResolver.DefaultLanguage()
}

// LanguageWithDefault 从 context 获取规范化语言；未设置有效请求语言时使用 fallback。
// 服务可以据此声明自己的默认语言，同时保留 Language 的兼容默认值。
func LanguageWithDefault(ctx context.Context, fallback string) string {
	if ctx != nil {
		if lang, ok := ctx.Value(LanguageKey).(string); ok && lang != "" {
			if normalized := Normalize(lang); normalized != "" {
				return normalized
			}
		}
	}
	if normalized := Normalize(fallback); normalized != "" {
		return normalized
	}
	return defaultResolver.DefaultLanguage()
}

// WithLanguage 将语言标签规范化后存入 context。
func WithLanguage(ctx context.Context, lang string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if normalized := Normalize(lang); normalized != "" {
		lang = normalized
	} else {
		lang = defaultResolver.DefaultLanguage()
	}
	return context.WithValue(ctx, LanguageKey, lang)
}

var defaultResolver = MustNewResolver(ResolverConfig{
	DefaultLanguage:    "zh-CN",
	SupportedLanguages: []string{"zh-CN", "en-US"},
})
