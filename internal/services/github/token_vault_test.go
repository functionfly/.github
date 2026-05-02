package github

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenVault(t *testing.T) {
	t.Run("valid 32-byte key", func(t *testing.T) {
		vault, err := NewTokenVault("12345678901234567890123456789012")
		require.NoError(t, err)
		assert.NotNil(t, vault)
	})

	t.Run("key too short", func(t *testing.T) {
		vault, err := NewTokenVault("short")
		assert.Error(t, err)
		assert.Nil(t, vault)
		assert.Contains(t, err.Error(), "32 bytes")
	})

	t.Run("key too long", func(t *testing.T) {
		vault, err := NewTokenVault("12345678901234567890123456789012extra")
		assert.Error(t, err)
		assert.Nil(t, vault)
	})

	t.Run("empty key", func(t *testing.T) {
		vault, err := NewTokenVault("")
		assert.Error(t, err)
		assert.Nil(t, vault)
	})
}

func TestTokenVault_Encrypt(t *testing.T) {
	vault, err := NewTokenVault("12345678901234567890123456789012")
	require.NoError(t, err)

	t.Run("encrypts plaintext", func(t *testing.T) {
		ciphertext, iv, tag, err := vault.Encrypt("my-secret-token")
		require.NoError(t, err)
		assert.NotEmpty(t, ciphertext)
		assert.NotEmpty(t, iv)
		assert.NotEmpty(t, tag)
		assert.NotEqual(t, "my-secret-token", ciphertext)
	})

	t.Run("produces valid base64", func(t *testing.T) {
		ciphertext, iv, tag, err := vault.Encrypt("test")
		require.NoError(t, err)

		_, err = base64.StdEncoding.DecodeString(ciphertext)
		assert.NoError(t, err, "ciphertext should be valid base64")

		_, err = base64.StdEncoding.DecodeString(iv)
		assert.NoError(t, err, "iv should be valid base64")

		_, err = base64.StdEncoding.DecodeString(tag)
		assert.NoError(t, err, "tag should be valid base64")
	})

	t.Run("different encryptions produce different ciphertexts", func(t *testing.T) {
		ct1, iv1, _, err := vault.Encrypt("same-input")
		require.NoError(t, err)
		ct2, iv2, _, err := vault.Encrypt("same-input")
		require.NoError(t, err)

		assert.NotEqual(t, ct1, ct2, "same plaintext should produce different ciphertexts (random nonce)")
		assert.NotEqual(t, iv1, iv2, "nonces should differ")
	})

	t.Run("encrypts empty string", func(t *testing.T) {
		ciphertext, iv, tag, err := vault.Encrypt("")
		require.NoError(t, err)
		// Empty plaintext produces empty ciphertext (0 bytes encrypted + GCM tag stripped)
		assert.NotEmpty(t, iv)
		assert.NotEmpty(t, tag)
		_ = ciphertext // may be empty for zero-length plaintext
	})

	t.Run("encrypts long plaintext", func(t *testing.T) {
		longToken := make([]byte, 10000)
		for i := range longToken {
			longToken[i] = 'A'
		}
		ciphertext, iv, tag, err := vault.Encrypt(string(longToken))
		require.NoError(t, err)
		assert.NotEmpty(t, ciphertext)
		assert.NotEmpty(t, iv)
		assert.NotEmpty(t, tag)
	})
}

func TestTokenVault_Decrypt(t *testing.T) {
	vault, err := NewTokenVault("12345678901234567890123456789012")
	require.NoError(t, err)

	t.Run("roundtrip encrypt/decrypt", func(t *testing.T) {
		plaintext := "ghp_abc123def456"
		ciphertext, iv, tag, err := vault.Encrypt(plaintext)
		require.NoError(t, err)

		decrypted, err := vault.Decrypt(ciphertext, iv, tag)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("roundtrip with empty string", func(t *testing.T) {
		ciphertext, iv, tag, err := vault.Encrypt("")
		require.NoError(t, err)

		decrypted, err := vault.Decrypt(ciphertext, iv, tag)
		require.NoError(t, err)
		assert.Equal(t, "", decrypted)
	})

	t.Run("roundtrip with long string", func(t *testing.T) {
		longToken := make([]byte, 10000)
		for i := range longToken {
			longToken[i] = 'X'
		}
		plaintext := string(longToken)
		ciphertext, iv, tag, err := vault.Encrypt(plaintext)
		require.NoError(t, err)

		decrypted, err := vault.Decrypt(ciphertext, iv, tag)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("roundtrip with special characters", func(t *testing.T) {
		plaintext := "token_with_特殊字符_and_émojis_🔐"
		ciphertext, iv, tag, err := vault.Encrypt(plaintext)
		require.NoError(t, err)

		decrypted, err := vault.Decrypt(ciphertext, iv, tag)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("invalid ciphertext base64", func(t *testing.T) {
		_, err := vault.Decrypt("not-valid-base64!!!", "dGVzdA==", "dGVzdA==")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode ciphertext")
	})

	t.Run("invalid iv base64", func(t *testing.T) {
		_, err := vault.Decrypt("dGVzdA==", "not-valid-base64!!!", "dGVzdA==")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode iv")
	})

	t.Run("invalid tag base64", func(t *testing.T) {
		_, err := vault.Decrypt("dGVzdA==", "dGVzdA==", "not-valid-base64!!!")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode tag")
	})

	t.Run("wrong tag fails decryption", func(t *testing.T) {
		ciphertext, iv, _, err := vault.Encrypt("secret")
		require.NoError(t, err)

		// Use a valid base64 but wrong tag
		wrongTag := base64.StdEncoding.EncodeToString([]byte("wrong-tag-value!"))
		_, err = vault.Decrypt(ciphertext, iv, wrongTag)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decrypt")
	})

	t.Run("wrong iv fails decryption", func(t *testing.T) {
		ciphertext, _, tag, err := vault.Encrypt("secret")
		require.NoError(t, err)

		wrongIV := base64.StdEncoding.EncodeToString([]byte("wrong-iv-value!!"))
		_, err = vault.Decrypt(ciphertext, wrongIV, tag)
		assert.Error(t, err)
	})

	t.Run("wrong key fails decryption", func(t *testing.T) {
		ciphertext, iv, tag, err := vault.Encrypt("secret")
		require.NoError(t, err)

		otherVault, err := NewTokenVault("abcdefghijklmnopqrstuvwxyz012345")
		require.NoError(t, err)

		_, err = otherVault.Decrypt(ciphertext, iv, tag)
		assert.Error(t, err)
	})

	t.Run("tampered ciphertext fails decryption", func(t *testing.T) {
		ciphertext, iv, tag, err := vault.Encrypt("secret")
		require.NoError(t, err)

		decoded, _ := base64.StdEncoding.DecodeString(ciphertext)
		decoded[0] ^= 0xFF
		tampered := base64.StdEncoding.EncodeToString(decoded)

		_, err = vault.Decrypt(tampered, iv, tag)
		assert.Error(t, err)
	})
}

func TestTokenVault_EncryptDecryptMultipleKeys(t *testing.T) {
	keys := []string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"12345678901234567890123456789012",
	}

	for _, key := range keys {
		t.Run("key_"+key[:8], func(t *testing.T) {
			vault, err := NewTokenVault(key)
			require.NoError(t, err)

			plaintext := "test-token-with-key-" + key[:8]
			ct, iv, tag, err := vault.Encrypt(plaintext)
			require.NoError(t, err)

			decrypted, err := vault.Decrypt(ct, iv, tag)
			require.NoError(t, err)
			assert.Equal(t, plaintext, decrypted)
		})
	}
}
