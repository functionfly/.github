package agent

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type chatSession struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	AgentID       string    `json:"agent_id"`
	MessageCount  int       `json:"message_count"`
	LastMessage   string    `json:"last_message,omitempty"`
	LastMessageAt time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type createSessionRequest struct {
	Title string `json:"title,omitempty"`
}

// HandleListChatSessions returns all chat sessions for an agent.
// GET /v1/agent/{agent_id}/chat/sessions?limit=50&offset=0
func (h *Handler) HandleListChatSessions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	db := h.rawDB
	if db == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "sessions": []chatSession{}})
		return
	}

	rows, err := db.QueryContext(r.Context(),
		`SELECT s.id, s.title, COALESCE(s.agent_id,''), COUNT(m.id) AS message_count,
		        MAX(m.created_at) AS last_message_at, s.created_at
		 FROM ai_chat_sessions s
		 LEFT JOIN ai_chat_messages m ON m.session_id = s.id
		 WHERE s.agent_id = $1 AND s.user_id = $2
		 GROUP BY s.id, s.title, s.agent_id, s.created_at
		 ORDER BY last_message_at DESC NULLS LAST, s.created_at DESC
		 LIMIT $3 OFFSET $4`,
		agent.AgentID, claims.UserID.String(), limit, offset,
	)
	if err != nil {
		logrus.WithError(err).Error("HandleListChatSessions: query failed")
		writeError(w, http.StatusInternalServerError, "QUERY_FAILED", "failed to list sessions")
		return
	}
	defer func() { _ = rows.Close() }()

	sessions := make([]chatSession, 0)
	for rows.Next() {
		var s chatSession
		var lastMsgAt *time.Time
		if err := rows.Scan(&s.ID, &s.Title, &s.AgentID, &s.MessageCount, &lastMsgAt, &s.CreatedAt); err != nil {
			logrus.WithError(err).Error("HandleListChatSessions: scan failed")
			continue
		}
		if lastMsgAt != nil {
			s.LastMessageAt = *lastMsgAt
		}
		sessions = append(sessions, s)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "sessions": sessions})
}

// HandleCreateChatSession creates a new chat session for an agent.
// POST /v1/agent/{agent_id}/chat/sessions
func (h *Handler) HandleCreateChatSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	var req createSessionRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	title := req.Title
	if title == "" {
		title = "New Chat"
	}

	sessionID := uuid.New()
	db := h.rawDB
	if db == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "session_id": sessionID.String()})
		return
	}

	if _, err := db.ExecContext(r.Context(),
		`INSERT INTO ai_chat_sessions (id, user_id, tenant_id, title, agent_id, context_type, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, 'agent', true, NOW(), NOW())`,
		sessionID, claims.UserID.String(), claims.TenantID.String(), title, agent.AgentID,
	); err != nil {
		logrus.WithError(err).Error("HandleCreateChatSession: insert failed")
		writeError(w, http.StatusInternalServerError, "CREATE_FAILED", "failed to create session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"session_id": sessionID.String(),
		"title":      title,
	})
}

// HandleDeleteChatSession deletes a chat session and all its messages.
// DELETE /v1/agent/{agent_id}/chat/sessions/{session_id}
func (h *Handler) HandleDeleteChatSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.getAgentByIDOrUUID(r, agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	sessionID := mux.Vars(r)["session_id"]
	db := h.rawDB
	if db == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
		return
	}

	result, err := db.ExecContext(r.Context(),
		`DELETE FROM ai_chat_sessions WHERE id = $1 AND agent_id = $2 AND user_id = $3`,
		sessionID, agent.AgentID, claims.UserID.String(),
	)
	if err != nil {
		logrus.WithError(err).Error("HandleDeleteChatSession: delete failed")
		writeError(w, http.StatusInternalServerError, "DELETE_FAILED", "failed to delete session")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "session not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}
