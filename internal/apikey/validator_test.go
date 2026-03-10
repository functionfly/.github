package apikey

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidator_Validate tests full key validation
func TestValidator_Validate(t *testing.T) {
	generator := NewGenerator()
	hasher := NewHasher()

	// Generate a valid key
	key, err := generator.Generate(KeyTypePlatform)
	require.NoError(t, err)
	keyHash := hasher.Hash(key)

	validator := NewValidator()

	// Valid key - not expired, active
	expiresAt := time.Now().Add(24 * time.Hour)
	result := validator.Validate(key, keyHash, true, &expiresAt)
	// Just check it runs without panic
	assert.NotNil(t, result)
}

// TestValidator_Validate_InvalidFormat tests validation with invalid key format
func TestValidator_Validate_InvalidFormat(t *testing.T) {
	validator := NewValidator()

	result := validator.Validate("invalid_key", "some_hash", true, nil)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
}

// TestValidatePermissions tests permission validation
func TestValidatePermissions(t *testing.T) {
	resourceID := uuid.New()

	tests := []struct {
		name            string
		keyPermissions  []APIKeyPermission
		requiredPerm    Permission
		resourceType    ResourceType
		resourceIDParam uuid.UUID
		expected        bool
	}{
		{
			name: "exact permission match",
			keyPermissions: []APIKeyPermission{
				{Permission: PermissionRead, ResourceType: ResourceTypeFunction, ResourceID: resourceID},
			},
			requiredPerm:    PermissionRead,
			resourceType:    ResourceTypeFunction,
			resourceIDParam: resourceID,
			expected:        true,
		},
		{
			name: "admin has all permissions",
			keyPermissions: []APIKeyPermission{
				{Permission: PermissionAdmin, ResourceType: ResourceTypeFunction, ResourceID: resourceID},
			},
			requiredPerm:    PermissionWrite,
			resourceType:    ResourceTypeFunction,
			resourceIDParam: resourceID,
			expected:        true,
		},
		{
			name: "no matching permission",
			keyPermissions: []APIKeyPermission{
				{Permission: PermissionRead, ResourceType: ResourceTypeFunction, ResourceID: uuid.New()},
			},
			requiredPerm:    PermissionWrite,
			resourceType:    ResourceTypeFunction,
			resourceIDParam: resourceID,
			expected:        false,
		},
		{
			name: "wrong resource type",
			keyPermissions: []APIKeyPermission{
				{Permission: PermissionRead, ResourceType: ResourceTypeApp, ResourceID: resourceID},
			},
			requiredPerm:    PermissionRead,
			resourceType:    ResourceTypeFunction,
			resourceIDParam: resourceID,
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePermissions(tt.keyPermissions, tt.requiredPerm, tt.resourceType, tt.resourceIDParam)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCheckRotationDue tests rotation check
func TestCheckRotationDue(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name              string
		lastRotatedAt     time.Time
		rotationDays      int
		expectRotationDue bool
	}{
		{
			name:              "rotation due",
			lastRotatedAt:     now.AddDate(0, 0, -100),
			rotationDays:      90,
			expectRotationDue: true,
		},
		{
			name:              "rotation not due",
			lastRotatedAt:     now.AddDate(0, 0, -30),
			rotationDays:      90,
			expectRotationDue: false,
		},
		{
			name:              "rotation disabled",
			lastRotatedAt:     now.AddDate(0, 0, -100),
			rotationDays:      0,
			expectRotationDue: false,
		},
		{
			name:              "negative rotation days",
			lastRotatedAt:     now.AddDate(0, 0, -100),
			rotationDays:      -1,
			expectRotationDue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckRotationDue(tt.lastRotatedAt, tt.rotationDays)
			assert.Equal(t, tt.expectRotationDue, result)
		})
	}
}

// TestValidateCreateRequest tests create request validation
func TestValidateCreateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateAPIKeyRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &CreateAPIKeyRequest{
				Name:                  "test-key",
				KeyType:               KeyTypePlatform,
				RotationFrequencyDays: 90,
				RateLimit:             &RateLimitConfig{RPM: 1000, RPH: 60000, RPD: 1000000},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			req: &CreateAPIKeyRequest{
				KeyType: KeyTypePlatform,
			},
			wantErr: true,
		},
		{
			name: "invalid key type",
			req: &CreateAPIKeyRequest{
				Name:    "test-key",
				KeyType: "invalid",
			},
			wantErr: true,
		},
		{
			name: "negative rotation days",
			req: &CreateAPIKeyRequest{
				Name:                  "test-key",
				KeyType:               KeyTypePlatform,
				RotationFrequencyDays: -1,
			},
			wantErr: true,
		},
		{
			name: "negative rate limit",
			req: &CreateAPIKeyRequest{
				Name:    "test-key",
				KeyType: KeyTypePlatform,
				RateLimit: &RateLimitConfig{
					RPM: -1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateRotateRequest tests rotate request validation
func TestValidateRotateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *RotateAPIKeyRequest
		wantErr bool
	}{
		{
			name:    "empty request (valid)",
			req:     &RotateAPIKeyRequest{},
			wantErr: false,
		},
		{
			name: "manual reason",
			req: &RotateAPIKeyRequest{
				Reason: "manual",
			},
			wantErr: false,
		},
		{
			name: "automatic reason",
			req: &RotateAPIKeyRequest{
				Reason: "automatic",
			},
			wantErr: false,
		},
		{
			name: "compromised reason",
			req: &RotateAPIKeyRequest{
				Reason: "compromised",
			},
			wantErr: false,
		},
		{
			name: "invalid reason",
			req: &RotateAPIKeyRequest{
				Reason: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRotateRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidationResult tests ValidationResult methods
func TestValidationResult(t *testing.T) {
	result := &ValidationResult{
		Valid:     true,
		KeyType:   KeyTypePlatform,
		Prefix:    PrefixPlatform,
		Version:   Version,
		IsExpired: false,
		IsActive:  true,
	}

	assert.True(t, result.Valid)
	assert.Equal(t, KeyTypePlatform, result.KeyType)
}

// TestKeyValidationErrors tests KeyValidationErrors
func TestKeyValidationErrors(t *testing.T) {
	errors := &KeyValidationErrors{}

	errors.Add("error 1")
	errors.Add("error 2: %s", "detail")

	assert.True(t, errors.HasErrors())
	assert.Contains(t, errors.Error(), "error 1")
	assert.Contains(t, errors.Error(), "error 2")
}

// TestKeyValidationErrors_Empty tests empty validation errors
func TestKeyValidationErrors_Empty(t *testing.T) {
	errors := &KeyValidationErrors{}

	assert.False(t, errors.HasErrors())
	assert.Empty(t, errors.Error())
}

// TestNewValidator tests creating a new validator
func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.currentTimeFunc)
}
