package providers

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// HandleListProviders returns the current user's connected providers (no tokens).
func (h *Handler) HandleListProviders(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	providers, err := h.repo.GetProvidersByUser(claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list providers")
		http.Error(w, "Failed to list providers", http.StatusInternalServerError)
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	providers, err := h.repo.GetProvidersByUser(claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list providers")
		http.Error(w, "Failed to list providers", http.StatusInternalServerError)
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
