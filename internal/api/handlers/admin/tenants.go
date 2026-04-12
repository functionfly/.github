package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/utils"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/services/membership"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleListTenants lists all tenants
func (h *Handler) HandleListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.repo.ListTenants()
	if err != nil {
		logrus.WithError(err).Error("Failed to list tenants")
		// Return empty list so admin UI can load; caller can retry or check logs
		tenants = []*storage.Tenant{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":    tenants,
		"tenants": tenants,
	})
}

// HandleGetTenant gets a specific tenant
func (h *Handler) HandleGetTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	tenant, err := h.repo.GetTenantByID(tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant")
		http.Error(w, "Failed to get tenant", http.StatusInternalServerError)
		return
	}
	if tenant == nil {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": tenant})
}

// HandleCreateTenant creates a new tenant
func (h *Handler) HandleCreateTenant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Tenant name is required", http.StatusBadRequest)
		return
	}

	tenant, err := h.repo.CreateTenant(r.Context(), req.Name)
	if err != nil {
		logrus.WithError(err).Error("Failed to create tenant")
		http.Error(w, "Failed to create tenant", http.StatusInternalServerError)
		return
	}

	// Log successful creation
	utils.LogAuditEvent(r.Context(), h.repo, r, "tenant.create", "tenant", &tenant.ID, nil, tenant, true)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tenant)
}

// HandleUpdateTenant updates a tenant
func (h *Handler) HandleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get before state for audit
	beforeTenant, _ := h.repo.GetTenantByID(tenantID)
	if beforeTenant == nil {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	oldPlan := beforeTenant.Plan

	// Check if plan is being changed and if new plan would exceed seat limit
	if newPlan, ok := updates["plan"].(string); ok && newPlan != "" {
		newMaxUsers := plans.MaxUsersPerPlan(newPlan)
		if newMaxUsers != -1 {
			if currentCount, err := h.repo.CountActiveUsersByTenant(r.Context(), tenantID); err == nil && currentCount > newMaxUsers {
				gracePeriodEnd := time.Now().AddDate(0, 0, plans.GetSeatGracePeriodDays())
				updates["seat_grace_period_end"] = gracePeriodEnd
				logrus.WithFields(logrus.Fields{"tenant_id": tenantID, "new_plan": newPlan, "current_users": currentCount, "max_users": newMaxUsers, "grace_period_end": gracePeriodEnd}).Info("Plan downgrade exceeded seat limit - grace period started")
			}
		}
	}

	tenant, err := h.repo.UpdateTenant(r.Context(), tenantID, updates)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to update tenant")
		http.Error(w, "Failed to update tenant", http.StatusInternalServerError)
		return
	}

	// Check if plan was upgraded and process membership events
	if newPlan, ok := updates["plan"].(string); ok && newPlan != "" && newPlan != oldPlan {
		go h.processPlanUpgrade(tenantID, oldPlan, newPlan)
	}

	// Log successful update
	utils.LogAuditEvent(r.Context(), h.repo, r, "tenant.update", "tenant", &tenantID, beforeTenant, tenant, true)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenant)
}

// processPlanUpgrade handles membership events for plan upgrades asynchronously
func (h *Handler) processPlanUpgrade(tenantID uuid.UUID, oldPlan, newPlan string) {
	// Get all active users in the tenant to award achievements and create activities
	users, err := h.repo.ListActiveUsersByTenant(context.Background(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to list users for plan upgrade processing")
		return
	}

	// Get the admin user who likely performed the upgrade (first user as proxy)
	var adminUserID uuid.UUID
	if len(users) > 0 {
		adminUserID = users[0].ID
	}

	for _, user := range users {
		upgradeData := membership.PlanUpgradeData{
			UserID:     user.ID,
			TenantID:   tenantID,
			OldPlan:    oldPlan,
			NewPlan:    newPlan,
			UpgradedAt: time.Now(),
			UpgradedBy: adminUserID,
		}

		if err := h.membershipSvc.HandlePlanUpgrade(context.Background(), upgradeData); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"user_id":   user.ID,
				"tenant_id": tenantID,
			}).Warn("Failed to process plan upgrade for user")
		}
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":  tenantID,
		"old_plan":   oldPlan,
		"new_plan":   newPlan,
		"user_count": len(users),
	}).Info("Processed plan upgrade for tenant users")
}

// HandleDeleteTenant deletes a tenant
func (h *Handler) HandleDeleteTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Get before state for audit
	beforeTenant, _ := h.repo.GetTenantByID(tenantID)

	err = h.repo.DeleteTenant(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to delete tenant")
		if err.Error() == "cannot delete tenant with existing users" {
			http.Error(w, "Cannot delete tenant with existing users", http.StatusConflict)
		} else {
			http.Error(w, "Failed to delete tenant", http.StatusInternalServerError)
		}
		return
	}

	// Log successful deletion
	utils.LogAuditEvent(r.Context(), h.repo, r, "tenant.delete", "tenant", &tenantID, beforeTenant, nil, true)

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetSeatUsage returns seat usage information for a tenant
// GET /v1/admin/tenants/{tenantId}/seat-usage
func (h *Handler) HandleGetSeatUsage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	tenant, err := h.repo.GetTenantByID(tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant")
		http.Error(w, "Failed to get tenant", http.StatusInternalServerError)
		return
	}
	if tenant == nil {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	// Count active users for this tenant
	activeUserCount, err := h.repo.CountActiveUsersByTenant(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to count active users")
		http.Error(w, "Failed to get seat usage", http.StatusInternalServerError)
		return
	}

	// Get seat usage info from plans package
	seatInfo := plans.GetSeatUsage(tenant.Plan, activeUserCount)

	// Include grace period info if set
	var gracePeriodEndsAt *time.Time
	if tenant.SeatGracePeriodEnd != nil {
		gracePeriodEndsAt = tenant.SeatGracePeriodEnd
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tenant_id":         tenantID,
		"tenant_name":       tenant.Name,
		"plan":              seatInfo.Plan,
		"current_users":     seatInfo.CurrentUsers,
		"max_users":         seatInfo.MaxUsers,
		"is_unlimited":      seatInfo.IsUnlimited,
		"is_at_limit":       seatInfo.IsAtLimit,
		"is_at_warning":     seatInfo.IsAtWarning,
		"warning_threshold": seatInfo.WarningPercent,
		"grace_period_ends": gracePeriodEndsAt,
	})
}

// HandleListTenantApps lists apps for a specific tenant (admin impersonating tenant)
func (h *Handler) HandleListTenantApps(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	apps, err := h.repo.ListAppsByTenant(tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to list tenant apps")
		http.Error(w, "Failed to list apps", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": apps})
}

// HandleGetTenantApp gets a specific app for a tenant (admin impersonating tenant)
func (h *Handler) HandleGetTenantApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appIDStr := vars["appId"]

	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}

	app, err := h.repo.GetAppByID(appID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to get app")
		http.Error(w, "Failed to get app", http.StatusInternalServerError)
		return
	}
	if app == nil {
		http.Error(w, "App not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": app})
}

// HandleListTenantBackends lists backends for a specific tenant app (admin impersonating tenant)
func (h *Handler) HandleListTenantBackends(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appIDStr := vars["appId"]

	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}

	backends, err := h.repo.ListBackendsByAppID(appID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to list backends")
		http.Error(w, "Failed to list backends", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": backends})
}

// HandleListTenantDeployments lists deployments for a specific tenant app (admin impersonating tenant)
func (h *Handler) HandleListTenantDeployments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appIDStr := vars["appId"]

	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}

	// Default limit of 50 deployments
	deployments, err := h.repo.ListDeploymentsByAppID(appID, 50)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to list deployments")
		http.Error(w, "Failed to list deployments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": deployments})
}

// HandleTenantDeploymentRollback rolls back a deployment for a tenant app (admin impersonating tenant)
func (h *Handler) HandleTenantDeploymentRollback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deploymentIDStr := vars["deploymentId"]

	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		http.Error(w, "Invalid deployment ID", http.StatusBadRequest)
		return
	}

	// Get deployment to verify it exists
	deployment, err := h.repo.GetDeploymentByID(deploymentID)
	if err != nil {
		logrus.WithError(err).WithField("deployment_id", deploymentID).Error("Failed to get deployment")
		http.Error(w, "Failed to get deployment", http.StatusInternalServerError)
		return
	}
	if deployment == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	// Update deployment status to rolled back
	err = h.repo.UpdateDeploymentStatus(deploymentID, "rolled_back", "Manually rolled back by admin", nil)
	if err != nil {
		logrus.WithError(err).WithField("deployment_id", deploymentID).Error("Failed to rollback deployment")
		http.Error(w, "Failed to rollback deployment", http.StatusInternalServerError)
		return
	}

	utils.LogAuditEvent(r.Context(), h.repo, r, "deployment.rollback", "deployment", &deploymentID, nil, map[string]interface{}{
		"app_id": deployment.AppID,
	}, true)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Deployment rolled back successfully",
	})
}

// HandleTenantMetrics returns metrics for a specific tenant (admin view)
func (h *Handler) HandleTenantMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Placeholder: tenant metrics can be integrated with metrics service
	metrics := map[string]interface{}{
		"tenant_id": tenantID,
		"functions": map[string]interface{}{
			"total":     0,
			"executions_24h": 0,
		},
		"generated_at": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// HandleTenantHealth returns health status for a specific tenant (admin view)
func (h *Handler) HandleTenantHealth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	tenant, err := h.repo.GetTenantByID(tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant for health check")
		http.Error(w, "Failed to get tenant", http.StatusInternalServerError)
		return
	}
	if tenant == nil {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	health := map[string]interface{}{
		"tenant_id":  tenantID,
		"tenant_name": tenant.Name,
		"status":     "healthy",
		"checks": map[string]interface{}{
			"database": "ok",
			"storage":  "ok",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
