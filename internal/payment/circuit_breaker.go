package payment

import (
	"sync/atomic"
	"time"
)

type CircuitState int32

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

const (
	circuitClosed int32 = 0
	circuitOpen   int32 = 1
	circuitHalfOpen int32 = 2
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

type CircuitBreaker struct {
	failures     atomic.Int64
	threshold   int64
	resetAfter  time.Duration
	state      atomic.Int32
	lastFailure atomic.Int64
}

func NewCircuitBreaker(threshold int64, resetAfter time.Duration) *CircuitBreaker {
	cb := &CircuitBreaker{
		threshold:  threshold,
		resetAfter: resetAfter,
	}
	cb.state.Store(circuitClosed)
	return cb
}

func (cb *CircuitBreaker) Allow() bool {
	switch cb.state.Load() {
	case circuitClosed:
		return true
	case circuitOpen:
		if time.Now().Unix()-cb.lastFailure.Load() > int64(cb.resetAfter.Seconds()) {
			cb.state.Store(circuitHalfOpen)
			return true
		}
		return false
	case circuitHalfOpen:
		return true
	}
	return false
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.failures.Add(1)
	cb.lastFailure.Store(time.Now().Unix())
	if cb.failures.Load() >= cb.threshold {
		cb.state.Store(circuitOpen)
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.failures.Store(0)
	cb.state.Store(circuitClosed)
}

func (cb *CircuitBreaker) State() CircuitState {
	return CircuitState(cb.state.Load())
}

func (cb *CircuitBreaker) FailureCount() int64 {
	return cb.failures.Load()
}

func (cb *CircuitBreaker) Reset() {
	cb.failures.Store(0)
	cb.state.Store(circuitClosed)
}
