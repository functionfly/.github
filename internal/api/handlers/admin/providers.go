package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type ProviderSettingsResponse struct {
	Provider       string  `json:"provider"`
	Disabled       bool    `json:"disabled"`
	DisabledReason *string `json:"disabled_reason,omitempty"`
	DisabledAt     *string `json:"disabled_at,omitempty"`
	DisabledBy     *string `json:"disabled_by,omitempty"`
}

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
		// Return empty list so admin UI can load; caller can retry or check logs
		providers = nil
	}

	// Build response with user and team information
	response := make([]map[string]interface{}, 0, len(providers))

	for _, provider := range providers {
		// Get user information
		user, err := h.repo.GetUserByID(r.Context(), provider.UserID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", provider.UserID).Warn("Failed to get user for provider")
			continue
		}
		if user == nil {
			logrus.WithField("user_id", provider.UserID).Warn("User not found for provider")
			continue
		}

		// Get tenant information
		tenant, err := h.repo.GetTenantByID(r.Context(), user.TenantID)
		if err != nil {
			logrus.WithError(err).WithField("tenant_id", user.TenantID).Warn("Failed to get tenant for provider")
			continue
		}
		tenantName := ""
		if tenant != nil {
			tenantName = tenant.Name
		}

		response = append(response, map[string]interface{}{
			"id":          provider.ID,
			"user_id":     provider.UserID.String(),
			"user_email":  user.Email,
			"tenant_id":   user.TenantID.String(),
			"tenant_name": tenantName,
			"provider":    provider.Provider,
			"status":      provider.Status,
			"is_shared":   provider.IsShared,
			"team_id":     provider.TeamID,
			"created_at":  provider.CreatedAt,
			"updated_at":  provider.UpdatedAt,
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
		apierror.WriteError(w, apierror.NewBadRequest("Provider ID required"))
		return
	}

	// Parse request body
	var req struct {
		Status   *string `json:"status,omitempty"`
		IsShared *bool   `json:"is_shared,omitempty"`
		TeamID   *string `json:"team_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Verify provider exists
	provider, err := h.repo.GetProviderByID(r.Context(), providerID)
	if err != nil {
		logrus.WithError(err).WithField("provider_id", providerID).Error("Failed to get provider")
		apierror.WriteError(w, apierror.NewNotFound("Provider not found"))
		return
	}
	if provider == nil {
		apierror.WriteError(w, apierror.NewNotFound("Provider not found"))
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
		apierror.WriteError(w, apierror.NewInternal("Failed to update provider"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         updatedProvider.ID,
		"user_id":    updatedProvider.UserID.String(),
		"provider":   updatedProvider.Provider,
		"status":     updatedProvider.Status,
		"is_shared":  updatedProvider.IsShared,
		"team_id":    updatedProvider.TeamID,
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
		apierror.WriteError(w, apierror.NewBadRequest("Provider ID required"))
		return
	}

	// Verify provider exists
	provider, err := h.repo.GetProviderByID(r.Context(), providerID)
	if err != nil {
		logrus.WithError(err).WithField("provider_id", providerID).Error("Failed to get provider")
		apierror.WriteError(w, apierror.NewNotFound("Provider not found"))
		return
	}
	if provider == nil {
		apierror.WriteError(w, apierror.NewNotFound("Provider not found"))
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
		apierror.WriteError(w, apierror.NewInternal("Failed to deactivate provider"))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Provider deactivated successfully",
	})
}

// HandleListProviderSettings handles GET /admin/providers/settings - returns all provider settings
func (h *ProvidersHandler) HandleListProviderSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.repo.ListProviderSettings(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to list provider settings")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve provider settings"))
		return
	}

	response := make([]ProviderSettingsResponse, 0, len(settings))
	for _, s := range settings {
		response = append(response, ProviderSettingsResponse{
			Provider:       s.Provider,
			Disabled:       s.Disabled,
			DisabledReason: s.DisabledReason,
			DisabledAt:     nullableTimeToString(s.DisabledAt),
			DisabledBy:     s.DisabledBy,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"settings": response,
	})
}

// HandleUpdateProviderSettings handles PATCH /admin/providers/settings/{provider} - update disabled state
func (h *ProvidersHandler) HandleUpdateProviderSettings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider"]
	if provider == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Provider is required"))
		return
	}

	var req struct {
		Disabled bool   `json:"disabled"`
		Reason   string `json:"reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	disabledBy := ""
	if claims := middleware.GetUserFromContext(r); claims != nil {
		disabledBy = claims.UserID.String()
	}

	err := h.repo.SetProviderDisabled(r.Context(), provider, req.Disabled, req.Reason, disabledBy)
	if err != nil {
		logrus.WithError(err).WithField("provider", provider).Error("Failed to update provider settings")
		apierror.WriteError(w, apierror.NewInternal("Failed to update provider settings"))
		return
	}

	updatedSettings, err := h.repo.GetProviderSettings(r.Context(), provider)
	if err != nil {
		logrus.WithError(err).WithField("provider", provider).Error("Failed to get updated provider settings")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve updated provider settings"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"settings": ProviderSettingsResponse{
			Provider:       updatedSettings.Provider,
			Disabled:       updatedSettings.Disabled,
			DisabledReason: updatedSettings.DisabledReason,
			DisabledAt:     nullableTimeToString(updatedSettings.DisabledAt),
			DisabledBy:     updatedSettings.DisabledBy,
		},
	})
}

func nullableTimeToString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}
