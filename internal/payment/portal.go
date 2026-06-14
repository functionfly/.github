package payment

import (
	"context"
	"fmt"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/billingportal/session"
	"github.com/stripe/stripe-go/v83/customer"
)

// CreateOrGetStripeCustomer ensures the tenant has a Stripe customer; creates one if missing.
// It returns the Stripe customer ID and updates the tenant's stripe_customer_id if a new customer was created.
// If the tenant has billing address/tax information, this is synced to Stripe for automatic tax calculation.
func CreateOrGetStripeCustomer(
	ctx context.Context,
	repo storage.Repository,
	tenantID uuid.UUID,
	email, name string,
) (customerID string, err error) {
	if stripeKey() == "" {
		return "", fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}

	t, err := repo.GetTenantByID(ctx, tenantID)
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

	// If tenant has billing address, include it for tax calculation
	if t.BillingCountry != nil && *t.BillingCountry != "" {
		params.Address = &stripe.AddressParams{
			Country: stripe.String(*t.BillingCountry),
		}
		if t.BillingState != nil && *t.BillingState != "" {
			params.Address.State = stripe.String(*t.BillingState)
		}
		if t.BillingPostalCode != nil && *t.BillingPostalCode != "" {
			params.Address.PostalCode = stripe.String(*t.BillingPostalCode)
		}
	}

	// If tenant has a tax ID, add it to the customer
	if t.TaxID != nil && *t.TaxID != "" && t.TaxIDType != nil {
		stripeTaxType := convertTaxIDTypeToStripe(*t.TaxIDType)
		if stripeTaxType != "" {
			params.TaxIDData = []*stripe.CustomerTaxIDDataParams{
				{
					Type:  stripe.String(stripeTaxType),
					Value: stripe.String(*t.TaxID),
				},
			}
		}
	}

	c, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("create stripe customer: %w", err)
	}

	_, _ = repo.UpdateTenant(ctx, tenantID, map[string]interface{}{"stripe_customer_id": c.ID})
	return c.ID, nil
}

// convertTaxIDTypeToStripe converts our tax ID types to Stripe's format
func convertTaxIDTypeToStripe(taxIDType string) string {
	stripeTypes := map[string]string{
		"eu_vat": "eu_vat",
		"uk_vat": "gb_vat",
		"us_ein": "us_ein",
		"ca_gst": "ca_bn",
		"au_abn": "au_abn",
		"nz_gst": "nz_gst",
		"sg_gst": "sg_gst",
		"ch_vat": "ch_vat",
		"no_vat": "no_vat",
		"jp_cn":  "jp_cn",
		"kr_brn": "kr_brn",
		"tw_vat": "tw_vat",
		"in_gst": "in_gst",
	}

	if stripeType, ok := stripeTypes[taxIDType]; ok {
		return stripeType
	}
	return ""
}

// CreateBillingPortalSession creates a Stripe Customer Billing Portal session and returns the URL to redirect the user.
// NOTE: The billing portal configuration (including tax ID collection) must be configured in the Stripe Dashboard:
// https://dashboard.stripe.com/settings/billing/portal
// Enable these features in the portal configuration:
// - Tax ID collection (for VAT/GST numbers)
// - Customer address updates (required for tax calculation)
// - Invoice history with tax breakdowns
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
