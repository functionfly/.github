package economy

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
func (s *Service) Credit(ctx context.Context, agentID string, amount float64, transactionType string, metadata map[string]any) (*identity.RevenueTransaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("credit amount must be positive")
	}

	wallet, err := s.GetOrCreateWallet(ctx, agentID)
	if err != nil {
		return nil, err
	}

	tx := s.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Update wallet
	wallet.BalanceUSD += amount
	wallet.TotalEarnedUSD += amount
	now := time.Now()
	wallet.LastEarningAt = &now
	wallet.UpdatedAt = now

	if err := tx.Save(wallet).Error; err != nil {
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
func (s *Service) Debit(ctx context.Context, agentID string, amount float64, transactionType string, metadata map[string]any) (*identity.RevenueTransaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("debit amount must be positive")
	}

	wallet, err := s.GetWallet(ctx, agentID)
	if err != nil {
		return nil, err
	}

	if wallet.BalanceUSD < amount {
		return nil, fmt.Errorf("insufficient funds: have %.2f, need %.2f", wallet.BalanceUSD, amount)
	}

	tx := s.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Update wallet
	wallet.BalanceUSD -= amount
	wallet.TotalSpentUSD += amount
	now := time.Now()
	wallet.LastSpendingAt = &now
	wallet.UpdatedAt = now

	if err := tx.Save(wallet).Error; err != nil {
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
func (s *Service) Transfer(ctx context.Context, fromAgentID, toAgentID string, amount float64, transactionType string, metadata map[string]any) (*identity.RevenueTransaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("transfer amount must be positive")
	}

	fromWallet, err := s.GetWallet(ctx, fromAgentID)
	if err != nil {
		return nil, err
	}

	if fromWallet.BalanceUSD < amount {
		return nil, fmt.Errorf("insufficient funds for transfer: have %.2f, need %.2f", fromWallet.BalanceUSD, amount)
	}

	toWallet, err := s.GetOrCreateWallet(ctx, toAgentID)
	if err != nil {
		return nil, err
	}

	tx := s.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	defer tx.Rollback()

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
func (s *Service) AddToEscrow(ctx context.Context, agentID string, amount float64) error {
	wallet, err := s.GetWallet(ctx, agentID)
	if err != nil {
		return err
	}

	if wallet.BalanceUSD < amount {
		return fmt.Errorf("insufficient funds for escrow: have %.2f, need %.2f", wallet.BalanceUSD, amount)
	}

	wallet.BalanceUSD -= amount
	wallet.EscrowBalanceUSD += amount
	wallet.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Save(wallet).Error
}

// ReleaseFromEscrow releases funds from escrow
func (s *Service) ReleaseFromEscrow(ctx context.Context, agentID string, amount float64, toAgentID string) error {
	wallet, err := s.GetWallet(ctx, agentID)
	if err != nil {
		return err
	}

	if wallet.EscrowBalanceUSD < amount {
		return fmt.Errorf("insufficient escrow balance: have %.2f, need %.2f", wallet.EscrowBalanceUSD, amount)
	}

	wallet.EscrowBalanceUSD -= amount
	wallet.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Save(wallet).Error; err != nil {
		return err
	}

	// If toAgentID is specified, transfer to that agent
	if toAgentID != "" {
		_, err = s.Credit(ctx, toAgentID, amount, identity.TransactionTypeDelegationPayment, nil)
		return err
	}

	return nil
}
