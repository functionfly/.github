package circuitbreaker

import (
	"sync"

	"github.com/google/uuid"
)

// Manager manages a collection of circuit breakers keyed by string.
// Use For(), ForBackend(), or ForProvider() to get or create breakers.
type Manager struct {
	breakers sync.Map // map[string]*Breaker
	config   Config
}

// NewManager creates a new breaker manager with the given configuration.
func NewManager(config Config) *Manager {
	return &Manager{
		config: config,
	}
}

// NewManagerFromEnv creates a new breaker manager with configuration from environment variables.
func NewManagerFromEnv() *Manager {
	return NewManager(ConfigFromEnv())
}

// For returns the circuit breaker for the given key, creating one if it doesn't exist.
func (m *Manager) For(key string) *Breaker {
	if v, ok := m.breakers.Load(key); ok {
		return v.(*Breaker)
	}
	breaker := New(key, m.config)
	actual, _ := m.breakers.LoadOrStore(key, breaker)
	return actual.(*Breaker)
}

// ForBackend returns the circuit breaker for a backend UUID.
func (m *Manager) ForBackend(id uuid.UUID) *Breaker {
	return m.For("backend:" + id.String())
}

// ForProvider returns the circuit breaker for a provider name (e.g., "wasm:python").
func (m *Manager) ForProvider(provider string) *Breaker {
	return m.For("provider:" + provider)
}

// ForBackendWithProvider returns the circuit breaker for a backend UUID with provider context.
// This is useful when the same backend ID might be used across different provider contexts.
func (m *Manager) ForBackendWithProvider(id uuid.UUID, provider string) *Breaker {
	return m.For(provider + ":" + id.String())
}

// SnapshotAll returns a map of all breaker keys to their current state.
func (m *Manager) SnapshotAll() map[string]StateInfo {
	snapshots := make(map[string]StateInfo)
	m.breakers.Range(func(key, value any) bool {
		snapshots[key.(string)] = value.(*Breaker).Snapshot()
		return true
	})
	return snapshots
}

// ForEach iterates over all breakers, calling fn for each.
// If fn returns false, iteration stops.
func (m *Manager) ForEach(fn func(key string, breaker *Breaker) bool) {
	m.breakers.Range(func(key, value any) bool {
		return fn(key.(string), value.(*Breaker))
	})
}
