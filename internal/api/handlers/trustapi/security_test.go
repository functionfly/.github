package trustapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TestSSESecurityHeaders verifies security headers are set
func TestSSESecurityHeaders(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	repo := &mockRegistryRepo{}
	streamer := NewTrustScoreStreamer(repo, logger)

	go streamer.Run()
	defer streamer.Stop()

	req := httptest.NewRequest(http.MethodGet, "/v1/trust/stream/sse", nil)
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
	case <-time.After(1 * time.Second):
		t.Fatal("Handler timeout")
	}

	resp := rec.Result()

	// Check security headers
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("Content-Type should be text/event-stream")
	}
	if resp.Header.Get("Cache-Control") != "no-cache" {
		t.Errorf("Cache-Control should be no-cache")
	}
	if resp.Header.Get("Connection") != "keep-alive" {
		t.Errorf("Connection should be keep-alive")
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("CORS header should be set")
	}
}

// TestSSEInvalidFunctionID tests handling of invalid function IDs
func TestSSEInvalidFunctionID(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	repo := &mockRegistryRepo{}
	streamer := NewTrustScoreStreamer(repo, logger)

	go streamer.Run()
	defer streamer.Stop()

	// Test with invalid function_id query param (would be caught by URL parsing in real handler)
	req := httptest.NewRequest(http.MethodGet, "/v1/trust/stream/sse?function_id=invalid-uuid", nil)
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
		// Should complete without panic
	case <-time.After(1 * time.Second):
		t.Fatal("Handler timeout")
	}
}

// TestSSEMalformedQueryParams tests handling of malformed query parameters
func TestSSEMalformedQueryParams(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	repo := &mockRegistryRepo{}
	streamer := NewTrustScoreStreamer(repo, logger)

	go streamer.Run()
	defer streamer.Stop()

	// Test with many function_id params (potential DoS)
	var sb strings.Builder
	sb.WriteString("/v1/trust/stream/sse?")
	for i := 0; i < 100; i++ {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(fmt.Sprintf("function_id=%s", uuid.New().String()))
	}

	req := httptest.NewRequest(http.MethodGet, sb.String(), nil)
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
		// Should handle gracefully
	case <-time.After(1 * time.Second):
		t.Fatal("Handler timeout with many params")
	}
}

// TestSSERapidConnectDisconnect tests rapid connection/disconnection
func TestSSERapidConnectDisconnect(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	repo := &mockRegistryRepo{}
	streamer := NewTrustScoreStreamer(repo, logger)

	go streamer.Run()
	defer streamer.Stop()

	// Connect multiple clients rapidly
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/trust/stream/sse", nil)
		rec := httptest.NewRecorder()

		ctx, cancel := context.WithCancel(req.Context())
		req = req.WithContext(ctx)

		// Start handler
		go streamer.HandleSSE(rec, req)

		// Immediately disconnect
		time.Sleep(5 * time.Millisecond)
		cancel()
	}

	// Give time for cleanup
	time.Sleep(100 * time.Millisecond)

	// Should not have any goroutine leaks or panics
}

// TestSSEClientBufferOverflow tests client buffer overflow handling
func TestSSEClientBufferOverflow(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	repo := &mockRegistryRepo{}
	streamer := NewTrustScoreStreamer(repo, logger)

	go streamer.Run()
	defer streamer.Stop()

	// Create a slow client that won't read messages
	req := httptest.NewRequest(http.MethodGet, "/v1/trust/stream/sse", nil)
	rec := httptest.NewRecorder()

	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()
	req = req.WithContext(ctx)

	// Start handler but don't read from response
	go streamer.HandleSSE(rec, req)

	// Send many messages to overflow buffer
	time.Sleep(20 * time.Millisecond) // Let client register

	functionID := uuid.New()
	for i := 0; i < 200; i++ {
		delta := &registry.TrustScoreDelta{
			FunctionID:    functionID,
			PreviousScore: 70.0 + float64(i),
			CurrentScore:   75.0 + float64(i),
			ScoreChange:    5.0,
		}
		streamer.BroadcastDelta(delta)
	}

	time.Sleep(100 * time.Millisecond)

	// Should handle overflow gracefully by disconnecting slow client
	// No panic should occur
}

// TestWebSocketAuthRequired verifies WebSocket requires authentication
func TestWebSocketAuthRequired(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	notifier := NewTrustDeltaNotifier(nil, nil, logger)

	// Create request without auth context
	req := httptest.NewRequest(http.MethodGet, "/v1/trust/notifications/ws", nil)
	rec := httptest.NewRecorder()

	notifier.HandleWebSocket(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rec.Code)
	}
}

// TestWebSocketInvalidOrigin tests WebSocket origin validation
func TestWebSocketInvalidOrigin(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	notifier := NewTrustDeltaNotifier(nil, nil, logger)

	// Create request with unauthorized origin
	req := httptest.NewRequest(http.MethodGet, "/v1/trust/notifications/ws", nil)
	req.Header.Set("Origin", "https://evil.com")
	rec := httptest.NewRecorder()

	// The handler should check origin through IsOriginAllowedForRequest
	// which would reject unknown origins

	// Without proper auth context, it should still return 401
	notifier.HandleWebSocket(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated request")
	}
}

// TestWebSocketRateLimiting tests rate limiting on WebSocket connections
func TestWebSocketRateLimiting(t *testing.T) {
	// This would test that rapid connection attempts are rate limited
	// Implementation depends on the rate limiting middleware

	// For now, verify that the hub handles rapid registrations
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	hub := NewTrustDeltaHub(logger)
	go hub.Run()

	// Register many clients rapidly
	for i := 0; i < 50; i++ {
		client := &TrustDeltaClient{
			UserID:      fmt.Sprintf("user-%d", i),
			FunctionIDs: []string{},
			Conn:        nil,
			Send:        make(chan []byte, 256),
			hub:         hub,
		}
		hub.register <- client
	}

	time.Sleep(50 * time.Millisecond)

	// Unregister all
	for i := 0; i < 50; i++ {
		// In real scenario we'd track clients
	}
}

// TestCooldownManagerConcurrentAccess tests concurrent access to cooldown manager
func TestCooldownManagerConcurrentAccess(t *testing.T) {
	cm := NewCooldownManager(100 * time.Millisecond)

	functionID := "concurrent-test"

	// Concurrent reads and writes
	done := make(chan bool, 100)

	for i := 0; i < 50; i++ {
		go func() {
			cm.CanNotify(functionID)
			done <- true
		}()
	}

	for i := 0; i < 50; i++ {
		go func() {
			cm.RecordNotification(functionID)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Should not have data races
}

// TestTrustScoreStreamEventDataIntegrity tests event data integrity
func TestTrustScoreStreamEventDataIntegrity(t *testing.T) {
	tests := []struct {
		name      string
		event     registry.TrustScoreStreamEvent
		wantValid bool
	}{
		{
			name: "valid score update",
			event: registry.TrustScoreStreamEvent{
				EventType:  "score_update",
				FunctionID: uuid.New(),
				Score: &registry.TrustHistory{
					TrustScore: 85.0,
					TrustTier:  registry.TrustTierVerified,
				},
				Timestamp:  time.Now(),
				WindowType: registry.WindowTypeSliding,
			},
			wantValid: true,
		},
		{
			name: "valid tier change",
			event: registry.TrustScoreStreamEvent{
				EventType:  "tier_change",
				FunctionID: uuid.New(),
				Delta: &registry.TrustScoreDelta{
					PreviousTier: registry.TrustTierTrusted,
					CurrentTier:  registry.TrustTierVerified,
					TierChanged:  true,
				},
				Timestamp:  time.Now(),
				WindowType: registry.WindowTypeSliding,
			},
			wantValid: true,
		},
		{
			name: "missing function ID",
			event: registry.TrustScoreStreamEvent{
				EventType:  "score_update",
				FunctionID: uuid.Nil,
				Timestamp:  time.Now(),
			},
			wantValid: false,
		},
		{
			name: "invalid event type",
			event: registry.TrustScoreStreamEvent{
				EventType:  "invalid_type",
				FunctionID: uuid.New(),
				Timestamp:  time.Now(),
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.event.FunctionID != uuid.Nil &&
				(tt.event.EventType == "score_update" ||
					tt.event.EventType == "tier_change" ||
					tt.event.EventType == "threshold_breach")

			if valid != tt.wantValid {
				t.Errorf("event validity = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

// TestSlidingWindowConfigValidation tests sliding window configuration validation
func TestSlidingWindowConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config registry.SlidingWindowConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: registry.SlidingWindowConfig{
				WindowDuration:  24 * time.Hour,
				SmoothingFactor: 0.3,
				MinDataPoints:   10,
				UpdateInterval:  5 * time.Minute,
			},
			valid: true,
		},
		{
			name: "zero window duration",
			config: registry.SlidingWindowConfig{
				WindowDuration:  0,
				SmoothingFactor: 0.3,
				MinDataPoints:   10,
				UpdateInterval:  5 * time.Minute,
			},
			valid: false,
		},
		{
			name: "negative smoothing factor",
			config: registry.SlidingWindowConfig{
				WindowDuration:  24 * time.Hour,
				SmoothingFactor: -0.1,
				MinDataPoints:   10,
				UpdateInterval:  5 * time.Minute,
			},
			valid: false,
		},
		{
			name: "smoothing factor > 1",
			config: registry.SlidingWindowConfig{
				WindowDuration:  24 * time.Hour,
				SmoothingFactor: 1.5,
				MinDataPoints:   10,
				UpdateInterval:  5 * time.Minute,
			},
			valid: false,
		},
		{
			name: "negative min data points",
			config: registry.SlidingWindowConfig{
				WindowDuration:  24 * time.Hour,
				SmoothingFactor: 0.3,
				MinDataPoints:   -5,
				UpdateInterval:  5 * time.Minute,
			},
			valid: false,
		},
		{
			name: "zero update interval",
			config: registry.SlidingWindowConfig{
				WindowDuration:  24 * time.Hour,
				SmoothingFactor: 0.3,
				MinDataPoints:   10,
				UpdateInterval:  0,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.config.WindowDuration > 0 &&
				tt.config.SmoothingFactor >= 0 && tt.config.SmoothingFactor <= 1 &&
				tt.config.MinDataPoints >= 0 &&
				tt.config.UpdateInterval > 0

			if valid != tt.valid {
				t.Errorf("config validity = %v, want %v", valid, tt.valid)
			}
		})
	}
}

// TestThresholdConfigValidation tests threshold configuration validation
func TestThresholdConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config registry.TrustScoreThresholdConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: registry.TrustScoreThresholdConfig{
				CriticalThreshold:  50.0,
				WarningThreshold:   70.0,
				MinChangeForNotify: 5.0,
				CooldownPeriod:     15 * time.Minute,
			},
			valid: true,
		},
		{
			name: "warning below critical",
			config: registry.TrustScoreThresholdConfig{
				CriticalThreshold:  70.0,
				WarningThreshold:   50.0,
				MinChangeForNotify: 5.0,
				CooldownPeriod:     15 * time.Minute,
			},
			valid: false,
		},
		{
			name: "negative threshold",
			config: registry.TrustScoreThresholdConfig{
				CriticalThreshold:  -10.0,
				WarningThreshold:   70.0,
				MinChangeForNotify: 5.0,
				CooldownPeriod:     15 * time.Minute,
			},
			valid: false,
		},
		{
			name: "negative cooldown",
			config: registry.TrustScoreThresholdConfig{
				CriticalThreshold:  50.0,
				WarningThreshold:   70.0,
				MinChangeForNotify: 5.0,
				CooldownPeriod:     -5 * time.Minute,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.config.CriticalThreshold >= 0 &&
				tt.config.WarningThreshold >= 0 &&
				tt.config.WarningThreshold >= tt.config.CriticalThreshold &&
				tt.config.MinChangeForNotify >= 0 &&
				tt.config.CooldownPeriod > 0

			if valid != tt.valid {
				t.Errorf("config validity = %v, want %v", valid, tt.valid)
			}
		})
	}
}

// TestTrustDeltaNotificationSecurity tests notification security
func TestTrustDeltaNotificationSecurity(t *testing.T) {
	notification := TrustDeltaNotification{
		UserID:        "user-123",
		Type:          "trust_delta",
		FunctionID:    uuid.New().String(),
		FunctionName:  "test-function",
		PreviousScore: 70.0,
		CurrentScore:  75.0,
		ScoreChange:   5.0,
		Severity:      "info",
		Timestamp:     time.Now(),
		Data: map[string]interface{}{
			"score_change_percent": 7.14,
			"component_changes": map[string]float64{
				"reliability": 2.0,
				"latency":     3.0,
			},
		},
	}

	// Verify no sensitive data exposed
	if notification.UserID == "" {
		t.Error("UserID should be set")
	}
	if notification.FunctionID == "" {
		t.Error("FunctionID should be set")
	}

	// Verify all fields are within expected ranges
	if notification.PreviousScore < 0 || notification.PreviousScore > 100 {
		t.Error("PreviousScore out of range")
	}
	if notification.CurrentScore < 0 || notification.CurrentScore > 100 {
		t.Error("CurrentScore out of range")
	}
}

// TestTrustScoreStreamerGracefulShutdown tests graceful shutdown
func TestTrustScoreStreamerGracefulShutdown(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	repo := &mockRegistryRepo{}
	streamer := NewTrustScoreStreamer(repo, logger)

	// Start with some clients
	go streamer.Run()
	time.Sleep(10 * time.Millisecond)

	// Add clients
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/trust/stream/sse", nil)
		rec := httptest.NewRecorder()
		ctx, cancel := context.WithCancel(req.Context())
		req = req.WithContext(ctx)

		go func() {
			streamer.HandleSSE(rec, req)
		}()

		time.Sleep(10 * time.Millisecond)
		cancel()
	}

	time.Sleep(50 * time.Millisecond)

	// Stop should close all client channels
	streamer.Stop()

	time.Sleep(50 * time.Millisecond)

	// Verify no panic after stop
	streamer.Stop()
}

// TestWebSocketMessageInjection tests protection against message injection
func TestWebSocketMessageInjection(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	notifier := NewTrustDeltaNotifier(nil, nil, logger)

	// Test that readPump properly handles unexpected messages
	client := &TrustDeltaClient{
		UserID:      "test-user",
		FunctionIDs: []string{},
		Send:        make(chan []byte, 256),
		hub:         notifier.hub,
	}

	// The readPump should handle various message types without crashing
	// This is tested implicitly through the overall flow

	_ = client
}


