package certification

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// ListTiers handles GET /v1/certification/tiers
// Returns all active certification tiers (public, no auth required)
func (h *Handler) ListTiers(w http.ResponseWriter, r *http.Request) {
	tiers, err := h.repo.ListTiers(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to list cert tiers")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve certification tiers")
		return
	}

	// Transform to response format
	result := make([]map[string]interface{}, 0, len(tiers))
	for _, t := range tiers {
		result = append(result, map[string]interface{}{
			"id":                 t.ID,
			"slug":               t.Slug,
			"name":               t.Name,
			"description":        t.Description,
			"icon":               t.Icon,
			"color":              t.Color,
			"sort_order":         t.SortOrder,
			"price_cents":        t.PriceCents,
			"currency":           t.Currency,
			"pass_threshold":     t.PassThreshold,
			"time_limit_minutes": t.TimeLimitMinutes,
			"question_count":     t.QuestionCount,
			"practical_count":    t.PracticalCount,
			"validity_months":    t.ValidityMonths,
			"is_coming_soon":     t.IsComingSoon,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tiers": result,
	})
}
