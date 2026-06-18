package client

import (
	"sync"
	"time"
)

// BreakerState represents the three states of a circuit breaker.
type BreakerState int

const (
	BreakerClosed   BreakerState = 0
	BreakerOpen     BreakerState = 1
	BreakerHalfOpen BreakerState = 2
)

// String returns the lowercase name used in Prometheus labels.
func (s BreakerState) String() string {
	switch s {
	case BreakerClosed:
		return "closed"
	case BreakerOpen:
		return "open"
	case BreakerHalfOpen:
		return "half_open"
	}
	return "unknown"
}

// CircuitBreaker implements the standard 5-consecutive-failures / 30s window
// breaker used by the orchestrator's engines_nocgo.go (lines 74–121).
//
// Behavior:
//   - Closed: calls pass through. Failures are counted. After `Threshold`
//     consecutive failures the breaker trips to Open.
//   - Open: calls are short-circuited (Allow() returns false) for `Window`,
//     after which the breaker moves to HalfOpen.
//   - HalfOpen: a single probe call is allowed. Success → Closed. Failure → Open.
type CircuitBreaker struct {
	Threshold int
	Window    time.Duration

	mu             sync.Mutex
	state          BreakerState
	failures       int
	openedAt       time.Time
	halfOpenInUse  bool
}

// NewCircuitBreaker returns a breaker with the plan's defaults: 5 / 30s.
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		Threshold: 5,
		Window:    30 * time.Second,
		state:     BreakerClosed,
	}
}

// State returns the current state. Thread-safe.
func (b *CircuitBreaker) State() BreakerState {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.currentStateLocked(time.Now())
}

func (b *CircuitBreaker) currentStateLocked(now time.Time) BreakerState {
	if b.state == BreakerOpen && now.Sub(b.openedAt) >= b.Window {
		b.state = BreakerHalfOpen
		b.halfOpenInUse = false
	}
	return b.state
}

// Allow returns true if the call should proceed. In HalfOpen, only the
// first call is allowed; concurrent callers are short-circuited.
func (b *CircuitBreaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.currentStateLocked(time.Now()) {
	case BreakerClosed:
		return true
	case BreakerOpen:
		return false
	case BreakerHalfOpen:
		if b.halfOpenInUse {
			return false
		}
		b.halfOpenInUse = true
		return true
	}
	return false
}

// OnSuccess records a successful call.
func (b *CircuitBreaker) OnSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = BreakerClosed
	b.halfOpenInUse = false
}

// OnFailure records a failed call.
func (b *CircuitBreaker) OnFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.state == BreakerHalfOpen {
		b.state = BreakerOpen
		b.openedAt = time.Now()
		b.halfOpenInUse = false
		return
	}
	if b.failures >= b.Threshold {
		b.state = BreakerOpen
		b.openedAt = time.Now()
	}
}
