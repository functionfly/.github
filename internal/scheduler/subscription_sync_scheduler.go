package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	stripe "github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/customer"
	"github.com/stripe/stripe-go/v83/paymentmethod"
	"github.com/stripe/stripe-go/v83/subscription"
)

// SubscriptionSyncScheduleConfig represents configuration for the subscription sync scheduler
type SubscriptionSyncScheduleConfig struct {
	Cron string
}

// SubscriptionSyncScheduler periodically syncs subscription status from Stripe to ensure consistency
type SubscriptionSyncScheduler struct {
	cron      *cron.Cron
	repo      storage.Repository
	logger    *logrus.Logger
	stopOnce  sync.Once
	cancel    context.CancelFunc
	notifySvc *notification.Service
	stripeKey string
}

// NewSubscriptionSyncScheduler creates a new subscription sync scheduler
func NewSubscriptionSyncScheduler(repo storage.Repository, notifySvc *notification.Service) *SubscriptionSyncScheduler {
	return &SubscriptionSyncScheduler{
		cron:      cron.New(),
		repo:      repo,
		logger:    logrus.New(),
		notifySvc: notifySvc,
		stripeKey: os.Getenv("STRIPE_SECRET_KEY"),
	}
}

// Start begins the subscription sync scheduler
func (s *SubscriptionSyncScheduler) Start(ctx context.Context, config SubscriptionSyncScheduleConfig) error {
	if config.Cron == "" {
		config.Cron = "0 */6 * * *" // Default: every 6 hours
	}

	if s.stripeKey == "" {
		s.logger.Warn("STRIPE_SECRET_KEY not configured, subscription sync scheduler will not run")
		return nil
	}

	stripe.Key = s.stripeKey

	var ctxWithCancel context.Context
	ctxWithCancel, s.cancel = context.WithCancel(ctx)

	// Add the sync job
	_, err := s.cron.AddFunc(config.Cron, func() {
		s.runSync(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add subscription sync cron job: %w", err)
	}

	s.cron.Start()
	s.logger.Infof("Subscription sync scheduler started with cron: %s", config.Cron)
	return nil
}

// Stop stops the subscription sync scheduler
func (s *SubscriptionSyncScheduler) Stop() error {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		<-s.cron.Stop().Done()
		s.logger.Info("Subscription sync scheduler stopped")
	})
	return nil
}

// runSync performs the subscription synchronization
func (s *SubscriptionSyncScheduler) runSync(ctx context.Context) {
	s.logger.Info("Starting subscription sync with Stripe")

	// Get all active subscriptions from our database
	subs, err := s.repo.ListAllSubscriptions(ctx, 1000, 0)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list subscriptions for sync")
		return
	}

	s.logger.Infof("Found %d subscriptions to sync", len(subs))

	var syncErrors int
	var updatedCount int

	for _, sub := range subs {
		if err := s.syncSubscription(ctx, sub); err != nil {
			s.logger.WithError(err).WithField("subscription_id", sub.ID).Error("Failed to sync subscription")
			syncErrors++
			continue
		}
		updatedCount++
	}

	s.logger.WithFields(logrus.Fields{
		"total":   len(subs),
		"updated": updatedCount,
		"errors":  syncErrors,
	}).Info("Subscription sync completed")
}

// syncSubscription syncs a single subscription with Stripe
func (s *SubscriptionSyncScheduler) syncSubscription(ctx context.Context, sub *storage.Subscription) error {
	if sub.StripeSubscriptionID == "" {
		// No Stripe subscription ID, skip
		return nil
	}

	// Fetch the subscription from Stripe
	stripeSub, err := subscription.Get(sub.StripeSubscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription from Stripe: %w", err)
	}

	// Map Stripe status to our status
	stripeStatus := string(stripeSub.Status)
	ourStatus := mapStripeStatus(stripeStatus)

	// Check if status has changed
	if sub.Status != ourStatus {
		s.logger.WithFields(logrus.Fields{
			"subscription_id": sub.ID,
			"stripe_id":       sub.StripeSubscriptionID,
			"old_status":      sub.Status,
			"new_status":      ourStatus,
		}).Info("Subscription status changed, updating")

		updates := map[string]interface{}{
			"status": ourStatus,
		}

		// Update additional fields from Stripe
		if stripeSub.CanceledAt > 0 {
			cancelledAt := time.Unix(stripeSub.CanceledAt, 0)
			updates["canceled_at"] = &cancelledAt
		}
		if stripeSub.TrialEnd > 0 {
			trialEnd := time.Unix(stripeSub.TrialEnd, 0)
			updates["trial_end"] = &trialEnd
		}
		updates["cancel_at_period_end"] = stripeSub.CancelAtPeriodEnd

		_, err := s.repo.UpdateSubscription(ctx, sub.ID, updates)
		if err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}

		// Send notification for important status changes
		if s.notifySvc != nil {
			s.sendStatusChangeNotification(ctx, sub, ourStatus)
		}
	}

	return nil
}

// mapStripeStatus maps Stripe subscription status to our internal status
func mapStripeStatus(stripeStatus string) string {
	switch stripeStatus {
	case "active":
		return "active"
	case "canceled":
		return "cancelled"
	case "incomplete":
		return "incomplete"
	case "incomplete_expired":
		return "expired"
	case "past_due":
		return "past_due"
	case "paused":
		return "paused"
	case "trialing":
		return "trialing"
	case "unpaid":
		return "unpaid"
	default:
		return stripeStatus
	}
}

// sendStatusChangeNotification sends a notification for important status changes
func (s *SubscriptionSyncScheduler) sendStatusChangeNotification(ctx context.Context, sub *storage.Subscription, newStatus string) {
	// Only notify for significant changes
	var title, body string
	priority := notification.PriorityNormal

	switch newStatus {
	case "past_due":
		title = "Payment Past Due"
		body = "Your subscription payment is past due. Please update your payment method to avoid service interruption."
		priority = notification.PriorityHigh
	case "cancelled":
		title = "Subscription Cancelled"
		body = "Your subscription has been cancelled. You will be downgraded to the free plan at the end of your billing period."
	case "unpaid":
		title = "Subscription Unpaid"
		body = "Your subscription is unpaid. Please make a payment to restore service."
		priority = notification.PriorityHigh
	case "active":
		// Don't spam for active status - this is the normal state
		return
	default:
		// Don't notify for other status changes
		return
	}

	_, err := s.notifySvc.Send(ctx, notification.SendRequest{
		UserID:   sub.TenantID, // Using tenant ID as proxy for user
		Type:     notification.TypeBillingAlert,
		Category: notification.CategoryBilling,
		Title:    title,
		Body:     body,
		Data: map[string]interface{}{
			"subscription_id": sub.ID.String(),
			"new_status":      newStatus,
			"synced_at":       time.Now().Format(time.RFC3339),
		},
		Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		Priority: priority,
	})
	if err != nil {
		s.logger.WithError(err).Warn("Failed to send subscription status change notification")
	}
}

// SyncPaymentMethods syncs payment methods for all tenants with Stripe customer IDs
// This is called as part of the periodic sync to ensure consistency
func (s *SubscriptionSyncScheduler) SyncPaymentMethods(ctx context.Context) {
	s.logger.Info("Starting payment method sync with Stripe")

	// Get all tenants with Stripe customer IDs
	tenants, err := s.repo.ListTenantsWithStripeCustomerID(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list tenants with Stripe customer IDs")
		return
	}

	s.logger.Infof("Found %d tenants to sync payment methods", len(tenants))

	for _, tenant := range tenants {
		if err := s.syncTenantPaymentMethod(ctx, tenant); err != nil {
			s.logger.WithError(err).WithField("tenant_id", tenant.ID).Error("Failed to sync payment method")
		}
	}

	s.logger.Info("Payment method sync completed")
}

// syncTenantPaymentMethod syncs a single tenant's default payment method from Stripe
func (s *SubscriptionSyncScheduler) syncTenantPaymentMethod(ctx context.Context, tenant *storage.Tenant) error {
	if tenant.StripeCustomerID == nil || *tenant.StripeCustomerID == "" {
		return nil
	}

	// Fetch customer from Stripe to get default payment method
	params := &stripe.CustomerParams{}
	params.AddExpand("default_payment_method")
	c, err := customer.Get(*tenant.StripeCustomerID, params)
	if err != nil {
		return fmt.Errorf("failed to fetch customer from Stripe: %w", err)
	}

	// Check for default payment method via InvoiceSettings
	if c.InvoiceSettings != nil && c.InvoiceSettings.DefaultPaymentMethod != nil && c.InvoiceSettings.DefaultPaymentMethod.ID != "" {
		pm, err := paymentmethod.Get(c.InvoiceSettings.DefaultPaymentMethod.ID, nil)
		if err != nil {
			s.logger.WithError(err).WithField("payment_method_id", c.InvoiceSettings.DefaultPaymentMethod.ID).Warn("Failed to fetch payment method details")
			return nil // Don't fail the sync for this
		}

		// Only sync card payment methods
		if pm.Card == nil {
			return nil
		}

		// Build billing details
		billingDetails := map[string]interface{}{}
		if pm.BillingDetails.Name != "" {
			billingDetails["name"] = pm.BillingDetails.Name
		}
		if pm.BillingDetails.Email != "" {
			billingDetails["email"] = pm.BillingDetails.Email
		}
		if pm.BillingDetails.Address != nil {
			billingDetails["address"] = map[string]interface{}{
				"line1":       pm.BillingDetails.Address.Line1,
				"line2":       pm.BillingDetails.Address.Line2,
				"city":        pm.BillingDetails.Address.City,
				"state":       pm.BillingDetails.Address.State,
				"postal_code": pm.BillingDetails.Address.PostalCode,
				"country":     pm.BillingDetails.Address.Country,
			}
		}
		billingDetailsJSON, _ := json.Marshal(billingDetails)

		// Update tenant payment method
		paymentMethod := &storage.PaymentMethodInfoExtended{
			StripePaymentMethodID: pm.ID,
			Brand:                 string(pm.Card.Brand),
			Last4:                 pm.Card.Last4,
			ExpMonth:              int(pm.Card.ExpMonth),
			ExpYear:               int(pm.Card.ExpYear),
			BillingDetails:        billingDetailsJSON,
		}

		if err := s.repo.UpdateTenantPaymentMethod(ctx, tenant.ID, paymentMethod); err != nil {
			return fmt.Errorf("failed to update tenant payment method: %w", err)
		}

		s.logger.WithFields(logrus.Fields{
			"tenant_id":         tenant.ID,
			"payment_method_id": pm.ID,
			"brand":             pm.Card.Brand,
			"last4":             pm.Card.Last4,
		}).Info("Payment method synced from Stripe")
	}

	return nil
}

// Helper imports for the scheduler - need to add to imports
func init() {
	// Ensure stripe.Key is set before making API calls
	if stripe.Key == "" && os.Getenv("STRIPE_SECRET_KEY") != "" {
		stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	}
}
