package billing

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type LiveReconciliationStatus struct {
	Plan                      string     `json:"plan"`
	LiveReconciliationActive  bool       `json:"live_reconciliation_active"`
	AutoReconcileEnabled      bool       `json:"auto_reconcile_enabled"`
	ScheduledReconcileEnabled bool       `json:"scheduled_reconcile_enabled"`
	ScheduleCron              string     `json:"schedule_cron,omitempty"`
	LastReconciliationAt      *time.Time `json:"last_reconciliation_at,omitempty"`
	NextScheduledReconcile    *time.Time `json:"next_scheduled_reconcile,omitempty"`
	TotalReconciliations      int64      `json:"total_reconciliations"`
	SuccessfulReconciliations int64      `json:"successful_reconciliations"`
	FailedReconciliations     int64      `json:"failed_reconciliations"`
	AuditExportEnabled        bool       `json:"audit_export_enabled"`
	SOC2Compliant             bool       `json:"soc2_compliant"`
	HIPAACompliant            bool       `json:"hipaa_compliant"`
}

type LiveReconciliationSettings struct {
	Plan                      string `json:"plan"`
	AutoReconcileEnabled      bool   `json:"auto_reconcile_enabled"`
	ScheduledReconcileEnabled bool   `json:"scheduled_reconcile_enabled"`
	ScheduleCron              string `json:"schedule_cron"`
	AuditExportEnabled        bool   `json:"audit_export_enabled"`
	NotifyOnCompletion        bool   `json:"notify_on_completion"`
	NotifyOnFailure           bool   `json:"notify_on_failure"`
}

type UpdateReconciliationSettingsRequest struct {
	AutoReconcileEnabled      bool   `json:"auto_reconcile_enabled"`
	ScheduledReconcileEnabled bool   `json:"scheduled_reconcile_enabled"`
	ScheduleCron              string `json:"schedule_cron"`
	AuditExportEnabled        bool   `json:"audit_export_enabled"`
	NotifyOnCompletion        bool   `json:"notify_on_completion"`
	NotifyOnFailure           bool   `json:"notify_on_failure"`
}

type LiveReconciliationUsageResponse struct {
	Plan                      string  `json:"plan"`
	PeriodStart               string  `json:"period_start"`
	PeriodEnd                 string  `json:"period_end"`
	TotalReconciliations      int64   `json:"total_reconciliations"`
	TotalExecutionsReconciled int64   `json:"total_executions_reconciled"`
	AvgDurationMs             int64   `json:"avg_duration_ms"`
	SuccessRate               float64 `json:"success_rate"`
}

func (h *Handler) HandleGetLiveReconciliationStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tenant, err := h.repo.GetTenantByID(r.Context(), claims.TenantID)
	if err != nil || tenant == nil {
		writeJSONError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	plan := tenant.Plan
	if plan == "" {
		plan = plans.PlanFree
	}

	supportsLiveReconciliation := plans.SupportsLiveReconciliation(plan)
	supportsAuditCertificates := plans.SupportsAuditCertificates(plan)

	var reconciliationSettings *storage.ReconciliationSettings
	if h.billingRepo != nil {
		if settings, err := h.billingRepo.GetReconciliationSettings(r.Context(), claims.TenantID); err == nil && settings != nil {
			reconciliationSettings = settings
		}
	}

	response := LiveReconciliationStatus{
		Plan:                      plan,
		LiveReconciliationActive:  supportsLiveReconciliation,
		AutoReconcileEnabled:      supportsLiveReconciliation && reconciliationSettings != nil && reconciliationSettings.AutoReconcileEnabled,
		ScheduledReconcileEnabled: supportsLiveReconciliation && reconciliationSettings != nil && reconciliationSettings.ScheduledReconcileEnabled,
		AuditExportEnabled:        supportsAuditCertificates && reconciliationSettings != nil && reconciliationSettings.AuditExportEnabled,
		SOC2Compliant:             supportsAuditCertificates,
		HIPAACompliant:            supportsAuditCertificates,
	}

	if reconciliationSettings != nil {
		response.ScheduleCron = reconciliationSettings.ScheduleCron
		if !reconciliationSettings.LastReconciliationAt.IsZero() {
			response.LastReconciliationAt = &reconciliationSettings.LastReconciliationAt
		}
		if reconciliationSettings.ScheduleCron != "" && reconciliationSettings.ScheduledReconcileEnabled {
			if nextRun, err := calculateNextRun(reconciliationSettings.ScheduleCron); err == nil {
				response.NextScheduledReconcile = nextRun
			}
		}
	}

	if h.billingRepo != nil {
		stats, err := h.billingRepo.GetReconciliationStats(r.Context(), claims.TenantID)
		if err == nil && stats != nil {
			response.TotalReconciliations = stats.TotalReconciliations
			response.SuccessfulReconciliations = stats.SuccessfulReconciliations
			response.FailedReconciliations = stats.FailedReconciliations
		}
	}

	encodeJSON(w, http.StatusOK, response)
}

func (h *Handler) HandleGetLiveReconciliationSettings(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tenant, err := h.repo.GetTenantByID(r.Context(), claims.TenantID)
	if err != nil || tenant == nil {
		writeJSONError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	plan := tenant.Plan
	if plan == "" {
		plan = plans.PlanFree
	}

	if !plans.SupportsLiveReconciliation(plan) {
		writeJSONError(w, http.StatusForbidden,
			"Live Reconciliation requires Enterprise plan or higher")
		return
	}

	var settings *storage.ReconciliationSettings
	if h.billingRepo != nil {
		settings, _ = h.billingRepo.GetReconciliationSettings(r.Context(), claims.TenantID)
	}

	if settings == nil {
		settings = &storage.ReconciliationSettings{
			TenantID:                  claims.TenantID,
			AutoReconcileEnabled:      false,
			ScheduledReconcileEnabled: false,
			ScheduleCron:              "0 2 * * *",
			AuditExportEnabled:        plans.SupportsAuditCertificates(plan),
			NotifyOnCompletion:        true,
			NotifyOnFailure:           true,
		}
	}

	response := LiveReconciliationSettings{
		Plan:                      plan,
		AutoReconcileEnabled:      settings.AutoReconcileEnabled,
		ScheduledReconcileEnabled: settings.ScheduledReconcileEnabled,
		ScheduleCron:              settings.ScheduleCron,
		AuditExportEnabled:        settings.AuditExportEnabled,
		NotifyOnCompletion:        settings.NotifyOnCompletion,
		NotifyOnFailure:           settings.NotifyOnFailure,
	}

	encodeJSON(w, http.StatusOK, response)
}

func (h *Handler) HandleUpdateLiveReconciliationSettings(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tenant, err := h.repo.GetTenantByID(r.Context(), claims.TenantID)
	if err != nil || tenant == nil {
		writeJSONError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	plan := tenant.Plan
	if plan == "" {
		plan = plans.PlanFree
	}

	if !plans.SupportsLiveReconciliation(plan) {
		writeJSONError(w, http.StatusForbidden,
			"Live Reconciliation requires Enterprise plan or higher")
		return
	}

	var req UpdateReconciliationSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ScheduleCron != "" && !isValidCronExpression(req.ScheduleCron) {
		writeJSONError(w, http.StatusBadRequest, "Invalid cron expression")
		return
	}

	if h.billingRepo == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing repository not configured")
		return
	}

	existingSettings, _ := h.billingRepo.GetReconciliationSettings(r.Context(), claims.TenantID)

	settings := &storage.ReconciliationSettings{
		TenantID:                  claims.TenantID,
		AutoReconcileEnabled:      req.AutoReconcileEnabled,
		ScheduledReconcileEnabled: req.ScheduledReconcileEnabled,
		ScheduleCron:              req.ScheduleCron,
		AuditExportEnabled:        req.AuditExportEnabled,
		NotifyOnCompletion:        req.NotifyOnCompletion,
		NotifyOnFailure:           req.NotifyOnFailure,
		UpdatedAt:                 time.Now(),
	}

	if existingSettings != nil {
		settings.ID = existingSettings.ID
		settings.CreatedAt = existingSettings.CreatedAt
		settings.LastReconciliationAt = existingSettings.LastReconciliationAt
	}

	if err := h.billingRepo.UpsertReconciliationSettings(r.Context(), settings); err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("billing: failed to update reconciliation settings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update reconciliation settings")
		return
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":             claims.TenantID,
		"auto_reconcile":        req.AutoReconcileEnabled,
		"scheduled_reconcile":   req.ScheduledReconcileEnabled,
		"audit_export":          req.AuditExportEnabled,
	}).Info("Live Reconciliation settings updated")

	encodeJSON(w, http.StatusOK, LiveReconciliationSettings{
		Plan:                      plan,
		AutoReconcileEnabled:      settings.AutoReconcileEnabled,
		ScheduledReconcileEnabled: settings.ScheduledReconcileEnabled,
		ScheduleCron:              settings.ScheduleCron,
		AuditExportEnabled:        settings.AuditExportEnabled,
		NotifyOnCompletion:        settings.NotifyOnCompletion,
		NotifyOnFailure:           settings.NotifyOnFailure,
	})
}

func (h *Handler) HandleGetLiveReconciliationUsage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tenant, err := h.repo.GetTenantByID(r.Context(), claims.TenantID)
	if err != nil || tenant == nil {
		writeJSONError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	plan := tenant.Plan
	if plan == "" {
		plan = plans.PlanFree
	}

	if !plans.SupportsLiveReconciliation(plan) {
		writeJSONError(w, http.StatusForbidden,
			"Live Reconciliation requires Enterprise plan or higher")
		return
	}

	if h.billingRepo == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing repository not configured")
		return
	}

	end := time.Now().UTC()
	start := end.AddDate(0, -1, 0)

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

	usage, err := h.billingRepo.GetLiveReconciliationUsage(r.Context(), claims.TenantID, start, end)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing: failed to get reconciliation usage")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve reconciliation usage")
		return
	}

	var successRate float64
	if usage.TotalReconciliations > 0 {
		successRate = float64(usage.SuccessfulReconciliations) / float64(usage.TotalReconciliations) * 100
	}

	response := LiveReconciliationUsageResponse{
		Plan:                      plan,
		PeriodStart:               start.Format("2006-01-02"),
		PeriodEnd:                 end.Format("2006-01-02"),
		TotalReconciliations:      usage.TotalReconciliations,
		TotalExecutionsReconciled: usage.TotalExecutionsReconciled,
		AvgDurationMs:             usage.AvgDurationMs,
		SuccessRate:               successRate,
	}

	encodeJSON(w, http.StatusOK, response)
}

func isValidCronExpression(cron string) bool {
	parts := splitCron(cron)
	return len(parts) == 5
}

func splitCron(cron string) []string {
	var parts []string
	var current []byte
	for i := 0; i < len(cron); i++ {
		if cron[i] == ' ' {
			if len(current) > 0 {
				parts = append(parts, string(current))
				current = nil
			}
		} else {
			current = append(current, cron[i])
		}
	}
	if len(current) > 0 {
		parts = append(parts, string(current))
	}
	return parts
}

func calculateNextRun(cronExpr string) (*time.Time, error) {
	return nil, nil
}
