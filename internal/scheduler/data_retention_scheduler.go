package scheduler

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// DataRetentionSchedulerConfig holds configuration for the data retention scheduler
type DataRetentionSchedulerConfig struct {
	// DetailedCleanupCron is the cron for detailed execution log cleanup (default: "0 3 * * *" - 3 AM daily)
	// These are high-volume entries older than 90 days
	DetailedCleanupCron string

	// FinancialDataRetentionYears is how long to keep financial aggregates (default: 7 years for SOX)
	FinancialDataRetentionYears int

	// DetailedDataRetentionDays is how long to keep detailed execution logs (default: 90 days)
	DetailedDataRetentionDays int

	// Enabled controls whether the scheduler is active
	Enabled bool

	// DryRunMode if true, logs what would be deleted without actually deleting
	DryRunMode bool

	// SkipIfLegalHold if true, skips deletion if any legal holds are active (default: true)
	SkipIfLegalHold bool
}

// DefaultDataRetentionSchedulerConfig returns default configuration
func DefaultDataRetentionSchedulerConfig() *DataRetentionSchedulerConfig {
	return &DataRetentionSchedulerConfig{
		DetailedCleanupCron:         "0 3 * * *", // 3 AM daily (after invoice generation at 2 AM)
		FinancialDataRetentionYears: 7,           // 7 years for SOX compliance
		DetailedDataRetentionDays:   90,          // 90 days for execution logs
		Enabled:                     true,
		DryRunMode:                  false,
		SkipIfLegalHold:             true,
	}
}

// LoadDataRetentionSchedulerConfig loads configuration from environment
func LoadDataRetentionSchedulerConfig() *DataRetentionSchedulerConfig {
	config := DefaultDataRetentionSchedulerConfig()

	if v := os.Getenv("DATA_RETENTION_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			config.Enabled = enabled
		}
	}

	if v := os.Getenv("DATA_RETENTION_CRON"); v != "" {
		config.DetailedCleanupCron = v
	}

	if v := os.Getenv("DATA_RETENTION_FINANCIAL_YEARS"); v != "" {
		if years, err := strconv.Atoi(v); err == nil && years > 0 {
			config.FinancialDataRetentionYears = years
		}
	}

	if v := os.Getenv("DATA_RETENTION_DETAILED_DAYS"); v != "" {
		if days, err := strconv.Atoi(v); err == nil && days > 0 {
			config.DetailedDataRetentionDays = days
		}
	}

	if v := os.Getenv("DATA_RETENTION_DRY_RUN"); v != "" {
		if dryRun, err := strconv.ParseBool(v); err == nil {
			config.DryRunMode = dryRun
		}
	}

	if v := os.Getenv("DATA_RETENTION_SKIP_IF_LEGAL_HOLD"); v != "" {
		if skip, err := strconv.ParseBool(v); err == nil {
			config.SkipIfLegalHold = skip
		}
	}

	return config
}

// DataRetentionScheduler manages periodic data retention cleanup
type DataRetentionScheduler struct {
	cron        *cron.Cron
	billingRepo *storage.BillingRepository
	notifySvc   *notification.Service
	logger      *logrus.Logger
	config      *DataRetentionSchedulerConfig
	stopOnce    sync.Once
	cancel      context.CancelFunc
}

// NewDataRetentionScheduler creates a new data retention scheduler
func NewDataRetentionScheduler(
	billingRepo *storage.BillingRepository,
	notifySvc *notification.Service,
) *DataRetentionScheduler {
	return &DataRetentionScheduler{
		cron:        cron.New(),
		billingRepo: billingRepo,
		notifySvc:   notifySvc,
		logger:      logrus.New(),
		config:      LoadDataRetentionSchedulerConfig(),
	}
}

// Start begins the data retention scheduler
func (s *DataRetentionScheduler) Start(ctx context.Context) error {
	if !s.config.Enabled {
		s.logger.Info("Data retention scheduler is disabled")
		return nil
	}

	// Validate cron expression
	if _, err := cron.ParseStandard(s.config.DetailedCleanupCron); err != nil {
		return fmt.Errorf("invalid data retention cron expression: %w", err)
	}

	var ctxWithCancel context.Context
	ctxWithCancel, s.cancel = context.WithCancel(ctx)

	// Add cleanup job
	_, err := s.cron.AddFunc(s.config.DetailedCleanupCron, func() {
		s.runRetentionCleanup(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add retention cleanup cron job: %w", err)
	}

	s.cron.Start()

	s.logger.WithFields(logrus.Fields{
		"cleanup_cron":        s.config.DetailedCleanupCron,
		"financial_retention": fmt.Sprintf("%d years", s.config.FinancialDataRetentionYears),
		"detailed_retention":  fmt.Sprintf("%d days", s.config.DetailedDataRetentionDays),
		"dry_run_mode":        s.config.DryRunMode,
		"skip_if_legal_hold":  s.config.SkipIfLegalHold,
	}).Info("Data retention scheduler started")

	return nil
}

// Stop stops the data retention scheduler
func (s *DataRetentionScheduler) Stop() error {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		<-s.cron.Stop().Done()
		s.logger.Info("Data retention scheduler stopped")
	})
	return nil
}

// runRetentionCleanup executes the retention cleanup
func (s *DataRetentionScheduler) runRetentionCleanup(ctx context.Context) {
	start := time.Now()
	s.logger.Info("Starting scheduled data retention cleanup")

	// Check for legal holds if configured
	if s.config.SkipIfLegalHold {
		hasHolds, err := s.checkLegalHolds(ctx)
		if err != nil {
			s.logger.WithError(err).Error("Failed to check legal holds")
			s.sendRetentionAlert(ctx, "legal_hold_check", err)
			return
		}
		if hasHolds {
			s.logger.Warn("Active legal holds detected, skipping retention cleanup")
			s.sendRetentionAlert(ctx, "legal_hold_skip", fmt.Errorf("cleanup skipped due to active legal holds"))
			return
		}
	}

	// Get summary before cleanup (for reporting)
	summary, err := s.billingRepo.GetCostAllocationRetentionSummary(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get retention summary")
	} else {
		s.logger.WithFields(logrus.Fields{
			"total_entries":               summary.TotalEntryCount,
			"detailed_eligible_deletion":  summary.DetailedEntriesEligibleForDeletion,
			"detailed_financial_value":    summary.DetailedFinancialValueCents,
			"financial_eligible_deletion": summary.FinancialEntriesEligibleForDeletion,
			"financial_data_value":        summary.FinancialDataValueCents,
			"oldest_entry":                summary.OldestEntryDate,
			"newest_entry":                summary.NewestEntryDate,
		}).Info("Retention cleanup summary before run")
	}

	if s.config.DryRunMode {
		s.logger.Info("DRY RUN MODE: Would delete detailed entries older than 90 days")
		if summary != nil {
			s.logger.WithField("would_delete_count", summary.DetailedEntriesEligibleForDeletion).
				Info("DRY RUN: No actual deletions performed")
		}
		return
	}

	// Step 1: Cleanup detailed execution logs (90 days)
	results, err := s.billingRepo.CleanupCostAllocationByRetention(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Detailed execution log cleanup failed")
		s.sendRetentionAlert(ctx, "detailed_cleanup", err)
		return
	}

	detailedDeleted := results["detailed_execution_logs_deleted"]
	duration := time.Since(start)

	s.logger.WithFields(logrus.Fields{
		"detailed_deleted": detailedDeleted,
		"duration_ms":      duration.Milliseconds(),
	}).Info("Detailed execution log cleanup completed")

	// Send success notification if deletions occurred
	if detailedDeleted > 0 && s.notifySvc != nil {
		s.sendRetentionSuccessNotification(ctx, detailedDeleted, duration)
	}
}

// checkLegalHolds queries if any active legal holds exist that should block cleanup
func (s *DataRetentionScheduler) checkLegalHolds(ctx context.Context) (bool, error) {
	return s.billingRepo.HasActiveLegalHolds(ctx)
}

// sendRetentionAlert sends a notification when retention cleanup fails
func (s *DataRetentionScheduler) sendRetentionAlert(ctx context.Context, operation string, err error) {
	if s.notifySvc == nil {
		return
	}

	_, notifyErr := s.notifySvc.Send(ctx, notification.SendRequest{
		Type:     notification.TypeBillingAlert,
		Category: notification.CategorySystem,
		Title:    fmt.Sprintf("Data Retention %s Failed", operation),
		Body:     fmt.Sprintf("The data retention operation '%s' failed: %v", operation, err),
		Data: map[string]interface{}{
			"operation": operation,
			"error":     err.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
			"dry_run":   s.config.DryRunMode,
		},
		Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		Priority: notification.PriorityHigh,
	})

	if notifyErr != nil {
		s.logger.WithError(notifyErr).Error("Failed to send retention alert")
	}
}

// sendRetentionSuccessNotification sends a notification about successful cleanup
func (s *DataRetentionScheduler) sendRetentionSuccessNotification(ctx context.Context, deletedCount int64, duration time.Duration) {
	if s.notifySvc == nil {
		return
	}

	_, notifyErr := s.notifySvc.Send(ctx, notification.SendRequest{
		Type:     notification.TypeSystemAnnouncement,
		Category: notification.CategorySystem,
		Title:    "Data Retention Cleanup Completed",
		Body:     fmt.Sprintf("Cleaned up %d old cost allocation entries (execution logs > 90 days) in %v", deletedCount, duration),
		Data: map[string]interface{}{
			"deleted_count": deletedCount,
			"duration_ms":   duration.Milliseconds(),
			"timestamp":     time.Now().Format(time.RFC3339),
		},
		Channels: []string{notification.ChannelInApp},
		Priority: notification.PriorityNormal,
	})

	if notifyErr != nil {
		s.logger.WithError(notifyErr).Error("Failed to send retention success notification")
	}
}

// GetSchedule returns the current schedule configuration
func (s *DataRetentionScheduler) GetSchedule() map[string]interface{} {
	nextRun := "unknown"

	if s.cron != nil {
		entries := s.cron.Entries()
		if len(entries) > 0 {
			nextRun = entries[0].Next.Format(time.RFC3339)
		}
	}

	return map[string]interface{}{
		"enabled":                   s.config.Enabled,
		"cleanup_cron":              s.config.DetailedCleanupCron,
		"next_run":                  nextRun,
		"financial_retention_years": s.config.FinancialDataRetentionYears,
		"detailed_retention_days":   s.config.DetailedDataRetentionDays,
		"dry_run_mode":              s.config.DryRunMode,
		"skip_if_legal_hold":        s.config.SkipIfLegalHold,
	}
}

// RunCleanupNow triggers an immediate retention cleanup (for manual/admin use)
func (s *DataRetentionScheduler) RunCleanupNow(ctx context.Context) error {
	s.logger.Info("Manually triggering data retention cleanup")
	s.runRetentionCleanup(ctx)
	return nil
}

// RunWithFinancialPurge triggers cleanup including financial data beyond 7 years
// WARNING: Only use this after legal review and confirmation no disputes are pending
func (s *DataRetentionScheduler) RunWithFinancialPurge(ctx context.Context) (int64, error) {
	s.logger.Warn("Manual financial data purge requested - this is typically only done after 7 years")

	retentionPeriod := time.Duration(s.config.FinancialDataRetentionYears) * 365 * 24 * time.Hour
	deleted, err := s.billingRepo.CleanupFinancialAggregatesAfterRetention(ctx, retentionPeriod)
	if err != nil {
		s.logger.WithError(err).Error("Financial data purge failed")
		return 0, err
	}

	s.logger.WithField("financial_records_deleted", deleted).Info("Financial data purge completed")
	return deleted, nil
}

// GetRetentionSummary returns the current retention statistics
func (s *DataRetentionScheduler) GetRetentionSummary(ctx context.Context) (*storage.CostAllocationRetentionSummary, error) {
	return s.billingRepo.GetCostAllocationRetentionSummary(ctx)
}
