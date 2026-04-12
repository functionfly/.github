package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RetentionHandler handles execution log retention policy management for admin API.
type RetentionHandler struct {
	postgresDB           *storage.PostgresDB
	executionLogCleanup  *storage.ExecutionLogCleanupService
}

// NewRetentionHandler creates a new admin retention handler.
func NewRetentionHandler(postgresDB *storage.PostgresDB, cleanupService *storage.ExecutionLogCleanupService) *RetentionHandler {
	return &RetentionHandler{
		postgresDB:          postgresDB,
		executionLogCleanup: cleanupService,
	}
}

// RetentionSettingsResponse represents the retention settings API response
type RetentionSettingsResponse struct {
	ExecutionRetentionDays       int       `json:"execution_retention_days"`
	PublicExecutionRetentionDays int       `json:"public_execution_retention_days"`
	ResourceUsageRetentionDays   int       `json:"resource_usage_retention_days"`
	MEGRecordRetentionDays       int       `json:"meg_record_retention_days"`
	DriftReportRetentionDays     int       `json:"drift_report_retention_days"`
	ExecutionCertRetentionDays   int       `json:"execution_cert_retention_days"`
	CleanupIntervalMinutes       int       `json:"cleanup_interval_minutes"`
	BatchSize                    int       `json:"batch_size"`
	VerboseLogging               bool      `json:"verbose_logging"`
	IsActive                     bool      `json:"is_active"`
	UpdatedAt                    time.Time `json:"updated_at"`
	UpdatedBy                    *uuid.UUID `json:"updated_by,omitempty"`
}

// RetentionStatsResponse represents the retention statistics API response
type RetentionStatsResponse struct {
	Tables map[string]TableStats `json:"tables"`
	Summary RetentionSummary       `json:"summary"`
}

// TableStats represents statistics for a single table
type TableStats struct {
	Total            int64 `json:"total"`
	OlderThan30d     int64 `json:"older_than_30d"`
	OlderThan90d     int64 `json:"older_than_90d"`
	OlderThan365d    int64 `json:"older_than_365d"`
}

// RetentionSummary provides a high-level summary of retention status
type RetentionSummary struct {
	TotalExecutions        int64 `json:"total_executions"`
	ExecutionsOlderThan90d int64 `json:"executions_older_than_90_days"`
	EstimatedCleanupImpact int64 `json:"estimated_cleanup_impact"`
}

// UpdateRetentionRequest represents a request to update retention settings
type UpdateRetentionRequest struct {
	ExecutionRetentionDays       *int `json:"execution_retention_days,omitempty"`
	PublicExecutionRetentionDays *int `json:"public_execution_retention_days,omitempty"`
	ResourceUsageRetentionDays   *int `json:"resource_usage_retention_days,omitempty"`
	MEGRecordRetentionDays       *int `json:"meg_record_retention_days,omitempty"`
	DriftReportRetentionDays     *int `json:"drift_report_retention_days,omitempty"`
	ExecutionCertRetentionDays   *int `json:"execution_cert_retention_days,omitempty"`
	CleanupIntervalMinutes     *int `json:"cleanup_interval_minutes,omitempty"`
	BatchSize                  *int `json:"batch_size,omitempty"`
	VerboseLogging             *bool `json:"verbose_logging,omitempty"`
}

// ManualCleanupResponse represents the response from a manual cleanup operation
type ManualCleanupResponse struct {
	Success    bool              `json:"success"`
	Message    string            `json:"message"`
	Deleted    map[string]int64  `json:"deleted"`
	DurationMs int64             `json:"duration_ms"`
	Errors     []string          `json:"errors,omitempty"`
}

// HandleGetRetentionSettings returns GET /v1/admin/retention/settings
func (h *RetentionHandler) HandleGetRetentionSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get or create default settings
	settings, err := h.postgresDB.GetOrCreateExecutionRetentionSettings(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get retention settings")
		http.Error(w, "Failed to get retention settings", http.StatusInternalServerError)
		return
	}

	resp := RetentionSettingsResponse{
		ExecutionRetentionDays:       settings.ExecutionRetentionDays,
		PublicExecutionRetentionDays: settings.PublicExecutionRetentionDays,
		ResourceUsageRetentionDays:   settings.ResourceUsageRetentionDays,
		MEGRecordRetentionDays:       settings.MEGRecordRetentionDays,
		DriftReportRetentionDays:     settings.DriftReportRetentionDays,
		ExecutionCertRetentionDays:   settings.ExecutionCertRetentionDays,
		CleanupIntervalMinutes:     settings.CleanupIntervalMinutes,
		BatchSize:                    settings.BatchSize,
		VerboseLogging:               settings.VerboseLogging,
		IsActive:                     settings.IsActive,
		UpdatedAt:                    settings.UpdatedAt,
		UpdatedBy:                    settings.UpdatedBy,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleUpdateRetentionSettings returns PUT /v1/admin/retention/settings
func (h *RetentionHandler) HandleUpdateRetentionSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req UpdateRetentionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user ID from context for audit
	var updatedBy *uuid.UUID
	if userIDStr := r.Header.Get("X-User-ID"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			updatedBy = &userID
		}
	}

	updates := &storage.ExecutionRetentionSettingsUpdate{
		ExecutionRetentionDays:       req.ExecutionRetentionDays,
		PublicExecutionRetentionDays: req.PublicExecutionRetentionDays,
		ResourceUsageRetentionDays:   req.ResourceUsageRetentionDays,
		MEGRecordRetentionDays:       req.MEGRecordRetentionDays,
		DriftReportRetentionDays:     req.DriftReportRetentionDays,
		ExecutionCertRetentionDays:   req.ExecutionCertRetentionDays,
		CleanupIntervalMinutes:       req.CleanupIntervalMinutes,
		BatchSize:                    req.BatchSize,
		VerboseLogging:               req.VerboseLogging,
		UpdatedBy:                    updatedBy,
	}

	settings, err := h.postgresDB.UpdateExecutionRetentionSettings(ctx, updates)
	if err != nil {
		logrus.WithError(err).Error("Failed to update retention settings")
		http.Error(w, "Failed to update retention settings", http.StatusInternalServerError)
		return
	}

	// Also update the running cleanup service config if available
	if h.executionLogCleanup != nil {
		newConfig := storage.ExecutionRetentionConfig{
			ExecutionRetentionDays:       settings.ExecutionRetentionDays,
			PublicExecutionRetentionDays: settings.PublicExecutionRetentionDays,
			ResourceUsageRetentionDays:   settings.ResourceUsageRetentionDays,
			MEGRecordRetentionDays:       settings.MEGRecordRetentionDays,
			DriftReportRetentionDays:     settings.DriftReportRetentionDays,
			ExecutionCertRetentionDays:   settings.ExecutionCertRetentionDays,
			CleanupInterval:              time.Duration(settings.CleanupIntervalMinutes) * time.Minute,
			BatchSize:                    settings.BatchSize,
			VerboseLogging:               settings.VerboseLogging,
		}
		h.executionLogCleanup.UpdateConfig(newConfig)
	}

	logrus.WithFields(logrus.Fields{
		"updated_by": updatedBy,
		"settings":   settings,
	}).Info("Retention settings updated")

	resp := RetentionSettingsResponse{
		ExecutionRetentionDays:       settings.ExecutionRetentionDays,
		PublicExecutionRetentionDays: settings.PublicExecutionRetentionDays,
		ResourceUsageRetentionDays:   settings.ResourceUsageRetentionDays,
		MEGRecordRetentionDays:       settings.MEGRecordRetentionDays,
		DriftReportRetentionDays:     settings.DriftReportRetentionDays,
		ExecutionCertRetentionDays:   settings.ExecutionCertRetentionDays,
		CleanupIntervalMinutes:       settings.CleanupIntervalMinutes,
		BatchSize:                    settings.BatchSize,
		VerboseLogging:               settings.VerboseLogging,
		IsActive:                     settings.IsActive,
		UpdatedAt:                    settings.UpdatedAt,
		UpdatedBy:                    settings.UpdatedBy,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleGetRetentionStats returns GET /v1/admin/retention/stats
func (h *RetentionHandler) HandleGetRetentionStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.postgresDB.GetExecutionRetentionStats(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get retention stats")
		http.Error(w, "Failed to get retention statistics", http.StatusInternalServerError)
		return
	}

	// Build table stats
	tables := make(map[string]TableStats)
	tableNames := []string{
		"registry_function_executions",
		"registry_executions_public",
		"execution_resource_usage",
		"execution_meg_records",
		"drift_reports",
		"execution_certificates",
	}

	for _, tableName := range tableNames {
		if tableData, ok := stats[tableName].(map[string]interface{}); ok {
			tables[tableName] = TableStats{
				Total:         getInt64(tableData["total"]),
				OlderThan30d:  getInt64(tableData["older_than_30d"]),
				OlderThan90d:  getInt64(tableData["older_than_90d"]),
				OlderThan365d: getInt64(tableData["older_than_365d"]),
			}
		}
	}

	// Build summary
	summary := RetentionSummary{
		TotalExecutions:        getInt64(stats["total_executions"]),
		ExecutionsOlderThan90d: getInt64(stats["executions_older_than_90_days"]),
	}

	// Calculate estimated cleanup impact based on current settings
	settings, err := h.postgresDB.GetOrCreateExecutionRetentionSettings(ctx)
	if err == nil && settings != nil {
		// Estimate records that would be cleaned up based on retention settings
		if execStats, ok := tables["registry_function_executions"]; ok {
			switch {
			case settings.ExecutionRetentionDays <= 30:
				summary.EstimatedCleanupImpact = execStats.OlderThan30d
			case settings.ExecutionRetentionDays <= 90:
				summary.EstimatedCleanupImpact = execStats.OlderThan90d
			case settings.ExecutionRetentionDays <= 365:
				summary.EstimatedCleanupImpact = execStats.OlderThan365d
			}
		}
	}

	resp := RetentionStatsResponse{
		Tables:  tables,
		Summary: summary,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleRunManualCleanup returns POST /v1/admin/retention/cleanup
func (h *RetentionHandler) HandleRunManualCleanup(w http.ResponseWriter, r *http.Request) {
	if h.executionLogCleanup == nil {
		http.Error(w, "Cleanup service not available", http.StatusServiceUnavailable)
		return
	}

	ctx := r.Context()
	start := time.Now()

	// Run the cleanup
	metrics, err := h.executionLogCleanup.RunManualCleanup(ctx)
	duration := time.Since(start)

	resp := ManualCleanupResponse{
		Success:    err == nil,
		DurationMs: duration.Milliseconds(),
		Deleted: map[string]int64{
			"registry_function_executions": metrics.ExecutionsDeleted,
			"registry_executions_public":   metrics.PublicExecutionsDeleted,
			"execution_resource_usage":     metrics.ResourceUsageDeleted,
			"execution_meg_records":       metrics.MEGRecordsDeleted,
			"drift_reports":                 metrics.DriftReportsDeleted,
			"execution_certificates":        metrics.CertificatesDeleted,
			"total":                         metrics.TotalDeleted,
		},
	}

	if err != nil {
		resp.Message = "Cleanup completed with errors"
		resp.Errors = append(resp.Errors, err.Error())
		logrus.WithError(err).Warn("Manual cleanup completed with errors")
	} else {
		resp.Message = "Cleanup completed successfully"
		logrus.WithFields(logrus.Fields{
			"duration_ms":   duration.Milliseconds(),
			"total_deleted": metrics.TotalDeleted,
		}).Info("Manual retention cleanup completed")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleResetRetentionDefaults returns POST /v1/admin/retention/reset
func (h *RetentionHandler) HandleResetRetentionDefaults(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context for audit
	var updatedBy *uuid.UUID
	if userIDStr := r.Header.Get("X-User-ID"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			updatedBy = &userID
		}
	}

	settings, err := h.postgresDB.ResetExecutionRetentionSettingsToDefaults(ctx, updatedBy)
	if err != nil {
		logrus.WithError(err).Error("Failed to reset retention settings to defaults")
		http.Error(w, "Failed to reset retention settings", http.StatusInternalServerError)
		return
	}

	// Update the running cleanup service config if available
	if h.executionLogCleanup != nil {
		newConfig := storage.ExecutionRetentionConfig{
			ExecutionRetentionDays:       settings.ExecutionRetentionDays,
			PublicExecutionRetentionDays: settings.PublicExecutionRetentionDays,
			ResourceUsageRetentionDays:   settings.ResourceUsageRetentionDays,
			MEGRecordRetentionDays:       settings.MEGRecordRetentionDays,
			DriftReportRetentionDays:     settings.DriftReportRetentionDays,
			ExecutionCertRetentionDays:   settings.ExecutionCertRetentionDays,
			CleanupInterval:              time.Duration(settings.CleanupIntervalMinutes) * time.Minute,
			BatchSize:                    settings.BatchSize,
			VerboseLogging:               settings.VerboseLogging,
		}
		h.executionLogCleanup.UpdateConfig(newConfig)
	}

	logrus.WithField("updated_by", updatedBy).Info("Retention settings reset to defaults")

	resp := RetentionSettingsResponse{
		ExecutionRetentionDays:       settings.ExecutionRetentionDays,
		PublicExecutionRetentionDays: settings.PublicExecutionRetentionDays,
		ResourceUsageRetentionDays:   settings.ResourceUsageRetentionDays,
		MEGRecordRetentionDays:       settings.MEGRecordRetentionDays,
		DriftReportRetentionDays:     settings.DriftReportRetentionDays,
		ExecutionCertRetentionDays:   settings.ExecutionCertRetentionDays,
		CleanupIntervalMinutes:       settings.CleanupIntervalMinutes,
		BatchSize:                    settings.BatchSize,
		VerboseLogging:               settings.VerboseLogging,
		IsActive:                     settings.IsActive,
		UpdatedAt:                    settings.UpdatedAt,
		UpdatedBy:                    settings.UpdatedBy,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Retention settings reset to defaults",
		"settings": resp,
	})
}

// HandleGetCleanupMetrics returns GET /v1/admin/retention/metrics
func (h *RetentionHandler) HandleGetCleanupMetrics(w http.ResponseWriter, r *http.Request) {
	if h.executionLogCleanup == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"cleanup_service_available": false,
		})
		return
	}

	metrics := h.executionLogCleanup.GetMetrics()
	config := h.executionLogCleanup.GetConfig()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"cleanup_service_available": true,
		"last_run_at":              metrics.LastRunAt,
		"executions_deleted":       metrics.ExecutionsDeleted,
		"public_executions_deleted": metrics.PublicExecutionsDeleted,
		"resource_usage_deleted":   metrics.ResourceUsageDeleted,
		"meg_records_deleted":      metrics.MEGRecordsDeleted,
		"drift_reports_deleted":    metrics.DriftReportsDeleted,
		"certificates_deleted":     metrics.CertificatesDeleted,
		"total_deleted":            metrics.TotalDeleted,
		"last_duration_ms":         metrics.DurationMs,
		"had_errors":               metrics.LastError != nil,
		"current_config": map[string]interface{}{
			"execution_retention_days":        config.ExecutionRetentionDays,
			"public_execution_retention_days":  config.PublicExecutionRetentionDays,
			"resource_usage_retention_days":    config.ResourceUsageRetentionDays,
			"meg_record_retention_days":        config.MEGRecordRetentionDays,
			"drift_report_retention_days":      config.DriftReportRetentionDays,
			"execution_cert_retention_days":    config.ExecutionCertRetentionDays,
			"cleanup_interval_minutes":         int(config.CleanupInterval.Minutes()),
			"batch_size":                       config.BatchSize,
			"verbose_logging":                  config.VerboseLogging,
		},
	})
}

// getInt64 safely extracts an int64 from interface{}
func getInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	default:
		return 0
	}
}
