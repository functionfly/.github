package health

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

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

// Monitor handles backend health monitoring and circuit breaker management
type Monitor struct {
	repo          storage.Repository
	probeInterval time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
	stopOnce      sync.Once
}

// NewMonitor creates a new health monitor
func NewMonitor(repo storage.Repository) *Monitor {
	return &Monitor{
		repo:          repo,
		probeInterval: 5 * time.Second, // Probe every 5 seconds
		stopChan:      make(chan struct{}),
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
	backends, err := m.getAllBackends()
	if err != nil {
		logrus.WithError(err).Error("Failed to get backends for health check")
		return
	}

	// Probe backends concurrently
	var wg sync.WaitGroup
	for _, backend := range backends {
		wg.Add(1)
		go func(b *storage.Backend) {
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
func (m *Monitor) getAllBackends() ([]*storage.Backend, error) {
	return m.repo.GetAllEnabledBackends()
}

// probeBackend probes a single backend for health
func (m *Monitor) probeBackend(backend *storage.Backend) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get the appropriate adapter for this backend's provider
	adapter := m.getAdapterForProvider(backend.Provider)
	if adapter == nil {
		m.recordHealthCheck(backend.ID, false, 0, 0, fmt.Sprintf("Unsupported provider '%s'", backend.Provider))
		m.handleCircuitBreaker(backend.ID, false)
		return
	}

	// Perform provider-specific health check
	result, err := adapter.HealthCheck(ctx, backend)
	if err != nil {
		m.recordHealthCheck(backend.ID, false, 0, 0, fmt.Sprintf("Health check error: %v", err))
		m.handleCircuitBreaker(backend.ID, false)
		return
	}

	// Record the health check result
	m.recordHealthCheck(backend.ID, result.OK, result.StatusCode, result.LatencyMs, result.ErrorMessage)
	m.handleCircuitBreaker(backend.ID, result.OK)

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
func (m *Monitor) recordHealthCheck(backendID uuid.UUID, ok bool, statusCode, latencyMs int, errorMessage string) {
	logger := logrus.WithFields(logrus.Fields{
		"backend_id":  backendID,
		"operation":   "health_check",
		"healthy":     ok,
		"status_code": statusCode,
		"latency_ms":  latencyMs,
	})

	err := m.repo.InsertHealthCheck(backendID, ok, statusCode, latencyMs, errorMessage)
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
	default:
		return nil
	}
}

// handleCircuitBreaker manages circuit breaker state transitions
func (m *Monitor) handleCircuitBreaker(backendID uuid.UUID, healthy bool) {
	requestID := fmt.Sprintf("circuit-breaker-%s-%d", backendID.String(), time.Now().Unix())

	logger := logrus.WithFields(logrus.Fields{
		"request_id": requestID,
		"backend_id": backendID,
		"operation":  "circuit_breaker",
	})

	state, err := m.repo.GetCircuitState(backendID)
	if err != nil {
		logger.WithError(err).Error("Failed to get circuit state")
		return
	}

	newState := state.State
	now := time.Now()

	// Circuit breaker logic
	switch state.State {
	case "closed":
		if !healthy {
			state.FailCount++
			// If we have 3 failures in a row, open the circuit
			if state.FailCount >= 3 {
				newState = "open"
				state.SinceTs = now
				state.LastFailureTs = &now
				logger.WithFields(logrus.Fields{
					"old_state":         "closed",
					"new_state":         "open",
					"fail_count":        state.FailCount,
					"transition_reason": "failure_threshold_reached",
				}).Warn("Circuit breaker opened")
			}
		} else {
			// Reset failure count on success
			state.FailCount = 0
			state.SuccessCount++
			state.LastSuccessTs = &now
		}

	case "open":
		// Check if we should move to half-open (after 30 seconds)
		if now.Sub(state.SinceTs) > 30*time.Second {
			newState = "half-open"
			state.SinceTs = now
			logger.WithFields(logrus.Fields{
				"old_state":         "open",
				"new_state":         "half-open",
				"time_since_open":   now.Sub(state.SinceTs).String(),
				"transition_reason": "timeout_expired",
			}).Info("Circuit breaker moving to half-open")
		}

	case "half-open":
		if healthy {
			// Success in half-open, close the circuit
			newState = "closed"
			state.SinceTs = now
			state.FailCount = 0
			state.SuccessCount++
			state.LastSuccessTs = &now
			logger.WithFields(logrus.Fields{
				"old_state":         "half-open",
				"new_state":         "closed",
				"transition_reason": "success_in_half_open",
			}).Info("Circuit breaker closed")
		} else {
			// Failure in half-open, go back to open
			newState = "open"
			state.SinceTs = now
			state.FailCount++
			state.LastFailureTs = &now
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
		err = m.repo.UpsertCircuitState(state)
		if err != nil {
			logger.WithError(err).Error("Failed to update circuit state")
		}
	} else {
		// Still update the state for failure/success counts
		err = m.repo.UpdateCircuitState(state)
		if err != nil {
			logger.WithError(err).Error("Failed to update circuit state")
		}
	}
}
