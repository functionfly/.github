package providers

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

type ProviderMaintenanceStatus struct {
	Disabled bool   `json:"disabled"`
	Reason   string `json:"reason,omitempty"`
}

// HandleListProviders returns the current user's connected providers (no tokens).
func (h *Handler) HandleListProviders(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}
	providers, err := h.repo.GetProvidersByUser(r.Context(), claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list providers")
		apierror.WriteError(w, apierror.NewInternal("Failed to list providers"))
		return
	}

	out := dedupeProviders(providers, func(p *storage.Provider) map[string]interface{} {
		return providerResponse(p)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// HandleGetProviderCredentials returns the user's connected providers with masked API keys.
// This is used for settings pages to show which providers are connected without exposing secrets.
func (h *Handler) HandleGetProviderCredentials(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}
	providers, err := h.repo.GetProvidersByUser(r.Context(), claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list providers")
		apierror.WriteError(w, apierror.NewInternal("Failed to list providers"))
		return
	}

	out := dedupeProviders(providers, func(p *storage.Provider) map[string]interface{} {
		entry := map[string]interface{}{
			"id":           p.ID,
			"name":         p.Provider,
			"status":       p.Status,
			"connectedAt":  p.CreatedAt.Format(time.RFC3339),
			"maskedApiKey": maskAPIKey(p.Token),
			"isStale":      isProviderStale(p),
		}
		if p.LastUsedAt != nil {
			entry["lastUsedAt"] = p.LastUsedAt.Format(time.RFC3339)
		}
		return entry
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// dedupeProviders returns one row per provider slug, preferring the newest when duplicates exist.
func dedupeProviders(providers []*storage.Provider, mapFn func(*storage.Provider) map[string]interface{}) []map[string]interface{} {
	sort.SliceStable(providers, func(i, j int) bool {
		return providers[i].CreatedAt.After(providers[j].CreatedAt)
	})
	seen := make(map[string]struct{}, len(providers))
	out := make([]map[string]interface{}, 0, len(providers))
	for _, p := range providers {
		if _, dup := seen[p.Provider]; dup {
			continue
		}
		seen[p.Provider] = struct{}{}
		out = append(out, mapFn(p))
	}
	return out
}

// HandleGetPlatformProviderStatus returns maintenance status for all provider types (public endpoint)
func (h *Handler) HandleGetPlatformProviderStatus(w http.ResponseWriter, r *http.Request) {
	settings, err := h.repo.ListProviderSettings(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to list provider settings")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve provider status"))
		return
	}

	status := make(map[string]ProviderMaintenanceStatus, len(settings))
	for _, s := range settings {
		status[s.Provider] = ProviderMaintenanceStatus{
			Disabled: s.Disabled,
			Reason:   "",
		}
		if s.DisabledReason != nil {
			status[s.Provider] = ProviderMaintenanceStatus{
				Disabled: s.Disabled,
				Reason:   *s.DisabledReason,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
