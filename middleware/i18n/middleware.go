// Package i18n 提供国际化（i18n）中间件和工具函数
package i18n

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
)

// Middleware i18n 中间件，使用兼容默认 resolver 提取语言并存入 context。
func Middleware() middleware.Middleware {
	return MiddlewareWithResolver(defaultResolver)
}

// MiddlewareWithResolver 创建使用显式语言配置的 i18n 中间件。
func MiddlewareWithResolver(resolver *Resolver) middleware.Middleware {
	if resolver == nil {
		resolver = defaultResolver
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			lang := extractLanguage(ctx, resolver)
			ctx = WithLanguage(ctx, lang)
			return handler(ctx, req)
		}
	}
}
