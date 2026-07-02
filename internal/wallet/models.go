package wallet

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// OwnerType indicates what kind of entity owns the wallet
const (
	OwnerTypeUser  = "user"
	OwnerTypeAgent = "agent"
)

// WalletType indicates the wallet's purpose
const (
	WalletTypeUnified   = "unified"   // Can be used for both registry and execution
	WalletTypeRegistry  = "registry"  // Only for registry fees (publish, version updates)
	WalletTypeExecution = "execution" // Only for function execution
)

// TransactionType categorizes wallet transactions
const (
	TransactionTypeCredit          = "credit"
	TransactionTypeDebit           = "debit"
	TransactionTypeFeePayment      = "fee_payment"
	TransactionTypeExecutionCharge = "execution_charge"
	TransactionTypeCommission      = "commission"
	TransactionTypeRefund          = "refund"
	TransactionTypeTransferIn      = "transfer_in"
	TransactionTypeTransferOut     = "transfer_out"
	TransactionTypeAdjustment      = "adjustment"
)

// TransactionStatus indicates the state of a transaction
const (
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
	TransactionStatusReversed  = "reversed"
)

// WalletStatus indicates the wallet's operational state
const (
	WalletStatusActive    = "active"
	WalletStatusSuspended = "suspended"
	WalletStatusClosed    = "closed"
)

// BillingMode indicates how agent billing is handled
const (
	BillingModePerWallet = "per_wallet"
	BillingModePerAgent  = "per_agent"
	BillingModePerTenant = "per_tenant"
	BillingModePerTeam   = "per_team"
)

// FeeType categorizes platform fees
const (
	FeeTypePublish       = "publish"
	FeeTypeVersionUpdate = "version_update"
	FeeTypeCommission    = "commission"
)

// Wallet represents a unified wallet that can be owned by a user or agent
// This merges the functionality of user_wallets and agent_billing_controls.credit_balance_usd
type Wallet struct {
	ID                    uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OwnerType             string          `json:"owner_type" gorm:"not null"`                        // 'user' or 'agent'
	OwnerID               string          `json:"owner_id" gorm:"not null;uniqueIndex:unique_owner"` // user_id or agent_id
	UserID                *uuid.UUID      `json:"user_id,omitempty" gorm:"type:uuid;index"`          // Set when owner_type='user'
	AgentID               *string         `json:"agent_id,omitempty" gorm:"index"`                   // Set when owner_type='agent'
	WalletType            string          `json:"wallet_type" gorm:"not null;default:'unified'"`     // 'unified', 'registry', 'execution'
	BalanceUSD            float64         `json:"balance_usd" gorm:"type:decimal(14,4);not null;default:0"`
	BalanceLocal          *float64        `json:"balance_local,omitempty" gorm:"type:decimal(14,4)"`
	Currency              string          `json:"currency" gorm:"not null;default:'USD'"`
	ExchangeRateToUSD     *float64        `json:"exchange_rate_to_usd,omitempty" gorm:"type:decimal(14,6)"`
	LifetimeEarningsUSD   float64         `json:"lifetime_earnings_usd" gorm:"type:decimal(14,4);not null;default:0"`
	LifetimeSpentUSD      float64         `json:"lifetime_spent_usd" gorm:"type:decimal(14,4);not null;default:0"`
	SpendCapMonthlyUSD    *float64        `json:"spend_cap_monthly_usd,omitempty" gorm:"type:decimal(10,2)"`
	SpendCapWeeklyUSD     *float64        `json:"spend_cap_weekly_usd,omitempty" gorm:"type:decimal(10,2)"`
	SpendCapDailyUSD      *float64        `json:"spend_cap_daily_usd,omitempty" gorm:"type:decimal(10,2)"`
	AlertThresholds       pq.Float64Array `json:"alert_thresholds" gorm:"type:decimal[];default:'{0.5,0.8,0.95}'"`
	BillingMode           string          `json:"billing_mode" gorm:"not null;default:'per_wallet'"`
	TeamID                *uuid.UUID      `json:"team_id,omitempty" gorm:"type:uuid"`
	Status                string          `json:"status" gorm:"not null;default:'active'"`
	Suspended             bool            `json:"suspended" gorm:"not null;default:false"`
	SuspendedAt           *time.Time      `json:"suspended_at,omitempty"`
	SuspensionReason      *string         `json:"suspension_reason,omitempty"`
	AutoTopupEnabled      bool            `json:"auto_topup_enabled" gorm:"not null;default:false"`
	AutoTopupThresholdUSD float64         `json:"auto_topup_threshold_usd" gorm:"type:decimal(10,2);default:0"`
	ClosedAt              *time.Time      `json:"closed_at,omitempty"`
	ClosureReason         *string         `json:"closure_reason,omitempty"`
	CreatedAt             time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time       `json:"updated_at" gorm:"autoUpdateTime"`

	// Security fields (not exposed in JSON)
	BalanceEncryptedJSON *string `json:"-" gorm:"column:balance_encrypted;type:text"` // Encrypted balance for verification
}

// TableName returns the database table name
func (Wallet) TableName() string {
	return "wallets"
}

// IsActive returns true if the wallet is active and can be used
func (w *Wallet) IsActive() bool {
	return w.Status == WalletStatusActive
}

// HasSufficientBalance returns true if the wallet has enough balance for the given amount
func (w *Wallet) HasSufficientBalance(amount float64) bool {
	return w.BalanceUSD >= amount
}

// WalletTransaction represents a single transaction in the unified wallet ledger
// This merges the functionality of fee_transactions and agent_financial_transactions
type WalletTransaction struct {
	ID                  uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WalletID            uuid.UUID  `json:"wallet_id" gorm:"type:uuid;not null;index"`
	TransactionType     string     `json:"transaction_type" gorm:"not null"`
	AmountUSD           float64    `json:"amount_usd" gorm:"type:decimal(14,4);not null"`
	BalanceBeforeUSD    float64    `json:"balance_before_usd" gorm:"type:decimal(14,4);not null"`
	BalanceAfterUSD     float64    `json:"balance_after_usd" gorm:"type:decimal(14,4);not null"`
	Status              string     `json:"status" gorm:"not null;default:'completed'"`
	Reference           *string    `json:"reference,omitempty" gorm:"type:text;index"`
	ParentTransactionID *uuid.UUID `json:"parent_transaction_id,omitempty" gorm:"type:uuid"`
	TriggeredByType     *string    `json:"triggered_by_type,omitempty"` // 'user', 'agent', 'system', 'admin', 'webhook'
	TriggeredByID       *string    `json:"triggered_by_id,omitempty"`
	ExecutionID         *uuid.UUID `json:"execution_id,omitempty" gorm:"type:uuid;index"`
	FunctionID          *uuid.UUID `json:"function_id,omitempty" gorm:"type:uuid"`
	FeeType             *string    `json:"fee_type,omitempty"` // 'publish', 'version_update', 'commission'
	Metadata            []byte     `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt           time.Time  `json:"created_at" gorm:"autoCreateTime"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	ReversedAt          *time.Time `json:"reversed_at,omitempty"`
	IdempotencyKey      *string    `json:"idempotency_key,omitempty" gorm:"type:text;uniqueIndex:idx_idempotency_unique"`

	// Relationships
	Wallet            *Wallet            `json:"wallet,omitempty" gorm:"foreignKey:WalletID;references:ID"`
	ParentTransaction *WalletTransaction `json:"parent_transaction,omitempty" gorm:"foreignKey:ParentTransactionID;references:ID"`
}

// TableName returns the database table name
func (WalletTransaction) TableName() string {
	return "wallet_transactions"
}

// IsCredit returns true if this transaction adds funds to the wallet
func (t *WalletTransaction) IsCredit() bool {
	switch t.TransactionType {
	case TransactionTypeCredit, TransactionTypeRefund, TransactionTypeTransferIn:
		return true
	default:
		return false
	}
}

// IsDebit returns true if this transaction removes funds from the wallet
func (t *WalletTransaction) IsDebit() bool {
	return !t.IsCredit()
}

// CreditRequest represents a request to credit a wallet
type CreditRequest struct {
	WalletID       uuid.UUID
	AmountUSD      float64
	Reference      string // External reference (Stripe payment intent)
	IdempotencyKey string // Unique key to prevent duplicates
	TriggeredBy    TriggeredByInfo
	Metadata       map[string]interface{}
}

// DebitRequest represents a request to debit a wallet
type DebitRequest struct {
	WalletID        uuid.UUID
	AmountUSD       float64
	TransactionType string // fee_payment, execution_charge, etc.
	Reference       string // Optional reference
	TriggeredBy     TriggeredByInfo
	ExecutionID     *uuid.UUID // For execution charges
	FunctionID      *uuid.UUID // For execution charges
	FeeType         *string    // For fee payments
	Metadata        map[string]interface{}
}

// TriggeredByInfo identifies who/what triggered a transaction
type TriggeredByInfo struct {
	Type string // 'user', 'agent', 'system', 'admin', 'webhook'
	ID   string // user_id or agent_id
}

// WalletSummary provides aggregated wallet statistics
type WalletSummary struct {
	WalletID                 uuid.UUID  `json:"wallet_id"`
	OwnerType                string     `json:"owner_type"`
	OwnerID                  string     `json:"owner_id"`
	UserID                   *uuid.UUID `json:"user_id,omitempty"`
	AgentID                  *string    `json:"agent_id,omitempty"`
	BalanceUSD               float64    `json:"balance_usd"`
	Status                   string     `json:"status"`
	TotalCreditsUSD          float64    `json:"total_credits_usd"`
	TotalDebitsUSD           float64    `json:"total_debits_usd"`
	TotalFeesPaidUSD         float64    `json:"total_fees_paid_usd"`
	TotalExecutionChargesUSD float64    `json:"total_execution_charges_usd"`
	TotalCommissionsUSD      float64    `json:"total_commissions_usd"`
	TotalTransactions        int64      `json:"total_transactions"`
	PendingTransactions      int64      `json:"pending_transactions"`
	LastTransactionAt        *time.Time `json:"last_transaction_at,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

// BalanceUpdate represents the result of a credit or debit operation
type BalanceUpdate struct {
	WalletID        uuid.UUID `json:"wallet_id"`
	PreviousBalance float64   `json:"previous_balance_usd"`
	CurrentBalance  float64   `json:"current_balance_usd"`
	Amount          float64   `json:"amount_usd"`
	TransactionID   uuid.UUID `json:"transaction_id"`
}

// SpendCapCheck represents the result of a spend cap validation
type SpendCapCheck struct {
	Allowed         bool    `json:"allowed"`
	DailySpendUSD   float64 `json:"daily_spend_usd,omitempty"`
	MonthlySpendUSD float64 `json:"monthly_spend_usd,omitempty"`
	DailyCapUSD     float64 `json:"daily_cap_usd,omitempty"`
	MonthlyCapUSD   float64 `json:"monthly_cap_usd,omitempty"`
	CapUtilization  float64 `json:"cap_utilization,omitempty"`
	Reason          string  `json:"reason,omitempty"`
}

// WalletCreationRequest represents a request to create a new wallet
type WalletCreationRequest struct {
	OwnerType          string
	OwnerID            string
	UserID             *uuid.UUID
	AgentID            *string
	WalletType         string
	InitialBalanceUSD  float64
	SpendCapMonthlyUSD *float64
	SpendCapWeeklyUSD  *float64
	SpendCapDailyUSD   *float64
	BillingMode        string
	TeamID             *uuid.UUID
}

// WalletBalanceAudit records balance drift incidents for auditing
type WalletBalanceAudit struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WalletID        uuid.UUID  `json:"wallet_id" gorm:"type:uuid;not null;index"`
	StoredBalance   float64   `json:"stored_balance" gorm:"type:decimal(15,6);not null"`
	ComputedBalance float64   `json:"computed_balance" gorm:"type:decimal(15,6);not null"`
	Drift           float64   `json:"drift" gorm:"type:decimal(15,6);not null"`
	Fixed           bool      `json:"fixed" gorm:"default:false"`
	FixedAt         *time.Time `json:"fixed_at,omitempty"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (WalletBalanceAudit) TableName() string {
	return "wallet_balance_audit"
}
