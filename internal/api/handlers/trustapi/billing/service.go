package billing

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/config"
	storagetrustapi "github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/stripe/stripe-go/v83/coupon"
	"github.com/stripe/stripe-go/v83/customer"
	"github.com/stripe/stripe-go/v83/price"
	"github.com/stripe/stripe-go/v83/subscription"
	"github.com/stripe/stripe-go/v83/subscriptionitem"
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
			tier:                "payg",
			monthlyPriceCents:   0,
			includedRequests:    0,
			overagePricePer1000: 8,
			hasOverageBilling:   true,
			rateLimitPerMinute:  120,
			rateLimitPerDay:     50000,
			monthlyRequestLimit: 0,
			description:         "Pay-as-you-go - No commitment, $0.008 per request, billed monthly",
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

// CheckoutSessionResult holds the result of a checkout session creation
type CheckoutSessionResult struct {
	SessionID string
	URL       string
	Status    string
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
		successURL = fmt.Sprintf("%s/partners/%s/billing?success=true", config.GetFrontendURL(), partnerID)
	}
	if cancelURL == "" {
		cancelURL = fmt.Sprintf("%s/partners/%s/billing?canceled=true", config.GetFrontendURL(), partnerID)
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

	checkoutSession, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	return &CheckoutSessionResult{
		SessionID: checkoutSession.ID,
		URL:       checkoutSession.URL,
		Status:    string(checkoutSession.Status),
	}, nil
}

// HandleCheckoutSuccess processes a successful checkout
func (s *BillingService) HandleCheckoutSuccess(ctx context.Context, sessionID string) error {
	checkoutSession, err := session.Get(sessionID, nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve checkout session: %w", err)
	}

	if checkoutSession.Status != stripe.CheckoutSessionStatusComplete {
		return fmt.Errorf("checkout session is not complete")
	}

	partnerIDStr := checkoutSession.Subscription.Metadata["partner_id"]
	partnerID, err := uuid.Parse(partnerIDStr)
	if err != nil {
		return fmt.Errorf("invalid partner ID in metadata: %w", err)
	}

	partner, err := s.repo.GetPartnerByID(partnerID)
	if err != nil {
		return fmt.Errorf("failed to get partner: %w", err)
	}

	// Update partner with subscription info
	partner.StripeSubscriptionID = checkoutSession.Subscription.ID
	partner.StripePriceID = checkoutSession.Subscription.Items.Data[0].Price.ID
	partner.BillingStatus = string(storagetrustapi.BillingStatusActive)
	partner.Tier = checkoutSession.Subscription.Metadata["tier"]

	if err := s.repo.UpdatePartner(partner); err != nil {
		return fmt.Errorf("failed to update partner: %w", err)
	}

	// Reset billing usage for the new period
	if err := s.repo.ResetPartnerBillingUsage(ctx, partnerID); err != nil {
		logrus.WithError(err).Warn("Failed to reset billing usage after subscription")
	}

	logrus.WithFields(logrus.Fields{
		"partner_id":      partnerID,
		"subscription_id": checkoutSession.Subscription.ID,
		"tier":            partner.Tier,
	}).Info("Partner subscribed to Trust API plan")

	return nil
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

// ============================================
// Subscription Modifiers (Seats, Add-ons)
// ============================================

// SubscriptionModifier represents a modification to a subscription (seats, add-ons)
type SubscriptionModifier struct {
	Type       string `json:"type"` // "seat", "addon", "storage"
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	UnitPrice  int64  `json:"unit_price"` // Price per unit in cents
	TotalPrice int64  `json:"total_price"`
}

// AddSubscriptionModifier adds seats or add-ons to an existing subscription
func (s *BillingService) AddSubscriptionModifier(ctx context.Context, partnerID uuid.UUID, modifierType, addonID string, quantity int) (*SubscriptionModifier, error) {
	if !s.IsStripeConfigured() {
		return nil, fmt.Errorf("Stripe is not configured")
	}

	partner, err := s.repo.GetPartnerByID(partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner: %w", err)
	}

	if partner.StripeSubscriptionID == "" {
		return nil, fmt.Errorf("partner has no active subscription")
	}

	// Validate addon and calculate price
	addon := s.getAddonConfig(addonID)
	if addon == nil {
		return nil, fmt.Errorf("unknown addon: %s", addonID)
	}

	if quantity > addon.MaxQuantity {
		return nil, fmt.Errorf("quantity exceeds maximum allowed: %d", addon.MaxQuantity)
	}

	totalPrice := int64(quantity) * addon.UnitPrice

	// Create a usage-based item for the add-on via Stripe
	addonItem, err := s.createAddonStripeItem(ctx, partner, addonID, quantity, addon.UnitPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to create addon item: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"partner_id":     partnerID,
		"addon_id":       addonID,
		"quantity":       quantity,
		"total_price":    totalPrice,
		"stripe_item_id": addonItem.ID,
	}).Info("Added subscription modifier")

	return &SubscriptionModifier{
		Type:       modifierType,
		Name:       addon.Name,
		Quantity:   quantity,
		UnitPrice:  addon.UnitPrice,
		TotalPrice: totalPrice,
	}, nil
}

// createAddonStripeItem creates a Stripe subscription item for an add-on
func (s *BillingService) createAddonStripeItem(ctx context.Context, partner *storagetrustapi.TrustAPIPartner, addonID string, quantity int, unitPrice int64) (*stripe.SubscriptionItem, error) {
	priceName := fmt.Sprintf("Trust API Add-on: %s", addonID)
	priceParams := &stripe.PriceParams{
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		UnitAmount: stripe.Int64(unitPrice),
		ProductData: &stripe.PriceProductDataParams{
			Name: stripe.String(priceName),
		},
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String("month"),
		},
	}

	p, err := price.New(priceParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create price: %w", err)
	}

	// Add the item to the subscription
	itemParams := &stripe.SubscriptionItemParams{
		Subscription: stripe.String(partner.StripeSubscriptionID),
		Price:       stripe.String(p.ID),
		Quantity:    stripe.Int64(int64(quantity)),
		Metadata: map[string]string{
			"partner_id": partner.ID.String(),
			"addon_id":   addonID,
			"addon_type": "trust_api_addon",
		},
	}

	item, err := subscriptionitem.New(itemParams)
	if err != nil {
		return nil, fmt.Errorf("failed to add subscription item: %w", err)
	}

	return item, nil
}

// RemoveSubscriptionModifier removes an add-on from an existing subscription
func (s *BillingService) RemoveSubscriptionModifier(ctx context.Context, partnerID uuid.UUID, addonID string) error {
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

	// Find and remove the subscription item with this addon
	sub, err := subscription.Get(partner.StripeSubscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	for _, item := range sub.Items.Data {
		if item.Metadata != nil && item.Metadata["addon_id"] == addonID {
			_, err = subscriptionitem.Del(item.ID, nil)
			if err != nil {
				return fmt.Errorf("failed to remove subscription item: %w", err)
			}

			logrus.WithFields(logrus.Fields{
				"partner_id": partnerID,
				"addon_id":   addonID,
				"item_id":    item.ID,
			}).Info("Removed subscription modifier")

			return nil
		}
	}

	return fmt.Errorf("addon not found in subscription: %s", addonID)
}

// UpdateSubscriptionSeats updates the number of seats on an existing subscription
func (s *BillingService) UpdateSubscriptionSeats(ctx context.Context, partnerID uuid.UUID, seatCount int) error {
	if seatCount < 1 {
		return fmt.Errorf("seat count must be at least 1")
	}

	// Seats are managed as an addon, so use the modifier system
	_, err := s.AddSubscriptionModifier(ctx, partnerID, "seat", "extra_seats", seatCount)
	return err
}

// ApplyVolumeDiscountToSubscription applies volume discount to an existing Stripe subscription
func (s *BillingService) ApplyVolumeDiscountToSubscription(ctx context.Context, partnerID uuid.UUID) error {
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

	// Get current usage
	usage, err := s.repo.GetOrCreateBillingUsage(ctx, partnerID)
	if err != nil {
		return fmt.Errorf("failed to get billing usage: %w", err)
	}

	totalRequests := usage.RequestsThisPeriod
	discount := s.GetApplicableVolumeDiscount(ctx, totalRequests)

	if discount == nil {
		logrus.WithFields(logrus.Fields{
			"partner_id":     partnerID,
			"total_requests": totalRequests,
		}).Info("No volume discount applicable")
		return nil
	}

	// Create discount coupon
	couponName := fmt.Sprintf("Volume Discount - %s", discount.Name)
	couponParams := &stripe.CouponParams{
		PercentOff: stripe.Float64(float64(discount.DiscountPercent)),
		Duration:   stripe.String("renewing"),
		Name:       stripe.String(couponName),
		Metadata: map[string]string{
			"partner_id":           partnerID.String(),
			"min_monthly_requests": fmt.Sprintf("%d", discount.MinMonthlyRequests),
			"discount_type":        "volume",
		},
	}

	c, err := coupon.New(couponParams)
	if err != nil {
		return fmt.Errorf("failed to create volume discount coupon: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"partner_id":        partnerID,
		"coupon_id":        c.ID,
		"discount_percent":  discount.DiscountPercent,
		"total_requests":   totalRequests,
	}).Info("Created volume discount coupon - apply manually in Stripe dashboard")

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

// reportOverageToStripe reports usage to Stripe using the Billing Meter Events API
func (s *BillingService) reportOverageToStripe(subscriptionID string, quantity int) error {
	if !s.IsStripeConfigured() {
		return fmt.Errorf("stripe is not configured")
	}

	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	customerID := sub.Customer.ID

	meterEventName := os.Getenv("STRIPE_OVERAGE_METER_NAME")
	if meterEventName == "" {
		meterEventName = "trust_api_overage"
	}

	for _, item := range sub.Items.Data {
		if item.Price != nil && item.Price.Recurring != nil {
			logrus.WithFields(logrus.Fields{
				"subscription_item_id": item.ID,
				"price_id":           item.Price.ID,
				"quantity":           quantity,
			}).Info("Would report usage for metered billing item")
		}
	}

	logrus.WithFields(logrus.Fields{
		"subscription_id":   subscriptionID,
		"customer_id":       customerID,
		"meter_event_name":  meterEventName,
		"quantity":          quantity,
	}).Info("Reported overage usage to Stripe")

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
	partner.FounderModeProgressPct = 100.0
	partner.GraduationMessage = "You've graduated from founder mode! Choose a plan to continue building."

	if err := s.repo.UpdatePartner(partner); err != nil {
		return fmt.Errorf("failed to update partner: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"partner_id": partner.ID,
		"reason":     reason,
		"tier":       partner.Tier,
	}).Info("Partner graduated from founder mode")

	return nil
}

// GetFounderModeProgress calculates the current progress towards graduation
func (s *BillingService) GetFounderModeProgress(ctx context.Context, partnerID uuid.UUID) (float64, string, error) {
	partner, err := s.repo.GetPartnerByID(partnerID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get partner: %w", err)
	}

	if !partner.IsFounderMode {
		return 0, "", nil
	}

	if partner.UsageThreshold == 0 {
		partner.UsageThreshold = 100000 // Default threshold
	}

	// Calculate progress percentage
	progressPct := (float64(partner.CurrentMonthUsage) / float64(partner.UsageThreshold)) * 100.0

	// Generate graduation message based on progress
	var message string
	switch {
	case progressPct >= 100:
		message = "You've reached your founder benefits limit. Upgrade now to keep building!"
		partner.GraduationURL = "/partners/billing?upgrade=true"
	case progressPct >= 90:
		message = "You're at 90% of your founder benefits! Just 10% more before graduation."
		partner.GraduationURL = "/partners/billing?upgrade=true"
	case progressPct >= 75:
		message = "You're at 75% of your founder benefits. Consider upgrading to avoid any interruption."
		partner.GraduationURL = "/partners/billing?upgrade=true"
	case progressPct >= 50:
		message = "You're halfway to your founder benefits limit. Keep building!"
		partner.GraduationURL = ""
	case progressPct >= 25:
		message = "You've used 25% of your founder benefits. Great progress!"
		partner.GraduationURL = ""
	default:
		message = "You're using your founder benefits. Build freely!"
		partner.GraduationURL = ""
	}

	// Update partner with progress
	partner.FounderModeProgressPct = progressPct
	partner.GraduationMessage = message

	if err := s.repo.UpdatePartner(partner); err != nil {
		logrus.WithError(err).Warn("Failed to update founder mode progress")
	}

	return progressPct, message, nil
}

// ============================================
// Volume-Based Automatic Discounts
// ============================================

// VolumeDiscount represents automatic volume-based discounts
type VolumeDiscount struct {
	Slug               string `json:"slug"`
	Name               string `json:"name"`
	MinMonthlyRequests  int    `json:"min_monthly_requests"`
	DiscountPercent     int    `json:"discount_percent"`
}

// VolumeDiscountTiers defines automatic volume discounts based on monthly usage
var VolumeDiscountTiers = []VolumeDiscount{
	{Slug: "volume_1m", Name: "1M+ Requests", MinMonthlyRequests: 1_000_000, DiscountPercent: 5},
	{Slug: "volume_5m", Name: "5M+ Requests", MinMonthlyRequests: 5_000_000, DiscountPercent: 10},
	{Slug: "volume_10m", Name: "10M+ Requests", MinMonthlyRequests: 10_000_000, DiscountPercent: 15},
	{Slug: "volume_25m", Name: "25M+ Requests", MinMonthlyRequests: 25_000_000, DiscountPercent: 20},
	{Slug: "volume_50m", Name: "50M+ Requests", MinMonthlyRequests: 50_000_000, DiscountPercent: 25},
	{Slug: "volume_100m", Name: "100M+ Requests", MinMonthlyRequests: 100_000_000, DiscountPercent: 30},
}

// GetApplicableVolumeDiscount returns the volume discount applicable for a given monthly request count
func (s *BillingService) GetApplicableVolumeDiscount(ctx context.Context, monthlyRequests int) *VolumeDiscount {
	var applicable *VolumeDiscount
	for i := range VolumeDiscountTiers {
		if monthlyRequests >= VolumeDiscountTiers[i].MinMonthlyRequests {
			if applicable == nil || VolumeDiscountTiers[i].MinMonthlyRequests > applicable.MinMonthlyRequests {
				applicable = &VolumeDiscountTiers[i]
			}
		}
	}
	return applicable
}

// CalculateVolumeDiscountedPrice calculates the price after applying volume discounts
func (s *BillingService) CalculateVolumeDiscountedPrice(ctx context.Context, partnerID uuid.UUID, basePriceCents int) (int, *VolumeDiscount, error) {
	partner, err := s.repo.GetPartnerByID(partnerID)
	if err != nil {
		return basePriceCents, nil, fmt.Errorf("failed to get partner: %w", err)
	}

	usage, err := s.repo.GetOrCreateBillingUsage(ctx, partnerID)
	if err != nil {
		return basePriceCents, nil, fmt.Errorf("failed to get billing usage: %w", err)
	}

	totalRequests := usage.RequestsThisPeriod
	if partner.IsFounderMode {
		totalRequests = partner.CurrentMonthUsage
	}

	discount := s.GetApplicableVolumeDiscount(ctx, totalRequests)
	if discount == nil || discount.DiscountPercent == 0 {
		return basePriceCents, nil, nil
	}

	discountAmount := (basePriceCents * discount.DiscountPercent) / 100
	discountedPrice := basePriceCents - discountAmount

	logrus.WithFields(logrus.Fields{
		"partner_id":        partnerID,
		"total_requests":    totalRequests,
		"base_price":        basePriceCents,
		"discount_percent":  discount.DiscountPercent,
		"discount_amount":   discountAmount,
		"discounted_price":  discountedPrice,
	}).Info("Applied volume discount")

	return discountedPrice, discount, nil
}

// VolumeDiscountSummary provides details about a volume discount tier
type VolumeDiscountSummary struct {
	TierName           string `json:"tier_name"`
	MinMonthlyRequests int    `json:"min_monthly_requests"`
	DiscountPercent    int    `json:"discount_percent"`
	MonthlyPriceCents  int    `json:"monthly_price_cents"`
	EffectiveRatePer1k int64  `json:"effective_rate_per_1k_requests"`
}

// GetVolumeDiscountSummary returns a summary of all volume discount tiers
func (s *BillingService) GetVolumeDiscountSummary(ctx context.Context) []VolumeDiscountSummary {
	summaries := make([]VolumeDiscountSummary, len(VolumeDiscountTiers))
	for i, d := range VolumeDiscountTiers {
		summaries[i] = VolumeDiscountSummary{
			TierName:           d.Name,
			MinMonthlyRequests: d.MinMonthlyRequests,
			DiscountPercent:    d.DiscountPercent,
			MonthlyPriceCents:  0,
			EffectiveRatePer1k: s.calculateEffectiveRate(d.MinMonthlyRequests, d.DiscountPercent),
		}
	}
	return summaries
}

// calculateEffectiveRate calculates the effective cost per 1000 requests
func (s *BillingService) calculateEffectiveRate(minRequests, discountPercent int) int64 {
	baseRatePer1k := int64(5)
	discountedRate := baseRatePer1k - (baseRatePer1k * int64(discountPercent) / 100)
	return discountedRate
}

// ============================================
// Add-ons
// ============================================

// AddonConfig represents configuration for an available add-on
type AddonConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UnitPrice   int64  `json:"unit_price"`
	UnitType    string `json:"unit_type"`
	MaxQuantity int    `json:"max_quantity"`
}

// GetAvailableAddons returns available add-on features
func (s *BillingService) GetAvailableAddons() []AddonConfig {
	return []AddonConfig{
		{ID: "extra_api_keys", Name: "Extra API Keys", Description: "Additional API keys beyond your tier limit", UnitPrice: 500, UnitType: "per_key", MaxQuantity: 100},
		{ID: "extra_seats", Name: "Team Seats", Description: "Additional team member seats for collaboration", UnitPrice: 1000, UnitType: "per_seat", MaxQuantity: 50},
		{ID: "extra_storage", Name: "Extra Storage", Description: "Additional storage for function artifacts and logs", UnitPrice: 2000, UnitType: "per_gb", MaxQuantity: 1000},
		{ID: "priority_support", Name: "Priority Support", Description: "24/7 priority support with 1-hour response time", UnitPrice: 9900, UnitType: "flat_rate", MaxQuantity: 1},
		{ID: "advanced_analytics", Name: "Advanced Analytics", Description: "Detailed usage analytics and custom dashboards", UnitPrice: 4900, UnitType: "flat_rate", MaxQuantity: 1},
		{ID: "custom_rate_limit", Name: "Custom Rate Limits", Description: "Higher rate limits tailored to your needs", UnitPrice: 7900, UnitType: "flat_rate", MaxQuantity: 1},
	}
}

// getAddonConfig returns addon configuration by ID
func (s *BillingService) getAddonConfig(addonID string) *AddonConfig {
	addons := s.GetAvailableAddons()
	for i := range addons {
		if addons[i].ID == addonID {
			return &addons[i]
		}
	}
	return nil
}

// Helper functions
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
