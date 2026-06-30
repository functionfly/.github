package aikeys

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/google/uuid"
)

const keyVersion = 1

// EncryptKey encrypts an API key using AES-256-GCM with a tenant-specific key
// derived from SERVER_MASTER_KEY + tenant_id. Returns ciphertext (with auth tag
// appended), nonce, and key version.
func EncryptKey(plaintext []byte, tenantID uuid.UUID) (ciphertext, nonce []byte, version int, err error) {
	key, err := deriveKey(tenantID)
	if err != nil {
		return nil, nil, 0, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("gcm: %w", err)
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, 0, fmt.Errorf("nonce: %w", err)
	}

	// gcm.Seal appends the 16-byte auth tag to ciphertext
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, keyVersion, nil
}

// DecryptKey decrypts an API key encrypted with EncryptKey.
// The ciphertext must include the 16-byte GCM auth tag appended.
func DecryptKey(ciphertext, nonce []byte, tenantID uuid.UUID) ([]byte, error) {
	key, err := deriveKey(tenantID)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}

// deriveKey derives a 32-byte AES key from SERVER_MASTER_KEY + tenant ID.
func deriveKey(tenantID uuid.UUID) ([]byte, error) {
	masterKey := os.Getenv("SERVER_MASTER_KEY")
	if masterKey == "" {
		return nil, fmt.Errorf("SERVER_MASTER_KEY not set")
	}

	material := sha256.Sum256([]byte(masterKey + tenantID.String()))
	return material[:32], nil
}

// KeyLast4 returns the last 4 characters of a key for display purposes.
func KeyLast4(key string) string {
	if len(key) <= 4 {
		return key
	}
	return key[len(key)-4:]
}

// GenerateID generates a random hex ID for ai_provider_keys records.
func GenerateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
