package wallet

import (
	"context"
	"testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=functionfly sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("Database not available, skipping test")
		return nil
	}
	return db
}

func TestCredit_IdempotencyPreCheck(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewRepository(db)
	_ = NewService(repo, nil)

	ctx := context.Background()
	wallet, err := repo.GetOrCreateWalletForUser(ctx, uuid.New())
	require.NoError(t, err)

	idempotencyKey := "test:idempotency:123"
	amount := 100.0

	req := CreditRequest{
		WalletID:       wallet.ID,
		AmountUSD:      amount,
		Reference:      "test-ref-1",
		IdempotencyKey: idempotencyKey,
		TriggeredBy: TriggeredByInfo{
			Type: "test",
			ID:   "test-user",
		},
	}

	update1, err := repo.Credit(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, amount, update1.Amount)
	assert.Equal(t, amount, update1.CurrentBalance)

	update2, err := repo.Credit(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, update1.TransactionID, update2.TransactionID)
	assert.Equal(t, update1.CurrentBalance, update2.CurrentBalance)
}

func TestCredit_ConcurrentSameIdempotencyKey(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewRepository(db)

	ctx := context.Background()
	wallet, err := repo.GetOrCreateWalletForUser(ctx, uuid.New())
	require.NoError(t, err)

	idempotencyKey := "test:concurrent:456"
	amount := 50.0

	req := CreditRequest{
		WalletID:       wallet.ID,
		AmountUSD:      amount,
		Reference:      "test-ref-2",
		IdempotencyKey: idempotencyKey,
		TriggeredBy: TriggeredByInfo{
			Type: "test",
			ID:   "test-user",
		},
	}

	done := make(chan bool, 2)

	go func() {
		_, err := repo.Credit(ctx, req)
		if err != nil {
			t.Logf("First credit error: %v", err)
		}
		done <- true
	}()

	go func() {
		_, err := repo.Credit(ctx, req)
		if err != nil {
			t.Logf("Second credit error: %v", err)
		}
		done <- true
	}()

	<-done
	<-done

	tx, err := repo.GetTransactionByIdempotencyKey(ctx, idempotencyKey)
	require.NoError(t, err)
	assert.NotNil(t, tx)
}

func TestDebit_InsufficientBalance(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewRepository(db)

	ctx := context.Background()
	wallet, err := repo.GetOrCreateWalletForUser(ctx, uuid.New())
	require.NoError(t, err)

	creditReq := CreditRequest{
		WalletID:       wallet.ID,
		AmountUSD:      100.0,
		Reference:      "test-credit",
		IdempotencyKey: "test:debit:insufficient:1",
		TriggeredBy:    TriggeredByInfo{Type: "test", ID: "test"},
	}
	_, err = repo.Credit(ctx, creditReq)
	require.NoError(t, err)

	debitReq := DebitRequest{
		WalletID:        wallet.ID,
		AmountUSD:       200.0,
		TransactionType: TransactionTypeDebit,
		TriggeredBy:     TriggeredByInfo{Type: "test", ID: "test"},
	}
	_, err = repo.Debit(ctx, debitReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")
}

func TestWalletCreation_RaceCondition(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewRepository(db)

	ctx := context.Background()
	userID := uuid.New()

	done := make(chan *Wallet, 2)

	go func() {
		wallet, err := repo.GetOrCreateWalletForUser(ctx, userID)
		if err != nil {
			t.Logf("First wallet error: %v", err)
		}
		done <- wallet
	}()

	go func() {
		wallet, err := repo.GetOrCreateWalletForUser(ctx, userID)
		if err != nil {
			t.Logf("Second wallet error: %v", err)
		}
		done <- wallet
	}()

	wallet1 := <-done
	wallet2 := <-done

	assert.Equal(t, wallet1.ID, wallet2.ID)
}

func TestBalanceValidator_NoDrift(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewRepository(db)
	validator := NewBalanceValidator(db, repo)

	ctx := context.Background()
	wallet, err := repo.GetOrCreateWalletForUser(ctx, uuid.New())
	require.NoError(t, err)

	creditReq := CreditRequest{
		WalletID:       wallet.ID,
		AmountUSD:      100.0,
		Reference:      "test-validation",
		IdempotencyKey: "test:validator:1",
		TriggeredBy:    TriggeredByInfo{Type: "test", ID: "test"},
	}
	_, err = repo.Credit(ctx, creditReq)
	require.NoError(t, err)

	result, err := validator.ReconcileWallet(ctx, wallet.ID)
	require.NoError(t, err)
	assert.Equal(t, "ok", result.Status)
	assert.False(t, result.Fixed)
}

func TestReconcileWallet_DriftDetected(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewRepository(db)
	validator := NewBalanceValidator(db, repo)
	validator.SetAutoFix(true, 10.0)

	ctx := context.Background()
	wallet, err := repo.GetOrCreateWalletForUser(ctx, uuid.New())
	require.NoError(t, err)

	creditReq := CreditRequest{
		WalletID:       wallet.ID,
		AmountUSD:      100.0,
		Reference:      "test-drift",
		IdempotencyKey: "test:drift:1",
		TriggeredBy:    TriggeredByInfo{Type: "test", ID: "test"},
	}
	_, err = repo.Credit(ctx, creditReq)
	require.NoError(t, err)

	db.Model(&Wallet{}).Where("id = ?", wallet.ID).Update("balance_usd", 50.0)

	result, err := validator.ReconcileWallet(ctx, wallet.ID)
	require.NoError(t, err)
	assert.Equal(t, "fixed", result.Status)
	assert.True(t, result.Fixed)
	assert.NotZero(t, result.Drift)
}
