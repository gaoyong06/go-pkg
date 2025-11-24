// Package response 提供统一响应格式中间件
package response

import (
	"context"
	"strings"
	"time"
)

// MatchPath 匹配路径（支持简单的通配符）
// 支持格式：
// - "/path" - 精确匹配
// - "/path/*" - 前缀匹配
// - "*/suffix" - 后缀匹配
func MatchPath(path, pattern string) bool {
	// 简单的通配符匹配实现
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(path, suffix)
	}

	return path == pattern
}

// GetTraceIdFromContext 从上下文获取 TraceId
func GetTraceIdFromContext(ctx context.Context) string {
	if traceId := ctx.Value("trace_id"); traceId != nil {
		if id, ok := traceId.(string); ok {
			return id
		}
	}

	if traceId := ctx.Value("X-Trace-Id"); traceId != nil {
		if id, ok := traceId.(string); ok {
			return id
		}
	}

	return ""
}

// SetTraceIdToContext 设置 TraceId 到上下文
func SetTraceIdToContext(ctx context.Context, traceId string) context.Context {
	ctx = context.WithValue(ctx, "trace_id", traceId)
	ctx = context.WithValue(ctx, "X-Trace-Id", traceId)
	return ctx
}

// GenerateUUID 生成简单的UUID（用于 TraceId）
func GenerateUUID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

