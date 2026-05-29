package economy

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PurchaseFunctionParams holds inputs for an atomic function purchase.
type PurchaseFunctionParams struct {
	TenantID         uuid.UUID
	UserID           uuid.UUID
	AgentID          string
	FunctionAuthor   string
	FunctionName     string
	PublishedID      uuid.UUID
	SellerAgentID    string
	PricePaidUSD     float64
	IdempotencyKey   string
	ClientIP         string
}

// CompleteFunctionPurchase debits buyer, credits seller, and records the purchase in one transaction.
// The second return value is true when an idempotency key replayed an existing purchase.
func (s *Service) CompleteFunctionPurchase(
	ctx context.Context,
	identityRepo *identity.Repository,
	params PurchaseFunctionParams,
) (*identity.FunctionPurchase, bool, error) {
	if params.IdempotencyKey != "" {
		existingID, err := identityRepo.GetIdempotentPurchaseID(ctx, params.TenantID, params.IdempotencyKey)
		if err != nil {
			return nil, false, err
		}
		if existingID != nil {
			purchase, err := identityRepo.GetFunctionPurchaseByID(ctx, params.TenantID, *existingID)
			if err != nil {
				return nil, false, err
			}
			return purchase, true, nil
		}
	}

	if _, err := identityRepo.GetCompletedFunctionPurchase(ctx, params.AgentID, params.FunctionAuthor, params.FunctionName); err == nil {
		return nil, false, identity.ErrPurchaseAlreadyExists
	} else if !errors.Is(err, identity.ErrPurchaseNotFound) {
		return nil, false, err
	}

	var purchase identity.FunctionPurchase

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if params.PricePaidUSD > 0 {
			if err := debitWalletInTx(tx, params.AgentID, params.PricePaidUSD, "function_purchase", map[string]any{
				"function_author": params.FunctionAuthor,
				"function_name":   params.FunctionName,
			}); err != nil {
				return err
			}

			sellerID := params.SellerAgentID
			if sellerID != "" && sellerID != params.AgentID {
				if err := creditWalletInTx(tx, sellerID, params.PricePaidUSD, "function_sale", map[string]any{
					"function_author": params.FunctionAuthor,
					"function_name":   params.FunctionName,
					"buyer_agent_id":  params.AgentID,
				}); err != nil {
					return fmt.Errorf("credit seller wallet: %w", err)
				}
			}
		}

		purchase = identity.FunctionPurchase{
			ID:             uuid.New(),
			AgentID:        params.AgentID,
			FunctionAuthor: params.FunctionAuthor,
			FunctionName:   params.FunctionName,
			PublishedID:    params.PublishedID,
			PricePaidUSD:   params.PricePaidUSD,
			Status:         "completed",
		}
		if err := createFunctionPurchaseInTx(tx, &purchase); err != nil {
			return err
		}
		if err := insertIdempotencyInTx(tx, params.TenantID, params.IdempotencyKey, purchase.ID); err != nil {
			return err
		}
		if err := insertAuditInTx(tx, params, purchase.ID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, false, err
	}

	return &purchase, false, nil
}

func debitWalletInTx(tx *gorm.DB, agentID string, amount float64, transactionType string, metadata map[string]any) error {
	var wallet identity.AgentWallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("agent_id = ?", agentID).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("insufficient funds: wallet not found for agent %s", agentID)
		}
		return err
	}
	if wallet.BalanceUSD < amount {
		return fmt.Errorf("insufficient funds: have %.4f, need %.4f", wallet.BalanceUSD, amount)
	}

	now := time.Now()
	wallet.BalanceUSD -= amount
	wallet.TotalSpentUSD += amount
	wallet.LastSpendingAt = &now
	wallet.UpdatedAt = now
	if err := tx.Save(&wallet).Error; err != nil {
		return err
	}

	transaction := identity.RevenueTransaction{
		ID:              uuid.New(),
		FromAgentID:     &agentID,
		ToAgentID:       "system",
		AmountUSD:       amount,
		TransactionType: transactionType,
		Status:          "completed",
		CreatedAt:       now,
	}
	applyTransactionMetadata(&transaction, metadata)
	return tx.Create(&transaction).Error
}

func creditWalletInTx(tx *gorm.DB, agentID string, amount float64, transactionType string, metadata map[string]any) error {
	var wallet identity.AgentWallet
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("agent_id = ?", agentID).First(&wallet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		wallet = identity.AgentWallet{
			ID:        uuid.New(),
			AgentID:   agentID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := tx.Create(&wallet).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	now := time.Now()
	wallet.BalanceUSD += amount
	wallet.TotalEarnedUSD += amount
	wallet.LastEarningAt = &now
	wallet.UpdatedAt = now
	if err := tx.Save(&wallet).Error; err != nil {
		return err
	}

	transaction := identity.RevenueTransaction{
		ID:              uuid.New(),
		ToAgentID:       agentID,
		AmountUSD:       amount,
		TransactionType: transactionType,
		Status:          "completed",
		CreatedAt:       now,
	}
	applyTransactionMetadata(&transaction, metadata)
	return tx.Create(&transaction).Error
}

func applyTransactionMetadata(transaction *identity.RevenueTransaction, metadata map[string]any) {
	if metadata == nil {
		return
	}
	if sessionID, ok := metadata["session_id"].(string); ok {
		transaction.SessionID = &sessionID
	}
	if executionID, ok := metadata["execution_id"].(string); ok {
		transaction.ExecutionID = &executionID
	}
}

func createFunctionPurchaseInTx(tx *gorm.DB, purchase *identity.FunctionPurchase) error {
	if purchase.ID == uuid.Nil {
		purchase.ID = uuid.New()
	}
	now := time.Now()
	purchase.CreatedAt = now
	purchase.UpdatedAt = now
	return tx.Create(purchase).Error
}

func insertIdempotencyInTx(tx *gorm.DB, tenantID uuid.UUID, key string, purchaseID uuid.UUID) error {
	if key == "" {
		return nil
	}
	return tx.Exec(`
		INSERT INTO marketplace_purchase_idempotency (tenant_id, idempotency_key, purchase_id)
		VALUES (?, ?, ?)`,
		tenantID, key, purchaseID,
	).Error
}

func insertAuditInTx(tx *gorm.DB, params PurchaseFunctionParams, purchaseID uuid.UUID) error {
	var userID interface{}
	if params.UserID != uuid.Nil {
		userID = params.UserID
	}
	var clientIP interface{}
	if params.ClientIP != "" {
		clientIP = params.ClientIP
	}
	return tx.Exec(`
		INSERT INTO marketplace_purchase_audit_log (
			tenant_id, user_id, agent_id, function_author, function_name,
			purchase_id, price_paid_usd, idempotency_key, client_ip, event_type
		) VALUES (?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?::inet, 'function_purchase')`,
		params.TenantID, userID, params.AgentID, params.FunctionAuthor, params.FunctionName,
		purchaseID, params.PricePaidUSD, params.IdempotencyKey, clientIP,
	).Error
}
