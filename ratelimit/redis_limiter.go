package ratelimit

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// luaScript Lua 脚本，用于原子性地检查所有时间窗口
// 优化：将多次 Redis 调用合并为一次
const luaScript = `
-- 检查单个时间窗口
local function check_window(key_suffix, limit, window_seconds, now_unix)
    if limit <= 0 then
        return {true, 0}
    end
    
    local window_start = now_unix - window_seconds
    local zset_key = "rate_limit:" .. KEYS[1] .. ":" .. key_suffix
    
    -- 删除过期数据
    redis.call('ZREMRANGEBYSCORE', zset_key, '0', tostring(window_start * 1000000))
    
    -- 获取当前计数
    local count = redis.call('ZCARD', zset_key)
    
    if count >= limit then
        return {false, count}
    end
    
    -- 添加当前请求
    local member = tostring(now_unix * 1000000 + math.random(0, 999999))
    redis.call('ZADD', zset_key, member, member)
    redis.call('EXPIRE', zset_key, window_seconds + 1)
    
    return {true, count + 1}
end

local now_unix = tonumber(ARGV[1])
local per_second = tonumber(ARGV[2])
local per_minute = tonumber(ARGV[3])
local per_hour = tonumber(ARGV[4])
local per_day = tonumber(ARGV[5])

-- 检查所有窗口
local windows = {
    {"per_second", per_second, 1},
    {"per_minute", per_minute, 60},
    {"per_hour", per_hour, 3600},
    {"per_day", per_day, 86400}
}

for i, window_config in ipairs(windows) do
    local suffix = window_config[1]
    local limit = window_config[2]
    local window = window_config[3]
    
    local result = check_window(suffix, limit, window, now_unix)
    if not result[1] then
        -- 返回: [失败标志, 窗口名称, 窗口秒数, 当前计数, 限制]
        return {0, suffix, window, result[2], limit}
    end
end

-- 返回: [成功标志, 空, 0, 0, 0]
return {1, "", 0, 0, 0}
`

// RedisLimiter 基于 Redis 的限流器
type RedisLimiter struct {
	rdb    *redis.Client
	script *redis.Script
}

// NewRedisLimiter 创建 Redis 限流器
func NewRedisLimiter(rdb *redis.Client) Limiter {
	return &RedisLimiter{
		rdb:    rdb,
		script: redis.NewScript(luaScript),
	}
}

// Allow 检查是否允许请求通过
func (r *RedisLimiter) Allow(ctx context.Context, key string, config *Config) error {
	if config == nil {
		return nil // 没有配置限流，允许通过
	}

	now := getNowUnix()

	// 执行 Lua 脚本
	result, err := r.script.Run(ctx, r.rdb, []string{key},
		now,
		config.PerSecond,
		config.PerMinute,
		config.PerHour,
		config.PerDay,
	).Result()

	if err != nil {
		return fmt.Errorf("rate limit check failed: %w", err)
	}

	// 解析结果
	res, ok := result.([]interface{})
	if !ok || len(res) < 5 {
		return fmt.Errorf("invalid lua script result")
	}

	success, ok := res[0].(int64)
	if !ok {
		return fmt.Errorf("invalid success flag in lua result")
	}

	if success == 0 {
		// 触发限流
		windowName := res[1].(string)
		windowSeconds := res[2].(int64)
		current := res[3].(int64)
		limit := res[4].(int64)

		return &RateLimitError{
			Key:           key,
			WindowName:    windowName,
			WindowSeconds: windowSeconds,
			Current:       current,
			Limit:         limit,
		}
	}

	return nil
}

// RateLimitError 限流错误
type RateLimitError struct {
	Key           string
	WindowName    string
	WindowSeconds int64
	Current       int64
	Limit         int64
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: key=%s, window=%s(%ds), current=%d, limit=%d",
		e.Key, e.WindowName, e.WindowSeconds, e.Current, e.Limit)
}

// IsRateLimitError 判断是否是限流错误
func IsRateLimitError(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}
