package billing

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDunningManager_IsStripeConfigured(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		expected bool
	}{
		{"configured", "sk_test_123", true},
		{"not configured", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				os.Setenv("STRIPE_SECRET_KEY", tt.envKey)
				defer os.Unsetenv("STRIPE_SECRET_KEY")
			}

			db, _, _ := sqlmock.New()
			defer db.Close()

			postgresDB := &storage.PostgresDB{GORM: nil, DB: db}
			dunningRepo := storage.NewDunningRepository(postgresDB)
			userRepo := &mockUserRepo{}

			mgr := NewDunningManager(dunningRepo, userRepo, nil)
			assert.Equal(t, tt.expected, mgr.IsStripeConfigured())
		})
	}
}

func TestDunningManager_InitiateDunningWorkflow(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postgresDB := &storage.PostgresDB{GORM: nil, DB: db}
	dunningRepo := storage.NewDunningRepository(postgresDB)
	userRepo := &mockUserRepo{}
	notificationSvc := &mockNotificationSvc{}

	mgr := NewDunningManager(dunningRepo, userRepo, notificationSvc)

	scheduleRows := sqlmock.NewRows([]string{"id", "schedule_type", "max_retries", "retry_intervals", "grace_period_days", "send_customer_notifications", "notify_admin_on_final_retry", "suspend_service_after_final_retry"}).
		AddRow(uuid.New(), "default", 4, "1,3,7,14", 14, true, true, true)

	mock.ExpectQuery(`SELECT .+ FROM payment_retry_schedules WHERE schedule_type = \$1`).
		WithArgs("default").
		WillReturnRows(scheduleRows)

	existingRows := sqlmock.NewRows([]string{"id", "tenant_id"})
	mock.ExpectQuery(`SELECT .+ FROM payment_retries WHERE invoice_id = \$1`).
		WithArgs("inv_123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id"}))

	mock.ExpectQuery(`INSERT INTO payment_retries`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(nil)

	params := DunningInitiationParams{
		TenantID:         uuid.New(),
		InvoiceID:        "inv_123",
		StripeCustomerID: "cus_123",
		CustomerEmail:    "test@example.com",
		AmountDueCents:   5000,
		Currency:         "USD",
		FailureCode:      "card_declined",
		FailureMessage:   "Your card was declined",
		DeclineCode:      "generic_decline",
		Metadata:         json.RawMessage(`{}`),
	}

	ctx := context.Background()
	retry, err := mgr.InitiateDunningWorkflow(ctx, params)
	require.NoError(t, err)
	assert.NotNil(t, retry)
	assert.Equal(t, "active", retry.Status)
	assert.Equal(t, 0, retry.CurrentAttempt)
	assert.Equal(t, 4, retry.MaxAttempts)
}

func TestDunningManager_InitiateDunningWorkflow_ExistingRetry(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postgresDB := &storage.PostgresDB{GORM: nil, DB: db}
	dunningRepo := storage.NewDunningRepository(postgresDB)
	userRepo := &mockUserRepo{}

	mgr := NewDunningManager(dunningRepo, userRepo, nil)

	scheduleRows := sqlmock.NewRows([]string{"id", "schedule_type", "max_retries", "retry_intervals", "grace_period_days", "send_customer_notifications", "notify_admin_on_final_retry", "suspend_service_after_final_retry"}).
		AddRow(uuid.New(), "default", 4, "1,3,7,14", 14, true, true, true)

	mock.ExpectQuery(`SELECT .+ FROM payment_retry_schedules WHERE schedule_type = \$1`).
		WithArgs("default").
		WillReturnRows(scheduleRows)

	existingRetry := &storage.PaymentRetry{
		ID:     uuid.New(),
		Status: "active",
	}
	existingRows := sqlmock.NewRows([]string{"id", "tenant_id", "invoice_id", "status"}).
		AddRow(existingRetry.ID, existingRetry.TenantID, "inv_123", existingRetry.Status)

	mock.ExpectQuery(`SELECT .+ FROM payment_retries WHERE invoice_id = \$1`).
		WithArgs("inv_123").
		WillReturnRows(existingRows)

	params := DunningInitiationParams{
		TenantID:         uuid.New(),
		InvoiceID:        "inv_123",
		StripeCustomerID: "cus_123",
		CustomerEmail:    "test@example.com",
		AmountDueCents:   5000,
		Currency:         "USD",
	}

	ctx := context.Background()
	retry, err := mgr.InitiateDunningWorkflow(ctx, params)
	require.NoError(t, err)
	assert.Equal(t, existingRetry.ID, retry.ID)
	assert.Equal(t, "active", retry.Status)
}

func TestDunningManager_ProcessGracePeriodExpirations(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postgresDB := &storage.PostgresDB{GORM: nil, DB: db}
	dunningRepo := storage.NewDunningRepository(postgresDB)
	userRepo := &mockUserRepo{}
	notificationSvc := &mockNotificationSvc{}

	mgr := NewDunningManager(dunningRepo, userRepo, notificationSvc)

	tenantID := uuid.New()
	retryID := uuid.New()

	retriesRows := sqlmock.NewRows([]string{"id", "tenant_id", "status", "grace_period_ends_at"}).
		AddRow(retryID, tenantID, "active", time.Now().Add(-24*time.Hour))

	mock.ExpectQuery(`SELECT .+ FROM payment_retries WHERE grace_period_ends_at < NOW\(\) AND status = 'active'`).
		WillReturnRows(retriesRows)

	mock.ExpectQuery(`SELECT .+ FROM service_suspensions WHERE tenant_id = \$1 AND suspended_at IS NOT NULL AND restored_at IS NULL`).
		WithArgs(tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id"}).AddRow(uuid.New(), tenantID))

	mock.ExpectQuery(`UPDATE tenants SET status = 'suspended'`).
		WithArgs("suspended", tenantID).
		WillReturnError(nil)

	ctx := context.Background()
	err = mgr.ProcessGracePeriodExpirations(ctx)
	require.NoError(t, err)
}

func TestDunningManager_SuspendService(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postgresDB := &storage.PostgresDB{GORM: nil, DB: db}
	dunningRepo := storage.NewDunningRepository(postgresDB)
	userRepo := &mockUserRepo{}
	notificationSvc := &mockNotificationSvc{}

	mgr := NewDunningManager(dunningRepo, userRepo, notificationSvc)

	tenantID := uuid.New()
	retryID := uuid.New()

	suspendedRows := sqlmock.NewRows([]string{"id", "tenant_id"}).AddRow(uuid.New(), tenantID)
	mock.ExpectQuery(`SELECT .+ FROM service_suspensions WHERE tenant_id = \$1 AND suspended_at IS NOT NULL AND restored_at IS NULL`).
		WithArgs(tenantID).
		WillReturnRows(suspendedRows)

	mock.ExpectQuery(`INSERT INTO service_suspensions`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(nil)

	mock.ExpectQuery(`UPDATE tenants SET status = 'suspended'`).
		WithArgs("suspended", tenantID).
		WillReturnError(nil)

	ctx := context.Background()
	err = mgr.SuspendService(ctx, retryID, tenantID, "grace_period_expired")
	require.NoError(t, err)
}

func TestDunningManager_SuspendService_AlreadySuspended(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postgresDB := &storage.PostgresDB{GORM: nil, DB: db}
	dunningRepo := storage.NewDunningRepository(postgresDB)
	userRepo := &mockUserRepo{}

	mgr := NewDunningManager(dunningRepo, userRepo, nil)

	tenantID := uuid.New()
	retryID := uuid.New()

	existingSuspensionRows := sqlmock.NewRows([]string{"id", "tenant_id", "suspended_at"}).
		AddRow(uuid.New(), tenantID, time.Now())

	mock.ExpectQuery(`SELECT .+ FROM service_suspensions WHERE tenant_id = \$1 AND suspended_at IS NOT NULL AND restored_at IS NULL`).
		WithArgs(tenantID).
		WillReturnRows(existingSuspensionRows)

	ctx := context.Background()
	err = mgr.SuspendService(ctx, retryID, tenantID, "grace_period_expired")
	require.NoError(t, err)
}

func TestDunningManager_RestoreService(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postgresDB := &storage.PostgresDB{GORM: nil, DB: db}
	dunningRepo := storage.NewDunningRepository(postgresDB)
	userRepo := &mockUserRepo{}
	notificationSvc := &mockNotificationSvc{}

	mgr := NewDunningManager(dunningRepo, userRepo, notificationSvc)

	tenantID := uuid.New()
	suspensionID := uuid.New()

	suspensionRows := sqlmock.NewRows([]string{"id", "tenant_id", "suspended_at"}).
		AddRow(suspensionID, tenantID, time.Now())

	mock.ExpectQuery(`SELECT .+ FROM service_suspensions WHERE tenant_id = \$1 AND suspended_at IS NOT NULL AND restored_at IS NULL`).
		WithArgs(tenantID).
		WillReturnRows(suspensionRows)

	mock.ExpectQuery(`UPDATE service_suspensions SET restored_at`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), suspensionID).
		WillReturnError(nil)

	mock.ExpectQuery(`UPDATE tenants SET status = 'active'`).
		WithArgs("active", tenantID).
		WillReturnError(nil)

	ctx := context.Background()
	err = mgr.RestoreService(ctx, tenantID, "admin", "payment_received")
	require.NoError(t, err)
}

func TestDunningManager_IsTenantSuspended(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postgresDB := &storage.PostgresDB{GORM: nil, DB: db}
	dunningRepo := storage.NewDunningRepository(postgresDB)
	userRepo := &mockUserRepo{}

	mgr := NewDunningManager(dunningRepo, userRepo, nil)

	tenantID := uuid.New()

	suspendedRows := sqlmock.NewRows([]string{"id", "tenant_id"}).
		AddRow(uuid.New(), tenantID)

	mock.ExpectQuery(`SELECT .+ FROM service_suspensions WHERE tenant_id = \$1 AND suspended_at IS NOT NULL AND restored_at IS NULL`).
		WithArgs(tenantID).
		WillReturnRows(suspendedRows)

	ctx := context.Background()
	suspended, err := mgr.IsTenantSuspended(ctx, tenantID)
	require.NoError(t, err)
	assert.True(t, suspended)
}

func TestDunningManager_IsTenantSuspended_NotSuspended(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postgresDB := &storage.PostgresDB{GORM: nil, DB: db}
	dunningRepo := storage.NewDunningRepository(postgresDB)
	userRepo := &mockUserRepo{}

	mgr := NewDunningManager(dunningRepo, userRepo, nil)

	tenantID := uuid.New()

	mock.ExpectQuery(`SELECT .+ FROM service_suspensions WHERE tenant_id = \$1 AND suspended_at IS NOT NULL AND restored_at IS NULL`).
		WithArgs(tenantID).
		WillReturnError(nil)

	ctx := context.Background()
	suspended, err := mgr.IsTenantSuspended(ctx, tenantID)
	require.NoError(t, err)
	assert.False(t, suspended)
}

func TestDunningManager_Stop(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postgresDB := &storage.PostgresDB{GORM: nil, DB: db}
	dunningRepo := storage.NewDunningRepository(postgresDB)
	userRepo := &mockUserRepo{}

	mgr := NewDunningManager(dunningRepo, userRepo, nil)

	stopCh := mgr.StopChan()
	assert.NotNil(t, stopCh)

	mgr.Stop()

	select {
	case _, ok := <-stopCh:
		if ok {
			t.Error("stop channel should be closed")
		}
	default:
		t.Error("stop channel should be closed immediately")
	}
}

func TestDunningInitiationParams_Validation(t *testing.T) {
	tests := []struct {
		name   string
		params DunningInitiationParams
		valid  bool
	}{
		{
			name: "valid params",
			params: DunningInitiationParams{
				TenantID:         uuid.New(),
				InvoiceID:        "inv_123",
				StripeCustomerID: "cus_123",
				CustomerEmail:    "test@example.com",
				AmountDueCents:   5000,
				Currency:         "USD",
			},
			valid: true,
		},
		{
			name: "missing invoice id",
			params: DunningInitiationParams{
				TenantID:         uuid.New(),
				StripeCustomerID: "cus_123",
				CustomerEmail:    "test@example.com",
				AmountDueCents:   5000,
				Currency:         "USD",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasInvoice := tt.params.InvoiceID != ""
			if tt.valid {
				assert.True(t, hasInvoice)
			}
		})
	}
}

func TestNotificationHelpers(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postgresDB := &storage.PostgresDB{GORM: nil, DB: db}
	dunningRepo := storage.NewDunningRepository(postgresDB)
	userRepo := &mockUserRepo{}
	notificationSvc := &mockNotificationSvc{}

	mgr := NewDunningManager(dunningRepo, userRepo, notificationSvc)

	retry := &storage.PaymentRetry{
		ID:                uuid.New(),
		TenantID:          uuid.New(),
		AmountDueCents:    5000,
		Currency:          "USD",
		MaxAttempts:       4,
		GracePeriodEndsAt: time.Now().Add(14 * 24 * time.Hour),
	}

	ctx := context.Background()

	err = mgr.sendInitialFailureNotification(ctx, retry, "test@example.com")
	require.NoError(t, err)

	err = mgr.sendPaymentRecoveredNotification(ctx, retry)
	require.NoError(t, err)

	err = mgr.sendServiceSuspendedNotification(ctx, &storage.ServiceSuspension{
		ID:          uuid.New(),
		TenantID:    retry.TenantID,
		SuspendedAt: time.Now(),
	})
	require.NoError(t, err)

	err = mgr.sendServiceRestoredNotification(ctx, retry.TenantID, &storage.ServiceSuspension{
		ID:       uuid.New(),
		TenantID: retry.TenantID,
	})
	require.NoError(t, err)
}

func TestNotificationHelpers_NoOpWhenNilService(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postgresDB := &storage.PostgresDB{GORM: nil, DB: db}
	dunningRepo := storage.NewDunningRepository(postgresDB)
	userRepo := &mockUserRepo{}

	mgr := NewDunningManager(dunningRepo, userRepo, nil)

	retry := &storage.PaymentRetry{
		ID:             uuid.New(),
		TenantID:       uuid.New(),
		AmountDueCents: 5000,
		Currency:       "USD",
	}

	ctx := context.Background()

	err = mgr.sendInitialFailureNotification(ctx, retry, "test@example.com")
	require.NoError(t, err)

	err = mgr.sendRetryReminderNotification(ctx, retry, 1)
	require.NoError(t, err)

	err = mgr.sendFinalFailureNotification(ctx, retry)
	require.NoError(t, err)

	err = mgr.sendAdminFinalFailureNotification(ctx, retry)
	require.NoError(t, err)
}

func TestGetAdminEmails(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	postgresDB := &storage.PostgresDB{GORM: nil, DB: db}
	dunningRepo := storage.NewDunningRepository(postgresDB)
	userRepo := &mockUserRepo{}

	mgr := NewDunningManager(dunningRepo, userRepo, nil)

	os.Setenv("BILLING_ALERTS_EMAIL", "billing@example.com")
	defer os.Unsetenv("BILLING_ALERTS_EMAIL")

	os.Setenv("FINANCE_TEAM_EMAIL", "finance@example.com")
	defer os.Unsetenv("FINANCE_TEAM_EMAIL")

	ctx := context.Background()
	emails := mgr.getAdminEmails(ctx, uuid.New())

	assert.Contains(t, emails, "billing@example.com")
	assert.Contains(t, emails, "finance@example.com")
}

type mockUserRepo struct{}

func (m *mockUserRepo) UpdateTenantStatus(ctx context.Context, tenantID uuid.UUID, status string) error {
	return nil
}

func (m *mockUserRepo) GetBundleSubscriptionByTenant(ctx context.Context, tenantID uuid.UUID) (*storage.BundleSubscription, error) {
	return nil, nil
}

func (m *mockUserRepo) UpdateBundleSubscription(ctx context.Context, sub *storage.BundleSubscription) error {
	return nil
}

func (m *mockUserRepo) ListActiveUsersByTenant(ctx context.Context, tenantID uuid.UUID) ([]storage.User, error) {
	return []storage.User{
		{ID: uuid.New(), Email: "admin@example.com", Role: "owner"},
	}, nil
}

func (m *mockUserRepo) GetTenantByID(tenantID uuid.UUID) (*storage.Tenant, error) {
	return &storage.Tenant{ID: tenantID, Name: "Test Tenant"}, nil
}

func (m *mockUserRepo) GetTenantByDomain(domain string) (*storage.Tenant, error) {
	return nil, nil
}

func (m *mockUserRepo) CreateTenant(tenant *storage.Tenant) error {
	return nil
}

func (m *mockUserRepo) UpdateTenant(tenant *storage.Tenant) error {
	return nil
}

func (m *mockUserRepo) ListTenants(limit, offset int) ([]storage.Tenant, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepo) CreateUser(user *storage.User) error {
	return nil
}

func (m *mockUserRepo) UpdateUser(user *storage.User) error {
	return nil
}

func (m *mockUserRepo) GetUserByID(id uuid.UUID) (*storage.User, error) {
	return nil, nil
}

func (m *mockUserRepo) GetUserByEmail(email string) (*storage.User, error) {
	return nil, nil
}

func (m *mockUserRepo) ListUsersByTenant(tenantID uuid.UUID, limit, offset int) ([]storage.User, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepo) DeleteUser(id uuid.UUID) error {
	return nil
}

func (m *mockUserRepo) GetRepository() interface{} {
	return nil
}

func (m *mockUserRepo) LogAuditEvent(ctx context.Context, event *storage.AuditEvent) error {
	return nil
}

func (m *mockUserRepo) ListSecurityScans(limit, offset int, filters map[string]interface{}) ([]*storage.SecurityScan, error) {
	return nil, nil
}

func (m *mockUserRepo) GetVulnerabilities(filters map[string]interface{}) ([]*storage.Vulnerability, error) {
	return nil, nil
}

func (m *mockUserRepo) GetVulnerabilityByID(id uuid.UUID) (*storage.Vulnerability, error) {
	return nil, nil
}

func (m *mockUserRepo) UpdateVulnerability(id uuid.UUID, updates map[string]interface{}) (*storage.Vulnerability, error) {
	return nil, nil
}

type mockNotificationSvc struct{}

func (m *mockNotificationSvc) SendBillingAlert(ctx context.Context, email, alertType string, data map[string]interface{}) error {
	return nil
}

func (m *mockNotificationSvc) Send(ctx context.Context, req notification.SendRequest) (string, error) {
	return "", nil
}
