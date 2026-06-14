package scheduler

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/atlas"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type ObservabilityRetentionConfig struct {
	Enabled bool
	Cron    string
	DryRun  bool
}

func DefaultObservabilityRetentionConfig() *ObservabilityRetentionConfig {
	return &ObservabilityRetentionConfig{
		Enabled: true,
		Cron:    "0 4 * * *",
		DryRun:  false,
	}
}

func LoadObservabilityRetentionConfig() *ObservabilityRetentionConfig {
	config := DefaultObservabilityRetentionConfig()

	if v := os.Getenv("OBSERVABILITY_RETENTION_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			config.Enabled = enabled
		}
	}

	if v := os.Getenv("OBSERVABILITY_RETENTION_CRON"); v != "" {
		config.Cron = v
	}

	if v := os.Getenv("OBSERVABILITY_RETENTION_DRY_RUN"); v != "" {
		if dryRun, err := strconv.ParseBool(v); err == nil {
			config.DryRun = dryRun
		}
	}

	return config
}

type ObservabilityRetentionScheduler struct {
	cron     *cron.Cron
	repo     *storage.AgentObservabilityRepository
	logger   *logrus.Logger
	config   *ObservabilityRetentionConfig
	stopOnce sync.Once
	cancel   context.CancelFunc
}

func NewObservabilityRetentionScheduler(repo *storage.AgentObservabilityRepository) *ObservabilityRetentionScheduler {
	return &ObservabilityRetentionScheduler{
		cron:   cron.New(),
		repo:   repo,
		logger: logrus.New(),
		config: LoadObservabilityRetentionConfig(),
	}
}

func (s *ObservabilityRetentionScheduler) Start(ctx context.Context) error {
	if !s.config.Enabled {
		s.logger.Info("Observability retention scheduler is disabled")
		return nil
	}

	if _, err := cron.ParseStandard(s.config.Cron); err != nil {
		return fmt.Errorf("invalid observability retention cron: %w", err)
	}

	var ctxWithCancel context.Context
	ctxWithCancel, s.cancel = context.WithCancel(ctx)

	_, err := s.cron.AddFunc(s.config.Cron, func() {
		s.runCleanup(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.cron.Start()

	s.logger.WithFields(logrus.Fields{
		"cron":    s.config.Cron,
		"dry_run": s.config.DryRun,
	}).Info("Observability retention scheduler started")

	return nil
}

func (s *ObservabilityRetentionScheduler) Stop() error {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		<-s.cron.Stop().Done()
		s.logger.Info("Observability retention scheduler stopped")
	})
	return nil
}

func (s *ObservabilityRetentionScheduler) runCleanup(ctx context.Context) {
	start := time.Now()
	s.logger.Info("Starting observability retention cleanup")

	tenants, err := s.repo.ListTenantsForRetention(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list tenants")
		return
	}

	totalDeleted := int64(0)

	for _, tenantID := range tenants {
		config, err := s.repo.GetConfig(ctx, tenantID)
		if err != nil {
			s.logger.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to get config")
			continue
		}

		retentionDays := config.RetentionDays
		if retentionDays <= 0 {
			retentionDays = 90
		}

		if s.config.DryRun {
			s.logger.WithFields(logrus.Fields{
				"tenant_id":      tenantID,
				"retention_days": retentionDays,
			}).Info("[DRY RUN] Would delete observability runs")
			continue
		}

		deleted, err := s.repo.DeleteOldRuns(ctx, tenantID, retentionDays)
		if err != nil {
			s.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to delete old runs")
			continue
		}

		if deleted > 0 {
			atlas.ObservabilityRunsCreated.WithLabelValues(
				tenantID.String(),
				"cleanup",
				"deleted",
			).Add(float64(deleted))
		}

		totalDeleted += deleted
	}

	duration := time.Since(start)
	s.logger.WithFields(logrus.Fields{
		"tenants_cleaned": len(tenants),
		"runs_deleted":    totalDeleted,
		"duration_ms":     duration.Milliseconds(),
	}).Info("Observability retention cleanup completed")
}

func (s *ObservabilityRetentionScheduler) GetSchedule() map[string]interface{} {
	nextRun := "unknown"
	if entries := s.cron.Entries(); len(entries) > 0 {
		nextRun = entries[0].Next.Format(time.RFC3339)
	}

	return map[string]interface{}{
		"enabled":  s.config.Enabled,
		"cron":     s.config.Cron,
		"next_run": nextRun,
		"dry_run":  s.config.DryRun,
	}
}
