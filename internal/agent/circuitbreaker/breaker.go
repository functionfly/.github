package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	// StateClosed allows requests to pass through
	StateClosed State = iota
	// StateOpen blocks all requests
	StateOpen
	// StateHalfOpen allows limited test requests
	StateHalfOpen
)

// String returns the string representation of the state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
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

// Breaker implements the circuit breaker pattern
type Breaker struct {
	mu            sync.RWMutex
	state         State
	failures      int
	successes     int
	lastFailure   time.Time
	halfOpenCount int
	config        Config
}

// New creates a new circuit breaker with the given configuration
func New(config Config) *Breaker {
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 5
	}
	if config.SuccessThreshold == 0 {
		config.SuccessThreshold = 2
	}
	if config.CooldownDuration == 0 {
		config.CooldownDuration = 30 * time.Second
	}
	if config.HalfOpenMaxRequests == 0 {
		config.HalfOpenMaxRequests = 1
	}

	return &Breaker{
		state:  StateClosed,
		config: config,
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
	b.mu.RLock()
	defer b.mu.RUnlock()

	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if cooldown has elapsed
		if time.Since(b.lastFailure) > b.config.CooldownDuration {
			// Transition to half-open
			b.mu.RUnlock()
			b.mu.Lock()
			b.transitionTo(StateHalfOpen)
			b.mu.Unlock()
			b.mu.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		return b.halfOpenCount < b.config.HalfOpenMaxRequests
	}
	return false
}

// Record records the result of a request
func (b *Breaker) Record(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		b.recordFailure()
	} else {
		b.recordSuccess()
	}
}

func (b *Breaker) recordFailure() {
	b.failures++
	b.successes = 0
	b.lastFailure = time.Now()

	switch b.state {
	case StateClosed:
		if b.failures >= b.config.FailureThreshold {
			b.transitionTo(StateOpen)
		}
	case StateHalfOpen:
		// Any failure in half-open transitions back to open
		b.transitionTo(StateOpen)
	}
}

func (b *Breaker) recordSuccess() {
	b.successes++
	b.failures = 0

	switch b.state {
	case StateHalfOpen:
		b.halfOpenCount++
		if b.successes >= b.config.SuccessThreshold {
			b.transitionTo(StateClosed)
		}
	}
}

func (b *Breaker) transitionTo(newState State) {
	oldState := b.state
	b.state = newState
	b.failures = 0
	b.successes = 0
	b.halfOpenCount = 0

	if b.config.OnStateChange != nil {
		go b.config.OnStateChange(oldState, newState)
	}
}

// State returns the current state of the circuit breaker
func (b *Breaker) State() State {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.state
}

// Failures returns the current failure count
func (b *Breaker) Failures() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.failures
}

// Successes returns the current success count
func (b *Breaker) Successes() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.successes
}

// Reset resets the circuit breaker to its initial state
func (b *Breaker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.state = StateClosed
	b.failures = 0
	b.successes = 0
	b.halfOpenCount = 0
}
