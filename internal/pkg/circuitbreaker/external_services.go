package circuitbreaker

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	ErrServiceUnavailable = errors.New("external service unavailable")
)

// ExternalServiceCircuitBreaker wraps circuit breaker logic for external service calls
type ExternalServiceCircuitBreaker struct {
	name       string
	cb         *WithMetrics
	serviceName string
}

// ExternalCircuitBreakerManager manages circuit breakers for all external dependencies
type ExternalCircuitBreakerManager struct {
	mu           sync.RWMutex
	breakers     map[string]*ExternalServiceCircuitBreaker
	redisBreaker *ExternalServiceCircuitBreaker
	postgresBreaker *ExternalServiceCircuitBreaker
	aiServiceBreaker *ExternalServiceCircuitBreaker
	httpClient   *http.Client
	logger       *logrus.Entry
}

// NewExternalCircuitBreakerManager creates a manager for external service circuit breakers
func NewExternalCircuitBreakerManager() *ExternalCircuitBreakerManager {
	m := &ExternalCircuitBreakerManager{
		breakers:   make(map[string]*ExternalServiceCircuitBreaker),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logrus.WithField("component", "external_circuit_breaker"),
	}

	// Initialize with default circuit breakers for critical services
	m.redisBreaker = m.GetOrCreate("redis", CircuitBreakerConfig{
		FailureThreshold:  5,
		SuccessThreshold:  2,
		OpenTimeout:       30 * time.Second,
		HalfOpenMaxCalls:  3,
	})

	m.postgresBreaker = m.GetOrCreate("postgres", CircuitBreakerConfig{
		FailureThreshold:  3,
		SuccessThreshold:  2,
		OpenTimeout:       30 * time.Second,
		HalfOpenMaxCalls:  3,
	})

	m.aiServiceBreaker = m.GetOrCreate("ai_service", CircuitBreakerConfig{
		FailureThreshold:  3,
		SuccessThreshold:  2,
		OpenTimeout:       60 * time.Second,
		HalfOpenMaxCalls:  2,
	})

	return m
}

// CircuitBreakerConfig holds configuration for a circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold  int
	SuccessThreshold  int
	OpenTimeout       time.Duration
	HalfOpenMaxCalls  int
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold:  3,
		SuccessThreshold:  2,
		OpenTimeout:       30 * time.Second,
		HalfOpenMaxCalls:  3,
	}
}

// GetOrCreate gets an existing circuit breaker or creates a new one
func (m *ExternalCircuitBreakerManager) GetOrCreate(name string, config CircuitBreakerConfig) *ExternalServiceCircuitBreaker {
	m.mu.RLock()
	if cb, exists := m.breakers[name]; exists {
		m.mu.RUnlock()
		return cb
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, exists := m.breakers[name]; exists {
		return cb
	}

	cfg := Config{
		Name:              name,
		FailureThreshold:  config.FailureThreshold,
		SuccessThreshold:  config.SuccessThreshold,
		OpenTimeout:       config.OpenTimeout,
		HalfOpenMaxCalls:  config.HalfOpenMaxCalls,
		Logger:            m.logger.WithField("circuit", name),
	}

	cb := &ExternalServiceCircuitBreaker{
		name:       name,
		cb:         NewWithMetrics(cfg, nil),
		serviceName: name,
	}

	m.breakers[name] = cb
	return cb
}

// Execute runs a function with circuit breaker protection
func (cb *ExternalServiceCircuitBreaker) Execute(ctx context.Context, operation func() error) error {
	if !cb.cb.Allow() {
		cb.cb.config.Logger.WithFields(logrus.Fields{
			"service": cb.serviceName,
		}).Warn("Circuit breaker open, rejecting request")
		return ErrCircuitOpen
	}

	err := operation()

	if err != nil {
		cb.cb.RecordFailure()
		cb.cb.config.Logger.WithFields(logrus.Fields{
			"service": cb.serviceName,
			"error":   err.Error(),
		}).Warn("Circuit breaker recorded failure")
	} else {
		cb.cb.RecordSuccess()
	}

	return err
}

// ExecuteWithFallback runs operation with circuit breaker, falls back to fallback on open circuit
func (cb *ExternalServiceCircuitBreaker) ExecuteWithFallback(ctx context.Context, operation func() error, fallback func() error) error {
	if !cb.cb.Allow() {
		cb.cb.config.Logger.WithFields(logrus.Fields{
			"service": cb.serviceName,
		}).Warn("Circuit breaker open, using fallback")
		if fallback != nil {
			return fallback()
		}
		return ErrCircuitOpen
	}

	err := operation()

	if err != nil {
		cb.cb.RecordFailure()
		if cb.cb.State() == StateOpen {
			if fallback != nil {
				cb.cb.config.Logger.WithFields(logrus.Fields{
					"service": cb.serviceName,
				}).Info("Circuit breaker open, falling back")
				return fallback()
			}
		}
		return err
	}

	cb.cb.RecordSuccess()
	return nil
}

// State returns the current state of a circuit breaker
func (m *ExternalCircuitBreakerManager) State(name string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if cb, exists := m.breakers[name]; exists {
		return cb.cb.State().String()
	}
	return "unknown"
}

// GetBreakerStates returns the state of all circuit breakers
func (m *ExternalCircuitBreakerManager) GetBreakerStates() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make(map[string]string)
	for name, cb := range m.breakers {
		states[name] = cb.cb.State().String()
	}
	return states
}

// GetRedisBreaker returns the circuit breaker for Redis
func (m *ExternalCircuitBreakerManager) GetRedisBreaker() *ExternalServiceCircuitBreaker {
	return m.redisBreaker
}

// GetPostgresBreaker returns the circuit breaker for PostgreSQL
func (m *ExternalCircuitBreakerManager) GetPostgresBreaker() *ExternalServiceCircuitBreaker {
	return m.postgresBreaker
}

// GetAIServiceBreaker returns the circuit breaker for AI service
func (m *ExternalCircuitBreakerManager) GetAIServiceBreaker() *ExternalServiceCircuitBreaker {
	return m.aiServiceBreaker
}

// HealthCheckCircuitBreaker is a circuit breaker for health check dependencies
type HealthCheckCircuitBreaker struct {
	*CircuitBreaker
	name string
}

// Global circuit breaker manager instance
var globalCBManager *ExternalCircuitBreakerManager
var globalCBOnce sync.Once

// GetGlobalCircuitBreakerManager returns the global circuit breaker manager
func GetGlobalCircuitBreakerManager() *ExternalCircuitBreakerManager {
	globalCBOnce.Do(func() {
		globalCBManager = NewExternalCircuitBreakerManager()
	})
	return globalCBManager
}
