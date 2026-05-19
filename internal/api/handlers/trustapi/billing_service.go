package trustapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/config"
	"github.com/functionfly/functionfly/internal/storage"
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
	repo               *storagetrustapi.BillingRepository
	usageReportingRepo *storage.UsageReportingRepository
	stripeKey          string
	environment        string
}

// NewBillingService creates a new billing service
func NewBillingService(repo *storagetrustapi.BillingRepository, usageReportingRepo *storage.UsageReportingRepository) *BillingService {
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	return &BillingService{
		repo:               repo,
		usageReportingRepo: usageReportingRepo,
		stripeKey:          stripeKey,
		environment:        env,
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

	if partner.StripeCustomerID != "" {
		_, err := customer.Get(partner.StripeCustomerID, nil)
		if err == nil {
			return partner.StripeCustomerID, nil
		}
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

	pricing, err := s.GetTierPricing(ctx, tier)
	if err != nil {
		return nil, fmt.Errorf("failed to get tier pricing: %w", err)
	}

	if pricing.MonthlyPriceCents == 0 {
		return nil, fmt.Errorf("cannot create checkout for free tier")
	}

	customerID := partner.StripeCustomerID
	if customerID == "" {
		customerID, err = s.CreateStripeCustomer(ctx, partnerID, partner.ContactEmail, partner.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to create Stripe customer: %w", err)
		}
	}

	var lineItems []*stripe.CheckoutSessionLineItemParams

	if pricing.StripePriceID != "" {
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(pricing.StripePriceID),
			Quantity: stripe.Int64(1),
		})
	} else {
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

	if pricing.HasOverageBilling && pricing.OveragePricePer1000 > 0 {
		if pricing.StripeMeteredPriceID != "" {
			lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
				Price:    stripe.String(pricing.StripeMeteredPriceID),
				Quantity: stripe.Int64(0),
			})
		} else {
			logrus.WithField("tier", tier).Warn(
				"No StripeMeteredPriceID configured for tier with overage billing. " +
					"Run Stripe CLI setup commands (documented in code) and update tier config.")
		}
	}

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

	partner.StripeSubscriptionID = checkoutSession.Subscription.ID
	partner.StripePriceID = checkoutSession.Subscription.Items.Data[0].Price.ID
	partner.BillingStatus = string(storagetrustapi.BillingStatusActive)
	partner.Tier = checkoutSession.Subscription.Metadata["tier"]

	if err := s.repo.UpdatePartner(partner); err != nil {
		return fmt.Errorf("failed to update partner: %w", err)
	}

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

// SubscriptionModifier represents a modification to a subscription (seats, add-ons)
type SubscriptionModifier struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	UnitPrice  int64  `json:"unit_price"`
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

	addon := s.getAddonConfig(addonID)
	if addon == nil {
		return nil, fmt.Errorf("unknown addon: %s", addonID)
	}

	if quantity > addon.MaxQuantity {
		return nil, fmt.Errorf("quantity exceeds maximum allowed: %d", addon.MaxQuantity)
	}

	totalPrice := int64(quantity) * addon.UnitPrice

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

func (s *BillingService) getAddonConfig(addonID string) *AddonConfig {
	addons := s.GetAvailableAddons()
	for i := range addons {
		if addons[i].ID == addonID {
			return &addons[i]
		}
	}
	return nil
}

func (s *BillingService) createAddonStripeItem(ctx context.Context, partner *storagetrustapi.TrustAPIPartner, addonID string, quantity int, unitPrice int64) (*stripe.SubscriptionItem, error) {
	priceName := fmt.Sprintf("Trust API Add-on: %s", addonID)
	priceParams := &stripe.PriceParams{
		Currency:   stripe.String(string(stripe.CurrencyUSD)),
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

	itemParams := &stripe.SubscriptionItemParams{
		Subscription: stripe.String(partner.StripeSubscriptionID),
		Price:        stripe.String(p.ID),
		Quantity:     stripe.Int64(int64(quantity)),
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
		"partner_id":       partnerID,
		"coupon_id":        c.ID,
		"discount_percent": discount.DiscountPercent,
		"total_requests":   totalRequests,
	}).Info("Created volume discount coupon - apply manually in Stripe dashboard")

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
// Usage Tracking & Metered Billing
// ============================================

// RecordUsage records API usage for billing
func (s *BillingService) RecordUsage(ctx context.Context, partnerID uuid.UUID, requestCount int) error {
	partner, err := s.repo.GetPartnerByID(partnerID)
	if err != nil {
		return fmt.Errorf("failed to get partner: %w", err)
	}

	pricing, err := s.GetTierPricing(ctx, partner.Tier)
	if err != nil {
		return fmt.Errorf("failed to get tier pricing: %w", err)
	}

	usage, err := s.repo.GetOrCreateBillingUsage(ctx, partnerID)
	if err != nil {
		return fmt.Errorf("failed to get or create billing usage: %w", err)
	}

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

	usage.RequestsThisPeriod = newTotalUsage
	usage.OveragesThisPeriod += overageCount

	if err := s.repo.UpdateBillingUsage(ctx, usage); err != nil {
		return fmt.Errorf("failed to update billing usage: %w", err)
	}

	if overageCount > 0 && partner.StripeSubscriptionID != "" && s.IsStripeConfigured() {
		if err := s.reportOverageToStripe(ctx, partnerID, partner.StripeSubscriptionID, overageCount); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"partner_id":    partnerID,
				"overage_count": overageCount,
			}).Warn("Failed to report overage to Stripe")
		}
	}

	partner.CurrentMonthUsage = newTotalUsage
	partner.CurrentOverageUsage = usage.OveragesThisPeriod

	if err := s.repo.UpdatePartner(partner); err != nil {
		logrus.WithError(err).Warn("Failed to update partner usage cache")
	}

	return nil
}

func (s *BillingService) reportOverageToStripe(ctx context.Context, partnerID uuid.UUID, subscriptionID string, quantity int) error {
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

	// Track whether we successfully reported to at least one metered item
	var reportedItems int
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	for _, item := range sub.Items.Data {
		// Check if this is a metered billing item
		if item.Price != nil && item.Price.Recurring != nil && item.Price.Recurring.UsageType == stripe.PriceRecurringUsageTypeMetered {

			// Create idempotency key for this usage report
			idempotencyKey := storage.GenerateIdempotencyKey(partnerID, item.ID, now.Unix())

			// Check if we've already reported this exact usage
			if s.usageReportingRepo != nil {
				existingReport, err := s.usageReportingRepo.GetUsageReportByIdempotencyKey(ctx, idempotencyKey)
				if err == nil && existingReport != nil {
					logrus.WithFields(logrus.Fields{
						"subscription_item_id": item.ID,
						"usage_report_id":      existingReport.ID,
					}).Info("Usage already reported for this period, skipping")
					continue
				}
			}

			// Create usage record in Stripe using billing meter event API
			// This is the new way to report usage in Stripe v83+ SDK
			// Meter events use a separate API endpoint (meter-events.stripe.com)
			meterPayload := map[string]interface{}{
				"event_name": meterEventName,
				"timestamp":  now.Unix(),
				"identifier": idempotencyKey,
				"payload": map[string]string{
					"value":              fmt.Sprintf("%d", quantity),
					"stripe_customer_id": customerID,
				},
			}

			meterEventID, err := s.createMeterEvent(ctx, meterPayload)
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"subscription_item_id": item.ID,
					"quantity":             quantity,
				}).Error("Failed to report usage to Stripe")

				// Record the failed attempt
				if s.usageReportingRepo != nil {
					report := &storage.StripeUsageReport{
						TenantID:           partnerID, // Using partnerID as tenantID for Trust API
						PartnerID:          &partnerID,
						SubscriptionID:     subscriptionID,
						SubscriptionItemID: item.ID,
						UsageQuantity:      quantity,
						UsagePeriodStart:   periodStart,
						UsagePeriodEnd:     periodEnd,
						StripeTimestamp:    now.Unix(),
						Status:             "failed",
						ErrorMessage:       err.Error(),
						IdempotencyKey:     idempotencyKey,
						MeterEventName:     meterEventName,
					}
					if err := s.usageReportingRepo.CreateUsageReport(ctx, report); err != nil {
						logrus.WithError(err).Warn("Failed to create usage report record for failed attempt")
					}
				}
				continue
			}

			// Successfully reported to Stripe
			reportedItems++

			// Record successful usage report
			if s.usageReportingRepo != nil {
				report := &storage.StripeUsageReport{
					TenantID:            partnerID, // Using partnerID as tenantID for Trust API
					PartnerID:           &partnerID,
					SubscriptionID:      subscriptionID,
					SubscriptionItemID:  item.ID,
					UsageQuantity:       quantity,
					UsagePeriodStart:    periodStart,
					UsagePeriodEnd:      periodEnd,
					StripeTimestamp:     now.Unix(),
					StripeUsageRecordID: meterEventID,
					Status:              "reported",
					IdempotencyKey:      idempotencyKey,
					MeterEventName:      meterEventName,
				}
				if err := s.usageReportingRepo.CreateUsageReport(ctx, report); err != nil {
					logrus.WithError(err).Warn("Failed to create usage report record for successful attempt")
				}
			}

			logrus.WithFields(logrus.Fields{
				"subscription_item_id": item.ID,
				"price_id":             item.Price.ID,
				"quantity":             quantity,
				"meter_event_id":       meterEventID,
				"timestamp":            now.Unix(),
			}).Info("Successfully reported usage to Stripe")
		}
	}

	if reportedItems == 0 {
		logrus.WithFields(logrus.Fields{
			"subscription_id": subscriptionID,
			"customer_id":     customerID,
			"quantity":        quantity,
		}).Warn("No metered billing items found for usage reporting")
		return fmt.Errorf("no metered billing items found in subscription")
	}

	logrus.WithFields(logrus.Fields{
		"subscription_id":  subscriptionID,
		"customer_id":      customerID,
		"meter_event_name": meterEventName,
		"quantity":         quantity,
		"items_reported":   reportedItems,
	}).Info("Successfully reported overage usage to Stripe")

	return nil
}

// createMeterEvent creates a meter event in Stripe using the Meter Events API
// This uses a separate endpoint (meter-events.stripe.com) for high-volume usage reporting
func (s *BillingService) createMeterEvent(ctx context.Context, payload map[string]interface{}) (string, error) {
	if s.stripeKey == "" {
		return "", fmt.Errorf("stripe API key not configured")
	}

	// Marshal payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request to Stripe Meter Events API
	// Note: Meter Events API uses a different base URL: https://meter-events.stripe.com
	req, err := http.NewRequestWithContext(ctx, "POST", "https://meter-events.stripe.com/v1/billing/meter_events", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.stripeKey)
	req.Header.Set("Stripe-Version", "2024-04-15") // Use a recent API version

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Stripe: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result struct {
		ID    string `json:"id"`
		Error *struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("stripe API error: %s (code: %s)", result.Error.Message, result.Error.Code)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return result.ID, nil
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

	overageRequests := usage.OveragesThisPeriod
	var overageChargeCents int
	if pricing.HasOverageBilling && overageRequests > 0 {
		units := (overageRequests + 999) / 1000
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

	if usageThreshold <= 0 {
		usageThreshold = 100000
	}
	if freeDays <= 0 {
		freeDays = 90
	}

	now := time.Now().UTC()
	endsAt := now.AddDate(0, 0, freeDays)

	partner.IsFounderMode = true
	partner.FounderModeStartedAt = &now
	partner.FounderModeEndsAt = &endsAt
	partner.UsageThreshold = usageThreshold
	partner.BillingStatus = string(storagetrustapi.BillingStatusFounder)
	partner.Tier = "developer"

	if err := s.repo.UpdatePartner(partner); err != nil {
		return fmt.Errorf("failed to update partner: %w", err)
	}

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

	if partner.FounderModeEndsAt != nil && time.Now().After(*partner.FounderModeEndsAt) {
		return true, s.graduateFromFounderMode(ctx, partner, "time_limit_reached")
	}

	if partner.CurrentMonthUsage >= partner.UsageThreshold {
		return true, s.graduateFromFounderMode(ctx, partner, "usage_threshold_reached")
	}

	return false, nil
}

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
		partner.UsageThreshold = 100000
	}

	progressPct := (float64(partner.CurrentMonthUsage) / float64(partner.UsageThreshold)) * 100.0

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
	MinMonthlyRequests int    `json:"min_monthly_requests"`
	DiscountPercent    int    `json:"discount_percent"`
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
		"partner_id":       partnerID,
		"total_requests":   totalRequests,
		"base_price":       basePriceCents,
		"discount_percent": discount.DiscountPercent,
		"discount_amount":  discountAmount,
		"discounted_price": discountedPrice,
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

// GetAvailableAddons returns available add-on features that can be purchased
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
