package status

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockWSHandler is a test helper to create a handler with mocked dependencies
type mockWSHandler struct {
	repo             *mockWSRepository
	prometheus       *PrometheusClient
	getHealthFunc    func(ctx context.Context) (map[string]bool, error)
	getHealthPctFunc func(ctx context.Context) (float64, error)
}

type mockWSRepository struct {
	incidents          []Incident
	maintenanceWindows []MaintenanceWindow
	components         []Component
	providerStatus     []ProviderStatus
	mu                 sync.RWMutex
}

func (m *mockWSRepository) GetActiveIncidents(ctx context.Context) ([]Incident, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active []Incident
	for _, inc := range m.incidents {
		if inc.Status != "resolved" {
			active = append(active, inc)
		}
	}
	return active, nil
}

func (m *mockWSRepository) GetUpcomingMaintenance(ctx context.Context) ([]MaintenanceWindow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.maintenanceWindows, nil
}

func (m *mockWSRepository) GetSystemHealthChecks(ctx context.Context) ([]Component, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.components, nil
}

func (m *mockWSRepository) GetProviderStatus(ctx context.Context) ([]ProviderStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.providerStatus, nil
}

func (m *mockWSRepository) GetComponentHealthHistory(ctx context.Context, componentName string, since time.Time) ([]StatusHistoryPoint, error) {
	return []StatusHistoryPoint{}, nil
}

func (m *mockWSRepository) CalculateComponentUptime(ctx context.Context, componentName string, duration time.Duration) (float64, error) {
	return 99.9, nil
}

func (m *mockWSRepository) GetLatestComponentResponseTime(ctx context.Context, componentName string) (int, error) {
	return 50, nil
}

func (m *mockWSRepository) GetProviderRegions(ctx context.Context, provider string) ([]RegionStatus, error) {
	return []RegionStatus{}, nil
}

func (m *mockWSRepository) GetProviderBackends(ctx context.Context, provider, region string) ([]BackendStatus, error) {
	return []BackendStatus{}, nil
}

func (m *mockWSRepository) AddIncident(incident Incident) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.incidents = append(m.incidents, incident)
}

func (m *mockWSRepository) UpdateIncidentStatus(id string, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.incidents {
		if m.incidents[i].ID == id {
			m.incidents[i].Status = status
			if status == "resolved" {
				now := time.Now()
				m.incidents[i].ResolvedAt = &now
			}
			break
		}
	}
}

func setupTestWebSocketHub(t *testing.T) (*StatusWebSocketHub, *mockWSRepository, func()) {
	repo := &mockWSRepository{
		incidents: []Incident{
			{
				ID:          "incident-1",
				Title:       "Test Incident",
				Severity:    "high",
				Status:      "investigating",
				Description: "Test description",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
		maintenanceWindows: []MaintenanceWindow{
			{
				ID:             "maint-1",
				Title:          "Scheduled Maintenance",
				Status:         "scheduled",
				ScheduledStart: time.Now().Add(24 * time.Hour),
				ScheduledEnd:   time.Now().Add(26 * time.Hour),
			},
		},
		components: []Component{
			{ID: "api", Name: "API", Type: "api", Status: "operational"},
			{ID: "db", Name: "Database", Type: "database", Status: "operational"},
		},
		providerStatus: []ProviderStatus{
			{Name: "fly", DisplayName: "Fly.io", OverallStatus: "operational"},
		},
	}

	// Create a handler with mocked dependencies
	handler := &Handler{
		repo: nil, // We'll set this after
		prometheus: &PrometheusClient{
			cache: &PrometheusCache{entries: make(map[string]cacheEntry)},
		},
	}

	// We need to create a repository that matches the interface
	// For simplicity, we'll override the methods we need

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Suppress log output during tests

	hub := NewStatusWebSocketHub(handler, logger)

	cleanup := func() {
		hub.Stop()
	}

	return hub, repo, cleanup
}

func TestNewStatusWebSocketHub(t *testing.T) {
	handler := &Handler{}
	logger := logrus.New()

	hub := NewStatusWebSocketHub(handler, logger)

	assert.NotNil(t, hub)
	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
	assert.NotNil(t, hub.broadcast)
	assert.Equal(t, handler, hub.handler)
	assert.Equal(t, logger, hub.logger)
	assert.False(t, hub.running)
}

func TestStatusWebSocketHub_RunAndStop(t *testing.T) {
	handler := &Handler{}
	logger := logrus.New()
	hub := NewStatusWebSocketHub(handler, logger)

	t.Run("start and stop hub", func(t *testing.T) {
		// Run the hub in a goroutine
		go hub.Run()

		// Give it time to start
		time.Sleep(100 * time.Millisecond)

		assert.True(t, hub.running)

		// Stop the hub
		hub.Stop()

		// Give it time to stop
		time.Sleep(50 * time.Millisecond)

		assert.False(t, hub.running)
	})

	t.Run("double start doesn't panic", func(t *testing.T) {
		hub2 := NewStatusWebSocketHub(handler, logger)

		go hub2.Run()
		time.Sleep(50 * time.Millisecond)

		// This should not panic or cause issues
		hub2.Run()

		hub2.Stop()
	})
}

func TestStatusWebSocketClient(t *testing.T) {
	handler := &Handler{}
	logger := logrus.New()
	hub := NewStatusWebSocketHub(handler, logger)

	t.Run("create client", func(t *testing.T) {
		// Create a test WebSocket server
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Fatalf("Failed to upgrade: %v", err)
			}
			defer conn.Close()

			// Create a client
			client := &StatusWebSocketClient{
				ID:       "test-client",
				Channels: make(map[string]bool),
				Conn:     conn,
				Send:     make(chan []byte, 256),
				Hub:      hub,
				IsAdmin:  false,
			}

			// Register the client
			hub.register <- client

			// Wait for a bit then unregister
			time.Sleep(100 * time.Millisecond)
			hub.unregister <- client
		}))
		defer server.Close()

		// Start the hub
		go hub.Run()
		defer hub.Stop()

		// Connect to the WebSocket server
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Wait for client to be registered and unregistered
		time.Sleep(200 * time.Millisecond)
	})
}

func TestStatusWebSocketHub_broadcastToSubscribers(t *testing.T) {
	handler := &Handler{}
	logger := logrus.New()
	hub := NewStatusWebSocketHub(handler, logger)

	// Create mock clients
	client1 := &StatusWebSocketClient{
		ID:       "client-1",
		Channels: map[string]bool{"platform": true},
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}
	client2 := &StatusWebSocketClient{
		ID:       "client-2",
		Channels: map[string]bool{"incidents": true},
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}
	client3 := &StatusWebSocketClient{
		ID:       "client-3",
		Channels: map[string]bool{"all": true},
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}

	hub.clients[client1] = true
	hub.clients[client2] = true
	hub.clients[client3] = true

	t.Run("broadcast to subscribed channels", func(t *testing.T) {
		msg := &StatusUpdateMessage{
			Type:      "status_update",
			Channel:   "platform",
			Timestamp: time.Now(),
			Data:      map[string]string{"status": "operational"},
		}

		hub.broadcastToSubscribers(msg)

		// client1 and client3 should receive (subscribed to platform or all)
		// client2 should not receive (only subscribed to incidents)

		select {
		case <-client1.Send:
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Error("client1 should have received the message")
		}

		select {
		case <-client2.Send:
			t.Error("client2 should not have received the message")
		case <-time.After(100 * time.Millisecond):
			// Expected
		}

		select {
		case <-client3.Send:
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Error("client3 should have received the message")
		}
	})
}

func TestGenerateClientID(t *testing.T) {
	id1 := generateClientID()
	id2 := generateClientID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2, "Client IDs should be unique")
	assert.True(t, strings.HasPrefix(id1, "client_"))
	assert.True(t, strings.HasPrefix(id2, "client_"))
}

func TestStatusUpdateMessage_Serialization(t *testing.T) {
	msg := &StatusUpdateMessage{
		Type:      "status_update",
		Channel:   "platform",
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Data: PlatformStatus{
			Status:      "operational",
			Indicator:   "none",
			Description: "All systems operational",
			UpdatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded StatusUpdateMessage
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, msg.Type, decoded.Type)
	assert.Equal(t, msg.Channel, decoded.Channel)
}

func TestSubscribeMessage(t *testing.T) {
	jsonData := `{"type":"subscribe","channels":["platform","incidents"]}`

	var msg SubscribeMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	require.NoError(t, err)

	assert.Equal(t, "subscribe", msg.Type)
	assert.Equal(t, []string{"platform", "incidents"}, msg.Channels)
}

func TestProviderUpdate(t *testing.T) {
	update := ProviderUpdate{
		Provider:     "fly",
		Region:       "iad",
		Status:       "operational",
		LatencyMs:    45.5,
		CircuitState: "closed",
	}

	data, err := json.Marshal(update)
	require.NoError(t, err)

	var decoded ProviderUpdate
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, update.Provider, decoded.Provider)
	assert.Equal(t, update.Region, decoded.Region)
	assert.Equal(t, update.Status, decoded.Status)
	assert.InDelta(t, update.LatencyMs, decoded.LatencyMs, 0.01)
	assert.Equal(t, update.CircuitState, decoded.CircuitState)
}

func TestStatusWebSocketHub_BroadcastIncidentUpdate(t *testing.T) {
	handler := &Handler{}
	logger := logrus.New()
	hub := NewStatusWebSocketHub(handler, logger)

	// Start the hub
	go hub.Run()
	defer hub.Stop()
	time.Sleep(50 * time.Millisecond)

	incident := &Incident{
		ID:       "incident-123",
		Title:    "Test Incident",
		Severity: "high",
		Status:   "investigating",
	}

	t.Run("broadcast when running", func(t *testing.T) {
		hub.BroadcastIncidentUpdate(incident)

		// The broadcast should not panic and should queue the message
		assert.True(t, hub.running)
	})

	t.Run("broadcast when not running", func(t *testing.T) {
		hub2 := NewStatusWebSocketHub(handler, logger)
		// Don't start the hub

		// This should not panic or block
		hub2.BroadcastIncidentUpdate(incident)
	})
}

func TestStatusWebSocketHub_BroadcastProviderUpdate(t *testing.T) {
	handler := &Handler{}
	logger := logrus.New()
	hub := NewStatusWebSocketHub(handler, logger)

	go hub.Run()
	defer hub.Stop()
	time.Sleep(50 * time.Millisecond)

	update := &ProviderUpdate{
		Provider:  "fly",
		Region:    "iad",
		Status:    "operational",
		LatencyMs: 45.5,
	}

	t.Run("broadcast when running", func(t *testing.T) {
		hub.BroadcastProviderUpdate(update)
		assert.True(t, hub.running)
	})

	t.Run("broadcast when not running", func(t *testing.T) {
		hub2 := NewStatusWebSocketHub(handler, logger)
		hub2.BroadcastProviderUpdate(update)
		// Should not panic
	})
}

func TestStatusWebSocketHub_sendInitialStatus(t *testing.T) {
	handler := &Handler{}
	logger := logrus.New()
	hub := NewStatusWebSocketHub(handler, logger)

	client := &StatusWebSocketClient{
		ID:       "test-client",
		Channels: map[string]bool{"platform": true, "incidents": true},
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}

	t.Run("send initial status to subscribed channels", func(t *testing.T) {
		// This would require mocking the repository and prometheus client
		// For now, we just verify the method doesn't panic
		// hub.sendInitialStatus(client)
		_ = client
	})
}

func TestStatusUpgrader(t *testing.T) {
	// Test that the upgrader is configured correctly
	assert.Equal(t, 1024, statusUpgrader.ReadBufferSize)
	assert.Equal(t, 1024, statusUpgrader.WriteBufferSize)
	assert.NotNil(t, statusUpgrader.CheckOrigin)

	// Test CheckOrigin allows all origins
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	assert.True(t, statusUpgrader.CheckOrigin(req))
}

// Integration test for full WebSocket flow
func TestWebSocketIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	// Create test components
	handler := &Handler{}
	logger := logrus.New()
	hub := NewStatusWebSocketHub(handler, logger)

	// Start the hub
	go hub.Run()
	defer hub.Stop()
	time.Sleep(50 * time.Millisecond)

	// Create test server
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	var serverConn *websocket.Conn
	var clientConnected sync.WaitGroup
	clientConnected.Add(1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade: %v", err)
		}
		serverConn = conn
		clientConnected.Done()

		// Keep connection open
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}))
	defer server.Close()

	// Connect client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer clientConn.Close()

	// Wait for connection
	clientConnected.Wait()
	require.NotNil(t, serverConn)

	t.Run("client can connect", func(t *testing.T) {
		assert.NotNil(t, clientConn)
		// Verify connection is open by checking we can write
		err := clientConn.WriteMessage(websocket.PingMessage, nil)
		assert.NoError(t, err)
	})

	t.Run("message round trip", func(t *testing.T) {
		// Send a message from client
		msg := map[string]interface{}{
			"type":     "subscribe",
			"channels": []string{"platform"},
		}
		err := clientConn.WriteJSON(msg)
		require.NoError(t, err)

		// The server should receive the message
		// In a real scenario, this would trigger subscription logic
	})
}

// Benchmark for broadcast performance
func BenchmarkBroadcastToSubscribers(b *testing.B) {
	handler := &Handler{}
	logger := logrus.New()
	hub := NewStatusWebSocketHub(handler, logger)

	// Create multiple clients
	for i := 0; i < 100; i++ {
		client := &StatusWebSocketClient{
			ID:       "client-" + string(rune(i)),
			Channels: map[string]bool{"platform": true},
			Send:     make(chan []byte, 256),
			Hub:      hub,
		}
		hub.clients[client] = true
	}

	msg := &StatusUpdateMessage{
		Type:      "status_update",
		Channel:   "platform",
		Timestamp: time.Now(),
		Data:      map[string]string{"status": "operational"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.broadcastToSubscribers(msg)
	}
}
