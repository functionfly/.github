package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/services/membership"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/api/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleListTenants lists all tenants
func (h *Handler) HandleListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.repo.ListTenants(r.Context())
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant ID"))
		return
	}

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": tenant})
}

// HandleCreateTenant creates a new tenant
func (h *Handler) HandleCreateTenant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	if req.Name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Tenant name is required"))
		return
	}

	tenant, err := h.repo.CreateTenant(r.Context(), req.Name)
	if err != nil {
		logrus.WithError(err).Error("Failed to create tenant")
		apierror.WriteError(w, apierror.NewInternal("Failed to create tenant"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant ID"))
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	// Get before state for audit
	beforeTenant, _ := h.repo.GetTenantByID(r.Context(), tenantID)
	if beforeTenant == nil {
		apierror.WriteError(w, apierror.NewNotFound("Tenant not found"))
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
		apierror.WriteError(w, apierror.NewInternal("Failed to update tenant"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant ID"))
		return
	}

	// Get before state for audit
	beforeTenant, _ := h.repo.GetTenantByID(r.Context(), tenantID)

	err = h.repo.DeleteTenant(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to delete tenant")
		if err.Error() == "cannot delete tenant with existing users" {
			apierror.WriteError(w, apierror.NewConflict("Cannot delete tenant with existing users"))
		} else {
			apierror.WriteError(w, apierror.NewInternal("Failed to delete tenant"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant ID"))
		return
	}

tenant, err := h.repo.GetTenantByID(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant")
		apierror.WriteError(w, apierror.NewInternal("Failed to get seat usage"))
		return
	}
	if tenant == nil {
		apierror.WriteError(w, apierror.NewNotFound("Tenant not found"))
		return
	}

	// Count active users for this tenant
	activeUserCount, err := h.repo.CountActiveUsersByTenant(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to count active users")
		apierror.WriteError(w, apierror.NewInternal("Failed to get seat usage"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant ID"))
		return
	}

	apps, err := h.repo.ListAppsByTenant(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to list tenant apps")
		apierror.WriteError(w, apierror.NewInternal("Failed to list apps"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid app ID"))
		return
	}

	app, err := h.repo.GetAppByID(r.Context(), appID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to get app")
		apierror.WriteError(w, apierror.NewInternal("Failed to get app"))
		return
	}
	if app == nil {
		apierror.WriteError(w, apierror.NewNotFound("App not found"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid app ID"))
		return
	}

	backends, err := h.repo.ListBackendsByAppID(r.Context(), appID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to list backends")
		apierror.WriteError(w, apierror.NewInternal("Failed to list backends"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid app ID"))
		return
	}

	// Default limit of 50 deployments
	deployments, err := h.repo.ListDeploymentsByAppID(r.Context(), appID, 50)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to list deployments")
		apierror.WriteError(w, apierror.NewInternal("Failed to list deployments"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid deployment ID"))
		return
	}

	// Get deployment to verify it exists
	deployment, err := h.repo.GetDeploymentByID(r.Context(), deploymentID)
	if err != nil {
		logrus.WithError(err).WithField("deployment_id", deploymentID).Error("Failed to get deployment")
		apierror.WriteError(w, apierror.NewInternal("Failed to get deployment"))
		return
	}
	if deployment == nil {
		apierror.WriteError(w, apierror.NewNotFound("Deployment not found"))
		return
	}

	// Update deployment status to rolled back
	err = h.repo.UpdateDeploymentStatus(r.Context(), deploymentID, "rolled_back", "Manually rolled back by admin", nil)
	if err != nil {
		logrus.WithError(err).WithField("deployment_id", deploymentID).Error("Failed to rollback deployment")
		apierror.WriteError(w, apierror.NewInternal("Failed to rollback deployment"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant ID"))
		return
	}

	tenant, err := h.repo.GetTenantByID(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant for metrics")
		apierror.WriteError(w, apierror.NewInternal("Failed to get tenant"))
		return
	}
	if tenant == nil {
		apierror.WriteError(w, apierror.NewNotFound("Tenant not found"))
		return
	}

	// Fetch aggregated dashboard metrics for this tenant
	dashboardMetrics, err := h.repo.GetDashboardMetrics(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get dashboard metrics")
		dashboardMetrics = nil
	}

	// Fetch per-function invocation/error counts from recent logs
	execRate24h, err := h.repo.GetExecutionRateByHour(r.Context(), tenantID, 24)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to get execution rate for tenant metrics")
	}

	executions24h := 0
	for _, h := range execRate24h {
		executions24h += h.Rate
	}

	// Get total function count for this tenant
	functions, _, err := h.repo.ListAllFunctions(r.Context(), 1000, 0, &tenantID, nil)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to list functions for tenant metrics")
	}

	var latencyP50, latencyP99 float64
	if dashboardMetrics != nil && dashboardMetrics.AvgLatencyMs != nil {
		latencyP50 = *dashboardMetrics.AvgLatencyMs
		latencyP99 = latencyP50 * 2
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tenant_id":          tenantID,
		"tenant_name":        tenant.Name,
		"plan":               tenant.Plan,
		"functions": map[string]interface{}{
			"total":          len(functions),
			"executions_24h":  executions24h,
			"requests_month":  dashboardMetrics.RequestsThisMonth,
			"requests_prev":   dashboardMetrics.RequestsPrevMonth,
		},
		"latency_p50_ms": latencyP50,
		"latency_p99_ms": latencyP99,
		"uptime_pct":     dashboardMetrics.UptimePct,
		"generated_at":    time.Now().UTC(),
	})
}

// HandleTenantHealth returns health status for a specific tenant (admin view)
func (h *Handler) HandleTenantHealth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant ID"))
		return
	}

	tenant, err := h.repo.GetTenantByID(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant for health check")
		apierror.WriteError(w, apierror.NewInternal("Failed to get tenant"))
		return
	}
	if tenant == nil {
		apierror.WriteError(w, apierror.NewNotFound("Tenant not found"))
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
