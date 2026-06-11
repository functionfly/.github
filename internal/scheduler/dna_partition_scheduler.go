package scheduler

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	dnaStorage "github.com/functionfly/functionfly/internal/storage/dna"
	"github.com/functionfly/functionfly/internal/tracing"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// DNAPartitionSchedulerConfig holds configuration for the DNA partition scheduler.
type DNAPartitionSchedulerConfig struct {
	// Cron is the cron expression for running the partition job (default: "0 2 1 * *" — 2 AM on 1st of month)
	Cron string

	// MonthsAhead is how many future months to create partitions for (default: 3)
	MonthsAhead int

	// RetentionMonths is how many months of old partitions to keep (default: 6)
	RetentionMonths int

	// Enabled controls whether the scheduler is active
	Enabled bool
}

// DefaultDNAPartitionSchedulerConfig returns default configuration.
func DefaultDNAPartitionSchedulerConfig() *DNAPartitionSchedulerConfig {
	return &DNAPartitionSchedulerConfig{
		Cron:            "0 2 1 * *",
		MonthsAhead:     3,
		RetentionMonths: 6,
		Enabled:         true,
	}
}

// LoadDNAPartitionSchedulerConfig loads configuration from environment.
func LoadDNAPartitionSchedulerConfig() *DNAPartitionSchedulerConfig {
	config := DefaultDNAPartitionSchedulerConfig()

	if v := os.Getenv("DNA_PARTITION_CRON"); v != "" {
		config.Cron = v
	}
	if v := os.Getenv("DNA_PARTITION_MONTHS_AHEAD"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.MonthsAhead = n
		}
	}
	if v := os.Getenv("DNA_PARTITION_RETENTION_MONTHS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.RetentionMonths = n
		}
	}
	if v := os.Getenv("DNA_PARTITION_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			config.Enabled = enabled
		}
	}

	return config
}

// DNAPartitionScheduler manages monthly partition creation and cleanup
// for the function_dna_execution_metrics table.
type DNAPartitionScheduler struct {
	cron     *cron.Cron
	repo     *dnaStorage.Repository
	logger   *logrus.Logger
	config   *DNAPartitionSchedulerConfig
	stopOnce sync.Once
	cancel   context.CancelFunc
}

// NewDNAPartitionScheduler creates a new DNA partition scheduler.
func NewDNAPartitionScheduler(repo *dnaStorage.Repository) *DNAPartitionScheduler {
	return &DNAPartitionScheduler{
		cron:   cron.New(),
		repo:   repo,
		logger: logrus.StandardLogger(),
		config: LoadDNAPartitionSchedulerConfig(),
	}
}

// Start begins the partition scheduler.
func (s *DNAPartitionScheduler) Start(ctx context.Context) error {
	if !s.config.Enabled {
		s.logger.Info("DNA partition scheduler is disabled")
		return nil
	}

	if _, err := cron.ParseStandard(s.config.Cron); err != nil {
		return fmt.Errorf("invalid DNA partition cron expression: %w", err)
	}

	var ctxWithCancel context.Context
	ctxWithCancel, s.cancel = context.WithCancel(ctx)

	_, err := s.cron.AddFunc(s.config.Cron, func() {
		s.runPartitionMaintenance(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add partition cron job: %w", err)
	}

	s.cron.Start()

	s.logger.WithFields(logrus.Fields{
		"cron":              s.config.Cron,
		"months_ahead":      s.config.MonthsAhead,
		"retention_months":  s.config.RetentionMonths,
	}).Info("DNA partition scheduler started")

	return nil
}

// Stop stops the partition scheduler.
func (s *DNAPartitionScheduler) Stop() error {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		<-s.cron.Stop().Done()
		s.logger.Info("DNA partition scheduler stopped")
	})
	return nil
}

// runPartitionMaintenance creates future partitions and drops old ones.
func (s *DNAPartitionScheduler) runPartitionMaintenance(ctx context.Context) {
	ctx, span := tracing.StartSpan(ctx, "dna.partition_maintenance")
	defer tracing.Finish(ctx)

	start := time.Now()
	s.logger.Info("Starting DNA partition maintenance")

	tracing.SetAttribute(ctx, "months_ahead", s.config.MonthsAhead)
	tracing.SetAttribute(ctx, "retention_months", s.config.RetentionMonths)

	// Create future partitions
	created, err := s.repo.CreateFuturePartitions(ctx, s.config.MonthsAhead)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create future DNA partitions")
		tracing.RecordError(ctx, err)
	} else {
		s.logger.WithField("partitions_created", created).Info("Future DNA partitions created")
	}
	tracing.SetAttribute(ctx, "partitions_created", created)

	// Drop old partitions
	dropped, err := s.repo.DropOldPartitions(ctx, s.config.RetentionMonths)
	if err != nil {
		s.logger.WithError(err).Error("Failed to drop old DNA partitions")
		tracing.RecordError(ctx, err)
	} else {
		s.logger.WithField("partitions_dropped", dropped).Info("Old DNA partitions dropped")
	}
	tracing.SetAttribute(ctx, "partitions_dropped", dropped)

	duration := time.Since(start)
	tracing.SetAttribute(ctx, "duration_ms", duration.Milliseconds())

	s.logger.WithFields(logrus.Fields{
		"created":      created,
		"dropped":      dropped,
		"duration_ms":  duration.Milliseconds(),
	}).Info("DNA partition maintenance completed")
}

// RunNow triggers an immediate partition maintenance cycle (for admin/manual use).
func (s *DNAPartitionScheduler) RunNow(ctx context.Context) error {
	s.logger.Info("Manually triggering DNA partition maintenance")
	s.runPartitionMaintenance(ctx)
	return nil
}

// GetStatus returns the current scheduler status.
func (s *DNAPartitionScheduler) GetStatus(ctx context.Context) map[string]interface{} {
	nextRun := "unknown"
	if s.cron != nil {
		entries := s.cron.Entries()
		if len(entries) > 0 {
			nextRun = entries[0].Next.Format(time.RFC3339)
		}
	}

	partitions, _ := s.repo.ListPartitions(ctx)

	return map[string]interface{}{
		"enabled":           s.config.Enabled,
		"cron":              s.config.Cron,
		"next_run":          nextRun,
		"months_ahead":      s.config.MonthsAhead,
		"retention_months":  s.config.RetentionMonths,
		"current_partitions": partitions,
	}
}
