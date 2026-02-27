package functionregistry

import (
	"testing"
)

func TestTrustScoreCalculator_Calculate(t *testing.T) {
	calc := NewTrustScoreCalculator()

	tests := []struct {
		name     string
		metrics  *TrustMetrics
		wantPass bool
		minScore float64
		maxScore float64
	}{
		{
			name: "insufficient data - less than 10 calls",
			metrics: &TrustMetrics{
				TotalCalls: 5,
			},
			wantPass: false,
			minScore: 0,
			maxScore: 0,
		},
		{
			name: "excellent function - high success, low latency, good diversity",
			metrics: &TrustMetrics{
				SuccessRate:     99.5,
				P50LatencyMs:    25,
				P95LatencyMs:    80,
				TimeoutRate:     0.1,
				ErrorRate:       0.2,
				TotalCalls:      1000,
				IsDeterministic: true,
				UniqueTenants:   50,
				UniqueUsers:     200,
				UniqueIPs:       300,
			},
			wantPass: true,
			minScore: 70,
			maxScore: 100,
		},
		{
			name: "poor function - low success rate",
			metrics: &TrustMetrics{
				SuccessRate:     50.0,
				P50LatencyMs:    100,
				P95LatencyMs:    500,
				TimeoutRate:     10.0,
				ErrorRate:       30.0,
				TotalCalls:      100,
				IsDeterministic: false,
				UniqueTenants:   5,
				UniqueUsers:     10,
				UniqueIPs:       20,
			},
			wantPass: true,
			minScore: 0,
			maxScore: 60,
		},
		{
			name: "deterministic bonus",
			metrics: &TrustMetrics{
				SuccessRate:     95.0,
				P50LatencyMs:    50,
				P95LatencyMs:    200,
				TimeoutRate:     1.0,
				ErrorRate:       2.0,
				TotalCalls:      100,
				IsDeterministic: true,
				UniqueTenants:   10,
				UniqueUsers:     20,
				UniqueIPs:       30,
			},
			wantPass: true,
			minScore: 60,
			maxScore: 90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.Calculate(tt.metrics)
			if tt.wantPass && !result.HasEnoughData {
				t.Errorf("expected HasEnoughData to be true")
			}
			if !tt.wantPass && result.HasEnoughData {
				t.Errorf("expected HasEnoughData to be false")
			}
			if result.HasEnoughData {
				if result.TrustScore < tt.minScore || result.TrustScore > tt.maxScore {
					t.Errorf("trust score = %v, expected range [%v, %v]", result.TrustScore, tt.minScore, tt.maxScore)
				}
			}
		})
	}
}

func TestTrustScoreCalculator_CalculateSuccessScore(t *testing.T) {
	calc := NewTrustScoreCalculator()

	tests := []struct {
		successRate float64
		wantScore   float64
	}{
		{100.0, 100.0},
		{95.0, 95.0},
		{50.0, 50.0},
		{0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := calc.calculateSuccessScore(tt.successRate)
			if got != tt.wantScore {
				t.Errorf("calculateSuccessScore(%v) = %v, want %v", tt.successRate, got, tt.wantScore)
			}
		})
	}
}

func TestTrustScoreCalculator_CalculateLatencyScore(t *testing.T) {
	calc := NewTrustScoreCalculator()

	tests := []struct {
		p50       int
		p95       int
		wantScore float64
	}{
		{10, 30, 100},    // Very fast - weighted ~18ms
		{45, 49, 100},    // Under 50ms weighted
		{90, 199, 80},    // weighted ~133ms
		{200, 499, 60},   // weighted ~320ms
		{500, 999, 40},   // weighted ~700ms
		{1000, 1999, 20}, // weighted ~1400ms
		{2000, 5000, 20}, // Over 1000ms
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := calc.calculateLatencyScore(tt.p50, tt.p95)
			if got != tt.wantScore {
				t.Errorf("calculateLatencyScore(%d, %d) = %v, want %v", tt.p50, tt.p95, got, tt.wantScore)
			}
		})
	}
}

func TestTrustScoreCalculator_CalculateReliabilityScore(t *testing.T) {
	calc := NewTrustScoreCalculator()

	tests := []struct {
		errorRate   float64
		timeoutRate float64
		wantScore   float64
	}{
		{0.0, 0.0, 100.0},  // Perfect
		{5.0, 0.0, 95.0},   // Only errors
		{0.0, 5.0, 97.5},   // Only timeouts (weighted less)
		{10.0, 10.0, 85.0}, // Both
		{50.0, 50.0, 25.0}, // High failure
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := calc.calculateReliabilityScore(tt.errorRate, tt.timeoutRate)
			if got != tt.wantScore {
				t.Errorf("calculateReliabilityScore(%v, %v) = %v, want %v", tt.errorRate, tt.timeoutRate, got, tt.wantScore)
			}
		})
	}
}

func TestTrustScoreCalculator_CalculateDeterminismScore(t *testing.T) {
	calc := NewTrustScoreCalculator()

	tests := []struct {
		isDeterministic bool
		wantScore       float64
	}{
		{true, 100.0},
		{false, 50.0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := calc.calculateDeterminismScore(tt.isDeterministic)
			if got != tt.wantScore {
				t.Errorf("calculateDeterminismScore(%v) = %v, want %v", tt.isDeterministic, got, tt.wantScore)
			}
		})
	}
}

func TestTrustScoreCalculator_CalculateVolumeScore(t *testing.T) {
	calc := NewTrustScoreCalculator()

	tests := []struct {
		totalCalls int
		wantScore  float64
	}{
		{0, 0},
		{1, 0},       // log10(1) = 0
		{10, 25},     // log10(10) * 25 = 25
		{100, 50},    // log10(100) * 25 = 50
		{1000, 75},   // log10(1000) * 25 = 75
		{10000, 100}, // capped at 100
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := calc.calculateVolumeScore(tt.totalCalls)
			if got != tt.wantScore {
				t.Errorf("calculateVolumeScore(%d) = %v, want %v", tt.totalCalls, got, tt.wantScore)
			}
		})
	}
}

func TestTrustScoreCalculator_GetTrustLevel(t *testing.T) {
	calc := NewTrustScoreCalculator()

	tests := []struct {
		score     float64
		wantLevel string
	}{
		{100, "excellent"},
		{80, "excellent"},
		{60, "good"},
		{40, "fair"},
		{20, "poor"},
		{10, "very_poor"},
		{0, "very_poor"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := calc.getTrustLevel(tt.score)
			if got != tt.wantLevel {
				t.Errorf("getTrustLevel(%v) = %v, want %v", tt.score, got, tt.wantLevel)
			}
		})
	}
}

func TestTrustScoreCalculator_ApplyWeights(t *testing.T) {
	calc := NewTrustScoreCalculator()

	// Test the weighted formula
	// trust_score = success * 0.20 + latency * 0.15 + reliability * 0.20 + determinism * 0.10 + volume * 0.10 + diversity * 0.25
	success := 100.0
	latency := 100.0
	reliability := 100.0
	determinism := 100.0
	volume := 100.0
	diversity := 100.0

	want := 100.0
	got := calc.applyWeights(success, latency, reliability, determinism, volume, diversity)
	if got != want {
		t.Errorf("applyWeights with all 100 = %v, want %v", got, want)
	}

	// Test with zeros
	got = calc.applyWeights(0, 0, 0, 0, 0, 0)
	if got != 0 {
		t.Errorf("applyWeights with all 0 = %v, want 0", got)
	}

	// Test mixed values
	// 80*0.20 + 80*0.15 + 80*0.20 + 50*0.10 + 50*0.10 + 50*0.25 = 16 + 12 + 16 + 5 + 5 + 12.5 = 66.5
	got = calc.applyWeights(80, 80, 80, 50, 50, 50)
	want = 66.5
	if got != want {
		t.Errorf("applyWeights(80, 80, 80, 50, 50, 50) = %v, want %v", got, want)
	}
}

func TestTrustScoreCalculator_DiversityScore(t *testing.T) {
	calc := NewTrustScoreCalculator()

	tests := []struct {
		name          string
		uniqueTenants int
		uniqueUsers   int
		uniqueIPs     int
		totalCalls    int
		maxScore      float64 // Should be less than or equal to this
	}{
		{
			name:          "high diversity",
			uniqueTenants: 100,
			uniqueUsers:   500,
			uniqueIPs:     1000,
			totalCalls:    10000,
			maxScore:      100, // Should be high but capped at 100
		},
		{
			name:          "low diversity",
			uniqueTenants: 1,
			uniqueUsers:   1,
			uniqueIPs:     1,
			totalCalls:    100,
			maxScore:      10, // Should be low
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.calculateDiversityScore(tt.uniqueTenants, tt.uniqueUsers, tt.uniqueIPs, tt.totalCalls)
			if got > tt.maxScore {
				t.Errorf("calculateDiversityScore() = %v, expected <= %v", got, tt.maxScore)
			}
		})
	}
}
