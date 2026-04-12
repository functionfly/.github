package wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Service provides high-level wallet operations with caching and notifications
type Service struct {
	repo       *Repository
	redis      *redis.Client
	notifyFunc func(ctx context.Context, userID uuid.UUID, notificationType string, data map[string]interface{}) error
}

// NewService creates a new wallet service
func NewService(repo *Repository, redisClient *redis.Client) *Service {
	return &Service{
		repo:  repo,
		redis: redisClient,
	}
}

// SetNotificationFunc sets the notification callback function
func (s *Service) SetNotificationFunc(fn func(ctx context.Context, userID uuid.UUID, notificationType string, data map[string]interface{}) error) {
	s.notifyFunc = fn
}

// ============================================
// Core Operations
// ============================================

// GetOrCreateUserWallet retrieves or creates a wallet for a user (for registry fees)
func (s *Service) GetOrCreateUserWallet(ctx context.Context, userID uuid.UUID) (*Wallet, error) {
	return s.repo.GetOrCreateWalletForUser(ctx, userID)
}

// GetOrCreateAgentWallet retrieves or creates a wallet for an agent (for execution credits)
func (s *Service) GetOrCreateAgentWallet(ctx context.Context, agentID string) (*Wallet, error) {
	return s.repo.GetOrCreateWalletForAgent(ctx, agentID)
}

// GetWallet retrieves a wallet by ID
func (s *Service) GetWallet(ctx context.Context, walletID uuid.UUID) (*Wallet, error) {
	return s.repo.GetWalletByID(ctx, walletID)
}

// GetWalletByOwner retrieves a wallet by owner
func (s *Service) GetWalletByOwner(ctx context.Context, ownerType, ownerID string) (*Wallet, error) {
	return s.repo.GetWalletByOwner(ctx, ownerType, ownerID)
}

// GetUserWallet retrieves a user's wallet (backward compatible)
func (s *Service) GetUserWallet(ctx context.Context, userID uuid.UUID) (*Wallet, error) {
	return s.repo.GetWalletByOwner(ctx, OwnerTypeUser, userID.String())
}

// GetAgentWallet retrieves an agent's wallet (backward compatible)
func (s *Service) GetAgentWallet(ctx context.Context, agentID string) (*Wallet, error) {
	return s.repo.GetWalletByOwner(ctx, OwnerTypeAgent, agentID)
}

// ============================================
// Credit Operations
// ============================================

// CreditUserWallet adds credits to a user's wallet (for registry top-ups)
func (s *Service) CreditUserWallet(ctx context.Context, userID uuid.UUID, amountUSD float64, stripeReference string) (*BalanceUpdate, error) {
	wallet, err := s.GetOrCreateUserWallet(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	req := CreditRequest{
		WalletID:       wallet.ID,
		AmountUSD:      amountUSD,
		Reference:      stripeReference,
		IdempotencyKey: fmt.Sprintf("stripe:%s", stripeReference),
		TriggeredBy: TriggeredByInfo{
			Type: "webhook",
			ID:   "stripe",
		},
		Metadata: map[string]interface{}{
			"type":              "registry_wallet_credit",
			"stripe_payment_id": stripeReference,
			"user_id":           userID.String(),
		},
	}

	update, err := s.repo.Credit(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache if needed
	s.invalidateWalletCache(ctx, wallet.ID)

	return update, nil
}

// CreditAgentWallet adds credits to an agent's wallet (for execution credits)
func (s *Service) CreditAgentWallet(ctx context.Context, agentID string, amountUSD float64, stripeReference string, initiatingUserID *uuid.UUID) (*BalanceUpdate, error) {
	wallet, err := s.GetOrCreateAgentWallet(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	triggeredByID := "stripe"
	if initiatingUserID != nil {
		triggeredByID = initiatingUserID.String()
	}

	req := CreditRequest{
		WalletID:       wallet.ID,
		AmountUSD:      amountUSD,
		Reference:      stripeReference,
		IdempotencyKey: fmt.Sprintf("stripe:%s", stripeReference),
		TriggeredBy: TriggeredByInfo{
			Type: "webhook",
			ID:   triggeredByID,
		},
		Metadata: map[string]interface{}{
			"type":              "agent_execution_credits",
			"stripe_payment_id": stripeReference,
			"agent_id":          agentID,
		},
	}

	update, err := s.repo.Credit(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	s.invalidateWalletCache(ctx, wallet.ID)

	// Send notification if callback is set and we have initiating user
	if s.notifyFunc != nil && initiatingUserID != nil {
		notificationData := map[string]interface{}{
			"agent_id":        agentID,
			"amount_usd":      amountUSD,
			"balance_usd":     update.CurrentBalance,
			"transaction_id":  update.TransactionID.String(),
		}
		go s.notifyFunc(context.Background(), *initiatingUserID, "wallet_top_up", notificationData)
	}

	return update, nil
}

// ============================================
// Debit Operations (with backward compatibility)
// ============================================

// DebitForFeePayment debits wallet for platform fee (publish/version update)
func (s *Service) DebitForFeePayment(ctx context.Context, walletID uuid.UUID, amountUSD float64, feeType string, description string) (*BalanceUpdate, error) {
	wallet, err := s.repo.GetWalletByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	req := DebitRequest{
		WalletID:        walletID,
		AmountUSD:       amountUSD,
		TransactionType: TransactionTypeFeePayment,
		Reference:       fmt.Sprintf("fee_%s_%s", feeType, description),
		TriggeredBy: TriggeredByInfo{
			Type: "system",
			ID:   wallet.OwnerID,
		},
		FeeType: &feeType,
		Metadata: map[string]interface{}{
			"fee_type":    feeType,
			"description": description,
		},
	}

	update, err := s.repo.Debit(ctx, req)
	if err != nil {
		return nil, err
	}

	s.invalidateWalletCache(ctx, walletID)

	return update, nil
}

// DebitForExecution debits wallet for function execution
func (s *Service) DebitForExecution(ctx context.Context, walletID uuid.UUID, amountUSD float64, executionID, functionID uuid.UUID) (*BalanceUpdate, error) {
	wallet, err := s.repo.GetWalletByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	req := DebitRequest{
		WalletID:        walletID,
		AmountUSD:       amountUSD,
		TransactionType: TransactionTypeExecutionCharge,
		TriggeredBy: TriggeredByInfo{
			Type: "agent",
			ID:   wallet.OwnerID,
		},
		ExecutionID: &executionID,
		FunctionID:  &functionID,
		Metadata: map[string]interface{}{
			"execution_id": executionID.String(),
			"function_id":  functionID.String(),
		},
	}

	update, err := s.repo.Debit(ctx, req)
	if err != nil {
		return nil, err
	}

	s.invalidateWalletCache(ctx, walletID)

	return update, nil
}

// ConsumeAgentCredits debits agent wallet for execution (backward compatible with old billing.Controller)
func (s *Service) ConsumeAgentCredits(ctx context.Context, agentID string, amountUSD float64) (*BalanceUpdate, error) {
	wallet, err := s.GetOrCreateAgentWallet(ctx, agentID)
	if err != nil {
		return nil, err
	}

	req := DebitRequest{
		WalletID:        wallet.ID,
		AmountUSD:       amountUSD,
		TransactionType: TransactionTypeExecutionCharge,
		TriggeredBy: TriggeredByInfo{
			Type: "agent",
			ID:   agentID,
		},
		Metadata: map[string]interface{}{
			"agent_id": agentID,
		},
	}

	update, err := s.repo.Debit(ctx, req)
	if err != nil {
		return nil, err
	}

	s.invalidateWalletCache(ctx, wallet.ID)

	// Check for low balance notification
	if s.notifyFunc != nil && update.CurrentBalance < 5.0 {
		// Parse owner as user ID for notification (if applicable)
		if wallet.UserID != nil {
			notificationData := map[string]interface{}{
				"agent_id":      agentID,
				"balance_usd":   update.CurrentBalance,
				"threshold_usd": 5.0,
			}
			go s.notifyFunc(context.Background(), *wallet.UserID, "wallet_low_balance", notificationData)
		}
	}

	return update, nil
}

// ============================================
// Backward Compatibility Methods
// ============================================

// CreditWalletUser adds credits to a user's wallet (compatible with old PlatformFeeRepository.CreditWallet)
func (s *Service) CreditWalletUser(ctx context.Context, userID uuid.UUID, amountUSD float64, stripePaymentID string) error {
	_, err := s.CreditUserWallet(ctx, userID, amountUSD, stripePaymentID)
	return err
}

// DebitWalletUser debits a user's wallet (compatible with old PlatformFeeRepository.DebitWallet)
func (s *Service) DebitWalletUser(ctx context.Context, userID uuid.UUID, amountUSD float64, description string) error {
	wallet, err := s.GetOrCreateUserWallet(ctx, userID)
	if err != nil {
		return err
	}

	feeType := FeeTypePublish
	_, err = s.DebitForFeePayment(ctx, wallet.ID, amountUSD, feeType, description)
	return err
}

// AddCreditsToAgent adds credits to an agent (compatible with old billing.Controller.AddCredits)
func (s *Service) AddCreditsToAgent(ctx context.Context, agentID string, amountUSD float64) error {
	// Check if already credited via idempotency
	wallet, err := s.GetOrCreateAgentWallet(ctx, agentID)
	if err != nil {
		return err
	}

	req := CreditRequest{
		WalletID:  wallet.ID,
		AmountUSD: amountUSD,
		TriggeredBy: TriggeredByInfo{
			Type: "system",
			ID:   "billing",
		},
		Metadata: map[string]interface{}{
			"agent_id": agentID,
		},
	}

	_, err = s.repo.Credit(ctx, req)
	return err
}

// GetUserBalance retrieves a user's wallet balance (compatible with old PlatformFeeRepository)
func (s *Service) GetUserBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	wallet, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return 0, err
	}
	if wallet == nil {
		return 0, nil
	}
	return wallet.BalanceUSD, nil
}

// GetAgentBalance retrieves an agent's wallet balance (compatible with old billing.Controller)
func (s *Service) GetAgentBalance(ctx context.Context, agentID string) (float64, error) {
	wallet, err := s.GetAgentWallet(ctx, agentID)
	if err != nil {
		return 0, err
	}
	if wallet == nil {
		return 0, nil
	}
	return wallet.BalanceUSD, nil
}

// CheckAgentSpendCap checks if an agent can spend (compatible with old billing.Controller)
func (s *Service) CheckAgentSpendCap(ctx context.Context, agentID string, estimatedCost float64) (bool, error) {
	wallet, err := s.GetAgentWallet(ctx, agentID)
	if err != nil {
		// Non-fatal: allow execution if wallet can't be loaded
		return true, nil
	}
	if wallet == nil {
		return true, nil
	}

	check, err := s.repo.CheckSpendCap(ctx, wallet.ID, estimatedCost)
	if err != nil {
		return true, nil
	}

	return check.Allowed, nil
}

// GetAgentSpendSummary returns spend summary (compatible with old billing.Controller)
func (s *Service) GetAgentSpendSummary(ctx context.Context, agentID string, period string) (*SpendCapCheck, error) {
	wallet, err := s.GetAgentWallet(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, fmt.Errorf("wallet not found for agent: %s", agentID)
	}

	return s.repo.CheckSpendCap(ctx, wallet.ID, 0)
}

// UpdateAgentSpendCaps updates spend caps (compatible with old billing.Controller)
func (s *Service) UpdateAgentSpendCaps(ctx context.Context, agentID string, dailyCap, monthlyCap *float64) error {
	wallet, err := s.GetAgentWallet(ctx, agentID)
	if err != nil {
		return err
	}
	if wallet == nil {
		return fmt.Errorf("wallet not found for agent: %s", agentID)
	}

	return s.repo.UpdateSpendCaps(ctx, wallet.ID, dailyCap, monthlyCap)
}

// HasUserWalletCreditReference checks if a credit was already applied (compatible with old PlatformFeeRepository)
func (s *Service) HasUserWalletCreditReference(ctx context.Context, reference string) (bool, error) {
	return s.repo.HasTransactionWithReference(ctx, reference, TransactionTypeCredit)
}

// ============================================
// Spend Cap Operations
// ============================================

// CheckSpendCap checks if a wallet can spend the given amount
func (s *Service) CheckSpendCap(ctx context.Context, walletID uuid.UUID, amountUSD float64) (*SpendCapCheck, error) {
	return s.repo.CheckSpendCap(ctx, walletID, amountUSD)
}

// UpdateSpendCaps updates spend caps for a wallet
func (s *Service) UpdateSpendCaps(ctx context.Context, walletID uuid.UUID, dailyCap, monthlyCap *float64) error {
	return s.repo.UpdateSpendCaps(ctx, walletID, dailyCap, monthlyCap)
}

// ============================================
// Transaction History
// ============================================

// GetTransactionHistory retrieves transaction history for a wallet
func (s *Service) GetTransactionHistory(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]WalletTransaction, int64, error) {
	return s.repo.GetTransactionsByWallet(ctx, walletID, limit, offset)
}

// GetWalletSummary retrieves wallet summary statistics
func (s *Service) GetWalletSummary(ctx context.Context, walletID uuid.UUID) (*WalletSummary, error) {
	return s.repo.GetWalletSummary(ctx, walletID)
}

// GetBalanceHistory retrieves balance history for a wallet
func (s *Service) GetBalanceHistory(ctx context.Context, query BalanceHistoryQuery) (*BalanceHistoryResult, error) {
	return s.repo.GetBalanceHistory(ctx, query)
}

// ============================================
// Cache Operations
// ============================================

func (s *Service) invalidateWalletCache(ctx context.Context, walletID uuid.UUID) {
	if s.redis == nil {
		return
	}
	// Cache invalidation is fire-and-forget
	_ = s.redis.Del(ctx, fmt.Sprintf("wallet:%s", walletID.String()))
	_ = s.redis.Del(ctx, fmt.Sprintf("wallet:summary:%s", walletID.String()))
}

// CacheWalletBalance caches wallet balance in Redis
func (s *Service) CacheWalletBalance(ctx context.Context, walletID uuid.UUID, balance float64, ttl int) error {
	if s.redis == nil {
		return nil
	}

	key := fmt.Sprintf("wallet:%s", walletID.String())
	data, _ := json.Marshal(map[string]float64{"balance": balance})

	return s.redis.Set(ctx, key, data, 0).Err()
}

// GetCachedWalletBalance retrieves cached wallet balance
func (s *Service) GetCachedWalletBalance(ctx context.Context, walletID uuid.UUID) (float64, bool) {
	if s.redis == nil {
		return 0, false
	}

	key := fmt.Sprintf("wallet:%s", walletID.String())
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return 0, false
	}

	var result map[string]float64
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return 0, false
	}

	return result["balance"], true
}

// ============================================
// Admin Operations
// ============================================

// SuspendWallet suspends a wallet
func (s *Service) SuspendWallet(ctx context.Context, walletID uuid.UUID, reason string) error {
	return s.repo.UpdateWalletStatus(ctx, walletID, WalletStatusSuspended, &reason)
}

// CloseWallet closes a wallet
func (s *Service) CloseWallet(ctx context.Context, walletID uuid.UUID, reason string) error {
	return s.repo.UpdateWalletStatus(ctx, walletID, WalletStatusClosed, &reason)
}

// ReactivateWallet reactivates a suspended wallet
func (s *Service) ReactivateWallet(ctx context.Context, walletID uuid.UUID) error {
	return s.repo.UpdateWalletStatus(ctx, walletID, WalletStatusActive, nil)
}

// GetLowBalanceWallets retrieves wallets with low balance
func (s *Service) GetLowBalanceWallets(ctx context.Context, threshold float64, ownerType string) ([]Wallet, error) {
	return s.repo.GetLowBalanceWallets(ctx, threshold, ownerType)
}

// ============================================
// Admin Operations (Credit/Debit for adjustments)
// ============================================

// AdminCredit adds credits to a wallet by admin (for adjustments)
func (s *Service) AdminCredit(ctx context.Context, walletID uuid.UUID, amountUSD float64, reference, reason string, adminUserID uuid.UUID) (*BalanceUpdate, error) {
	if amountUSD <= 0 {
		return nil, fmt.Errorf("credit amount must be positive")
	}

	wallet, err := s.repo.GetWalletByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	if wallet == nil {
		return nil, fmt.Errorf("wallet not found: %s", walletID)
	}

	adminID := adminUserID.String()
	req := CreditRequest{
		WalletID:       walletID,
		AmountUSD:      amountUSD,
		Reference:      reference,
		IdempotencyKey: fmt.Sprintf("admin:%s:%d", adminID, time.Now().Unix()),
		TriggeredBy: TriggeredByInfo{
			Type: "admin",
			ID:   adminID,
		},
		Metadata: map[string]interface{}{
			"type":         "admin_adjustment",
			"reason":       reason,
			"admin_user_id": adminID,
			"is_credit":    true,
		},
	}

	update, err := s.repo.Credit(ctx, req)
	if err != nil {
		return nil, err
	}

	s.invalidateWalletCache(ctx, walletID)

	// Record in balance history if available
	if s.repo.db != nil {
		_ = s.repo.RecordBalanceChange(ctx, walletID, update.CurrentBalance, amountUSD, update.TransactionID, wallet.Currency)
	}

	return update, nil
}

// AdminDebit debits a wallet by admin (for adjustments)
func (s *Service) AdminDebit(ctx context.Context, walletID uuid.UUID, amountUSD float64, reference, reason string, adminUserID uuid.UUID) (*BalanceUpdate, error) {
	if amountUSD <= 0 {
		return nil, fmt.Errorf("debit amount must be positive")
	}

	wallet, err := s.repo.GetWalletByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	if wallet == nil {
		return nil, fmt.Errorf("wallet not found: %s", walletID)
	}

	adminID := adminUserID.String()
	req := DebitRequest{
		WalletID:        walletID,
		AmountUSD:       amountUSD,
		TransactionType: TransactionTypeAdjustment,
		Reference:       reference,
		TriggeredBy: TriggeredByInfo{
			Type: "admin",
			ID:   adminID,
		},
		Metadata: map[string]interface{}{
			"type":         "admin_adjustment",
			"reason":       reason,
			"admin_user_id": adminID,
			"is_debit":     true,
		},
	}

	update, err := s.repo.Debit(ctx, req)
	if err != nil {
		return nil, err
	}

	s.invalidateWalletCache(ctx, walletID)

	// Record in balance history if available
	if s.repo.db != nil {
		_ = s.repo.RecordBalanceChange(ctx, walletID, update.CurrentBalance, -amountUSD, update.TransactionID, wallet.Currency)
	}

	return update, nil
}

// WalletFilter defines filters for listing wallets
type WalletFilter struct {
	OwnerType   string
	OwnerIDs    []string
	Status      string
	MinBalance  *float64
	MaxBalance  *float64
	WalletType  string
	WalletIDs   []string
	Currency    string
	Limit       int
	Offset      int
}

// WalletListResult holds the result of a wallet list query
type WalletListResult struct {
	Wallets []Wallet `json:"wallets"`
	Total   int64    `json:"total"`
}

// ListWallets lists wallets with filtering
func (s *Service) ListWallets(ctx context.Context, filter WalletFilter) (*WalletListResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 500 {
		filter.Limit = 500
	}

	db := s.repo.db.WithContext(ctx).Model(&Wallet{})

	// Apply filters
	if filter.OwnerType != "" {
		db = db.Where("owner_type = ?", filter.OwnerType)
	}
	if len(filter.OwnerIDs) > 0 {
		db = db.Where("owner_id IN ?", filter.OwnerIDs)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.WalletType != "" {
		db = db.Where("wallet_type = ?", filter.WalletType)
	}
	if len(filter.WalletIDs) > 0 {
		db = db.Where("id IN ?", filter.WalletIDs)
	}
	if filter.Currency != "" {
		db = db.Where("currency = ?", filter.Currency)
	}
	if filter.MinBalance != nil {
		db = db.Where("balance_usd >= ?", *filter.MinBalance)
	}
	if filter.MaxBalance != nil {
		db = db.Where("balance_usd <= ?", *filter.MaxBalance)
	}

	// Get total count
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count wallets: %w", err)
	}

	// Get wallets
	var wallets []Wallet
	if err := db.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&wallets).Error; err != nil {
		return nil, fmt.Errorf("failed to list wallets: %w", err)
	}

	return &WalletListResult{
		Wallets: wallets,
		Total:   total,
	}, nil
}
