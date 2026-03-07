package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WebAuthnCredential represents a WebAuthn/Passkey credential stored in the database
type WebAuthnCredential struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	User           *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CredentialID   []byte     `json:"credential_id" gorm:"type:bytea;not null"` // Raw credential ID from WebAuthn
	PublicKey      []byte     `json:"public_key" gorm:"type:bytea;not null"`    // CBOR-encoded public key
	SignCount      uint32     `json:"sign_count" gorm:"not null;default:0"`     // Signature counter
	BackupEligible bool       `json:"backup_eligible" gorm:"default:false"`     // Whether credential is eligible for backup
	BackupState    bool       `json:"backup_state" gorm:"default:false"`        // Current backup state
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
}

// TableName returns the table name for WebAuthnCredential
func (WebAuthnCredential) TableName() string {
	return "webauthn_credentials"
}

// WebAuthnRepository handles database operations for WebAuthn credentials
type WebAuthnRepository struct {
	db *gorm.DB
}

// NewWebAuthnRepository creates a new WebAuthn repository
func NewWebAuthnRepository(db *gorm.DB) *WebAuthnRepository {
	return &WebAuthnRepository{
		db: db,
	}
}

// Create creates a new WebAuthn credential
func (r *WebAuthnRepository) Create(credential *WebAuthnCredential) error {
	return r.db.Create(credential).Error
}

// GetByUserID retrieves all credentials for a user
func (r *WebAuthnRepository) GetByUserID(userID uuid.UUID) ([]*WebAuthnCredential, error) {
	var credentials []*WebAuthnCredential
	err := r.db.Where("user_id = ?", userID).Find(&credentials).Error
	return credentials, err
}

// GetByCredentialID retrieves a credential by its credential ID (raw bytes)
func (r *WebAuthnRepository) GetByCredentialID(credentialID []byte) (*WebAuthnCredential, error) {
	var credential WebAuthnCredential
	err := r.db.Where("credential_id = ?", credentialID).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

// GetByID retrieves a credential by its UUID
func (r *WebAuthnRepository) GetByID(id uuid.UUID) (*WebAuthnCredential, error) {
	var credential WebAuthnCredential
	err := r.db.First(&credential, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

// UpdateSignCount updates the signature counter for a credential
func (r *WebAuthnRepository) UpdateSignCount(id uuid.UUID, signCount uint32) error {
	return r.db.Model(&WebAuthnCredential{}).Where("id = ?", id).Update("sign_count", signCount).Error
}

// UpdateLastUsed updates the last used timestamp for a credential
func (r *WebAuthnRepository) UpdateLastUsed(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&WebAuthnCredential{}).Where("id = ?", id).Update("last_used_at", now).Error
}

// Delete deletes a credential by its UUID
func (r *WebAuthnRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&WebAuthnCredential{}, "id = ?", id).Error
}

// DeleteByUserID deletes all credentials for a user
func (r *WebAuthnRepository) DeleteByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&WebAuthnCredential{}).Error
}

// CountByUserID returns the number of credentials for a user
func (r *WebAuthnRepository) CountByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&WebAuthnCredential{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// GetAllCredentialsForUser returns all credentials for a user (used for login)
func (r *WebAuthnRepository) GetAllCredentialsForUser(userID uuid.UUID) ([]*WebAuthnCredential, error) {
	var credentials []*WebAuthnCredential
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&credentials).Error
	return credentials, err
}
