// Package auth 提供认证中间件和工具函数
package auth

import "strings"

// Config 认证中间件配置
type Config struct {
	// 跳过认证的路径（支持通配符）
	// 例如：["/health", "/swagger/*", "/v1/public/*"]
	SkipPaths []string `json:"skip_paths" yaml:"skip_paths"`
}

// ShouldSkipPath 判断是否应该跳过某个路径
func (c *Config) ShouldSkipPath(path string) bool {
	if c == nil {
		return false
	}

	for _, skipPath := range c.SkipPaths {
		if MatchPath(path, skipPath) {
			return true
		}
	}

	return false
}

// MatchPath 匹配路径（支持简单的通配符）
// 支持格式：
// - "/path" - 精确匹配
// - "/path/*" - 前缀匹配
// - "*/suffix" - 后缀匹配
func MatchPath(path, pattern string) bool {
	// 简单的通配符匹配实现
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(path, suffix)
	}

	return path == pattern
}

