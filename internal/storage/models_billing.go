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
	ID               uuid.UUID  `json:"id"`
	TenantID         uuid.UUID  `json:"tenant_id"`
	SubscriptionID   *uuid.UUID `json:"subscription_id,omitempty"`
	Status           string     `json:"status"`
	AmountDueCents   int        `json:"amount_due_cents"`
	AmountPaidCents  int        `json:"amount_paid_cents"`
	Currency         string     `json:"currency"`
	InvoicePdfURL    string     `json:"invoice_pdf_url,omitempty"`
	HostedInvoiceURL string     `json:"hosted_invoice_url,omitempty"`
	PeriodStart      *time.Time `json:"period_start,omitempty"`
	PeriodEnd        *time.Time `json:"period_end,omitempty"`
	DueDate          *time.Time `json:"due_date,omitempty"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
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
