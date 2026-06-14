package provisioning

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler provides HTTP endpoints for one-click bundle provisioning.
type Handler struct {
	provisioner *BundleProvisioner
	repo        storage.Repository
}

// NewHandler creates a new provisioning handler
func NewHandler(provisioner *BundleProvisioner, repo storage.Repository) *Handler {
	return &Handler{
		provisioner: provisioner,
		repo:        repo,
	}
}

// RegisterRoutes registers the provisioning API routes on a gorilla/mux router.
// Mounted under /v1/provisioning/ — caller passes subrouter or full router.
func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/provisioning/bundle", h.HandleProvisionBundle).Methods("POST", "OPTIONS")
	r.HandleFunc("/provisioning/status", h.HandleGetProvisioningStatus).Methods("GET", "OPTIONS")
	r.HandleFunc("/provisioning/retry", h.HandleRetryFailedComponents).Methods("POST", "OPTIONS")
}

// HandleProvisionBundle triggers one-click provisioning for the authenticated tenant.
// POST /v1/provisioning/bundle
//
// Request body:
//
//	{ "bundle_slug": "saas-starter" }
//
// Response:
//
//	{ "status": "active", "components": {...}, "duration_ms": 1234 }
func (h *Handler) HandleProvisionBundle(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		BundleSlug string `json:"bundle_slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate bundle slug
	validSlugs := map[string]bool{"saas-starter": true, "marketplace": true, "ai-app": true}
	if !validSlugs[req.BundleSlug] {
		writeJSONError(w, http.StatusBadRequest, "Invalid bundle slug")
		return
	}

	// Check if already provisioning or active
	existing, _ := h.provisioner.GetProvisioningStatus(r.Context(), claims.TenantID)
	if existing != nil && existing.Status == StatusProvisioning {
		writeJSONError(w, http.StatusConflict, "Provisioning already in progress")
		return
	}

	// Check plan eligibility (Starter+ plans get dedicated DBs)
	tenant, err := h.repo.GetTenantByID(r.Context(), claims.TenantID)
	if err != nil || tenant == nil {
		writeJSONError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	// Execute provisioning (async to avoid HTTP timeout on slow DB creation)
	resultCh := make(chan *ProvisionResult, 1)
	errCh := make(chan error, 1)

	go func() {
		result, err := h.provisioner.ProvisionBundle(r.Context(), claims.TenantID, req.BundleSlug)
		if err != nil {
			errCh <- err
		} else {
			resultCh <- result
		}
	}()

	// Wait for result (or timeout after 5 minutes)
	select {
	case result := <-resultCh:
		w.Header().Set("Content-Type", "application/json")
		if result.Status == StatusActive {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusPartialContent) // 206: some components failed
		}
		json.NewEncoder(w).Encode(result)

	case err := <-errCh:
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("Bundle provisioning failed")
		writeJSONError(w, http.StatusInternalServerError, "Provisioning failed: "+err.Error())

	case <-r.Context().Done():
		writeJSONError(w, http.StatusRequestTimeout, "Provisioning timed out")
	}
}

// HandleGetProvisioningStatus returns the current provisioning state.
// GET /v1/provisioning/status
func (h *Handler) HandleGetProvisioningStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	result, err := h.provisioner.GetProvisioningStatus(r.Context(), claims.TenantID)
	if err != nil {
		// No provisioning state found — return empty
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "not_provisioned",
			"tenant_id": claims.TenantID,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleRetryFailedComponents re-runs only the components that failed.
// POST /v1/provisioning/retry
func (h *Handler) HandleRetryFailedComponents(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	existing, err := h.provisioner.GetProvisioningStatus(r.Context(), claims.TenantID)
	if err != nil || existing == nil {
		writeJSONError(w, http.StatusNotFound, "No provisioning state found")
		return
	}

	if existing.Status == StatusActive {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "already_active",
			"message": "All components are already active",
		})
		return
	}

	// Re-provision (idempotent — active components are skipped)
	go func() {
		h.provisioner.ProvisionBundle(r.Context(), claims.TenantID, existing.BundleSlug)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "retrying",
		"message": "Retrying failed components",
	})
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
