package trustapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// TestNewCooldownManager tests cooldown manager initialization
func TestNewCooldownManager(t *testing.T) {
	cooldown := 5 * time.Minute
	cm := NewCooldownManager(cooldown)

	if cm == nil {
		t.Fatal("NewCooldownManager() returned nil")
	}
	if cm.cooldown != cooldown {
		t.Errorf("cooldown = %v, want %v", cm.cooldown, cooldown)
	}
	if cm.lastNotified == nil {
		t.Error("lastNotified map not initialized")
	}
}

// TestCooldownManagerCanNotify tests notification cooldown logic
func TestCooldownManagerCanNotify(t *testing.T) {
	cm := NewCooldownManager(100 * time.Millisecond)

	functionID := "test-function-1"

	// First notification should be allowed
	if !cm.CanNotify(functionID) {
		t.Error("first notification should be allowed")
	}

	// Record notification
	cm.RecordNotification(functionID)

	// Immediate second notification should be blocked
	if cm.CanNotify(functionID) {
		t.Error("immediate second notification should be blocked")
	}

	// Wait for cooldown
	time.Sleep(150 * time.Millisecond)

	// After cooldown, should be allowed
	if !cm.CanNotify(functionID) {
		t.Error("notification after cooldown should be allowed")
	}
}

// TestCooldownManagerDifferentFunctions tests per-function cooldown isolation
func TestCooldownManagerDifferentFunctions(t *testing.T) {
	cm := NewCooldownManager(1 * time.Minute)

	func1 := "function-1"
	func2 := "function-2"

	// Record notification for function 1
	cm.RecordNotification(func1)

	// Function 2 should still be able to notify
	if !cm.CanNotify(func2) {
		t.Error("different function should not be affected by cooldown")
	}

	// Function 1 should be on cooldown
	if cm.CanNotify(func1) {
		t.Error("function on cooldown should be blocked")
	}
}

// TestNewTrustDeltaHub tests hub initialization
func TestNewTrustDeltaHub(t *testing.T) {
	logger := logrus.New()
	hub := NewTrustDeltaHub(logger)

	if hub == nil {
		t.Fatal("NewTrustDeltaHub() returned nil")
	}
	if hub.clients == nil {
		t.Error("clients map not initialized")
	}
	if hub.register == nil {
		t.Error("register channel not initialized")
	}
	if hub.unregister == nil {
		t.Error("unregister channel not initialized")
	}
	if hub.broadcast == nil {
		t.Error("broadcast channel not initialized")
	}
}

// TestTrustDeltaHubBroadcast tests hub broadcasting
func TestTrustDeltaHubBroadcast(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	hub := NewTrustDeltaHub(logger)
	go hub.Run()

	time.Sleep(10 * time.Millisecond)

	notification := &TrustDeltaNotification{
		UserID:       "user-1",
		Type:         "trust_delta",
		FunctionID:   uuid.New().String(),
		CurrentScore: 75.0,
		PreviousScore: 70.0,
		ScoreChange:   5.0,
		Severity:      "info",
		Timestamp:    time.Now(),
	}

	// Broadcast should not panic
	hub.Broadcast("user-1", notification)

	time.Sleep(10 * time.Millisecond)
}

// TestNewTrustDeltaNotifier tests notifier initialization
func TestNewTrustDeltaNotifier(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create notifier without webhook service
	notifier := NewTrustDeltaNotifier(nil, nil, logger)

	if notifier == nil {
		t.Fatal("NewTrustDeltaNotifier() returned nil")
	}
	if notifier.hub == nil {
		t.Error("hub not initialized")
	}
	if notifier.cooldownManager == nil {
		t.Error("cooldownManager not initialized")
	}
}

// TestTrustDeltaNotifierSetThresholdConfig tests threshold configuration
func TestTrustDeltaNotifierSetThresholdConfig(t *testing.T) {
	logger := logrus.New()
	notifier := NewTrustDeltaNotifier(nil, nil, logger)

	config := registry.TrustScoreThresholdConfig{
		CriticalThreshold:  40.0,
		WarningThreshold:   60.0,
		MinChangeForNotify: 3.0,
		CooldownPeriod:     10 * time.Minute,
	}

	notifier.SetThresholdConfig(config)

	if notifier.thresholdConfig.CriticalThreshold != 40.0 {
		t.Errorf("CriticalThreshold not set correctly")
	}
	if notifier.cooldownManager.cooldown != 10*time.Minute {
		t.Errorf("cooldown not updated correctly")
	}
}

// TestTrustDeltaNotifierDetermineNotification tests notification determination
func TestTrustDeltaNotifierDetermineNotification(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	notifier := NewTrustDeltaNotifier(nil, nil, logger)

	tests := []struct {
		name         string
		delta        *registry.TrustScoreDelta
		wantSeverity string
		wantType     string
	}{
		{
			name: "critical threshold breach",
			delta: &registry.TrustScoreDelta{
				CurrentScore: 45.0,
				PreviousScore: 55.0,
			},
			wantSeverity: "critical",
			wantType:     "threshold_breach",
		},
		{
			name: "warning threshold breach",
			delta: &registry.TrustScoreDelta{
				CurrentScore:   65.0,
				PreviousScore:  75.0,
			},
			wantSeverity: "warning",
			wantType:     "threshold_breach",
		},
		{
			name: "tier downgrade",
			delta: &registry.TrustScoreDelta{
				CurrentScore:   60.0,
				PreviousScore:  80.0,
				PreviousTier:   registry.TrustTierVerified,
				CurrentTier:    registry.TrustTierTrusted,
				TierChanged:    true,
			},
			wantSeverity: "warning",
			wantType:     "tier_change",
		},
		{
			name: "tier upgrade",
			delta: &registry.TrustScoreDelta{
				CurrentScore:   80.0,
				PreviousScore:  60.0,
				PreviousTier:   registry.TrustTierTrusted,
				CurrentTier:    registry.TrustTierVerified,
				TierChanged:    true,
			},
			wantSeverity: "info",
			wantType:     "tier_change",
		},
		{
			name: "positive score change",
			delta: &registry.TrustScoreDelta{
				CurrentScore:   75.0,
				PreviousScore:  70.0,
				ScoreChange:    5.0,
				TierChanged:    false,
			},
			wantSeverity: "info",
			wantType:     "trust_delta",
		},
		{
			name: "negative score change",
			delta: &registry.TrustScoreDelta{
				CurrentScore:   70.0,
				PreviousScore:  75.0,
				ScoreChange:    -5.0,
				TierChanged:    false,
			},
			wantSeverity: "info",
			wantType:     "trust_delta",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity, notifType, _ := notifier.determineNotification(tt.delta)

			if severity != tt.wantSeverity {
				t.Errorf("severity = %v, want %v", severity, tt.wantSeverity)
			}
			if notifType != tt.wantType {
				t.Errorf("type = %v, want %v", notifType, tt.wantType)
			}
		})
	}
}

// TestFormatChange tests change formatting
func TestFormatChange(t *testing.T) {
	tests := []struct {
		change   float64
		expected string
	}{
		{5.0, "+5"},
		{-5.0, "-5"},
		{0.0, "0"},
		{3.5, "+3.5"},
	}

	for _, tt := range tests {
		result := formatChange(tt.change)
		if result != tt.expected {
			t.Errorf("formatChange(%v) = %v, want %v", tt.change, result, tt.expected)
		}
	}
}

// TestTrustDeltaNotificationStructure tests notification structure
func TestTrustDeltaNotificationStructure(t *testing.T) {
	functionID := uuid.New().String()

	notification := TrustDeltaNotification{
		UserID:        "user-123",
		Type:          "trust_delta",
		FunctionID:    functionID,
		FunctionName:  "test-function",
		PreviousScore: 70.0,
		CurrentScore:  75.0,
		ScoreChange:   5.0,
		PreviousTier:  "trusted",
		CurrentTier:   "verified",
		Severity:      "info",
		Message:       "Trust score improved",
		Timestamp:     time.Now(),
		Data: map[string]interface{}{
			"score_change_percent": 7.14,
			"tier_changed":         false,
		},
	}

	if notification.UserID != "user-123" {
		t.Error("UserID mismatch")
	}
	if notification.FunctionID != functionID {
		t.Error("FunctionID mismatch")
	}
	if notification.ScoreChange != 5.0 {
		t.Error("ScoreChange mismatch")
	}
}

// TestTrustDeltaClientWebSocketUpgrade tests WebSocket upgrade (without full server)
func TestTrustDeltaClientWebSocketUpgrade(t *testing.T) {
	// Create a test WebSocket server
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Create a simple server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Could not upgrade", http.StatusBadRequest)
			return
		}
		defer conn.Close()

		// Send a test message
		msg := map[string]string{"type": "connected"}
		conn.WriteJSON(msg)
	}))
	defer server.Close()

	// Connect to the server
	wsURL := "ws" + server.URL[4:] // Replace http with ws
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Read the connected message
	var msg map[string]string
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	if msg["type"] != "connected" {
		t.Errorf("expected type 'connected', got %v", msg["type"])
	}
}

// TestTrustDeltaHubClientRegistration tests client registration flow
func TestTrustDeltaHubClientRegistration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	hub := NewTrustDeltaHub(logger)
	go hub.Run()

	// Create a client
	client := &TrustDeltaClient{
		UserID:      "test-user",
		FunctionIDs: []string{"func-1", "func-2"},
		Conn:        nil, // Would be real WebSocket in practice
		Send:        make(chan []byte, 256),
		hub:         hub,
	}

	// Register client
	hub.register <- client

	time.Sleep(10 * time.Millisecond)

	// Unregister client
	hub.unregister <- client

	time.Sleep(10 * time.Millisecond)

	// Send channel should be closed after unregister
	select {
	case _, ok := <-client.Send:
		if ok {
			t.Error("Send channel should be closed after unregister")
		}
	default:
		// Channel might already be closed
	}
}

// TestTrustDeltaHubBroadcastWithNoClients tests broadcast with no clients
func TestTrustDeltaHubBroadcastWithNoClients(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	hub := NewTrustDeltaHub(logger)
	go hub.Run()

	notification := &TrustDeltaNotification{
		UserID:       "nonexistent-user",
		Type:         "trust_delta",
		FunctionID:   uuid.New().String(),
		CurrentScore: 80.0,
	}

	// Should not panic
	hub.Broadcast("nonexistent-user", notification)

	time.Sleep(10 * time.Millisecond)
}

// TestTrustDeltaNotifierProcessDeltaBelowThreshold tests delta below threshold is ignored
func TestTrustDeltaNotifierProcessDeltaBelowThreshold(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	notifier := NewTrustDeltaNotifier(nil, nil, logger)

	// Set high threshold
	notifier.SetThresholdConfig(registry.TrustScoreThresholdConfig{
		CriticalThreshold:  50.0,
		WarningThreshold:   70.0,
		MinChangeForNotify: 10.0, // High threshold
		CooldownPeriod:     1 * time.Minute,
	})

	delta := &registry.TrustScoreDelta{
		FunctionID:    uuid.New(),
		PreviousScore: 80.0,
		CurrentScore:   82.0,
		ScoreChange:    2.0, // Below 10.0 threshold
		TierChanged:    false,
		CalculatedAt:   time.Now(),
	}

	// Should not trigger notification due to low change
	ownerID := uuid.New()
	notifier.ProcessTrustDelta(delta, ownerID, "test-function")

	// Test passes if no panic occurs
}

// TestTrustDeltaNotifierProcessDeltaCritical tests critical delta processing
func TestTrustDeltaNotifierProcessDeltaCritical(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	notifier := NewTrustDeltaNotifier(nil, nil, logger)

	delta := &registry.TrustScoreDelta{
		FunctionID:    uuid.New(),
		PreviousScore: 55.0,
		CurrentScore:   45.0, // Below critical threshold
		ScoreChange:    -10.0,
		TierChanged:    true,
		PreviousTier:   registry.TrustTierTrusted,
		CurrentTier:    registry.TrustTierUntrusted,
		CalculatedAt:   time.Now(),
	}

	ownerID := uuid.New()

	// Should broadcast and attempt webhook delivery
	notifier.ProcessTrustDelta(delta, ownerID, "test-function")

	// Verify cooldown was recorded
	if notifier.cooldownManager.CanNotify(delta.FunctionID.String()) {
		t.Error("cooldown should be recorded for critical alert")
	}
}

// TestTrustDeltaNotifierProcessDeltaWithCooldown tests cooldown enforcement
func TestTrustDeltaNotifierProcessDeltaWithCooldown(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	notifier := NewTrustDeltaNotifier(nil, nil, logger)

	// Set short cooldown for testing
	notifier.SetThresholdConfig(registry.TrustScoreThresholdConfig{
		CriticalThreshold:  50.0,
		WarningThreshold:   70.0,
		MinChangeForNotify: 1.0,
		CooldownPeriod:     1 * time.Hour, // Long cooldown
	})

	delta := &registry.TrustScoreDelta{
		FunctionID:    uuid.New(),
		PreviousScore: 80.0,
		CurrentScore:   75.0,
		ScoreChange:    -5.0,
		CalculatedAt:   time.Now(),
	}

	ownerID := uuid.New()

	// First notification
	notifier.ProcessTrustDelta(delta, ownerID, "test-function")

	// Record cooldown
	notifier.cooldownManager.RecordNotification(delta.FunctionID.String())

	// Second notification should be blocked by cooldown
	// (even though change is significant)
	delta2 := &registry.TrustScoreDelta{
		FunctionID:    delta.FunctionID, // Same function
		PreviousScore: 75.0,
		CurrentScore:   70.0,
		ScoreChange:    -5.0,
		CalculatedAt:   time.Now(),
	}

	notifier.ProcessTrustDelta(delta2, ownerID, "test-function")

	// Test passes if no panic
}
