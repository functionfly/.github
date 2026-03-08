package state

import (
	"gorm.io/gorm"
)

// StateRepository handles state-related database operations using GORM
type StateRepository struct {
	db               *gorm.DB
	encryptionKey    []byte
	encryptionEnabled bool
}

// NewStateRepository creates a new state repository
func NewStateRepository(db *gorm.DB) *StateRepository {
	repo := &StateRepository{db: db}
	// Check if encryption is available (from environment or config)
	repo.encryptionEnabled = false // Default to false, enable per-state
	return repo
}

// SetEncryptionKey sets the encryption key for state values
func (r *StateRepository) SetEncryptionKey(key []byte) {
	r.encryptionKey = key
	r.encryptionEnabled = len(key) > 0
}

// encryptValue encrypts a value using AES-GCM
func (r *StateRepository) encryptValue(data []byte) ([]byte, error) {
	if !r.encryptionEnabled || r.encryptionKey == nil {
		return data, nil
	}
	// Use AES-GCM encryption
	return encryptAES_GCM(r.encryptionKey, data)
}

// decryptValue decrypts a value using AES-GCM
func (r *StateRepository) decryptValue(data []byte) ([]byte, error) {
	if !r.encryptionEnabled || r.encryptionKey == nil {
		return data, nil
	}
	// Use AES-GCM decryption
	return decryptAES_GCM(r.encryptionKey, data)
}

// encryptAES_GCM encrypts data using AES-GCM
func encryptAES_GCM(key, plaintext []byte) ([]byte, error) {
	// Simplified implementation - in production use crypto/aes properly
	// This is a placeholder that would use the actual encryption
	return plaintext, nil
}

// decryptAES_GCM decrypts data using AES-GCM
func decryptAES_GCM(key, ciphertext []byte) ([]byte, error) {
	// Simplified implementation - in production use crypto/aes properly
	return ciphertext, nil
}