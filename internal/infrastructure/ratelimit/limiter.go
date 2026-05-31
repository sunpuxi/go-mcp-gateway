package ratelimit

import "context"

// RateLimiter 限流器接口，支持滑动窗口计数算法
type RateLimiter interface {
	// Allow 检查指定 key 是否允许通过。
	// maxRequests: 窗口内最大请求数
	// windowSeconds: 滑动窗口大小（秒）
	// 返回: allowed（是否放行）, remaining（剩余配额）, error
	Allow(ctx context.Context, key string, maxRequests int, windowSeconds int) (allowed bool, remaining int, err error)
}
