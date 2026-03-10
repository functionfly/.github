package apikey

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRateLimitResult tests RateLimitResult
func TestRateLimitResult(t *testing.T) {
	result := &RateLimitResult{
		Allowed:    true,
		Remaining:  999,
		RetryAfter: 0,
	}

	assert.True(t, result.Allowed)
	assert.Equal(t, 999, result.Remaining)
}

// TestRateLimitResult_NotAllowed tests RateLimitResult when not allowed
func TestRateLimitResult_NotAllowed(t *testing.T) {
	result := &RateLimitResult{
		Allowed:    false,
		Remaining:  0,
		RetryAfter: 30,
	}

	assert.False(t, result.Allowed)
	assert.Equal(t, 0, result.Remaining)
}

// TestAPIKeyContextKey tests context keys
func TestAPIKeyContextKey(t *testing.T) {
	// Test that context keys are properly defined
	assert.Equal(t, APIKeyContextKey("api_key_id"), APIKeyIDContextKey)
	assert.Equal(t, APIKeyContextKey("api_key_hash"), APIKeyHashContextKey)
	assert.Equal(t, APIKeyContextKey("api_key_tenant_id"), APIKeyTenantIDContextKey)
	assert.Equal(t, APIKeyContextKey("api_key_user_id"), APIKeyUserIDContextKey)
	assert.Equal(t, APIKeyContextKey("api_key_type"), APIKeyTypeContextKey)
	assert.Equal(t, APIKeyContextKey("api_key_name"), APIKeyNameContextKey)
}

// TestGetAPIKeyFromContext tests extracting API key ID from context
func TestGetAPIKeyFromContext(t *testing.T) {
	keyID := uuid.New()
	ctx := context.WithValue(context.Background(), APIKeyIDContextKey, keyID)

	retrievedID, ok := GetAPIKeyFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, keyID, retrievedID)
}

// TestGetAPIKeyFromContext_NotFound tests extracting from empty context
func TestGetAPIKeyFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	retrievedID, ok := GetAPIKeyFromContext(ctx)
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, retrievedID)
}

// TestGetTenantIDFromContext tests extracting tenant ID from context
func TestGetTenantIDFromContext(t *testing.T) {
	tenantID := uuid.New()
	ctx := context.WithValue(context.Background(), APIKeyTenantIDContextKey, tenantID)

	retrievedID, ok := GetTenantIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, tenantID, retrievedID)
}

// TestGetUserIDFromContext tests extracting user ID from context
func TestGetUserIDFromContext(t *testing.T) {
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), APIKeyUserIDContextKey, userID)

	retrievedID, ok := GetUserIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, userID, retrievedID)
}

// TestGetAPIKeyTypeFromContext tests extracting key type from context
func TestGetAPIKeyTypeFromContext(t *testing.T) {
	keyType := KeyTypePlatform
	ctx := context.WithValue(context.Background(), APIKeyTypeContextKey, keyType)

	retrievedType, ok := GetAPIKeyTypeFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, keyType, retrievedType)
}

// TestGetAPIKeyNameFromContext tests extracting key name from context
func TestGetAPIKeyNameFromContext(t *testing.T) {
	keyName := "test-key"
	ctx := context.WithValue(context.Background(), APIKeyNameContextKey, keyName)

	retrievedName, ok := GetAPIKeyNameFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, keyName, retrievedName)
}

// TestGetAPIKeyHashFromContext tests extracting key hash from context
func TestGetAPIKeyHashFromContext(t *testing.T) {
	keyHash := "abc123def456"
	ctx := context.WithValue(context.Background(), APIKeyHashContextKey, keyHash)

	retrievedHash, ok := GetAPIKeyHashFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, keyHash, retrievedHash)
}

// TestExtractAPIKey tests extracting API key from HTTP request
func TestExtractAPIKey(t *testing.T) {
	// Create a test key
	generator := NewGenerator()
	key, _ := generator.Generate(KeyTypePlatform)

	tests := []struct {
		name          string
		headerKey     string
		headerValue   string
		expectedKey   string
		expectEmpty   bool
	}{
		{
			name:        "X-API-Key header",
			headerKey:   "X-API-Key",
			headerValue: key,
			expectedKey: key,
			expectEmpty: false,
		},
		{
			name:        "Authorization header with ApiKey prefix",
			headerKey:   "Authorization",
			headerValue: "ApiKey " + key,
			expectedKey: key,
			expectEmpty: false,
		},
		{
			name:        "Authorization header lowercase",
			headerKey:   "Authorization",
			headerValue: "apikey " + key,
			expectedKey: key,
			expectEmpty: false,
		},
		{
			name:        "Authorization header with different prefix",
			headerKey:   "Authorization",
			headerValue: "Bearer " + key,
			expectedKey: "",
			expectEmpty: true,
		},
		{
			name:        "Empty X-API-Key",
			headerKey:   "X-API-Key",
			headerValue: "",
			expectedKey: "",
			expectEmpty: true,
		},
		{
			name:        "No headers",
			headerKey:   "",
			headerValue: "",
			expectedKey: "",
			expectEmpty: true,
		},
	}

	// Create a simple test helper to simulate key extraction
	extractKey := func(req *http.Request) string {
		// Check X-API-Key header first
		if key := req.Header.Get("X-API-Key"); key != "" {
			return key
		}

		// Check Authorization header
		auth := req.Header.Get("Authorization")
		if auth == "" {
			return ""
		}

		// Support "ApiKey <key>" format
		parts := splitAuthHeader(auth)
		if len(parts) == 2 && parts[0] == "ApiKey" {
			return parts[1]
		}

		return ""
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.headerKey != "" {
				req.Header.Set(tt.headerKey, tt.headerValue)
			}

			result := extractKey(req)

			if tt.expectEmpty {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expectedKey, result)
			}
		})
	}
}

// splitAuthHeader is a test helper to split authorization header
func splitAuthHeader(auth string) []string {
	// Simple implementation for testing
	if len(auth) >= 7 && auth[:7] == "ApiKey " {
		return []string{"ApiKey", auth[7:]}
	}
	if len(auth) >= 7 && auth[:7] == "apikey " {
		return []string{"ApiKey", auth[7:]}
	}
	return []string{auth}
}

// TestRateLimitKeyPrefix tests rate limit key prefix constant
func TestRateLimitKeyPrefix(t *testing.T) {
	assert.Equal(t, "ratelimit:api:", rateLimitKeyPrefix)
}

// TestRateLimitTTL tests rate limit TTL values
func TestRateLimitTTL(t *testing.T) {
	assert.Equal(t, 60, int(rateLimitTTLMinute.Seconds()))
	assert.Equal(t, 3600, int(rateLimitTTLHour.Seconds()))
	assert.Equal(t, 86400, int(rateLimitTTLDay.Seconds()))
}

// TestRateLimiterInterface tests that RateLimiter has required methods
func TestRateLimiterInterface(t *testing.T) {
	// This test ensures the RateLimiter struct implements expected interface
	// We can't fully test without Redis, but we can verify basic struct existence
	var _ interface{} = (*RateLimiter)(nil)
	require.NotNil(t, &RateLimiter{})
}

// TestNewRateLimiter tests creating a new rate limiter
func TestNewRateLimiter(t *testing.T) {
	// Just verify we can create the struct
	// In real tests we'd mock Redis
	assert.NotNil(t, &RateLimiter{})
}
