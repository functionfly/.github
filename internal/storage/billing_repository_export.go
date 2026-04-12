package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ExportRepository handles export and external billing integration operations
type ExportRepository struct {
	db *sql.DB
}

// NewExportRepository creates a new export repository
func NewExportRepository(db *sql.DB) *ExportRepository {
	return &ExportRepository{db: db}
}

// ==================== Usage Export Configuration ====================

// CreateUsageExportConfiguration creates a new export configuration
func (r *ExportRepository) CreateUsageExportConfiguration(ctx context.Context, config *UsageExportConfiguration) error {
	query := `
		INSERT INTO usage_export_configurations (
			id, tenant_id, name, description, format, data_types, granularity,
			include_metadata, include_breakdown, date_range_type, function_filter,
			region_filter, outcome_filter, is_scheduled, schedule_frequency,
			schedule_day_of_month, schedule_day_of_week, schedule_hour,
			delivery_method, email_recipients, webhook_url, s3_bucket, s3_prefix,
			external_system_id, field_mapping, transform_config, is_active,
			created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30)
	`

	_, err := r.db.ExecContext(ctx, query,
		config.ID, config.TenantID, config.Name, config.Description,
		config.Format, pq.Array(config.DataTypes), config.Granularity,
		config.IncludeMetadata, config.IncludeBreakdown, config.DateRangeType,
		pq.Array(config.FunctionFilter), pq.Array(config.RegionFilter), pq.Array(config.OutcomeFilter),
		config.IsScheduled, config.ScheduleFrequency,
		config.ScheduleDayOfMonth, config.ScheduleDayOfWeek, config.ScheduleHour,
		config.DeliveryMethod, pq.Array(config.EmailRecipients), config.WebhookURL,
		config.S3Bucket, config.S3Prefix, config.ExternalSystemID,
		config.FieldMapping, config.TransformConfig, config.IsActive,
		config.CreatedBy, config.CreatedAt, config.UpdatedAt,
	)

	return err
}

// GetUsageExportConfiguration retrieves an export configuration by ID
func (r *ExportRepository) GetUsageExportConfiguration(ctx context.Context, id uuid.UUID) (*UsageExportConfiguration, error) {
	query := `
		SELECT id, tenant_id, name, description, format, data_types, granularity,
			include_metadata, include_breakdown, date_range_type, function_filter,
			region_filter, outcome_filter, is_scheduled, schedule_frequency,
			schedule_day_of_month, schedule_day_of_week, schedule_hour,
			delivery_method, email_recipients, webhook_url, s3_bucket, s3_prefix,
			external_system_id, field_mapping, transform_config, is_active,
			created_by, created_at, updated_at, last_executed_at, last_export_id
		FROM usage_export_configurations WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanExportConfig(row)
}

// ListUsageExportConfigurations lists all export configurations for a tenant
func (r *ExportRepository) ListUsageExportConfigurations(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*UsageExportConfiguration, error) {
	query := `
		SELECT id, tenant_id, name, description, format, data_types, granularity,
			include_metadata, include_breakdown, date_range_type, function_filter,
			region_filter, outcome_filter, is_scheduled, schedule_frequency,
			schedule_day_of_month, schedule_day_of_week, schedule_hour,
			delivery_method, email_recipients, webhook_url, s3_bucket, s3_prefix,
			external_system_id, field_mapping, transform_config, is_active,
			created_by, created_at, updated_at, last_executed_at, last_export_id
		FROM usage_export_configurations
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanExportConfigRows(rows)
}

// UpdateUsageExportConfiguration updates an export configuration
func (r *ExportRepository) UpdateUsageExportConfiguration(ctx context.Context, config *UsageExportConfiguration) error {
	query := `
		UPDATE usage_export_configurations SET
			name = $2, description = $3, format = $4, data_types = $5, granularity = $6,
			include_metadata = $7, include_breakdown = $8, date_range_type = $9,
			function_filter = $10, region_filter = $11, outcome_filter = $12,
			is_scheduled = $13, schedule_frequency = $14,
			schedule_day_of_month = $15, schedule_day_of_week = $16, schedule_hour = $17,
			delivery_method = $18, email_recipients = $19, webhook_url = $20,
			s3_bucket = $21, s3_prefix = $22, external_system_id = $23,
			field_mapping = $24, transform_config = $25, is_active = $26,
			updated_at = $27
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		config.ID, config.Name, config.Description,
		config.Format, pq.Array(config.DataTypes), config.Granularity,
		config.IncludeMetadata, config.IncludeBreakdown, config.DateRangeType,
		pq.Array(config.FunctionFilter), pq.Array(config.RegionFilter), pq.Array(config.OutcomeFilter),
		config.IsScheduled, config.ScheduleFrequency,
		config.ScheduleDayOfMonth, config.ScheduleDayOfWeek, config.ScheduleHour,
		config.DeliveryMethod, pq.Array(config.EmailRecipients), config.WebhookURL,
		config.S3Bucket, config.S3Prefix, config.ExternalSystemID,
		config.FieldMapping, config.TransformConfig, config.IsActive,
		time.Now(),
	)

	return err
}

// DeleteUsageExportConfiguration deletes an export configuration
func (r *ExportRepository) DeleteUsageExportConfiguration(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM usage_export_configurations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// UpdateLastExecution updates the last execution timestamp and export ID
func (r *ExportRepository) UpdateLastExecution(ctx context.Context, configID, exportID uuid.UUID, executedAt time.Time) error {
	query := `
		UPDATE usage_export_configurations
		SET last_executed_at = $2, last_export_id = $3, updated_at = $4
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, configID, executedAt, exportID, time.Now())
	return err
}

// ==================== Usage Export Jobs ====================

// CreateUsageExportJob creates a new export job
func (r *ExportRepository) CreateUsageExportJob(ctx context.Context, job *UsageExportJob) error {
	query := `
		INSERT INTO usage_export_jobs (
			id, configuration_id, tenant_id, status, format, data_types,
			period_start, period_end, record_count, file_size_bytes,
			storage_provider, storage_path, storage_url, checksum,
			started_at, completed_at, expires_at, error_message, retry_count,
			delivered_at, delivery_method, delivery_status, delivery_error,
			created_at, triggered_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)
	`

	_, err := r.db.ExecContext(ctx, query,
		job.ID, job.ConfigurationID, job.TenantID, job.Status, job.Format,
		pq.Array(job.DataTypes), job.PeriodStart, job.PeriodEnd,
		job.RecordCount, job.FileSizeBytes, job.StorageProvider, job.StoragePath,
		job.StorageURL, job.Checksum, job.StartedAt, job.CompletedAt, job.ExpiresAt,
		job.ErrorMessage, job.RetryCount, job.DeliveredAt, job.DeliveryMethod,
		job.DeliveryStatus, job.DeliveryError, job.CreatedAt, job.TriggeredBy,
	)

	return err
}

// GetUsageExportJob retrieves an export job by ID
func (r *ExportRepository) GetUsageExportJob(ctx context.Context, id uuid.UUID) (*UsageExportJob, error) {
	query := `
		SELECT id, configuration_id, tenant_id, status, format, data_types,
			period_start, period_end, record_count, file_size_bytes,
			storage_provider, storage_path, storage_url, checksum,
			started_at, completed_at, expires_at, error_message, retry_count,
			delivered_at, delivery_method, delivery_status, delivery_error,
			created_at, triggered_by
		FROM usage_export_jobs WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanExportJob(row)
}

// ListUsageExportJobs lists export jobs for a tenant
func (r *ExportRepository) ListUsageExportJobs(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*UsageExportJob, error) {
	query := `
		SELECT id, configuration_id, tenant_id, status, format, data_types,
			period_start, period_end, record_count, file_size_bytes,
			storage_provider, storage_path, storage_url, checksum,
			started_at, completed_at, expires_at, error_message, retry_count,
			delivered_at, delivery_method, delivery_status, delivery_error,
			created_at, triggered_by
		FROM usage_export_jobs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanExportJobRows(rows)
}

// UpdateUsageExportJobStatus updates the status of an export job
func (r *ExportRepository) UpdateUsageExportJobStatus(ctx context.Context, id uuid.UUID, status UsageExportStatus, errorMsg string) error {
	query := `UPDATE usage_export_jobs SET status = $2, error_message = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status, errorMsg)
	return err
}

// UpdateUsageExportJobProgress updates job progress metrics
func (r *ExportRepository) UpdateUsageExportJobProgress(ctx context.Context, id uuid.UUID, recordCount, fileSize int64, storagePath, storageURL, checksum string) error {
	query := `
		UPDATE usage_export_jobs SET
			record_count = $2, file_size_bytes = $3,
			storage_path = $4, storage_url = $5, checksum = $6
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, recordCount, fileSize, storagePath, storageURL, checksum)
	return err
}

// CompleteUsageExportJob marks a job as completed
func (r *ExportRepository) CompleteUsageExportJob(ctx context.Context, id uuid.UUID, storagePath, storageURL, checksum string, recordCount, fileSize int64) error {
	query := `
		UPDATE usage_export_jobs SET
			status = $2, storage_path = $3, storage_url = $4, checksum = $5,
			record_count = $6, file_size_bytes = $7, completed_at = $8
		WHERE id = $1
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		id, ExportStatusCompleted, storagePath, storageURL, checksum,
		recordCount, fileSize, now,
	)
	return err
}

// UpdateDeliveryStatus updates the delivery status of an export job
func (r *ExportRepository) UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status, errorMsg string) error {
	query := `
		UPDATE usage_export_jobs SET
			delivery_status = $2, delivery_error = $3, delivered_at = $4
		WHERE id = $1
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, id, status, errorMsg, now)
	return err
}

// GetPendingScheduledConfigs gets configurations that are due for scheduled execution
func (r *ExportRepository) GetPendingScheduledConfigs(ctx context.Context, now time.Time) ([]*UsageExportConfiguration, error) {
	query := `
		SELECT id, tenant_id, name, description, format, data_types, granularity,
			include_metadata, include_breakdown, date_range_type, function_filter,
			region_filter, outcome_filter, is_scheduled, schedule_frequency,
			schedule_day_of_month, schedule_day_of_week, schedule_hour,
			delivery_method, email_recipients, webhook_url, s3_bucket, s3_prefix,
			external_system_id, field_mapping, transform_config, is_active,
			created_by, created_at, updated_at, last_executed_at, last_export_id
		FROM usage_export_configurations
		WHERE is_scheduled = true AND is_active = true
		AND (
			last_executed_at IS NULL
			OR (
				(schedule_frequency = 'daily' AND last_executed_at < $1 - INTERVAL '1 day')
				OR (schedule_frequency = 'weekly' AND last_executed_at < $1 - INTERVAL '7 days')
				OR (schedule_frequency = 'monthly' AND last_executed_at < $1 - INTERVAL '1 month')
			)
		)
		AND (
			schedule_hour IS NULL
			OR EXTRACT(HOUR FROM $1) >= schedule_hour
		)
	`

	rows, err := r.db.QueryContext(ctx, query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanExportConfigRows(rows)
}

// ==================== External Billing Systems ====================

// CreateExternalBillingSystem creates a new external billing integration
func (r *ExportRepository) CreateExternalBillingSystem(ctx context.Context, system *ExternalBillingSystem) error {
	query := `
		INSERT INTO external_billing_systems (
			id, tenant_id, name, description, system_type, api_endpoint, auth_type,
			api_credential_key, api_credential_secret, oauth_token, oauth_refresh_token,
			oauth_expires_at, is_active, last_tested_at, last_test_status, last_test_error,
			sync_enabled, sync_frequency, sync_direction, last_sync_at, last_sync_status,
			field_mappings, value_mappings, transform_rules, webhook_secret, webhook_url,
			created_at, updated_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)
	`

	_, err := r.db.ExecContext(ctx, query,
		system.ID, system.TenantID, system.Name, system.Description,
		system.SystemType, system.APIEndpoint, system.AuthType,
		system.APICredentialKey, system.APICredentialSecret,
		system.OAuthToken, system.OAuthRefreshToken, system.OAuthExpiresAt,
		system.IsActive, system.LastTestedAt, system.LastTestStatus, system.LastTestError,
		system.SyncEnabled, system.SyncFrequency, system.SyncDirection,
		system.LastSyncAt, system.LastSyncStatus,
		system.FieldMappings, system.ValueMappings, system.TransformRules,
		system.WebhookSecret, system.WebhookURL,
		system.CreatedAt, system.UpdatedAt, system.CreatedBy,
	)

	return err
}

// GetExternalBillingSystem retrieves an external billing system by ID
func (r *ExportRepository) GetExternalBillingSystem(ctx context.Context, id uuid.UUID) (*ExternalBillingSystem, error) {
	query := `
		SELECT id, tenant_id, name, description, system_type, api_endpoint, auth_type,
			is_active, last_tested_at, last_test_status, last_test_error,
			sync_enabled, sync_frequency, sync_direction, last_sync_at, last_sync_status,
			field_mappings, value_mappings, transform_rules, webhook_url,
			created_at, updated_at, created_by
		FROM external_billing_systems WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanExternalBillingSystem(row)
}

// GetExternalBillingSystemWithCredentials retrieves system with credentials
func (r *ExportRepository) GetExternalBillingSystemWithCredentials(ctx context.Context, id uuid.UUID) (*ExternalBillingSystem, error) {
	query := `
		SELECT id, tenant_id, name, description, system_type, api_endpoint, auth_type,
			api_credential_key, api_credential_secret, oauth_token, oauth_refresh_token,
			oauth_expires_at, is_active, last_tested_at, last_test_status, last_test_error,
			sync_enabled, sync_frequency, sync_direction, last_sync_at, last_sync_status,
			field_mappings, value_mappings, transform_rules, webhook_secret, webhook_url,
			created_at, updated_at, created_by
		FROM external_billing_systems WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanExternalBillingSystemWithCredentials(row)
}

// ListExternalBillingSystems lists all external billing systems for a tenant with pagination
func (r *ExportRepository) ListExternalBillingSystems(ctx context.Context, tenantID uuid.UUID, limit, offset int, activeOnly bool) ([]*ExternalBillingSystem, error) {
	query := `
		SELECT id, tenant_id, name, description, system_type, api_endpoint, auth_type,
			is_active, last_tested_at, last_test_status, last_test_error,
			sync_enabled, sync_frequency, sync_direction, last_sync_at, last_sync_status,
			field_mappings, value_mappings, transform_rules, webhook_url,
			created_at, updated_at, created_by
		FROM external_billing_systems
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	systems, err := r.scanExternalBillingSystemRows(rows)
	if err != nil {
		return nil, err
	}

	// Filter by active if requested
	if activeOnly {
		var filtered []*ExternalBillingSystem
		for _, s := range systems {
			if s.IsActive {
				filtered = append(filtered, s)
			}
		}
		return filtered, nil
	}

	return systems, nil
}

// UpdateExternalBillingSystem updates an external billing system
func (r *ExportRepository) UpdateExternalBillingSystem(ctx context.Context, system *ExternalBillingSystem) error {
	query := `
		UPDATE external_billing_systems SET
			name = $2, description = $3, system_type = $4, api_endpoint = $5, auth_type = $6,
			is_active = $7, sync_enabled = $8, sync_frequency = $9, sync_direction = $10,
			field_mappings = $11, value_mappings = $12, transform_rules = $13,
			webhook_url = $14, updated_at = $15
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		system.ID, system.Name, system.Description, system.SystemType,
		system.APIEndpoint, system.AuthType, system.IsActive,
		system.SyncEnabled, system.SyncFrequency, system.SyncDirection,
		system.FieldMappings, system.ValueMappings, system.TransformRules,
		system.WebhookURL, time.Now(),
	)

	return err
}

// UpdateExternalBillingSystemCredentials updates only the credentials
func (r *ExportRepository) UpdateExternalBillingSystemCredentials(ctx context.Context, id uuid.UUID, credentialKey, credentialSecret, oauthToken, oauthRefresh string, oauthExpires *time.Time) error {
	query := `
		UPDATE external_billing_systems SET
			api_credential_key = $2, api_credential_secret = $3,
			oauth_token = $4, oauth_refresh_token = $5, oauth_expires_at = $6,
			updated_at = $7
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		id, credentialKey, credentialSecret, oauthToken, oauthRefresh, oauthExpires, time.Now(),
	)
	return err
}

// UpdateExternalBillingSystemTestStatus updates test connection status
func (r *ExportRepository) UpdateExternalBillingSystemTestStatus(ctx context.Context, id uuid.UUID, status, errorMsg string) error {
	query := `
		UPDATE external_billing_systems SET
			last_tested_at = $2, last_test_status = $3, last_test_error = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, time.Now(), status, errorMsg, time.Now())
	return err
}

// UpdateExternalBillingSystemSyncStatus updates sync status
func (r *ExportRepository) UpdateExternalBillingSystemSyncStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE external_billing_systems SET
			last_sync_at = $2, last_sync_status = $3, updated_at = $4
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, time.Now(), status, time.Now())
	return err
}

// DeleteExternalBillingSystem deletes an external billing system
func (r *ExportRepository) DeleteExternalBillingSystem(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM external_billing_systems WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ==================== Billing Integration Sync ====================

// CreateBillingIntegrationSync creates a new sync record
func (r *ExportRepository) CreateBillingIntegrationSync(ctx context.Context, sync *BillingIntegrationSync) error {
	query := `
		INSERT INTO billing_integration_syncs (
			id, external_system_id, tenant_id, sync_type, direction, status,
			started_at, completed_at, records_processed, records_created,
			records_updated, records_failed, records_skipped,
			error_message, error_details, external_batch_id, external_references,
			created_at, triggered_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	_, err := r.db.ExecContext(ctx, query,
		sync.ID, sync.ExternalSystemID, sync.TenantID, sync.SyncType,
		sync.Direction, sync.Status, sync.StartedAt, sync.CompletedAt,
		sync.RecordsProcessed, sync.RecordsCreated, sync.RecordsUpdated,
		sync.RecordsFailed, sync.RecordsSkipped, sync.ErrorMessage,
		sync.ErrorDetails, sync.ExternalBatchID, sync.ExternalReferences,
		sync.CreatedAt, sync.TriggeredBy,
	)

	return err
}

// GetBillingIntegrationSync retrieves a sync record by ID
func (r *ExportRepository) GetBillingIntegrationSync(ctx context.Context, id uuid.UUID) (*BillingIntegrationSync, error) {
	query := `
		SELECT id, external_system_id, tenant_id, sync_type, direction, status,
			started_at, completed_at, records_processed, records_created,
			records_updated, records_failed, records_skipped,
			error_message, error_details, external_batch_id, external_references,
			created_at, triggered_by
		FROM billing_integration_syncs WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanBillingSync(row)
}

// ListBillingIntegrationSyncs lists sync records for a tenant with optional filtering
func (r *ExportRepository) ListBillingIntegrationSyncs(ctx context.Context, tenantID uuid.UUID, systemID *uuid.UUID, status string, limit, offset int) ([]*BillingIntegrationSync, error) {
	var query string
	var rows *sql.Rows
	var err error

	if systemID != nil {
		query = `
			SELECT s.id, s.external_system_id, s.tenant_id, s.sync_type, s.direction, s.status,
				s.started_at, s.completed_at, s.records_processed, s.records_created,
				s.records_updated, s.records_failed, s.records_skipped,
				s.error_message, s.error_details, s.external_batch_id, s.external_references,
				s.created_at, s.triggered_by
			FROM billing_integration_syncs s
			JOIN external_billing_systems ebs ON s.external_system_id = ebs.id
			WHERE s.tenant_id = $1 AND s.external_system_id = $2
			ORDER BY s.created_at DESC
			LIMIT $3 OFFSET $4
		`
		rows, err = r.db.QueryContext(ctx, query, tenantID, *systemID, limit, offset)
	} else if status != "" {
		query = `
			SELECT s.id, s.external_system_id, s.tenant_id, s.sync_type, s.direction, s.status,
				s.started_at, s.completed_at, s.records_processed, s.records_created,
				s.records_updated, s.records_failed, s.records_skipped,
				s.error_message, s.error_details, s.external_batch_id, s.external_references,
				s.created_at, s.triggered_by
			FROM billing_integration_syncs s
			WHERE s.tenant_id = $1 AND s.status = $2
			ORDER BY s.created_at DESC
			LIMIT $3 OFFSET $4
		`
		rows, err = r.db.QueryContext(ctx, query, tenantID, status, limit, offset)
	} else {
		query = `
			SELECT s.id, s.external_system_id, s.tenant_id, s.sync_type, s.direction, s.status,
				s.started_at, s.completed_at, s.records_processed, s.records_created,
				s.records_updated, s.records_failed, s.records_skipped,
				s.error_message, s.error_details, s.external_batch_id, s.external_references,
				s.created_at, s.triggered_by
			FROM billing_integration_syncs s
			WHERE s.tenant_id = $1
			ORDER BY s.created_at DESC
			LIMIT $2 OFFSET $3
		`
		rows, err = r.db.QueryContext(ctx, query, tenantID, limit, offset)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanBillingSyncRows(rows)
}

// UpdateBillingIntegrationSyncStatus updates sync status
func (r *ExportRepository) UpdateBillingIntegrationSyncStatus(ctx context.Context, id uuid.UUID, status string, completedAt *time.Time) error {
	query := `
		UPDATE billing_integration_syncs
		SET status = $2, completed_at = $3
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, status, completedAt)
	return err
}

// UpdateBillingIntegrationSyncStats updates sync statistics
func (r *ExportRepository) UpdateBillingIntegrationSyncStats(ctx context.Context, id uuid.UUID, processed, created, updated, failed, skipped int64) error {
	query := `
		UPDATE billing_integration_syncs SET
			records_processed = $2, records_created = $3, records_updated = $4,
			records_failed = $5, records_skipped = $6
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, processed, created, updated, failed, skipped)
	return err
}

// ==================== Export Templates ====================

// GetUsageExportTemplate retrieves an export template by ID
func (r *ExportRepository) GetUsageExportTemplate(ctx context.Context, id uuid.UUID) (*UsageExportTemplate, error) {
	query := `
		SELECT id, name, description, category, format, data_types, granularity,
			include_metadata, include_breakdown, default_fields, field_order,
			column_headers, data_transforms, is_active, is_system, created_at, updated_at
		FROM usage_export_templates WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanExportTemplate(row)
}

// ListUsageExportTemplates lists all export templates
func (r *ExportRepository) ListUsageExportTemplates(ctx context.Context, category string) ([]*UsageExportTemplate, error) {
	var query string
	var rows *sql.Rows
	var err error

	if category != "" {
		query = `
			SELECT id, name, description, category, format, data_types, granularity,
				include_metadata, include_breakdown, default_fields, field_order,
				column_headers, data_transforms, is_active, is_system, created_at, updated_at
			FROM usage_export_templates
			WHERE category = $1 AND is_active = true
			ORDER BY is_system DESC, name ASC
		`
		rows, err = r.db.QueryContext(ctx, query, category)
	} else {
		query = `
			SELECT id, name, description, category, format, data_types, granularity,
				include_metadata, include_breakdown, default_fields, field_order,
				column_headers, data_transforms, is_active, is_system, created_at, updated_at
			FROM usage_export_templates
			WHERE is_active = true
			ORDER BY is_system DESC, name ASC
		`
		rows, err = r.db.QueryContext(ctx, query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanExportTemplateRows(rows)
}

// ==================== Helper Methods ====================

func (r *ExportRepository) scanExportConfig(row *sql.Row) (*UsageExportConfiguration, error) {
	var c UsageExportConfiguration
	var lastExecutedAt sql.NullTime
	var lastExportID uuid.NullUUID
	var externalSystemID uuid.NullUUID
	var scheduleDayOfMonth, scheduleDayOfWeek, scheduleHour sql.NullInt32

	err := row.Scan(
		&c.ID, &c.TenantID, &c.Name, &c.Description, &c.Format,
		pq.Array(&c.DataTypes), &c.Granularity, &c.IncludeMetadata, &c.IncludeBreakdown,
		&c.DateRangeType, pq.Array(&c.FunctionFilter), pq.Array(&c.RegionFilter), pq.Array(&c.OutcomeFilter),
		&c.IsScheduled, &c.ScheduleFrequency, &scheduleDayOfMonth, &scheduleDayOfWeek, &scheduleHour,
		&c.DeliveryMethod, pq.Array(&c.EmailRecipients), &c.WebhookURL, &c.S3Bucket, &c.S3Prefix,
		&externalSystemID, &c.FieldMapping, &c.TransformConfig, &c.IsActive,
		&c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &lastExecutedAt, &lastExportID,
	)
	if err != nil {
		return nil, err
	}

	if lastExecutedAt.Valid {
		c.LastExecutedAt = &lastExecutedAt.Time
	}
	if lastExportID.Valid {
		c.LastExportID = &lastExportID.UUID
	}
	if externalSystemID.Valid {
		c.ExternalSystemID = &externalSystemID.UUID
	}
	if scheduleDayOfMonth.Valid {
		v := int(scheduleDayOfMonth.Int32)
		c.ScheduleDayOfMonth = &v
	}
	if scheduleDayOfWeek.Valid {
		v := int(scheduleDayOfWeek.Int32)
		c.ScheduleDayOfWeek = &v
	}
	if scheduleHour.Valid {
		v := int(scheduleHour.Int32)
		c.ScheduleHour = &v
	}

	return &c, nil
}

func (r *ExportRepository) scanExportConfigRows(rows *sql.Rows) ([]*UsageExportConfiguration, error) {
	var configs []*UsageExportConfiguration
	for rows.Next() {
		var c UsageExportConfiguration
		var lastExecutedAt sql.NullTime
		var lastExportID uuid.NullUUID
		var externalSystemID uuid.NullUUID
		var scheduleDayOfMonth, scheduleDayOfWeek, scheduleHour sql.NullInt32

		err := rows.Scan(
			&c.ID, &c.TenantID, &c.Name, &c.Description, &c.Format,
			pq.Array(&c.DataTypes), &c.Granularity, &c.IncludeMetadata, &c.IncludeBreakdown,
			&c.DateRangeType, pq.Array(&c.FunctionFilter), pq.Array(&c.RegionFilter), pq.Array(&c.OutcomeFilter),
			&c.IsScheduled, &c.ScheduleFrequency, &scheduleDayOfMonth, &scheduleDayOfWeek, &scheduleHour,
			&c.DeliveryMethod, pq.Array(&c.EmailRecipients), &c.WebhookURL, &c.S3Bucket, &c.S3Prefix,
			&externalSystemID, &c.FieldMapping, &c.TransformConfig, &c.IsActive,
			&c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &lastExecutedAt, &lastExportID,
		)
		if err != nil {
			return nil, err
		}

		if lastExecutedAt.Valid {
			c.LastExecutedAt = &lastExecutedAt.Time
		}
		if lastExportID.Valid {
			c.LastExportID = &lastExportID.UUID
		}
		if externalSystemID.Valid {
			c.ExternalSystemID = &externalSystemID.UUID
		}
		if scheduleDayOfMonth.Valid {
			v := int(scheduleDayOfMonth.Int32)
			c.ScheduleDayOfMonth = &v
		}
		if scheduleDayOfWeek.Valid {
			v := int(scheduleDayOfWeek.Int32)
			c.ScheduleDayOfWeek = &v
		}
		if scheduleHour.Valid {
			v := int(scheduleHour.Int32)
			c.ScheduleHour = &v
		}

		configs = append(configs, &c)
	}
	return configs, rows.Err()
}

func (r *ExportRepository) scanExportJob(row *sql.Row) (*UsageExportJob, error) {
	var j UsageExportJob
	var startedAt, completedAt, expiresAt, deliveredAt sql.NullTime

	err := row.Scan(
		&j.ID, &j.ConfigurationID, &j.TenantID, &j.Status, &j.Format,
		pq.Array(&j.DataTypes), &j.PeriodStart, &j.PeriodEnd,
		&j.RecordCount, &j.FileSizeBytes, &j.StorageProvider, &j.StoragePath,
		&j.StorageURL, &j.Checksum, &startedAt, &completedAt, &expiresAt,
		&j.ErrorMessage, &j.RetryCount, &deliveredAt, &j.DeliveryMethod,
		&j.DeliveryStatus, &j.DeliveryError, &j.CreatedAt, &j.TriggeredBy,
	)
	if err != nil {
		return nil, err
	}

	if startedAt.Valid {
		j.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		j.CompletedAt = &completedAt.Time
	}
	if expiresAt.Valid {
		j.ExpiresAt = &expiresAt.Time
	}
	if deliveredAt.Valid {
		j.DeliveredAt = &deliveredAt.Time
	}

	return &j, nil
}

func (r *ExportRepository) scanExportJobRows(rows *sql.Rows) ([]*UsageExportJob, error) {
	var jobs []*UsageExportJob
	for rows.Next() {
		var j UsageExportJob
		var startedAt, completedAt, expiresAt, deliveredAt sql.NullTime

		err := rows.Scan(
			&j.ID, &j.ConfigurationID, &j.TenantID, &j.Status, &j.Format,
			pq.Array(&j.DataTypes), &j.PeriodStart, &j.PeriodEnd,
			&j.RecordCount, &j.FileSizeBytes, &j.StorageProvider, &j.StoragePath,
			&j.StorageURL, &j.Checksum, &startedAt, &completedAt, &expiresAt,
			&j.ErrorMessage, &j.RetryCount, &deliveredAt, &j.DeliveryMethod,
			&j.DeliveryStatus, &j.DeliveryError, &j.CreatedAt, &j.TriggeredBy,
		)
		if err != nil {
			return nil, err
		}

		if startedAt.Valid {
			j.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			j.CompletedAt = &completedAt.Time
		}
		if expiresAt.Valid {
			j.ExpiresAt = &expiresAt.Time
		}
		if deliveredAt.Valid {
			j.DeliveredAt = &deliveredAt.Time
		}

		jobs = append(jobs, &j)
	}
	return jobs, rows.Err()
}

func (r *ExportRepository) scanExternalBillingSystem(row *sql.Row) (*ExternalBillingSystem, error) {
	var s ExternalBillingSystem
	var lastTestedAt, lastSyncAt, oauthExpiresAt sql.NullTime
	var transformRulesJSON []byte

	err := row.Scan(
		&s.ID, &s.TenantID, &s.Name, &s.Description, &s.SystemType, &s.APIEndpoint, &s.AuthType,
		&s.IsActive, &lastTestedAt, &s.LastTestStatus, &s.LastTestError,
		&s.SyncEnabled, &s.SyncFrequency, &s.SyncDirection, &lastSyncAt, &s.LastSyncStatus,
		&s.FieldMappings, &s.ValueMappings, &transformRulesJSON, &s.WebhookURL,
		&s.CreatedAt, &s.UpdatedAt, &s.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	if lastTestedAt.Valid {
		s.LastTestedAt = &lastTestedAt.Time
	}
	if lastSyncAt.Valid {
		s.LastSyncAt = &lastSyncAt.Time
	}
	if oauthExpiresAt.Valid {
		s.OAuthExpiresAt = &oauthExpiresAt.Time
	}
	if len(transformRulesJSON) > 0 {
		json.Unmarshal(transformRulesJSON, &s.TransformRules)
	}

	return &s, nil
}

func (r *ExportRepository) scanExternalBillingSystemWithCredentials(row *sql.Row) (*ExternalBillingSystem, error) {
	var s ExternalBillingSystem
	var lastTestedAt, lastSyncAt, oauthExpiresAt sql.NullTime
	var transformRulesJSON []byte

	err := row.Scan(
		&s.ID, &s.TenantID, &s.Name, &s.Description, &s.SystemType, &s.APIEndpoint, &s.AuthType,
		&s.APICredentialKey, &s.APICredentialSecret, &s.OAuthToken, &s.OAuthRefreshToken, &oauthExpiresAt,
		&s.IsActive, &lastTestedAt, &s.LastTestStatus, &s.LastTestError,
		&s.SyncEnabled, &s.SyncFrequency, &s.SyncDirection, &lastSyncAt, &s.LastSyncStatus,
		&s.FieldMappings, &s.ValueMappings, &transformRulesJSON, &s.WebhookSecret, &s.WebhookURL,
		&s.CreatedAt, &s.UpdatedAt, &s.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	if lastTestedAt.Valid {
		s.LastTestedAt = &lastTestedAt.Time
	}
	if lastSyncAt.Valid {
		s.LastSyncAt = &lastSyncAt.Time
	}
	if oauthExpiresAt.Valid {
		s.OAuthExpiresAt = &oauthExpiresAt.Time
	}
	if len(transformRulesJSON) > 0 {
		json.Unmarshal(transformRulesJSON, &s.TransformRules)
	}

	return &s, nil
}

func (r *ExportRepository) scanExternalBillingSystemRows(rows *sql.Rows) ([]*ExternalBillingSystem, error) {
	var systems []*ExternalBillingSystem
	for rows.Next() {
		var s ExternalBillingSystem
		var lastTestedAt, lastSyncAt sql.NullTime
		var transformRulesJSON []byte

		err := rows.Scan(
			&s.ID, &s.TenantID, &s.Name, &s.Description, &s.SystemType, &s.APIEndpoint, &s.AuthType,
			&s.IsActive, &lastTestedAt, &s.LastTestStatus, &s.LastTestError,
			&s.SyncEnabled, &s.SyncFrequency, &s.SyncDirection, &lastSyncAt, &s.LastSyncStatus,
			&s.FieldMappings, &s.ValueMappings, &transformRulesJSON, &s.WebhookURL,
			&s.CreatedAt, &s.UpdatedAt, &s.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		if lastTestedAt.Valid {
			s.LastTestedAt = &lastTestedAt.Time
		}
		if lastSyncAt.Valid {
			s.LastSyncAt = &lastSyncAt.Time
		}
		if len(transformRulesJSON) > 0 {
			json.Unmarshal(transformRulesJSON, &s.TransformRules)
		}

		systems = append(systems, &s)
	}
	return systems, rows.Err()
}

func (r *ExportRepository) scanBillingSync(row *sql.Row) (*BillingIntegrationSync, error) {
	var s BillingIntegrationSync
	var startedAt, completedAt sql.NullTime
	var errorDetailsJSON []byte

	err := row.Scan(
		&s.ID, &s.ExternalSystemID, &s.TenantID, &s.SyncType, &s.Direction, &s.Status,
		&startedAt, &completedAt, &s.RecordsProcessed, &s.RecordsCreated,
		&s.RecordsUpdated, &s.RecordsFailed, &s.RecordsSkipped,
		&s.ErrorMessage, &errorDetailsJSON, &s.ExternalBatchID, &s.ExternalReferences,
		&s.CreatedAt, &s.TriggeredBy,
	)
	if err != nil {
		return nil, err
	}

	if startedAt.Valid {
		s.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		s.CompletedAt = &completedAt.Time
	}
	if len(errorDetailsJSON) > 0 {
		json.Unmarshal(errorDetailsJSON, &s.ErrorDetails)
	}

	return &s, nil
}

func (r *ExportRepository) scanBillingSyncRows(rows *sql.Rows) ([]*BillingIntegrationSync, error) {
	var syncs []*BillingIntegrationSync
	for rows.Next() {
		var s BillingIntegrationSync
		var startedAt, completedAt sql.NullTime
		var errorDetailsJSON []byte

		err := rows.Scan(
			&s.ID, &s.ExternalSystemID, &s.TenantID, &s.SyncType, &s.Direction, &s.Status,
			&startedAt, &completedAt, &s.RecordsProcessed, &s.RecordsCreated,
			&s.RecordsUpdated, &s.RecordsFailed, &s.RecordsSkipped,
			&s.ErrorMessage, &errorDetailsJSON, &s.ExternalBatchID, &s.ExternalReferences,
			&s.CreatedAt, &s.TriggeredBy,
		)
		if err != nil {
			return nil, err
		}

		if startedAt.Valid {
			s.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			s.CompletedAt = &completedAt.Time
		}
		if len(errorDetailsJSON) > 0 {
			json.Unmarshal(errorDetailsJSON, &s.ErrorDetails)
		}

		syncs = append(syncs, &s)
	}
	return syncs, rows.Err()
}

func (r *ExportRepository) scanExportTemplate(row *sql.Row) (*UsageExportTemplate, error) {
	var t UsageExportTemplate
	var dataTransformsJSON []byte

	err := row.Scan(
		&t.ID, &t.Name, &t.Description, &t.Category, &t.Format, pq.Array(&t.DataTypes),
		&t.Granularity, &t.IncludeMetadata, &t.IncludeBreakdown, pq.Array(&t.DefaultFields),
		pq.Array(&t.FieldOrder), &t.ColumnHeaders, &dataTransformsJSON,
		&t.IsActive, &t.IsSystem, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(dataTransformsJSON) > 0 {
		json.Unmarshal(dataTransformsJSON, &t.DataTransforms)
	}

	return &t, nil
}

func (r *ExportRepository) scanExportTemplateRows(rows *sql.Rows) ([]*UsageExportTemplate, error) {
	var templates []*UsageExportTemplate
	for rows.Next() {
		var t UsageExportTemplate
		var dataTransformsJSON []byte

		err := rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.Category, &t.Format, pq.Array(&t.DataTypes),
			&t.Granularity, &t.IncludeMetadata, &t.IncludeBreakdown, pq.Array(&t.DefaultFields),
			pq.Array(&t.FieldOrder), &t.ColumnHeaders, &dataTransformsJSON,
			&t.IsActive, &t.IsSystem, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(dataTransformsJSON) > 0 {
			json.Unmarshal(dataTransformsJSON, &t.DataTransforms)
		}

		templates = append(templates, &t)
	}
	return templates, rows.Err()
}

// EnsureTables creates the necessary tables if they don't exist
func (r *ExportRepository) EnsureTables(ctx context.Context) error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS usage_export_configurations (
			id UUID PRIMARY KEY,
			tenant_id UUID NOT NULL REFERENCES tenants(id),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			format VARCHAR(20) NOT NULL,
			data_types TEXT[],
			granularity VARCHAR(20),
			include_metadata BOOLEAN DEFAULT false,
			include_breakdown BOOLEAN DEFAULT false,
			date_range_type VARCHAR(30),
			function_filter UUID[],
			region_filter TEXT[],
			outcome_filter TEXT[],
			is_scheduled BOOLEAN DEFAULT false,
			schedule_frequency VARCHAR(20),
			schedule_day_of_month INTEGER,
			schedule_day_of_week INTEGER,
			schedule_hour INTEGER,
			delivery_method VARCHAR(30),
			email_recipients TEXT[],
			webhook_url TEXT,
			s3_bucket VARCHAR(255),
			s3_prefix VARCHAR(255),
			external_system_id UUID,
			field_mapping JSONB,
			transform_config JSONB,
			is_active BOOLEAN DEFAULT true,
			created_by UUID NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			last_executed_at TIMESTAMP,
			last_export_id UUID
		)`,
		`CREATE TABLE IF NOT EXISTS usage_export_jobs (
			id UUID PRIMARY KEY,
			configuration_id UUID REFERENCES usage_export_configurations(id),
			tenant_id UUID NOT NULL REFERENCES tenants(id),
			status VARCHAR(30) NOT NULL,
			format VARCHAR(20) NOT NULL,
			data_types TEXT[],
			period_start TIMESTAMP NOT NULL,
			period_end TIMESTAMP NOT NULL,
			record_count BIGINT DEFAULT 0,
			file_size_bytes BIGINT DEFAULT 0,
			storage_provider VARCHAR(30),
			storage_path TEXT,
			storage_url TEXT,
			checksum VARCHAR(64),
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			expires_at TIMESTAMP,
			error_message TEXT,
			retry_count INTEGER DEFAULT 0,
			delivered_at TIMESTAMP,
			delivery_method VARCHAR(30),
			delivery_status VARCHAR(30),
			delivery_error TEXT,
			created_at TIMESTAMP NOT NULL,
			triggered_by VARCHAR(30)
		)`,
		`CREATE TABLE IF NOT EXISTS external_billing_systems (
			id UUID PRIMARY KEY,
			tenant_id UUID NOT NULL REFERENCES tenants(id),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			system_type VARCHAR(50) NOT NULL,
			api_endpoint TEXT,
			auth_type VARCHAR(30) NOT NULL,
			api_credential_key TEXT,
			api_credential_secret TEXT,
			oauth_token TEXT,
			oauth_refresh_token TEXT,
			oauth_expires_at TIMESTAMP,
			is_active BOOLEAN DEFAULT true,
			last_tested_at TIMESTAMP,
			last_test_status VARCHAR(30),
			last_test_error TEXT,
			sync_enabled BOOLEAN DEFAULT false,
			sync_frequency VARCHAR(20),
			sync_direction VARCHAR(20),
			last_sync_at TIMESTAMP,
			last_sync_status VARCHAR(30),
			field_mappings JSONB,
			value_mappings JSONB,
			transform_rules JSONB,
			webhook_secret TEXT,
			webhook_url TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			created_by UUID NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS billing_integration_syncs (
			id UUID PRIMARY KEY,
			external_system_id UUID NOT NULL REFERENCES external_billing_systems(id),
			tenant_id UUID NOT NULL REFERENCES tenants(id),
			sync_type VARCHAR(30) NOT NULL,
			direction VARCHAR(20) NOT NULL,
			status VARCHAR(30) NOT NULL,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			records_processed BIGINT DEFAULT 0,
			records_created BIGINT DEFAULT 0,
			records_updated BIGINT DEFAULT 0,
			records_failed BIGINT DEFAULT 0,
			records_skipped BIGINT DEFAULT 0,
			error_message TEXT,
			error_details JSONB,
			external_batch_id VARCHAR(255),
			external_references JSONB,
			created_at TIMESTAMP NOT NULL,
			triggered_by VARCHAR(30)
		)`,
		`CREATE TABLE IF NOT EXISTS usage_export_templates (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			category VARCHAR(50) NOT NULL,
			format VARCHAR(20) NOT NULL,
			data_types TEXT[],
			granularity VARCHAR(20),
			include_metadata BOOLEAN DEFAULT false,
			include_breakdown BOOLEAN DEFAULT false,
			default_fields TEXT[],
			field_order TEXT[],
			column_headers JSONB,
			data_transforms JSONB,
			is_active BOOLEAN DEFAULT true,
			is_system BOOLEAN DEFAULT false,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_export_configs_tenant ON usage_export_configurations(tenant_id)`,
		`CREATE INDEX IF NOT EXISTS idx_export_jobs_config ON usage_export_jobs(configuration_id)`,
		`CREATE INDEX IF NOT EXISTS idx_export_jobs_tenant ON usage_export_jobs(tenant_id)`,
		`CREATE INDEX IF NOT EXISTS idx_export_jobs_status ON usage_export_jobs(status)`,
		`CREATE INDEX IF NOT EXISTS idx_external_billing_tenant ON external_billing_systems(tenant_id)`,
		`CREATE INDEX IF NOT EXISTS idx_billing_syncs_system ON billing_integration_syncs(external_system_id)`,
	}

	for _, table := range tables {
		if _, err := r.db.ExecContext(ctx, table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}
