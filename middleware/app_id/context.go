// Package app_id 提供 appId 中间件和工具函数
package app_id

import (
	"context"
)

// appIDKey 是 context 中存储 appId 信息的键
type appIDKey struct{}

// AppIDKey 导出 appId 键，供外部使用
var AppIDKey = appIDKey{}

// GetAppIDFromContext 从 Context 获取 appId
// 如果 context 中没有 appId 信息，返回空字符串
func GetAppIDFromContext(ctx context.Context) string {
	if appID := ctx.Value(AppIDKey); appID != nil {
		if id, ok := appID.(string); ok {
			return id
		}
	}
	return ""
}

// WithAppID 将 appId 存入 context
func WithAppID(ctx context.Context, appID string) context.Context {
	return context.WithValue(ctx, AppIDKey, appID)
}
