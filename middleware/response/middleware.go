// Package response 提供统一响应格式中间件
package response

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// Middleware 统一响应格式中间件
// config: 配置信息
// errorHandler: 错误处理接口
// logger: 日志记录器
func Middleware(config *Config, errorHandler ErrorHandler, logger log.Logger) middleware.Middleware {
	logHelper := log.NewHelper(logger)

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			// 获取传输信息
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				logHelper.Debug("Failed to get transport from context")
			}

			// 检查是否应该跳过统一响应格式
			if tr != nil {
				operation := tr.Operation()
				if config.ShouldSkipPath(operation) {
					// 跳过统一响应格式，直接返回原始响应
					return handler(ctx, req)
				}
			}

			// 执行业务逻辑
			reply, err = handler(ctx, req)

			// 生成或获取 trace ID
			traceId := ""
			if config.IncludeTraceId {
				traceId = GetTraceIdFromContext(ctx)
				if traceId == "" {
					traceId = GenerateUUID()
					ctx = SetTraceIdToContext(ctx, traceId)
				}
			}

			// 获取主机信息
			host := ""
			if config.IncludeHost {
				if ht, ok := tr.(*kratoshttp.Transport); ok && ht.Request() != nil {
					host = ht.Request().Host
				}
			}

			// 如果有错误，统一处理错误响应
			if err != nil {
				logHelper.Errorf("API error: %v, TraceId: %s", err, traceId)
				return nil, err
			}

			// 成功响应的统一格式
			successResponse := &ResponseStructure{
				Success:      true,
				Data:         reply,
				ErrorCode:    "",
				ErrorMessage: "",
				ShowType:     ShowTypeSilent,
				TraceId:      traceId,
				Host:         host,
			}

			return successResponse, nil
		}
	}
}
