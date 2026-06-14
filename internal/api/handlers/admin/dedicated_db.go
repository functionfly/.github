package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// TenantDBServiceInterface defines the interface for tenant database management
type TenantDBServiceInterface interface {
	ProvisionForTenant(ctx context.Context, tenantID uuid.UUID) error
	DeprovisionForTenant(ctx context.Context, tenantID uuid.UUID) error
	SuspendTenant(ctx context.Context, tenantID uuid.UUID) error
	ResumeTenant(ctx context.Context, tenantID uuid.UUID) error
	GetHealthStatus(tenantID uuid.UUID) (*storage.TenantHealthStatus, error)
	GetPoolStats() storage.TenantPoolStats
}

// SetTenantDBService sets the tenant database service for admin operations
func (h *Handler) SetTenantDBService(svc TenantDBServiceInterface) {
	h.tenantDBService = svc
}

// HandleGetTenantDedicatedDBStatus returns the status of a tenant's dedicated database
// GET /v1/admin/tenants/{tenantId}/dedicated-db/status
func (h *Handler) HandleGetTenantDedicatedDBStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant ID"))
		return
	}

	// Check if tenant exists
	tenant, err := h.repo.GetTenantByID(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant")
		apierror.WriteError(w, apierror.NewInternal("Failed to get tenant"))
		return
	}
	if tenant == nil {
		apierror.WriteError(w, apierror.NewNotFound("Tenant not found"))
		return
	}

	// Check if tenant DB service is available
	if h.tenantDBService == nil {
		// Return disabled status if service not configured
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tenant_id":         tenantID,
			"tenant_name":       tenant.Name,
			"enabled":           false,
			"status":            "not_configured",
			"message":           "Tenant database service not configured",
			"dedicated_db_plan": tenant.Plan,
		})
		return
	}

	// Get health status
	health, err := h.tenantDBService.GetHealthStatus(tenantID)
	poolStats := h.tenantDBService.GetPoolStats()

	status := "unknown"
	healthStatus := "unknown"
	var latencyMs int64

	if err == nil && health != nil {
		status = string(health.Status)
		healthStatus = string(health.Status)
		latencyMs = health.LatencyMs
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tenant_id":           tenantID,
		"tenant_name":         tenant.Name,
		"enabled":             true,
		"status":              status,
		"health_status":       healthStatus,
		"latency_ms":          latencyMs,
		"plan":                tenant.Plan,
		"pool_stats": map[string]interface{}{
			"total_pools":      poolStats.TotalPools,
			"healthy_pools":    poolStats.HealthyPools,
			"degraded_pools":    poolStats.DegradedPools,
			"unhealthy_pools":   poolStats.UnhealthyPools,
			"total_connections": poolStats.TotalConns,
		},
		"checked_at": time.Now().UTC(),
	})
}

// HandleProvisionTenantDedicatedDB provisions a dedicated database for a tenant
// POST /v1/admin/tenants/{tenantId}/dedicated-db/provision
func (h *Handler) HandleProvisionTenantDedicatedDB(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant ID"))
		return
	}

	// Check if tenant exists
	tenant, err := h.repo.GetTenantByID(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant")
		apierror.WriteError(w, apierror.NewInternal("Failed to get tenant"))
		return
	}
	if tenant == nil {
		apierror.WriteError(w, apierror.NewNotFound("Tenant not found"))
		return
	}

	// Check if tenant DB service is available
	if h.tenantDBService == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Tenant database service not configured"))
		return
	}

	// Check if plan qualifies for dedicated DB
	qualifyingPlans := map[string]bool{
		"starter":     true,
		"professional": true,
		"enterprise":  true,
	}
	if !qualifyingPlans[tenant.Plan] {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "plan_not_eligible",
			"message": "Plan does not qualify for dedicated database",
			"plan":    tenant.Plan,
		})
		return
	}

	// Provision the dedicated database
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	if err := h.tenantDBService.ProvisionForTenant(ctx, tenantID); err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to provision dedicated database")
		apierror.WriteError(w, apierror.NewInternal("Failed to provision dedicated database: "+err.Error()))
		return
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":  tenantID,
		"admin_user": adminUserFromRequest(r),
	}).Info("Admin manually provisioned dedicated database for tenant")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Dedicated database provisioned successfully",
		"tenant_id": tenantID,
		"status":    "active",
	})
}

// HandleSuspendTenantDedicatedDB suspends a tenant's dedicated database
// POST /v1/admin/tenants/{tenantId}/dedicated-db/suspend
func (h *Handler) HandleSuspendTenantDedicatedDB(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant ID"))
		return
	}

	// Check if tenant exists
	tenant, err := h.repo.GetTenantByID(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant")
		apierror.WriteError(w, apierror.NewInternal("Failed to get tenant"))
		return
	}
	if tenant == nil {
		apierror.WriteError(w, apierror.NewNotFound("Tenant not found"))
		return
	}

	// Check if tenant DB service is available
	if h.tenantDBService == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Tenant database service not configured"))
		return
	}

	// Suspend the dedicated database
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := h.tenantDBService.SuspendTenant(ctx, tenantID); err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to suspend dedicated database")
		apierror.WriteError(w, apierror.NewInternal("Failed to suspend dedicated database: "+err.Error()))
		return
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":  tenantID,
		"admin_user": adminUserFromRequest(r),
	}).Info("Admin suspended dedicated database for tenant")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Dedicated database suspended successfully",
		"tenant_id": tenantID,
		"status":    "suspended",
	})
}

// HandleResumeTenantDedicatedDB resumes a suspended tenant's dedicated database
// POST /v1/admin/tenants/{tenantId}/dedicated-db/resume
func (h *Handler) HandleResumeTenantDedicatedDB(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant ID"))
		return
	}

	// Check if tenant exists
	tenant, err := h.repo.GetTenantByID(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant")
		apierror.WriteError(w, apierror.NewInternal("Failed to get tenant"))
		return
	}
	if tenant == nil {
		apierror.WriteError(w, apierror.NewNotFound("Tenant not found"))
		return
	}

	// Check if tenant DB service is available
	if h.tenantDBService == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Tenant database service not configured"))
		return
	}

	// Resume the dedicated database
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := h.tenantDBService.ResumeTenant(ctx, tenantID); err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to resume dedicated database")
		apierror.WriteError(w, apierror.NewInternal("Failed to resume dedicated database: "+err.Error()))
		return
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":  tenantID,
		"admin_user": adminUserFromRequest(r),
	}).Info("Admin resumed dedicated database for tenant")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Dedicated database resumed successfully",
		"tenant_id": tenantID,
		"status":    "active",
	})
}

// HandleDeprovisionTenantDedicatedDB deprovisions a tenant's dedicated database
// DELETE /v1/admin/tenants/{tenantId}/dedicated-db
func (h *Handler) HandleDeprovisionTenantDedicatedDB(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant ID"))
		return
	}

	// Check if tenant exists
	tenant, err := h.repo.GetTenantByID(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant")
		apierror.WriteError(w, apierror.NewInternal("Failed to get tenant"))
		return
	}
	if tenant == nil {
		apierror.WriteError(w, apierror.NewNotFound("Tenant not found"))
		return
	}

	// Check if tenant DB service is available
	if h.tenantDBService == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Tenant database service not configured"))
		return
	}

	// Deprovision the dedicated database
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	if err := h.tenantDBService.DeprovisionForTenant(ctx, tenantID); err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to deprovision dedicated database")
		apierror.WriteError(w, apierror.NewInternal("Failed to deprovision dedicated database: "+err.Error()))
		return
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":  tenantID,
		"admin_user": adminUserFromRequest(r),
	}).Info("Admin deprovisioned dedicated database for tenant")

	w.WriteHeader(http.StatusNoContent)
}

// HandleListTenantDedicatedDBs lists all tenant databases (admin view)
// GET /v1/admin/dedicated-dbs
func (h *Handler) HandleListTenantDedicatedDBs(w http.ResponseWriter, r *http.Request) {
	if h.tenantDBService == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Tenant database service not configured"))
		return
	}

	// Get pool stats as a summary
	stats := h.tenantDBService.GetPoolStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"summary": map[string]interface{}{
			"total_pools":      stats.TotalPools,
			"healthy_pools":    stats.HealthyPools,
			"degraded_pools":   stats.DegradedPools,
			"unhealthy_pools":  stats.UnhealthyPools,
			"total_connections": stats.TotalConns,
		},
	})
}

// adminUserFromRequest extracts the admin username from the request context
func adminUserFromRequest(r *http.Request) string {
	if claims := middleware.GetUserFromContext(r); claims != nil {
		return claims.Username
	}
	return "unknown"
}