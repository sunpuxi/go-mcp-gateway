package entity

import "encoding/json"

// RetryConfig 定义工具调用失败时的重试策略
type RetryConfig struct {
	MaxRetries     int      `json:"max_retries"`      // 最大重试次数，0 = 不重试
	BackoffType    string   `json:"backoff_type"`     // "fixed" / "exponential"，默认 exponential
	RetryOnStatus  []int    `json:"retry_on_status"`  // 触发重试的 HTTP 状态码，默认 [502, 503, 504]
	RetryOnMethods []string `json:"retry_on_methods"` // 允许重试的 HTTP 方法，默认 ["GET"]
}

// ParseRetryConfig 将 RawMessage 解析为 RetryConfig，nil 输入表示未配置
func ParseRetryConfig(raw *json.RawMessage) (*RetryConfig, error) {
	if raw == nil || len(*raw) == 0 {
		return nil, nil
	}
	var cfg RetryConfig
	if err := json.Unmarshal(*raw, &cfg); err != nil {
		return nil, err
	}
	if cfg.MaxRetries <= 0 {
		return nil, nil // 显式配置为 0 也视为未开启
	}
	if cfg.BackoffType == "" {
		cfg.BackoffType = "exponential"
	}
	if len(cfg.RetryOnStatus) == 0 {
		cfg.RetryOnStatus = []int{502, 503, 504}
	}
	if len(cfg.RetryOnMethods) == 0 {
		cfg.RetryOnMethods = []string{"GET"}
	}
	return &cfg, nil
}

// GetMaxRetries 返回最大重试次数（retryConfigProvider 接口适配）
func (c *RetryConfig) GetMaxRetries() int {
	if c == nil {
		return 0
	}
	return c.MaxRetries
}

// GetBackoffType 返回退避类型（retryConfigProvider 接口适配）
func (c *RetryConfig) GetBackoffType() string {
	if c == nil {
		return ""
	}
	return c.BackoffType
}

// ShouldRetry 判断给定的 HTTP 方法和状态码是否应触发重试
func (c *RetryConfig) ShouldRetry(method string, statusCode int) bool {
	if c == nil {
		return false
	}
	if !containsMethod(c.RetryOnMethods, method) {
		return false
	}
	if !containsStatus(c.RetryOnStatus, statusCode) {
		return false
	}
	return true
}

func containsMethod(methods []string, method string) bool {
	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}

func containsStatus(statuses []int, code int) bool {
	for _, s := range statuses {
		if s == code {
			return true
		}
	}
	return false
}
