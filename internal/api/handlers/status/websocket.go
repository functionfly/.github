package status

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// WebSocket upgrader configuration (origin check aligned with CORS allowlist)
var statusUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     middleware.IsOriginAllowedForRequest,
}

// StatusWebSocketClient represents a connected WebSocket client
type StatusWebSocketClient struct {
	ID       string
	Channels map[string]bool
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *StatusWebSocketHub
	IsAdmin  bool
}

// StatusWebSocketHub manages all WebSocket connections for status updates
type StatusWebSocketHub struct {
	clients    map[*StatusWebSocketClient]bool
	register   chan *StatusWebSocketClient
	unregister chan *StatusWebSocketClient
	broadcast  chan *StatusUpdateMessage
	handler    *Handler
	logger     *logrus.Logger
	running    bool
}

// NewStatusWebSocketHub creates a new WebSocket hub for status updates
func NewStatusWebSocketHub(handler *Handler, logger *logrus.Logger) *StatusWebSocketHub {
	return &StatusWebSocketHub{
		clients:    make(map[*StatusWebSocketClient]bool),
		register:   make(chan *StatusWebSocketClient),
		unregister: make(chan *StatusWebSocketClient),
		broadcast:  make(chan *StatusUpdateMessage, 100),
		handler:    handler,
		logger:     logger,
		running:    false,
	}
}

// Run starts the WebSocket hub
func (h *StatusWebSocketHub) Run() {
	if h.running {
		return
	}
	h.running = true

	// Start background broadcasters
	go h.platformStatusBroadcaster()
	go h.providerMetricsBroadcaster()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.logger.WithField("client_id", client.ID).Debug("Status WebSocket client registered")

			// Send initial status to client
			go h.sendInitialStatus(client)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				h.logger.WithField("client_id", client.ID).Debug("Status WebSocket client unregistered")
			}

		case message := <-h.broadcast:
			h.broadcastToSubscribers(message)
		}
	}
}

// Stop stops the WebSocket hub
func (h *StatusWebSocketHub) Stop() {
	h.running = false
}

// broadcastToSubscribers sends a message to all subscribed clients
func (h *StatusWebSocketHub) broadcastToSubscribers(message *StatusUpdateMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal WebSocket message")
		return
	}

	for client := range h.clients {
		// Check if client is subscribed to this channel
		if client.Channels[message.Channel] || client.Channels["all"] {
			select {
			case client.Send <- data:
			default:
				// Client buffer full, close connection
				close(client.Send)
				delete(h.clients, client)
				client.Conn.Close()
			}
		}
	}
}

// sendInitialStatus sends the current status to a newly connected client
func (h *StatusWebSocketHub) sendInitialStatus(client *StatusWebSocketClient) {
	ctx := context.Background()

	// Send platform status
	if client.Channels["platform"] || client.Channels["all"] {
		status, err := h.getCurrentPlatformStatus()
		if err == nil {
			msg := &StatusUpdateMessage{
				Type:      "status_update",
				Channel:   "platform",
				Timestamp: time.Now(),
				Data:      status,
			}
			data, _ := json.Marshal(msg)
			select {
			case client.Send <- data:
			default:
			}
		}
	}

	// Send provider status
	if client.Channels["providers"] || client.Channels["all"] {
		providers, err := h.getCurrentProviderStatus(ctx)
		if err == nil {
			msg := &StatusUpdateMessage{
				Type:      "status_update",
				Channel:   "providers",
				Timestamp: time.Now(),
				Data:      providers,
			}
			data, _ := json.Marshal(msg)
			select {
			case client.Send <- data:
			default:
			}
		}
	}

	// Send active incidents
	if client.Channels["incidents"] || client.Channels["all"] {
		incidents, err := h.handler.repo.GetActiveIncidents(ctx)
		if err == nil {
			msg := &StatusUpdateMessage{
				Type:      "status_update",
				Channel:   "incidents",
				Timestamp: time.Now(),
				Data:      incidents,
			}
			data, _ := json.Marshal(msg)
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

// platformStatusBroadcaster broadcasts platform status every 10 seconds
func (h *StatusWebSocketHub) platformStatusBroadcaster() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		if !h.running {
			return
		}

		select {
		case <-ticker.C:
			status, err := h.getCurrentPlatformStatus()
			if err != nil {
				h.logger.WithError(err).Warn("Failed to get platform status for broadcast")
				continue
			}

			msg := &StatusUpdateMessage{
				Type:      "status_update",
				Channel:   "platform",
				Timestamp: time.Now(),
				Data:      status,
			}

			h.broadcast <- msg
		}
	}
}

// providerMetricsBroadcaster broadcasts provider metrics every 5 seconds
func (h *StatusWebSocketHub) providerMetricsBroadcaster() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	ctx := context.Background()

	for {
		if !h.running {
			return
		}

		select {
		case <-ticker.C:
			providers, err := h.getCurrentProviderStatus(ctx)
			if err != nil {
				h.logger.WithError(err).Warn("Failed to get provider status for broadcast")
				continue
			}

			msg := &StatusUpdateMessage{
				Type:      "status_update",
				Channel:   "providers",
				Timestamp: time.Now(),
				Data:      providers,
			}

			h.broadcast <- msg
		}
	}
}

// getCurrentPlatformStatus gets the current platform status
func (h *StatusWebSocketHub) getCurrentPlatformStatus() (*PlatformStatus, error) {
	ctx := context.Background()

	// Get platform health percentage
	healthPercent, _ := h.handler.prometheus.GetPlatformHealthPercentage(ctx)

	// Get active incidents
	incidents, err := h.handler.repo.GetActiveIncidents(ctx)
	if err != nil {
		incidents = []Incident{}
	}

	// Get upcoming maintenance
	maintenance, err := h.handler.repo.GetUpcomingMaintenance(ctx)
	if err != nil {
		maintenance = []MaintenanceWindow{}
	}

	// Determine status
	status, indicator, description := h.handler.determinePlatformStatus(healthPercent, incidents)

	// Get component summaries
	components := h.handler.getComponentSummaries(ctx)

	// Convert maintenance to summaries
	maintenanceSummaries := make([]MaintenanceSummary, len(maintenance))
	for i, m := range maintenance {
		maintenanceSummaries[i] = MaintenanceSummary{
			ID:             m.ID,
			Title:          m.Title,
			Status:         m.Status,
			ScheduledStart: m.ScheduledStart,
			ScheduledEnd:   m.ScheduledEnd,
		}
	}

	return &PlatformStatus{
		Status:      status,
		Indicator:   indicator,
		Description: description,
		UpdatedAt:   time.Now(),
		Components:  components,
		Incidents:   incidents,
		Maintenance: maintenanceSummaries,
	}, nil
}

// getCurrentProviderStatus gets the current provider status
func (h *StatusWebSocketHub) getCurrentProviderStatus(ctx context.Context) ([]ProviderStatus, error) {
	providers, err := h.handler.repo.GetProviderStatus(ctx)
	if err != nil {
		return nil, err
	}

	// Enhance with Prometheus metrics
	for i := range providers {
		latency, err := h.handler.prometheus.GetProbeLatency(ctx, providers[i].Name, "", 0.95)
		if err == nil && latency.Data != nil {
			for _, result := range latency.Data.Result {
				if len(result.Value) >= 2 {
					val := parseValue(result.Value[1])
					providers[i].Summary.AvgLatencyMs = val
				}
			}
		}
	}

	return providers, nil
}

// BroadcastIncidentUpdate broadcasts an incident update to all subscribers
func (h *StatusWebSocketHub) BroadcastIncidentUpdate(incident *Incident) {
	if !h.running {
		return
	}

	msg := &StatusUpdateMessage{
		Type:      "status_update",
		Channel:   "incidents",
		Timestamp: time.Now(),
		Data:      incident,
	}

	select {
	case h.broadcast <- msg:
	default:
		h.logger.Warn("Broadcast channel full, dropping incident update")
	}
}

// BroadcastProviderUpdate broadcasts a provider update
func (h *StatusWebSocketHub) BroadcastProviderUpdate(update *ProviderUpdate) {
	if !h.running {
		return
	}

	msg := &StatusUpdateMessage{
		Type:      "status_update",
		Channel:   "providers",
		Timestamp: time.Now(),
		Data:      update,
	}

	select {
	case h.broadcast <- msg:
	default:
		h.logger.Warn("Broadcast channel full, dropping provider update")
	}
}

// HandleWebSocket handles WebSocket connections for status updates
func (h *Handler) HandleWebSocket(hub *StatusWebSocketHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Upgrade HTTP connection to WebSocket
		conn, err := statusUpgrader.Upgrade(w, r, nil)
		if err != nil {
			logrus.WithError(err).Error("Failed to upgrade WebSocket connection")
			return
		}

		// Check if user is admin (optional token in query param)
		isAdmin := false
		if token := r.URL.Query().Get("token"); token != "" {
			claims, err := h.authSvc.ValidateToken(token)
			if err == nil && h.isAdmin(claims) {
				isAdmin = true
			}
		}

		// Create client
		client := &StatusWebSocketClient{
			ID:       generateClientID(),
			Channels: make(map[string]bool),
			Conn:     conn,
			Send:     make(chan []byte, 256),
			Hub:      hub,
			IsAdmin:  isAdmin,
		}

		// Register client
		hub.register <- client

		// Start goroutines for reading and writing
		go client.writePump()
		go client.readPump(hub)
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *StatusWebSocketClient) readPump(hub *StatusWebSocketHub) {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.WithError(err).Warn("WebSocket error")
			}
			break
		}

		// Handle subscription messages
		var subscribeMsg SubscribeMessage
		if err := json.Unmarshal(message, &subscribeMsg); err == nil {
			if subscribeMsg.Type == "subscribe" {
				for _, channel := range subscribeMsg.Channels {
					c.Channels[channel] = true
				}

				// Send acknowledgment
				ack := map[string]interface{}{
					"type":     "subscribed",
					"channels": subscribeMsg.Channels,
				}
				data, _ := json.Marshal(ack)
				c.Send <- data
			}
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *StatusWebSocketClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return "client_" + time.Now().Format("20060102150405") + "_" + strconv.Itoa(int(time.Now().UnixNano()%10000))
}

// HandleWebSocketStatus is the HTTP handler for status WebSocket connections
func (h *Handler) HandleWebSocketStatus(w http.ResponseWriter, r *http.Request) {
	// This method creates a temporary hub if needed, or uses an existing one
	// In production, the hub should be created once and shared
	hub := NewStatusWebSocketHub(h, logrus.New())
	go hub.Run()

	// Use the WebSocket handler
	h.HandleWebSocket(hub)(w, r)
}
