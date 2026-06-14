package users

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	presenceKeyPrefix    = "presence:user:"
	presenceTTL          = 5 * time.Minute
	presenceHubBuffer    = 256
	presencePingInterval = 30 * time.Second
	presenceReadDeadline = 60 * time.Second
)

var presenceUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     middleware.IsOriginAllowedForRequest,
}

type PresenceStatus string

const (
	StatusOnline  PresenceStatus = "online"
	StatusAway    PresenceStatus = "away"
	StatusOffline PresenceStatus = "offline"
)

type UserPresence struct {
	UserID      uuid.UUID      `json:"userId"`
	Status      PresenceStatus `json:"status"`
	LastActive  time.Time      `json:"lastActive"`
	TenantID    uuid.UUID      `json:"tenantId,omitempty"`
	Username    string         `json:"username,omitempty"`
	DisplayName string         `json:"displayName,omitempty"`
	Avatar      string         `json:"avatar,omitempty"`
}

type PresenceHeartbeat struct {
	UserID     uuid.UUID `json:"userId"`
	TenantID   uuid.UUID `json:"tenantId"`
	Username   string    `json:"username,omitempty"`
	ActiveAt   time.Time `json:"activeAt"`
	LastActive time.Time `json:"lastActive"`
}

type PresenceMessage struct {
	Type      string             `json:"type"`
	UserID    string             `json:"userId,omitempty"`
	Status    PresenceStatus     `json:"status,omitempty"`
	Heartbeat *PresenceHeartbeat `json:"heartbeat,omitempty"`
	Users     []UserPresence     `json:"users,omitempty"`
	Timestamp time.Time          `json:"timestamp"`
}

type PresenceWebSocketClient struct {
	ID       string
	UserID   uuid.UUID
	TenantID uuid.UUID
	Username string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *PresenceHub
	closed   bool
	mu       sync.Mutex
}

type PresenceHub struct {
	clients    map[*PresenceWebSocketClient]bool
	register   chan *PresenceWebSocketClient
	unregister chan *PresenceWebSocketClient
	broadcast  chan *PresenceMessage
	heartbeats map[uuid.UUID]*PresenceHeartbeat
	hubMu      sync.RWMutex
	logger     *logrus.Logger
	running    bool
}

type PresenceHandler struct {
	repo        storage.Repository
	authSvc     *auth.AuthService
	redisClient *redis.Client
	hub         *PresenceHub
	logger      *logrus.Logger
}

func NewPresenceHandler(repo storage.Repository, authSvc *auth.AuthService, redisClient *redis.Client, logger *logrus.Logger) *PresenceHandler {
	if logger == nil {
		logger = logrus.New()
	}
	h := &PresenceHub{
		clients:    make(map[*PresenceWebSocketClient]bool),
		register:   make(chan *PresenceWebSocketClient),
		unregister: make(chan *PresenceWebSocketClient),
		broadcast:  make(chan *PresenceMessage, 100),
		heartbeats: make(map[uuid.UUID]*PresenceHeartbeat),
		logger:     logger,
	}
	go h.Run()
	return &PresenceHandler{
		repo:        repo,
		authSvc:     authSvc,
		redisClient: redisClient,
		hub:         h,
		logger:      logger,
	}
}

func (h *PresenceHub) Run() {
	if h.running {
		return
	}
	h.running = true

	ticker := time.NewTicker(presencePingInterval)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.hubMu.Lock()
			h.clients[client] = true
			h.hubMu.Unlock()
			h.logger.WithField("user_id", client.UserID).Debug("Presence WebSocket client registered")

		case client := <-h.unregister:
			h.hubMu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.mu.Lock()
				if !client.closed {
					client.closed = true
					close(client.Send)
				}
				client.mu.Unlock()
				h.logger.WithField("user_id", client.UserID).Debug("Presence WebSocket client unregistered")
			}
			h.hubMu.Unlock()

		case msg := <-h.broadcast:
			data, err := json.Marshal(msg)
			if err != nil {
				h.logger.WithError(err).Error("Failed to marshal presence message")
				continue
			}
			h.hubMu.RLock()
			for client := range h.clients {
				if msg.UserID != "" && client.UserID.String() == msg.UserID {
					continue
				}
				client.mu.Lock()
				if !client.closed {
					select {
					case client.Send <- data:
					default:
						close(client.Send)
						delete(h.clients, client)
					}
				}
				client.mu.Unlock()
			}
			h.hubMu.RUnlock()

		case <-ticker.C:
			h.pruneInactive()
		}
	}
}

func (h *PresenceHub) pruneInactive() {
	h.hubMu.Lock()
	defer h.hubMu.Unlock()

	now := time.Now()
	for userID, hb := range h.heartbeats {
		if now.Sub(hb.LastActive) > presenceTTL {
			delete(h.heartbeats, userID)
			msg := &PresenceMessage{
				Type:      "presence_leave",
				UserID:    userID.String(),
				Status:    StatusOffline,
				Timestamp: now,
			}
			data, _ := json.Marshal(msg)
			for client := range h.clients {
				if client.UserID != userID {
					select {
					case client.Send <- data:
					default:
					}
				}
			}
		}
	}
}

func (h *PresenceHub) UpdateHeartbeat(hb *PresenceHeartbeat) {
	h.hubMu.Lock()
	defer h.hubMu.Unlock()

	wasNew := false
	if _, exists := h.heartbeats[hb.UserID]; !exists {
		wasNew = true
	}
	h.heartbeats[hb.UserID] = hb

	msg := &PresenceMessage{
		Type:      "presence_update",
		UserID:    hb.UserID.String(),
		Status:    StatusOnline,
		Heartbeat: hb,
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(msg)

	for client := range h.clients {
		if client.UserID != hb.UserID {
			select {
			case client.Send <- data:
			default:
			}
		}
	}

	if wasNew {
		joinMsg := &PresenceMessage{
			Type:      "presence_join",
			UserID:    hb.UserID.String(),
			Status:    StatusOnline,
			Heartbeat: hb,
			Timestamp: time.Now(),
		}
		joinData, _ := json.Marshal(joinMsg)
		for client := range h.clients {
			if client.UserID != hb.UserID {
				select {
				case client.Send <- joinData:
				default:
				}
			}
		}
	}
}

func (h *PresenceHub) GetOnlineUsers() []UserPresence {
	h.hubMu.RLock()
	defer h.hubMu.RUnlock()

	users := make([]UserPresence, 0, len(h.heartbeats))
	now := time.Now()
	for userID, hb := range h.heartbeats {
		status := StatusOnline
		if now.Sub(hb.LastActive) > 2*time.Minute {
			status = StatusAway
		}
		users = append(users, UserPresence{
			UserID:     userID,
			Status:     status,
			LastActive: hb.LastActive,
			TenantID:   hb.TenantID,
			Username:   hb.Username,
		})
	}
	return users
}

func (h *PresenceHandler) HandleWebSocketPresence(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		token := r.URL.Query().Get("token")
		if token != "" {
			validatedClaims, err := h.authSvc.ValidateToken(r.Context(), token)
			if err != nil {
				tokenPrefix := token
				if len(token) > 50 {
					tokenPrefix = token[:50] + "..."
				}
				h.logger.WithError(err).WithField("token_prefix", tokenPrefix).Warn("Presence WebSocket auth failed via token")
				apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
				return
			}
			claims = validatedClaims
			h.logger.WithField("user_id", claims.UserID).Info("Presence WebSocket auth successful via token")
		}
	}
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		apierror.WriteError(w, apierror.NewNotFound("User not found"))
		return
	}

	tenantID := user.TenantID
	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	conn, err := presenceUpgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to upgrade presence WebSocket")
		return
	}

	client := &PresenceWebSocketClient{
		ID:       generatePresenceClientID(),
		UserID:   claims.UserID,
		TenantID: tenantID,
		Username: username,
		Conn:     conn,
		Send:     make(chan []byte, presenceHubBuffer),
		Hub:      h.hub,
	}

	h.hub.register <- client

	hb := &PresenceHeartbeat{
		UserID:     claims.UserID,
		TenantID:   tenantID,
		Username:   username,
		ActiveAt:   time.Now(),
		LastActive: time.Now(),
	}
	h.hub.UpdateHeartbeat(hb)

	go client.writePump()
	go client.readPump()
}

func (c *PresenceWebSocketClient) readPump() {
	defer func() {
		c.Hub.hubMu.Lock()
		c.Hub.hubMu.Unlock()
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(presenceReadDeadline))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(presenceReadDeadline))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.WithError(err).Warn("Presence WebSocket read error")
			}
			break
		}

		var msg PresenceMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "heartbeat":
			hb := &PresenceHeartbeat{
				UserID:     c.UserID,
				TenantID:   c.TenantID,
				Username:   c.Username,
				ActiveAt:   time.Now(),
				LastActive: time.Now(),
			}
			c.Hub.UpdateHeartbeat(hb)

		case "ping":
			pong := map[string]interface{}{
				"type":      "pong",
				"timestamp": time.Now(),
			}
			data, _ := json.Marshal(pong)
			c.mu.Lock()
			if !c.closed {
				c.Send <- data
			}
			c.mu.Unlock()
		}
	}
}

func (c *PresenceWebSocketClient) writePump() {
	ticker := time.NewTicker(presencePingInterval)
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

func generatePresenceClientID() string {
	return "presence_" + time.Now().Format("20060102150405") + "_" + strconv.Itoa(int(time.Now().UnixNano()%10000))
}

func (h *PresenceHandler) HandleGetPresence(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	onlineUsers := h.hub.GetOnlineUsers()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users":   onlineUsers,
		"count":   len(onlineUsers),
		"updated": time.Now(),
	})
}

func (h *PresenceHandler) HandleGetMyPresence(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	h.hub.hubMu.RLock()
	hb, exists := h.hub.heartbeats[claims.UserID]
	h.hub.hubMu.RUnlock()

	now := time.Now()
	var status PresenceStatus = StatusOffline
	var lastActive time.Time = now

	if exists {
		lastActive = hb.LastActive
		if now.Sub(lastActive) < 2*time.Minute {
			status = StatusOnline
		} else if now.Sub(lastActive) < presenceTTL {
			status = StatusAway
		}
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"userId":     user.ID,
		"status":     status,
		"lastActive": lastActive,
		"tenantId":   user.TenantID,
		"username":   username,
		"name":       user.Name,
	})
}

func (h *PresenceHandler) HandleUpdateMyPresence(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	hb := &PresenceHeartbeat{
		UserID:     claims.UserID,
		TenantID:   user.TenantID,
		Username:   username,
		ActiveAt:   time.Now(),
		LastActive: time.Now(),
	}
	h.hub.UpdateHeartbeat(hb)

	if h.redisClient != nil {
		ctx := context.Background()
		key := presenceKeyPrefix + claims.UserID.String()
		data, _ := json.Marshal(hb)
		h.redisClient.Set(ctx, key, data, presenceTTL)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"updated": time.Now(),
	})
}

func (h *PresenceHandler) HandleGetPresenceByIDs(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userIDsParam := r.URL.Query().Get("user_ids")
	if userIDsParam == "" {
		writeJSONError(w, http.StatusBadRequest, "user_ids is required")
		return
	}

	idStrs := strings.Split(userIDsParam, ",")
	userIDs := make([]uuid.UUID, 0, len(idStrs))
	for _, idStr := range idStrs {
		id, err := uuid.Parse(strings.TrimSpace(idStr))
		if err == nil {
			userIDs = append(userIDs, id)
		}
	}

	if len(userIDs) == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{"users": []UserPresence{}})
		return
	}

	now := time.Now()
	users := make([]UserPresence, 0, len(userIDs))

	h.hub.hubMu.RLock()
	for _, userID := range userIDs {
		var status PresenceStatus = StatusOffline
		var lastActive time.Time

		if hb, exists := h.hub.heartbeats[userID]; exists {
			lastActive = hb.LastActive
			if now.Sub(lastActive) < 2*time.Minute {
				status = StatusOnline
			} else if now.Sub(lastActive) < presenceTTL {
				status = StatusAway
			}
		} else {
			lastActive = now
		}

		users = append(users, UserPresence{
			UserID:     userID,
			Status:     status,
			LastActive: lastActive,
		})
	}
	h.hub.hubMu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
	})
}

func (h *PresenceHandler) HandleListOnlineUsers(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	onlineUsers := h.hub.GetOnlineUsers()

	result := make([]UserPresence, 0, len(onlineUsers))
	for _, u := range onlineUsers {
		if u.TenantID == claims.TenantID {
			user, _ := h.repo.GetUserByID(r.Context(), u.UserID)
			if user != nil {
				username := ""
				if user.Username != nil {
					username = *user.Username
				}
				u.Username = username
				u.DisplayName = user.Name
			}
			result = append(result, u)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users": result,
		"count": len(result),
	})
}

func (h *PresenceHandler) RegisterRoutes(router *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	router.HandleFunc("/users/presence", authMiddleware.RequireAuth(h.HandleGetPresence)).Methods("GET", "OPTIONS")
	router.HandleFunc("/users/presence/online", authMiddleware.RequireAuth(h.HandleListOnlineUsers)).Methods("GET", "OPTIONS")
	router.HandleFunc("/users/presence/users", authMiddleware.RequireAuth(h.HandleGetPresenceByIDs)).Methods("GET", "OPTIONS")
	router.HandleFunc("/users/presence/me", authMiddleware.RequireAuth(h.HandleGetMyPresence)).Methods("GET", "OPTIONS")
	router.HandleFunc("/users/presence/me", authMiddleware.RequireAuth(h.HandleUpdateMyPresence)).Methods("POST", "OPTIONS")
	// WebSocket endpoint uses custom auth (token query param fallback) so no RequireAuth wrapper
	router.HandleFunc("/users/presence/ws", h.HandleWebSocketPresence).Methods("GET")
}

func (h *PresenceHandler) SetRedisClient(client *redis.Client) {
	h.redisClient = client
}
