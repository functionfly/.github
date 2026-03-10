package apikey

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestAPIKey_ToResponse tests converting APIKey to response
func TestAPIKey_ToResponse(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	apiKey := &APIKey{
		ID:                    uuid.New(),
		TenantID:              uuid.New(),
		UserID:                uuid.New(),
		Name:                  "test-key",
		Description:           "A test key",
		KeyType:               KeyTypePlatform,
		KeyPrefix:             PrefixPlatform,
		KeyHash:               "somehash",
		KeyVersion:            1,
		ExpiresAt:             &expiresAt,
		LastRotatedAt:         now,
		RotationFrequencyDays: 90,
		RateLimitRPM:          1000,
		RateLimitRPH:          60000,
		RateLimitRPD:          1000000,
		IsActive:              true,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	resp := apiKey.ToResponse()

	assert.Equal(t, apiKey.ID, resp.ID)
	assert.Equal(t, apiKey.Name, resp.Name)
	assert.Equal(t, apiKey.Description, resp.Description)
	assert.Equal(t, apiKey.KeyType, resp.KeyType)
	assert.Equal(t, apiKey.KeyPrefix, resp.Prefix)
	assert.Equal(t, apiKey.ExpiresAt, resp.ExpiresAt)
	assert.Equal(t, apiKey.LastRotatedAt, resp.LastRotatedAt)
	assert.Equal(t, apiKey.RotationFrequencyDays, resp.RotationFrequencyDays)
	assert.Equal(t, apiKey.IsActive, resp.IsActive)
	assert.Equal(t, apiKey.CreatedAt, resp.CreatedAt)
	assert.NotNil(t, resp.RateLimit)
}

// TestAPIKey_ToResponse_NoExpiration tests converting APIKey without expiration
func TestAPIKey_ToResponse_NoExpiration(t *testing.T) {
	now := time.Now()

	apiKey := &APIKey{
		ID:                    uuid.New(),
		Name:                  "test-key",
		KeyType:               KeyTypePlatform,
		KeyPrefix:             PrefixPlatform,
		IsActive:              true,
		CreatedAt:             now,
		UpdatedAt:             now,
		LastRotatedAt:         now,
		RotationFrequencyDays: 90,
		RateLimitRPM:          1000,
		RateLimitRPH:          60000,
		RateLimitRPD:          1000000,
	}

	resp := apiKey.ToResponse()

	assert.Nil(t, resp.ExpiresAt)
	assert.True(t, resp.IsActive)
}

// TestAPIKey_GetRateLimitConfig tests getting rate limit config
func TestAPIKey_GetRateLimitConfig(t *testing.T) {
	apiKey := &APIKey{
		RateLimitRPM: 1000,
		RateLimitRPH: 60000,
		RateLimitRPD: 1000000,
	}

	cfg := apiKey.GetRateLimitConfig()

	assert.Equal(t, 1000, cfg.RPM)
	assert.Equal(t, 60000, cfg.RPH)
	assert.Equal(t, 1000000, cfg.RPD)
}

// TestAPIKey_SetRateLimitConfig tests setting rate limit config
func TestAPIKey_SetRateLimitConfig(t *testing.T) {
	apiKey := &APIKey{}

	cfg := &RateLimitConfig{
		RPM: 500,
		RPH: 30000,
		RPD: 500000,
	}

	apiKey.SetRateLimitConfig(cfg)

	assert.Equal(t, 500, apiKey.RateLimitRPM)
	assert.Equal(t, 30000, apiKey.RateLimitRPH)
	assert.Equal(t, 500000, apiKey.RateLimitRPD)
}

// TestAPIKey_SetRateLimitConfig_Nil tests setting rate limit config with nil
func TestAPIKey_SetRateLimitConfig_Nil(t *testing.T) {
	apiKey := &APIKey{
		RateLimitRPM: 1000,
		RateLimitRPH: 60000,
		RateLimitRPD: 1000000,
	}

	// Setting nil should not change values
	apiKey.SetRateLimitConfig(nil)

	assert.Equal(t, 1000, apiKey.RateLimitRPM)
	assert.Equal(t, 60000, apiKey.RateLimitRPH)
	assert.Equal(t, 1000000, apiKey.RateLimitRPD)
}

// TestGetPrefixForKeyType tests getting prefix for key type
func TestGetPrefixForKeyType(t *testing.T) {
	tests := []struct {
		keyType  KeyType
		expected string
	}{
		{KeyTypePlatform, PrefixPlatform},
		{KeyTypeFunction, PrefixFunction},
		{KeyTypeAgent, PrefixAgent},
		{KeyTypeEnvironment, PrefixEnvironment},
		{KeyTypeOAuth, PrefixOAuth},
		{KeyType("unknown"), PrefixPlatform}, // Default
	}

	for _, tt := range tests {
		result := GetPrefixForKeyType(tt.keyType)
		assert.Equal(t, tt.expected, result)
	}
}

// TestDefaultRateLimitConfig tests default rate limit config
func TestDefaultRateLimitConfig(t *testing.T) {
	cfg := DefaultRateLimitConfig()

	assert.Equal(t, DefaultRateLimitRPM, cfg.RPM)
	assert.Equal(t, DefaultRateLimitRPH, cfg.RPH)
	assert.Equal(t, DefaultRateLimitRPD, cfg.RPD)
}

// TestIsValidKeyType tests key type validation
func TestIsValidKeyType(t *testing.T) {
	tests := []struct {
		keyType  string
		expected bool
	}{
		{"platform", true},
		{"function", true},
		{"agent", true},
		{"environment", true},
		{"oauth", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsValidKeyType(tt.keyType)
		assert.Equal(t, tt.expected, result)
	}
}

// TestValidationError tests ValidationError
func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "name",
		Message: "is required",
	}

	assert.Equal(t, "name: is required", err.Error())
}

// TestValidationErrors tests ValidationErrors collection
func TestValidationErrors(t *testing.T) {
	errs := ValidationErrors{
		{Field: "name", Message: "is required"},
		{Field: "key_type", Message: "is required"},
	}

	assert.True(t, errs.HasErrors())
	assert.Contains(t, errs.Error(), "name: is required")
	assert.Contains(t, errs.Error(), "key_type: is required")
}

// TestValidationErrors_Empty tests empty ValidationErrors
func TestValidationErrors_Empty(t *testing.T) {
	errs := ValidationErrors{}

	assert.False(t, errs.HasErrors())
	assert.Equal(t, "validation errors:", errs.Error())
}

// TestListFilters tests ListFilters
func TestListFilters(t *testing.T) {
	tenantID := uuid.New()
	keyType := KeyTypePlatform
	active := true

	filters := &ListFilters{
		TenantID: &tenantID,
		KeyType:  &keyType,
		IsActive: &active,
		Search:   "test",
	}

	assert.NotNil(t, filters.TenantID)
	assert.NotNil(t, filters.KeyType)
	assert.NotNil(t, filters.IsActive)
	assert.Equal(t, "test", filters.Search)
}

// TestCreateAPIKeyRequest tests CreateAPIKeyRequest
func TestCreateAPIKeyRequest(t *testing.T) {
	req := &CreateAPIKeyRequest{
		Name:                  "test-key",
		Description:           "A test key",
		KeyType:               KeyTypePlatform,
		RotationFrequencyDays: 90,
		RateLimit:             DefaultRateLimitConfig(),
	}

	assert.Equal(t, "test-key", req.Name)
	assert.Equal(t, KeyTypePlatform, req.KeyType)
}

// TestRotateAPIKeyRequest tests RotateAPIKeyRequest
func TestRotateAPIKeyRequest(t *testing.T) {
	req := &RotateAPIKeyRequest{
		Reason:   "manual",
		Metadata: map[string]any{"note": "test"},
	}

	assert.Equal(t, RotationReason("manual"), RotationReason(req.Reason))
	assert.NotNil(t, req.Metadata)
}

// TestAPIKeyCreateResponse tests APIKeyCreateResponse
func TestAPIKeyCreateResponse(t *testing.T) {
	resp := &APIKeyCreateResponse{
		APIKeyResponse: APIKeyResponse{
			ID:   uuid.New(),
			Name: "test-key",
		},
		Plaintext: "ffp_v1_abc...",
	}

	assert.NotEmpty(t, resp.Plaintext)
	assert.Equal(t, "test-key", resp.Name)
}

// TestPermissionGrant tests PermissionGrant
func TestPermissionGrant(t *testing.T) {
	perm := PermissionGrant{
		Permission:   PermissionRead,
		ResourceType: ResourceTypeFunction,
		ResourceID:   uuid.New(),
	}

	assert.Equal(t, PermissionRead, perm.Permission)
	assert.Equal(t, ResourceTypeFunction, perm.ResourceType)
}

// TestAPIKeyResponse tests APIKeyResponse
func TestAPIKeyResponse(t *testing.T) {
	now := time.Now()

	resp := &APIKeyResponse{
		ID:                    uuid.New(),
		Name:                  "test-key",
		Description:           "Test description",
		KeyType:               KeyTypePlatform,
		Prefix:                PrefixPlatform,
		ExpiresAt:             nil,
		LastRotatedAt:         now,
		RotationFrequencyDays: 90,
		RateLimit:             DefaultRateLimitConfig(),
		IsActive:              true,
		CreatedAt:             now,
	}

	assert.Equal(t, "test-key", resp.Name)
	assert.Equal(t, KeyTypePlatform, resp.KeyType)
	assert.True(t, resp.IsActive)
}

// TestAPIKeyPermission tests APIKeyPermission
func TestAPIKeyPermission(t *testing.T) {
	perm := APIKeyPermission{
		ID:           uuid.New(),
		APIKeyID:     uuid.New(),
		Permission:   PermissionWrite,
		ResourceType: ResourceTypeFunction,
		ResourceID:   uuid.New(),
		CreatedAt:    time.Now(),
	}

	assert.Equal(t, PermissionWrite, perm.Permission)
	assert.Equal(t, ResourceTypeFunction, perm.ResourceType)
}

// TestAPIKeyEnvironment tests APIKeyEnvironment
func TestAPIKeyEnvironment(t *testing.T) {
	env := APIKeyEnvironment{
		ID:              uuid.New(),
		APIKeyID:        uuid.New(),
		EnvironmentID:   uuid.New(),
		EnvironmentName: "production",
		CreatedAt:       time.Now(),
	}

	assert.Equal(t, "production", env.EnvironmentName)
}

// TestAPIKeyRotation tests APIKeyRotation
func TestAPIKeyRotation(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	userID := uuid.New()

	rotation := &APIKeyRotation{
		ID:             uuid.New(),
		APIKeyID:       uuid.New(),
		RotatedAt:      now,
		ExpiresAt:      &expiresAt,
		CreatedBy:      &userID,
		KeyHash:        "somehash",
		RotationReason: RotationReasonManual,
		Metadata:       map[string]any{"note": "test"},
	}

	assert.Equal(t, RotationReasonManual, rotation.RotationReason)
	assert.NotNil(t, rotation.ExpiresAt)
}
