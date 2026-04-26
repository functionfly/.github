package storage

import (
	"time"

	"github.com/google/uuid"
)

const (
	RevenueTypeSubscription = "subscription"
	RevenueTypeUsage        = "usage"
	RevenueTypeOneTime      = "one_time"
	RevenueTypeRecognition  = "recognition"
)

const (
	RecognitionMethodOverTime  = "over_time"
	RecognitionMethodPointInTime = "point_in_time"
)

const (
	PerformanceObligationTypeAccess    = "access"     // Access to software
	PerformanceObligationTypeUsage      = "usage"      // Usage-based
	PerformanceObligationTypeLicense    = "license"    // License key
	PerformanceObligationTypeSupport    = "support"    // Support services
	PerformanceObligationTypeCustom     = "custom"     // Custom development
)

const (
	ContractAssetType = "contract_asset"
	ContractLiabilityType = "contract_liability"
)

type PerformanceObligation struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	InvoiceID uuid.UUID `json:"invoice_id" gorm:"type:uuid;not null;index"`

	Name string `json:"name" gorm:"not null;size:255"`
	Description string `json:"description" gorm:"size:1000"`

	// Classification
	Type string `json:"type" gorm:"size:50;not null"` // access, usage, license, support, custom

	// Transaction price allocation
	TransactionPriceCents int `json:"transaction_price_cents" gorm:"not null;default:0"`
	AllocatedPriceCents int `json:"allocated_price_cents" gorm:"not null;default:0"`

	// SSP (Standalone Selling Price)
	SSPCents int `json:"ssp_cents" gorm:"not null;default:0"`
	SSPCurrency string `json:"ssp_currency" gorm:"size:3;default:'USD'"`
	SSPBasis string `json:"ssp_basis" gorm:"size:50"` // 'total', 'per_unit', 'tiered'

	// Recognition pattern
	RecognitionMethod string `json:"recognition_method" gorm:"size:50;not null"` // over_time, point_in_time
	RecognitionStartDate time.Time `json:"recognition_start_date"`
	RecognitionEndDate *time.Time `json:"recognition_end_date,omitempty"`

	// For over-time: delivery pattern
	DeliveryPattern string `json:"delivery_pattern" gorm:"size:50"` // 'linear', 'milestone', 'usage_based'
	Milestones []byte `json:"milestones" gorm:"type:jsonb"` // [{name, date, percentage}]

	// For usage-based: billing period
	BillablePeriodStart time.Time `json:"billable_period_start"`
	BillablePeriodEnd time.Time `json:"billable_period_end"`

	// Delivery status
	IsDelivered bool `json:"is_delivered" gorm:"default:false"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`

	// Recognition status
	IsFullyRecognized bool `json:"is_fully_recognized" gorm:"default:false"`
	FullyRecognizedAt *time.Time `json:"fully_recognized_at,omitempty"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" gorm:"-"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ContractAsset struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`

	// Reference
	InvoiceID *uuid.UUID `json:"invoice_id,omitempty" gorm:"type:uuid;index"`
	CustomerID string `json:"customer_id" gorm:"size:255"` // Stripe customer ID

	// Type
	AssetType string `json:"asset_type" gorm:"size:50;not null"` // contract_asset, contract_liability

	// Amounts (in cents)
	AmountCents int `json:"amount_cents" gorm:"not null;default:0"`
	Currency string `json:"currency" gorm:"size:3;default:'USD'"`

	// Description
	Description string `json:"description" gorm:"size:500"`

	// Period
	ReportingPeriod string `json:"reporting_period" gorm:"size:7"` // YYYY-MM

	// Status
	Status string `json:"status" gorm:"size:50;not null"` // 'active', 'reduced', 'settled'
	ReducedAmountCents int `json:"reduced_amount_cents" gorm:"default:0"`

	// Reversal tracking
	IsReversed bool `json:"is_reversed" gorm:"default:false"`
	ReversedAt *time.Time `json:"reversed_at,omitempty"`

	// Reduction reason
	ReductionReason string `json:"reduction_reason,omitempty" gorm:"size:255"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type RevenueRecognitionSchedule struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	InvoiceID uuid.UUID `json:"invoice_id" gorm:"type:uuid;not null;index"`
	PerformanceObligationID uuid.UUID `json:"performance_obligation_id" gorm:"type:uuid;index"`

	// Schedule details
	RecognitionMonth string `json:"recognition_month" gorm:"size:7;not null;index"` // YYYY-MM
	PeriodStartDate time.Time `json:"period_start_date"`
	PeriodEndDate time.Time `json:"period_end_date"`

	// Amounts
	AllocatedAmountCents int `json:"allocated_amount_cents" gorm:"not null;default:0"`
	RecognizedAmountCents int `json:"recognized_amount_cents" gorm:"not null;default:0"`
	DeferredAmountCents int `json:"deferred_amount_cents" gorm:"not null;default:0"`

	// Recognition type
	RevenueType string `json:"revenue_type" gorm:"size:50;not null"` // subscription, usage, one_time

	// Status
	IsRecognized bool `json:"is_recognized" gorm:"default:false"`
	RecognizedAt *time.Time `json:"recognized_at,omitempty"`

	// Overwrite of the original revenue_recognition table entry
	OriginalTotalCents int `json:"original_total_cents" gorm:"default:0"`

	// Adjustment tracking
	IsAdjustment bool `json:"is_adjustment" gorm:"default:false"`
	AdjustmentReason string `json:"adjustment_reason,omitempty" gorm:"size:255"`
	PreviousScheduleID *uuid.UUID `json:"previous_schedule_id,omitempty" gorm:"type:uuid"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type RevenueRecognitionEvent struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	InvoiceID uuid.UUID `json:"invoice_id" gorm:"type:uuid;not null;index"`

	// Event classification
	EventType string `json:"event_type" gorm:"size:50;not null"` // 'invoice_paid', 'delivery_completed', 'milestone_reached', 'contract_modified', 'credit_note_issued'
	RevenueType string `json:"revenue_type" gorm:"size:50;not null"` // subscription, usage, one_time

	// Financial impact (in cents)
	GrossAmountCents int `json:"gross_amount_cents" gorm:"not null;default:0"`
	DeferredAmountCents int `json:"deferred_amount_cents" gorm:"not null;default:0"`
	RecognizedAmountCents int `json:"recognized_amount_cents" gorm:"not null;default:0"`

	// Period
	EventDate time.Time `json:"event_date" gorm:"not null"`
	ReportingPeriod string `json:"reporting_period" gorm:"size:7"` // YYYY-MM

	// Reference
	PerformanceObligationID *uuid.UUID `json:"performance_obligation_id,omitempty" gorm:"type:uuid;index"`
	ScheduleID *uuid.UUID `json:"schedule_id,omitempty" gorm:"type:uuid;index"`

	// Contract modification
	PreviousInvoiceID *uuid.UUID `json:"previous_invoice_id,omitempty" gorm:"type:uuid"`
	ModificationType string `json:"modification_type,omitempty" gorm:"size:50"` // 'add_scope', 'terminate', 'price_change'

	// Metadata
	Description string `json:"description" gorm:"size:500"`
	Metadata map[string]interface{} `json:"metadata,omitempty" gorm:"-"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type DeferredRevenueSummary struct {
	TenantID uuid.UUID `json:"tenant_id"`
	ReportingPeriod string `json:"reporting_period"` // YYYY-MM

	// Opening balance
	OpeningBalanceCents int `json:"opening_balance_cents"`

	// Movements
	NewDeferredCents int `json:"new_deferred_cents"`
	RecognizedCents int `json:"recognized_cents"`
	AdjustmentsCents int `json:"adjustments_cents"`
	WriteOffsCents int `json:"write_offs_cents"`

	// Closing balance
	ClosingBalanceCents int `json:"closing_balance_cents"`

	// Breakdown by type
	SubscriptionDeferredCents int `json:"subscription_deferred_cents"`
	UsageDeferredCents int `json:"usage_deferred_cents"`
	OneTimeDeferredCents int `json:"one_time_deferred_cents"`
}

type RecognizedRevenueSummary struct {
	TenantID uuid.UUID `json:"tenant_id"`
	ReportingPeriod string `json:"reporting_period"` // YYYY-MM

	// By revenue type
	SubscriptionRevenueCents int `json:"subscription_revenue_cents"`
	UsageRevenueCents int `json:"usage_revenue_cents"`
	OneTimeRevenueCents int `json:"one_time_revenue_cents"`
	TotalRevenueCents int `json:"total_revenue_cents"`

	// Over-time vs point-in-time
	OverTimeRevenueCents int `json:"over_time_revenue_cents"`
	PointInTimeRevenueCents int `json:"point_in_time_revenue_cents"`

	// Contract assets
	ContractAssetIncreaseCents int `json:"contract_asset_increase_cents"`
	ContractAssetDecreaseCents int `json:"contract_asset_decrease_cents"`
}

type RevenueRecognitionReport struct {
	ReportID uuid.UUID `json:"report_id"`
	GeneratedAt time.Time `json:"generated_at"`

	// Period
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd time.Time `json:"period_end"`
	ReportingPeriod string `json:"reporting_period"` // YYYY-MM

	// Summary
	TotalRevenueCents int `json:"total_revenue_cents"`
	TotalDeferredCents int `json:"total_deferred_cents"`
	TotalRecognizedCents int `json:"total_recognized_cents"`

	// Deferred revenue
	OpeningDeferredCents int `json:"opening_deferred_cents"`
	NewDeferredCents int `json:"new_deferred_cents"`
	RecognizedFromDeferredCents int `json:"recognized_from_deferred_cents"`
	ClosingDeferredCents int `json:"closing_deferred_cents"`

	// Contract assets
	OpeningContractAssetCents int `json:"opening_contract_asset_cents"`
	NetContractAssetMovementCents int `json:"net_contract_asset_movement_cents"`
	ClosingContractAssetCents int `json:"closing_contract_asset_cents"`

	// Breakdown by revenue type
	ByRevenueType map[string]int `json:"by_revenue_type,omitempty"`

	// Breakdown by performance obligation
	ByPerformanceObligation map[string]int `json:"by_performance_obligation,omitempty"`

	// Recognition method split
	OverTimeRevenueCents int `json:"over_time_revenue_cents"`
	PointInTimeRevenueCents int `json:"point_in_time_revenue_cents"`

	// Adjustments
	NumberOfAdjustments int `json:"number_of_adjustments"`
	AdjustmentAmountCents int `json:"adjustment_amount_cents"`
}