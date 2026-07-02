package routing

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/health"
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
	GetRecentRoutingEventsByBackend(ctx context.Context, backendID uuid.UUID, limit int) ([]*storage.RoutingEvent, error)
}

// Router handles backend selection and routing decisions
type Router struct {
	repo         RouterRepository
	healthMon    *health.Monitor
	decisionCache sync.Map // map[uuid.UUID]*cachedDecision
	cacheTTL      time.Duration
	ewmaSource    string   // "health" or "real"
}

// BackendScore represents a backend with its routing score
type BackendScore struct {
	Backend      *storage.Backend
	Score        float64
	CircuitState string
	HealthOK     bool
	LatencyMs    int
}

// RoutingDecision represents the result of a routing decision
type RoutingDecision struct {
	SelectedBackend  *storage.Backend `json:"selected_backend"`
	FailoverBackends []*storage.Backend `json:"failover_backends,omitempty"`
	AllBackends      []*BackendScore  `json:"all_backends"`
	Reason           string           `json:"reason"`
}

// cachedDecision holds a routing decision cached for a short TTL.
type cachedDecision struct {
	decision *RoutingDecision
	cachedAt time.Time
}

// NewRouter creates a new router instance
func NewRouter(repo RouterRepository) *Router {
	ewmaSource := os.Getenv("EWMA_SOURCE")
	if ewmaSource != "real" {
		ewmaSource = "health"
	}
	return &Router{
		repo:       repo,
		cacheTTL:   1 * time.Second,
		ewmaSource: ewmaSource,
	}
}

// SetHealthMonitor provides the health monitor for circuit breaker access.
func (r *Router) SetHealthMonitor(mon *health.Monitor) {
	r.healthMon = mon
}

// InvalidateCache removes a cached routing decision for the given app.
func (r *Router) InvalidateCache(appID uuid.UUID) {
	r.decisionCache.Delete(appID)
}

// SelectBackend selects the best backend for routing based on health, circuit state, and latency.
// Uses a 1-second decision cache to avoid repeated DB queries under high traffic.
func (r *Router) SelectBackend(appID uuid.UUID, method string, requestID string, plan string) (*RoutingDecision, error) {
	// Check decision cache
	if v, ok := r.decisionCache.Load(appID); ok {
		cached := v.(*cachedDecision)
		if time.Since(cached.cachedAt) < r.cacheTTL {
			return cached.decision, nil
		}
	}

	backends, err := r.repo.ListBackendsByAppID(context.Background(), appID)
	if err != nil {
		return nil, fmt.Errorf("failed to list backends: %w", err)
	}

	if len(backends) == 0 {
		return &RoutingDecision{
			Reason: "No backends configured for this app",
		}, nil
	}

	scoredBackends := r.scoreBackends(backends)
	healthyBackends := r.filterHealthyBackends(scoredBackends)

	if len(healthyBackends) == 0 {
		return &RoutingDecision{
			AllBackends: scoredBackends,
			Reason:      "No healthy backends available (all circuits open or unhealthy)",
		}, nil
	}

	selectedBackends := r.selectBestBackend(healthyBackends, method, plan)
	if len(selectedBackends) == 0 {
		return &RoutingDecision{
			AllBackends: scoredBackends,
			Reason:      "No backends available after filtering",
		}, nil
	}

	var failoverBackends []*storage.Backend
	if len(selectedBackends) > 1 {
		for _, backend := range selectedBackends[1:] {
			failoverBackends = append(failoverBackends, backend.Backend)
		}
	}

	decision := &RoutingDecision{
		SelectedBackend:  selectedBackends[0].Backend,
		FailoverBackends: failoverBackends,
		AllBackends:      scoredBackends,
		Reason:           r.buildSelectionReason(selectedBackends, method),
	}

	// Cache the decision
	r.decisionCache.Store(appID, &cachedDecision{
		decision: decision,
		cachedAt: time.Now(),
	})

	return decision, nil
}

// scoreBackends calculates routing scores for all backends
func (r *Router) scoreBackends(backends []*storage.Backend) []*BackendScore {
	scored := make([]*BackendScore, 0, len(backends))

	for _, backend := range backends {
		score := &BackendScore{
			Backend: backend,
			Score:   0.0,
		}

		circuitState, err := r.repo.GetCircuitState(context.Background(), backend.ID)
		if err != nil {
			logrus.WithError(err).WithField("backend_id", backend.ID).Error("Failed to get circuit state")
			score.CircuitState = "unknown"
			score.HealthOK = false
		} else {
			score.CircuitState = circuitState.State
			score.HealthOK = circuitState.State != "open"
		}

		if score.HealthOK {
			ewmaScore := r.calculateEWMAScore(backend.ID)
			score.Score = ewmaScore
			score.LatencyMs = int(ewmaScore)
		}

		scored = append(scored, score)
	}

	return scored
}

// calculateEWMAScore calculates the EWMA latency score.
// Supports shadow mode: always computes health-check EWMA, and optionally
// computes routing-event EWMA for comparison logging.
func (r *Router) calculateEWMAScore(backendID uuid.UUID) float64 {
	healthEWMA := r.calculateEWMAFromHealthChecks(backendID)

	if r.ewmaSource == "real" {
		realEWMA := r.calculateEWMAFromRoutingEvents(backendID)
		if realEWMA > 0 {
			return realEWMA
		}
		// Fall back to health if no routing events yet
		return healthEWMA
	}

	// Shadow mode: compute real EWMA for comparison but use health EWMA
	realEWMA := r.calculateEWMAFromRoutingEvents(backendID)
	if realEWMA > 0 && healthEWMA > 0 {
		delta := math.Abs(healthEWMA-realEWMA) / healthEWMA * 100
		if delta > 20 {
			logrus.WithFields(logrus.Fields{
				"backend_id":  backendID,
				"health_ewma": healthEWMA,
				"real_ewma":   realEWMA,
				"delta_pct":   delta,
			}).Debug("EWMA shadow comparison")
		}
	}

	return healthEWMA
}

func (r *Router) calculateEWMAFromHealthChecks(backendID uuid.UUID) float64 {
	checks, err := r.repo.GetRecentHealthChecks(context.Background(), backendID, 10)
	if err != nil {
		logrus.WithError(err).WithField("backend_id", backendID).Error("Failed to get health checks")
		return 1000.0
	}

	if len(checks) == 0 {
		return 1000.0
	}

	var latencies []float64
	for _, check := range checks {
		if check.OK && check.LatencyMs > 0 {
			latencies = append(latencies, float64(check.LatencyMs))
		}
	}

	if len(latencies) == 0 {
		return 1000.0
	}

	return computeEWMA(latencies)
}

func (r *Router) calculateEWMAFromRoutingEvents(backendID uuid.UUID) float64 {
	events, err := r.repo.GetRecentRoutingEventsByBackend(context.Background(), backendID, 20)
	if err != nil || len(events) == 0 {
		return 0
	}

	var latencies []float64
	for _, event := range events {
		if event.Outcome == "success" && event.LatencyMs > 0 {
			latencies = append(latencies, float64(event.LatencyMs))
		}
	}

	if len(latencies) == 0 {
		return 0
	}

	return computeEWMA(latencies)
}

func computeEWMA(latencies []float64) float64 {
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
func (r *Router) selectBestBackend(backends []*BackendScore, method string, plan string) []*BackendScore {
	if len(backends) == 0 {
		return nil
	}

	if plan == "professional" {
		sort.Slice(backends, func(i, j int) bool {
			iPriority := backends[i].Backend.Priority
			jPriority := backends[j].Backend.Priority

			if iPriority == nil && jPriority == nil {
				return backends[i].Score < backends[j].Score
			}
			if iPriority == nil {
				return false
			}
			if jPriority == nil {
				return true
			}

			if *iPriority != *jPriority {
				return *iPriority < *jPriority
			}

			return backends[i].Score < backends[j].Score
		})
	} else {
		sort.Slice(backends, func(i, j int) bool {
			return backends[i].Score < backends[j].Score
		})
	}

	if IsIdempotentMethod(method) {
		maxFailover := 3
		if len(backends) < maxFailover {
			maxFailover = len(backends)
		}
		return backends[:maxFailover]
	}

	return backends[:1]
}

// RecordRoutingResult records the result of a routing attempt for future scoring
func (r *Router) RecordRoutingResult(appID, backendID uuid.UUID, latencyMs int, outcome, requestID string) error {
	// Invalidate decision cache for this app on any routing result
	r.InvalidateCache(appID)
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