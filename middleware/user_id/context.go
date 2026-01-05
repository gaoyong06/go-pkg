// Package user_id 提供终端用户 ID 中间件和工具函数
package user_id

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/grpc/metadata"
)

// userIDKey 是 context 中存储终端用户 ID 信息的键
type userIDKey struct{}

// UserIDKey 导出终端用户 ID 键，供外部使用
var UserIDKey = userIDKey{}

// GetUserIDFromContext 从 Context 获取终端用户 ID
// 优先级：
// 1. 从 context 中获取（由中间件设置，标准方式）
// 2. 如果 context 中没有，尝试从 HTTP Header 获取（兜底方案）
// 3. 如果 HTTP Header 也没有，尝试从 gRPC metadata 获取（兜底方案）
// 这样即使中间件没有运行，也能从 header/metadata 中获取 userID
func GetUserIDFromContext(ctx context.Context) string {
	// 1. 优先从 context 中获取（由中间件设置）
	if userID := ctx.Value(UserIDKey); userID != nil {
		if id, ok := userID.(string); ok && id != "" {
			return id
		}
	}

	// 2. 兜底方案：从 HTTP Header 获取
	if tr, ok := transport.FromServerContext(ctx); ok {
		header := tr.RequestHeader()
		if userID := header.Get("X-End-User-Id"); userID != "" {
			return strings.TrimSpace(userID)
		}
		// 尝试小写版本（某些情况下 header 可能被转换为小写）
		if userID := header.Get("x-end-user-id"); userID != "" {
			return strings.TrimSpace(userID)
		}
	}

	// 3. 兜底方案：从 gRPC metadata 获取
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("x-end-user-id")
		if len(values) > 0 && values[0] != "" {
			return strings.TrimSpace(values[0])
		}
	}

	return ""
}

// WithUserID 将终端用户 ID 存入 context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
