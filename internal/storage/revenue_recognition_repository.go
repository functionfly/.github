package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RevenueRecognitionRepository struct {
	db *PostgresDB
}

func NewRevenueRecognitionRepository(db *PostgresDB) *RevenueRecognitionRepository {
	return &RevenueRecognitionRepository{db: db}
}

func (r *RevenueRecognitionRepository) CreatePerformanceObligation(ctx context.Context, po *PerformanceObligation) error {
	if po.ID == uuid.Nil {
		po.ID = uuid.New()
	}
	po.CreatedAt = time.Now()
	po.UpdatedAt = time.Now()

	var milestonesJSON []byte
	if po.Milestones != nil {
		milestonesJSON = po.Milestones
	} else {
		milestonesJSON, _ = json.Marshal([]map[string]interface{}{})
	}

	query := `
		INSERT INTO performance_obligations
		(id, tenant_id, invoice_id, name, description, type,
		 transaction_price_cents, allocated_price_cents, ssp_cents, ssp_currency, ssp_basis,
		 recognition_method, recognition_start_date, recognition_end_date, delivery_pattern, milestones,
		 billable_period_start, billable_period_end, is_delivered, delivered_at,
		 is_fully_recognized, fully_recognized_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)`

	_, err := r.db.ExecContext(ctx, query,
		po.ID, po.TenantID, po.InvoiceID, po.Name, po.Description, po.Type,
		po.TransactionPriceCents, po.AllocatedPriceCents, po.SSPCents, po.SSPCurrency, po.SSPBasis,
		po.RecognitionMethod, po.RecognitionStartDate, po.RecognitionEndDate, po.DeliveryPattern, milestonesJSON,
		po.BillablePeriodStart, po.BillablePeriodEnd, po.IsDelivered, po.DeliveredAt,
		po.IsFullyRecognized, po.FullyRecognizedAt, po.CreatedAt, po.UpdatedAt,
	)
	return err
}

func (r *RevenueRecognitionRepository) GetPerformanceObligationByID(ctx context.Context, id uuid.UUID) (*PerformanceObligation, error) {
	query := `
		SELECT id, tenant_id, invoice_id, name, description, type,
			transaction_price_cents, allocated_price_cents, ssp_cents, ssp_currency, ssp_basis,
			recognition_method, recognition_start_date, recognition_end_date, delivery_pattern, milestones,
			billable_period_start, billable_period_end, is_delivered, delivered_at,
			is_fully_recognized, fully_recognized_at, created_at, updated_at
		FROM performance_obligations WHERE id = $1`

	po := &PerformanceObligation{}
	var milestonesJSON []byte
	var desc, sspBasis, deliveryPattern sql.NullString
	var recognitionEndDate, deliveredAt, fullyRecognizedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&po.ID, &po.TenantID, &po.InvoiceID, &po.Name, &desc, &po.Type,
		&po.TransactionPriceCents, &po.AllocatedPriceCents, &po.SSPCents, &po.SSPCurrency, &sspBasis,
		&po.RecognitionMethod, &po.RecognitionStartDate, &recognitionEndDate, &deliveryPattern, &milestonesJSON,
		&po.BillablePeriodStart, &po.BillablePeriodEnd, &po.IsDelivered, &deliveredAt,
		&po.IsFullyRecognized, &fullyRecognizedAt, &po.CreatedAt, &po.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if desc.Valid {
		po.Description = desc.String
	}
	if sspBasis.Valid {
		po.SSPBasis = sspBasis.String
	}
	if deliveryPattern.Valid {
		po.DeliveryPattern = deliveryPattern.String
	}
	if recognitionEndDate.Valid {
		po.RecognitionEndDate = &recognitionEndDate.Time
	}
	if deliveredAt.Valid {
		po.DeliveredAt = &deliveredAt.Time
	}
	if fullyRecognizedAt.Valid {
		po.FullyRecognizedAt = &fullyRecognizedAt.Time
	}
	if milestonesJSON != nil {
		_ = json.Unmarshal(milestonesJSON, &po.Milestones)
	}

	return po, nil
}

func (r *RevenueRecognitionRepository) GetPerformanceObligationsByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]*PerformanceObligation, error) {
	query := `
		SELECT id, tenant_id, invoice_id, name, description, type,
			transaction_price_cents, allocated_price_cents, ssp_cents, ssp_currency, ssp_basis,
			recognition_method, recognition_start_date, recognition_end_date, delivery_pattern, milestones,
			billable_period_start, billable_period_end, is_delivered, delivered_at,
			is_fully_recognized, fully_recognized_at, created_at, updated_at
		FROM performance_obligations WHERE invoice_id = $1
		ORDER BY created_at`

	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var obligations []*PerformanceObligation
	for rows.Next() {
		po := &PerformanceObligation{}
		var milestonesJSON []byte
		var desc, sspBasis, deliveryPattern sql.NullString
		var recognitionEndDate, deliveredAt, fullyRecognizedAt sql.NullTime

		err := rows.Scan(
			&po.ID, &po.TenantID, &po.InvoiceID, &po.Name, &desc, &po.Type,
			&po.TransactionPriceCents, &po.AllocatedPriceCents, &po.SSPCents, &po.SSPCurrency, &sspBasis,
			&po.RecognitionMethod, &po.RecognitionStartDate, &recognitionEndDate, &deliveryPattern, &milestonesJSON,
			&po.BillablePeriodStart, &po.BillablePeriodEnd, &po.IsDelivered, &deliveredAt,
			&po.IsFullyRecognized, &fullyRecognizedAt, &po.CreatedAt, &po.UpdatedAt,
		)
		if err != nil {
			continue
		}
		if desc.Valid {
			po.Description = desc.String
		}
		if sspBasis.Valid {
			po.SSPBasis = sspBasis.String
		}
		if deliveryPattern.Valid {
			po.DeliveryPattern = deliveryPattern.String
		}
		if recognitionEndDate.Valid {
			po.RecognitionEndDate = &recognitionEndDate.Time
		}
		if deliveredAt.Valid {
			po.DeliveredAt = &deliveredAt.Time
		}
		if fullyRecognizedAt.Valid {
			po.FullyRecognizedAt = &fullyRecognizedAt.Time
		}
		if milestonesJSON != nil {
			_ = json.Unmarshal(milestonesJSON, &po.Milestones)
		}
		obligations = append(obligations, po)
	}
	return obligations, rows.Err()
}

func (r *RevenueRecognitionRepository) UpdatePerformanceObligationDeliveryStatus(ctx context.Context, id uuid.UUID, isDelivered bool) error {
	query := `
		UPDATE performance_obligations
		SET is_delivered = $2,
			delivered_at = CASE WHEN $2 = true THEN NOW() ELSE delivered_at END,
			is_fully_recognized = CASE WHEN $2 = true THEN false ELSE is_fully_recognized END,
			fully_recognized_at = CASE WHEN $2 = false THEN NULL ELSE fully_recognized_at END,
			updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, isDelivered)
	return err
}

func (r *RevenueRecognitionRepository) CreateContractAsset(ctx context.Context, ca *ContractAsset) error {
	if ca.ID == uuid.Nil {
		ca.ID = uuid.New()
	}
	ca.CreatedAt = time.Now()
	ca.UpdatedAt = time.Now()

	query := `
		INSERT INTO contract_assets
		(id, tenant_id, invoice_id, customer_id, asset_type, amount_cents, currency,
		 description, reporting_period, status, reduced_amount_cents, is_reversed, reversed_at,
		 reduction_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := r.db.ExecContext(ctx, query,
		ca.ID, ca.TenantID, ca.InvoiceID, ca.CustomerID, ca.AssetType, ca.AmountCents, ca.Currency,
		ca.Description, ca.ReportingPeriod, ca.Status, ca.ReducedAmountCents, ca.IsReversed, ca.ReversedAt,
		ca.ReductionReason, ca.CreatedAt, ca.UpdatedAt,
	)
	return err
}

func (r *RevenueRecognitionRepository) GetContractAssetsByTenant(ctx context.Context, tenantID uuid.UUID, period string) ([]*ContractAsset, error) {
	query := `
		SELECT id, tenant_id, invoice_id, customer_id, asset_type, amount_cents, currency,
			description, reporting_period, status, reduced_amount_cents, is_reversed, reversed_at,
			reduction_reason, created_at, updated_at
		FROM contract_assets
		WHERE tenant_id = $1 AND reporting_period = $2
		ORDER BY created_at`

	rows, err := r.db.QueryContext(ctx, query, tenantID, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*ContractAsset
	for rows.Next() {
		ca := &ContractAsset{}
		var invoiceID sql.NullString
		var desc, reductionReason sql.NullString
		var reversedAt sql.NullTime

		err := rows.Scan(
			&ca.ID, &ca.TenantID, &invoiceID, &ca.CustomerID, &ca.AssetType, &ca.AmountCents, &ca.Currency,
			&desc, &ca.ReportingPeriod, &ca.Status, &ca.ReducedAmountCents, &ca.IsReversed, &reversedAt,
			&reductionReason, &ca.CreatedAt, &ca.UpdatedAt,
		)
		if err != nil {
			continue
		}
		if invoiceID.Valid {
			id, _ := uuid.Parse(invoiceID.String)
			ca.InvoiceID = &id
		}
		if desc.Valid {
			ca.Description = desc.String
		}
		if reductionReason.Valid {
			ca.ReductionReason = reductionReason.String
		}
		if reversedAt.Valid {
			ca.ReversedAt = &reversedAt.Time
		}
		assets = append(assets, ca)
	}
	return assets, rows.Err()
}

func (r *RevenueRecognitionRepository) GetContractAssetSummary(ctx context.Context, tenantID uuid.UUID, period string) (opening, closing int, err error) {
	openingQuery := `
		SELECT COALESCE(SUM(CASE WHEN asset_type = 'contract_asset' THEN amount_cents ELSE -amount_cents END), 0)
		FROM contract_assets
		WHERE tenant_id = $1 AND reporting_period < $2`

	var openingBalance int
	err = r.db.QueryRowContext(ctx, openingQuery, tenantID, period).Scan(&openingBalance)
	if err != nil {
		return 0, 0, err
	}

	closingQuery := `
		SELECT COALESCE(SUM(CASE WHEN asset_type = 'contract_asset' THEN amount_cents ELSE -amount_cents END), 0)
		FROM contract_assets
		WHERE tenant_id = $1 AND reporting_period = $2`

	var closingBalance int
	err = r.db.QueryRowContext(ctx, closingQuery, tenantID, period).Scan(&closingBalance)
	if err != nil {
		return 0, 0, err
	}

	return openingBalance, closingBalance, nil
}

func (r *RevenueRecognitionRepository) CreateRecognitionSchedule(ctx context.Context, schedule *RevenueRecognitionSchedule) error {
	if schedule.ID == uuid.Nil {
		schedule.ID = uuid.New()
	}
	schedule.CreatedAt = time.Now()
	schedule.UpdatedAt = time.Now()

	query := `
		INSERT INTO revenue_recognition_schedules
		(id, tenant_id, invoice_id, performance_obligation_id, recognition_month,
		 period_start_date, period_end_date, allocated_amount_cents, recognized_amount_cents,
		 deferred_amount_cents, revenue_type, is_recognized, recognized_at, original_total_cents,
		 is_adjustment, adjustment_reason, previous_schedule_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`

	_, err := r.db.ExecContext(ctx, query,
		schedule.ID, schedule.TenantID, schedule.InvoiceID, schedule.PerformanceObligationID, schedule.RecognitionMonth,
		schedule.PeriodStartDate, schedule.PeriodEndDate, schedule.AllocatedAmountCents, schedule.RecognizedAmountCents,
		schedule.DeferredAmountCents, schedule.RevenueType, schedule.IsRecognized, schedule.RecognizedAt, schedule.OriginalTotalCents,
		schedule.IsAdjustment, schedule.AdjustmentReason, schedule.PreviousScheduleID, schedule.CreatedAt, schedule.UpdatedAt,
	)
	return err
}

func (r *RevenueRecognitionRepository) GetRecognitionSchedulesByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]*RevenueRecognitionSchedule, error) {
	query := `
		SELECT id, tenant_id, invoice_id, performance_obligation_id, recognition_month,
			period_start_date, period_end_date, allocated_amount_cents, recognized_amount_cents,
			deferred_amount_cents, revenue_type, is_recognized, recognized_at, original_total_cents,
			is_adjustment, adjustment_reason, previous_schedule_id, created_at, updated_at
		FROM revenue_recognition_schedules
		WHERE invoice_id = $1
		ORDER BY recognition_month`

	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*RevenueRecognitionSchedule
	for rows.Next() {
		s := &RevenueRecognitionSchedule{}
		var poID, prevScheduleID sql.NullString
		var recognizedAt sql.NullTime
		var adjustmentReason sql.NullString

		err := rows.Scan(
			&s.ID, &s.TenantID, &s.InvoiceID, &poID, &s.RecognitionMonth,
			&s.PeriodStartDate, &s.PeriodEndDate, &s.AllocatedAmountCents, &s.RecognizedAmountCents,
			&s.DeferredAmountCents, &s.RevenueType, &s.IsRecognized, &recognizedAt, &s.OriginalTotalCents,
			&s.IsAdjustment, &adjustmentReason, &prevScheduleID, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			continue
		}
		if poID.Valid {
			id, _ := uuid.Parse(poID.String)
			s.PerformanceObligationID = id
		}
		if recognizedAt.Valid {
			s.RecognizedAt = &recognizedAt.Time
		}
		if adjustmentReason.Valid {
			s.AdjustmentReason = adjustmentReason.String
		}
		if prevScheduleID.Valid {
			id, _ := uuid.Parse(prevScheduleID.String)
			s.PreviousScheduleID = &id
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

func (r *RevenueRecognitionRepository) GetRecognitionSchedulesByPeriod(ctx context.Context, tenantID uuid.UUID, period string) ([]*RevenueRecognitionSchedule, error) {
	query := `
		SELECT id, tenant_id, invoice_id, performance_obligation_id, recognition_month,
			period_start_date, period_end_date, allocated_amount_cents, recognized_amount_cents,
			deferred_amount_cents, revenue_type, is_recognized, recognized_at, original_total_cents,
			is_adjustment, adjustment_reason, previous_schedule_id, created_at, updated_at
		FROM revenue_recognition_schedules
		WHERE tenant_id = $1 AND recognition_month = $2
		ORDER BY invoice_id, recognition_month`

	rows, err := r.db.QueryContext(ctx, query, tenantID, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*RevenueRecognitionSchedule
	for rows.Next() {
		s := &RevenueRecognitionSchedule{}
		var poID, prevScheduleID sql.NullString
		var recognizedAt sql.NullTime
		var adjustmentReason sql.NullString

		err := rows.Scan(
			&s.ID, &s.TenantID, &s.InvoiceID, &poID, &s.RecognitionMonth,
			&s.PeriodStartDate, &s.PeriodEndDate, &s.AllocatedAmountCents, &s.RecognizedAmountCents,
			&s.DeferredAmountCents, &s.RevenueType, &s.IsRecognized, &recognizedAt, &s.OriginalTotalCents,
			&s.IsAdjustment, &adjustmentReason, &prevScheduleID, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			continue
		}
		if poID.Valid {
			id, _ := uuid.Parse(poID.String)
			s.PerformanceObligationID = id
		}
		if recognizedAt.Valid {
			s.RecognizedAt = &recognizedAt.Time
		}
		if adjustmentReason.Valid {
			s.AdjustmentReason = adjustmentReason.String
		}
		if prevScheduleID.Valid {
			id, _ := uuid.Parse(prevScheduleID.String)
			s.PreviousScheduleID = &id
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

func (r *RevenueRecognitionRepository) MarkScheduleRecognized(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE revenue_recognition_schedules
		SET is_recognized = true, recognized_at = NOW(), updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *RevenueRecognitionRepository) CreateRecognitionEvent(ctx context.Context, event *RevenueRecognitionEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	event.CreatedAt = time.Now()

	var metadataJSON []byte
	if event.Metadata != nil {
		metadataJSON, _ = json.Marshal(event.Metadata)
	} else {
		metadataJSON = []byte("{}")
	}

	query := `
		INSERT INTO revenue_recognition_events
		(id, tenant_id, invoice_id, event_type, revenue_type, gross_amount_cents,
		 deferred_amount_cents, recognized_amount_cents, event_date, reporting_period,
		 performance_obligation_id, schedule_id, previous_invoice_id, modification_type,
		 description, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`

	_, err := r.db.ExecContext(ctx, query,
		event.ID, event.TenantID, event.InvoiceID, event.EventType, event.RevenueType, event.GrossAmountCents,
		event.DeferredAmountCents, event.RecognizedAmountCents, event.EventDate, event.ReportingPeriod,
		event.PerformanceObligationID, event.ScheduleID, event.PreviousInvoiceID, event.ModificationType,
		event.Description, metadataJSON, event.CreatedAt,
	)
	return err
}

func (r *RevenueRecognitionRepository) GetDeferredRevenueSummary(ctx context.Context, tenantID uuid.UUID, period string) (*DeferredRevenueSummary, error) {
	summary := &DeferredRevenueSummary{
		TenantID: tenantID,
		ReportingPeriod: period,
	}

	openingQuery := `
		SELECT COALESCE(SUM(deferred_amount_cents - recognized_amount_cents), 0)
		FROM revenue_recognition_schedules
		WHERE tenant_id = $1 AND recognition_month < $2`

	err := r.db.QueryRowContext(ctx, openingQuery, tenantID, period).Scan(&summary.OpeningBalanceCents)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	newDefQuery := `
		SELECT COALESCE(SUM(allocated_amount_cents - recognized_amount_cents), 0)
		FROM revenue_recognition_schedules
		WHERE tenant_id = $1 AND recognition_month = $2 AND is_recognized = false`

	err = r.db.QueryRowContext(ctx, newDefQuery, tenantID, period).Scan(&summary.NewDeferredCents)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	recQuery := `
		SELECT COALESCE(SUM(recognized_amount_cents), 0)
		FROM revenue_recognition_schedules
		WHERE tenant_id = $1 AND recognition_month = $2 AND is_recognized = true`

	err = r.db.QueryRowContext(ctx, recQuery, tenantID, period).Scan(&summary.RecognizedCents)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	summary.ClosingBalanceCents = summary.OpeningBalanceCents + summary.NewDeferredCents - summary.RecognizedCents

	subDefQuery := `
		SELECT COALESCE(SUM(allocated_amount_cents - recognized_amount_cents), 0)
		FROM revenue_recognition_schedules
		WHERE tenant_id = $1 AND recognition_month = $2 AND revenue_type = 'subscription'`

	err = r.db.QueryRowContext(ctx, subDefQuery, tenantID, period).Scan(&summary.SubscriptionDeferredCents)
	if err != nil && err != sql.ErrNoRows {
		summary.SubscriptionDeferredCents = 0
	}

	return summary, nil
}

func (r *RevenueRecognitionRepository) GetRecognizedRevenueSummary(ctx context.Context, tenantID uuid.UUID, period string) (*RecognizedRevenueSummary, error) {
	summary := &RecognizedRevenueSummary{
		TenantID: tenantID,
		ReportingPeriod: period,
	}

	query := `
		SELECT
			COALESCE(SUM(CASE WHEN revenue_type = 'subscription' THEN recognized_amount_cents ELSE 0 END), 0) as sub_rev,
			COALESCE(SUM(CASE WHEN revenue_type = 'usage' THEN recognized_amount_cents ELSE 0 END), 0) as usage_rev,
			COALESCE(SUM(CASE WHEN revenue_type = 'one_time' THEN recognized_amount_cents ELSE 0 END), 0) as ot_rev,
			COALESCE(SUM(recognized_amount_cents), 0) as total_rev
		FROM revenue_recognition_schedules
		WHERE tenant_id = $1 AND recognition_month = $2 AND is_recognized = true`

	err := r.db.QueryRowContext(ctx, query, tenantID, period).Scan(
		&summary.SubscriptionRevenueCents,
		&summary.UsageRevenueCents,
		&summary.OneTimeRevenueCents,
		&summary.TotalRevenueCents,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return summary, nil
}

func (r *RevenueRecognitionRepository) GetUnbilledRevenue(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(SUM(allocated_amount_cents - recognized_amount_cents), 0)
		FROM revenue_recognition_schedules
		WHERE tenant_id = $1 AND is_recognized = false`

	var total int
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&total)
	return total, err
}

func (r *RevenueRecognitionRepository) GetRemainingPerformanceObligations(ctx context.Context, invoiceID uuid.UUID) ([]*PerformanceObligation, error) {
	query := `
		SELECT id, tenant_id, invoice_id, name, description, type,
			transaction_price_cents, allocated_price_cents, ssp_cents, ssp_currency, ssp_basis,
			recognition_method, recognition_start_date, recognition_end_date, delivery_pattern, milestones,
			billable_period_start, billable_period_end, is_delivered, delivered_at,
			is_fully_recognized, fully_recognized_at, created_at, updated_at
		FROM performance_obligations
		WHERE invoice_id = $1 AND is_fully_recognized = false
		ORDER BY created_at`

	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var obligations []*PerformanceObligation
	for rows.Next() {
		po := &PerformanceObligation{}
		var milestonesJSON []byte
		var desc, sspBasis, deliveryPattern sql.NullString
		var recognitionEndDate, deliveredAt, fullyRecognizedAt sql.NullTime

		err := rows.Scan(
			&po.ID, &po.TenantID, &po.InvoiceID, &po.Name, &desc, &po.Type,
			&po.TransactionPriceCents, &po.AllocatedPriceCents, &po.SSPCents, &po.SSPCurrency, &sspBasis,
			&po.RecognitionMethod, &po.RecognitionStartDate, &recognitionEndDate, &deliveryPattern, &milestonesJSON,
			&po.BillablePeriodStart, &po.BillablePeriodEnd, &po.IsDelivered, &deliveredAt,
			&po.IsFullyRecognized, &fullyRecognizedAt, &po.CreatedAt, &po.UpdatedAt,
		)
		if err != nil {
			continue
		}
		if desc.Valid {
			po.Description = desc.String
		}
		if sspBasis.Valid {
			po.SSPBasis = sspBasis.String
		}
		if deliveryPattern.Valid {
			po.DeliveryPattern = deliveryPattern.String
		}
		if recognitionEndDate.Valid {
			po.RecognitionEndDate = &recognitionEndDate.Time
		}
		if deliveredAt.Valid {
			po.DeliveredAt = &deliveredAt.Time
		}
		if fullyRecognizedAt.Valid {
			po.FullyRecognizedAt = &fullyRecognizedAt.Time
		}
		if milestonesJSON != nil {
			_ = json.Unmarshal(milestonesJSON, &po.Milestones)
		}
		obligations = append(obligations, po)
	}
	return obligations, rows.Err()
}

func (r *RevenueRecognitionRepository) GetDB() *PostgresDB {
	return r.db
}

func (r *RevenueRecognitionRepository) GetScheduleByID(ctx context.Context, id uuid.UUID) (*RevenueRecognitionSchedule, error) {
	query := `
		SELECT id, tenant_id, invoice_id, performance_obligation_id, recognition_month,
			period_start_date, period_end_date, allocated_amount_cents, recognized_amount_cents,
			deferred_amount_cents, revenue_type, is_recognized, recognized_at, original_total_cents,
			is_adjustment, adjustment_reason, previous_schedule_id, created_at, updated_at
		FROM revenue_recognition_schedules
		WHERE id = $1`

	schedule := &RevenueRecognitionSchedule{}
	var poID, prevScheduleID sql.NullString
	var recognizedAt sql.NullTime
	var adjustmentReason sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&schedule.ID, &schedule.TenantID, &schedule.InvoiceID, &poID, &schedule.RecognitionMonth,
		&schedule.PeriodStartDate, &schedule.PeriodEndDate, &schedule.AllocatedAmountCents, &schedule.RecognizedAmountCents,
		&schedule.DeferredAmountCents, &schedule.RevenueType, &schedule.IsRecognized, &recognizedAt, &schedule.OriginalTotalCents,
		&schedule.IsAdjustment, &adjustmentReason, &prevScheduleID, &schedule.CreatedAt, &schedule.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if poID.Valid {
		parsedID, _ := uuid.Parse(poID.String)
		schedule.PerformanceObligationID = parsedID
	}
	if recognizedAt.Valid {
		schedule.RecognizedAt = &recognizedAt.Time
	}
	if adjustmentReason.Valid {
		schedule.AdjustmentReason = adjustmentReason.String
	}
	if prevScheduleID.Valid {
		parsedID, _ := uuid.Parse(prevScheduleID.String)
		schedule.PreviousScheduleID = &parsedID
	}

	return schedule, nil
}

func (r *RevenueRecognitionRepository) GetAllUnrecognizedSchedules(ctx context.Context, period string) ([]*RevenueRecognitionSchedule, error) {
	query := `
		SELECT id, tenant_id, invoice_id, performance_obligation_id, recognition_month,
			period_start_date, period_end_date, allocated_amount_cents, recognized_amount_cents,
			deferred_amount_cents, revenue_type, is_recognized, recognized_at, original_total_cents,
			is_adjustment, adjustment_reason, previous_schedule_id, created_at, updated_at
		FROM revenue_recognition_schedules
		WHERE recognition_month <= $1 AND is_recognized = false
		ORDER BY recognition_month`

	rows, err := r.db.QueryContext(ctx, query, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*RevenueRecognitionSchedule
	for rows.Next() {
		s := &RevenueRecognitionSchedule{}
		var poID, prevScheduleID sql.NullString
		var recognizedAt sql.NullTime
		var adjustmentReason sql.NullString

		err := rows.Scan(
			&s.ID, &s.TenantID, &s.InvoiceID, &poID, &s.RecognitionMonth,
			&s.PeriodStartDate, &s.PeriodEndDate, &s.AllocatedAmountCents, &s.RecognizedAmountCents,
			&s.DeferredAmountCents, &s.RevenueType, &s.IsRecognized, &recognizedAt, &s.OriginalTotalCents,
			&s.IsAdjustment, &adjustmentReason, &prevScheduleID, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			continue
		}
		if poID.Valid {
			parsedID, _ := uuid.Parse(poID.String)
			s.PerformanceObligationID = parsedID
		}
		if recognizedAt.Valid {
			s.RecognizedAt = &recognizedAt.Time
		}
		if adjustmentReason.Valid {
			s.AdjustmentReason = adjustmentReason.String
		}
		if prevScheduleID.Valid {
			parsedID, _ := uuid.Parse(prevScheduleID.String)
			s.PreviousScheduleID = &parsedID
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

func (r *RevenueRecognitionRepository) CalculateSSPDiscountPercentage(totalPrice, sspSum int) float64 {
	if totalPrice <= 0 || sspSum <= 0 {
		return 1.0
	}
	return float64(totalPrice) / float64(sspSum)
}

func (r *RevenueRecognitionRepository) AllocateTransactionPrice(totalPrice int, obligations []*PerformanceObligation) error {
	if len(obligations) == 0 {
		return fmt.Errorf("no performance obligations to allocate")
	}

	totalSSP := 0
	for _, po := range obligations {
		totalSSP += po.SSPCents
	}

	if totalSSP == 0 {
		return fmt.Errorf("total SSP is zero")
	}

	for _, po := range obligations {
		ratio := float64(po.SSPCents) / float64(totalSSP)
		po.AllocatedPriceCents = int(float64(totalPrice) * ratio)
	}

	allocatedSum := 0
	for _, po := range obligations {
		allocatedSum += po.AllocatedPriceCents
	}

	diff := totalPrice - allocatedSum
	if diff != 0 && len(obligations) > 0 {
		obligations[0].AllocatedPriceCents += diff
	}

	return nil
}