package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AgentFinancialTransaction is a ledger row for agent wallet / billing events.
type AgentFinancialTransaction struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID      `json:"tenant_id" gorm:"type:uuid;not null"`
	AgentID     string         `json:"agent_id" gorm:"not null"`
	Kind        string         `json:"kind" gorm:"not null"`
	AmountUSD   float64        `json:"amount_usd" gorm:"type:decimal(14,4);not null"`
	Status      string         `json:"status" gorm:"not null;default:completed"`
	Provider    *string        `json:"provider,omitempty"`
	ProviderRef *string        `json:"provider_ref,omitempty"`
	Metadata    datatypes.JSON `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt   time.Time      `json:"created_at" gorm:"not null;default:now()"`
}

// TableName for GORM.
func (AgentFinancialTransaction) TableName() string {
	return "agent_financial_transactions"
}

// FinancialTransactionRepository persists idempotent financial ledger rows.
type FinancialTransactionRepository struct {
	db *gorm.DB
}

// AgentWalletSummary is a ledger-derived aggregate for wallet views.
type AgentWalletSummary struct {
	BalanceUSD     float64    `json:"balance_usd"`
	TotalEarnedUSD float64    `json:"total_earned_usd"`
	TotalSpentUSD  float64    `json:"total_spent_usd"`
	LastEarningAt  *time.Time `json:"last_earning_at,omitempty"`
	LastSpendingAt *time.Time `json:"last_spending_at,omitempty"`
}

// NewFinancialTransactionRepository creates a repository backed by GORM.
func NewFinancialTransactionRepository(db *gorm.DB) *FinancialTransactionRepository {
	return &FinancialTransactionRepository{db: db}
}

// CreateIdempotent inserts a row; on unique (provider, provider_ref) conflict, does nothing.
// Returns (created=true) if a new row was inserted.
func (r *FinancialTransactionRepository) CreateIdempotent(ctx context.Context, tx *AgentFinancialTransaction) (created bool, err error) {
	if tx == nil {
		return false, errors.New("nil transaction")
	}
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now().UTC()
	}
	if len(tx.Metadata) == 0 {
		tx.Metadata = datatypes.JSON([]byte("{}"))
	}

	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "provider"}, {Name: "provider_ref"}},
		DoNothing: true,
	}).Create(tx)

	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// GetByProviderRef returns a row by provider + ref if present.
func (r *FinancialTransactionRepository) GetByProviderRef(ctx context.Context, provider, providerRef string) (*AgentFinancialTransaction, error) {
	if provider == "" || providerRef == "" {
		return nil, fmt.Errorf("provider and provider_ref required")
	}
	var row AgentFinancialTransaction
	err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_ref = ?", provider, providerRef).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ListByAgent returns ledger rows for an agent and tenant, newest first.
func (r *FinancialTransactionRepository) ListByAgent(ctx context.Context, tenantID uuid.UUID, agentID string, limit, offset int) ([]AgentFinancialTransaction, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	q := r.db.WithContext(ctx).Model(&AgentFinancialTransaction{}).
		Where("tenant_id = ? AND agent_id = ?", tenantID, agentID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []AgentFinancialTransaction
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}

// GetAgentWalletSummary computes wallet aggregates from completed financial transactions.
func (r *FinancialTransactionRepository) GetAgentWalletSummary(ctx context.Context, tenantID uuid.UUID, agentID string) (*AgentWalletSummary, error) {
	summary := &AgentWalletSummary{}

	// Treat known credit kinds as inflows and debit kinds as outflows.
	const kindCredits = "'credit_purchase','transfer_in','refund'"
	const kindDebits = "'execution_debit','transfer_out'"

	row := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(CASE
				WHEN kind IN (`+kindCredits+`) THEN amount_usd
				WHEN kind IN (`+kindDebits+`) THEN -amount_usd
				WHEN kind = 'adjustment' THEN amount_usd
				ELSE 0
			END), 0) AS balance_usd,
			COALESCE(SUM(CASE
				WHEN kind IN (`+kindCredits+`) THEN amount_usd
				WHEN kind = 'adjustment' AND amount_usd > 0 THEN amount_usd
				ELSE 0
			END), 0) AS total_earned_usd,
			COALESCE(SUM(CASE
				WHEN kind IN (`+kindDebits+`) THEN amount_usd
				WHEN kind = 'adjustment' AND amount_usd < 0 THEN -amount_usd
				ELSE 0
			END), 0) AS total_spent_usd,
			MAX(CASE
				WHEN kind IN (`+kindCredits+`) OR (kind = 'adjustment' AND amount_usd > 0) THEN created_at
				ELSE NULL
			END) AS last_earning_at,
			MAX(CASE
				WHEN kind IN (`+kindDebits+`) OR (kind = 'adjustment' AND amount_usd < 0) THEN created_at
				ELSE NULL
			END) AS last_spending_at
		FROM agent_financial_transactions
		WHERE tenant_id = ? AND agent_id = ? AND status = 'completed'
	`, tenantID, agentID).Row()

	if err := row.Scan(
		&summary.BalanceUSD,
		&summary.TotalEarnedUSD,
		&summary.TotalSpentUSD,
		&summary.LastEarningAt,
		&summary.LastSpendingAt,
	); err != nil {
		return nil, err
	}

	return summary, nil
}

// MetadataMap decodes metadata JSON to a map.
func (t *AgentFinancialTransaction) MetadataMap() map[string]any {
	if len(t.Metadata) == 0 {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal(t.Metadata, &m); err != nil || m == nil {
		return map[string]any{}
	}
	return m
}
