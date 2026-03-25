package registry

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestCalculateReliabilityScore tests the reliability score calculation
func TestCalculateReliabilityScore(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *ExecutionMetrics
		expected float64
	}{
		{
			name:     "nil metrics returns default",
			metrics:  nil,
			expected: 50.0,
		},
		{
			name:     "zero calls returns default",
			metrics:  &ExecutionMetrics{TotalCalls: 0},
			expected: 50.0,
		},
		{
			name:     "100% success rate",
			metrics:  &ExecutionMetrics{TotalCalls: 100, SuccessRate: 100.0},
			expected: 100.0,
		},
		{
			name:     "95% success rate",
			metrics:  &ExecutionMetrics{TotalCalls: 100, SuccessRate: 95.0},
			expected: 95.0,
		},
		{
			name:     "50% success rate",
			metrics:  &ExecutionMetrics{TotalCalls: 100, SuccessRate: 50.0},
			expected: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateReliabilityScore(tt.metrics)
			if result != tt.expected {
				t.Errorf("calculateReliabilityScore() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateLatencyScore tests the latency score calculation
func TestCalculateLatencyScore(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *ExecutionMetrics
		expected float64
	}{
		{
			name:     "nil metrics returns default",
			metrics:  nil,
			expected: 50.0,
		},
		{
			name:     "zero calls returns default",
			metrics:  &ExecutionMetrics{TotalCalls: 0},
			expected: 50.0,
		},
		{
			name:     "p95 < 50ms",
			metrics:  &ExecutionMetrics{TotalCalls: 100, LatencyP95: 40},
			expected: 100.0,
		},
		{
			name:     "p95 < 100ms",
			metrics:  &ExecutionMetrics{TotalCalls: 100, LatencyP95: 90},
			expected: 90.0,
		},
		{
			name:     "p95 < 200ms",
			metrics:  &ExecutionMetrics{TotalCalls: 100, LatencyP95: 150},
			expected: 80.0,
		},
		{
			name:     "p95 < 500ms",
			metrics:  &ExecutionMetrics{TotalCalls: 100, LatencyP95: 400},
			expected: 70.0,
		},
		{
			name:     "p95 < 1000ms",
			metrics:  &ExecutionMetrics{TotalCalls: 100, LatencyP95: 800},
			expected: 60.0,
		},
		{
			name:     "p95 < 2000ms",
			metrics:  &ExecutionMetrics{TotalCalls: 100, LatencyP95: 1500},
			expected: 50.0,
		},
		{
			name:     "p95 < 5000ms",
			metrics:  &ExecutionMetrics{TotalCalls: 100, LatencyP95: 4000},
			expected: 40.0,
		},
		{
			name:     "p95 >= 5000ms",
			metrics:  &ExecutionMetrics{TotalCalls: 100, LatencyP95: 6000},
			expected: 30.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateLatencyScore(tt.metrics)
			if result != tt.expected {
				t.Errorf("calculateLatencyScore() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateErrorRateScore tests the error rate score calculation
func TestCalculateErrorRateScore(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *ExecutionMetrics
		expected float64
	}{
		{
			name:     "nil metrics returns default",
			metrics:  nil,
			expected: 50.0,
		},
		{
			name:     "zero calls returns default",
			metrics:  &ExecutionMetrics{TotalCalls: 0},
			expected: 50.0,
		},
		{
			name:     "zero error rate",
			metrics:  &ExecutionMetrics{TotalCalls: 100, ErrorRate: 0, TimeoutRate: 0},
			expected: 100.0,
		},
		{
			name:     "error rate < 0.1%",
			metrics:  &ExecutionMetrics{TotalCalls: 100, ErrorRate: 0.05, TimeoutRate: 0},
			expected: 95.0,
		},
		{
			name:     "error rate < 0.5%",
			metrics:  &ExecutionMetrics{TotalCalls: 100, ErrorRate: 0.3, TimeoutRate: 0},
			expected: 85.0,
		},
		{
			name:     "error rate < 1.0%",
			metrics:  &ExecutionMetrics{TotalCalls: 100, ErrorRate: 0.7, TimeoutRate: 0},
			expected: 75.0,
		},
		{
			name:     "error rate < 2.0%",
			metrics:  &ExecutionMetrics{TotalCalls: 100, ErrorRate: 1.5, TimeoutRate: 0},
			expected: 60.0,
		},
		{
			name:     "error rate < 5.0%",
			metrics:  &ExecutionMetrics{TotalCalls: 100, ErrorRate: 3.0, TimeoutRate: 0},
			expected: 40.0,
		},
		{
			name:     "error rate >= 5.0%",
			metrics:  &ExecutionMetrics{TotalCalls: 100, ErrorRate: 6.0, TimeoutRate: 0},
			expected: 20.0,
		},
		{
			name:     "combined error + timeout < 0.1%",
			metrics:  &ExecutionMetrics{TotalCalls: 100, ErrorRate: 0.05, TimeoutRate: 0.04},
			expected: 95.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateErrorRateScore(tt.metrics)
			if result != tt.expected {
				t.Errorf("calculateErrorRateScore() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateUserRatingScore tests the user rating score calculation
func TestCalculateUserRatingScore(t *testing.T) {
	tests := []struct {
		name     string
		rating   *RegistryFunctionRating
		expected float64
	}{
		{
			name:     "nil rating returns default",
			rating:   nil,
			expected: 50.0,
		},
		{
			name:     "zero ratings returns default",
			rating:   &RegistryFunctionRating{TotalRatings: 0},
			expected: 50.0,
		},
		{
			name:     "perfect rating",
			rating:   &RegistryFunctionRating{TotalRatings: 10, OverallScore: 100.0},
			expected: 100.0,
		},
		{
			name:     "average rating",
			rating:   &RegistryFunctionRating{TotalRatings: 10, OverallScore: 75.0},
			expected: 75.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateUserRatingScore(tt.rating)
			if result != tt.expected {
				t.Errorf("calculateUserRatingScore() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateVerificationBonus tests the verification bonus calculation
func TestCalculateVerificationBonus(t *testing.T) {
	tests := []struct {
		name            string
		isVerified      bool
		verificationLevel string
		expected        float64
	}{
		{
			name:            "not verified",
			isVerified:      false,
			verificationLevel: "none",
			expected:        0.0,
		},
		{
			name:            "verified basic",
			isVerified:      true,
			verificationLevel: "basic",
			expected:        5.0,
		},
		{
			name:            "verified standard",
			isVerified:      true,
			verificationLevel: "standard",
			expected:        10.0,
		},
		{
			name:            "verified enterprise",
			isVerified:      true,
			verificationLevel: "enterprise",
			expected:        15.0,
		},
		{
			name:            "verified unknown level",
			isVerified:      true,
			verificationLevel: "unknown",
			expected:        5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateVerificationBonus(tt.isVerified, tt.verificationLevel)
			if result != tt.expected {
				t.Errorf("calculateVerificationBonus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestDetermineTrustTier tests the trust tier determination
func TestDetermineTrustTier(t *testing.T) {
	tests := []struct {
		name       string
		trustScore float64
		isVerified bool
		expected   TrustTier
	}{
		{
			name:       "score >= 90 and verified",
			trustScore: 95.0,
			isVerified: true,
			expected:   TrustTierHighlyTrusted,
		},
		{
			name:       "score >= 90 and not verified",
			trustScore: 92.0,
			isVerified: false,
			expected:   TrustTierVerified,
		},
		{
			name:       "score >= 70",
			trustScore: 75.0,
			isVerified: false,
			expected:   TrustTierVerified,
		},
		{
			name:       "score >= 50",
			trustScore: 55.0,
			isVerified: false,
			expected:   TrustTierTrusted,
		},
		{
			name:       "score < 50",
			trustScore: 45.0,
			isVerified: false,
			expected:   TrustTierUntrusted,
		},
		{
			name:       "score == 50",
			trustScore: 50.0,
			isVerified: false,
			expected:   TrustTierTrusted,
		},
		{
			name:       "score == 70",
			trustScore: 70.0,
			isVerified: false,
			expected:   TrustTierVerified,
		},
		{
			name:       "score == 90",
			trustScore: 90.0,
			isVerified: false,
			expected:   TrustTierVerified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineTrustTier(tt.trustScore, tt.isVerified)
			if result != tt.expected {
				t.Errorf("determineTrustTier() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestTrustHistoryTableName tests the TrustHistory table name
func TestTrustHistoryTableName(t *testing.T) {
	history := TrustHistory{}
	if history.TableName() != "trust_history" {
		t.Errorf("TrustHistory.TableName() = %v, want trust_history", history.TableName())
	}
}

// TestExecutionMetricsTableName tests the ExecutionMetrics table name
func TestExecutionMetricsTableName(t *testing.T) {
	metrics := ExecutionMetrics{}
	if metrics.TableName() != "execution_metrics" {
		t.Errorf("ExecutionMetrics.TableName() = %v, want execution_metrics", metrics.TableName())
	}
}

// TestTrustScoreWeightsConfigTableName tests the TrustScoreWeightsConfig table name
func TestTrustScoreWeightsConfigTableName(t *testing.T) {
	config := TrustScoreWeightsConfig{}
	if config.TableName() != "trust_score_weights" {
		t.Errorf("TrustScoreWeightsConfig.TableName() = %v, want trust_score_weights", config.TableName())
	}
}

// TestTrustScoreJobTableName tests the TrustScoreJob table name
func TestTrustScoreJobTableName(t *testing.T) {
	job := TrustScoreJob{}
	if job.TableName() != "trust_score_jobs" {
		t.Errorf("TrustScoreJob.TableName() = %v, want trust_score_jobs", job.TableName())
	}
}

// TestDefaultTrustScoreWeights tests the default weights
func TestDefaultTrustScoreWeights(t *testing.T) {
	weights := DefaultTrustScoreWeights()

	// Verify all weights are set
	if weights.Reliability != 0.30 {
		t.Errorf("DefaultTrustScoreWeights().Reliability = %v, want 0.30", weights.Reliability)
	}
	if weights.Latency != 0.20 {
		t.Errorf("DefaultTrustScoreWeights().Latency = %v, want 0.20", weights.Latency)
	}
	if weights.ErrorRate != 0.20 {
		t.Errorf("DefaultTrustScoreWeights().ErrorRate = %v, want 0.20", weights.ErrorRate)
	}
	if weights.UserRating != 0.15 {
		t.Errorf("DefaultTrustScoreWeights().UserRating = %v, want 0.15", weights.UserRating)
	}
	if weights.Verification != 0.15 {
		t.Errorf("DefaultTrustScoreWeights().Verification = %v, want 0.15", weights.Verification)
	}

	// Verify weights sum to 1.0
	total := weights.Reliability + weights.Latency + weights.ErrorRate + weights.UserRating + weights.Verification
	if total != 1.0 {
		t.Errorf("DefaultTrustScoreWeights() total = %v, want 1.0", total)
	}
}

// TestTrustTierConstants tests trust tier constants
func TestTrustTierConstants(t *testing.T) {
	if TrustTierUntrusted != "untrusted" {
		t.Errorf("TrustTierUntrusted = %v, want untrusted", TrustTierUntrusted)
	}
	if TrustTierTrusted != "trusted" {
		t.Errorf("TrustTierTrusted = %v, want trusted", TrustTierTrusted)
	}
	if TrustTierVerified != "verified" {
		t.Errorf("TrustTierVerified = %v, want verified", TrustTierVerified)
	}
	if TrustTierHighlyTrusted != "highly_trusted" {
		t.Errorf("TrustTierHighlyTrusted = %v, want highly_trusted", TrustTierHighlyTrusted)
	}
}

// TestTimePtr tests the timePtr helper function
func TestTimePtr(t *testing.T) {
	now := time.Now()
	ptr := timePtr(now)

	if ptr == nil {
		t.Error("timePtr() returned nil")
		return
	}

	if !ptr.Equal(now) {
		t.Errorf("timePtr() = %v, want %v", *ptr, now)
	}
}

// TestTrustScoreResponseStructure tests that TrustScoreResponse has all expected fields
func TestTrustScoreResponseStructure(t *testing.T) {
	response := TrustScoreResponse{
		FunctionID: uuid.New(),
		TrustScore: 85.5,
		TrustTier: TrustTierVerified,
		IsVerified: true,
		VerificationLevel: "standard",
		LastUpdated: time.Now(),
		WindowStart: time.Now().Add(-24 * time.Hour),
		WindowEnd: time.Now(),
	}

	// Verify fields are set correctly
	if response.TrustScore != 85.5 {
		t.Errorf("TrustScoreResponse.TrustScore = %v, want 85.5", response.TrustScore)
	}
	if response.TrustTier != TrustTierVerified {
		t.Errorf("TrustScoreResponse.TrustTier = %v, want verified", response.TrustTier)
	}
	if !response.IsVerified {
		t.Error("TrustScoreResponse.IsVerified should be true")
	}
}

// TestTrustHistoryResponseStructure tests that TrustHistoryResponse has all expected fields
func TestTrustHistoryResponseStructure(t *testing.T) {
	response := TrustHistoryResponse{
		FunctionID: uuid.New(),
		History: []TrustHistory{
			{TrustScore: 90.0},
			{TrustScore: 85.0},
		},
		TotalCount: 2,
		Page: 1,
		PageSize: 20,
	}

	// Verify fields are set correctly
	if len(response.History) != 2 {
		t.Errorf("TrustHistoryResponse.History length = %v, want 2", len(response.History))
	}
	if response.TotalCount != 2 {
		t.Errorf("TrustHistoryResponse.TotalCount = %v, want 2", response.TotalCount)
	}
	if response.Page != 1 {
		t.Errorf("TrustHistoryResponse.Page = %v, want 1", response.Page)
	}
	if response.PageSize != 20 {
		t.Errorf("TrustHistoryResponse.PageSize = %v, want 20", response.PageSize)
	}
}
