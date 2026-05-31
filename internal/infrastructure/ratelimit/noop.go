package ratelimit

import "context"

// NoopLimiter 空实现，始终放行（Redis 未配置时使用）
type NoopLimiter struct{}

func (n *NoopLimiter) Allow(ctx context.Context, key string, maxRequests int, windowSeconds int) (bool, int, error) {
	return true, maxRequests, nil
}
