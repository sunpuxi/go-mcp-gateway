package ratelimit

import (
	"github.com/redis/go-redis/v9"
)

// RedisRateLimiter implements sliding window rate limiting with Redis.
type RedisRateLimiter struct {
	client *redis.Client
}

// NewRedisRateLimiter creates a new RedisRateLimiter.
func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{client: client}
}

// Check checks if a request should be allowed.
func (r *RedisRateLimiter) Check(scopeType, scopeID string) error {
	return nil // TODO: implement
}

// LoadRules loads rate limit rules from the database.
func (r *RedisRateLimiter) LoadRules() error {
	return nil // TODO: implement
}
