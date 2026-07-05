package billing

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// UsageForecastHandler handles usage forecasting and alert APIs
type UsageForecastHandler struct {
	alertRepo   *storage.UsageAlertRepository
	billingRepo storage.Repository
	forecaster  *services.UsageForecaster
	alerter     *services.UsageAlerter
}

// NewUsageForecastHandler creates a new handler
func NewUsageForecastHandler(alertRepo *storage.UsageAlertRepository, billingRepo storage.Repository, forecaster *services.UsageForecaster, alerter *services.UsageAlerter) *UsageForecastHandler {
	return &UsageForecastHandler{
		alertRepo:   alertRepo,
		billingRepo: billingRepo,
		forecaster:  forecaster,
		alerter:     alerter,
	}
}

// RegisterRoutes registers usage forecast and alert routes on the provided router
func (h *UsageForecastHandler) RegisterRoutes(r *mux.Router) {
	// Forecast endpoints
	r.HandleFunc("/usage/forecast", h.GetCurrentForecast).Methods("GET", "OPTIONS")
	r.HandleFunc("/usage/forecast/{type}", h.GetForecastByType).Methods("GET", "OPTIONS")
	r.HandleFunc("/usage/forecast/refresh", h.RefreshForecast).Methods("POST", "OPTIONS")

	// Alert configuration endpoints
	r.HandleFunc("/usage/alerts", h.ListAlerts).Methods("GET", "OPTIONS")
	r.HandleFunc("/usage/alerts", h.CreateAlert).Methods("POST", "OPTIONS")
	r.HandleFunc("/usage/alerts/{id}", h.GetAlert).Methods("GET", "OPTIONS")
	r.HandleFunc("/usage/alerts/{id}", h.UpdateAlert).Methods("PUT", "OPTIONS")
	r.HandleFunc("/usage/alerts/{id}", h.DeleteAlert).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/usage/alerts/history", h.GetAlertHistory).Methods("GET", "OPTIONS")

	// Spend cap endpoints
	r.HandleFunc("/usage/spend-cap", h.GetSpendCap).Methods("GET", "OPTIONS")
	r.HandleFunc("/usage/spend-cap", h.UpdateSpendCap).Methods("PUT", "OPTIONS")

	// Trend analysis
	r.HandleFunc("/usage/trends", h.GetUsageTrends).Methods("GET", "OPTIONS")
}

// ForecastResponse represents a forecast API response
type ForecastResponse struct {
	TenantID       string  `json:"tenant_id"`
	ForecastType   string  `json:"forecast_type"`
	PeriodStart    string  `json:"period_start"`
	PeriodEnd      string  `json:"period_end"`
	CurrentValue   float64 `json:"current_value"`
	PredictedValue float64 `json:"predicted_value"`
	LowerBound     float64 `json:"lower_bound"`
	UpperBound     float64 `json:"upper_bound"`
	Confidence     float64 `json:"confidence"`
	MethodUsed     string  `json:"method_used"`
	GrowthRate     float64 `json:"growth_rate"`
	DaysRemaining  int     `json:"days_remaining"`
	PercentOfCap   float64 `json:"percent_of_cap,omitempty"`
	AtRisk         bool    `json:"at_risk"`
}

// GetCurrentForecast returns the current usage and spend forecasts
func (h *UsageForecastHandler) GetCurrentForecast(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		apierror.WriteError(w, apierror.NewUnauthorized("Tenant not found"))
		return
	}

	ctx := r.Context()

	// Get spend forecast
	spendForecast, err := h.forecaster.GetForecastForTenant(ctx, tenantID, "spend")
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get spend forecast"))
		return
	}

	// Get execution forecast
	execForecast, err := h.forecaster.GetForecastForTenant(ctx, tenantID, "function_execution")
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get execution forecast"))
		return
	}

	// Get compute forecast
	computeForecast, err := h.forecaster.GetForecastForTenant(ctx, tenantID, "compute_time_ms")
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get compute forecast"))
		return
	}

	response := map[string]interface{}{
		"spend":         h.formatForecastResponse(spendForecast, tenantID),
		"executions":    h.formatForecastResponse(execForecast, tenantID),
		"compute_time":  h.formatForecastResponse(computeForecast, tenantID),
	}

	encodeJSON(w, http.StatusOK, response)
}

// GetForecastByType returns a forecast for a specific metric type
func (h *UsageForecastHandler) GetForecastByType(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		apierror.WriteError(w, apierror.NewUnauthorized("Tenant not found"))
		return
	}

	vars := mux.Vars(r)
	forecastType := vars["type"]
	validTypes := map[string]bool{"spend": true, "function_execution": true, "compute_time_ms": true}
	if !validTypes[forecastType] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid forecast type"))
		return
	}

	ctx := r.Context()
	forecast, err := h.forecaster.GetForecastForTenant(ctx, tenantID, forecastType)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get forecast"))
		return
	}

	if forecast == nil {
		encodeJSON(w, http.StatusOK, map[string]interface{}{
			"error": "No forecast available yet. Insufficient historical data.",
		})
		return
	}

	encodeJSON(w, http.StatusOK, h.formatForecastResponse(forecast, tenantID))
}

// RefreshForecast triggers a fresh forecast generation
func (h *UsageForecastHandler) RefreshForecast(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		apierror.WriteError(w, apierror.NewUnauthorized("Tenant not found"))
		return
	}

	// Get subscription for period info
	sub, err := h.billingRepo.GetSubscriptionByTenantID(r.Context(), tenantID)
	if err != nil || sub == nil {
		apierror.WriteError(w, apierror.NewNotFound("Subscription not found"))
		return
	}

	ctx := r.Context()

	// Generate new forecasts
	if err := h.forecaster.GenerateSpendForecast(ctx, tenantID, sub); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to generate spend forecast"))
		return
	}

	if err := h.forecaster.GenerateUsageForecast(ctx, tenantID, "function_execution", sub); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to generate execution forecast"))
		return
	}

	encodeJSON(w, http.StatusOK, map[string]string{"status": "forecasts refreshed"})
}

// formatForecastResponse formats a forecast for API response
func (h *UsageForecastHandler) formatForecastResponse(forecast *storage.UsageForecast, tenantID uuid.UUID) *ForecastResponse {
	if forecast == nil {
		return nil
	}

	daysRemaining := int(forecast.PeriodEnd.Sub(time.Now().UTC()).Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	// Check if at risk of exceeding cap
	atRisk := false
	if forecast.ForecastType == "spend" {
		// Get spend cap to calculate risk
		cap, _ := h.alertRepo.GetSpendCapByTenant(context.Background(), tenantID, forecast.PeriodStart)
		if cap != nil && cap.IsEnabled && cap.CapAmountCents > 0 {
			percentOfCap := (forecast.PredictedValue / float64(cap.CapAmountCents)) * 100
			atRisk = percentOfCap >= 80
			return &ForecastResponse{
				TenantID:       tenantID.String(),
				ForecastType:   forecast.ForecastType,
				PeriodStart:    forecast.PeriodStart.Format(time.RFC3339),
				PeriodEnd:      forecast.PeriodEnd.Format(time.RFC3339),
				CurrentValue:   forecast.CurrentValue,
				PredictedValue: forecast.PredictedValue,
				LowerBound:     forecast.LowerBound,
				UpperBound:     forecast.UpperBound,
				Confidence:     forecast.Confidence,
				MethodUsed:     forecast.MethodUsed,
				GrowthRate:     forecast.GrowthRate,
				DaysRemaining:  daysRemaining,
				PercentOfCap:   percentOfCap,
				AtRisk:         atRisk,
			}
		}
	}

	return &ForecastResponse{
		TenantID:       tenantID.String(),
		ForecastType:   forecast.ForecastType,
		PeriodStart:    forecast.PeriodStart.Format(time.RFC3339),
		PeriodEnd:      forecast.PeriodEnd.Format(time.RFC3339),
		CurrentValue:   forecast.CurrentValue,
		PredictedValue: forecast.PredictedValue,
		LowerBound:     forecast.LowerBound,
		UpperBound:     forecast.UpperBound,
		Confidence:     forecast.Confidence,
		MethodUsed:     forecast.MethodUsed,
		GrowthRate:     forecast.GrowthRate,
		DaysRemaining:  daysRemaining,
		AtRisk:         atRisk,
	}
}

// CreateAlertRequest represents a request to create an alert
type CreateAlertRequest struct {
	Name                 string   `json:"name"`
	AlertType            string   `json:"alert_type"`
	ThresholdValue       float64  `json:"threshold_value"`
	ThresholdOperator    string   `json:"threshold_operator"`
	PeriodType           string   `json:"period_type"`
	NotificationChannels []string `json:"notification_channels"`
	CooldownMinutes      int      `json:"cooldown_minutes"`
}

// CreateAlert creates a new usage alert
func (h *UsageForecastHandler) CreateAlert(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		apierror.WriteError(w, apierror.NewUnauthorized("Tenant not found"))
		return
	}

	var req CreateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Validate
	validTypes := map[string]bool{"spend_cap": true, "usage_spike": true, "threshold": true, "forecast_exceeded": true}
	if !validTypes[req.AlertType] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid alert type"))
		return
	}

	validOperators := map[string]bool{"gte": true, "lte": true, "percentage_of_cap": true}
	if !validOperators[req.ThresholdOperator] {
		req.ThresholdOperator = "gte"
	}

	if req.CooldownMinutes < 1 {
		req.CooldownMinutes = 60
	}

	alert := &storage.UsageAlert{
		TenantID:             tenantID,
		Name:                 req.Name,
		AlertType:            req.AlertType,
		ThresholdValue:       req.ThresholdValue,
		ThresholdOperator:    req.ThresholdOperator,
		PeriodType:           req.PeriodType,
		NotificationChannels: req.NotificationChannels,
		CooldownMinutes:      req.CooldownMinutes,
		IsEnabled:            true,
	}

	ctx := r.Context()
	if err := h.alertRepo.CreateUsageAlert(ctx, alert); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to create alert"))
		return
	}

	encodeJSON(w, http.StatusCreated, alert)
}

// ListAlerts returns all alerts for the tenant
func (h *UsageForecastHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		apierror.WriteError(w, apierror.NewUnauthorized("Tenant not found"))
		return
	}

	ctx := r.Context()
	alerts, err := h.alertRepo.ListUsageAlertsByTenant(ctx, tenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to list alerts"))
		return
	}

	encodeJSON(w, http.StatusOK, alerts)
}

// GetAlert returns a specific alert
func (h *UsageForecastHandler) GetAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid alert ID"))
		return
	}

	ctx := r.Context()
	alert, err := h.alertRepo.GetUsageAlertByID(ctx, alertID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get alert"))
		return
	}

	if alert == nil {
		apierror.WriteError(w, apierror.NewNotFound("Alert not found"))
		return
	}

	// Verify tenant ownership
	tenantID, _ := middleware.GetTenantID(r)
	if alert.TenantID != tenantID {
		apierror.WriteError(w, apierror.NewForbidden("Not authorized"))
		return
	}

	encodeJSON(w, http.StatusOK, alert)
}

// UpdateAlert updates an existing alert
func (h *UsageForecastHandler) UpdateAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid alert ID"))
		return
	}

	// Get existing alert to verify ownership
	ctx := r.Context()
	existing, err := h.alertRepo.GetUsageAlertByID(ctx, alertID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get alert"))
		return
	}
	if existing == nil {
		apierror.WriteError(w, apierror.NewNotFound("Alert not found"))
		return
	}

	tenantID, _ := middleware.GetTenantID(r)
	if existing.TenantID != tenantID {
		apierror.WriteError(w, apierror.NewForbidden("Not authorized"))
		return
	}

	var req CreateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Update fields
	existing.Name = req.Name
	existing.AlertType = req.AlertType
	existing.ThresholdValue = req.ThresholdValue
	existing.ThresholdOperator = req.ThresholdOperator
	existing.PeriodType = req.PeriodType
	existing.NotificationChannels = req.NotificationChannels
	existing.CooldownMinutes = req.CooldownMinutes

	if err := h.alertRepo.UpdateUsageAlert(ctx, existing); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to update alert"))
		return
	}

	encodeJSON(w, http.StatusOK, existing)
}

// DeleteAlert deletes an alert
func (h *UsageForecastHandler) DeleteAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid alert ID"))
		return
	}

	// Get existing to verify ownership
	ctx := r.Context()
	existing, err := h.alertRepo.GetUsageAlertByID(ctx, alertID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get alert"))
		return
	}
	if existing == nil {
		apierror.WriteError(w, apierror.NewNotFound("Alert not found"))
		return
	}

	tenantID, _ := middleware.GetTenantID(r)
	if existing.TenantID != tenantID {
		apierror.WriteError(w, apierror.NewForbidden("Not authorized"))
		return
	}

	if err := h.alertRepo.DeleteUsageAlert(ctx, alertID); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to delete alert"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAlertHistory returns the alert history for the tenant
func (h *UsageForecastHandler) GetAlertHistory(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		apierror.WriteError(w, apierror.NewUnauthorized("Tenant not found"))
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	ctx := r.Context()
	history, err := h.alertRepo.GetAlertHistoryByTenant(ctx, tenantID, limit)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get alert history"))
		return
	}

	encodeJSON(w, http.StatusOK, history)
}

// SpendCapRequest represents a request to update spend cap
type SpendCapRequest struct {
	CapAmountCents    int      `json:"cap_amount_cents"`
	WarningThresholds []int    `json:"warning_thresholds"`
	ActionOnCap       string   `json:"action_on_cap"`
	IsHardCap         bool     `json:"is_hard_cap"`
	IsEnabled         bool     `json:"is_enabled"`
}

// GetSpendCap returns the current spend cap configuration
func (h *UsageForecastHandler) GetSpendCap(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		apierror.WriteError(w, apierror.NewUnauthorized("Tenant not found"))
		return
	}

	// Get subscription for period
	sub, err := h.billingRepo.GetSubscriptionByTenantID(r.Context(), tenantID)
	if err != nil || sub == nil {
		apierror.WriteError(w, apierror.NewNotFound("Subscription not found"))
		return
	}

	ctx := r.Context()
	cap, err := h.alertRepo.GetSpendCapByTenant(ctx, tenantID, sub.CurrentPeriodStart)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get spend cap"))
		return
	}

	if cap == nil {
		encodeJSON(w, http.StatusOK, map[string]interface{}{
			"configured": false,
			"message":    "No spend cap configured. Set a cap to receive proactive alerts.",
		})
		return
	}

	encodeJSON(w, http.StatusOK, cap)
}

// UpdateSpendCap creates or updates the spend cap
func (h *UsageForecastHandler) UpdateSpendCap(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		apierror.WriteError(w, apierror.NewUnauthorized("Tenant not found"))
		return
	}

	var req SpendCapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Validate
	if req.CapAmountCents < 0 {
		apierror.WriteError(w, apierror.NewBadRequest("Cap amount must be non-negative"))
		return
	}

	// Get subscription for period dates
	sub, err := h.billingRepo.GetSubscriptionByTenantID(r.Context(), tenantID)
	if err != nil || sub == nil {
		apierror.WriteError(w, apierror.NewNotFound("Subscription not found"))
		return
	}

	cap := &storage.SpendCap{
		TenantID:          tenantID,
		CapAmountCents:    req.CapAmountCents,
		WarningThresholds: req.WarningThresholds,
		PeriodStart:       sub.CurrentPeriodStart,
		PeriodEnd:         sub.CurrentPeriodEnd,
		ActionOnCap:       req.ActionOnCap,
		IsHardCap:         req.IsHardCap,
		IsEnabled:         req.IsEnabled,
	}

	ctx := r.Context()
	if err := h.alertRepo.CreateOrUpdateSpendCap(ctx, cap); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to update spend cap"))
		return
	}

	encodeJSON(w, http.StatusOK, cap)
}

// GetUsageTrends returns usage trend analysis
func (h *UsageForecastHandler) GetUsageTrends(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r)
	if !ok {
		apierror.WriteError(w, apierror.NewUnauthorized("Tenant not found"))
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}

	days := 30
	switch period {
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	}

	ctx := r.Context()

	// Get daily usage for trend analysis
	execHistory, err := h.alertRepo.GetDailyUsageHistory(ctx, tenantID, "function_execution", days)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get usage history"))
		return
	}

	spendHistory, err := h.alertRepo.GetDailySpendHistory(ctx, tenantID, days)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get spend history"))
		return
	}

	response := map[string]interface{}{
		"period":    period,
		"days":      days,
		"executions": h.calculateTrend(execHistory, days),
		"spend":     h.calculateTrend(spendHistory, days),
	}

		encodeJSON(w, http.StatusOK, response)
	}

	// calculateTrend calculates trend metrics from history
func (h *UsageForecastHandler) calculateTrend(history []*storage.DailyUsagePoint, days int) map[string]interface{} {
	if len(history) < 2 {
		return map[string]interface{}{
			"data_points": len(history),
			"message":     "Insufficient data for trend analysis",
		}
	}

	var sum, min, max float64
	min = history[0].Value
	max = history[0].Value

	for _, point := range history {
		val := point.Value
		sum += val
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
	}

	avg := sum / float64(len(history))

	// Calculate trend direction
	firstWeek := 0.0
	lastWeek := 0.0
	weekSize := 7

	if len(history) >= weekSize*2 {
		for i := 0; i < weekSize && i < len(history); i++ {
			firstWeek += history[i].Value
		}
		for i := len(history) - weekSize; i < len(history); i++ {
			lastWeek += history[i].Value
		}
		firstWeek /= float64(weekSize)
		lastWeek /= float64(weekSize)
	}

	trendDirection := "stable"
	percentChange := 0.0

	if firstWeek > 0 {
		percentChange = ((lastWeek - firstWeek) / firstWeek) * 100
		if percentChange > 10 {
			trendDirection = "increasing"
		} else if percentChange < -10 {
			trendDirection = "decreasing"
		}
	}

	return map[string]interface{}{
		"data_points":      len(history),
		"avg_daily":        avg,
		"peak_daily":       max,
		"min_daily":        min,
		"trend_direction":  trendDirection,
		"percent_change":   percentChange,
		"anomaly_detected": false, // Would be populated by anomaly detection
	}
}
