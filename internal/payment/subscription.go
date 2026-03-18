package payment

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/customer"
	"github.com/stripe/stripe-go/v83/paymentmethod"
)

// PaymentMethodInfo represents the payment method details for display
type PaymentMethodInfo struct {
	Brand    string `json:"brand"`
	Last4    string `json:"last4"`
	ExpMonth int    `json:"exp_month"`
	ExpYear  int    `json:"exp_year"`
}

// GetPaymentMethodForCustomer retrieves the default payment method for a Stripe customer
func GetPaymentMethodForCustomer(ctx context.Context, customerID string) (*PaymentMethodInfo, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	if customerID == "" {
		return nil, fmt.Errorf("customer_id is required")
	}

	// Get the customer to find the default payment method
	params := &stripe.CustomerParams{}
	c, err := customer.Get(customerID, params)
	if err != nil {
		return nil, fmt.Errorf("get customer: %w", err)
	}

	// Check for default payment method on the customer
	if c.InvoiceSettings != nil && c.InvoiceSettings.DefaultPaymentMethod != nil {
		defaultPM := c.InvoiceSettings.DefaultPaymentMethod
		if defaultPM.ID != "" {
			pm, err := paymentmethod.Get(defaultPM.ID, nil)
			if err == nil && pm != nil && pm.Card != nil {
				return &PaymentMethodInfo{
					Brand:    string(pm.Card.Brand),
					Last4:    pm.Card.Last4,
					ExpMonth: int(pm.Card.ExpMonth),
					ExpYear:  int(pm.Card.ExpYear),
				}, nil
			}
			// If Get failed or card is nil, fall through to legacy sources or placeholder
		}

		// Fallback: check legacy card sources on the customer
		if c.Sources != nil && c.Sources.Data != nil {
			for _, src := range c.Sources.Data {
				if src.Card != nil {
					return &PaymentMethodInfo{
						Brand:    string(src.Card.Brand),
						Last4:    src.Card.Last4,
						ExpMonth: int(src.Card.ExpMonth),
						ExpYear:  int(src.Card.ExpYear),
					}, nil
				}
			}
		}

		// Last resort: we have a default payment method ID but couldn't get details
		if defaultPM.ID != "" {
			return &PaymentMethodInfo{
				Brand: "Card",
				Last4: "****",
			}, nil
		}
	}

	return nil, fmt.Errorf("no payment method found for customer")
}
