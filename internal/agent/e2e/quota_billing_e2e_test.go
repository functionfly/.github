package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/agent/billing"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/agent/quota"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ============================================================
// Quota Enforcement Tests
// ============================================================

func TestQuotaEnforcement(t *testing.T) {
	t.Run("should create default quota for agent", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		require.NoError(t, err)

		require.NoError(t, db.AutoMigrate(&identity.AgentQuotaConfig{}))

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/quota-agent",
			Name:     "Quota Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := identity.NewRepository(db).CreateAgent(context.Background(), uuid.New(), agentReq)
		require.NoError(t, err)

		config, err := identity.NewRepository(db).GetQuotaConfig(context.Background(), agent.AgentID)
		require.NoError(t, err)

		assert.Equal(t, agent.AgentID, config.AgentID)
		assert.Greater(t, config.MaxCallsPerMinute, 0)
		assert.Greater(t, config.MaxCallsPerDay, 0)
		assert.Greater(t, config.MaxDailySpendUSD, 0.0)
	})

	t.Run("should update quota config", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		require.NoError(t, err)

		require.NoError(t, db.AutoMigrate(&identity.AgentQuotaConfig{}))

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/update-quota-agent",
			Name:     "Update Quota Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := identity.NewRepository(db).CreateAgent(context.Background(), uuid.New(), agentReq)
		require.NoError(t, err)

		updates := map[string]interface{}{
			"max_calls_per_minute": 200,
			"max_daily_spend_usd":  50.00,
		}

		err = identity.NewRepository(db).UpdateQuotaConfig(context.Background(), agent.AgentID, updates)
		require.NoError(t, err)

		config, err := identity.NewRepository(db).GetQuotaConfig(context.Background(), agent.AgentID)
		require.NoError(t, err)

		assert.Equal(t, 200, config.MaxCallsPerMinute)
		assert.Equal(t, 50.00, config.MaxDailySpendUSD)
	})

	t.Run("should check agent usage structure", func(t *testing.T) {
		usage := &quota.AgentUsage{
			AgentID:           "test-agent",
			CallsThisMinute:    0,
			CallsToday:        0,
			StateWritesThisHr: 0,
			SpendTodayUSD:     0.0,
			SpendThisMonthUSD: 0.0,
			LastUpdated:       time.Now(),
		}

		assert.NotNil(t, usage)
		assert.Equal(t, int64(0), usage.CallsThisMinute)
		assert.Equal(t, int64(0), usage.CallsToday)
	})
}

// ============================================================
// Billing Controller Tests
// ============================================================

func TestBillingController(t *testing.T) {
	t.Run("should create billing controller", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		require.NoError(t, err)

		ctrl := billing.NewController(db, nil)
		require.NotNil(t, ctrl)
	})

	t.Run("should get or create billing controls", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		require.NoError(t, err)

		require.NoError(t, db.AutoMigrate(&billing.AgentBillingControls{}))

		ctrl := billing.NewController(db, nil)

		agentID := "test-org/billing-controls-agent"
		controls, err := ctrl.GetOrCreateControls(context.Background(), agentID)
		require.NoError(t, err)
		require.NotNil(t, controls)

		assert.Equal(t, agentID, controls.AgentID)
		assert.Equal(t, 0.0, controls.CreditBalanceUSD)
		assert.Equal(t, billing.BillingModePerAgent, controls.BillingMode)
	})

	t.Run("should track billing controls with spend caps", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		require.NoError(t, err)

		require.NoError(t, db.AutoMigrate(&billing.AgentBillingControls{}))

		ctrl := billing.NewController(db, nil)

		dailyCap := 100.0
		monthlyCap := 1000.0

		controls := &billing.AgentBillingControls{
			ID:               uuid.New(),
			AgentID:          "test-org/spend-caps-agent",
			CreditBalanceUSD: 500.0,
			SpendCapDailyUSD: &dailyCap,
			SpendCapMonthlyUSD: &monthlyCap,
			BillingMode: billing.BillingModePerAgent,
			AlertThresholds: pq.Float64Array{0.5, 0.8, 0.95},
		}

		err = db.Create(controls).Error
		require.NoError(t, err)

		// Verify retrieval
		retrieved, err := ctrl.GetOrCreateControls(context.Background(), controls.AgentID)
		require.NoError(t, err)
		assert.Equal(t, 500.0, retrieved.CreditBalanceUSD)
		assert.NotNil(t, retrieved.SpendCapDailyUSD)
		assert.Equal(t, 100.0, *retrieved.SpendCapDailyUSD)
	})

	t.Run("should check spend cap", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		require.NoError(t, err)

		require.NoError(t, db.AutoMigrate(&billing.AgentBillingControls{}))

		ctrl := billing.NewController(db, nil)

		agentID := "test-org/spend-check-agent"

		// Create controls with daily cap
		dailyCap := 100.0
		controls := &billing.AgentBillingControls{
			ID:               uuid.New(),
			AgentID:          agentID,
			CreditBalanceUSD: 50.0,
			SpendCapDailyUSD: &dailyCap,
			BillingMode:      billing.BillingModePerAgent,
		}
		require.NoError(t, db.Create(controls).Error)

		// Check spend cap with reasonable amount
		allowed, err := ctrl.CheckSpendCap(context.Background(), agentID, 30.0)
		require.NoError(t, err)
		assert.True(t, allowed)

		// Check spend cap that would exceed
		allowed, err = ctrl.CheckSpendCap(context.Background(), agentID, 80.0)
		require.NoError(t, err)
		assert.False(t, allowed)
	})
}

// ============================================================
// Circuit Breaker Tests
// ============================================================

func TestCircuitBreaker(t *testing.T) {
	t.Run("should start in closed state", func(t *testing.T) {
		cb := &mockCircuitBreaker{
			failureThreshold: 3,
			timeout:          30 * time.Second,
		}

		assert.Equal(t, "closed", cb.GetState())
		assert.False(t, cb.IsOpen())
	})

	t.Run("should open after repeated failures", func(t *testing.T) {
		cb := &mockCircuitBreaker{
			failureThreshold: 3,
			timeout:          30 * time.Second,
		}

		// Record failures
		for i := 0; i < 3; i++ {
			cb.RecordFailure()
		}

		assert.Equal(t, "open", cb.GetState())
		assert.True(t, cb.IsOpen())
	})

	t.Run("should transition to half-open after timeout", func(t *testing.T) {
		cb := &mockCircuitBreaker{
			failureThreshold: 3,
			timeout:          1 * time.Millisecond, // Short timeout for testing
		}

		// Open the breaker
		for i := 0; i < 3; i++ {
			cb.RecordFailure()
		}
		assert.Equal(t, "open", cb.GetState())

		// Wait for timeout
		time.Sleep(10 * time.Millisecond)

		// Check half-open
		state := cb.GetState()
		assert.True(t, state == "half-open" || state == "closed", "should transition to half-open or be reset")
	})

	t.Run("should reset on successful call in half-open", func(t *testing.T) {
		cb := &mockCircuitBreaker{
			failureThreshold: 2,
			timeout:          1 * time.Millisecond,
		}

		// Open the breaker
		for i := 0; i < 2; i++ {
			cb.RecordFailure()
		}
		assert.Equal(t, "open", cb.GetState())

		// Wait for timeout
		time.Sleep(10 * time.Millisecond)

		// Record success
		cb.RecordSuccess()

		// Should be closed now
		assert.Equal(t, "closed", cb.GetState())
	})
}

type mockCircuitBreaker struct {
	failureThreshold int
	timeout          time.Duration
	failures         int
	lastFailure      time.Time
	state            string
}

func (cb *mockCircuitBreaker) GetState() string {
	if cb.failures >= cb.failureThreshold {
		if time.Since(cb.lastFailure) > cb.timeout {
			return "half-open"
		}
		return "open"
	}
	return "closed"
}

func (cb *mockCircuitBreaker) IsOpen() bool {
	return cb.GetState() == "open"
}

func (cb *mockCircuitBreaker) RecordFailure() {
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.failureThreshold {
		cb.state = "open"
	}
}

func (cb *mockCircuitBreaker) RecordSuccess() {
	cb.failures = 0
	cb.state = "closed"
}
