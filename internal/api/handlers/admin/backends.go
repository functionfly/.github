package admin

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// BackendsHandler handles admin backend management endpoints
type BackendsHandler struct {
	repo    storage.Repository
	authSvc *auth.AuthService
}

// NewBackendsHandler creates a new admin backends handler
func NewBackendsHandler(repo storage.Repository, authSvc *auth.AuthService) *BackendsHandler {
	return &BackendsHandler{
		repo:    repo,
		authSvc: authSvc,
	}
}

// HandleListPlatformBackends handles GET /admin/backends - lists all platform backends
func (h *BackendsHandler) HandleListPlatformBackends(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get all backends
	backends, err := h.repo.ListAllBackends(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to list all backends")
		apierror.WriteError(w, apierror.NewInternal("Failed to list backends"))
		return
	}

	// Build response with additional app/tenant info
	response := make([]map[string]interface{}, 0, len(backends))

	for _, backend := range backends {
		// Get app details
		app, err := h.repo.GetAppByID(ctx, backend.AppID)
		if err != nil {
			logrus.WithError(err).WithField("app_id", backend.AppID).Warn("Failed to get app details for backend")
			continue
		}

		// Get tenant details
		tenant, err := h.repo.GetTenantByID(ctx, app.TenantID)
		if err != nil {
			logrus.WithError(err).WithField("tenant_id", app.TenantID).Warn("Failed to get tenant details for backend")
			continue
		}

		// Mask the URL for security (show only host)
		maskedURL := backend.URL
		if len(maskedURL) > 0 {
			if idx := strings.Index(maskedURL, "://"); idx != -1 {
				maskedURL = maskedURL[idx+3:]
			}
			if idx := strings.Index(maskedURL, "/"); idx != -1 {
				maskedURL = maskedURL[:idx]
			}
		}

		response = append(response, map[string]interface{}{
			"id":          backend.ID.String(),
			"app_id":      backend.AppID.String(),
			"app_name":    app.Name,
			"tenant_id":   app.TenantID.String(),
			"tenant_name": tenant.Name,
			"provider":    backend.Provider,
			"region":      backend.Region,
			"url":         maskedURL,
			"enabled":     backend.Enabled,
			"priority":    backend.Priority,
			"created_at":  backend.CreatedAt,
			"updated_at":  backend.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"backends": response,
	})
}

// HandleUpdateBackendEnabled handles PATCH /admin/backends/:backendId - updates backend enabled status
func (h *BackendsHandler) HandleUpdateBackendEnabled(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse backend ID from path
	vars := mux.Vars(r)
	backendIDStr, ok := vars["backendId"]
	if !ok {
		apierror.WriteError(w, apierror.NewBadRequest("Backend ID required"))
		return
	}

	backendID, err := uuid.Parse(backendIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid backend ID"))
		return
	}

	// Parse request body
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Verify backend exists
	backend, err := h.repo.GetBackendByID(ctx, backendID)
	if err != nil {
		logrus.WithError(err).WithField("backend_id", backendID).Error("Failed to get backend")
		apierror.WriteError(w, apierror.NewNotFound("Backend not found"))
		return
	}
	if backend == nil {
		apierror.WriteError(w, apierror.NewNotFound("Backend not found"))
		return
	}

	// Update backend enabled status
	if err := h.repo.UpdateBackendEnabled(ctx, backendID, req.Enabled); err != nil {
		logrus.WithError(err).WithField("backend_id", backendID).Error("Failed to update backend enabled status")
		apierror.WriteError(w, apierror.NewInternal("Failed to update backend"))
		return
	}

	// Get updated backend
	updatedBackend, err := h.repo.GetBackendByID(ctx, backendID)
	if err != nil {
		logrus.WithError(err).WithField("backend_id", backendID).Error("Failed to get updated backend")
		apierror.WriteError(w, apierror.NewInternal("Failed to get updated backend"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         updatedBackend.ID.String(),
		"app_id":     updatedBackend.AppID.String(),
		"provider":   updatedBackend.Provider,
		"region":     updatedBackend.Region,
		"url":        updatedBackend.URL,
		"enabled":    updatedBackend.Enabled,
		"priority":   updatedBackend.Priority,
		"created_at": updatedBackend.CreatedAt,
		"updated_at": updatedBackend.UpdatedAt,
	})
}
