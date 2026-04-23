package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// WebSocket upgrader configuration (origin check aligned with CORS allowlist)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     middleware.IsOriginAllowedForRequest,
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
	hub    *WebSocketHub
}

// WebSocketHub manages all WebSocket connections
type WebSocketHub struct {
	clients    map[string][]*WebSocketClient
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	broadcast  chan *WebSocketMessage
	logger     *logrus.Logger
	stop       chan struct{} // Stop signal for graceful shutdown
}

// WebSocketMessage represents a message sent via WebSocket
type WebSocketMessage struct {
	UserID  string          `json:"-"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(logger *logrus.Logger) *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[string][]*WebSocketClient),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		broadcast:  make(chan *WebSocketMessage, 100),
		logger:     logger,
		stop:       make(chan struct{}),
	}
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.UserID] = append(h.clients[client.UserID], client)
			h.logger.WithField("user_id", client.UserID).Debug("WebSocket client registered")

		case client := <-h.unregister:
			if clients, ok := h.clients[client.UserID]; ok {
				for i, c := range clients {
					if c == client {
						h.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
				if len(h.clients[client.UserID]) == 0 {
					delete(h.clients, client.UserID)
				}
			}
			close(client.Send)
			h.logger.WithField("user_id", client.UserID).Debug("WebSocket client unregistered")

		case message := <-h.broadcast:
			if clients, ok := h.clients[message.UserID]; ok {
				data, err := json.Marshal(message)
				if err != nil {
					h.logger.WithError(err).Error("Failed to marshal WebSocket message")
					continue
				}

				for _, client := range clients {
					select {
					case client.Send <- data:
					default:
						// Client buffer full, close connection
						close(client.Send)
						h.unregister <- client
						client.Conn.Close()
					}
				}
			}

		case <-h.stop:
			h.logger.Info("WebSocket hub stopping")
			return
		}
	}
}

// Stop signals the WebSocket hub to stop
func (h *WebSocketHub) Stop() {
	close(h.stop)
}

// StopHub stops the underlying WebSocket hub and the PostgreSQL LISTEN subscription
func (h *WebSocketHandler) StopHub() {
	// Cancel the PostgreSQL LISTEN subscription context first
	if h.cancelSub != nil {
		h.cancelSub()
	}
	// Then stop the WebSocket hub
	if h.hub != nil {
		h.hub.Stop()
	}
}

// Broadcast sends a message to a specific user
func (h *WebSocketHub) Broadcast(userID string, messageType string, payload interface{}) {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal WebSocket payload")
		return
	}

	h.broadcast <- &WebSocketMessage{
		UserID:  userID,
		Type:    messageType,
		Payload: payloadData,
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *WebSocketClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.WithError(err).Error("WebSocket read error")
			}
			break
		}
		// We don't expect messages from clients for now
		// In the future, we could handle client commands here
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// WebSocketHandler handles WebSocket connections for real-time notifications
type WebSocketHandler struct {
	hub       *WebSocketHub
	logger    *logrus.Logger
	cancelSub context.CancelFunc // Cancels the PostgreSQL LISTEN subscription
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *WebSocketHub, logger *logrus.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:    hub,
		logger: logger,
	}
}

// RunHub starts the WebSocket hub's event loop in a background goroutine.
// Call this once during server startup.
func (h *WebSocketHandler) RunHub() {
	if h.hub == nil {
		return
	}
	go h.hub.Run()
}

// RunNotificationSubscription starts the PostgreSQL LISTEN loop that pushes
// notifications to connected WebSocket clients. Call this after RunHub.
func (h *WebSocketHandler) RunNotificationSubscription(ctx context.Context, pool *pgxpool.Pool) {
	if pool == nil {
		return
	}
	// Create a cancellable context so we can stop the subscription during shutdown
	subCtx, cancel := context.WithCancel(ctx)
	h.cancelSub = cancel
	poolFactory := func() (*pgxpool.Pool, error) { return pool, nil }
	go h.SubscribeToNotifications(subCtx, poolFactory)
}

// HandleWebSocket upgrades the HTTP connection to WebSocket and handles real-time notifications
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		return
	}

	// Create client
	client := &WebSocketClient{
		UserID: user.UserID.String(),
		Conn:   conn,
		Send:   make(chan []byte, 256),
		hub:    h.hub,
	}

	// Register client
	client.hub.register <- client

	// Start pumps
	go client.writePump()
	go client.readPump()

	// Send initial connection success message
	welcomeMsg := map[string]interface{}{
		"type": "connected",
		"payload": map[string]string{
			"message": "Connected to notification stream",
		},
	}
	data, _ := json.Marshal(welcomeMsg)
	client.Send <- data

	h.logger.WithField("user_id", user.UserID).Info("WebSocket connection established")
}

// pgListenerKey is the PostgreSQL LISTEN channel used for broadcasting to all connected WebSocket clients.
const pgListenerKey = "notification_broadcast"

// SubscribeToNotifications listens to PostgreSQL NOTIFY events and pushes them to WebSocket clients.
//
// It acquires a dedicated connection from the pool so LISTEN state is isolated from regular queries
// and is cleared automatically when the connection is closed. The caller passes a factory that
// returns a fresh pgx.Pool on each invocation so the listener connection can be replaced after
// transient errors without poisoning the main pool.
func (h *WebSocketHandler) SubscribeToNotifications(ctx context.Context, poolFactory func() (*pgxpool.Pool, error)) {
	h.logger.Info("Starting PostgreSQL notification subscription")

	// Track all active subscriptions so they can be unregistered on shutdown.
	var wg sync.WaitGroup
	var subMu sync.Mutex
	subs := make(map[string]context.CancelFunc)

	// Helper to (re)connect a LISTEN loop for a given channel.
	connect := func(ctx context.Context, channel string) {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			pool, err := poolFactory()
			if err != nil {
				h.logger.WithError(err).Error("Failed to get connection pool for notification subscription")
				time.Sleep(5 * time.Second)
				continue
			}

			conn, err := pool.Acquire(ctx)
			if err != nil {
				h.logger.WithError(err).Error("Failed to acquire connection for notification subscription")
				time.Sleep(5 * time.Second)
				continue
			}

			// Every session starts with LISTEN so pg_notify payloads are received
			// as Notifications; Exec here creates the persistent subscription.
			if _, err := conn.Conn().Exec(ctx, "LISTEN "+channel); err != nil {
				conn.Release()
				h.logger.WithError(err).Errorf("Failed to execute LISTEN %s", channel)
				time.Sleep(5 * time.Second)
				continue
			}

			h.logger.WithField("channel", channel).Info("LISTEN subscription active")

			// Receive loop – exits only on context cancellation or connection error.
			for {
				notification, err := conn.Conn().WaitForNotification(ctx)
				if err != nil {
					if ctx.Err() != nil {
						break
					}
					h.logger.WithError(err).Warn("Notification subscription connection dropped, reconnecting")
					break
				}
				// Payload is JSON: {"type":"notification","user_id":"...","notification_id":"..."}
				var payload struct {
					Type           string `json:"type"`
					UserID         string `json:"user_id"`
					NotificationID string `json:"notification_id"`
					Title          string `json:"title"`
					Body           string `json:"body"`
					Category       string `json:"category"`
					Priority       string `json:"priority"`
					CreatedAt      string `json:"created_at"`
				}
				if err := json.Unmarshal([]byte(notification.Payload), &payload); err != nil {
					h.logger.WithError(err).Warn("Failed to parse notification payload")
					return
				}

				// Only broadcast "notification" type messages.
				if payload.Type != "notification" || payload.UserID == "" {
					continue
				}

				wsPayload := NotificationPayload{
					ID:        payload.NotificationID,
					Type:      "notification",
					Category:  payload.Category,
					Title:     payload.Title,
					Body:      payload.Body,
					Priority:  payload.Priority,
					CreatedAt: time.Now(),
					Data: map[string]interface{}{
						"user_id": payload.UserID,
					},
				}

				h.hub.Broadcast(payload.UserID, "notification", wsPayload)
			}

			conn.Release()

			if ctx.Err() != nil {
				return
			}

			// Connection dropped or errored – back off and retry.
			if err != nil && !strings.Contains(err.Error(), "context canceled") {
				h.logger.WithError(err).Warn("Notification subscription connection dropped, reconnecting")
			}
			time.Sleep(2 * time.Second)
		}
	}

	// Subscribe to the global broadcast channel and to each user's personal channel.
	channels := []string{pgListenerKey}
	for {
		// Re-check context before creating subscriptions.
		if ctx.Err() != nil {
			break
		}

		subCtx, cancel := context.WithCancel(ctx)
		subMu.Lock()
		for _, ch := range channels {
			wg.Add(1)
			go connect(subCtx, ch)
			subs[ch] = cancel
		}
		subMu.Unlock()

		// Wait until context is cancelled to exit.
		<-ctx.Done()

		// Clean up all subscription contexts.
		subMu.Lock()
		for _, cancel := range subs {
			cancel()
		}
		subMu.Unlock()

		wg.Wait()
		break
	}

	h.logger.Info("PostgreSQL notification subscription stopped")
}

// NotificationPayload represents the payload for notification messages
type NotificationPayload struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Category  string                 `json:"category"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Priority  string                 `json:"priority"`
	CreatedAt time.Time              `json:"created_at"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// BroadcastNotification broadcasts a notification to a user's WebSocket clients
func (h *WebSocketHandler) BroadcastNotification(userID string, n *notification.Notification) {
	payload := NotificationPayload{
		ID:        n.ID.String(),
		Type:      n.Type,
		Category:  n.Category,
		Title:     n.Title,
		Body:      n.Body,
		Priority:  n.Priority,
		CreatedAt: n.CreatedAt,
		Data:      n.Data,
	}

	h.hub.Broadcast(userID, "notification", payload)
}

// RegisterWebSocketRoute registers the WebSocket route
func (h *WebSocketHandler) RegisterWebSocketRoute(router interface{}) {
	// This would be called during route registration
	// The actual route registration would be done in the server setup
}
