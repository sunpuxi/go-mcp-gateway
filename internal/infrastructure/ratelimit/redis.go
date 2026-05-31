package ratelimit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// slidingWindowScript 使用 Sorted Set 实现滑动窗口限流
//
// KEYS[1] = 限流 key
// ARGV[1] = 当前时间戳（毫秒）
// ARGV[2] = 窗口大小（毫秒）
// ARGV[3] = 最大请求数
// ARGV[4] = 唯一成员标识（timestamp + random）
//
// 返回: {allowed (1|0), remaining}
var slidingWindowScript = `
local now = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local max_req = tonumber(ARGV[3])
local member = ARGV[4]

-- 1. 移除窗口外的过期条目
local window_start = now - window_ms
redis.call('ZREMRANGEBYSCORE', KEYS[1], 0, window_start)

-- 2. 统计当前窗口内的请求数
local current = redis.call('ZCARD', KEYS[1])

-- 3. 判断是否放行
if current < max_req then
    redis.call('ZADD', KEYS[1], now, member)
    -- 设置过期时间：窗口 + 1 秒缓冲
    redis.call('EXPIRE', KEYS[1], math.ceil(window_ms / 1000) + 1)
    return {1, max_req - current - 1}
else
    return {0, max_req - current}
end
`

// RedisLimiter 基于 Redis Sorted Set 的滑动窗口限流器
type RedisLimiter struct {
	client *goredis.Client
	script *goredis.Script
}

// NewRedisLimiter 创建 Redis 限流器
func NewRedisLimiter(client *goredis.Client) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		script: goredis.NewScript(slidingWindowScript),
	}
}

// Allow 检查指定 key 是否允许通过
func (l *RedisLimiter) Allow(ctx context.Context, key string, maxRequests int, windowSeconds int) (bool, int, error) {
	if maxRequests <= 0 {
		return true, 0, nil
	}
	if windowSeconds <= 0 {
		windowSeconds = 1
	}

	windowMS := windowSeconds * 1000
	nowMS := time.Now().UnixMilli()
	member := fmt.Sprintf("%d-%s", nowMS, randomHex(8))

	res, err := l.script.Run(ctx, l.client, []string{key},
		nowMS,     // ARGV[1]
		windowMS,  // ARGV[2]
		maxRequests, // ARGV[3]
		member,    // ARGV[4]
	).Result()
	if err != nil {
		// Redis 出错时保守放行，避免限流器成为单点故障
		return true, maxRequests, fmt.Errorf("redis 限流执行失败，降级放行: %w", err)
	}

	// Lua 返回的是 []interface{}: {allowed, remaining}
	arr, ok := res.([]interface{})
	if !ok || len(arr) < 2 {
		return true, maxRequests, nil
	}

	allowed := toInt64(arr[0]) == 1
	remaining := int(toInt64(arr[1]))

	return allowed, remaining, nil
}

// randomHex 生成指定长度的随机 hex 字符串
func randomHex(n int) string {
	b := make([]byte, n/2+1)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	default:
		return 0
	}
}
