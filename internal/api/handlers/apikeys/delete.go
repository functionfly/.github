package apikeys

import (
	"context"
	"net/http"

	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HandleDelete handles DELETE /api/v1/api-keys/:id
// This is a soft delete - sets is_active = false
func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	// Get user claims
	claims, ok := getUserClaims(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	// Get API key ID from path
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "API key ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid API key ID format")
		return
	}

	// Fetch existing API key to verify ownership
	ctx := context.Background()
	apiKey, err := h.repo.GetByID(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "not_found", "API key not found")
		return
	}

	// Verify ownership
	if apiKey.TenantID != claims.TenantID {
		h.writeError(w, http.StatusNotFound, "not_found", "API key not found")
		return
	}

	// Soft delete (revoke) - sets is_active = false
	err = h.repo.Delete(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to revoke API key")
		return
	}

	// Return success response
	h.writeSuccess(w, map[string]string{
		"message": "API key revoked successfully",
		"id":      id.String(),
	})
}

// RegisterDeleteRoutes registers the delete route
func RegisterDeleteRoutes(router *mux.Router, repo *apikey.Repository) {
	h := NewHandler(repo, nil)
	router.HandleFunc("/api-keys/{id}", h.HandleDelete).Methods("DELETE", "OPTIONS")
}
