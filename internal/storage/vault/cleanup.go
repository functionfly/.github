package vault

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type RetentionConfig struct {
	RetentionDays      int
	KeepLatestVersions int
	CleanupInterval    time.Duration
	BatchSize          int
	VerboseLogging     bool
	DryRun             bool
}

func DefaultRetentionConfig() RetentionConfig {
	return RetentionConfig{
		RetentionDays:      90,
		KeepLatestVersions: 5,
		CleanupInterval:    24 * time.Hour,
		BatchSize:          1000,
		VerboseLogging:     false,
		DryRun:             false,
	}
}

func RetentionConfigFromEnv() RetentionConfig {
	config := DefaultRetentionConfig()

	if v := os.Getenv("SECRET_VERSION_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			config.RetentionDays = n
		}
	}

	if v := os.Getenv("SECRET_VERSION_KEEP_LATEST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			config.KeepLatestVersions = n
		}
	}

	if v := os.Getenv("SECRET_VERSION_CLEANUP_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d >= time.Hour {
			config.CleanupInterval = d
		}
	}

	if v := os.Getenv("SECRET_VERSION_CLEANUP_BATCH_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.BatchSize = n
		}
	}

	if os.Getenv("SECRET_VERSION_CLEANUP_VERBOSE") == "true" {
		config.VerboseLogging = true
	}

	if os.Getenv("SECRET_VERSION_CLEANUP_DRY_RUN") == "true" {
		config.DryRun = true
	}

	return config
}

type CleanupMetrics struct {
	LastRunAt          time.Time
	VersionsDeleted    int64
	LastError          error
	DurationMs         int64
	RetentionDays      int
	KeepLatestVersions int
}

type CleanupService struct {
	repo    *Repository
	config  RetentionConfig
	logger  *logrus.Logger
	metrics CleanupMetrics
}

func NewCleanupService(repo *Repository, config RetentionConfig) *CleanupService {
	if config.CleanupInterval == 0 {
		config = DefaultRetentionConfig()
	}
	if config.KeepLatestVersions == 0 {
		config.KeepLatestVersions = 5
	}
	if config.BatchSize == 0 {
		config.BatchSize = 1000
	}

	logger := logrus.New()
	if config.VerboseLogging {
		logger.SetLevel(logrus.DebugLevel)
	}

	return &CleanupService{
		repo:   repo,
		config: config,
		logger: logger,
	}
}

func (s *CleanupService) StartCleanupRoutine(ctx context.Context) {
	if s.config.RetentionDays == 0 {
		s.logger.Info("Secret version retention disabled (retention days = 0)")
		return
	}

	s.logger.WithFields(logrus.Fields{
		"interval":            s.config.CleanupInterval.String(),
		"batch_size":          s.config.BatchSize,
		"retention_days":      s.config.RetentionDays,
		"keep_latest":         s.config.KeepLatestVersions,
		"dry_run":             s.config.DryRun,
	}).Info("Starting secret version cleanup routine")

	s.runCleanup(ctx)

	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping secret version cleanup routine")
			return
		case <-ticker.C:
			s.runCleanup(ctx)
		}
	}
}

func (s *CleanupService) runCleanup(ctx context.Context) {
	start := time.Now()

	s.logger.Info("Starting secret version cleanup cycle")

	s.metrics = CleanupMetrics{
		LastRunAt:          start,
		RetentionDays:      s.config.RetentionDays,
		KeepLatestVersions: s.config.KeepLatestVersions,
	}

	if s.config.DryRun {
		s.logger.Warn("DRY RUN MODE: No versions will be deleted")
	}

	olderThan := time.Duration(s.config.RetentionDays) * 24 * time.Hour

	tenants, err := s.repo.GetTenantsWithSecrets(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get tenants with secrets")
		s.metrics.LastError = err
		return
	}

	s.logger.WithField("tenant_count", len(tenants)).Debug("Processing tenants")

	var totalDeleted int64

	for _, tenantID := range tenants {
		if s.config.DryRun {
			count, err := s.repo.CountSecretVersionsByAge(ctx, olderThan)
			if err != nil {
				s.logger.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to count old versions")
				continue
			}
			s.logger.WithFields(logrus.Fields{
				"tenant_id": tenantID,
				"count":     count,
			}).Info("[DRY RUN] Would delete old versions")
			totalDeleted += count
		} else {
			deleted, err := s.repo.DeleteSecretVersionsByTenant(ctx, tenantID, olderThan, s.config.KeepLatestVersions)
			if err != nil {
				s.logger.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to delete old versions")
				s.metrics.LastError = err
				continue
			}
			if deleted > 0 {
				s.logger.WithFields(logrus.Fields{
					"tenant_id": tenantID,
					"deleted":   deleted,
				}).Debug("Deleted old secret versions")
			}
			totalDeleted += deleted
		}
	}

	s.metrics.VersionsDeleted = totalDeleted
	s.metrics.DurationMs = time.Since(start).Milliseconds()

	s.logger.WithFields(logrus.Fields{
		"total_deleted": totalDeleted,
		"duration_ms":   s.metrics.DurationMs,
	}).Info("Secret version cleanup cycle completed")
}

func (s *CleanupService) RunManualCleanup(ctx context.Context) (int64, error) {
	start := time.Now()

	if s.config.DryRun {
		s.logger.Warn("DRY RUN MODE: No versions will be deleted")
	}

	olderThan := time.Duration(s.config.RetentionDays) * 24 * time.Hour

	tenants, err := s.repo.GetTenantsWithSecrets(ctx)
	if err != nil {
		return 0, err
	}

	var totalDeleted int64

	for _, tenantID := range tenants {
		if s.config.DryRun {
			count, err := s.repo.CountSecretVersionsByAge(ctx, olderThan)
			if err != nil {
				continue
			}
			totalDeleted += count
		} else {
			deleted, err := s.repo.DeleteSecretVersionsByTenant(ctx, tenantID, olderThan, s.config.KeepLatestVersions)
			if err != nil {
				continue
			}
			totalDeleted += deleted
		}
	}

	s.logger.WithFields(logrus.Fields{
		"total_deleted": totalDeleted,
		"duration":      time.Since(start),
	}).Info("Manual secret version cleanup completed")

	return totalDeleted, nil
}

func (s *CleanupService) GetMetrics() CleanupMetrics {
	return s.metrics
}

func (s *CleanupService) GetConfig() RetentionConfig {
	return s.config
}

func (s *CleanupService) UpdateConfig(config RetentionConfig) {
	s.config = config
	if config.VerboseLogging {
		s.logger.SetLevel(logrus.DebugLevel)
	} else {
		s.logger.SetLevel(logrus.InfoLevel)
	}
}