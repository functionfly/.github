package flywheel

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// GetThreadTimeline handles GET /api/v1/flywheel/threads/:id/timeline
func (h *Handler) GetThreadTimeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, ok := h.parseUUID(w, r, vars["id"], "thread ID")
	if !ok {
		return
	}

	// Get thread with executions for replay
	thread, err := h.service.GetThread(r.Context(), threadID)
	if err != nil {
		http.Error(w, `{"error":"Thread not found"}`, http.StatusNotFound)
		return
	}

	// Build timeline from thread history
	timeline := []map[string]interface{}{
		{
			"timestamp": thread.CreatedAt,
			"type":      "thread_created",
			"data":      thread,
		},
	}

	response := map[string]interface{}{
		"thread_id": threadID,
		"timeline":  timeline,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ReplayThread handles POST /api/v1/flywheel/threads/:id/replay
func (h *Handler) ReplayThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, ok := h.parseUUID(w, r, vars["id"], "thread ID")
	if !ok {
		return
	}

	// Verify thread exists
	_, err := h.service.GetThread(r.Context(), threadID)
	if err != nil {
		http.Error(w, `{"error":"Thread not found"}`, http.StatusNotFound)
		return
	}

	// Thread replay is not yet implemented
	h.logger.WithField("thread_id", threadID).Warn("Thread replay requested but not implemented")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "not_implemented",
		"message": "Thread replay functionality is coming soon",
	})
}
