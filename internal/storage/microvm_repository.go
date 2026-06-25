package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MicroVMRepository struct {
	db *sql.DB
}

func NewMicroVMRepository(db *sql.DB) *MicroVMRepository {
	return &MicroVMRepository{db: db}
}

func (r *MicroVMRepository) CreateExecution(ctx context.Context, exec *MicroVMExecution) error {
	query := `
		INSERT INTO microvm_executions (
			id, tenant_id, function_id, function_version, execution_id,
			started_at, completed_at, duration_ms, memory_mb, vcpus,
			status, outcome, error_message, network_allowed, packages_cached, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.db.ExecContext(ctx, query,
		exec.ID, exec.TenantID, exec.FunctionID, exec.FunctionVersion, exec.ExecutionID,
		exec.StartedAt, exec.CompletedAt, exec.DurationMs, exec.MemoryMB, exec.VCPUs,
		exec.Status, exec.Outcome, exec.ErrorMessage, exec.NetworkAllowed, exec.PackagesCached, exec.CreatedAt,
	)
	return err
}

func (r *MicroVMRepository) UpdateExecutionStatus(ctx context.Context, id uuid.UUID, status string, outcome, errorMsg *string, completedAt time.Time, durationMs int) error {
	query := `
		UPDATE microvm_executions
		SET status = $2, outcome = $3, error_message = $4, completed_at = $5, duration_ms = $6
		WHERE id = $1
	`
	var outcomeVal, errorVal sql.NullString
	if outcome != nil {
		outcomeVal = sql.NullString{String: *outcome, Valid: true}
	}
	if errorMsg != nil {
		errorVal = sql.NullString{String: *errorMsg, Valid: true}
	}
	_, err := r.db.ExecContext(ctx, query, id, status, outcomeVal, errorVal, completedAt, durationMs)
	return err
}

func (r *MicroVMRepository) GetExecutionByID(ctx context.Context, id uuid.UUID) (*MicroVMExecution, error) {
	query := `
		SELECT id, tenant_id, function_id, function_version, execution_id,
			started_at, completed_at, duration_ms, memory_mb, vcpus,
			status, outcome, error_message, network_allowed, packages_cached, created_at
		FROM microvm_executions WHERE id = $1
	`
	exec := &MicroVMExecution{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&exec.ID, &exec.TenantID, &exec.FunctionID, &exec.FunctionVersion, &exec.ExecutionID,
		&exec.StartedAt, &exec.CompletedAt, &exec.DurationMs, &exec.MemoryMB, &exec.VCPUs,
		&exec.Status, &exec.Outcome, &exec.ErrorMessage, &exec.NetworkAllowed, &exec.PackagesCached, &exec.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return exec, err
}

func (r *MicroVMRepository) GetExecutionsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*MicroVMExecution, error) {
	query := `
		SELECT id, tenant_id, function_id, function_version, execution_id,
			started_at, completed_at, duration_ms, memory_mb, vcpus,
			status, outcome, error_message, network_allowed, packages_cached, created_at
		FROM microvm_executions
		WHERE tenant_id = $1
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var execs []*MicroVMExecution
	for rows.Next() {
		exec := &MicroVMExecution{}
		err := rows.Scan(
			&exec.ID, &exec.TenantID, &exec.FunctionID, &exec.FunctionVersion, &exec.ExecutionID,
			&exec.StartedAt, &exec.CompletedAt, &exec.DurationMs, &exec.MemoryMB, &exec.VCPUs,
			&exec.Status, &exec.Outcome, &exec.ErrorMessage, &exec.NetworkAllowed, &exec.PackagesCached, &exec.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		execs = append(execs, exec)
	}
	return execs, rows.Err()
}

func (r *MicroVMRepository) GetRunningExecutionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*MicroVMExecution, error) {
	query := `
		SELECT id, tenant_id, function_id, function_version, execution_id,
			started_at, completed_at, duration_ms, memory_mb, vcpus,
			status, outcome, error_message, network_allowed, packages_cached, created_at
		FROM microvm_executions
		WHERE tenant_id = $1 AND status = 'running'
		ORDER BY started_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var execs []*MicroVMExecution
	for rows.Next() {
		exec := &MicroVMExecution{}
		err := rows.Scan(
			&exec.ID, &exec.TenantID, &exec.FunctionID, &exec.FunctionVersion, &exec.ExecutionID,
			&exec.StartedAt, &exec.CompletedAt, &exec.DurationMs, &exec.MemoryMB, &exec.VCPUs,
			&exec.Status, &exec.Outcome, &exec.ErrorMessage, &exec.NetworkAllowed, &exec.PackagesCached, &exec.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		execs = append(execs, exec)
	}
	return execs, rows.Err()
}

func (r *MicroVMRepository) GetUsageStats(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (*MicroVMStats, error) {
	query := `
		SELECT
			COUNT(*) as total_executions,
			COUNT(*) FILTER (WHERE status = 'running') as running_vms,
			COALESCE(AVG(duration_ms) FILTER (WHERE status = 'completed'), 0) as avg_duration_ms,
			COUNT(*) FILTER (WHERE outcome = 'success') * 100.0 / NULLIF(COUNT(*) FILTER (WHERE outcome IS NOT NULL), 0) as success_rate,
			COALESCE(SUM(duration_ms) / 1000.0, 0) as total_compute_seconds
		FROM microvm_executions
		WHERE tenant_id = $1 AND started_at >= $2 AND started_at < $3
	`
	stats := &MicroVMStats{}
	err := r.db.QueryRowContext(ctx, query, tenantID, start, end).Scan(
		&stats.TotalExecutions, &stats.RunningVMs, &stats.AvgDurationMs, &stats.SuccessRate, &stats.TotalComputeSeconds,
	)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *MicroVMRepository) GetTenantQuota(ctx context.Context, tenantID uuid.UUID) (*MicroVMTenantQuota, error) {
	query := `
		SELECT tenant_id, max_concurrent_vms, max_memory_mb, max_vcpus, max_timeout_ms,
			current_compute_usage, current_memory_usage, updated_at
		FROM microvm_tenant_quotas WHERE tenant_id = $1
	`
	quota := &MicroVMTenantQuota{}
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&quota.TenantID, &quota.MaxConcurrentVMs, &quota.MaxMemoryMB, &quota.MaxVCPUs, &quota.MaxTimeoutMs,
		&quota.CurrentComputeUsage, &quota.CurrentMemoryUsage, &quota.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return quota, err
}

func (r *MicroVMRepository) UpsertTenantQuota(ctx context.Context, quota *MicroVMTenantQuota) error {
	query := `
		INSERT INTO microvm_tenant_quotas (
			tenant_id, max_concurrent_vms, max_memory_mb, max_vcpus, max_timeout_ms,
			current_compute_usage, current_memory_usage, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (tenant_id) DO UPDATE SET
			max_concurrent_vms = EXCLUDED.max_concurrent_vms,
			max_memory_mb = EXCLUDED.max_memory_mb,
			max_vcpus = EXCLUDED.max_vcpus,
			max_timeout_ms = EXCLUDED.max_timeout_ms,
			current_compute_usage = EXCLUDED.current_compute_usage,
			current_memory_usage = EXCLUDED.current_memory_usage,
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		quota.TenantID, quota.MaxConcurrentVMs, quota.MaxMemoryMB, quota.MaxVCPUs, quota.MaxTimeoutMs,
		quota.CurrentComputeUsage, quota.CurrentMemoryUsage,
	)
	return err
}

func (r *MicroVMRepository) CreateBillingRecord(ctx context.Context, record *MicroVMBillingRecord) error {
	query := `
		INSERT INTO microvm_billing_records (
			id, tenant_id, billing_period, total_executions, total_compute_seconds,
			total_memory_seconds, avg_memory_mb, avg_vcpus, base_fee_cents,
			compute_charge_cents, memory_charge_cents, total_charge_cents, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		ON CONFLICT (tenant_id, billing_period) DO UPDATE SET
			total_executions = EXCLUDED.total_executions,
			total_compute_seconds = EXCLUDED.total_compute_seconds,
			total_memory_seconds = EXCLUDED.total_memory_seconds,
			avg_memory_mb = EXCLUDED.avg_memory_mb,
			avg_vcpus = EXCLUDED.avg_vcpus,
			base_fee_cents = EXCLUDED.base_fee_cents,
			compute_charge_cents = EXCLUDED.compute_charge_cents,
			memory_charge_cents = EXCLUDED.memory_charge_cents,
			total_charge_cents = EXCLUDED.total_charge_cents,
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		record.ID, record.TenantID, record.BillingPeriod, record.TotalExecutions, record.TotalComputeSeconds,
		record.TotalMemorySeconds, record.AvgMemoryMB, record.AvgVCPUs, record.BaseFeeCents,
		record.ComputeChargeCents, record.MemoryChargeCents, record.TotalChargeCents,
	)
	return err
}

func (r *MicroVMRepository) GetBillingRecord(ctx context.Context, tenantID uuid.UUID, billingPeriod string) (*MicroVMBillingRecord, error) {
	query := `
		SELECT id, tenant_id, billing_period, total_executions, total_compute_seconds,
			total_memory_seconds, avg_memory_mb, avg_vcpus, base_fee_cents,
			compute_charge_cents, memory_charge_cents, total_charge_cents, created_at, updated_at
		FROM microvm_billing_records
		WHERE tenant_id = $1 AND billing_period = $2
	`
	record := &MicroVMBillingRecord{}
	err := r.db.QueryRowContext(ctx, query, tenantID, billingPeriod).Scan(
		&record.ID, &record.TenantID, &record.BillingPeriod, &record.TotalExecutions, &record.TotalComputeSeconds,
		&record.TotalMemorySeconds, &record.AvgMemoryMB, &record.AvgVCPUs, &record.BaseFeeCents,
		&record.ComputeChargeCents, &record.MemoryChargeCents, &record.TotalChargeCents, &record.CreatedAt, &record.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return record, err
}

func (r *MicroVMRepository) GetBillingHistory(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*MicroVMBillingRecord, error) {
	query := `
		SELECT id, tenant_id, billing_period, total_executions, total_compute_seconds,
			total_memory_seconds, avg_memory_mb, avg_vcpus, base_fee_cents,
			compute_charge_cents, memory_charge_cents, total_charge_cents, created_at, updated_at
		FROM microvm_billing_records
		WHERE tenant_id = $1
		ORDER BY billing_period DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*MicroVMBillingRecord
	for rows.Next() {
		record := &MicroVMBillingRecord{}
		err := rows.Scan(
			&record.ID, &record.TenantID, &record.BillingPeriod, &record.TotalExecutions, &record.TotalComputeSeconds,
			&record.TotalMemorySeconds, &record.AvgMemoryMB, &record.AvgVCPUs, &record.BaseFeeCents,
			&record.ComputeChargeCents, &record.MemoryChargeCents, &record.TotalChargeCents, &record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (r *MicroVMRepository) CreateAuditLog(ctx context.Context, log *MicroVMAuditLog) error {
	query := `
		INSERT INTO microvm_audit_log (
			id, tenant_id, user_id, action, resource_type, resource_id, details, ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.TenantID, log.UserID, log.Action, log.ResourceType, log.ResourceID,
		log.Details, log.IPAddress, log.UserAgent, log.CreatedAt,
	)
	return err
}

func (r *MicroVMRepository) GetAuditLog(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*MicroVMAuditLog, error) {
	query := `
		SELECT id, tenant_id, user_id, action, resource_type, resource_id, details, ip_address, user_agent, created_at
		FROM microvm_audit_log
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*MicroVMAuditLog
	for rows.Next() {
		log := &MicroVMAuditLog{}
		err := rows.Scan(
			&log.ID, &log.TenantID, &log.UserID, &log.Action, &log.ResourceType, &log.ResourceID,
			&log.Details, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (r *MicroVMRepository) AggregateUsageForBilling(ctx context.Context, tenantID uuid.UUID, billingPeriod string) (*MicroVMBillingRecord, error) {
	startDate := billingPeriod + "-01"
	endDate := time.Date(
		time.Now().Year(), time.Now().Month()+1, 1, 0, 0, 0, 0, time.UTC,
	).Format("2006-01-02")

	if tenantID != uuid.Nil {
		startDate = billingPeriod + "-01"
		month, _ := time.Parse("2006-01", billingPeriod)
		endDate = month.AddDate(0, 1, 0).Format("2006-01-02")
	}

	query := `
		SELECT
			$1::uuid as tenant_id,
			$2 as billing_period,
			COUNT(*) as total_executions,
			COALESCE(SUM(duration_ms) / 1000.0, 0) as total_compute_seconds,
			COALESCE(SUM((duration_ms / 1000.0) * (memory_mb / 1024.0)), 0) as total_memory_seconds,
			COALESCE(AVG(memory_mb) FILTER (WHERE status = 'completed'), 0)::integer as avg_memory_mb,
			COALESCE(AVG(vcpus) FILTER (WHERE status = 'completed'), 0) as avg_vcpus
		FROM microvm_executions
		WHERE tenant_id = $1
			AND started_at >= $3::timestamptz
			AND started_at < $4::timestamptz
	`
	record := &MicroVMBillingRecord{}
	err := r.db.QueryRowContext(ctx, query, tenantID, billingPeriod, startDate, endDate).Scan(
		&record.TenantID, &record.BillingPeriod, &record.TotalExecutions,
		&record.TotalComputeSeconds, &record.TotalMemorySeconds,
		&record.AvgMemoryMB, &record.AvgVCPUs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate MicroVM usage: %w", err)
	}
	return record, nil
}

func (r *MicroVMRepository) CleanupOldExecutions(ctx context.Context, retentionDays int) (int, error) {
	query := `SELECT cleanup_microvm_executions($1)`
	var deletedCount int
	err := r.db.QueryRowContext(ctx, query, retentionDays).Scan(&deletedCount)
	return deletedCount, err
}

func (r *MicroVMRepository) GetActiveVMCount(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM microvm_executions WHERE tenant_id = $1 AND status = 'running'`
	var count int
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	return count, err
}
