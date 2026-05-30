package service

import (
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/sunpuxi/go-mcp-gateway/pkg/logger"
)

func init() {
	// 测试环境初始化一个静默 logger
	_ = logger.Init(logger.Config{Level: "error", Format: "console"})
}

// mockRetryConfig 实现 retryConfigProvider 接口
type mockRetryConfig struct {
	maxRetries     int
	backoffType    string
	retryOnStatus  []int
	retryOnMethods []string
}

func (m *mockRetryConfig) GetMaxRetries() int    { return m.maxRetries }
func (m *mockRetryConfig) GetBackoffType() string { return m.backoffType }

func (m *mockRetryConfig) ShouldRetry(method string, statusCode int) bool {
	methodOk := false
	for _, allowed := range m.retryOnMethods {
		if method == allowed {
			methodOk = true
			break
		}
	}
	statusOk := false
	for _, allowed := range m.retryOnStatus {
		if statusCode == allowed {
			statusOk = true
			break
		}
	}
	return methodOk && statusOk
}

// mockHTTPClient 实现 httpRequester 接口
type mockHTTPClient struct {
	responses []mockResponse
	mu        sync.Mutex
	callCount int
}

type mockResponse struct {
	statusCode int
	body       []byte
	err        error
}

func (m *mockHTTPClient) DoRequest(method, url string, header http.Header, body []byte, timeoutMs int) (int, []byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.callCount >= len(m.responses) {
		return 0, nil, errors.New("unexpected call: no more mocked responses")
	}
	r := m.responses[m.callCount]
	m.callCount++
	return r.statusCode, r.body, r.err
}

func makeHeader() http.Header {
	return http.Header{"Content-Type": {"application/json"}}
}

// 测试期间把 backoff 设为 0，避免等待
var testBackoff = func(attempt int, backoffType string) time.Duration {
	return 0
}

func TestDoRequestWithRetry_NilConfig(t *testing.T) {
	mock := &mockHTTPClient{
		responses: []mockResponse{
			{statusCode: 200, body: []byte(`"ok"`)},
		},
	}

	code, body, err := doRequestWithRetry(mock, "GET", "http://test/api", makeHeader(), nil, 5000, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 200 {
		t.Errorf("status = %d, want 200", code)
	}
	if string(body) != `"ok"` {
		t.Errorf("body = %s, want \"ok\"", body)
	}
}

func TestDoRequestWithRetry_FirstAttemptSucceeds(t *testing.T) {
	cfg := &mockRetryConfig{maxRetries: 3, retryOnStatus: []int{502}, retryOnMethods: []string{"GET"}}
	mock := &mockHTTPClient{
		responses: []mockResponse{{statusCode: 200}},
	}

	// 临时替换 backoff 避免等待
	origBackoff := backoffDuration
	backoffDuration = testBackoff
	defer func() { backoffDuration = origBackoff }()

	code, _, err := doRequestWithRetry(mock, "GET", "http://test/api", makeHeader(), nil, 5000, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 200 {
		t.Errorf("status = %d, want 200", code)
	}
	if mock.callCount != 1 {
		t.Errorf("callCount = %d, want 1", mock.callCount)
	}
}

func TestDoRequestWithRetry_RetryThenSucceed(t *testing.T) {
	cfg := &mockRetryConfig{maxRetries: 3, retryOnStatus: []int{502}, retryOnMethods: []string{"GET"}}
	mock := &mockHTTPClient{
		responses: []mockResponse{
			{statusCode: 502},
			{statusCode: 502},
			{statusCode: 200, body: []byte(`"recovered"`)},
		},
	}

	origBackoff := backoffDuration
	backoffDuration = testBackoff
	defer func() { backoffDuration = origBackoff }()

	code, body, err := doRequestWithRetry(mock, "GET", "http://test/api", makeHeader(), nil, 5000, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 200 {
		t.Errorf("status = %d, want 200", code)
	}
	if string(body) != `"recovered"` {
		t.Errorf("body = %s, want \"recovered\"", body)
	}
	if mock.callCount != 3 {
		t.Errorf("callCount = %d, want 3", mock.callCount)
	}
}

func TestDoRequestWithRetry_AllRetriesExhausted(t *testing.T) {
	cfg := &mockRetryConfig{maxRetries: 2, retryOnStatus: []int{502}, retryOnMethods: []string{"GET"}}
	mock := &mockHTTPClient{
		responses: []mockResponse{
			{statusCode: 502}, // 原始
			{statusCode: 502}, // 重试 1
			{statusCode: 502}, // 重试 2
		},
	}

	origBackoff := backoffDuration
	backoffDuration = testBackoff
	defer func() { backoffDuration = origBackoff }()

	code, _, err := doRequestWithRetry(mock, "GET", "http://test/api", makeHeader(), nil, 5000, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 502 {
		t.Errorf("status = %d, want 502", code)
	}
	if mock.callCount != 3 {
		t.Errorf("callCount = %d, want 3 (1 original + 2 retries)", mock.callCount)
	}
}

func TestDoRequestWithRetry_NonRetryableStatus(t *testing.T) {
	cfg := &mockRetryConfig{maxRetries: 3, retryOnStatus: []int{502, 503, 504}, retryOnMethods: []string{"GET"}}
	mock := &mockHTTPClient{
		responses: []mockResponse{{statusCode: 400}}, // 400 不在重试列表
	}

	code, _, err := doRequestWithRetry(mock, "GET", "http://test/api", makeHeader(), nil, 5000, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 400 {
		t.Errorf("status = %d, want 400", code)
	}
	if mock.callCount != 1 {
		t.Errorf("callCount = %d, want 1 (no retry for 400)", mock.callCount)
	}
}

func TestDoRequestWithRetry_PostNotRetriedWhenNotAllowed(t *testing.T) {
	cfg := &mockRetryConfig{maxRetries: 3, retryOnStatus: []int{502}, retryOnMethods: []string{"GET"}}
	mock := &mockHTTPClient{
		responses: []mockResponse{{statusCode: 502}},
	}

	code, _, err := doRequestWithRetry(mock, "POST", "http://test/api", makeHeader(), nil, 5000, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 502 {
		t.Errorf("status = %d, want 502", code)
	}
	if mock.callCount != 1 {
		t.Errorf("callCount = %d, want 1 (POST not allowed)", mock.callCount)
	}
}

func TestDoRequestWithRetry_NetworkErrorRetried(t *testing.T) {
	cfg := &mockRetryConfig{maxRetries: 2, retryOnStatus: []int{502}, retryOnMethods: []string{"GET"}}
	mock := &mockHTTPClient{
		responses: []mockResponse{
			{err: errors.New("connection refused")}, // 网络错误 → 重试
			{err: errors.New("connection refused")}, // 网络错误 → 重试
			{statusCode: 200, body: []byte(`"ok"`)},
		},
	}

	origBackoff := backoffDuration
	backoffDuration = testBackoff
	defer func() { backoffDuration = origBackoff }()

	code, body, err := doRequestWithRetry(mock, "GET", "http://test/api", makeHeader(), nil, 5000, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 200 {
		t.Errorf("status = %d, want 200", code)
	}
	if string(body) != `"ok"` {
		t.Errorf("body = %s, want \"ok\"", body)
	}
	if mock.callCount != 3 {
		t.Errorf("callCount = %d, want 3", mock.callCount)
	}
}

func TestBackoffDuration_Fixed(t *testing.T) {
	for i := 0; i < 5; i++ {
		d := backoffDuration(i, "fixed")
		if d < 750*time.Millisecond || d > 1250*time.Millisecond {
			t.Errorf("attempt %d: duration = %v, want ~1s (±25%% jitter)", i, d)
		}
	}
}

func TestBackoffDuration_Exponential(t *testing.T) {
	tests := []struct {
		attempt int
		min     time.Duration
		max     time.Duration
	}{
		{0, 750 * time.Millisecond, 1250 * time.Millisecond},
		{1, 1500 * time.Millisecond, 2500 * time.Millisecond},
		{2, 3000 * time.Millisecond, 5000 * time.Millisecond},
	}

	for _, tt := range tests {
		d := backoffDuration(tt.attempt, "exponential")
		if d < tt.min || d > tt.max {
			t.Errorf("attempt %d: duration = %v, want [%v, %v]", tt.attempt, d, tt.min, tt.max)
		}
	}
}
