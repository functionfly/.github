package storage

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// ExecutionRetentionConfig holds configuration for execution log retention policies
type ExecutionRetentionConfig struct {
	// Retention periods (0 = disable cleanup for this table)
	ExecutionRetentionDays        int           // Default: 90 days for registry_function_executions
	PublicExecutionRetentionDays  int           // Default: 30 days for registry_executions_public
	ResourceUsageRetentionDays    int           // Default: 90 days for execution_resource_usage
	MEGRecordRetentionDays        int           // Default: 365 days for execution_meg_records
	DriftReportRetentionDays      int           // Default: 365 days for drift_reports
	ExecutionCertRetentionDays    int           // Default: 365 days for execution_certificates

	// Cleanup settings
	CleanupInterval time.Duration // Default: 24 hours
	BatchSize       int           // Default: 1000 records per batch
	VerboseLogging  bool          // Default: false
}

// DefaultExecutionRetentionConfig returns the default retention configuration
func DefaultExecutionRetentionConfig() ExecutionRetentionConfig {
	return ExecutionRetentionConfig{
		ExecutionRetentionDays:       90,
		PublicExecutionRetentionDays: 30,
		ResourceUsageRetentionDays:   90,
		MEGRecordRetentionDays:       365,
		DriftReportRetentionDays:     365,
		ExecutionCertRetentionDays:   365,
		CleanupInterval:              24 * time.Hour,
		BatchSize:                    1000,
		VerboseLogging:               false,
	}
}

// ExecutionRetentionConfigFromEnv loads retention configuration from environment variables
func ExecutionRetentionConfigFromEnv() ExecutionRetentionConfig {
	config := DefaultExecutionRetentionConfig()

	// Load retention days from environment
	if v := os.Getenv("EXECUTION_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			config.ExecutionRetentionDays = n
		}
	}
	if v := os.Getenv("PUBLIC_EXECUTION_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			config.PublicExecutionRetentionDays = n
		}
	}
	if v := os.Getenv("RESOURCE_USAGE_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			config.ResourceUsageRetentionDays = n
		}
	}
	if v := os.Getenv("MEG_RECORD_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			config.MEGRecordRetentionDays = n
		}
	}
	if v := os.Getenv("DRIFT_REPORT_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			config.DriftReportRetentionDays = n
		}
	}
	if v := os.Getenv("EXECUTION_CERT_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			config.ExecutionCertRetentionDays = n
		}
	}

	// Load cleanup settings
	if v := os.Getenv("EXECUTION_CLEANUP_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d >= time.Hour {
			config.CleanupInterval = d
		}
	}
	if v := os.Getenv("EXECUTION_CLEANUP_BATCH_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.BatchSize = n
		}
	}
	if os.Getenv("EXECUTION_CLEANUP_VERBOSE") == "true" {
		config.VerboseLogging = true
	}

	return config
}

// ExecutionLogCleanupMetrics tracks cleanup operation metrics
type ExecutionLogCleanupMetrics struct {
	LastRunAt               time.Time
	ExecutionsDeleted       int64
	PublicExecutionsDeleted int64
	ResourceUsageDeleted    int64
	MEGRecordsDeleted       int64
	DriftReportsDeleted     int64
	CertificatesDeleted     int64
	TotalDeleted            int64
	LastError               error
	DurationMs              int64
}

// CleanupCallback is called during cleanup operations for external metric reporting
type CleanupCallback func(tableName string, deleted int64, err error)

// ExecutionLogCleanupOptions contains optional callbacks and settings
type ExecutionLogCleanupOptions struct {
	OnTableCleaned CleanupCallback
}

// ExecutionLogCleanupService handles periodic cleanup of old execution logs
type ExecutionLogCleanupService struct {
	repo    Repository
	config  ExecutionRetentionConfig
	logger  *logrus.Logger
	metrics ExecutionLogCleanupMetrics
	callback CleanupCallback
}

// NewExecutionLogCleanupService creates a new execution log cleanup service
func NewExecutionLogCleanupService(repo Repository, config ExecutionRetentionConfig) *ExecutionLogCleanupService {
	return NewExecutionLogCleanupServiceWithCallback(repo, config, nil)
}

// NewExecutionLogCleanupServiceWithCallback creates a cleanup service with optional callback for metrics
func NewExecutionLogCleanupServiceWithCallback(repo Repository, config ExecutionRetentionConfig, callback CleanupCallback) *ExecutionLogCleanupService {
	if config.CleanupInterval == 0 {
		config = DefaultExecutionRetentionConfig()
	}
	if config.BatchSize == 0 {
		config.BatchSize = 1000
	}

	logger := logrus.New()
	if config.VerboseLogging {
		logger.SetLevel(logrus.DebugLevel)
	}

	return &ExecutionLogCleanupService{
		repo:     repo,
		config:   config,
		logger:   logger,
		callback: callback,
	}
}

// StartCleanupRoutine starts a background goroutine that periodically cleans up old execution logs
func (s *ExecutionLogCleanupService) StartCleanupRoutine(ctx context.Context) {
	s.logger.WithFields(logrus.Fields{
		"interval":              s.config.CleanupInterval.String(),
		"batch_size":            s.config.BatchSize,
		"execution_days":        s.config.ExecutionRetentionDays,
		"public_execution_days": s.config.PublicExecutionRetentionDays,
		"resource_usage_days":   s.config.ResourceUsageRetentionDays,
		"meg_record_days":       s.config.MEGRecordRetentionDays,
		"drift_report_days":     s.config.DriftReportRetentionDays,
		"cert_days":             s.config.ExecutionCertRetentionDays,
	}).Info("Starting execution log cleanup routine")

	// Run initial cleanup
	s.runCleanup(ctx)

	// Start ticker for periodic cleanup
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping execution log cleanup routine")
			return
		case <-ticker.C:
			s.runCleanup(ctx)
		}
	}
}

// runCleanup performs the cleanup of all execution log tables
func (s *ExecutionLogCleanupService) runCleanup(ctx context.Context) {
	start := time.Now()

	s.logger.Info("Starting execution log cleanup cycle")

	// Reset metrics for this run
	s.metrics = ExecutionLogCleanupMetrics{
		LastRunAt: start,
	}

	// Cleanup registry_function_executions
	if s.config.ExecutionRetentionDays > 0 {
		if err := s.cleanupExecutions(ctx); err != nil {
			s.logger.WithError(err).Error("Failed to cleanup executions")
			s.metrics.LastError = err
			if s.callback != nil {
				s.callback("registry_function_executions", 0, err)
			}
		}
	} else {
		s.logger.Debug("Execution cleanup disabled (ExecutionRetentionDays = 0)")
	}

	// Cleanup registry_executions_public
	if s.config.PublicExecutionRetentionDays > 0 {
		if err := s.cleanupPublicExecutions(ctx); err != nil {
			s.logger.WithError(err).Error("Failed to cleanup public executions")
			s.metrics.LastError = err
			if s.callback != nil {
				s.callback("registry_executions_public", 0, err)
			}
		}
	} else {
		s.logger.Debug("Public execution cleanup disabled (PublicExecutionRetentionDays = 0)")
	}

	// Cleanup execution_resource_usage
	if s.config.ResourceUsageRetentionDays > 0 {
		if err := s.cleanupResourceUsage(ctx); err != nil {
			s.logger.WithError(err).Error("Failed to cleanup resource usage")
			s.metrics.LastError = err
			if s.callback != nil {
				s.callback("execution_resource_usage", 0, err)
			}
		}
	} else {
		s.logger.Debug("Resource usage cleanup disabled (ResourceUsageRetentionDays = 0)")
	}

	// Cleanup execution_meg_records
	if s.config.MEGRecordRetentionDays > 0 {
		if err := s.cleanupMEGRecords(ctx); err != nil {
			s.logger.WithError(err).Error("Failed to cleanup MEG records")
			s.metrics.LastError = err
			if s.callback != nil {
				s.callback("execution_meg_records", 0, err)
			}
		}
	} else {
		s.logger.Debug("MEG record cleanup disabled (MEGRecordRetentionDays = 0)")
	}

	// Cleanup drift_reports
	if s.config.DriftReportRetentionDays > 0 {
		if err := s.cleanupDriftReports(ctx); err != nil {
			s.logger.WithError(err).Error("Failed to cleanup drift reports")
			s.metrics.LastError = err
			if s.callback != nil {
				s.callback("drift_reports", 0, err)
			}
		}
	} else {
		s.logger.Debug("Drift report cleanup disabled (DriftReportRetentionDays = 0)")
	}

	// Cleanup execution_certificates
	if s.config.ExecutionCertRetentionDays > 0 {
		if err := s.cleanupExecutionCertificates(ctx); err != nil {
			s.logger.WithError(err).Error("Failed to cleanup execution certificates")
			s.metrics.LastError = err
			if s.callback != nil {
				s.callback("execution_certificates", 0, err)
			}
		}
	} else {
		s.logger.Debug("Execution certificate cleanup disabled (ExecutionCertRetentionDays = 0)")
	}

	// Update metrics
	duration := time.Since(start)
	s.metrics.DurationMs = duration.Milliseconds()
	s.metrics.TotalDeleted = s.metrics.ExecutionsDeleted +
		s.metrics.PublicExecutionsDeleted +
		s.metrics.ResourceUsageDeleted +
		s.metrics.MEGRecordsDeleted +
		s.metrics.DriftReportsDeleted +
		s.metrics.CertificatesDeleted

	// Log summary
	fields := logrus.Fields{
		"duration_ms":              duration.Milliseconds(),
		"executions_deleted":       s.metrics.ExecutionsDeleted,
		"public_executions_deleted": s.metrics.PublicExecutionsDeleted,
		"resource_usage_deleted":   s.metrics.ResourceUsageDeleted,
		"meg_records_deleted":      s.metrics.MEGRecordsDeleted,
		"drift_reports_deleted":    s.metrics.DriftReportsDeleted,
		"certificates_deleted":     s.metrics.CertificatesDeleted,
		"total_deleted":            s.metrics.TotalDeleted,
	}

	if s.metrics.LastError != nil {
		fields["had_errors"] = true
		s.logger.WithFields(fields).Warn("Execution log cleanup completed with errors")
	} else {
		s.logger.WithFields(fields).Info("Execution log cleanup completed successfully")
	}
}

// cleanupExecutions removes old executions from registry_function_executions
func (s *ExecutionLogCleanupService) cleanupExecutions(ctx context.Context) error {
	cutoff := time.Now().AddDate(0, 0, -s.config.ExecutionRetentionDays)

	deleted, err := s.repo.DeleteOldExecutions(ctx, cutoff, s.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to delete old executions: %w", err)
	}

	s.metrics.ExecutionsDeleted = deleted

	if deleted > 0 || s.config.VerboseLogging {
		s.logger.WithFields(logrus.Fields{
			"deleted":       deleted,
			"retention_days": s.config.ExecutionRetentionDays,
			"cutoff":        cutoff.Format(time.RFC3339),
		}).Info("Cleaned up old executions")
	}

	return nil
}

// cleanupPublicExecutions removes old executions from registry_executions_public
func (s *ExecutionLogCleanupService) cleanupPublicExecutions(ctx context.Context) error {
	cutoff := time.Now().AddDate(0, 0, -s.config.PublicExecutionRetentionDays)

	deleted, err := s.repo.DeleteOldPublicExecutions(ctx, cutoff, s.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to delete old public executions: %w", err)
	}

	s.metrics.PublicExecutionsDeleted = deleted

	if deleted > 0 || s.config.VerboseLogging {
		s.logger.WithFields(logrus.Fields{
			"deleted":       deleted,
			"retention_days": s.config.PublicExecutionRetentionDays,
			"cutoff":        cutoff.Format(time.RFC3339),
		}).Info("Cleaned up old public executions")
	}

	return nil
}

// cleanupResourceUsage removes old records from execution_resource_usage
func (s *ExecutionLogCleanupService) cleanupResourceUsage(ctx context.Context) error {
	cutoff := time.Now().AddDate(0, 0, -s.config.ResourceUsageRetentionDays)

	deleted, err := s.repo.DeleteOldResourceUsage(ctx, cutoff, s.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to delete old resource usage: %w", err)
	}

	s.metrics.ResourceUsageDeleted = deleted

	if deleted > 0 || s.config.VerboseLogging {
		s.logger.WithFields(logrus.Fields{
			"deleted":       deleted,
			"retention_days": s.config.ResourceUsageRetentionDays,
			"cutoff":        cutoff.Format(time.RFC3339),
		}).Info("Cleaned up old resource usage records")
	}

	return nil
}

// cleanupMEGRecords removes old records from execution_meg_records
func (s *ExecutionLogCleanupService) cleanupMEGRecords(ctx context.Context) error {
	cutoff := time.Now().AddDate(0, 0, -s.config.MEGRecordRetentionDays)

	deleted, err := s.repo.DeleteOldMEGRecords(ctx, cutoff, s.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to delete old MEG records: %w", err)
	}

	s.metrics.MEGRecordsDeleted = deleted

	if deleted > 0 || s.config.VerboseLogging {
		s.logger.WithFields(logrus.Fields{
			"deleted":       deleted,
			"retention_days": s.config.MEGRecordRetentionDays,
			"cutoff":        cutoff.Format(time.RFC3339),
		}).Info("Cleaned up old MEG records")
	}

	return nil
}

// cleanupDriftReports removes old records from drift_reports
func (s *ExecutionLogCleanupService) cleanupDriftReports(ctx context.Context) error {
	cutoff := time.Now().AddDate(0, 0, -s.config.DriftReportRetentionDays)

	deleted, err := s.repo.DeleteOldDriftReports(ctx, cutoff, s.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to delete old drift reports: %w", err)
	}

	s.metrics.DriftReportsDeleted = deleted

	if deleted > 0 || s.config.VerboseLogging {
		s.logger.WithFields(logrus.Fields{
			"deleted":       deleted,
			"retention_days": s.config.DriftReportRetentionDays,
			"cutoff":        cutoff.Format(time.RFC3339),
		}).Info("Cleaned up old drift reports")
	}

	return nil
}

// cleanupExecutionCertificates removes old records from execution_certificates
func (s *ExecutionLogCleanupService) cleanupExecutionCertificates(ctx context.Context) error {
	cutoff := time.Now().AddDate(0, 0, -s.config.ExecutionCertRetentionDays)

	deleted, err := s.repo.DeleteOldExecutionCertificates(ctx, cutoff, s.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to delete old execution certificates: %w", err)
	}

	s.metrics.CertificatesDeleted = deleted

	if deleted > 0 || s.config.VerboseLogging {
		s.logger.WithFields(logrus.Fields{
			"deleted":       deleted,
			"retention_days": s.config.ExecutionCertRetentionDays,
			"cutoff":        cutoff.Format(time.RFC3339),
		}).Info("Cleaned up old execution certificates")
	}

	return nil
}

// GetMetrics returns the current cleanup metrics
func (s *ExecutionLogCleanupService) GetMetrics() ExecutionLogCleanupMetrics {
	return s.metrics
}

// GetConfig returns the current retention configuration
func (s *ExecutionLogCleanupService) GetConfig() ExecutionRetentionConfig {
	return s.config
}

// UpdateConfig updates the retention configuration (can be used for runtime config changes)
func (s *ExecutionLogCleanupService) UpdateConfig(config ExecutionRetentionConfig) {
	s.config = config
	s.logger.WithFields(logrus.Fields{
		"execution_days":        config.ExecutionRetentionDays,
		"public_execution_days": config.PublicExecutionRetentionDays,
		"resource_usage_days":   config.ResourceUsageRetentionDays,
		"meg_record_days":       config.MEGRecordRetentionDays,
		"drift_report_days":     config.DriftReportRetentionDays,
		"cert_days":             config.ExecutionCertRetentionDays,
	}).Info("Updated execution log retention configuration")
}

// RunManualCleanup triggers an immediate cleanup (for admin use)
func (s *ExecutionLogCleanupService) RunManualCleanup(ctx context.Context) (*ExecutionLogCleanupMetrics, error) {
	s.runCleanup(ctx)
	return &s.metrics, s.metrics.LastError
}

// UpdateTableStats retrieves current table statistics for monitoring
func (s *ExecutionLogCleanupService) UpdateTableStats(ctx context.Context) (map[string]interface{}, error) {
	stats, err := s.repo.GetExecutionRetentionStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution retention stats: %w", err)
	}

	return stats, nil
}

// GetCleanupConfigSummary returns a summary of the current cleanup configuration
func (s *ExecutionLogCleanupService) GetCleanupConfigSummary() map[string]interface{} {
	return map[string]interface{}{
		"execution_retention_days":        s.config.ExecutionRetentionDays,
		"public_execution_retention_days":  s.config.PublicExecutionRetentionDays,
		"resource_usage_retention_days":   s.config.ResourceUsageRetentionDays,
		"meg_record_retention_days":       s.config.MEGRecordRetentionDays,
		"drift_report_retention_days":     s.config.DriftReportRetentionDays,
		"execution_cert_retention_days":   s.config.ExecutionCertRetentionDays,
		"cleanup_interval_minutes":        int(s.config.CleanupInterval.Minutes()),
		"batch_size":                      s.config.BatchSize,
		"verbose_logging":                 s.config.VerboseLogging,
	}
}
