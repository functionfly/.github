package statefabric

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/plans"
	statestore "github.com/functionfly/functionfly/internal/storage/state"
)

func (h *Handler) requireFabricPermission(
	w http.ResponseWriter,
	r *http.Request,
	tenantID, userID, fabricID uuid.UUID,
	permission string,
) bool {
	if h.repo == nil {
		return true
	}
	allowed, err := h.repo.UserHasFabricPermission(r.Context(), tenantID, fabricID, userID, permission)
	if err != nil {
		logrus.WithError(err).Error("fabric permission check failed")
		http.Error(w, "permission check failed", http.StatusInternalServerError)
		return false
	}
	if !allowed {
		http.Error(w, "permission denied", http.StatusForbidden)
		return false
	}
	return true
}

func (h *Handler) requireFabricQuota(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) bool {
	if h.planResolver == nil {
		return true
	}

	plan := h.planResolver.GetTenantPlan(r.Context(), tenantID)
	if !plans.PlanHasStateFabricFeature(plan) {
		writeErr(w, http.StatusForbidden, "State Fabric requires a Starter plan or higher")
		return false
	}

	maxFabrics := plans.MaxStateFabricsPerPlan(plan)
	if maxFabrics == 0 {
		writeErr(w, http.StatusForbidden, "State Fabric is not available on your plan")
		return false
	}
	if maxFabrics < 0 {
		return true
	}

	if h.repo == nil {
		return true
	}

	count, err := h.repo.CountFabricsByTenant(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).Error("failed to count state fabrics")
		http.Error(w, "failed to verify fabric quota", http.StatusInternalServerError)
		return false
	}
	if count >= maxFabrics {
		writeErr(w, http.StatusPaymentRequired, "state fabric limit reached for your plan")
		return false
	}
	return true
}

// Permission shortcuts using shared state permission columns.
const (
	fabricPermRead    = statestore.PermRead
	fabricPermWrite   = statestore.PermWrite
	fabricPermDelete  = statestore.PermDelete
	fabricPermAdmin   = statestore.PermAdmin
	fabricPermTrigger = statestore.PermTrigger
)
