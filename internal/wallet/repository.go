package wallet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository provides unified wallet operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new wallet repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ============================================
// Wallet CRUD Operations
// ============================================

// CreateWallet creates a new wallet for a user or agent
func (r *Repository) CreateWallet(ctx context.Context, req WalletCreationRequest) (*Wallet, error) {
	wallet := &Wallet{
		ID:                 uuid.New(),
		OwnerType:          req.OwnerType,
		OwnerID:            req.OwnerID,
		UserID:             req.UserID,
		AgentID:            req.AgentID,
		WalletType:         req.WalletType,
		BalanceUSD:         req.InitialBalanceUSD,
		SpendCapMonthlyUSD: req.SpendCapMonthlyUSD,
		SpendCapDailyUSD:   req.SpendCapDailyUSD,
		BillingMode:        req.BillingMode,
		TeamID:             req.TeamID,
		Status:             WalletStatusActive,
		AlertThresholds:    []float64{0.5, 0.8, 0.95},
	}

	if wallet.BillingMode == "" {
		wallet.BillingMode = BillingModePerWallet
	}
	if wallet.WalletType == "" {
		wallet.WalletType = WalletTypeUnified
	}

	if err := r.db.WithContext(ctx).Create(wallet).Error; err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return wallet, nil
}

// GetWalletByID retrieves a wallet by its ID
func (r *Repository) GetWalletByID(ctx context.Context, walletID uuid.UUID) (*Wallet, error) {
	var wallet Wallet
	if err := r.db.WithContext(ctx).First(&wallet, "id = ?", walletID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	return &wallet, nil
}

// GetWalletByOwner retrieves a wallet by owner type and ID
func (r *Repository) GetWalletByOwner(ctx context.Context, ownerType, ownerID string) (*Wallet, error) {
	var wallet Wallet
	if err := r.db.WithContext(ctx).First(&wallet, "owner_type = ? AND owner_id = ?", ownerType, ownerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	return &wallet, nil
}

// GetOrCreateWalletForUser retrieves or creates a wallet for a user
func (r *Repository) GetOrCreateWalletForUser(ctx context.Context, userID uuid.UUID) (*Wallet, error) {
	ownerID := userID.String()
	wallet, err := r.GetWalletByOwner(ctx, OwnerTypeUser, ownerID)
	if err != nil {
		return nil, err
	}
	if wallet != nil {
		return wallet, nil
	}

	// Create new wallet
	return r.CreateWallet(ctx, WalletCreationRequest{
		OwnerType:  OwnerTypeUser,
		OwnerID:    ownerID,
		UserID:     &userID,
		WalletType: WalletTypeRegistry,
	})
}

// GetOrCreateWalletForAgent retrieves or creates a wallet for an agent
func (r *Repository) GetOrCreateWalletForAgent(ctx context.Context, agentID string) (*Wallet, error) {
	wallet, err := r.GetWalletByOwner(ctx, OwnerTypeAgent, agentID)
	if err != nil {
		return nil, err
	}
	if wallet != nil {
		return wallet, nil
	}

	// Create new wallet
	return r.CreateWallet(ctx, WalletCreationRequest{
		OwnerType:  OwnerTypeAgent,
		OwnerID:    agentID,
		AgentID:    &agentID,
		WalletType: WalletTypeExecution,
	})
}

// GetBalance retrieves the current balance for a wallet
func (r *Repository) GetBalance(ctx context.Context, walletID uuid.UUID) (float64, error) {
	wallet, err := r.GetWalletByID(ctx, walletID)
	if err != nil {
		return 0, err
	}
	if wallet == nil {
		return 0, nil
	}
	return wallet.BalanceUSD, nil
}

// UpdateWalletStatus updates the wallet status
func (r *Repository) UpdateWalletStatus(ctx context.Context, walletID uuid.UUID, status string, reason *string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == WalletStatusClosed {
		now := time.Now()
		updates["closed_at"] = now
		if reason != nil {
			updates["closure_reason"] = *reason
		}
	}

	if err := r.db.WithContext(ctx).Model(&Wallet{}).Where("id = ?", walletID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update wallet status: %w", err)
	}

	return nil
}

// UpdateSpendCaps updates the spend caps for a wallet
func (r *Repository) UpdateSpendCaps(ctx context.Context, walletID uuid.UUID, dailyCap, monthlyCap *float64) error {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if dailyCap != nil {
		updates["spend_cap_daily_usd"] = *dailyCap
	}
	if monthlyCap != nil {
		updates["spend_cap_monthly_usd"] = *monthlyCap
	}

	result := r.db.WithContext(ctx).Model(&Wallet{}).Where("id = ?", walletID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update spend caps: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("wallet not found: %s", walletID)
	}

	return nil
}

// ============================================
// Credit/Debit Operations
// ============================================

// Credit adds funds to a wallet
func (r *Repository) Credit(ctx context.Context, req CreditRequest) (*BalanceUpdate, error) {
	if req.AmountUSD <= 0 {
		return nil, fmt.Errorf("credit amount must be positive")
	}

	var update BalanceUpdate
	update.WalletID = req.WalletID
	update.Amount = req.AmountUSD

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock wallet for update
		var wallet Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&wallet, "id = ?", req.WalletID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("wallet not found: %s", req.WalletID)
			}
			return fmt.Errorf("failed to lock wallet: %w", err)
		}

		if wallet.Status != WalletStatusActive {
			return fmt.Errorf("wallet is not active: %s", wallet.Status)
		}

		// Store previous balance
		update.PreviousBalance = wallet.BalanceUSD

		// Update wallet balance
		wallet.BalanceUSD += req.AmountUSD
		wallet.LifetimeEarningsUSD += req.AmountUSD

		if err := tx.Save(&wallet).Error; err != nil {
			return fmt.Errorf("failed to update wallet balance: %w", err)
		}

		update.CurrentBalance = wallet.BalanceUSD

		// Create transaction record
		triggeredByType := req.TriggeredBy.Type
		triggeredByID := req.TriggeredBy.ID
		metadata, _ := json.Marshal(req.Metadata)

		txRecord := WalletTransaction{
			ID:               uuid.New(),
			WalletID:         req.WalletID,
			TransactionType:  TransactionTypeCredit,
			AmountUSD:        req.AmountUSD,
			BalanceBeforeUSD: update.PreviousBalance,
			BalanceAfterUSD:  update.CurrentBalance,
			Status:           TransactionStatusCompleted,
			Reference:        &req.Reference,
			TriggeredByType:  &triggeredByType,
			TriggeredByID:    &triggeredByID,
			Metadata:         metadata,
			IdempotencyKey:   &req.IdempotencyKey,
			CompletedAt:      ptr(time.Now()),
		}

		if err := tx.Create(&txRecord).Error; err != nil {
			return fmt.Errorf("failed to create transaction record: %w", err)
		}

		update.TransactionID = txRecord.ID

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &update, nil
}

// Debit removes funds from a wallet
func (r *Repository) Debit(ctx context.Context, req DebitRequest) (*BalanceUpdate, error) {
	if req.AmountUSD <= 0 {
		return nil, fmt.Errorf("debit amount must be positive")
	}

	var update BalanceUpdate
	update.WalletID = req.WalletID
	update.Amount = req.AmountUSD

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock wallet for update
		var wallet Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&wallet, "id = ?", req.WalletID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("wallet not found: %s", req.WalletID)
			}
			return fmt.Errorf("failed to lock wallet: %w", err)
		}

		if wallet.Status != WalletStatusActive {
			return fmt.Errorf("wallet is not active: %s", wallet.Status)
		}

		// Check sufficient balance
		if wallet.BalanceUSD < req.AmountUSD {
			return fmt.Errorf("insufficient balance: have %.4f, need %.4f", wallet.BalanceUSD, req.AmountUSD)
		}

		// Store previous balance
		update.PreviousBalance = wallet.BalanceUSD

		// Update wallet balance
		wallet.BalanceUSD -= req.AmountUSD
		wallet.LifetimeSpentUSD += req.AmountUSD

		if err := tx.Save(&wallet).Error; err != nil {
			return fmt.Errorf("failed to update wallet balance: %w", err)
		}

		update.CurrentBalance = wallet.BalanceUSD

		// Create transaction record
		triggeredByType := req.TriggeredBy.Type
		triggeredByID := req.TriggeredBy.ID
		metadata, _ := json.Marshal(req.Metadata)

		txRecord := WalletTransaction{
			ID:               uuid.New(),
			WalletID:         req.WalletID,
			TransactionType:  req.TransactionType,
			AmountUSD:        req.AmountUSD,
			BalanceBeforeUSD: update.PreviousBalance,
			BalanceAfterUSD:  update.CurrentBalance,
			Status:           TransactionStatusCompleted,
			Reference:        &req.Reference,
			TriggeredByType:  &triggeredByType,
			TriggeredByID:    &triggeredByID,
			ExecutionID:      req.ExecutionID,
			FunctionID:       req.FunctionID,
			FeeType:          req.FeeType,
			Metadata:         metadata,
			CompletedAt:      ptr(time.Now()),
		}

		if err := tx.Create(&txRecord).Error; err != nil {
			return fmt.Errorf("failed to create transaction record: %w", err)
		}

		update.TransactionID = txRecord.ID

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &update, nil
}

// ============================================
// Transaction Queries
// ============================================

// GetTransactionByID retrieves a transaction by ID
func (r *Repository) GetTransactionByID(ctx context.Context, transactionID uuid.UUID) (*WalletTransaction, error) {
	var tx WalletTransaction
	if err := r.db.WithContext(ctx).First(&tx, "id = ?", transactionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	return &tx, nil
}

// GetTransactionsByWallet retrieves transactions for a wallet
func (r *Repository) GetTransactionsByWallet(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]WalletTransaction, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&WalletTransaction{}).
		Where("wallet_id = ?", walletID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	var transactions []WalletTransaction
	if err := r.db.WithContext(ctx).
		Where("wallet_id = ?", walletID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	return transactions, total, nil
}

// GetTransactionsByType retrieves transactions by type
func (r *Repository) GetTransactionsByType(ctx context.Context, walletID uuid.UUID, txType string, limit int) ([]WalletTransaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	var transactions []WalletTransaction
	if err := r.db.WithContext(ctx).
		Where("wallet_id = ? AND transaction_type = ?", walletID, txType).
		Order("created_at DESC").
		Limit(limit).
		Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	return transactions, nil
}

// HasTransactionWithReference checks if a transaction with the given reference already exists
func (r *Repository) HasTransactionWithReference(ctx context.Context, reference string, txType string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&WalletTransaction{}).
		Where("reference = ? AND transaction_type = ?", reference, txType).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check transaction reference: %w", err)
	}
	return count > 0, nil
}

// HasTransactionWithIdempotencyKey checks if a transaction with the given idempotency key already exists
func (r *Repository) HasTransactionWithIdempotencyKey(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, nil
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&WalletTransaction{}).
		Where("idempotency_key = ?", key).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check idempotency key: %w", err)
	}
	return count > 0, nil
}

// ============================================
// Summary and Analytics
// ============================================

// GetWalletSummary retrieves aggregated wallet statistics
func (r *Repository) GetWalletSummary(ctx context.Context, walletID uuid.UUID) (*WalletSummary, error) {
	var summary WalletSummary

	// Get wallet base info
	wallet, err := r.GetWalletByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, nil
	}

	summary.WalletID = wallet.ID
	summary.OwnerType = wallet.OwnerType
	summary.OwnerID = wallet.OwnerID
	summary.UserID = wallet.UserID
	summary.AgentID = wallet.AgentID
	summary.BalanceUSD = wallet.BalanceUSD
	summary.Status = wallet.Status
	summary.CreatedAt = wallet.CreatedAt
	summary.UpdatedAt = wallet.UpdatedAt

	// Calculate aggregates
	type aggregates struct {
		TotalCredits      float64 `gorm:"column:total_credits"`
		TotalDebits       float64 `gorm:"column:total_debits"`
		TotalFees         float64 `gorm:"column:total_fees"`
		TotalExecution    float64 `gorm:"column:total_execution"`
		TotalCommissions  float64 `gorm:"column:total_commissions"`
		TotalTxCount      int64   `gorm:"column:total_tx_count"`
		PendingTxCount    int64   `gorm:"column:pending_tx_count"`
		LastTransactionAt *time.Time
	}

	var agg aggregates
	err = r.db.WithContext(ctx).Model(&WalletTransaction{}).
		Select(`
			COALESCE(SUM(CASE WHEN transaction_type = 'credit' AND status = 'completed' THEN amount_usd ELSE 0 END), 0) as total_credits,
			COALESCE(SUM(CASE WHEN transaction_type = 'debit' AND status = 'completed' THEN amount_usd ELSE 0 END), 0) as total_debits,
			COALESCE(SUM(CASE WHEN transaction_type = 'fee_payment' AND status = 'completed' THEN amount_usd ELSE 0 END), 0) as total_fees,
			COALESCE(SUM(CASE WHEN transaction_type = 'execution_charge' AND status = 'completed' THEN amount_usd ELSE 0 END), 0) as total_execution,
			COALESCE(SUM(CASE WHEN transaction_type = 'commission' AND status = 'completed' THEN amount_usd ELSE 0 END), 0) as total_commissions,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as total_tx_count,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_tx_count,
			MAX(CASE WHEN status = 'completed' THEN created_at END) as last_transaction_at
		`).
		Where("wallet_id = ?", walletID).
		Scan(&agg).Error

	if err != nil {
		return nil, fmt.Errorf("failed to calculate wallet summary: %w", err)
	}

	summary.TotalCreditsUSD = agg.TotalCredits
	summary.TotalDebitsUSD = agg.TotalDebits
	summary.TotalFeesPaidUSD = agg.TotalFees
	summary.TotalExecutionChargesUSD = agg.TotalExecution
	summary.TotalCommissionsUSD = agg.TotalCommissions
	summary.TotalTransactions = agg.TotalTxCount
	summary.PendingTransactions = agg.PendingTxCount
	summary.LastTransactionAt = agg.LastTransactionAt

	return &summary, nil
}

// ============================================
// Spend Cap Operations
// ============================================

// CheckSpendCap validates if a wallet can spend the estimated amount
func (r *Repository) CheckSpendCap(ctx context.Context, walletID uuid.UUID, estimatedCost float64) (*SpendCapCheck, error) {
	wallet, err := r.GetWalletByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return &SpendCapCheck{Allowed: false, Reason: "wallet not found"}, nil
	}

	result := &SpendCapCheck{
		Allowed: true,
	}

	// Check daily cap
	if wallet.SpendCapDailyUSD != nil && *wallet.SpendCapDailyUSD > 0 {
		dailySpend, err := r.getDailySpend(ctx, walletID)
		if err == nil {
			result.DailySpendUSD = dailySpend
			result.DailyCapUSD = *wallet.SpendCapDailyUSD

			if dailySpend+estimatedCost > *wallet.SpendCapDailyUSD {
				result.Allowed = false
				result.Reason = fmt.Sprintf("daily spend cap of $%.2f would be exceeded (current: $%.4f, estimated: $%.6f)",
					*wallet.SpendCapDailyUSD, dailySpend, estimatedCost)
				return result, nil
			}
		}
	}

	// Check monthly cap
	if wallet.SpendCapMonthlyUSD != nil && *wallet.SpendCapMonthlyUSD > 0 {
		monthlySpend, err := r.getMonthlySpend(ctx, walletID)
		if err == nil {
			result.MonthlySpendUSD = monthlySpend
			result.MonthlyCapUSD = *wallet.SpendCapMonthlyUSD

			if monthlySpend+estimatedCost > *wallet.SpendCapMonthlyUSD {
				result.Allowed = false
				result.Reason = fmt.Sprintf("monthly spend cap of $%.2f would be exceeded (current: $%.4f, estimated: $%.6f)",
					*wallet.SpendCapMonthlyUSD, monthlySpend, estimatedCost)
				return result, nil
			}
		}
	}

	return result, nil
}

// getDailySpend calculates the total spend for today
func (r *Repository) getDailySpend(ctx context.Context, walletID uuid.UUID) (float64, error) {
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	var totalSpend float64
	err := r.db.WithContext(ctx).Model(&WalletTransaction{}).
		Where("wallet_id = ? AND created_at >= ? AND status = ? AND transaction_type IN ?",
			walletID, todayStart, TransactionStatusCompleted,
			[]string{TransactionTypeDebit, TransactionTypeFeePayment, TransactionTypeExecutionCharge, TransactionTypeCommission}).
		Select("COALESCE(SUM(amount_usd), 0)").
		Scan(&totalSpend).Error

	return totalSpend, err
}

// getMonthlySpend calculates the total spend for this month
func (r *Repository) getMonthlySpend(ctx context.Context, walletID uuid.UUID) (float64, error) {
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	var totalSpend float64
	err := r.db.WithContext(ctx).Model(&WalletTransaction{}).
		Where("wallet_id = ? AND created_at >= ? AND status = ? AND transaction_type IN ?",
			walletID, monthStart, TransactionStatusCompleted,
			[]string{TransactionTypeDebit, TransactionTypeFeePayment, TransactionTypeExecutionCharge, TransactionTypeCommission}).
		Select("COALESCE(SUM(amount_usd), 0)").
		Scan(&totalSpend).Error

	return totalSpend, err
}

// ============================================
// Bulk Operations
// ============================================

// ListWalletsByOwnerType retrieves all wallets of a specific owner type
func (r *Repository) ListWalletsByOwnerType(ctx context.Context, ownerType string, limit, offset int) ([]Wallet, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&Wallet{}).
		Where("owner_type = ?", ownerType).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count wallets: %w", err)
	}

	var wallets []Wallet
	if err := r.db.WithContext(ctx).
		Where("owner_type = ?", ownerType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&wallets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list wallets: %w", err)
	}

	return wallets, total, nil
}

// GetLowBalanceWallets retrieves wallets with balance below threshold
func (r *Repository) GetLowBalanceWallets(ctx context.Context, threshold float64, ownerType string) ([]Wallet, error) {
	var wallets []Wallet
	query := r.db.WithContext(ctx).
		Where("balance_usd < ? AND status = ?", threshold, WalletStatusActive)

	if ownerType != "" {
		query = query.Where("owner_type = ?", ownerType)
	}

	if err := query.Find(&wallets).Error; err != nil {
		return nil, fmt.Errorf("failed to get low balance wallets: %w", err)
	}

	return wallets, nil
}

// ============================================
// Utility Functions
// ============================================

func ptr(t time.Time) *time.Time {
	return &t
}
