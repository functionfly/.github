package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ==================== MRR/ARR Calculation Methods ====================

// CalculateMRR calculates Monthly Recurring Revenue for a given month
func (r *BillingRepository) CalculateMRR(ctx context.Context, year, month int) (*MRRMetrics, error) {
	periodMonth := fmt.Sprintf("%04d-%02d", year, month)
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	metrics := &MRRMetrics{
		PeriodMonth:     periodMonth,
		CalculationDate: time.Now(),
	}

	// Calculate active MRR at end of period
	query := `
		SELECT COALESCE(SUM(b.display_price_cents), 0)
		FROM bundle_subscriptions bs
		JOIN pricing_bundles b ON bs.bundle_id = b.id
		WHERE bs.status = 'active'
		AND bs.current_period_start < $1
		AND bs.current_period_end >= $2
	`

	err := r.db.GORM.Raw(query, endDate, startDate).Scan(&metrics.MRR_Cents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate MRR: %w", err)
	}

	metrics.MRR_USD = float64(metrics.MRR_Cents) / 100.0

	// Calculate MRR components using churn events
	// New MRR: customers who started in this period
	newMRRQuery := `
		SELECT COALESCE(SUM(b.display_price_cents), 0)
		FROM bundle_subscriptions bs
		JOIN pricing_bundles b ON bs.bundle_id = b.id
		WHERE bs.status = 'active'
		AND bs.created_at >= $1 AND bs.created_at < $2
		AND NOT EXISTS (
			SELECT 1 FROM subscription_churn_events sce
			WHERE sce.subscription_id = bs.id
			AND sce.event_type IN ('cancellation', 'downgrade')
		)
	`

	var newMRRCents int
	err = r.db.GORM.Raw(newMRRQuery, startDate, endDate).Scan(&newMRRCents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate new MRR: %w", err)
	}
	metrics.NewMRR_Cents = newMRRCents

	// Expansion MRR: upgrades in this period
	expansionQuery := `
		SELECT COALESCE(SUM(sce.new_mrr_cents - sce.previous_mrr_cents), 0)
		FROM subscription_churn_events sce
		WHERE sce.event_type = 'upgrade'
		AND sce.event_date >= $1 AND sce.event_date < $2
	`

	var expansionMRRCents int
	err = r.db.GORM.Raw(expansionQuery, startDate, endDate).Scan(&expansionMRRCents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate expansion MRR: %w", err)
	}
	metrics.ExpansionMRR_Cents = expansionMRRCents

	// Contraction MRR: downgrades
	contractionQuery := `
		SELECT COALESCE(SUM(sce.previous_mrr_cents - sce.new_mrr_cents), 0)
		FROM subscription_churn_events sce
		WHERE sce.event_type = 'downgrade'
		AND sce.event_date >= $1 AND sce.event_date < $2
	`

	var contractionMRRCents int
	err = r.db.GORM.Raw(contractionQuery, startDate, endDate).Scan(&contractionMRRCents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate contraction MRR: %w", err)
	}
	metrics.ContractionMRR_Cents = contractionMRRCents

	// Churned MRR: cancellations
	churnedQuery := `
		SELECT COALESCE(SUM(mrr_lost_cents), 0)
		FROM subscription_churn_events
		WHERE event_type = 'cancellation'
		AND event_date >= $1 AND event_date < $2
	`

	var churnedMRRCents int
	err = r.db.GORM.Raw(churnedQuery, startDate, endDate).Scan(&churnedMRRCents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate churned MRR: %w", err)
	}
	metrics.ChurnedMRR_Cents = churnedMRRCents

	// Reactivation MRR: recovered customers
	reactivationQuery := `
		SELECT COALESCE(SUM(b.display_price_cents), 0)
		FROM bundle_subscriptions bs
		JOIN pricing_bundles b ON bs.bundle_id = b.id
		WHERE bs.status = 'active'
		AND bs.created_at >= $1 AND bs.created_at < $2
		AND EXISTS (
			SELECT 1 FROM subscription_churn_events sce
			WHERE sce.tenant_id = bs.tenant_id
			AND sce.event_type = 'cancellation'
			AND sce.is_recovered = true
			AND sce.recovered_at >= $1 AND sce.recovered_at < $2
		)
	`

	var reactivationMRRCents int
	err = r.db.GORM.Raw(reactivationQuery, startDate, endDate).Scan(&reactivationMRRCents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate reactivation MRR: %w", err)
	}
	metrics.ReactivationMRR_Cents = reactivationMRRCents

	// Net MRR Change
	metrics.NetMRRChange_Cents = metrics.NewMRR_Cents + metrics.ExpansionMRR_Cents +
		metrics.ReactivationMRR_Cents - metrics.ContractionMRR_Cents - metrics.ChurnedMRR_Cents

	return metrics, nil
}

// CalculateARR calculates Annual Recurring Revenue (MRR * 12)
func (r *BillingRepository) CalculateARR(ctx context.Context, year, month int) (*ARRMetrics, error) {
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
	arr.ExpansionARR_Cents = mrr.ExpansionMRR_Cents * 12
	arr.ContractionARR_Cents = mrr.ContractionMRR_Cents * 12
	arr.ChurnedARR_Cents = mrr.ChurnedMRR_Cents * 12
	arr.ReactivationARR_Cents = mrr.ReactivationMRR_Cents * 12

	return arr, nil
}

// GetMRRTimeseries returns MRR data over time for charts
func (r *BillingRepository) GetMRRTimeseries(ctx context.Context, months int) ([]MRRMetrics, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0)

	results := make([]MRRMetrics, 0, months)

	for i := 0; i < months; i++ {
		year := startDate.Year()
		month := int(startDate.Month())

		mrr, err := r.CalculateMRR(ctx, year, month)
		if err != nil {
			return nil, err
		}

		results = append(results, *mrr)
		startDate = startDate.AddDate(0, 1, 0)
	}

	return results, nil
}

// ==================== Churn Metrics Methods ====================

// CalculateChurnMetrics calculates churn metrics for a given month
func (r *BillingRepository) CalculateChurnMetrics(ctx context.Context, year, month int) (*ChurnMetrics, error) {
	periodMonth := fmt.Sprintf("%04d-%02d", year, month)
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	metrics := &ChurnMetrics{
		PeriodMonth:     periodMonth,
		CalculationDate: time.Now(),
	}

	// Customers at start of period
	customersAtStartQuery := `
		SELECT COUNT(DISTINCT tenant_id)
		FROM bundle_subscriptions
		WHERE status = 'active'
		AND created_at < $1
		AND (canceled_at IS NULL OR canceled_at >= $1)
	`

	err := r.db.GORM.Raw(customersAtStartQuery, startDate).Scan(&metrics.CustomersAtStart).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count customers at start: %w", err)
	}

	// Churned customers in this period
	churnedQuery := `
		SELECT COUNT(DISTINCT tenant_id)
		FROM subscription_churn_events
		WHERE event_type = 'cancellation'
		AND event_date >= $1 AND event_date < $2
		AND is_recovered = false
	`

	err = r.db.GORM.Raw(churnedQuery, startDate, endDate).Scan(&metrics.CustomersChurned).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count churned customers: %w", err)
	}

	// Calculate customer churn rate
	if metrics.CustomersAtStart > 0 {
		metrics.CustomerChurnRate = float64(metrics.CustomersChurned) / float64(metrics.CustomersAtStart) * 100.0
	}

	// MRR at start of period
	mrrAtStartQuery := `
		SELECT COALESCE(SUM(b.display_price_cents), 0)
		FROM bundle_subscriptions bs
		JOIN pricing_bundles b ON bs.bundle_id = b.id
		WHERE bs.status = 'active'
		AND bs.created_at < $1
		AND (bs.canceled_at IS NULL OR bs.canceled_at >= $1)
	`

	err = r.db.GORM.Raw(mrrAtStartQuery, startDate).Scan(&metrics.MRRAtStart_Cents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate MRR at start: %w", err)
	}

	// MRR churned
	mrrChurnedQuery := `
		SELECT COALESCE(SUM(mrr_lost_cents), 0)
		FROM subscription_churn_events
		WHERE event_date >= $1 AND event_date < $2
		AND event_type IN ('cancellation', 'downgrade')
		AND is_recovered = false
	`

	err = r.db.GORM.Raw(mrrChurnedQuery, startDate, endDate).Scan(&metrics.MRRChurned_Cents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate churned MRR: %w", err)
	}

	// Calculate revenue churn rate
	if metrics.MRRAtStart_Cents > 0 {
		metrics.RevenueChurnRate = float64(metrics.MRRChurned_Cents) / float64(metrics.MRRAtStart_Cents) * 100.0
	}

	// Gross churn (downgrades + cancellations without expansion)
	grossChurnQuery := `
		SELECT COALESCE(SUM(mrr_lost_cents), 0)
		FROM subscription_churn_events
		WHERE event_date >= $1 AND event_date < $2
		AND event_type IN ('cancellation', 'downgrade')
	`

	var grossChurnCents int
	err = r.db.GORM.Raw(grossChurnQuery, startDate, endDate).Scan(&grossChurnCents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate gross churn: %w", err)
	}

	if metrics.MRRAtStart_Cents > 0 {
		metrics.GrossChurnRate = float64(grossChurnCents) / float64(metrics.MRRAtStart_Cents) * 100.0
		metrics.GrossRevenueRetention = 100.0 - metrics.GrossChurnRate
	}

	// Calculate expansion MRR for net churn calculation
	expansionQuery := `
		SELECT COALESCE(SUM(sce.new_mrr_cents - sce.previous_mrr_cents), 0)
		FROM subscription_churn_events sce
		WHERE sce.event_type = 'upgrade'
		AND sce.event_date >= $1 AND sce.event_date < $2
	`
	var expansionMRRCents int
	err = r.db.GORM.Raw(expansionQuery, startDate, endDate).Scan(&expansionMRRCents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate expansion MRR: %w", err)
	}

	// Net churn (including expansion)
	netChurnCents := grossChurnCents - expansionMRRCents
	if metrics.MRRAtStart_Cents > 0 {
		metrics.NetChurnRate = float64(netChurnCents) / float64(metrics.MRRAtStart_Cents) * 100.0
		metrics.NetRevenueRetention = 100.0 - metrics.NetChurnRate
	}

	// Breakdown by churn type
	// Voluntary churn
	voluntaryQuery := `
		SELECT COUNT(*)
		FROM subscription_churn_events
		WHERE event_type = 'cancellation'
		AND event_date >= $1 AND event_date < $2
		AND cancel_reason NOT LIKE '%payment%'
		AND cancel_reason NOT LIKE '%failed%'
	`

	err = r.db.GORM.Raw(voluntaryQuery, startDate, endDate).Scan(&metrics.VoluntaryChurn).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count voluntary churn: %w", err)
	}

	// Involuntary churn (payment failures)
	involuntaryQuery := `
		SELECT COUNT(*)
		FROM subscription_churn_events
		WHERE event_type IN ('failed_renewal', 'payment_failure')
		AND event_date >= $1 AND event_date < $2
	`

	err = r.db.GORM.Raw(involuntaryQuery, startDate, endDate).Scan(&metrics.InvoluntaryChurn).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count involuntary churn: %w", err)
	}

	// Downgrades
	downgradeQuery := `
		SELECT COUNT(*)
		FROM subscription_churn_events
		WHERE event_type = 'downgrade'
		AND event_date >= $1 AND event_date < $2
	`

	err = r.db.GORM.Raw(downgradeQuery, startDate, endDate).Scan(&metrics.DowngradeChurn).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count downgrades: %w", err)
	}

	// Failed renewals
	failedRenewalQuery := `
		SELECT COUNT(*)
		FROM subscription_churn_events
		WHERE event_type = 'failed_renewal'
		AND event_date >= $1 AND event_date < $2
	`

	err = r.db.GORM.Raw(failedRenewalQuery, startDate, endDate).Scan(&metrics.FailedRenewalChurn).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count failed renewals: %w", err)
	}

	return metrics, nil
}

// RecordSubscriptionChurnEvent records a churn event for tracking
func (r *BillingRepository) RecordSubscriptionChurnEvent(ctx context.Context, event *SubscriptionChurnEvent) error {
	return r.db.GORM.WithContext(ctx).Create(event).Error
}

// GetChurnMetricsTimeseries returns churn metrics over time
func (r *BillingRepository) GetChurnMetricsTimeseries(ctx context.Context, months int) ([]ChurnMetrics, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0)

	results := make([]ChurnMetrics, 0, months)

	for i := 0; i < months; i++ {
		year := startDate.Year()
		month := int(startDate.Month())

		churn, err := r.CalculateChurnMetrics(ctx, year, month)
		if err != nil {
			return nil, err
		}

		results = append(results, *churn)
		startDate = startDate.AddDate(0, 1, 0)
	}

	return results, nil
}

// ==================== Cohort Analysis Methods ====================

// GetCohortRetention returns cohort retention data for analysis
func (r *BillingRepository) GetCohortRetention(ctx context.Context, cohortMonths int) ([]CohortRetention, error) {
	query := `
		SELECT
			TO_CHAR(DATE_TRUNC('month', bs.created_at), 'YYYY-MM') as cohort_month,
			COUNT(DISTINCT bs.tenant_id) as cohort_size,
			generate_series(0, 12) as period_index
		FROM bundle_subscriptions bs
		WHERE bs.created_at >= DATE_TRUNC('month', NOW() - INTERVAL '%d months')
		AND bs.status = 'active'
		GROUP BY cohort_month, period_index
		ORDER BY cohort_month, period_index
	`

	// Note: This is a simplified version. A full cohort analysis would require
	// more complex queries to track customers across months.

	_ = query // suppress unused warning

	// Placeholder implementation
	return []CohortRetention{}, nil
}

// GetLTVMetrics calculates lifetime value metrics
func (r *BillingRepository) GetLTVMetrics(ctx context.Context) (*LTVMetrics, error) {
	ltv := &LTVMetrics{
		CalculationDate: time.Now(),
	}

	// Average LTV calculation
	avgLTVQuery := `
		SELECT AVG(total_paid)
		FROM (
			SELECT tenant_id, SUM(amount_paid_cents) as total_paid
			FROM invoices
			WHERE paid_at IS NOT NULL
			GROUP BY tenant_id
		) as customer_totals
	`

	err := r.db.GORM.Raw(avgLTVQuery).Scan(&ltv.AverageLTV_Cents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate average LTV: %w", err)
	}

	// Median LTV calculation (simplified)
	medianLTVQuery := `
		SELECT PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY total_paid)
		FROM (
			SELECT tenant_id, SUM(amount_paid_cents) as total_paid
			FROM invoices
			WHERE paid_at IS NOT NULL
			GROUP BY tenant_id
		) as customer_totals
	`

	err = r.db.GORM.Raw(medianLTVQuery).Scan(&ltv.MedianLTV_Cents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate median LTV: %w", err)
	}

	return ltv, nil
}

// ==================== Financial Reporting Methods ====================

// GenerateFinancialReport generates a comprehensive financial report
func (r *BillingRepository) GenerateFinancialReport(ctx context.Context, reportType string, periodStart, periodEnd time.Time) (*FinancialReport, error) {
	report := &FinancialReport{
		ReportID:    uuid.New(),
		GeneratedAt: time.Now(),
		ReportType:  reportType,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}

	// Total revenue
	revenueQuery := `
		SELECT COALESCE(SUM(amount_paid_cents), 0)
		FROM invoices
		WHERE paid_at >= $1 AND paid_at < $2
		AND status = 'paid'
	`

	err := r.db.GORM.Raw(revenueQuery, periodStart, periodEnd).Scan(&report.TotalRevenue_Cents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total revenue: %w", err)
	}

	// Subscription revenue
	subRevenueQuery := `
		SELECT COALESCE(SUM(amount_paid_cents), 0)
		FROM invoices
		WHERE paid_at >= $1 AND paid_at < $2
		AND status = 'paid'
		AND subscription_id IS NOT NULL
	`

	err = r.db.GORM.Raw(subRevenueQuery, periodStart, periodEnd).Scan(&report.SubscriptionRevenue_Cents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate subscription revenue: %w", err)
	}

	// One-time revenue
	report.OneTimeRevenue_Cents = report.TotalRevenue_Cents - report.SubscriptionRevenue_Cents

	// Refunds
	refundsQuery := `
		SELECT COALESCE(SUM(amount_cents), 0)
		FROM payment_refunds
		WHERE created_at >= $1 AND created_at < $2
		AND status = 'succeeded'
	`

	err = r.db.GORM.Raw(refundsQuery, periodStart, periodEnd).Scan(&report.Refunds_Cents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate refunds: %w", err)
	}

	// Disputes (chargebacks)
	disputesQuery := `
		SELECT COALESCE(SUM(amount_cents), 0)
		FROM payment_disputes
		WHERE created_at >= $1 AND created_at < $2
		AND status = 'lost'
	`

	err = r.db.GORM.Raw(disputesQuery, periodStart, periodEnd).Scan(&report.Disputes_Cents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate disputes: %w", err)
	}

	// Outstanding invoices
	outstandingQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN status IN ('open', 'pending') THEN amount_due_cents - amount_paid_cents ELSE 0 END), 0) as outstanding,
			COALESCE(SUM(CASE WHEN status = 'overdue' THEN amount_due_cents - amount_paid_cents ELSE 0 END), 0) as overdue
		FROM invoices
		WHERE due_date < $2
	`

	var outstanding, overdue int
	err = r.db.GORM.Raw(outstandingQuery, periodStart, periodEnd).Scan(&[]interface{}{&outstanding, &overdue}).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate outstanding: %w", err)
	}

	report.OutstandingInvoices_Cents = outstanding
	report.OverdueInvoices_Cents = overdue

	// Tax collected
	taxQuery := `
		SELECT COALESCE(SUM(tax_amount_cents), 0)
		FROM invoice_tax_details itd
		JOIN invoices i ON itd.invoice_id = i.id
		WHERE i.paid_at >= $1 AND i.paid_at < $2
	`

	err = r.db.GORM.Raw(taxQuery, periodStart, periodEnd).Scan(&report.TaxCollected_Cents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate tax: %w", err)
	}

	return report, nil
}

// GetTaxJurisdictionReport returns tax collection by jurisdiction
func (r *BillingRepository) GetTaxJurisdictionReport(ctx context.Context, periodMonth string) ([]TaxJurisdictionReport, error) {
	query := `
		SELECT
			tr.country,
			tr.state,
			tr.tax_type as jurisdiction_type,
			tr.display_name as jurisdiction,
			tr.percentage as tax_rate,
			COUNT(*) as transactions_count,
			COALESCE(SUM(itd.taxable_amount_cents), 0) as taxable_revenue_cents,
			COALESCE(SUM(itd.tax_amount_cents), 0) as tax_collected_cents
		FROM invoice_tax_details itd
		JOIN tax_rates tr ON itd.tax_rate_id = tr.id
		JOIN invoices i ON itd.invoice_id = i.id
		WHERE TO_CHAR(i.paid_at, 'YYYY-MM') = $1
		GROUP BY tr.country, tr.state, tr.tax_type, tr.display_name, tr.percentage
		ORDER BY tr.country, tr.state
	`

	results := []TaxJurisdictionReport{}
	err := r.db.GORM.Raw(query, periodMonth).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get tax jurisdiction report: %w", err)
	}

	return results, nil
}

// GetSubscriptionLifecycleMetrics calculates lifecycle metrics
func (r *BillingRepository) GetSubscriptionLifecycleMetrics(ctx context.Context, year, month int) (*SubscriptionLifecycleMetrics, error) {
	periodMonth := fmt.Sprintf("%04d-%02d", year, month)
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	metrics := &SubscriptionLifecycleMetrics{
		PeriodMonth: periodMonth,
	}

	// Trials started
	trialsQuery := `
		SELECT COUNT(*)
		FROM subscriptions
		WHERE trial_end IS NOT NULL
		AND created_at >= $1 AND created_at < $2
	`

	err := r.db.GORM.Raw(trialsQuery, startDate, endDate).Scan(&metrics.TrialsStarted).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count trials: %w", err)
	}

	// Trial conversions
	conversionsQuery := `
		SELECT COUNT(DISTINCT s.id)
		FROM subscriptions s
		JOIN invoices i ON i.subscription_id = s.id
		WHERE s.trial_end IS NOT NULL
		AND i.paid_at >= s.trial_end
		AND i.paid_at >= $1 AND i.paid_at < $2
	`

	err = r.db.GORM.Raw(conversionsQuery, startDate, endDate).Scan(&metrics.TrialsConverted).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count conversions: %w", err)
	}

	// Conversion rate
	if metrics.TrialsStarted > 0 {
		metrics.TrialConversionRate = float64(metrics.TrialsConverted) / float64(metrics.TrialsStarted) * 100.0
	}

	return metrics, nil
}
