package collaboration

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type AuthService interface {
	GetUserFromRequest(r *http.Request) (*storage.User, error)
	GetTenantByUser(userID string) (string, string, error)
}

type PlanService interface {
	GetTenantPlan(tenantID string) (string, error)
}

type Handler struct {
	svc     *CollaborationService
	authSvc AuthService
	planSvc PlanService
	logger  *logrus.Logger
}

func NewHandler(svc *CollaborationService, authSvc AuthService, planSvc PlanService, logger *logrus.Logger) *Handler {
	return &Handler{
		svc:     svc,
		authSvc: authSvc,
		planSvc: planSvc,
		logger:  logger,
	}
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionKey := vars["sessionKey"]

	user, err := h.authSvc.GetUserFromRequest(r)
	if err != nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	tenantID, _, err := h.authSvc.GetTenantByUser(user.ID.String())
	if err != nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	plan, err := h.planSvc.GetTenantPlan(tenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get plan"))
		return
	}

	if !plans.HasFeature(plan, plans.FeatureCollaborativeSessions) {
		apierror.WriteError(w, apierror.NewForbidden("Collaborative sessions require Enterprise plan"))
		return
	}

	session, err := h.svc.GetSession(r.Context(), sessionKey)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Session not found"))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to upgrade WebSocket")
		return
	}

	displayName := user.Email
	if user.Name != "" {
		displayName = user.Name
	}

	_, participant, err := h.svc.JoinSession(r.Context(), sessionKey, user.ID, displayName)
	if err != nil {
		h.logger.WithError(err).Error("Failed to join session")
		conn.Close()
		return
	}

	client := &WSClient{
		SessionKey: sessionKey,
		UserID:     user.ID,
		Conn:       conn,
		Send:       make(chan []byte, 256),
		Hub:        h.svc.WebSocketHub(),
	}

	h.svc.WebSocketHub().Register(client)
	go h.writePump(client)
	go h.readPump(client, participant)

	h.sendSessionState(client, session, participant)
}

func (h *Handler) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	user, err := h.authSvc.GetUserFromRequest(r)
	if err != nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	tenantID, _, err := h.authSvc.GetTenantByUser(user.ID.String())
	if err != nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	plan, err := h.planSvc.GetTenantPlan(tenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get plan"))
		return
	}

	if !plans.HasFeature(plan, plans.FeatureCollaborativeSessions) {
		apierror.WriteError(w, apierror.NewForbidden("Collaborative sessions require Enterprise plan"))
		return
	}

	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionID"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	var req struct {
		InitialInput json.RawMessage `json:"initial_input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.InitialInput = json.RawMessage("{}")
	}

	session, err := h.svc.CreateSession(r.Context(), functionID, user.ID, req.InitialInput)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to create session"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_key": session.SessionKey,
		"session_id":  session.ID.String(),
	})
}

func (h *Handler) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionKey := vars["sessionKey"]

	session, err := h.svc.GetSession(r.Context(), sessionKey)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Session not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (h *Handler) HandleLeaveSession(w http.ResponseWriter, r *http.Request) {
	user, err := h.authSvc.GetUserFromRequest(r)
	if err != nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	sessionKey := vars["sessionKey"]

	if err := h.svc.LeaveSession(r.Context(), sessionKey, user.ID); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to leave session"))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) sendSessionState(client *WSClient, session *Session, participant *Participant) {
	state := map[string]interface{}{
		"type":         "session_state",
		"session":      session,
		"participant":  participant,
		"timestamp":    time.Now().UnixMilli(),
	}
	data, _ := json.Marshal(state)
	client.Send <- data
}

func (h *Handler) readPump(client *WSClient, participant *Participant) {
	defer func() {
		h.svc.WebSocketHub().Unregister(client)
		h.svc.LeaveSession(nil, client.SessionKey, client.UserID)
		client.Conn.Close()
	}()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.WithError(err).Error("Collaboration WS read error")
			}
			break
		}

		var msg struct {
			Type     string          `json:"type"`
			Path     string          `json:"path"`
			Value    json.RawMessage `json:"value,omitempty"`
			Position int             `json:"position"`
			Start    int             `json:"start"`
			End      int             `json:"end"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "operation":
			op := &Operation{
				Type:   "set",
				UserID: client.UserID,
				Path:   msg.Path,
				Value:  msg.Value,
			}
			h.svc.ApplyOperation(nil, client.SessionKey, op)

		case "cursor":
			h.svc.UpdateCursor(nil, client.SessionKey, client.UserID, msg.Position, msg.Start, msg.End)
		}
	}
}

func (h *Handler) writePump(client *WSClient) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}