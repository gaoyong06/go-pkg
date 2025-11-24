// Package i18n 提供国际化（i18n）中间件和工具函数
package i18n

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"
)

// extractLanguage 从请求中提取语言
func extractLanguage(ctx context.Context) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return "zh-CN"
	}

	// 1. 从 URL 路径提取（如 /zh/xxx 或 /en/xxx）
	operation := tr.Operation()
	if strings.HasPrefix(operation, "/zh") {
		return "zh-CN"
	} else if strings.HasPrefix(operation, "/en") {
		return "en-US"
	}

	// 2. 从 HTTP Header 提取
	if httpTr, ok := tr.(interface {
		RequestHeader() map[string][]string
	}); ok {
		headers := httpTr.RequestHeader()
		if acceptLang, ok := headers["Accept-Language"]; ok && len(acceptLang) > 0 {
			lang := parseAcceptLanguage(acceptLang[0])
			if lang != "" {
				return lang
			}
		}
	}

	return "zh-CN" // 默认语言
}

// parseAcceptLanguage 解析 Accept-Language header
// 支持格式：zh-CN,zh;q=0.9,en;q=0.8
func parseAcceptLanguage(acceptLang string) string {
	parts := strings.Split(acceptLang, ",")
	if len(parts) > 0 {
		lang := strings.TrimSpace(strings.Split(parts[0], ";")[0])
		if strings.HasPrefix(lang, "zh") {
			return "zh-CN"
		} else if strings.HasPrefix(lang, "en") {
			return "en-US"
		}
	}
	return ""
}

