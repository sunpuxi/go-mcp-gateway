package circuitbreaker

import (
	"fmt"
	"sync"
	"time"
)

// State 表示熔断器的当前状态
type State int

const (
	StateClosed   State = iota // 正常，请求放行，累计失败
	StateOpen                  // 熔断打开，拒绝请求
	StateHalfOpen              // 半开，放行少量探测请求
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// Config 熔断器配置
type Config struct {
	FailureThreshold    int           // 连续失败多少次后打开熔断，默认 5
	OpenTimeout         time.Duration // 打开状态持续多久后进入半开，默认 30s
	HalfOpenMaxRequests int           // 半开状态最多放行多少个探测请求，默认 1
	SuccessThreshold    int           // 半开状态连续成功多少次后关闭熔断，默认 1
}

// DefaultConfig 返回默认熔断配置
func DefaultConfig() Config {
	return Config{
		FailureThreshold:    5,
		OpenTimeout:         30 * time.Second,
		HalfOpenMaxRequests: 1,
		SuccessThreshold:    1,
	}
}

// CircuitBreaker 是一个按 Project（下游服务）维度的熔断器
// 所有方法线程安全
type CircuitBreaker struct {
	mu sync.Mutex

	state    State
	failures int // 连续失败计数（Closed / HalfOpen）
	successes int // 连续成功计数（HalfOpen）

	cfg      Config
	openedAt time.Time // 熔断打开的时间

	// 半开状态下当前正在执行的探测请求数
	halfOpenInFlight int
}

// New 创建一个新的熔断器，初始状态为 Closed
func New(cfg Config) *CircuitBreaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = DefaultConfig().FailureThreshold
	}
	if cfg.OpenTimeout <= 0 {
		cfg.OpenTimeout = DefaultConfig().OpenTimeout
	}
	if cfg.HalfOpenMaxRequests <= 0 {
		cfg.HalfOpenMaxRequests = DefaultConfig().HalfOpenMaxRequests
	}
	if cfg.SuccessThreshold <= 0 {
		cfg.SuccessThreshold = DefaultConfig().SuccessThreshold
	}
	return &CircuitBreaker{
		state: StateClosed,
		cfg:   cfg,
	}
}

// Allow 检查当前是否允许发送请求。
// 返回 true 表示可以发送请求，调用方必须在请求完成后调用 RecordSuccess 或 RecordFailure。
// 返回 false 表示熔断打开，请求被拒绝。
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		if time.Since(cb.openedAt) >= cb.cfg.OpenTimeout {
			// 超时，进入半开状态
			cb.state = StateHalfOpen
			cb.successes = 0
			cb.halfOpenInFlight = 0
			// 不 return，继续走半开逻辑
		} else {
			return false
		}
		fallthrough

	case StateHalfOpen:
		if cb.halfOpenInFlight < cb.cfg.HalfOpenMaxRequests {
			cb.halfOpenInFlight++
			return true
		}
		return false

	default:
		return false
	}
}

// RecordSuccess 记录一次成功的请求
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		// 成功一次，重置连续失败计数
		cb.failures = 0

	case StateHalfOpen:
		cb.halfOpenInFlight--
		cb.successes++
		if cb.successes >= cb.cfg.SuccessThreshold {
			// 连续成功达到阈值，关闭熔断
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
			cb.halfOpenInFlight = 0
		}

	case StateOpen:
		// 不应该出现：Open 状态不允许请求通过
		// 忽略
	}
}

// RecordFailure 记录一次失败的请求
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.cfg.FailureThreshold {
			// 连续失败达到阈值，打开熔断
			cb.state = StateOpen
			cb.openedAt = time.Now()
		}

	case StateHalfOpen:
		cb.halfOpenInFlight--
		// 半开状态下探测失败，立即重新打开熔断
		cb.state = StateOpen
		cb.openedAt = time.Now()
		cb.successes = 0

	case StateOpen:
		// 不应该出现：Open 状态不允许请求通过
		// 忽略
	}
}

// State 返回当前熔断器状态（只读）
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Stats 返回当前熔断器的统计信息，用于 Admin API 暴露
func (cb *CircuitBreaker) Stats() Stats {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return Stats{
		State:             cb.state,
		Failures:          cb.failures,
		Successes:         cb.successes,
		HalfOpenInFlight:  cb.halfOpenInFlight,
		FailureThreshold:  cb.cfg.FailureThreshold,
		SuccessThreshold:  cb.cfg.SuccessThreshold,
		OpenedAt:          cb.openedAt,
	}
}

// Stats 熔断器统计信息快照
type Stats struct {
	State             State     `json:"state"`
	Failures          int       `json:"failures"`
	Successes         int       `json:"successes"`
	HalfOpenInFlight  int       `json:"half_open_in_flight"`
	FailureThreshold  int       `json:"failure_threshold"`
	SuccessThreshold  int       `json:"success_threshold"`
	OpenedAt          time.Time `json:"opened_at"`
}
