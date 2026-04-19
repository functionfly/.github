package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateMagicLinkToken(t *testing.T) {
	token1, err := generateMagicLinkToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token1)
	assert.Equal(t, 64, len(token1)) // 32 bytes = 64 hex chars

	// Ensure tokens are unique
	token2, err := generateMagicLinkToken()
	require.NoError(t, err)
	assert.NotEqual(t, token1, token2)

	// Ensure tokens are hex-encoded
	for _, c := range token1 {
		assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'), "token should be hex encoded")
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase email",
			input:    "user@example.com",
			expected: "user@example.com",
		},
		{
			name:     "uppercase email",
			input:    "USER@EXAMPLE.COM",
			expected: "user@example.com",
		},
		{
			name:     "mixed case email",
			input:    "User@Example.COM",
			expected: "user@example.com",
		},
		{
			name:     "email with whitespace",
			input:    "  user@example.com  ",
			expected: "user@example.com",
		},
		{
			name:     "mixed case with whitespace",
			input:    "  USER@Example.COM  ",
			expected: "user@example.com",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeEmail(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultMagicLinkConfig(t *testing.T) {
	config := DefaultMagicLinkConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 15*time.Minute, config.TokenExpiry, "default expiry should be 15 minutes")
	assert.Equal(t, 5, config.MaxAttempts, "default max attempts should be 5")
	assert.True(t, config.AllowSignup, "default AllowSignup should be true")
}

func TestMagicLinkConfig_CustomValues(t *testing.T) {
	// Test that custom values can be set
	config := &MagicLinkConfig{
		TokenExpiry: 30 * time.Minute,
		MaxAttempts: 10,
		AllowSignup: false,
	}

	assert.Equal(t, 30*time.Minute, config.TokenExpiry)
	assert.Equal(t, 10, config.MaxAttempts)
	assert.False(t, config.AllowSignup)
}

func TestMagicLinkRequest_Validation(t *testing.T) {
	// Test valid request
	validReq := MagicLinkRequest{
		Email:        "test@example.com",
		RedirectPath: "/dashboard",
	}
	assert.Equal(t, "test@example.com", validReq.Email)
	assert.Equal(t, "/dashboard", validReq.RedirectPath)

	// Test with empty redirect path
	noRedirectReq := MagicLinkRequest{
		Email: "test@example.com",
	}
	assert.Empty(t, noRedirectReq.RedirectPath)
}

func TestMagicLinkVerifyRequest_Validation(t *testing.T) {
	// Test valid request
	validReq := MagicLinkVerifyRequest{
		Token:     "valid_token_here",
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0",
	}
	assert.Equal(t, "valid_token_here", validReq.Token)
	assert.Equal(t, "192.168.1.1", validReq.IPAddress)
	assert.Equal(t, "Mozilla/5.0", validReq.UserAgent)

	// Test request without IP/UserAgent (set from context)
	minimalReq := MagicLinkVerifyRequest{
		Token: "token_only",
	}
	assert.Equal(t, "token_only", minimalReq.Token)
	assert.Empty(t, minimalReq.IPAddress)
	assert.Empty(t, minimalReq.UserAgent)
}

func TestMagicLinkResponse_Struct(t *testing.T) {
	// Test response with email sent
	sentResp := MagicLinkResponse{
		Message:   "If that email is registered, a magic link has been sent.",
		EmailSent: true,
	}
	assert.Equal(t, "If that email is registered, a magic link has been sent.", sentResp.Message)
	assert.True(t, sentResp.EmailSent)

	// Test response without email sent
	notSentResp := MagicLinkResponse{
		Message:   "If that email is registered, a magic link has been sent.",
		EmailSent: false,
	}
	assert.False(t, notSentResp.EmailSent)
}

func TestMagicLinkVerifyResponse_Struct(t *testing.T) {
	user := &LoginUser{
		ID:    "user-id-123",
		Email: "test@example.com",
		Role:  "user",
	}

	// Test for existing user
	existingUserResp := MagicLinkVerifyResponse{
		Token:        "jwt_token_here",
		RefreshToken: "refresh_token_here",
		User:         user,
		NewUser:      false,
	}
	assert.Equal(t, "jwt_token_here", existingUserResp.Token)
	assert.Equal(t, "refresh_token_here", existingUserResp.RefreshToken)
	assert.Equal(t, user, existingUserResp.User)
	assert.False(t, existingUserResp.NewUser)

	// Test for new user
	newUserResp := MagicLinkVerifyResponse{
		Token:        "jwt_token_here",
		RefreshToken: "refresh_token_here",
		User:         user,
		NewUser:      true,
	}
	assert.True(t, newUserResp.NewUser)
}

func TestMagicLinkToken_Length(t *testing.T) {
	// Generate multiple tokens and verify they all have correct length
	for i := 0; i < 10; i++ {
		token, err := generateMagicLinkToken()
		require.NoError(t, err)
		assert.Equal(t, 64, len(token), "token should be exactly 64 characters (32 bytes hex encoded)")
	}
}

func TestMagicLinkToken_Uniqueness(t *testing.T) {
	// Generate many tokens and ensure they're all unique
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := generateMagicLinkToken()
		require.NoError(t, err)
		assert.False(t, tokens[token], "generated token should be unique")
		tokens[token] = true
	}
}
