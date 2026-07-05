package provisioning

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestBundleProvisioner() (*BundleProvisioner, error) {
	cfg := &storage.TenantDatabaseConfig{
		Enabled: false,
	}
	dbProvisioner, err := storage.NewTenantDBProvisioner(cfg, nil)
	if err != nil {
		return nil, err
	}
	return NewBundleProvisioner(nil, nil, nil, dbProvisioner, nil), nil
}

func TestBundleProvisioner_ProvisionBundle_SaaSStarter(t *testing.T) {
	bp, err := newTestBundleProvisioner()
	require.NoError(t, err)

	tenantID := uuid.New()
	result, err := bp.ProvisionBundle(context.Background(), tenantID, "saas-starter")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, tenantID, result.TenantID)
	assert.Equal(t, "saas-starter", result.BundleSlug)
	assert.Equal(t, StatusActive, result.Status)
	assert.NotNil(t, result.Components)
	assert.NotEmpty(t, result.Components)

	expectedComponents := []string{"user_db", "auth", "payments", "email_workflows", "analytics"}
	for _, comp := range expectedComponents {
		assert.Contains(t, result.Components, comp, "Expected component %s to be provisioned", comp)
		compState := result.Components[comp]
		assert.Equal(t, StatusActive, compState.Status, "Component %s should be active", comp)
	}

	assert.Less(t, result.Duration, int64(5000), "Provisioning should complete quickly")
}

func TestBundleProvisioner_ProvisionBundle_Marketplace(t *testing.T) {
	bp, err := newTestBundleProvisioner()
	require.NoError(t, err)

	tenantID := uuid.New()
	result, err := bp.ProvisionBundle(context.Background(), tenantID, "marketplace")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, tenantID, result.TenantID)
	assert.Equal(t, "marketplace", result.BundleSlug)
	assert.Equal(t, StatusActive, result.Status)

	assert.Contains(t, result.Components, "marketplace", "Marketplace bundle should have marketplace component")
	assert.Equal(t, StatusActive, result.Components["marketplace"].Status)
}

func TestBundleProvisioner_ProvisionBundle_AIApp(t *testing.T) {
	bp, err := newTestBundleProvisioner()
	require.NoError(t, err)

	tenantID := uuid.New()
	result, err := bp.ProvisionBundle(context.Background(), tenantID, "ai-app")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, tenantID, result.TenantID)
	assert.Equal(t, "ai-app", result.BundleSlug)
	assert.Equal(t, StatusActive, result.Status)

	assert.Contains(t, result.Components, "ai_app", "AI App bundle should have ai_app component")
	assert.Equal(t, StatusActive, result.Components["ai_app"].Status)
}

func TestBundleProvisioner_ProvisionBundle_Isolation(t *testing.T) {
	bp, err := newTestBundleProvisioner()
	require.NoError(t, err)

	tenant1 := uuid.New()
	tenant2 := uuid.New()

	result1, err := bp.ProvisionBundle(context.Background(), tenant1, "saas-starter")
	require.NoError(t, err)

	result2, err := bp.ProvisionBundle(context.Background(), tenant2, "saas-starter")
	require.NoError(t, err)

	assert.NotEqual(t, result1.Components["auth"].ResourceID, result2.Components["auth"].ResourceID,
		"Each tenant should get isolated auth resources")

	assert.NotEqual(t, result1.Components["user_db"].ResourceID, result2.Components["user_db"].ResourceID,
		"Each tenant should get isolated databases")

	assert.NotEqual(t, result1.Components["payments"].ResourceID, result2.Components["payments"].ResourceID,
		"Each tenant should get isolated payment resources")
}

func TestBundleProvisioner_ProvisionBundle_Idempotency(t *testing.T) {
	bp, err := newTestBundleProvisioner()
	require.NoError(t, err)

	tenantID := uuid.New()

	result1, err := bp.ProvisionBundle(context.Background(), tenantID, "saas-starter")
	require.NoError(t, err)
	assert.Equal(t, StatusActive, result1.Status)

	result2, err := bp.ProvisionBundle(context.Background(), tenantID, "saas-starter")
	require.NoError(t, err)
	assert.Equal(t, StatusActive, result2.Status)

	assert.Equal(t, result1.Components["auth"].ResourceID, result2.Components["auth"].ResourceID,
		"Re-provisioning should return same resource ID (idempotent)")
}

func TestBundleProvisioner_TrackProvisioningStart(t *testing.T) {
	bp := &BundleProvisioner{
		platformDB: nil,
	}

	tenantID := uuid.New()
	err := bp.trackProvisioningStart(context.Background(), tenantID, "saas-starter")
	assert.NoError(t, err)
}

func TestBundleProvisioner_TrackProvisioningComplete(t *testing.T) {
	bp := &BundleProvisioner{
		platformDB: nil,
	}

	tenantID := uuid.New()
	result := &ProvisionResult{
		TenantID:   tenantID,
		BundleSlug: "saas-starter",
		Status:     StatusActive,
		Components: map[string]*ComponentState{
			"auth": {Status: StatusActive, ResourceID: "key_123"},
		},
		ErrorLog: []string{},
	}

	err := bp.trackProvisioningComplete(context.Background(), tenantID, result)
	assert.NoError(t, err)
}

func TestProvisionResult_JSON(t *testing.T) {
	tenantID := uuid.New()
	result := &ProvisionResult{
		TenantID:   tenantID,
		BundleSlug: "saas-starter",
		Status:     StatusActive,
		Components: map[string]*ComponentState{
			"user_db": {Status: StatusActive, ResourceID: "db_123", Timestamp: time.Now()},
			"auth":    {Status: StatusActive, ResourceID: "key_456", Timestamp: time.Now()},
		},
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
		Duration:   2000,
		ErrorLog:   []string{},
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded ProvisionResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.TenantID, decoded.TenantID)
	assert.Equal(t, result.BundleSlug, decoded.BundleSlug)
	assert.Equal(t, result.Status, decoded.Status)
	assert.Len(t, decoded.Components, 2)
}

func TestComponentState_JSON(t *testing.T) {
	state := &ComponentState{
		Status:     StatusActive,
		Timestamp:  time.Now(),
		ResourceID: "key_abc123",
		Error:      "",
	}

	data, err := json.Marshal(state)
	require.NoError(t, err)

	var decoded ComponentState
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, state.Status, decoded.Status)
	assert.Equal(t, state.ResourceID, decoded.ResourceID)
}

func TestBundleProvisioner_ComponentState_StatusTransitions(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  ProvisionerStatus
		expectedStatus ProvisionerStatus
	}{
		{"pending_to_provisioning", StatusPending, StatusProvisioning},
		{"provisioning_to_active", StatusProvisioning, StatusActive},
		{"provisioning_to_failed", StatusProvisioning, StatusFailed},
		{"failed_to_rolledback", StatusFailed, StatusRolledBack},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &ComponentState{Status: tt.initialStatus}
			state.Status = tt.expectedStatus
			assert.Equal(t, tt.expectedStatus, state.Status)
		})
	}
}

func TestBundleProvisioner_ProvisionResult_Duration(t *testing.T) {
	start := time.Now()
	end := start.Add(1500 * time.Millisecond)

	result := &ProvisionResult{
		StartedAt:  start,
		FinishedAt: end,
	}

	result.Duration = result.FinishedAt.Sub(result.StartedAt).Milliseconds()

	assert.Equal(t, int64(1500), result.Duration)
}

func TestNewBundleProvisioner(t *testing.T) {
	cfg := &storage.TenantDatabaseConfig{
		Enabled: false,
	}

	dbProvisioner, err := storage.NewTenantDBProvisioner(cfg, nil)
	require.NoError(t, err)

	bp := NewBundleProvisioner(nil, nil, nil, dbProvisioner, nil)
	require.NotNil(t, bp)
	assert.NotNil(t, bp.authProvisioner)
	assert.NotNil(t, bp.paymentsProvisioner)
	assert.NotNil(t, bp.emailWfProvisioner)
	assert.NotNil(t, bp.analyticsProvisioner)
	assert.NotNil(t, bp.userDBProvisioner)
	assert.NotNil(t, bp.marketplaceProvisioner)
	assert.NotNil(t, bp.aiAppProvisioner)
}

func TestProvisionerStatus_Constants(t *testing.T) {
	assert.Equal(t, ProvisionerStatus("pending"), StatusPending)
	assert.Equal(t, ProvisionerStatus("provisioning"), StatusProvisioning)
	assert.Equal(t, ProvisionerStatus("active"), StatusActive)
	assert.Equal(t, ProvisionerStatus("failed"), StatusFailed)
	assert.Equal(t, ProvisionerStatus("rolled_back"), StatusRolledBack)
}

func TestProvisionBundleForBilling(t *testing.T) {
	bp, err := newTestBundleProvisioner()
	require.NoError(t, err)

	adapter := ProvisionBundleForBilling(bp)
	require.NotNil(t, adapter)

	tenantID := uuid.New()
	status, count, err := adapter(context.Background(), tenantID, "saas-starter")
	require.NoError(t, err)
	assert.Equal(t, "active", status)
	assert.Greater(t, count, 0)
}

func TestBundleProvisioner_UnknownBundleSlug(t *testing.T) {
	bp, err := newTestBundleProvisioner()
	require.NoError(t, err)

	tenantID := uuid.New()
	result, err := bp.ProvisionBundle(context.Background(), tenantID, "unknown-bundle")
	require.NoError(t, err)
	assert.Equal(t, StatusActive, result.Status)
}
