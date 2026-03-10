package apikey

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerator_Generate tests key generation for all key types
func TestGenerator_Generate(t *testing.T) {
	generator := NewGenerator()

	tests := []struct {
		name     string
		keyType  KeyType
		expected string
	}{
		{"platform", KeyTypePlatform, PrefixPlatform},
		{"function", KeyTypeFunction, PrefixFunction},
		{"agent", KeyTypeAgent, PrefixAgent},
		{"environment", KeyTypeEnvironment, PrefixEnvironment},
		{"oauth", KeyTypeOAuth, PrefixOAuth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := generator.Generate(tt.keyType)
			require.NoError(t, err)
			assert.NotEmpty(t, key)
			assert.True(t, strings.HasPrefix(key, tt.expected), "key should start with prefix %s, got %s", tt.expected, key)
		})
	}
}

// TestGenerator_Generate_Randomness tests that generated keys are unique
func TestGenerator_Generate_Randomness(t *testing.T) {
	generator := NewGenerator()

	keys := make(map[string]bool)
	count := 100

	for i := 0; i < count; i++ {
		key, err := generator.Generate(KeyTypePlatform)
		require.NoError(t, err)
		keys[key] = true
	}

	// All keys should be unique
	assert.Equal(t, count, len(keys), "all generated keys should be unique")
}

// TestGenerator_GenerateWithPrefix tests generating keys with custom prefixes
func TestGenerator_GenerateWithPrefix(t *testing.T) {
	generator := NewGenerator()

	key, err := generator.GenerateWithPrefix("test_")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(key, "test_v1_"))
	assert.NotEmpty(t, key)
}

// TestGenerator_Generate_Error tests error handling when random source fails
func TestGenerator_Generate_Error(t *testing.T) {
	generator := &Generator{
		randomSource: func([]byte) (int, error) {
			return 0, assert.AnError
		},
	}

	key, err := generator.Generate(KeyTypePlatform)
	assert.Error(t, err)
	assert.Empty(t, key)
	assert.Contains(t, err.Error(), "failed to generate random bytes")
}

// TestParseKey tests parsing a valid key
func TestParseKey(t *testing.T) {
	generator := NewGenerator()
	key, err := generator.Generate(KeyTypePlatform)
	require.NoError(t, err)

	prefix, version, randomHex, checksumHex, err := ParseKey(key)

	require.NoError(t, err)
	assert.NotEmpty(t, prefix)
	assert.Equal(t, Version, version)
	assert.Equal(t, 32, len(randomHex), "random hex should be 32 characters")
	assert.Equal(t, 2, len(checksumHex), "checksum hex should be 2 characters")
}

// TestParseKey_InvalidFormat tests parsing invalid key formats
func TestParseKey_InvalidFormat(t *testing.T) {
	// Empty key
	_, _, _, _, err := ParseKey("")
	assert.Error(t, err)

	// Wrong number of parts
	_, _, _, _, err = ParseKey("ffp_v1_abc")
	assert.Error(t, err)
}

// TestValidateKeyFormat tests key format validation
func TestValidateKeyFormat(t *testing.T) {
	generator := NewGenerator()
	validKey, err := generator.Generate(KeyTypePlatform)
	require.NoError(t, err)

	// Valid key should pass
	err = ValidateKeyFormat(validKey)
	assert.NoError(t, err)

	// Empty should fail
	err = ValidateKeyFormat("")
	assert.Error(t, err)
}

// TestVerifyChecksum tests checksum verification with tampered key
func TestVerifyChecksum_Invalid(t *testing.T) {
	generator := NewGenerator()
	key, _ := generator.Generate(KeyTypePlatform)

	// Tampered key - change a character
	tamperedKey := key[:len(key)-3] + "xx"
	assert.False(t, VerifyChecksum(tamperedKey), "tampered key should have invalid checksum")
}

// TestGetKeyTypeFromPrefix tests extracting key type from prefix
func TestGetKeyTypeFromPrefix(t *testing.T) {
	// Valid prefixes
	kt, err := GetKeyTypeFromPrefix(PrefixPlatform)
	assert.NoError(t, err)
	assert.Equal(t, KeyTypePlatform, kt)

	kt, err = GetKeyTypeFromPrefix(PrefixFunction)
	assert.NoError(t, err)
	assert.Equal(t, KeyTypeFunction, kt)

	// Invalid prefix
	_, err = GetKeyTypeFromPrefix("invalid_")
	assert.Error(t, err)
}

// TestGenerateForType tests the convenience function
func TestGenerateForType(t *testing.T) {
	key, err := GenerateForType(KeyTypePlatform)
	require.NoError(t, err)
	assert.NotEmpty(t, key)
	assert.True(t, strings.HasPrefix(key, PrefixPlatform))
}
