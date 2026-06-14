package routing

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RouterRepository defines the subset of Repository methods used by Router
type RouterRepository interface {
	ListBackendsByAppID(ctx context.Context, appID uuid.UUID) ([]*storage.Backend, error)
	GetCircuitState(ctx context.Context, backendID uuid.UUID) (*storage.CircuitState, error)
	GetRecentHealthChecks(ctx context.Context, backendID uuid.UUID, limit int) ([]*storage.HealthCheck, error)
	InsertRoutingEvent(ctx context.Context, appID, backendID uuid.UUID, latencyMs int, outcome, requestID string) error
}

// Router handles backend selection and routing decisions
type Router struct {
	repo RouterRepository
}

// BackendScore represents a backend with its routing score
type BackendScore struct {
	Backend     *storage.Backend
	Score       float64
	CircuitState string
	HealthOK    bool
	LatencyMs   int
}

// RoutingDecision represents the result of a routing decision
type RoutingDecision struct {
	SelectedBackend  *storage.Backend `json:"selected_backend"`
	FailoverBackends []*storage.Backend `json:"failover_backends,omitempty"`
	AllBackends      []*BackendScore  `json:"all_backends"`
	Reason           string           `json:"reason"`
}

// NewRouter creates a new router instance
func NewRouter(repo RouterRepository) *Router {
	return &Router{repo: repo}
}

// SelectBackend selects the best backend for routing based on health, circuit state, and latency
func (r *Router) SelectBackend(appID uuid.UUID, method string, requestID string, plan string) (*RoutingDecision, error) {
	// Get all backends for the app
	backends, err := r.repo.ListBackendsByAppID(context.Background(), appID)
	if err != nil {
		return nil, fmt.Errorf("failed to list backends: %w", err)
	}

	if len(backends) == 0 {
		return &RoutingDecision{
			Reason: "No backends configured for this app",
		}, nil
	}

	// Score all backends
	scoredBackends := r.scoreBackends(backends)

	// Filter healthy backends (circuit not open)
	healthyBackends := r.filterHealthyBackends(scoredBackends)

	if len(healthyBackends) == 0 {
		return &RoutingDecision{
			AllBackends: scoredBackends,
			Reason:      "No healthy backends available (all circuits open or unhealthy)",
		}, nil
	}

	// Select the best backend(s)
	selectedBackends := r.selectBestBackend(healthyBackends, method, plan)
	if len(selectedBackends) == 0 {
		return &RoutingDecision{
			AllBackends: scoredBackends,
			Reason:      "No backends available after filtering",
		}, nil
	}

	// Prepare failover backends (all except the first)
	var failoverBackends []*storage.Backend
	if len(selectedBackends) > 1 {
		for _, backend := range selectedBackends[1:] {
			failoverBackends = append(failoverBackends, backend.Backend)
		}
	}

	return &RoutingDecision{
		SelectedBackend:  selectedBackends[0].Backend,
		FailoverBackends: failoverBackends,
		AllBackends:      scoredBackends,
		Reason:           r.buildSelectionReason(selectedBackends, method),
	}, nil
}

// scoreBackends calculates routing scores for all backends
func (r *Router) scoreBackends(backends []*storage.Backend) []*BackendScore {
	scored := make([]*BackendScore, 0, len(backends))

	for _, backend := range backends {
		score := &BackendScore{
			Backend: backend,
			Score:   0.0,
		}

		// Get circuit state
		circuitState, err := r.repo.GetCircuitState(context.Background(), backend.ID)
		if err != nil {
			logrus.WithError(err).WithField("backend_id", backend.ID).Error("Failed to get circuit state")
			score.CircuitState = "unknown"
			score.HealthOK = false
		} else {
			score.CircuitState = circuitState.State
			score.HealthOK = circuitState.State != "open"
		}

		// Calculate EWMA score if healthy
		if score.HealthOK {
			ewmaScore := r.calculateEWMAScore(backend.ID)
			score.Score = ewmaScore
			score.LatencyMs = int(ewmaScore) // Approximate latency for display
		}

		scored = append(scored, score)
	}

	return scored
}

// calculateEWMAScore calculates the exponentially weighted moving average latency score
func (r *Router) calculateEWMAScore(backendID uuid.UUID) float64 {
	// Get recent health checks (last 10)
	checks, err := r.repo.GetRecentHealthChecks(context.Background(), backendID, 10)
	if err != nil {
		logrus.WithError(err).WithField("backend_id", backendID).Error("Failed to get health checks")
		return 1000.0 // Default high latency
	}

	if len(checks) == 0 {
		return 1000.0 // No data, assume high latency
	}

	// Filter successful checks
	var latencies []float64
	for _, check := range checks {
		if check.OK && check.LatencyMs > 0 {
			latencies = append(latencies, float64(check.LatencyMs))
		}
	}

	if len(latencies) == 0 {
		return 1000.0 // No successful checks
	}

	// Calculate EWMA (alpha = 0.3 for recent bias)
	alpha := 0.3
	ewma := latencies[0]
	for i := 1; i < len(latencies); i++ {
		ewma = alpha*latencies[i] + (1-alpha)*ewma
	}

	return ewma
}

// filterHealthyBackends returns only backends that are healthy (circuit not open)
func (r *Router) filterHealthyBackends(backends []*BackendScore) []*BackendScore {
	var healthy []*BackendScore
	for _, backend := range backends {
		if backend.HealthOK {
			healthy = append(healthy, backend)
		}
	}
	return healthy
}

// selectBestBackend selects the backend(s) with the lowest score(s) (latency)
// For idempotent methods, returns top 3 backends for fast failover
// For non-idempotent methods, returns only the best backend
func (r *Router) selectBestBackend(backends []*BackendScore, method string, plan string) []*BackendScore {
	if len(backends) == 0 {
		return nil
	}

	// Sort backends based on plan
	if plan == "professional" {
		// For Pro plan: sort by priority (asc, nulls last) then by score (lower is better)
		sort.Slice(backends, func(i, j int) bool {
			// Handle null priorities (nulls last)
			iPriority := backends[i].Backend.Priority
			jPriority := backends[j].Backend.Priority

			if iPriority == nil && jPriority == nil {
				// Both null, compare by score
				return backends[i].Score < backends[j].Score
			}
			if iPriority == nil {
				// i is null, j comes first
				return false
			}
			if jPriority == nil {
				// j is null, i comes first
				return true
			}

			// Both have priorities, compare them first
			if *iPriority != *jPriority {
				return *iPriority < *jPriority
			}

			// Same priority, compare by score
			return backends[i].Score < backends[j].Score
		})
	} else {
		// For Starter plan: sort by score only (existing behavior)
		sort.Slice(backends, func(i, j int) bool {
			return backends[i].Score < backends[j].Score
		})
	}

	// For idempotent methods, return top 3 backends for fast failover
	if IsIdempotentMethod(method) {
		maxFailover := 3
		if len(backends) < maxFailover {
			maxFailover = len(backends)
		}
		return backends[:maxFailover]
	}

	// For non-idempotent methods, return only the best backend
	return backends[:1]
}

// RecordRoutingResult records the result of a routing attempt for future scoring
func (r *Router) RecordRoutingResult(appID, backendID uuid.UUID, latencyMs int, outcome, requestID string) error {
	return r.repo.InsertRoutingEvent(context.Background(), appID, backendID, latencyMs, outcome, requestID)
}

// IsIdempotentMethod checks if an HTTP method is considered idempotent for fast failover
func IsIdempotentMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

// IsIdempotentMethod checks if an HTTP method is considered idempotent for fast failover (public version)
func (r *Router) IsIdempotentMethod(method string) bool {
	return IsIdempotentMethod(method)
}

// buildSelectionReason builds a descriptive reason for backend selection
func (r *Router) buildSelectionReason(selectedBackends []*BackendScore, method string) string {
	if len(selectedBackends) == 0 {
		return "No backends selected"
	}

	primary := selectedBackends[0]
	reason := fmt.Sprintf("Selected backend %s with score %.2f", primary.Backend.ID, primary.Score)

	if len(selectedBackends) > 1 {
		reason += fmt.Sprintf(" (%d failover backends available for %s method)", len(selectedBackends)-1, method)
	}

	return reason
}