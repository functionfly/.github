package circuitbreaker

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBreaker_BasicClosed(t *testing.T) {
	b := New("test", DefaultConfig())

	if b.State() != StateClosed {
		t.Fatalf("expected closed, got %s", b.State())
	}
	if !b.Allow() {
		t.Fatal("expected Allow() to return true in closed state")
	}
}

func TestBreaker_OpenAfterThreshold(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FailureThreshold = 3
	cfg.BaseCooldown = 100 * time.Millisecond
	b := New("test", cfg)

	for i := 0; i < 3; i++ {
		b.Record(errors.New("fail"))
	}

	if b.State() != StateOpen {
		t.Fatalf("expected open after 3 failures, got %s", b.State())
	}
	if b.Allow() {
		t.Fatal("expected Allow() to return false in open state")
	}
}

func TestBreaker_ProbeAllowInOpen(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FailureThreshold = 1
	cfg.BaseCooldown = 1 * time.Hour // Long cooldown so it stays open
	b := New("test", cfg)

	b.Record(errors.New("fail"))

	if !b.ProbeAllow() {
		t.Fatal("expected ProbeAllow() to return true even in open state")
	}
}

func TestBreaker_HalfOpenAfterCooldown(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FailureThreshold = 1
	cfg.BaseCooldown = 50 * time.Millisecond
	cfg.SuccessThreshold = 1
	b := New("test", cfg)

	b.Record(errors.New("fail"))
	if b.State() != StateOpen {
		t.Fatalf("expected open, got %s", b.State())
	}

	time.Sleep(60 * time.Millisecond)

	if !b.Allow() {
		t.Fatal("expected Allow() to return true after cooldown (half-open)")
	}
	if b.State() != StateHalfOpen {
		t.Fatalf("expected half-open, got %s", b.State())
	}
}

func TestBreaker_ClosedAfterSuccessInHalfOpen(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FailureThreshold = 1
	cfg.BaseCooldown = 50 * time.Millisecond
	cfg.SuccessThreshold = 2
	cfg.HalfOpenMaxRequests = 3
	b := New("test", cfg)

	b.Record(errors.New("fail"))
	time.Sleep(60 * time.Millisecond)

	b.Allow() // triggers half-open
	b.Record(nil)
	b.Record(nil)

	if b.State() != StateClosed {
		t.Fatalf("expected closed after 2 successes in half-open, got %s", b.State())
	}
}

func TestBreaker_ReopenOnFailureInHalfOpen(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FailureThreshold = 1
	cfg.BaseCooldown = 50 * time.Millisecond
	b := New("test", cfg)

	b.Record(errors.New("fail"))
	time.Sleep(60 * time.Millisecond)

	b.Allow() // triggers half-open
	b.Record(errors.New("fail again"))

	if b.State() != StateOpen {
		t.Fatalf("expected open after failure in half-open, got %s", b.State())
	}

	snap := b.Snapshot()
	if snap.ReopenCount != 1 {
		t.Fatalf("expected reopen count 1, got %d", snap.ReopenCount)
	}
}

func TestBreaker_ExponentialBackoff(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FailureThreshold = 1
	cfg.BaseCooldown = 100 * time.Millisecond
	cfg.BackoffMultiplier = 2.0
	cfg.MaxCooldown = 10 * time.Second
	b := New("test", cfg)

	// First open
	b.Record(errors.New("fail"))
	if b.currentCooldown() != 100*time.Millisecond {
		t.Fatalf("expected 100ms cooldown, got %s", b.currentCooldown())
	}

	// Simulate reopen by calling transition
	time.Sleep(110 * time.Millisecond)
	b.Allow() // half-open
	b.Record(errors.New("fail"))

	snap := b.Snapshot()
	if snap.ReopenCount != 1 {
		t.Fatalf("expected reopen count 1, got %d", snap.ReopenCount)
	}
}

func TestBreaker_HalfOpenMaxRequests(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FailureThreshold = 1
	cfg.BaseCooldown = 50 * time.Millisecond
	cfg.HalfOpenMaxRequests = 2
	cfg.SuccessThreshold = 2
	b := New("test", cfg)

	b.Record(errors.New("fail"))
	time.Sleep(60 * time.Millisecond)

	// First request allowed in half-open
	if !b.Allow() {
		t.Fatal("first half-open request should be allowed")
	}
	b.Record(nil) // success increments halfOpenCount to 1

	// Second request allowed
	if !b.Allow() {
		t.Fatal("second half-open request should be allowed")
	}
	b.Record(nil) // success increments halfOpenCount to 2, circuit should close

	// Circuit should be closed now (2 successes met threshold)
	if b.State() != StateClosed {
		t.Fatalf("expected closed after max successful half-open requests, got %s", b.State())
	}
}

func TestBreaker_Reset(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FailureThreshold = 1
	b := New("test", cfg)

	b.Record(errors.New("fail"))
	if b.State() != StateOpen {
		t.Fatalf("expected open, got %s", b.State())
	}

	b.Reset()
	if b.State() != StateClosed {
		t.Fatalf("expected closed after reset, got %s", b.State())
	}
}

func TestBreaker_RestoreState(t *testing.T) {
	cfg := DefaultConfig()
	b := New("test", cfg)

	b.RestoreState(StateOpen, 5, 2)

	snap := b.Snapshot()
	if snap.State != StateOpen {
		t.Fatalf("expected open, got %s", snap.State)
	}
	if snap.Failures != 5 {
		t.Fatalf("expected 5 failures, got %d", snap.Failures)
	}
	if snap.ReopenCount != 2 {
		t.Fatalf("expected 2 reopen count, got %d", snap.ReopenCount)
	}
}

func TestBreaker_OnStateChange(t *testing.T) {
	var transitions []string
	cfg := DefaultConfig()
	cfg.FailureThreshold = 1
	cfg.BaseCooldown = 50 * time.Millisecond
	cfg.SuccessThreshold = 1
	cfg.OnStateChange = func(key string, from, to State) {
		transitions = append(transitions, from.String()+"->"+to.String())
	}
	b := New("test", cfg)

	b.Record(errors.New("fail"))
	time.Sleep(10 * time.Millisecond) // let async callback run

	if len(transitions) != 1 || transitions[0] != "closed->open" {
		t.Fatalf("expected [closed->open], got %v", transitions)
	}
}

func TestManager_For(t *testing.T) {
	m := NewManager(DefaultConfig())

	b1 := m.For("key1")
	b2 := m.For("key1")
	if b1 != b2 {
		t.Fatal("expected same breaker instance for same key")
	}

	b3 := m.For("key2")
	if b1 == b3 {
		t.Fatal("expected different breaker instances for different keys")
	}
}

func TestManager_ForBackend(t *testing.T) {
	m := NewManager(DefaultConfig())

	b := m.ForBackend(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	if b == nil {
		t.Fatal("expected non-nil breaker")
	}
	if b.State() != StateClosed {
		t.Fatalf("expected closed, got %s", b.State())
	}
}

func TestManager_SnapshotAll(t *testing.T) {
	m := NewManager(DefaultConfig())

	m.For("a").Record(errors.New("fail"))
	m.For("b")

	snapshots := m.SnapshotAll()
	if len(snapshots) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snapshots))
	}
}
