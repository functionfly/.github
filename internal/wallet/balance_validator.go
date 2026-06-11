package wallet

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BalanceValidator struct {
	db                 *gorm.DB
	repo              *Repository
	autoFixEnabled    bool
	autoFixThreshold  float64
}

type ReconciliationResult struct {
	WalletID        uuid.UUID
	StoredBalance   float64
	ComputedBalance float64
	Drift           float64
	TxCount         int
	Status          string
	Fixed           bool
}

func NewBalanceValidator(db *gorm.DB, repo *Repository) *BalanceValidator {
	return &BalanceValidator{
		db:                db,
		repo:              repo,
		autoFixEnabled:    false,
		autoFixThreshold: 1.0,
	}
}

func (v *BalanceValidator) SetAutoFix(enabled bool, threshold float64) {
	v.autoFixEnabled = enabled
	v.autoFixThreshold = threshold
}

func (v *BalanceValidator) ReconcileWallet(ctx context.Context, walletID uuid.UUID) (*ReconciliationResult, error) {
	wallet, err := v.repo.GetWalletByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, fmt.Errorf("wallet not found: %s", walletID)
	}

	var computedBalance float64
	err = v.db.Model(&WalletTransaction{}).
		Where("wallet_id = ? AND status = ?", walletID, TransactionStatusCompleted).
		Select("COALESCE(SUM(amount_usd), 0)").
		Scan(&computedBalance).Error
	if err != nil {
		return nil, fmt.Errorf("failed to compute balance: %w", err)
	}

	drift := wallet.BalanceUSD - computedBalance

	result := &ReconciliationResult{
		WalletID:        walletID,
		StoredBalance:   wallet.BalanceUSD,
		ComputedBalance: computedBalance,
		Drift:           drift,
		Status:          "ok",
	}

	if walletAbs(drift) > 0.001 {
		result.Status = "drift_detected"
		BalanceDriftDetected.Inc()

		if v.autoFixEnabled && walletAbs(drift) < v.autoFixThreshold {
			if err := v.FixBalance(ctx, walletID, computedBalance); err != nil {
				result.Status = "error"
				logrus.Error("Failed to fix balance drift",
					"wallet_id", walletID,
					"drift", drift,
					"err", err)
				return result, err
			}
			result.Fixed = true
			result.Status = "fixed"
		}

		v.logDriftAudit(ctx, result)
	}

	return result, nil
}

func (v *BalanceValidator) ReconcileAll(ctx context.Context) ([]ReconciliationResult, error) {
	var walletIDs []uuid.UUID
	err := v.db.Model(&Wallet{}).
		Where("status = ?", WalletStatusActive).
		Pluck("id", &walletIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list wallets: %w", err)
	}

	var results []ReconciliationResult
	for _, walletID := range walletIDs {
		result, err := v.ReconcileWallet(ctx, walletID)
		if err != nil {
			logrus.Error("Reconciliation error for wallet",
				"wallet_id", walletID,
				"err", err)
			continue
		}
		results = append(results, *result)
	}
	return results, nil
}

func (v *BalanceValidator) FixBalance(ctx context.Context, walletID uuid.UUID, correctBalance float64) error {
	return v.db.Model(&Wallet{}).
		Where("id = ?", walletID).
		Update("balance_usd", correctBalance).Error
}

func (v *BalanceValidator) logDriftAudit(ctx context.Context, result *ReconciliationResult) {
	audit := &WalletBalanceAudit{
		ID:              uuid.New(),
		WalletID:        result.WalletID,
		StoredBalance:   result.StoredBalance,
		ComputedBalance: result.ComputedBalance,
		Drift:           result.Drift,
		Fixed:           result.Fixed,
	}
	if result.Fixed {
		now := time.Now()
		audit.FixedAt = &now
	}

	v.db.WithContext(ctx).Create(audit)
}

func walletAbs(x float64) float64 {
	return math.Abs(x)
}
