package wallet

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestRepository_NewRepository(t *testing.T) {
	repo := NewRepository(nil)
	assert.NotNil(t, repo)
}

func TestRepository_Credit_ZeroAmount(t *testing.T) {
	repo := NewRepository(nil)

	req := CreditRequest{
		WalletID:  testUUID(),
		AmountUSD: 0,
	}

	_, err := repo.Credit(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestRepository_Credit_NegativeAmount(t *testing.T) {
	repo := NewRepository(nil)

	req := CreditRequest{
		WalletID:  testUUID(),
		AmountUSD: -50.0,
	}

	_, err := repo.Credit(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestRepository_Debit_ZeroAmount(t *testing.T) {
	repo := NewRepository(nil)

	req := DebitRequest{
		WalletID:  testUUID(),
		AmountUSD: 0,
	}

	_, err := repo.Debit(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestRepository_Debit_NegativeAmount(t *testing.T) {
	repo := NewRepository(nil)

	req := DebitRequest{
		WalletID:  testUUID(),
		AmountUSD: -50.0,
	}

	_, err := repo.Debit(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestRepository_HasTransactionWithIdempotencyKey_EmptyKey(t *testing.T) {
	repo := NewRepository(nil)

	has, err := repo.HasTransactionWithIdempotencyKey(context.Background(), "")
	assert.NoError(t, err)
	assert.False(t, has)
}

func Testptr(t *testing.T) {
	now := time.Now()
	p := ptr(now)
	assert.Equal(t, &now, p)
}

func TestWallet_IsActive(t *testing.T) {
	wallet := &Wallet{Status: WalletStatusActive}
	assert.True(t, wallet.IsActive())

	wallet.Status = WalletStatusSuspended
	assert.False(t, wallet.IsActive())

	wallet.Status = WalletStatusClosed
	assert.False(t, wallet.IsActive())
}

func TestWallet_HasSufficientBalance(t *testing.T) {
	wallet := &Wallet{BalanceUSD: 100.0}

	assert.True(t, wallet.HasSufficientBalance(50.0))
	assert.True(t, wallet.HasSufficientBalance(100.0))
	assert.False(t, wallet.HasSufficientBalance(100.01))
}

func TestWalletTransaction_IsCredit(t *testing.T) {
	tx := &WalletTransaction{TransactionType: TransactionTypeCredit}
	assert.True(t, tx.IsCredit())
	assert.False(t, tx.IsDebit())

	tx.TransactionType = TransactionTypeRefund
	assert.True(t, tx.IsCredit())

	tx.TransactionType = TransactionTypeTransferIn
	assert.True(t, tx.IsCredit())
}

func TestWalletTransaction_IsDebit(t *testing.T) {
	tx := &WalletTransaction{TransactionType: TransactionTypeDebit}
	assert.True(t, tx.IsDebit())
	assert.False(t, tx.IsCredit())

	tx.TransactionType = TransactionTypeFeePayment
	assert.True(t, tx.IsDebit())

	tx.TransactionType = TransactionTypeExecutionCharge
	assert.True(t, tx.IsDebit())

	tx.TransactionType = TransactionTypeCommission
	assert.True(t, tx.IsDebit())
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "user", OwnerTypeUser)
	assert.Equal(t, "agent", OwnerTypeAgent)

	assert.Equal(t, "unified", WalletTypeUnified)
	assert.Equal(t, "registry", WalletTypeRegistry)
	assert.Equal(t, "execution", WalletTypeExecution)

	assert.Equal(t, "credit", TransactionTypeCredit)
	assert.Equal(t, "debit", TransactionTypeDebit)
	assert.Equal(t, "fee_payment", TransactionTypeFeePayment)
	assert.Equal(t, "execution_charge", TransactionTypeExecutionCharge)
	assert.Equal(t, "commission", TransactionTypeCommission)
	assert.Equal(t, "refund", TransactionTypeRefund)
	assert.Equal(t, "transfer_in", TransactionTypeTransferIn)
	assert.Equal(t, "transfer_out", TransactionTypeTransferOut)
	assert.Equal(t, "adjustment", TransactionTypeAdjustment)

	assert.Equal(t, "pending", TransactionStatusPending)
	assert.Equal(t, "completed", TransactionStatusCompleted)
	assert.Equal(t, "failed", TransactionStatusFailed)
	assert.Equal(t, "reversed", TransactionStatusReversed)

	assert.Equal(t, "active", WalletStatusActive)
	assert.Equal(t, "suspended", WalletStatusSuspended)
	assert.Equal(t, "closed", WalletStatusClosed)

	assert.Equal(t, "per_wallet", BillingModePerWallet)
	assert.Equal(t, "per_agent", BillingModePerAgent)
	assert.Equal(t, "per_tenant", BillingModePerTenant)
	assert.Equal(t, "per_team", BillingModePerTeam)

	assert.Equal(t, "publish", FeeTypePublish)
	assert.Equal(t, "version_update", FeeTypeVersionUpdate)
	assert.Equal(t, "commission", FeeTypeCommission)
}

func TestWallet_TableName(t *testing.T) {
	wallet := &Wallet{}
	assert.Equal(t, "wallets", wallet.TableName())
}

func TestWalletTransaction_TableName(t *testing.T) {
	tx := &WalletTransaction{}
	assert.Equal(t, "wallet_transactions", tx.TableName())
}

func TestSpendCapCheck_Defaults(t *testing.T) {
	check := &SpendCapCheck{}
	assert.True(t, check.Allowed)
	assert.Empty(t, check.Reason)
}

func TestBalanceUpdate_Fields(t *testing.T) {
	update := &BalanceUpdate{
		WalletID:        testUUID(),
		PreviousBalance: 100.0,
		CurrentBalance:  150.0,
		Amount:          50.0,
		TransactionID:   testUUID(),
	}

	assert.Equal(t, testUUID(), update.WalletID)
	assert.Equal(t, 100.0, update.PreviousBalance)
	assert.Equal(t, 150.0, update.CurrentBalance)
	assert.Equal(t, 50.0, update.Amount)
	assert.Equal(t, testUUID(), update.TransactionID)
}

func TestCreditRequest_Fields(t *testing.T) {
	req := CreditRequest{
		WalletID:       testUUID(),
		AmountUSD:      100.0,
		Reference:      "ref_123",
		IdempotencyKey: "idem_123",
		TriggeredBy: TriggeredByInfo{
			Type: "user",
			ID:   "user_123",
		},
		Metadata: map[string]interface{}{"key": "value"},
	}

	assert.Equal(t, testUUID(), req.WalletID)
	assert.Equal(t, 100.0, req.AmountUSD)
	assert.Equal(t, "ref_123", req.Reference)
	assert.Equal(t, "idem_123", req.IdempotencyKey)
	assert.Equal(t, "user", req.TriggeredBy.Type)
	assert.Equal(t, "user_123", req.TriggeredBy.ID)
	assert.Equal(t, "value", req.Metadata["key"])
}

func TestDebitRequest_Fields(t *testing.T) {
	execID := testUUID()
	funcID := testUUID()
	feeType := "publish"

	req := DebitRequest{
		WalletID:        testUUID(),
		AmountUSD:       50.0,
		TransactionType: TransactionTypeFeePayment,
		Reference:       "ref_456",
		TriggeredBy: TriggeredByInfo{
			Type: "system",
			ID:   "billing",
		},
		ExecutionID: &execID,
		FunctionID:  &funcID,
		FeeType:     &feeType,
		Metadata:    map[string]interface{}{"execution": "true"},
	}

	assert.Equal(t, testUUID(), req.WalletID)
	assert.Equal(t, 50.0, req.AmountUSD)
	assert.Equal(t, TransactionTypeFeePayment, req.TransactionType)
	assert.Equal(t, "ref_456", req.Reference)
	assert.Equal(t, "system", req.TriggeredBy.Type)
	assert.Equal(t, "billing", req.TriggeredBy.ID)
	assert.Equal(t, &execID, req.ExecutionID)
	assert.Equal(t, &funcID, req.FunctionID)
	assert.Equal(t, &feeType, req.FeeType)
	assert.Equal(t, "true", req.Metadata["execution"])
}

func TestWalletCreationRequest_Defaults(t *testing.T) {
	req := WalletCreationRequest{}

	assert.Empty(t, req.OwnerType)
	assert.Empty(t, req.OwnerID)
	assert.Nil(t, req.UserID)
	assert.Nil(t, req.AgentID)
	assert.Empty(t, req.WalletType)
	assert.Equal(t, float64(0), req.InitialBalanceUSD)
	assert.Nil(t, req.SpendCapMonthlyUSD)
	assert.Nil(t, req.SpendCapDailyUSD)
	assert.Empty(t, req.BillingMode)
	assert.Nil(t, req.TeamID)
}

func TestWalletSummary_Fields(t *testing.T) {
	agentID := "agent_123"
	summary := &WalletSummary{
		WalletID:                 testUUID(),
		OwnerType:                OwnerTypeAgent,
		OwnerID:                  agentID,
		AgentID:                  &agentID,
		BalanceUSD:               500.0,
		Status:                   WalletStatusActive,
		TotalCreditsUSD:          1000.0,
		TotalDebitsUSD:           500.0,
		TotalFeesPaidUSD:         100.0,
		TotalExecutionChargesUSD: 300.0,
		TotalCommissionsUSD:      50.0,
		TotalTransactions:        50,
		PendingTransactions:      2,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	assert.Equal(t, testUUID(), summary.WalletID)
	assert.Equal(t, OwnerTypeAgent, summary.OwnerType)
	assert.Equal(t, agentID, summary.OwnerID)
	assert.Equal(t, &agentID, summary.AgentID)
	assert.Equal(t, 500.0, summary.BalanceUSD)
	assert.Equal(t, WalletStatusActive, summary.Status)
	assert.Equal(t, 1000.0, summary.TotalCreditsUSD)
	assert.Equal(t, 500.0, summary.TotalDebitsUSD)
	assert.Equal(t, int64(50), summary.TotalTransactions)
}

type mockGORM struct {
	db *gorm.DB
}

func (m *mockGORM) WithContext(ctx context.Context) *mockGORM {
	return m
}
