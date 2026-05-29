package wasm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStateFabricPath_KeyOnly(t *testing.T) {
	tenantID := uuid.New()
	fabricID := uuid.New()

	gotTenant, gotFabric, key, err := parseStateFabricPath("user123", tenantID, fabricID)
	require.NoError(t, err)
	assert.Equal(t, tenantID, gotTenant)
	assert.Equal(t, fabricID, gotFabric)
	assert.Equal(t, "user123", key)
}

func TestParseStateFabricPath_FabricKey(t *testing.T) {
	tenantID := uuid.New()
	fabricID := uuid.New()

	gotTenant, gotFabric, key, err := parseStateFabricPath(fabricID.String()+"/cart", tenantID, fabricID)
	require.NoError(t, err)
	assert.Equal(t, tenantID, gotTenant)
	assert.Equal(t, fabricID, gotFabric)
	assert.Equal(t, "cart", key)
}

func TestParseStateFabricPath_FullPath(t *testing.T) {
	tenantID := uuid.New()
	fabricID := uuid.New()

	path := tenantID.String() + "/" + fabricID.String() + "/orders"
	gotTenant, gotFabric, key, err := parseStateFabricPath(path, tenantID, fabricID)
	require.NoError(t, err)
	assert.Equal(t, tenantID, gotTenant)
	assert.Equal(t, fabricID, gotFabric)
	assert.Equal(t, "orders", key)
}

func TestParseStateFabricPath_CrossTenantDenied(t *testing.T) {
	executionTenant := uuid.New()
	otherTenant := uuid.New()
	fabricID := uuid.New()

	path := otherTenant.String() + "/" + fabricID.String() + "/secret"
	_, _, _, err := parseStateFabricPath(path, executionTenant, fabricID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cross-tenant")
}
