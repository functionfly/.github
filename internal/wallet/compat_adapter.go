package wallet

import (
	"context"
	"fmt"

	"github.com/functionfly/functionfly/internal/agent/billing"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// CompatAdapter provides backward compatibility by implementing old interfaces
// while using the new unified wallet system internally.
// This allows gradual migration without breaking existing code.
type CompatAdapter struct {
	service         *Service
	legacyUserRepo  storageregistry.PlatformFeeRepository
	legacyAgentRepo *billing.Controller
}

// NewCompatAdapter creates a new compatibility adapter
func NewCompatAdapter(service *Service) *CompatAdapter {
	return &CompatAdapter{
		service: service,
	}
}

// SetLegacyRepos sets the legacy repositories for fallback
func (a *CompatAdapter) SetLegacyRepos(userRepo storageregistry.PlatformFeeRepository, agentRepo *billing.Controller) {
	a.legacyUserRepo = userRepo
	a.legacyAgentRepo = agentRepo
}

// ============================================
// User Wallet Compatibility (storageregistry.PlatformFeeRepository interface)
// ============================================

// PlatformFeeRepositoryWrapper wraps the new wallet service to match the old PlatformFeeRepository interface
type PlatformFeeRepositoryWrapper struct {
	service *Service
}

// NewPlatformFeeRepositoryWrapper creates a wrapper for the old interface
func NewPlatformFeeRepositoryWrapper(service *Service) *PlatformFeeRepositoryWrapper {
	return &PlatformFeeRepositoryWrapper{service: service}
}

// GetWallet retrieves a user's wallet (backward compatible)
func (w *PlatformFeeRepositoryWrapper) GetWallet(ctx context.Context, userID uuid.UUID) (*storageregistry.UserWallet, error) {
	wallet, err := w.service.GetUserWallet(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, nil
	}

	return &storageregistry.UserWallet{
		UserID:              userID,
		BalanceUSD:          wallet.BalanceUSD,
		LifetimeEarningsUSD: wallet.LifetimeEarningsUSD,
		LifetimeFeesUSD:     wallet.LifetimeSpentUSD,
		UpdatedAt:           wallet.UpdatedAt,
		CreatedAt:           wallet.CreatedAt,
	}, nil
}

// GetOrCreateWallet retrieves or creates a user's wallet (backward compatible)
func (w *PlatformFeeRepositoryWrapper) GetOrCreateWallet(ctx context.Context, userID uuid.UUID) (*storageregistry.UserWallet, error) {
	wallet, err := w.service.GetOrCreateUserWallet(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &storageregistry.UserWallet{
		UserID:              userID,
		BalanceUSD:          wallet.BalanceUSD,
		LifetimeEarningsUSD: wallet.LifetimeEarningsUSD,
		LifetimeFeesUSD:     wallet.LifetimeSpentUSD,
		UpdatedAt:           wallet.UpdatedAt,
		CreatedAt:           wallet.CreatedAt,
	}, nil
}

// GetWalletBalance retrieves a user's wallet balance (backward compatible)
func (w *PlatformFeeRepositoryWrapper) GetWalletBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	return w.service.GetUserBalance(ctx, userID)
}

// CreditWallet adds credits to a user's wallet (backward compatible)
func (w *PlatformFeeRepositoryWrapper) CreditWallet(ctx context.Context, userID uuid.UUID, amountUSD float64, stripePaymentID string) error {
	return w.service.CreditWalletUser(ctx, userID, amountUSD, stripePaymentID)
}

// DebitWallet debits a user's wallet (backward compatible)
func (w *PlatformFeeRepositoryWrapper) DebitWallet(ctx context.Context, userID uuid.UUID, amountUSD float64, description string) error {
	return w.service.DebitWalletUser(ctx, userID, amountUSD, description)
}

// HasWalletCreditReference checks if a credit reference exists (backward compatible)
func (w *PlatformFeeRepositoryWrapper) HasWalletCreditReference(ctx context.Context, reference string) (bool, error) {
	return w.service.HasUserWalletCreditReference(ctx, reference)
}

// ============================================
// Agent Billing Compatibility (billing.Controller interface)
// ============================================

// BillingControllerWrapper wraps the new wallet service to match the old billing.Controller interface
type BillingControllerWrapper struct {
	service *Service
}

// NewBillingControllerWrapper creates a wrapper for the old interface
func NewBillingControllerWrapper(service *Service) *BillingControllerWrapper {
	return &BillingControllerWrapper{service: service}
}

// GetOrCreateControls retrieves or creates billing controls (backward compatible)
func (w *BillingControllerWrapper) GetOrCreateControls(ctx context.Context, agentID string) (*billing.AgentBillingControls, error) {
	wallet, err := w.service.GetWalletByOwner(ctx, OwnerTypeUser, agentID)
	if err != nil || wallet == nil {
		wallet, err = w.service.GetWalletByOwner(ctx, OwnerTypeAgent, agentID)
		if err != nil {
			return nil, err
		}
	}
	if wallet == nil {
		return nil, fmt.Errorf("wallet not found for owner: %s", agentID)
	}

	return walletToBillingControls(wallet), nil
}

// CheckSpendCap checks if an agent can spend (backward compatible)
func (w *BillingControllerWrapper) CheckSpendCap(ctx context.Context, agentID string, estimatedCost float64) (bool, error) {
	return w.service.CheckAgentSpendCap(ctx, agentID, estimatedCost)
}

// ConsumeCredits consumes credits from an agent's wallet (backward compatible)
func (w *BillingControllerWrapper) ConsumeCredits(ctx context.Context, agentID string, amount float64) (*billing.CreditBalanceUpdate, error) {
	update, err := w.service.ConsumeUserOrAgentCredits(ctx, agentID, amount)
	if err != nil {
		return nil, err
	}

	return &billing.CreditBalanceUpdate{
		AgentID:     agentID,
		AmountUSD:   amount,
		PreviousUSD: update.PreviousBalance,
		CurrentUSD:  update.CurrentBalance,
	}, nil
}

// AddCredits adds credits to an agent's wallet (backward compatible)
func (w *BillingControllerWrapper) AddCredits(ctx context.Context, agentID string, amount float64) error {
	return w.service.AddCreditsToAgent(ctx, agentID, amount)
}

// GetAgentSpend returns spend summary for an agent (backward compatible)
func (w *BillingControllerWrapper) GetAgentSpend(ctx context.Context, agentID string, period string) (*billing.SpendSummary, error) {
	wallet, err := w.service.GetAgentWallet(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, fmt.Errorf("wallet not found for agent: %s", agentID)
	}

	check, err := w.service.CheckSpendCap(ctx, wallet.ID, 0)
	if err != nil {
		return nil, err
	}

	return &billing.SpendSummary{
		AgentID:         agentID,
		Period:          period,
		CreditBalance:   wallet.BalanceUSD,
		SpendCapMonthly: wallet.SpendCapMonthlyUSD,
		SpendCapDaily:   wallet.SpendCapDailyUSD,
		CapUtilization:  check.CapUtilization,
		GeneratedAt:     wallet.UpdatedAt,
	}, nil
}

// UpdateSpendCap updates spend caps for an agent (backward compatible)
func (w *BillingControllerWrapper) UpdateSpendCap(ctx context.Context, agentID string, dailyCap, monthlyCap *float64) error {
	return w.service.UpdateAgentSpendCaps(ctx, agentID, dailyCap, monthlyCap)
}

// Helper function to convert Wallet to AgentBillingControls
func walletToBillingControls(wallet *Wallet) *billing.AgentBillingControls {
	controls := &billing.AgentBillingControls{
		ID:               wallet.ID,
		AgentID:          wallet.OwnerID,
		CreditBalanceUSD: wallet.BalanceUSD,
		BillingMode:      wallet.BillingMode,
		CreatedAt:        wallet.CreatedAt,
		UpdatedAt:        wallet.UpdatedAt,
	}

	if wallet.SpendCapMonthlyUSD != nil {
		controls.SpendCapMonthlyUSD = wallet.SpendCapMonthlyUSD
	}
	if wallet.SpendCapDailyUSD != nil {
		controls.SpendCapDailyUSD = wallet.SpendCapDailyUSD
	}
	if wallet.TeamID != nil {
		controls.TeamID = wallet.TeamID
	}

	return controls
}
