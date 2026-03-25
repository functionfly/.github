package admin

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/utils"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HandleListStateFabricAddonCatalog returns canonical add-on catalog for admin tooling.
// GET /v1/admin/billing/state-fabric-add-ons/catalog
func (h *Handler) HandleListStateFabricAddonCatalog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    statefabricaddons.Catalog(),
	})
}

// HandleListStateFabricTenantEntitlements lists add-on entitlements for one tenant.
// GET /v1/admin/billing/state-fabric-add-ons/entitlements/{tenantId}
func (h *Handler) HandleListStateFabricTenantEntitlements(w http.ResponseWriter, r *http.Request) {
	if h.sfAddons == nil {
		http.Error(w, "state fabric add-on repository not configured", http.StatusServiceUnavailable)
		return
	}
	tenantID, err := uuid.Parse(mux.Vars(r)["tenantId"])
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}
	items, err := h.sfAddons.ListEntitlementsByTenant(r.Context(), tenantID)
	if err != nil {
		http.Error(w, "Failed to list entitlements", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    items,
	})
}

type adminUpsertAddonEntitlementRequest struct {
	Status                   string  `json:"status"`
	StripeSubscriptionID     *string `json:"stripe_subscription_id"`
	StripeSubscriptionItemID *string `json:"stripe_subscription_item_id"`
}

// HandleUpsertStateFabricTenantEntitlement upserts tenant entitlement by addon id.
// PATCH /v1/admin/billing/state-fabric-add-ons/entitlements/{tenantId}/{addonId}
func (h *Handler) HandleUpsertStateFabricTenantEntitlement(w http.ResponseWriter, r *http.Request) {
	if h.sfAddons == nil {
		http.Error(w, "state fabric add-on repository not configured", http.StatusServiceUnavailable)
		return
	}
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["tenantId"])
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}
	addonID := vars["addonId"]
	if _, ok := statefabricaddons.GetByID(addonID); !ok {
		http.Error(w, "Invalid addon ID", http.StatusBadRequest)
		return
	}

	var req adminUpsertAddonEntitlementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if req.Status != "active" && req.Status != "inactive" && req.Status != "suspended" {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	before, err := h.sfAddons.GetEntitlement(r.Context(), tenantID, addonID)
	if err != nil {
		http.Error(w, "Failed to load entitlement", http.StatusInternalServerError)
		return
	}

	if err := h.sfAddons.UpsertEntitlement(
		r.Context(),
		tenantID,
		addonID,
		req.Status,
		req.StripeSubscriptionID,
		req.StripeSubscriptionItemID,
	); err != nil {
		utils.LogAuditEvent(r.Context(), h.repo, r,
			"billing.state_fabric_addon.entitlement.upsert",
			"state_fabric_addon_entitlement",
			&tenantID,
			auditEntitlementSnapshot(tenantID, addonID, before),
			map[string]interface{}{
				"error": "upsert_failed",
			},
			false,
		)
		http.Error(w, "Failed to update entitlement", http.StatusInternalServerError)
		return
	}

	after, errAfter := h.sfAddons.GetEntitlement(r.Context(), tenantID, addonID)
	afterSnap := auditEntitlementSnapshot(tenantID, addonID, after)
	if errAfter != nil || after == nil {
		afterSnap = auditEntitlementAfterFallback(tenantID, addonID, req)
	}
	utils.LogAuditEvent(r.Context(), h.repo, r,
		"billing.state_fabric_addon.entitlement.upsert",
		"state_fabric_addon_entitlement",
		&tenantID,
		auditEntitlementSnapshot(tenantID, addonID, before),
		afterSnap,
		true,
	)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]string{
			"tenant_id": tenantID.String(),
			"addon_id":  addonID,
			"status":    req.Status,
		},
	})
}

func auditEntitlementSnapshot(tenantID uuid.UUID, addonID string, e *statefabricaddons.Entitlement) map[string]interface{} {
	if e == nil {
		return map[string]interface{}{
			"tenant_id": tenantID.String(),
			"addon_id":  addonID,
			"exists":    false,
		}
	}
	out := map[string]interface{}{
		"tenant_id": e.TenantID.String(),
		"addon_id":  e.AddonID,
		"status":    e.Status,
		"exists":    true,
	}
	if e.StripeSubscriptionID != nil {
		out["stripe_subscription_id"] = *e.StripeSubscriptionID
	}
	if e.StripeSubscriptionItemID != nil {
		out["stripe_subscription_item_id"] = *e.StripeSubscriptionItemID
	}
	return out
}

func auditEntitlementAfterFallback(
	tenantID uuid.UUID,
	addonID string,
	req adminUpsertAddonEntitlementRequest,
) map[string]interface{} {
	out := map[string]interface{}{
		"tenant_id": tenantID.String(),
		"addon_id":  addonID,
		"status":    req.Status,
		"exists":    true,
		"note":      "after_state_from_request",
	}
	if req.StripeSubscriptionID != nil {
		out["stripe_subscription_id"] = *req.StripeSubscriptionID
	}
	if req.StripeSubscriptionItemID != nil {
		out["stripe_subscription_item_id"] = *req.StripeSubscriptionItemID
	}
	return out
}
