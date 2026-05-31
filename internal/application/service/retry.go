package service

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/sunpuxi/go-mcp-gateway/pkg/logger"
)

// httpRequester 抽象 HTTP 请求能力，方便单元测试 mock
type httpRequester interface {
	DoRequest(method, url string, header http.Header, body []byte, timeoutMs int) (int, []byte, error)
}

// retryResult 封装单次 HTTP 请求的结果
type retryResult struct {
	statusCode int
	body       []byte
	err        error
}

// doRequestWithRetry 根据 RetryConfig 执行带重试的 HTTP 请求
// cfg 为 nil 时直接发一次请求，不走重试逻辑
func doRequestWithRetry(
	client httpRequester,
	method, url string,
	header map[string][]string,
	body []byte,
	timeoutMs int,
	cfg retryConfigProvider,
) (int, []byte, error) {
	if cfg == nil || cfg.GetMaxRetries() <= 0 {
		return client.DoRequest(method, url, header, body, timeoutMs)
	}

	var lastResult retryResult
	maxAttempts := cfg.GetMaxRetries() + 1 // 总尝试次数 = 1 次原始 + N 次重试

	for attempt := 0; attempt < maxAttempts; attempt++ {
		statusCode, respBody, httpErr := client.DoRequest(method, url, header, body, timeoutMs)

		// 请求成功 → 不重试
		if httpErr == nil && !cfg.ShouldRetry(method, statusCode) {
			return statusCode, respBody, nil
		}

		// 记录最后一次失败结果
		lastResult = retryResult{statusCode, respBody, httpErr}

		// 最后一次尝试，不再等待
		if attempt >= maxAttempts-1 {
			break
		}

		delay := backoffDuration(attempt, cfg.GetBackoffType())
		logger.Warn("下游调用失败，准备重试",
			"method", method,
			"attempt", attempt+1,
			"max_retries", cfg.GetMaxRetries(),
			"retry_after_ms", delay.Milliseconds(),
		)
		time.Sleep(delay)
	}

	logger.Error("重试耗尽，下游仍不可用",
		"method", method,
		"total_attempts", maxAttempts,
	)
	return lastResult.statusCode, lastResult.body, lastResult.err
}

// backoffDuration 计算退避等待时间（含 jitter）
//   - fixed: 每次固定 1s
//   - exponential: 1s, 2s, 4s, 8s, ...
//   - jitter: ±25% 随机抖动，防止惊群效应
//
// 声明为 var 而非 func，方便测试时替换
var backoffDuration = func(attempt int, backoffType string) time.Duration {
	const base = time.Second

	var d time.Duration
	switch backoffType {
	case "fixed":
		d = base
	default: // exponential
		d = base * time.Duration(1<<attempt) // 1s, 2s, 4s, 8s...
	}

	// jitter: ±25%
	jitterRange := d / 4
	jitter := time.Duration(rand.Int63n(int64(jitterRange*2+1))) - jitterRange

	return d + jitter
}

// retryConfigProvider 接口，解耦 RetryConfig 具体类型
type retryConfigProvider interface {
	// GetMaxRetries 获取最大的重试次数
	GetMaxRetries() int

	// GetBackoffType 获取降级类型
	GetBackoffType() string

	// ShouldRetry 是否应该重试
	ShouldRetry(method string, statusCode int) bool
}
