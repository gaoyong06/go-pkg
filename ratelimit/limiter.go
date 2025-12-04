package ratelimit

import "context"

// Limiter 限流器接口
type Limiter interface {
	// Allow 检查是否允许请求通过
	// key: 限流键（如 "sms:aliyun:user:123"）
	// config: 限流配置
	// 返回 error 表示触发限流
	Allow(ctx context.Context, key string, config *Config) error
}

// Config 限流配置
type Config struct {
	PerSecond int32 // 每秒限制次数，0 表示不限制
	PerMinute int32 // 每分钟限制次数，0 表示不限制
	PerHour   int32 // 每小时限制次数，0 表示不限制
	PerDay    int32 // 每天限制次数，0 表示不限制
}
