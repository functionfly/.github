package studio

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// extensionDeprecationSunset is the date the extension_registry API
// is removed. Clients should migrate to /plugins before this date.
// See docs/PLUGIN_MIGRATION.md.
var extensionDeprecationSunset = time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)

// markExtensionDeprecation sets the deprecation response headers on
// every response from the legacy extension_registry endpoints and
// logs a one-line warning per call so we can track laggards in
// staging. The Sunset header follows RFC 8594.
func markExtensionDeprecation(w http.ResponseWriter, route string) {
	w.Header().Set("Deprecation", "true")
	w.Header().Set("Sunset", extensionDeprecationSunset.Format(http.TimeFormat))
	w.Header().Set("X-Deprecated-Use", "/plugins")
	w.Header().Set("Link", `</docs/PLUGIN_MIGRATION.md>; rel="deprecation"`)
	logrus.WithFields(logrus.Fields{
		"route":  route,
		"sunset": extensionDeprecationSunset.Format(time.RFC3339),
	}).Warn("extension_registry: deprecated endpoint called; clients must migrate to /plugins")
}

// ExtensionsHandler handles studio extension HTTP requests
type ExtensionsHandler struct {
	extRepo *ExtensionRepository
}

// NewExtensionsHandler creates a new extensions handler
func NewExtensionsHandler(extRepo *ExtensionRepository) *ExtensionsHandler {
	return &ExtensionsHandler{extRepo: extRepo}
}

// HandleListExtensions handles GET /v1/extensions
func (h *ExtensionsHandler) HandleListExtensions(w http.ResponseWriter, r *http.Request) {
	markExtensionDeprecation(w, "GET /v1/extensions")
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	category := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")
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

	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	var statusPtr *ExtensionStatus
	if status != "" {
		s := ExtensionStatus(status)
		if s != ExtensionStatusEnabled && s != ExtensionStatusDisabled && s != ExtensionStatusError {
			writeJSONError(w, http.StatusBadRequest, "Invalid status")
			return
		}
		statusPtr = &s
	}

	params := ListExtensionsParams{
		TenantID: tenantID,
		Category: categoryPtr,
		Status:   statusPtr,
		Limit:    limit,
		Offset:   offset,
	}

	extensions, err := h.extRepo.ListExtensions(r.Context(), params)
	if err != nil {
		logrus.WithError(err).Warn("extensions: failed to list extensions")
		writeJSON(w, http.StatusOK, map[string]interface{}{"extensions": []Extension{}})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"extensions": extensions})
}

// HandleInstallExtension handles POST /v1/extensions/{id}/install
func (h *ExtensionsHandler) HandleInstallExtension(w http.ResponseWriter, r *http.Request) {
	markExtensionDeprecation(w, "POST /v1/extensions/{id}/install")
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Name        string   `json:"name"`
		Version     string   `json:"version"`
		Description string   `json:"description"`
		AuthorName  string   `json:"author_name"`
		Category    string   `json:"category"`
		Permissions []string `json:"permissions"`
		Hooks       []string `json:"hooks"`
		SizeKB      int      `json:"size_kb"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Version == "" {
		writeJSONError(w, http.StatusBadRequest, "name and version are required")
		return
	}

	ext := &Extension{
		TenantID:    tenantID,
		Name:        req.Name,
		Version:     req.Version,
		Description: req.Description,
		AuthorName:  req.AuthorName,
		Category:    req.Category,
		Status:      ExtensionStatusDisabled,
		Permissions: req.Permissions,
		Hooks:       req.Hooks,
		SizeKB:      req.SizeKB,
	}

	if err := h.extRepo.InstallExtension(r.Context(), ext); err != nil {
		logrus.WithError(err).Error("extensions: failed to install extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to install extension")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"extension": ext})
}

// HandleUninstallExtension handles DELETE /v1/extensions/{id}
func (h *ExtensionsHandler) HandleUninstallExtension(w http.ResponseWriter, r *http.Request) {
	markExtensionDeprecation(w, "DELETE /v1/extensions/{id}")
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	if err := h.extRepo.UninstallExtension(r.Context(), tenantID, extID); err != nil {
		logrus.WithError(err).Error("extensions: failed to uninstall extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to uninstall extension")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Extension uninstalled"})
}

// HandleEnableExtension handles POST /v1/extensions/{id}/enable
func (h *ExtensionsHandler) HandleEnableExtension(w http.ResponseWriter, r *http.Request) {
	markExtensionDeprecation(w, "POST /v1/extensions/{id}/enable")
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	if err := h.extRepo.EnableExtension(r.Context(), tenantID, extID); err != nil {
		logrus.WithError(err).Error("extensions: failed to enable extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to enable extension")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Extension enabled"})
}

// HandleDisableExtension handles POST /v1/extensions/{id}/disable
func (h *ExtensionsHandler) HandleDisableExtension(w http.ResponseWriter, r *http.Request) {
	markExtensionDeprecation(w, "POST /v1/extensions/{id}/disable")
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	if err := h.extRepo.DisableExtension(r.Context(), tenantID, extID); err != nil {
		logrus.WithError(err).Error("extensions: failed to disable extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to disable extension")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Extension disabled"})
}

// HandleConfigureExtension handles PUT /v1/extensions/{id}/config
func (h *ExtensionsHandler) HandleConfigureExtension(w http.ResponseWriter, r *http.Request) {
	markExtensionDeprecation(w, "PUT /v1/extensions/{id}/config")
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req struct {
		Config []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	config := make(map[string]string)
	for _, c := range req.Config {
		config[c.Key] = c.Value
	}

	if err := h.extRepo.UpdateExtensionConfig(r.Context(), tenantID, extID, config); err != nil {
		logrus.WithError(err).Error("extensions: failed to configure extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to configure extension")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Extension configured"})
}

// HandleListHooks handles GET /v1/extensions/hooks
func (h *ExtensionsHandler) HandleListHooks(w http.ResponseWriter, r *http.Request) {
	markExtensionDeprecation(w, "GET /v1/extensions/hooks")
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	hooks, err := h.extRepo.ListHooks(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).Warn("extensions: failed to list hooks")
		writeJSON(w, http.StatusOK, map[string]interface{}{"hooks": []ExtensionHook{}})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"hooks": hooks})
}
