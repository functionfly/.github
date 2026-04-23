package storage

import (
	"time"

	"github.com/google/uuid"
)

type DeployKey struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Name        string     `json:"name" gorm:"size:255;not null"`
	PublicKey   string     `json:"public_key" gorm:"type:text;not null"`
	Fingerprint string     `json:"fingerprint" gorm:"size:64;not null;uniqueIndex"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime;not null"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty" gorm:"type:uuid"`
}

func (DeployKey) TableName() string {
	return "deploy_keys"
}

type DeployKeyCreateRequest struct {
	Name      string     `json:"name" binding:"required,min=1,max=255"`
	PublicKey string     `json:"public_key" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type DeployKeyResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	PublicKey   string     `json:"public_key"`
	Fingerprint string     `json:"fingerprint"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
}

type DeployKeyListResponse struct {
	DeployKeys []DeployKeyResponse `json:"deploy_keys"`
	TotalCount int64               `json:"total_count"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}
