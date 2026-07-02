package circuitbreaker

import (
	"errors"
	"math"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	// StateClosed allows requests to pass through (normal operation).
	StateClosed State = iota
	// StateOpen blocks all requests (failing fast).
	StateOpen
	// StateHalfOpen allows limited test requests (testing recovery).
	StateHalfOpen
)

// String returns the string representation of the state.
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

// StateFromInt converts an int (0=closed, 1=open, 2=half-open) to State.
func StateFromInt(i int) State {
	switch i {
	case 0:
		return StateClosed
	case 1:
		return StateOpen
	case 2:
		return StateHalfOpen
	default:
		return StateClosed
	}
}

var (
	// ErrCircuitOpen is returned when the circuit breaker is open.
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// StateInfo captures the current state of a breaker for inspection.
type StateInfo struct {
	State       State
	Failures    int
	Successes   int
	ReopenCount int
	LastFailure time.Time
	Since       time.Time
}

// Breaker implements the circuit breaker pattern with exponential backoff.
type Breaker struct {
	mu            sync.Mutex
	key           string
	state         State
	failures      int
	successes     int
	halfOpenCount int
	reopenCount   int
	lastFailure   time.Time
	since         time.Time
	config        Config
}

// New creates a new circuit breaker with the given key and configuration.
func New(key string, config Config) *Breaker {
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 3
	}
	if config.SuccessThreshold <= 0 {
		config.SuccessThreshold = 2
	}
	if config.BaseCooldown <= 0 {
		config.BaseCooldown = 30 * time.Second
	}
	if config.MaxCooldown <= 0 {
		config.MaxCooldown = 5 * time.Minute
	}
	if config.BackoffMultiplier <= 0 {
		config.BackoffMultiplier = 2.0
	}
	if config.HalfOpenMaxRequests <= 0 {
		config.HalfOpenMaxRequests = 3
	}

	return &Breaker{
		key:    key,
		state:  StateClosed,
		since:  time.Now(),
		config: config,
	}
}

// Allow returns true if the circuit breaker allows a request (for routing/gating).
// Returns false when the circuit is open and cooldown has not elapsed.
func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(b.lastFailure) > b.currentCooldown() {
			b.transitionToLocked(StateHalfOpen)
			return true
		}
		return false
	case StateHalfOpen:
		return b.halfOpenCount < b.config.HalfOpenMaxRequests
	}
	return false
}

// ProbeAllow always returns true. Use this for health monitoring probes that
// need to continue even when the circuit is open, without counting against
// the half-open request limit.
func (b *Breaker) ProbeAllow() bool {
	return true
}

// Record records the result of a request. A non-nil error is a failure.
func (b *Breaker) Record(err error) {
	if err != nil {
		b.recordFailure()
	} else {
		b.recordSuccess()
	}
}

// RecordSuccess records a successful request.
func (b *Breaker) RecordSuccess() {
	b.recordSuccess()
}

// RecordFailure records a failed request.
func (b *Breaker) RecordFailure() {
	b.recordFailure()
}

func (b *Breaker) recordFailure() {
	b.mu.Lock()
	oldState := b.state
	b.failures++
	b.successes = 0
	b.lastFailure = time.Now()

	switch b.state {
	case StateClosed:
		if b.failures >= b.config.FailureThreshold {
			b.transitionToLocked(StateOpen)
		}
	case StateHalfOpen:
		// Any failure in half-open transitions back to open with incremented reopen count.
		b.reopenCount++
		b.transitionToLocked(StateOpen)
	}
	newState := b.state
	b.mu.Unlock()

	if oldState != newState && b.config.OnStateChange != nil {
		go b.config.OnStateChange(b.key, oldState, newState)
	}
}

func (b *Breaker) recordSuccess() {
	b.mu.Lock()
	oldState := b.state
	b.successes++
	b.failures = 0

	switch b.state {
	case StateHalfOpen:
		b.halfOpenCount++
		if b.successes >= b.config.SuccessThreshold {
			b.transitionToLocked(StateClosed)
		}
	}
	newState := b.state
	b.mu.Unlock()

	if oldState != newState && b.config.OnStateChange != nil {
		go b.config.OnStateChange(b.key, oldState, newState)
	}
}

// transitionToLocked must be called with b.mu held.
func (b *Breaker) transitionToLocked(newState State) {
	b.state = newState
	b.failures = 0
	b.successes = 0
	b.halfOpenCount = 0
	b.since = time.Now()

	if newState == StateClosed {
		b.reopenCount = 0
	}
}

// currentCooldown calculates the cooldown duration with exponential backoff.
// Must be called with b.mu held.
func (b *Breaker) currentCooldown() time.Duration {
	cooldown := time.Duration(float64(b.config.BaseCooldown) * math.Pow(b.config.BackoffMultiplier, float64(b.reopenCount)))
	if cooldown > b.config.MaxCooldown {
		cooldown = b.config.MaxCooldown
	}
	return cooldown
}

// State returns the current state of the circuit breaker.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// Snapshot returns a point-in-time view of the breaker state.
func (b *Breaker) Snapshot() StateInfo {
	b.mu.Lock()
	defer b.mu.Unlock()
	return StateInfo{
		State:       b.state,
		Failures:    b.failures,
		Successes:   b.successes,
		ReopenCount: b.reopenCount,
		LastFailure: b.lastFailure,
		Since:       b.since,
	}
}

// Reset resets the circuit breaker to closed state.
func (b *Breaker) Reset() {
	b.mu.Lock()
	oldState := b.state
	b.state = StateClosed
	b.failures = 0
	b.successes = 0
	b.halfOpenCount = 0
	b.reopenCount = 0
	b.since = time.Now()
	b.mu.Unlock()

	if oldState != StateClosed && b.config.OnStateChange != nil {
		go b.config.OnStateChange(b.key, oldState, StateClosed)
	}
}

// RestoreState sets the breaker state from a persisted value.
// Used during initialization to restore state from DB.
func (b *Breaker) RestoreState(state State, failCount, reopenCount int) {
	b.mu.Lock()
	b.state = state
	b.failures = failCount
	b.reopenCount = reopenCount
	b.since = time.Now()
	b.mu.Unlock()
}
