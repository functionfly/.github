package storage

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PricingTier represents a billing pricing tier
type PricingTier struct {
	ID               uuid.UUID   `json:"id"`
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	PriceCents       int         `json:"price_cents"`
	AnnualPriceCents *int        `json:"annual_price_cents,omitempty"`
	Currency         string      `json:"currency"`
	BillingCycle     string      `json:"billing_cycle"` // 'monthly' or 'annual'
	Features         interface{} `json:"features"`     // JSON features/limits
	IsActive         bool        `json:"is_active"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

// Subscription represents a tenant's subscription
type Subscription struct {
	ID                   uuid.UUID    `json:"id"`
	TenantID             uuid.UUID    `json:"tenant_id"`
	PricingTierID        uuid.UUID    `json:"pricing_tier_id"`
	Status               string       `json:"status"`
	BillingCycle         string       `json:"billing_cycle"` // 'monthly' or 'annual'
	StripeSubscriptionID string       `json:"stripe_subscription_id,omitempty"`
	CurrentPeriodStart   time.Time    `json:"current_period_start"`
	CurrentPeriodEnd     time.Time    `json:"current_period_end"`
	TrialEnd             *time.Time   `json:"trial_end,omitempty"`
	CancelAtPeriodEnd    bool         `json:"cancel_at_period_end"`
	CanceledAt           *time.Time   `json:"canceled_at,omitempty"`
	CreatedAt            time.Time    `json:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at"`
	PricingTier          *PricingTier `json:"pricing_tier,omitempty"` // Populated in queries
}

// Invoice represents a billing invoice
type Invoice struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	SubscriptionID    *uuid.UUID `json:"subscription_id,omitempty"`
	Status            string     `json:"status"`
	AmountDueCents    int        `json:"amount_due_cents"`
	AmountPaidCents   int        `json:"amount_paid_cents"`
	Currency          string     `json:"currency"`
	StripeInvoiceID   *string    `json:"stripe_invoice_id,omitempty"`
	ExternalReference *string    `json:"external_reference,omitempty"`
	InvoicePdfURL     string     `json:"invoice_pdf_url,omitempty"`
	HostedInvoiceURL  string     `json:"hosted_invoice_url,omitempty"`
	PeriodStart       *time.Time `json:"period_start,omitempty"`
	PeriodEnd         *time.Time `json:"period_end,omitempty"`
	DueDate           *time.Time `json:"due_date,omitempty"`
	PaidAt            *time.Time `json:"paid_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// UsageEvent represents a usage event for billing
type UsageEvent struct {
	ID             uuid.UUID   `json:"id"`
	TenantID       uuid.UUID   `json:"tenant_id"`
	EventType      string      `json:"event_type"`
	Quantity       int         `json:"quantity"`
	UnitPriceCents *int        `json:"unit_price_cents,omitempty"`
	Metadata       interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time   `json:"timestamp"`
}

// UsageRollup represents aggregated usage data
type UsageRollup struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	EventType     string    `json:"event_type"`
	PeriodDate    time.Time `json:"period_date"`
	TotalQuantity int       `json:"total_quantity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Coupon represents a discount coupon
type Coupon struct {
	ID             uuid.UUID  `json:"id"`
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	DiscountType   string     `json:"discount_type"` // 'percent' or 'amount'
	DiscountValue  int        `json:"discount_value"`
	MaxRedemptions *int       `json:"max_redemptions,omitempty"`
	TimesRedeemed  int        `json:"times_redeemed"`
	ValidFrom      *time.Time `json:"valid_from,omitempty"`
	ValidUntil     *time.Time `json:"valid_until,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CouponRedemption represents a coupon redemption
type CouponRedemption struct {
	ID             uuid.UUID  `json:"id"`
	CouponID       uuid.UUID  `json:"coupon_id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	SubscriptionID *uuid.UUID `json:"subscription_id,omitempty"`
	RedeemedAt     time.Time  `json:"redeemed_at"`
	Coupon         *Coupon    `json:"coupon,omitempty"` // Populated in queries
}

// CostAllocationEntry represents a detailed cost allocation record for a function execution
// This enables fine-grained chargebacks and cost transparency

type CostAllocationEntry struct {
	ID               uuid.UUID `json:"id"`
	TenantID         uuid.UUID `json:"tenant_id"`
	APIKeyID         uuid.UUID `json:"api_key_id,omitempty"` // Optional: tracks which API key was used
	FunctionID       uuid.UUID `json:"function_id"`
	FunctionName     string    `json:"function_name"`
	FunctionAuthor   string    `json:"function_author"`
	ExecutionID      uuid.UUID `json:"execution_id"`
	ExecutionOutcome string    `json:"execution_outcome"`
	Cached           bool      `json:"cached"`

	// Resource usage
	DurationMs   int64 `json:"duration_ms"`
	CPUTimeMs    int64 `json:"cpu_time_ms"`
	MemoryUsedMB int64 `json:"memory_used_mb"`
	WallTimeMs   int64 `json:"wall_time_ms"`

	// Cost breakdown (in cents for precision)
	ExecutionCostCents int64 `json:"execution_cost_cents"`
	ComputeCostCents   int64 `json:"compute_cost_cents"`
	PlatformFeeCents   int64 `json:"platform_fee_cents"`
	DataTransferCents  int64 `json:"data_transfer_cents"`
	TotalCostCents     int64 `json:"total_cost_cents"`

	// Metadata
	Region      string                 `json:"region"`
	Timestamp   time.Time              `json:"timestamp"`
	PeriodStart time.Time              `json:"period_start"`
	PeriodEnd   time.Time              `json:"period_end"`
	Tags        map[string]string      `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// CostAllocationSummary provides aggregated cost data by function
type CostAllocationSummary struct {
	FunctionID     uuid.UUID `json:"function_id"`
	FunctionName   string    `json:"function_name"`
	FunctionAuthor string    `json:"function_author"`

	// Execution counts
	TotalExecutions   int64 `json:"total_executions"`
	SuccessExecutions int64 `json:"success_executions"`
	ErrorExecutions   int64 `json:"error_executions"`
	CachedExecutions  int64 `json:"cached_executions"`

	// Resource totals
	TotalDurationMs   int64 `json:"total_duration_ms"`
	TotalCPUTimeMs    int64 `json:"total_cpu_time_ms"`
	TotalMemoryUsedMB int64 `json:"total_memory_used_mb"`

	// Cost totals (in cents)
	ExecutionCostCents int64 `json:"execution_cost_cents"`
	ComputeCostCents   int64 `json:"compute_cost_cents"`
	PlatformFeeCents   int64 `json:"platform_fee_cents"`
	DataTransferCents  int64 `json:"data_transfer_cents"`
	TotalCostCents     int64 `json:"total_cost_cents"`

	// Averages
	AvgDurationMs float64 `json:"avg_duration_ms"`
	AvgCostCents  float64 `json:"avg_cost_cents"`
}

// TenantCostSummary provides tenant-level cost aggregation
type TenantCostSummary struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	TenantName string    `json:"tenant_name,omitempty"`

	// Period
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`

	// Executions
	TotalExecutions int64 `json:"total_executions"`
	UniqueFunctions int   `json:"unique_functions"`

	// Costs (in cents)
	ExecutionCostCents int64 `json:"execution_cost_cents"`
	ComputeCostCents   int64 `json:"compute_cost_cents"`
	PlatformFeeCents   int64 `json:"platform_fee_cents"`
	DataTransferCents  int64 `json:"data_transfer_cents"`
	TotalCostCents     int64 `json:"total_cost_cents"`

	// Function breakdown
	FunctionSummaries []CostAllocationSummary `json:"function_summaries,omitempty"`

	// Daily breakdown
	DailyBreakdown []DailyCostBreakdown `json:"daily_breakdown,omitempty"`
}

// DailyCostBreakdown provides daily cost aggregation
type DailyCostBreakdown struct {
	Date       time.Time `json:"date"`
	Executions int64     `json:"executions"`
	CostCents  int64     `json:"cost_cents"`
}

// CostAllocationFilter provides filtering options for cost queries
type CostAllocationFilter struct {
	TenantID     *uuid.UUID
	FunctionID   *uuid.UUID
	FunctionName *string
	Author       *string
	StartDate    *time.Time
	EndDate      *time.Time
	Outcome      *string
	Cached       *bool
	Region       *string
	MinCostCents *int64
	MaxCostCents *int64
	Tags         map[string]string
}

// PricingBundle represents a Backend-in-a-Box pricing bundle
type PricingBundle struct {
	ID                    uuid.UUID      `json:"id"`
	Slug                  string         `json:"slug"`
	Name                  string         `json:"name"`
	DisplayName           string         `json:"display_name"`
	Description           string         `json:"description"`
	ShortDescription      string         `json:"short_description"`
	DisplayPriceCents     int            `json:"display_price_cents"`
	BillingInterval       string         `json:"billing_interval"`
	StripePriceID         string         `json:"stripe_price_id"`
	Icon                  string         `json:"icon"`
	Color                 string         `json:"color"`
	FeaturesIncluded      []string       `json:"features_included"`
	FeatureLimits         map[string]int `json:"feature_limits"`
	ProvisioningTemplates []string       `json:"provisioning_templates"`
	SortOrder             int            `json:"sort_order"`
	IsActive              bool           `json:"is_active"`
	IsPopular             bool           `json:"is_popular"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

// FounderModeRegistration represents a founder mode (free until trigger) registration
type FounderModeRegistration struct {
	ID                   uuid.UUID  `json:"id"`
	TenantID             uuid.UUID  `json:"tenant_id"`
	BundleID             uuid.UUID  `json:"bundle_id"`
	ModeType             string     `json:"mode_type"` // 'time_based', 'revenue_based', 'hybrid'
	StartedAt            time.Time  `json:"started_at"`
	EndsAt               *time.Time `json:"ends_at,omitempty"`
	FreeDays             int        `json:"free_days"`
	MRRThresholdCents    int        `json:"mrr_threshold_cents"`
	Status               string     `json:"status"` // 'active', 'grace_period', 'converted', 'expired', 'canceled'
	ConvertedToBundleID  *uuid.UUID `json:"converted_to_bundle_id,omitempty"`
	ConvertedAt          *time.Time `json:"converted_at,omitempty"`
	StripeSubscriptionID string     `json:"stripe_subscription_id,omitempty"`
	GracePeriodStartedAt *time.Time `json:"grace_period_started_at,omitempty"`
	GracePeriodEndsAt    *time.Time `json:"grace_period_ends_at,omitempty"`
	MaxUsersSeen         int        `json:"max_users_seen"`
	MaxMRRSeenCents      int        `json:"max_mrr_seen_cents"`
	MaxAPICallsMonthly   int        `json:"max_api_calls_monthly"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// BundleSubscription represents a subscription to a Backend-in-a-Box bundle
type BundleSubscription struct {
	ID                       uuid.UUID  `json:"id"`
	TenantID                 uuid.UUID  `json:"tenant_id"`
	BundleID                 uuid.UUID  `json:"bundle_id"`
	FounderModeID            *uuid.UUID `json:"founder_mode_id,omitempty"`
	ConvertedFromFounderMode bool       `json:"converted_from_founder_mode"`
	Status                   string     `json:"status"` // 'active', 'deferred', 'canceled', 'past_due'
	StripeSubscriptionID     string     `json:"stripe_subscription_id,omitempty"`
	DefaultAppID             *uuid.UUID `json:"default_app_id,omitempty"` // App created during bundle provisioning
	CurrentPeriodStart       time.Time  `json:"current_period_start"`
	CurrentPeriodEnd         time.Time  `json:"current_period_end"`
	CancelAtPeriodEnd        bool       `json:"cancel_at_period_end"`
	CanceledAt               *time.Time `json:"canceled_at,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

// DeferredBillingConfig represents configuration for deferred billing triggers
type DeferredBillingConfig struct {
	ID                      uuid.UUID  `json:"id"`
	BundleID                uuid.UUID  `json:"bundle_id"`
	IsDefault               bool       `json:"is_default"`
	TriggerUserCount        *int       `json:"trigger_user_count,omitempty"`
	TriggerRevenueCents     *int       `json:"trigger_revenue_cents,omitempty"`
	TriggerAPICalls         *int       `json:"trigger_api_calls,omitempty"`
	TriggerDaysElapsed      *int       `json:"trigger_days_elapsed,omitempty"`
	GracePeriodDays         int        `json:"grace_period_days"`
	ConvertToBundleID       *uuid.UUID `json:"convert_to_bundle_id,omitempty"`
	WarningEmailTemplate    string     `json:"warning_email_template"`
	TriggerEmailTemplate    string     `json:"trigger_email_template"`
	ConversionEmailTemplate string     `json:"conversion_email_template"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// CostAllocationReport represents a comprehensive cost report
type CostAllocationReport struct {
	ReportID    uuid.UUID `json:"report_id"`
	GeneratedAt time.Time `json:"generated_at"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`

	// Summary
	TenantCount     int   `json:"tenant_count"`
	FunctionCount   int   `json:"function_count"`
	TotalExecutions int64 `json:"total_executions"`
	TotalCostCents  int64 `json:"total_cost_cents"`

	// Chargeback data for internal billing
	ChargebackEntries []CostAllocationChargeback `json:"chargeback_entries"`

	// Detailed data
	TenantSummaries []TenantCostSummary `json:"tenant_summaries"`
}

// CostAllocationChargeback represents a chargeback entry for internal billing
type CostAllocationChargeback struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	TenantName     string    `json:"tenant_name"`
	CostCenter     string    `json:"cost_center,omitempty"`
	Department     string    `json:"department,omitempty"`
	Project        string    `json:"project,omitempty"`
	TotalCostCents int64     `json:"total_cost_cents"`

	// Breakdown by resource type
	ExecutionCostCents int64 `json:"execution_cost_cents"`
	ComputeCostCents   int64 `json:"compute_cost_cents"`
	PlatformFeeCents   int64 `json:"platform_fee_cents"`
	DataTransferCents  int64 `json:"data_transfer_cents"`

	// Metadata for invoicing
	InvoicePeriod string    `json:"invoice_period"`
	GeneratedAt   time.Time `json:"generated_at"`
}

// TeamCostAllocation holds aggregated cost data for a specific team within a tenant.
type TeamCostAllocation struct {
	TeamID   uuid.UUID `json:"team_id"`
	TeamName string    `json:"team_name"`
	TenantID uuid.UUID `json:"tenant_id"`

	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`

	TotalCostCents     int64 `json:"total_cost_cents"`
	ExecutionCostCents int64 `json:"execution_cost_cents"`
	ComputeCostCents   int64 `json:"compute_cost_cents"`
	PlatformFeeCents   int64 `json:"platform_fee_cents"`
	DataTransferCents  int64 `json:"data_transfer_cents"`

	TotalExecutions int64 `json:"total_executions"`
	UniqueFunctions int   `json:"unique_functions"`

	// Budget context
	BudgetCents   int64   `json:"budget_cents,omitempty"`
	SpentPercent  float64 `json:"spent_percent,omitempty"`
	DaysRemaining int     `json:"days_remaining,omitempty"`
}

// TeamCostBreakdown is per-function cost within a team.
type TeamCostBreakdown struct {
	TeamID          uuid.UUID `json:"team_id"`
	FunctionID      uuid.UUID `json:"function_id"`
	FunctionName    string    `json:"function_name"`
	FunctionAuthor  string    `json:"function_author"`
	TotalCostCents  int64     `json:"total_cost_cents"`
	TotalExecutions int64     `json:"total_executions"`
	AvgCostCents    float64   `json:"avg_cost_cents"`
	SuccessRate     float64   `json:"success_rate"`
}

// DepartmentBudget defines a spending cap and alerting thresholds for a department.
type DepartmentBudget struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`

	// Budget amounts in cents
	BudgetCents          int64 `json:"budget_cents"`
	WarningThresholdPct  int   `json:"warning_threshold_pct"`  // e.g. 75 = warn at 75% spent
	CriticalThresholdPct int   `json:"critical_threshold_pct"` // e.g. 90 = critical at 90%

	// Time period
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`

	// Current spend (populated on queries)
	SpentCents     int64   `json:"spent_cents,omitempty"`
	RemainingCents int64   `json:"remaining_cents,omitempty"`
	SpentPercent   float64 `json:"spent_percent,omitempty"`

	// Which cost centers / teams are covered
	TeamIDs    []uuid.UUID       `json:"team_ids,omitempty"`
	TagFilters map[string]string `json:"tag_filters,omitempty"` // e.g. {"department": "engineering"}

	// Alerts
	AlertEmail string `json:"alert_email,omitempty"`

	IsActive  bool      `json:"is_active"`
	CreatedBy uuid.UUID `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BudgetAlert records a triggered budget warning/critical notification.
type BudgetAlert struct {
	ID       uuid.UUID `json:"id"`
	BudgetID uuid.UUID `json:"budget_id"`
	TenantID uuid.UUID `json:"tenant_id"`

	Level       string  `json:"level"` // "warning", "critical", "resolved"
	SpentPct    float64 `json:"spent_pct"`
	SpentCents  int64   `json:"spent_cents"`
	BudgetCents int64   `json:"budget_cents"`

	AlertSentTo  string `json:"alert_sent_to"`
	AlertMessage string `json:"alert_message,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

// CostAnomaly represents a detected abnormal spending pattern.
type CostAnomaly struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	AnomalyType string `json:"anomaly_type"` // "spike", "trend", "new_source", "budget_breach"
	Severity    string `json:"severity"`     // "low", "medium", "high", "critical"

	// Context
	TeamID     *uuid.UUID `json:"team_id,omitempty"`
	FunctionID *uuid.UUID `json:"function_id,omitempty"`
	Region     *string    `json:"region,omitempty"`

	// What changed
	ExpectedCostCents int64   `json:"expected_cost_cents"`
	ActualCostCents   int64   `json:"actual_cost_cents"`
	DeltaCents        int64   `json:"delta_cents"`
	DeltaPercent      float64 `json:"delta_percent"`

	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// ==================== API Key Billing Types ====================

// APIKeyBudget represents spending budget and alerting thresholds for a specific API key
type APIKeyBudget struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	APIKeyID uuid.UUID `json:"api_key_id" gorm:"type:uuid;not null;uniqueIndex"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`

	// Budget amounts in cents
	BudgetCents          int64 `json:"budget_cents"`           // Monthly budget limit
	WarningThresholdPct  int   `json:"warning_threshold_pct"`  // e.g. 75 = warn at 75% spent (default 80)
	CriticalThresholdPct int   `json:"critical_threshold_pct"` // e.g. 90 = critical at 90% (default 95)

	// Current spend (populated on queries)
	SpentCents     int64   `json:"spent_cents,omitempty"`
	RemainingCents int64   `json:"remaining_cents,omitempty"`
	SpentPercent   float64 `json:"spent_percent,omitempty"`

	// Period tracking
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`

	// Alerts
	AlertEmail     string     `json:"alert_email,omitempty"`
	LastAlertAt    *time.Time `json:"last_alert_at,omitempty"`
	LastAlertLevel string     `json:"last_alert_level,omitempty"` // "warning", "critical"

	// Auto-disable at limit
	DisableAtLimit bool `json:"disable_at_limit" gorm:"default:false"`

	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// APIKeyCostSummary provides aggregated cost data for a specific API key
type APIKeyCostSummary struct {
	APIKeyID   uuid.UUID `json:"api_key_id"`
	APIKeyName string    `json:"api_key_name"`

	// Period
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`

	// Executions
	TotalExecutions int64 `json:"total_executions"`

	// Costs (in cents)
	ExecutionCostCents int64 `json:"execution_cost_cents"`
	ComputeCostCents   int64 `json:"compute_cost_cents"`
	PlatformFeeCents   int64 `json:"platform_fee_cents"`
	DataTransferCents  int64 `json:"data_transfer_cents"`
	TotalCostCents     int64 `json:"total_cost_cents"`

	// Budget context
	BudgetCents   int64   `json:"budget_cents,omitempty"`
	SpentPercent  float64 `json:"spent_percent,omitempty"`
	DaysRemaining int     `json:"days_remaining,omitempty"`
	IsOverBudget  bool    `json:"is_over_budget"`
	IsNearBudget  bool    `json:"is_near_budget"` // Within warning threshold
}

// ==================== Tax/VAT Compliance Types ====================

// TaxRate represents a cached tax rate from Stripe Tax
// This is used for reporting and displaying tax information
type TaxRate struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Country         string     `json:"country" gorm:"not null;size:2"`       // ISO 3166-1 alpha-2
	State           *string    `json:"state,omitempty" gorm:"size:50"`       // State/Province
	PostalCode      *string    `json:"postal_code,omitempty" gorm:"size:20"` // Postal/ZIP code
	TaxType         string     `json:"tax_type" gorm:"not null;size:50"`     // 'vat', 'sales_tax', 'gst', etc.
	DisplayName     string     `json:"display_name" gorm:"not null;size:100"`
	Percentage      float64    `json:"percentage" gorm:"type:decimal(5,2);not null"`
	Inclusive       bool       `json:"inclusive" gorm:"default:false"` // Whether tax is included in price
	StripeTaxRateID *string    `json:"stripe_tax_rate_id,omitempty" gorm:"size:255"`
	Jurisdiction    *string    `json:"jurisdiction,omitempty" gorm:"size:100"`
	EffectiveFrom   *time.Time `json:"effective_from,omitempty"`
	EffectiveUntil  *time.Time `json:"effective_until,omitempty"`
	IsActive        bool       `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// InvoiceTaxDetail represents tax breakdown for an invoice
type InvoiceTaxDetail struct {
	ID                     uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	InvoiceID              uuid.UUID  `json:"invoice_id" gorm:"type:uuid;not null;index"`
	TaxRateID              *uuid.UUID `json:"tax_rate_id,omitempty" gorm:"type:uuid;index"`
	TaxAmountCents         int        `json:"tax_amount_cents" gorm:"not null;default:0"`
	SubtotalCents          int        `json:"subtotal_cents" gorm:"not null;default:0"`
	TotalCents             int        `json:"total_cents" gorm:"not null;default:0"`
	Currency               string     `json:"currency" gorm:"not null;default:'USD';size:3"`
	TaxName                *string    `json:"tax_name,omitempty" gorm:"size:100"`
	TaxPercentage          *float64   `json:"tax_percentage,omitempty" gorm:"type:decimal(5,2)"`
	StripeTaxCalculationID *string    `json:"stripe_tax_calculation_id,omitempty" gorm:"size:255"`
	CreatedAt              time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt              time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TaxIDValidationLog represents an audit log for tax ID validation attempts
type TaxIDValidationLog struct {
	ID                 uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID           uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	TaxID              string     `json:"tax_id" gorm:"not null;size:50"`
	TaxIDType          string     `json:"tax_id_type" gorm:"not null;size:20"`        // eu_vat, us_ein, ca_gst, etc.
	ValidationStatus   string     `json:"validation_status" gorm:"not null;size:20"`  // valid, invalid, pending
	ValidationSource   *string    `json:"validation_source,omitempty" gorm:"size:50"` // vies, stripe, manual
	ValidationResponse *string    `json:"validation_response,omitempty" gorm:"type:text"`
	ValidatedAt        *time.Time `json:"validated_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// TaxSettings represents tax configuration for a tenant
// This is used for managing tax settings through the API
type TaxSettings struct {
	TenantID            uuid.UUID `json:"tenant_id"`
	BillingCountry      *string   `json:"billing_country,omitempty"`
	BillingState        *string   `json:"billing_state,omitempty"`
	BillingPostalCode   *string   `json:"billing_postal_code,omitempty"`
	TaxID               *string   `json:"tax_id,omitempty"`
	TaxIDType           *string   `json:"tax_id_type,omitempty"`
	TaxStatus           string    `json:"tax_status"`
	TaxExempt           bool      `json:"tax_exempt"`
	StripeTaxLocationID *string   `json:"stripe_tax_location_id,omitempty"`
	StripeCustomerTaxID *string   `json:"stripe_customer_tax_id,omitempty"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// TaxCalculationRequest represents a request to calculate tax for a transaction
type TaxCalculationRequest struct {
	TenantID        uuid.UUID `json:"tenant_id"`
	AmountCents     int       `json:"amount_cents"`
	Currency        string    `json:"currency"`
	TransactionType string    `json:"transaction_type"` // 'subscription', 'one_time', 'usage'
}

// TaxCalculationResult represents the result of a tax calculation
type TaxCalculationResult struct {
	TaxAmountCents         int     `json:"tax_amount_cents"`
	SubtotalCents          int     `json:"subtotal_cents"`
	TotalCents             int     `json:"total_cents"`
	Currency               string  `json:"currency"`
	TaxRatePercentage      float64 `json:"tax_rate_percentage"`
	TaxName                string  `json:"tax_name"`
	Jurisdiction           string  `json:"jurisdiction"`
	StripeTaxCalculationID string  `json:"stripe_tax_calculation_id,omitempty"`
}

// TaxIDType represents valid tax ID types
type TaxIDType string

const (
	TaxIDTypeEUVAT   TaxIDType = "eu_vat"   // European Union VAT
	TaxIDTypeUKVAT   TaxIDType = "uk_vat"   // United Kingdom VAT
	TaxIDTypeUSEIN   TaxIDType = "us_ein"   // US Employer Identification Number
	TaxIDTypeUSSales TaxIDType = "us_sales" // US Sales Tax Permit
	TaxIDTypeCAGST   TaxIDType = "ca_gst"   // Canada GST/HST
	TaxIDTypeAUABN   TaxIDType = "au_abn"   // Australia ABN
	TaxIDTypeNZGST   TaxIDType = "nz_gst"   // New Zealand GST
	TaxIDTypeSGGST   TaxIDType = "sg_gst"   // Singapore GST
	TaxIDTypeCHVAT   TaxIDType = "ch_vat"   // Switzerland VAT
	TaxIDTypeNOVAT   TaxIDType = "no_vat"   // Norway VAT
	TaxIDTypeOther   TaxIDType = "other"    // Other tax ID type
)

// CreditNote represents an accounting credit note for refunds
// Credit notes are issued when refunding a customer and provide proper accounting
// reconciliation for SOX compliance
type CreditNote struct {
	ID              uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID        uuid.UUID   `json:"tenant_id" gorm:"type:uuid;not null;index"`
	InvoiceID       *uuid.UUID  `json:"invoice_id,omitempty" gorm:"type:uuid;index"`        // Original invoice being credited
	ReferenceNumber string      `json:"reference_number" gorm:"not null;size:50;uniqueIndex"` // Human-readable credit note number
	Status          string      `json:"status" gorm:"not null;size:20;default:'draft'"`       // draft, issued, applied, void

	// Amount fields (all in cents)
	SubtotalCents   int `json:"subtotal_cents" gorm:"not null;default:0"`
	TaxCents        int `json:"tax_cents" gorm:"not null;default:0"`
	TotalCents      int `json:"total_cents" gorm:"not null;default:0"`

	Currency string `json:"currency" gorm:"not null;size:3;default:'USD'"`

	// Reason and description
	Reason          string `json:"reason" gorm:"size:255"` // refund_reason from original transaction
	Description     string `json:"description,omitempty" gorm:"size:500"`

	// Refund association
	PaymentRefundID *uuid.UUID `json:"payment_refund_id,omitempty" gorm:"type:uuid;index"`

	// Timestamps
	IssuedAt  *time.Time `json:"issued_at,omitempty"`
	AppliedAt *time.Time `json:"applied_at,omitempty"`
	VoidedAt  *time.Time `json:"voided_at,omitempty"`

	// Audit fields
	IssuedBy uuid.UUID `json:"issued_by" gorm:"type:uuid;not null"` // Admin user who issued
	Notes    string    `json:"notes,omitempty" gorm:"size:1000"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Related entities (not stored, populated on query)
	Invoice        *Invoice        `json:"invoice,omitempty" gorm:"-"`
	LineItems      []*CreditNoteLineItem `json:"line_items,omitempty" gorm:"-"`
	PaymentRefund  *PaymentRefund  `json:"payment_refund,omitempty" gorm:"-"`
}

// CreditNoteLineItem represents a line item on a credit note
type CreditNoteLineItem struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreditNoteID   uuid.UUID  `json:"credit_note_id" gorm:"type:uuid;not null;index"`
	Description    string     `json:"description" gorm:"not null;size:500"`
	Quantity       int        `json:"quantity" gorm:"not null;default:1"`
	UnitPriceCents int        `json:"unit_price_cents" gorm:"not null;default:0"`
	TaxCents       int        `json:"tax_cents" gorm:"not null;default:0"`
	AmountCents    int        `json:"amount_cents" gorm:"not null;default:0"` // quantity * unit_price
	TotalCents     int        `json:"total_cents" gorm:"not null;default:0"` // amount + tax
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName overrides the table name for CreditNote
func (CreditNote) TableName() string {
	return "credit_notes"
}

// TableName overrides the table name for CreditNoteLineItem
func (CreditNoteLineItem) TableName() string {
	return "credit_note_line_items"
}

// CreditNoteStatus constants
const (
	CreditNoteStatusDraft  = "draft"
	CreditNoteStatusIssued = "issued"
	CreditNoteStatusApplied = "applied"
	CreditNoteStatusVoid   = "void"
)

// TaxStatus represents valid tax statuses
const (
	TaxStatusPending  = "pending"  // Tax status not yet determined
	TaxStatusValid    = "valid"    // Tax ID validated successfully
	TaxStatusInvalid  = "invalid"  // Tax ID validation failed
	TaxStatusExempt   = "exempt"   // Customer is tax exempt
	TaxStatusRequired = "required" // Tax ID is required but not provided
)

// ValidationSource represents sources of tax ID validation
const (
	ValidationSourceVIES   = "vies"   // EU VIES (VAT Information Exchange System)
	ValidationSourceStripe = "stripe" // Stripe Tax validation
	ValidationSourceManual = "manual" // Manual admin verification
)

// ==================== Webhook Replay & Audit Types ====================

// StoredWebhookPayload represents a stored raw webhook payload for replay capability
// Payloads are retained for 30 days for operational recovery
type StoredWebhookPayload struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StripeEventID     string          `json:"stripe_event_id" gorm:"not null;size:255;index:idx_webhook_event_id"`
	EventType         string          `json:"event_type" gorm:"not null;size:100;index:idx_webhook_event_type"`
	Payload           json.RawMessage `json:"payload" gorm:"type:jsonb;not null"`
	Signature         string          `json:"signature" gorm:"size:255"`                                                            // Stored for verification replay
	ProcessingStatus  string          `json:"processing_status" gorm:"not null;size:20;default:'pending';index:idx_webhook_status"` // pending, processed, failed, replayed
	ProcessedAt       *time.Time      `json:"processed_at,omitempty"`
	ReplayedAt        *time.Time      `json:"replayed_at,omitempty"`
	ReplayedBy        *uuid.UUID      `json:"replayed_by,omitempty" gorm:"type:uuid"`
	ReplayReason      string          `json:"replay_reason,omitempty" gorm:"size:500"`
	ProcessingError   string          `json:"processing_error,omitempty" gorm:"size:1000"`
	Attempts          int             `json:"attempts" gorm:"default:0"`
	WebhookSecretHash string          `json:"webhook_secret_hash,omitempty" gorm:"size:64"` // Hash of secret used (for audit)
	CreatedAt         time.Time       `json:"created_at" gorm:"autoCreateTime"`
	ExpiresAt         time.Time       `json:"expires_at" gorm:"not null;index:idx_webhook_expires"` // 30-day retention
}

// TableName overrides the table name
func (StoredWebhookPayload) TableName() string {
	return "stored_webhook_payloads"
}

// WebhookReplayRequest represents a manual webhook replay request
type WebhookReplayRequest struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WebhookPayloadID uuid.UUID  `json:"webhook_payload_id" gorm:"type:uuid;not null;index"`
	RequestedBy      uuid.UUID  `json:"requested_by" gorm:"type:uuid;not null"`
	RequestedAt      time.Time  `json:"requested_at" gorm:"autoCreateTime"`
	Reason           string     `json:"reason" gorm:"not null;size:500"`
	Status           string     `json:"status" gorm:"not null;size:20;default:'pending'"` // pending, completed, failed
	ResultMessage    string     `json:"result_message,omitempty" gorm:"size:1000"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

// ==================== Tax Exemption Certificate Types ====================

// TaxExemptionCertificate represents a tax exemption certificate uploaded by a US entity
type TaxExemptionCertificate struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index:idx_tax_cert_tenant"`
	UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index:idx_tax_cert_user"`

	// Certificate details
	CertificateNumber string `json:"certificate_number" gorm:"not null;size:100;index:idx_tax_cert_number"`
	State             string `json:"state" gorm:"not null;size:2;index:idx_tax_cert_state"` // US state code
	ExemptionType     string `json:"exemption_type" gorm:"not null;size:50"`                // resale, nonprofit, government, agricultural, etc.
	ExemptionReason   string `json:"exemption_reason" gorm:"size:500"`

	// File storage
	FileURL  string `json:"file_url" gorm:"not null;size:500"`
	FileName string `json:"file_name" gorm:"not null;size:255"`
	FileSize int64  `json:"file_size" gorm:"not null"`
	FileHash string `json:"file_hash" gorm:"not null;size:64"` // SHA-256 hash for integrity

	// Validity period
	ValidFrom  time.Time  `json:"valid_from" gorm:"not null"`
	ValidUntil *time.Time `json:"valid_until,omitempty"`

	// Review workflow
	Status          string     `json:"status" gorm:"not null;size:20;default:'pending';index:idx_tax_cert_status"` // pending, approved, rejected, expired
	ReviewedBy      *uuid.UUID `json:"reviewed_by,omitempty" gorm:"type:uuid"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	ReviewNotes     string     `json:"review_notes,omitempty" gorm:"size:1000"`
	RejectionReason string     `json:"rejection_reason,omitempty" gorm:"size:500"`

	// Stripe integration
	StripeExemptionID *string    `json:"stripe_exemption_id,omitempty" gorm:"size:255"`
	AppliedToStripeAt *time.Time `json:"applied_to_stripe_at,omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName overrides the table name
func (TaxExemptionCertificate) TableName() string {
	return "tax_exemption_certificates"
}

// EUVATValidation represents VAT ID validation results from VIES API
type EUVATValidation struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index:idx_vat_tenant"`
	UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index:idx_vat_user"`

	// VAT ID details
	VATID       string `json:"vat_id" gorm:"not null;size:20;index:idx_vat_id"`
	CountryCode string `json:"country_code" gorm:"not null;size:2"` // EU country code

	// VIES API response
	IsValid          bool      `json:"is_valid" gorm:"not null"`
	RequestDate      time.Time `json:"request_date" gorm:"not null"`
	ValidationSource string    `json:"validation_source" gorm:"not null;size:20;default:'vies'"` // vies, fallback

	// VIES response details (stored for audit)
	VIESRequestID     string `json:"vies_request_id,omitempty" gorm:"size:100"`
	VIESResponseCode  string `json:"vies_response_code,omitempty" gorm:"size:50"`
	VIESTraderName    string `json:"vies_trader_name,omitempty" gorm:"size:255"`
	VIESTraderAddress string `json:"vies_trader_address,omitempty" gorm:"size:500"`

	// Error handling
	ErrorCode    string `json:"error_code,omitempty" gorm:"size:50"`
	ErrorMessage string `json:"error_message,omitempty" gorm:"size:500"`

	// Retry logic for VIES unavailability
	RetryCount  int        `json:"retry_count" gorm:"default:0"`
	NextRetryAt *time.Time `json:"next_retry_at,omitempty"`

	// Status
	Status string `json:"status" gorm:"not null;size:20;default:'pending';index:idx_vat_status"` // pending, valid, invalid, error, timeout

	// Applied to tenant tax settings
	AppliedToSettings bool       `json:"applied_to_settings" gorm:"default:false"`
	AppliedAt         *time.Time `json:"applied_at,omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName overrides the table name
func (EUVATValidation) TableName() string {
	return "eu_vat_validations"
}

// ==================== Stripe Two-Way Sync Types ====================

// StripeSyncEvent represents an event received from Stripe webhooks
// This enables two-way sync between Stripe and internal records
type StripeSyncEvent struct {
	ID             uuid.UUID       `json:"id"`
	StripeEventID  string          `json:"stripe_event_id"`
	StripeObjectID string          `json:"stripe_object_id"` // ID of the Stripe object (subscription, payment_method, etc.)
	EventType      string          `json:"event_type"`
	EventData      json.RawMessage `json:"event_data"`
	TenantID       *uuid.UUID      `json:"tenant_id,omitempty"`
	Status         string          `json:"status"` // 'pending', 'processed', 'failed', 'ignored'
	ErrorMessage   *string         `json:"error_message,omitempty"`
	ProcessedAt    *time.Time      `json:"processed_at,omitempty"`
	RetryCount     int             `json:"retry_count"`
	IdempotencyKey *string         `json:"idempotency_key,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// PaymentMethodInfoExtended represents stored payment method information
type PaymentMethodInfoExtended struct {
	ID                    uuid.UUID       `json:"id"`
	TenantID              uuid.UUID       `json:"tenant_id"`
	StripePaymentMethodID string          `json:"stripe_payment_method_id"`
	Brand                 string          `json:"brand"`
	Last4                 string          `json:"last4"`
	ExpMonth              int             `json:"exp_month"`
	ExpYear               int             `json:"exp_year"`
	IsDefault             bool            `json:"is_default"`
	BillingDetails        json.RawMessage `json:"billing_details,omitempty"` // Name, email, address from Stripe
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

// SubscriptionSyncStatus represents the sync status constants
const (
	StripeSyncStatusPending   = "pending"
	StripeSyncStatusProcessed = "processed"
	StripeSyncStatusFailed    = "failed"
	StripeSyncStatusIgnored   = "ignored"
)

// =============================================================================
// Database-Driven Agent Tier Pricing (replaces hardcoded constants)
// =============================================================================

// AgentTierPricing represents a database-driven agent subscription tier
// This replaces the hardcoded constants in internal/plans/limits.go
type AgentTierPricing struct {
	ID                       uuid.UUID              `json:"id"`
	TierSlug                 string                 `json:"tier_slug"` // 'agent-starter', 'agent-scale', 'agent-pro', 'agent-enterprise'
	DisplayName              string                 `json:"display_name"`
	Description              string                 `json:"description"`
	MonthlyPriceCents        int                    `json:"monthly_price_cents"`          // Base price in base_currency
	AnnualPriceCents         *int                   `json:"annual_price_cents,omitempty"` // NULL = no annual discount
	BaseCurrency             string                 `json:"base_currency"`                // ISO 4217 code (USD, EUR, etc.)
	RegionPricing            map[string]interface{} `json:"region_pricing,omitempty"`     // Region-specific pricing overrides
	MaxAgents                int                    `json:"max_agents"`                   // -1 = unlimited
	IncludedAICalls          int                    `json:"included_ai_calls"`            // -1 = unlimited
	IncludedExecutions       int                    `json:"included_executions"`          // -1 = unlimited
	IncludedStorageGB        int                    `json:"included_storage_gb"`          // -1 = unlimited
	OveragePricePer1000Cents int                    `json:"overage_price_per_1000_cents"` // Price per 1000 overage units
	StripePriceIDMonthly     *string                `json:"stripe_price_id_monthly,omitempty"`
	StripePriceIDAnnual      *string                `json:"stripe_price_id_annual,omitempty"`
	FeaturesIncluded         []string               `json:"features_included"` // Feature flags enabled for this tier
	IsActive                 bool                   `json:"is_active"`
	SortOrder                int                    `json:"sort_order"`
	PricingVariant           string                 `json:"pricing_variant"` // 'default', 'experiment_a', 'promo'
	ValidFrom                time.Time              `json:"valid_from"`
	ValidUntil               *time.Time             `json:"valid_until,omitempty"`
	CreatedAt                time.Time              `json:"created_at"`
	UpdatedAt                time.Time              `json:"updated_at"`
}

// GetMonthlyPrice returns the monthly price in the specified currency
// If a region-specific price exists for the currency, it will be used
func (t *AgentTierPricing) GetMonthlyPrice(currencyCode string) int {
	if currencyCode == t.BaseCurrency {
		return t.MonthlyPriceCents
	}
	// Check for region-specific pricing
	if t.RegionPricing != nil {
		if regionPrice, ok := t.RegionPricing[currencyCode]; ok {
			if priceMap, ok := regionPrice.(map[string]interface{}); ok {
				if monthly, ok := priceMap["monthly"].(float64); ok {
					return int(monthly)
				}
			}
		}
	}
	// Return base currency price (caller should convert using currency service)
	return t.MonthlyPriceCents
}

// GetAnnualPrice returns the annual price in the specified currency
func (t *AgentTierPricing) GetAnnualPrice(currencyCode string) *int {
	if t.AnnualPriceCents == nil {
		return nil
	}
	if currencyCode == t.BaseCurrency {
		return t.AnnualPriceCents
	}
	// Check for region-specific pricing
	if t.RegionPricing != nil {
		if regionPrice, ok := t.RegionPricing[currencyCode]; ok {
			if priceMap, ok := regionPrice.(map[string]interface{}); ok {
				if annual, ok := priceMap["annual"].(float64); ok {
					annualInt := int(annual)
					return &annualInt
				}
			}
		}
	}
	return t.AnnualPriceCents
}

// IsUnlimited returns true if the tier has unlimited resources
func (t *AgentTierPricing) IsUnlimited(field string) bool {
	switch field {
	case "agents":
		return t.MaxAgents < 0
	case "ai_calls":
		return t.IncludedAICalls < 0
	case "executions":
		return t.IncludedExecutions < 0
	case "storage":
		return t.IncludedStorageGB < 0
	default:
		return false
	}
}

// IsValidAt returns true if the pricing is valid at the given time
func (t *AgentTierPricing) IsValidAt(checkTime time.Time) bool {
	if checkTime.Before(t.ValidFrom) {
		return false
	}
	if t.ValidUntil != nil && checkTime.After(*t.ValidUntil) {
		return false
	}
	return t.IsActive
}

// =============================================================================
// Multi-Currency Support
// =============================================================================

// CurrencyExchangeRate represents an exchange rate for currency conversion
// Uses integer math for financial precision to avoid floating-point errors
type CurrencyExchangeRate struct {
	ID               uuid.UUID  `json:"id"`
	BaseCurrency     string     `json:"base_currency"`    // e.g., 'USD'
	QuoteCurrency    string     `json:"quote_currency"`   // e.g., 'EUR'
	Rate             float64    `json:"rate"`             // How much quote currency for 1 base currency (display/legacy)
	RateNumerator    int64      `json:"rate_numerator"`   // Integer rate numerator for precision math
	RateDenominator  int64      `json:"rate_denominator"` // Integer rate denominator (default: 1_000_000)
	Source           string     `json:"source"`           // 'ecb', 'openexchange', 'manual', 'stripe'
	SourceURL        *string    `json:"source_url,omitempty"`
	EffectiveDate    string     `json:"effective_date"` // YYYY-MM-DD
	FetchedAt        *time.Time `json:"fetched_at,omitempty"`
	IsManualOverride bool       `json:"is_manual_override"`
	OverrideReason   *string    `json:"override_reason,omitempty"`
	IsStripeRate     bool       `json:"is_stripe_rate"`   // Stripe adds ~1-2% markup
	StripePrecision  string     `json:"stripe_precision"` // '0', '2', '4' decimal places
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

const (
	// DefaultRateDenominator is the default denominator for rate calculations (1 million = 6 decimal places)
	DefaultRateDenominator int64 = 1_000_000
)

// GetRateDenominator returns the rate denominator for calculations
// Uses DefaultRateDenominator if not set
func (r *CurrencyExchangeRate) GetRateDenominator() int64 {
	if r.RateDenominator <= 0 {
		return DefaultRateDenominator
	}
	return r.RateDenominator
}

// GetRateNumerator returns the rate numerator for calculations
// Derives from float Rate if not explicitly set (for backward compatibility)
func (r *CurrencyExchangeRate) GetRateNumerator() int64 {
	if r.RateNumerator > 0 {
		return r.RateNumerator
	}
	// Derive from float rate for backward compatibility
	denom := r.GetRateDenominator()
	return int64(r.Rate * float64(denom))
}

// Convert converts an amount from base currency to quote currency using integer math
// Formula: (amountCents * rateNumerator) / rateDenominator
// This avoids floating-point precision issues in financial calculations
func (r *CurrencyExchangeRate) Convert(amountCents int) int {
	numerator := r.GetRateNumerator()
	denominator := r.GetRateDenominator()

	// Use 128-bit intermediate to prevent overflow for large amounts
	// amountCents is in cents, so max around $100M = 10^10 cents
	// numerator is around 10^6, so product is around 10^16 which fits in int64
	converted := (int64(amountCents) * numerator) / denominator

	// Round to nearest cent (already integer, but ensure positive rounding)
	return int(converted)
}

// ConvertTo converts an amount from the base currency to a target currency
// This is the inverse of Convert (amount is in base currency, returns quote currency)
func (r *CurrencyExchangeRate) ConvertTo(amountCents int) int {
	return r.Convert(amountCents)
}

// SetRateFromFloat sets the integer rate fields from a float rate value
// This should be called when creating/updating rates to ensure consistency
func (r *CurrencyExchangeRate) SetRateFromFloat(rate float64) {
	r.Rate = rate
	r.RateDenominator = DefaultRateDenominator
	r.RateNumerator = int64(rate * float64(DefaultRateDenominator))
}

// SupportedCurrency represents a currency that can be used in the system
type SupportedCurrency struct {
	Code               string    `json:"code"` // ISO 4217 code (USD, EUR, etc.)
	Name               string    `json:"name"`
	Symbol             string    `json:"symbol"`
	SymbolPosition     string    `json:"symbol_position"` // 'before' ($100) or 'after' (100 €)
	DecimalPlaces      int       `json:"decimal_places"`
	ThousandsSeparator string    `json:"thousands_separator"`
	DecimalSeparator   string    `json:"decimal_separator"`
	IsActive           bool      `json:"is_active"`
	IsStablecoin       bool      `json:"is_stablecoin"`              // USDC, USDT, etc.
	ContractAddress    *string   `json:"contract_address,omitempty"` // For stablecoins
	ChainID            *int      `json:"chain_id,omitempty"`
	DefaultCountry     *string   `json:"default_country,omitempty"`
	SupportedCountries []string  `json:"supported_countries,omitempty"`
	RoundingMode       string    `json:"rounding_mode"`        // 'half_up', 'half_down', 'up', 'down'
	MinimumChargeCents int       `json:"minimum_charge_cents"` // Stripe minimum (usually 50 cents)
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// FormatAmount formats an amount in cents for display
func (c *SupportedCurrency) FormatAmount(amountCents int) string {
	amount := float64(amountCents) / math.Pow(10, float64(c.DecimalPlaces))

	// Format with thousands separator
	parts := strings.Split(fmt.Sprintf("%.*f", c.DecimalPlaces, amount), ".")
	intPart := parts[0]
	decPart := ""
	if len(parts) > 1 {
		decPart = c.DecimalSeparator + parts[1]
	}

	// Add thousands separators
	var result strings.Builder
	for i, ch := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteString(c.ThousandsSeparator)
		}
		result.WriteRune(ch)
	}
	result.WriteString(decPart)

	if c.SymbolPosition == "before" {
		return c.Symbol + result.String()
	}
	return result.String() + " " + c.Symbol
}

// ConvertToStripeAmount converts cents to Stripe's smallest unit
// Some currencies like JPY have no decimal places
func (c *SupportedCurrency) ConvertToStripeAmount(cents int) int64 {
	if c.DecimalPlaces == 0 {
		// JPY, HUF, etc. - Stripe uses the main unit directly
		return int64(cents / 100)
	}
	// Standard currencies - Stripe uses cents
	return int64(cents)
}

// ConvertFromStripeAmount converts Stripe's smallest unit to cents
func (c *SupportedCurrency) ConvertFromStripeAmount(stripeAmount int64) int {
	if c.DecimalPlaces == 0 {
		return int(stripeAmount * 100)
	}
	return int(stripeAmount)
}

// =============================================================================
// Affiliate / Referral Commission System
// =============================================================================

const (
	AffiliateStatusActive   = "active"
	AffiliateStatusPaused  = "paused"
	AffiliateStatusCanceled = "canceled"
)

const (
	CommissionStatusPending   = "pending"
	CommissionStatusApproved  = "approved"
	CommissionStatusPaid      = "paid"
	CommissionStatusCanceled  = "canceled"
)

// AffiliateCode represents an affiliate/referral code for promoter commissions
type AffiliateCode struct {
	ID          uuid.UUID  `json:"id"`
	Code        string    `json:"code"` // Unique referral code (e.g., "SAVE20")
	PublisherID uuid.UUID  `json:"publisher_id"` // User who owns this code
	TenantID    *uuid.UUID `json:"tenant_id,omitempty"`

	Name        string    `json:"name"` // Display name for the campaign
	Description string    `json:"description,omitempty"`

	// Commission structure
	CommissionType  string  `json:"commission_type"`  // 'percent' or 'fixed'
	CommissionValue float64 `json:"commission_value"` // e.g., 20.00 for 20% or $20 fixed

	// Limits
	MaxCommissions *int `json:"max_commissions,omitempty"` // Max total commissions payout
	MaxReferrals    *int `json:"max_referrals,omitempty"`    // Max number of referrals

	// Current counts
	TotalReferrals     int `json:"total_referrals"`
	TotalCommissions  int `json:"total_commissions"` // Count of paid commissions
	PendingCommissions int `json:"pending_commissions"`

	// Earnings tracking (in cents for precision)
	PendingEarningsCents  int64 `json:"pending_earnings_cents"`
	TotalEarningsCents   int64 `json:"total_earnings_cents"`
	PaidOutEarningsCents  int64 `json:"paid_out_earnings_cents"`

	// Validity
	ValidFrom  *time.Time `json:"valid_from,omitempty"`
	ValidUntil *time.Time `json:"valid_until,omitempty"`
	IsActive   bool      `json:"is_active"`

	// Metadata
	UTMSource   string `json:"utm_source,omitempty"`
	UTMCampaign string `json:"utm_campaign,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AffiliateReferral represents a referral tracking record
type AffiliateReferral struct {
	ID              uuid.UUID  `json:"id"`
	AffiliateCodeID uuid.UUID  `json:"affiliate_code_id"`
	AffiliateCode   *AffiliateCode `json:"affiliate_code,omitempty"`

	// Referred tenant info
	ReferredTenantID uuid.UUID `json:"referred_tenant_id"`
	ReferralTenant   *Tenant   `json:"referral_tenant,omitempty"`

	// Subscription info (if subscribed)
	SubscriptionID *uuid.UUID `json:"subscription_id,omitempty"`

	// Attribution
	UTMSource   string `json:"utm_source,omitempty"`
	UTMCampaign string `json:"utm_campaign,omitempty"`
	UTContent   string `json:"utm_content,omitempty"`
	UTMTerm     string `json:"utm_term,omitempty"`

	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`

	// Tracking status
	Status string `json:"status"` // 'pending', 'converted', 'qualified', 'canceled'

	// When the referral was made
	ReferredAt time.Time `json:"referred_at"`

	// When they subscribed (if they did)
	ConvertedAt *time.Time `json:"converted_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AffiliateCommission represents a commission earned by an affiliate
type AffiliateCommission struct {
	ID              uuid.UUID  `json:"id"`
	AffiliateCodeID uuid.UUID  `json:"affiliate_code_id"`
	AffiliateCode   *AffiliateCode `json:"affiliate_code,omitempty"`

	ReferralID       uuid.UUID        `json:"referral_id"`
	Referral        *AffiliateReferral `json:"referral,omitempty"`

	// Commission details
	CommissionType  string  `json:"commission_type"`  // 'percent' or 'fixed'
	CommissionValue float64 `json:"commission_value"` // e.g., 20.00

	// What the commission is based on
	BaseAmountCents int64   `json:"base_amount_cents"` // Subscription amount in cents
	BaseAmountUSD    float64 `json:"base_amount_usd"`

	// Calculated commission
	CommissionCents int64   `json:"commission_cents"`
	CommissionUSD   float64 `json:"commission_usd"`

	// Status
	Status string `json:"status"` // 'pending', 'approved', 'paid', 'canceled'

	// For tracking payment
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	PaymentBatchID  *uuid.UUID `json:"payment_batch_id,omitempty"`
	PaymentBatch    string     `json:"payment_batch,omitempty"` // Human-readable batch ID

	// Associated subscription for commission period
	SubscriptionID *uuid.UUID `json:"subscription_id,omitempty"`

	Notes string `json:"notes,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AffiliateCodeStatus constants
const (
	ReferralStatusPending   = "pending"
	ReferralStatusConverted = "converted"
	ReferralStatusQualified = "qualified"
	ReferralStatusCanceled  = "canceled"
)
