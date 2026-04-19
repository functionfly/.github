package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AnalyticsSettings struct {
	ID          uuid.UUID              `json:"id"`
	ServiceName string                 `json:"service_name"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type AnalyticsRepository struct {
	db *PostgresDB
}

func NewAnalyticsRepository(db *PostgresDB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// CreateAnalyticsSettings creates new analytics settings for a service
func (r *AnalyticsRepository) CreateAnalyticsSettings(settings *AnalyticsSettings) (*AnalyticsSettings, error) {
	configJSON, err := json.Marshal(settings.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if settings.ID == uuid.Nil {
		settings.ID = uuid.New()
	}
	now := time.Now()
	settings.CreatedAt = now
	settings.UpdatedAt = now

	var returnedConfigJSON []byte
	err = r.db.QueryRow(`
		INSERT INTO analytics_settings (id, service_name, enabled, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (service_name) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			config = EXCLUDED.config,
			updated_at = EXCLUDED.updated_at
		RETURNING id, service_name, enabled, config, created_at, updated_at`,
		settings.ID, settings.ServiceName, settings.Enabled, configJSON, settings.CreatedAt, settings.UpdatedAt).Scan(
		&settings.ID, &settings.ServiceName, &settings.Enabled, &returnedConfigJSON,
		&settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create analytics settings: %w", err)
	}

	if err := json.Unmarshal(returnedConfigJSON, &settings.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal returned config: %w", err)
	}

	return settings, nil
}

// InitializeTenantAnalytics creates default analytics configuration for a tenant
func (r *AnalyticsRepository) InitializeTenantAnalytics(tenantID uuid.UUID) error {
	// Create tenant-specific analytics tracking entry
	settings := &AnalyticsSettings{
		ID:          uuid.New(),
		ServiceName: fmt.Sprintf("tenant_%s_analytics", tenantID.String()),
		Enabled:     true,
		Config: map[string]interface{}{
			"tenant_id":         tenantID.String(),
			"tracking_enabled":  true,
			"track_executions":  true,
			"track_errors":      true,
			"track_performance": true,
			"retention_days":    90,
		},
	}

	_, err := r.CreateAnalyticsSettings(settings)
	return err
}

func (r *AnalyticsRepository) GetAnalyticsSettings(serviceName string) (*AnalyticsSettings, error) {
	settings := &AnalyticsSettings{}
	var configJSON []byte
	var config map[string]interface{}

	err := r.db.QueryRow(`
		SELECT id, service_name, enabled, config, created_at, updated_at
		FROM analytics_settings
		WHERE service_name = $1`, serviceName).Scan(
		&settings.ID, &settings.ServiceName, &settings.Enabled, &configJSON,
		&settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("analytics settings not found for service: %s", serviceName)
		}
		return nil, fmt.Errorf("failed to get analytics settings: %w", err)
	}

	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	settings.Config = config

	return settings, nil
}

func (r *AnalyticsRepository) GetAllAnalyticsSettings() ([]AnalyticsSettings, error) {
	rows, err := r.db.Query(`
		SELECT id, service_name, enabled, config, created_at, updated_at
		FROM analytics_settings
		ORDER BY service_name`)
	if err != nil {
		return nil, fmt.Errorf("failed to query analytics settings: %w", err)
	}
	defer rows.Close()

	var settings []AnalyticsSettings
	for rows.Next() {
		s := AnalyticsSettings{}
		var configJSON []byte
		var config map[string]interface{}

		err := rows.Scan(&s.ID, &s.ServiceName, &s.Enabled, &configJSON, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analytics settings: %w", err)
		}

		if err := json.Unmarshal(configJSON, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
		s.Config = config
		settings = append(settings, s)
	}

	return settings, nil
}

func (r *AnalyticsRepository) UpdateAnalyticsSettings(serviceName string, enabled bool, config map[string]interface{}) (*AnalyticsSettings, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	settings := &AnalyticsSettings{}
	var returnedConfigJSON []byte

	err = r.db.QueryRow(`
		UPDATE analytics_settings
		SET enabled = $2, config = $3, updated_at = NOW()
		WHERE service_name = $1
		RETURNING id, service_name, enabled, config, created_at, updated_at`,
		serviceName, enabled, configJSON).Scan(
		&settings.ID, &settings.ServiceName, &settings.Enabled, &returnedConfigJSON,
		&settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("analytics settings not found for service: %s", serviceName)
		}
		return nil, fmt.Errorf("failed to update analytics settings: %w", err)
	}

	if err := json.Unmarshal(returnedConfigJSON, &settings.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal returned config: %w", err)
	}

	return settings, nil
}

// ==================== MRR/ARR Calculation Methods ====================

// CalculateMRR calculates Monthly Recurring Revenue for a given month
func (r *AnalyticsRepository) CalculateMRR(ctx context.Context, year, month int) (*MRRMetrics, error) {
	periodMonth := fmt.Sprintf("%04d-%02d", year, month)
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	metrics := &MRRMetrics{
		PeriodMonth:     periodMonth,
		CalculationDate: time.Now(),
	}

	// Calculate active MRR at end of period (active bundle subscriptions)
	query := `
		SELECT COALESCE(SUM(b.display_price_cents), 0)
		FROM bundle_subscriptions bs
		JOIN pricing_bundles b ON bs.bundle_id = b.id
		WHERE bs.status = 'active'
		AND bs.current_period_start < $1
		AND bs.current_period_end >= $2
	`

	var mrrCents sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, endDate, startDate).Scan(&mrrCents)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate MRR: %w", err)
	}
	metrics.MRR_Cents = int(mrrCents.Int64)
	metrics.MRR_USD = float64(metrics.MRR_Cents) / 100.0

	// New MRR: subscriptions created in this period
	newMRRQuery := `
		SELECT COALESCE(SUM(b.display_price_cents), 0)
		FROM bundle_subscriptions bs
		JOIN pricing_bundles b ON bs.bundle_id = b.id
		WHERE bs.status = 'active'
		AND bs.created_at >= $1 AND bs.created_at < $2
	`

	var newMRRCents sql.NullInt64
	err = r.db.QueryRowContext(ctx, newMRRQuery, startDate, endDate).Scan(&newMRRCents)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate new MRR: %w", err)
	}
	metrics.NewMRR_Cents = int(newMRRCents.Int64)

	// Churned MRR: lost this period
	churnedQuery := `
		SELECT COALESCE(SUM(mrr_lost_cents), 0)
		FROM subscription_churn_events
		WHERE event_type IN ('cancellation', 'downgrade')
		AND event_date >= $1 AND event_date < $2
		AND is_recovered = false
	`

	var churnedMRRCents sql.NullInt64
	err = r.db.QueryRowContext(ctx, churnedQuery, startDate, endDate).Scan(&churnedMRRCents)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate churned MRR: %w", err)
	}
	metrics.ChurnedMRR_Cents = int(churnedMRRCents.Int64)

	// Net MRR Change
	metrics.NetMRRChange_Cents = metrics.NewMRR_Cents - metrics.ChurnedMRR_Cents

	return metrics, nil
}

// CalculateARR calculates Annual Recurring Revenue (MRR * 12)
func (r *AnalyticsRepository) CalculateARR(ctx context.Context, year, month int) (*ARRMetrics, error) {
	mrr, err := r.CalculateMRR(ctx, year, month)
	if err != nil {
		return nil, err
	}

	arr := &ARRMetrics{
		CalculationDate: mrr.CalculationDate,
	}

	// ARR = MRR * 12
	arr.ARR_Cents = mrr.MRR_Cents * 12
	arr.ARR_USD = float64(arr.ARR_Cents) / 100.0

	// Calculate ARR components
	arr.NewARR_Cents = mrr.NewMRR_Cents * 12
	arr.ChurnedARR_Cents = mrr.ChurnedMRR_Cents * 12

	return arr, nil
}

// GetMRRTimeseries returns MRR data over time for charts
func (r *AnalyticsRepository) GetMRRTimeseries(ctx context.Context, months int) ([]MRRMetrics, error) {
	results := make([]MRRMetrics, 0, months)
	endDate := time.Now()

	for i := 0; i < months; i++ {
		currentMonth := endDate.AddDate(0, -months+1+i, 0)
		year := currentMonth.Year()
		month := int(currentMonth.Month())

		mrr, err := r.CalculateMRR(ctx, year, month)
		if err != nil {
			return nil, err
		}

		results = append(results, *mrr)
	}

	return results, nil
}

// ==================== Churn Metrics Methods ====================

// CalculateChurnMetrics calculates churn metrics for a given month
func (r *AnalyticsRepository) CalculateChurnMetrics(ctx context.Context, year, month int) (*ChurnMetrics, error) {
	periodMonth := fmt.Sprintf("%04d-%02d", year, month)
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	metrics := &ChurnMetrics{
		PeriodMonth:     periodMonth,
		CalculationDate: time.Now(),
	}

	// Customers at start of period (active subscriptions at start)
	customersAtStartQuery := `
		SELECT COUNT(DISTINCT tenant_id)
		FROM bundle_subscriptions
		WHERE created_at < $1
		AND (canceled_at IS NULL OR canceled_at >= $1)
	`

	var customersAtStart sql.NullInt64
	err := r.db.QueryRowContext(ctx, customersAtStartQuery, startDate).Scan(&customersAtStart)
	if err != nil {
		return nil, fmt.Errorf("failed to count customers at start: %w", err)
	}
	metrics.CustomersAtStart = int(customersAtStart.Int64)

	// Churned customers in this period (new cancellations)
	churnedQuery := `
		SELECT COUNT(DISTINCT tenant_id)
		FROM subscription_churn_events
		WHERE event_type = 'cancellation'
		AND event_date >= $1 AND event_date < $2
		AND is_recovered = false
	`

	var customersChurned sql.NullInt64
	err = r.db.QueryRowContext(ctx, churnedQuery, startDate, endDate).Scan(&customersChurned)
	if err != nil {
		return nil, fmt.Errorf("failed to count churned customers: %w", err)
	}
	metrics.CustomersChurned = int(customersChurned.Int64)

	// Calculate customer churn rate
	if metrics.CustomersAtStart > 0 {
		metrics.CustomerChurnRate = float64(metrics.CustomersChurned) / float64(metrics.CustomersAtStart) * 100.0
	}

	// MRR at start of period
	mrrAtStartQuery := `
		SELECT COALESCE(SUM(b.display_price_cents), 0)
		FROM bundle_subscriptions bs
		JOIN pricing_bundles b ON bs.bundle_id = b.id
		WHERE bs.created_at < $1
		AND (bs.canceled_at IS NULL OR bs.canceled_at >= $1)
	`

	var mrrAtStart sql.NullInt64
	err = r.db.QueryRowContext(ctx, mrrAtStartQuery, startDate).Scan(&mrrAtStart)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate MRR at start: %w", err)
	}
	metrics.MRRAtStart_Cents = int(mrrAtStart.Int64)

	// MRR churned in this period
	mrrChurnedQuery := `
		SELECT COALESCE(SUM(mrr_lost_cents), 0)
		FROM subscription_churn_events
		WHERE event_date >= $1 AND event_date < $2
		AND event_type IN ('cancellation', 'downgrade')
		AND is_recovered = false
	`

	var mrrChurned sql.NullInt64
	err = r.db.QueryRowContext(ctx, mrrChurnedQuery, startDate, endDate).Scan(&mrrChurned)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate churned MRR: %w", err)
	}
	metrics.MRRChurned_Cents = int(mrrChurned.Int64)

	// Calculate revenue churn rate
	if metrics.MRRAtStart_Cents > 0 {
		metrics.RevenueChurnRate = float64(metrics.MRRChurned_Cents) / float64(metrics.MRRAtStart_Cents) * 100.0
	}

	// Gross Revenue Retention (GRR) = 100% - gross churn
	if metrics.MRRAtStart_Cents > 0 {
		metrics.GrossRevenueRetention = 100.0 - (float64(metrics.MRRChurned_Cents) / float64(metrics.MRRAtStart_Cents) * 100.0)
	}

	// Breakdown by churn type
	// Voluntary churn (cancellations without payment issues)
	voluntaryQuery := `
		SELECT COUNT(*)
		FROM subscription_churn_events
		WHERE event_type = 'cancellation'
		AND event_date >= $1 AND event_date < $2
		AND (cancel_reason IS NULL OR 
		     (cancel_reason NOT LIKE '%payment%' AND cancel_reason NOT LIKE '%failed%' AND cancel_reason NOT LIKE '%card%'))
	`

	var voluntaryChurn sql.NullInt64
	err = r.db.QueryRowContext(ctx, voluntaryQuery, startDate, endDate).Scan(&voluntaryChurn)
	if err != nil {
		return nil, fmt.Errorf("failed to count voluntary churn: %w", err)
	}
	metrics.VoluntaryChurn = int(voluntaryChurn.Int64)

	// Involuntary churn (payment failures)
	involuntaryQuery := `
		SELECT COUNT(*)
		FROM subscription_churn_events
		WHERE event_type IN ('failed_renewal', 'payment_failure')
		AND event_date >= $1 AND event_date < $2
	`

	var involuntaryChurn sql.NullInt64
	err = r.db.QueryRowContext(ctx, involuntaryQuery, startDate, endDate).Scan(&involuntaryChurn)
	if err != nil {
		return nil, fmt.Errorf("failed to count involuntary churn: %w", err)
	}
	metrics.InvoluntaryChurn = int(involuntaryChurn.Int64)

	// Downgrades
	downgradeQuery := `
		SELECT COUNT(*)
		FROM subscription_churn_events
		WHERE event_type = 'downgrade'
		AND event_date >= $1 AND event_date < $2
	`

	var downgrades sql.NullInt64
	err = r.db.QueryRowContext(ctx, downgradeQuery, startDate, endDate).Scan(&downgrades)
	if err != nil {
		return nil, fmt.Errorf("failed to count downgrades: %w", err)
	}
	metrics.DowngradeChurn = int(downgrades.Int64)

	// Failed renewals (payment failures that led to cancellation)
	failedRenewalQuery := `
		SELECT COUNT(*)
		FROM subscription_churn_events
		WHERE event_type = 'failed_renewal'
		AND event_date >= $1 AND event_date < $2
	`

	var failedRenewals sql.NullInt64
	err = r.db.QueryRowContext(ctx, failedRenewalQuery, startDate, endDate).Scan(&failedRenewals)
	if err != nil {
		return nil, fmt.Errorf("failed to count failed renewals: %w", err)
	}
	metrics.FailedRenewalChurn = int(failedRenewals.Int64)

	return metrics, nil
}

// RecordSubscriptionChurnEvent records a churn event for tracking
func (r *AnalyticsRepository) RecordSubscriptionChurnEvent(ctx context.Context, event *SubscriptionChurnEvent) error {
	query := `
		INSERT INTO subscription_churn_events (
			id, tenant_id, subscription_id, bundle_id, event_type, event_date, cohort_month,
			previous_mrr_cents, new_mrr_cents, mrr_lost_cents, previous_bundle_id, new_bundle_id,
			cancel_reason, cancel_at_period_end, cancel_date, failed_payment_amount_cents,
			failure_code, attempt_count, stripe_subscription_id, stripe_event_id,
			is_recovered, recovered_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
		ON CONFLICT (id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query,
		event.ID, event.TenantID, event.SubscriptionID, event.BundleID,
		event.EventType, event.EventDate, event.CohortMonth,
		event.PreviousMRR_Cents, event.NewMRR_Cents, event.MRRLost_Cents,
		event.PreviousBundleID, event.NewBundleID,
		event.CancelReason, event.CancelAtPeriodEnd, event.CancelDate,
		event.FailedPaymentAmount_Cents, event.FailureCode, event.AttemptCount,
		event.StripeSubscriptionID, event.StripeEventID,
		event.IsRecovered, event.RecoveredAt,
		event.CreatedAt, event.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to record churn event: %w", err)
	}

	return nil
}

// GetChurnMetricsTimeseries returns churn metrics over time
func (r *AnalyticsRepository) GetChurnMetricsTimeseries(ctx context.Context, months int) ([]ChurnMetrics, error) {
	results := make([]ChurnMetrics, 0, months)
	endDate := time.Now()

	for i := 0; i < months; i++ {
		currentMonth := endDate.AddDate(0, -months+1+i, 0)
		year := currentMonth.Year()
		month := int(currentMonth.Month())

		churn, err := r.CalculateChurnMetrics(ctx, year, month)
		if err != nil {
			return nil, err
		}

		results = append(results, *churn)
	}

	return results, nil
}

// ==================== Financial Reporting Methods ====================

// GenerateFinancialReport generates a comprehensive financial report
func (r *AnalyticsRepository) GenerateFinancialReport(ctx context.Context, reportType string, periodStart, periodEnd time.Time) (*FinancialReport, error) {
	report := &FinancialReport{
		ReportID:    uuid.New(),
		GeneratedAt: time.Now(),
		ReportType:  reportType,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}

	// Total revenue (paid invoices in period)
	revenueQuery := `
		SELECT COALESCE(SUM(amount_paid_cents), 0)
		FROM invoices
		WHERE paid_at >= $1 AND paid_at < $2
		AND status = 'paid'
	`

	var totalRevenue sql.NullInt64
	err := r.db.QueryRowContext(ctx, revenueQuery, periodStart, periodEnd).Scan(&totalRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total revenue: %w", err)
	}
	report.TotalRevenue_Cents = int(totalRevenue.Int64)

	// Subscription revenue (invoices with subscription_id)
	subRevenueQuery := `
		SELECT COALESCE(SUM(amount_paid_cents), 0)
		FROM invoices
		WHERE paid_at >= $1 AND paid_at < $2
		AND status = 'paid'
		AND subscription_id IS NOT NULL
	`

	var subRevenue sql.NullInt64
	err = r.db.QueryRowContext(ctx, subRevenueQuery, periodStart, periodEnd).Scan(&subRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate subscription revenue: %w", err)
	}
	report.SubscriptionRevenue_Cents = int(subRevenue.Int64)
	report.OneTimeRevenue_Cents = report.TotalRevenue_Cents - report.SubscriptionRevenue_Cents

	// Outstanding invoices (amount still due on unpaid invoices)
	outstandingQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN status IN ('open', 'pending') THEN amount_due_cents - amount_paid_cents ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'overdue' THEN amount_due_cents - amount_paid_cents ELSE 0 END), 0)
		FROM invoices
		WHERE due_date < $1
		AND status IN ('open', 'pending', 'overdue')
	`

	var outstanding, overdue sql.NullInt64
	err = r.db.QueryRowContext(ctx, outstandingQuery, periodEnd).Scan(&outstanding, &overdue)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate outstanding: %w", err)
	}

	report.OutstandingInvoices_Cents = int(outstanding.Int64)
	report.OverdueInvoices_Cents = int(overdue.Int64)

	// Tax collected
	taxQuery := `
		SELECT COALESCE(SUM(tax_amount_cents), 0)
		FROM invoice_tax_details itd
		JOIN invoices i ON itd.invoice_id = i.id
		WHERE i.paid_at >= $1 AND i.paid_at < $2
	`

	var taxCollected sql.NullInt64
	err = r.db.QueryRowContext(ctx, taxQuery, periodStart, periodEnd).Scan(&taxCollected)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate tax: %w", err)
	}
	report.TaxCollected_Cents = int(taxCollected.Int64)

	// MRR at end of period for ARR calculation
	currentMRR, err := r.CalculateMRR(ctx, periodEnd.Year(), int(periodEnd.Month()))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate MRR for ARR: %w", err)
	}
	report.MRR_Cents = currentMRR.MRR_Cents
	report.ARR_Cents = currentMRR.MRR_Cents * 12

	return report, nil
}

// GetTaxJurisdictionReport returns tax collection by jurisdiction
func (r *AnalyticsRepository) GetTaxJurisdictionReport(ctx context.Context, periodMonth string) ([]TaxJurisdictionReport, error) {
	query := `
		SELECT
			tr.country,
			tr.state,
			tr.tax_type as jurisdiction_type,
			tr.display_name as jurisdiction,
			tr.percentage as tax_rate,
			COUNT(*) as transactions_count,
			COALESCE(SUM(itd.subtotal_cents), 0) as taxable_revenue_cents,
			COALESCE(SUM(itd.tax_amount_cents), 0) as tax_collected_cents
		FROM invoice_tax_details itd
		JOIN tax_rates tr ON itd.tax_rate_id = tr.id
		JOIN invoices i ON itd.invoice_id = i.id
		WHERE TO_CHAR(i.paid_at, 'YYYY-MM') = $1
		GROUP BY tr.country, tr.state, tr.tax_type, tr.display_name, tr.percentage
		ORDER BY tr.country, tr.state
	`

	results := []TaxJurisdictionReport{}
	rows, err := r.db.QueryContext(ctx, query, periodMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get tax jurisdiction report: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var report TaxJurisdictionReport
		err := rows.Scan(
			&report.Country,
			&report.State,
			&report.JurisdictionType,
			&report.Jurisdiction,
			&report.TaxRate,
			&report.TransactionsCount,
			&report.TaxableRevenue_Cents,
			&report.TaxCollected_Cents,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tax jurisdiction row: %w", err)
		}
		results = append(results, report)
	}

	return results, nil
}

// ==================== LTV Metrics ====================

// GetLTVMetrics calculates lifetime value metrics
func (r *AnalyticsRepository) GetLTVMetrics(ctx context.Context) (*LTVMetrics, error) {
	ltv := &LTVMetrics{
		CalculationDate: time.Now(),
	}

	// Average LTV (total paid per customer)
	avgLTVQuery := `
		SELECT AVG(total_paid)
		FROM (
			SELECT tenant_id, SUM(amount_paid_cents) as total_paid
			FROM invoices
			WHERE paid_at IS NOT NULL
			GROUP BY tenant_id
		) as customer_totals
	`

	var avgLTV sql.NullFloat64
	err := r.db.QueryRowContext(ctx, avgLTVQuery).Scan(&avgLTV)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate average LTV: %w", err)
	}
	ltv.AverageLTV_Cents = int(avgLTV.Float64)

	return ltv, nil
}
