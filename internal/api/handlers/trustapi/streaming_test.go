package trustapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// mockRegistryRepo is a mock implementation for testing
type mockRegistryRepo struct {
	functions []registry.RegistryFunction
	deltas    []registry.TrustScoreDelta
}

func (m *mockRegistryRepo) UpdateSlidingWindowScores(ctx context.Context, config registry.SlidingWindowConfig) ([]registry.TrustScoreDelta, error) {
	return m.deltas, nil
}

func (m *mockRegistryRepo) GetAllFunctionsWithTrustScores() ([]registry.RegistryFunction, error) {
	return m.functions, nil
}

// TestNewTrustScoreStreamer tests streamer initialization
func TestNewTrustScoreStreamer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	repo := &mockRegistryRepo{}
	streamer := NewTrustScoreStreamer(repo, logger)

	if streamer == nil {
		t.Fatal("NewTrustScoreStreamer() returned nil")
	}
	if streamer.repo != repo {
		t.Error("streamer.repo not set correctly")
	}
	if streamer.logger != logger {
		t.Error("streamer.logger not set correctly")
	}
	if streamer.clients == nil {
		t.Error("clients map not initialized")
	}
	if streamer.register == nil {
		t.Error("register channel not initialized")
	}
	if streamer.unregister == nil {
		t.Error("unregister channel not initialized")
	}
	if streamer.broadcast == nil {
		t.Error("broadcast channel not initialized")
	}
}

// TestTrustScoreStreamerConfig tests configuration methods
func TestTrustScoreStreamerConfig(t *testing.T) {
	logger := logrus.New()
	repo := &mockRegistryRepo{}
	streamer := NewTrustScoreStreamer(repo, logger)

	// Test window config
	windowConfig := registry.SlidingWindowConfig{
		WindowDuration:  12 * time.Hour,
		SmoothingFactor: 0.5,
		MinDataPoints:   20,
		UpdateInterval:  1 * time.Minute,
	}
	streamer.SetWindowConfig(windowConfig)

	if streamer.windowConfig.WindowDuration != 12*time.Hour {
		t.Errorf("WindowDuration not set correctly")
	}
	if streamer.windowConfig.SmoothingFactor != 0.5 {
		t.Errorf("SmoothingFactor not set correctly")
	}

	// Test threshold config
	thresholdConfig := registry.TrustScoreThresholdConfig{
		CriticalThreshold:  40.0,
		WarningThreshold:   60.0,
		MinChangeForNotify: 3.0,
		CooldownPeriod:     10 * time.Minute,
	}
	streamer.SetThresholdConfig(thresholdConfig)

	if streamer.thresholdConfig.CriticalThreshold != 40.0 {
		t.Errorf("CriticalThreshold not set correctly")
	}
	if streamer.thresholdConfig.CooldownPeriod != 10*time.Minute {
		t.Errorf("CooldownPeriod not set correctly")
	}
}

// TestTrustScoreStreamerBroadcastDelta tests delta broadcasting
func TestTrustScoreStreamerBroadcastDelta(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	repo := &mockRegistryRepo{}
	streamer := NewTrustScoreStreamer(repo, logger)

	// Run the hub in background
	go streamer.Run()
	defer streamer.Stop()

	time.Sleep(10 * time.Millisecond) // Let hub start

	// Create a test delta
	functionID := uuid.New()
	delta := &registry.TrustScoreDelta{
		FunctionID:    functionID,
		PreviousScore: 70.0,
		CurrentScore:   60.0,
		ScoreChange:    -10.0,
		PreviousTier:   registry.TrustTierTrusted,
		CurrentTier:    registry.TrustTierTrusted,
		TierChanged:    false,
		CalculatedAt:   time.Now(),
		WindowType:     registry.WindowTypeSliding,
	}

	// Broadcast should not panic
	streamer.BroadcastDelta(delta)

	time.Sleep(10 * time.Millisecond) // Let broadcast process

	// Test critical threshold breach
	criticalDelta := &registry.TrustScoreDelta{
		FunctionID:    uuid.New(),
		PreviousScore: 55.0,
		CurrentScore:   45.0,
		ScoreChange:    -10.0,
		PreviousTier:   registry.TrustTierTrusted,
		CurrentTier:    registry.TrustTierUntrusted,
		TierChanged:    true,
		CalculatedAt:   time.Now(),
		WindowType:     registry.WindowTypeSliding,
	}

	streamer.BroadcastDelta(criticalDelta)
	time.Sleep(10 * time.Millisecond)
}

// TestTrustScoreStreamerBroadcastScore tests score broadcasting
func TestTrustScoreStreamerBroadcastScore(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	repo := &mockRegistryRepo{}
	streamer := NewTrustScoreStreamer(repo, logger)

	go streamer.Run()
	defer streamer.Stop()

	time.Sleep(10 * time.Millisecond)

	score := &registry.TrustHistory{
		FunctionID: uuid.New(),
		TrustScore: 85.0,
		TrustTier:  registry.TrustTierVerified,
		CalculatedAt: time.Now(),
	}

	streamer.BroadcastScore(score)
	time.Sleep(10 * time.Millisecond)
}

// TestTrustScoreStreamerHandleSSE tests SSE handler
func TestTrustScoreStreamerHandleSSE(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	repo := &mockRegistryRepo{}
	streamer := NewTrustScoreStreamer(repo, logger)

	go streamer.Run()
	defer streamer.Stop()

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/v1/trust/stream/sse", nil)
	rec := httptest.NewRecorder()

	// Set up context with cancel to simulate client disconnect
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)

	// Handle SSE in background
	done := make(chan struct{})
	go func() {
		streamer.HandleSSE(rec, req)
		close(done)
	}()

	// Wait for connection message
	time.Sleep(50 * time.Millisecond)

	// Cancel context to simulate disconnect
	cancel()

	// Wait for handler to finish
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Handler did not finish in time")
	}

	// Verify response
	resp := rec.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %s", resp.Header.Get("Content-Type"))
	}
	if resp.Header.Get("Cache-Control") != "no-cache" {
		t.Errorf("expected Cache-Control no-cache, got %s", resp.Header.Get("Cache-Control"))
	}
}

// TestTrustScoreStreamerHandleSSEWithFunctionIDs tests SSE with function filtering
func TestTrustScoreStreamerHandleSSEWithFunctionIDs(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	repo := &mockRegistryRepo{}
	streamer := NewTrustScoreStreamer(repo, logger)

	go streamer.Run()
	defer streamer.Stop()

	functionID1 := uuid.New().String()
	functionID2 := uuid.New().String()

	// Create test request with function_id query params
	req := httptest.NewRequest(http.MethodGet, "/v1/trust/stream/sse?function_id="+functionID1+"&function_id="+functionID2, nil)
	rec := httptest.NewRecorder()

	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)

	done := make(chan struct{})
	go func() {
		streamer.HandleSSE(rec, req)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Handler did not finish in time")
	}
}

// TestTrustScoreStreamerShouldSendToClient tests client filtering logic
func TestTrustScoreStreamerShouldSendToClient(t *testing.T) {
	logger := logrus.New()
	streamer := NewTrustScoreStreamer(nil, logger)

	functionID1 := uuid.New()
	functionID2 := uuid.New()

	tests := []struct {
		name       string
		client     *SSEClient
		event      *registry.TrustScoreStreamEvent
		shouldSend bool
	}{
		{
			name: "no filter watches all",
			client: &SSEClient{
				FunctionIDs: []string{},
			},
			event: &registry.TrustScoreStreamEvent{
				FunctionID: functionID1,
			},
			shouldSend: true,
		},
		{
			name: "matching function id",
			client: &SSEClient{
				FunctionIDs: []string{functionID1.String()},
			},
			event: &registry.TrustScoreStreamEvent{
				FunctionID: functionID1,
			},
			shouldSend: true,
		},
		{
			name: "non-matching function id",
			client: &SSEClient{
				FunctionIDs: []string{functionID1.String()},
			},
			event: &registry.TrustScoreStreamEvent{
				FunctionID: functionID2,
			},
			shouldSend: false,
		},
		{
			name: "multiple function ids - one matches",
			client: &SSEClient{
				FunctionIDs: []string{functionID1.String(), functionID2.String()},
			},
			event: &registry.TrustScoreStreamEvent{
				FunctionID: functionID2,
			},
			shouldSend: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := streamer.shouldSendToClient(tt.client, tt.event)
			if result != tt.shouldSend {
				t.Errorf("shouldSendToClient() = %v, want %v", result, tt.shouldSend)
			}
		})
	}
}

// TestTrustScoreStreamerShouldSendToClientWithTierFilter tests tier filtering
func TestTrustScoreStreamerShouldSendToClientWithTierFilter(t *testing.T) {
	logger := logrus.New()
	streamer := NewTrustScoreStreamer(nil, logger)

	functionID := uuid.New()

	tests := []struct {
		name       string
		filterTier registry.TrustTier
		scoreTier  registry.TrustTier
		shouldSend bool
	}{
		{
			name:       "no tier filter",
			filterTier: "",
			scoreTier:  registry.TrustTierTrusted,
			shouldSend: true,
		},
		{
			name:       "tier meets minimum",
			filterTier: registry.TrustTierTrusted,
			scoreTier:  registry.TrustTierVerified,
			shouldSend: true,
		},
		{
			name:       "tier below minimum",
			filterTier: registry.TrustTierVerified,
			scoreTier:  registry.TrustTierTrusted,
			shouldSend: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &SSEClient{
				FunctionIDs: []string{},
				FilterTier:  tt.filterTier,
			}
			event := &registry.TrustScoreStreamEvent{
				FunctionID: functionID,
				Score: &registry.TrustHistory{
					TrustTier: tt.scoreTier,
				},
			}

			result := streamer.shouldSendToClient(client, event)
			if result != tt.shouldSend {
				t.Errorf("shouldSendToClient() = %v, want %v", result, tt.shouldSend)
			}
		})
	}
}

// TestTrustScoreStreamerGenerateClientID tests client ID generation
func TestTrustScoreStreamerGenerateClientID(t *testing.T) {
	logger := logrus.New()
	streamer := NewTrustScoreStreamer(nil, logger)

	client1 := &SSEClient{FunctionIDs: []string{}}
	client2 := &SSEClient{FunctionIDs: []string{uuid.New().String()}}

	id1 := streamer.generateClientID(client1)
	id2 := streamer.generateClientID(client2)

	// IDs should be different
	if id1 == id2 {
		t.Error("generateClientID() returned same ID for different clients")
	}

	// ID should not be empty
	if id1 == "" {
		t.Error("generateClientID() returned empty ID")
	}
}

// TestTrustScoreStreamerTriggerImmediateUpdate tests immediate update trigger
func TestTrustScoreStreamerTriggerImmediateUpdate(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	repo := &mockRegistryRepo{
		deltas: []registry.TrustScoreDelta{
			{
				FunctionID:    uuid.New(),
				PreviousScore: 70.0,
				CurrentScore:   75.0,
				ScoreChange:    5.0,
			},
		},
	}

	streamer := NewTrustScoreStreamer(repo, logger)

	// Should not block
	streamer.TriggerImmediateUpdate()

	// Give goroutine time to start
	time.Sleep(50 * time.Millisecond)
}

// TestTrustScoreStreamerRunStop tests Run and Stop
func TestTrustScoreStreamerRunStop(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	repo := &mockRegistryRepo{}
	streamer := NewTrustScoreStreamer(repo, logger)

	// Run should not block, Stop should clean up
	go streamer.Run()

	// Let it run briefly
	time.Sleep(50 * time.Millisecond)

	// Stop should clean up
	streamer.Stop()

	// Should not panic when stopped again
	streamer.Stop()
}

// TestToJSONStringArray tests helper function
func TestToJSONStringArray(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "empty array",
			input:    []string{},
			expected: "[]",
		},
		{
			name:     "single element",
			input:    []string{"id1"},
			expected: `["id1"]`,
		},
		{
			name:     "multiple elements",
			input:    []string{"id1", "id2", "id3"},
			expected: `["id1","id2","id3"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toJSONStringArray(tt.input)
			if result != tt.expected {
				t.Errorf("toJSONStringArray() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestAbsHelper tests abs helper function
func TestAbsHelper(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{5.0, 5.0},
		{-5.0, 5.0},
		{0.0, 0.0},
		{-0.0, 0.0},
	}

	for _, tt := range tests {
		result := abs(tt.input)
		if result != tt.expected {
			t.Errorf("abs(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

// TestTrustScoreStreamEventSerialization tests JSON serialization
func TestTrustScoreStreamEventSerialization(t *testing.T) {
	functionID := uuid.New()
	event := registry.TrustScoreStreamEvent{
		EventType:  "score_update",
		FunctionID: functionID,
		Score: &registry.TrustHistory{
			TrustScore: 85.0,
			TrustTier:  registry.TrustTierVerified,
		},
		Timestamp:  time.Now(),
		WindowType: registry.WindowTypeSliding,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	var decoded registry.TrustScoreStreamEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if decoded.EventType != event.EventType {
		t.Errorf("EventType mismatch")
	}
	if decoded.FunctionID != event.FunctionID {
		t.Errorf("FunctionID mismatch")
	}
}
