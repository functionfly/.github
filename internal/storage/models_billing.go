package storage

import (
	"time"

	"github.com/google/uuid"
)

// PricingTier represents a billing pricing tier
type PricingTier struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	PriceCents  int         `json:"price_cents"`
	Currency    string      `json:"currency"`
	Features    interface{} `json:"features"` // JSON features/limits
	IsActive    bool        `json:"is_active"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Subscription represents a tenant's subscription
type Subscription struct {
	ID                   uuid.UUID    `json:"id"`
	TenantID             uuid.UUID    `json:"tenant_id"`
	PricingTierID        uuid.UUID    `json:"pricing_tier_id"`
	Status               string       `json:"status"`
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
