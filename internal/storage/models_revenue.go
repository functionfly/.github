package storage

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// VerificationFee represents the fee structure for function verification
type VerificationFee struct {
	ID          uuid.UUID  `json:"id"`
	Level       string    `json:"level"` // 'basic', 'standard', 'full'
	PriceCents  int       `json:"price_cents"`
	Currency    string    `json:"currency"`
	IsActive    bool      `json:"is_active"`
	MinPlan     *string   `json:"min_plan,omitempty"` // NULL = available to all plans
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FunctionVerificationPayment represents a payment for function verification
type FunctionVerificationPayment struct {
	ID                      uuid.UUID  `json:"id"`
	FunctionID              uuid.UUID  `json:"function_id"`
	VerificationLevel       string     `json:"verification_level"`
	AmountCents             int        `json:"amount_cents"`
	Currency                string     `json:"currency"`
	Status                  string     `json:"status"` // 'pending', 'paid', 'refunded', 'failed'
	StripePaymentIntentID   *string    `json:"stripe_payment_intent_id,omitempty"`
	StripeCheckoutSessionID *string    `json:"stripe_checkout_session_id,omitempty"`
	TenantID                uuid.UUID  `json:"tenant_id"`
	PaidBy                  *uuid.UUID `json:"paid_by,omitempty"`
	VerificationJobID       *uuid.UUID `json:"verification_job_id,omitempty"`
	PaidAt                  *time.Time `json:"paid_at,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// PublisherEarning represents earnings from function sales
type PublisherEarning struct {
	ID                  uuid.UUID        `json:"id"`
	TenantID            uuid.UUID        `json:"tenant_id"`
	PublisherUserID     uuid.UUID        `json:"publisher_user_id"`
	FunctionID          *uuid.UUID       `json:"function_id,omitempty"`
	FunctionName        string           `json:"function_name,omitempty"`
	TransactionType     string           `json:"transaction_type"` // 'sale', 'refund', 'payout'
	AmountCents         int              `json:"amount_cents"`
	Currency            string           `json:"currency"`
	GrossAmountCents    int              `json:"gross_amount_cents"`
	PlatformFeeCents    int              `json:"platform_fee_cents"`
	NetAmountCents      int              `json:"net_amount_cents"`
	PlatformFeePercent  float64          `json:"platform_fee_percent"`
	Status              string           `json:"status"` // 'pending', 'available', 'withdrawn', 'withheld'
	StripePayoutID      *string          `json:"stripe_payout_id,omitempty"`
	PeriodMonth         *int             `json:"period_month,omitempty"`
	PeriodYear          *int             `json:"period_year,omitempty"`
	Metadata            json.RawMessage  `json:"metadata,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
	EarnedAt            time.Time        `json:"earned_at"`
}

// AgentSubscription represents an agent-based subscription
type AgentSubscription struct {
	ID                   uuid.UUID  `json:"id"`
	AgentID              uuid.UUID  `json:"agent_id"`
	TenantID             uuid.UUID  `json:"tenant_id"`
	PlanName             string     `json:"plan_name"` // 'per_agent', 'unlimited'
	PricePerAgentCents   int        `json:"price_per_agent_cents"`
	Currency             string     `json:"currency"`
	MaxAgents            int        `json:"max_agents"`
	Status               string     `json:"status"` // 'active', 'suspended', 'cancelled'
	CurrentPeriodStart   time.Time  `json:"current_period_start"`
	CurrentPeriodEnd     time.Time  `json:"current_period_end"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty"`
	StripeCustomerID     *string    `json:"stripe_customer_id,omitempty"`
	LastPaymentStatus    *string    `json:"last_payment_status,omitempty"`
	LastPaymentAt        *time.Time `json:"last_payment_at,omitempty"`
	CancelAtPeriodEnd    bool       `json:"cancel_at_period_end"`
	CancelledAt          *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// AgentUsage represents usage tracking for agent billing
type AgentUsage struct {
	ID                 uuid.UUID  `json:"id"`
	AgentID            uuid.UUID  `json:"agent_id"`
	TenantID           uuid.UUID  `json:"tenant_id"`
	SubscriptionID     *uuid.UUID `json:"subscription_id,omitempty"`
	PeriodStart        time.Time  `json:"period_start"`
	PeriodEnd          time.Time  `json:"period_end"`
	TotalCalls         int        `json:"total_calls"`
	TotalExecutions    int        `json:"total_executions"`
	TotalErrors        int        `json:"total_errors"`
	TotalLatencyMs     int64      `json:"total_latency_ms"`
	BillableCalls      int        `json:"billable_calls"`
	OverageCalls       int        `json:"overage_calls"`
	EstimatedCostCents int        `json:"estimated_cost_cents"`
	Status             string     `json:"status"` // 'active', 'billed', 'disputed'
	StripeInvoiceID    *string    `json:"stripe_invoice_id,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// PlatformFee represents platform fees collected from marketplace transactions
type PlatformFee struct {
	ID                   uuid.UUID       `json:"id"`
	FeeType              string          `json:"fee_type"` // 'marketplace_sale', 'verification', 'agent_subscription', 'tier_upgrade'
	SourceTransactionID  *uuid.UUID      `json:"source_transaction_id,omitempty"`
	SourceType           string          `json:"source_type"` // 'publisher_earnings', 'function_verification_payment', 'agent_subscription'
	GrossAmountCents     int             `json:"gross_amount_cents"`
	PlatformFeeCents    int             `json:"platform_fee_cents"`
	NetAmountCents      int             `json:"net_amount_cents"`
	PlatformFeePercent  float64         `json:"platform_fee_percent"`
	Currency            string          `json:"currency"`
	TenantID            *uuid.UUID      `json:"tenant_id,omitempty"`
	UserID              *uuid.UUID      `json:"user_id,omitempty"`
	FunctionID          *uuid.UUID      `json:"function_id,omitempty"`
	AgentID             *uuid.UUID      `json:"agent_id,omitempty"`
	Status              string          `json:"status"` // 'collected', 'refunded', 'disputed', 'paid_out'
	StripeTransferID    *string         `json:"stripe_transfer_id,omitempty"`
	PaidOutAt           *time.Time      `json:"paid_out_at,omitempty"`
	PeriodMonth         *int            `json:"period_month,omitempty"`
	PeriodYear          *int            `json:"period_year,omitempty"`
	Metadata            json.RawMessage `json:"metadata,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
}

// PricingTierExtended extends PricingTier with new Moat fields
type PricingTierExtended struct {
	ID                    uuid.UUID       `json:"id"`
	Name                  string          `json:"name"`
	Description           string          `json:"description"`
	PriceCents            int             `json:"price_cents"`
	AnnualPriceCents      *int            `json:"annual_price_cents,omitempty"`
	Currency              string          `json:"currency"`
	BillingCycle          string          `json:"billing_cycle"` // 'monthly' or 'annual'
	Features              json.RawMessage `json:"features"`
	IsActive              bool            `json:"is_active"`
	TierType              string          `json:"tier_type"` // 'subscription'
	StripePriceID         *string         `json:"stripe_price_id,omitempty"`
	StripePriceIDAnnual   *string         `json:"stripe_price_id_annual,omitempty"`
	TrialDays             int             `json:"trial_days"`
	MaxAgents             int             `json:"max_agents"`
	MaxFunctions          int             `json:"max_functions"`
	MaxExecutionsPerMonth int             `json:"max_executions_per_month"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}
