package routing

import (
	"context"
	"fmt"
	"testing"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// MockRepository provides a simple mock implementation for benchmarking
type MockRepository struct {
	backends    map[uuid.UUID][]*storage.Backend
	healthData  map[uuid.UUID][]*storage.HealthCheck
	circuitData map[uuid.UUID]*storage.CircuitState
}

// Ensure MockRepository implements RouterRepository
var _ RouterRepository = (*MockRepository)(nil)

func NewMockRepository() *MockRepository {
	return &MockRepository{
		backends:    make(map[uuid.UUID][]*storage.Backend),
		healthData:  make(map[uuid.UUID][]*storage.HealthCheck),
		circuitData: make(map[uuid.UUID]*storage.CircuitState),
	}
}

func (m *MockRepository) ListBackendsByAppID(ctx context.Context, appID uuid.UUID) ([]*storage.Backend, error) {
	return m.backends[appID], nil
}

func (m *MockRepository) GetRecentHealthChecks(ctx context.Context, backendID uuid.UUID, limit int) ([]*storage.HealthCheck, error) {
	checks := m.healthData[backendID]
	if len(checks) > limit {
		return checks[:limit], nil
	}
	return checks, nil
}

func (m *MockRepository) GetCircuitState(ctx context.Context, backendID uuid.UUID) (*storage.CircuitState, error) {
	if state, exists := m.circuitData[backendID]; exists {
		return state, nil
	}
	// Default to closed state
	return &storage.CircuitState{State: "closed"}, nil
}

func (m *MockRepository) InsertRoutingEvent(ctx context.Context, appID, backendID uuid.UUID, latencyMs int, outcome, requestID string) error {
	// Mock implementation - just return success
	return nil
}

func (m *MockRepository) GetRecentRoutingEventsByBackend(ctx context.Context, backendID uuid.UUID, limit int) ([]*storage.RoutingEvent, error) {
	return nil, nil
}

// setupBenchmarkRouter creates a router with mock storage for benchmarking
func setupBenchmarkRouter(b *testing.B) (*Router, func()) {
	mockRepo := NewMockRepository()

	// Create test app ID
	appID := uuid.New()

	// Create multiple backends for realistic testing
	backends := make([]*storage.Backend, 10)
	for i := 0; i < 10; i++ {
		var priority *int
		priorities := []*int{&[]int{1}[0], &[]int{2}[0], &[]int{3}[0], nil}
		priority = priorities[i%4]
		backend := &storage.Backend{
			ID:       uuid.New(),
			AppID:    appID,
			Provider: "cloudflare",
			Region:   "us-east-1",
			Priority: priority, // Mix of priorities and nulls
		}
		backends[i] = backend

		// Add health check data
		healthChecks := make([]*storage.HealthCheck, 10)
		for j := 0; j < 10; j++ {
			healthChecks[j] = &storage.HealthCheck{
				BackendID: backend.ID,
				OK:        j < 8, // 80% success rate
				LatencyMs: 50 + (j * 10),
			}
		}
		mockRepo.healthData[backend.ID] = healthChecks

		// Set circuit state
		state := "closed"
		if i%5 == 0 { // 20% have open circuits
			state = "open"
		}
		mockRepo.circuitData[backend.ID] = &storage.CircuitState{State: state}
	}

	mockRepo.backends[appID] = backends

	router := NewRouter(mockRepo)
	return router, func() {} // No cleanup needed for mock
}

// BenchmarkSelectBackendSingle benchmarks backend selection for a single request
func BenchmarkSelectBackendSingle(b *testing.B) {
	router, cleanup := setupBenchmarkRouter(b)
	defer cleanup()

	// Get the app ID from the mock repository
	mockRepo := router.repo.(*MockRepository)
	var appID uuid.UUID
	for id := range mockRepo.backends {
		appID = id
		break
	}

	requestID := "bench-request-1"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := router.SelectBackend(appID, "GET", requestID, "starter")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkSelectBackendWithHealthChecks benchmarks backend selection with health check data
func BenchmarkSelectBackendWithHealthChecks(b *testing.B) {
	router, cleanup := setupBenchmarkRouter(b)
	defer cleanup()

	// Get the app ID from the mock repository
	mockRepo := router.repo.(*MockRepository)
	var appID uuid.UUID
	for id := range mockRepo.backends {
		appID = id
		break
	}

	requestID := "bench-request-health"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := router.SelectBackend(appID, "GET", fmt.Sprintf("%s-%d", requestID, i), "starter")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEWMACalculation benchmarks EWMA score calculation
func BenchmarkEWMACalculation(b *testing.B) {
	router, cleanup := setupBenchmarkRouter(b)
	defer cleanup()

	// Get a backend from the mock repository
	mockRepo := router.repo.(*MockRepository)
	var backendID uuid.UUID
	for _, backends := range mockRepo.backends {
		if len(backends) > 0 {
			backendID = backends[0].ID
			break
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		score := router.calculateEWMAScore(backendID)
		_ = score
	}
}

// BenchmarkBackendScoring benchmarks the scoring of multiple backends
func BenchmarkBackendScoring(b *testing.B) {
	router, cleanup := setupBenchmarkRouter(b)
	defer cleanup()

	// Get backends from the mock repository
	mockRepo := router.repo.(*MockRepository)
	var backends []*storage.Backend
	for _, backendList := range mockRepo.backends {
		backends = backendList
		break
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scored := router.scoreBackends(backends)
		_ = scored
	}
}

// BenchmarkPrioritySorting benchmarks backend sorting by priority and score
func BenchmarkPrioritySorting(b *testing.B) {
	router, cleanup := setupBenchmarkRouter(b)
	defer cleanup()

	// Create backends with different priorities and scores
	backends := make([]*BackendScore, 50)
	for i := 0; i < 50; i++ {
		priorities := []*int{&[]int{1}[0], &[]int{2}[0], &[]int{3}[0], nil}
		priority := priorities[i%4]
		score := float64(100 + (i * 5)) // Different scores

		backends[i] = &BackendScore{
			Backend: &storage.Backend{
				Priority: priority,
			},
			Score:    score,
			HealthOK: true,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test Pro plan sorting (priority + score)
		router.selectBestBackend(backends, "GET", "pro")

		// Test Starter plan sorting (score only)
		router.selectBestBackend(backends, "POST", "starter")
	}
}

// BenchmarkCircuitBreakerLogic benchmarks circuit breaker state checking
func BenchmarkCircuitBreakerLogic(b *testing.B) {
	router, cleanup := setupBenchmarkRouter(b)
	defer cleanup()

	// Get backends from the mock repository
	mockRepo := router.repo.(*MockRepository)
	var backends []*storage.Backend
	for _, backendList := range mockRepo.backends {
		backends = backendList
		break
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scored := router.scoreBackends(backends)
		healthy := router.filterHealthyBackends(scored)
		_ = healthy
	}
}

// BenchmarkFailoverSelection benchmarks selecting failover backends for idempotent methods
func BenchmarkFailoverSelection(b *testing.B) {
	router, cleanup := setupBenchmarkRouter(b)
	defer cleanup()

	// Get backends from the mock repository and convert to BackendScore
	mockRepo := router.repo.(*MockRepository)
	var backends []*BackendScore
	for _, backendList := range mockRepo.backends {
		for _, backend := range backendList {
			backends = append(backends, &BackendScore{
				Backend:  backend,
				Score:    float64(100 + (len(backends) * 10)), // Different latencies
				HealthOK: true,
			})
		}
		break
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test GET (idempotent) - should return top 3
		selected := router.selectBestBackend(backends, "GET", "starter")
		if len(selected) != 3 {
			b.Fatalf("Expected 3 backends for GET, got %d", len(selected))
		}

		// Test POST (non-idempotent) - should return top 1
		selected = router.selectBestBackend(backends, "POST", "starter")
		if len(selected) != 1 {
			b.Fatalf("Expected 1 backend for POST, got %d", len(selected))
		}
	}
}