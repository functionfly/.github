package registry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Fee configuration constants
const (
	FeeTypePublish       = "publish"
	FeeTypeVersionUpdate = "version_update"
	FeeTypeCommission    = "commission"

	FeeStatusPending   = "pending"
	FeeStatusCompleted = "completed"
	FeeStatusFailed    = "failed"
	FeeStatusRefunded  = "refunded"

	// Publish Fee: $2.99 per function
	PublishFeeAmount = 2.99

	// Version Update Fee: $0.99 per version
	VersionUpdateFeeAmount = 0.99

	// Platform Commission: 15%
	PlatformCommissionRate = 0.15
)

// Exempt authors from fees
var ExemptAuthors = []string{"functionfly"}

// FeeTransaction is an internal ledger record for wallet transactions
type FeeTransaction struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	Kind        string         `json:"kind" gorm:"not null"` // 'credit', 'debit', 'fee_payment', 'commission'
	AmountUSD   float64        `json:"amount_usd" gorm:"type:decimal(14,4);not null"`
	Status      string         `json:"status" gorm:"not null;default:'completed'"`
	Reference   string         `json:"reference,omitempty" gorm:"type:text"` // fee_id or stripe ref
	Metadata    datatypes.JSON `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt   time.Time      `json:"created_at" gorm:"not null;default:now()"`
}

// TableName returns the database table name for FeeTransaction.
func (FeeTransaction) TableName() string {
	return "fee_transactions"
}

// PlatformFeeRepository handles platform fee and wallet database operations
type PlatformFeeRepository struct {
	db *gorm.DB
}

// NewPlatformFeeRepository creates a new platform fee repository
func NewPlatformFeeRepository(db *gorm.DB) *PlatformFeeRepository {
	return &PlatformFeeRepository{db: db}
}

// EnableUnifiedWallet switches this repository to use the unified wallet system
// This is called after the migration (000255) completes
func (r *PlatformFeeRepository) EnableUnifiedWallet() {
	// Migration complete - platform fee repo will use wallets table directly
	// In Phase 2, this will be removed and callers will use wallet.Service directly
}

// GetWallet retrieves a user's wallet
func (r *PlatformFeeRepository) GetWallet(ctx context.Context, userID uuid.UUID) (*UserWallet, error) {
	var wallet UserWallet
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	return &wallet, nil
}

// GetOrCreateWallet retrieves or creates a user's wallet
func (r *PlatformFeeRepository) GetOrCreateWallet(ctx context.Context, userID uuid.UUID) (*UserWallet, error) {
	var wallet UserWallet
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error
	if err == nil {
		return &wallet, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Create new wallet
	wallet = UserWallet{
		UserID:              userID,
		BalanceUSD:          0,
		LifetimeEarningsUSD: 0,
		LifetimeFeesUSD:     0,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	if err := r.db.WithContext(ctx).Create(&wallet).Error; err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}
	return &wallet, nil
}

// GetWalletBalance retrieves the current balance for a user
func (r *PlatformFeeRepository) GetWalletBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	wallet, err := r.GetWallet(ctx, userID)
	if err != nil {
		return 0, err
	}
	if wallet == nil {
		return 0, nil
	}
	return wallet.BalanceUSD, nil
}

// CreditWallet adds funds to a user's wallet. It is idempotent based on stripePaymentID reference.
// If the reference already exists, the operation succeeds without adding duplicate credits.
func (r *PlatformFeeRepository) CreditWallet(ctx context.Context, userID uuid.UUID, amountUSD float64, stripePaymentID string) error {
	if amountUSD <= 0 {
		return fmt.Errorf("credit amount must be positive")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if this credit has already been applied (idempotency check)
		var existingCount int64
		if err := tx.Model(&FeeTransaction{}).
			Where("reference = ? AND kind = ?", stripePaymentID, "credit").
			Count(&existingCount).Error; err != nil {
			return fmt.Errorf("failed to check for existing credit: %w", err)
		}
		if existingCount > 0 {
			// Already credited, return success without adding duplicate
			return nil
		}

		// Lock the wallet row for update
		var wallet UserWallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&wallet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create wallet if it doesn't exist
				wallet = UserWallet{
					UserID:              userID,
					BalanceUSD:          amountUSD,
					LifetimeEarningsUSD: amountUSD,
					LifetimeFeesUSD:     0,
					CreatedAt:           time.Now(),
					UpdatedAt:           time.Now(),
				}
				if err := tx.Create(&wallet).Error; err != nil {
					return fmt.Errorf("failed to create wallet: %w", err)
				}
				return nil
			}
			return fmt.Errorf("failed to lock wallet: %w", err)
		}

		// Update balance and lifetime earnings
		wallet.BalanceUSD += amountUSD
		wallet.LifetimeEarningsUSD += amountUSD
		wallet.UpdatedAt = time.Now()

		if err := tx.Save(&wallet).Error; err != nil {
			return fmt.Errorf("failed to update wallet: %w", err)
		}

		// Record the financial transaction
		txRecord := FeeTransaction{
			UserID:    userID,
			Kind:      "credit",
			AmountUSD: amountUSD,
			Status:    "completed",
			Reference: stripePaymentID,
			Metadata:  datatypes.JSON([]byte(fmt.Sprintf(`{"type":"wallet_credit","stripe_payment_id":"%s"}`, stripePaymentID))),
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&txRecord).Error; err != nil {
			// Check if it's a unique constraint violation (another concurrent request beat us)
			if isUniqueViolation(err) {
				// Credit already applied, rollback the transaction but return success
				return nil
			}
			return fmt.Errorf("failed to record transaction: %w", err)
		}

		return nil
	})
}

// DebitWallet removes funds from a user's wallet (for fee payment)
func (r *PlatformFeeRepository) DebitWallet(ctx context.Context, userID uuid.UUID, amountUSD float64, description string) error {
	if amountUSD <= 0 {
		return fmt.Errorf("debit amount must be positive")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock the wallet row for update
		var wallet UserWallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&wallet).Error; err != nil {
			return fmt.Errorf("failed to lock wallet: %w", err)
		}

		// Check sufficient balance
		if wallet.BalanceUSD < amountUSD {
			return fmt.Errorf("insufficient balance: have %.4f, need %.4f", wallet.BalanceUSD, amountUSD)
		}

		// Update balance and lifetime fees
		wallet.BalanceUSD -= amountUSD
		wallet.LifetimeFeesUSD += amountUSD
		wallet.UpdatedAt = time.Now()

		if err := tx.Save(&wallet).Error; err != nil {
			return fmt.Errorf("failed to update wallet: %w", err)
		}

		// Record the financial transaction
		txRecord := FeeTransaction{
			UserID:    userID,
			Kind:      "fee_payment",
			AmountUSD: amountUSD,
			Status:    "completed",
			Reference: fmt.Sprintf("fee_payment_%s", description),
			Metadata:  datatypes.JSON([]byte(fmt.Sprintf(`{"type":"fee_payment","description":"%s"}`, description))),
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&txRecord).Error; err != nil {
			return fmt.Errorf("failed to record transaction: %w", err)
		}

		return nil
	})
}

// RecordPlatformFee creates a platform fee record
func (r *PlatformFeeRepository) RecordPlatformFee(ctx context.Context, fee *PlatformFee) error {
	if fee.ID == uuid.Nil {
		fee.ID = uuid.New()
	}
	if fee.ChargedAt.IsZero() {
		fee.ChargedAt = time.Now()
	}
	if err := r.db.WithContext(ctx).Create(fee).Error; err != nil {
		return fmt.Errorf("failed to record platform fee: %w", err)
	}
	return nil
}

// GetPlatformFeesByFunction retrieves all fees for a function
func (r *PlatformFeeRepository) GetPlatformFeesByFunction(ctx context.Context, functionID uuid.UUID) ([]PlatformFee, error) {
	var fees []PlatformFee
	if err := r.db.WithContext(ctx).
		Where("function_id = ?", functionID).
		Order("charged_at DESC").
		Find(&fees).Error; err != nil {
		return nil, fmt.Errorf("failed to get platform fees: %w", err)
	}
	return fees, nil
}

// GetPlatformFeesByUser retrieves all fees for a user
func (r *PlatformFeeRepository) GetPlatformFeesByUser(ctx context.Context, userID uuid.UUID) ([]PlatformFee, error) {
	var fees []PlatformFee
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("charged_at DESC").
		Find(&fees).Error; err != nil {
		return nil, fmt.Errorf("failed to get platform fees: %w", err)
	}
	return fees, nil
}

// ListPlatformFeesByUserPaged returns a page of platform fee rows for billing history.
func (r *PlatformFeeRepository) ListPlatformFeesByUserPaged(ctx context.Context, userID uuid.UUID, limit, offset int) ([]PlatformFee, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	var total int64
	base := r.db.WithContext(ctx).Model(&PlatformFee{}).Where("user_id = ?", userID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count platform fees: %w", err)
	}
	var fees []PlatformFee
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("charged_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&fees).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list platform fees: %w", err)
	}
	return fees, total, nil
}

// HasWalletCreditReference reports whether a wallet credit was already recorded for this reference (e.g. Stripe Checkout session id).
func (r *PlatformFeeRepository) HasWalletCreditReference(ctx context.Context, reference string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&FeeTransaction{}).
		Where("reference = ? AND kind = ?", reference, "credit").
		Count(&n).Error
	if err != nil {
		return false, fmt.Errorf("failed to check wallet credit reference: %w", err)
	}
	return n > 0, nil
}

// UpdatePlatformFeeStatus updates the status of a platform fee
func (r *PlatformFeeRepository) UpdatePlatformFeeStatus(ctx context.Context, feeID uuid.UUID, status string, stripePaymentID string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if stripePaymentID != "" {
		updates["stripe_payment_id"] = stripePaymentID
	}

	if err := r.db.WithContext(ctx).
		Model(&PlatformFee{}).
		Where("id = ?", feeID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update platform fee status: %w", err)
	}
	return nil
}

// IsAuthorExempt checks if an author is exempt from platform fees
func IsAuthorExempt(author string) bool {
	for _, exempt := range ExemptAuthors {
		if author == exempt {
			return true
		}
	}
	return false
}

// CalculateCommission calculates the platform commission for a transaction
func CalculateCommission(saleAmountUSD float64) float64 {
	return saleAmountUSD * PlatformCommissionRate
}

// CalculatePublishFee returns the publish fee amount
func CalculatePublishFee(author string) float64 {
	if IsAuthorExempt(author) {
		return 0
	}
	return PublishFeeAmount
}

// CalculateVersionUpdateFee returns the version update fee amount
func CalculateVersionUpdateFee(author string) float64 {
	if IsAuthorExempt(author) {
		return 0
	}
	return VersionUpdateFeeAmount
}

// isUniqueViolation checks if an error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		// PostgreSQL error code for unique_violation is 23505
		return pgErr.Code == "23505"
	}
	return false
}
