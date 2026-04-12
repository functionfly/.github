package flywheel

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/flywheel"
	"github.com/gorilla/mux"
)

// CreateThread handles POST /api/v1/flywheel/threads
func (h *Handler) CreateThread(w http.ResponseWriter, r *http.Request) {
	user := h.getUser(w, r)
	if user == nil {
		return
	}

	var req CreateThreadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" || req.Type == "" {
		http.Error(w, `{"error":"Title and type are required"}`, http.StatusBadRequest)
		return
	}

	// Sanitize user input to prevent XSS attacks
	sanitizedTitle, sanitizedContent := SanitizeThread(req.Title, req.Content)

	thread := &flywheel.Thread{
		Title:    sanitizedTitle,
		Type:     flywheel.ThreadType(req.Type),
		AuthorID: user.UserID,
		Tags:     req.Tags,
	}

	if req.CategoryID != nil {
		id, err := parseUUIDStr(*req.CategoryID)
		if err != nil {
			http.Error(w, `{"error":"Invalid category_id"}`, http.StatusBadRequest)
			return
		}
		thread.CategoryID = &id
	}

	if req.AttachedCapsuleID != nil {
		id, err := parseUUIDStr(*req.AttachedCapsuleID)
		if err != nil {
			http.Error(w, `{"error":"Invalid attached_capsule_id"}`, http.StatusBadRequest)
			return
		}
		thread.AttachedCapsuleID = &id
	}

	// Set problem data for problem threads
	if thread.Type == flywheel.ThreadTypeProblem {
		if len(req.ProblemData) > 0 {
			thread.ProblemData = req.ProblemData
		} else if sanitizedContent != "" {
			thread.ProblemData = json.RawMessage(`{"description":` + strconv.Quote(sanitizedContent) + `}`)
		}
		thread.EnvironmentSpecs = req.EnvironmentSpecs
		thread.ExpectedOutput = req.ExpectedOutput
	}

	if err := h.service.CreateThread(r.Context(), thread); err != nil {
		h.logger.WithError(err).Error("Failed to create thread")
		http.Error(w, `{"error":"Failed to create thread"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(thread)
}

// ListThreads handles GET /api/v1/flywheel/threads
func (h *Handler) ListThreads(w http.ResponseWriter, r *http.Request) {
	filter := flywheel.ThreadFilter{
		Type:        flywheel.ThreadType(r.URL.Query().Get("type")),
		Status:      flywheel.ThreadStatus(r.URL.Query().Get("status")),
		SearchQuery: r.URL.Query().Get("q"),
		SortBy:      r.URL.Query().Get("sort"),
	}

	if filter.SortBy == "" {
		filter.SortBy = "newest"
	}

	// Parse category_id
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		id, err := parseUUIDStr(categoryID)
		if err == nil {
			filter.CategoryID = &id
		}
	}

	// Parse pagination
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}
	if filter.Limit == 0 || filter.Limit > 100 {
		filter.Limit = 20
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	threads, count, err := h.service.ListThreads(r.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list threads")
		http.Error(w, `{"error":"Failed to list threads"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"threads": threads,
		"total":   count,
		"limit":   filter.Limit,
		"offset":  filter.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetThread handles GET /api/v1/flywheel/threads/:id
func (h *Handler) GetThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := h.parseUUID(w, r, vars["id"], "thread ID")
	if !ok {
		return
	}

	thread, err := h.service.GetThread(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"Thread not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thread)
}

// UpdateThread handles PATCH /api/v1/flywheel/threads/:id
func (h *Handler) UpdateThread(w http.ResponseWriter, r *http.Request) {
	user := h.getUser(w, r)
	if user == nil {
		return
	}

	vars := mux.Vars(r)
	id, ok := h.parseUUID(w, r, vars["id"], "thread ID")
	if !ok {
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Sanitize user-provided fields in updates
	if title, ok := updates["title"].(string); ok {
		updates["title"] = SanitizeContent(title)
	}

	callerCanModerate := flywheel.IsModeratorRole(user.Role)
	if err := h.service.UpdateThread(r.Context(), id, updates, user.UserID, callerCanModerate); err != nil {
		h.logger.WithError(err).Error("Failed to update thread")
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// ResolveThread handles POST /api/v1/flywheel/threads/:id/resolve
func (h *Handler) ResolveThread(w http.ResponseWriter, r *http.Request) {
	user := h.getUser(w, r)
	if user == nil {
		return
	}

	vars := mux.Vars(r)
	threadID, ok := h.parseUUID(w, r, vars["id"], "thread ID")
	if !ok {
		return
	}

	var req ResolveThreadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	replyID, err := parseUUIDStr(req.ReplyID)
	if err != nil {
		http.Error(w, `{"error":"Invalid reply_id"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.ResolveThread(r.Context(), threadID, replyID, user.UserID); err != nil {
		h.logger.WithError(err).Error("Failed to resolve thread")
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "resolved"})
}
