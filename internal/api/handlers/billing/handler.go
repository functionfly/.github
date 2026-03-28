package billing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	"github.com/functionfly/functionfly/internal/storage"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
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
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
	PaymentMethod        *PaymentMethodInfo `json:"payment_method,omitempty"`
}

// Handler handles billing portal and subscription management (Stripe).
type Handler struct {
	repo storage.Repository
	// Platform-fee wallet (registry credits balance for publish fees, etc.)
	platformFees *storageregistry.PlatformFeeRepository
	// State Fabric add-on entitlements (optional; nil returns empty entitlements).
	sfAddons *statefabricaddons.Repository
}

// NewHandler creates a new billing handler.
func NewHandler(repo storage.Repository, platformFees *storageregistry.PlatformFeeRepository, sfAddons *statefabricaddons.Repository) *Handler {
	return &Handler{repo: repo, platformFees: platformFees, sfAddons: sfAddons}
}

// CreatePortalSessionRequest is the request body for creating a billing portal session.
type CreatePortalSessionRequest struct {
	ReturnURL string `json:"return_url"`
}

// CreateCheckoutSessionRequest is the request body for creating a checkout session.
type CreateCheckoutSessionRequest struct {
	PriceID    string `json:"price_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
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
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing portal: user not found")
		writeJSONError(w, http.StatusNotFound, "User not found")
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
		writeJSONError(w, http.StatusBadRequest, "price_id is required")
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

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
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
	response := SubscriptionResponse{
		ID:                   subscription.ID,
		TenantID:             subscription.TenantID,
		Plan:                 plan,
		Status:               subscription.Status,
		StripeSubscriptionID: subscription.ID.String(), // Use subscription ID as reference
		CurrentPeriodStart:   &subscription.CurrentPeriodStart,
		CurrentPeriodEnd:     &subscription.CurrentPeriodEnd,
		CancelAtPeriodEnd:    subscription.CancelAtPeriodEnd,
		CanceledAt:           subscription.CanceledAt,
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
		TenantID     uuid.UUID `json:"tenant_id"`
		OldPlan      string    `json:"old_plan"`
		NewPlan      string    `json:"new_plan"`
		UserID       uuid.UUID `json:"user_id"`
		UpgradedBy   uuid.UUID `json:"upgraded_by"`
		UpgradedAt   time.Time `json:"upgraded_at"`
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
