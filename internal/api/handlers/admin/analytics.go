package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/utils"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetAnalyticsSettings returns current analytics configuration
func (h *Handler) HandleGetAnalyticsSettings(w http.ResponseWriter, r *http.Request) {
	allSettings, err := h.analyticsRepo.GetAllAnalyticsSettings()
	if err != nil {
		logrus.WithError(err).Error("Failed to get analytics settings from database")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve analytics settings"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
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
			apierror.WriteError(w, apierror.NewInternal("Failed to update Google Analytics settings"))
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
			apierror.WriteError(w, apierror.NewInternal("Failed to update Hotjar settings"))
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
	// Default to last 30 days
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -30)

	summary := map[string]interface{}{
		"platform": map[string]interface{}{
			"total_functions":       0,
			"total_executions":      0,
			"active_tenants":        0,
			"total_users":           0,
			"total_state_read_ops":  0,
			"total_state_write_ops": 0,
			"total_agent_calls":     0,
			"total_registry_execs":  0,
			"period":                "last_30_days",
			"period_start":          start,
			"period_end":            end,
			"generated_at":          time.Now().UTC(),
		},
	}

	// Use unified analytics if available for real data
	if h.unifiedAnalytics != nil {
		ctx := r.Context()
		platformSummary, err := h.unifiedAnalytics.PlatformSummary(ctx, start, end)
		if err == nil && platformSummary != nil {
			summary["platform"] = map[string]interface{}{
				"active_tenants":        platformSummary.TotalTenantsActive,
				"total_executions":      platformSummary.TotalFunctionExecs,
				"total_state_read_ops":  platformSummary.TotalStateReadOps,
				"total_state_write_ops": platformSummary.TotalStateWriteOps,
				"total_agent_calls":     platformSummary.TotalAgentCalls,
				"total_registry_execs":  platformSummary.TotalRegistryExecs,
				"period":                "last_30_days",
				"period_start":          platformSummary.Start,
				"period_end":            platformSummary.End,
				"generated_at":          platformSummary.GeneratedAt,
			}
			logrus.WithFields(logrus.Fields{
				"active_tenants":   platformSummary.TotalTenantsActive,
				"total_executions": platformSummary.TotalFunctionExecs,
			}).Debug("Platform analytics summary generated from unified analytics")
		} else if err != nil {
			logrus.WithError(err).Warn("Failed to get platform summary from unified analytics, returning defaults")
		}
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
		apierror.WriteError(w, apierror.NewBadRequest("invalid tenant ID"))
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
			"read_ops":      0,
			"write_ops":     0,
		},
		"billing": map[string]interface{}{
			"quantity": 0,
			"cost_usd": 0,
		},
		"agent": map[string]interface{}{
			"calls":         0,
			"cost_usd":      0,
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
				"executions":          tenantSummary.FunctionExecutions,
				"registry_executions": tenantSummary.RegistryExecutions,
			}
			summary["state"] = map[string]interface{}{
				"storage_bytes": tenantSummary.StateStorageBytes,
				"read_ops":      tenantSummary.StateReadOps,
				"write_ops":     tenantSummary.StateWriteOps,
				"active_states": tenantSummary.StateActiveStates,
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

// HandleAnalyticsMRR returns MRR (Monthly Recurring Revenue) metrics
// GET /v1/admin/analytics/mrr?year=2024&month=4
func (h *Handler) HandleAnalyticsMRR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse year and month from query params, default to current month
	year := time.Now().Year()
	month := int(time.Now().Month())

	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 12 {
			month = parsed
		}
	}

	mrr, err := h.analyticsRepo.CalculateMRR(ctx, year, month)
	if err != nil {
		logrus.WithError(err).Error("Failed to calculate MRR")
		apierror.WriteError(w, apierror.NewInternal("Failed to calculate MRR"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":    mrr,
		"success": true,
	})
}

// HandleAnalyticsMRRSeries returns MRR timeseries data for charts
// GET /v1/admin/analytics/mrr-series?months=12
func (h *Handler) HandleAnalyticsMRRSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	months := 12
	if m := r.URL.Query().Get("months"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed > 0 && parsed <= 24 {
			months = parsed
		}
	}

	series, err := h.analyticsRepo.GetMRRTimeseries(ctx, months)
	if err != nil {
		logrus.WithError(err).Error("Failed to get MRR series")
		apierror.WriteError(w, apierror.NewInternal("Failed to get MRR series"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":    series,
		"success": true,
	})
}

// HandleAnalyticsARR returns ARR (Annual Recurring Revenue) metrics
// GET /v1/admin/analytics/arr?year=2024&month=4
func (h *Handler) HandleAnalyticsARR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	year := time.Now().Year()
	month := int(time.Now().Month())

	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 12 {
			month = parsed
		}
	}

	arr, err := h.analyticsRepo.CalculateARR(ctx, year, month)
	if err != nil {
		logrus.WithError(err).Error("Failed to calculate ARR")
		apierror.WriteError(w, apierror.NewInternal("Failed to calculate ARR"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":    arr,
		"success": true,
	})
}

// HandleAnalyticsChurn returns churn metrics for a specific period
// GET /v1/admin/analytics/churn?year=2024&month=4
func (h *Handler) HandleAnalyticsChurn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	year := time.Now().Year()
	month := int(time.Now().Month())

	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 12 {
			month = parsed
		}
	}

	churn, err := h.analyticsRepo.CalculateChurnMetrics(ctx, year, month)
	if err != nil {
		logrus.WithError(err).Error("Failed to calculate churn metrics")
		apierror.WriteError(w, apierror.NewInternal("Failed to calculate churn metrics"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":    churn,
		"success": true,
	})
}

// HandleAnalyticsChurnSeries returns churn metrics timeseries
// GET /v1/admin/analytics/churn-series?months=12
func (h *Handler) HandleAnalyticsChurnSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	months := 12
	if m := r.URL.Query().Get("months"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed > 0 && parsed <= 24 {
			months = parsed
		}
	}

	series, err := h.analyticsRepo.GetChurnMetricsTimeseries(ctx, months)
	if err != nil {
		logrus.WithError(err).Error("Failed to get churn series")
		apierror.WriteError(w, apierror.NewInternal("Failed to get churn series"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":    series,
		"success": true,
	})
}

// HandleAnalyticsLTV returns Lifetime Value metrics
// GET /v1/admin/analytics/ltv
func (h *Handler) HandleAnalyticsLTV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ltv, err := h.analyticsRepo.GetLTVMetrics(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to calculate LTV")
		apierror.WriteError(w, apierror.NewInternal("Failed to calculate LTV"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":    ltv,
		"success": true,
	})
}

// HandleFinancialReport generates a comprehensive financial report
// GET /v1/admin/analytics/financial-report?type=cash&start=2024-01-01&end=2024-12-31
func (h *Handler) HandleFinancialReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reportType := r.URL.Query().Get("type")
	if reportType == "" {
		reportType = "cash" // Default to cash basis
	}

	// Parse date range, default to current month
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	if s := r.URL.Query().Get("start"); s != "" {
		if parsed, err := time.Parse("2006-01-02", s); err == nil {
			start = parsed
		}
	}
	if e := r.URL.Query().Get("end"); e != "" {
		if parsed, err := time.Parse("2006-01-02", e); err == nil {
			end = parsed
		}
	}

	report, err := h.analyticsRepo.GenerateFinancialReport(ctx, reportType, start, end)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate financial report")
		apierror.WriteError(w, apierror.NewInternal("Failed to generate financial report"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":    report,
		"success": true,
	})
}

// HandleTaxJurisdictionReport returns tax collection by jurisdiction
// GET /v1/admin/analytics/tax-jurisdiction?month=2024-04
func (h *Handler) HandleTaxJurisdictionReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Default to current month
	month := time.Now().Format("2006-01")
	if m := r.URL.Query().Get("month"); m != "" {
		month = m
	}

	report, err := h.analyticsRepo.GetTaxJurisdictionReport(ctx, month)
	if err != nil {
		logrus.WithError(err).Error("Failed to get tax jurisdiction report")
		apierror.WriteError(w, apierror.NewInternal("Failed to get tax jurisdiction report"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":    report,
		"success": true,
	})
}
