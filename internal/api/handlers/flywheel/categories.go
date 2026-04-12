package flywheel

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// ListCategories handles GET /api/v1/flywheel/categories
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.ListCategories(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list categories")
		http.Error(w, `{"error":"Failed to list categories"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories": categories,
	})
}

// SubscribeToThread handles POST /api/v1/flywheel/threads/:id/subscribe
func (h *Handler) SubscribeToThread(w http.ResponseWriter, r *http.Request) {
	user := h.getUser(w, r)
	if user == nil {
		return
	}

	vars := mux.Vars(r)
	threadID, ok := h.parseUUID(w, r, vars["id"], "thread ID")
	if !ok {
		return
	}

	var req NotificationLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.NotificationLevel = "all"
	}

	if err := h.service.SubscribeToThread(r.Context(), user.UserID, threadID, req.NotificationLevel); err != nil {
		h.logger.WithError(err).Error("Failed to subscribe to thread")
		http.Error(w, `{"error":"Failed to subscribe"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "subscribed"})
}

// UnsubscribeFromThread handles DELETE /api/v1/flywheel/threads/:id/subscribe
func (h *Handler) UnsubscribeFromThread(w http.ResponseWriter, r *http.Request) {
	user := h.getUser(w, r)
	if user == nil {
		return
	}

	vars := mux.Vars(r)
	threadID, ok := h.parseUUID(w, r, vars["id"], "thread ID")
	if !ok {
		return
	}

	if err := h.service.UnsubscribeFromThread(r.Context(), user.UserID, threadID); err != nil {
		h.logger.WithError(err).Error("Failed to unsubscribe from thread")
		http.Error(w, `{"error":"Failed to unsubscribe"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "unsubscribed"})
}
