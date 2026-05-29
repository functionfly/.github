package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SendWalletTopUp notifies a user after credits are added to an agent wallet.
func (s *Service) SendWalletTopUp(ctx context.Context, userID uuid.UUID, agentID string, amountUSD, newBalanceUSD float64) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeBillingWalletToppedUp,
		Category: CategoryBilling,
		Title:    "Wallet Top-Up Successful",
		Body:     fmt.Sprintf("Your wallet was topped up by $%.2f. New balance: $%.2f.", amountUSD, newBalanceUSD),
		Data: JSONMap{
			"agent_id":        agentID,
			"amount_usd":      amountUSD,
			"new_balance_usd": newBalanceUSD,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendRegistryWalletTopUp notifies a user after credits are added to their registry wallet.
func (s *Service) SendRegistryWalletTopUp(ctx context.Context, userID uuid.UUID, amountUSD, newBalanceUSD float64) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeBillingWalletToppedUp,
		Category: CategoryBilling,
		Title:    "Registry Wallet Top-Up Successful",
		Body:     fmt.Sprintf("Your registry wallet was topped up by $%.2f. New balance: $%.2f.", amountUSD, newBalanceUSD),
		Data: JSONMap{
			"amount_usd":      amountUSD,
			"new_balance_usd": newBalanceUSD,
			"wallet_type":     "registry",
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendWalletLowBalance notifies a user when wallet balance drops below threshold.
func (s *Service) SendWalletLowBalance(ctx context.Context, userID uuid.UUID, agentID string, balanceUSD, thresholdUSD float64) error {
	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeBillingWalletLowBalance,
		Category: CategoryBilling,
		Title:    "Low Wallet Balance",
		Body: fmt.Sprintf(
			"Wallet balance is low ($%.2f). It is now at or below your alert threshold of $%.2f.",
			balanceUSD,
			thresholdUSD,
		),
		Data: JSONMap{
			"agent_id":      agentID,
			"balance_usd":   balanceUSD,
			"threshold_usd": thresholdUSD,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

// SendBillingInvoiceGenerated sends an invoice generated notification when a new invoice is created.
func (s *Service) SendBillingInvoiceGenerated(ctx context.Context, userEmail, period string, amountDueUSD float64, invoiceURL, invoiceID string) error {
	user, err := s.db.GetUserByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found for email: %s", userEmail)
	}

	_, err = s.Send(ctx, SendRequest{
		UserID:   user.ID,
		Type:     TypeBillingInvoiceGenerated,
		Category: CategoryBilling,
		Title:    "New Invoice Available",
		Body:     fmt.Sprintf("Your invoice for %s ($%.2f) is ready for payment.", period, amountDueUSD),
		Data: JSONMap{
			"period":      period,
			"amount_usd":  amountDueUSD,
			"invoice_url": invoiceURL,
			"invoice_id":  invoiceID,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendBillingPaymentSuccess sends a payment success notification for subscription payments.
func (s *Service) SendBillingPaymentSuccess(ctx context.Context, userEmail, period string, amountPaidUSD float64, invoiceID string) error {
	user, err := s.db.GetUserByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found for email: %s", userEmail)
	}

	_, err = s.Send(ctx, SendRequest{
		UserID:   user.ID,
		Type:     TypeBillingPaymentSuccess,
		Category: CategoryBilling,
		Title:    "Payment Successful",
		Body:     fmt.Sprintf("Your payment of $%.2f for %s has been successfully processed. Thank you!", amountPaidUSD, period),
		Data: JSONMap{
			"period":       period,
			"amount_usd":   amountPaidUSD,
			"invoice_id":   invoiceID,
			"payment_date": time.Now().Format(time.RFC3339),
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}

// SendUpcomingRenewalNotice sends a notification about an upcoming subscription renewal.
func (s *Service) SendUpcomingRenewalNotice(ctx context.Context, userID uuid.UUID, period string, amountUSD float64, renewalDate time.Time, daysUntil int) error {
	var title, body string
	priority := PriorityNormal

	switch daysUntil {
	case 7:
		title = "Subscription Renewal in 1 Week"
		body = fmt.Sprintf("Your subscription will renew on %s for $%.2f. Ensure your payment method is up to date.", renewalDate.Format("Jan 2, 2006"), amountUSD)
	case 3:
		title = "Subscription Renewal in 3 Days"
		body = fmt.Sprintf("Your subscription renewal ($%.2f for %s) is approaching. Update your payment method if needed.", amountUSD, period)
		priority = PriorityHigh
	case 1:
		title = "Subscription Renews Tomorrow"
		body = fmt.Sprintf("Your subscription for %s ($%.2f) will renew tomorrow. Make sure your payment method is valid.", period, amountUSD)
		priority = PriorityHigh
	default:
		title = "Upcoming Subscription Renewal"
		body = fmt.Sprintf("Your subscription for %s ($%.2f) will renew on %s.", period, amountUSD, renewalDate.Format("Jan 2, 2006"))
	}

	_, err := s.Send(ctx, SendRequest{
		UserID:   userID,
		Type:     TypeBillingSubscriptionExpiring,
		Category: CategoryBilling,
		Title:    title,
		Body:     body,
		Data: JSONMap{
			"period":       period,
			"amount_usd":   amountUSD,
			"renewal_date": renewalDate.Format(time.RFC3339),
			"days_until":   daysUntil,
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: priority,
	})
	return err
}

// SendBillingAlert sends a billing-related alert notification (e.g., payment failed, invoice overdue).
func (s *Service) SendBillingAlert(ctx context.Context, userEmail string, alertType string, data map[string]interface{}) error {
	user, err := s.db.GetUserByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found for email: %s", userEmail)
	}

	var title, body string
	priority := PriorityHigh

	switch alertType {
	case "payment_failed":
		amountDue, _ := data["amount_due"].(float64)
		currency, _ := data["currency"].(string)
		attemptCount, _ := data["attempt_count"].(int64)
		title = "Payment Failed"
		body = fmt.Sprintf("Your payment of %.2f %s failed (attempt %d). Please update your payment method to avoid service interruption.", amountDue, currency, attemptCount)
	case "invoice_overdue":
		title = "Invoice Overdue"
		body = "You have an overdue invoice. Please make a payment to avoid service suspension."
	case "subscription_cancelled":
		title = "Subscription Cancelled"
		body = "Your subscription has been cancelled. You will be downgraded to the free plan at the end of your billing period."
	default:
		title = "Billing Alert"
		body = "There is an issue with your billing account. Please review your payment settings."
	}

	_, err = s.Send(ctx, SendRequest{
		UserID:   user.ID,
		Type:     TypeBillingAlert,
		Category: CategoryBilling,
		Title:    title,
		Body:     body,
		Data:     data,
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: priority,
	})
	return err
}

// SendLowBalance sends a low balance alert to a user via email
func (s *Service) SendLowBalance(ctx context.Context, userEmail string, data map[string]interface{}) error {
	user, err := s.db.GetUserByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found for email: %s", userEmail)
	}

	balanceUSD, _ := data["balance_usd"].(float64)
	thresholdUSD, _ := data["threshold_usd"].(float64)
	severity, _ := data["severity"].(string)
	autoTopupEnabled, _ := data["auto_topup_enabled"].(bool)

	title := "Low Wallet Balance"
	body := fmt.Sprintf("Your wallet balance is low ($%.2f USD). It is now at or below your alert threshold of $%.2f USD.", balanceUSD, thresholdUSD)

	if severity == "critical" {
		title = "Critical: Very Low Wallet Balance"
		body = fmt.Sprintf("URGENT: Your wallet balance is critically low ($%.2f USD). Add funds immediately to avoid service interruption.", balanceUSD)
	}

	_, err = s.Send(ctx, SendRequest{
		UserID:   user.ID,
		Type:     TypeBillingWalletLowBalance,
		Category: CategoryBilling,
		Title:    title,
		Body:     body,
		Data: JSONMap{
			"balance_usd":        balanceUSD,
			"threshold_usd":      thresholdUSD,
			"severity":           severity,
			"auto_topup_enabled": autoTopupEnabled,
			"currency":           data["currency"],
			"balance_local":      data["balance_local"],
		},
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityHigh,
	})
	return err
}

// SendLowBalanceNotification sends an in-app low balance notification
func (s *Service) SendLowBalanceNotification(ctx context.Context, userID interface{}, data map[string]interface{}) error {
	uid, err := parseFlexibleUserID(userID)
	if err != nil {
		return err
	}

	balanceUSD, _ := data["balance_usd"].(float64)
	thresholdUSD, _ := data["threshold_usd"].(float64)
	severity, _ := data["severity"].(string)

	title := "Low Wallet Balance"
	body := fmt.Sprintf("Your wallet balance ($%.2f) is below the threshold ($%.2f).", balanceUSD, thresholdUSD)

	if severity == "critical" {
		title = "Critical: Low Balance"
		body = fmt.Sprintf("URGENT: Your balance ($%.2f) is critically low. Add funds now.", balanceUSD)
	}

	_, err = s.Send(ctx, SendRequest{
		UserID:   uid,
		Type:     TypeBillingWalletLowBalance,
		Category: CategoryBilling,
		Title:    title,
		Body:     body,
		Data:     data,
		Channels: []string{ChannelInApp},
		Priority: PriorityHigh,
	})
	return err
}

// SendAutoTopupApproaching sends a notification that auto-topup threshold is approaching
func (s *Service) SendAutoTopupApproaching(ctx context.Context, userID interface{}, data map[string]interface{}) error {
	uid, err := parseFlexibleUserID(userID)
	if err != nil {
		return err
	}

	balanceUSD, _ := data["balance_usd"].(float64)
	autoTopupThreshold, _ := data["auto_topup_threshold"].(float64)

	_, err = s.Send(ctx, SendRequest{
		UserID:   uid,
		Type:     TypeBillingWalletLowBalance,
		Category: CategoryBilling,
		Title:    "Auto-Topup Approaching",
		Body:     fmt.Sprintf("Your balance ($%.2f) is approaching your auto-topup threshold ($%.2f). Funds will be added automatically soon.", balanceUSD, autoTopupThreshold),
		Data:     data,
		Channels: []string{ChannelInApp, ChannelEmail},
		Priority: PriorityNormal,
	})
	return err
}
