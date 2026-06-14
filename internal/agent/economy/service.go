package economy

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Service handles agent economic transactions
type Service struct {
	db *gorm.DB
}

// NewService creates a new economy service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GetOrCreateWallet gets or creates a wallet for an agent
func (s *Service) GetOrCreateWallet(ctx context.Context, agentID string) (*identity.AgentWallet, error) {
	var wallet identity.AgentWallet
	err := s.db.WithContext(ctx).Where("agent_id = ?", agentID).First(&wallet).Error
	if err == nil {
		return &wallet, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create wallet
	wallet = identity.AgentWallet{
		ID:               uuid.New(),
		AgentID:          agentID,
		BalanceUSD:       0,
		EscrowBalanceUSD: 0,
		TotalEarnedUSD:   0,
		TotalSpentUSD:    0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

// GetWallet gets the wallet for an agent
func (s *Service) GetWallet(ctx context.Context, agentID string) (*identity.AgentWallet, error) {
	var wallet identity.AgentWallet
	err := s.db.WithContext(ctx).Where("agent_id = ?", agentID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// Credit adds funds to an agent's wallet
// SECURITY FIX: Uses SELECT FOR UPDATE to prevent TOCTOU race conditions
func (s *Service) Credit(ctx context.Context, agentID string, amount float64, transactionType string, metadata map[string]any) (*identity.RevenueTransaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("credit amount must be positive")
	}

	tx := s.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// SECURITY FIX: Use SELECT FOR UPDATE to lock the wallet row and prevent race conditions
	var wallet identity.AgentWallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("agent_id = ?", agentID).First(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create wallet if not exists, within transaction
			wallet = identity.AgentWallet{
				ID:               uuid.New(),
				AgentID:          agentID,
				BalanceUSD:       0,
				EscrowBalanceUSD: 0,
				TotalEarnedUSD:   0,
				TotalSpentUSD:    0,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			if err := tx.Create(&wallet).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Update wallet (now under row lock)
	wallet.BalanceUSD += amount
	wallet.TotalEarnedUSD += amount
	now := time.Now()
	wallet.LastEarningAt = &now
	wallet.UpdatedAt = now

	if err := tx.Save(&wallet).Error; err != nil {
		return nil, err
	}

	// Create transaction record
	transaction := &identity.RevenueTransaction{
		ID:              uuid.New(),
		ToAgentID:       agentID,
		AmountUSD:       amount,
		TransactionType: transactionType,
		Status:          "completed",
		CreatedAt:       now,
	}

	if metadata != nil {
		if sessionID, ok := metadata["session_id"].(string); ok {
			transaction.SessionID = &sessionID
		}
		if executionID, ok := metadata["execution_id"].(string); ok {
			transaction.ExecutionID = &executionID
		}
		if parentExecID, ok := metadata["parent_execution_id"].(string); ok {
			transaction.ParentExecutionID = &parentExecID
		}
	}

	if err := tx.Create(transaction).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return transaction, nil
}

// Debit removes funds from an agent's wallet
// SECURITY FIX: Uses SELECT FOR UPDATE to prevent TOCTOU race conditions
func (s *Service) Debit(ctx context.Context, agentID string, amount float64, transactionType string, metadata map[string]any) (*identity.RevenueTransaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("debit amount must be positive")
	}

	tx := s.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// SECURITY FIX: Use SELECT FOR UPDATE to lock the wallet row and prevent race conditions
	var wallet identity.AgentWallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("agent_id = ?", agentID).First(&wallet).Error; err != nil {
		return nil, fmt.Errorf("failed to get wallet for debit: %w", err)
	}

	if wallet.BalanceUSD < amount {
		return nil, fmt.Errorf("insufficient funds: have %.2f, need %.2f", wallet.BalanceUSD, amount)
	}

	// Update wallet (now under row lock)
	wallet.BalanceUSD -= amount
	wallet.TotalSpentUSD += amount
	now := time.Now()
	wallet.LastSpendingAt = &now
	wallet.UpdatedAt = now

	if err := tx.Save(&wallet).Error; err != nil {
		return nil, err
	}

	// Create transaction record
	transaction := &identity.RevenueTransaction{
		ID:              uuid.New(),
		FromAgentID:     &agentID,
		ToAgentID:       "system", // System receives payment
		AmountUSD:       amount,
		TransactionType: transactionType,
		Status:          "completed",
		CreatedAt:       now,
	}

	if metadata != nil {
		if sessionID, ok := metadata["session_id"].(string); ok {
			transaction.SessionID = &sessionID
		}
		if executionID, ok := metadata["execution_id"].(string); ok {
			transaction.ExecutionID = &executionID
		}
	}

	if err := tx.Create(transaction).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return transaction, nil
}

// Transfer transfers funds between agents
// SECURITY FIX: Uses SELECT FOR UPDATE on both wallets to prevent race conditions
func (s *Service) Transfer(ctx context.Context, fromAgentID, toAgentID string, amount float64, transactionType string, metadata map[string]any) (*identity.RevenueTransaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("transfer amount must be positive")
	}

	tx := s.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// SECURITY FIX: Lock both wallets in consistent order to prevent deadlocks
	// Always lock in alphabetical order by agentID to prevent A-B/A-B deadlocks
	firstID, secondID := fromAgentID, toAgentID
	if fromAgentID > toAgentID {
		firstID, secondID = toAgentID, fromAgentID
	}

	var firstWallet, secondWallet identity.AgentWallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("agent_id = ?", firstID).First(&firstWallet).Error; err != nil {
		return nil, fmt.Errorf("failed to lock first wallet: %w", err)
	}
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("agent_id = ?", secondID).First(&secondWallet).Error; err != nil {
		return nil, fmt.Errorf("failed to lock second wallet: %w", err)
	}

	// Identify sender and receiver from the locked wallets
	var fromWallet, toWallet *identity.AgentWallet
	if firstID == fromAgentID {
		fromWallet = &firstWallet
		toWallet = &secondWallet
	} else {
		fromWallet = &secondWallet
		toWallet = &firstWallet
	}

	if fromWallet.BalanceUSD < amount {
		return nil, fmt.Errorf("insufficient funds for transfer: have %.2f, need %.2f", fromWallet.BalanceUSD, amount)
	}

	// Debit from sender
	fromWallet.BalanceUSD -= amount
	fromWallet.TotalSpentUSD += amount
	now := time.Now()
	fromWallet.LastSpendingAt = &now
	fromWallet.UpdatedAt = now
	if err := tx.Save(fromWallet).Error; err != nil {
		return nil, err
	}

	// Credit to receiver
	toWallet.BalanceUSD += amount
	toWallet.TotalEarnedUSD += amount
	toWallet.LastEarningAt = &now
	toWallet.UpdatedAt = now
	if err := tx.Save(toWallet).Error; err != nil {
		return nil, err
	}

	// Create transaction record
	transaction := &identity.RevenueTransaction{
		ID:              uuid.New(),
		FromAgentID:     &fromAgentID,
		ToAgentID:       toAgentID,
		AmountUSD:       amount,
		TransactionType: transactionType,
		Status:          "completed",
		CreatedAt:       now,
	}

	if metadata != nil {
		if sessionID, ok := metadata["session_id"].(string); ok {
			transaction.SessionID = &sessionID
		}
		if executionID, ok := metadata["execution_id"].(string); ok {
			transaction.ExecutionID = &executionID
		}
		if parentExecID, ok := metadata["parent_execution_id"].(string); ok {
			transaction.ParentExecutionID = &parentExecID
		}
	}

	if err := tx.Create(transaction).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return transaction, nil
}

// ProcessDelegationPayment processes payment for task delegation
// Revenue split: parent gets retained amount, child gets delegation amount
func (s *Service) ProcessDelegationPayment(ctx context.Context, parentAgentID, childAgentID string, totalAmount float64, splitPercent float64, metadata map[string]any) error {
	if totalAmount <= 0 {
		return fmt.Errorf("total amount must be positive")
	}

	// Calculate split
	parentShare := totalAmount * (splitPercent / 100)
	childShare := totalAmount - parentShare

	// Credit parent (they keep their share)
	_, err := s.Credit(ctx, parentAgentID, parentShare, identity.TransactionTypeDelegationPayment, metadata)
	if err != nil {
		return fmt.Errorf("failed to credit parent: %w", err)
	}

	// Credit child
	metadata["parent_agent_id"] = parentAgentID
	_, err = s.Credit(ctx, childAgentID, childShare, identity.TransactionTypeDelegationPayment, metadata)
	if err != nil {
		return fmt.Errorf("failed to credit child: %w", err)
	}

	return nil
}

// GetTransactions retrieves transactions for an agent
func (s *Service) GetTransactions(ctx context.Context, agentID string, limit, offset int) ([]identity.RevenueTransaction, int64, error) {
	if limit <= 0 {
		limit = 50
	}

	var transactions []identity.RevenueTransaction
	var total int64

	query := s.db.WithContext(ctx).Model(&identity.RevenueTransaction{}).
		Where("to_agent_id = ? OR from_agent_id = ?", agentID, agentID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// AddToEscrow adds funds to escrow
// SECURITY FIX: Uses SELECT FOR UPDATE and transaction to prevent race conditions
func (s *Service) AddToEscrow(ctx context.Context, agentID string, amount float64) error {
	tx := s.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return err
	}
	defer tx.Rollback()

	// SECURITY FIX: Use SELECT FOR UPDATE to lock wallet row
	var wallet identity.AgentWallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("agent_id = ?", agentID).First(&wallet).Error; err != nil {
		return fmt.Errorf("failed to get wallet for escrow: %w", err)
	}

	if wallet.BalanceUSD < amount {
		return fmt.Errorf("insufficient funds for escrow: have %.2f, need %.2f", wallet.BalanceUSD, amount)
	}

	wallet.BalanceUSD -= amount
	wallet.EscrowBalanceUSD += amount
	wallet.UpdatedAt = time.Now()

	if err := tx.Save(&wallet).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}

// ReleaseFromEscrow releases funds from escrow
// SECURITY FIX: Uses SELECT FOR UPDATE and transaction to prevent race conditions
func (s *Service) ReleaseFromEscrow(ctx context.Context, agentID string, amount float64, toAgentID string) error {
	tx := s.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return err
	}
	defer tx.Rollback()

	// SECURITY FIX: Use SELECT FOR UPDATE to lock wallet row
	var wallet identity.AgentWallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("agent_id = ?", agentID).First(&wallet).Error; err != nil {
		return fmt.Errorf("failed to get wallet for escrow release: %w", err)
	}

	if wallet.EscrowBalanceUSD < amount {
		return fmt.Errorf("insufficient escrow balance: have %.2f, need %.2f", wallet.EscrowBalanceUSD, amount)
	}

	wallet.EscrowBalanceUSD -= amount
	wallet.UpdatedAt = time.Now()

	if err := tx.Save(&wallet).Error; err != nil {
		return err
	}

	// If toAgentID is specified, credit to that agent within the same transaction
	if toAgentID != "" {
		// Get or create the recipient wallet
		var toWallet identity.AgentWallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("agent_id = ?", toAgentID).First(&toWallet).Error; err == gorm.ErrRecordNotFound {
			toWallet = identity.AgentWallet{
				ID:               uuid.New(),
				AgentID:          toAgentID,
				BalanceUSD:       0,
				EscrowBalanceUSD:  0,
				TotalEarnedUSD:   0,
				TotalSpentUSD:    0,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			if err := tx.Create(&toWallet).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		toWallet.BalanceUSD += amount
		toWallet.TotalEarnedUSD += amount
		now := time.Now()
		toWallet.LastEarningAt = &now
		toWallet.UpdatedAt = now
		if err := tx.Save(&toWallet).Error; err != nil {
			return err
		}
	}

	return tx.Commit().Error
}
