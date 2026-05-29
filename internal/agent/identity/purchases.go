package identity

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrPurchaseAlreadyExists = errors.New("function already purchased by this agent")
	ErrPurchaseNotFound        = errors.New("purchase not found")
)

// ListFunctionPurchasesForTenant returns completed purchases for agents owned by the tenant.
func (r *Repository) ListFunctionPurchasesForTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]FunctionPurchase, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	var purchases []FunctionPurchase
	err := r.db.WithContext(ctx).
		Table("function_purchases AS fp").
		Select("fp.*").
		Joins("INNER JOIN agent_identities ai ON ai.agent_id = fp.agent_id").
		Where("ai.tenant_id = ? AND fp.status = ?", tenantID, "completed").
		Order("fp.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&purchases).Error
	if err != nil {
		return nil, fmt.Errorf("list function purchases: %w", err)
	}
	return purchases, nil
}

// GetCompletedFunctionPurchase returns an active purchase for an agent/function pair.
func (r *Repository) GetCompletedFunctionPurchase(ctx context.Context, agentID, author, name string) (*FunctionPurchase, error) {
	var purchase FunctionPurchase
	err := r.db.WithContext(ctx).
		Where("agent_id = ? AND function_author = ? AND function_name = ? AND status = ?",
			agentID, author, name, "completed").
		First(&purchase).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPurchaseNotFound
		}
		return nil, err
	}
	return &purchase, nil
}

// ListAgentHiringsForTenant returns hiring records initiated by the tenant.
func (r *Repository) ListAgentHiringsForTenant(ctx context.Context, tenantID string, limit, offset int) ([]AgentHiring, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	var hirings []AgentHiring
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&hirings).Error
	if err != nil {
		return nil, fmt.Errorf("list agent hirings: %w", err)
	}
	return hirings, nil
}

// GetFunctionPurchaseByID returns a purchase if it belongs to the tenant (via agent).
func (r *Repository) GetFunctionPurchaseByID(ctx context.Context, tenantID uuid.UUID, purchaseID uuid.UUID) (*FunctionPurchase, error) {
	var purchase FunctionPurchase
	err := r.db.WithContext(ctx).
		Table("function_purchases AS fp").
		Select("fp.*").
		Joins("INNER JOIN agent_identities ai ON ai.agent_id = fp.agent_id").
		Where("ai.tenant_id = ? AND fp.id = ?", tenantID, purchaseID).
		First(&purchase).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPurchaseNotFound
		}
		return nil, err
	}
	return &purchase, nil
}

// GetIdempotentPurchaseID looks up a prior purchase for an idempotency key.
func (r *Repository) GetIdempotentPurchaseID(ctx context.Context, tenantID uuid.UUID, key string) (*uuid.UUID, error) {
	if key == "" {
		return nil, nil
	}
	var purchaseID uuid.UUID
	result := r.db.WithContext(ctx).Raw(`
		SELECT purchase_id FROM marketplace_purchase_idempotency
		WHERE tenant_id = ? AND idempotency_key = ?
		LIMIT 1`, tenantID, key).Scan(&purchaseID)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 || purchaseID == uuid.Nil {
		return nil, nil
	}
	return &purchaseID, nil
}

// PurchaseAuditEntry is written after a successful purchase.
type PurchaseAuditEntry struct {
	TenantID        uuid.UUID
	UserID          uuid.UUID
	AgentID         string
	FunctionAuthor  string
	FunctionName    string
	PurchaseID      uuid.UUID
	PricePaidUSD    float64
	IdempotencyKey  string
	ClientIP        string
}

// InsertPurchaseAuditLog records a marketplace purchase audit event.
func (r *Repository) InsertPurchaseAuditLog(ctx context.Context, entry PurchaseAuditEntry) error {
	var userID interface{}
	if entry.UserID != uuid.Nil {
		userID = entry.UserID
	}
	var clientIP interface{}
	if entry.ClientIP != "" {
		clientIP = entry.ClientIP
	}
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO marketplace_purchase_audit_log (
			tenant_id, user_id, agent_id, function_author, function_name,
			purchase_id, price_paid_usd, idempotency_key, client_ip, event_type
		) VALUES (?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?::inet, 'function_purchase')`,
		entry.TenantID, userID, entry.AgentID, entry.FunctionAuthor, entry.FunctionName,
		entry.PurchaseID, entry.PricePaidUSD, entry.IdempotencyKey, clientIP,
	).Error
}
