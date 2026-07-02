package dna

import (
	"time"

	"github.com/functionfly/functionfly/internal/circuitbreaker"
)

// circuitBreaker is a thin wrapper around the shared circuitbreaker.Breaker,
// maintaining the DNA-specific metrics integration (SetCircuitBreakerState, etc.).
type circuitBreaker struct {
	inner *circuitbreaker.Breaker
}

func newCircuitBreaker(threshold int, cooldown time.Duration) *circuitBreaker {
	cfg := circuitbreaker.Config{
		FailureThreshold:    threshold,
		SuccessThreshold:    1,
		BaseCooldown:        cooldown,
		MaxCooldown:         cooldown,
		BackoffMultiplier:   1.0,
		HalfOpenMaxRequests: 1,
		OnStateChange: func(_ string, _ circuitbreaker.State, to circuitbreaker.State) {
			SetCircuitBreakerState(float64(to))
		},
	}
	return &circuitBreaker{
		inner: circuitbreaker.New("dna-ai", cfg),
	}
}

func (cb *circuitBreaker) allow() bool {
	state := cb.inner.State()
	switch state {
	case circuitbreaker.StateClosed:
		SetCircuitBreakerState(0)
	case circuitbreaker.StateOpen:
		SetCircuitBreakerState(1)
	case circuitbreaker.StateHalfOpen:
		SetCircuitBreakerState(2)
	}
	return cb.inner.Allow()
}

func (cb *circuitBreaker) recordSuccess() {
	cb.inner.RecordSuccess()
	SetCircuitBreakerState(0)
	SetCircuitBreakerSuccesses(1)
	SetCircuitBreakerFailures(0)
}

func (cb *circuitBreaker) recordFailure() {
	cb.inner.RecordFailure()
	snap := cb.inner.Snapshot()
	SetCircuitBreakerFailures(float64(snap.Failures))
}

func (cb *circuitBreaker) GetState() int {
	return int(cb.inner.State())
}
