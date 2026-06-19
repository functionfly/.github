package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     middleware.IsOriginAllowedForRequest,
}

type WebSocketHub struct {
	sessions   map[uuid.UUID]map[*WebSocketConn]bool
	register   chan *WebSocketConn
	unregister chan *WebSocketConn
	broadcast  chan *BroadcastMessage
	service    *Service
	repo       *Repository
	aiClient   *AIServiceClient
	logger     *logrus.Logger
	mu         sync.RWMutex
}

type BroadcastMessage struct {
	SessionID uuid.UUID
	Message  *WebSocketMessage
	Exclude  *WebSocketConn
}

type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type WebSocketConn struct {
	conn     *websocket.Conn
	userID   uuid.UUID
	tenantID uuid.UUID
	session  uuid.UUID
	send     chan []byte
	done     chan struct{}
}

func NewWebSocketHub(service *Service, repo *Repository, aiClient *AIServiceClient, logger *logrus.Logger) *WebSocketHub {
	if logger == nil {
		logger = logrus.New()
	}
	return &WebSocketHub{
		sessions:   make(map[uuid.UUID]map[*WebSocketConn]bool),
		register:   make(chan *WebSocketConn),
		unregister: make(chan *WebSocketConn),
		broadcast:  make(chan *BroadcastMessage),
		service:   service,
		repo:      repo,
		aiClient:  aiClient,
		logger:    logger,
	}
}

func (h *WebSocketHub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			if _, ok := h.sessions[conn.session]; !ok {
				h.sessions[conn.session] = make(map[*WebSocketConn]bool)
			}
			h.sessions[conn.session][conn] = true
			h.mu.Unlock()
			h.logger.WithFields(logrus.Fields{"user_id": conn.userID, "session": conn.session}).Info("WS client connected")

		case conn := <-h.unregister:
			h.mu.Lock()
			if conns, ok := h.sessions[conn.session]; ok {
				delete(conns, conn)
				if len(conns) == 0 {
					delete(h.sessions, conn.session)
				}
			}
			close(conn.send)
			h.mu.Unlock()
			h.logger.WithField("user_id", conn.userID).Info("WS client disconnected")

		case msg := <-h.broadcast:
			h.mu.RLock()
			if conns, ok := h.sessions[msg.SessionID]; ok {
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

func (h *WebSocketHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sessionIDStr := query.Get("session_id")
	userIDStr := query.Get("user_id")
	tenantIDStr := query.Get("tenant_id")

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid session_id"))
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid user_id"))
		return
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant_id"))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.WithError(err).Error("WebSocket upgrade failed")
		return
	}

	wsConn := &WebSocketConn{
		conn:     conn,
		userID:   userID,
		tenantID: tenantID,
		session:  sessionID,
		send:     make(chan []byte, 256),
		done:     make(chan struct{}),
	}

	h.register <- wsConn
	go h.writePump(wsConn)
	go h.readPump(wsConn)
}

func (h *WebSocketHub) readPump(conn *WebSocketConn) {
	defer func() {
		close(conn.done)
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
				h.logger.WithError(err).Error("WS read error")
			}
			break
		}

		var wsMsg WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			h.logger.WithError(err).Error("Failed to parse WS message")
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
		case <-conn.done:
			return
		}
	}
}

func (h *WebSocketHub) handleMessage(conn *WebSocketConn, msg *WebSocketMessage) {
	switch msg.Type {
	case "chat_message":
		var payload struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}

		userMsg := &ChatMessage{
			SessionID: conn.session,
			Role:      "user",
			Content:   payload.Content,
		}
		h.repo.CreateMessage(context.Background(), userMsg)

		history, _ := h.repo.ListMessages(context.Background(), conn.session, 50, 0)
		resp, err := h.service.GenerateResponse(context.Background(), &ChatRequest{
			SessionID:  conn.session,
			UserID:     conn.userID,
			TenantID:   conn.tenantID,
			Message:    payload.Content,
			History:    history,
		})

		if err != nil {
			resp = &AIResponse{Message: "Error processing request"}
		}

		assistantMsg := &ChatMessage{
			SessionID: conn.session,
			Role:      "assistant",
			Content:   resp.Message,
		}
		h.repo.CreateMessage(context.Background(), assistantMsg)

		msgBytes, _ := json.Marshal(assistantMsg)
		responseMsg := &WebSocketMessage{Type: "new_message", Payload: msgBytes}
		h.broadcast <- &BroadcastMessage{SessionID: conn.session, Message: responseMsg}

	case "typing":
		h.broadcast <- &BroadcastMessage{
			SessionID: conn.session,
			Message:   &WebSocketMessage{Type: "user_typing", Payload: json.RawMessage(`{"user_id":"` + conn.userID.String() + `"}`)},
			Exclude:   conn,
		}

	case "ping":
		response, _ := json.Marshal(&WebSocketMessage{Type: "pong"})
		conn.send <- response
	}
}
