// Package i18n 提供国际化（i18n）中间件和工具函数
package i18n

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"
)

// extractLanguage 从请求中提取语言。
func extractLanguage(ctx context.Context, resolver *Resolver) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return resolver.DefaultLanguage()
	}

	// 1. 从 URL 路径提取（如 /zh/xxx 或 /en/xxx）
	operation := tr.Operation()
	if strings.HasPrefix(operation, "/zh") {
		return resolver.Resolve("zh-CN")
	} else if strings.HasPrefix(operation, "/en") {
		return resolver.Resolve("en-US")
	}

	// 2. 从 HTTP Header 提取
	if acceptLang := strings.TrimSpace(tr.RequestHeader().Get("Accept-Language")); acceptLang != "" {
		return resolver.Resolve(acceptLang)
	}

	return resolver.DefaultLanguage()
}
