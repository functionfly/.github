package apikey

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestAPIKeyAudit_ToMap tests converting audit struct to map
func TestAPIKeyAudit_ToMap(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	keyID := uuid.New()
	now := time.Now()

	audit := &APIKeyAudit{
		EventType:     API_KEY_CREATED,
		TenantID:      &tenantID,
		UserID:        &userID,
		APIKeyID:      keyID,
		APIKeyName:    "test-key",
		KeyHash:       "abc12345",
		IPAddress:     "192.168.1.1",
		UserAgent:     "TestAgent",
		Metadata:      map[string]interface{}{"foo": "bar"},
		Success:       true,
		FailureReason: "",
		Timestamp:     now,
	}

	m := audit.ToMap()

	assert.Equal(t, API_KEY_CREATED, m["event_type"])
	assert.Equal(t, &tenantID, m["tenant_id"])
	assert.Equal(t, keyID, m["api_key_id"])
	assert.Equal(t, "test-key", m["api_key_name"])
	assert.Equal(t, "abc12345", m["key_hash"])
	assert.Equal(t, "192.168.1.1", m["ip_address"])
	assert.Equal(t, true, m["success"])
}

// TestMaskKeyName tests key name masking for privacy
func TestMaskKeyName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"testkey", "te**ey"},
		{"ab", "****"},
		{"abcde", "ab**de"},
		{"", "****"},
		{"a", "****"},
		{"abcdefghij", "ab**ij"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := MaskKeyName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTruncateHash tests hash truncation
func TestTruncateHash(t *testing.T) {
	tests := []struct {
		hash     string
		n        int
		expected string
	}{
		{"abcdef1234567890", 8, "abcdef12"},
		{"abc", 8, "abc"},
		{"abcdef1234567890", 16, "abcdef1234567890"},
		{"abcdef1234567890", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.hash, func(t *testing.T) {
			result := TruncateHash(tt.hash, tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAuditConstants tests audit event type constants
func TestAuditConstants(t *testing.T) {
	assert.Equal(t, "api_key_created", API_KEY_CREATED)
	assert.Equal(t, "api_key_viewed", API_KEY_VIEWED)
	assert.Equal(t, "api_key_updated", API_KEY_UPDATED)
	assert.Equal(t, "api_key_deleted", API_KEY_DELETED)
	assert.Equal(t, "api_key_rotated", API_KEY_ROTATED)
	assert.Equal(t, "api_key_authenticated", API_KEY_AUTHENTICATED)
	assert.Equal(t, "api_key_rate_limited", API_KEY_RATE_LIMITED)
}

// TestAuditLoggerInterface tests that AuditLogger implements Logger interface
func TestAuditLoggerInterface(t *testing.T) {
	// This test ensures the AuditLogger struct implements expected interface
	var _ Logger = (*AuditLogger)(nil)
}

// TestNoOpAuditLogger tests NoOpAuditLogger
func TestNoOpAuditLogger(t *testing.T) {
	logger := NewNoOpAuditLogger()

	ctx := context.Background()
	tenantID := uuid.New()
	keyID := uuid.New()

	// All methods should return nil (no-op)
	assert.Nil(t, logger.LogAPIKeyEvent(ctx, &APIKeyAudit{APIKeyID: keyID}))
	assert.Nil(t, logger.LogAPIKeyCreated(ctx, &tenantID, nil, keyID, "test", "hash", "127.0.0.1", "test", nil))
	assert.Nil(t, logger.LogAPIKeyViewed(ctx, &tenantID, nil, keyID, "test", "127.0.0.1", "test"))
	assert.Nil(t, logger.LogAPIKeyUpdated(ctx, &tenantID, nil, keyID, "test", "hash", "127.0.0.1", "test", nil))
	assert.Nil(t, logger.LogAPIKeyDeleted(ctx, &tenantID, nil, keyID, "test", "hash", "127.0.0.1", "test", nil))
	assert.Nil(t, logger.LogAPIKeyRotated(ctx, &tenantID, nil, keyID, "test", "hash", "127.0.0.1", "test", "manual", nil))
	assert.Nil(t, logger.LogAPIKeyAuthenticated(ctx, &tenantID, keyID, "test", "hash", "127.0.0.1", "test", true, ""))
	assert.Nil(t, logger.LogAPIKeyRateLimited(ctx, &tenantID, keyID, "test", "hash", "127.0.0.1", "test", nil))
}

// TestNewAuditLogger tests creating a new audit logger
func TestNewAuditLogger(t *testing.T) {
	// Just verify we can create the struct
	assert.NotNil(t, &AuditLogger{})
}

// TestNewNoOpAuditLogger tests creating a no-op logger
func TestNewNoOpAuditLogger(t *testing.T) {
	logger := NewNoOpAuditLogger()
	assert.NotNil(t, logger)
}

// TestNewAuditLoggerFromContext tests creating audit logger from context
func TestNewAuditLoggerFromContext(t *testing.T) {
	// Empty context should return NoOp logger
	ctx := context.Background()
	logger := NewAuditLoggerFromContext(ctx)
	assert.IsType(t, &NoOpAuditLogger{}, logger)
}

// TestWithAuditLogger tests context with audit logger
func TestWithAuditLogger(t *testing.T) {
	logger := NewNoOpAuditLogger()
	ctx := context.Background()

	ctxWithLogger := WithAuditLogger(ctx, logger)
	assert.NotNil(t, ctxWithLogger)
}

// TestAuditLoggerContextKey tests audit logger context key
func TestAuditLoggerContextKey(t *testing.T) {
	assert.Equal(t, AuditLoggerContextKeyType("audit_logger"), AuditLoggerContextKey)
}

// TestGetClientIP tests client IP extraction
func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name          string
		xForwardedFor string
		xRealIP       string
		remoteAddr    string
		expected      string
	}{
		{
			name:          "X-Forwarded-For",
			xForwardedFor: "192.168.1.100, 192.168.1.1",
			xRealIP:       "",
			remoteAddr:    "",
			expected:      "192.168.1.100",
		},
		{
			name:          "X-Real-IP only",
			xForwardedFor: "",
			xRealIP:       "10.0.0.1",
			remoteAddr:    "",
			expected:      "10.0.0.1",
		},
		{
			name:          "Remote addr fallback",
			xForwardedFor: "",
			xRealIP:       "",
			remoteAddr:    "192.168.1.1:8080",
			expected:      "192.168.1.1:8080",
		},
		{
			name:          "All empty",
			xForwardedFor: "",
			xRealIP:       "",
			remoteAddr:    "",
			expected:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetClientIP(tt.xForwardedFor, tt.xRealIP, tt.remoteAddr)
			assert.Equal(t, tt.expected, result)
		})
	}
}
