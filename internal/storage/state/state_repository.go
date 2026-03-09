package state

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"sync"

	"gorm.io/gorm"
)

// AES key size constants (bytes). AES-128, AES-192, AES-256.
const (
	AESKeySize128 = 16
	AESKeySize192 = 24
	AESKeySize256 = 32
)

// GCM sizes (NIST recommendation: 96-bit nonce; Go's GCM tag is 16 bytes).
const (
	gcmNonceSize = 12
	gcmTagSize   = 16
)

var (
	ErrInvalidKeySize   = errors.New("state: encryption key must be 16, 24, or 32 bytes")
	ErrEncryptionFailed = errors.New("state: encryption failed")
	ErrDecryptionFailed = errors.New("state: decryption failed")
)

// StateRepository handles state-related database operations using GORM.
// When an encryption key is set, values can be encrypted at rest using AES-GCM.
type StateRepository struct {
	db  *gorm.DB
	mu  sync.RWMutex
	key []byte // copied on set; never expose via getter
	enc bool
}

// NewStateRepository creates a new state repository.
func NewStateRepository(db *gorm.DB) *StateRepository {
	return &StateRepository{
		db:  db,
		key: nil,
		enc: false,
	}
}

// SetEncryptionKey sets the key for encrypting/decrypting state values.
// Key must be 16, 24, or 32 bytes (AES-128, AES-192, AES-256).
// A nil or empty key disables encryption. The key is copied and not retained by the caller.
func (r *StateRepository) SetEncryptionKey(key []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Zero previous key if present
	if len(r.key) > 0 {
		for i := range r.key {
			r.key[i] = 0
		}
	}
	r.key = nil
	r.enc = false
	if len(key) == 0 {
		return
	}
	if !validAESKeySize(len(key)) {
		return
	}
	r.key = make([]byte, len(key))
	copy(r.key, key)
	r.enc = true
}

func validAESKeySize(n int) bool {
	return n == AESKeySize128 || n == AESKeySize192 || n == AESKeySize256
}

// encryptValue encrypts data using AES-GCM. Returns plaintext unchanged if encryption is disabled.
func (r *StateRepository) encryptValue(data []byte) ([]byte, error) {
	r.mu.RLock()
	key := r.key
	enc := r.enc
	r.mu.RUnlock()
	if !enc || len(key) == 0 {
		return data, nil
	}
	return encryptAES_GCM(key, data)
}

// decryptValue decrypts data using AES-GCM. Returns ciphertext unchanged if encryption is disabled.
func (r *StateRepository) decryptValue(data []byte) ([]byte, error) {
	r.mu.RLock()
	key := r.key
	enc := r.enc
	r.mu.RUnlock()
	if !enc || len(key) == 0 {
		return data, nil
	}
	return decryptAES_GCM(key, data)
}

// encryptAES_GCM encrypts plaintext with AES-GCM. Output format: nonce (12 bytes) || ciphertext || tag (16 bytes).
// Key must be 16, 24, or 32 bytes.
func encryptAES_GCM(key, plaintext []byte) ([]byte, error) {
	if !validAESKeySize(len(key)) {
		return nil, ErrInvalidKeySize
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrEncryptionFailed
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrEncryptionFailed
	}
	nonce := make([]byte, gcmNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, ErrEncryptionFailed
	}
	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptAES_GCM decrypts data produced by encryptAES_GCM. Expects: nonce (12 bytes) || ciphertext || tag (16 bytes).
func decryptAES_GCM(key, ciphertext []byte) ([]byte, error) {
	if !validAESKeySize(len(key)) {
		return nil, ErrInvalidKeySize
	}
	if len(ciphertext) < gcmNonceSize+gcmTagSize {
		return nil, ErrDecryptionFailed
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	nonce := ciphertext[:gcmNonceSize]
	payload := ciphertext[gcmNonceSize:]
	plaintext, err := aead.Open(nil, nonce, payload, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	return plaintext, nil
}
