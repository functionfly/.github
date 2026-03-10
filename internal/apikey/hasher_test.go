package apikey

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHasher_Hash tests SHA-256 hashing
func TestHasher_Hash(t *testing.T) {
	hasher := NewHasher()

	key := "testkey123"
	hash := hasher.Hash(key)

	assert.NotEmpty(t, hash)
	assert.Equal(t, 64, len(hash), "SHA-256 hash should be 64 hex characters")
}

// TestHasher_Hash_Deterministic tests that hashing is deterministic
func TestHasher_Hash_Deterministic(t *testing.T) {
	hasher := NewHasher()

	key := "testkey123"
	hash1 := hasher.Hash(key)
	hash2 := hasher.Hash(key)

	assert.Equal(t, hash1, hash2, "same key should produce same hash")
}

// TestHasher_Hash_UniqueKeys tests that different keys produce different hashes
func TestHasher_Hash_UniqueKeys(t *testing.T) {
	hasher := NewHasher()

	key1 := "key1"
	key2 := "key2"

	hash1 := hasher.Hash(key1)
	hash2 := hasher.Hash(key2)

	assert.NotEqual(t, hash1, hash2, "different keys should produce different hashes")
}

// TestHasher_HashBytes tests hashing byte slices
func TestHasher_HashBytes(t *testing.T) {
	hasher := NewHasher()

	data := []byte("test data")
	hash := hasher.HashBytes(data)

	assert.NotEmpty(t, hash)
	assert.Equal(t, 64, len(hash), "SHA-256 hash should be 64 hex characters")
}

// TestHasher_Hash_WithSalt tests hashing with a salt
func TestHasher_Hash_WithSalt(t *testing.T) {
	hasher := NewHasherWithSalt("mysalt")

	key := "testkey123"
	hash := hasher.Hash(key)

	assert.NotEmpty(t, hash)

	// Same key with different salt should produce different hash
	hasher2 := NewHasherWithSalt("differentsalt")
	hash2 := hasher2.Hash(key)

	assert.NotEqual(t, hash, hash2, "different salt should produce different hash")
}

// TestHasher_Verify tests constant-time comparison
func TestHasher_Verify(t *testing.T) {
	hasher := NewHasher()

	key := "testkey123"
	hash := hasher.Hash(key)

	// Correct key should verify
	assert.True(t, hasher.Verify(key, hash), "correct key should verify")

	// Wrong key should not verify
	assert.False(t, hasher.Verify("wrong_key", hash), "wrong key should not verify")

	// Empty key should not verify
	assert.False(t, hasher.Verify("", hash), "empty key should not verify")
}

// TestVerifyPrefix tests prefix verification
func TestVerifyPrefix(t *testing.T) {
	key := "ffp_v1_abc123"

	// Correct prefix (partial match)
	assert.True(t, VerifyPrefix(key, "ffp"))

	// Wrong prefix
	assert.False(t, VerifyPrefix(key, "fff"))
}

// TestVerifyMultiplePrefixes tests checking multiple prefixes
func TestVerifyMultiplePrefixes(t *testing.T) {
	key := "ffp_v1_abc123"

	// Key matches one of the prefixes
	assert.True(t, VerifyMultiplePrefixes(key, "fff", "ffp", "aep"))

	// Key matches none of the prefixes
	assert.False(t, VerifyMultiplePrefixes(key, "fff", "aep"))
}

// TestShortHash tests short hash generation
func TestShortHash(t *testing.T) {
	hash := "abcdef1234567890"
	short := ShortHash(hash)

	assert.Equal(t, "abcdef12", short)
	assert.Equal(t, 8, len(short))
}

// TestShortHash_ShortInput tests short hash with input shorter than 8 chars
func TestShortHash_ShortInput(t *testing.T) {
	short := ShortHash("abc")
	assert.Equal(t, "abc", short)
}

// TestMaskKey tests key masking
func TestMaskKey(t *testing.T) {
	key := "ffp_v1_abc123def456abc123def456abc123de_12ab"
	masked := MaskKey(key)

	assert.NotEqual(t, key, masked, "masked key should be different from original")
	assert.Contains(t, masked, "xxxx", "masked key should contain xxxx")
	assert.True(t, len(masked) < len(key), "masked key should be shorter than original")
}

// TestMaskKey_ShortInput tests masking very short keys
func TestMaskKey_ShortInput(t *testing.T) {
	short := "abc"
	masked := MaskKey(short)

	assert.Equal(t, "****", masked)
}

// TestHashKey tests the convenience function
func TestHashKey(t *testing.T) {
	key := "testkey123"
	hash := HashKey(key)

	assert.NotEmpty(t, hash)
	assert.Equal(t, 64, len(hash))
}

// TestVerifyKey tests the convenience verification function
func TestVerifyKey(t *testing.T) {
	key := "testkey123"
	hash := HashKey(key)

	assert.True(t, VerifyKey(key, hash))
	assert.False(t, VerifyKey("wrong_key", hash))
}

// TestPrepareKeyForStorage tests key preparation for storage
func TestPrepareKeyForStorage(t *testing.T) {
	// Valid key should be accepted
	generator := NewGenerator()
	validKey, _ := generator.Generate(KeyTypePlatform)

	hash, err := PrepareKeyForStorage(validKey)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Empty key should be rejected
	_, err = PrepareKeyForStorage("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key cannot be empty")
}

// TestExtractKeyInfo tests extracting key info without full validation
func TestExtractKeyInfo(t *testing.T) {
	generator := NewGenerator()
	key, _ := generator.Generate(KeyTypePlatform)

	info := ExtractKeyInfo(key)
	require.NotNil(t, info, "valid key should return info")

	assert.NotEmpty(t, info.Prefix)
	assert.NotEmpty(t, info.Version)
	assert.Equal(t, 32, len(info.RandomHex))
	assert.Equal(t, 2, len(info.Checksum))
}

// TestExtractKeyInfo_Invalid tests extracting info from invalid keys
func TestExtractKeyInfo_Invalid(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"empty", ""},
		{"invalid format", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ExtractKeyInfo(tt.key)
			assert.Nil(t, info, "invalid key should return nil")
		})
	}
}
