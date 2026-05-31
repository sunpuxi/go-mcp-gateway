package circuitbreaker

import (
	"sync"
	"testing"
	"time"
)

func TestCircuitBreaker_Closed_AllowsRequest(t *testing.T) {
	cb := New(DefaultConfig())
	if !cb.Allow() {
		t.Error("Allow should return true in Closed state")
	}
	if cb.State() != StateClosed {
		t.Error("should remain Closed after Allow")
	}
}

func TestCircuitBreaker_Closed_RecordsSuccessResetsFailures(t *testing.T) {
	cb := New(Config{
		FailureThreshold: 3,
		OpenTimeout:      30 * time.Second,
	})

	// Accumulate 2 failures
	cb.Allow()
	cb.RecordFailure()
	cb.Allow()
	cb.RecordFailure()

	if cb.Stats().Failures != 2 {
		t.Fatalf("failures = %d, want 2", cb.Stats().Failures)
	}

	// One success resets the counter
	cb.Allow()
	cb.RecordSuccess()

	if cb.Stats().Failures != 0 {
		t.Errorf("failures after success = %d, want 0", cb.Stats().Failures)
	}
	if cb.State() != StateClosed {
		t.Error("should remain Closed after success")
	}
}

func TestCircuitBreaker_ClosedToOpen_AfterThresholdFailures(t *testing.T) {
	cb := New(Config{
		FailureThreshold: 3,
		OpenTimeout:      30 * time.Second,
	})

	for i := 0; i < 3; i++ {
		if !cb.Allow() {
			t.Fatalf("attempt %d: Allow should return true", i)
		}
		cb.RecordFailure()
	}

	if cb.State() != StateOpen {
		t.Fatalf("state = %s, want Open after %d failures", cb.State(), 3)
	}

	// Subsequent Allow should reject
	if cb.Allow() {
		t.Error("Allow should return false in Open state")
	}
}

func TestCircuitBreaker_Open_RejectsRequests(t *testing.T) {
	cb := New(Config{
		FailureThreshold: 1,
		OpenTimeout:      30 * time.Second,
	})

	// Trip the breaker
	cb.Allow()
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Fatal("expected Open state")
	}

	for i := 0; i < 10; i++ {
		if cb.Allow() {
			t.Errorf("attempt %d: Allow should return false in Open state", i)
		}
	}
}

func TestCircuitBreaker_OpenToHalfOpen_AfterTimeout(t *testing.T) {
	cb := New(Config{
		FailureThreshold: 1,
		OpenTimeout:      10 * time.Millisecond,
	})

	// Trip the breaker
	cb.Allow()
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatal("expected Open state")
	}

	// Wait for timeout
	time.Sleep(15 * time.Millisecond)

	// Next Allow should transition to HalfOpen and allow
	if !cb.Allow() {
		t.Error("Allow should return true, transitioning to HalfOpen after timeout")
	}
	if cb.State() != StateHalfOpen {
		t.Errorf("state = %s, want HalfOpen", cb.State())
	}
}

func TestCircuitBreaker_HalfOpen_SuccessClosesBreaker(t *testing.T) {
	cb := New(Config{
		FailureThreshold:    1,
		OpenTimeout:         1 * time.Millisecond,
		HalfOpenMaxRequests: 3,
		SuccessThreshold:    2,
	})

	// Trip the breaker
	cb.Allow()
	cb.RecordFailure()
	time.Sleep(2 * time.Millisecond)

	// Transition to HalfOpen
	if !cb.Allow() {
		t.Fatal("Allow should return true in HalfOpen")
	}

	// First success
	cb.RecordSuccess()
	if cb.State() != StateHalfOpen {
		t.Error("should remain HalfOpen after first success (need 2)")
	}

	// Second request
	if !cb.Allow() {
		t.Fatal("Allow should return true")
	}
	cb.RecordSuccess()

	// Should close after 2 consecutive successes
	if cb.State() != StateClosed {
		t.Errorf("state = %s, want Closed after 2 successes", cb.State())
	}
}

func TestCircuitBreaker_HalfOpen_FailureReopensBreaker(t *testing.T) {
	cb := New(Config{
		FailureThreshold:    1,
		OpenTimeout:         1 * time.Millisecond,
		HalfOpenMaxRequests: 3,
		SuccessThreshold:    2,
	})

	// Trip the breaker
	cb.Allow()
	cb.RecordFailure()
	time.Sleep(2 * time.Millisecond)

	// Transition to HalfOpen
	if !cb.Allow() {
		t.Fatal("Allow should return true in HalfOpen")
	}

	// Fail the probe
	cb.RecordFailure()

	// Should reopen
	if cb.State() != StateOpen {
		t.Errorf("state = %s, want Open after probe failure", cb.State())
	}

	// Should reject subsequent requests
	if cb.Allow() {
		t.Error("Allow should return false after reopening")
	}
}

func TestCircuitBreaker_HalfOpen_LimitsConcurrentProbes(t *testing.T) {
	cb := New(Config{
		FailureThreshold:    1,
		OpenTimeout:         1 * time.Millisecond,
		HalfOpenMaxRequests: 2,
	})

	// Trip and wait for HalfOpen
	cb.Allow()
	cb.RecordFailure()
	time.Sleep(2 * time.Millisecond)

	// First two probes pass
	if !cb.Allow() {
		t.Error("probe 1: should pass")
	}
	if !cb.Allow() {
		t.Error("probe 2: should pass")
	}

	// Third probe rejected — halfOpenMaxRequests is 2
	if cb.Allow() {
		t.Error("probe 3: should be rejected (max 2 probes)")
	}

	// After one completes (success), a new slot opens
	cb.RecordSuccess()
	if !cb.Allow() {
		t.Error("probe after success: should be allowed (slot freed)")
	}
}

func TestCircuitBreaker_Closed_NonConsecutiveFailuresDontTrip(t *testing.T) {
	cb := New(Config{
		FailureThreshold: 3,
		OpenTimeout:      30 * time.Second,
	})

	// Interleave success and failure
	for i := 0; i < 5; i++ {
		cb.Allow()
		cb.RecordFailure()
		cb.Allow()
		cb.RecordSuccess() // resets counter each time
	}

	if cb.State() != StateClosed {
		t.Error("should remain Closed when failures are not consecutive")
	}
}

func TestCircuitBreaker_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.FailureThreshold != 5 {
		t.Errorf("FailureThreshold = %d, want 5", cfg.FailureThreshold)
	}
	if cfg.OpenTimeout != 30*time.Second {
		t.Errorf("OpenTimeout = %v, want 30s", cfg.OpenTimeout)
	}
	if cfg.HalfOpenMaxRequests != 1 {
		t.Errorf("HalfOpenMaxRequests = %d, want 1", cfg.HalfOpenMaxRequests)
	}
	if cfg.SuccessThreshold != 1 {
		t.Errorf("SuccessThreshold = %d, want 1", cfg.SuccessThreshold)
	}
}

func TestCircuitBreaker_NewWithZeroConfig_UsesDefaults(t *testing.T) {
	cb := New(Config{})
	stats := cb.Stats()
	if stats.FailureThreshold != 5 {
		t.Errorf("FailureThreshold = %d, want 5", stats.FailureThreshold)
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := New(Config{
		FailureThreshold: 100,
		OpenTimeout:      30 * time.Second,
	})

	var wg sync.WaitGroup
	goroutines := 50
	iterations := 200

	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				if cb.Allow() {
					// Randomly succeed or fail
					if i%2 == 0 {
						cb.RecordSuccess()
					} else {
						cb.RecordFailure()
					}
				}
			}
		}()
	}

	wg.Wait()

	// Should not panic or deadlock; state should be valid
	state := cb.State()
	if state != StateClosed && state != StateOpen && state != StateHalfOpen {
		t.Errorf("invalid state after concurrent access: %s", state)
	}
}

func TestCircuitBreaker_Stats_Snapshot(t *testing.T) {
	cb := New(Config{
		FailureThreshold: 5,
		OpenTimeout:      30 * time.Second,
	})

	cb.Allow()
	cb.RecordFailure()
	cb.Allow()
	cb.RecordFailure()

	stats := cb.Stats()
	if stats.State != StateClosed {
		t.Errorf("state = %s, want Closed", stats.State)
	}
	if stats.Failures != 2 {
		t.Errorf("failures = %d, want 2", stats.Failures)
	}
	if stats.FailureThreshold != 5 {
		t.Errorf("threshold = %d, want 5", stats.FailureThreshold)
	}
}

func TestState_String(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(99), "unknown(99)"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestCircuitBreaker_StateRemainsOpenBeforeTimeout(t *testing.T) {
	cb := New(Config{
		FailureThreshold: 1,
		OpenTimeout:      5 * time.Second,
	})

	cb.Allow()
	cb.RecordFailure()

	if cb.Allow() {
		t.Error("should reject immediately after opening (timeout not elapsed)")
	}
}

func TestCircuitBreaker_HalfOpenSuccessResetsAllCounters(t *testing.T) {
	cb := New(Config{
		FailureThreshold:    5,
		OpenTimeout:         1 * time.Millisecond,
		HalfOpenMaxRequests: 1,
		SuccessThreshold:    1,
	})

	cb.Allow()
	cb.RecordFailure() // trip
	time.Sleep(2 * time.Millisecond)

	cb.Allow()
	cb.RecordSuccess() // probe success → close

	if cb.State() != StateClosed {
		t.Fatal("expected Closed after successful probe")
	}

	stats := cb.Stats()
	if stats.Failures != 0 {
		t.Errorf("failures = %d, want 0 (reset after close)", stats.Failures)
	}
	if stats.Successes != 0 {
		t.Errorf("successes = %d, want 0 (reset after close)", stats.Successes)
	}
}
