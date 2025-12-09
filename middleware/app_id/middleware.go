// Package app_id 提供 appId 中间件和工具函数
package app_id

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/grpc/metadata"
)

// Middleware appId 中间件，提取 appId 并存入 context
// appId 提取优先级：
// 1. HTTP Header X-App-Id（由 API Gateway 设置）
// 2. gRPC metadata X-App-Id（服务间调用时传递）
func Middleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			appID := extractAppID(ctx)
			if appID != "" {
				ctx = WithAppID(ctx, appID)
			}
			return handler(ctx, req)
		}
	}
}

// extractAppID 从请求中提取 appId
// 优先级：
// 1. HTTP Header X-App-Id（由 API Gateway 统一设置）
// 2. gRPC metadata X-App-Id（服务间调用时传递）
func extractAppID(ctx context.Context) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		// 如果没有 transport，尝试从 gRPC metadata 提取
		return extractAppIDFromGRPCMetadata(ctx)
	}

	// 1. 从 HTTP Header 提取 X-App-Id（标准名称，由 API Gateway 统一设置）
	if httpTr, ok := tr.(interface {
		RequestHeader() map[string][]string
	}); ok {
		headers := httpTr.RequestHeader()
		if values, ok := headers["X-App-Id"]; ok && len(values) > 0 && values[0] != "" {
			return values[0]
		}
	}

	// 2. 从 gRPC metadata 提取 X-App-Id（服务间调用时传递）
	return extractAppIDFromGRPCMetadata(ctx)
}

// extractAppIDFromGRPCMetadata 从 gRPC metadata 中提取 appId
// 只支持 X-App-Id 一个标准名称（gRPC metadata 的 key 会被转换为小写）
func extractAppIDFromGRPCMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	// gRPC metadata 的 key 会被转换为小写，所以使用 "x-app-id"
	values := md.Get("x-app-id")
	if len(values) > 0 && values[0] != "" {
		return strings.TrimSpace(values[0])
	}

	return ""
}
