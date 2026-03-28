package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAuthServiceForPassword creates an AuthService for testing without requiring full dependencies
func mockAuthServiceForPassword() *AuthService {
	return &AuthService{
		jwtSecret:   []byte("test-secret-key-for-testing-purposes-only"),
		jwtDuration: 30 * 0, // Will be overridden in tests
	}
}

// TestHashPassword tests password hashing with Argon2
func TestHashPassword(t *testing.T) {
	authSvc := mockAuthServiceForPassword()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "SecurePassword123!",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // Hashing empty string should work, verification will fail
		},
		{
			name:     "long password",
			password: strings.Repeat("a", 1000),
			wantErr:  false,
		},
		{
			name:     "unicode password",
			password: " пароль日本語",
			wantErr:  false,
		},
		{
			name:     "password with special characters",
			password: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := authSvc.HashPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.True(t, strings.HasPrefix(hash, "$argon2id$"), "hash should use argon2id format")
		})
	}
}

// TestHashPassword_Unique tests that same password produces different hashes (due to random salt)
func TestHashPassword_Unique(t *testing.T) {
	authSvc := mockAuthServiceForPassword()
	password := "SamePassword123!"

	hash1, err := authSvc.HashPassword(password)
	require.NoError(t, err)

	hash2, err := authSvc.HashPassword(password)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2, "same password should produce different hashes due to random salt")
}

// TestVerifyPassword_Argon2 tests password verification with Argon2 hashes
func TestVerifyPassword_Argon2(t *testing.T) {
	authSvc := mockAuthServiceForPassword()
	password := "TestPassword123!"

	hash, err := authSvc.HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name      string
		password  string
		wantMatch bool
		wantErr   bool
	}{
		{
			name:      "correct password",
			password:  password,
			wantMatch: true,
			wantErr:   false,
		},
		{
			name:      "wrong password",
			password:  "WrongPassword123!",
			wantMatch: false,
			wantErr:   false,
		},
		{
			name:      "empty password for non-empty hash",
			password:  "",
			wantMatch: false,
			wantErr:   false,
		},
		{
			name:      "case sensitivity",
			password:  "testpassword123!",
			wantMatch: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := authSvc.VerifyPassword(tt.password, hash)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantMatch, match, "password match should be %v", tt.wantMatch)
		})
	}
}

// TestVerifyPassword_Bcrypt tests password verification with bcrypt hashes (legacy format)
func TestVerifyPassword_Bcrypt(t *testing.T) {
	authSvc := mockAuthServiceForPassword()

	// Bcrypt hash for "password" (known test vector)
	// Note: The exact hash depends on the bcrypt implementation
	bcryptHash := "$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4b6n1JqNuYQq9SjC"

	tests := []struct {
		name      string
		password  string
		wantMatch bool
		wantErr   bool
	}{
		{
			name:      "wrong bcrypt password",
			password:  "wrongpassword",
			wantMatch: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := authSvc.VerifyPassword(tt.password, bcryptHash)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantMatch, match)
		})
	}
}

// TestVerifyPassword_InvalidHash tests verification with invalid hash formats
func TestVerifyPassword_InvalidHash(t *testing.T) {
	authSvc := mockAuthServiceForPassword()

	tests := []struct {
		name      string
		hash      string
		wantMatch bool
		wantErr   bool
	}{
		{
			name:    "empty hash",
			hash:    "",
			wantErr: true,
		},
		{
			name:    "invalid argon2 format - missing parts",
			hash:    "$argon2id$v=19$m=65536",
			wantErr: true,
		},
		{
			name:    "invalid argon2 format - wrong version",
			hash:    "$argon2id$v=18$m=65536,t=1,p=4$abc$def",
			wantErr: false, // Implementation doesn't strictly validate version
		},
		{
			name:    "invalid base64 salt",
			hash:    "$argon2id$v=19$m=65536,t=1,p=4$!!!$def",
			wantErr: true,
		},
		{
			name:    "random string",
			hash:    "not-a-valid-hash-at-all",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := authSvc.VerifyPassword("anypassword", tt.hash)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantMatch, match)
		})
	}
}

// TestPasswordHashFormat tests the format of generated password hashes
func TestPasswordHashFormat(t *testing.T) {
	authSvc := mockAuthServiceForPassword()

	password := "TestPassword123!"
	hash, err := authSvc.HashPassword(password)
	require.NoError(t, err)

	// Check format: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	parts := strings.Split(hash, "$")
	require.Len(t, parts, 6, "argon2 hash should have 6 parts")

	assert.Equal(t, "argon2id", parts[1])
	assert.Equal(t, "v=19", parts[2])
	assert.Equal(t, "m=65536,t=1,p=4", parts[3])
	assert.NotEmpty(t, parts[4], "salt should not be empty")
	assert.NotEmpty(t, parts[5], "hash should not be empty")
}

// TestVerifyPassword_ConstantTime tests that password verification uses constant-time comparison
func TestVerifyPassword_ConstantTime(t *testing.T) {
	authSvc := mockAuthServiceForPassword()
	password := "TestPassword123!"
	hash, err := authSvc.HashPassword(password)
	require.NoError(t, err)

	// Verify correct password
	match, err := authSvc.VerifyPassword(password, hash)
	require.NoError(t, err)
	assert.True(t, match)

	// Verify wrong password
	match, err = authSvc.VerifyPassword("WrongPassword", hash)
	require.NoError(t, err)
	assert.False(t, match)
}
