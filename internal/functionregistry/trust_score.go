package functionregistry

import (
	"math"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// TrustScoreCalculator calculates trust scores for functions based on various metrics
type TrustScoreCalculator struct {
	// Minimum executions required to calculate a trust score
	MinExecutions int
	// Minimum calls per tenant to count toward diversity
	MinCallsPerTenant int
	// Days of sustained usage required
	SustainedUsageDays int
	// Days of inactivity before trust decays
	InactivityDays int
}

// NewTrustScoreCalculator creates a new trust score calculator with default values
func NewTrustScoreCalculator() *TrustScoreCalculator {
	return &TrustScoreCalculator{
		MinExecutions:      10,
		MinCallsPerTenant:  20,
		SustainedUsageDays: 7,
		InactivityDays:     30,
	}
}

// TrustMetrics holds all the metrics needed for trust score calculation
type TrustMetrics struct {
	SuccessRate      float64
	P50LatencyMs     int
	P95LatencyMs     int
	AvgLatencyMs     int
	TimeoutRate      float64
	ErrorRate        float64
	TotalCalls       int
	IsDeterministic  bool
	UniqueTenants    int
	UniqueUsers      int
	UniqueIPs        int
	FunctionAgeDays  int
	LastActivityDays int
}

// TrustScoreResult contains the calculated trust score and its components
type TrustScoreResult struct {
	TrustScore       float64 `json:"trust_score"`
	SuccessScore     float64 `json:"success_score"`
	LatencyScore     float64 `json:"latency_score"`
	ReliabilityScore float64 `json:"reliability_score"`
	DeterminismScore float64 `json:"determinism_score"`
	VolumeScore      float64 `json:"volume_score"`
	DiversityScore   float64 `json:"diversity_score"`
	HasEnoughData    bool    `json:"has_enough_data"`
	TrustLevel       string  `json:"trust_level"`
}

// Calculate computes the trust score from the given metrics
func (t *TrustScoreCalculator) Calculate(metrics *TrustMetrics) *TrustScoreResult {
	// Check if we have enough data
	hasEnoughData := metrics.TotalCalls >= t.MinExecutions

	if !hasEnoughData {
		return &TrustScoreResult{
			TrustScore:    0,
			HasEnoughData: false,
			TrustLevel:    "insufficient_data",
		}
	}

	// Calculate individual component scores
	successScore := t.calculateSuccessScore(metrics.SuccessRate)
	latencyScore := t.calculateLatencyScore(metrics.P50LatencyMs, metrics.P95LatencyMs)
	reliabilityScore := t.calculateReliabilityScore(metrics.ErrorRate, metrics.TimeoutRate)
	determinismScore := t.calculateDeterminismScore(metrics.IsDeterministic)
	volumeScore := t.calculateVolumeScore(metrics.TotalCalls)
	diversityScore := t.calculateDiversityScore(metrics.UniqueTenants, metrics.UniqueUsers, metrics.UniqueIPs, metrics.TotalCalls)

	// Apply anti-gaming decay for inactive functions
	trustScore := t.applyWeights(successScore, latencyScore, reliabilityScore, determinismScore, volumeScore, diversityScore)

	// Apply freshness decay
	if metrics.LastActivityDays > t.InactivityDays {
		decayFactor := math.Max(0.5, 1.0-float64(metrics.LastActivityDays-t.InactivityDays)/100.0)
		trustScore *= decayFactor
	}

	// Clamp to 0-100 range
	trustScore = math.Min(100, math.Max(0, trustScore))

	return &TrustScoreResult{
		TrustScore:       trustScore,
		SuccessScore:     successScore,
		LatencyScore:     latencyScore,
		ReliabilityScore: reliabilityScore,
		DeterminismScore: determinismScore,
		VolumeScore:      volumeScore,
		DiversityScore:   diversityScore,
		HasEnoughData:    true,
		TrustLevel:       t.getTrustLevel(trustScore),
	}
}

// calculateSuccessScore calculates the success score component
func (t *TrustScoreCalculator) calculateSuccessScore(successRate float64) float64 {
	// Success rate directly maps to score
	return successRate
}

// calculateLatencyScore calculates the latency score component (lower is better)
func (t *TrustScoreCalculator) calculateLatencyScore(p50LatencyMs, p95LatencyMs int) float64 {
	// Use weighted average: 60% p50, 40% p95
	weightedLatency := float64(p50LatencyMs)*0.6 + float64(p95LatencyMs)*0.4

	// Score based on latency thresholds (in ms)
	// < 50ms = 100, < 100ms = 90, < 200ms = 80, < 500ms = 60, < 1000ms = 40, >= 1000ms = 20
	if weightedLatency < 50 {
		return 100
	} else if weightedLatency < 100 {
		return 90
	} else if weightedLatency < 200 {
		return 80
	} else if weightedLatency < 500 {
		return 60
	} else if weightedLatency < 1000 {
		return 40
	}
	return 20
}

// calculateReliabilityScore calculates the reliability score based on error and timeout rates
func (t *TrustScoreCalculator) calculateReliabilityScore(errorRate, timeoutRate float64) float64 {
	// Reliability = 100 - (error_rate * 1.0 + timeout_rate * 0.5)
	// Timeouts are weighted less heavily since they may be due to caller issues
	reliability := 100.0 - (errorRate*1.0 + timeoutRate*0.5)
	return math.Max(0, reliability)
}

// calculateDeterminismScore calculates the determinism score
func (t *TrustScoreCalculator) calculateDeterminismScore(isDeterministic bool) float64 {
	if isDeterministic {
		return 100
	}
	return 50 // Non-deterministic functions get half score
}

// calculateVolumeScore calculates the volume score using logarithmic scale
func (t *TrustScoreCalculator) calculateVolumeScore(totalCalls int) float64 {
	// Use log scale to prevent large functions from dominating
	// Base score on log10 of calls, capped at 100
	if totalCalls <= 0 {
		return 0
	}
	logScore := math.Log10(float64(totalCalls)) * 25 // log10(100) * 25 = 50
	return math.Min(100, logScore)
}

// calculateDiversityScore calculates the diversity score based on consumer distribution
func (t *TrustScoreCalculator) calculateDiversityScore(uniqueTenants, uniqueUsers, uniqueIPs, totalCalls int) float64 {
	if totalCalls == 0 {
		return 0
	}

	// Calculate diversity percentages
	tenantDiversity := float64(uniqueTenants) / float64(totalCalls) * 100
	userDiversity := float64(uniqueUsers) / float64(totalCalls) * 100
	ipDiversity := float64(uniqueIPs) / float64(totalCalls) * 100

	// Weighted composite:
	// - Tenant diversity (40%): hardest to fake, tied to billing
	// - User diversity (20%): real adoption signal
	// - IP diversity (10%): supporting signal for global distribution
	// - Even distribution bonus (30%): rewards balanced usage across consumers

	// Calculate evenness score (how evenly distributed are calls)
	evennessScore := 0.0
	if uniqueTenants > 1 {
		// Simple Gini-like coefficient approximation
		avgCallsPerTenant := float64(totalCalls) / float64(uniqueTenants)
		variance := float64(totalCalls) * 0.3 / float64(uniqueTenants) // Simplified
		evennessScore = math.Max(0, 100-(variance/avgCallsPerTenant*10))
	}

	diversityScore := (tenantDiversity * 0.4) + (userDiversity * 0.2) + (ipDiversity * 0.1) + (evennessScore * 0.3)

	// Normalize to 0-100
	return math.Min(100, diversityScore)
}

// applyWeights combines all component scores using the weighted formula
func (t *TrustScoreCalculator) applyWeights(success, latency, reliability, determinism, volume, diversity float64) float64 {
	// Trust Score Formula:
	// success_rate * 0.20 + latency_score * 0.15 + reliability_score * 0.20 +
	// determinism_score * 0.10 + volume_score * 0.10 + diversity_score * 0.25

	trustScore := (success * 0.20) +
		(latency * 0.15) +
		(reliability * 0.20) +
		(determinism * 0.10) +
		(volume * 0.10) +
		(diversity * 0.25)

	return trustScore
}

// getTrustLevel returns a human-readable trust level
func (t *TrustScoreCalculator) getTrustLevel(score float64) string {
	switch {
	case score >= 80:
		return "excellent"
	case score >= 60:
		return "good"
	case score >= 40:
		return "fair"
	case score >= 20:
		return "poor"
	default:
		return "very_poor"
	}
}

// CalculateTrustForRating calculates trust score and updates the rating record
func (t *TrustScoreCalculator) CalculateTrustForRating(rating *storage.RegistryFunctionRating, function *storage.RegistryFunction, functionAgeDays int) *TrustScoreResult {
	metrics := &TrustMetrics{
		SuccessRate:      rating.SuccessRate,
		P50LatencyMs:     rating.P50LatencyMs,
		P95LatencyMs:     rating.P95LatencyMs,
		AvgLatencyMs:     rating.AvgLatencyMs,
		TimeoutRate:      rating.TimeoutRate,
		ErrorRate:        rating.ErrorRate,
		TotalCalls:       rating.TotalRatings,
		IsDeterministic:  function != nil && function.DeterministicScore > 0,
		UniqueTenants:    rating.TenantDiversity,
		UniqueUsers:      rating.UserDiversity,
		UniqueIPs:        int(rating.ConsumerDiversity), // Convert percentage to int for calculation
		FunctionAgeDays:  functionAgeDays,
		LastActivityDays: 0, // Will be calculated from trust_updated_at
	}

	result := t.Calculate(metrics)

	// Update rating with trust score
	rating.TrustScore = result.TrustScore

	return result
}

// GetFunctionAgeDays calculates the age of a function in days
func GetFunctionAgeDays(createdAt time.Time) int {
	return int(time.Since(createdAt).Hours() / 24)
}

// GenerateTrustScore generates a trust score for a function given its execution data
func GenerateTrustScore(
	functionID uuid.UUID,
	stats *storage.RegistryFunctionExecution,
	rating *storage.RegistryFunctionRating,
	function *storage.RegistryFunction,
	since time.Time,
) (*TrustScoreResult, error) {
	calc := NewTrustScoreCalculator()

	// Get function age
	functionAgeDays := 0
	if function != nil && !function.CreatedAt.IsZero() {
		functionAgeDays = GetFunctionAgeDays(function.CreatedAt)
	}

	// Calculate metrics from rating
	metrics := &TrustMetrics{
		SuccessRate:     rating.SuccessRate,
		P50LatencyMs:    rating.P50LatencyMs,
		P95LatencyMs:    rating.P95LatencyMs,
		AvgLatencyMs:    rating.AvgLatencyMs,
		TimeoutRate:     rating.TimeoutRate,
		ErrorRate:       rating.ErrorRate,
		TotalCalls:      rating.TotalRatings,
		IsDeterministic: function != nil && function.DeterministicScore > 50,
		UniqueTenants:   rating.TenantDiversity,
		UniqueUsers:     rating.UserDiversity,
		UniqueIPs:       int(rating.ConsumerDiversity),
		FunctionAgeDays: functionAgeDays,
	}

	result := calc.Calculate(metrics)

	// Update rating with calculated trust score
	rating.TrustScore = result.TrustScore

	return result, nil
}
