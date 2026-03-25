package registry

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// ── Fee Calculation & Exemption Tests ─────────────────────────────────────

func TestIsAuthorExempt(t *testing.T) {
	tests := []struct {
		author string
		exempt bool
	}{
		{author: "functionfly", exempt: true},
		{author: "FunctionFly", exempt: false}, // case-sensitive
		{author: "FUNCTIONFLY", exempt: false},
		{author: "functionfly2", exempt: false},
		{author: "acme", exempt: false},
		{author: "test-author", exempt: false},
		{author: "", exempt: false},
	}

	for _, tt := range tests {
		t.Run(tt.author, func(t *testing.T) {
			result := IsAuthorExempt(tt.author)
			assert.Equal(t, tt.exempt, result, "IsAuthorExempt(%q) = %v, want %v", tt.author, result, tt.exempt)
		})
	}
}

func TestCalculatePublishFee(t *testing.T) {
	tests := []struct {
		author string
		fee    float64
	}{
		{author: "functionfly", fee: 0},         // exempt
		{author: "acme", fee: PublishFeeAmount}, // $2.99
		{author: "test", fee: PublishFeeAmount},  // $2.99
		{author: "", fee: PublishFeeAmount},      // non-exempt for empty author
	}

	for _, tt := range tests {
		t.Run(tt.author, func(t *testing.T) {
			result := CalculatePublishFee(tt.author)
			assert.Equal(t, tt.fee, result, "CalculatePublishFee(%q) = %v, want %v", tt.author, result, tt.fee)
		})
	}

	// Verify the constant value
	assert.Equal(t, 2.99, PublishFeeAmount)
}

func TestCalculateVersionUpdateFee(t *testing.T) {
	tests := []struct {
		author string
		fee    float64
	}{
		{author: "functionfly", fee: 0},                   // exempt
		{author: "acme", fee: VersionUpdateFeeAmount},     // $0.99
		{author: "test", fee: VersionUpdateFeeAmount},      // $0.99
		{author: "", fee: VersionUpdateFeeAmount},          // non-exempt for empty author
	}

	for _, tt := range tests {
		t.Run(tt.author, func(t *testing.T) {
			result := CalculateVersionUpdateFee(tt.author)
			assert.Equal(t, tt.fee, result, "CalculateVersionUpdateFee(%q) = %v, want %v", tt.author, result, tt.fee)
		})
	}

	// Verify the constant value
	assert.Equal(t, 0.99, VersionUpdateFeeAmount)
}

func TestCalculateCommission(t *testing.T) {
	tests := []struct {
		saleAmount float64
		commission float64
	}{
		{saleAmount: 100.00, commission: 15.00},  // 15%
		{saleAmount: 10.00, commission: 1.50},    // 15%
		{saleAmount: 0.00, commission: 0.00},     // 0% of 0
		{saleAmount: 1.00, commission: 0.15},     // 15% of $1
		{saleAmount: 33.33, commission: 5.00},    // 15% of $33.33 ≈ $5.00
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := CalculateCommission(tt.saleAmount)
			assert.InDelta(t, tt.commission, result, 0.01, "CalculateCommission(%f) = %f, want %f", tt.saleAmount, result, tt.commission)
		})
	}

	// Verify the constant value
	assert.Equal(t, 0.15, PlatformCommissionRate)
}

// ── Fee Constants Tests ─────────────────────────────────────────────────────

func TestFeeConstants(t *testing.T) {
	// Fee types
	assert.Equal(t, "publish", FeeTypePublish)
	assert.Equal(t, "version_update", FeeTypeVersionUpdate)
	assert.Equal(t, "commission", FeeTypeCommission)

	// Fee statuses
	assert.Equal(t, "pending", FeeStatusPending)
	assert.Equal(t, "completed", FeeStatusCompleted)
	assert.Equal(t, "failed", FeeStatusFailed)
	assert.Equal(t, "refunded", FeeStatusRefunded)

	// Fee amounts
	assert.Equal(t, 2.99, PublishFeeAmount)
	assert.Equal(t, 0.99, VersionUpdateFeeAmount)
	assert.Equal(t, 0.15, PlatformCommissionRate)
}

// ── Exempt Authors List Test ────────────────────────────────────────────────

func TestExemptAuthorsList(t *testing.T) {
	assert.Contains(t, ExemptAuthors, "functionfly")
	assert.Len(t, ExemptAuthors, 1, "Expected only one exempt author")
}

// ── FeeTransaction TableName Test ───────────────────────────────────────────

func TestFeeTransactionTableName(t *testing.T) {
	tx := FeeTransaction{}
	assert.Equal(t, "fee_transactions", tx.TableName())
}

// ── PlatformFee TableName Test ─────────────────────────────────────────────

func TestPlatformFeeTableName(t *testing.T) {
	pf := PlatformFee{}
	assert.Equal(t, "platform_fees_legacy_publish_audit", pf.TableName())
}

// ── UserWallet TableName Test ──────────────────────────────────────────────

func TestUserWalletTableName(t *testing.T) {
	uw := UserWallet{}
	assert.Equal(t, "user_wallets", uw.TableName())
}

// ── Fee Type Validation Tests ───────────────────────────────────────────────

func TestFeeTypes(t *testing.T) {
	// Verify fee types are distinct
	assert.NotEqual(t, FeeTypePublish, FeeTypeVersionUpdate)
	assert.NotEqual(t, FeeTypePublish, FeeTypeCommission)
	assert.NotEqual(t, FeeTypeVersionUpdate, FeeTypeCommission)
}

// ── Fee Status Validation Tests ────────────────────────────────────────────

func TestFeeStatuses(t *testing.T) {
	// Verify statuses are distinct
	assert.NotEqual(t, FeeStatusPending, FeeStatusCompleted)
	assert.NotEqual(t, FeeStatusPending, FeeStatusFailed)
	assert.NotEqual(t, FeeStatusPending, FeeStatusRefunded)
	assert.NotEqual(t, FeeStatusCompleted, FeeStatusFailed)
	assert.NotEqual(t, FeeStatusCompleted, FeeStatusRefunded)
	assert.NotEqual(t, FeeStatusFailed, FeeStatusRefunded)
}

// ── Platform Fee Model Tests ───────────────────────────────────────────────

func TestPlatformFeeModel(t *testing.T) {
	fnID := uuid.New()
	userID := uuid.New()

	pf := PlatformFee{
		ID:         uuid.New(),
		FunctionID: fnID,
		UserID:     userID,
		FeeType:    FeeTypePublish,
		AmountUSD:  2.99,
		Status:     FeeStatusCompleted,
	}

	assert.NotEqual(t, uuid.Nil, pf.ID)
	assert.Equal(t, fnID, pf.FunctionID)
	assert.Equal(t, userID, pf.UserID)
	assert.Equal(t, FeeTypePublish, pf.FeeType)
	assert.Equal(t, 2.99, pf.AmountUSD)
	assert.Equal(t, FeeStatusCompleted, pf.Status)
}

// ── UserWallet Model Tests ─────────────────────────────────────────────────

func TestUserWalletModel(t *testing.T) {
	userID := uuid.New()

	wallet := UserWallet{
		UserID:              userID,
		BalanceUSD:          100.50,
		LifetimeEarningsUSD: 200.00,
		LifetimeFeesUSD:     50.00,
	}

	assert.Equal(t, userID, wallet.UserID)
	assert.Equal(t, 100.50, wallet.BalanceUSD)
	assert.Equal(t, 200.00, wallet.LifetimeEarningsUSD)
	assert.Equal(t, 50.00, wallet.LifetimeFeesUSD)
}

// ── FeeTransaction Model Tests ─────────────────────────────────────────────

func TestFeeTransactionModel(t *testing.T) {
	userID := uuid.New()

	tx := FeeTransaction{
		ID:        uuid.New(),
		UserID:    userID,
		Kind:      "credit",
		AmountUSD: 50.00,
		Status:    "completed",
		Reference: "stripe_123",
	}

	assert.NotEqual(t, uuid.Nil, tx.ID)
	assert.Equal(t, userID, tx.UserID)
	assert.Equal(t, "credit", tx.Kind)
	assert.Equal(t, 50.00, tx.AmountUSD)
	assert.Equal(t, "completed", tx.Status)
	assert.Equal(t, "stripe_123", tx.Reference)
}

// ── Publish Fee Amount Boundaries Tests ─────────────────────────────────────

func TestPublishFeeAmountBoundaries(t *testing.T) {
	// Verify publish fee is positive
	assert.Greater(t, PublishFeeAmount, 0.0)

	// Verify publish fee is reasonable (not too high)
	assert.Less(t, PublishFeeAmount, 10.0)

	// Verify version update fee is less than publish fee
	assert.Less(t, VersionUpdateFeeAmount, PublishFeeAmount)
}

// ── Commission Rate Boundaries Tests ───────────────────────────────────────

func TestCommissionRateBoundaries(t *testing.T) {
	// Verify commission rate is between 0 and 1 (0% to 100%)
	assert.GreaterOrEqual(t, PlatformCommissionRate, 0.0)
	assert.LessOrEqual(t, PlatformCommissionRate, 1.0)

	// Verify it's exactly 15%
	assert.InDelta(t, 0.15, PlatformCommissionRate, 0.001)
}

// ── Multiple Exemption Checks ───────────────────────────────────────────────

func TestIsAuthorExempt_MultipleChecks(t *testing.T) {
	// Run multiple times to catch any race conditions or state issues
	for i := 0; i < 100; i++ {
		assert.True(t, IsAuthorExempt("functionfly"), "Iteration %d: functionfly should be exempt", i)
		assert.False(t, IsAuthorExempt("other"), "Iteration %d: other should not be exempt", i)
	}
}

// ── Fee Calculation Consistency Tests ───────────────────────────────────────

func TestCalculatePublishFee_Consistency(t *testing.T) {
	// Run multiple times to ensure consistent results
	author := "testauthor"

	for i := 0; i < 100; i++ {
		fee := CalculatePublishFee(author)
		assert.Equal(t, 2.99, fee, "Iteration %d: fee should be consistent", i)
	}
}

func TestCalculateCommission_Consistency(t *testing.T) {
	// Run multiple times to ensure consistent results
	saleAmount := 100.00

	for i := 0; i < 100; i++ {
		comm := CalculateCommission(saleAmount)
		assert.InDelta(t, 15.00, comm, 0.01, "Iteration %d: commission should be consistent", i)
	}
}
