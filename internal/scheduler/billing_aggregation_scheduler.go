package scheduler

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// BillingAggregationSchedulerConfig holds configuration for the billing aggregation scheduler
type BillingAggregationSchedulerConfig struct {
	// AggregationCron is the cron expression for execution aggregation (default: "0 * * * *" - every hour)
	AggregationCron string

	// RollupCron is the cron expression for rollup aggregation (default: "0 * * * *" - every hour)
	RollupCron string

	// InvoiceGenerationCron is the cron expression for invoice generation (default: "0 2 * * *" - 2 AM daily)
	InvoiceGenerationCron string

	// ForecastCron is the cron expression for usage forecasting (default: "0 6 * * *" - 6 AM daily)
	ForecastCron string

	// AlertCheckCron is the cron expression for alert checking (default: "0 */4 * * *" - every 4 hours)
	AlertCheckCron string

	// Enabled controls whether the scheduler is active
	Enabled bool
}

// DefaultBillingAggregationSchedulerConfig returns default configuration
func DefaultBillingAggregationSchedulerConfig() *BillingAggregationSchedulerConfig {
	return &BillingAggregationSchedulerConfig{
		AggregationCron:       "0 * * * *",     // Every hour
		RollupCron:            "0 * * * *",     // Every hour
		InvoiceGenerationCron: "0 2 * * *",     // 2 AM daily
		ForecastCron:          "0 6 * * *",     // 6 AM daily
		AlertCheckCron:        "0 */4 * * *",   // Every 4 hours
		Enabled:               true,
	}
}

// LoadBillingAggregationSchedulerConfig loads configuration from environment
func LoadBillingAggregationSchedulerConfig() *BillingAggregationSchedulerConfig {
	config := DefaultBillingAggregationSchedulerConfig()

	if v := os.Getenv("BILLING_AGGREGATION_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			config.Enabled = enabled
		}
	}

	if v := os.Getenv("BILLING_AGGREGATION_CRON"); v != "" {
		config.AggregationCron = v
	}

	if v := os.Getenv("BILLING_ROLLUP_CRON"); v != "" {
		config.RollupCron = v
	}

	if v := os.Getenv("BILLING_INVOICE_GENERATION_CRON"); v != "" {
		config.InvoiceGenerationCron = v
	}

	if v := os.Getenv("BILLING_FORECAST_CRON"); v != "" {
		config.ForecastCron = v
	}

	if v := os.Getenv("BILLING_ALERT_CHECK_CRON"); v != "" {
		config.AlertCheckCron = v
	}

	return config
}

// BillingAggregationScheduler manages periodic billing aggregation and invoice generation
type BillingAggregationScheduler struct {
	cron         *cron.Cron
	repo         storage.Repository
	aggregator   *services.RegistryUsageAggregator
	forecaster   *services.UsageForecaster
	alerter      *services.UsageAlerter
	notifySvc    *notification.Service
	logger       *logrus.Logger
	config       *BillingAggregationSchedulerConfig
	stopOnce     sync.Once
	cancel       context.CancelFunc
}

// NewBillingAggregationScheduler creates a new billing aggregation scheduler
func NewBillingAggregationScheduler(
	repo storage.Repository,
	aggregator *services.RegistryUsageAggregator,
	notifySvc *notification.Service,
) *BillingAggregationScheduler {
	return &BillingAggregationScheduler{
		cron:       cron.New(),
		repo:       repo,
		aggregator: aggregator,
		notifySvc:  notifySvc,
		logger:     logrus.New(),
		config:     LoadBillingAggregationSchedulerConfig(),
	}
}

// SetForecaster sets the usage forecaster service
func (s *BillingAggregationScheduler) SetForecaster(forecaster *services.UsageForecaster) {
	s.forecaster = forecaster
}

// SetAlerter sets the usage alerter service
func (s *BillingAggregationScheduler) SetAlerter(alerter *services.UsageAlerter) {
	s.alerter = alerter
}

// Start begins the billing aggregation scheduler
func (s *BillingAggregationScheduler) Start(ctx context.Context) error {
	if !s.config.Enabled {
		s.logger.Info("Billing aggregation scheduler is disabled")
		return nil
	}

	// Validate cron expressions
	if _, err := cron.ParseStandard(s.config.AggregationCron); err != nil {
		return fmt.Errorf("invalid aggregation cron expression: %w", err)
	}
	if _, err := cron.ParseStandard(s.config.RollupCron); err != nil {
		return fmt.Errorf("invalid rollup cron expression: %w", err)
	}
	if _, err := cron.ParseStandard(s.config.InvoiceGenerationCron); err != nil {
		return fmt.Errorf("invalid invoice generation cron expression: %w", err)
	}
	if _, err := cron.ParseStandard(s.config.ForecastCron); err != nil {
		return fmt.Errorf("invalid forecast cron expression: %w", err)
	}
	if _, err := cron.ParseStandard(s.config.AlertCheckCron); err != nil {
		return fmt.Errorf("invalid alert check cron expression: %w", err)
	}

	var ctxWithCancel context.Context
	ctxWithCancel, s.cancel = context.WithCancel(ctx)

	// Add aggregation job
	_, err := s.cron.AddFunc(s.config.AggregationCron, func() {
		s.runAggregation(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add aggregation cron job: %w", err)
	}

	// Add rollup job
	_, err = s.cron.AddFunc(s.config.RollupCron, func() {
		s.runRollup(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add rollup cron job: %w", err)
	}

	// Add invoice generation job
	_, err = s.cron.AddFunc(s.config.InvoiceGenerationCron, func() {
		s.runInvoiceGeneration(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add invoice generation cron job: %w", err)
	}

	// Add forecast generation job
	_, err = s.cron.AddFunc(s.config.ForecastCron, func() {
		s.runForecast(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add forecast cron job: %w", err)
	}

	// Add alert check job
	_, err = s.cron.AddFunc(s.config.AlertCheckCron, func() {
		s.runAlertCheck(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add alert check cron job: %w", err)
	}

	s.cron.Start()

	s.logger.WithFields(logrus.Fields{
		"aggregation_cron":        s.config.AggregationCron,
		"rollup_cron":             s.config.RollupCron,
		"invoice_generation_cron": s.config.InvoiceGenerationCron,
		"forecast_cron":           s.config.ForecastCron,
		"alert_check_cron":        s.config.AlertCheckCron,
	}).Info("Billing aggregation scheduler started")

	return nil
}

// Stop stops the billing aggregation scheduler
func (s *BillingAggregationScheduler) Stop() error {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		<-s.cron.Stop().Done()
		s.logger.Info("Billing aggregation scheduler stopped")
	})
	return nil
}

// runAggregation executes the execution-to-usage-events aggregation
func (s *BillingAggregationScheduler) runAggregation(ctx context.Context) {
	start := time.Now()
	s.logger.Info("Starting scheduled execution aggregation")

	if err := s.aggregator.AggregateExecutionsToUsageEvents(ctx); err != nil {
		s.logger.WithError(err).Error("Execution aggregation failed")

		// Send alert if notification service is available
		if s.notifySvc != nil {
			s.sendAggregationAlert(ctx, "execution", err)
		}
		return
	}

	duration := time.Since(start)
	s.logger.WithField("duration_ms", duration.Milliseconds()).Info("Execution aggregation completed")
}

// runRollup executes the usage-events-to-rollups aggregation
func (s *BillingAggregationScheduler) runRollup(ctx context.Context) {
	start := time.Now()
	s.logger.Info("Starting scheduled usage rollup")

	if err := s.aggregator.AggregateUsageEventsToRollups(ctx); err != nil {
		s.logger.WithError(err).Error("Usage rollup failed")

		if s.notifySvc != nil {
			s.sendAggregationAlert(ctx, "rollup", err)
		}
		return
	}

	duration := time.Since(start)
	s.logger.WithField("duration_ms", duration.Milliseconds()).Info("Usage rollup completed")
}

// runInvoiceGeneration executes the invoice generation
func (s *BillingAggregationScheduler) runInvoiceGeneration(ctx context.Context) {
	start := time.Now()
	s.logger.Info("Starting scheduled invoice generation")

	if err := s.aggregator.GenerateDraftInvoices(ctx); err != nil {
		s.logger.WithError(err).Error("Invoice generation failed")

		if s.notifySvc != nil {
			s.sendAggregationAlert(ctx, "invoice", err)
		}
		return
	}

	duration := time.Since(start)
	s.logger.WithField("duration_ms", duration.Milliseconds()).Info("Invoice generation completed")
}

// runForecast executes the usage forecasting
func (s *BillingAggregationScheduler) runForecast(ctx context.Context) {
	if s.forecaster == nil {
		s.logger.Debug("Forecaster not configured, skipping forecast generation")
		return
	}

	start := time.Now()
	s.logger.Info("Starting scheduled forecast generation")

	if err := s.forecaster.GenerateAllForecasts(ctx); err != nil {
		s.logger.WithError(err).Error("Forecast generation failed")

		if s.notifySvc != nil {
			s.sendAggregationAlert(ctx, "forecast", err)
		}
		return
	}

	duration := time.Since(start)
	s.logger.WithField("duration_ms", duration.Milliseconds()).Info("Forecast generation completed")
}

// runAlertCheck executes the alert checking
func (s *BillingAggregationScheduler) runAlertCheck(ctx context.Context) {
	if s.alerter == nil {
		s.logger.Debug("Alerter not configured, skipping alert check")
		return
	}

	start := time.Now()
	s.logger.Info("Starting scheduled alert check")

	if err := s.alerter.CheckAllAlerts(ctx); err != nil {
		s.logger.WithError(err).Error("Alert check failed")

		if s.notifySvc != nil {
			s.sendAggregationAlert(ctx, "alert_check", err)
		}
		return
	}

	duration := time.Since(start)
	s.logger.WithField("duration_ms", duration.Milliseconds()).Info("Alert check completed")
}

// sendAggregationAlert sends a notification when aggregation fails
func (s *BillingAggregationScheduler) sendAggregationAlert(ctx context.Context, operation string, err error) {
	if s.notifySvc == nil {
		return
	}

	_, notifyErr := s.notifySvc.Send(ctx, notification.SendRequest{
		Type:     notification.TypeBillingAlert,
		Category: notification.CategoryBilling,
		Title:    fmt.Sprintf("Billing %s Failed", operation),
		Body:     fmt.Sprintf("The billing %s operation failed: %v", operation, err),
		Data: map[string]interface{}{
			"operation":   operation,
			"error":       err.Error(),
			"timestamp":   time.Now().Format(time.RFC3339),
		},
		Channels: []string{notification.ChannelInApp},
		Priority: notification.PriorityHigh,
	})

	if notifyErr != nil {
		s.logger.WithError(notifyErr).Error("Failed to send aggregation alert")
	}
}

// GetSchedule returns the current schedule configuration
func (s *BillingAggregationScheduler) GetSchedule() map[string]interface{} {
	nextAggregation := "unknown"
	nextRollup := "unknown"
	nextInvoice := "unknown"
	nextForecast := "unknown"
	nextAlertCheck := "unknown"

	if s.cron != nil {
		entries := s.cron.Entries()
		if len(entries) > 0 {
			nextAggregation = entries[0].Next.Format(time.RFC3339)
		}
		if len(entries) > 1 {
			nextRollup = entries[1].Next.Format(time.RFC3339)
		}
		if len(entries) > 2 {
			nextInvoice = entries[2].Next.Format(time.RFC3339)
		}
		if len(entries) > 3 {
			nextForecast = entries[3].Next.Format(time.RFC3339)
		}
		if len(entries) > 4 {
			nextAlertCheck = entries[4].Next.Format(time.RFC3339)
		}
	}

	return map[string]interface{}{
		"enabled":                 s.config.Enabled,
		"aggregation_cron":        s.config.AggregationCron,
		"next_aggregation_run":    nextAggregation,
		"rollup_cron":             s.config.RollupCron,
		"next_rollup_run":         nextRollup,
		"invoice_generation_cron": s.config.InvoiceGenerationCron,
		"next_invoice_generation": nextInvoice,
		"forecast_cron":           s.config.ForecastCron,
		"next_forecast_run":       nextForecast,
		"alert_check_cron":        s.config.AlertCheckCron,
		"next_alert_check":        nextAlertCheck,
		"forecaster_enabled":      s.forecaster != nil,
		"alerter_enabled":         s.alerter != nil,
	}
}

// RunAggregationNow triggers an immediate aggregation run (for manual/admin use)
func (s *BillingAggregationScheduler) RunAggregationNow(ctx context.Context) error {
	s.logger.Info("Manually triggering execution aggregation")
	return s.aggregator.AggregateExecutionsToUsageEvents(ctx)
}

// RunRollupNow triggers an immediate rollup run (for manual/admin use)
func (s *BillingAggregationScheduler) RunRollupNow(ctx context.Context) error {
	s.logger.Info("Manually triggering usage rollup")
	return s.aggregator.AggregateUsageEventsToRollups(ctx)
}

// RunInvoiceGenerationNow triggers immediate invoice generation (for manual/admin use)
func (s *BillingAggregationScheduler) RunInvoiceGenerationNow(ctx context.Context) error {
	s.logger.Info("Manually triggering invoice generation")
	return s.aggregator.GenerateDraftInvoices(ctx)
}

// RunForecastNow triggers immediate forecast generation (for manual/admin use)
func (s *BillingAggregationScheduler) RunForecastNow(ctx context.Context) error {
	if s.forecaster == nil {
		return fmt.Errorf("forecaster not configured")
	}
	s.logger.Info("Manually triggering forecast generation")
	return s.forecaster.GenerateAllForecasts(ctx)
}

// RunAlertCheckNow triggers immediate alert check (for manual/admin use)
func (s *BillingAggregationScheduler) RunAlertCheckNow(ctx context.Context) error {
	if s.alerter == nil {
		return fmt.Errorf("alerter not configured")
	}
	s.logger.Info("Manually triggering alert check")
	return s.alerter.CheckAllAlerts(ctx)
}
