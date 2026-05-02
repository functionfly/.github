package github

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/storage"
)

// HandleListTemplates lists all import templates for the tenant.
func (h *Handler) HandleListTemplates(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	templates, err := h.githubRepo.ListTemplatesByTenant(r.Context(), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list templates")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to list templates")
		return
	}

	results := make([]*TemplateResponse, len(templates))
	for i, tmpl := range templates {
		results[i] = h.mapTemplateResponse(tmpl)
	}

	h.respondJSON(w, http.StatusOK, results)
}

// HandleCreateTemplate creates a new import template.
func (h *Handler) HandleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	var req TemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if req.Name == "" {
		h.respondError(w, http.StatusBadRequest, "missing_name", "Template name is required")
		return
	}
	if req.Config == nil {
		req.Config = json.RawMessage(`{}`)
	}
	if req.DetectionRules == nil {
		req.DetectionRules = json.RawMessage(`{}`)
	}

	tmpl := &storage.GitHubImportTemplate{
		TenantID:       claims.TenantID,
		UserID:         claims.UserID,
		Name:           req.Name,
		Description:    req.Description,
		Config:         req.Config,
		DetectionRules: req.DetectionRules,
		IsDefault:      req.IsDefault,
	}

	created, err := h.githubRepo.CreateTemplate(r.Context(), tmpl)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create template")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to create template")
		return
	}

	h.respondJSON(w, http.StatusCreated, h.mapTemplateResponse(created))
}

// HandleUpdateTemplate updates an existing import template.
func (h *Handler) HandleUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	tmplID, err := h.parseUUID(r, "id")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid template ID")
		return
	}

	existing, err := h.githubRepo.GetTemplateByID(r.Context(), tmplID)
	if err != nil || existing == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "Template not found")
		return
	}
	if existing.TenantID != claims.TenantID {
		h.respondError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	var req TemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.Config != nil {
		updates["config"] = req.Config
	}
	if req.DetectionRules != nil {
		updates["detection_rules"] = req.DetectionRules
	}
	updates["is_default"] = req.IsDefault

	if err := h.githubRepo.UpdateTemplate(r.Context(), tmplID, updates); err != nil {
		h.logger.WithError(err).Error("Failed to update template")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to update template")
		return
	}

	updated, _ := h.githubRepo.GetTemplateByID(r.Context(), tmplID)
	h.respondJSON(w, http.StatusOK, h.mapTemplateResponse(updated))
}

// HandleDeleteTemplate deletes an import template.
func (h *Handler) HandleDeleteTemplate(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	tmplID, err := h.parseUUID(r, "id")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid template ID")
		return
	}

	existing, err := h.githubRepo.GetTemplateByID(r.Context(), tmplID)
	if err != nil || existing == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "Template not found")
		return
	}
	if existing.TenantID != claims.TenantID {
		h.respondError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	if err := h.githubRepo.DeleteTemplate(r.Context(), tmplID); err != nil {
		h.logger.WithError(err).Error("Failed to delete template")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to delete template")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) mapTemplateResponse(tmpl *storage.GitHubImportTemplate) *TemplateResponse {
	return &TemplateResponse{
		ID:             tmpl.ID,
		Name:           tmpl.Name,
		Description:    tmpl.Description,
		Config:         tmpl.Config,
		DetectionRules: tmpl.DetectionRules,
		IsDefault:      tmpl.IsDefault,
		UsageCount:     tmpl.UsageCount,
		CreatedAt:      tmpl.CreatedAt,
		UpdatedAt:      tmpl.UpdatedAt,
	}
}
