package conversations

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/conversations"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var convWSUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     middleware.IsOriginAllowedForRequest,
}

// ConvWSMessage represents a message sent over the conversations WebSocket.
type ConvWSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ConvWSClient represents a single WebSocket connection.
type ConvWSClient struct {
	conn   *websocket.Conn
	userID uuid.UUID
	send   chan []byte
}

// ConvBroadcastMessage is queued for delivery to all participants of a conversation.
type ConvBroadcastMessage struct {
	ConversationID uuid.UUID
	Message        *ConvWSMessage
	ExcludeConn    *ConvWSClient
}

// ConversationWebSocketHub manages WebSocket connections for real-time conversation updates.
type ConversationWebSocketHub struct {
	// conversations maps conversation IDs to the set of connected clients.
	conversations map[uuid.UUID]map[*ConvWSClient]bool

	// users maps user IDs to their connected clients (for targeted sends).
	users map[uuid.UUID]map[*ConvWSClient]bool

	// register / unregister channels for connection lifecycle.
	register   chan *ConvWSClient
	unregister chan *ConvWSClient

	// broadcast delivers a message to all clients in a conversation room.
	broadcast chan *ConvBroadcastMessage

	logger *logrus.Logger
	mu     sync.RWMutex
}

// NewConversationWebSocketHub creates a new hub.
func NewConversationWebSocketHub(logger *logrus.Logger) *ConversationWebSocketHub {
	if logger == nil {
		logger = logrus.New()
	}
	return &ConversationWebSocketHub{
		conversations: make(map[uuid.UUID]map[*ConvWSClient]bool),
		users:         make(map[uuid.UUID]map[*ConvWSClient]bool),
		register:      make(chan *ConvWSClient),
		unregister:    make(chan *ConvWSClient),
		broadcast:     make(chan *ConvBroadcastMessage, 64),
		logger:        logger,
	}
}

// Run is the hub event loop — must be called in a goroutine.
func (h *ConversationWebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.users[client.userID]; !ok {
				h.users[client.userID] = make(map[*ConvWSClient]bool)
			}
			h.users[client.userID][client] = true
			h.mu.Unlock()
			h.logger.WithField("user_id", client.userID).Debug("Conversations WS client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if conns, ok := h.users[client.userID]; ok {
				delete(conns, client)
				if len(conns) == 0 {
					delete(h.users, client.userID)
				}
			}
			for convID := range h.conversations {
				delete(h.conversations[convID], client)
			}
			close(client.send)
			h.mu.Unlock()
			h.logger.WithField("user_id", client.userID).Debug("Conversations WS client disconnected")

		case msg := <-h.broadcast:
			h.mu.RLock()
			if conns, ok := h.conversations[msg.ConversationID]; ok {
				data, _ := json.Marshal(msg.Message)
				for client := range conns {
					if client == msg.ExcludeConn {
						continue
					}
					select {
					case client.send <- data:
					default:
						close(client.send)
						delete(conns, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// JoinConversation subscribes a client to a conversation room.
func (h *ConversationWebSocketHub) JoinConversation(client *ConvWSClient, conversationID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.conversations[conversationID]; !ok {
		h.conversations[conversationID] = make(map[*ConvWSClient]bool)
	}
	h.conversations[conversationID][client] = true
}

// LeaveConversation unsubscribes a client from a conversation room.
func (h *ConversationWebSocketHub) LeaveConversation(client *ConvWSClient, conversationID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.conversations[conversationID]; ok {
		delete(conns, client)
	}
}

// BroadcastToConversation queues a message for delivery to all clients in a conversation.
func (h *ConversationWebSocketHub) BroadcastToConversation(conversationID uuid.UUID, msg *ConvWSMessage, exclude *ConvWSClient) {
	h.broadcast <- &ConvBroadcastMessage{
		ConversationID: conversationID,
		Message:        msg,
		ExcludeConn:    exclude,
	}
}

// BroadcastNewMessage is a convenience method for the common case of
// broadcasting a newly-created message to all participants.
func (h *ConversationWebSocketHub) BroadcastNewMessage(msg *conversations.ConversationMessage) {
	payload, err := json.Marshal(msg)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal new message payload")
		return
	}
	h.BroadcastToConversation(msg.ConversationID, &ConvWSMessage{
		Type:    "new_message",
		Payload: payload,
	}, nil)
}

// BroadcastConversationResolved notifies participants that a conversation was resolved.
func (h *ConversationWebSocketHub) BroadcastConversationResolved(conv *conversations.Conversation) {
	payload, err := json.Marshal(conv)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal resolved conversation payload")
		return
	}
	h.BroadcastToConversation(conv.ID, &ConvWSMessage{
		Type:    "conversation_resolved",
		Payload: payload,
	}, nil)
}

// BroadcastMessageUpdated notifies participants that a message was edited.
func (h *ConversationWebSocketHub) BroadcastMessageUpdated(msg *conversations.ConversationMessage) {
	payload, err := json.Marshal(msg)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal updated message payload")
		return
	}
	h.BroadcastToConversation(msg.ConversationID, &ConvWSMessage{
		Type:    "message_updated",
		Payload: payload,
	}, nil)
}

// BroadcastMessageDeleted notifies participants that a message was soft-deleted.
func (h *ConversationWebSocketHub) BroadcastMessageDeleted(conversationID, messageID uuid.UUID) {
	payload, _ := json.Marshal(map[string]interface{}{
		"conversation_id": conversationID,
		"message_id":      messageID,
	})
	h.BroadcastToConversation(conversationID, &ConvWSMessage{
		Type:    "message_deleted",
		Payload: payload,
	}, nil)
}

// SendToUser delivers a message to every connection of a particular user.
func (h *ConversationWebSocketHub) SendToUser(userID uuid.UUID, msg *ConvWSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, _ := json.Marshal(msg)
	if conns, ok := h.users[userID]; ok {
		for client := range conns {
			select {
			case client.send <- data:
			default:
			}
		}
	}
}

// HandleWebSocket is the HTTP handler that upgrades to a WebSocket connection.
func (h *ConversationWebSocketHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	conn, err := convWSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.WithError(err).Error("Conversations WS upgrade failed")
		return
	}

	client := &ConvWSClient{
		conn:   conn,
		userID: user.UserID,
		send:   make(chan []byte, 256),
	}

	h.register <- client

	go h.writePump(client)
	go h.readPump(client)
}

func (h *ConversationWebSocketHub) readPump(client *ConvWSClient) {
	defer func() {
		h.unregister <- client
		client.conn.Close()
	}()

	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.WithError(err).Error("Conversations WS read error")
			}
			break
		}

		var wsMsg ConvWSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			h.logger.WithError(err).Error("Failed to parse conversations WS message")
			continue
		}

		h.handleClientMessage(client, &wsMsg)
	}
}

func (h *ConversationWebSocketHub) writePump(client *ConvWSClient) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Drain queued messages into the same write batch.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *ConversationWebSocketHub) handleClientMessage(client *ConvWSClient, msg *ConvWSMessage) {
	switch msg.Type {
	case "join_conversation":
		var payload struct {
			ConversationID string `json:"conversation_id"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		convID, err := uuid.Parse(payload.ConversationID)
		if err != nil {
			return
		}
		h.JoinConversation(client, convID)

	case "leave_conversation":
		var payload struct {
			ConversationID string `json:"conversation_id"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		convID, err := uuid.Parse(payload.ConversationID)
		if err != nil {
			return
		}
		h.LeaveConversation(client, convID)

	case "typing":
		var payload struct {
			ConversationID string `json:"conversation_id"`
			Typing         bool   `json:"typing"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		convID, err := uuid.Parse(payload.ConversationID)
		if err != nil {
			return
		}
		typingPayload, _ := json.Marshal(map[string]interface{}{
			"user_id":         client.userID,
			"conversation_id": convID,
			"typing":          payload.Typing,
		})
		h.BroadcastToConversation(convID, &ConvWSMessage{
			Type:    "user_typing",
			Payload: typingPayload,
		}, client)

	case "ping":
		pong, _ := json.Marshal(&ConvWSMessage{Type: "pong"})
		client.send <- pong
	}
}

// RegisterConversationWSRoute registers the conversations WebSocket endpoint on the given router.
func RegisterConversationWSRoute(router *mux.Router, hub *ConversationWebSocketHub) {
	router.HandleFunc("/conversations/ws", hub.HandleWebSocket).Methods("GET")
}
