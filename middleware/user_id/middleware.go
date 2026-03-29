// Package user_id 提供终端用户 ID 中间件和工具函数
package user_id

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/grpc/metadata"
)

// Middleware 终端用户 ID 中间件，提取终端用户 ID 并存入 context
// 终端用户 ID 提取优先级：
// 1. HTTP Header X-End-User-Id（由 API Gateway 的 jwt-user 插件设置）
// 2. gRPC metadata X-End-User-Id（服务间调用时传递）
func Middleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			userID := extractUserID(ctx)
			if userID != "" {
				ctx = WithUserID(ctx, userID)
				// 调试日志：记录成功提取的 userID
				log.NewHelper(log.GetLogger()).Infof("user_id middleware: extracted userId=%s", userID)
			} else {
				// 调试日志：记录未找到 userID 的情况
				log.NewHelper(log.GetLogger()).Infof("user_id middleware: no userId found in headers or metadata")
			}
			return handler(ctx, req)
		}
	}
}

// extractUserID 从请求中提取终端用户 ID
// 优先级：
// 1. HTTP Header X-End-User-Id（由 API Gateway 的 jwt-user 插件统一设置）
// 2. gRPC metadata X-End-User-Id（服务间调用时传递）
func extractUserID(ctx context.Context) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		// 如果没有 transport，尝试从 gRPC metadata 提取
		return extractUserIDFromGRPCMetadata(ctx)
	}

	// 1. 从 HTTP Header 提取 X-End-User-Id（标准名称，由 API Gateway 的 jwt-user 插件统一设置）
	// 直接使用 transport 的 RequestHeader().Get() 方法（参考 app_id 中间件的实现）
	header := tr.RequestHeader()
	if userID := header.Get("X-End-User-Id"); userID != "" {
		return strings.TrimSpace(userID)
	}
	// 如果标准名称不存在，尝试其他可能的名称
	if userID := header.Get("x-end-user-id"); userID != "" {
		return strings.TrimSpace(userID)
	}

	// 2. 从 gRPC metadata 提取 X-End-User-Id（服务间调用时传递）
	return extractUserIDFromGRPCMetadata(ctx)
}

// extractUserIDFromGRPCMetadata 从 gRPC metadata 中提取终端用户 ID
// 只支持 X-End-User-Id 一个标准名称（gRPC metadata 的 key 会被转换为小写）
func extractUserIDFromGRPCMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	// gRPC metadata 的 key 会被转换为小写，所以使用 "x-end-user-id"
	values := md.Get("x-end-user-id")
	if len(values) > 0 && values[0] != "" {
		return strings.TrimSpace(values[0])
	}

	return ""
}
