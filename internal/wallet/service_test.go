package wallet

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestService_NewService(t *testing.T) {
	svc := NewService(nil, nil)
	assert.NotNil(t, svc)
}

func TestService_SetNotificationFunc(t *testing.T) {
	svc := NewService(nil, nil)

	fn := func(ctx context.Context, userID uuid.UUID, notificationType string, data map[string]interface{}) error {
		return nil
	}

	svc.SetNotificationFunc(fn)
	assert.NotNil(t, svc.notifyFunc)
}

func TestService_invalidateWalletCache(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := NewService(nil, rdb)

	key := "wallet:12345678-1234-1234-1234-123456789012"
	rdb.Set(context.Background(), key, "test", 0)
	rdb.Set(context.Background(), "wallet:summary:"+key, "test", 0)

	svc.invalidateWalletCache(context.Background(), testUUID())

	exists, err := rdb.Exists(context.Background(), key).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}

func TestService_invalidateWalletCache_NilRedis(t *testing.T) {
	svc := NewService(nil, nil)
	svc.invalidateWalletCache(context.Background(), testUUID())
}

func TestService_CacheWalletBalance(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := NewService(nil, rdb)

	walletID := testUUID()
	err = svc.CacheWalletBalance(context.Background(), walletID, 100.50, 0)
	assert.NoError(t, err)

	data, err := rdb.Get(context.Background(), "wallet:"+walletID.String()).Result()
	assert.NoError(t, err)
	assert.Contains(t, data, "100.5")
}

func TestService_CacheWalletBalance_NilRedis(t *testing.T) {
	svc := NewService(nil, nil)
	err := svc.CacheWalletBalance(context.Background(), testUUID(), 100.50, 0)
	assert.NoError(t, err)
}

func TestService_GetCachedWalletBalance(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := NewService(nil, rdb)

	walletID := testUUID()

	balance, found := svc.GetCachedWalletBalance(context.Background(), walletID)
	assert.False(t, found)
	assert.Equal(t, float64(0), balance)

	rdb.Set(context.Background(), "wallet:"+walletID.String(), `{"balance":150.25}`, 0)

	balance, found = svc.GetCachedWalletBalance(context.Background(), walletID)
	assert.True(t, found)
	assert.Equal(t, 150.25, balance)
}

func TestService_GetCachedWalletBalance_NilRedis(t *testing.T) {
	svc := NewService(nil, nil)
	balance, found := svc.GetCachedWalletBalance(context.Background(), testUUID())
	assert.False(t, found)
	assert.Equal(t, float64(0), balance)
}

func TestService_GetCachedWalletBalance_InvalidJSON(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := NewService(nil, rdb)

	walletID := testUUID()
	rdb.Set(context.Background(), "wallet:"+walletID.String(), "invalid json", 0)

	balance, found := svc.GetCachedWalletBalance(context.Background(), walletID)
	assert.False(t, found)
	assert.Equal(t, float64(0), balance)
}

func TestService_CreditUserWallet_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.CreditUserWallet(context.Background(), testUUID(), 100.0, "ref_123")
	assert.Error(t, err)
}

func TestService_CreditAgentWallet_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.CreditAgentWallet(context.Background(), "agent_123", 100.0, "ref_123", nil)
	assert.Error(t, err)
}

func TestService_GetOrCreateUserWallet_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.GetOrCreateUserWallet(context.Background(), testUUID())
	assert.Error(t, err)
}

func TestService_GetOrCreateAgentWallet_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.GetOrCreateAgentWallet(context.Background(), "agent_123")
	assert.Error(t, err)
}

func TestService_GetWallet_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.GetWallet(context.Background(), testUUID())
	assert.Error(t, err)
}

func TestService_GetWalletByOwner_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.GetWalletByOwner(context.Background(), OwnerTypeUser, "user_123")
	assert.Error(t, err)
}

func TestService_DebitForFeePayment_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.DebitForFeePayment(context.Background(), testUUID(), 10.0, "publish", "test_desc")
	assert.Error(t, err)
}

func TestService_DebitForExecution_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.DebitForExecution(context.Background(), testUUID(), 0.5, testUUID(), testUUID())
	assert.Error(t, err)
}

func TestService_ConsumeUserOrAgentCredits_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.ConsumeUserOrAgentCredits(context.Background(), "user_123", 1.0)
	assert.Error(t, err)
}

func TestService_ConsumeAgentCredits_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.ConsumeAgentCredits(context.Background(), "agent_123", 1.0)
	assert.Error(t, err)
}

func TestService_AdminCredit_ZeroAmount(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.AdminCredit(context.Background(), testUUID(), 0, "ref", "reason", testUUID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestService_AdminCredit_NegativeAmount(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.AdminCredit(context.Background(), testUUID(), -50.0, "ref", "reason", testUUID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestService_AdminDebit_ZeroAmount(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.AdminDebit(context.Background(), testUUID(), 0, "ref", "reason", mustParseUUID("12345678-1234-1234-123456789012"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestService_AdminDebit_NegativeAmount(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.AdminDebit(context.Background(), testUUID(), -50.0, "ref", "reason", mustParseUUID("12345678-1234-1234-123456789012"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestWalletFilter_Defaults(t *testing.T) {
	filter := WalletFilter{}

	assert.Equal(t, 0, filter.Limit)
	assert.Equal(t, 0, filter.Offset)
	assert.Empty(t, filter.OwnerType)
	assert.Empty(t, filter.Status)
}

func TestListWallets_InvalidLimits(t *testing.T) {
	svc := NewService(nil, nil)

	tests := []struct {
		name   string
		filter WalletFilter
	}{
		{"zero limit", WalletFilter{Limit: 0}},
		{"negative limit", WalletFilter{Limit: -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ListWallets(context.Background(), tt.filter)
			assert.Error(t, err)
		})
	}
}

func TestService_SuspendWallet_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	err := svc.SuspendWallet(context.Background(), testUUID(), "test reason")
	assert.Error(t, err)
}

func TestService_CloseWallet_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	err := svc.CloseWallet(context.Background(), testUUID(), "test reason")
	assert.Error(t, err)
}

func TestService_ReactivateWallet_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	err := svc.ReactivateWallet(context.Background(), testUUID())
	assert.Error(t, err)
}

func TestService_GetLowBalanceWallets_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.GetLowBalanceWallets(context.Background(), 10.0, "user")
	assert.Error(t, err)
}

func TestService_CheckSpendCap_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.CheckSpendCap(context.Background(), testUUID(), 5.0)
	assert.Error(t, err)
}

func TestService_UpdateSpendCaps_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	dailyCap := 100.0
	err := svc.UpdateSpendCaps(context.Background(), testUUID(), &dailyCap, nil)
	assert.Error(t, err)
}

func TestService_GetTransactionHistory_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, _, err := svc.GetTransactionHistory(context.Background(), testUUID(), 20, 0)
	assert.Error(t, err)
}

func TestService_GetWalletSummary_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.GetWalletSummary(context.Background(), testUUID())
	assert.Error(t, err)
}

func TestService_GetBalanceHistory_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.GetBalanceHistory(context.Background(), BalanceHistoryQuery{})
	assert.Error(t, err)
}

func TestService_CheckAgentSpendCap_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	allowed, err := svc.CheckAgentSpendCap(context.Background(), "agent_123", 10.0)
	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestService_GetAgentSpendSummary_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.GetAgentSpendSummary(context.Background(), "agent_123", "daily")
	assert.Error(t, err)
}

func TestService_UpdateAgentSpendCaps_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	dailyCap := 100.0
	err := svc.UpdateAgentSpendCaps(context.Background(), "agent_123", &dailyCap, nil)
	assert.Error(t, err)
}

func TestService_HasUserWalletCreditReference_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.HasUserWalletCreditReference(context.Background(), "ref_123")
	assert.Error(t, err)
}

func TestService_GetUserBalance_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.GetUserBalance(context.Background(), testUUID())
	assert.Error(t, err)
}

func TestService_GetAgentBalance_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.GetAgentBalance(context.Background(), "agent_123")
	assert.Error(t, err)
}

func TestService_CreditWalletUser_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	err := svc.CreditWalletUser(context.Background(), testUUID(), 100.0, "ref_123")
	assert.Error(t, err)
}

func TestService_DebitWalletUser_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	err := svc.DebitWalletUser(context.Background(), testUUID(), 10.0, "test desc")
	assert.Error(t, err)
}

func TestService_AddCreditsToAgent_NilRepo(t *testing.T) {
	svc := NewService(nil, nil)
	err := svc.AddCreditsToAgent(context.Background(), "agent_123", 100.0)
	assert.Error(t, err)
}

func testUUID() uuid.UUID {
	return uuid.MustParse("12345678-1234-1234-1234-123456789012")
}

func mustParseUUID(s string) uuid.UUID {
	return uuid.MustParse(s)
}
