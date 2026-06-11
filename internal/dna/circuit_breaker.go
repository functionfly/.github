package dna

import (
	"sync"
	"time"
)

const (
	stateClosed   = iota // normal operation
	stateOpen            // failing fast
	stateHalfOpen        // testing recovery
)

// circuitBreaker prevents cascading failures by short-circuiting calls to a
// failing service. After `threshold` consecutive failures the breaker opens
// and rejects calls for `cooldown`. A single success in half-open resets it.
type circuitBreaker struct {
	mu          sync.Mutex
	state       int
	failures    int
	threshold   int
	cooldown    time.Duration
	lastFailure time.Time
}

func newCircuitBreaker(threshold int, cooldown time.Duration) *circuitBreaker {
	return &circuitBreaker{
		state:     stateClosed,
		threshold: threshold,
		cooldown:  cooldown,
	}
}

// allow returns true if the call should proceed.
func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case stateClosed:
		SetCircuitBreakerState(0)
		return true
	case stateOpen:
		SetCircuitBreakerState(1)
		if time.Since(cb.lastFailure) > cb.cooldown {
			cb.state = stateHalfOpen
			SetCircuitBreakerState(2)
			return true
		}
		return false
	case stateHalfOpen:
		SetCircuitBreakerState(2)
		return true
	default:
		return true
	}
}

// recordSuccess resets the breaker to closed state.
func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = stateClosed
	SetCircuitBreakerState(0)
	SetCircuitBreakerSuccesses(1) // increment success counter for monitoring
	SetCircuitBreakerFailures(0)  // reset failure counter
}

// recordFailure increments the failure counter and trips the breaker if threshold is reached.
func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	SetCircuitBreakerFailures(float64(cb.failures)) // expose failure count for early warning
	if cb.failures >= cb.threshold {
		cb.state = stateOpen
		SetCircuitBreakerState(1)
	}
}

// GetState returns the current circuit breaker state for health checks.
func (cb *circuitBreaker) GetState() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
