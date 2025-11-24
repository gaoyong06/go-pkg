// Package response 提供统一响应格式中间件
package response

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
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
				logHelper.Debug("无法获取传输信息")
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
				// 使用 ResponseEncoder 方式时，主机信息将在 HTTP 层处理
				// 这里暂时使用默认值
				host = "api-server" // 可配置的服务标识
			}

			// 如果有错误，统一处理错误响应
			if err != nil {
				errorResponse := &ResponseStructure{
					Success:      false,
					Data:         nil,
					ErrorCode:    errorHandler.GetErrorCode(err),
					ErrorMessage: errorHandler.GetErrorMessage(err, config.IncludeDetailedError),
					ShowType:     errorHandler.GetErrorShowType(err),
					TraceId:      traceId,
					Host:         host,
				}

				logHelper.Errorf("API错误: %v, TraceId: %s", err, traceId)
				return errorResponse, nil // 返回nil错误，让框架正常处理响应
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

