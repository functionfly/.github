package functionregistry

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// StatsAggregator periodically updates function statistics
type StatsAggregator struct {
	repo     *registry.RegistryRepository
	interval time.Duration
	stopChan chan struct{}
}

// NewStatsAggregator creates a new stats aggregator
func NewStatsAggregator(repo *registry.RegistryRepository, interval time.Duration) *StatsAggregator {
	return &StatsAggregator{
		repo:     repo,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start begins the aggregation loop
func (s *StatsAggregator) Start(ctx context.Context) {
	logrus.Info("Starting function registry stats aggregator")

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run immediately on start
	s.aggregateAll(ctx)

	for {
		select {
		case <-ticker.C:
			s.aggregateAll(ctx)
		case <-s.stopChan:
			logrus.Info("Stopping function registry stats aggregator")
			return
		case <-ctx.Done():
			logrus.Info("Context cancelled, stopping stats aggregator")
			return
		}
	}
}

// Stop stops the aggregator
func (s *StatsAggregator) Stop() {
	close(s.stopChan)
}

// aggregateAll updates stats for all functions
func (s *StatsAggregator) aggregateAll(ctx context.Context) {
	logrus.Debug("Running function stats aggregation")

	// Get all public functions
	functions, _, err := s.repo.ListFunctions("", "", nil, "public", 1000, 0)
	if err != nil {
		logrus.WithError(err).Error("Failed to list functions for aggregation")
		return
	}

	for _, fn := range functions {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := s.aggregateFunction(ctx, fn.ID); err != nil {
			logrus.WithError(err).WithField("function_id", fn.ID).Error("Failed to aggregate function stats")
		}
	}

	logrus.Debugf("Aggregated stats for %d functions", len(functions))
}

// aggregateFunction updates stats for a single function
func (s *StatsAggregator) aggregateFunction(ctx context.Context, functionID uuid.UUID) error {
	// Get stats for last 24 hours
	since := time.Now().Add(-24 * time.Hour)

	// Get extended trust stats including p50, timeout_rate, error_rate
	totalCalls, successRate, avgLatency, p50Latency, p95Latency, timeoutRate, errorRate, err := s.repo.GetFunctionTrustStats(functionID, since)
	if err != nil {
		return fmt.Errorf("failed to get trust stats: %w", err)
	}

	// Get consumer diversity metrics
	uniqueIPs, uniqueTenants, uniqueUsers, err := s.repo.GetConsumerDiversity(functionID, since)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get consumer diversity, using defaults")
		uniqueIPs = 0
		uniqueTenants = 0
		uniqueUsers = 0
	}

	// Get or create rating
	rating, err := s.repo.GetOrCreateRating(functionID)
	if err != nil {
		return fmt.Errorf("failed to get rating: %w", err)
	}

	// Update rating with new stats
	rating.SuccessRate = successRate
	rating.AvgLatencyMs = avgLatency
	rating.P50LatencyMs = p50Latency
	rating.P95LatencyMs = p95Latency
	rating.TimeoutRate = timeoutRate
	rating.ErrorRate = errorRate
	rating.TotalRatings = totalCalls

	// Update diversity metrics
	rating.TenantDiversity = uniqueTenants
	rating.UserDiversity = uniqueUsers

	// Calculate consumer diversity score (percentage of unique IPs vs total calls)
	if totalCalls > 0 {
		rating.ConsumerDiversity = float64(uniqueIPs) / float64(totalCalls) * 100
	} else {
		rating.ConsumerDiversity = 0
	}

	// Calculate reliability score (based on success rate)
	rating.ReliabilityScore = successRate

	// Calculate latency score (inverse - lower latency = higher score)
	latencyScore := calculateLatencyScore(p95Latency)
	rating.LatencyScore = latencyScore

	// Calculate trust score using the new trust score calculator
	trustCalc := NewTrustScoreCalculator()

	// Get function for deterministic score
	fn, err := s.repo.GetFunctionByID(functionID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get function for trust calculation")
	}

	functionAgeDays := 0
	if fn != nil && !fn.CreatedAt.IsZero() {
		functionAgeDays = GetFunctionAgeDays(fn.CreatedAt)
	}

	metrics := &TrustMetrics{
		SuccessRate:     successRate,
		P50LatencyMs:    p50Latency,
		P95LatencyMs:    p95Latency,
		AvgLatencyMs:    avgLatency,
		TimeoutRate:     timeoutRate,
		ErrorRate:       errorRate,
		TotalCalls:      totalCalls,
		IsDeterministic: fn != nil && fn.DeterministicScore > 50,
		UniqueTenants:   uniqueTenants,
		UniqueUsers:     uniqueUsers,
		UniqueIPs:       uniqueIPs,
		FunctionAgeDays: functionAgeDays,
	}

	trustResult := trustCalc.Calculate(metrics)
	rating.TrustScore = trustResult.TrustScore

	// Calculate overall score (weighted average)
	rating.OverallScore = (rating.ReliabilityScore + rating.LatencyScore + rating.DocumentationScore) / 3

	// Update timestamp for trust score
	now := time.Now()
	rating.TrustUpdatedAt = &now

	if err := s.repo.UpdateRating(rating); err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"function_id": functionID,
		"trust_score": rating.TrustScore,
		"total_calls": totalCalls,
	}).Debug("Updated function trust score")

	return nil
}

// calculateLatencyScore calculates the latency score based on p95 latency
func calculateLatencyScore(p95Latency int) float64 {
	if p95Latency < 50 {
		return 100
	} else if p95Latency < 200 {
		return 80
	} else if p95Latency < 500 {
		return 60
	} else if p95Latency < 1000 {
		return 40
	}
	return 20
}
