package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	"github.com/functionfly/functionfly/internal/storage"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/wallet"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// PaymentMethodInfo represents the payment method details for display
type PaymentMethodInfo struct {
	Brand    string `json:"brand"`
	Last4    string `json:"last4"`
	ExpMonth int    `json:"exp_month"`
	ExpYear  int    `json:"exp_year"`
}

// TenantInvoiceJSON is the dashboard-facing shape for GET /v1/billing/invoices (amounts in cents).
type TenantInvoiceJSON struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	StripeInvoiceID  *string    `json:"stripe_invoice_id"`
	Amount           int        `json:"amount"`
	Currency         string     `json:"currency"`
	Status           string     `json:"status"`
	InvoiceDate      *time.Time `json:"invoice_date"`
	DueDate          *time.Time `json:"due_date"`
	InvoicePDF       *string    `json:"invoice_pdf"`
	HostedInvoiceURL *string    `json:"hosted_invoice_url"`
	CreatedAt        time.Time  `json:"created_at"`
}

// SubscriptionResponse is the response for subscription details
type SubscriptionResponse struct {
	ID                   uuid.UUID          `json:"id"`
	TenantID             uuid.UUID          `json:"tenant_id"`
	Plan                 string             `json:"plan"`
	Status               string             `json:"status"`
	StripeSubscriptionID string             `json:"stripe_subscription_id,omitempty"`
	CurrentPeriodStart   *time.Time         `json:"current_period_start,omitempty"`
	CurrentPeriodEnd     *time.Time         `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd    bool               `json:"cancel_at_period_end"`
	CanceledAt           *time.Time         `json:"canceled_at,omitempty"`
	TrialEnd             *time.Time         `json:"trial_end,omitempty"`
	IsTrialing           bool               `json:"is_trialing"`
	TrialDaysRemaining   int                `json:"trial_days_remaining"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
	PaymentMethod        *PaymentMethodInfo `json:"payment_method,omitempty"`
}

// Handler handles billing portal and subscription management (Stripe).
type Handler struct {
	repo storage.Repository
	// Platform-fee wallet (registry credits balance for publish fees, etc.)
	platformFees *storageregistry.PlatformFeeRepository
	// Wallet service for unified wallet operations
	walletService *wallet.Service
	// State Fabric add-on entitlements (optional; nil returns empty entitlements).
	sfAddons *statefabricaddons.Repository
	// Redis client for rate limiting
	redisClient *redis.Client
	// Isolated bundle provisioner callback for one-click SaaS Starter provisioning (optional).
	// Returns (status, componentCount, error). Set via SetBundleProvisioner during server init.
	provisionBundleFn func(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (string, int, error)
}

// SetWalletService injects the unified wallet service into the billing handler.
// This allows the billing handler to read wallet info from the unified wallets table
// instead of the legacy user_wallets table.
func (h *Handler) SetWalletService(walletSvc *wallet.Service) {
	h.walletService = walletSvc
}

// NewHandler creates a new billing handler.
func NewHandler(repo storage.Repository, platformFees *storageregistry.PlatformFeeRepository, sfAddons *statefabricaddons.Repository, redisClient *redis.Client) *Handler {
	return &Handler{repo: repo, platformFees: platformFees, sfAddons: sfAddons, redisClient: redisClient}
}

// SetBundleProvisioner injects the isolated bundle provisioner into the billing handler.
// The provisioner function is called asynchronously during bundle signup/checkout.
// This is called during server initialization when TENANT_DB_ENABLED=true.
func (h *Handler) SetBundleProvisioner(fn func(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (string, int, error)) {
	h.provisionBundleFn = fn
}

// CreatePortalSessionRequest is the request body for creating a billing portal session.
type CreatePortalSessionRequest struct {
	ReturnURL string `json:"return_url"`
}

// Validate implements the ValidatedRequest interface
func (r CreatePortalSessionRequest) Validate() error {
	// Portal session requests are optional - empty is valid
	return nil
}

// CreateCheckoutSessionRequest is the request body for creating a checkout session.
type CreateCheckoutSessionRequest struct {
	PriceID    string `json:"price_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

// Validate implements the ValidatedRequest interface
func (r CreateCheckoutSessionRequest) Validate() error {
	if r.PriceID == "" {
		return fmt.Errorf("price_id is required")
	}
	return nil
}

// CreatePortalSessionResponse is the response with the Stripe portal URL.
type CreatePortalSessionResponse struct {
	URL string `json:"url"`
}

// CreateCheckoutSessionResponse is the response with the Stripe checkout URL.
type CreateCheckoutSessionResponse struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

// HandleCreatePortalSession creates a Stripe Customer Billing Portal session and returns the URL.
// POST /v1/billing/portal-session
func (h *Handler) HandleCreatePortalSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	if !payment.IsConfigured() {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Billing is not configured"))
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing portal: user not found")
		apierror.WriteError(w, apierror.NewNotFound("User not found"))
		return
	}

	var req CreatePortalSessionRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = "/settings"
	}

	// Stripe requires a full URL; build from request if path-only
	if strings.HasPrefix(returnURL, "/") {
		scheme := "https"
		if r.TLS == nil && (r.URL == nil || r.URL.Scheme == "") {
			scheme = "http"
		}
		if r.URL != nil && r.URL.Scheme != "" {
			scheme = r.URL.Scheme
		}
		returnURL = scheme + "://" + r.Host + returnURL
	}

	// Validate return URL for security (prevent open redirect)
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		logrus.Warn("APP_URL not set - using default app URL")
		appURL = "https://app.functionfly.com" // Must be set explicitly in production
	}
	returnURL = payment.SanitizeReturnURL(returnURL, appURL+"/settings")

	name := user.Name
	if name == "" && user.ProviderData != nil {
		if n, ok := user.ProviderData["name"].(string); ok {
			name = n
		}
	}
	if name == "" {
		name = user.Email
	}

	customerID, err := payment.CreateOrGetStripeCustomer(r.Context(), h.repo, claims.TenantID, user.Email, name)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("billing portal: create or get stripe customer")
		msg := "Failed to prepare billing session"
		if os.Getenv("DEVELOPMENT") == "true" {
			msg = err.Error()
		}
		writeJSONError(w, http.StatusInternalServerError, msg)
		return
	}

	url, err := payment.CreateBillingPortalSession(r.Context(), customerID, returnURL)
	if err != nil {
		logrus.WithError(err).WithField("customer_id", customerID).Error("billing portal: create session")
		msg := "Failed to create billing session"
		if os.Getenv("DEVELOPMENT") == "true" {
			msg = err.Error()
		}
		writeJSONError(w, http.StatusInternalServerError, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CreatePortalSessionResponse{URL: url})
}

// HandleCreateCheckoutSession creates a Stripe Checkout session for subscription checkout.
// POST /v1/billing/checkout
func (h *Handler) HandleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing checkout: user not found")
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	var req CreateCheckoutSessionRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	if req.PriceID == "" {
		apierror.WriteError(w, apierror.ValidationFieldError("price_id", "price_id is required"))
		return
	}

	name := user.Name
	if name == "" && user.ProviderData != nil {
		if n, ok := user.ProviderData["name"].(string); ok {
			name = n
		}
	}
	if name == "" {
		name = user.Email
	}

	resp, err := payment.CreateCheckoutSession(
		r.Context(),
		h.repo,
		claims.TenantID,
		user.Email,
		name,
		payment.CreateCheckoutSessionRequest{
			PriceID:    req.PriceID,
			SuccessURL: req.SuccessURL,
			CancelURL:  req.CancelURL,
		},
	)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("billing checkout: create checkout session")
		msg := "Failed to create checkout session"
		if os.Getenv("DEVELOPMENT") == "true" {
			msg = err.Error()
		}
		writeJSONError(w, http.StatusInternalServerError, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// writeJSONError writes a standardized JSON error response
// Deprecated: Use apierror.WriteError() directly for new code
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	err := &apierror.APIError{
		Status:  status,
		Code:    apierror.ErrCodeInternal,
		Message: msg,
	}
	apierror.WriteError(w, err)
}

// HandleGetSubscription returns the current user's subscription details.
// GET /v1/billing/subscription
func (h *Handler) HandleGetSubscription(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	subscription, err := h.repo.GetSubscriptionByTenantID(claims.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing: failed to get subscription")
		writeJSONError(w, http.StatusNotFound, "No subscription found")
		return
	}
	if subscription == nil {
		writeJSONError(w, http.StatusNotFound, "No subscription found")
		return
	}

	// Get tenant to access Stripe customer ID for payment method info
	tenant, err := h.repo.GetTenantByID(claims.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing: failed to get tenant")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve subscription details")
		return
	}

	// Build extended subscription response with payment method info
	var paymentMethod *PaymentMethodInfo
	if tenant.StripeCustomerID != nil && *tenant.StripeCustomerID != "" {
		pm, err := payment.GetPaymentMethodForCustomer(r.Context(), *tenant.StripeCustomerID)
		if err != nil {
			logrus.WithError(err).Warn("billing: failed to get payment method")
			// Don't fail the request, just don't include payment method
		} else if pm != nil {
			paymentMethod = &PaymentMethodInfo{
				Brand:    pm.Brand,
				Last4:    pm.Last4,
				ExpMonth: pm.ExpMonth,
				ExpYear:  pm.ExpYear,
			}
		}
	}

	// Convert storage.Subscription to response format (PricingTier may be nil if tier was deleted or not loaded)
	plan := ""
	if subscription.PricingTier != nil {
		plan = subscription.PricingTier.Name
	}

	// Compute trial status
	var isTrialing bool = false
	var daysRemaining int = 0
	if subscription.TrialEnd != nil {
		now := time.Now()
		isTrialing = now.Before(*subscription.TrialEnd)
		diff := subscription.TrialEnd.Sub(now)
		daysRemaining = int(diff.Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}
	} else if !isTrialing && subscription.Status == "trialing" {
		// If status says trialing but no trial_end, assume active
		isTrialing = true
		daysRemaining = 14 // Default fallback
	}

	response := SubscriptionResponse{
		ID:                   subscription.ID,
		TenantID:             subscription.TenantID,
		Plan:                 plan,
		Status:               subscription.Status,
		StripeSubscriptionID: subscription.ID.String(),
		CurrentPeriodStart:   &subscription.CurrentPeriodStart,
		CurrentPeriodEnd:     &subscription.CurrentPeriodEnd,
		CancelAtPeriodEnd:    subscription.CancelAtPeriodEnd,
		CanceledAt:           subscription.CanceledAt,
		TrialEnd:             subscription.TrialEnd,
		IsTrialing:           isTrialing,
		TrialDaysRemaining:   daysRemaining,
		CreatedAt:            subscription.CreatedAt,
		UpdatedAt:            subscription.UpdatedAt,
		PaymentMethod:        paymentMethod,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleListInvoices returns the current user's invoices.
// GET /v1/billing/invoices
func (h *Handler) HandleListInvoices(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := 10
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	invoices, err := h.repo.ListInvoicesByTenant(claims.TenantID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing: failed to list invoices")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve invoices")
		return
	}

	total, err := h.repo.CountInvoicesByTenant(claims.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing: failed to count invoices")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve invoices")
		return
	}

	out := make([]TenantInvoiceJSON, 0, len(invoices))
	for _, inv := range invoices {
		if inv == nil {
			continue
		}
		amt := inv.AmountPaidCents
		if amt <= 0 {
			amt = inv.AmountDueCents
		}
		var invDate *time.Time
		if inv.PaidAt != nil {
			invDate = inv.PaidAt
		} else {
			invDate = &inv.CreatedAt
		}
		var pdfPtr, hostedPtr *string
		if inv.InvoicePdfURL != "" {
			s := inv.InvoicePdfURL
			pdfPtr = &s
		}
		if inv.HostedInvoiceURL != "" {
			s := inv.HostedInvoiceURL
			hostedPtr = &s
		}
		curr := strings.ToLower(strings.TrimSpace(inv.Currency))
		if curr == "" {
			curr = "usd"
		}
		out = append(out, TenantInvoiceJSON{
			ID:               inv.ID.String(),
			TenantID:         inv.TenantID.String(),
			StripeInvoiceID:  inv.StripeInvoiceID,
			Amount:           amt,
			Currency:         curr,
			Status:           inv.Status,
			InvoiceDate:      invDate,
			DueDate:          inv.DueDate,
			InvoicePDF:       pdfPtr,
			HostedInvoiceURL: hostedPtr,
			CreatedAt:        inv.CreatedAt,
		})
	}

	response := struct {
		Invoices []TenantInvoiceJSON `json:"invoices"`
		Limit    int                 `json:"limit"`
		Offset   int                 `json:"offset"`
		Total    int                 `json:"total"`
	}{
		Invoices: out,
		Limit:    limit,
		Offset:   offset,
		Total:    total,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleGetUsage returns the current user's usage details.
// GET /v1/billing/usage
func (h *Handler) HandleGetUsage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get usage from the past 30 days by default
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -30)

	if s := r.URL.Query().Get("start"); s != "" {
		if parsed, err := time.Parse("2006-01-02", s); err == nil {
			start = parsed
		}
	}
	if e := r.URL.Query().Get("end"); e != "" {
		if parsed, err := time.Parse("2006-01-02", e); err == nil {
			end = parsed
		}
	}

	usage, err := h.repo.GetUsageByTenant(claims.TenantID, "", start, end)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing: failed to get usage")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve usage")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"usage": usage,
		"start": start.Format("2006-01-02"),
		"end":   end.Format("2006-01-02"),
	})
}

// CancelSubscriptionRequest is the request body for cancelling a subscription
type CancelSubscriptionRequest struct {
	Immediately bool `json:"immediately"`
}

// Validate implements the ValidatedRequest interface
func (r CancelSubscriptionRequest) Validate() error {
	// Cancel subscription requests are valid with any values
	return nil
}

// HandleCancelSubscription cancels the current user's subscription.
// POST /v1/billing/subscription/cancel
func (h *Handler) HandleCancelSubscription(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	// Get subscription to find Stripe subscription ID
	subscription, err := h.repo.GetSubscriptionByTenantID(claims.TenantID)
	if err != nil || subscription == nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing cancel: no subscription found")
		writeJSONError(w, http.StatusNotFound, "No active subscription found")
		return
	}

	// Parse request body
	var req CancelSubscriptionRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	// Cancel the subscription via the repository
	err = h.repo.CancelSubscription(r.Context(), subscription.ID)
	if err != nil {
		logrus.WithError(err).WithField("subscription_id", subscription.ID).Error("billing cancel: failed to cancel subscription")
		msg := "Failed to cancel subscription"
		if os.Getenv("DEVELOPMENT") == "true" {
			msg = err.Error()
		}
		writeJSONError(w, http.StatusInternalServerError, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Subscription cancelled successfully",
	})
}

// HandleSubscriptionWebhook processes subscription updates from webhooks (e.g., Stripe).
// This is called by webhook handlers when a subscription is created or updated.
// POST /v1/billing/subscription/webhook (internal use)
// Requires INTERNAL_WEBHOOK_SECRET header for authentication.
func (h *Handler) HandleSubscriptionWebhook(w http.ResponseWriter, r *http.Request) {
	// Verify internal webhook authentication
	internalSecret := os.Getenv("INTERNAL_WEBHOOK_SECRET")
	if internalSecret != "" {
		authHeader := r.Header.Get("X-Internal-Webhook-Secret")
		if authHeader != internalSecret {
			logrus.Warn("billing webhook: invalid or missing internal webhook secret")
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
	} else {
		// In development mode without secret configured, allow localhost requests only
		host := r.Host
		remoteAddr := r.RemoteAddr
		if !strings.Contains(host, "localhost") && !strings.Contains(remoteAddr, "127.0.0.1") && !strings.Contains(remoteAddr, "[::1]") {
			logrus.Warn("billing webhook: rejected non-localhost request without INTERNAL_WEBHOOK_SECRET")
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized - INTERNAL_WEBHOOK_SECRET not configured")
			return
		}
	}

	var req struct {
		TenantID   uuid.UUID `json:"tenant_id"`
		OldPlan    string    `json:"old_plan"`
		NewPlan    string    `json:"new_plan"`
		UserID     uuid.UUID `json:"user_id"`
		UpgradedBy uuid.UUID `json:"upgraded_by"`
		UpgradedAt time.Time `json:"upgraded_at"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.TenantID == uuid.Nil || req.NewPlan == "" {
		writeJSONError(w, http.StatusBadRequest, "tenant_id and new_plan are required")
		return
	}

	// Get tenant users to process upgrade for all users in the tenant
	users, err := h.repo.ListActiveUsersByTenant(r.Context(), req.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", req.TenantID).Warn("billing webhook: failed to list users for plan upgrade")
		writeJSONError(w, http.StatusInternalServerError, "Failed to process subscription update")
		return
	}

	// Determine who performed the upgrade
	adminUserID := req.UpgradedBy
	if adminUserID == uuid.Nil && req.UserID != uuid.Nil {
		adminUserID = req.UserID
	}
	if adminUserID == uuid.Nil && len(users) > 0 {
		adminUserID = users[0].ID
	}

	upgradedAt := req.UpgradedAt
	if upgradedAt.IsZero() {
		upgradedAt = time.Now()
	}

	// Process upgrade for each user in the tenant
	for _, user := range users {
		activity := &storage.UserActivity{
			UserID:       user.ID,
			ActivityType: "membership_upgraded",
			Title:        fmt.Sprintf("Upgraded to %s", formatPlanName(req.NewPlan)),
			Description:  getUpgradeDescription(req.NewPlan),
			Metadata: map[string]interface{}{
				"plan":         req.NewPlan,
				"previousPlan": req.OldPlan,
				"upgradedAt":   upgradedAt.Format(time.RFC3339),
			},
			IsPublic: true,
		}

		if err := h.repo.CreateUserActivity(activity); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"user_id":   user.ID,
				"tenant_id": req.TenantID,
			}).Warn("billing webhook: failed to create membership upgrade activity")
		}

		// Award enterprise achievement if upgrading to enterprise
		if isEnterprisePlan(req.NewPlan) && !isEnterprisePlan(req.OldPlan) {
			if err := awardEnterpriseAchievement(h.repo, user.ID); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"user_id":   user.ID,
					"tenant_id": req.TenantID,
				}).Warn("billing webhook: failed to award enterprise achievement")
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":  req.TenantID,
		"old_plan":   req.OldPlan,
		"new_plan":   req.NewPlan,
		"user_count": len(users),
	}).Info("billing webhook: processed subscription plan upgrade")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "success",
		"user_count": fmt.Sprintf("%d", len(users)),
	})
}

// formatPlanName formats a plan name for display.
func formatPlanName(plan string) string {
	switch plan {
	case "enterprise", "Enterprise":
		return "Enterprise"
	case "professional", "pro", "Pro":
		return "Professional"
	case "starter", "Starter":
		return "Starter"
	case "free", "Free":
		return "Free"
	default:
		if len(plan) > 0 {
			return string(plan[0]-32) + plan[1:]
		}
		return plan
	}
}

// HandleListPaymentMethods returns all payment methods for the current user's Stripe customer.
// GET /v1/billing/payment-methods
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
		writeJSONError(w, http.StatusInternalServerError, "Failed to get tenant")
		return
	}

	if tenant.StripeCustomerID == nil || *tenant.StripeCustomerID == "" {
		writeJSONError(w, http.StatusNotFound, "No Stripe customer found")
		return
	}

	methods, err := payment.ListPaymentMethodsForCustomer(r.Context(), *tenant.StripeCustomerID)
	if err != nil {
		logrus.WithError(err).WithField("customer_id", *tenant.StripeCustomerID).Warn("billing: failed to list payment methods")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list payment methods")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payment_methods": methods,
	})
}

// HandleCreateSetupIntent creates a Stripe SetupIntent for client-side payment method collection.
// POST /v1/billing/payment-methods/setup-intent
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

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	name := user.Name
	if name == "" && user.ProviderData != nil {
		if n, ok := user.ProviderData["name"].(string); ok {
			name = n
		}
	}
	if name == "" {
		name = user.Email
	}

	customerID, err := payment.CreateOrGetStripeCustomer(r.Context(), h.repo, claims.TenantID, user.Email, name)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("setup intent: create or get stripe customer")
		writeJSONError(w, http.StatusInternalServerError, "Failed to prepare billing")
		return
	}

	result, err := payment.CreateSetupIntent(r.Context(), customerID)
	if err != nil {
		logrus.WithError(err).WithField("customer_id", customerID).Error("setup intent: create")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create setup intent")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleSetDefaultPaymentMethod sets a payment method as the default for the customer.
// POST /v1/billing/payment-methods/default
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

	tenant, err := h.repo.GetTenantByID(claims.TenantID)
	if err != nil || tenant == nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to get tenant")
		return
	}

	if tenant.StripeCustomerID == nil || *tenant.StripeCustomerID == "" {
		writeJSONError(w, http.StatusNotFound, "No Stripe customer found")
		return
	}

	var req struct {
		PaymentMethodID string `json:"payment_method_id"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
	}
	if req.PaymentMethodID == "" {
		writeJSONError(w, http.StatusBadRequest, "payment_method_id is required")
		return
	}

	err = payment.SetDefaultPaymentMethod(r.Context(), *tenant.StripeCustomerID, req.PaymentMethodID)
	if err != nil {
		logrus.WithError(err).WithField("customer_id", *tenant.StripeCustomerID).Warn("billing: failed to set default payment method")
		writeJSONError(w, http.StatusInternalServerError, "Failed to set default payment method")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Default payment method updated"})
}

// HandleDetachPaymentMethod detaches a payment method from the customer.
// DELETE /v1/billing/payment-methods/:id
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

	// Extract payment method ID from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var paymentMethodID string
	for i, part := range parts {
		if part == "payment-methods" && i+1 < len(parts) {
			paymentMethodID = parts[i+1]
			break
		}
	}
	if paymentMethodID == "" {
		writeJSONError(w, http.StatusBadRequest, "Payment method ID is required")
		return
	}

	err := payment.DetachPaymentMethod(r.Context(), paymentMethodID)
	if err != nil {
		logrus.WithError(err).WithField("payment_method_id", paymentMethodID).Warn("billing: failed to detach payment method")
		writeJSONError(w, http.StatusInternalServerError, "Failed to remove payment method")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Payment method removed"})
}

// getUpgradeDescription returns a description based on the plan tier.
func getUpgradeDescription(plan string) string {
	switch plan {
	case "enterprise":
		return "Unlimited functions, dedicated support, and premium enterprise features"
	case "professional", "pro":
		return "Advanced features, priority support, and increased limits"
	case "starter":
		return "Expanded features and higher execution limits"
	default:
		return "Membership upgraded with new features and benefits"
	}
}

// isEnterprisePlan checks if a plan name represents an enterprise tier.
func isEnterprisePlan(plan string) bool {
	return plan == "enterprise" || plan == "Enterprise"
}

// HandleGetMyAffiliateCodes returns the affiliate codes belonging to the authenticated user.
// GET /v1/affiliate/my-codes
func (h *Handler) HandleGetMyAffiliateCodes(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	codes, err := h.repo.ListAffiliateCodesByPublisher(claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list affiliate codes for user")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve affiliate codes")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"affiliate_codes": codes,
	})
}

// HandleGetMyAffiliateCommissions returns commissions for the authenticated user's affiliate codes.
// GET /v1/affiliate/my-commissions
func (h *Handler) HandleGetMyAffiliateCommissions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	commissions, err := h.repo.ListAffiliateCommissionsByPublisher(claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list affiliate commissions for user")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve commissions")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"commissions": commissions,
	})
}

// HandleGetMyAffiliateReferrals returns referrals for the authenticated user's affiliate codes.
// GET /v1/affiliate/referrals
func (h *Handler) HandleGetMyAffiliateReferrals(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	codes, err := h.repo.ListAffiliateCodesByPublisher(claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list affiliate codes for user")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve referrals")
		return
	}

	var allReferrals []*storage.AffiliateReferral
	for _, code := range codes {
		referrals, err := h.repo.ListAffiliateReferralsByCode(code.ID)
		if err != nil {
			logrus.WithError(err).WithField("code_id", code.ID).Warn("Failed to list referrals for code")
			continue
		}
		allReferrals = append(allReferrals, referrals...)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"referrals": allReferrals,
	})
}

// HandleGetAffiliateEarningsSummary returns earnings summary for the authenticated user.
// GET /v1/affiliate/earnings-summary
func (h *Handler) HandleGetAffiliateEarningsSummary(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	codes, err := h.repo.ListAffiliateCodesByPublisher(claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list affiliate codes for user")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve earnings summary")
		return
	}

	var totalPendingEarningsCents int64
	var totalEarningsCents int64
	var totalPaidOutCents int64
	var totalReferrals int
	var pendingCommissions int

	for _, code := range codes {
		totalPendingEarningsCents += code.PendingEarningsCents
		totalEarningsCents += code.TotalEarningsCents
		totalPaidOutCents += code.PaidOutEarningsCents
		totalReferrals += code.TotalReferrals
		pendingCommissions += code.PendingCommissions
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pending_earnings_cents": totalPendingEarningsCents,
		"total_earnings_cents":   totalEarningsCents,
		"paid_out_cents":         totalPaidOutCents,
		"total_referrals":        totalReferrals,
		"pending_commissions":    pendingCommissions,
		"codes_count":           len(codes),
	})
}

// HandleApplyAffiliateCode applies an affiliate code during registration.
// POST /v1/affiliate/apply-code
func (h *Handler) HandleApplyAffiliateCode(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req struct {
		Code        string `json:"code"`
		UTMSource   string `json:"utm_source,omitempty"`
		UTMCampaign string `json:"utm_campaign,omitempty"`
		UTContent   string `json:"utm_content,omitempty"`
		UTMTerm     string `json:"utm_term,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Code == "" {
		writeJSONError(w, http.StatusBadRequest, "Affiliate code is required")
		return
	}

	code, err := h.repo.GetAffiliateCodeByCode(req.Code)
	if err != nil {
		logrus.WithError(err).Error("Failed to look up affiliate code")
		writeJSONError(w, http.StatusInternalServerError, "Failed to validate affiliate code")
		return
	}

	if code == nil {
		writeJSONError(w, http.StatusNotFound, "Affiliate code not found or inactive")
		return
	}

	// Store the applied code association for the user
	// This could be stored in user preferences or a separate affiliate_applications table
	// For now, we just return success if the code is valid
	_ = claims // Mark as used to avoid compile warning

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":          true,
		"code":             code.Code,
		"name":             code.Name,
		"commission_type":  code.CommissionType,
		"commission_value": code.CommissionValue,
	})
}

// awardEnterpriseAchievement awards the "Enterprise Pioneer" achievement to a user.
func awardEnterpriseAchievement(repo storage.Repository, userID uuid.UUID) error {
	achievement, err := repo.GetAchievementBySlug("enterprise_pioneer")
	if err != nil {
		return fmt.Errorf("failed to get enterprise pioneer achievement: %w", err)
	}

	if achievement == nil {
		logrus.Warn("Enterprise Pioneer achievement not found in database - skipping award")
		return nil
	}

	// Check if user already has this achievement
	existingAchievements, err := repo.GetUserAchievements(userID)
	if err != nil {
		return fmt.Errorf("failed to check existing achievements: %w", err)
	}

	for _, ea := range existingAchievements {
		if ea.AchievementID == achievement.ID {
			return nil // User already has this achievement
		}
	}

	metadata := map[string]interface{}{
		"awarded_at": time.Now().Format(time.RFC3339),
		"reason":     "Upgraded to Enterprise plan",
	}

	if err := repo.AwardAchievement(userID, achievement.ID, metadata); err != nil {
		return fmt.Errorf("failed to award achievement: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":        userID,
		"achievement_id": achievement.ID,
	}).Info("Awarded Enterprise Pioneer achievement via billing")

	return nil
}
