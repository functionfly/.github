package billing

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/invoice"
)

// DeferredBillingChecker runs background jobs to check founder mode trigger thresholds
// and transition accounts from "building" to "grace_period" to "converted" status
type DeferredBillingChecker struct {
	repo            storage.Repository
	notificationSvc *notification.Service
	ticker          *time.Ticker
	stop            chan bool
	checkInterval   time.Duration
}

// NewDeferredBillingChecker creates a new checker with the given check interval
func NewDeferredBillingChecker(repo storage.Repository, notificationSvc *notification.Service, checkInterval time.Duration) *DeferredBillingChecker {
	if checkInterval <= 0 {
		checkInterval = 24 * time.Hour // Default: check once per day
	}
	return &DeferredBillingChecker{
		repo:            repo,
		notificationSvc: notificationSvc,
		stop:            make(chan bool),
		checkInterval:   checkInterval,
	}
}

// Start begins the background checking loop
func (d *DeferredBillingChecker) Start() {
	d.ticker = time.NewTicker(d.checkInterval)

	go func() {
		// Run immediately on start
		if err := d.checkThresholds(context.Background()); err != nil {
			logrus.WithError(err).Error("deferred billing: initial threshold check failed")
		}

		for {
			select {
			case <-d.ticker.C:
				if err := d.checkThresholds(context.Background()); err != nil {
					logrus.WithError(err).Error("deferred billing: threshold check failed")
				}
			case <-d.stop:
				return
			}
		}
	}()

	logrus.Info("Deferred billing checker started")
}

// Stop halts the background checking loop
func (d *DeferredBillingChecker) Stop() {
	if d.ticker != nil {
		d.ticker.Stop()
	}
	close(d.stop)
	logrus.Info("Deferred billing checker stopped")
}

// checkThresholds evaluates all active founder mode registrations
// and transitions them based on trigger thresholds
func (d *DeferredBillingChecker) checkThresholds(ctx context.Context) error {
	// Get all active founder mode registrations
	registrations, err := d.repo.ListAllActiveFounderModes(ctx)
	if err != nil {
		return err
	}

	if len(registrations) == 0 {
		logrus.Debug("No active founder mode registrations to check")
		return nil
	}

	logrus.WithField("count", len(registrations)).Info("Checking founder mode thresholds")

	now := time.Now().UTC()

	for _, reg := range registrations {
		if err := d.checkSingleRegistration(ctx, reg, now); err != nil {
			logrus.WithError(err).WithField("registration_id", reg.ID).Error("Failed to check registration")
			// Continue with next registration
		}
	}

	return nil
}

// checkSingleRegistration checks a single founder mode registration
func (d *DeferredBillingChecker) checkSingleRegistration(ctx context.Context, reg *storage.FounderModeRegistration, now time.Time) error {
	log := logrus.WithFields(logrus.Fields{
		"registration_id": reg.ID,
		"tenant_id":       reg.TenantID,
		"bundle_id":       reg.BundleID,
	})

	// Get deferred billing config for this bundle
	config, err := d.repo.GetDeferredBillingConfig(ctx, reg.BundleID)
	if err != nil {
		return err
	}
	if config == nil {
		log.Warn("No deferred billing config found, using defaults")
		config = &storage.DeferredBillingConfig{
			TriggerUserCount:    intPtr(100),
			TriggerRevenueCents: intPtr(100000), // $1000
			TriggerDaysElapsed:  intPtr(90),
			GracePeriodDays:     7,
		}
	}

	// Get tenant metrics
	metrics := d.getTenantMetrics(ctx, reg.TenantID, now)

	// Update progress tracking
	if err := d.repo.UpdateFounderModeProgress(ctx, reg.ID, metrics.UserCount, metrics.MRRCents, metrics.APICalls); err != nil {
		log.WithError(err).Error("Failed to update founder mode progress")
	}

	// Build thresholds
	thresholds := TriggerThresholds{
		UserCount:   derefOrDefault(config.TriggerUserCount, 100),
		MRRCents:    derefOrDefault(config.TriggerRevenueCents, 100000),
		APICalls:    derefOrDefault(config.TriggerAPICalls, 0),
		DaysElapsed: derefOrDefault(config.TriggerDaysElapsed, 90),
	}

	current := CurrentProgress{
		UserCount:       metrics.UserCount,
		MRRCents:        metrics.MRRCents,
		APICalls:        metrics.APICalls,
		DaysElapsed:     int(now.Sub(reg.StartedAt).Hours() / 24),
		ProgressPercent: 0,
	}

	// Calculate progress
	current = calculateProgress(thresholds, current)

	log.WithFields(logrus.Fields{
		"progress_percent": current.ProgressPercent,
		"users":            current.UserCount,
		"mrr_cents":        current.MRRCents,
		"days_elapsed":     current.DaysElapsed,
	}).Debug("Founder mode progress")

	// Get bundle for notifications
	bundle, _ := d.repo.GetPricingBundleByID(ctx, reg.BundleID)
	bundleName := "Bundle"
	if bundle != nil {
		bundleName = bundle.Name
	}

	// Check if we should trigger billing (any threshold reached)
	if shouldTriggerBilling(thresholds, current) {
		log.Info("Threshold reached - starting grace period")

		// Start grace period
		if err := d.repo.StartGracePeriod(ctx, reg.ID, config.GracePeriodDays); err != nil {
			return err
		}

		// Send threshold reached notification
		if d.notificationSvc != nil {
			// Get first user in tenant for notification
			users, err := d.repo.ListActiveUsersByTenant(ctx, reg.TenantID)
			if err == nil && len(users) > 0 {
				thresholdType := d.getExceededThreshold(thresholds, current)
				if err := d.notificationSvc.SendFounderModeThresholdReached(ctx, users[0].ID, bundleName, thresholdType, config.GracePeriodDays); err != nil {
					log.WithError(err).Warn("Failed to send threshold reached notification")
				}
			}
		}

		return nil
	}

	// Check if approaching threshold (80% of any metric) - only notify once per threshold
	if current.ProgressPercent >= 80 && current.ProgressPercent < 100 {
		if reg.Status == "active" {
			log.Info("Approaching threshold - sending warning")

			if d.notificationSvc != nil {
				users, err := d.repo.ListActiveUsersByTenant(ctx, reg.TenantID)
				if err == nil && len(users) > 0 {
					thresholdType, _ := d.getClosestThreshold(thresholds, current)
					thresholdValue := d.getThresholdValue(thresholds, thresholdType)
					if err := d.notificationSvc.SendFounderModeThresholdWarning(ctx, users[0].ID, bundleName, current.ProgressPercent, thresholdType, thresholdValue); err != nil {
						log.WithError(err).Warn("Failed to send threshold warning notification")
					}
				}
			}
		}
	}

	// Check if in grace period and ending soon
	if reg.Status == "grace_period" && reg.GracePeriodEndsAt != nil {
		daysUntilExpiry := int(reg.GracePeriodEndsAt.Sub(now).Hours() / 24)

		// Send notification at 3 days and 1 day before expiry
		if daysUntilExpiry == 3 || daysUntilExpiry == 1 {
			log.WithField("days_left", daysUntilExpiry).Info("Grace period ending soon")

			if d.notificationSvc != nil {
				users, err := d.repo.ListActiveUsersByTenant(ctx, reg.TenantID)
				if err == nil && len(users) > 0 {
					if err := d.notificationSvc.SendFounderModeGracePeriodEnding(ctx, users[0].ID, bundleName, daysUntilExpiry); err != nil {
						log.WithError(err).Warn("Failed to send grace period ending notification")
					}
				}
			}
		}
	}

	return nil
}

// getTenantMetrics retrieves current metrics for a tenant from the database
func (d *DeferredBillingChecker) getTenantMetrics(ctx context.Context, tenantID uuid.UUID, now time.Time) TenantMetrics {
	metrics := TenantMetrics{
		UserCount: 0,
		MRRCents:  0,
		APICalls:  0,
	}

	// Get user count for this tenant
	// This counts active users in the tenant
	userCount, err := d.repo.CountActiveUsersByTenant(ctx, tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to get tenant user count")
	} else {
		metrics.UserCount = userCount
	}

	// Get API calls from usage rollups for the current month
	// Look at function_execution events
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	execRollups, err := d.repo.GetUsageByTenant(ctx, tenantID, "function_execution", periodStart, periodEnd)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to get usage rollups")
	} else {
		for _, rollup := range execRollups {
			metrics.APICalls += rollup.TotalQuantity
		}
	}

	// Get MRR from Stripe if customer ID exists
	tenant, err := d.repo.GetTenantByID(ctx, tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to get tenant for MRR lookup")
	} else if tenant != nil && tenant.StripeCustomerID != nil && *tenant.StripeCustomerID != "" {
		stripeMRR, err := d.getMRRFromStripe(ctx, *tenant.StripeCustomerID)
		if err != nil {
			logrus.WithError(err).WithField("stripe_customer_id", *tenant.StripeCustomerID).Warn("Failed to fetch MRR from Stripe")
		} else {
			metrics.MRRCents = stripeMRR
			logrus.WithFields(logrus.Fields{
				"tenant_id": tenantID,
				"mrr_cents": stripeMRR,
			}).Debug("Fetched MRR from Stripe invoices")
		}
	}

	return metrics
}

// getMRRFromStripe fetches paid subscription invoice totals from Stripe for the last 30 days
func (d *DeferredBillingChecker) getMRRFromStripe(ctx context.Context, stripeCustomerID string) (int, error) {
	if stripeCustomerID == "" {
		return 0, nil
	}

	// Set API key if not already configured
	if stripe.Key == "" {
		stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	}
	if stripe.Key == "" {
		return 0, fmt.Errorf("STRIPE_SECRET_KEY not configured")
	}

	// Calculate month boundaries for MRR calculation
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	// Query paid invoices for this customer
	params := &stripe.InvoiceListParams{
		Customer: stripe.String(stripeCustomerID),
		Status:   stripe.String("paid"),
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(100),
		},
	}
	// Filter to current month's invoices for accurate MRR
	params.CreatedRange = &stripe.RangeQueryParams{
		GreaterThanOrEqual: monthStart.Unix(),
		LesserThanOrEqual:  monthEnd.Unix(),
	}

	totalCents := 0
	subscriptionInvoices := 0

	i := invoice.List(params)
	for i.Next() {
		inv := i.Invoice()
		// Only count recurring subscription invoices (exclude one-time purchases like wallet credits)
		// BillingReason values: subscription_cycle, subscription_create, subscription_update, manual, etc.
		if inv.BillingReason == "subscription_cycle" || inv.BillingReason == "subscription_create" {
			totalCents += int(inv.AmountPaid)
			subscriptionInvoices++
		}
	}

	if err := i.Err(); err != nil {
		return 0, fmt.Errorf("stripe invoice list failed: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"stripe_customer_id":    stripeCustomerID,
		"subscription_invoices": subscriptionInvoices,
		"total_mrr_cents":       totalCents,
	}).Debug("Calculated MRR from Stripe invoices")

	return totalCents, nil
}

// getExceededThreshold returns which threshold was exceeded
func (d *DeferredBillingChecker) getExceededThreshold(thresholds TriggerThresholds, current CurrentProgress) string {
	if thresholds.UserCount > 0 && current.UserCount >= thresholds.UserCount {
		return "user_count"
	}
	if thresholds.MRRCents > 0 && current.MRRCents >= thresholds.MRRCents {
		return "mrr"
	}
	if thresholds.APICalls > 0 && current.APICalls >= thresholds.APICalls {
		return "api_calls"
	}
	if thresholds.DaysElapsed > 0 && current.DaysElapsed >= thresholds.DaysElapsed {
		return "time"
	}
	return "unknown"
}

// getClosestThreshold returns which threshold is closest to being exceeded and the percentage
func (d *DeferredBillingChecker) getClosestThreshold(thresholds TriggerThresholds, current CurrentProgress) (string, float64) {
	maxPercent := 0.0
	closest := ""

	if thresholds.UserCount > 0 {
		p := float64(current.UserCount) / float64(thresholds.UserCount) * 100
		if p > maxPercent {
			maxPercent = p
			closest = "user_count"
		}
	}
	if thresholds.MRRCents > 0 {
		p := float64(current.MRRCents) / float64(thresholds.MRRCents) * 100
		if p > maxPercent {
			maxPercent = p
			closest = "mrr"
		}
	}
	if thresholds.APICalls > 0 {
		p := float64(current.APICalls) / float64(thresholds.APICalls) * 100
		if p > maxPercent {
			maxPercent = p
			closest = "api_calls"
		}
	}
	if thresholds.DaysElapsed > 0 {
		p := float64(current.DaysElapsed) / float64(thresholds.DaysElapsed) * 100
		if p > maxPercent {
			maxPercent = p
			closest = "time"
		}
	}

	return closest, maxPercent
}

// getThresholdValue returns the value of the specified threshold
func (d *DeferredBillingChecker) getThresholdValue(thresholds TriggerThresholds, thresholdType string) int {
	switch thresholdType {
	case "user_count":
		return thresholds.UserCount
	case "mrr":
		return thresholds.MRRCents
	case "api_calls":
		return thresholds.APICalls
	case "time":
		return thresholds.DaysElapsed
	default:
		return 0
	}
}

// TenantMetrics represents the current metrics for a tenant
type TenantMetrics struct {
	UserCount int
	MRRCents  int
	APICalls  int
}

// TriggerThresholds represents the configured trigger thresholds for a bundle
type TriggerThresholds struct {
	UserCount   int // Number of users
	MRRCents    int // Monthly recurring revenue in cents
	APICalls    int // API call volume per month
	DaysElapsed int // Days since signup
}

// CurrentProgress represents the current progress toward thresholds
type CurrentProgress struct {
	UserCount       int
	MRRCents        int
	APICalls        int
	DaysElapsed     int
	ProgressPercent float64
}

// calculateProgress determines how close a tenant is to hitting their thresholds
func calculateProgress(thresholds TriggerThresholds, current CurrentProgress) CurrentProgress {
	// Calculate progress as percentage of closest threshold
	progress := CurrentProgress{
		UserCount:   current.UserCount,
		MRRCents:    current.MRRCents,
		APICalls:    current.APICalls,
		DaysElapsed: current.DaysElapsed,
	}

	// Find which threshold is closest to being hit
	var percentages []float64

	if thresholds.UserCount > 0 {
		p := float64(current.UserCount) / float64(thresholds.UserCount) * 100
		percentages = append(percentages, p)
	}
	if thresholds.MRRCents > 0 {
		p := float64(current.MRRCents) / float64(thresholds.MRRCents) * 100
		percentages = append(percentages, p)
	}
	if thresholds.APICalls > 0 {
		p := float64(current.APICalls) / float64(thresholds.APICalls) * 100
		percentages = append(percentages, p)
	}
	if thresholds.DaysElapsed > 0 {
		p := float64(current.DaysElapsed) / float64(thresholds.DaysElapsed) * 100
		percentages = append(percentages, p)
	}

	// Take the maximum percentage (closest to triggering)
	var maxPercent float64
	for _, p := range percentages {
		if p > maxPercent {
			maxPercent = p
		}
	}
	progress.ProgressPercent = maxPercent

	return progress
}

// shouldTriggerBilling determines if any threshold has been exceeded
func shouldTriggerBilling(thresholds TriggerThresholds, current CurrentProgress) bool {
	return (thresholds.UserCount > 0 && current.UserCount >= thresholds.UserCount) ||
		(thresholds.MRRCents > 0 && current.MRRCents >= thresholds.MRRCents) ||
		(thresholds.APICalls > 0 && current.APICalls >= thresholds.APICalls) ||
		(thresholds.DaysElapsed > 0 && current.DaysElapsed >= thresholds.DaysElapsed)
}

// derefOrDefault returns the dereferenced value or a default if nil
func derefOrDefault(ptr *int, defaultVal int) int {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}
