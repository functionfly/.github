package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/google/uuid"
)

// ServerEncrypt encrypts data with a tenant-specific key using AES-256-GCM.
// The key is derived from SERVER_MASTER_KEY + tenant_id using SHA-256.
// Returns ciphertext, iv, salt, tag (all base64 encoded).
func ServerEncrypt(plaintext []byte, tenantID uuid.UUID) (ciphertext, iv, salt, tag []byte, err error) {
	masterKey := os.Getenv("SERVER_MASTER_KEY")
	if masterKey == "" {
		return nil, nil, nil, nil, fmt.Errorf("SERVER_MASTER_KEY environment variable not set")
	}

	// Derive key from master key + tenant ID
	keyMaterial := sha256.Sum256([]byte(masterKey + tenantID.String()))
	key := keyMaterial[:32] // 256 bits

	// Generate random salt and IV
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	ivBytes := make([]byte, 12) // 96 bits for GCM
	if _, err := rand.Read(ivBytes); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Create AES-GCM cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Encrypt (adds auth tag automatically)
	ciphertextWithTag := gcm.Seal(nil, ivBytes, plaintext, nil)

	// Split ciphertext and tag (last 16 bytes are tag)
	tagLen := 16
	ciphertextBytes := ciphertextWithTag[:len(ciphertextWithTag)-tagLen]
	tagBytes := ciphertextWithTag[len(ciphertextWithTag)-tagLen:]

	return ciphertextBytes, ivBytes, saltBytes, tagBytes, nil
}

// ServerDecrypt decrypts data encrypted with ServerEncrypt.
// Expects ciphertext, iv, salt, tag all as base64 encoded strings.
func ServerDecrypt(ciphertextB64, ivB64, saltB64, tagB64 []byte, tenantID uuid.UUID) ([]byte, error) {
	masterKey := os.Getenv("SERVER_MASTER_KEY")
	if masterKey == "" {
		return nil, fmt.Errorf("SERVER_MASTER_KEY environment variable not set")
	}

	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(string(ciphertextB64))
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(string(ivB64))
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %w", err)
	}

	tag, err := base64.StdEncoding.DecodeString(string(tagB64))
	if err != nil {
		return nil, fmt.Errorf("failed to decode tag: %w", err)
	}

	// Derive key
	keyMaterial := sha256.Sum256([]byte(masterKey + tenantID.String()))
	key := keyMaterial[:32]

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Combine ciphertext and tag for decryption
	combined := make([]byte, len(ciphertext)+len(tag))
	copy(combined, ciphertext)
	copy(combined[len(ciphertext):], tag)

	// Decrypt
	plaintext, err := gcm.Open(nil, iv, combined, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// EncryptToJSON encrypts plaintext and returns a JSON-serializable map
func EncryptToJSON(plaintext []byte, tenantID uuid.UUID) (map[string]interface{}, error) {
	ct, iv, salt, tag, err := ServerEncrypt(plaintext, tenantID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"ciphertext": base64.StdEncoding.EncodeToString(ct),
		"iv":         base64.StdEncoding.EncodeToString(iv),
		"salt":       base64.StdEncoding.EncodeToString(salt),
		"tag":        base64.StdEncoding.EncodeToString(tag),
		"key_version": 1,
	}, nil
}

// DecryptFromJSON decrypts a JSON map created by EncryptToJSON
func DecryptFromJSON(data map[string]interface{}, tenantID uuid.UUID) ([]byte, error) {
	ctB64, ok := data["ciphertext"].(string)
	if !ok {
		return nil, fmt.Errorf("missing ciphertext")
	}

	ivB64, ok := data["iv"].(string)
	if !ok {
		return nil, fmt.Errorf("missing iv")
	}

	saltB64, ok := data["salt"].(string)
	if !ok {
		return nil, fmt.Errorf("missing salt")
	}

	tagB64, ok := data["tag"].(string)
	if !ok {
		return nil, fmt.Errorf("missing tag")
	}

	return ServerDecrypt([]byte(ctB64), []byte(ivB64), []byte(saltB64), []byte(tagB64), tenantID)
}
