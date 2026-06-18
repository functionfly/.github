package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BillingRepositoryTestSuite struct {
	test.TestSuite
	db *PostgresDB
}

func TestBillingRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(BillingRepositoryTestSuite))
}

func (s *BillingRepositoryTestSuite) SetupTest() {
	s.TestSuite.SetupTest()

	db, err := s.setupTestDatabase()
	require.NoError(s.T(), err)
	s.db = db
	s.SetDB(db)

	err = RunMigrations(db)
	require.NoError(s.T(), err)
}

func (s *BillingRepositoryTestSuite) setupTestDatabase() (*PostgresDB, error) {
	os.Setenv("DB_HOST", getEnvOrDefault("TEST_DB_HOST", "localhost"))
	os.Setenv("DB_PORT", getEnvOrDefault("TEST_DB_PORT", "5434"))
	os.Setenv("DB_USER", getEnvOrDefault("TEST_DB_USER", "postgres"))
	os.Setenv("DB_PASSWORD", getEnvOrDefault("TEST_DB_PASSWORD", "postgres"))
	os.Setenv("DB_NAME", getEnvOrDefault("TEST_DB_NAME", "functionfly_test"))
	os.Setenv("DB_SSLMODE", "disable")

	os.Setenv("DB_MAX_OPEN_CONNS", "5")
	os.Setenv("DB_MAX_IDLE_CONNS", "2")
	os.Setenv("DB_CONN_MAX_LIFETIME", "5m")
	os.Setenv("DB_CONN_MAX_IDLE_TIME", "1m")

	return NewPostgresDB()
}

func (s *BillingRepositoryTestSuite) TearDownTest() {
	if s.db != nil {
		s.CleanupTestData()
	}
}

func (s *BillingRepositoryTestSuite) CleanupTestData() {
	tables := []string{
		"subscriptions",
		"pricing_tiers",
		"invoices",
		"coupons",
		"legal_holds",
		"tenants",
	}
	for _, table := range tables {
		_, err := s.db.DB.Exec("TRUNCATE TABLE " + table + " CASCADE")
		if err != nil {
			s.T().Logf("Failed to truncate table %s: %v", table, err)
		}
	}
}

func (s *BillingRepositoryTestSuite) createTestTenant() (*Tenant, error) {
	repo := s.db.Repository()
	ctx := context.Background()
	return repo.CreateTenant(ctx, "test-tenant-"+uuid.New().String())
}

func (s *BillingRepositoryTestSuite) createPricingTier(name string, priceCents int, currency string) (*PricingTier, error) {
	query := `
		INSERT INTO pricing_tiers (id, name, description, price_cents, currency, features, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, true, $7, $8)
		RETURNING id, name, description, price_cents, currency, features, is_active, created_at, updated_at`

	tier := &PricingTier{
		ID:          uuid.New(),
		Name:        name,
		Description: "Test tier",
		PriceCents:  priceCents,
		Currency:    currency,
		Features:    []string{"feature1", "feature2"},
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	featuresJSON := `["feature1", "feature2"]`
	_, err := s.db.DB.Exec(query, tier.ID, tier.Name, tier.Description, tier.PriceCents, tier.Currency, featuresJSON, tier.CreatedAt, tier.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return tier, nil
}

func (s *BillingRepositoryTestSuite) TestHasActiveLegalHolds_NoHolds() {
	repo := NewBillingRepository(s.db)
	ctx := context.Background()

	hasHolds, err := repo.HasActiveLegalHolds(ctx)
	require.NoError(s.T(), err)
	assert.False(s.T(), hasHolds)
}

func (s *BillingRepositoryTestSuite) TestHasActiveLegalHolds_WithActiveHold() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	_, err = s.db.DB.Exec(`
		INSERT INTO legal_holds (id, tenant_id, status, reason, created_at, created_by)
		VALUES ($1, $2, 'active', 'investigation', $3, $4)`,
		uuid.New(), tenant.ID, time.Now(), uuid.New())
	require.NoError(s.T(), err)

	repo := NewBillingRepository(s.db)
	ctx := context.Background()

	hasHolds, err := repo.HasActiveLegalHolds(ctx)
	require.NoError(s.T(), err)
	assert.True(s.T(), hasHolds)
}

func (s *BillingRepositoryTestSuite) TestHasActiveLegalHolds_WithExpiredHold() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	_, err = s.db.DB.Exec(`
		INSERT INTO legal_holds (id, tenant_id, status, reason, expires_at, created_at, created_by)
		VALUES ($1, $2, 'active', 'investigation', $3, $4, $5)`,
		uuid.New(), tenant.ID, time.Now().Add(-24*time.Hour), time.Now(), uuid.New())
	require.NoError(s.T(), err)

	repo := NewBillingRepository(s.db)
	ctx := context.Background()

	hasHolds, err := repo.HasActiveLegalHolds(ctx)
	require.NoError(s.T(), err)
	assert.False(s.T(), hasHolds)
}

func (s *BillingRepositoryTestSuite) TestCreateSubscription() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	tier, err := s.createPricingTier("Pro", 2900, "USD")
	require.NoError(s.T(), err)

	repo := NewBillingRepository(s.db)
	ctx := context.Background()

	now := time.Now()
	sub := &Subscription{
		TenantID:            tenant.ID,
		PricingTierID:       tier.ID,
		CurrentPeriodStart:  now,
		CurrentPeriodEnd:    now.Add(30 * 24 * time.Hour),
	}

	created, err := repo.CreateSubscription(ctx, sub)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), created.ID)
	assert.Equal(s.T(), tenant.ID, created.TenantID)
	assert.Equal(s.T(), "active", created.Status)
}

func (s *BillingRepositoryTestSuite) TestGetSubscriptionByTenantID() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	tier, err := s.createPricingTier("Basic", 900, "USD")
	require.NoError(s.T(), err)

	repo := NewBillingRepository(s.db)
	ctx := context.Background()

	now := time.Now()
	sub := &Subscription{
		TenantID:            tenant.ID,
		PricingTierID:       tier.ID,
		CurrentPeriodStart:  now,
		CurrentPeriodEnd:    now.Add(30 * 24 * time.Hour),
	}

	_, err = repo.CreateSubscription(ctx, sub)
	require.NoError(s.T(), err)

	retrieved, err := repo.GetSubscriptionByTenantID(ctx, tenant.ID)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), retrieved)
	assert.Equal(s.T(), tenant.ID, retrieved.TenantID)
	assert.Equal(s.T(), "Basic", retrieved.PricingTier.Name)
}

func (s *BillingRepositoryTestSuite) TestGetSubscriptionByTenantIDNotFound() {
	repo := NewBillingRepository(s.db)
	ctx := context.Background()

	retrieved, err := repo.GetSubscriptionByTenantID(ctx, uuid.New())
	require.NoError(s.T(), err)
	assert.Nil(s.T(), retrieved)
}

func (s *BillingRepositoryTestSuite) TestGetSubscriptionByID() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	tier, err := s.createPricingTier("Enterprise", 9900, "USD")
	require.NoError(s.T(), err)

	repo := NewBillingRepository(s.db)
	ctx := context.Background()

	now := time.Now()
	sub := &Subscription{
		TenantID:            tenant.ID,
		PricingTierID:       tier.ID,
		CurrentPeriodStart:  now,
		CurrentPeriodEnd:    now.Add(30 * 24 * time.Hour),
	}

	created, err := repo.CreateSubscription(ctx, sub)
	require.NoError(s.T(), err)

	retrieved, err := repo.GetSubscriptionByID(ctx, created.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), created.ID, retrieved.ID)
	assert.Equal(s.T(), "Enterprise", retrieved.PricingTier.Name)
}

func (s *BillingRepositoryTestSuite) TestGetSubscriptionByIDNotFound() {
	repo := NewBillingRepository(s.db)
	ctx := context.Background()

	retrieved, err := repo.GetSubscriptionByID(ctx, uuid.New())
	require.NoError(s.T(), err)
	assert.Nil(s.T(), retrieved)
}

func (s *BillingRepositoryTestSuite) TestUpdateSubscription() {
	tenant, err := s.createTestTenant()
	require.NoError(s.T(), err)

	tier, err := s.createPricingTier("Premium", 4900, "USD")
	require.NoError(s.T(), err)

	repo := NewBillingRepository(s.db)
	ctx := context.Background()

	now := time.Now()
	sub := &Subscription{
		TenantID:            tenant.ID,
		PricingTierID:       tier.ID,
		CurrentPeriodStart:  now,
		CurrentPeriodEnd:    now.Add(30 * 24 * time.Hour),
	}

	created, err := repo.CreateSubscription(ctx, sub)
	require.NoError(s.T(), err)

	updates := map[string]interface{}{
		"status": "past_due",
	}

	updated, err := repo.UpdateSubscription(ctx, created.ID, updates)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "past_due", updated.Status)
}

func (s *BillingRepositoryTestSuite) TestStringPtr() {
	result := StringPtr("test")
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), "test", *result)

	emptyResult := StringPtr("")
	assert.Nil(s.T(), emptyResult)
}

func (s *BillingRepositoryTestSuite) TestUUIDPtr() {
	id := uuid.New()
	result := UUIDPtr(id)
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), id, *result)
}

func (s *BillingRepositoryTestSuite) TestNormalizeInvoiceCurrency() {
	tests := []struct {
		input    string
		expected string
	}{
		{"usd", "USD"},
		{"EUR", "EUR"},
		{"gbp ", "GBP"},
		{"", "USD"},
		{"TOOLONG", "TOO"},
	}

	for _, tt := range tests {
		result := normalizeInvoiceCurrency(tt.input)
		assert.Equal(s.T(), tt.expected, result)
	}
}
