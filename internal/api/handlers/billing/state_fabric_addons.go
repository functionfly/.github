package billing

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// StateFabricAddOnsCatalogResponse is the public catalog (pricing page, docs).
type StateFabricAddOnsCatalogResponse struct {
	AddOns []statefabricaddons.AddOn `json:"add_ons"`
}

// StateFabricAddOnEntitlementsResponse lists active add-ons for the current tenant.
type StateFabricAddOnEntitlementsResponse struct {
	AddonIDs []string `json:"addon_ids"`
}

type CreateStateFabricAddOnCheckoutRequest struct {
	AddonID    string `json:"addon_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

// HandleListStateFabricAddOnCatalog returns the canonical State Fabric add-on catalog.
// GET /v1/billing/state-fabric/add-ons — public (no auth required).
func (h *Handler) HandleListStateFabricAddOnCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(StateFabricAddOnsCatalogResponse{
		AddOns: statefabricaddons.Catalog(),
	})
}

// HandleGetStateFabricAddOnEntitlements returns active add-on IDs for the signed-in tenant.
// GET /v1/billing/state-fabric/add-ons/entitlements
func (h *Handler) HandleGetStateFabricAddOnEntitlements(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if h.sfAddons == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(StateFabricAddOnEntitlementsResponse{AddonIDs: []string{}})
		return
	}
	ids, err := h.sfAddons.ListActiveAddonIDs(r.Context(), claims.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("billing: list state fabric add-on entitlements")
		writeJSONError(w, http.StatusInternalServerError, "Failed to load add-ons")
		return
	}
	if ids == nil {
		ids = []string{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(StateFabricAddOnEntitlementsResponse{AddonIDs: ids})
}

// HandleCreateStateFabricAddOnCheckout starts Stripe checkout for a state fabric add-on.
// POST /v1/billing/state-fabric/add-ons/checkout
func (h *Handler) HandleCreateStateFabricAddOnCheckout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}
	var req CreateStateFabricAddOnCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	addon, ok := statefabricaddons.GetByID(req.AddonID)
	if !ok {
		writeJSONError(w, http.StatusBadRequest, "unknown addon_id")
		return
	}
	if addon.StripePriceID == "" {
		writeJSONError(w, http.StatusServiceUnavailable, "Add-on checkout is not configured")
		return
	}
	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing addon checkout: user not found")
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}
	name := user.Name
	if name == "" {
		name = user.Email
	}

	successURL := req.SuccessURL
	if successURL == "" {
		successURL = os.Getenv("APP_URL")
	}
	cancelURL := req.CancelURL
	if cancelURL == "" {
		cancelURL = os.Getenv("APP_URL")
	}
	resp, err := payment.CreateStateFabricAddonCheckoutSession(
		r.Context(),
		h.repo,
		claims.TenantID,
		user.Email,
		name,
		payment.CreateAddonCheckoutSessionRequest{
			PriceID:    addon.StripePriceID,
			SuccessURL: successURL,
			CancelURL:  cancelURL,
			TenantID:   claims.TenantID,
			AddonID:    addon.ID,
		},
	)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"tenant_id": claims.TenantID,
			"addon_id":  addon.ID,
		}).Error("billing addon checkout: create session")
		msg := "Failed to create checkout session"
		if os.Getenv("DEVELOPMENT") == "true" {
			msg = err.Error()
		}
		writeJSONError(w, http.StatusInternalServerError, msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
