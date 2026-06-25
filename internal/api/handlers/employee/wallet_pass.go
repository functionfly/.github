package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
)

type generatePassFileRequest struct {
	Name         string                   `json:"name"`
	PassType     string                   `json:"pass_type"`
	Platform     string                   `json:"platform"`
	TemplateData map[string]interface{}   `json:"template_data,omitempty"`
}

func (h *Handler) HandleGeneratePassFile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req generatePassFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Name == "" || req.PassType == "" || req.Platform == "" {
		apierror.WriteError(w, apierror.NewBadRequest("name, pass_type, and platform are required"))
		return
	}

	tmpl := &storage.WalletPassTemplate{
		TenantID:  claims.TenantID,
		Name:      req.Name,
		PassType:  req.PassType,
		Platform:  req.Platform,
		IsActive:  true,
	}
	if req.TemplateData != nil {
		tmpl.TemplateData = storage.JSONMap(req.TemplateData)
	}

	created, err := h.repo.CreateWalletPassTemplate(r.Context(), tmpl)
	if err != nil {
		h.log.WithError(err).Error("Failed to generate pass template")
		apierror.WriteError(w, apierror.NewInternal("Failed to generate pass template"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"template": created,
	})
}

func (h *Handler) HandleListPassTemplates(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListWalletPassTemplatesOpts{
		Limit:  20,
		Offset: 0,
	}
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			opts.Limit = n
		}
	}
	if o := q.Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			opts.Offset = n
		}
	}
	if p := q.Get("platform"); p != "" {
		opts.Platform = &p
	}
	if a := q.Get("is_active"); a != "" {
		active := a == "true"
		opts.IsActive = &active
	}

	templates, total, err := h.repo.ListWalletPassTemplates(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list pass templates")
		apierror.WriteError(w, apierror.NewInternal("Failed to list pass templates"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": templates,
		"total":     total,
		"limit":     opts.Limit,
		"offset":    opts.Offset,
	})
}
