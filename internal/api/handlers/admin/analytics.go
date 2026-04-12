package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetAnalyticsSettings returns current analytics configuration
func (h *Handler) HandleGetAnalyticsSettings(w http.ResponseWriter, r *http.Request) {
	allSettings, err := h.analyticsRepo.GetAllAnalyticsSettings()
	if err != nil {
		logrus.WithError(err).Error("Failed to get analytics settings from database")
		http.Error(w, "Failed to retrieve analytics settings", http.StatusInternalServerError)
		return
	}

	settings := map[string]interface{}{
		"services": []map[string]interface{}{},
	}

	for _, s := range allSettings {
		serviceEntry := map[string]interface{}{
			"name":      s.ServiceName,
			"enabled":   s.Enabled,
			"status":    "loaded",
			"config":    s.Config,
			"last_used": nil,
		}

		switch s.ServiceName {
		case "google_analytics":
			settings["google_analytics"] = map[string]interface{}{
				"measurement_id": s.Config["measurement_id"],
				"enabled":        s.Enabled,
			}
		case "hotjar":
			settings["hotjar"] = map[string]interface{}{
				"site_id": s.Config["site_id"],
				"enabled": s.Enabled,
			}
		}

		servicesList := settings["services"].([]map[string]interface{})
		servicesList = append(servicesList, serviceEntry)
		settings["services"] = servicesList
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// HandleUpdateAnalyticsSettings updates analytics configuration
func (h *Handler) HandleUpdateAnalyticsSettings(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if googleAnalytics, ok := req["googleAnalytics"].(map[string]interface{}); ok {
		measurementID, _ := googleAnalytics["measurementId"].(string)
		enabled, _ := googleAnalytics["enabled"].(bool)
		_, err := h.analyticsRepo.UpdateAnalyticsSettings("google_analytics", enabled, map[string]interface{}{
			"measurement_id": measurementID,
		})
		if err != nil {
			logrus.WithError(err).Error("Failed to update Google Analytics settings")
			http.Error(w, "Failed to update Google Analytics settings", http.StatusInternalServerError)
			return
		}
	}

	if hotjar, ok := req["hotjar"].(map[string]interface{}); ok {
		siteID, _ := hotjar["siteId"].(string)
		enabled, _ := hotjar["enabled"].(bool)
		_, err := h.analyticsRepo.UpdateAnalyticsSettings("hotjar", enabled, map[string]interface{}{
			"site_id": siteID,
		})
		if err != nil {
			logrus.WithError(err).Error("Failed to update Hotjar settings")
			http.Error(w, "Failed to update Hotjar settings", http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"message":  "Analytics settings updated successfully",
		"settings": req,
	}

	utils.LogAuditEvent(r.Context(), h.repo, r, "analytics.settings.update", "system", nil, nil, req, true)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandlePlatformAnalyticsSummary returns platform-wide analytics summary
func (h *Handler) HandlePlatformAnalyticsSummary(w http.ResponseWriter, r *http.Request) {
	summary := map[string]interface{}{
		"platform": map[string]interface{}{
			"total_functions":       0,
			"total_executions":      0,
			"active_tenants":        0,
			"total_users":           0,
			"period":                "last_30_days",
			"generated_at":          time.Now().UTC(),
		},
	}

	// Use unified analytics if available
	if h.unifiedAnalytics != nil {
		// TODO: integrate with unified analytics for real platform summary
		logrus.Debug("Unified analytics available for platform summary")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// HandleTenantAnalyticsSummary returns analytics summary for a specific tenant
func (h *Handler) HandleTenantAnalyticsSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid tenant ID"}`, http.StatusBadRequest)
		return
	}

	// Default to last 30 days
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -30)

	summary := map[string]interface{}{
		"tenant_id": tenantID,
		"period": map[string]interface{}{
			"start": start,
			"end":   end,
		},
		"functions": map[string]interface{}{
			"executions": 0,
			"errors":     0,
		},
		"state": map[string]interface{}{
			"storage_bytes": 0,
			"read_ops":    0,
			"write_ops":   0,
		},
		"billing": map[string]interface{}{
			"quantity": 0,
			"cost_usd": 0,
		},
		"agent": map[string]interface{}{
			"calls":        0,
			"cost_usd":     0,
			"success_count": 0,
			"error_count":   0,
		},
		"generated_at": time.Now().UTC(),
	}

	// Use unified analytics if available for real data
	if h.unifiedAnalytics != nil {
		ctx := r.Context()
		tenantSummary, err := h.unifiedAnalytics.TenantSummary(ctx, tenantID, start, end)
		if err == nil && tenantSummary != nil {
			summary["functions"] = map[string]interface{}{
				"executions": tenantSummary.FunctionExecutions,
				"registry_executions": tenantSummary.RegistryExecutions,
			}
			summary["state"] = map[string]interface{}{
				"storage_bytes":   tenantSummary.StateStorageBytes,
				"read_ops":        tenantSummary.StateReadOps,
				"write_ops":       tenantSummary.StateWriteOps,
				"active_states":   tenantSummary.StateActiveStates,
			}
			summary["billing"] = map[string]interface{}{
				"quantity": tenantSummary.BillingQuantity,
			}
			summary["agent"] = map[string]interface{}{
				"calls":         tenantSummary.AgentCalls,
				"cost_usd":      tenantSummary.AgentCostUSD,
				"success_count": tenantSummary.AgentSuccessCount,
				"error_count":   tenantSummary.AgentErrorCount,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
