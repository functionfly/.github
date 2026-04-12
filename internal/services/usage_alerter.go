package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// UsageAlerterConfig holds configuration for the alerting service
type UsageAlerterConfig struct {
	Enabled            bool
	CheckInterval      time.Duration
	CooldownMultiplier time.Duration
}

// DefaultUsageAlerterConfig returns default configuration
func DefaultUsageAlerterConfig() *UsageAlerterConfig {
	return &UsageAlerterConfig{
		Enabled:            true,
		CheckInterval:      15 * time.Minute,
		CooldownMultiplier: time.Minute,
	}
}

// LoadUsageAlerterConfig loads configuration from environment
func LoadUsageAlerterConfig() *UsageAlerterConfig {
	config := DefaultUsageAlerterConfig()

	if v := os.Getenv("USAGE_ALERTING_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			config.Enabled = enabled
		}
	}

	if v := os.Getenv("USAGE_ALERT_CHECK_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			config.CheckInterval = d
		}
	}

	return config
}

// UsageAlerter provides proactive alerts for usage and spend thresholds
type UsageAlerter struct {
	alertRepo         *storage.UsageAlertRepository
	billingRepo       storage.Repository
	notificationSvc   *notification.Service
	forecaster        *UsageForecaster
	config            *UsageAlerterConfig
	logger            *logrus.Logger
	stopChan          chan struct{}
	stopOnce          sync.Once
	lastAlertSent     map[string]time.Time // Key: "tenantID:alertType"
	lastAlertMu       sync.RWMutex
}

// NewUsageAlerter creates a new usage alerter
func NewUsageAlerter(alertRepo *storage.UsageAlertRepository, billingRepo storage.Repository, notificationSvc *notification.Service, forecaster *UsageForecaster, config *UsageAlerterConfig) *UsageAlerter {
	return &UsageAlerter{
		alertRepo:       alertRepo,
		billingRepo:     billingRepo,
		notificationSvc: notificationSvc,
		forecaster:      forecaster,
		config:          config,
		logger:          logrus.New(),
		stopChan:        make(chan struct{}),
		lastAlertSent:   make(map[string]time.Time),
	}
}

// Start begins the alerting service
func (a *UsageAlerter) Start(ctx context.Context) {
	if !a.config.Enabled {
		a.logger.Info("Usage alerting service is disabled")
		return
	}

	a.logger.WithFields(logrus.Fields{
		"check_interval": a.config.CheckInterval,
	}).Info("Starting usage alerting service")

	// Run initial check
	go a.runInitialCheck(ctx)

	// Start check loop
	go a.runCheckLoop(ctx)
}

// Stop stops the alerting service
func (a *UsageAlerter) Stop() {
	a.stopOnce.Do(func() {
		close(a.stopChan)
	})
}

// runInitialCheck runs the initial alert check on startup
func (a *UsageAlerter) runInitialCheck(ctx context.Context) {
	if err := a.CheckAllAlerts(ctx); err != nil {
		a.logger.WithError(err).Error("Initial alert check failed")
	}
}

// runCheckLoop runs the periodic alert checking loop
func (a *UsageAlerter) runCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(a.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Alert check loop stopping due to context cancellation")
			return
		case <-a.stopChan:
			a.logger.Info("Alert check loop stopped")
			return
		case <-ticker.C:
			if err := a.CheckAllAlerts(ctx); err != nil {
				a.logger.WithError(err).Error("Alert check failed")
			}
		}
	}
}

// CheckAllAlerts checks all configured alerts for all tenants
func (a *UsageAlerter) CheckAllAlerts(ctx context.Context) error {
	start := time.Now()

	// Get all active subscriptions
	subs, err := a.billingRepo.ListAllSubscriptions(1000, 0)
	if err != nil {
		return fmt.Errorf("failed to list subscriptions: %w", err)
	}

	totalAlerts := 0
	triggeredAlerts := 0

	for _, sub := range subs {
		if sub.Status != "active" && sub.Status != "trialing" {
			continue
		}

		// Check configured alerts for tenant
		alerts, err := a.alertRepo.ListUsageAlertsByTenant(ctx, sub.TenantID)
		if err != nil {
			a.logger.WithError(err).WithField("tenant_id", sub.TenantID).Warn("Failed to list alerts")
			continue
		}

		for _, alert := range alerts {
			totalAlerts++
			if !alert.IsEnabled {
				continue
			}

			// Check cooldown
			if a.isInCooldown(sub.TenantID, alert) {
				continue
			}

			// Check if alert should trigger
			shouldTrigger, triggeredValue, message, err := a.evaluateAlert(ctx, alert, sub)
			if err != nil {
				a.logger.WithError(err).WithFields(logrus.Fields{
					"alert_id": alert.ID,
					"tenant_id": sub.TenantID,
				}).Warn("Failed to evaluate alert")
				continue
			}

			if shouldTrigger {
				if err := a.triggerAlert(ctx, alert, sub, triggeredValue, message); err != nil {
					a.logger.WithError(err).WithFields(logrus.Fields{
						"alert_id": alert.ID,
						"tenant_id": sub.TenantID,
					}).Error("Failed to trigger alert")
				} else {
					triggeredAlerts++
					a.setCooldown(sub.TenantID, alert)
				}
			}
		}

		// Check spend cap alerts
		if err := a.checkSpendCapAlerts(ctx, sub); err != nil {
			a.logger.WithError(err).WithField("tenant_id", sub.TenantID).Warn("Failed to check spend cap alerts")
		}
	}

	a.logger.WithFields(logrus.Fields{
		"duration":         time.Since(start),
		"total_alerts":     totalAlerts,
		"triggered_alerts": triggeredAlerts,
		"tenants_checked":  len(subs),
	}).Info("Alert check completed")

	return nil
}

// evaluateAlert evaluates if an alert should trigger
func (a *UsageAlerter) evaluateAlert(ctx context.Context, alert *storage.UsageAlert, sub *storage.Subscription) (bool, float64, string, error) {
	// Determine the period to check
	periodStart, periodEnd := a.getAlertPeriod(alert, sub)

	// Get current value based on alert type
	var currentValue float64
	var err error

	switch alert.AlertType {
	case "spend_cap":
		currentValue, err = a.getCurrentSpend(ctx, sub.TenantID, periodStart, periodEnd)
	case "usage_spike":
		currentValue, err = a.getCurrentUsage(ctx, sub.TenantID, "function_execution", periodStart, periodEnd)
	case "threshold":
		currentValue, err = a.getCurrentUsage(ctx, sub.TenantID, "function_execution", periodStart, periodEnd)
	case "forecast_exceeded":
		return a.evaluateForecastAlert(ctx, alert, sub)
	default:
		return false, 0, "", fmt.Errorf("unknown alert type: %s", alert.AlertType)
	}

	if err != nil {
		return false, 0, "", err
	}

	// Evaluate based on operator
	triggered := false
	message := ""

	switch alert.ThresholdOperator {
	case "gte":
		if currentValue >= alert.ThresholdValue {
			triggered = true
			message = fmt.Sprintf("Usage has reached %.0f, exceeding threshold of %.0f",
				currentValue, alert.ThresholdValue)
		}
	case "lte":
		if currentValue <= alert.ThresholdValue {
			triggered = true
			message = fmt.Sprintf("Usage has dropped to %.0f, below threshold of %.0f",
				currentValue, alert.ThresholdValue)
		}
	case "percentage_of_cap":
		// Get spend cap to calculate percentage
		cap, err := a.alertRepo.GetSpendCapByTenant(ctx, sub.TenantID, sub.CurrentPeriodStart)
		if err != nil || cap == nil || !cap.IsEnabled {
			return false, 0, "", nil
		}
		percentage := (currentValue / float64(cap.CapAmountCents)) * 100
		if percentage >= alert.ThresholdValue {
			triggered = true
			message = fmt.Sprintf("Spend is at %.1f%% of your $%.2f cap (%.0f%% threshold)",
				percentage, float64(cap.CapAmountCents)/100, alert.ThresholdValue)
		}
	}

	return triggered, currentValue, message, nil
}

// evaluateForecastAlert checks if forecast predicts exceeding threshold
func (a *UsageAlerter) evaluateForecastAlert(ctx context.Context, alert *storage.UsageAlert, sub *storage.Subscription) (bool, float64, string, error) {
	// Get the latest forecast
	forecast, err := a.alertRepo.GetLatestForecast(ctx, sub.TenantID, "spend")
	if err != nil || forecast == nil {
		return false, 0, "", nil
	}

	// Check if predicted value exceeds threshold
	if forecast.PredictedValue >= alert.ThresholdValue {
		return true, forecast.PredictedValue,
			fmt.Sprintf("Forecast predicts spend of $%.2f by period end, exceeding threshold of $%.2f",
				forecast.PredictedValue/100, alert.ThresholdValue/100),
			nil
	}

	return false, 0, "", nil
}

// checkSpendCapAlerts checks default spend cap thresholds
func (a *UsageAlerter) checkSpendCapAlerts(ctx context.Context, sub *storage.Subscription) error {
	// Get spend cap for tenant
	cap, err := a.alertRepo.GetSpendCapByTenant(ctx, sub.TenantID, sub.CurrentPeriodStart)
	if err != nil || cap == nil || !cap.IsEnabled {
		return nil
	}

	// Update current spend
	currentSpend, err := a.getCurrentSpend(ctx, sub.TenantID, sub.CurrentPeriodStart, sub.CurrentPeriodEnd)
	if err != nil {
		return err
	}

	// Update spend in cap record
	if err := a.alertRepo.UpdateCurrentSpend(ctx, cap.ID, int(currentSpend)); err != nil {
		a.logger.WithError(err).Warn("Failed to update current spend")
	}

	// Check warning thresholds
	if len(cap.WarningThresholds) > 0 {
		percentageOfCap := (currentSpend / float64(cap.CapAmountCents)) * 100

		for _, threshold := range cap.WarningThresholds {
			// Check if this threshold has been crossed
			if percentageOfCap >= float64(threshold) {
				alertKey := fmt.Sprintf("%s:cap_threshold_%d", sub.TenantID.String(), threshold)

				a.lastAlertMu.RLock()
				lastSent, exists := a.lastAlertSent[alertKey]
				a.lastAlertMu.RUnlock()

				// Only send once per day per threshold
				if !exists || time.Since(lastSent) > 24*time.Hour {
					if err := a.sendSpendCapAlert(ctx, sub, cap, currentSpend, threshold); err != nil {
						a.logger.WithError(err).Warn("Failed to send spend cap alert")
					} else {
						a.lastAlertMu.Lock()
						a.lastAlertSent[alertKey] = time.Now()
						a.lastAlertMu.Unlock()
					}
				}
			}
		}
	}

	return nil
}

// triggerAlert sends an alert notification
func (a *UsageAlerter) triggerAlert(ctx context.Context, alert *storage.UsageAlert, sub *storage.Subscription, triggeredValue float64, message string) error {
	// Get tenant users to notify
	users, err := a.billingRepo.ListActiveUsersByTenant(ctx, sub.TenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant users: %w", err)
	}

	// Record the alert trigger
	history := &storage.UsageAlertHistory{
		AlertID:        alert.ID,
		TenantID:       sub.TenantID,
		TriggeredValue: triggeredValue,
		ThresholdValue: alert.ThresholdValue,
		Message:        message,
		Metadata: map[string]interface{}{
			"alert_name":   alert.Name,
			"alert_type":   alert.AlertType,
			"period_start": sub.CurrentPeriodStart,
			"period_end":   sub.CurrentPeriodEnd,
		},
	}

	if err := a.alertRepo.RecordAlertTrigger(ctx, history); err != nil {
		a.logger.WithError(err).Warn("Failed to record alert trigger")
	}

	// Send notifications to all tenant users
	for _, user := range users {
		notificationType := a.getNotificationTypeForAlert(alert.AlertType)
		priority := a.getPriorityForAlert(alert.AlertType)

		_, err := a.notificationSvc.Send(ctx, notification.SendRequest{
			UserID:   user.ID,
			Type:     notificationType,
			Category: notification.CategoryBilling,
			Title:    alert.Name,
			Body:     message,
			Data: notification.JSONMap{
				"alert_id":         alert.ID.String(),
				"alert_type":       alert.AlertType,
				"threshold_value":  alert.ThresholdValue,
				"triggered_value":  triggeredValue,
				"tenant_id":        sub.TenantID.String(),
				"period_start":     sub.CurrentPeriodStart,
				"period_end":       sub.CurrentPeriodEnd,
			},
			Channels: alert.NotificationChannels,
			Priority: priority,
		})

		if err != nil {
			a.logger.WithError(err).WithField("user_id", user.ID).Warn("Failed to send alert notification")
		}
	}

	a.logger.WithFields(logrus.Fields{
		"alert_id":         alert.ID,
		"tenant_id":        sub.TenantID,
		"alert_type":       alert.AlertType,
		"triggered_value":  triggeredValue,
		"users_notified":   len(users),
	}).Info("Alert triggered and notifications sent")

	return nil
}

// sendSpendCapAlert sends a spend cap alert notification
func (a *UsageAlerter) sendSpendCapAlert(ctx context.Context, sub *storage.Subscription, cap *storage.SpendCap, currentSpend float64, threshold int) error {
	users, err := a.billingRepo.ListActiveUsersByTenant(ctx, sub.TenantID)
	if err != nil {
		return err
	}

	percentage := (currentSpend / float64(cap.CapAmountCents)) * 100
	message := fmt.Sprintf("You have used %.1f%% ($%.2f of $%.2f) of your monthly spend cap.",
		percentage, currentSpend/100, float64(cap.CapAmountCents)/100)

	if cap.IsHardCap && percentage >= 100 {
		message = fmt.Sprintf("CRITICAL: You have exceeded your $%.2f spend cap. Current spend: $%.2f. "+
			"Action may be taken: %s", float64(cap.CapAmountCents)/100, currentSpend/100, cap.ActionOnCap)
	}

	for _, user := range users {
		priority := notification.PriorityNormal
		if threshold >= 90 || (cap.IsHardCap && percentage >= 100) {
			priority = notification.PriorityHigh
		}

		_, err := a.notificationSvc.Send(ctx, notification.SendRequest{
			UserID:   user.ID,
			Type:     notification.TypeBillingAlert,
			Category: notification.CategoryBilling,
			Title:    fmt.Sprintf("Spend Cap Alert: %d%% Threshold", threshold),
			Body:     message,
			Data: notification.JSONMap{
				"spend_cap_id":     cap.ID.String(),
				"cap_amount":       cap.CapAmountCents,
				"current_spend":    currentSpend,
				"threshold":        threshold,
				"percentage_used":  percentage,
				"action_on_cap":    cap.ActionOnCap,
				"is_hard_cap":      cap.IsHardCap,
				"tenant_id":        sub.TenantID.String(),
			},
			Channels: []string{notification.ChannelEmail, notification.ChannelInApp},
			Priority: priority,
		})

		if err != nil {
			a.logger.WithError(err).WithField("user_id", user.ID).Warn("Failed to send spend cap alert")
		}
	}

	return nil
}

// isInCooldown checks if an alert is in cooldown period
func (a *UsageAlerter) isInCooldown(tenantID uuid.UUID, alert *storage.UsageAlert) bool {
	if alert.LastTriggeredAt == nil {
		return false
	}

	cooldownDuration := time.Duration(alert.CooldownMinutes) * a.config.CooldownMultiplier
	return time.Since(*alert.LastTriggeredAt) < cooldownDuration
}

// setCooldown marks an alert as having been sent recently
func (a *UsageAlerter) setCooldown(tenantID uuid.UUID, alert *storage.UsageAlert) {
	key := fmt.Sprintf("%s:%s", tenantID.String(), alert.ID.String())
	a.lastAlertMu.Lock()
	a.lastAlertSent[key] = time.Now()
	a.lastAlertMu.Unlock()
}

// getAlertPeriod returns the start and end dates for an alert's period type
func (a *UsageAlerter) getAlertPeriod(alert *storage.UsageAlert, sub *storage.Subscription) (time.Time, time.Time) {
	switch alert.PeriodType {
	case "billing_period":
		return sub.CurrentPeriodStart, sub.CurrentPeriodEnd
	case "daily":
		now := time.Now().UTC()
		return now.Truncate(24 * time.Hour), now.Truncate(24*time.Hour).Add(24*time.Hour)
	case "weekly":
		now := time.Now().UTC()
		start := now.AddDate(0, 0, -7)
		return start, now
	default:
		return sub.CurrentPeriodStart, sub.CurrentPeriodEnd
	}
}

// getCurrentSpend retrieves current spend for a tenant in a period
func (a *UsageAlerter) getCurrentSpend(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (float64, error) {
	// Get invoices for the period
	invoices, err := a.billingRepo.ListInvoicesByTenant(tenantID, 100, 0)
	if err != nil {
		return 0, err
	}

	totalCents := 0
	for _, inv := range invoices {
		if inv.Status != "void" && inv.PeriodStart != nil && inv.PeriodEnd != nil {
			if !inv.PeriodEnd.Before(start) && !inv.PeriodStart.After(end) {
				totalCents += inv.AmountDueCents
			}
		}
	}

	return float64(totalCents), nil
}

// getCurrentUsage retrieves current usage for a tenant
func (a *UsageAlerter) getCurrentUsage(ctx context.Context, tenantID uuid.UUID, eventType string, start, end time.Time) (float64, error) {
	summary, err := a.alertRepo.GetCurrentPeriodUsage(ctx, tenantID, start, end)
	if err != nil {
		return 0, err
	}

	switch eventType {
	case "function_execution":
		return float64(summary.TotalExecutions), nil
	case "compute_time_ms":
		return float64(summary.TotalComputeMs), nil
	default:
		return 0, nil
	}
}

// getNotificationTypeForAlert returns the appropriate notification type
func (a *UsageAlerter) getNotificationTypeForAlert(alertType string) string {
	switch alertType {
	case "spend_cap", "threshold":
		return notification.TypeBillingAlert
	case "usage_spike":
		return notification.TypeFunctionExecuted
	case "forecast_exceeded":
		return notification.TypeBillingAlert
	default:
		return notification.TypeBillingAlert
	}
}

// getPriorityForAlert returns the appropriate priority level
func (a *UsageAlerter) getPriorityForAlert(alertType string) string {
	switch alertType {
	case "spend_cap":
		return notification.PriorityHigh
	case "usage_spike":
		return notification.PriorityNormal
	case "threshold":
		return notification.PriorityNormal
	case "forecast_exceeded":
		return notification.PriorityHigh
	default:
		return notification.PriorityNormal
	}
}
