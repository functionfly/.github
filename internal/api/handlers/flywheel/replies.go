package flywheel

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/flywheel"
	"github.com/gorilla/mux"
)

// CreateReply handles POST /api/v1/flywheel/threads/:id/replies
func (h *Handler) CreateReply(w http.ResponseWriter, r *http.Request) {
	user := h.getUser(w, r)
	if user == nil {
		return
	}

	vars := mux.Vars(r)
	threadID, ok := h.parseUUID(w, r, vars["id"], "thread ID")
	if !ok {
		return
	}

	var req CreateReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Content == "" && len(req.CodeBlocks) == 0 {
		http.Error(w, `{"error":"Content or code_blocks required"}`, http.StatusBadRequest)
		return
	}

	// Sanitize user content to prevent XSS attacks
	sanitizedContent := SanitizeReply(req.Content)

	reply := &flywheel.Reply{
		ThreadID: threadID,
		AuthorID: user.UserID,
		Content:  sanitizedContent,
	}

	if req.ParentReplyID != nil {
		id, err := parseUUIDStr(*req.ParentReplyID)
		if err != nil {
			http.Error(w, `{"error":"Invalid parent_reply_id"}`, http.StatusBadRequest)
			return
		}
		reply.ParentReplyID = &id
	}

	if len(req.CodeBlocks) > 0 {
		reply.CodeBlocks = req.CodeBlocks
	}

	if req.AttachedCapsuleID != nil {
		id, err := parseUUIDStr(*req.AttachedCapsuleID)
		if err != nil {
			http.Error(w, `{"error":"Invalid attached_capsule_id"}`, http.StatusBadRequest)
			return
		}
		reply.AttachedCapsuleID = &id
	}

	if err := h.service.CreateReply(r.Context(), reply); err != nil {
		h.logger.WithError(err).Error("Failed to create reply")
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reply)
}

// ListReplies handles GET /api/v1/flywheel/threads/:id/replies
func (h *Handler) ListReplies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, ok := h.parseUUID(w, r, vars["id"], "thread ID")
	if !ok {
		return
	}

	filter := flywheel.ReplyFilter{}
	if r.URL.Query().Get("parent_only") == "true" {
		filter.ParentOnly = true
	}

	replies, count, err := h.service.ListReplies(r.Context(), threadID, filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list replies")
		http.Error(w, `{"error":"Failed to list replies"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"replies": replies,
		"total":   count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExecuteReply handles POST /api/v1/flywheel/replies/:id/execute
func (h *Handler) ExecuteReply(w http.ResponseWriter, r *http.Request) {
	user := h.getUser(w, r)
	if user == nil {
		return
	}

	vars := mux.Vars(r)
	replyID, ok := h.parseUUID(w, r, vars["id"], "reply ID")
	if !ok {
		return
	}

	var req ExecuteReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Input = nil
	}

	execution, err := h.service.ExecuteReply(r.Context(), replyID, user.UserID, req.Input)
	if err != nil {
		h.logger.WithError(err).Error("Failed to execute reply")
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(execution)
}

// VerifyReply handles POST /api/v1/flywheel/replies/:id/verify
func (h *Handler) VerifyReply(w http.ResponseWriter, r *http.Request) {
	user := h.getUser(w, r)
	if user == nil {
		return
	}

	vars := mux.Vars(r)
	replyID, ok := h.parseUUID(w, r, vars["id"], "reply ID")
	if !ok {
		return
	}

	if err := h.service.VerifyReply(r.Context(), replyID); err != nil {
		h.logger.WithError(err).Error("Failed to verify reply")
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "verified"})
}
