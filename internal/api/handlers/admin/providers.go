package admin

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ProvidersHandler handles admin provider management endpoints
type ProvidersHandler struct {
	repo    storage.Repository
	authSvc *auth.AuthService
}

// NewProvidersHandler creates a new admin providers handler
func NewProvidersHandler(repo storage.Repository, authSvc *auth.AuthService) *ProvidersHandler {
	return &ProvidersHandler{
		repo:    repo,
		authSvc: authSvc,
	}
}

// HandleListProviders handles GET /admin/providers - lists all providers across all users
func (h *ProvidersHandler) HandleListProviders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get all providers
	providers, err := h.repo.ListAllProviders(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to list all providers")
		http.Error(w, "Failed to list providers", http.StatusInternalServerError)
		return
	}

	// Build response with user and team information
	response := make([]map[string]interface{}, 0, len(providers))

	for _, provider := range providers {
		// Get user information
		user, err := h.repo.GetUserByID(provider.UserID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", provider.UserID).Warn("Failed to get user for provider")
			continue
		}

		// Get tenant information
		tenant, err := h.repo.GetTenantByID(user.TenantID)
		if err != nil {
			logrus.WithError(err).WithField("tenant_id", user.TenantID).Warn("Failed to get tenant for provider")
			continue
		}

		response = append(response, map[string]interface{}{
			"id":        provider.ID,
			"user_id":   provider.UserID.String(),
			"user_email": user.Email,
			"tenant_id": user.TenantID.String(),
			"tenant_name": tenant.Name,
			"provider":  provider.Provider,
			"status":    provider.Status,
			"is_shared": provider.IsShared,
			"team_id":   provider.TeamID,
			"created_at": provider.CreatedAt,
			"updated_at": provider.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"providers": response,
	})
}

// HandleUpdateProvider handles PATCH /admin/providers/:providerId - updates provider status
func (h *ProvidersHandler) HandleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse provider ID from path
	vars := mux.Vars(r)
	providerID := vars["providerId"]
	if providerID == "" {
		http.Error(w, "Provider ID required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Status   *string `json:"status,omitempty"`
		IsShared *bool   `json:"is_shared,omitempty"`
		TeamID   *string `json:"team_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify provider exists
	provider, err := h.repo.GetProviderByID(providerID)
	if err != nil {
		logrus.WithError(err).WithField("provider_id", providerID).Error("Failed to get provider")
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}
	if provider == nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.IsShared != nil {
		updates["is_shared"] = *req.IsShared
	}
	if req.TeamID != nil {
		updates["team_id"] = *req.TeamID
	}

	// Update provider
	updatedProvider, err := h.repo.UpdateProvider(ctx, providerID, updates)
	if err != nil {
		logrus.WithError(err).WithField("provider_id", providerID).Error("Failed to update provider")
		http.Error(w, "Failed to update provider", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        updatedProvider.ID,
		"user_id":   updatedProvider.UserID.String(),
		"provider":  updatedProvider.Provider,
		"status":    updatedProvider.Status,
		"is_shared": updatedProvider.IsShared,
		"team_id":   updatedProvider.TeamID,
		"created_at": updatedProvider.CreatedAt,
		"updated_at": updatedProvider.UpdatedAt,
	})
}

// HandleDeleteProvider handles DELETE /admin/providers/:providerId - deletes a provider
func (h *ProvidersHandler) HandleDeleteProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse provider ID from path
	vars := mux.Vars(r)
	providerID := vars["providerId"]
	if providerID == "" {
		http.Error(w, "Provider ID required", http.StatusBadRequest)
		return
	}

	// Verify provider exists
	provider, err := h.repo.GetProviderByID(providerID)
	if err != nil {
		logrus.WithError(err).WithField("provider_id", providerID).Error("Failed to get provider")
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}
	if provider == nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	// For now, we'll just deactivate the provider instead of deleting it
	// This preserves data integrity and allows reactivation if needed
	updates := map[string]interface{}{
		"status": "inactive",
	}

	_, err = h.repo.UpdateProvider(ctx, providerID, updates)
	if err != nil {
		logrus.WithError(err).WithField("provider_id", providerID).Error("Failed to deactivate provider")
		http.Error(w, "Failed to deactivate provider", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Provider deactivated successfully",
	})
}