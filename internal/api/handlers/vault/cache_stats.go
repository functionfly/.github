package vault

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage/vault"
)

// HandleCacheStats handles GET /v1/vault/cache/stats.
//
// Returns the current count of cached metadata + token entries
// across all tenants. Operators use this to size Redis and to
// verify cache invalidation is working.
func (h *Handler) HandleCacheStats(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	if h.Cache == nil {
		apierror.WriteError(w, apierror.NewInternal("Cache not configured"))
		return
	}
	var stats vault.CacheStats
	if h.Cache.Enabled() {
		stats = h.Cache.Stats(r.Context())
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":  h.Cache.Enabled(),
		"meta":     stats.MetaKeys,
		"tokens":   stats.TokenKeys,
		"resource": "vault:cache",
	})
}

// _ keeps the import alive in case we later move the helper.
