package flywheel

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/flywheel"
	"github.com/gorilla/mux"
)

// ListThreadAgents handles GET /api/v1/flywheel/threads/:id/agents
func (h *Handler) ListThreadAgents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, ok := h.parseUUID(w, r, vars["id"], "thread ID")
	if !ok {
		return
	}

	// Agent collaboration is not yet implemented - return empty list with 501
	h.logger.WithField("thread_id", threadID).Debug("ListThreadAgents not implemented")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"thread_id": threadID,
		"agents":    []interface{}{},
		"message":   "Agent collaboration is coming soon",
	})
}

// InviteAgent handles POST /api/v1/flywheel/threads/:id/agents/:agent_id/invite
func (h *Handler) InviteAgent(w http.ResponseWriter, r *http.Request) {
	user := h.getUser(w, r)
	if user == nil {
		return
	}

	vars := mux.Vars(r)
	threadID, ok := h.parseUUID(w, r, vars["id"], "thread ID")
	if !ok {
		return
	}

	agentID, ok := h.parseUUID(w, r, vars["agent_id"], "agent ID")
	if !ok {
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"thread_id": threadID,
		"agent_id":  agentID,
		"user_id":   user.UserID,
	}).Warn("InviteAgent not implemented")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "not_implemented",
		"message": "Agent invitation is coming soon",
	})
}

// RemoveAgent handles DELETE /api/v1/flywheel/threads/:id/agents/:agent_id
func (h *Handler) RemoveAgent(w http.ResponseWriter, r *http.Request) {
	user := h.getUser(w, r)
	if user == nil {
		return
	}

	vars := mux.Vars(r)
	threadID, ok := h.parseUUID(w, r, vars["id"], "thread ID")
	if !ok {
		return
	}

	agentID, ok := h.parseUUID(w, r, vars["agent_id"], "agent ID")
	if !ok {
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"thread_id": threadID,
		"agent_id":  agentID,
		"user_id":   user.UserID,
	}).Warn("RemoveAgent not implemented")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "not_implemented",
		"message": "Agent removal is coming soon",
	})
}

// AgentRespond handles POST /api/v1/flywheel/threads/:id/agents/:agent_id/respond
func (h *Handler) AgentRespond(w http.ResponseWriter, r *http.Request) {
	user := h.getUser(w, r)
	if user == nil {
		return
	}

	vars := mux.Vars(r)
	threadID, ok := h.parseUUID(w, r, vars["id"], "thread ID")
	if !ok {
		return
	}

	agentID, ok := h.parseUUID(w, r, vars["agent_id"], "agent ID")
	if !ok {
		return
	}

	var req AgentResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Create reply from agent
	reply := &flywheel.Reply{
		ThreadID:   threadID,
		AuthorID:   agentID,
		AuthorType: flywheel.ReplyAuthorTypeAgent,
		Content:    req.Content,
	}

	if err := h.service.CreateReply(r.Context(), reply); err != nil {
		h.logger.WithError(err).Error("Failed to create agent reply")
		http.Error(w, `{"error":"Failed to create reply"}`, http.StatusInternalServerError)
		return
	}

	// Broadcast to thread subscribers
	h.BroadcastNewReply(threadID.String(), reply)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reply)
}
