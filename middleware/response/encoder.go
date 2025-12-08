// Package response 提供统一响应格式中间件
package response

import (
	"encoding/json"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

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
			traceId := GenerateUUID()
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
			traceId := GenerateUUID()
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
		traceId := GenerateUUID()
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
func NewErrorEncoder(errorHandler ErrorHandler) func(http.ResponseWriter, *http.Request, error) {
	if errorHandler == nil {
		panic("ErrorHandler cannot be nil")
	}

	return func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")

		// 获取HTTP状态码
		statusCode := errorHandler.GetHTTPStatusCode(err)
		w.WriteHeader(statusCode)

		// 生成错误响应
		traceId := GenerateUUID()
		host := r.Host

		response := &ResponseStructure{
			Success:      false,
			Data:         nil,
			ErrorCode:    errorHandler.GetErrorCode(err),
			ErrorMessage: errorHandler.GetErrorMessage(err, false),
			ShowType:     errorHandler.GetErrorShowType(err),
			TraceId:      traceId,
			Host:         host,
		}

		json.NewEncoder(w).Encode(response)
	}
}
