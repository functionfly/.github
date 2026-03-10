// Package apikey provides API key generation, hashing, and validation functionality.
package apikey

import (
	"time"

	"github.com/google/uuid"
)

// Key types supported by the API key system
type KeyType string

const (
	KeyTypePlatform    KeyType = "platform"
	KeyTypeFunction    KeyType = "function"
	KeyTypeAgent       KeyType = "agent"
	KeyTypeEnvironment KeyType = "environment"
	KeyTypeOAuth       KeyType = "oauth"
)

// Key prefixes by type
const (
	PrefixPlatform    = "ffp_"
	PrefixFunction    = "fff_"
	PrefixAgent       = "aep_"
	PrefixEnvironment = "ffe_"
	PrefixOAuth       = "ffo_"
)

// Default rate limits
const (
	DefaultRateLimitRPM = 1000
	DefaultRateLimitRPH = 60000
	DefaultRateLimitRPD = 1000000
	DefaultRotationDays = 90
)

// Default key version
const (
	DefaultKeyVersion = 1
)

// KeyLength constants
const (
	// RandomBytesLength is the number of random bytes in the key (16 bytes = 32 hex chars)
	RandomBytesLength = 16
	// KeyLengthTotal is the total length of the key (prefix + version + random + checksum)
	KeyLengthTotal = 47 // ffp_v1_ + 32 hex chars + _ + 8 checksum = 47
)

// Permission types for API key resources
type Permission string

const (
	PermissionRead    Permission = "read"
	PermissionWrite   Permission = "write"
	PermissionExecute Permission = "execute"
	PermissionAdmin   Permission = "admin"
)

// Resource types that can be accessed via API keys
type ResourceType string

const (
	ResourceTypeFunction   ResourceType = "function"
	ResourceTypeApp        ResourceType = "app"
	ResourceTypeTenant     ResourceType = "tenant"
	ResourceTypeRegistry   ResourceType = "registry"
	ResourceTypeDeployment ResourceType = "deployment"
	ResourceTypeSecret     ResourceType = "secret"
)

// RotationReason reasons for key rotation
type RotationReason string

const (
	RotationReasonManual      RotationReason = "manual"
	RotationReasonAutomatic   RotationReason = "automatic"
	RotationReasonCompromised RotationReason = "compromised"
)

// RateLimitConfig holds rate limiting configuration for an API key
type RateLimitConfig struct {
	RPM int `json:"rpm"` // Requests per minute
	RPH int `json:"rph"` // Requests per hour
	RPD int `json:"rpd"` // Requests per day
}

// DefaultRateLimitConfig returns the default rate limit configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RPM: DefaultRateLimitRPM,
		RPH: DefaultRateLimitRPH,
		RPD: DefaultRateLimitRPD,
	}
}

// APIKey represents an API key entity
type APIKey struct {
	ID                    uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID              uuid.UUID      `json:"tenant_id" gorm:"type:uuid;not null;index:idx_api_keys_tenant"`
	UserID                uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	Name                  string         `json:"name" gorm:"size:255;not null"`
	Description           string         `json:"description" gorm:"type:text"`
	KeyType               KeyType        `json:"key_type" gorm:"type:varchar(50);not null;index:idx_api_keys_type"`
	KeyPrefix             string         `json:"key_prefix" gorm:"size:10;not null;index:idx_api_keys_prefix"`
	KeyHash               string         `json:"-" gorm:"type:text;not null;index:idx_api_keys_hash"`
	KeyVersion            int            `json:"key_version" gorm:"not null;default:1"`
	ExpiresAt             *time.Time     `json:"expires_at" gorm:"index:idx_api_keys_expires"`
	LastRotatedAt         time.Time      `json:"last_rotated_at" gorm:"not null"`
	RotationFrequencyDays int            `json:"rotation_frequency_days" gorm:"not null;default:90"`
	RateLimitRPM          int            `json:"rate_limit_rpm" gorm:"column:rate_limit_rpm;not null;default:1000"`
	RateLimitRPH          int            `json:"rate_limit_rph" gorm:"column:rate_limit_rph;not null;default:60000"`
	RateLimitRPD          int            `json:"rate_limit_rpd" gorm:"column:rate_limit_rpd;not null;default:1000000"`
	IsActive              bool           `json:"is_active" gorm:"not null;default:true;index:idx_api_keys_active"`
	Metadata              map[string]any `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt             time.Time      `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt             time.Time      `json:"updated_at" gorm:"not null;autoUpdateTime"`
	LastUsedAt            *time.Time     `json:"last_used_at" gorm:"index"`

	// Associations
	Permissions  []APIKeyPermission  `json:"permissions,omitempty" gorm:"foreignKey:APIKeyID"`
	Environments []APIKeyEnvironment `json:"environments,omitempty" gorm:"foreignKey:APIKeyID"`
}

// APIKeyPermission represents a permission granted to an API key
type APIKeyPermission struct {
	ID           uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	APIKeyID     uuid.UUID    `json:"api_key_id" gorm:"type:uuid;not null;index:idx_api_key_permissions_key_id"`
	Permission   Permission   `json:"permission" gorm:"type:varchar(50);not null"`
	ResourceType ResourceType `json:"resource_type" gorm:"type:varchar(50);not null;index:idx_api_key_permissions_resource"`
	ResourceID   uuid.UUID    `json:"resource_id" gorm:"type:uuid;not null;index:idx_api_key_permissions_resource"`
	CreatedAt    time.Time    `json:"created_at" gorm:"not null;autoCreateTime"`
}

// APIKeyEnvironment represents an environment association for an API key
type APIKeyEnvironment struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	APIKeyID        uuid.UUID `json:"api_key_id" gorm:"type:uuid;not null;index:idx_api_key_environments_key_id"`
	EnvironmentID   uuid.UUID `json:"environment_id" gorm:"type:uuid;not null;index:idx_api_key_environments_env_id"`
	EnvironmentName string    `json:"environment_name" gorm:"size:255"`
	CreatedAt       time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
}

// APIKeyRotation represents a rotation event for an API key
type APIKeyRotation struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	APIKeyID       uuid.UUID      `json:"api_key_id" gorm:"type:uuid;not null;index:idx_api_key_rotations_key_id"`
	RotatedAt      time.Time      `json:"rotated_at" gorm:"not null;autoCreateTime;index:idx_api_key_rotations_rotated_at"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	CreatedBy      *uuid.UUID     `json:"created_by,omitempty" gorm:"type:uuid"`
	KeyHash        string         `json:"key_hash" gorm:"type:text;not null"`
	RotationReason RotationReason `json:"rotation_reason" gorm:"type:varchar(50);not null"`
	Metadata       map[string]any `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
}

// CreateAPIKeyRequest represents a request to create a new API key
type CreateAPIKeyRequest struct {
	Name                  string            `json:"name"`
	Description           string            `json:"description,omitempty"`
	KeyType               KeyType           `json:"key_type"`
	Permissions           []PermissionGrant `json:"permissions,omitempty"`
	Environments          []uuid.UUID       `json:"environments,omitempty"`
	ExpiresAt             *time.Time        `json:"expires_at,omitempty"`
	RotationFrequencyDays int               `json:"rotation_frequency_days,omitempty"`
	RateLimit             *RateLimitConfig  `json:"rate_limit,omitempty"`
	Metadata              map[string]any    `json:"metadata,omitempty"`
}

// PermissionGrant represents a permission to be granted to an API key
type PermissionGrant struct {
	Permission   Permission   `json:"permission"`
	ResourceType ResourceType `json:"resource_type"`
	ResourceID   uuid.UUID    `json:"resource_id"`
}

// RotateAPIKeyRequest represents a request to rotate an API key
type RotateAPIKeyRequest struct {
	Reason   RotationReason `json:"reason,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ListFilters holds filter options for listing API keys
type ListFilters struct {
	TenantID      *uuid.UUID `json:"tenant_id,omitempty"`
	UserID        *uuid.UUID `json:"user_id,omitempty"`
	KeyType       *KeyType   `json:"key_type,omitempty"`
	IsActive      *bool      `json:"is_active,omitempty"`
	ExpiresBefore *time.Time `json:"expires_before,omitempty"`
	ExpiresAfter  *time.Time `json:"expires_after,omitempty"`
	Search        string     `json:"search,omitempty"` // Search by name or description
}

// APIKeyResponse represents the response for API key operations
type APIKeyResponse struct {
	ID                    uuid.UUID           `json:"id"`
	Name                  string              `json:"name"`
	Description           string              `json:"description,omitempty"`
	KeyType               KeyType             `json:"key_type"`
	Prefix                string              `json:"prefix"`
	ExpiresAt             *time.Time          `json:"expires_at,omitempty"`
	LastRotatedAt         time.Time           `json:"last_rotated_at"`
	RotationFrequencyDays int                 `json:"rotation_frequency_days"`
	RateLimit             *RateLimitConfig    `json:"rate_limit"`
	IsActive              bool                `json:"is_active"`
	CreatedAt             time.Time           `json:"created_at"`
	Permissions           []APIKeyPermission  `json:"permissions,omitempty"`
	Environments          []APIKeyEnvironment `json:"environments,omitempty"`
}

// APIKeyCreateResponse represents the response when creating an API key (includes plaintext)
type APIKeyCreateResponse struct {
	APIKeyResponse
	Plaintext string `json:"plaintext"` // Only returned on creation
}

// ValidationError represents a validation error for API keys
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error returns the validation error message
func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors []ValidationError

// Error returns all validation error messages
func (e ValidationErrors) Error() string {
	result := "validation errors:"
	for _, err := range e {
		result += " " + err.Error() + ";"
	}
	return result
}

// HasErrors returns true if there are any validation errors
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// GetPrefixForKeyType returns the key prefix for a given key type
func GetPrefixForKeyType(keyType KeyType) string {
	switch keyType {
	case KeyTypePlatform:
		return PrefixPlatform
	case KeyTypeFunction:
		return PrefixFunction
	case KeyTypeAgent:
		return PrefixAgent
	case KeyTypeEnvironment:
		return PrefixEnvironment
	case KeyTypeOAuth:
		return PrefixOAuth
	default:
		return PrefixPlatform
	}
}

// GetRateLimitConfig returns the rate limit config from the flat fields
func (k *APIKey) GetRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RPM: k.RateLimitRPM,
		RPH: k.RateLimitRPH,
		RPD: k.RateLimitRPD,
	}
}

// SetRateLimitConfig sets the rate limit from a config struct
func (k *APIKey) SetRateLimitConfig(cfg *RateLimitConfig) {
	if cfg != nil {
		k.RateLimitRPM = cfg.RPM
		k.RateLimitRPH = cfg.RPH
		k.RateLimitRPD = cfg.RPD
	}
}

// ToResponse converts APIKey to APIKeyResponse
func (k *APIKey) ToResponse() *APIKeyResponse {
	resp := &APIKeyResponse{
		ID:                    k.ID,
		Name:                  k.Name,
		Description:           k.Description,
		KeyType:               k.KeyType,
		Prefix:                k.KeyPrefix,
		ExpiresAt:             k.ExpiresAt,
		LastRotatedAt:         k.LastRotatedAt,
		RotationFrequencyDays: k.RotationFrequencyDays,
		RateLimit:             k.GetRateLimitConfig(),
		IsActive:              k.IsActive,
		CreatedAt:             k.CreatedAt,
	}
	return resp
}

// IsValidKeyType checks if the given string is a valid key type
func IsValidKeyType(s string) bool {
	switch KeyType(s) {
	case KeyTypePlatform, KeyTypeFunction, KeyTypeAgent, KeyTypeEnvironment, KeyTypeOAuth:
		return true
	default:
		return false
	}
}
