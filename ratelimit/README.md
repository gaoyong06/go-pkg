# Rate Limiter

基于 Redis 的分布式限流器，支持多时间窗口限流。

## 特性

- ✅ 支持多时间窗口（per_second, per_minute, per_hour, per_day）
- ✅ 使用 Lua 脚本保证原子性
- ✅ 滑动窗口算法，精确度高
- ✅ 支持分布式场景
- ✅ 性能优秀（单次 Redis 调用）

## 安装

```bash
go get github.com/gaoyong06/go-pkg/ratelimit
```

## 使用示例

### 基本用法

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/gaoyong06/go-pkg/ratelimit"
    "github.com/go-redis/redis/v8"
)

func main() {
    // 创建 Redis 客户端
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // 创建限流器
    limiter := ratelimit.NewRedisLimiter(rdb)
    
    // 配置限流规则
    config := &ratelimit.Config{
        PerSecond: 10,    // 每秒最多 10 次
        PerMinute: 100,   // 每分钟最多 100 次
        PerHour:   1000,  // 每小时最多 1000 次
        PerDay:    10000, // 每天最多 10000 次
    }
    
    // 检查限流
    ctx := context.Background()
    err := limiter.Allow(ctx, "user:123", config)
    if err != nil {
        if ratelimit.IsRateLimitError(err) {
            fmt.Println("触发限流:", err)
            return
        }
        fmt.Println("检查失败:", err)
        return
    }
    
    fmt.Println("请求通过")
}
```

### 在 HTTP Handler 中使用

```go
func SendSMSHandler(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    
    // 限流检查
    config := &ratelimit.Config{
        PerSecond: 1,   // 每秒最多 1 次
        PerMinute: 5,   // 每分钟最多 5 次
        PerDay:    20,  // 每天最多 20 次
    }
    
    key := fmt.Sprintf("sms:user:%s", userID)
    if err := limiter.Allow(r.Context(), key, config); err != nil {
        if ratelimit.IsRateLimitError(err) {
            http.Error(w, "Too many requests", http.StatusTooManyRequests)
            return
        }
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }
    
    // 发送短信...
}
```

### 不同场景使用不同限流规则

```go
// SMS 限流
smsConfig := &ratelimit.Config{
    PerSecond: 1,
    PerMinute: 5,
    PerDay:    20,
}

// Email 限流
emailConfig := &ratelimit.Config{
    PerSecond: 2,
    PerMinute: 10,
    PerDay:    100,
}

// API 限流
apiConfig := &ratelimit.Config{
    PerSecond: 100,
    PerMinute: 1000,
}

// 使用
limiter.Allow(ctx, "sms:user:123", smsConfig)
limiter.Allow(ctx, "email:user:123", emailConfig)
limiter.Allow(ctx, "api:user:123", apiConfig)
```

## 配置说明

### Config 结构

```go
type Config struct {
    PerSecond int32 // 每秒限制次数，0 表示不限制
    PerMinute int32 // 每分钟限制次数，0 表示不限制
    PerHour   int32 // 每小时限制次数，0 表示不限制
    PerDay    int32 // 每天限制次数，0 表示不限制
}
```

- 所有字段都是可选的，设置为 0 表示不限制该时间窗口
- 可以同时配置多个时间窗口，任何一个窗口触发限流都会拒绝请求
- 建议根据实际业务需求合理配置

### Key 命名规范

建议使用以下格式：
- `{resource}:{provider}:{identifier}` - 如 `sms:aliyun:user:123`
- `{service}:{user}:{identifier}` - 如 `api:user:123`
- `{action}:{identifier}` - 如 `login:user:123`

## 算法说明

### 滑动窗口 (Sliding Window)

本限流器使用滑动窗口算法，相比固定窗口更加精确：

- **固定窗口**: 在窗口边界可能出现突发流量（如 00:59 和 01:00 各 100 次）
- **滑动窗口**: 任意时刻往前推一个窗口时间，都不会超过限制

### 实现细节

- 使用 Redis Sorted Set 存储请求时间戳
- Score 和 Member 都是时间戳（微秒），保证唯一性
- 定期清理过期数据，避免内存泄漏
- 使用 Lua 脚本保证原子性，避免并发问题

## 性能

### 优化点

1. **Lua 脚本**: 将多次 Redis 调用合并为一次，大幅提升性能
2. **原子操作**: Lua 脚本在 Redis 中原子执行，无需额外锁
3. **自动过期**: 使用 EXPIRE 自动清理过期数据

### 基准测试

```bash
go test -bench=. -benchmem
```

典型结果：
```
BenchmarkRedisLimiter_Allow-8    10000    100000 ns/op    1024 B/op    20 allocs/op
```

## 错误处理

### RateLimitError

当触发限流时，返回 `*RateLimitError`：

```go
type RateLimitError struct {
    Key           string  // 限流键
    WindowName    string  // 触发限流的窗口名称（per_second, per_minute, per_hour, per_day）
    WindowSeconds int64   // 窗口大小（秒）
    Current       int64   // 当前请求数
    Limit         int64   // 限制数
}
```

### 判断是否是限流错误

```go
err := limiter.Allow(ctx, key, config)
if ratelimit.IsRateLimitError(err) {
    // 触发限流
    rateLimitErr := err.(*ratelimit.RateLimitError)
    log.Printf("Rate limit exceeded: %s, current=%d, limit=%d",
        rateLimitErr.WindowName,
        rateLimitErr.Current,
        rateLimitErr.Limit)
}
```

## 最佳实践

1. **合理设置限流值**: 根据实际业务需求和系统容量设置
2. **使用有意义的 Key**: 便于监控和调试
3. **错误处理**: 区分限流错误和系统错误
4. **监控告警**: 监控限流触发情况，及时调整配置
5. **降级策略**: 触发限流后的降级处理

## License

MIT
