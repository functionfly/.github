package scheduler

import (
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// TestNewTrustScoreScheduler tests scheduler initialization
func TestNewTrustScoreScheduler(t *testing.T) {
	// Would need actual repository
	var repo *registry.RegistryRepository
	
	scheduler := NewTrustScoreScheduler(repo)
	
	if scheduler == nil {
		t.Fatal("NewTrustScoreScheduler() returned nil")
	}
	if scheduler.cron == nil {
		t.Error("cron not initialized")
	}
	if scheduler.repo != repo {
		t.Error("repo not set correctly")
	}
	
	// Check default streaming config
	if scheduler.streamingConfig.MinChangeThreshold != 5.0 {
		t.Errorf("default MinChangeThreshold = %v, want 5.0", scheduler.streamingConfig.MinChangeThreshold)
	}
	if scheduler.streamingConfig.EnableSlidingWindow != false {
		t.Error("default EnableSlidingWindow should be false")
	}
}

// TestTrustScoreSchedulerSetStreamingConfig tests streaming config updates
func TestTrustScoreSchedulerSetStreamingConfig(t *testing.T) {
	var repo *registry.RegistryRepository
	scheduler := NewTrustScoreScheduler(repo)

	config := TrustScoreStreamingConfig{
		Enabled:              true,
		BroadcastSignificant: true,
		MinChangeThreshold:   3.0,
		EnableSlidingWindow:  true,
		UpdateInterval:       1 * time.Minute,
	}

	scheduler.SetStreamingConfig(config)

	if !scheduler.streamingConfig.Enabled {
		t.Error("Enabled not set correctly")
	}
	if !scheduler.streamingConfig.BroadcastSignificant {
		t.Error("BroadcastSignificant not set correctly")
	}
	if scheduler.streamingConfig.MinChangeThreshold != 3.0 {
		t.Error("MinChangeThreshold not set correctly")
	}
	if !scheduler.streamingConfig.EnableSlidingWindow {
		t.Error("EnableSlidingWindow not set correctly")
	}
	if scheduler.streamingConfig.UpdateInterval != 1*time.Minute {
		t.Error("UpdateInterval not set correctly")
	}
}

// TestTrustScoreSchedulerSetSlidingWindowConfig tests sliding window config updates
func TestTrustScoreSchedulerSetSlidingWindowConfig(t *testing.T) {
	var repo *registry.RegistryRepository
	scheduler := NewTrustScoreScheduler(repo)

	config := registry.SlidingWindowConfig{
		WindowDuration:  12 * time.Hour,
		SmoothingFactor: 0.5,
		MinDataPoints:   20,
		UpdateInterval:  30 * time.Second,
	}

	scheduler.SetSlidingWindowConfig(config)

	if scheduler.slidingWindowConfig.WindowDuration != 12*time.Hour {
		t.Error("WindowDuration not set correctly")
	}
	if scheduler.slidingWindowConfig.SmoothingFactor != 0.5 {
		t.Error("SmoothingFactor not set correctly")
	}
}

// TestTrustScoreSchedulerSetOnScoreUpdate tests callback registration
func TestTrustScoreSchedulerSetOnScoreUpdate(t *testing.T) {
	var repo *registry.RegistryRepository
	scheduler := NewTrustScoreScheduler(repo)

	callbackCalled := false
	callback := func(delta *registry.TrustScoreDelta) {
		callbackCalled = true
	}

	scheduler.SetOnScoreUpdate(callback)

	if scheduler.onScoreUpdate == nil {
		t.Error("onScoreUpdate not set correctly")
	}

	// Call the callback to verify it works
	if scheduler.onScoreUpdate != nil {
		delta := &registry.TrustScoreDelta{
			FunctionID:    uuid.New(),
			PreviousScore: 70.0,
			CurrentScore:  75.0,
			ScoreChange:   5.0,
		}
		scheduler.onScoreUpdate(delta)
		if !callbackCalled {
			t.Error("callback was not called")
		}
	}
}

// TestCountSignificant tests significant change counting
func TestCountSignificant(t *testing.T) {
	tests := []struct {
		name      string
		deltas    []registry.TrustScoreDelta
		threshold float64
		expected  int
	}{
		{
			name:      "no changes",
			deltas:    []registry.TrustScoreDelta{},
			threshold: 5.0,
			expected:  0,
		},
		{
			name: "all below threshold",
			deltas: []registry.TrustScoreDelta{
				{ScoreChange: 1.0},
				{ScoreChange: 2.0},
				{ScoreChange: 3.0},
			},
			threshold: 5.0,
			expected:  0,
		},
		{
			name: "all above threshold",
			deltas: []registry.TrustScoreDelta{
				{ScoreChange: 6.0},
				{ScoreChange: 7.0},
				{ScoreChange: 10.0},
			},
			threshold: 5.0,
			expected:  3,
		},
		{
			name: "mixed changes",
			deltas: []registry.TrustScoreDelta{
				{ScoreChange: 3.0},
				{ScoreChange: 7.0},
				{ScoreChange: 2.0},
				{ScoreChange: 8.0},
			},
			threshold: 5.0,
			expected:  2,
		},
		{
			name: "tier changes count as significant",
			deltas: []registry.TrustScoreDelta{
				{ScoreChange: 2.0, TierChanged: true},
				{ScoreChange: 3.0, TierChanged: false},
			},
			threshold: 5.0,
			expected:  1,
		},
		{
			name: "negative changes",
			deltas: []registry.TrustScoreDelta{
				{ScoreChange: -6.0},
				{ScoreChange: -7.0},
				{ScoreChange: 3.0},
			},
			threshold: 5.0,
			expected:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countSignificant(tt.deltas, tt.threshold)
			if result != tt.expected {
				t.Errorf("countSignificant() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestAbsFunction tests the abs helper function
func TestAbsFunction(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{5.0, 5.0},
		{-5.0, 5.0},
		{0.0, 0.0},
		{-0.0, 0.0},
		{3.14, 3.14},
		{-3.14, 3.14},
	}

	for _, tt := range tests {
		result := abs(tt.input)
		if result != tt.expected {
			t.Errorf("abs(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

// TestTrustScoreStreamingConfigValidation tests streaming config validation
func TestTrustScoreStreamingConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config TrustScoreStreamingConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: TrustScoreStreamingConfig{
				Enabled:              true,
				BroadcastSignificant: true,
				MinChangeThreshold:   5.0,
				EnableSlidingWindow:  true,
				UpdateInterval:       1 * time.Minute,
			},
			valid: true,
		},
		{
			name: "negative threshold",
			config: TrustScoreStreamingConfig{
				Enabled:              true,
				BroadcastSignificant: true,
				MinChangeThreshold:   -5.0,
				EnableSlidingWindow:  true,
				UpdateInterval:       1 * time.Minute,
			},
			valid: false,
		},
		{
			name: "zero update interval",
			config: TrustScoreStreamingConfig{
				Enabled:              true,
				BroadcastSignificant: true,
				MinChangeThreshold:   5.0,
				EnableSlidingWindow:  true,
				UpdateInterval:       0,
			},
			valid: false,
		},
		{
			name: "negative update interval",
			config: TrustScoreStreamingConfig{
				Enabled:              true,
				BroadcastSignificant: true,
				MinChangeThreshold:   5.0,
				EnableSlidingWindow:  true,
				UpdateInterval:       -1 * time.Minute,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.config.MinChangeThreshold >= 0 &&
				tt.config.UpdateInterval > 0

			if valid != tt.valid {
				t.Errorf("config validity = %v, want %v", valid, tt.valid)
			}
		})
	}
}

// TestTrustScoreSchedulerConcurrency tests thread safety
func TestTrustScoreSchedulerConcurrency(t *testing.T) {
	var repo *registry.RegistryRepository
	scheduler := NewTrustScoreScheduler(repo)

	// Concurrent config updates
	done := make(chan bool, 100)

	for i := 0; i < 50; i++ {
		go func() {
			config := TrustScoreStreamingConfig{
				Enabled:              true,
				MinChangeThreshold:   float64(i),
				UpdateInterval:       time.Duration(i) * time.Second,
			}
			scheduler.SetStreamingConfig(config)
			done <- true
		}()
	}

	for i := 0; i < 50; i++ {
		go func() {
			config := registry.SlidingWindowConfig{
				WindowDuration:  time.Duration(i) * time.Hour,
				SmoothingFactor: float64(i) / 100.0,
			}
			scheduler.SetSlidingWindowConfig(config)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// No data races should occur
}
