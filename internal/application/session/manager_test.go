package session

import (
	"testing"
	"time"
)

func TestBroadcast_DeliversToSessionWithActiveSSECh(t *testing.T) {
	m := NewManager(0)
	s := m.Create("client-1", []string{"tool1"})
	s.SSECh = make(chan []byte, 16)

	msg := []byte(`{"jsonrpc":"2.0","method":"notifications/tools/list_changed"}`)
	m.Broadcast(msg)

	select {
	case received := <-s.SSECh:
		if string(received) != string(msg) {
			t.Errorf("received = %s, want %s", received, msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected message on SSE channel, got none")
	}
}

func TestBroadcast_SkipsNilSSECh(t *testing.T) {
	m := NewManager(0)
	s := m.Create("client-1", []string{"tool1"})
	// SSECh is nil by default (not set)

	msg := []byte(`{"jsonrpc":"2.0","method":"notifications/tools/list_changed"}`)
	// Should not panic
	m.Broadcast(msg)

	// Verify session still exists and is healthy
	got, ok := m.Get(s.ID)
	if !ok {
		t.Fatal("session should still exist")
	}
	if got.SSECh != nil {
		t.Error("SSECh should remain nil")
	}
}

func TestBroadcast_NonBlockingWhenChannelFull(t *testing.T) {
	m := NewManager(0)
	s := m.Create("client-1", []string{"tool1"})
	// 容量为 1 的通道，先填满
	s.SSECh = make(chan []byte, 1)
	s.SSECh <- []byte("existing")

	msg := []byte(`{"jsonrpc":"2.0","method":"notifications/tools/list_changed"}`)

	// 不应阻塞，不应 panic
	done := make(chan struct{})
	go func() {
		m.Broadcast(msg)
		close(done)
	}()

	select {
	case <-done:
		// OK: Broadcast 立即返回
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Broadcast blocked on full channel")
	}

	// 通道中应仍然是旧消息
	select {
	case received := <-s.SSECh:
		if string(received) != "existing" {
			t.Errorf("unexpected message: %s", received)
		}
	default:
		t.Fatal("expected existing message in channel")
	}
}

func TestBroadcast_NoSessions(t *testing.T) {
	m := NewManager(0)
	// No sessions created

	msg := []byte(`{"jsonrpc":"2.0","method":"notifications/tools/list_changed"}`)
	// Should not panic
	m.Broadcast(msg)
}

func TestBroadcast_DeliversToMultipleSessions(t *testing.T) {
	m := NewManager(0)

	s1 := m.Create("client-1", []string{"tool1"})
	s1.SSECh = make(chan []byte, 16)
	s2 := m.Create("client-2", []string{"tool1", "tool2"})
	s2.SSECh = make(chan []byte, 16)
	_ = m.Create("client-3", []string{"tool3"})
	// s3 has nil SSECh — should be skipped

	msg := []byte(`{"jsonrpc":"2.0","method":"notifications/tools/list_changed"}`)
	m.Broadcast(msg)

	// Both s1 and s2 should receive
	for i, ch := range []chan []byte{s1.SSECh, s2.SSECh} {
		select {
		case received := <-ch:
			if string(received) != string(msg) {
				t.Errorf("session %d: received = %s, want %s", i+1, received, msg)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("session %d: expected message, got none", i+1)
		}
	}

	// s3's SSECh is nil, no message delivered (verified by not panicking)
}

func TestBroadcast_EmptyData(t *testing.T) {
	m := NewManager(0)
	s := m.Create("client-1", []string{"tool1"})
	s.SSECh = make(chan []byte, 16)

	m.Broadcast([]byte{})

	select {
	case received := <-s.SSECh:
		if len(received) != 0 {
			t.Errorf("expected empty message, got %s", received)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected empty message on SSE channel, got none")
	}
}
