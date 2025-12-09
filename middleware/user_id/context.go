// Package user_id 提供终端用户 ID 中间件和工具函数
package user_id

import (
	"context"
)

// userIDKey 是 context 中存储终端用户 ID 信息的键
type userIDKey struct{}

// UserIDKey 导出终端用户 ID 键，供外部使用
var UserIDKey = userIDKey{}

// GetUserIDFromContext 从 Context 获取终端用户 ID
// 如果 context 中没有终端用户 ID 信息，返回空字符串
func GetUserIDFromContext(ctx context.Context) string {
	if userID := ctx.Value(UserIDKey); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// WithUserID 将终端用户 ID 存入 context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
