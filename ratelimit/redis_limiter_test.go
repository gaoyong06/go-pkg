package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	// 使用 miniredis 模拟 Redis
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return client, cleanup
}

func TestRedisLimiter_PerSecond(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	limiter := NewRedisLimiter(rdb)
	ctx := context.Background()

	config := &Config{
		PerSecond: 3,
	}

	// 前 3 次应该成功
	for i := 0; i < 3; i++ {
		err := limiter.Allow(ctx, "test:user:123", config)
		assert.NoError(t, err, "request %d should be allowed", i+1)
	}

	// 第 4 次应该被限流
	err := limiter.Allow(ctx, "test:user:123", config)
	assert.Error(t, err)
	assert.True(t, IsRateLimitError(err))

	rateLimitErr, ok := err.(*RateLimitError)
	require.True(t, ok)
	assert.Equal(t, "per_second", rateLimitErr.WindowName)
	assert.Equal(t, int64(3), rateLimitErr.Current)
	assert.Equal(t, int64(3), rateLimitErr.Limit)
}

func TestRedisLimiter_PerMinute(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	limiter := NewRedisLimiter(rdb)
	ctx := context.Background()

	config := &Config{
		PerMinute: 5,
	}

	// 前 5 次应该成功
	for i := 0; i < 5; i++ {
		err := limiter.Allow(ctx, "test:user:456", config)
		assert.NoError(t, err, "request %d should be allowed", i+1)
	}

	// 第 6 次应该被限流
	err := limiter.Allow(ctx, "test:user:456", config)
	assert.Error(t, err)
	assert.True(t, IsRateLimitError(err))
}

func TestRedisLimiter_MultipleWindows(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	limiter := NewRedisLimiter(rdb)
	ctx := context.Background()

	config := &Config{
		PerSecond: 2,
		PerMinute: 5,
	}

	// 前 2 次应该成功（受 per_second 限制）
	for i := 0; i < 2; i++ {
		err := limiter.Allow(ctx, "test:user:789", config)
		assert.NoError(t, err)
	}

	// 第 3 次应该被 per_second 限流
	err := limiter.Allow(ctx, "test:user:789", config)
	assert.Error(t, err)
	rateLimitErr, ok := err.(*RateLimitError)
	require.True(t, ok)
	assert.Equal(t, "per_second", rateLimitErr.WindowName)
}

func TestRedisLimiter_DifferentKeys(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	limiter := NewRedisLimiter(rdb)
	ctx := context.Background()

	config := &Config{
		PerSecond: 2,
	}

	// user1 的前 2 次应该成功
	for i := 0; i < 2; i++ {
		err := limiter.Allow(ctx, "test:user:1", config)
		assert.NoError(t, err)
	}

	// user2 的前 2 次也应该成功（不同的 key）
	for i := 0; i < 2; i++ {
		err := limiter.Allow(ctx, "test:user:2", config)
		assert.NoError(t, err)
	}

	// user1 的第 3 次应该被限流
	err := limiter.Allow(ctx, "test:user:1", config)
	assert.Error(t, err)

	// user2 的第 3 次也应该被限流
	err = limiter.Allow(ctx, "test:user:2", config)
	assert.Error(t, err)
}

func TestRedisLimiter_NoLimit(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	limiter := NewRedisLimiter(rdb)
	ctx := context.Background()

	// 没有配置限流
	err := limiter.Allow(ctx, "test:user:999", nil)
	assert.NoError(t, err)

	// 配置为 0 表示不限制
	config := &Config{
		PerSecond: 0,
		PerMinute: 0,
	}

	for i := 0; i < 100; i++ {
		err := limiter.Allow(ctx, "test:user:999", config)
		assert.NoError(t, err)
	}
}

func TestRedisLimiter_SlidingWindow(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	limiter := NewRedisLimiter(rdb)
	ctx := context.Background()

	config := &Config{
		PerSecond: 2,
	}

	// Mock 时间
	mockTime := int64(1000)
	getNowUnix = func() int64 {
		return mockTime
	}
	defer func() {
		getNowUnix = func() int64 {
			return time.Now().Unix()
		}
	}()

	// 时间 1000: 第 1 次请求
	err := limiter.Allow(ctx, "test:sliding", config)
	assert.NoError(t, err)

	// 时间 1000: 第 2 次请求
	err = limiter.Allow(ctx, "test:sliding", config)
	assert.NoError(t, err)

	// 时间 1000: 第 3 次请求，应该被限流
	err = limiter.Allow(ctx, "test:sliding", config)
	assert.Error(t, err)

	// 时间前进 2 秒到 1002，窗口完全滑动（超过 1 秒窗口）
	mockTime = 1002

	// 时间 1002: 应该可以再次请求（窗口已滑动，时间 1000 的请求已过期）
	err = limiter.Allow(ctx, "test:sliding", config)
	assert.NoError(t, err)
}

func BenchmarkRedisLimiter_Allow(b *testing.B) {
	rdb, cleanup := setupTestRedis(&testing.T{})
	defer cleanup()

	limiter := NewRedisLimiter(rdb)
	ctx := context.Background()

	config := &Config{
		PerSecond: 1000000, // 设置很大的限制，避免触发限流
		PerMinute: 1000000,
		PerHour:   1000000,
		PerDay:    1000000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = limiter.Allow(ctx, "bench:test", config)
	}
}
