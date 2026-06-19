package circuitbreaker

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/sirupsen/logrus"
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

type State int32

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

type Config struct {
	Name              string
	FailureThreshold  int
	SuccessThreshold  int
	OpenTimeout       time.Duration
	HalfOpenMaxCalls  int
	Logger            *logrus.Entry
}

func DefaultConfig(name string) Config {
	return Config{
		Name:              name,
		FailureThreshold:  3,
		SuccessThreshold:  2,
		OpenTimeout:       30 * time.Second,
		HalfOpenMaxCalls:  3,
		Logger:            logrus.WithField("component", "circuit_breaker"),
	}
}

type CircuitBreaker struct {
	config     Config
	state     atomic.Int32
	mu         sync.RWMutex
	failures   atomic.Int32
	successes  atomic.Int32
	lastFailure atomic.Value
	halfOpenCalls atomic.Int32
	openedAt   atomic.Value
}

func New(cfg Config) *CircuitBreaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 3
	}
	if cfg.SuccessThreshold <= 0 {
		cfg.SuccessThreshold = 2
	}
	if cfg.OpenTimeout <= 0 {
		cfg.OpenTimeout = 30 * time.Second
	}
	if cfg.HalfOpenMaxCalls <= 0 {
		cfg.HalfOpenMaxCalls = 3
	}
	if cfg.Logger == nil {
		cfg.Logger = logrus.WithField("component", "circuit_breaker")
	}

	cb := &CircuitBreaker{
		config: cfg,
	}
	cb.state.Store(int32(StateClosed))
	return cb
}

func (cb *CircuitBreaker) Allow() bool {
	for {
		switch State(cb.state.Load()) {
		case StateClosed:
			return true

		case StateOpen:
			if cb.shouldAttemptReset() {
				cb.toHalfOpen()
				return cb.Allow()
			}
			return false

		case StateHalfOpen:
			if cb.halfOpenCalls.Load() >= int32(cb.config.HalfOpenMaxCalls) {
				return false
			}
			cb.halfOpenCalls.Add(1)
			return true
		}
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch State(cb.state.Load()) {
	case StateClosed:
		cb.failures.Store(0)

	case StateHalfOpen:
		cb.successes.Add(1)
		if cb.successes.Load() >= int32(cb.config.SuccessThreshold) {
			cb.toClosedLocked()
		}
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailure.Store(time.Now())

	switch State(cb.state.Load()) {
	case StateClosed:
		cb.failures.Add(1)
		if cb.failures.Load() >= int32(cb.config.FailureThreshold) {
			cb.toOpenLocked()
		}

	case StateHalfOpen:
		cb.toOpenLocked()
	}
}

func (cb *CircuitBreaker) State() State {
	return State(cb.state.Load())
}

func (cb *CircuitBreaker) shouldAttemptReset() bool {
	openedAt, ok := cb.openedAt.Load().(time.Time)
	if !ok {
		return true
	}
	return time.Since(openedAt) >= cb.config.OpenTimeout
}

func (cb *CircuitBreaker) toHalfOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if State(cb.state.Load()) != StateOpen {
		return
	}

	cb.state.Store(int32(StateHalfOpen))
	cb.successes.Store(0)
	cb.halfOpenCalls.Store(0)
	cb.config.Logger.WithFields(logrus.Fields{
		"circuit":    cb.config.Name,
		"from_state": "open",
		"to_state":   "half-open",
	}).Info("Circuit breaker transitioning")
}

func (cb *CircuitBreaker) toOpenLocked() {
	cb.state.Store(int32(StateOpen))
	cb.openedAt.Store(time.Now())
	cb.failures.Store(0)
	cb.halfOpenCalls.Store(0)
	cb.config.Logger.WithFields(logrus.Fields{
		"circuit":    cb.config.Name,
		"from_state": State(cb.state.Load()).String(),
		"to_state":   "open",
	}).Warn("Circuit breaker opened")
}

func (cb *CircuitBreaker) toClosedLocked() {
	cb.state.Store(int32(StateClosed))
	cb.failures.Store(0)
	cb.successes.Store(0)
	cb.halfOpenCalls.Store(0)
	cb.config.Logger.WithFields(logrus.Fields{
		"circuit":    cb.config.Name,
		"from_state": "half-open",
		"to_state":   "closed",
	}).Info("Circuit breaker closed")
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.toClosedLocked()
}

func (cb *CircuitBreaker) String() string {
	return cb.config.Name
}

type metrics interface {
	RecordCircuitStateChange(name string, from, to string)
}

type noopMetrics struct{}

func (m *noopMetrics) RecordCircuitStateChange(name string, from, to string) {}

type prometheusMetrics struct{}

func (m *prometheusMetrics) RecordCircuitStateChange(name string, from, to string) {
	monitoring.RecordCircuitBreakerTransition(name, from, to)
}

func NewPrometheusMetrics() metrics {
	return &prometheusMetrics{}
}

type WithMetrics struct {
	*CircuitBreaker
	metrics metrics
}

func NewWithMetrics(cfg Config, m metrics) *WithMetrics {
	if m == nil {
		m = &prometheusMetrics{}
	}
	return &WithMetrics{
		CircuitBreaker: New(cfg),
		metrics:        m,
	}
}

func (cb *WithMetrics) toHalfOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if State(cb.state.Load()) != StateOpen {
		return
	}

	cb.state.Store(int32(StateHalfOpen))
	cb.successes.Store(0)
	cb.halfOpenCalls.Store(0)
	cb.metrics.RecordCircuitStateChange(cb.config.Name, "open", "half-open")
}

func (cb *WithMetrics) toOpenLocked() {
	oldState := State(cb.state.Load())
	cb.state.Store(int32(StateOpen))
	cb.openedAt.Store(time.Now())
	cb.failures.Store(0)
	cb.halfOpenCalls.Store(0)
	cb.metrics.RecordCircuitStateChange(cb.config.Name, oldState.String(), "open")
}

func (cb *WithMetrics) toClosedLocked() {
	oldState := State(cb.state.Load())
	cb.state.Store(int32(StateClosed))
	cb.failures.Store(0)
	cb.successes.Store(0)
	cb.halfOpenCalls.Store(0)
	cb.metrics.RecordCircuitStateChange(cb.config.Name, oldState.String(), "closed")
}
