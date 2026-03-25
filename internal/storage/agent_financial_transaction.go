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
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID   uuid.UUID      `json:"tenant_id" gorm:"type:uuid;not null"`
	AgentID    string         `json:"agent_id" gorm:"not null"`
	Kind       string         `json:"kind" gorm:"not null"`
	AmountUSD  float64        `json:"amount_usd" gorm:"type:decimal(14,4);not null"`
	Status     string         `json:"status" gorm:"not null;default:completed"`
	Provider   *string        `json:"provider,omitempty"`
	ProviderRef *string       `json:"provider_ref,omitempty"`
	Metadata   datatypes.JSON `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt  time.Time      `json:"created_at" gorm:"not null;default:now()"`
}

// TableName for GORM.
func (AgentFinancialTransaction) TableName() string {
	return "agent_financial_transactions"
}

// FinancialTransactionRepository persists idempotent financial ledger rows.
type FinancialTransactionRepository struct {
	db *gorm.DB
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
