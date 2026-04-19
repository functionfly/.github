package scheduler

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// UpcomingRenewalConfig represents configuration for the upcoming renewal scheduler
type UpcomingRenewalConfig struct {
	Cron string
}

// UpcomingRenewalScheduler sends notifications about upcoming subscription renewals
type UpcomingRenewalScheduler struct {
	cron       *cron.Cron
	repo       storage.Repository
	logger     *logrus.Logger
	stopOnce   sync.Once
	cancel     context.CancelFunc
	notifySvc  *notification.Service
	noticeDays []int // Days before renewal to send notice (e.g., [7, 3, 1])
}

// NewUpcomingRenewalScheduler creates a new upcoming renewal scheduler
func NewUpcomingRenewalScheduler(repo storage.Repository, notifySvc *notification.Service) *UpcomingRenewalScheduler {
	return &UpcomingRenewalScheduler{
		cron:       cron.New(),
		repo:       repo,
		logger:     logrus.New(),
		notifySvc:  notifySvc,
		noticeDays: []int{7, 3, 1}, // Notify at 7 days, 3 days, and 1 day before renewal
	}
}

// Start begins the upcoming renewal scheduler
func (s *UpcomingRenewalScheduler) Start(ctx context.Context, config UpcomingRenewalConfig) error {
	if config.Cron == "" {
		config.Cron = "0 9 * * *" // Default: daily at 9 AM
	}

	if s.notifySvc == nil {
		s.logger.Warn("Notification service not configured, upcoming renewal scheduler will not run")
		return nil
	}

	var ctxWithCancel context.Context
	ctxWithCancel, s.cancel = context.WithCancel(ctx)

	// Add the renewal check job
	_, err := s.cron.AddFunc(config.Cron, func() {
		s.runRenewalCheck(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add upcoming renewal cron job: %w", err)
	}

	s.cron.Start()
	s.logger.Infof("Upcoming renewal scheduler started with cron: %s", config.Cron)
	return nil
}

// Stop stops the upcoming renewal scheduler
func (s *UpcomingRenewalScheduler) Stop() error {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		<-s.cron.Stop().Done()
		s.logger.Info("Upcoming renewal scheduler stopped")
	})
	return nil
}

// runRenewalCheck performs the renewal check for all active subscriptions
func (s *UpcomingRenewalScheduler) runRenewalCheck(ctx context.Context) {
	s.logger.Info("Starting upcoming renewal check")

	// Check each notice window
	for _, daysUntil := range s.noticeDays {
		targetDate := time.Now().UTC().AddDate(0, 0, daysUntil)
		if err := s.checkRenewalsForDate(ctx, targetDate, daysUntil); err != nil {
			s.logger.WithError(err).WithField("days_until", daysUntil).Error("Failed to check renewals")
		}
	}

	s.logger.Info("Upcoming renewal check completed")
}

// checkRenewalsForDate finds subscriptions renewing on the target date and sends notifications
func (s *UpcomingRenewalScheduler) checkRenewalsForDate(ctx context.Context, targetDate time.Time, daysUntil int) error {
	// Get all active subscriptions
	subs, err := s.repo.ListAllSubscriptions(1000, 0)
	if err != nil {
		return fmt.Errorf("failed to list subscriptions: %w", err)
	}

	// Track how many notifications were sent
	var notificationCount int

	for _, sub := range subs {
		// Skip if not an active subscription
		if sub.Status != "active" && sub.Status != "trialing" {
			continue
		}

		// Check if renewal is on the target date (within a day window)
		if !s.isRenewalOnDate(sub, targetDate) {
			continue
		}

		// Get tenant info for billing details
		tenant, err := s.repo.GetTenantByID(sub.TenantID)
		if err != nil {
			s.logger.WithError(err).WithField("tenant_id", sub.TenantID).Warn("Failed to get tenant for renewal notice")
			continue
		}

		// Get users to notify (primary billing contact)
		users, err := s.repo.ListActiveUsersByTenant(ctx, sub.TenantID)
		if err != nil || len(users) == 0 {
			s.logger.WithError(err).WithField("tenant_id", sub.TenantID).Warn("No active users found for renewal notice")
			continue
		}

		// Calculate renewal amount
		amountUSD := s.calculateRenewalAmount(sub, tenant)

		// Send notification to each user
		for _, user := range users {
			// Skip users without email
			if user.Email == "" {
				continue
			}

			period := "Monthly Subscription"
			if sub.CurrentPeriodStart.IsZero() || sub.CurrentPeriodEnd.IsZero() {
				// Use current period if available
				period = fmt.Sprintf("%s - %s",
					sub.CurrentPeriodStart.Format("Jan 2, 2006"),
					sub.CurrentPeriodEnd.Format("Jan 2, 2006"))
			}

			if err := s.notifySvc.SendUpcomingRenewalNotice(ctx, user.ID, period, amountUSD, sub.CurrentPeriodEnd, daysUntil); err != nil {
				s.logger.WithError(err).WithFields(logrus.Fields{
					"user_id":    user.ID,
					"tenant_id":  sub.TenantID,
					"days_until": daysUntil,
				}).Warn("Failed to send upcoming renewal notice")
			} else {
				notificationCount++
				s.logger.WithFields(logrus.Fields{
					"user_id":    user.ID,
					"tenant_id":  sub.TenantID,
					"days_until": daysUntil,
					"amount_usd": amountUSD,
				}).Debug("Upcoming renewal notice sent")
			}

			// Only notify the first user (primary contact) per tenant
			break
		}
	}

	if notificationCount > 0 {
		s.logger.WithFields(logrus.Fields{
			"days_until":         daysUntil,
			"notifications_sent": notificationCount,
		}).Info("Upcoming renewal notices sent")
	}

	return nil
}

// isRenewalOnDate checks if a subscription renews on or near the target date
func (s *UpcomingRenewalScheduler) isRenewalOnDate(sub *storage.Subscription, targetDate time.Time) bool {
	// Handle zero time
	if sub.CurrentPeriodEnd.IsZero() {
		return false
	}

	// Normalize to date only (ignore time)
	renewalDate := sub.CurrentPeriodEnd.Truncate(24 * time.Hour)
	target := targetDate.Truncate(24 * time.Hour)

	// Check if renewal date matches target date (within a small window for safety)
	diff := renewalDate.Sub(target)
	if diff < 0 {
		diff = -diff
	}

	// Allow for a 1-day window to handle any edge cases
	return diff <= 24*time.Hour
}

// calculateRenewalAmount calculates the renewal amount in USD
func (s *UpcomingRenewalScheduler) calculateRenewalAmount(sub *storage.Subscription, tenant *storage.Tenant) float64 {
	// If subscription has a pricing tier, get the price from that
	if sub.PricingTierID != uuid.Nil && sub.PricingTier != nil {
		return float64(sub.PricingTier.PriceCents) / 100.0
	}

	// Default to checking Stripe subscription if available
	if sub.StripeSubscriptionID != "" {
		// In a real implementation, you might want to fetch from Stripe API
		// For now, return a placeholder that will be resolved at billing time
		return 0.0 // Unknown - will be resolved by Stripe
	}

	return 0.0
}

// getPricingTier retrieves a pricing tier by ID (helper to avoid import cycles)
func (s *UpcomingRenewalScheduler) getPricingTier(tierID uuid.UUID) (*storage.PricingTier, error) {
	// This is a simplified implementation
	// In production, you'd use a proper repository method
	return nil, fmt.Errorf("not implemented")
}

// SendImmediateRenewalNotice sends an immediate renewal notice for a specific subscription
// This can be called programmatically when needed
func (s *UpcomingRenewalScheduler) SendImmediateRenewalNotice(ctx context.Context, tenantID uuid.UUID, amountUSD float64, renewalDate time.Time) error {
	if s.notifySvc == nil {
		return fmt.Errorf("notification service not available")
	}

	// Calculate days until renewal
	daysUntil := int(renewalDate.Sub(time.Now().UTC()).Hours() / 24)
	if daysUntil < 0 {
		daysUntil = 0
	}

	// Get users to notify
	users, err := s.repo.ListActiveUsersByTenant(ctx, tenantID)
	if err != nil || len(users) == 0 {
		return fmt.Errorf("no active users found: %w", err)
	}

	// Send to first user
	period := renewalDate.Format("Jan 2, 2006")
	return s.notifySvc.SendUpcomingRenewalNotice(ctx, users[0].ID, period, amountUSD, renewalDate, daysUntil)
}

// init function to ensure logrus is configured
func init() {
	if os.Getenv("DEBUG") == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
