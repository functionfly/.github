package mfa

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MFATOTP represents a TOTP configuration for a user
type MFATOTP struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_mfa_totp_user"`
	Secret    string         `gorm:"type:text;not null"` // Encrypted TOTP secret
	Enabled   bool           `gorm:"default:false"`
	Verified  bool           `gorm:"default:false"`
	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the MFATOTP model
func (MFATOTP) TableName() string {
	return "gba_mfa_totp"
}

// BeforeCreate hook to set timestamps
func (m *MFATOTP) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.Must(uuid.NewRandom())
	}
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	return nil
}

// BeforeUpdate hook to update timestamps
func (m *MFATOTP) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}

// MFABackupCode represents a single-use backup code for MFA recovery
type MFABackupCode struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:idx_backup_code_user_code"`
	CodeHash  string     `gorm:"type:text;not null"` // bcrypt hash of the backup code
	Used      bool       `gorm:"default:false"`
	UsedAt    *time.Time `gorm:"default:null"`
	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for the MFABackupCode model
func (MFABackupCode) TableName() string {
	return "gba_mfa_backup_codes"
}

// BeforeCreate hook to set timestamps
func (m *MFABackupCode) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.Must(uuid.NewRandom())
	}
	m.CreatedAt = time.Now()
	return nil
}

// MarkAsUsed marks a backup code as used
func (m *MFABackupCode) MarkAsUsed() {
	now := time.Now()
	m.Used = true
	m.UsedAt = &now
}

// TOTPSetup represents the response for TOTP setup
type TOTPSetup struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// TOTPVerifyRequest represents a TOTP verification request
type TOTPVerifyRequest struct {
	Code string `json:"code"`
}

// TOTPVerifyResponse represents a TOTP verification response
type TOTPVerifyResponse struct {
	Enabled bool   `json:"enabled"`
	Message string `json:"message,omitempty"`
}

// MFAChallengeRequest represents an MFA challenge request during login
type MFAChallengeRequest struct {
	Code       string `json:"code,omitempty"`
	BackupCode string `json:"backup_code,omitempty"`
}

// MFAChallengeResponse represents an MFA challenge response
type MFAChallengeResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

// MFADisableRequest represents an MFA disable request
type MFADisableRequest struct {
	Code       string `json:"code,omitempty"`
	BackupCode string `json:"backup_code,omitempty"`
	Password   string `json:"password"` // Required for security
}

// MFADisableResponse represents an MFA disable response
type MFADisableResponse struct {
	Disabled bool   `json:"disabled"`
	Message  string `json:"message,omitempty"`
}

// BackupCodesRegenerateRequest represents a backup code regeneration request
type BackupCodesRegenerateRequest struct {
	Code string `json:"code"`
}

// BackupCodesRegenerateResponse represents a backup code regeneration response
type BackupCodesRegenerateResponse struct {
	BackupCodes []string `json:"backup_codes"`
}

// MFAStatusResponse represents the MFA status response
type MFAStatusResponse struct {
	Enabled         bool `json:"enabled"`
	Verified        bool `json:"verified"`
	HasBackupCodes  bool `json:"has_backup_codes"`
	BackupCodeCount int  `json:"backup_code_count"`
}
