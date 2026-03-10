package apikeys

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// AddEnvironmentRequest represents a request to link an environment
type AddEnvironmentRequest struct {
	EnvironmentID   uuid.UUID `json:"environment_id"`
	EnvironmentName string    `json:"environment_name,omitempty"`
}

// HandleListEnvironments handles GET /api/v1/api-keys/:id/environments
func (h *Handler) HandleListEnvironments(w http.ResponseWriter, r *http.Request) {
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

	// Verify ownership
	ctx := context.Background()
	apiKey, err := h.repo.GetByID(ctx, id)
	if err != nil || apiKey.TenantID != claims.TenantID {
		h.writeError(w, http.StatusNotFound, "not_found", "API key not found")
		return
	}

	// Fetch environments
	environments, err := h.repo.GetEnvironments(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch environments")
		return
	}

	h.writeSuccess(w, environments)
}

// HandleAddEnvironment handles POST /api/v1/api-keys/:id/environments
func (h *Handler) HandleAddEnvironment(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var req AddEnvironmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate request
	if req.EnvironmentID == uuid.Nil {
		h.writeError(w, http.StatusBadRequest, "validation_error", "environment_id is required")
		return
	}

	// Verify ownership
	ctx := context.Background()
	apiKey, err := h.repo.GetByID(ctx, id)
	if err != nil || apiKey.TenantID != claims.TenantID {
		h.writeError(w, http.StatusNotFound, "not_found", "API key not found")
		return
	}

	// Link environment
	err = h.repo.LinkEnvironment(ctx, id, req.EnvironmentID, req.EnvironmentName)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to link environment")
		return
	}

	// Fetch updated environments
	environments, err := h.repo.GetEnvironments(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch environments")
		return
	}

	h.writeSuccess(w, environments)
}

// HandleRemoveEnvironment handles DELETE /api/v1/api-keys/:id/environments/:env_id
func (h *Handler) HandleRemoveEnvironment(w http.ResponseWriter, r *http.Request) {
	// Get user claims
	claims, ok := getUserClaims(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	// Get API key ID and environment ID from path
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "API key ID is required")
		return
	}

	envIDStr, ok := vars["env_id"]
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Environment ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid API key ID format")
		return
	}

	envID, err := uuid.Parse(envIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid environment ID format")
		return
	}

	// Verify ownership
	ctx := context.Background()
	apiKey, err := h.repo.GetByID(ctx, id)
	if err != nil || apiKey.TenantID != claims.TenantID {
		h.writeError(w, http.StatusNotFound, "not_found", "API key not found")
		return
	}

	// Unlink environment
	err = h.repo.UnlinkEnvironment(ctx, id, envID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to unlink environment")
		return
	}

	// Fetch updated environments
	environments, err := h.repo.GetEnvironments(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch environments")
		return
	}

	h.writeSuccess(w, environments)
}

// RegisterEnvironmentRoutes registers the environment routes
func RegisterEnvironmentRoutes(router *mux.Router, repo *apikey.Repository) {
	h := NewHandler(repo)
	router.HandleFunc("/api-keys/{id}/environments", h.HandleListEnvironments).Methods("GET", "OPTIONS")
	router.HandleFunc("/api-keys/{id}/environments", h.HandleAddEnvironment).Methods("POST", "OPTIONS")
	router.HandleFunc("/api-keys/{id}/environments/{env_id}", h.HandleRemoveEnvironment).Methods("DELETE", "OPTIONS")
}
