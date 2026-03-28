package storage

import (
	"time"

	"github.com/google/uuid"
)

// SignupInviteCode stores hashed platform signup invite codes (invite-only launch).
type SignupInviteCode struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CodeFingerprint string     `gorm:"column:code_fingerprint;size:64;uniqueIndex:idx_signup_invite_fingerprint;not null"`
	CodeHash        string     `gorm:"column:code_hash;type:text;not null"`
	Label           string     `gorm:"column:label;size:512"`
	MaxUses         *int       `gorm:"column:max_uses"`
	UsesCount       int        `gorm:"column:uses_count;not null;default:0"`
	ExpiresAt       *time.Time `gorm:"column:expires_at"`
	RevokedAt       *time.Time `gorm:"column:revoked_at"`
	CreatedBy       *uuid.UUID `gorm:"column:created_by;type:uuid"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (SignupInviteCode) TableName() string {
	return "signup_invite_codes"
}

// SignupInviteCodeAdminList is a safe subset for admin API responses (no secrets).
type SignupInviteCodeAdminList struct {
	ID        uuid.UUID  `json:"id"`
	Label     string     `json:"label"`
	MaxUses   *int       `json:"maxUses,omitempty"`
	UsesCount int        `json:"usesCount"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
	CreatedBy *uuid.UUID `json:"createdBy,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}
