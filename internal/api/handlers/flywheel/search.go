package flywheel

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/flywheel"
)

// Search handles GET /api/v1/flywheel/search
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, `{"error":"Search query required"}`, http.StatusBadRequest)
		return
	}

	// Search in threads
	threadFilter := flywheel.ThreadFilter{
		SearchQuery: query,
		Limit:       20,
	}

	threads, _, err := h.service.ListThreads(r.Context(), threadFilter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to search threads")
		http.Error(w, `{"error":"Search failed"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"query":   query,
		"threads": threads,
		"total":   len(threads),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListVerifiedSolutions handles GET /api/v1/flywheel/solutions/verified
func (h *Handler) ListVerifiedSolutions(w http.ResponseWriter, r *http.Request) {
	// This would query for replies with verified execution status
	// For now, return an empty list
	response := map[string]interface{}{
		"solutions": []interface{}{},
		"total":     0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
