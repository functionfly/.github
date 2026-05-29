package studio

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetMonetization handles GET /v1/marketplace/functions/{id}/monetization
func (h *MarketplaceHandler) HandleGetMonetization(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	functionID := mux.Vars(r)["id"]
	if functionID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	analysis, err := h.repo.GetMonetizationAnalysis(r.Context(), tenantID, functionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeJSONError(w, http.StatusNotFound, "Function not found")
			return
		}
		logrus.WithError(err).Error("marketplace: failed to load monetization analysis")
		writeJSONError(w, http.StatusInternalServerError, "Failed to load monetization analysis")
		return
	}

	writeJSON(w, http.StatusOK, analysis)
}

// HandleApplyMonetization handles POST /v1/marketplace/functions/{id}/monetization/apply
func (h *MarketplaceHandler) HandleApplyMonetization(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	functionID := mux.Vars(r)["id"]
	if functionID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req struct {
		Model string  `json:"model"`
		Price float64 `json:"price"`
		Trial bool    `json:"trial"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Model = strings.TrimSpace(strings.ToLower(req.Model))
	if req.Model == "" {
		writeJSONError(w, http.StatusBadRequest, "model is required")
		return
	}
	if req.Price < 0 {
		writeJSONError(w, http.StatusBadRequest, "price must be non-negative")
		return
	}

	if err := h.repo.ApplyMonetizationRecommendation(r.Context(), tenantID, functionID, req.Model, req.Price); err != nil {
		if strings.Contains(err.Error(), "not owned") || strings.Contains(err.Error(), "not found") {
			writeJSONError(w, http.StatusNotFound, "Function not found or not owned by tenant")
			return
		}
		if strings.Contains(err.Error(), "unsupported pricing model") {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		logrus.WithError(err).Error("marketplace: failed to apply monetization recommendation")
		writeJSONError(w, http.StatusInternalServerError, "Failed to apply monetization recommendation")
		return
	}

	message := "Pricing updated"
	if req.Trial {
		message = "Trial pricing applied"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": message,
		"model":   req.Model,
		"price":   req.Price,
		"trial":   req.Trial,
	})
}
