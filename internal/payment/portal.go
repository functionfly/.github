package payment

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/billingportal/session"
	"github.com/stripe/stripe-go/v83/customer"
)

// CreateOrGetStripeCustomer ensures the tenant has a Stripe customer; creates one if missing.
// It returns the Stripe customer ID and updates the tenant's stripe_customer_id if a new customer was created.
func CreateOrGetStripeCustomer(
	ctx context.Context,
	repo storage.Repository,
	tenantID uuid.UUID,
	email, name string,
) (customerID string, err error) {
	if stripeKey() == "" {
		return "", fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}

	t, err := repo.GetTenantByID(tenantID)
	if err != nil {
		return "", fmt.Errorf("get tenant: %w", err)
	}
	if t == nil {
		return "", fmt.Errorf("tenant not found")
	}
	if t.StripeCustomerID != nil && *t.StripeCustomerID != "" {
		return *t.StripeCustomerID, nil
	}

	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
		Metadata: map[string]string{
			"tenant_id": tenantID.String(),
		},
	}

	c, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("create stripe customer: %w", err)
	}

	_, _ = repo.UpdateTenant(ctx, tenantID, map[string]interface{}{"stripe_customer_id": c.ID})
	return c.ID, nil
}

// CreateBillingPortalSession creates a Stripe Customer Billing Portal session and returns the URL to redirect the user.
func CreateBillingPortalSession(ctx context.Context, customerID, returnURL string) (string, error) {
	if stripeKey() == "" {
		return "", fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	if customerID == "" {
		return "", fmt.Errorf("customer_id is required")
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("create billing portal session: %w", err)
	}
	if sess.URL == "" {
		return "", fmt.Errorf("portal session has no URL")
	}
	return sess.URL, nil
}
