package trustapi

import (
	"context"
	"fmt"
	"os"
	"time"

	storagetrustapi "github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/stripe/stripe-go/v83/customer"
	"github.com/stripe/stripe-go/v83/subscription"
)

// BillingService handles Trust API partner billing
type BillingService struct {
	repo        *storagetrustapi.BillingRepository
	stripeKey   string
	environment string
}

// NewBillingService creates a new billing service
func NewBillingService(repo *storagetrustapi.BillingRepository) *BillingService {
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	return &BillingService{
		repo:        repo,
		stripeKey:   stripeKey,
		environment: env,
	}
}

// IsStripeConfigured returns whether Stripe is configured
func (s *BillingService) IsStripeConfigured() bool {
	return s.stripeKey != ""
}

// ============================================
// Tier Pricing
// ============================================

// InitializeTierPricing sets up default pricing tiers if they don't exist
func (s *BillingService) InitializeTierPricing(ctx context.Context) error {
	tiers := []struct {
		tier                string
		monthlyPriceCents   int
		includedRequests    int
		overagePricePer1000 int
		hasOverageBilling   bool
		rateLimitPerMinute  int
		rateLimitPerDay     int
		monthlyRequestLimit int
		description         string
	}{
		{
			tier:                "developer",
			monthlyPriceCents:   0,
			includedRequests:    50000,
			overagePricePer1000: 0,
			hasOverageBilling:   false,
			rateLimitPerMinute:  60,
			rateLimitPerDay:     10000,
			monthlyRequestLimit: 50000,
			description:         "Free tier for developers - 50K requests/month, hard limit",
		},
		{
			tier:                "startup",
			monthlyPriceCents:   4900,
			includedRequests:    500000,
			overagePricePer1000: 5,
			hasOverageBilling:   true,
			rateLimitPerMinute:  300,
			rateLimitPerDay:     100000,
			monthlyRequestLimit: 500000,
			description:         "Startup tier - $49/mo for 500K requests, $0.005 per overage",
		},
		{
			tier:                "business",
			monthlyPriceCents:   19900,
			includedRequests:    2000000,
			overagePricePer1000: 3,
			hasOverageBilling:   true,
			rateLimitPerMinute:  1000,
			rateLimitPerDay:     500000,
			monthlyRequestLimit: 2000000,
			description:         "Business tier - $199/mo for 2M requests, $0.003 per overage",
		},
		{
			tier:                "enterprise",
			monthlyPriceCents:   0,
			includedRequests:    0,
			overagePricePer1000: 0,
			hasOverageBilling:   false,
			rateLimitPerMinute:  10000,
			rateLimitPerDay:     10000000,
			monthlyRequestLimit: 100000000,
			description:         "Enterprise tier - Custom pricing, contact sales",
		},
	}

	for _, t := range tiers {
		pricing := &storagetrustapi.PartnerTierPricing{
			Tier:                t.tier,
			MonthlyPriceCents:   t.monthlyPriceCents,
			IncludedRequests:    t.includedRequests,
			OveragePricePer1000: t.overagePricePer1000,
			HasOverageBilling:   t.hasOverageBilling,
			RateLimitPerMinute:  t.rateLimitPerMinute,
			RateLimitPerDay:     t.rateLimitPerDay,
			MonthlyRequestLimit: t.monthlyRequestLimit,
			Description:         t.description,
			IsActive:            true,
		}

		if err := s.repo.UpsertTierPricing(ctx, pricing); err != nil {
			return fmt.Errorf("failed to upsert tier pricing for %s: %w", t.tier, err)
		}
	}

	return nil
}

// GetTierPricing retrieves pricing for a specific tier
func (s *BillingService) GetTierPricing(ctx context.Context, tier string) (*storagetrustapi.PartnerTierPricing, error) {
	return s.repo.GetTierPricing(ctx, tier)
}

// ListTierPricing lists all active tier pricing
func (s *BillingService) ListTierPricing(ctx context.Context) ([]storagetrustapi.PartnerTierPricing, error) {
	return s.repo.ListTierPricing(ctx)
}

// ============================================
// Stripe Integration
// ============================================

// CreateStripeCustomer creates a Stripe customer for a partner
func (s *BillingService) CreateStripeCustomer(ctx context.Context, partnerID uuid.UUID, email, name string) (string, error) {
	if !s.IsStripeConfigured() {
		return "", fmt.Errorf("Stripe is not configured")
	}

	partner, err := s.repo.GetPartnerByID(partnerID)
	if err != nil {
		return "", fmt.Errorf("failed to get partner: %w", err)
	}

	// Check if customer already exists
	if partner.StripeCustomerID != "" {
		// Verify the customer still exists in Stripe
		_, err := customer.Get(partner.StripeCustomerID, nil)
		if err == nil {
			return partner.StripeCustomerID, nil
		}
		// Customer doesn't exist in Stripe, create new one
	}

	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
		Metadata: map[string]string{
			"partner_id":   partnerID.String(),
			"partner_slug": partner.Slug,
			"tier":         partner.Tier,
		},
	}

	c, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	// Save the Stripe customer ID
	partner.StripeCustomerID = c.ID
	if err := s.repo.UpdatePartner(partner); err != nil {
		return "", fmt.Errorf("failed to update partner with Stripe customer ID: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"partner_id":         partnerID,
		"stripe_customer_id": c.ID,
	}).Info("Created Stripe customer for partner")

	return c.ID, nil
}

// CreateCheckoutSession creates a Stripe checkout session for tier upgrade
func (s *BillingService) CreateCheckoutSession(ctx context.Context, partnerID uuid.UUID, tier, successURL, cancelURL string) (*CheckoutSessionResult, error) {
	if !s.IsStripeConfigured() {
		return nil, fmt.Errorf("Stripe is not configured")
	}

	partner, err := s.repo.GetPartnerByID(partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner: %w", err)
	}

	// Get tier pricing
	pricing, err := s.GetTierPricing(ctx, tier)
	if err != nil {
		return nil, fmt.Errorf("failed to get tier pricing: %w", err)
	}

	if pricing.MonthlyPriceCents == 0 {
		return nil, fmt.Errorf("cannot create checkout for free tier")
	}

	// Ensure Stripe customer exists
	customerID := partner.StripeCustomerID
	if customerID == "" {
		customerID, err = s.CreateStripeCustomer(ctx, partnerID, partner.ContactEmail, partner.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to create Stripe customer: %w", err)
		}
	}

	// Build line items
	var lineItems []*stripe.CheckoutSessionLineItemParams

	// Base subscription
	if pricing.StripePriceID != "" {
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(pricing.StripePriceID),
			Quantity: stripe.Int64(1),
		})
	} else {
		// Create ad-hoc price if not pre-configured
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String(string(stripe.CurrencyUSD)),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name:        stripe.String(fmt.Sprintf("Trust API %s Plan", tier)),
					Description: stripe.String(pricing.Description),
				},
				UnitAmount: stripe.Int64(int64(pricing.MonthlyPriceCents)),
				Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
					Interval: stripe.String("month"),
				},
			},
			Quantity: stripe.Int64(1),
		})
	}

	// Add metered billing for overages if applicable
	if pricing.HasOverageBilling && pricing.OveragePricePer1000 > 0 {
		// Metered billing prices are pre-created in Stripe and stored in tier config.
		// These prices have recurring[usage_type]=metered and a billing meter attached.
		//
		// STRIPE SETUP (already done via CLI):
		//   # 1. Create billing meter for tracking overages
		//   stripe billing meters create --event-name="trust_api_overage" \
		//     --display-name="Trust API Overage Requests" \
		//     --customer-mapping.type="by_id" \
		//     --customer-mapping.event-payload-key="stripe_customer_id" \
		//     --value-settings.event-payload-key="value" \
		//     --default-aggregation.formula="sum"
		//   # Meter ID: mtr_test_61UUsbycP39uDpgl841Kxe78JyppiD6e
		//
		//   # 2. Create metered prices for overage billing
		//   # Startup: $0.005 per 1000 requests (5 cents / 1000)
		//   stripe prices create --unit-amount=5 --currency=usd \
		//     --recurring.interval=month --recurring.usage-type=metered \
		//     --recurring.meter="mtr_test_61UUsbycP39uDpgl841Kxe78JyppiD6e" \
		//     -d "product_data[name]"="Trust API Overage - Startup Tier"
		//   # Price ID: price_1TLUewKxe78JyppibuckHSro
		//
		//   # Business: $0.003 per 1000 requests (3 cents / 1000)
		//   stripe prices create --unit-amount=3 --currency=usd \
		//     --recurring.interval=month --recurring.usage-type=metered \
		//     --recurring.meter="mtr_test_61UUsbycP39uDpgl841Kxe78JyppiD6e" \
		//     --transform-quantity.divide-by=1000 --transform-quantity.round=up \
		//     -d "product_data[name]"="Trust API Overage - Business Tier"
		//   # Price ID: price_1TLUf2Kxe78JyppiqC0y7kbq
		//
		//   # 3. Create base subscription prices (licensed)
		//   # Startup: $49/month
		//   stripe prices create --unit-amount=4900 --currency=usd \
		//     --recurring.interval=month \
		//     -d "product_data[name]"="Trust API - Startup Tier"
		//   # Price ID: price_1TLUf7Kxe78Jyppidoow5T4w
		//
		//   # Business: $199/month
		//   stripe prices create --unit-amount=19900 --currency=usd \
		//     --recurring.interval=month \
		//     -d "product_data[name]"="Trust API - Business Tier"
		//   # Price ID: price_1TLUfAKxe78Jyppi1bUAn1xW
		//
		// TEST METER EVENTS with Stripe CLI:
		//   stripe billing meter_events create \
		//     --event-name="trust_api_overage" \
		//     --payload="stripe_customer_id=cus_xxx,value=1000"
		if pricing.StripeMeteredPriceID != "" {
			lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
				Price:    stripe.String(pricing.StripeMeteredPriceID),
				Quantity: stripe.Int64(0), // Metered billing starts at 0, accumulates via meter events
			})
		} else {
			logrus.WithField("tier", tier).Warn(
				"No StripeMeteredPriceID configured for tier with overage billing. " +
					"Run Stripe CLI setup commands (documented in code) and update tier config.")
		}
	}

	// Default URLs if not provided
	if successURL == "" {
		successURL = fmt.Sprintf("http://localhost:3000/partners/%s/billing?success=true", partnerID)
	}
	if cancelURL == "" {
		cancelURL = fmt.Sprintf("http://localhost:3000/partners/%s/billing?canceled=true", partnerID)
	}

	params := &stripe.CheckoutSessionParams{
		Customer:   stripe.String(customerID),
		LineItems:  lineItems,
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"partner_id":   partnerID.String(),
				"partner_slug": partner.Slug,
				"tier":         tier,
				"type":         "trust_api_subscription",
			},
		},
	}

	session, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	return &CheckoutSessionResult{
		SessionID: session.ID,
		URL:       session.URL,
		Status:    string(session.Status),
	}, nil
}

// CheckoutSessionResult holds the result of a checkout session creation
type CheckoutSessionResult struct {
	SessionID string
	URL       string
	Status    string
}

// HandleCheckoutSuccess processes a successful checkout
func (s *BillingService) HandleCheckoutSuccess(ctx context.Context, sessionID string) error {
	session, err := session.Get(sessionID, nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve checkout session: %w", err)
	}

	if session.Status != stripe.CheckoutSessionStatusComplete {
		return fmt.Errorf("checkout session is not complete")
	}

	partnerIDStr := session.Subscription.Metadata["partner_id"]
	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		return fmt.Errorf("invalid partner ID in metadata: %w", err)
	}

	partner, err := s.repo.GetPartnerByID(partnerID)
	if err != nil {
		return fmt.Errorf("failed to get partner: %w", err)
	}

	// Update partner with subscription info
	partner.StripeSubscriptionID = session.Subscription.ID
	partner.StripePriceID = session.Subscription.Items.Data[0].Price.ID
	partner.BillingStatus = string(storagetrustapi.BillingStatusActive)
	partner.Tier = session.Subscription.Metadata["tier"]

	if err := s.repo.UpdatePartner(partner); err != nil {
		return fmt.Errorf("failed to update partner: %w", err)
	}

	// Reset billing usage for the new period
	if err := s.repo.ResetPartnerBillingUsage(ctx, partnerID); err != nil {
		logrus.WithError(err).Warn("Failed to reset billing usage after subscription")
	}

	logrus.WithFields(logrus.Fields{
		"partner_id":      partnerID,
		"subscription_id": session.Subscription.ID,
		"tier":            partner.Tier,
	}).Info("Partner subscribed to Trust API plan")

	return nil
}

// ============================================
// Usage Tracking & Metered Billing
// ============================================

// RecordUsage records API usage for billing
func (s *BillingService) RecordUsage(ctx context.Context, partnerID uuid.UUID, requestCount int) error {
	partner, err := s.repo.GetPartnerByID(partnerID)
	if err != nil {
		return fmt.Errorf("failed to get partner: %w", err)
	}

	// Get tier pricing
	pricing, err := s.GetTierPricing(ctx, partner.Tier)
	if err != nil {
		return fmt.Errorf("failed to get tier pricing: %w", err)
	}

	// Update usage in database
	usage, err := s.repo.GetOrCreateBillingUsage(ctx, partnerID)
	if err != nil {
		return fmt.Errorf("failed to get or create billing usage: %w", err)
	}

	// Calculate overage
	newTotalUsage := usage.RequestsThisPeriod + requestCount
	var overageCount int

	if newTotalUsage > pricing.IncludedRequests && pricing.HasOverageBilling {
		// Calculate how many of these requests are overages
		previousOverage := usage.OveragesThisPeriod
		newOverage := newTotalUsage - pricing.IncludedRequests
		if newOverage > 0 {
			overageCount = newOverage - previousOverage
			if overageCount > requestCount {
				overageCount = requestCount
			}
		}
	}

	// Update usage
	usage.RequestsThisPeriod = newTotalUsage
	usage.OveragesThisPeriod += overageCount

	if err := s.repo.UpdateBillingUsage(ctx, usage); err != nil {
		return fmt.Errorf("failed to update billing usage: %w", err)
	}

	// Report overage to Stripe for metered billing
	if overageCount > 0 && partner.StripeSubscriptionID != "" && s.IsStripeConfigured() {
		if err := s.reportOverageToStripe(partner.StripeSubscriptionID, overageCount); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"partner_id":    partnerID,
				"overage_count": overageCount,
			}).Warn("Failed to report overage to Stripe")
			// Don't fail the request, just log the error
		}
	}

	// Update partner's cached usage
	partner.CurrentMonthUsage = newTotalUsage
	partner.CurrentOverageUsage = usage.OveragesThisPeriod

	if err := s.repo.UpdatePartner(partner); err != nil {
		logrus.WithError(err).Warn("Failed to update partner usage cache")
	}

	return nil
}

// reportOverageToStripe reports usage to Stripe using the Billing Meter Events API.
// Requires: billing meter "trust_api_overage" created via Stripe CLI (see CreateCheckoutSession comments)
// Usage is reported in real-time and aggregated by Stripe for invoicing.
func (s *BillingService) reportOverageToStripe(subscriptionID string, quantity int) error {
	if !s.IsStripeConfigured() {
		return fmt.Errorf("stripe is not configured")
	}

	// Get subscription to find the customer ID
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	customerID := sub.Customer.ID

	// Get the meter event name from environment or use default
	meterEventName := os.Getenv("STRIPE_OVERAGE_METER_NAME")
	if meterEventName == "" {
		meterEventName = "trust_api_overage"
	}

	// Create Stripe client for new API services
	client := stripe.NewClient(s.stripeKey)

	// Create billing meter event
	// Note: This requires a billing meter to be pre-created in Stripe with:
	// - event_name: "trust_api_overage" (or your custom name)
	// - customer_mapping: {event_payload_key: "stripe_customer_id"}
	// - value_settings: {event_payload_key: "value"}
	//
	// To test with Stripe CLI:
	//   stripe billing meters create --event-name="trust_api_overage" \
	//     --customer-mapping="stripe_customer_id" --value-settings="value"
	//
	// Then send test events:
	//   stripe billing meter_events create --event-name="trust_api_overage" \
	//     --payload="stripe_customer_id=cus_xxx,value=100"
	params := &stripe.BillingMeterEventCreateParams{
		EventName:  stripe.String(meterEventName),
		Identifier: stripe.String(fmt.Sprintf("overage_%s_%d_%d", subscriptionID, time.Now().Unix(), quantity)),
		Payload: map[string]string{
			"stripe_customer_id": customerID,
			"value":              fmt.Sprintf("%d", quantity),
		},
		Timestamp: stripe.Int64(time.Now().Unix()),
	}

	_, err = client.V1BillingMeterEvents.Create(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to create billing meter event: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"subscription_id":  subscriptionID,
		"customer_id":      customerID,
		"meter_event_name": meterEventName,
		"quantity":         quantity,
	}).Info("Reported overage usage to Stripe Billing Meter")

	return nil
}

// GetCurrentUsage retrieves current billing usage for a partner
func (s *BillingService) GetCurrentUsage(ctx context.Context, partnerID uuid.UUID) (*UsageInfo, error) {
	partner, err := s.repo.GetPartnerByID(partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner: %w", err)
	}

	pricing, err := s.GetTierPricing(ctx, partner.Tier)
	if err != nil {
		return nil, fmt.Errorf("failed to get tier pricing: %w", err)
	}

	usage, err := s.repo.GetOrCreateBillingUsage(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing usage: %w", err)
	}

	// Calculate charges
	overageRequests := usage.OveragesThisPeriod
	var overageChargeCents int
	if pricing.HasOverageBilling && overageRequests > 0 {
		// Charge per 1000 requests
		units := (overageRequests + 999) / 1000 // Round up
		overageChargeCents = units * pricing.OveragePricePer1000
	}

	return &UsageInfo{
		PartnerID:          partnerID,
		Tier:               partner.Tier,
		MonthlyPriceCents:  pricing.MonthlyPriceCents,
		IncludedRequests:   pricing.IncludedRequests,
		CurrentUsage:       usage.RequestsThisPeriod,
		RemainingRequests:  max(0, pricing.IncludedRequests-usage.RequestsThisPeriod),
		OverageRequests:    overageRequests,
		OverageChargeCents: overageChargeCents,
		BillingPeriodStart: usage.BillingPeriodStart,
		BillingPeriodEnd:   usage.BillingPeriodEnd,
		IsHardLimit:        !pricing.HasOverageBilling,
	}, nil
}

// UsageInfo contains current usage information
type UsageInfo struct {
	PartnerID          uuid.UUID `json:"partner_id"`
	Tier               string    `json:"tier"`
	MonthlyPriceCents  int       `json:"monthly_price_cents"`
	IncludedRequests   int       `json:"included_requests"`
	CurrentUsage       int       `json:"current_usage"`
	RemainingRequests  int       `json:"remaining_requests"`
	OverageRequests    int       `json:"overage_requests"`
	OverageChargeCents int       `json:"overage_charge_cents"`
	BillingPeriodStart time.Time `json:"billing_period_start"`
	BillingPeriodEnd   time.Time `json:"billing_period_end"`
	IsHardLimit        bool      `json:"is_hard_limit"`
}

// ============================================
// Founder Mode
// ============================================

// EnrollFounderMode enrolls a partner in founder mode (free tier with limits)
func (s *BillingService) EnrollFounderMode(ctx context.Context, partnerID uuid.UUID, usageThreshold, freeDays int) error {
	partner, err := s.repo.GetPartnerByID(partnerID)
	if err != nil {
		return fmt.Errorf("failed to get partner: %w", err)
	}

	if partner.IsFounderMode {
		return fmt.Errorf("partner is already in founder mode")
	}

	// Set defaults
	if usageThreshold <= 0 {
		usageThreshold = 100000 // 100K requests default
	}
	if freeDays <= 0 {
		freeDays = 90 // 90 days default
	}

	now := time.Now().UTC()
	endsAt := now.AddDate(0, 0, freeDays)

	partner.IsFounderMode = true
	partner.FounderModeStartedAt = &now
	partner.FounderModeEndsAt = &endsAt
	partner.UsageThreshold = usageThreshold
	partner.BillingStatus = string(storagetrustapi.BillingStatusFounder)
	partner.Tier = "developer" // Start on developer tier

	if err := s.repo.UpdatePartner(partner); err != nil {
		return fmt.Errorf("failed to update partner: %w", err)
	}

	// Initialize billing usage record
	if _, err := s.repo.GetOrCreateBillingUsage(ctx, partnerID); err != nil {
		return fmt.Errorf("failed to initialize billing usage: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"partner_id":      partnerID,
		"usage_threshold": usageThreshold,
		"free_days":       freeDays,
		"ends_at":         endsAt,
	}).Info("Partner enrolled in founder mode")

	return nil
}

// CheckFounderModeStatus checks if founder mode should end
func (s *BillingService) CheckFounderModeStatus(ctx context.Context, partnerID uuid.UUID) (bool, error) {
	partner, err := s.repo.GetPartnerByID(partnerID)
	if err != nil {
		return false, fmt.Errorf("failed to get partner: %w", err)
	}

	if !partner.IsFounderMode {
		return false, nil
	}

	// Check if time limit reached
	if partner.FounderModeEndsAt != nil && time.Now().After(*partner.FounderModeEndsAt) {
		return true, s.graduateFromFounderMode(ctx, partner, "time_limit_reached")
	}

	// Check if usage threshold reached
	if partner.CurrentMonthUsage >= partner.UsageThreshold {
		return true, s.graduateFromFounderMode(ctx, partner, "usage_threshold_reached")
	}

	return false, nil
}

// graduateFromFounderMode transitions a partner out of founder mode
func (s *BillingService) graduateFromFounderMode(ctx context.Context, partner *storagetrustapi.TrustAPIPartner, reason string) error {
	partner.IsFounderMode = false
	partner.BillingStatus = string(storagetrustapi.BillingStatusTrial)

	if err := s.repo.UpdatePartner(partner); err != nil {
		return fmt.Errorf("failed to update partner: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"partner_id": partner.ID,
		"reason":     reason,
		"tier":       partner.Tier,
	}).Info("Partner graduated from founder mode")

	// TODO: Send notification to partner about graduation
	// This would trigger an email with upgrade options

	return nil
}

// ============================================
// Monthly Billing Cycle
// ============================================

// GenerateMonthlyInvoice generates an invoice for the previous billing period
func (s *BillingService) GenerateMonthlyInvoice(ctx context.Context, partnerID uuid.UUID) (*storagetrustapi.TrustAPIBillingRecord, error) {
	usage, err := s.repo.GetOrCreateBillingUsage(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing usage: %w", err)
	}

	partner, err := s.repo.GetPartnerByID(partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner: %w", err)
	}

	pricing, err := s.GetTierPricing(ctx, partner.Tier)
	if err != nil {
		return nil, fmt.Errorf("failed to get tier pricing: %w", err)
	}

	// Calculate charges
	baseCharge := pricing.MonthlyPriceCents
	overageCharge := 0
	if pricing.HasOverageBilling && usage.OveragesThisPeriod > 0 {
		units := (usage.OveragesThisPeriod + 999) / 1000
		overageCharge = units * pricing.OveragePricePer1000
	}

	record := &storagetrustapi.TrustAPIBillingRecord{
		PartnerID:          partnerID,
		PeriodStart:        usage.BillingPeriodStart,
		PeriodEnd:          usage.BillingPeriodEnd,
		BaseRequests:       usage.RequestsThisPeriod - usage.OveragesThisPeriod,
		OverageRequests:    usage.OveragesThisPeriod,
		TotalRequests:      usage.RequestsThisPeriod,
		BaseChargeCents:    baseCharge,
		OverageChargeCents: overageCharge,
		TotalChargeCents:   baseCharge + overageCharge,
		Status:             "draft",
	}

	if err := s.repo.CreateBillingRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create billing record: %w", err)
	}

	// Reset usage for new period
	if err := s.repo.ResetPartnerBillingUsage(ctx, partnerID); err != nil {
		logrus.WithError(err).Warn("Failed to reset billing usage after invoice generation")
	}

	return record, nil
}

// CancelSubscription cancels a partner's Stripe subscription
func (s *BillingService) CancelSubscription(ctx context.Context, partnerID uuid.UUID) error {
	if !s.IsStripeConfigured() {
		return fmt.Errorf("Stripe is not configured")
	}

	partner, err := s.repo.GetPartnerByID(partnerID)
	if err != nil {
		return fmt.Errorf("failed to get partner: %w", err)
	}

	if partner.StripeSubscriptionID == "" {
		return fmt.Errorf("partner has no active subscription")
	}

	// Cancel at period end
	params := &stripe.SubscriptionCancelParams{
		InvoiceNow: stripe.Bool(false),
		Prorate:    stripe.Bool(false),
	}

	_, err = subscription.Cancel(partner.StripeSubscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	partner.BillingStatus = string(storagetrustapi.BillingStatusCancelled)
	partner.StripeSubscriptionID = ""

	if err := s.repo.UpdatePartner(partner); err != nil {
		return fmt.Errorf("failed to update partner: %w", err)
	}

	return nil
}

// Helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
