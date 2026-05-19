package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo HandlerRepo
}

type HandlerRepo interface {
	List(ctx context.Context, params ListPluginsParams) ([]Plugin, error)
	Get(ctx context.Context, tenantID, pluginID string) (*Plugin, error)
	Create(ctx context.Context, plugin *Plugin) error
	Update(ctx context.Context, plugin *Plugin) error
	Delete(ctx context.Context, tenantID, pluginID string) error
	SetStatus(ctx context.Context, tenantID, pluginID string, status PluginStatus) error
	SetError(ctx context.Context, tenantID, pluginID string, errMsg string) error
	UpdateConfig(ctx context.Context, tenantID, pluginID string, config map[string]string) error
	GetEnabledByType(ctx context.Context, tenantID string, pluginType PluginType) (*Plugin, error)
	GetSandbox(ctx context.Context, pluginID string) (*PluginSandbox, error)
	UpsertSandbox(ctx context.Context, sandbox *PluginSandbox) error
	ListPermissions(ctx context.Context, pluginID string) ([]PluginPermission, error)
	SetPermission(ctx context.Context, perm *PluginPermission) error
	CreateVersion(ctx context.Context, version *PluginVersion) error
	ListVersions(ctx context.Context, pluginID string) ([]PluginVersion, error)
	GetPreviousVersion(ctx context.Context, pluginID, currentVersion string) (*PluginVersion, error)
}

func NewHandler(repo HandlerRepo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) HandleListPlugins(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	var pluginTypePtr *PluginType
	if pluginType != "" {
		pt := PluginType(pluginType)
		pluginTypePtr = &pt
	}

	var statusPtr *PluginStatus
	if status != "" {
		s := PluginStatus(status)
		if s != PluginStatusEnabled && s != PluginStatusDisabled && s != PluginStatusError && s != PluginStatusPaused {
			writeJSONError(w, http.StatusBadRequest, "Invalid status")
			return
		}
		statusPtr = &s
	}

	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	params := ListPluginsParams{
		TenantID:   tenantID,
		PluginType: pluginTypePtr,
		Status:     statusPtr,
		Category:   categoryPtr,
		Search:     searchPtr,
		Limit:      limit,
		Offset:     offset,
	}

	plugins, err := h.repo.List(r.Context(), params)
	if err != nil {
		logrus.WithError(err).Warn("plugins: failed to list plugins")
		writeJSON(w, http.StatusOK, map[string]interface{}{"plugins": []Plugin{}})
		return
	}

	if plugins == nil {
		plugins = []Plugin{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"plugins": plugins})
}

func (h *Handler) HandleGetPlugin(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginID := mux.Vars(r)["id"]
	if pluginID == "" {
		writeJSONError(w, http.StatusBadRequest, "plugin id is required")
		return
	}

	plugin, err := h.repo.Get(r.Context(), tenantID, pluginID)
	if err != nil {
		logrus.WithError(err).Error("plugins: failed to get plugin")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get plugin")
		return
	}
	if plugin == nil {
		writeJSONError(w, http.StatusNotFound, "Plugin not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"plugin": plugin})
}

type InstallPluginRequest struct {
	Manifest        map[string]interface{} `json:"manifest"`
	PluginType      PluginType             `json:"plugin_type"`
	Name            string                  `json:"name"`
	Version         string                  `json:"version"`
	Description     string                  `json:"description"`
	AuthorName      string                  `json:"author_name"`
	AuthorEmail     string                  `json:"author_email"`
	AuthorWebsite   string                  `json:"author_website"`
	Category        string                  `json:"category"`
	IconURL         string                  `json:"icon_url"`
	HomepageURL     string                  `json:"homepage_url"`
	RepositoryURL   string                  `json:"repository_url"`
	License         string                  `json:"license"`
	SizeBytes       int                     `json:"size_bytes"`
	Signature       string                  `json:"signature"`
	SandboxTier     SandboxTier             `json:"sandbox_tier"`
	SandboxConfig   *SandboxConfig          `json:"sandbox_config"`
}

type SandboxConfig struct {
	CPULimit       float64   `json:"cpu_limit"`
	MemoryLimitMB  int       `json:"memory_limit_mb"`
	TimeoutSeconds int       `json:"timeout_seconds"`
	AllowedDomains []string  `json:"allowed_domains"`
	BlockedDomains []string  `json:"blocked_domains"`
}

func (h *Handler) HandleInstallPlugin(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req InstallPluginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Version == "" {
		writeJSONError(w, http.StatusBadRequest, "name and version are required")
		return
	}

	plugin := &Plugin{
		TenantID:      tenantID,
		Manifest:      req.Manifest,
		PluginType:    req.PluginType,
		Name:          req.Name,
		Version:       req.Version,
		Description:   req.Description,
		AuthorName:    req.AuthorName,
		AuthorEmail:   req.AuthorEmail,
		AuthorWebsite: req.AuthorWebsite,
		Category:      req.Category,
		Status:        PluginStatusDisabled,
		IconURL:       req.IconURL,
		HomepageURL:   req.HomepageURL,
		RepositoryURL: req.RepositoryURL,
		License:       req.License,
		SizeBytes:     req.SizeBytes,
		Signature:     req.Signature,
	}

	if err := h.repo.Create(r.Context(), plugin); err != nil {
		logrus.WithError(err).Error("plugins: failed to install plugin")
		writeJSONError(w, http.StatusInternalServerError, "Failed to install plugin")
		return
	}

	if req.SandboxTier != "" || req.SandboxConfig != nil {
		sandbox := &PluginSandbox{
			PluginID:        plugin.ID,
			Tier:            req.SandboxTier,
			CPULimit:        0.5,
			MemoryLimitMB:   256,
			TimeoutSeconds:  30,
		}
		if req.SandboxConfig != nil {
			sandbox.CPULimit = req.SandboxConfig.CPULimit
			sandbox.MemoryLimitMB = req.SandboxConfig.MemoryLimitMB
			sandbox.TimeoutSeconds = req.SandboxConfig.TimeoutSeconds
			sandbox.AllowedDomains = req.SandboxConfig.AllowedDomains
			sandbox.BlockedDomains = req.SandboxConfig.BlockedDomains
		}
		if sandbox.Tier == "" {
			sandbox.Tier = SandboxTierWorker
		}
		_ = h.repo.UpsertSandbox(r.Context(), sandbox)
	}

	version := &PluginVersion{
		PluginID:  plugin.ID,
		Version:   plugin.Version,
		Manifest:  req.Manifest,
		SizeBytes: req.SizeBytes,
		Signature: req.Signature,
	}
	_ = h.repo.CreateVersion(r.Context(), version)

	writeJSON(w, http.StatusCreated, map[string]interface{}{"plugin": plugin})
}

func (h *Handler) HandleUpdatePlugin(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginID := mux.Vars(r)["id"]
	if pluginID == "" {
		writeJSONError(w, http.StatusBadRequest, "plugin id is required")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	plugin, err := h.repo.Get(r.Context(), tenantID, pluginID)
	if err != nil {
		logrus.WithError(err).Error("plugins: failed to get plugin")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update plugin")
		return
	}
	if plugin == nil {
		writeJSONError(w, http.StatusNotFound, "Plugin not found")
		return
	}

	if v, ok := req["manifest"].(map[string]interface{}); ok {
		plugin.Manifest = v
	}
	if v, ok := req["version"].(string); ok {
		plugin.Version = v
	}
	if v, ok := req["description"].(string); ok {
		plugin.Description = v
	}
	if v, ok := req["category"].(string); ok {
		plugin.Category = v
	}
	if v, ok := req["icon_url"].(string); ok {
		plugin.IconURL = v
	}

	if err := h.repo.Update(r.Context(), plugin); err != nil {
		logrus.WithError(err).Error("plugins: failed to update plugin")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update plugin")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"plugin": plugin})
}

func (h *Handler) HandleUninstallPlugin(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginID := mux.Vars(r)["id"]
	if pluginID == "" {
		writeJSONError(w, http.StatusBadRequest, "plugin id is required")
		return
	}

	if err := h.repo.Delete(r.Context(), tenantID, pluginID); err != nil {
		logrus.WithError(err).Error("plugins: failed to uninstall plugin")
		writeJSONError(w, http.StatusInternalServerError, "Failed to uninstall plugin")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Plugin uninstalled"})
}

func (h *Handler) HandleEnablePlugin(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginID := mux.Vars(r)["id"]
	if pluginID == "" {
		writeJSONError(w, http.StatusBadRequest, "plugin id is required")
		return
	}

	if err := h.repo.SetStatus(r.Context(), tenantID, pluginID, PluginStatusEnabled); err != nil {
		logrus.WithError(err).Error("plugins: failed to enable plugin")
		writeJSONError(w, http.StatusInternalServerError, "Failed to enable plugin")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Plugin enabled"})
}

func (h *Handler) HandleDisablePlugin(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginID := mux.Vars(r)["id"]
	if pluginID == "" {
		writeJSONError(w, http.StatusBadRequest, "plugin id is required")
		return
	}

	if err := h.repo.SetStatus(r.Context(), tenantID, pluginID, PluginStatusDisabled); err != nil {
		logrus.WithError(err).Error("plugins: failed to disable plugin")
		writeJSONError(w, http.StatusInternalServerError, "Failed to disable plugin")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Plugin disabled"})
}

func (h *Handler) HandlePausePlugin(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginID := mux.Vars(r)["id"]
	if pluginID == "" {
		writeJSONError(w, http.StatusBadRequest, "plugin id is required")
		return
	}

	if err := h.repo.SetStatus(r.Context(), tenantID, pluginID, PluginStatusPaused); err != nil {
		logrus.WithError(err).Error("plugins: failed to pause plugin")
		writeJSONError(w, http.StatusInternalServerError, "Failed to pause plugin")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Plugin paused"})
}

func (h *Handler) HandleRollbackPlugin(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginID := mux.Vars(r)["id"]
	if pluginID == "" {
		writeJSONError(w, http.StatusBadRequest, "plugin id is required")
		return
	}

	plugin, err := h.repo.Get(r.Context(), tenantID, pluginID)
	if err != nil {
		logrus.WithError(err).Error("plugins: failed to get plugin for rollback")
		writeJSONError(w, http.StatusInternalServerError, "Failed to rollback plugin")
		return
	}
	if plugin == nil {
		writeJSONError(w, http.StatusNotFound, "Plugin not found")
		return
	}

	prevVersion, err := h.repo.GetPreviousVersion(r.Context(), plugin.ID, plugin.Version)
	if err != nil {
		logrus.WithError(err).Error("plugins: failed to get previous version for rollback")
		writeJSONError(w, http.StatusInternalServerError, "Failed to rollback plugin")
		return
	}
	if prevVersion == nil {
		writeJSONError(w, http.StatusBadRequest, "No previous version available for rollback")
		return
	}

	plugin.Version = prevVersion.Version
	plugin.Manifest = prevVersion.Manifest
	if err := h.repo.Update(r.Context(), plugin); err != nil {
		logrus.WithError(err).Error("plugins: failed to rollback plugin")
		writeJSONError(w, http.StatusInternalServerError, "Failed to rollback plugin")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"plugin": plugin, "rolled_back_to": prevVersion.Version})
}

func (h *Handler) HandleGetSandbox(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginID := mux.Vars(r)["id"]
	if pluginID == "" {
		writeJSONError(w, http.StatusBadRequest, "plugin id is required")
		return
	}

	sandbox, err := h.repo.GetSandbox(r.Context(), pluginID)
	if err != nil {
		logrus.WithError(err).Error("plugins: failed to get sandbox")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get sandbox")
		return
	}
	if sandbox == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"sandbox": nil})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"sandbox": sandbox})
}

type UpdateSandboxRequest struct {
	Tier           SandboxTier `json:"tier"`
	CPULimit       float64     `json:"cpu_limit"`
	MemoryLimitMB  int         `json:"memory_limit_mb"`
	TimeoutSeconds int         `json:"timeout_seconds"`
	AllowedDomains []string    `json:"allowed_domains"`
	BlockedDomains []string    `json:"blocked_domains"`
}

func (h *Handler) HandleUpdateSandbox(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginID := mux.Vars(r)["id"]
	if pluginID == "" {
		writeJSONError(w, http.StatusBadRequest, "plugin id is required")
		return
	}

	var req UpdateSandboxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	existingSandbox, err := h.repo.GetSandbox(r.Context(), pluginID)
	if err != nil {
		logrus.WithError(err).Error("plugins: failed to get existing sandbox")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update sandbox")
		return
	}

	sandbox := &PluginSandbox{
		PluginID: pluginID,
	}
	if existingSandbox != nil {
		sandbox.ID = existingSandbox.ID
	}

	if req.Tier != "" {
		sandbox.Tier = req.Tier
	} else if existingSandbox != nil {
		sandbox.Tier = existingSandbox.Tier
	} else {
		sandbox.Tier = SandboxTierWorker
	}

	if req.CPULimit > 0 {
		sandbox.CPULimit = req.CPULimit
	} else if existingSandbox != nil {
		sandbox.CPULimit = existingSandbox.CPULimit
	}

	if req.MemoryLimitMB > 0 {
		sandbox.MemoryLimitMB = req.MemoryLimitMB
	} else if existingSandbox != nil {
		sandbox.MemoryLimitMB = existingSandbox.MemoryLimitMB
	}

	if req.TimeoutSeconds > 0 {
		sandbox.TimeoutSeconds = req.TimeoutSeconds
	} else if existingSandbox != nil {
		sandbox.TimeoutSeconds = existingSandbox.TimeoutSeconds
	}

	sandbox.AllowedDomains = req.AllowedDomains
	sandbox.BlockedDomains = req.BlockedDomains

	if err := h.repo.UpsertSandbox(r.Context(), sandbox); err != nil {
		logrus.WithError(err).Error("plugins: failed to update sandbox")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update sandbox")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"sandbox": sandbox})
}

func (h *Handler) HandleGetPermissions(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginID := mux.Vars(r)["id"]
	if pluginID == "" {
		writeJSONError(w, http.StatusBadRequest, "plugin id is required")
		return
	}

	permissions, err := h.repo.ListPermissions(r.Context(), pluginID)
	if err != nil {
		logrus.WithError(err).Error("plugins: failed to list permissions")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list permissions")
		return
	}

	if permissions == nil {
		permissions = []PluginPermission{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"permissions": permissions})
}

type SetPermissionRequest struct {
	PermissionType   string `json:"permission_type"`
	PermissionAction string `json:"permission_action"`
	Resource         string `json:"resource"`
	Granted          bool   `json:"granted"`
	ExpiresAt        string `json:"expires_at"`
}

func (h *Handler) HandleSetPermission(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginID := mux.Vars(r)["id"]
	if pluginID == "" {
		writeJSONError(w, http.StatusBadRequest, "plugin id is required")
		return
	}

	var req SetPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	perm := &PluginPermission{
		PluginID:         pluginID,
		PermissionType:  req.PermissionType,
		PermissionAction: req.PermissionAction,
		Resource:        req.Resource,
		Granted:         req.Granted,
	}

	if req.ExpiresAt != "" {
		expiredAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err == nil {
			perm.ExpiresAt = &expiredAt
		}
	}

	if err := h.repo.SetPermission(r.Context(), perm); err != nil {
		logrus.WithError(err).Error("plugins: failed to set permission")
		writeJSONError(w, http.StatusInternalServerError, "Failed to set permission")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Permission updated"})
}

func (h *Handler) HandleListVersions(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginID := mux.Vars(r)["id"]
	if pluginID == "" {
		writeJSONError(w, http.StatusBadRequest, "plugin id is required")
		return
	}

	versions, err := h.repo.ListVersions(r.Context(), pluginID)
	if err != nil {
		logrus.WithError(err).Error("plugins: failed to list versions")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list versions")
		return
	}

	if versions == nil {
		versions = []PluginVersion{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"versions": versions})
}

func (h *Handler) HandleConfigurePlugin(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginID := mux.Vars(r)["id"]
	if pluginID == "" {
		writeJSONError(w, http.StatusBadRequest, "plugin id is required")
		return
	}

	var req struct {
		Config map[string]string `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.repo.UpdateConfig(r.Context(), tenantID, pluginID, req.Config); err != nil {
		logrus.WithError(err).Error("plugins: failed to configure plugin")
		writeJSONError(w, http.StatusInternalServerError, "Failed to configure plugin")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Plugin configured"})
}

func (h *Handler) HandleSetError(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pluginID := mux.Vars(r)["id"]
	if pluginID == "" {
		writeJSONError(w, http.StatusBadRequest, "plugin id is required")
		return
	}

	var req struct {
		Error string `json:"error"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.repo.SetError(r.Context(), tenantID, pluginID, req.Error); err != nil {
		logrus.WithError(err).Error("plugins: failed to set error")
		writeJSONError(w, http.StatusInternalServerError, "Failed to set error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Error updated"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func getTenantID(r *http.Request) string {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		return ""
	}
	return claims.TenantID.String()
}

type ListPluginsParams struct {
	TenantID   string
	PluginType *PluginType
	Status     *PluginStatus
	Category   *string
	Search     *string
	Limit      int
	Offset     int
}

type Plugin struct {
	ID            string                 `json:"id"`
	TenantID      string                 `json:"tenant_id"`
	Manifest      map[string]interface{} `json:"manifest"`
	PluginType    PluginType             `json:"plugin_type"`
	Name          string                 `json:"name"`
	Version       string                 `json:"version"`
	Description   string                 `json:"description"`
	AuthorName    string                 `json:"author_name"`
	AuthorEmail   string                 `json:"author_email"`
	AuthorWebsite string                 `json:"author_website"`
	Category      string                 `json:"category"`
	Status        PluginStatus           `json:"status"`
	IconURL       string                 `json:"icon_url"`
	HomepageURL   string                 `json:"homepage_url"`
	RepositoryURL string                 `json:"repository_url"`
	License       string                 `json:"license"`
	SizeBytes     int                    `json:"size_bytes"`
	Signature     string                 `json:"signature"`
	Verified      bool                   `json:"verified"`
	Config        map[string]string      `json:"config"`
	Metadata      map[string]interface{} `json:"metadata"`
	InstalledAt   time.Time              `json:"installed_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	EnabledAt     *time.Time             `json:"enabled_at,omitempty"`
	ErrorMessage  *string                `json:"error_message,omitempty"`
}

type PluginStatus string

const (
	PluginStatusEnabled  PluginStatus = "enabled"
	PluginStatusDisabled PluginStatus = "disabled"
	PluginStatusError    PluginStatus = "error"
	PluginStatusPaused   PluginStatus = "paused"
)

type PluginType string

const (
	PluginTypeUI            PluginType = "ui"
	PluginTypeGraph         PluginType = "graph"
	PluginTypeAITool        PluginType = "ai_tool"
	PluginTypeRuntime       PluginType = "runtime"
	PluginTypeInfrastructure PluginType = "infrastructure"
	PluginTypeMarketplace   PluginType = "marketplace"
)

type SandboxTier string

const (
	SandboxTierWASM       SandboxTier = "wasm"
	SandboxTierWorker     SandboxTier = "worker"
	SandboxTierMicroVM    SandboxTier = "microvm"
	SandboxTierEnterprise SandboxTier = "enterprise"
)

type PluginVersion struct {
	ID         string
	PluginID   string
	Version    string
	Changelog  string
	Manifest   map[string]interface{}
	SizeBytes  int
	Signature  string
	ReleaseAt  time.Time
	CreatedAt  time.Time
}

type PluginSandbox struct {
	ID              string
	PluginID        string
	Tier            SandboxTier
	CPULimit        float64
	MemoryLimitMB   int
	TimeoutSeconds  int
	NetworkIsolated bool
	FilesystemScope string
	MaxInstances    int
	EnvVars         map[string]string
	AllowedDomains  []string
	BlockedDomains  []string
	RateLimitRPM    *int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type PluginPermission struct {
	ID               string
	PluginID         string
	PermissionType  string
	PermissionAction string
	Resource         string
	Granted          bool
	GrantedAt        *time.Time
	GrantedBy        *string
	ExpiresAt        *time.Time
	CreatedAt        time.Time
}

type RateLimitCheckRequest struct {
	IP        string `json:"ip"`
	Endpoint  string `json:"endpoint"`
	UserID    string `json:"user_id"`
	Limit     int    `json:"limit"`
	WindowSec int    `json:"window_sec"`
}

type RateLimitCheckResponse struct {
	Allowed   bool   `json:"allowed"`
	Remaining int    `json:"remaining"`
	ResetAt   int64  `json:"reset_at"`
	Limit     int    `json:"limit"`
}

func (h *Handler) HandleCheckRateLimit(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rateLimiterPlugin, err := h.repo.GetEnabledByType(r.Context(), tenantID, PluginTypeInfrastructure)
	if err != nil {
		logrus.WithError(err).Error("plugins: failed to check rate limiter plugin")
		writeJSONError(w, http.StatusInternalServerError, "Failed to check rate limiter")
		return
	}

	if rateLimiterPlugin == nil || rateLimiterPlugin.Name != "Rate Limiter" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"enabled": false,
			"message": "Rate limiter plugin not enabled",
		})
		return
	}

	var req RateLimitCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.WindowSec <= 0 {
		req.WindowSec = 60
	}

	_ = fmt.Sprintf("ratelimit:%s:%s:%s", tenantID, req.Endpoint, req.IP)
	remaining := req.Limit - 1
	resetAt := time.Now().Add(time.Duration(req.WindowSec) * time.Second).Unix()

	writeJSON(w, http.StatusOK, RateLimitCheckResponse{
		Allowed:   true,
		Remaining: remaining,
		ResetAt:   resetAt,
		Limit:     req.Limit,
	})
}