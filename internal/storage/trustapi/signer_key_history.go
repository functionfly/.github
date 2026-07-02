package trustapi

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SigningKeyRecord represents a historical signing key in the database
type SigningKeyRecord struct {
	ID            uuid.UUID          `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	KeyID         string             `json:"key_id" gorm:"size:64;not null;uniqueIndex"`
	PublicKeyHex  string             `json:"public_key_hex" gorm:"type:text;not null"`
	Algorithm     SignatureAlgorithm `json:"algorithm" gorm:"size:50;not null"`
	Backend       string             `json:"backend" gorm:"size:50;not null"`
	IsActive      bool               `json:"is_active" gorm:"default:false"`
	Fingerprint   string             `json:"fingerprint" gorm:"size:64;not null;index"`
	ActivatedAt   time.Time          `json:"activated_at" gorm:"not null"`
	DeactivatedAt *time.Time         `json:"deactivated_at,omitempty"`
	CreatedAt     time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time          `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for SigningKeyRecord
func (SigningKeyRecord) TableName() string {
	return "signing_key_history"
}

// SigningKeyRepository handles database operations for signing key history
type SigningKeyRepository struct {
	db *gorm.DB
}

var signingKeyRepo *SigningKeyRepository

// SetSigningKeyRepository sets the package-level signing key repository
func SetSigningKeyRepository(repo *SigningKeyRepository) {
	signingKeyRepo = repo
}

// NewSigningKeyRepository creates a new signing key repository
func NewSigningKeyRepository(db *gorm.DB) *SigningKeyRepository {
	return &SigningKeyRepository{db: db}
}

// RecordKey saves the current signing key to the history table
func RecordKey(signer Signer, backend string) error {
	if signingKeyRepo == nil || signer == nil {
		return nil
	}
	return signingKeyRepo.RecordKey(signer, backend)
}

// GetKeyByID looks up a historical key by its key_id
func GetKeyByID(keyID string) (*SigningKeyRecord, error) {
	if signingKeyRepo == nil {
		return nil, fmt.Errorf("signing key repository not initialized")
	}
	return signingKeyRepo.GetKeyByID(keyID)
}

// GetAllKeys returns all historical signing keys ordered by activated_at desc
func GetAllKeys() ([]SigningKeyRecord, error) {
	if signingKeyRepo == nil {
		return nil, fmt.Errorf("signing key repository not initialized")
	}
	return signingKeyRepo.GetAllKeys()
}

// DeactivateKeyByID marks a key as no longer active by key ID
func DeactivateKeyByID(keyID string) error {
	if signingKeyRepo == nil {
		return fmt.Errorf("signing key repository not initialized")
	}
	return signingKeyRepo.DeactivateKey(keyID)
}

// RecordKey saves the current signing key to the history table (repository method)
func (r *SigningKeyRepository) RecordKey(signer Signer, backend string) error {
	if signer == nil {
		return nil
	}

	fingerprint := sha256.Sum256([]byte(signer.PublicKeyHex()))
	fingerprintHex := hex.EncodeToString(fingerprint[:])

	if r == nil || r.db == nil {
		fmt.Fprintf(os.Stderr, "attestation: signing key repository not initialized, skipping key record\n")
		return nil
	}

	// Deactivate any previously active keys
	r.db.Model(&SigningKeyRecord{}).
		Where("is_active = ?", true).
		Updates(map[string]interface{}{
			"is_active":      false,
			"deactivated_at": time.Now(),
		})

	record := &SigningKeyRecord{
		KeyID:        signer.KeyID(),
		PublicKeyHex: signer.PublicKeyHex(),
		Algorithm:    signer.Algorithm(),
		Backend:      backend,
		IsActive:     true,
		Fingerprint:  fingerprintHex,
		ActivatedAt:  time.Now(),
	}

	result := r.db.Create(record)
	return result.Error
}

// GetKeyByID looks up a historical key by its key_id
func (r *SigningKeyRepository) GetKeyByID(keyID string) (*SigningKeyRecord, error) {
	var record SigningKeyRecord
	result := r.db.Where("key_id = ?", keyID).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

// GetAllKeys returns all historical signing keys ordered by activated_at desc
func (r *SigningKeyRepository) GetAllKeys() ([]SigningKeyRecord, error) {
	var records []SigningKeyRecord
	result := r.db.Order("activated_at DESC").Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	return records, nil
}

// DeactivateKey marks a key as no longer active
func (r *SigningKeyRepository) DeactivateKey(keyID string) error {
	now := time.Now()
	result := r.db.Model(&SigningKeyRecord{}).
		Where("key_id = ?", keyID).
		Updates(map[string]interface{}{
			"is_active":      false,
			"deactivated_at": now,
		})
	return result.Error
}

// GetActiveKey returns the currently active signing key record
func (r *SigningKeyRepository) GetActiveKey() (*SigningKeyRecord, error) {
	var record SigningKeyRecord
	result := r.db.Where("is_active = ?", true).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

// KeyExists checks if a key with the given fingerprint already exists
func (r *SigningKeyRepository) KeyExists(fingerprint string) bool {
	var count int64
	r.db.Model(&SigningKeyRecord{}).Where("fingerprint = ?", fingerprint).Count(&count)
	return count > 0
}
