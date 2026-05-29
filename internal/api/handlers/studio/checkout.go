package studio

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/functionfly/functionfly/internal/config"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
)

// MarketplacePlanCheckoutResult is returned when starting Stripe checkout for a creator plan.
type MarketplacePlanCheckoutResult struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

// CreateMarketplacePlanCheckoutSession starts a Stripe subscription checkout for a marketplace plan.
func CreateMarketplacePlanCheckoutSession(
	ctx context.Context,
	userRepo storage.Repository,
	subscriberTenantID uuid.UUID,
	subscriberUserID uuid.UUID,
	subscriberEmail, subscriberName string,
	plan *SubscriptionPlan,
	successURL, cancelURL string,
) (*MarketplacePlanCheckoutResult, error) {
	if plan == nil {
		return nil, fmt.Errorf("plan is required")
	}
	if plan.Price <= 0 {
		return nil, fmt.Errorf("free plans do not require checkout")
	}
	if !payment.IsConfigured() {
		return nil, fmt.Errorf("Stripe is not configured")
	}

	customerID, err := payment.CreateOrGetStripeCustomer(ctx, userRepo, subscriberTenantID, subscriberEmail, subscriberName)
	if err != nil {
		return nil, fmt.Errorf("create or get stripe customer: %w", err)
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = config.GetFrontendURL()
	}
	if successURL == "" {
		successURL = appURL + "/studio?marketplace_subscription=success"
	}
	if cancelURL == "" {
		cancelURL = appURL + "/studio?marketplace_subscription=cancel"
	}
	successURL = payment.SanitizeReturnURL(successURL, appURL+"/studio?marketplace_subscription=success")
	cancelURL = payment.SanitizeReturnURL(cancelURL, appURL+"/studio?marketplace_subscription=cancel")

	billingCycle := plan.BillingCycle
	if billingCycle == "" {
		billingCycle = "monthly"
	}
	interval, intervalCount := stripeBillingInterval(billingCycle)
	amountCents := int64(math.Round(plan.Price * 100))
	if amountCents <= 0 {
		return nil, fmt.Errorf("invalid plan price")
	}

	metadata := map[string]string{
		"purpose":              "marketplace_plan",
		"creator_tenant_id":    plan.TenantID,
		"plan_id":              plan.ID,
		"plan_name":            plan.Name,
		"subscriber_tenant_id": subscriberTenantID.String(),
		"subscriber_user_id":   subscriberUserID.String(),
		"subscriber_name":      subscriberName,
		"subscriber_email":     subscriberEmail,
		"billing_cycle":        billingCycle,
	}

	params := &stripe.CheckoutSessionParams{
		Customer:   stripe.String(customerID),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyUSD)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(fmt.Sprintf("%s (Creator Plan)", plan.Name)),
					},
					Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
						Interval:      stripe.String(interval),
						IntervalCount: stripe.Int64(intervalCount),
					},
					UnitAmount: stripe.Int64(amountCents),
				},
			},
		},
		AllowPromotionCodes: stripe.Bool(true),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
		Metadata: metadata,
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: metadata,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("create checkout session: %w", err)
	}
	if sess.URL == "" {
		return nil, fmt.Errorf("checkout session has no URL")
	}

	return &MarketplacePlanCheckoutResult{
		SessionID: sess.ID,
		URL:       sess.URL,
	}, nil
}

func stripeBillingInterval(billingCycle string) (string, int64) {
	switch billingCycle {
	case "quarterly":
		return string(stripe.PriceRecurringIntervalMonth), 3
	case "annual":
		return string(stripe.PriceRecurringIntervalYear), 1
	default:
		return string(stripe.PriceRecurringIntervalMonth), 1
	}
}
