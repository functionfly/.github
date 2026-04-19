package jobs

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// MagicLinkCleanupJob handles periodic cleanup of expired magic links
type MagicLinkCleanupJob struct {
	repo          storage.Repository
	cron          *cron.Cron
	enabled       bool
	schedule      string
	retentionDays int
}

// MagicLinkCleanupConfig holds configuration for the cleanup job
type MagicLinkCleanupConfig struct {
	Enabled       bool
	Schedule      string // Cron expression
	RetentionDays int    // How long to keep used/expired links before deletion
}

// DefaultMagicLinkCleanupConfig returns the default configuration
func DefaultMagicLinkCleanupConfig() *MagicLinkCleanupConfig {
	return &MagicLinkCleanupConfig{
		Enabled:       true,
		Schedule:      "0 2 * * *", // Daily at 2 AM
		RetentionDays: 7,           // Keep for 7 days (for audit purposes)
	}
}

// MagicLinkCleanupConfigFromEnv loads configuration from environment variables
func MagicLinkCleanupConfigFromEnv() *MagicLinkCleanupConfig {
	config := DefaultMagicLinkCleanupConfig()

	// Check if cleanup is disabled
	if os.Getenv("MAGIC_LINK_CLEANUP_DISABLED") == "true" {
		config.Enabled = false
	}

	// Override schedule
	if schedule := os.Getenv("MAGIC_LINK_CLEANUP_SCHEDULE"); schedule != "" {
		config.Schedule = schedule
	}

	// Override retention days
	if retentionStr := os.Getenv("MAGIC_LINK_CLEANUP_RETENTION_DAYS"); retentionStr != "" {
		if retention, err := strconv.Atoi(retentionStr); err == nil && retention >= 0 {
			config.RetentionDays = retention
		}
	}

	return config
}

// NewMagicLinkCleanupJob creates a new magic link cleanup job
func NewMagicLinkCleanupJob(repo storage.Repository, config *MagicLinkCleanupConfig) *MagicLinkCleanupJob {
	if config == nil {
		config = MagicLinkCleanupConfigFromEnv()
	}

	return &MagicLinkCleanupJob{
		repo:          repo,
		cron:          cron.New(),
		enabled:       config.Enabled,
		schedule:      config.Schedule,
		retentionDays: config.RetentionDays,
	}
}

// Start begins the cleanup job scheduler
func (j *MagicLinkCleanupJob) Start() error {
	if !j.enabled {
		logrus.Info("Magic link cleanup job is disabled")
		return nil
	}

	// Register the cleanup task
	_, err := j.cron.AddFunc(j.schedule, j.runCleanup)
	if err != nil {
		return err
	}

	j.cron.Start()
	logrus.WithFields(logrus.Fields{
		"schedule":       j.schedule,
		"retention_days": j.retentionDays,
	}).Info("Magic link cleanup job started")

	return nil
}

// Stop halts the cleanup job scheduler
func (j *MagicLinkCleanupJob) Stop() {
	if j.cron != nil {
		j.cron.Stop()
		logrus.Info("Magic link cleanup job stopped")
	}
}

// runCleanup performs the actual cleanup of expired magic links
func (j *MagicLinkCleanupJob) runCleanup() {
	ctx := context.Background()

	logrus.Info("Starting magic link cleanup job")

	// First, delete expired and used links (older than retention period)
	deletedCount, err := j.repo.DeleteExpiredMagicLinks(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete expired magic links")
		return
	}

	// Log results
	logrus.WithFields(logrus.Fields{
		"deleted_count": deletedCount,
	}).Info("Magic link cleanup completed")
}

// RunOnce runs the cleanup job immediately (useful for manual triggers)
func (j *MagicLinkCleanupJob) RunOnce() (int64, error) {
	ctx := context.Background()
	return j.repo.DeleteExpiredMagicLinks(ctx)
}

// GetNextRun returns the next scheduled run time
func (j *MagicLinkCleanupJob) GetNextRun() time.Time {
	if !j.enabled || j.cron == nil {
		return time.Time{}
	}

	// Get next run from the cron schedule
	entries := j.cron.Entries()
	if len(entries) > 0 {
		return entries[0].Next
	}

	return time.Time{}
}

// IsEnabled returns whether the job is enabled
func (j *MagicLinkCleanupJob) IsEnabled() bool {
	return j.enabled
}

// GetSchedule returns the cron schedule
func (j *MagicLinkCleanupJob) GetSchedule() string {
	return j.schedule
}

// GetRetentionDays returns the retention period
func (j *MagicLinkCleanupJob) GetRetentionDays() int {
	return j.retentionDays
}
