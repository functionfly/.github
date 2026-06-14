package health

import (
	"context"
	"fmt"
	"math"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/aws"
	"github.com/functionfly/functionfly/internal/adapters/cloudflare"
	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/functionfly/functionfly/internal/adapters/deno"
	"github.com/functionfly/functionfly/internal/adapters/fly"
	"github.com/functionfly/functionfly/internal/adapters/functionfly"
	"github.com/functionfly/functionfly/internal/adapters/vercel"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CircuitBreakerConfig holds configurable circuit breaker settings
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of consecutive failures before opening the circuit
	FailureThreshold int
	// SuccessThreshold is the number of successes in half-open state before closing
	SuccessThreshold int
	// OpenTimeout is the base timeout before moving to half-open state
	OpenTimeout time.Duration
	// MaxOpenTimeout is the maximum timeout with exponential backoff
	MaxOpenTimeout time.Duration
	// BackoffMultiplier is the multiplier for exponential backoff
	BackoffMultiplier float64
	// HalfOpenMaxRequests is the maximum number of requests allowed in half-open state
	HalfOpenMaxRequests int
}

// DefaultCircuitBreakerConfig returns production-ready circuit breaker defaults
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold:    3,
		SuccessThreshold:    2,
		OpenTimeout:         30 * time.Second,
		MaxOpenTimeout:      5 * time.Minute,
		BackoffMultiplier:   2.0,
		HalfOpenMaxRequests: 3,
	}
}

// Monitor handles backend health monitoring and circuit breaker management
type Monitor struct {
	repo              storage.Repository
	probeInterval     time.Duration
	stopChan          chan struct{}
	wg                sync.WaitGroup
	stopOnce          sync.Once
	circuitConfig     *CircuitBreakerConfig
	openCircuits      map[uuid.UUID]time.Time // Track when circuits were opened for backoff
	openCircuitsMutex sync.RWMutex
}

// NewMonitor creates a new health monitor
func NewMonitor(repo storage.Repository) *Monitor {
	return &Monitor{
		repo:          repo,
		probeInterval: 5 * time.Second,
		stopChan:      make(chan struct{}),
		circuitConfig: DefaultCircuitBreakerConfig(),
		openCircuits:  make(map[uuid.UUID]time.Time),
	}
}

// NewMonitorWithConfig creates a new health monitor with custom circuit breaker config
func NewMonitorWithConfig(repo storage.Repository, config *CircuitBreakerConfig) *Monitor {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	return &Monitor{
		repo:          repo,
		probeInterval: 5 * time.Second,
		stopChan:      make(chan struct{}),
		circuitConfig: config,
		openCircuits:  make(map[uuid.UUID]time.Time),
	}
}

// Start begins the health monitoring loop
func (m *Monitor) Start() {
	logrus.Info("Starting health monitor")

	m.wg.Add(1)
	go m.monitorLoop()
}

// Stop gracefully stops the health monitoring
func (m *Monitor) Stop() {
	m.stopOnce.Do(func() {
		logrus.Info("Stopping health monitor")
		close(m.stopChan)
		m.wg.Wait()
		logrus.Info("Health monitor stopped")
	})
}

// monitorLoop runs the continuous health monitoring
func (m *Monitor) monitorLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.probeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.probeAllBackends()
		}
	}
}

// probeAllBackends probes all backends for health
func (m *Monitor) probeAllBackends() {
	// For MVP, we'll get all backends from all apps
	// Optimized with database indexes on backends.enabled and (enabled, created_at) to avoid full table scans
	backends, err := m.getAllBackends(context.Background())
	if err != nil {
		logrus.WithError(err).Error("Failed to get backends for health check")
		return
	}

	// Probe backends concurrently
	var wg sync.WaitGroup
	for _, backend := range backends {
		wg.Add(1)
		go func(b *storage.Backend) {
			defer func() {
				if rec := recover(); rec != nil {
					logrus.WithFields(logrus.Fields{
						"panic":      rec,
						"stack":      string(debug.Stack()),
						"backend_id": b.ID,
						"backend":    b.URL,
					}).Error("Health monitor probeBackend goroutine panicked")
					wg.Done()
				}
			}()
			defer wg.Done()
			m.probeBackend(b)
		}(backend)
	}

	wg.Wait()

	// System edge probe: when EDGE_HEALTH_URL is set, probe it and update edge metrics/stats (no DB backend required)
	if edgeURL := os.Getenv("EDGE_HEALTH_URL"); edgeURL != "" {
		m.probeSystemEdge(edgeURL, os.Getenv("EDGE_HEALTH_SECRET"))
	}
}

// probeSystemEdge probes the configured edge URL and updates Prometheus + in-memory edge stats.
func (m *Monitor) probeSystemEdge(edgeURL, sharedSecret string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	backend := &storage.Backend{
		ID:           uuid.Nil,
		URL:          edgeURL,
		Region:       "system",
		SharedSecret: sharedSecret,
		Provider:     "functionfly-edge",
	}
	adapter := functionfly.NewFunctionFlyAdapter()
	result, err := adapter.HealthCheck(ctx, backend)
	if err != nil {
		monitoring.UpdateEdgeProbeAndMetrics(false, 0, err.Error())
		logrus.WithError(err).WithField("edge_url", edgeURL).Warn("Edge health probe error")
		return
	}
	monitoring.UpdateEdgeProbeAndMetrics(result.OK, result.LatencyMs, result.ErrorMessage)
	if !result.OK {
		logrus.WithFields(logrus.Fields{
			"edge_url": edgeURL,
			"status":   result.StatusCode,
			"error":    result.ErrorMessage,
		}).Warn("Edge health probe failed")
	}
}

// getAllBackends gets all enabled backends
func (m *Monitor) getAllBackends(ctx context.Context) ([]*storage.Backend, error) {
	return m.repo.GetAllEnabledBackends(ctx)
}

// probeBackend probes a single backend for health
func (m *Monitor) probeBackend(backend *storage.Backend) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get the appropriate adapter for this backend's provider
	adapter := m.getAdapterForProvider(backend.Provider)
	if adapter == nil {
		m.recordHealthCheck(ctx, backend.ID, false, 0, 0, fmt.Sprintf("Unsupported provider '%s'", backend.Provider))
		m.handleCircuitBreaker(ctx, backend.ID, false)
		return
	}

	// Perform provider-specific health check
	result, err := adapter.HealthCheck(ctx, backend)
	if err != nil {
		m.recordHealthCheck(ctx, backend.ID, false, 0, 0, fmt.Sprintf("Health check error: %v", err))
		m.handleCircuitBreaker(ctx, backend.ID, false)
		return
	}

	// Record the health check result
	m.recordHealthCheck(ctx, backend.ID, result.OK, result.StatusCode, result.LatencyMs, result.ErrorMessage)
	m.handleCircuitBreaker(ctx, backend.ID, result.OK)

	// Log additional provider-specific information
	if result.Version != "" {
		logrus.WithFields(logrus.Fields{
			"backend_id": backend.ID,
			"provider":   backend.Provider,
			"region":     result.Region,
			"version":    result.Version,
		}).Debug("Health check completed with provider info")
	}
}

// recordHealthCheck records the result of a health check
func (m *Monitor) recordHealthCheck(ctx context.Context, backendID uuid.UUID, ok bool, statusCode, latencyMs int, errorMessage string) {
	logger := logrus.WithFields(logrus.Fields{
		"backend_id":  backendID,
		"operation":   "health_check",
		"healthy":     ok,
		"status_code": statusCode,
		"latency_ms":  latencyMs,
	})

	err := m.repo.InsertHealthCheck(ctx, backendID, ok, statusCode, latencyMs, errorMessage)
	if err != nil {
		logger.WithError(err).Error("Failed to record health check")
		return
	}

	if ok {
		logger.Debug("Health check passed")
	} else {
		logger.WithField("error", errorMessage).Warn("Health check failed")
	}
}

// getAdapterForProvider returns the appropriate adapter for a provider
func (m *Monitor) getAdapterForProvider(provider string) common.ProviderAdapter {
	switch provider {
	case "workers":
		return cloudflare.NewCloudflareAdapter()
	case "vercel":
		return vercel.NewVercelAdapter()
	case "fly":
		return fly.NewFlyAdapter()
	case "deno-deploy":
		return deno.NewDenoAdapter()
	case "functionfly-edge", "functionfly":
		return functionfly.NewFunctionFlyAdapter()
	case "aws-lambda":
		return aws.NewAWSAdapter()
	default:
		return nil
	}
}

// handleCircuitBreaker manages circuit breaker state transitions with configurable thresholds and exponential backoff
func (m *Monitor) handleCircuitBreaker(ctx context.Context, backendID uuid.UUID, healthy bool) {
	requestID := fmt.Sprintf("circuit-breaker-%s-%d", backendID.String(), time.Now().Unix())

	logger := logrus.WithFields(logrus.Fields{
		"request_id": requestID,
		"backend_id": backendID,
		"operation":  "circuit_breaker",
	})

	state, err := m.repo.GetCircuitState(ctx, backendID)
	if err != nil {
		logger.WithError(err).Error("Failed to get circuit state")
		return
	}

	newState := state.State
	now := time.Now()

	// Circuit breaker logic with configurable thresholds
	switch state.State {
	case "closed":
		if !healthy {
			state.FailCount++
			// Use configurable failure threshold
			if state.FailCount >= m.circuitConfig.FailureThreshold {
				newState = "open"
				state.SinceTs = now
				state.LastFailureTs = &now

				// Track when circuit was opened for exponential backoff
				m.openCircuitsMutex.Lock()
				m.openCircuits[backendID] = now
				m.openCircuitsMutex.Unlock()

				logger.WithFields(logrus.Fields{
					"old_state":          "closed",
					"new_state":          "open",
					"fail_count":         state.FailCount,
					"failure_threshold":  m.circuitConfig.FailureThreshold,
					"transition_reason":  "failure_threshold_reached",
				}).Warn("Circuit breaker opened")
			}
		} else {
			// Reset failure count on success
			state.FailCount = 0
			state.SuccessCount++
			state.LastSuccessTs = &now
		}

	case "open":
		// Calculate timeout with exponential backoff
		timeout := m.calculateBackoffTimeout(backendID)

		if now.Sub(state.SinceTs) > timeout {
			newState = "half-open"
			state.SinceTs = now
			logger.WithFields(logrus.Fields{
				"old_state":         "open",
				"new_state":         "half-open",
				"time_since_open":   now.Sub(state.SinceTs).String(),
				"backoff_timeout":   timeout.String(),
				"transition_reason": "timeout_expired",
			}).Info("Circuit breaker moving to half-open")
		}

	case "half-open":
		if healthy {
			// Use configurable success threshold
			state.SuccessCount++
			if state.SuccessCount >= m.circuitConfig.SuccessThreshold {
				newState = "closed"
				state.SinceTs = now
				state.FailCount = 0

				// Remove from open circuits tracking
				m.openCircuitsMutex.Lock()
				delete(m.openCircuits, backendID)
				m.openCircuitsMutex.Unlock()

				logger.WithFields(logrus.Fields{
					"old_state":          "half-open",
					"new_state":          "closed",
					"success_count":      state.SuccessCount,
					"success_threshold":  m.circuitConfig.SuccessThreshold,
					"transition_reason":  "success_in_half_open",
				}).Info("Circuit breaker closed")
			}
		} else {
			// Failure in half-open, go back to open with incremented backoff
			newState = "open"
			state.SinceTs = now
			state.FailCount++
			state.SuccessCount = 0
			state.LastFailureTs = &now

			// Update open time for backoff calculation
			m.openCircuitsMutex.Lock()
			m.openCircuits[backendID] = now
			m.openCircuitsMutex.Unlock()

			logger.WithFields(logrus.Fields{
				"old_state":         "half-open",
				"new_state":         "open",
				"fail_count":        state.FailCount,
				"transition_reason": "failure_in_half_open",
			}).Warn("Circuit breaker reopened")
		}
	}

	// Update state if changed
	if newState != state.State {
		state.State = newState
		err = m.repo.UpsertCircuitState(ctx, state)
		if err != nil {
			logger.WithError(err).Error("Failed to update circuit state")
		}
	} else {
		// Still update the state for failure/success counts
		err = m.repo.UpdateCircuitState(ctx, state)
		if err != nil {
			logger.WithError(err).Error("Failed to update circuit state")
		}
	}
}

// calculateBackoffTimeout calculates exponential backoff timeout for circuit breaker
func (m *Monitor) calculateBackoffTimeout(backendID uuid.UUID) time.Duration {
	m.openCircuitsMutex.RLock()
	openTime, exists := m.openCircuits[backendID]
	m.openCircuitsMutex.RUnlock()

	if !exists {
		return m.circuitConfig.OpenTimeout
	}

	// Calculate how many times the circuit has been opened
	elapsed := time.Since(openTime)
	baseTimeout := m.circuitConfig.OpenTimeout

	// Exponential backoff: timeout = base * (multiplier ^ attempts)
	// Approximate attempts from elapsed time
	attempts := int(elapsed / baseTimeout)
	if attempts < 0 {
		attempts = 0
	}

	backoffTimeout := time.Duration(float64(baseTimeout) * math.Pow(m.circuitConfig.BackoffMultiplier, float64(attempts)))

	// Cap at maximum timeout
	if backoffTimeout > m.circuitConfig.MaxOpenTimeout {
		backoffTimeout = m.circuitConfig.MaxOpenTimeout
	}

	return backoffTimeout
}

