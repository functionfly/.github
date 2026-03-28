package support

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/support"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // No origin header (e.g., direct WebSocket client)
		}

		allowedOriginsStr := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
		isDev := os.Getenv("DEVELOPMENT") == "true" || os.Getenv("NODE_ENV") == "development"
		isLocalhost := strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")

		// In development, allow localhost origins without explicit CORS_ALLOWED_ORIGINS
		if isDev && isLocalhost && allowedOriginsStr == "" {
			return true
		}

		// Check against explicit allowed origins list
		if allowedOriginsStr != "" {
			allowedOrigins := strings.Split(allowedOriginsStr, ",")
			for _, allowed := range allowedOrigins {
				allowed = strings.TrimSpace(allowed)
				if allowed == "*" || allowed == origin {
					return true
				}
			}
		}

		// Reject in production or if origin doesn't match
		logrus.WithFields(logrus.Fields{"origin": origin}).Warn("WebSocket origin rejected")
		return false
	},
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ChatMessagePayload represents a chat message payload
type ChatMessagePayload struct {
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
	MessageID      string `json:"message_id,omitempty"`
	AuthorType     string `json:"author_type"`
	Timestamp      int64  `json:"timestamp"`
}

// WebSocketHub manages WebSocket connections for support chat
type WebSocketHub struct {
	// Registered connections by conversation ID
	conversations map[uuid.UUID]map[*WebSocketConn]bool

	// Registered connections by user ID
	users map[uuid.UUID]map[*WebSocketConn]bool

	// Staff connections
	staff map[*WebSocketConn]bool

	// Register requests
	register chan *WebSocketConn

	// Unregister requests
	unregister chan *WebSocketConn

	// Broadcast to conversation
	broadcast chan *BroadcastMessage

	// Service reference
	service *support.Service

	// Redis-backed pub/sub for server -> client real-time updates.
	redis support.RedisClient

	// Auth service for WebSocket connections that pass tokens via query params.
	authSvc *auth.AuthService

	// Logger
	logger *logrus.Logger

	// Mutex for thread safety
	mu sync.RWMutex
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	ConversationID uuid.UUID
	Message       *WebSocketMessage
	Exclude       *WebSocketConn
}

// WebSocketConn represents a WebSocket connection
type WebSocketConn struct {
	conn    *websocket.Conn
	userID  uuid.UUID
	isStaff bool
	send    chan []byte
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(service *support.Service, redis support.RedisClient, authSvc *auth.AuthService, logger *logrus.Logger) *WebSocketHub {
	if logger == nil {
		logger = logrus.New()
	}
	return &WebSocketHub{
		conversations: make(map[uuid.UUID]map[*WebSocketConn]bool),
		users:         make(map[uuid.UUID]map[*WebSocketConn]bool),
		staff:         make(map[*WebSocketConn]bool),
		register:       make(chan *WebSocketConn),
		unregister:    make(chan *WebSocketConn),
		broadcast:     make(chan *BroadcastMessage),
		service:       service,
		redis:         redis,
		authSvc:       authSvc,
		logger:        logger,
	}
}

// Run starts the hub
func (h *WebSocketHub) Run() {
	if h.redis != nil {
		go h.subscribeToMessageCreated()
	}

	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			// Initialize user map if needed
			if _, ok := h.users[conn.userID]; !ok {
				h.users[conn.userID] = make(map[*WebSocketConn]bool)
			}
			h.users[conn.userID][conn] = true
			if conn.isStaff {
				h.staff[conn] = true
			}
			h.mu.Unlock()
			h.logger.WithFields(logrus.Fields{
				"user_id":  conn.userID,
				"is_staff": conn.isStaff,
			}).Info("WebSocket client connected")

		case conn := <-h.unregister:
			h.mu.Lock()
			// Remove from users
			if conns, ok := h.users[conn.userID]; ok {
				delete(conns, conn)
				if len(conns) == 0 {
					delete(h.users, conn.userID)
				}
			}
			// Remove from staff
			if conn.isStaff {
				delete(h.staff, conn)
			}
			// Remove from all conversations
			for convID := range h.conversations {
				delete(h.conversations[convID], conn)
			}
			close(conn.send)
			h.mu.Unlock()
			h.logger.WithField("user_id", conn.userID).Info("WebSocket client disconnected")

		case msg := <-h.broadcast:
			h.mu.RLock()
			if conns, ok := h.conversations[msg.ConversationID]; ok {
				data, _ := json.Marshal(msg.Message)
				for conn := range conns {
					if conn != msg.Exclude {
						select {
						case conn.send <- data:
						default:
							close(conn.send)
							delete(conns, conn)
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// subscribeToMessageCreated forwards support messages published to Redis
// to all WebSocket connections joined to the corresponding conversation.
func (h *WebSocketHub) subscribeToMessageCreated() {
	ch, err := h.redis.Subscribe(context.Background(), "message.created")
	if err != nil {
		h.logger.WithError(err).Warn("support websocket: failed to subscribe to message.created")
		return
	}

	for payload := range ch {
		var msg support.SupportMessage
		if err := json.Unmarshal([]byte(payload), &msg); err != nil {
			h.logger.WithError(err).Warn("support websocket: failed to parse message.created payload")
			continue
		}

		// Convert service message to a websocket event payload.
		// Frontend expects the SupportMessage JSON shape.
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			h.logger.WithError(err).Warn("support websocket: failed to marshal message payload")
			continue
		}

		wsMsg := &WebSocketMessage{
			Type:    "new_message",
			Payload: msgBytes,
		}
		h.BroadcastToConversation(msg.ConversationID, wsMsg, nil)
	}
}

// JoinConversation adds a connection to a conversation room
func (h *WebSocketHub) JoinConversation(conn *WebSocketConn, conversationID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.conversations[conversationID]; !ok {
		h.conversations[conversationID] = make(map[*WebSocketConn]bool)
	}
	h.conversations[conversationID][conn] = true

	// Notify others that user joined
	msg := &WebSocketMessage{
		Type: "user_joined",
	}
	data, _ := json.Marshal(msg)
	for c := range h.conversations[conversationID] {
		if c != conn {
			select {
			case c.send <- data:
			default:
			}
		}
	}
}

// LeaveConversation removes a connection from a conversation room
func (h *WebSocketHub) LeaveConversation(conn *WebSocketConn, conversationID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.conversations[conversationID]; ok {
		delete(conns, conn)
	}
}

// BroadcastToConversation sends a message to all users in a conversation
func (h *WebSocketHub) BroadcastToConversation(conversationID uuid.UUID, msg *WebSocketMessage, exclude *WebSocketConn) {
	h.broadcast <- &BroadcastMessage{
		ConversationID: conversationID,
		Message:        msg,
		Exclude:       exclude,
	}
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *WebSocketHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)

	// For WebSocket connections, auth is often passed via query param because
	// browsers can't set Authorization headers on the handshake.
	if user == nil && h.authSvc != nil {
		if token := r.URL.Query().Get("token"); token != "" {
			claims, err := h.authSvc.ValidateToken(token)
			if err == nil {
				user = claims
			}
		}
	}

	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	isStaff := false
	switch user.Role {
	case "super_admin", "admin", "support":
		isStaff = true
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.WithError(err).Error("WebSocket upgrade failed")
		return
	}

	wsConn := &WebSocketConn{
		conn:    conn,
		userID:  user.UserID,
		isStaff: isStaff,
		send:    make(chan []byte, 256),
	}

	h.register <- wsConn

	// Start goroutines for reading and writing
	go h.writePump(wsConn)
	go h.readPump(wsConn)
}

func (h *WebSocketHub) readPump(conn *WebSocketConn) {
	defer func() {
		h.unregister <- conn
		conn.conn.Close()
	}()

	conn.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.conn.SetPongHandler(func(string) error {
		conn.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := conn.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.WithError(err).Error("WebSocket read error")
			}
			break
		}

		var wsMsg WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			h.logger.WithError(err).Error("Failed to parse WebSocket message")
			continue
		}

		h.handleMessage(conn, &wsMsg)
	}
}

func (h *WebSocketHub) writePump(conn *WebSocketConn) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		conn.conn.Close()
	}()

	for {
		select {
		case message, ok := <-conn.send:
			conn.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Drain queued messages
			n := len(conn.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-conn.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			conn.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WebSocketHub) handleMessage(conn *WebSocketConn, msg *WebSocketMessage) {
	switch msg.Type {
	case "join_conversation":
		var payload struct {
			ConversationID string `json:"conversation_id"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			h.logger.WithError(err).Error("Failed to parse join_conversation payload")
			return
		}
		convID, err := uuid.Parse(payload.ConversationID)
		if err != nil {
			h.logger.WithError(err).Error("Invalid conversation ID")
			return
		}
		h.JoinConversation(conn, convID)

	case "leave_conversation":
		var payload struct {
			ConversationID string `json:"conversation_id"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			h.logger.WithError(err).Error("Failed to parse leave_conversation payload")
			return
		}
		convID, err := uuid.Parse(payload.ConversationID)
		if err != nil {
			return
		}
		h.LeaveConversation(conn, convID)

	case "chat_message":
		var payload ChatMessagePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			h.logger.WithError(err).Error("Failed to parse chat_message payload")
			return
		}

		convID, err := uuid.Parse(payload.ConversationID)
		if err != nil {
			h.logger.WithError(err).Error("Invalid conversation ID in chat_message")
			return
		}

		// Get author type
		authorType := support.AuthorUser
		if conn.isStaff {
			authorType = support.AuthorStaff
		}

		// Send message via service
		message, err := h.service.SendMessage(context.Background(), convID, conn.userID, authorType, payload.Content)
		if err != nil {
			h.logger.WithError(err).Error("Failed to send message via service")
			return
		}

		// When Redis is enabled, subscribers will forward the persisted message.
		// When Redis is disabled, fall back to broadcasting directly.
		if h.redis == nil {
			msgBytes, marshalErr := json.Marshal(message)
			if marshalErr == nil {
				responseMsg := &WebSocketMessage{Type: "new_message", Payload: msgBytes}
				h.BroadcastToConversation(convID, responseMsg, nil)
			}
		}

	case "typing":
		var payload struct {
			ConversationID string `json:"conversation_id"`
			Typing         bool   `json:"typing"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		convID, _ := uuid.Parse(payload.ConversationID)

		// Broadcast typing indicator
		typingPayload, _ := json.Marshal(map[string]interface{}{
			"user_id":         conn.userID,
			"conversation_id": convID,
			"typing":          payload.Typing,
		})
		typingMsg := &WebSocketMessage{
			Type:    "user_typing",
			Payload: typingPayload,
		}
		h.BroadcastToConversation(convID, typingMsg, conn)

	case "ping":
		response, _ := json.Marshal(&WebSocketMessage{Type: "pong"})
		conn.send <- response
	}
}

// HandleWebSocketWithRouter handles WebSocket connections with mux routing
func (h *WebSocketHub) HandleWebSocketWithRouter(w http.ResponseWriter, r *http.Request) {
	h.HandleWebSocket(w, r)
}

// RegisterWebSocketRoutes registers WebSocket routes on a router
func RegisterWebSocketRoutes(router *mux.Router, hub *WebSocketHub) {
	router.HandleFunc("/v1/support/ws", hub.HandleWebSocketWithRouter).Methods("GET")
}

// Helper to send WebSocket message to a specific user
func (h *WebSocketHub) SendToUser(userID uuid.UUID, msg *WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, _ := json.Marshal(msg)
	if conns, ok := h.users[userID]; ok {
		for conn := range conns {
			select {
			case conn.send <- data:
			default:
			}
		}
	}
}

// Helper to notify staff of new emergency
func (h *WebSocketHub) NotifyStaffNewEmergency(emergency *support.EmergencyFixRequest) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	payload, _ := json.Marshal(emergency)
	msg := &WebSocketMessage{
		Type:    "new_emergency",
		Payload: payload,
	}
	data, _ := json.Marshal(msg)

	for conn := range h.staff {
		select {
		case conn.send <- data:
		default:
		}
	}
}

// Helper to notify user of staff joining
func (h *WebSocketHub) NotifyUserStaffJoined(conversationID, staffID uuid.UUID) {
	payload, _ := json.Marshal(map[string]interface{}{
		"conversation_id": conversationID,
		"staff_id":        staffID,
		"message":         "A support staff member has joined",
	})
	msg := &WebSocketMessage{
		Type:    "staff_joined",
		Payload: payload,
	}

	h.BroadcastToConversation(conversationID, msg, nil)
}

// Log the WebSocketHub creation
func init() {
	log.Println("Support WebSocket hub module loaded")
}
