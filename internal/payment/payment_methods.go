package payment

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/customer"
	"github.com/stripe/stripe-go/v83/paymentmethod"
	"github.com/stripe/stripe-go/v83/setupintent"
)

// ListPaymentMethodsForCustomer returns all payment methods attached to a Stripe customer
func ListPaymentMethodsForCustomer(ctx context.Context, customerID string) ([]PaymentMethodInfo, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	if customerID == "" {
		return nil, fmt.Errorf("customer_id is required")
	}

	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"),
	}

	iter := paymentmethod.List(params)
	var methods []PaymentMethodInfo

	// Get the customer to check default payment method
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("get customer: %w", err)
	}

	var defaultPMID string
	if cust.InvoiceSettings != nil && cust.InvoiceSettings.DefaultPaymentMethod != nil {
		defaultPMID = cust.InvoiceSettings.DefaultPaymentMethod.ID
	}

	for iter.Next() {
		pm := iter.PaymentMethod()
		if pm == nil || pm.Card == nil {
			continue
		}

		method := PaymentMethodInfo{
			StripePaymentMethodID: pm.ID,
			Brand:                 string(pm.Card.Brand),
			Last4:                 pm.Card.Last4,
			ExpMonth:              int(pm.Card.ExpMonth),
			ExpYear:               int(pm.Card.ExpYear),
			IsDefault:             pm.ID == defaultPMID,
		}
		methods = append(methods, method)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("list payment methods: %w", err)
	}

	return methods, nil
}

// AttachPaymentMethodToCustomer attaches a Stripe payment method to a customer
func AttachPaymentMethodToCustomer(ctx context.Context, customerID, paymentMethodID string) error {
	if stripeKey() == "" {
		return fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	if customerID == "" {
		return fmt.Errorf("customer_id is required")
	}
	if paymentMethodID == "" {
		return fmt.Errorf("payment_method_id is required")
	}

	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}

	_, err := paymentmethod.Attach(paymentMethodID, params)
	if err != nil {
		return fmt.Errorf("attach payment method: %w", err)
	}

	return nil
}

// DetachPaymentMethod detaches a Stripe payment method from its customer
func DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	if stripeKey() == "" {
		return fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	if paymentMethodID == "" {
		return fmt.Errorf("payment_method_id is required")
	}

	_, err := paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("detach payment method: %w", err)
	}

	return nil
}

// SetDefaultPaymentMethod sets the default payment method for a customer
func SetDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error {
	if stripeKey() == "" {
		return fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	if customerID == "" {
		return fmt.Errorf("customer_id is required")
	}
	if paymentMethodID == "" {
		return fmt.Errorf("payment_method_id is required")
	}

	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	}

	_, err := customer.Update(customerID, params)
	if err != nil {
		return fmt.Errorf("set default payment method: %w", err)
	}

	return nil
}

// CreateSetupIntent creates a Stripe SetupIntent for secure client-side payment method collection
func CreateSetupIntent(ctx context.Context, customerID string) (*SetupIntentResult, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	if customerID == "" {
		return nil, fmt.Errorf("customer_id is required")
	}

	params := &stripe.SetupIntentParams{
		Customer: stripe.String(customerID),
		PaymentMethodTypes: []*string{
			stripe.String("card"),
		},
	}

	si, err := setupintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("create setup intent: %w", err)
	}

	return &SetupIntentResult{
		ClientSecret: si.ClientSecret,
	}, nil
}

// SetupIntentResult holds the result of creating a SetupIntent
type SetupIntentResult struct {
	ClientSecret string `json:"client_secret"`
}
