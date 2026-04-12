package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RecordUsageEvent records a usage event
func (r *BillingRepository) RecordUsageEvent(ctx context.Context, event *UsageEvent) error {
	event.ID = uuid.New()
	event.Timestamp = time.Now()

	query := `INSERT INTO usage_events (id, tenant_id, event_type, quantity, unit_price_cents, metadata, timestamp)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	var metadata []byte
	if event.Metadata != nil {
		metadata, _ = json.Marshal(event.Metadata)
	}

	_, err := r.db.Exec(query, event.ID, event.TenantID, event.EventType, event.Quantity,
		event.UnitPriceCents, metadata, event.Timestamp)

	if err != nil {
		return fmt.Errorf("failed to record usage event: %w", err)
	}

	return nil
}

// GetUsageByTenant gets usage rollups for a tenant
func (r *BillingRepository) GetUsageByTenant(tenantID uuid.UUID, eventType string, start, end time.Time) ([]*UsageRollup, error) {
	query := `
		SELECT id, tenant_id, event_type, period_date, total_quantity, created_at, updated_at
		FROM usage_rollups
		WHERE tenant_id = $1 AND event_type = $2 AND period_date >= $3 AND period_date <= $4
		ORDER BY period_date ASC`

	rows, err := r.db.Query(query, tenantID, eventType, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}
	defer rows.Close()

	var rollups []*UsageRollup
	for rows.Next() {
		rollup := &UsageRollup{}
		err := rows.Scan(&rollup.ID, &rollup.TenantID, &rollup.EventType, &rollup.PeriodDate,
			&rollup.TotalQuantity, &rollup.CreatedAt, &rollup.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage rollup: %w", err)
		}
		rollups = append(rollups, rollup)
	}

	return rollups, nil
}

// CreateOrUpdateUsageRollup creates or updates a usage rollup for a tenant/event_type/date
func (r *BillingRepository) CreateOrUpdateUsageRollup(ctx context.Context, rollup *UsageRollup) error {
	rollup.ID = uuid.New()
	rollup.CreatedAt = time.Now()
	rollup.UpdatedAt = time.Now()

	query := `
		INSERT INTO usage_rollups (id, tenant_id, event_type, period_date, total_quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id, event_type, period_date)
		DO UPDATE SET
			total_quantity = usage_rollups.total_quantity + EXCLUDED.total_quantity,
			updated_at = NOW()
		RETURNING id, total_quantity, created_at, updated_at
	`

	var totalQuantity int
	err := r.db.QueryRowContext(ctx, query,
		rollup.ID, rollup.TenantID, rollup.EventType, rollup.PeriodDate,
		rollup.TotalQuantity, rollup.CreatedAt, rollup.UpdatedAt,
	).Scan(&rollup.ID, &totalQuantity, &rollup.CreatedAt, &rollup.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create or update usage rollup: %w", err)
	}

	// Update the total quantity to reflect the merged value
	rollup.TotalQuantity = totalQuantity

	return nil
}
