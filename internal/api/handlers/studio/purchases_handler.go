package studio

import (
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/marketplaceconfig"
	"github.com/sirupsen/logrus"
)

// HandleListMyPurchases handles GET /v1/marketplace/purchases (buyer-scoped).
func (h *MarketplaceHandler) HandleListMyPurchases(w http.ResponseWriter, r *http.Request) {
	if !marketplaceconfig.PurchasesEnabled() {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"enabled": false,
			"functions":     []FunctionPurchaseItem{},
			"agents":        []AgentHiringItem{},
			"licenses":      []BuyerLicenseItem{},
			"subscriptions": []BuyerSubscriptionItem{},
		})
		return
	}

	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	summary, err := h.repo.ListMyPurchases(r.Context(), tenantID, getUserID(r), limit, offset)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to list purchases")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list purchases")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":            true,
		"functions":          summary.Functions,
		"agents":             summary.Agents,
		"licenses":           summary.Licenses,
		"subscriptions":      summary.Subscriptions,
		"totalFunctions":     summary.TotalFunctions,
		"totalAgents":        summary.TotalAgents,
		"totalLicenses":      summary.TotalLicenses,
		"totalSubscriptions": summary.TotalSubscriptions,
		"limit":              limit,
		"offset":             offset,
	})
}
