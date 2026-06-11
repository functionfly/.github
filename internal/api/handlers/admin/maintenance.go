package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// MaintenanceHandler contains maintenance mode handlers
type MaintenanceHandler struct {
	maintenanceRepo *storage.MaintenanceRepository
	authSvc         *auth.AuthService
}

// NewMaintenanceHandler creates a new maintenance handler
func NewMaintenanceHandler(maintenanceRepo *storage.MaintenanceRepository, authSvc *auth.AuthService) *MaintenanceHandler {
	return &MaintenanceHandler{
		maintenanceRepo: maintenanceRepo,
		authSvc:         authSvc,
	}
}

// HandleGetMaintenanceStatus handles GET /v1/admin/maintenance
func (h *MaintenanceHandler) HandleGetMaintenanceStatus(w http.ResponseWriter, r *http.Request) {
	maintenance, err := h.maintenanceRepo.GetPlatformMaintenance(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get maintenance status")
		apierror.WriteError(w, apierror.NewInternal("Failed to get maintenance status"))
		return
	}

	if maintenance == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"enabled": false,
		})
		return
	}

	// Get template
	template, _ := h.maintenanceRepo.GetMaintenanceTemplate(r.Context(), maintenance.PageTemplate)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled":             maintenance.Enabled,
		"name":                maintenance.Name,
		"description":         maintenance.Description,
		"message":             maintenance.Message,
		"page_template":       maintenance.PageTemplate,
		"retry_after_seconds": maintenance.RetryAfterSeconds,
		"rollout_percentage":  maintenance.RolloutPercentage,
		"scheduled_start":     maintenance.ScheduledStart,
		"scheduled_end":       maintenance.ScheduledEnd,
		"is_scheduled":        maintenance.IsScheduled,
		"timezone":            maintenance.Timezone,
		"template":            template,
		"created_at":          maintenance.CreatedAt,
		"updated_at":          maintenance.UpdatedAt,
	})
}

// HandleEnableMaintenance handles POST /v1/admin/maintenance (enable)
func (h *MaintenanceHandler) HandleEnableMaintenance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name              string     `json:"name"`
		Description       *string    `json:"description"`
		Message           *string    `json:"message"`
		PageTemplate      string     `json:"page_template"`
		RetryAfterSeconds int        `json:"retry_after_seconds"`
		RolloutPercentage int        `json:"rollout_percentage"`
		ScheduledStart    *time.Time `json:"scheduled_start"`
		ScheduledEnd      *time.Time `json:"scheduled_end"`
		IsScheduled       bool       `json:"is_scheduled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Get current maintenance config
	maintenance, err := h.maintenanceRepo.GetPlatformMaintenance(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get maintenance config")
		apierror.WriteError(w, apierror.NewInternal("Failed to get maintenance config"))
		return
	}

	if maintenance == nil {
		// Create new if doesn't exist
		maintenance = &types.PlatformMaintenance{
			Name:              req.Name,
			Description:       req.Description,
			Message:           req.Message,
			PageTemplate:      req.PageTemplate,
			RetryAfterSeconds: req.RetryAfterSeconds,
			RolloutPercentage: req.RolloutPercentage,
			ScheduledStart:    req.ScheduledStart,
			ScheduledEnd:      req.ScheduledEnd,
			IsScheduled:       req.IsScheduled,
			Timezone:          "UTC",
		}
		err = h.maintenanceRepo.UpdatePlatformMaintenance(r.Context(), maintenance)
		if err != nil {
			logrus.WithError(err).Error("Failed to create maintenance config")
			apierror.WriteError(w, apierror.NewInternal("Failed to create maintenance config"))
			return
		}
	}

	// Get user ID from context
	userID := h.getUserID(r)

	// Store old values for audit
	oldValues := maintenance.Enabled

	// Enable maintenance
	maintenance.Enabled = true
	if req.Name != "" {
		maintenance.Name = req.Name
	}
	if req.Description != nil {
		maintenance.Description = req.Description
	}
	if req.Message != nil {
		maintenance.Message = req.Message
	}
	if req.PageTemplate != "" {
		maintenance.PageTemplate = req.PageTemplate
	}
	if req.RetryAfterSeconds > 0 {
		maintenance.RetryAfterSeconds = req.RetryAfterSeconds
	}
	if req.RolloutPercentage > 0 {
		maintenance.RolloutPercentage = req.RolloutPercentage
	}
	if req.ScheduledStart != nil {
		maintenance.ScheduledStart = req.ScheduledStart
	}
	if req.ScheduledEnd != nil {
		maintenance.ScheduledEnd = req.ScheduledEnd
	}
	maintenance.IsScheduled = req.IsScheduled

	err = h.maintenanceRepo.EnableMaintenance(r.Context(), maintenance)
	if err != nil {
		logrus.WithError(err).Error("Failed to enable maintenance")
		apierror.WriteError(w, apierror.NewInternal("Failed to enable maintenance"))
		return
	}

	// Create audit log
	h.createAuditLog(r, maintenance, "enabled", map[string]interface{}{
		"enabled_before": oldValues,
		"enabled_after":  true,
	}, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled": true,
		"message": "Maintenance mode enabled",
	})
}

// HandleDisableMaintenance handles DELETE /v1/admin/maintenance
func (h *MaintenanceHandler) HandleDisableMaintenance(w http.ResponseWriter, r *http.Request) {
	maintenance, err := h.maintenanceRepo.GetPlatformMaintenance(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get maintenance config")
		apierror.WriteError(w, apierror.NewInternal("Failed to get maintenance config"))
		return
	}

	if maintenance == nil {
		apierror.WriteError(w, apierror.NewNotFound("No maintenance configuration found"))
		return
	}

	// Store old value for audit
	oldEnabled := maintenance.Enabled

	err = h.maintenanceRepo.DisableMaintenance(r.Context(), maintenance)
	if err != nil {
		logrus.WithError(err).Error("Failed to disable maintenance")
		apierror.WriteError(w, apierror.NewInternal("Failed to disable maintenance"))
		return
	}

	// Get user ID from context
	userID := h.getUserID(r)

	// Create audit log
	h.createAuditLog(r, maintenance, "disabled", map[string]interface{}{
		"enabled_before": oldEnabled,
		"enabled_after":  false,
	}, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled": false,
		"message": "Maintenance mode disabled",
	})
}

// HandleUpdateMaintenance handles PUT /v1/admin/maintenance
func (h *MaintenanceHandler) HandleUpdateMaintenance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name              *string    `json:"name"`
		Description       *string    `json:"description"`
		Message           *string    `json:"message"`
		PageTemplate      *string    `json:"page_template"`
		RetryAfterSeconds *int       `json:"retry_after_seconds"`
		RolloutPercentage *int       `json:"rollout_percentage"`
		ScheduledStart    *time.Time `json:"scheduled_start"`
		ScheduledEnd      *time.Time `json:"scheduled_end"`
		IsScheduled       *bool      `json:"is_scheduled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	maintenance, err := h.maintenanceRepo.GetPlatformMaintenance(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get maintenance config")
		apierror.WriteError(w, apierror.NewInternal("Failed to get maintenance config"))
		return
	}

	if maintenance == nil {
		apierror.WriteError(w, apierror.NewNotFound("No maintenance configuration found"))
		return
	}

	// Update fields
	if req.Name != nil {
		maintenance.Name = *req.Name
	}
	if req.Description != nil {
		maintenance.Description = req.Description
	}
	if req.Message != nil {
		maintenance.Message = req.Message
	}
	if req.PageTemplate != nil {
		maintenance.PageTemplate = *req.PageTemplate
	}
	if req.RetryAfterSeconds != nil {
		maintenance.RetryAfterSeconds = *req.RetryAfterSeconds
	}
	if req.RolloutPercentage != nil {
		maintenance.RolloutPercentage = *req.RolloutPercentage
	}
	if req.ScheduledStart != nil {
		maintenance.ScheduledStart = req.ScheduledStart
	}
	if req.ScheduledEnd != nil {
		maintenance.ScheduledEnd = req.ScheduledEnd
	}
	if req.IsScheduled != nil {
		maintenance.IsScheduled = *req.IsScheduled
	}

	err = h.maintenanceRepo.UpdatePlatformMaintenance(r.Context(), maintenance)
	if err != nil {
		logrus.WithError(err).Error("Failed to update maintenance")
		apierror.WriteError(w, apierror.NewInternal("Failed to update maintenance"))
		return
	}

	// Get user ID from context
	userID := h.getUserID(r)

	// Create audit log
	h.createAuditLog(r, maintenance, "updated", nil, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Maintenance configuration updated",
	})
}

// HandleGetTemplates handles GET /v1/admin/maintenance/templates
func (h *MaintenanceHandler) HandleGetTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.maintenanceRepo.ListTemplates(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to list templates")
		apierror.WriteError(w, apierror.NewInternal("Failed to list templates"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": templates,
	})
}

// HandleCreateTemplate handles POST /v1/admin/maintenance/templates
func (h *MaintenanceHandler) HandleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name            string  `json:"name"`
		Title           *string `json:"title"`
		MessageHTML     *string `json:"message_html"`
		LogoURL         *string `json:"logo_url"`
		BackgroundColor string  `json:"background_color"`
		TextColor       string  `json:"text_color"`
		AccentColor     string  `json:"accent_color"`
		ShowContactInfo bool    `json:"show_contact_info"`
		ContactEmail    *string `json:"contact_email"`
		ShowSocialLinks bool    `json:"show_social_links"`
		IsDefault       bool    `json:"is_default"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	template := &types.MaintenancePageTemplate{
		Name:            req.Name,
		Title:           req.Title,
		MessageHTML:     req.MessageHTML,
		LogoURL:         req.LogoURL,
		BackgroundColor: req.BackgroundColor,
		TextColor:       req.TextColor,
		AccentColor:     req.AccentColor,
		ShowContactInfo: req.ShowContactInfo,
		ContactEmail:    req.ContactEmail,
		ShowSocialLinks: req.ShowSocialLinks,
		IsDefault:       req.IsDefault,
	}

	err := h.maintenanceRepo.CreateTemplate(r.Context(), template)
	if err != nil {
		logrus.WithError(err).Error("Failed to create template")
		apierror.WriteError(w, apierror.NewInternal("Failed to create template"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"template": template,
	})
}

// HandleUpdateTemplate handles PUT /v1/admin/maintenance/templates/{id}
func (h *MaintenanceHandler) HandleUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid template ID"))
		return
	}

	var req struct {
		Name            *string `json:"name"`
		Title           *string `json:"title"`
		MessageHTML     *string `json:"message_html"`
		LogoURL         *string `json:"logo_url"`
		BackgroundColor *string `json:"background_color"`
		TextColor       *string `json:"text_color"`
		AccentColor     *string `json:"accent_color"`
		ShowContactInfo *bool   `json:"show_contact_info"`
		ContactEmail    *string `json:"contact_email"`
		ShowSocialLinks *bool   `json:"show_social_links"`
		IsDefault       *bool   `json:"is_default"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	template, err := h.maintenanceRepo.GetTemplateByID(r.Context(), templateID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get template")
		apierror.WriteError(w, apierror.NewInternal("Failed to get template"))
		return
	}

	if template == nil {
		apierror.WriteError(w, apierror.NewNotFound("Template not found"))
		return
	}

	// Update fields
	if req.Name != nil {
		template.Name = *req.Name
	}
	if req.Title != nil {
		template.Title = req.Title
	}
	if req.MessageHTML != nil {
		template.MessageHTML = req.MessageHTML
	}
	if req.LogoURL != nil {
		template.LogoURL = req.LogoURL
	}
	if req.BackgroundColor != nil {
		template.BackgroundColor = *req.BackgroundColor
	}
	if req.TextColor != nil {
		template.TextColor = *req.TextColor
	}
	if req.AccentColor != nil {
		template.AccentColor = *req.AccentColor
	}
	if req.ShowContactInfo != nil {
		template.ShowContactInfo = *req.ShowContactInfo
	}
	if req.ContactEmail != nil {
		template.ContactEmail = req.ContactEmail
	}
	if req.ShowSocialLinks != nil {
		template.ShowSocialLinks = *req.ShowSocialLinks
	}
	if req.IsDefault != nil {
		template.IsDefault = *req.IsDefault
	}

	err = h.maintenanceRepo.UpdateTemplate(r.Context(), template)
	if err != nil {
		logrus.WithError(err).Error("Failed to update template")
		apierror.WriteError(w, apierror.NewInternal("Failed to update template"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"template": template,
	})
}

// HandleDeleteTemplate handles DELETE /v1/admin/maintenance/templates/{id}
func (h *MaintenanceHandler) HandleDeleteTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid template ID"))
		return
	}

	err = h.maintenanceRepo.DeleteTemplate(r.Context(), templateID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete template")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete template"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Template deleted",
	})
}

// HandleGetScheduledMaintenance handles GET /v1/admin/maintenance/schedule
func (h *MaintenanceHandler) HandleGetScheduledMaintenance(w http.ResponseWriter, r *http.Request) {
	maintenances, err := h.maintenanceRepo.GetScheduledMaintenance(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get scheduled maintenance")
		apierror.WriteError(w, apierror.NewInternal("Failed to get scheduled maintenance"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scheduled": maintenances,
	})
}

// HandleGetAuditLog handles GET /v1/admin/maintenance/audit
func (h *MaintenanceHandler) HandleGetAuditLog(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	logs, err := h.maintenanceRepo.GetAuditLog(r.Context(), limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get audit log")
		apierror.WriteError(w, apierror.NewInternal("Failed to get audit log"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"audit_log": logs,
	})
}

// HandleGetPublicStatus handles GET /maintenance/status (public endpoint)
func (h *MaintenanceHandler) HandleGetPublicStatus(w http.ResponseWriter, r *http.Request) {
	// Check if maintenance is enabled
	maintenance, err := h.maintenanceRepo.GetEnabledMaintenance(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get maintenance status")
		// Return safe default on error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.MaintenanceStatus{
			MaintenanceMode: false,
		})
		return
	}

	// Get scheduled maintenance
	scheduled, _ := h.maintenanceRepo.GetScheduledMaintenance(r.Context())

	var scheduledMaintenance []types.ScheduledMaintenance
	for _, m := range scheduled {
		if m.ScheduledStart != nil {
			scheduledMaintenance = append(scheduledMaintenance, types.ScheduledMaintenance{
				Name:           m.Name,
				ScheduledStart: *m.ScheduledStart,
				ScheduledEnd:   *m.ScheduledEnd,
				Status:         "scheduled",
			})
		}
	}

	status := types.MaintenanceStatus{
		MaintenanceMode:      maintenance != nil && maintenance.Enabled,
		ScheduledMaintenance: scheduledMaintenance,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Helper to get user ID from request context
func (h *MaintenanceHandler) getUserID(r *http.Request) *uuid.UUID {
	// Try to get from context - this depends on auth middleware
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			return &userID
		}
	}
	return nil
}

// Helper to create audit log
func (h *MaintenanceHandler) createAuditLog(r *http.Request, maintenance *types.PlatformMaintenance, action string, changes map[string]interface{}, userID *uuid.UUID) {
	oldJSON, _ := json.Marshal(changes)
	ip := middleware.GetRealIP(r)
	userAgent := r.UserAgent()

	auditLog := &types.MaintenanceAuditLog{
		Action:    action,
		OldValues: stringPtr(string(oldJSON)),
		ChangedBy: userID,
		IPAddress: &ip,
		UserAgent: &userAgent,
	}

	_ = h.maintenanceRepo.CreateAuditLog(r.Context(), auditLog)
}

// stringPtr helper
func stringPtr(s string) *string {
	return &s
}
