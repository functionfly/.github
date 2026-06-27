package apikeys

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// UpdateAPIKeyRequest represents a request to update an API key
type UpdateAPIKeyRequest struct {
	Name                  *string    `json:"name,omitempty"`
	Description           *string    `json:"description,omitempty"`
	ExpiresAt             *time.Time `json:"expires_at,omitempty"`
	RateLimitRPM          *int       `json:"rate_limit_rpm,omitempty"`
	RateLimitRPH          *int       `json:"rate_limit_rph,omitempty"`
	RateLimitRPD          *int       `json:"rate_limit_rpd,omitempty"`
	IsActive              *bool      `json:"is_active,omitempty"`
	RotationFrequencyDays *int       `json:"rotation_frequency_days,omitempty"`
}

// HandleUpdate handles PATCH /api/v1/api-keys/:id
func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
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
	var req UpdateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Fetch existing API key
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

	// Apply updates
	if req.Name != nil {
		apiKey.Name = *req.Name
	}
	if req.Description != nil {
		apiKey.Description = *req.Description
	}
	if req.ExpiresAt != nil {
		apiKey.ExpiresAt = req.ExpiresAt
	}
	if req.IsActive != nil {
		apiKey.IsActive = *req.IsActive
	}
	if req.RotationFrequencyDays != nil {
		apiKey.RotationFrequencyDays = *req.RotationFrequencyDays
	}
	if req.RateLimitRPM != nil {
		apiKey.RateLimitRPM = *req.RateLimitRPM
	}
	if req.RateLimitRPH != nil {
		apiKey.RateLimitRPH = *req.RateLimitRPH
	}
	if req.RateLimitRPD != nil {
		apiKey.RateLimitRPD = *req.RateLimitRPD
	}

	// Save changes
	err = h.repo.Update(ctx, apiKey)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update API key")
		return
	}

	// Fetch updated key with associations
	apiKey, err = h.repo.GetByIDWithAssociations(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch updated API key")
		return
	}

	// Build response
	resp := apiKey.ToResponse()
	resp.Permissions = apiKey.Permissions
	resp.Environments = apiKey.Environments

	h.writeSuccess(w, resp)
}

// RegisterUpdateRoutes registers the update route
func RegisterUpdateRoutes(router *mux.Router, repo *apikey.Repository) {
	h := NewHandler(repo, nil)
	router.HandleFunc("/api-keys/{id}", h.HandleUpdate).Methods("PATCH", "OPTIONS")
}
