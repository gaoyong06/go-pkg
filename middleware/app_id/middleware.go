// Package app_id 提供 appId 中间件和工具函数
package app_id

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
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
				// 调试日志：记录成功提取的 appId
				log.NewHelper(log.GetLogger()).Debugf("app_id middleware: extracted appId=%s", appID)
			} else {
				// 调试日志：记录未找到 appId 的情况
				log.NewHelper(log.GetLogger()).Debugf("app_id middleware: no appId found in headers or metadata")
			}
			return handler(ctx, req)
		}
	}
}

// extractAppID 从请求中提取 appId
// 优先级：
// 1. HTTP Header X-App-Id（由 API Gateway 统一设置）
// 2. Query String appId（前端通过 Query String 传递，作为后备方案）
// 3. gRPC metadata X-App-Id（服务间调用时传递）
func extractAppID(ctx context.Context) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		// 如果没有 transport，尝试从 gRPC metadata 提取
		return extractAppIDFromGRPCMetadata(ctx)
	}

	// 1. 从 HTTP Header 提取 X-App-Id（标准名称，由 API Gateway 统一设置）
	// 直接使用 transport 的 RequestHeader().Get() 方法（参考其他中间件的实现）
	header := tr.RequestHeader()
	if appID := header.Get("X-App-Id"); appID != "" {
		return strings.TrimSpace(appID)
	}
	// 如果标准名称不存在，尝试其他可能的名称
	if appID := header.Get("x-app-id"); appID != "" {
		return strings.TrimSpace(appID)
	}

	// 调试：如果 Get() 方法失败，尝试使用 map 方式（用于调试）
	if httpTr, ok := tr.(interface {
		RequestHeader() map[string][]string
	}); ok {
		headers := httpTr.RequestHeader()
		// 调试：打印所有 Header 键名和值
		headerKeys := make([]string, 0, len(headers))
		headerValues := make(map[string]string)
		for key, values := range headers {
			headerKeys = append(headerKeys, key)
			if len(values) > 0 {
				headerValues[key] = values[0]
			}
		}
		log.NewHelper(log.GetLogger()).Infof("app_id middleware: available headers: %v", headerKeys)
		log.NewHelper(log.GetLogger()).Infof("app_id middleware: header values: %v", headerValues)
		// 遍历所有 Header 查找（处理大小写不匹配）
		for key, values := range headers {
			if strings.EqualFold(key, "X-App-Id") && len(values) > 0 && values[0] != "" {
				log.NewHelper(log.GetLogger()).Infof("app_id middleware: found X-App-Id in header %s=%s", key, values[0])
				return strings.TrimSpace(values[0])
			}
		}
	}

	// 2. 从 Query String 提取 appId（作为后备方案，用于直接访问服务的情况）
	if httpTr, ok := tr.(*kratoshttp.Transport); ok {
		if req := httpTr.Request(); req != nil {
			if reqURL := req.URL; reqURL != nil {
				if appID := reqURL.Query().Get("appId"); appID != "" {
					log.NewHelper(log.GetLogger()).Infof("app_id middleware: found appId in query string: %s", appID)
					return strings.TrimSpace(appID)
				}
			}
		}
	}

	// 3. 从 gRPC metadata 提取 X-App-Id（服务间调用时传递）
	return extractAppIDFromGRPCMetadata(ctx)
}

// extractAppIDFromGRPCMetadata 从 gRPC metadata 中提取 appId
// 只支持 X-App-Id 一个标准名称（gRPC metadata 的 key 会被转换为小写）
func extractAppIDFromGRPCMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	// Debug: print all metadata keys
	log.NewHelper(log.GetLogger()).Infof("app_id middleware: gRPC metadata keys: %v", md)

	// gRPC metadata 的 key 会被转换为小写，所以使用 "x-app-id"
	values := md.Get("x-app-id")
	if len(values) > 0 && values[0] != "" {
		return strings.TrimSpace(values[0])
	}

	return ""
}
