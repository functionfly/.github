package billing

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// HandleListPaymentMethods returns all payment methods for the authenticated tenant.
// GET /billing/payment-methods
func (h *Handler) HandleListPaymentMethods(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	tenant, err := h.repo.GetTenantByID(claims.TenantID)
	if err != nil || tenant == nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("list payment methods: failed to get tenant")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve tenant")
		return
	}

	if tenant.StripeCustomerID == nil || *tenant.StripeCustomerID == "" {
		writeJSONError(w, http.StatusBadRequest, "No billing account found")
		return
	}

	methods, err := payment.ListPaymentMethodsForCustomer(r.Context(), *tenant.StripeCustomerID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("list payment methods: failed to list payment methods")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve payment methods")
		return
	}

	out := make([]PaymentMethodInfo, 0, len(methods))
	for _, m := range methods {
		out = append(out, PaymentMethodInfo{
			Brand:    m.Brand,
			Last4:    m.Last4,
			ExpMonth: m.ExpMonth,
			ExpYear:  m.ExpYear,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payment_methods": out,
	})
}

// HandleCreateSetupIntent creates a Stripe SetupIntent for adding a new payment method.
// POST /billing/payment-methods/setup-intent
func (h *Handler) HandleCreateSetupIntent(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	tenant, err := h.repo.GetTenantByID(claims.TenantID)
	if err != nil || tenant == nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("create setup intent: failed to get tenant")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve tenant")
		return
	}

	if tenant.StripeCustomerID == nil || *tenant.StripeCustomerID == "" {
		writeJSONError(w, http.StatusBadRequest, "No billing account found")
		return
	}

	result, err := payment.CreateSetupIntent(r.Context(), *tenant.StripeCustomerID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("create setup intent: failed to create setup intent")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create setup intent")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"client_secret": result.ClientSecret,
	})
}

// HandleSetDefaultPaymentMethod sets the default payment method for the tenant.
// POST /billing/payment-methods/default
func (h *Handler) HandleSetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	var req struct {
		PaymentMethodID string `json:"payment_method_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PaymentMethodID == "" {
		writeJSONError(w, http.StatusBadRequest, "payment_method_id is required")
		return
	}

	tenant, err := h.repo.GetTenantByID(claims.TenantID)
	if err != nil || tenant == nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("set default payment method: failed to get tenant")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve tenant")
		return
	}

	if tenant.StripeCustomerID == nil || *tenant.StripeCustomerID == "" {
		writeJSONError(w, http.StatusBadRequest, "No billing account found")
		return
	}

	if err := payment.SetDefaultPaymentMethod(r.Context(), *tenant.StripeCustomerID, req.PaymentMethodID); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"tenant_id":         claims.TenantID,
			"payment_method_id": req.PaymentMethodID,
		}).Error("set default payment method: failed to set default payment method")
		writeJSONError(w, http.StatusInternalServerError, "Failed to set default payment method")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Default payment method updated successfully",
	})
}

// HandleDetachPaymentMethod removes a payment method from the tenant.
// DELETE /billing/payment-methods/{id}
func (h *Handler) HandleDetachPaymentMethod(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	paymentMethodID := r.PathValue("id")
	if paymentMethodID == "" {
		writeJSONError(w, http.StatusBadRequest, "Payment method ID is required")
		return
	}

	tenant, err := h.repo.GetTenantByID(claims.TenantID)
	if err != nil || tenant == nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("detach payment method: failed to get tenant")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve tenant")
		return
	}

	if tenant.StripeCustomerID == nil || *tenant.StripeCustomerID == "" {
		writeJSONError(w, http.StatusBadRequest, "No billing account found")
		return
	}

	if err := payment.DetachPaymentMethod(r.Context(), paymentMethodID); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"tenant_id":         claims.TenantID,
			"payment_method_id": paymentMethodID,
		}).Error("detach payment method: failed to detach payment method")
		writeJSONError(w, http.StatusInternalServerError, "Failed to remove payment method")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Payment method removed successfully",
	})
}