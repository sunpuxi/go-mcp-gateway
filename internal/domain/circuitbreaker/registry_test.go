package circuitbreaker

import (
	"sync"
	"testing"
	"time"
)

func TestRegistry_Get_CreatesLazily(t *testing.T) {
	r := NewRegistry(DefaultConfig())

	cb1 := r.Get("project-a")
	if cb1 == nil {
		t.Fatal("Get should return a non-nil CircuitBreaker")
	}
	if cb1.State() != StateClosed {
		t.Error("new breaker should be Closed")
	}
}

func TestRegistry_Get_ReturnsSameInstance(t *testing.T) {
	r := NewRegistry(DefaultConfig())

	cb1 := r.Get("project-a")
	cb2 := r.Get("project-a")

	if cb1 != cb2 {
		t.Error("Get should return the same instance for the same project")
	}
}

func TestRegistry_Get_DifferentProjectsIndependent(t *testing.T) {
	r := NewRegistry(Config{
		FailureThreshold: 1,
		OpenTimeout:      30 * time.Second,
	})

	// Trip project-a
	cbA := r.Get("project-a")
	cbA.Allow()
	cbA.RecordFailure()

	if cbA.State() != StateOpen {
		t.Fatal("project-a should be Open")
	}

	// project-b should still be Closed
	cbB := r.Get("project-b")
	if cbB.State() != StateClosed {
		t.Errorf("project-b state = %s, want Closed (independent)", cbB.State())
	}
}

func TestRegistry_Stats_ReturnsAll(t *testing.T) {
	r := NewRegistry(DefaultConfig())

	r.Get("project-a")
	r.Get("project-b")
	r.Get("project-c")

	stats := r.Stats()
	if len(stats) != 3 {
		t.Errorf("len(stats) = %d, want 3", len(stats))
	}

	for _, projectID := range []string{"project-a", "project-b", "project-c"} {
		if _, ok := stats[projectID]; !ok {
			t.Errorf("stats missing project: %s", projectID)
		}
	}
}

func TestRegistry_Reset_RemovesInstance(t *testing.T) {
	r := NewRegistry(DefaultConfig())

	cb1 := r.Get("project-a")
	cb1.Allow()
	cb1.RecordFailure()

	r.Reset("project-a")

	// After reset, a new instance should be created
	cb2 := r.Get("project-a")
	if cb2.State() != StateClosed {
		t.Errorf("reset breaker state = %s, want Closed", cb2.State())
	}
	if cb1 == cb2 {
		t.Error("Reset should create a new instance")
	}

	stats := r.Stats()
	if _, ok := stats["project-a"]; !ok {
		t.Error("project-a should reappear in stats after re-creation")
	}
}

func TestRegistry_ResetNonExistent(t *testing.T) {
	r := NewRegistry(DefaultConfig())
	// Should not panic
	r.Reset("nonexistent")
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := NewRegistry(DefaultConfig())

	var wg sync.WaitGroup
	goroutines := 20
	projectIDs := []string{"a", "b", "c", "d"}

	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(idx int) {
			defer wg.Done()
			projectID := projectIDs[idx%len(projectIDs)]
			for i := 0; i < 100; i++ {
				cb := r.Get(projectID)
				if cb.Allow() {
					if i%3 == 0 {
						cb.RecordFailure()
					} else {
						cb.RecordSuccess()
					}
				}
			}
		}(g)
	}

	wg.Wait()

	// Stats should contain exactly 4 projects
	stats := r.Stats()
	if len(stats) != 4 {
		t.Errorf("len(stats) = %d, want 4", len(stats))
	}
}

func TestRegistry_Stats_Empty(t *testing.T) {
	r := NewRegistry(DefaultConfig())
	stats := r.Stats()
	if len(stats) != 0 {
		t.Errorf("empty registry stats should be empty, got %d entries", len(stats))
	}
}
