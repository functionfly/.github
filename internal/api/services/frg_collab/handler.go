package frg_collab

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	svc     *FRGCollaborationService
	authSvc *auth.AuthService
	logger  *logrus.Logger
}

func NewHandler(svc *FRGCollaborationService, authSvc *auth.AuthService, logger *logrus.Logger) *Handler {
	return &Handler{
		svc:     svc,
		authSvc: authSvc,
		logger:  logger,
	}
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	graphIDStr := vars["graphId"]

	graphID, err := uuid.Parse(graphIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid graph ID"))
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		apierror.WriteError(w, apierror.NewUnauthorized("Missing auth token"))
		return
	}

	claims, err := h.authSvc.ValidateToken(r.Context(), token)
	if err != nil {
		h.logger.WithError(err).Warn("FRG collab WebSocket auth failed via token")
		apierror.WriteError(w, apierror.NewUnauthorized("Invalid or expired token"))
		return
	}

	session, err := h.svc.GetOrCreateSession(r.Context(), graphID, claims.TenantID, claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get or create FRG session")
		apierror.WriteError(w, apierror.NewInternal("Failed to join collaboration session"))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to upgrade FRG WebSocket")
		return
	}

	displayName := claims.Email
	if claims.Username != "" {
		displayName = claims.Username
	}

	_, participant, err := h.svc.JoinSession(r.Context(), graphID, claims.UserID, displayName)
	if err != nil {
		h.logger.WithError(err).Error("Failed to join FRG session")
		_ = conn.Close()
		return
	}

	client := &FRGWSClient{
		GraphID:   graphID.String(),
		UserID:    claims.UserID,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		Hub:       h.svc.WebSocketHub(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	h.svc.WebSocketHub().Register(client)
	go h.writePump(client)
	go h.readPump(client, session, participant)

	h.sendSessionState(client, session, participant)
}

func (h *Handler) readPump(client *FRGWSClient, session *GraphSession, participant *Participant) {
	defer func() {
		h.svc.WebSocketHub().Unregister(client)
		if graphID, err := uuid.Parse(client.GraphID); err == nil {
			_ = h.svc.LeaveSession(context.Background(), graphID, client.UserID)
		}
		_ = client.Conn.Close()
	}()

	_ = client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		_ = client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.WithError(err).Debug("FRG collab WS read error")
			}
			break
		}

		var msg struct {
			Type         string  `json:"type"`
			X            float64 `json:"x"`
			Y            float64 `json:"y"`
			ViewportX    float64 `json:"viewport_x"`
			ViewportY    float64 `json:"viewport_y"`
			ViewportZoom float64 `json:"viewport_zoom"`
			SelectedNode string  `json:"selected_node"`
			NodeID       string  `json:"node_id"`
			Selected     bool    `json:"selected"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "cursor":
			cursor := &CursorPosition{
				X:            msg.X,
				Y:            msg.Y,
				ViewportX:     msg.ViewportX,
				ViewportY:     msg.ViewportY,
				ViewportZoom:  msg.ViewportZoom,
				SelectedNode:  msg.SelectedNode,
			}
			_ = h.svc.UpdateCursor(context.Background(), client.GraphID, client.UserID, cursor)

		case "viewport":
			_ = h.svc.UpdateViewport(context.Background(), client.GraphID, client.UserID, msg.ViewportX, msg.ViewportY, msg.ViewportZoom)

		case "node_selection":
			_ = h.svc.UpdateNodeSelection(context.Background(), client.GraphID, client.UserID, msg.NodeID, msg.Selected)
		}
	}
}

func (h *Handler) writePump(client *FRGWSClient) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		_ = client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Handler) sendSessionState(client *FRGWSClient, session *GraphSession, participant *Participant) {
	participants := make([]Participant, 0)
	for _, p := range session.Participants {
		if p.UserID != client.UserID {
			participants = append(participants, p)
		}
	}

	state := map[string]interface{}{
		"type":         "session_state",
		"graph_id":     session.GraphID.String(),
		"participants": participants,
		"you":          participant,
		"timestamp":    time.Now().UnixMilli(),
	}
	data, _ := json.Marshal(state)
	client.Send <- data
}
