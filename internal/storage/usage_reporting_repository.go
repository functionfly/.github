package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// StripeUsageReport tracks usage reported to Stripe for metered billing
type StripeUsageReport struct {
	ID                  uuid.UUID       `json:"id" db:"id"`
	TenantID            uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	PartnerID           *uuid.UUID      `json:"partner_id,omitempty" db:"partner_id"`
	SubscriptionID      string          `json:"subscription_id" db:"subscription_id"`
	SubscriptionItemID  string          `json:"subscription_item_id" db:"subscription_item_id"`
	UsageQuantity       int             `json:"usage_quantity" db:"usage_quantity"`
	UsagePeriodStart    time.Time       `json:"usage_period_start" db:"usage_period_start"`
	UsagePeriodEnd      time.Time       `json:"usage_period_end" db:"usage_period_end"`
	StripeTimestamp     int64           `json:"stripe_timestamp" db:"stripe_timestamp"`
	StripeUsageRecordID string          `json:"stripe_usage_record_id,omitempty" db:"stripe_usage_record_id"`
	Status              string          `json:"status" db:"status"` // 'pending', 'reported', 'failed', 'reconciled'
	ErrorMessage        string          `json:"error_message,omitempty" db:"error_message"`
	RetryCount          int             `json:"retry_count" db:"retry_count"`
	IdempotencyKey      string          `json:"idempotency_key" db:"idempotency_key"`
	MeterEventName      string          `json:"meter_event_name,omitempty" db:"meter_event_name"`
	Metadata            json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

// BillingUsageReconciliation tracks reconciliation between internal and Stripe usage
type BillingUsageReconciliation struct {
	ID                  uuid.UUID       `json:"id" db:"id"`
	TenantID            uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	SubscriptionID      string          `json:"subscription_id" db:"subscription_id"`
	PeriodStart         time.Time       `json:"period_start" db:"period_start"`
	PeriodEnd           time.Time       `json:"period_end" db:"period_end"`
	InternalUsageCount  int             `json:"internal_usage_count" db:"internal_usage_count"`
	InternalUsageValue  int             `json:"internal_usage_value" db:"internal_usage_value"`
	StripeReportedCount *int            `json:"stripe_reported_count,omitempty" db:"stripe_reported_count"`
	StripeReportedValue *int            `json:"stripe_reported_value,omitempty" db:"stripe_reported_value"`
	Status              string          `json:"status" db:"status"` // 'pending', 'matched', 'discrepancy', 'resolved'
	DiscrepancyAmount   *int            `json:"discrepancy_amount,omitempty" db:"discrepancy_amount"`
	DiscrepancyReason   string          `json:"discrepancy_reason,omitempty" db:"discrepancy_reason"`
	ResolvedAt          *time.Time      `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedBy          *uuid.UUID      `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolutionNotes     string          `json:"resolution_notes,omitempty" db:"resolution_notes"`
	Metadata            json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

// UsageReportingRepository handles usage reporting to Stripe
type UsageReportingRepository struct {
	db *PostgresDB
}

// NewUsageReportingRepository creates a new usage reporting repository
func NewUsageReportingRepository(db *PostgresDB) *UsageReportingRepository {
	return &UsageReportingRepository{db: db}
}

// CreateUsageReport creates a new usage report record
func (r *UsageReportingRepository) CreateUsageReport(ctx context.Context, report *StripeUsageReport) error {
	query := `
		INSERT INTO stripe_usage_reports (
			id, tenant_id, partner_id, subscription_id, subscription_item_id,
			usage_quantity, usage_period_start, usage_period_end, stripe_timestamp,
			stripe_usage_record_id, status, idempotency_key, meter_event_name,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	if report.ID == uuid.Nil {
		report.ID = uuid.New()
	}
	now := time.Now().UTC()
	report.CreatedAt = now
	report.UpdatedAt = now
	if report.Status == "" {
		report.Status = "pending"
	}
	if report.Metadata == nil {
		report.Metadata = json.RawMessage("{}")
	}

	_, err := r.db.ExecContext(ctx, query,
		report.ID, report.TenantID, report.PartnerID, report.SubscriptionID, report.SubscriptionItemID,
		report.UsageQuantity, report.UsagePeriodStart, report.UsagePeriodEnd, report.StripeTimestamp,
		report.StripeUsageRecordID, report.Status, report.IdempotencyKey, report.MeterEventName,
		report.Metadata, report.CreatedAt, report.UpdatedAt,
	)
	return err
}

// GetUsageReportByIdempotencyKey retrieves a usage report by its idempotency key
func (r *UsageReportingRepository) GetUsageReportByIdempotencyKey(ctx context.Context, idempotencyKey string) (*StripeUsageReport, error) {
	query := `
		SELECT id, tenant_id, partner_id, subscription_id, subscription_item_id,
		       usage_quantity, usage_period_start, usage_period_end, stripe_timestamp,
		       stripe_usage_record_id, status, error_message, retry_count, idempotency_key,
		       meter_event_name, metadata, created_at, updated_at
		FROM stripe_usage_reports
		WHERE idempotency_key = $1
	`
	row := r.db.QueryRowContext(ctx, query, idempotencyKey)
	return scanStripeUsageReport(row)
}

// UpdateUsageReportStatus updates the status of a usage report
func (r *UsageReportingRepository) UpdateUsageReportStatus(ctx context.Context, reportID uuid.UUID, status, stripeRecordID string) error {
	query := `
		UPDATE stripe_usage_reports SET
			status = $1,
			stripe_usage_record_id = $2,
			updated_at = $3
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, status, stripeRecordID, time.Now().UTC(), reportID)
	return err
}

// MarkUsageReportFailed marks a usage report as failed with error message
func (r *UsageReportingRepository) MarkUsageReportFailed(ctx context.Context, reportID uuid.UUID, errorMessage string) error {
	query := `
		UPDATE stripe_usage_reports SET
			status = 'failed',
			error_message = $1,
			retry_count = retry_count + 1,
			updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, errorMessage, time.Now().UTC(), reportID)
	return err
}

// GetPendingUsageReports retrieves all pending usage reports that need to be sent to Stripe
func (r *UsageReportingRepository) GetPendingUsageReports(ctx context.Context) ([]StripeUsageReport, error) {
	query := `
		SELECT id, tenant_id, partner_id, subscription_id, subscription_item_id,
		       usage_quantity, usage_period_start, usage_period_end, stripe_timestamp,
		       stripe_usage_record_id, status, error_message, retry_count, idempotency_key,
		       meter_event_name, metadata, created_at, updated_at
		FROM stripe_usage_reports
		WHERE status IN ('pending', 'failed')
		  AND retry_count < 5
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStripeUsageReportRows(rows)
}

// GetUsageReportByPeriod retrieves usage report for a specific period
func (r *UsageReportingRepository) GetUsageReportByPeriod(ctx context.Context, tenantID uuid.UUID, subscriptionItemID string, periodStart, periodEnd time.Time) (*StripeUsageReport, error) {
	query := `
		SELECT id, tenant_id, partner_id, subscription_id, subscription_item_id,
		       usage_quantity, usage_period_start, usage_period_end, stripe_timestamp,
		       stripe_usage_record_id, status, error_message, retry_count, idempotency_key,
		       meter_event_name, metadata, created_at, updated_at
		FROM stripe_usage_reports
		WHERE tenant_id = $1
		  AND subscription_item_id = $2
		  AND usage_period_start = $3
		  AND usage_period_end = $4
		  AND status IN ('reported', 'reconciled')
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, tenantID, subscriptionItemID, periodStart, periodEnd)
	return scanStripeUsageReport(row)
}

// GetTotalReportedUsageForPeriod gets the total usage already reported to Stripe for a period
func (r *UsageReportingRepository) GetTotalReportedUsageForPeriod(ctx context.Context, tenantID uuid.UUID, subscriptionItemID string, periodStart, periodEnd time.Time) (int, error) {
	var total sql.NullInt64
	query := `
		SELECT COALESCE(SUM(usage_quantity), 0) as total
		FROM stripe_usage_reports
		WHERE tenant_id = $1
		  AND subscription_item_id = $2
		  AND usage_period_start >= $3
		  AND usage_period_end <= $4
		  AND status IN ('reported', 'reconciled')
	`
	row := r.db.QueryRowContext(ctx, query, tenantID, subscriptionItemID, periodStart, periodEnd)
	err := row.Scan(&total)
	if err != nil {
		return 0, err
	}
	return int(total.Int64), nil
}

// CreateReconciliation creates a reconciliation record
func (r *UsageReportingRepository) CreateReconciliation(ctx context.Context, rec *BillingUsageReconciliation) error {
	query := `
		INSERT INTO billing_usage_reconciliation (
			id, tenant_id, subscription_id, period_start, period_end,
			internal_usage_count, internal_usage_value, status, metadata,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	if rec.ID == uuid.Nil {
		rec.ID = uuid.New()
	}
	now := time.Now().UTC()
	rec.CreatedAt = now
	rec.UpdatedAt = now
	if rec.Status == "" {
		rec.Status = "pending"
	}
	if rec.Metadata == nil {
		rec.Metadata = json.RawMessage("{}")
	}

	_, err := r.db.ExecContext(ctx, query,
		rec.ID, rec.TenantID, rec.SubscriptionID, rec.PeriodStart, rec.PeriodEnd,
		rec.InternalUsageCount, rec.InternalUsageValue, rec.Status, rec.Metadata,
		rec.CreatedAt, rec.UpdatedAt,
	)
	return err
}

// UpdateReconciliationWithStripeData updates reconciliation with Stripe reported data
func (r *UsageReportingRepository) UpdateReconciliationWithStripeData(ctx context.Context, recID uuid.UUID, stripeCount, stripeValue int, status string) error {
	query := `
		UPDATE billing_usage_reconciliation SET
			stripe_reported_count = $1,
			stripe_reported_value = $2,
			status = $3,
			updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query, stripeCount, stripeValue, status, time.Now().UTC(), recID)
	return err
}

// GetReconciliationByPeriod retrieves reconciliation for a specific period
func (r *UsageReportingRepository) GetReconciliationByPeriod(ctx context.Context, tenantID uuid.UUID, subscriptionID string, periodStart, periodEnd time.Time) (*BillingUsageReconciliation, error) {
	query := `
		SELECT id, tenant_id, subscription_id, period_start, period_end,
		       internal_usage_count, internal_usage_value, stripe_reported_count,
		       stripe_reported_value, status, discrepancy_amount, discrepancy_reason,
		       resolved_at, resolved_by, resolution_notes, metadata, created_at, updated_at
		FROM billing_usage_reconciliation
		WHERE tenant_id = $1
		  AND subscription_id = $2
		  AND period_start = $3
		  AND period_end = $4
	`
	row := r.db.QueryRowContext(ctx, query, tenantID, subscriptionID, periodStart, periodEnd)
	return scanBillingUsageReconciliation(row)
}

// ResolveReconciliation marks a reconciliation as resolved
func (r *UsageReportingRepository) ResolveReconciliation(ctx context.Context, recID uuid.UUID, resolvedBy uuid.UUID, notes string) error {
	query := `
		UPDATE billing_usage_reconciliation SET
			status = 'resolved',
			resolved_at = $1,
			resolved_by = $2,
			resolution_notes = $3,
			updated_at = $1
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, time.Now().UTC(), resolvedBy, notes, recID)
	return err
}

// GetReconciliationsNeedingReview retrieves reconciliations with discrepancies
func (r *UsageReportingRepository) GetReconciliationsNeedingReview(ctx context.Context, limit int) ([]BillingUsageReconciliation, error) {
	query := `
		SELECT id, tenant_id, subscription_id, period_start, period_end,
		       internal_usage_count, internal_usage_value, stripe_reported_count,
		       stripe_reported_value, status, discrepancy_amount, discrepancy_reason,
		       resolved_at, resolved_by, resolution_notes, metadata, created_at, updated_at
		FROM billing_usage_reconciliation
		WHERE status IN ('pending', 'discrepancy')
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBillingUsageReconciliationRows(rows)
}

// GenerateIdempotencyKey generates a unique idempotency key for usage reporting
func GenerateIdempotencyKey(tenantID uuid.UUID, subscriptionItemID string, timestamp int64) string {
	return fmt.Sprintf("%s_%s_%d", tenantID.String(), subscriptionItemID, timestamp)
}

// scanStripeUsageReport scans a StripeUsageReport from a database row
func scanStripeUsageReport(row *sql.Row) (*StripeUsageReport, error) {
	var r StripeUsageReport
	err := row.Scan(
		&r.ID, &r.TenantID, &r.PartnerID, &r.SubscriptionID, &r.SubscriptionItemID,
		&r.UsageQuantity, &r.UsagePeriodStart, &r.UsagePeriodEnd, &r.StripeTimestamp,
		&r.StripeUsageRecordID, &r.Status, &r.ErrorMessage, &r.RetryCount, &r.IdempotencyKey,
		&r.MeterEventName, &r.Metadata, &r.CreatedAt, &r.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// scanStripeUsageReportRows scans multiple StripeUsageReport rows
func scanStripeUsageReportRows(rows *sql.Rows) ([]StripeUsageReport, error) {
	var reports []StripeUsageReport
	for rows.Next() {
		var r StripeUsageReport
		err := rows.Scan(
			&r.ID, &r.TenantID, &r.PartnerID, &r.SubscriptionID, &r.SubscriptionItemID,
			&r.UsageQuantity, &r.UsagePeriodStart, &r.UsagePeriodEnd, &r.StripeTimestamp,
			&r.StripeUsageRecordID, &r.Status, &r.ErrorMessage, &r.RetryCount, &r.IdempotencyKey,
			&r.MeterEventName, &r.Metadata, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	return reports, rows.Err()
}

// scanBillingUsageReconciliation scans a BillingUsageReconciliation from a database row
func scanBillingUsageReconciliation(row *sql.Row) (*BillingUsageReconciliation, error) {
	var rec BillingUsageReconciliation
	err := row.Scan(
		&rec.ID, &rec.TenantID, &rec.SubscriptionID, &rec.PeriodStart, &rec.PeriodEnd,
		&rec.InternalUsageCount, &rec.InternalUsageValue, &rec.StripeReportedCount,
		&rec.StripeReportedValue, &rec.Status, &rec.DiscrepancyAmount, &rec.DiscrepancyReason,
		&rec.ResolvedAt, &rec.ResolvedBy, &rec.ResolutionNotes, &rec.Metadata,
		&rec.CreatedAt, &rec.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// scanBillingUsageReconciliationRows scans multiple BillingUsageReconciliation rows
func scanBillingUsageReconciliationRows(rows *sql.Rows) ([]BillingUsageReconciliation, error) {
	var recs []BillingUsageReconciliation
	for rows.Next() {
		var rec BillingUsageReconciliation
		err := rows.Scan(
			&rec.ID, &rec.TenantID, &rec.SubscriptionID, &rec.PeriodStart, &rec.PeriodEnd,
			&rec.InternalUsageCount, &rec.InternalUsageValue, &rec.StripeReportedCount,
			&rec.StripeReportedValue, &rec.Status, &rec.DiscrepancyAmount, &rec.DiscrepancyReason,
			&rec.ResolvedAt, &rec.ResolvedBy, &rec.ResolutionNotes, &rec.Metadata,
			&rec.CreatedAt, &rec.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		recs = append(recs, rec)
	}
	return recs, rows.Err()
}
