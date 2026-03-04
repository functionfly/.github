package flywheel

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/flywheel"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// WebSocketHub manages WebSocket connections for real-time updates
type WebSocketHub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     *logrus.Logger
}

// Client represents a WebSocket client
type Client struct {
	hub      *WebSocketHub
	conn     *websocket.Conn
	send     chan []byte
	userID   string
	threadID string
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(logger *logrus.Logger) *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.WithField("user_id", client.userID).Info("Client registered")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.logger.WithField("user_id", client.userID).Info("Client unregistered")

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *WebSocketHub) Broadcast(msgType string, payload interface{}) {
	msg := WebSocketMessage{
		Type:    msgType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal broadcast message")
		return
	}
	h.broadcast <- data
}

// BroadcastToThread sends a message to clients subscribed to a specific thread
func (h *WebSocketHub) BroadcastToThread(threadID string, msgType string, payload interface{}) {
	msg := WebSocketMessage{
		Type:    msgType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal thread broadcast message")
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.threadID == threadID {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// HandleWebSocket handles WebSocket connections
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user from context (may be nil for anonymous)
	user := middleware.GetUserFromContext(r)

	// Get thread ID from query params
	threadID := r.URL.Query().Get("thread_id")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to upgrade WebSocket")
		return
	}

	userID := ""
	if user != nil {
		userID = user.UserID.String()
	}

	client := &Client{
		hub:      h.wsHub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		threadID: threadID,
	}

	client.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024) // 512KB max message size

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.WithError(err).Error("WebSocket error")
			}
			break
		}

		// Handle incoming messages (ping/pong, subscriptions, etc.)
		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err == nil {
			switch msg.Type {
			case "subscribe_thread":
				if payload, ok := msg.Payload.(map[string]interface{}); ok {
					if threadID, ok := payload["thread_id"].(string); ok {
						c.threadID = threadID
					}
				}
			case "ping":
				c.send <- []byte(`{"type":"pong"}`)
			}
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.hub.logger.WithError(err).Error("Failed to write WebSocket message")
				return
			}
		}
	}
}

// BroadcastNewReply broadcasts a new reply to thread subscribers
func (h *Handler) BroadcastNewReply(threadID string, reply *flywheel.Reply) {
	if h.wsHub == nil {
		return
	}
	h.wsHub.BroadcastToThread(threadID, "new_reply", reply)
}

// BroadcastReputationUpdate broadcasts a reputation update
func (h *Handler) BroadcastReputationUpdate(userID string, scores *flywheel.ReputationScores) {
	if h.wsHub == nil {
		return
	}
	h.wsHub.Broadcast("reputation_update", map[string]interface{}{
		"user_id": userID,
		"scores":  scores,
	})
}

// BroadcastExecutionComplete broadcasts execution completion
func (h *Handler) BroadcastExecutionComplete(threadID string, execution *flywheel.Execution) {
	if h.wsHub == nil {
		return
	}
	h.wsHub.BroadcastToThread(threadID, "execution_complete", execution)
}

// BroadcastChallengeUpdate broadcasts challenge updates
func (h *Handler) BroadcastChallengeUpdate(challengeID string, update interface{}) {
	if h.wsHub == nil {
		return
	}
	h.wsHub.Broadcast("challenge_update", map[string]interface{}{
		"challenge_id": challengeID,
		"update":       update,
	})
}
