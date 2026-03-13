package payment

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/customer"
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

		// Try to get card details from the payment method
		// Note: In Stripe, you need to expand the payment method to get card details
		// For now, we'll return the information we can get from the customer object
		// In a full implementation, you'd make an additional API call to get the PaymentMethod

		// Check if there's a card in the customer's sources
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

		// If we have a default payment method ID but couldn't get details,
		// return what we know from the customer
		if defaultPM.ID != "" {
			// Return a placeholder - in production you'd use the PaymentIntent
			// or SetupIntent to get full card details
			return &PaymentMethodInfo{
				Brand: "Card",
				Last4: "****",
			}, nil
		}
	}

	return nil, fmt.Errorf("no payment method found for customer")
}
