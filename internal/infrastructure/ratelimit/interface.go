package ratelimit

// RateLimiter defines the interface for rate limiting.
type RateLimiter interface {
	// Check checks if a request should be allowed.
	// Returns error if rate limit exceeded.
	Check(scopeType string, scopeID string) error

	// LoadRules loads rate limit rules from the database.
	LoadRules() error
}
