// Package i18n 提供国际化（i18n）中间件和工具函数
package i18n

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
)

// Middleware i18n 中间件，提取语言并存入 context
// 语言提取优先级：
// 1. URL 路径（如 /zh/xxx 或 /en/xxx）
// 2. HTTP Header Accept-Language
// 3. 默认语言 zh-CN
func Middleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			lang := extractLanguage(ctx)
			ctx = WithLanguage(ctx, lang)
			return handler(ctx, req)
		}
	}
}

