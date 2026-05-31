package entity

import "encoding/json"

// RateLimitConfig 定义对某个工具调用的限流策略（Tool 级别）
type RateLimitConfig struct {
	MaxRequests   int `json:"max_requests"`   // 窗口内最大请求数，0 = 不限流
	WindowSeconds int `json:"window_seconds"` // 滑动窗口大小（秒），默认 1
}

// ParseRateLimitConfig 将 RawMessage 解析为 RateLimitConfig，nil 输入表示未配置
func ParseRateLimitConfig(raw *json.RawMessage) (*RateLimitConfig, error) {
	if raw == nil || len(*raw) == 0 {
		return nil, nil
	}
	var cfg RateLimitConfig
	if err := json.Unmarshal(*raw, &cfg); err != nil {
		return nil, err
	}
	if cfg.MaxRequests <= 0 {
		return nil, nil // MaxRequests=0 或负数视为不限流
	}
	if cfg.WindowSeconds <= 0 {
		cfg.WindowSeconds = 1 // 默认 1 秒窗口
	}
	return &cfg, nil
}

// IsEnabled 返回限流配置是否生效
func (c *RateLimitConfig) IsEnabled() bool {
	return c != nil && c.MaxRequests > 0
}

// GetMaxRequests 返回窗口内最大请求数（rateLimitConfigProvider 接口适配）
func (c *RateLimitConfig) GetMaxRequests() int {
	if c == nil {
		return 0
	}
	return c.MaxRequests
}

// GetWindowSeconds 返回窗口大小秒数（rateLimitConfigProvider 接口适配）
func (c *RateLimitConfig) GetWindowSeconds() int {
	if c == nil {
		return 0
	}
	return c.WindowSeconds
}
