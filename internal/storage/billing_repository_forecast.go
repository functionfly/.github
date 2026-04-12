package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// UsageAlertRepository provides methods for managing usage alerts and spend caps
type UsageAlertRepository struct {
	db *sql.DB
}

// NewUsageAlertRepository creates a new usage alert repository
func NewUsageAlertRepository(db *sql.DB) *UsageAlertRepository {
	return &UsageAlertRepository{db: db}
}

// CreateUsageAlert creates a new usage alert configuration
func (r *UsageAlertRepository) CreateUsageAlert(ctx context.Context, alert *UsageAlert) error {
	alert.ID = uuid.New()
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()
	alert.TriggerCount = 0
	alert.IsEnabled = true

	query := `
		INSERT INTO usage_alerts (id, tenant_id, name, alert_type, threshold_value, threshold_operator,
			period_type, notification_channels, is_enabled, cooldown_minutes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		alert.ID, alert.TenantID, alert.Name, alert.AlertType, alert.ThresholdValue,
		alert.ThresholdOperator, alert.PeriodType, pq.Array(alert.NotificationChannels),
		alert.IsEnabled, alert.CooldownMinutes, alert.CreatedAt, alert.UpdatedAt,
	).Scan(&alert.ID)
}

// GetUsageAlertByID retrieves a usage alert by ID
func (r *UsageAlertRepository) GetUsageAlertByID(ctx context.Context, id uuid.UUID) (*UsageAlert, error) {
	query := `
		SELECT id, tenant_id, name, alert_type, threshold_value, threshold_operator,
			period_type, notification_channels, is_enabled, last_triggered_at, trigger_count,
			cooldown_minutes, created_at, updated_at
		FROM usage_alerts WHERE id = $1`

	alert := &UsageAlert{}
	var notificationChannels []string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&alert.ID, &alert.TenantID, &alert.Name, &alert.AlertType, &alert.ThresholdValue,
		&alert.ThresholdOperator, &alert.PeriodType, pq.Array(&notificationChannels),
		&alert.IsEnabled, &alert.LastTriggeredAt, &alert.TriggerCount,
		&alert.CooldownMinutes, &alert.CreatedAt, &alert.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get usage alert: %w", err)
	}

	alert.NotificationChannels = notificationChannels
	return alert, nil
}

// ListUsageAlertsByTenant lists all usage alerts for a tenant
func (r *UsageAlertRepository) ListUsageAlertsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*UsageAlert, error) {
	query := `
		SELECT id, tenant_id, name, alert_type, threshold_value, threshold_operator,
			period_type, notification_channels, is_enabled, last_triggered_at, trigger_count,
			cooldown_minutes, created_at, updated_at
		FROM usage_alerts WHERE tenant_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list usage alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*UsageAlert
	for rows.Next() {
		alert := &UsageAlert{}
		var notificationChannels []string

		err := rows.Scan(
			&alert.ID, &alert.TenantID, &alert.Name, &alert.AlertType, &alert.ThresholdValue,
			&alert.ThresholdOperator, &alert.PeriodType, pq.Array(&notificationChannels),
			&alert.IsEnabled, &alert.LastTriggeredAt, &alert.TriggerCount,
			&alert.CooldownMinutes, &alert.CreatedAt, &alert.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage alert: %w", err)
		}

		alert.NotificationChannels = notificationChannels
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// UpdateUsageAlert updates a usage alert configuration
func (r *UsageAlertRepository) UpdateUsageAlert(ctx context.Context, alert *UsageAlert) error {
	alert.UpdatedAt = time.Now()

	query := `
		UPDATE usage_alerts SET
			name = $2, alert_type = $3, threshold_value = $4, threshold_operator = $5,
			period_type = $6, notification_channels = $7, is_enabled = $8,
			cooldown_minutes = $9, updated_at = $10
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		alert.ID, alert.Name, alert.AlertType, alert.ThresholdValue,
		alert.ThresholdOperator, alert.PeriodType, pq.Array(alert.NotificationChannels),
		alert.IsEnabled, alert.CooldownMinutes, alert.UpdatedAt,
	)
	return err
}

// DeleteUsageAlert deletes a usage alert
func (r *UsageAlertRepository) DeleteUsageAlert(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM usage_alerts WHERE id = $1", id)
	return err
}

// RecordAlertTrigger records when an alert was triggered
func (r *UsageAlertRepository) RecordAlertTrigger(ctx context.Context, history *UsageAlertHistory) error {
	history.ID = uuid.New()
	history.TriggeredAt = time.Now()

	query := `
		INSERT INTO usage_alert_history (id, alert_id, tenant_id, triggered_at, triggered_value,
			threshold_value, message, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		history.ID, history.AlertID, history.TenantID, history.TriggeredAt,
		history.TriggeredValue, history.ThresholdValue, history.Message, history.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to record alert trigger: %w", err)
	}

	// Update the alert's last triggered time and count
	_, err = r.db.ExecContext(ctx, `
		UPDATE usage_alerts SET
			last_triggered_at = $2,
			trigger_count = trigger_count + 1,
			updated_at = NOW()
		WHERE id = $1`,
		history.AlertID, history.TriggeredAt,
	)

	return err
}

// GetAlertHistoryByTenant retrieves alert history for a tenant
func (r *UsageAlertRepository) GetAlertHistoryByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*UsageAlertHistory, error) {
	query := `
		SELECT id, alert_id, tenant_id, triggered_at, triggered_value, threshold_value,
			message, metadata, acknowledged_at, acknowledged_by
		FROM usage_alert_history
		WHERE tenant_id = $1
		ORDER BY triggered_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert history: %w", err)
	}
	defer rows.Close()

	var history []*UsageAlertHistory
	for rows.Next() {
		h := &UsageAlertHistory{}
		err := rows.Scan(
			&h.ID, &h.AlertID, &h.TenantID, &h.TriggeredAt, &h.TriggeredValue,
			&h.ThresholdValue, &h.Message, &h.Metadata, &h.AcknowledgedAt, &h.AcknowledgedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

// CreateOrUpdateSpendCap creates or updates a spend cap for a tenant
func (r *UsageAlertRepository) CreateOrUpdateSpendCap(ctx context.Context, cap *SpendCap) error {
	cap.ID = uuid.New()
	cap.CreatedAt = time.Now()
	cap.UpdatedAt = time.Now()

	query := `
		INSERT INTO spend_caps (id, tenant_id, cap_amount_cents, warning_thresholds, current_spend_cents,
			period_start, period_end, action_on_cap, is_hard_cap, is_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (tenant_id, period_start) DO UPDATE SET
			cap_amount_cents = EXCLUDED.cap_amount_cents,
			warning_thresholds = EXCLUDED.warning_thresholds,
			action_on_cap = EXCLUDED.action_on_cap,
			is_hard_cap = EXCLUDED.is_hard_cap,
			is_enabled = EXCLUDED.is_enabled,
			updated_at = NOW()
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		cap.ID, cap.TenantID, cap.CapAmountCents, pq.Array(cap.WarningThresholds),
		cap.CurrentSpendCents, cap.PeriodStart, cap.PeriodEnd, cap.ActionOnCap,
		cap.IsHardCap, cap.IsEnabled, cap.CreatedAt, cap.UpdatedAt,
	).Scan(&cap.ID)
}

// GetSpendCapByTenant retrieves the current spend cap for a tenant
func (r *UsageAlertRepository) GetSpendCapByTenant(ctx context.Context, tenantID uuid.UUID, periodStart time.Time) (*SpendCap, error) {
	query := `
		SELECT id, tenant_id, cap_amount_cents, warning_thresholds, current_spend_cents,
			period_start, period_end, action_on_cap, is_hard_cap, is_enabled, created_at, updated_at
		FROM spend_caps
		WHERE tenant_id = $1 AND period_start = $2`

	cap := &SpendCap{}
	var warningThresholds []int

	err := r.db.QueryRowContext(ctx, query, tenantID, periodStart).Scan(
		&cap.ID, &cap.TenantID, &cap.CapAmountCents, pq.Array(&warningThresholds),
		&cap.CurrentSpendCents, &cap.PeriodStart, &cap.PeriodEnd, &cap.ActionOnCap,
		&cap.IsHardCap, &cap.IsEnabled, &cap.CreatedAt, &cap.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get spend cap: %w", err)
	}

	cap.WarningThresholds = warningThresholds
	return cap, nil
}

// UpdateCurrentSpend updates the current spend amount for a cap
func (r *UsageAlertRepository) UpdateCurrentSpend(ctx context.Context, capID uuid.UUID, spendCents int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE spend_caps SET
			current_spend_cents = $2,
			updated_at = NOW()
		WHERE id = $1`,
		capID, spendCents,
	)
	return err
}

// SaveUsageForecast saves a usage forecast
func (r *UsageAlertRepository) SaveUsageForecast(ctx context.Context, forecast *UsageForecast) error {
	forecast.ID = uuid.New()
	forecast.CreatedAt = time.Now()

	query := `
		INSERT INTO usage_forecasts (id, tenant_id, forecast_type, period_start, period_end,
			current_value, predicted_value, lower_bound, upper_bound, confidence, method_used,
			growth_rate, days_of_history, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err := r.db.ExecContext(ctx, query,
		forecast.ID, forecast.TenantID, forecast.ForecastType, forecast.PeriodStart, forecast.PeriodEnd,
		forecast.CurrentValue, forecast.PredictedValue, forecast.LowerBound, forecast.UpperBound,
		forecast.Confidence, forecast.MethodUsed, forecast.GrowthRate, forecast.DaysOfHistory,
		forecast.Metadata, forecast.CreatedAt,
	)
	return err
}

// GetLatestForecast retrieves the latest forecast for a tenant and type
func (r *UsageAlertRepository) GetLatestForecast(ctx context.Context, tenantID uuid.UUID, forecastType string) (*UsageForecast, error) {
	query := `
		SELECT id, tenant_id, forecast_type, period_start, period_end, current_value,
			predicted_value, lower_bound, upper_bound, confidence, method_used,
			growth_rate, days_of_history, metadata, created_at
		FROM usage_forecasts
		WHERE tenant_id = $1 AND forecast_type = $2
		ORDER BY created_at DESC
		LIMIT 1`

	forecast := &UsageForecast{}
	err := r.db.QueryRowContext(ctx, query, tenantID, forecastType).Scan(
		&forecast.ID, &forecast.TenantID, &forecast.ForecastType, &forecast.PeriodStart,
		&forecast.PeriodEnd, &forecast.CurrentValue, &forecast.PredictedValue,
		&forecast.LowerBound, &forecast.UpperBound, &forecast.Confidence,
		&forecast.MethodUsed, &forecast.GrowthRate, &forecast.DaysOfHistory,
		&forecast.Metadata, &forecast.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest forecast: %w", err)
	}

	return forecast, nil
}

// GetDailyUsageHistory retrieves daily usage data for time series analysis
func (r *UsageAlertRepository) GetDailyUsageHistory(ctx context.Context, tenantID uuid.UUID, eventType string, days int) ([]*DailyUsagePoint, error) {
	query := `
		SELECT period_date, total_quantity
		FROM usage_rollups
		WHERE tenant_id = $1
			AND event_type = $2
			AND period_date >= $3
		ORDER BY period_date ASC`

	startDate := time.Now().UTC().AddDate(0, 0, -days).Truncate(24 * time.Hour)

	rows, err := r.db.QueryContext(ctx, query, tenantID, eventType, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily usage history: %w", err)
	}
	defer rows.Close()

	var points []*DailyUsagePoint
	for rows.Next() {
		point := &DailyUsagePoint{}
		err := rows.Scan(&point.Date, &point.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily usage: %w", err)
		}
		points = append(points, point)
	}

	return points, nil
}

// GetDailySpendHistory retrieves daily spend data for forecasting
func (r *UsageAlertRepository) GetDailySpendHistory(ctx context.Context, tenantID uuid.UUID, days int) ([]*DailyUsagePoint, error) {
	query := `
		SELECT DATE(period_end) as date, SUM(amount_due_cents) as spend
		FROM invoices
		WHERE tenant_id = $1
			AND status IN ('draft', 'open', 'paid')
			AND period_end >= $2
		GROUP BY DATE(period_end)
		ORDER BY date ASC`

	startDate := time.Now().UTC().AddDate(0, 0, -days).Truncate(24 * time.Hour)

	rows, err := r.db.QueryContext(ctx, query, tenantID, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily spend history: %w", err)
	}
	defer rows.Close()

	var points []*DailyUsagePoint
	for rows.Next() {
		point := &DailyUsagePoint{}
		err := rows.Scan(&point.Date, &point.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily spend: %w", err)
		}
		points = append(points, point)
	}

	return points, nil
}

// GetCurrentPeriodUsage retrieves current period usage summary for a tenant
func (r *UsageAlertRepository) GetCurrentPeriodUsage(ctx context.Context, tenantID uuid.UUID, periodStart, periodEnd time.Time) (*UsageSummary, error) {
	query := `
		SELECT COALESCE(SUM(CASE WHEN event_type = 'function_execution' THEN total_quantity ELSE 0 END), 0) as total_executions,
		       COALESCE(SUM(CASE WHEN event_type = 'compute_time_ms' THEN total_quantity ELSE 0 END), 0) as total_compute_ms
		FROM usage_rollups
		WHERE tenant_id = $1 AND period_date >= $2 AND period_date <= $3`

	summary := &UsageSummary{
		TenantID:    tenantID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}

	err := r.db.QueryRowContext(ctx, query, tenantID, periodStart, periodEnd).Scan(
		&summary.TotalExecutions, &summary.TotalComputeMs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get current period usage: %w", err)
	}

	return summary, nil
}

// UsageSummary provides a summary of tenant usage
type UsageSummary struct {
	TenantID        uuid.UUID
	PeriodStart     time.Time
	PeriodEnd       time.Time
	TotalExecutions int
	TotalComputeMs  int
	EstimatedCost   int // in cents
}
