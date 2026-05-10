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
		return true
	case stateOpen:
		if time.Since(cb.lastFailure) > cb.cooldown {
			cb.state = stateHalfOpen
			return true
		}
		return false
	case stateHalfOpen:
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
}

// recordFailure increments the failure counter and trips the breaker if threshold is reached.
func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.threshold {
		cb.state = stateOpen
	}
}
