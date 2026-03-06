package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/gorilla/websocket"
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
		}
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
	hub    *WebSocketHub
	logger *logrus.Logger
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *WebSocketHub, logger *logrus.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:    hub,
		logger: logger,
	}
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

// SubscribeToNotifications listens to PostgreSQL notifications and broadcasts to WebSocket clients
func (h *WebSocketHandler) SubscribeToNotifications(ctx context.Context, db notification.Repository) {
	// This would typically listen to PostgreSQL NOTIFY events
	// For now, we'll leave this as a placeholder for the implementation
	h.logger.Info("Starting PostgreSQL notification subscription")

	// In a real implementation, you would:
	// 1. Listen to the PostgreSQL NOTIFY channel
	// 2. When a notification is received, broadcast it to the WebSocket clients
	// 3. Handle context cancellation for graceful shutdown

	<-ctx.Done()
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
