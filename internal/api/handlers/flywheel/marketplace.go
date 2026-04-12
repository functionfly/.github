package flywheel

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// PublishToMarketplace handles POST /api/v1/flywheel/replies/:id/publish-to-marketplace
func (h *Handler) PublishToMarketplace(w http.ResponseWriter, r *http.Request) {
	user := h.getUser(w, r)
	if user == nil {
		return
	}

	vars := mux.Vars(r)
	replyID, ok := h.parseUUID(w, r, vars["id"], "reply ID")
	if !ok {
		return
	}

	// Get the reply
	reply, err := h.service.GetReply(r.Context(), replyID)
	if err != nil {
		http.Error(w, `{"error":"Reply not found"}`, http.StatusNotFound)
		return
	}

	// Verify the user owns this reply or has permission
	if reply.AuthorID != user.UserID {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusForbidden)
		return
	}

	var req MarketplacePublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Marketplace publishing is not yet implemented
	h.logger.WithFields(map[string]interface{}{
		"reply_id":   replyID,
		"user_id":    user.UserID,
		"name":       req.Name,
		"visibility": req.Visibility,
	}).Warn("PublishToMarketplace not implemented")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "not_implemented",
		"message": "Marketplace publishing is coming soon",
	})
}
