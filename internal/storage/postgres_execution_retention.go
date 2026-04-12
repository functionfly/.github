package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Execution retention operations - delegate to the GORM DB directly

func (db *PostgresDB) DeleteOldExecutions(ctx context.Context, cutoff time.Time, batchSize int) (int64, error) {
	return deleteOldRecordsInBatches(ctx, db.GORM, "registry_function_executions", "timestamp", cutoff, batchSize)
}

func (db *PostgresDB) DeleteOldPublicExecutions(ctx context.Context, cutoff time.Time, batchSize int) (int64, error) {
	return deleteOldRecordsInBatches(ctx, db.GORM, "registry_executions_public", "created_at", cutoff, batchSize)
}

func (db *PostgresDB) DeleteOldResourceUsage(ctx context.Context, cutoff time.Time, batchSize int) (int64, error) {
	return deleteOldRecordsInBatches(ctx, db.GORM, "execution_resource_usage", "created_at", cutoff, batchSize)
}

func (db *PostgresDB) DeleteOldMEGRecords(ctx context.Context, cutoff time.Time, batchSize int) (int64, error) {
	return deleteOldRecordsInBatches(ctx, db.GORM, "execution_meg_records", "created_at", cutoff, batchSize)
}

func (db *PostgresDB) DeleteOldDriftReports(ctx context.Context, cutoff time.Time, batchSize int) (int64, error) {
	return deleteOldRecordsInBatches(ctx, db.GORM, "drift_reports", "detected_at", cutoff, batchSize)
}

func (db *PostgresDB) DeleteOldExecutionCertificates(ctx context.Context, cutoff time.Time, batchSize int) (int64, error) {
	return deleteOldRecordsInBatches(ctx, db.GORM, "execution_certificates", "created_at", cutoff, batchSize)
}

// deleteOldRecordsInBatches deletes old records in batches to avoid long-running transactions
func deleteOldRecordsInBatches(ctx context.Context, db *gorm.DB, tableName, timestampColumn string, cutoff time.Time, batchSize int) (int64, error) {
	var totalDeleted int64

	for {
		// Get IDs of old records in a batch
		var ids []uuid.UUID
		err := db.WithContext(ctx).
			Table(tableName).
			Where(fmt.Sprintf("%s < ?", timestampColumn), cutoff).
			Limit(batchSize).
			Pluck("id", &ids).Error

		if err != nil {
			return totalDeleted, fmt.Errorf("failed to query old records from %s: %w", tableName, err)
		}

		if len(ids) == 0 {
			break
		}

		// Delete the batch
		result := db.WithContext(ctx).
			Table(tableName).
			Where("id IN ?", ids).
			Delete(nil)

		if result.Error != nil {
			return totalDeleted, fmt.Errorf("failed to delete %s batch: %w", tableName, result.Error)
		}

		totalDeleted += result.RowsAffected

		if result.RowsAffected < int64(batchSize) {
			break
		}
	}

	return totalDeleted, nil
}

func (db *PostgresDB) GetExecutionRetentionStats(ctx context.Context) (map[string]interface{}, error) {
	return getExecutionRetentionStats(ctx, db.GORM)
}

// getExecutionRetentionStats returns statistics about execution data for retention planning
func getExecutionRetentionStats(ctx context.Context, db *gorm.DB) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	now := time.Now()

	// Helper to get stats for a table
	getTableStats := func(tableName, timestampColumn string) (map[string]interface{}, error) {
		tableStats := make(map[string]interface{})

		// Total count
		var total int64
		if err := db.WithContext(ctx).Table(tableName).Count(&total).Error; err != nil {
			return nil, fmt.Errorf("failed to count %s: %w", tableName, err)
		}
		tableStats["total"] = total

		// Count by age ranges
		var count30d, count90d, count365d int64

		if err := db.WithContext(ctx).Table(tableName).
			Where(fmt.Sprintf("%s >= ?", timestampColumn), now.AddDate(0, 0, -30)).
			Count(&count30d).Error; err != nil {
			return nil, fmt.Errorf("failed to count 30-day %s: %w", tableName, err)
		}
		tableStats["older_than_30d"] = total - count30d

		if err := db.WithContext(ctx).Table(tableName).
			Where(fmt.Sprintf("%s >= ?", timestampColumn), now.AddDate(0, 0, -90)).
			Count(&count90d).Error; err != nil {
			return nil, fmt.Errorf("failed to count 90-day %s: %w", tableName, err)
		}
		tableStats["older_than_90d"] = total - count90d

		if err := db.WithContext(ctx).Table(tableName).
			Where(fmt.Sprintf("%s >= ?", timestampColumn), now.AddDate(0, 0, -365)).
			Count(&count365d).Error; err != nil {
			return nil, fmt.Errorf("failed to count 365-day %s: %w", tableName, err)
		}
		tableStats["older_than_365d"] = total - count365d

		return tableStats, nil
	}

	// Get stats for each table
	tables := map[string]string{
		"registry_function_executions": "timestamp",
		"registry_executions_public":   "created_at",
		"execution_resource_usage":     "created_at",
		"execution_meg_records":        "created_at",
		"drift_reports":                "detected_at",
		"execution_certificates":       "created_at",
	}

	for tableName, timestampColumn := range tables {
		tableStats, err := getTableStats(tableName, timestampColumn)
		if err != nil {
			return nil, err
		}
		stats[tableName] = tableStats
	}

	// Add legacy summary stats for backward compatibility
	if execStats, ok := stats["registry_function_executions"].(map[string]interface{}); ok {
		if total, ok := execStats["total"].(int64); ok {
			stats["total_executions"] = total
		}
		if older90d, ok := execStats["older_than_90d"].(int64); ok {
			stats["executions_older_than_90_days"] = older90d
		}
	}

	return stats, nil
}

// Execution retention settings operations

func (db *PostgresDB) GetExecutionRetentionSettings(ctx context.Context) (*ExecutionRetentionSettings, error) {
	var settings ExecutionRetentionSettings

	err := db.GORM.WithContext(ctx).
		Where("is_active = ?", true).
		First(&settings).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get execution retention settings: %w", err)
	}

	return &settings, nil
}

func (db *PostgresDB) UpdateExecutionRetentionSettings(ctx context.Context, updates *ExecutionRetentionSettingsUpdate) (*ExecutionRetentionSettings, error) {
	// First get the current settings
	settings, err := db.GetExecutionRetentionSettings(ctx)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.ExecutionRetentionDays != nil && *updates.ExecutionRetentionDays >= 0 {
		settings.ExecutionRetentionDays = *updates.ExecutionRetentionDays
	}
	if updates.PublicExecutionRetentionDays != nil && *updates.PublicExecutionRetentionDays >= 0 {
		settings.PublicExecutionRetentionDays = *updates.PublicExecutionRetentionDays
	}
	if updates.ResourceUsageRetentionDays != nil && *updates.ResourceUsageRetentionDays >= 0 {
		settings.ResourceUsageRetentionDays = *updates.ResourceUsageRetentionDays
	}
	if updates.MEGRecordRetentionDays != nil && *updates.MEGRecordRetentionDays >= 0 {
		settings.MEGRecordRetentionDays = *updates.MEGRecordRetentionDays
	}
	if updates.DriftReportRetentionDays != nil && *updates.DriftReportRetentionDays >= 0 {
		settings.DriftReportRetentionDays = *updates.DriftReportRetentionDays
	}
	if updates.ExecutionCertRetentionDays != nil && *updates.ExecutionCertRetentionDays >= 0 {
		settings.ExecutionCertRetentionDays = *updates.ExecutionCertRetentionDays
	}
	if updates.CleanupIntervalMinutes != nil && *updates.CleanupIntervalMinutes >= 60 {
		settings.CleanupIntervalMinutes = *updates.CleanupIntervalMinutes
	}
	if updates.BatchSize != nil && *updates.BatchSize > 0 {
		settings.BatchSize = *updates.BatchSize
	}
	if updates.VerboseLogging != nil {
		settings.VerboseLogging = *updates.VerboseLogging
	}
	if updates.UpdatedBy != nil {
		settings.UpdatedBy = updates.UpdatedBy
	}
	settings.UpdatedAt = time.Now()

	// Save updated settings
	err = db.GORM.WithContext(ctx).Save(settings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update execution retention settings: %w", err)
	}

	return settings, nil
}

func (db *PostgresDB) GetOrCreateExecutionRetentionSettings(ctx context.Context) (*ExecutionRetentionSettings, error) {
	settings, err := db.GetExecutionRetentionSettings(ctx)
	if err == nil {
		return settings, nil
	}

	// Create default settings if none exist
	defaultConfig := DefaultExecutionRetentionConfig()
	settings = &ExecutionRetentionSettings{
		ExecutionRetentionDays:       defaultConfig.ExecutionRetentionDays,
		PublicExecutionRetentionDays: defaultConfig.PublicExecutionRetentionDays,
		ResourceUsageRetentionDays:   defaultConfig.ResourceUsageRetentionDays,
		MEGRecordRetentionDays:       defaultConfig.MEGRecordRetentionDays,
		DriftReportRetentionDays:     defaultConfig.DriftReportRetentionDays,
		ExecutionCertRetentionDays:   defaultConfig.ExecutionCertRetentionDays,
		CleanupIntervalMinutes:       int(defaultConfig.CleanupInterval.Minutes()),
		BatchSize:                    defaultConfig.BatchSize,
		VerboseLogging:               defaultConfig.VerboseLogging,
		IsActive:                     true,
	}

	err = db.GORM.WithContext(ctx).Create(settings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create default execution retention settings: %w", err)
	}

	return settings, nil
}

func (db *PostgresDB) ResetExecutionRetentionSettingsToDefaults(ctx context.Context, updatedBy *uuid.UUID) (*ExecutionRetentionSettings, error) {
	defaultConfig := DefaultExecutionRetentionConfig()

	cleanupIntervalMinutes := int(defaultConfig.CleanupInterval.Minutes())
	updates := &ExecutionRetentionSettingsUpdate{
		ExecutionRetentionDays:       &defaultConfig.ExecutionRetentionDays,
		PublicExecutionRetentionDays: &defaultConfig.PublicExecutionRetentionDays,
		ResourceUsageRetentionDays:   &defaultConfig.ResourceUsageRetentionDays,
		MEGRecordRetentionDays:       &defaultConfig.MEGRecordRetentionDays,
		DriftReportRetentionDays:     &defaultConfig.DriftReportRetentionDays,
		ExecutionCertRetentionDays:   &defaultConfig.ExecutionCertRetentionDays,
		CleanupIntervalMinutes:       &cleanupIntervalMinutes,
		BatchSize:                    &defaultConfig.BatchSize,
		VerboseLogging:               &defaultConfig.VerboseLogging,
		UpdatedBy:                    updatedBy,
	}

	return db.UpdateExecutionRetentionSettings(ctx, updates)
}
