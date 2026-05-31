package entity

import (
	"encoding/json"
	"testing"
)

func rawMsg(s string) *json.RawMessage {
	raw := json.RawMessage(s)
	return &raw
}

func TestParseRateLimitConfig_NilInput(t *testing.T) {
	cfg, err := ParseRateLimitConfig(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Error("nil input should return nil config")
	}
}

func TestParseRateLimitConfig_EmptyInput(t *testing.T) {
	cfg, err := ParseRateLimitConfig(rawMsg(``))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Error("empty input should return nil config")
	}
}

func TestParseRateLimitConfig_ZeroMaxRequests(t *testing.T) {
	cfg, err := ParseRateLimitConfig(rawMsg(`{"max_requests":0,"window_seconds":1}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Error("max_requests=0 should return nil config")
	}
}

func TestParseRateLimitConfig_Valid(t *testing.T) {
	cfg, err := ParseRateLimitConfig(rawMsg(`{"max_requests":100,"window_seconds":1}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxRequests != 100 {
		t.Errorf("MaxRequests = %d, want 100", cfg.MaxRequests)
	}
	if cfg.WindowSeconds != 1 {
		t.Errorf("WindowSeconds = %d, want 1", cfg.WindowSeconds)
	}
}

func TestParseRateLimitConfig_DefaultWindowSeconds(t *testing.T) {
	cfg, err := ParseRateLimitConfig(rawMsg(`{"max_requests":50}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.WindowSeconds != 1 {
		t.Errorf("WindowSeconds = %d, want 1 (default)", cfg.WindowSeconds)
	}
}

func TestParseRateLimitConfig_InvalidJSON(t *testing.T) {
	_, err := ParseRateLimitConfig(rawMsg(`{bad json}`))
	if err == nil {
		t.Error("invalid JSON should return error")
	}
}

func TestRateLimitConfig_IsEnabled(t *testing.T) {
	var nilCfg *RateLimitConfig
	if nilCfg.IsEnabled() {
		t.Error("nil config should not be enabled")
	}

	disabled := &RateLimitConfig{MaxRequests: 0, WindowSeconds: 1}
	if disabled.IsEnabled() {
		t.Error("MaxRequests=0 should not be enabled")
	}

	enabled := &RateLimitConfig{MaxRequests: 100, WindowSeconds: 1}
	if !enabled.IsEnabled() {
		t.Error("MaxRequests=100 should be enabled")
	}
}

func TestRateLimitConfig_ProviderInterface(t *testing.T) {
	var nilCfg *RateLimitConfig
	if nilCfg.GetMaxRequests() != 0 {
		t.Error("nil GetMaxRequests should be 0")
	}
	if nilCfg.GetWindowSeconds() != 0 {
		t.Error("nil GetWindowSeconds should be 0")
	}

	cfg := &RateLimitConfig{MaxRequests: 30, WindowSeconds: 5}
	if cfg.GetMaxRequests() != 30 {
		t.Errorf("GetMaxRequests = %d, want 30", cfg.GetMaxRequests())
	}
	if cfg.GetWindowSeconds() != 5 {
		t.Errorf("GetWindowSeconds = %d, want 5", cfg.GetWindowSeconds())
	}
}
