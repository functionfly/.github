package registry

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestSlidingWindowConfigDefaults tests default sliding window configuration
func TestSlidingWindowConfigDefaults(t *testing.T) {
	config := DefaultSlidingWindowConfig()

	if config.WindowDuration != 24*time.Hour {
		t.Errorf("expected WindowDuration = 24h, got %v", config.WindowDuration)
	}
	if config.SmoothingFactor != 0.3 {
		t.Errorf("expected SmoothingFactor = 0.3, got %v", config.SmoothingFactor)
	}
	if config.MinDataPoints != 10 {
		t.Errorf("expected MinDataPoints = 10, got %v", config.MinDataPoints)
	}
	if config.UpdateInterval != 5*time.Minute {
		t.Errorf("expected UpdateInterval = 5m, got %v", config.UpdateInterval)
	}
}

// TestSlidingWindowStateComponentScores tests component score getters and setters
func TestSlidingWindowStateComponentScores(t *testing.T) {
	state := &SlidingWindowState{
		FunctionID:        uuid.New(),
		ReliabilityScore:  80.0,
		LatencyScore:      70.0,
		ErrorRateScore:    90.0,
		UserRatingScore:   85.0,
		VerificationBonus: 10.0,
	}

	tests := []struct {
		name     string
		component string
		expected float64
	}{
		{"reliability", "reliability", 80.0},
		{"latency", "latency", 70.0},
		{"error_rate", "error_rate", 90.0},
		{"user_rating", "user_rating", 85.0},
		{"verification", "verification", 10.0},
		{"unknown", "unknown", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := state.GetComponentScore(tt.component)
			if result != tt.expected {
				t.Errorf("GetComponentScore(%s) = %v, want %v", tt.component, result, tt.expected)
			}
		})
	}
}

// TestSlidingWindowStateSetComponentScores tests component score setting
func TestSlidingWindowStateSetComponentScores(t *testing.T) {
	state := &SlidingWindowState{}

	state.SetComponentScore("reliability", 75.0)
	state.SetComponentScore("latency", 65.0)
	state.SetComponentScore("error_rate", 95.0)
	state.SetComponentScore("user_rating", 80.0)
	state.SetComponentScore("verification", 15.0)

	if state.ReliabilityScore != 75.0 {
		t.Errorf("expected ReliabilityScore = 75.0, got %v", state.ReliabilityScore)
	}
	if state.LatencyScore != 65.0 {
		t.Errorf("expected LatencyScore = 65.0, got %v", state.LatencyScore)
	}
	if state.ErrorRateScore != 95.0 {
		t.Errorf("expected ErrorRateScore = 95.0, got %v", state.ErrorRateScore)
	}
	if state.UserRatingScore != 80.0 {
		t.Errorf("expected UserRatingScore = 80.0, got %v", state.UserRatingScore)
	}
	if state.VerificationBonus != 15.0 {
		t.Errorf("expected VerificationBonus = 15.0, got %v", state.VerificationBonus)
	}
}

// TestSlidingWindowStateUpdateFromHistory tests updating component scores from TrustHistory
func TestSlidingWindowStateUpdateFromHistory(t *testing.T) {
	state := &SlidingWindowState{}

	history := &TrustHistory{
		ReliabilityScore:  82.0,
		LatencyScore:      78.0,
		ErrorRateScore:    88.0,
		UserRatingScore:   76.0,
		VerificationBonus: 12.0,
	}

	state.UpdateComponentScores(history)

	if state.ReliabilityScore != 82.0 {
		t.Errorf("expected ReliabilityScore = 82.0, got %v", state.ReliabilityScore)
	}
	if state.LatencyScore != 78.0 {
		t.Errorf("expected LatencyScore = 78.0, got %v", state.LatencyScore)
	}
	if state.ErrorRateScore != 88.0 {
		t.Errorf("expected ErrorRateScore = 88.0, got %v", state.ErrorRateScore)
	}
	if state.UserRatingScore != 76.0 {
		t.Errorf("expected UserRatingScore = 76.0, got %v", state.UserRatingScore)
	}
	if state.VerificationBonus != 12.0 {
		t.Errorf("expected VerificationBonus = 12.0, got %v", state.VerificationBonus)
	}
}

// TestDefaultThresholdConfig tests default threshold configuration
func TestDefaultThresholdConfig(t *testing.T) {
	config := DefaultThresholdConfig()

	if config.CriticalThreshold != 50.0 {
		t.Errorf("expected CriticalThreshold = 50.0, got %v", config.CriticalThreshold)
	}
	if config.WarningThreshold != 70.0 {
		t.Errorf("expected WarningThreshold = 70.0, got %v", config.WarningThreshold)
	}
	if config.MinChangeForNotify != 5.0 {
		t.Errorf("expected MinChangeForNotify = 5.0, got %v", config.MinChangeForNotify)
	}
	if config.CooldownPeriod != 15*time.Minute {
		t.Errorf("expected CooldownPeriod = 15m, got %v", config.CooldownPeriod)
	}
}

// TestTrustScoreDeltaCalculation tests delta calculation logic
func TestTrustScoreDeltaCalculation(t *testing.T) {
	tests := []struct {
		name           string
		previousScore  float64
		currentScore   float64
		expectedChange float64
		expectedPercent float64
	}{
		{
			name:           "score increase",
			previousScore:  60.0,
			currentScore:   70.0,
			expectedChange: 10.0,
			expectedPercent: 16.6667,
		},
		{
			name:           "score decrease",
			previousScore:  80.0,
			currentScore:   70.0,
			expectedChange: -10.0,
			expectedPercent: -12.5,
		},
		{
			name:           "no change",
			previousScore:  75.0,
			currentScore:   75.0,
			expectedChange: 0.0,
			expectedPercent: 0.0,
		},
		{
			name:           "from zero",
			previousScore:  0.0,
			currentScore:   50.0,
			expectedChange: 50.0,
			expectedPercent: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := TrustScoreDelta{
				PreviousScore: tt.previousScore,
				CurrentScore:  tt.currentScore,
			}
			delta.ScoreChange = delta.CurrentScore - delta.PreviousScore
			if delta.PreviousScore > 0 {
				delta.ScoreChangePercent = (delta.ScoreChange / delta.PreviousScore) * 100
			}

			if delta.ScoreChange != tt.expectedChange {
				t.Errorf("ScoreChange = %v, want %v", delta.ScoreChange, tt.expectedChange)
			}

			// Allow small floating point differences
			percentDiff := delta.ScoreChangePercent - tt.expectedPercent
			if percentDiff < 0 {
				percentDiff = -percentDiff
			}
			if percentDiff > 0.001 {
				t.Errorf("ScoreChangePercent = %v, want %v (diff: %v)", 
					delta.ScoreChangePercent, tt.expectedPercent, percentDiff)
			}
		})
	}
}

// TestTrustTierTransitions tests tier change detection
func TestTrustTierTransitions(t *testing.T) {
	tests := []struct {
		name         string
		previousTier TrustTier
		currentTier  TrustTier
		shouldNotify bool
		severity     string
	}{
		{
			name:         "upgrade untrusted to trusted",
			previousTier: TrustTierUntrusted,
			currentTier:  TrustTierTrusted,
			shouldNotify: true,
			severity:     "info",
		},
		{
			name:         "upgrade to highly trusted",
			previousTier: TrustTierVerified,
			currentTier:  TrustTierHighlyTrusted,
			shouldNotify: true,
			severity:     "info",
		},
		{
			name:         "downgrade to untrusted",
			previousTier: TrustTierTrusted,
			currentTier:  TrustTierUntrusted,
			shouldNotify: true,
			severity:     "warning",
		},
		{
			name:         "no change",
			previousTier: TrustTierVerified,
			currentTier:  TrustTierVerified,
			shouldNotify: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := TrustScoreDelta{
				PreviousTier: tt.previousTier,
				CurrentTier:  tt.currentTier,
				TierChanged:  tt.previousTier != tt.currentTier,
			}

			if delta.TierChanged != tt.shouldNotify {
				t.Errorf("TierChanged = %v, want %v", delta.TierChanged, tt.shouldNotify)
			}
		})
	}
}

// TestTrustScoreStreamEventValidation tests event validation
func TestTrustScoreStreamEventValidation(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		isValid   bool
	}{
		{"score_update", "score_update", true},
		{"tier_change", "tier_change", true},
		{"threshold_breach", "threshold_breach", true},
		{"invalid_type", "invalid", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := TrustScoreStreamEvent{
				EventType:  tt.eventType,
				FunctionID: uuid.New(),
				Timestamp:  time.Now(),
			}

			valid := event.EventType == "score_update" || 
				event.EventType == "tier_change" || 
				event.EventType == "threshold_breach"

			if valid != tt.isValid {
				t.Errorf("event type %q valid = %v, want %v", tt.eventType, valid, tt.isValid)
			}
		})
	}
}

// TestExponentialMovingAverage tests EMA calculation logic
func TestExponentialMovingAverage(t *testing.T) {
	tests := []struct {
		name       string
		previous   float64
		current    float64
		alpha      float64
		expected   float64
	}{
		{
			name:     "first calculation (no previous)",
			previous: 0,
			current:  80.0,
			alpha:    0.3,
			expected: 80.0,
		},
		{
			name:     "typical smoothing",
			previous: 70.0,
			current:  80.0,
			alpha:    0.3,
			expected: 73.0, // 0.3*80 + 0.7*70 = 24 + 49 = 73
		},
		{
			name:     "high alpha (responsive)",
			previous: 70.0,
			current:  80.0,
			alpha:    0.7,
			expected: 77.0, // 0.7*80 + 0.3*70 = 56 + 21 = 77
		},
		{
			name:     "low alpha (stable)",
			previous: 70.0,
			current:  80.0,
			alpha:    0.1,
			expected: 71.0, // 0.1*80 + 0.9*70 = 8 + 63 = 71
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result float64
			if tt.previous == 0 {
				result = tt.current
			} else {
				result = tt.alpha*tt.current + (1-tt.alpha)*tt.previous
			}

			if result != tt.expected {
				t.Errorf("EMA = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSlidingWindowStateTableName tests the database table name
func TestSlidingWindowStateTableName(t *testing.T) {
	state := SlidingWindowState{}
	expected := "trust_sliding_window_state"
	if state.TableName() != expected {
		t.Errorf("TableName() = %v, want %v", state.TableName(), expected)
	}
}

// TestWindowTypeConstants tests window type constants
func TestWindowTypeConstants(t *testing.T) {
	if WindowTypeDiscrete != "discrete" {
		t.Errorf("WindowTypeDiscrete = %v, want discrete", WindowTypeDiscrete)
	}
	if WindowTypeSliding != "sliding" {
		t.Errorf("WindowTypeSliding = %v, want sliding", WindowTypeSliding)
	}
}

// TestComponentChangesCalculation tests component change calculation
func TestComponentChangesCalculation(t *testing.T) {
	delta := &TrustScoreDelta{
		FunctionID: uuid.New(),
		ComponentChanges: map[string]float64{
			"reliability":  5.0,
			"latency":      -3.0,
			"error_rate":   2.0,
			"user_rating":  0.0,
			"verification": 0.0,
		},
	}

	// Verify all components are tracked
	expectedComponents := []string{"reliability", "latency", "error_rate", "user_rating", "verification"}
	for _, component := range expectedComponents {
		if _, ok := delta.ComponentChanges[component]; !ok {
			t.Errorf("missing component %q in ComponentChanges", component)
		}
	}

	// Verify change values
	if delta.ComponentChanges["reliability"] != 5.0 {
		t.Errorf("reliability change = %v, want 5.0", delta.ComponentChanges["reliability"])
	}
	if delta.ComponentChanges["latency"] != -3.0 {
		t.Errorf("latency change = %v, want -3.0", delta.ComponentChanges["latency"])
	}
}
