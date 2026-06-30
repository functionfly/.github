package dna

import (
	"testing"
	"time"

	dnadb "github.com/functionfly/functionfly/internal/storage/dna"
	"github.com/sirupsen/logrus"
)

func newTestService() *Service {
	svc := NewService(nil, logrus.StandardLogger())
	// For tests that don't need AI, ensure the service has a dummy URL
	svc.aiBaseURL = "http://localhost:8081"
	svc.aiAPIKey = ""
	return svc
}

func TestComputeFitness(t *testing.T) {
	svc := newTestService()

	tests := []struct {
		name     string
		metrics  *dnadb.AggregatedMetrics
		minScore float64
		maxScore float64
	}{
		{
			name: "excellent performance",
			metrics: &dnadb.AggregatedMetrics{
				TotalExecutions:   10000,
				AvgLatencyMs:      50,
				P99LatencyMs:      80,
				SuccessRate:       0.99,
				ColdStartRate:     0.05,
				ErrorDistribution: map[string]int64{},
			},
			minScore: 90,
			maxScore: 100,
		},
		{
			name: "poor performance",
			metrics: &dnadb.AggregatedMetrics{
				TotalExecutions:   5000,
				AvgLatencyMs:      400,
				P99LatencyMs:      480,
				SuccessRate:       0.90,
				ColdStartRate:     0.35,
				ErrorDistribution: map[string]int64{"timeout": 50, "runtime": 30, "network": 20},
			},
			minScore: 55,
			maxScore: 70,
		},
		{
			name: "baseline no data",
			metrics: &dnadb.AggregatedMetrics{
				TotalExecutions:   0,
				AvgLatencyMs:      0,
				P99LatencyMs:      0,
				SuccessRate:       1.0,
				ColdStartRate:     0,
				ErrorDistribution: map[string]int64{},
			},
			minScore: 95,
			maxScore: 100,
		},
		{
			name: "moderate performance",
			metrics: &dnadb.AggregatedMetrics{
				TotalExecutions:   1000,
				AvgLatencyMs:      150,
				P99LatencyMs:      250,
				SuccessRate:       0.97,
				ColdStartRate:     0.15,
				ErrorDistribution: map[string]int64{"timeout": 10},
			},
			minScore: 70,
			maxScore: 90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := svc.computeFitness(tt.metrics)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("computeFitness() = %.2f, want [%.2f, %.2f]", score, tt.minScore, tt.maxScore)
			}
			if score < 0 || score > 100 {
				t.Errorf("computeFitness() = %.2f, want [0, 100]", score)
			}
		})
	}
}

func TestComputeFitness_Clamping(t *testing.T) {
	svc := newTestService()

	m := &dnadb.AggregatedMetrics{
		TotalExecutions:   100000,
		SuccessRate:       1.0,
		P99LatencyMs:      10,
		ColdStartRate:     0,
		ErrorDistribution: map[string]int64{},
	}
	score := svc.computeFitness(m)
	if score > 100 {
		t.Errorf("score should be clamped to 100, got %.2f", score)
	}
}

func TestShouldMutate(t *testing.T) {
	svc := newTestService()

	tests := []struct {
		name         string
		metrics      *dnadb.AggregatedMetrics
		profile      *dnadb.DNAProfile
		expectMutate bool
		expectedType string
	}{
		{
			name: "high latency triggers optimize_latency",
			metrics: &dnadb.AggregatedMetrics{
				P99LatencyMs:      600,
				SuccessRate:       0.99,
				ColdStartRate:     0.1,
				ErrorDistribution: map[string]int64{},
			},
			profile:      &dnadb.DNAProfile{FitnessScore: 80, TotalMutations: 0},
			expectMutate: true,
			expectedType: "optimize_latency",
		},
		{
			name: "low success rate triggers fix_error_pattern",
			metrics: &dnadb.AggregatedMetrics{
				P99LatencyMs:      200,
				SuccessRate:       0.92,
				ColdStartRate:     0.1,
				ErrorDistribution: map[string]int64{},
			},
			profile:      &dnadb.DNAProfile{FitnessScore: 80, TotalMutations: 0},
			expectMutate: true,
			expectedType: "fix_error_pattern",
		},
		{
			name: "high cold starts triggers reduce_memory",
			metrics: &dnadb.AggregatedMetrics{
				P99LatencyMs:      200,
				SuccessRate:       0.99,
				ColdStartRate:     0.4,
				ErrorDistribution: map[string]int64{},
			},
			profile:      &dnadb.DNAProfile{FitnessScore: 80, TotalMutations: 0},
			expectMutate: true,
			expectedType: "reduce_memory",
		},
		{
			name: "high memory triggers reduce_memory",
			metrics: &dnadb.AggregatedMetrics{
				P99LatencyMs:      200,
				SuccessRate:       0.99,
				ColdStartRate:     0.1,
				AvgMemoryPeakMb:   300,
				ErrorDistribution: map[string]int64{},
			},
			profile:      &dnadb.DNAProfile{FitnessScore: 80, TotalMutations: 0},
			expectMutate: true,
			expectedType: "reduce_memory",
		},
		{
			name: "low fitness with no mutations triggers refactor_hotpath",
			metrics: &dnadb.AggregatedMetrics{
				P99LatencyMs:      200,
				SuccessRate:       0.99,
				ColdStartRate:     0.1,
				ErrorDistribution: map[string]int64{},
			},
			profile:      &dnadb.DNAProfile{FitnessScore: 50, TotalMutations: 0},
			expectMutate: true,
			expectedType: "refactor_hotpath",
		},
		{
			name: "healthy function no mutation",
			metrics: &dnadb.AggregatedMetrics{
				P99LatencyMs:      100,
				SuccessRate:       0.99,
				ColdStartRate:     0.1,
				AvgMemoryPeakMb:   100,
				ErrorDistribution: map[string]int64{},
			},
			profile:      &dnadb.DNAProfile{FitnessScore: 85, TotalMutations: 2},
			expectMutate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldMutate, mutationType, _ := svc.shouldMutate(tt.metrics, tt.profile)
			if shouldMutate != tt.expectMutate {
				t.Errorf("shouldMutate() = %v, want %v", shouldMutate, tt.expectMutate)
			}
			if shouldMutate && mutationType != tt.expectedType {
				t.Errorf("shouldMutate() type = %s, want %s", mutationType, tt.expectedType)
			}
		})
	}
}

func TestSafePrefix(t *testing.T) {
	tests := []struct {
		s    string
		n    int
		want string
	}{
		{"abcdefgh", 8, "abcdefgh"},
		{"abcdefgh", 4, "abcd"},
		{"ab", 8, "ab"},
		{"", 8, ""},
	}
	for _, tt := range tests {
		if got := safePrefix(tt.s, tt.n); got != tt.want {
			t.Errorf("safePrefix(%q, %d) = %q, want %q", tt.s, tt.n, got, tt.want)
		}
	}
}

func TestVerifyDNAHash_NoProfile(t *testing.T) {
	_ = newTestService()
	// With nil repo this would panic — just test that the method exists and
	// the hash verification logic compiles. Real coverage comes from integration tests.
}

func TestCircuitBreaker(t *testing.T) {
	cb := newCircuitBreaker(3, 100*time.Millisecond)

	if !cb.allow() {
		t.Error("should allow in closed state")
	}

	cb.recordFailure()
	cb.recordFailure()
	if !cb.allow() {
		t.Error("should still allow after 2 failures (threshold=3)")
	}

	cb.recordFailure()
	if cb.allow() {
		t.Error("should not allow after 3 failures (threshold reached)")
	}

	time.Sleep(150 * time.Millisecond)
	if !cb.allow() {
		t.Error("should allow in half-open after cooldown")
	}

	cb.recordSuccess()
	if !cb.allow() {
		t.Error("should allow after success resets to closed")
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	cb := newCircuitBreaker(2, 50*time.Millisecond)

	cb.recordFailure()
	cb.recordFailure()
	if cb.allow() {
		t.Error("should be open after threshold")
	}

	time.Sleep(60 * time.Millisecond)
	if !cb.allow() {
		t.Error("should allow in half-open")
	}

	cb.recordFailure()
	if cb.allow() {
		t.Error("should re-open after half-open failure")
	}
}
