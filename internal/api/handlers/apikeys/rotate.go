package apikeys

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// RotateAPIKeyRequest represents a request to rotate an API key
type RotateAPIKeyRequest struct {
	Reason   string         `json:"reason,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// HandleRotate handles POST /api/v1/api-keys/:id/rotate
func (h *Handler) HandleRotate(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body (reason is optional)
	var req RotateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
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

	// Determine rotation reason
	reason := apikey.RotationReasonManual
	if req.Reason != "" {
		reason = apikey.RotationReason(req.Reason)
	}

	// Generate new API key
	plaintext, err := h.keyGen.Generate(apiKey.KeyType)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to generate new API key")
		return
	}

	// Perform rotation
	err = h.repo.Rotate(ctx, id, plaintext, reason, &claims.UserID, req.Metadata)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to rotate API key")
		return
	}

	// Fetch updated key with associations
	apiKey, err = h.repo.GetByIDWithAssociations(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch rotated API key")
		return
	}

	// Build response with new plaintext (only shown once!)
	resp := apikey.APIKeyCreateResponse{
		APIKeyResponse: *apiKey.ToResponse(),
		Plaintext:      plaintext,
	}
	resp.Permissions = apiKey.Permissions
	resp.Environments = apiKey.Environments

	h.writeSuccess(w, resp)
}

// RegisterRotateRoutes registers the rotate route
func RegisterRotateRoutes(router *mux.Router, repo *apikey.Repository) {
	h := NewHandler(repo, nil)
	router.HandleFunc("/api-keys/{id}/rotate", h.HandleRotate).Methods("POST", "OPTIONS")
}
