// Package circuitbreaker provides a circuit breaker implementation for agent operations.
// This is a thin wrapper around the shared internal/circuitbreaker package,
// maintaining backward compatibility for agent subsystems.
package circuitbreaker

import (
	"time"

	cb "github.com/functionfly/functionfly/internal/circuitbreaker"
)

// State represents the circuit breaker state
type State = cb.State

const (
	// StateClosed allows requests to pass through
	StateClosed = cb.StateClosed
	// StateOpen blocks all requests
	StateOpen = cb.StateOpen
	// StateHalfOpen allows limited test requests
	StateHalfOpen = cb.StateHalfOpen
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = cb.ErrCircuitOpen
)

// Config holds circuit breaker configuration
type Config struct {
	// FailureThreshold is the number of failures before opening the circuit
	FailureThreshold int
	// SuccessThreshold is the number of successes before closing the circuit
	SuccessThreshold int
	// CooldownDuration is how long to wait in OPEN before transitioning to HALF_OPEN
	CooldownDuration time.Duration
	// HalfOpenMaxRequests is the maximum number of requests allowed in HALF_OPEN state
	HalfOpenMaxRequests int
	// OnStateChange is called when the state changes
	OnStateChange func(from, to State)
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() Config {
	return Config{
		FailureThreshold:    5,
		SuccessThreshold:    2,
		CooldownDuration:    30 * time.Second,
		HalfOpenMaxRequests: 1,
	}
}

// Breaker implements the circuit breaker pattern using the shared implementation.
type Breaker struct {
	inner *cb.Breaker
}

// New creates a new circuit breaker with the given configuration
func New(config Config) *Breaker {
	sharedConfig := cb.Config{
		FailureThreshold:    config.FailureThreshold,
		SuccessThreshold:    config.SuccessThreshold,
		BaseCooldown:        config.CooldownDuration,
		MaxCooldown:         config.CooldownDuration, // No backoff in agent context
		BackoffMultiplier:   1.0,                      // No backoff
		HalfOpenMaxRequests: config.HalfOpenMaxRequests,
	}
	if config.OnStateChange != nil {
		sharedConfig.OnStateChange = func(_ string, from, to cb.State) {
			config.OnStateChange(from, to)
		}
	}

	return &Breaker{
		inner: cb.New("agent", sharedConfig),
	}
}

// Execute runs the given function if the circuit breaker allows it
func (b *Breaker) Execute(fn func() error) error {
	if !b.Allow() {
		return ErrCircuitOpen
	}

	err := fn()
	b.Record(err)
	return err
}

// Allow returns true if the circuit breaker allows a request
func (b *Breaker) Allow() bool {
	return b.inner.Allow()
}

// Record records the result of a request
func (b *Breaker) Record(err error) {
	b.inner.Record(err)
}

// State returns the current state of the circuit breaker
func (b *Breaker) State() State {
	return b.inner.State()
}

// Failures returns the current failure count
func (b *Breaker) Failures() int {
	return b.inner.Snapshot().Failures
}

// Successes returns the current success count
func (b *Breaker) Successes() int {
	return b.inner.Snapshot().Successes
}

// Reset resets the circuit breaker to its initial state
func (b *Breaker) Reset() {
	b.inner.Reset()
}
