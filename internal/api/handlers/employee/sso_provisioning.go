package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListSSOConfigs(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListSSOProvisioningConfigsOpts{
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
	if p := q.Get("provider"); p != "" {
		opts.Provider = &p
	}
	if ia := q.Get("is_active"); ia != "" {
		active := ia == "true"
		opts.IsActive = &active
	}

	configs, total, err := h.repo.ListSSOProvisioningConfigs(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list SSO configs")
		apierror.WriteError(w, apierror.NewInternal("Failed to list SSO configs"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"configs": configs,
		"total":   total,
		"limit":   opts.Limit,
		"offset":  opts.Offset,
	})
}

type createSSOConfigRequest struct {
	Provider            string                 `json:"provider"`
	ProviderURL         *string                `json:"provider_url,omitempty"`
	ClientID            *string                `json:"client_id,omitempty"`
	ClientSecret        *string                `json:"client_secret,omitempty"`
	SCIMEndpoint        *string                `json:"scim_endpoint,omitempty"`
	SCIMToken           *string                `json:"scim_token,omitempty"`
	AutoCreateEmployee  *bool                  `json:"auto_create_employee,omitempty"`
	AutoUpdateEmployee  *bool                  `json:"auto_update_employee,omitempty"`
	AutoDeactivate      *bool                  `json:"auto_deactivate,omitempty"`
	DefaultDepartmentID *int64                 `json:"default_department_id,omitempty"`
	DefaultClearance    string                 `json:"default_clearance,omitempty"`
	FieldMappings       map[string]interface{} `json:"field_mappings,omitempty"`
}

func (h *Handler) HandleCreateSSOConfig(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req createSSOConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Provider == "" {
		apierror.WriteError(w, apierror.NewBadRequest("provider is required"))
		return
	}

	cfg := &storage.SSOProvisioningConfig{
		TenantID:            claims.TenantID,
		Provider:            req.Provider,
		ProviderURL:         req.ProviderURL,
		ClientID:            req.ClientID,
		SCIMEndpoint:        req.SCIMEndpoint,
		AutoCreateEmployee:  true,
		AutoUpdateEmployee:  true,
		DefaultClearance:    "standard",
		IsActive:            true,
		DefaultDepartmentID: req.DefaultDepartmentID,
	}
	if req.ClientSecret != nil {
		cfg.ClientSecretEncrypted = req.ClientSecret
	}
	if req.SCIMToken != nil {
		cfg.SCIMTokenEncrypted = req.SCIMToken
	}
	if req.AutoCreateEmployee != nil {
		cfg.AutoCreateEmployee = *req.AutoCreateEmployee
	}
	if req.AutoUpdateEmployee != nil {
		cfg.AutoUpdateEmployee = *req.AutoUpdateEmployee
	}
	if req.AutoDeactivate != nil {
		cfg.AutoDeactivate = *req.AutoDeactivate
	}
	if req.DefaultClearance != "" {
		cfg.DefaultClearance = req.DefaultClearance
	}
	if req.FieldMappings != nil {
		cfg.FieldMappings = storage.JSONMap(req.FieldMappings)
	}

	created, err := h.repo.CreateSSOProvisioningConfig(r.Context(), cfg)
	if err != nil {
		h.log.WithError(err).Error("Failed to create SSO config")
		apierror.WriteError(w, apierror.NewInternal("Failed to create SSO config"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"config": created,
	})
}

func (h *Handler) HandleSyncSSO(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	configID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid config ID"))
		return
	}

	cfg, err := h.repo.GetSSOProvisioningConfigByID(r.Context(), configID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get SSO config")
		apierror.WriteError(w, apierror.NewInternal("Failed to get SSO config"))
		return
	}
	if cfg == nil {
		apierror.WriteError(w, apierror.NewNotFound("SSO config not found"))
		return
	}

	h.repo.UpdateSSOProvisioningConfig(r.Context(), configID, map[string]interface{}{
		"last_sync_at": nil,
	})

	log := &storage.SSOProvisioningLog{
		ConfigID: configID,
		Action:   "sync_started",
		Details:  storage.JSONMap{"triggered_by": claims.UserID.String()},
	}
	h.repo.CreateSSOProvisioningLog(r.Context(), log)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"config_id": configID,
		"message":  "SSO sync initiated",
	})
}

func (h *Handler) HandleGetProvisioningLogs(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	configID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid config ID"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListSSOProvisioningLogsOpts{
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
	if a := q.Get("action"); a != "" {
		opts.Action = &a
	}

	logs, total, err := h.repo.ListSSOProvisioningLogs(r.Context(), configID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list provisioning logs")
		apierror.WriteError(w, apierror.NewInternal("Failed to list provisioning logs"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":   logs,
		"total":  total,
		"limit":  opts.Limit,
		"offset": opts.Offset,
	})
}
