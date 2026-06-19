package billing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ExportHandler handles usage data export API endpoints
type ExportHandler struct {
	exportRepo *storage.ExportRepository
	exportSvc  *services.ExportService
	repo       storage.Repository
	logger     *logrus.Logger
}

// NewExportHandler creates a new export handler
func NewExportHandler(exportRepo *storage.ExportRepository, exportSvc *services.ExportService, repo storage.Repository) *ExportHandler {
	return &ExportHandler{
		exportRepo: exportRepo,
		exportSvc:  exportSvc,
		repo:       repo,
		logger:     logrus.New(),
	}
}

// ==================== Export Configuration Endpoints ====================

// CreateExportConfiguration creates a new export configuration
//
// POST /api/v1/exports/configurations
//
// Request body:
//
//	{
//	  "name": "Monthly Usage Report",
//	  "description": "Automated monthly usage export",
//	  "format": "csv",
//	  "data_types": ["usage", "costs"],
//	  "granularity": "daily",
//	  "include_metadata": true,
//	  "date_range_type": "last_30d",
//	  "is_scheduled": true,
//	  "schedule_frequency": "monthly",
//	  "delivery_method": "download"
//	}
func (h *ExportHandler) CreateExportConfiguration(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	var req struct {
		Name               string                 `json:"name"`
		Description        string                 `json:"description"`
		Format             storage.UsageExportFormat `json:"format"`
		DataTypes          []string               `json:"data_types"`
		Granularity        string                 `json:"granularity"`
		IncludeMetadata    bool                   `json:"include_metadata"`
		IncludeBreakdown   bool                   `json:"include_breakdown"`
		DateRangeType      string                 `json:"date_range_type"`
		FunctionFilter     []uuid.UUID            `json:"function_filter"`
		RegionFilter       []string               `json:"region_filter"`
		OutcomeFilter      []string               `json:"outcome_filter"`
		IsScheduled        bool                   `json:"is_scheduled"`
		ScheduleFrequency  string                 `json:"schedule_frequency"`
		ScheduleDayOfMonth *int                   `json:"schedule_day_of_month"`
		ScheduleDayOfWeek  *int                   `json:"schedule_day_of_week"`
		ScheduleHour       *int                   `json:"schedule_hour"`
		DeliveryMethod     string                 `json:"delivery_method"`
		EmailRecipients    []string               `json:"email_recipients"`
		WebhookURL         string                 `json:"webhook_url"`
		S3Bucket           string                 `json:"s3_bucket"`
		S3Prefix           string                 `json:"s3_prefix"`
		ExternalSystemID   *uuid.UUID             `json:"external_system_id"`
		FieldMapping       map[string]string      `json:"field_mapping"`
		TransformConfig    map[string]interface{} `json:"transform_config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "Invalid Request", "export handler request", err, h.writeError)
		return
	}

	// Validate required fields
	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "name is required")
		return
	}
	if req.Format == "" {
		req.Format = storage.ExportFormatCSV
	}
	if len(req.DataTypes) == 0 {
		req.DataTypes = []string{"usage"}
	}
	if req.DeliveryMethod == "" {
		req.DeliveryMethod = "download"
	}

	// Get user ID from context
	userID := h.extractUserID(r)
	if userID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	config := &storage.UsageExportConfiguration{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		Name:               req.Name,
		Description:        req.Description,
		Format:             req.Format,
		DataTypes:          req.DataTypes,
		Granularity:        req.Granularity,
		IncludeMetadata:    req.IncludeMetadata,
		IncludeBreakdown:   req.IncludeBreakdown,
		DateRangeType:      req.DateRangeType,
		FunctionFilter:     req.FunctionFilter,
		RegionFilter:       req.RegionFilter,
		OutcomeFilter:      req.OutcomeFilter,
		IsScheduled:        req.IsScheduled,
		ScheduleFrequency:  req.ScheduleFrequency,
		ScheduleDayOfMonth: req.ScheduleDayOfMonth,
		ScheduleDayOfWeek:  req.ScheduleDayOfWeek,
		ScheduleHour:       req.ScheduleHour,
		DeliveryMethod:     req.DeliveryMethod,
		EmailRecipients:    req.EmailRecipients,
		WebhookURL:         req.WebhookURL,
		S3Bucket:           req.S3Bucket,
		S3Prefix:           req.S3Prefix,
		ExternalSystemID:   req.ExternalSystemID,
		FieldMapping:       req.FieldMapping,
		TransformConfig:    req.TransformConfig,
		IsActive:           true,
		CreatedBy:          userID,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := h.exportRepo.CreateUsageExportConfiguration(r.Context(), config); err != nil {
		h.logger.WithError(err).Error("Failed to create export configuration")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to create export configuration")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(h.formatExportConfiguration(config))
}

// ListExportConfigurations lists all export configurations for the tenant
//
// GET /api/v1/exports/configurations
//
// Query params:
//   - limit: Page size (default: 20, max: 100)
//   - offset: Pagination offset
func (h *ExportHandler) ListExportConfigurations(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	configs, err := h.exportRepo.ListUsageExportConfigurations(r.Context(), tenantID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list export configurations")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to list configurations")
		return
	}

	formatted := make([]map[string]interface{}, len(configs))
	for i, config := range configs {
		formatted[i] = h.formatExportConfiguration(config)
	}

	response := map[string]interface{}{
		"configurations": formatted,
		"limit":          limit,
		"offset":         offset,
		"total":          len(formatted),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetExportConfiguration gets a specific export configuration
//
// GET /api/v1/exports/configurations/{id}
func (h *ExportHandler) GetExportConfiguration(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	vars := mux.Vars(r)
	configID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid configuration ID")
		return
	}

	config, err := h.exportRepo.GetUsageExportConfiguration(r.Context(), configID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get export configuration")
		h.writeError(w, http.StatusNotFound, "Not Found", "Configuration not found")
		return
	}

	// Verify tenant ownership
	if config.TenantID != tenantID {
		h.writeError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.formatExportConfiguration(config))
}

// UpdateExportConfiguration updates an export configuration
//
// PUT /api/v1/exports/configurations/{id}
func (h *ExportHandler) UpdateExportConfiguration(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	vars := mux.Vars(r)
	configID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid configuration ID")
		return
	}

	existing, err := h.exportRepo.GetUsageExportConfiguration(r.Context(), configID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Not Found", "Configuration not found")
		return
	}

	if existing.TenantID != tenantID {
		h.writeError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	var updates struct {
		Name               string                 `json:"name"`
		Description        string                 `json:"description"`
		Format             storage.UsageExportFormat `json:"format"`
		DataTypes          []string               `json:"data_types"`
		Granularity        string                 `json:"granularity"`
		IncludeMetadata    bool                   `json:"include_metadata"`
		IncludeBreakdown   bool                   `json:"include_breakdown"`
		DateRangeType      string                 `json:"date_range_type"`
		FunctionFilter     []uuid.UUID            `json:"function_filter"`
		RegionFilter       []string               `json:"region_filter"`
		OutcomeFilter      []string               `json:"outcome_filter"`
		IsScheduled        bool                   `json:"is_scheduled"`
		ScheduleFrequency  string                 `json:"schedule_frequency"`
		ScheduleDayOfMonth *int                   `json:"schedule_day_of_month"`
		ScheduleDayOfWeek  *int                   `json:"schedule_day_of_week"`
		ScheduleHour       *int                   `json:"schedule_hour"`
		DeliveryMethod     string                 `json:"delivery_method"`
		EmailRecipients    []string               `json:"email_recipients"`
		WebhookURL         string                 `json:"webhook_url"`
		S3Bucket           string                 `json:"s3_bucket"`
		S3Prefix           string                 `json:"s3_prefix"`
		IsActive           *bool                  `json:"is_active"`
		FieldMapping       map[string]string      `json:"field_mapping"`
		TransformConfig    map[string]interface{} `json:"transform_config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "Invalid Request", "export handler request", err, h.writeError)
		return
	}

	// Update fields
	if updates.Name != "" {
		existing.Name = updates.Name
	}
	if updates.Description != "" {
		existing.Description = updates.Description
	}
	if updates.Format != "" {
		existing.Format = updates.Format
	}
	if updates.DataTypes != nil {
		existing.DataTypes = updates.DataTypes
	}
	if updates.Granularity != "" {
		existing.Granularity = updates.Granularity
	}
	existing.IncludeMetadata = updates.IncludeMetadata
	existing.IncludeBreakdown = updates.IncludeBreakdown
	if updates.DateRangeType != "" {
		existing.DateRangeType = updates.DateRangeType
	}
	if updates.FunctionFilter != nil {
		existing.FunctionFilter = updates.FunctionFilter
	}
	if updates.RegionFilter != nil {
		existing.RegionFilter = updates.RegionFilter
	}
	if updates.OutcomeFilter != nil {
		existing.OutcomeFilter = updates.OutcomeFilter
	}
	existing.IsScheduled = updates.IsScheduled
	if updates.ScheduleFrequency != "" {
		existing.ScheduleFrequency = updates.ScheduleFrequency
	}
	if updates.ScheduleDayOfMonth != nil {
		existing.ScheduleDayOfMonth = updates.ScheduleDayOfMonth
	}
	if updates.ScheduleDayOfWeek != nil {
		existing.ScheduleDayOfWeek = updates.ScheduleDayOfWeek
	}
	if updates.ScheduleHour != nil {
		existing.ScheduleHour = updates.ScheduleHour
	}
	if updates.DeliveryMethod != "" {
		existing.DeliveryMethod = updates.DeliveryMethod
	}
	if updates.EmailRecipients != nil {
		existing.EmailRecipients = updates.EmailRecipients
	}
	if updates.WebhookURL != "" {
		existing.WebhookURL = updates.WebhookURL
	}
	if updates.S3Bucket != "" {
		existing.S3Bucket = updates.S3Bucket
	}
	if updates.S3Prefix != "" {
		existing.S3Prefix = updates.S3Prefix
	}
	if updates.IsActive != nil {
		existing.IsActive = *updates.IsActive
	}
	if updates.FieldMapping != nil {
		existing.FieldMapping = updates.FieldMapping
	}
	if updates.TransformConfig != nil {
		existing.TransformConfig = updates.TransformConfig
	}

	existing.UpdatedAt = time.Now()

	if err := h.exportRepo.UpdateUsageExportConfiguration(r.Context(), existing); err != nil {
		h.logger.WithError(err).Error("Failed to update export configuration")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to update configuration")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.formatExportConfiguration(existing))
}

// DeleteExportConfiguration deletes an export configuration
//
// DELETE /api/v1/exports/configurations/{id}
func (h *ExportHandler) DeleteExportConfiguration(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	vars := mux.Vars(r)
	configID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid configuration ID")
		return
	}

	existing, err := h.exportRepo.GetUsageExportConfiguration(r.Context(), configID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Not Found", "Configuration not found")
		return
	}

	if existing.TenantID != tenantID {
		h.writeError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	if err := h.exportRepo.DeleteUsageExportConfiguration(r.Context(), configID); err != nil {
		h.logger.WithError(err).Error("Failed to delete export configuration")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to delete configuration")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ==================== Export Job Endpoints ====================

// ExecuteExport executes an export based on a configuration
//
// POST /api/v1/exports/execute
//
// Request body:
//
//	{
//	  "configuration_id": "uuid",
//	  "period_start": "2026-04-01",
//	  "period_end": "2026-04-30"
//	}
func (h *ExportHandler) ExecuteExport(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	var req struct {
		ConfigurationID uuid.UUID  `json:"configuration_id"`
		PeriodStart     *time.Time `json:"period_start"`
		PeriodEnd       *time.Time `json:"period_end"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "Invalid Request", "export handler request", err, h.writeError)
		return
	}

	if req.ConfigurationID == uuid.Nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "configuration_id is required")
		return
	}

	// Get the configuration
	config, err := h.exportRepo.GetUsageExportConfiguration(r.Context(), req.ConfigurationID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Not Found", "Configuration not found")
		return
	}

	if config.TenantID != tenantID {
		h.writeError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	// Override period if specified
	if req.PeriodStart != nil {
		// Would need to add PeriodStart/End to config temporarily
	}

	// Execute the export
	result, err := h.exportSvc.ExecuteExport(r.Context(), config, "manual")
	if err != nil {
		h.logger.WithError(err).Error("Failed to execute export")
		writeErrorFromErr(r, w, http.StatusInternalServerError, "Internal Error", "execute export", err, h.writeError)
		return
	}

	response := map[string]interface{}{
		"job_id":          result.JobID.String(),
		"status":          "completed",
		"download_url":    result.DownloadURL,
		"record_count":    result.RecordCount,
		"file_size_bytes": result.FileSizeBytes,
		"checksum":        result.Checksum,
		"format":          result.Format,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

// ListExportJobs lists export jobs for the tenant
//
// GET /api/v1/exports/jobs
//
// Query params:
//   - limit: Page size (default: 20, max: 100)
//   - offset: Pagination offset
//   - status: Filter by status (pending, processing, completed, failed)
func (h *ExportHandler) ListExportJobs(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	jobs, err := h.exportRepo.ListUsageExportJobs(r.Context(), tenantID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list export jobs")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to list jobs")
		return
	}

	formatted := make([]map[string]interface{}, len(jobs))
	for i, job := range jobs {
		formatted[i] = h.formatExportJob(job)
	}

	response := map[string]interface{}{
		"jobs":   formatted,
		"limit":  limit,
		"offset": offset,
		"total":  len(formatted),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetExportJob gets a specific export job
//
// GET /api/v1/exports/jobs/{id}
func (h *ExportHandler) GetExportJob(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid job ID")
		return
	}

	job, err := h.exportRepo.GetUsageExportJob(r.Context(), jobID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Not Found", "Job not found")
		return
	}

	if job.TenantID != tenantID {
		h.writeError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.formatExportJob(job))
}

// DownloadExport downloads an export file
//
// GET /api/v1/exports/{id}/download
func (h *ExportHandler) DownloadExport(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid job ID")
		return
	}

	// Verify job ownership
	job, err := h.exportRepo.GetUsageExportJob(r.Context(), jobID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Not Found", "Job not found")
		return
	}

	if job.TenantID != tenantID {
		h.writeError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	if job.Status != storage.ExportStatusCompleted {
		h.writeError(w, http.StatusBadRequest, "Bad Request", "Export not yet completed")
		return
	}

	// Get the export file
	data, contentType, err := h.exportSvc.GetExportFile(r.Context(), jobID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get export file")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve export file")
		return
	}

	// Determine filename extension
	ext := ".bin"
	switch job.Format {
	case storage.ExportFormatCSV:
		ext = ".csv"
	case storage.ExportFormatJSON:
		ext = ".json"
	case storage.ExportFormatExcel:
		ext = ".xlsx"
	case storage.ExportFormatParquet:
		ext = ".parquet"
	}

	filename := fmt.Sprintf("usage_export_%s%s", jobID.String(), ext)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

// ==================== Export Templates Endpoints ====================

// ListExportTemplates lists available export templates
//
// GET /api/v1/exports/templates
//
// Query params:
//   - category: Filter by category (financial, operational, compliance)
func (h *ExportHandler) ListExportTemplates(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	templates, err := h.exportRepo.ListUsageExportTemplates(r.Context(), category)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list export templates")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to list templates")
		return
	}

	formatted := make([]map[string]interface{}, len(templates))
	for i, template := range templates {
		formatted[i] = h.formatExportTemplate(template)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": formatted,
	})
}

// GetExportTemplate gets a specific export template
//
// GET /api/v1/exports/templates/{id}
func (h *ExportHandler) GetExportTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid template ID")
		return
	}

	template, err := h.exportRepo.GetUsageExportTemplate(r.Context(), templateID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Not Found", "Template not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.formatExportTemplate(template))
}

// CreateConfigurationFromTemplate creates an export configuration from a template
//
// POST /api/v1/exports/templates/{id}/use
func (h *ExportHandler) CreateConfigurationFromTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	vars := mux.Vars(r)
	templateID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid template ID")
		return
	}

	template, err := h.exportRepo.GetUsageExportTemplate(r.Context(), templateID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Not Found", "Template not found")
		return
	}

	var req struct {
		Name           string `json:"name"`
		DeliveryMethod string `json:"delivery_method"`
		IsScheduled    bool   `json:"is_scheduled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "Invalid Request", "export handler request", err, h.writeError)
		return
	}

	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "name is required")
		return
	}

	userID := h.extractUserID(r)
	if userID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	config := &storage.UsageExportConfiguration{
		ID:              uuid.New(),
		TenantID:        tenantID,
		Name:            req.Name,
		Description:     template.Description,
		Format:          template.Format,
		DataTypes:       template.DataTypes,
		Granularity:     template.Granularity,
		IncludeMetadata: template.IncludeMetadata,
		IncludeBreakdown: template.IncludeBreakdown,
		DateRangeType:   "last_30d",
		IsScheduled:     req.IsScheduled,
		DeliveryMethod:  req.DeliveryMethod,
		IsActive:        true,
		CreatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if config.DeliveryMethod == "" {
		config.DeliveryMethod = "download"
	}

	if err := h.exportRepo.CreateUsageExportConfiguration(r.Context(), config); err != nil {
		h.logger.WithError(err).Error("Failed to create export configuration from template")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to create configuration")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(h.formatExportConfiguration(config))
}

// ==================== Helper Methods ====================

func (h *ExportHandler) extractTenantID(r *http.Request) uuid.UUID {
	if tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID); ok {
		return tenantID
	}

	if user, ok := r.Context().Value("user").(*storage.User); ok && user.TenantID != uuid.Nil {
		return user.TenantID
	}

	return uuid.Nil
}

func (h *ExportHandler) extractUserID(r *http.Request) uuid.UUID {
	if user, ok := r.Context().Value("user").(*storage.User); ok {
		return user.ID
	}
	return uuid.Nil
}

func (h *ExportHandler) formatExportConfiguration(config *storage.UsageExportConfiguration) map[string]interface{} {
	return map[string]interface{}{
		"id":                  config.ID.String(),
		"tenant_id":           config.TenantID.String(),
		"name":                config.Name,
		"description":         config.Description,
		"format":              config.Format,
		"data_types":          config.DataTypes,
		"granularity":         config.Granularity,
		"include_metadata":    config.IncludeMetadata,
		"include_breakdown":   config.IncludeBreakdown,
		"date_range_type":     config.DateRangeType,
		"function_filter":     config.FunctionFilter,
		"region_filter":       config.RegionFilter,
		"outcome_filter":      config.OutcomeFilter,
		"is_scheduled":        config.IsScheduled,
		"schedule_frequency":  config.ScheduleFrequency,
		"schedule_day_of_month": config.ScheduleDayOfMonth,
		"schedule_day_of_week":  config.ScheduleDayOfWeek,
		"schedule_hour":       config.ScheduleHour,
		"delivery_method":     config.DeliveryMethod,
		"email_recipients":    config.EmailRecipients,
		"webhook_url":         config.WebhookURL,
		"s3_bucket":           config.S3Bucket,
		"s3_prefix":           config.S3Prefix,
		"external_system_id":  config.ExternalSystemID,
		"is_active":           config.IsActive,
		"created_by":          config.CreatedBy.String(),
		"created_at":          config.CreatedAt.Format(time.RFC3339),
		"updated_at":          config.UpdatedAt.Format(time.RFC3339),
		"last_executed_at":    config.LastExecutedAt,
		"last_export_id":      config.LastExportID,
	}
}

func (h *ExportHandler) formatExportJob(job *storage.UsageExportJob) map[string]interface{} {
	return map[string]interface{}{
		"id":               job.ID.String(),
		"configuration_id": job.ConfigurationID.String(),
		"tenant_id":        job.TenantID.String(),
		"status":           job.Status,
		"format":           job.Format,
		"data_types":       job.DataTypes,
		"period_start":     job.PeriodStart.Format("2006-01-02"),
		"period_end":       job.PeriodEnd.Format("2006-01-02"),
		"record_count":     job.RecordCount,
		"file_size_bytes":  job.FileSizeBytes,
		"storage_provider": job.StorageProvider,
		"storage_url":      job.StorageURL,
		"checksum":         job.Checksum,
		"started_at":       job.StartedAt,
		"completed_at":     job.CompletedAt,
		"expires_at":       job.ExpiresAt,
		"error_message":    job.ErrorMessage,
		"retry_count":      job.RetryCount,
		"delivered_at":     job.DeliveredAt,
		"delivery_method":  job.DeliveryMethod,
		"delivery_status":  job.DeliveryStatus,
		"created_at":       job.CreatedAt.Format(time.RFC3339),
		"triggered_by":     job.TriggeredBy,
	}
}

func (h *ExportHandler) formatExportTemplate(template *storage.UsageExportTemplate) map[string]interface{} {
	return map[string]interface{}{
		"id":               template.ID.String(),
		"name":             template.Name,
		"description":      template.Description,
		"category":         template.Category,
		"format":           template.Format,
		"data_types":       template.DataTypes,
		"granularity":      template.Granularity,
		"include_metadata": template.IncludeMetadata,
		"include_breakdown": template.IncludeBreakdown,
		"default_fields":   template.DefaultFields,
		"is_active":        template.IsActive,
		"is_system":        template.IsSystem,
		"created_at":       template.CreatedAt.Format(time.RFC3339),
		"updated_at":       template.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *ExportHandler) writeError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}
