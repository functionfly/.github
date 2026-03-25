package circuitbreaker

import (
	"errors"
	"testing"
	"time"
)

func TestBreaker_ClosedToOpen(t *testing.T) {
	b := New(Config{
		FailureThreshold:    3,
		SuccessThreshold:    2,
		CooldownDuration:    time.Second,
		HalfOpenMaxRequests: 1,
	})

	// Record failures
	for i := 0; i < 3; i++ {
		b.Record(errors.New("fail"))
	}

	if b.State() != StateOpen {
		t.Errorf("expected OPEN, got %v", b.State())
	}
}

func TestBreaker_HalfOpenToClosed(t *testing.T) {
	b := New(Config{
		FailureThreshold:    3,
		SuccessThreshold:    2,
		CooldownDuration:    time.Millisecond,
		HalfOpenMaxRequests: 2,
	})

	// Force to OPEN
	for i := 0; i < 3; i++ {
		b.Record(errors.New("fail"))
	}

	// Wait for cooldown
	time.Sleep(2 * time.Millisecond)

	// Record successes
	b.Record(nil)
	b.Record(nil)

	if b.State() != StateClosed {
		t.Errorf("expected CLOSED, got %v", b.State())
	}
}

func TestBreaker_HalfOpenToOpen(t *testing.T) {
	b := New(Config{
		FailureThreshold:    3,
		SuccessThreshold:    2,
		CooldownDuration:    time.Millisecond,
		HalfOpenMaxRequests: 2,
	})

	// Force to OPEN
	for i := 0; i < 3; i++ {
		b.Record(errors.New("fail"))
	}

	// Wait for cooldown
	time.Sleep(2 * time.Millisecond)

	// Record failure in half-open
	b.Record(errors.New("fail"))

	if b.State() != StateOpen {
		t.Errorf("expected OPEN, got %v", b.State())
	}
}

func TestBreaker_Execute(t *testing.T) {
	b := New(DefaultConfig())

	// Successful execution
	err := b.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Failed execution
	err = b.Execute(func() error {
		return errors.New("fail")
	})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestBreaker_ExecuteWhenOpen(t *testing.T) {
	b := New(Config{
		FailureThreshold:    1,
		SuccessThreshold:    2,
		CooldownDuration:    time.Second,
		HalfOpenMaxRequests: 1,
	})

	// Force to OPEN
	b.Record(errors.New("fail"))

	// Try to execute
	err := b.Execute(func() error {
		return nil
	})
	if err != ErrCircuitOpen {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestBreaker_Allow(t *testing.T) {
	b := New(DefaultConfig())

	// Should allow in CLOSED state
	if !b.Allow() {
		t.Error("expected Allow() to return true in CLOSED state")
	}
}

func TestBreaker_AllowWhenOpen(t *testing.T) {
	b := New(Config{
		FailureThreshold:    1,
		SuccessThreshold:    2,
		CooldownDuration:    time.Second,
		HalfOpenMaxRequests: 1,
	})

	// Force to OPEN
	b.Record(errors.New("fail"))

	// Should not allow in OPEN state
	if b.Allow() {
		t.Error("expected Allow() to return false in OPEN state")
	}
}

func TestBreaker_AllowAfterCooldown(t *testing.T) {
	b := New(Config{
		FailureThreshold:    1,
		SuccessThreshold:    2,
		CooldownDuration:    time.Millisecond,
		HalfOpenMaxRequests: 1,
	})

	// Force to OPEN
	b.Record(errors.New("fail"))

	// Wait for cooldown
	time.Sleep(2 * time.Millisecond)

	// Should allow after cooldown
	if !b.Allow() {
		t.Error("expected Allow() to return true after cooldown")
	}
}

func TestBreaker_Reset(t *testing.T) {
	b := New(Config{
		FailureThreshold:    1,
		SuccessThreshold:    2,
		CooldownDuration:    time.Second,
		HalfOpenMaxRequests: 1,
	})

	// Force to OPEN
	b.Record(errors.New("fail"))

	// Reset
	b.Reset()

	if b.State() != StateClosed {
		t.Errorf("expected CLOSED after reset, got %v", b.State())
	}
	if b.Failures() != 0 {
		t.Errorf("expected 0 failures after reset, got %d", b.Failures())
	}
}

func TestBreaker_OnStateChange(t *testing.T) {
	var fromState, toState State
	stateChanged := false

	b := New(Config{
		FailureThreshold:    1,
		SuccessThreshold:    2,
		CooldownDuration:    time.Second,
		HalfOpenMaxRequests: 1,
		OnStateChange: func(from, to State) {
			fromState = from
			toState = to
			stateChanged = true
		},
	})

	// Force to OPEN
	b.Record(errors.New("fail"))

	// Wait for goroutine
	time.Sleep(10 * time.Millisecond)

	if !stateChanged {
		t.Error("expected OnStateChange to be called")
	}
	if fromState != StateClosed {
		t.Errorf("expected from=CLOSED, got %v", fromState)
	}
	if toState != StateOpen {
		t.Errorf("expected to=OPEN, got %v", toState)
	}
}

func TestBreaker_HalfOpenMaxRequests(t *testing.T) {
	b := New(Config{
		FailureThreshold:    1,
		SuccessThreshold:    2,
		CooldownDuration:    time.Millisecond,
		HalfOpenMaxRequests: 1,
	})

	// Force to OPEN
	b.Record(errors.New("fail"))

	// Wait for cooldown
	time.Sleep(2 * time.Millisecond)

	// First request should be allowed
	if !b.Allow() {
		t.Error("expected first request to be allowed in HALF_OPEN")
	}

	// Second request should not be allowed
	if b.Allow() {
		t.Error("expected second request to be denied in HALF_OPEN")
	}
}

func TestBreaker_StateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
	}

	for _, tt := range tests {
		if tt.state.String() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.state.String())
		}
	}
}
