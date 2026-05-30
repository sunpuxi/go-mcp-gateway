package entity

import (
	"encoding/json"
	"testing"
)

func rawJSON(s string) *json.RawMessage {
	r := json.RawMessage(s)
	return &r
}

func TestParseRetryConfig_NilInput(t *testing.T) {
	cfg, err := ParseRetryConfig(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Fatal("expected nil config for nil input")
	}
}

func TestParseRetryConfig_EmptyInput(t *testing.T) {
	cfg, err := ParseRetryConfig(rawJSON(``))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Fatal("expected nil config for empty input")
	}
}

func TestParseRetryConfig_ZeroMaxRetries(t *testing.T) {
	cfg, err := ParseRetryConfig(rawJSON(`{"max_retries":0}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Fatal("expected nil when max_retries=0")
	}
}

func TestParseRetryConfig_DefaultsApplied(t *testing.T) {
	cfg, err := ParseRetryConfig(rawJSON(`{"max_retries":3}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", cfg.MaxRetries)
	}
	if cfg.BackoffType != "exponential" {
		t.Errorf("BackoffType = %s, want exponential", cfg.BackoffType)
	}
	if len(cfg.RetryOnStatus) != 3 {
		t.Fatalf("RetryOnStatus length = %d, want 3", len(cfg.RetryOnStatus))
	}
	if cfg.RetryOnMethods[0] != "GET" {
		t.Errorf("RetryOnMethods[0] = %s, want GET", cfg.RetryOnMethods[0])
	}
}

func TestParseRetryConfig_FullCustom(t *testing.T) {
	jsonStr := `{
		"max_retries": 5,
		"backoff_type": "fixed",
		"retry_on_status": [500, 502],
		"retry_on_methods": ["GET", "POST"]
	}`
	cfg, err := ParseRetryConfig(rawJSON(jsonStr))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", cfg.MaxRetries)
	}
	if cfg.BackoffType != "fixed" {
		t.Errorf("BackoffType = %s, want fixed", cfg.BackoffType)
	}
	if len(cfg.RetryOnStatus) != 2 || cfg.RetryOnStatus[0] != 500 || cfg.RetryOnStatus[1] != 502 {
		t.Errorf("RetryOnStatus = %v, want [500, 502]", cfg.RetryOnStatus)
	}
	if len(cfg.RetryOnMethods) != 2 || cfg.RetryOnMethods[0] != "GET" || cfg.RetryOnMethods[1] != "POST" {
		t.Errorf("RetryOnMethods = %v, want [GET, POST]", cfg.RetryOnMethods)
	}
}

func TestShouldRetry_NilConfig(t *testing.T) {
	var cfg *RetryConfig
	if cfg.ShouldRetry("GET", 502) {
		t.Error("nil config should not retry")
	}
}

func TestShouldRetry_MethodNotAllowed(t *testing.T) {
	cfg := &RetryConfig{
		MaxRetries:     3,
		RetryOnMethods: []string{"GET"},
		RetryOnStatus:  []int{502, 503, 504},
	}
	if cfg.ShouldRetry("POST", 502) {
		t.Error("POST should not retry when only GET is allowed")
	}
}

func TestShouldRetry_StatusNotMatched(t *testing.T) {
	cfg := &RetryConfig{
		MaxRetries:     3,
		RetryOnMethods: []string{"GET"},
		RetryOnStatus:  []int{502, 503, 504},
	}
	if cfg.ShouldRetry("GET", 500) {
		t.Error("500 should not retry when only 502/503/504 configured")
	}
}

func TestShouldRetry_AllConditionsMet(t *testing.T) {
	cfg := &RetryConfig{
		MaxRetries:     3,
		RetryOnMethods: []string{"GET", "POST"},
		RetryOnStatus:  []int{502, 503, 504},
	}

	tests := []struct {
		method string
		status int
		want   bool
	}{
		{"GET", 502, true},
		{"GET", 503, true},
		{"GET", 504, true},
		{"POST", 502, true},
		{"POST", 200, false},
		{"DELETE", 502, false},
		{"GET", 400, false},
	}
	for _, tt := range tests {
		got := cfg.ShouldRetry(tt.method, tt.status)
		if got != tt.want {
			t.Errorf("ShouldRetry(%s, %d) = %v, want %v", tt.method, tt.status, got, tt.want)
		}
	}
}
