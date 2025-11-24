// Package response 提供统一响应格式中间件
package response

// Config 响应格式配置
// 注意：不包含业务相关的默认值，需要在项目中自定义
type Config struct {
	// 是否启用统一响应格式
	EnableUnifiedResponse bool `json:"enable_unified_response" yaml:"enable_unified_response"`

	// 跳过统一响应格式的路径（支持通配符）
	// 注意：需要在项目中自定义，不提供默认值
	SkipPaths []string `json:"skip_paths" yaml:"skip_paths"`

	// 是否包含详细的错误信息
	IncludeDetailedError bool `json:"include_detailed_error" yaml:"include_detailed_error"`

	// 是否在响应中包含主机信息（请求的主机名，如 localhost:8080）
	IncludeHost bool `json:"include_host" yaml:"include_host"`

	// 是否在响应中包含 TraceId
	IncludeTraceId bool `json:"include_trace_id" yaml:"include_trace_id"`

	// 自定义 TraceId 头部名称
	TraceIdHeader string `json:"trace_id_header" yaml:"trace_id_header"`
}

// ShouldSkipPath 判断是否应该跳过某个路径
func (c *Config) ShouldSkipPath(path string) bool {
	if !c.EnableUnifiedResponse {
		return true
	}

	for _, skipPath := range c.SkipPaths {
		if MatchPath(path, skipPath) {
			return true
		}
	}

	return false
}

