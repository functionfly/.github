package apikeys

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HandleGetRotationHistory handles GET /api/v1/api-keys/:id/rotations
func (h *Handler) HandleGetRotationHistory(w http.ResponseWriter, r *http.Request) {
	claims, ok := getUserClaims(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

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

	ctx := r.Context()
	apiKey, err := h.repo.GetByID(ctx, id)
	if err != nil || apiKey.TenantID != claims.TenantID {
		h.writeError(w, http.StatusNotFound, "not_found", "API key not found")
		return
	}

	rotations, err := h.repo.GetRotationHistory(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch rotation history")
		return
	}

	h.writeSuccess(w, rotations)
}

// RegisterRotationHistoryRoutes registers the rotation history route.
func RegisterRotationHistoryRoutes(router *mux.Router, repo *apikey.Repository) {
	h := NewHandler(repo)
	router.HandleFunc("/api-keys/{id}/rotations", h.HandleGetRotationHistory).Methods("GET", "OPTIONS")
}
