// Package response 提供统一响应格式中间件
package response

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ErrorEncoderOption 配置错误编码器的可选观测能力。
type ErrorEncoderOption func(*errorEncoderOptions)

type errorEncoderOptions struct {
	logger          *log.Helper
	requestMetadata func(*http.Request) map[string]string
}

// WithErrorLogger 启用统一错误日志记录。
func WithErrorLogger(logger log.Logger) ErrorEncoderOption {
	return func(options *errorEncoderOptions) {
		if logger != nil {
			options.logger = log.NewHelper(logger)
		}
	}
}

// WithErrorRequestMetadata 添加业务服务可选的请求元数据提取器。
func WithErrorRequestMetadata(extractor func(*http.Request) map[string]string) ErrorEncoderOption {
	return func(options *errorEncoderOptions) {
		options.requestMetadata = extractor
	}
}

// NewResponseEncoder 创建响应编码器
// errorHandler: 错误处理接口，如果为 nil，使用默认处理
// config: 配置信息，如果为 nil，不跳过任何路径
func NewResponseEncoder(errorHandler ErrorHandler, config *Config) func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		// 检查是否应该跳过统一响应格式
		if config != nil && config.ShouldSkipPath(r.URL.Path) {
			// 跳过统一响应格式，直接返回原始响应
			// 注意：这里需要确保响应已经被正确设置
			return nil
		}

		// 检查是否已经设置了非 JSON 的 Content-Type（如文件下载）
		if contentType := w.Header().Get("Content-Type"); contentType != "" && contentType != "application/json" {
			// 响应已经被处理（如文件下载），不需要编码
			return nil
		}

		w.Header().Set("Content-Type", "application/json")

		// 如果 v 为 nil（服务返回 nil, nil），返回 data 为 null 的响应
		if v == nil {
			traceId := traceIDFromRequest(r)
			host := r.Host
			response := &ResponseStructure{
				Success:      true,
				Data:         nil,
				ErrorCode:    "",
				ErrorMessage: "",
				ShowType:     ShowTypeSilent,
				TraceId:      traceId,
				Host:         host,
			}
			return json.NewEncoder(w).Encode(response)
		}

		// 如果已经是ResponseStructure格式，更新host信息后序列化
		if resp, ok := v.(*ResponseStructure); ok {
			// 更新host信息为真实的请求主机名
			resp.Host = r.Host

			// 对于protobuf消息，使用protojson序列化以处理零值字段
			if msg, ok := resp.Data.(proto.Message); ok {
				jsonBytes, err := protojson.MarshalOptions{
					EmitUnpopulated: true,  // 包含零值字段
					UseProtoNames:   false, // 使用JSON字段名（驼峰命名）
				}.Marshal(msg)
				if err != nil {
					return err
				}

				// 将序列化后的JSON转换为interface{}
				var jsonData interface{}
				if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
					return err
				}

				// 更新data字段为处理后的JSON数据
				resp.Data = jsonData
			}

			return json.NewEncoder(w).Encode(resp)
		}

		// 如果是protobuf消息，包装为ResponseStructure
		if msg, ok := v.(proto.Message); ok {
			traceId := traceIDFromRequest(r)
			host := r.Host

			// 使用protojson序列化以处理零值字段
			jsonBytes, err := protojson.MarshalOptions{
				EmitUnpopulated: true,  // 包含零值字段
				UseProtoNames:   false, // 使用JSON字段名（驼峰命名）
			}.Marshal(msg)
			if err != nil {
				return err
			}

			// 将序列化后的JSON转换为interface{}
			var jsonData interface{}
			if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
				return err
			}

			response := &ResponseStructure{
				Success:      true,
				Data:         jsonData,
				ErrorCode:    "",
				ErrorMessage: "",
				ShowType:     ShowTypeSilent,
				TraceId:      traceId,
				Host:         host,
			}

			return json.NewEncoder(w).Encode(response)
		}

		// 其他情况，包装为ResponseStructure
		traceId := traceIDFromRequest(r)
		host := r.Host

		response := &ResponseStructure{
			Success:      true,
			Data:         v,
			ErrorCode:    "",
			ErrorMessage: "",
			ShowType:     ShowTypeSilent,
			TraceId:      traceId,
			Host:         host,
		}

		return json.NewEncoder(w).Encode(response)
	}
}

// NewErrorEncoder 创建错误编码器
// errorHandler: 错误处理接口，必须提供
func NewErrorEncoder(errorHandler ErrorHandler, opts ...ErrorEncoderOption) func(http.ResponseWriter, *http.Request, error) {
	if errorHandler == nil {
		panic("ErrorHandler cannot be nil")
	}
	options := &errorEncoderOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}

	return func(w http.ResponseWriter, r *http.Request, err error) {
		traceID := traceIDFromRequest(r)
		w.Header().Set("X-Trace-Id", traceID)
		if options.logger != nil {
			metadata := map[string]string{}
			if options.requestMetadata != nil {
				metadata = options.requestMetadata(r)
			}
			options.logger.Errorf(
				"HTTP request failed: method=%s path=%s trace_id=%s metadata=%s error=%v",
				r.Method,
				r.URL.Path,
				traceID,
				formatRequestMetadata(metadata),
				err,
			)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errorHandler.GetHTTPStatusCode(err))
		host := r.Host

		response := &ResponseStructure{
			Success:      false,
			Data:         nil,
			ErrorCode:    errorHandler.GetErrorCode(err),
			ErrorMessage: errorHandler.GetErrorMessage(err, false),
			ShowType:     errorHandler.GetErrorShowType(err),
			TraceId:      traceID,
			Host:         host,
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}

func formatRequestMetadata(metadata map[string]string) string {
	if len(metadata) == 0 {
		return "-"
	}
	keys := make([]string, 0, len(metadata))
	for key := range metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		value := strings.ReplaceAll(metadata[key], " ", "_")
		parts = append(parts, key+"="+value)
	}
	return strings.Join(parts, ",")
}

func traceIDFromRequest(request *http.Request) string {
	if traceID := GetTraceIdFromContext(request.Context()); traceID != "" {
		return traceID
	}
	if traceID := request.Header.Get("X-Trace-Id"); traceID != "" {
		return traceID
	}
	return GenerateUUID()
}
