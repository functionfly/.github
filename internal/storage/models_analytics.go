package storage

import (
	"time"

	"github.com/google/uuid"
)

// ==================== Revenue Analytics Types ====================

// MRRMetrics represents Monthly Recurring Revenue metrics
type MRRMetrics struct {
	MRR_Cents             int       `json:"mrr_cents"`
	MRR_USD               float64   `json:"mrr_usd"`
	NewMRR_Cents          int       `json:"new_mrr_cents"`
	ExpansionMRR_Cents    int       `json:"expansion_mrr_cents"`
	ContractionMRR_Cents  int       `json:"contraction_mrr_cents"`
	ChurnedMRR_Cents      int       `json:"churned_mrr_cents"`
	ReactivationMRR_Cents int       `json:"reactivation_mrr_cents"`
	NetMRRChange_Cents    int       `json:"net_mrr_change_cents"`
	CalculationDate       time.Time `json:"calculation_date"`
	PeriodMonth           string    `json:"period_month"` // YYYY-MM
}

// ARRMetrics represents Annual Recurring Revenue metrics
type ARRMetrics struct {
	ARR_USD               float64   `json:"arr_usd"`
	ARR_Cents             int       `json:"arr_cents"`
	NewARR_Cents          int       `json:"new_arr_cents"`
	ExpansionARR_Cents    int       `json:"expansion_arr_cents"`
	ContractionARR_Cents  int       `json:"contraction_arr_cents"`
	ChurnedARR_Cents      int       `json:"churned_arr_cents"`
	ReactivationARR_Cents int       `json:"reactivation_arr_cents"`
	CalculationDate       time.Time `json:"calculation_date"`
}

// MRRCohortData represents MRR data for cohort analysis
type MRRCohortData struct {
	CohortMonth         string  `json:"cohort_month"` // YYYY-MM
	StartingMRR_Cents   int     `json:"starting_mrr_cents"`
	CurrentMRR_Cents    int     `json:"current_mrr_cents"`
	RetentionRate       float64 `json:"retention_rate"`
	MonthsSinceCohort   int     `json:"months_since_cohort"`
	CustomerCount       int     `json:"customer_count"`
	ActiveCustomerCount int     `json:"active_customer_count"`
}

// ==================== Churn Analytics Types ====================

// ChurnMetrics represents churn analysis metrics
type ChurnMetrics struct {
	CalculationDate time.Time `json:"calculation_date"`
	PeriodMonth     string    `json:"period_month"` // YYYY-MM

	// Customer Churn
	CustomersAtStart  int     `json:"customers_at_start"`
	CustomersChurned  int     `json:"customers_churned"`
	CustomerChurnRate float64 `json:"customer_churn_rate"` // percentage

	// Revenue Churn
	MRRAtStart_Cents int     `json:"mrr_at_start_cents"`
	MRRChurned_Cents int     `json:"mrr_churned_cents"`
	RevenueChurnRate float64 `json:"revenue_churn_rate"` // percentage

	// Gross vs Net
	GrossChurnRate        float64 `json:"gross_churn_rate"`
	NetChurnRate          float64 `json:"net_churn_rate"`
	NetRevenueRetention   float64 `json:"net_revenue_retention"`   // NRR percentage
	GrossRevenueRetention float64 `json:"gross_revenue_retention"` // GRR percentage

	// Types of churn
	VoluntaryChurn     int `json:"voluntary_churn"`
	InvoluntaryChurn   int `json:"involuntary_churn"`
	DowngradeChurn     int `json:"downgrade_churn"`
	FailedRenewalChurn int `json:"failed_renewal_churn"`

	// Cohorts
	ChurnedCustomerIDs []uuid.UUID `json:"churned_customer_ids,omitempty"`
}

// SubscriptionChurnEvent represents a single churn event for detailed tracking
type SubscriptionChurnEvent struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID       uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	SubscriptionID uuid.UUID  `json:"subscription_id" gorm:"type:uuid;not null;index"`
	BundleID       *uuid.UUID `json:"bundle_id,omitempty" gorm:"type:uuid;index"`

	EventType   string    `json:"event_type" gorm:"not null;size:50"` // 'cancellation', 'downgrade', 'failed_renewal', 'payment_failure'
	EventDate   time.Time `json:"event_date" gorm:"not null"`
	CohortMonth string    `json:"cohort_month" gorm:"size:7;index"` // YYYY-MM

	// Financial impact
	PreviousMRR_Cents int `json:"previous_mrr_cents"`
	NewMRR_Cents      int `json:"new_mrr_cents"`
	MRRLost_Cents     int `json:"mrr_lost_cents"`

	// Downgrade specific
	PreviousBundleID *uuid.UUID `json:"previous_bundle_id,omitempty" gorm:"type:uuid"`
	NewBundleID      *uuid.UUID `json:"new_bundle_id,omitempty" gorm:"type:uuid"`

	// Cancellation specific
	CancelReason      string     `json:"cancel_reason,omitempty" gorm:"size:255"`
	CancelAtPeriodEnd bool       `json:"cancel_at_period_end"`
	CancelDate        *time.Time `json:"cancel_date,omitempty"`

	// Payment failure specific
	FailedPaymentAmount_Cents int    `json:"failed_payment_amount_cents"`
	FailureCode               string `json:"failure_code,omitempty" gorm:"size:100"`
	AttemptCount              int    `json:"attempt_count"`

	// Stripe reference
	StripeSubscriptionID string `json:"stripe_subscription_id,omitempty" gorm:"size:255"`
	StripeEventID        string `json:"stripe_event_id,omitempty" gorm:"size:255"`

	IsRecovered bool       `json:"is_recovered" gorm:"default:false"` // true if customer reactivated
	RecoveredAt *time.Time `json:"recovered_at,omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// CohortRetention represents monthly cohort retention data
type CohortRetention struct {
	CohortMonth     string    `json:"cohort_month"` // YYYY-MM
	CohortSize      int       `json:"cohort_size"`
	PeriodIndex     int       `json:"period_index"` // 0 = first month, 1 = second month, etc.
	PeriodDate      time.Time `json:"period_date"`
	ActiveCustomers int       `json:"active_customers"`
	RetentionRate   float64   `json:"retention_rate"`
	Revenue_Cents   int       `json:"revenue_cents"`
}

// ==================== Financial Reporting Types ====================

// FinancialReport represents a comprehensive financial report for accounting
type FinancialReport struct {
	ReportID    uuid.UUID `json:"report_id"`
	GeneratedAt time.Time `json:"generated_at"`
	ReportType  string    `json:"report_type"` // 'cash', 'accrual', 'tax'
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`

	// Revenue summary
	TotalRevenue_Cents        int `json:"total_revenue_cents"`
	SubscriptionRevenue_Cents int `json:"subscription_revenue_cents"`
	UsageRevenue_Cents        int `json:"usage_revenue_cents"`
	OneTimeRevenue_Cents      int `json:"one_time_revenue_cents"`

	// Breakdown by type
	MRR_Cents int `json:"mrr_cents"`
	ARR_Cents int `json:"arr_cents"`

	// Adjustments
	Refunds_Cents         int `json:"refunds_cents"`
	Disputes_Cents        int `json:"disputes_cents"`
	CreditsIssued_Cents   int `json:"credits_issued_cents"`
	CouponsRedeemed_Cents int `json:"coupons_redeemed_cents"`

	// Outstanding
	OutstandingInvoices_Cents int `json:"outstanding_invoices_cents"`
	OverdueInvoices_Cents     int `json:"overdue_invoices_cents"`

	// Tax
	TaxCollected_Cents int            `json:"tax_collected_cents"`
	TaxByJurisdiction  map[string]int `json:"tax_by_jurisdiction,omitempty"`
}

// RevenueRecognitionEntry represents revenue recognition data for accrual accounting
type RevenueRecognitionEntry struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	InvoiceID uuid.UUID `json:"invoice_id" gorm:"type:uuid;not null;index"`
	TenantID  uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`

	// Period
	RecognitionMonth string    `json:"recognition_month" gorm:"size:7;not null;index"` // YYYY-MM
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`

	// Amounts
	TotalInvoice_Cents int `json:"total_invoice_cents"`
	Recognized_Cents   int `json:"recognized_cents"`
	Deferred_Cents     int `json:"deferred_cents"`

	// Recognition type
	RevenueType string `json:"revenue_type" gorm:"size:50"` // 'subscription', 'usage', 'one_time'

	// Status
	IsRecognized bool       `json:"is_recognized" gorm:"default:false"`
	RecognizedAt *time.Time `json:"recognized_at,omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TaxJurisdictionReport represents tax collection by jurisdiction
type TaxJurisdictionReport struct {
	Jurisdiction     string `json:"jurisdiction"`      // Country or State code
	JurisdictionType string `json:"jurisdiction_type"` // 'country', 'state', 'province'
	Country          string `json:"country"`
	State            string `json:"state,omitempty"`

	PeriodMonth string `json:"period_month"` // YYYY-MM

	// Tax details
	TaxRate                 float64 `json:"tax_rate"`
	TransactionsCount       int     `json:"transactions_count"`
	TaxableRevenue_Cents    int     `json:"taxable_revenue_cents"`
	TaxCollected_Cents      int     `json:"tax_collected_cents"`
	NonTaxableRevenue_Cents int     `json:"non_taxable_revenue_cents"`
	ExemptRevenue_Cents     int     `json:"exempt_revenue_cents"`

	// Breakdown
	ByTaxRate map[float64]int `json:"by_tax_rate,omitempty"` // map[tax_rate]cents
}

// ==================== Dashboard Analytics Types ====================

// SubscriptionLifecycleMetrics tracks conversions and upgrades
type SubscriptionLifecycleMetrics struct {
	PeriodMonth string `json:"period_month"`

	// Trial to paid
	TrialsStarted       int     `json:"trials_started"`
	TrialsConverted     int     `json:"trials_converted"`
	TrialConversionRate float64 `json:"trial_conversion_rate"`

	// Upgrades/Downgrades
	Upgrades                 int `json:"upgrades"`
	Downgrades               int `json:"downgrades"`
	ExpansionRevenue_Cents   int `json:"expansion_revenue_cents"`
	ContractionRevenue_Cents int `json:"contraction_revenue_cents"`

	// Plan distribution
	PlanDistribution map[string]int `json:"plan_distribution,omitempty"` // plan_name -> count
}

// LTVMetrics represents customer lifetime value metrics
type LTVMetrics struct {
	CalculationDate  time.Time `json:"calculation_date"`
	AverageLTV_Cents int       `json:"average_ltv_cents"`
	MedianLTV_Cents  int       `json:"median_ltv_cents"`
	LTVCACRatio      float64   `json:"ltv_cac_ratio,omitempty"`

	// By cohort
	LTVByCohort map[string]int `json:"ltv_by_cohort,omitempty"` // cohort_month -> avg LTV

	// Payback period
	AveragePaybackMonths float64 `json:"average_payback_months"`
}

// FailedPaymentMetrics tracks dunning and recovery
type FailedPaymentMetrics struct {
	PeriodMonth string `json:"period_month"`

	// Dunning
	FailedPayments    int     `json:"failed_payments"`
	SuccessfulRetries int     `json:"successful_retries"`
	RetrySuccessRate  float64 `json:"retry_success_rate"`

	// Recovery
	InvoluntaryChurn   int     `json:"involuntary_churn"`
	RecoveredCustomers int     `json:"recovered_customers"`
	RecoveryRate       float64 `json:"recovery_rate"`

	// Aging
	FailedPaymentByAge map[string]int `json:"failed_payment_by_age,omitempty"` // '0-7', '8-14', '15-30', '30+' -> count
}
