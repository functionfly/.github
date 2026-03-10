package apikeys

import (
	"context"
	"net/http"

	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HandleGet handles GET /api/v1/api-keys/:id
func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
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

	// Fetch API key with associations (permissions and environments)
	ctx := context.Background()
	apiKey, err := h.repo.GetByIDWithAssociations(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "not_found", "API key not found")
		return
	}

	// Verify ownership (tenant must match)
	if apiKey.TenantID != claims.TenantID {
		h.writeError(w, http.StatusNotFound, "not_found", "API key not found")
		return
	}

	// Build response (without plaintext)
	resp := apiKey.ToResponse()
	resp.Permissions = apiKey.Permissions
	resp.Environments = apiKey.Environments

	h.writeSuccess(w, resp)
}

// RegisterGetRoutes registers the get route
func RegisterGetRoutes(router *mux.Router, repo *apikey.Repository) {
	h := NewHandler(repo)
	router.HandleFunc("/api-keys/{id}", h.HandleGet).Methods("GET", "OPTIONS")
}
