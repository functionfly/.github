package routing

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CanaryRouter handles canary traffic routing
type CanaryRouter struct {
	canaryRepo     *registry.CanaryConfigRepository
	metricsTracker *CanaryMetricsTracker
	mu             sync.RWMutex
}

// CanaryMetricsTracker tracks canary metrics for analysis
type CanaryMetricsTracker struct {
	metrics map[uuid.UUID][]*CanaryMetricSnapshot
	mu      sync.Mutex
}

// CanaryMetricSnapshot represents a point in time metric snapshot
type CanaryMetricSnapshot struct {
	Timestamp    time.Time `json:"timestamp"`
	ErrorRate    float64   `json:"error_rate"`
	LatencyP50   float64   `json:"latency_p50"`
	LatencyP95   float64   `json:"latency_p95"`
	LatencyP99   float64   `json:"latency_p99"`
	RequestCount int       `json:"request_count"`
}

// NewCanaryRouter creates a new canary router
func NewCanaryRouter(canaryRepo *registry.CanaryConfigRepository) *CanaryRouter {
	return &CanaryRouter{
		canaryRepo:     canaryRepo,
		metricsTracker: NewCanaryMetricsTracker(),
	}
}

// NewCanaryMetricsTracker creates a new metrics tracker
func NewCanaryMetricsTracker() *CanaryMetricsTracker {
	return &CanaryMetricsTracker{
		metrics: make(map[uuid.UUID][]*CanaryMetricSnapshot),
	}
}

// RouteDecision represents the result of a canary routing decision
type RouteDecision struct {
	FunctionID   uuid.UUID `json:"function_id"`
	Version      string    `json:"version"`
	IsCanary     bool      `json:"is_canary"`
	CanaryConfig *registry.CanaryConfig
	Reason       string `json:"reason"`
}

// Route determines which version to route a request to
func (cr *CanaryRouter) Route(ctx context.Context, functionID uuid.UUID, stableVersion string, requestHash string) (*RouteDecision, error) {
	// Get canary config for the function
	canary, err := cr.canaryRepo.GetByFunctionID(functionID)
	if err != nil || canary == nil {
		// No canary, use stable version
		return &RouteDecision{
			FunctionID: functionID,
			Version:    stableVersion,
			IsCanary:   false,
			Reason:     "no active canary",
		}, nil
	}

	// Check if canary is still active
	if canary.Status != "active" {
		return &RouteDecision{
			FunctionID: functionID,
			Version:    stableVersion,
			IsCanary:   false,
			Reason:     "canary not active",
		}, nil
	}

	// Check if error rate exceeds threshold
	if canary.ErrorRate > canary.PromoteThreshold {
		logrus.Warnf("Canary error rate %.2f%% exceeds threshold %.2f%% for function %s",
			canary.ErrorRate*100, canary.PromoteThreshold*100, functionID)

		return &RouteDecision{
			FunctionID:   functionID,
			Version:      stableVersion,
			IsCanary:     false,
			CanaryConfig: canary,
			Reason:       "error rate exceeds threshold",
		}, nil
	}

	// Determine routing based on traffic percentage
	hash := generateHash(requestHash)
	percent := hash % 100

	if percent < canary.TrafficPercent {
		return &RouteDecision{
			FunctionID:   functionID,
			Version:      canary.Version,
			IsCanary:     true,
			CanaryConfig: canary,
			Reason:       fmt.Sprintf("routing %.0f%% to canary", float64(canary.TrafficPercent)),
		}, nil
	}

	return &RouteDecision{
		FunctionID:   functionID,
		Version:      stableVersion,
		IsCanary:     false,
		CanaryConfig: canary,
		Reason:       fmt.Sprintf("routing %d%% to stable", 100-canary.TrafficPercent),
	}, nil
}

// RouteByRequest routes based on an HTTP request
func (cr *CanaryRouter) RouteByRequest(ctx context.Context, functionID uuid.UUID, stableVersion string, r *http.Request) (*RouteDecision, error) {
	hash := generateHashFromRequest(r)
	return cr.Route(ctx, functionID, stableVersion, hash)
}

// generateHash creates a deterministic hash from a string
func generateHash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}

// generateHashFromRequest creates a hash from request attributes for consistent routing
func generateHashFromRequest(r *http.Request) string {
	// Use combination of IP and path for consistent routing
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}

	// Include user agent for better distribution
	ua := r.Header.Get("User-Agent")

	return fmt.Sprintf("%s:%s:%s", r.URL.Path, ip, ua)
}

// RecordRequest records a request for metrics
func (cr *CanaryRouter) RecordRequest(canaryID uuid.UUID, success bool, latencyMs float64) {
	cr.metricsTracker.RecordMetric(canaryID, success, latencyMs)

	// Also update the canary config repository
	if err := cr.canaryRepo.RecordCanaryRequest(canaryID, success); err != nil {
		logrus.WithError(err).Error("Failed to record canary request")
	}
}

// RecordMetric records a metric snapshot (one latency sample per request).
// Percentiles (P50/P95/P99) are computed in CalculateMetrics from the sample window.
func (cmt *CanaryMetricsTracker) RecordMetric(canaryID uuid.UUID, success bool, latencyMs float64) {
	cmt.mu.Lock()
	defer cmt.mu.Unlock()

	snapshot := &CanaryMetricSnapshot{
		Timestamp:    time.Now(),
		RequestCount: 1,
	}

	if !success {
		snapshot.ErrorRate = 1.0
	}

	// Store single sample; P50/P95/P99 are derived when aggregating in CalculateMetrics
	snapshot.LatencyP50 = latencyMs
	snapshot.LatencyP95 = 0
	snapshot.LatencyP99 = 0

	cmt.metrics[canaryID] = append(cmt.metrics[canaryID], snapshot)

	// Keep only last hour of metrics
	cutoff := time.Now().Add(-time.Hour)
	filtered := make([]*CanaryMetricSnapshot, 0)
	for _, m := range cmt.metrics[canaryID] {
		if m.Timestamp.After(cutoff) {
			filtered = append(filtered, m)
		}
	}
	cmt.metrics[canaryID] = filtered
}

// GetMetrics returns metrics for a canary
func (cmt *CanaryMetricsTracker) GetMetrics(canaryID uuid.UUID) []*CanaryMetricSnapshot {
	cmt.mu.Lock()
	defer cmt.mu.Unlock()

	return cmt.metrics[canaryID]
}

// ShouldAutoPromote determines if a canary should be auto-promoted
func (cr *CanaryRouter) ShouldAutoPromote(canaryID uuid.UUID) (bool, error) {
	canary, err := cr.canaryRepo.GetByID(canaryID)
	if err != nil {
		return false, err
	}

	// Check if auto-promote is enabled
	if !canary.AutoPromote {
		return false, nil
	}

	// Check if enough time has passed
	windowStart := canary.CreatedAt.Add(time.Duration(canary.PromoteWindow) * time.Second)
	if time.Now().Before(windowStart) {
		return false, nil
	}

	// Check error rate
	if canary.ErrorRate <= canary.PromoteThreshold {
		return true, nil
	}

	return false, nil
}

// GetCanaryStatus returns the current status of a canary
func (cr *CanaryRouter) GetCanaryStatus(functionID uuid.UUID) (*registry.CanaryConfig, error) {
	return cr.canaryRepo.GetByFunctionID(functionID)
}

// CalculateMetrics calculates aggregate metrics for a canary
func (cr *CanaryRouter) CalculateMetrics(canaryID uuid.UUID) (*CanaryMetricSnapshot, error) {
	canary, err := cr.canaryRepo.GetByID(canaryID)
	if err != nil {
		return nil, err
	}

	// Get metrics from tracker
	metrics := cr.metricsTracker.GetMetrics(canaryID)

	if len(metrics) == 0 {
		// Calculate from stored data
		errorRate := canary.ErrorRate
		requestCount := canary.RequestCount

		return &CanaryMetricSnapshot{
			Timestamp:    time.Now(),
			ErrorRate:    errorRate,
			RequestCount: requestCount,
		}, nil
	}

	// Aggregate from snapshots: one latency sample per snapshot (LatencyP50)
	var totalRequests int
	var totalErrors int
	latencies := make([]float64, 0, len(metrics))

	for _, m := range metrics {
		totalRequests += m.RequestCount
		if m.ErrorRate > 0 {
			totalErrors += int(float64(m.RequestCount) * m.ErrorRate)
		}
		latencies = append(latencies, m.LatencyP50)
	}

	errorRate := 0.0
	if totalRequests > 0 {
		errorRate = float64(totalErrors) / float64(totalRequests)
	}

	// Compute percentiles (nearest-rank) from sample window
	ps := percentileNearestRank(latencies, 50, 95, 99)
	p50, p95, p99 := ps[0], ps[1], ps[2]

	return &CanaryMetricSnapshot{
		Timestamp:    time.Now(),
		ErrorRate:    errorRate,
		LatencyP50:   p50,
		LatencyP95:   p95,
		LatencyP99:   p99,
		RequestCount: totalRequests,
	}, nil
}

// percentileNearestRank returns the requested percentiles from samples using nearest-rank.
// Each p is 0-100 (e.g. 50, 95, 99). Returns 0 for any percentile when samples is empty.
func percentileNearestRank(samples []float64, percentiles ...int) []float64 {
	out := make([]float64, len(percentiles))
	if len(samples) == 0 {
		return out
	}
	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	sort.Float64s(sorted)
	n := float64(len(sorted))

	for i, p := range percentiles {
		if p <= 0 {
			out[i] = sorted[0]
			continue
		}
		if p >= 100 {
			out[i] = sorted[len(sorted)-1]
			continue
		}
		// Nearest-rank: index = ceil(p/100 * n) - 1, 0-based
		idx := int(math.Ceil(n*float64(p)/100)) - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sorted) {
			idx = len(sorted) - 1
		}
		out[i] = sorted[idx]
	}
	return out
}

// WeightedRandom selects a version based on weights
func WeightedRandom(stableWeight, canaryWeight int) bool {
	if canaryWeight <= 0 {
		return false
	}
	if stableWeight+canaryWeight <= 0 {
		return false
	}

	// Simple weighted random selection
	r := rand.Intn(stableWeight + canaryWeight)
	return r < canaryWeight
}

// DeterministicWeightedRandom selects a version based on weights using a hash
func DeterministicWeightedRandom(hash string, canaryPercent int) bool {
	if canaryPercent <= 0 {
		return false
	}
	if canaryPercent >= 100 {
		return true
	}

	h := generateHash(hash)
	return h%100 < canaryPercent
}
