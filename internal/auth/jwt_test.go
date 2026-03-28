package auth

import (
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAuthServiceForJWT creates an AuthService for JWT testing
func mockAuthServiceForJWT(jwtSecret string) *AuthService {
	return &AuthService{
		jwtSecret:   []byte(jwtSecret),
		jwtDuration: 1 * time.Hour,
	}
}

// mockUser creates a mock user for testing
func mockUser() *storage.User {
	tenantID := uuid.New()
	username := "testuser"
	return &storage.User{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Email:     "test@example.com",
		Username:  &username,
		Role:      "user",
	}
}

// mockAdminUser creates a mock admin user for testing
func mockAdminUser() *storage.User {
	tenantID := uuid.New()
	username := "admin"
	return &storage.User{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Email:     "admin@example.com",
		Username:  &username,
		Role:      "admin",
	}
}

// TestGenerateToken tests JWT token generation
func TestGenerateToken(t *testing.T) {
	authSvc := mockAuthServiceForJWT("test-secret-key")
	user := mockUser()

	token, err := authSvc.GenerateToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

// TestGenerateToken_AdminRole tests token generation for admin users
func TestGenerateToken_AdminRole(t *testing.T) {
	authSvc := mockAuthServiceForJWT("test-secret-key")
	user := mockAdminUser()

	token, err := authSvc.GenerateToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

// TestValidateToken tests token validation
func TestValidateToken(t *testing.T) {
	authSvc := mockAuthServiceForJWT("test-secret-key")
	user := mockUser()

	// Generate a token
	token, err := authSvc.GenerateToken(user)
	require.NoError(t, err)

	// Validate the token
	claims, err := authSvc.ValidateToken(token)
	require.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.TenantID, claims.TenantID)
}

// TestValidateToken_InvalidToken tests validation of invalid tokens
func TestValidateToken_InvalidToken(t *testing.T) {
	authSvc := mockAuthServiceForJWT("test-secret-key")

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "malformed token",
			token: "not-a-valid-jwt",
		},
		{
			name:  "random string",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzNDU2Nzg5MCJ9.invalid-signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := authSvc.ValidateToken(tt.token)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

// TestValidateToken_WrongSecret tests validation with wrong secret
func TestValidateToken_WrongSecret(t *testing.T) {
	authSvc1 := mockAuthServiceForJWT("secret-one")
	authSvc2 := mockAuthServiceForJWT("secret-two")
	user := mockUser()

	// Generate token with one secret
	token, err := authSvc1.GenerateToken(user)
	require.NoError(t, err)

	// Try to validate with different secret
	claims, err := authSvc2.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestGenerateRefreshToken tests refresh token generation
func TestGenerateRefreshToken(t *testing.T) {
	authSvc := mockAuthServiceForJWT("test-secret-key")

	token, hash, err := authSvc.GenerateRefreshToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, token, hash, "token and hash should be different")
}

// TestGenerateRefreshToken_Unique tests that refresh tokens are unique
func TestGenerateRefreshToken_Unique(t *testing.T) {
	authSvc := mockAuthServiceForJWT("test-secret-key")

	token1, _, err := authSvc.GenerateRefreshToken()
	require.NoError(t, err)

	token2, _, err := authSvc.GenerateRefreshToken()
	require.NoError(t, err)

	assert.NotEqual(t, token1, token2, "refresh tokens should be unique")
}

// TestGenerateRefreshToken_Length tests refresh token length (64 bytes = 128 hex chars)
func TestGenerateRefreshToken_Length(t *testing.T) {
	authSvc := mockAuthServiceForJWT("test-secret-key")

	token, _, err := authSvc.GenerateRefreshToken()
	require.NoError(t, err)
	assert.Equal(t, 128, len(token), "refresh token should be 128 hex characters (64 bytes)")
}

// TestGenerateInviteToken tests invite token generation
func TestGenerateInviteToken(t *testing.T) {
	token, expiresAt := GenerateInviteToken()

	assert.NotEmpty(t, token)
	assert.Equal(t, 64, len(token), "invite token should be 64 hex characters")
	assert.True(t, expiresAt.After(time.Now()), "expiresAt should be in the future")

	// Default expiration is 7 days
	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	assert.True(t, expiresAt.Before(expectedExpiry.Add(time.Second)), "expiresAt should be approximately 7 days from now")
}

// TestGenerateInviteToken_Unique tests that invite tokens are unique
func TestGenerateInviteToken_Unique(t *testing.T) {
	token1, _ := GenerateInviteToken()
	token2, _ := GenerateInviteToken()

	assert.NotEqual(t, token1, token2, "invite tokens should be unique")
}

// TestClaims_UserID tests that UserID claim is properly set
func TestClaims_UserID(t *testing.T) {
	authSvc := mockAuthServiceForJWT("test-secret-key")
	user := mockUser()

	token, err := authSvc.GenerateToken(user)
	require.NoError(t, err)

	claims, err := authSvc.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, user.ID, claims.UserID)
}

// TestClaims_Email tests that Email claim is properly set
func TestClaims_Email(t *testing.T) {
	authSvc := mockAuthServiceForJWT("test-secret-key")
	user := mockUser()

	token, err := authSvc.GenerateToken(user)
	require.NoError(t, err)

	claims, err := authSvc.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, user.Email, claims.Email)
}

// TestClaims_TenantID tests that TenantID claim is properly set
func TestClaims_TenantID(t *testing.T) {
	authSvc := mockAuthServiceForJWT("test-secret-key")
	user := mockUser()

	token, err := authSvc.GenerateToken(user)
	require.NoError(t, err)

	claims, err := authSvc.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, user.TenantID, claims.TenantID)
}

// TestClaims_Role tests that Role claim is properly set
func TestClaims_Role(t *testing.T) {
	authSvc := mockAuthServiceForJWT("test-secret-key")
	user := mockAdminUser()

	token, err := authSvc.GenerateToken(user)
	require.NoError(t, err)

	claims, err := authSvc.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, user.Role, claims.Role)
}

// TestClaims_Permissions tests that permissions are set based on role
func TestClaims_Permissions(t *testing.T) {
	authSvc := mockAuthServiceForJWT("test-secret-key")

	// Test admin user permissions
	adminUser := mockAdminUser()
	token, err := authSvc.GenerateToken(adminUser)
	require.NoError(t, err)

	claims, err := authSvc.ValidateToken(token)
	require.NoError(t, err)

	assert.Contains(t, claims.Permissions, "*", "admin should have wildcard permission")
}

// TestClaims_Username tests that Username claim is set when present
func TestClaims_Username(t *testing.T) {
	authSvc := mockAuthServiceForJWT("test-secret-key")
	user := mockUser()

	token, err := authSvc.GenerateToken(user)
	require.NoError(t, err)

	claims, err := authSvc.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, *user.Username, claims.Username)
}

// TestGetPermissionsForRole tests the role-based permissions
func TestGetPermissionsForRole(t *testing.T) {
	authSvc := mockAuthServiceForJWT("test-secret-key")

	tests := []struct {
		role            string
		expectedPerms   []string
		shouldBeEmpty   bool
	}{
		{
			role:          "admin",
			expectedPerms: []string{"*"},
		},
		{
			role:          "user",
			expectedPerms: []string{"functions:read", "functions:execute", "api_keys:read", "api_keys:write"},
		},
		{
			role:          "viewer",
			expectedPerms: []string{"functions:read"},
		},
		{
			role:          "unknown",
			shouldBeEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			perms := authSvc.getPermissionsForRole(tt.role)
			if tt.shouldBeEmpty {
				assert.Empty(t, perms)
			} else {
				assert.Equal(t, tt.expectedPerms, perms)
			}
		})
	}
}
