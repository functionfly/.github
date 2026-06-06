package apikey

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles persistence for API keys
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new API key repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new API key and returns it with the plaintext key (only on creation)
func (r *Repository) Create(ctx context.Context, tenantID, userID uuid.UUID, req *CreateAPIKeyRequest) (*APIKey, string, error) {
	// Generate the API key
	generator := NewGenerator()
	plaintext, err := generator.Generate(req.KeyType)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate API key: %w", err)
	}

	// Hash the key for storage
	hasher := NewHasher()
	keyHash := hasher.Hash(plaintext)

	// Get prefix for key type
	prefix := GetPrefixForKeyType(req.KeyType)

	// Set defaults
	rotationDays := req.RotationFrequencyDays
	if rotationDays == 0 {
		rotationDays = DefaultRotationDays
	}

	rateLimit := req.RateLimit
	if rateLimit == nil {
		rateLimit = DefaultRateLimitConfig()
	}

	now := time.Now()
	apiKey := &APIKey{
		ID:                    uuid.New(),
		TenantID:              tenantID,
		UserID:                userID,
		Name:                  req.Name,
		Description:           req.Description,
		KeyType:               req.KeyType,
		KeyID:                 plaintext, // Plaintext is the public key identifier
		KeyPrefix:             prefix,
		KeyHash:               keyHash,
		KeyVersion:            DefaultKeyVersion,
		ExpiresAt:             req.ExpiresAt,
		LastRotatedAt:         now,
		RotationFrequencyDays: rotationDays,
		RateLimitRPM:          rateLimit.RPM,
		RateLimitRPH:          rateLimit.RPH,
		RateLimitRPD:          rateLimit.RPD,
		IsActive:              true,
		Metadata:              JSONBMap(req.Metadata),
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the API key
		if err := tx.Create(apiKey).Error; err != nil {
			return fmt.Errorf("failed to create API key: %w", err)
		}

		// Add permissions if provided
		for _, perm := range req.Permissions {
			permission := &APIKeyPermission{
				ID:           uuid.New(),
				APIKeyID:     apiKey.ID,
				Permission:   perm.Permission,
				ResourceType: perm.ResourceType,
				ResourceID:   perm.ResourceID,
				CreatedAt:    now,
			}
			if err := tx.Create(permission).Error; err != nil {
				return fmt.Errorf("failed to create permission: %w", err)
			}
		}

		// Link environments if provided
		for _, envID := range req.Environments {
			envLink := &APIKeyEnvironment{
				ID:            uuid.New(),
				APIKeyID:      apiKey.ID,
				EnvironmentID: envID,
				CreatedAt:     now,
			}
			if err := tx.Create(envLink).Error; err != nil {
				return fmt.Errorf("failed to link environment: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, "", err
	}

	return apiKey, plaintext, nil
}

// GetByID retrieves an API key by its ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*APIKey, error) {
	var apiKey APIKey
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&apiKey).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("API key not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}
	return &apiKey, nil
}

// GetByIDWithAssociations retrieves an API key by its ID with permissions and environments
func (r *Repository) GetByIDWithAssociations(ctx context.Context, id uuid.UUID) (*APIKey, error) {
	var apiKey APIKey
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Preload("Environments").
		Where("id = ?", id).
		First(&apiKey).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("API key not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}
	return &apiKey, nil
}

// GetByHash retrieves an API key by its hash (for authentication)
func (r *Repository) GetByHash(ctx context.Context, keyHash string) (*APIKey, error) {
	var apiKey APIKey
	err := r.db.WithContext(ctx).
		Where("key_hash = ?", keyHash).
		First(&apiKey).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid API key")
		}
		return nil, fmt.Errorf("failed to authenticate API key: %w", err)
	}
	return &apiKey, nil
}

// Update updates an existing API key
func (r *Repository) Update(ctx context.Context, apiKey *APIKey) error {
	apiKey.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).Save(apiKey).Error
	if err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}
	return nil
}

// List lists API keys with filtering and pagination.
// Avoids a separate COUNT when the result set fits in one page (common case) to prevent slow count queries.
func (r *Repository) List(ctx context.Context, filters *ListFilters, limit, offset int) ([]*APIKey, int64, error) {
	var keys []*APIKey

	query := r.db.WithContext(ctx).Model(&APIKey{})

	// Apply filters
	if filters != nil {
		if filters.TenantID != nil {
			query = query.Where("tenant_id = ?", *filters.TenantID)
		}
		if filters.UserID != nil {
			query = query.Where("user_id = ?", *filters.UserID)
		}
		if filters.KeyType != nil {
			query = query.Where("key_type = ?", *filters.KeyType)
		}
		if filters.IsActive != nil {
			query = query.Where("is_active = ?", *filters.IsActive)
		}
		if filters.ExpiresBefore != nil {
			query = query.Where("expires_at < ?", *filters.ExpiresBefore)
		}
		if filters.ExpiresAfter != nil {
			query = query.Where("expires_at > ?", *filters.ExpiresAfter)
		}
		if filters.Search != "" {
			searchPattern := "%" + filters.Search + "%"
			query = query.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
		}
	}

	// Fetch limit+1 to detect whether there is another page (avoids slow COUNT when result fits in one page)
	if err := query.Order("created_at DESC").Limit(limit + 1).Offset(offset).Find(&keys).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list API keys: %w", err)
	}

	var total int64
	if len(keys) <= limit {
		total = int64(offset + len(keys))
	} else {
		// More than one page: truncate to requested limit and run count
		keys = keys[:limit]
		countQuery := r.db.WithContext(ctx).Model(&APIKey{})
		if filters != nil {
			if filters.TenantID != nil {
				countQuery = countQuery.Where("tenant_id = ?", *filters.TenantID)
			}
			if filters.UserID != nil {
				countQuery = countQuery.Where("user_id = ?", *filters.UserID)
			}
			if filters.KeyType != nil {
				countQuery = countQuery.Where("key_type = ?", *filters.KeyType)
			}
			if filters.IsActive != nil {
				countQuery = countQuery.Where("is_active = ?", *filters.IsActive)
			}
			if filters.ExpiresBefore != nil {
				countQuery = countQuery.Where("expires_at < ?", *filters.ExpiresBefore)
			}
			if filters.ExpiresAfter != nil {
				countQuery = countQuery.Where("expires_at > ?", *filters.ExpiresAfter)
			}
			if filters.Search != "" {
				searchPattern := "%" + filters.Search + "%"
				countQuery = countQuery.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
			}
		}
		if err := countQuery.Count(&total).Error; err != nil {
			return nil, 0, fmt.Errorf("failed to count API keys: %w", err)
		}
	}

	return keys, total, nil
}

// Delete performs a soft delete (sets is_active = false)
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&APIKey{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("failed to delete API key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("API key not found: %s", id)
	}
	return nil
}

// HardDelete permanently deletes an API key
func (r *Repository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&APIKey{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to hard delete API key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("API key not found: %s", id)
	}
	return nil
}

// Rotate handles key rotation - creates a new key and records the rotation
func (r *Repository) Rotate(ctx context.Context, keyID uuid.UUID, newPlaintext string, reason RotationReason, createdBy *uuid.UUID, metadata map[string]any) error {
	// Get the existing key
	existingKey, err := r.GetByID(ctx, keyID)
	if err != nil {
		return err
	}

	// Hash the new key
	hasher := NewHasher()
	newHash := hasher.Hash(newPlaintext)

	now := time.Now()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create rotation record with old key hash
		rotation := &APIKeyRotation{
			ID:             uuid.New(),
			APIKeyID:       keyID,
			RotatedAt:      now,
			ExpiresAt:      existingKey.ExpiresAt,
			CreatedBy:      createdBy,
			KeyHash:        existingKey.KeyHash,
			RotationReason: reason,
			Metadata:       JSONBMap(metadata),
		}
		if err := tx.Create(rotation).Error; err != nil {
			return fmt.Errorf("failed to create rotation record: %w", err)
		}

		// Update the key with new hash and version
		updates := map[string]interface{}{
			"key_hash":        newHash,
			"key_version":     existingKey.KeyVersion + 1,
			"last_rotated_at": now,
			"updated_at":      now,
		}
		if err := tx.Model(existingKey).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to rotate API key: %w", err)
		}

		return nil
	})
}

// GetRotationHistory retrieves the rotation history for an API key
func (r *Repository) GetRotationHistory(ctx context.Context, keyID uuid.UUID) ([]*APIKeyRotation, error) {
	var rotations []*APIKeyRotation
	err := r.db.WithContext(ctx).
		Where("api_key_id = ?", keyID).
		Order("rotated_at DESC").
		Find(&rotations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get rotation history: %w", err)
	}
	return rotations, nil
}

// AddPermission adds a permission to an API key
func (r *Repository) AddPermission(ctx context.Context, keyID uuid.UUID, perm *PermissionGrant) error {
	permission := &APIKeyPermission{
		ID:           uuid.New(),
		APIKeyID:     keyID,
		Permission:   perm.Permission,
		ResourceType: perm.ResourceType,
		ResourceID:   perm.ResourceID,
		CreatedAt:    time.Now(),
	}
	err := r.db.WithContext(ctx).Create(permission).Error
	if err != nil {
		return fmt.Errorf("failed to add permission: %w", err)
	}
	return nil
}

// RemovePermission removes a permission from an API key
func (r *Repository) RemovePermission(ctx context.Context, keyID uuid.UUID, perm Permission, resourceType ResourceType, resourceID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("api_key_id = ? AND permission = ? AND resource_type = ? AND resource_id = ?",
			keyID, perm, resourceType, resourceID).
		Delete(&APIKeyPermission{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove permission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("permission not found")
	}
	return nil
}

// GetPermissions retrieves all permissions for an API key
func (r *Repository) GetPermissions(ctx context.Context, keyID uuid.UUID) ([]*APIKeyPermission, error) {
	var permissions []*APIKeyPermission
	err := r.db.WithContext(ctx).
		Where("api_key_id = ?", keyID).
		Find(&permissions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}
	return permissions, nil
}

// LinkEnvironment links an environment to an API key
func (r *Repository) LinkEnvironment(ctx context.Context, keyID, envID uuid.UUID, envName string) error {
	envLink := &APIKeyEnvironment{
		ID:              uuid.New(),
		APIKeyID:        keyID,
		EnvironmentID:   envID,
		EnvironmentName: envName,
		CreatedAt:       time.Now(),
	}
	err := r.db.WithContext(ctx).Create(envLink).Error
	if err != nil {
		return fmt.Errorf("failed to link environment: %w", err)
	}
	return nil
}

// UnlinkEnvironment unlinks an environment from an API key
func (r *Repository) UnlinkEnvironment(ctx context.Context, keyID, envID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("api_key_id = ? AND environment_id = ?", keyID, envID).
		Delete(&APIKeyEnvironment{})
	if result.Error != nil {
		return fmt.Errorf("failed to unlink environment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("environment link not found")
	}
	return nil
}

// GetEnvironments retrieves all linked environments for an API key
func (r *Repository) GetEnvironments(ctx context.Context, keyID uuid.UUID) ([]*APIKeyEnvironment, error) {
	var environments []*APIKeyEnvironment
	err := r.db.WithContext(ctx).
		Where("api_key_id = ?", keyID).
		Find(&environments).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get environments: %w", err)
	}
	return environments, nil
}

// UpdateLastUsed updates the last_used_at timestamp for an API key
func (r *Repository) UpdateLastUsed(ctx context.Context, keyID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&APIKey{}).
		Where("id = ?", keyID).
		Update("last_used_at", time.Now())
	if result.Error != nil {
		return fmt.Errorf("failed to update last used: %w", result.Error)
	}
	return nil
}

// UpdateExpiresAt sets the expires_at timestamp for an API key.
// Pass nil to clear the expiry.
func (r *Repository) UpdateExpiresAt(ctx context.Context, keyID uuid.UUID, expiresAt *time.Time) error {
	updates := map[string]interface{}{
		"expires_at": expiresAt,
		"updated_at": time.Now(),
	}
	result := r.db.WithContext(ctx).Model(&APIKey{}).
		Where("id = ?", keyID).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update expires_at: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("API key not found: %s", keyID)
	}
	return nil
}

// CountByTenant counts API keys for a tenant
func (r *Repository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&APIKey{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count API keys: %w", err)
	}
	return count, nil
}

// GetExpiringKeys retrieves API keys that are expiring within the specified duration
func (r *Repository) GetExpiringKeys(ctx context.Context, before time.Time) ([]*APIKey, error) {
	var keys []*APIKey
	err := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at <= ? AND is_active = true", before).
		Find(&keys).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get expiring keys: %w", err)
	}
	return keys, nil
}

// GetKeysNeedingRotation retrieves API keys that need rotation based on their rotation frequency
func (r *Repository) GetKeysNeedingRotation(ctx context.Context) ([]*APIKey, error) {
	var keys []*APIKey

	// Keys where last_rotated_at + rotation_frequency_days <= now
	cutoffDate := time.Now()

	err := r.db.WithContext(ctx).
		Where("is_active = true AND rotation_frequency_days > 0 AND last_rotated_at + (rotation_frequency_days * interval '1 day') <= ?", cutoffDate).
		Find(&keys).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get keys needing rotation: %w", err)
	}
	return keys, nil
}

// CreateTrustAPIKey creates a new Trust API key for a partner.
// It uses the fft_ prefix and the platform v1 key format (prefix_v1_random_checksum).
func (r *Repository) CreateTrustAPIKey(ctx context.Context, req *CreateTrustAPIKeyRequest) (*APIKey, string, error) {
	// Generate the API key using the Trust type (fft_v1_...)
	generator := NewGenerator()
	plaintext, err := generator.Generate(KeyTypeTrust)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate Trust API key: %w", err)
	}

	// Hash the key for storage
	hasher := NewHasher()
	keyHash := hasher.Hash(plaintext)

	now := time.Now()
	apiKey := &APIKey{
		ID:                    uuid.New(),
		TenantID:              req.PartnerID, // PartnerID used as TenantID for Trust keys
		UserID:                req.PartnerID,
		Name:                  req.Name,
		Description:           req.Description,
		KeyType:               KeyTypeTrust,
		KeyID:                 plaintext, // Trust API keys use plaintext as the public identifier
		KeyPrefix:             PrefixTrust,
		KeyHash:               keyHash,
		KeyVersion:            DefaultKeyVersion,
		ExpiresAt:             req.ExpiresAt,
		LastRotatedAt:         now,
		RotationFrequencyDays: DefaultRotationDays,
		RateLimitRPM:          DefaultRateLimitRPM,
		RateLimitRPH:          DefaultRateLimitRPH,
		RateLimitRPD:          DefaultRateLimitRPD,
		IsActive:              true,
		IsRevoked:             false,
		UseCount:              0,
		PartnerID:             req.PartnerID,
		Scopes:                JSONBMap{"trust:read": true},
		AllowedIPs:            JSONBMap{},
		Metadata:              JSONBMap{},
		CreatedBy:             req.CreatedBy,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	// Set scopes
	if len(req.Scopes) > 0 {
		scopesMap := JSONBMap{}
		for _, s := range req.Scopes {
			scopesMap[s] = true
		}
		apiKey.Scopes = scopesMap
	}

	// Set allowed IPs (validate each entry; skip invalid ones with a warning)
	if len(req.AllowedIPs) > 0 {
		allowedIPsMap := JSONBMap{}
		for _, ip := range req.AllowedIPs {
			if !isValidIPOrCIDR(ip) {
				continue
			}
			allowedIPsMap[ip] = true
		}
		apiKey.AllowedIPs = allowedIPsMap
	}

	if err := r.db.WithContext(ctx).Create(apiKey).Error; err != nil {
		return nil, "", fmt.Errorf("failed to create Trust API key: %w", err)
	}

	return apiKey, plaintext, nil
}

// ValidateAPIKey validates a Trust API key (fft_ prefixed) and returns the key + partner info.
// It supports both the new platform format (fft_v1_xxx_crc8) and legacy format (fft_xxx).
func (r *Repository) ValidateAPIKey(rawKey string) (*APIKey, error) {
	var keyHash string

	// Determine format and compute hash
	if strings.HasPrefix(rawKey, PrefixTrust+"v1_") {
		// New platform format: fft_v1_random_crc8
		hasher := NewHasher()
		keyHash = hasher.Hash(rawKey)
	} else if strings.HasPrefix(rawKey, PrefixTrust) {
		// Legacy Trust format: fft_xxx (no version/checksum)
		hash := sha256.Sum256([]byte(rawKey))
		keyHash = hex.EncodeToString(hash[:])
	} else {
		return nil, fmt.Errorf("invalid API key prefix")
	}

	var apiKey APIKey
	err := r.db.Where("key_hash = ?", keyHash).First(&apiKey).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid API key")
		}
		return nil, fmt.Errorf("failed to validate API key: %w", err)
	}

	// Check if revoked
	if apiKey.IsRevoked {
		return nil, fmt.Errorf("API key has been revoked")
	}

	// Check if expired
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Check if active
	if !apiKey.IsActive {
		return nil, fmt.Errorf("API key is not active")
	}

	return &apiKey, nil
}

// CheckIPAllowed checks if a client IP is in the key's allowed IPs list.
// If the key has no allowed IPs configured, all IPs are allowed.
// Supports both individual IP addresses and CIDR ranges (e.g., "192.168.1.0/24").
func (r *Repository) CheckIPAllowed(apiKey *APIKey, clientIP string) bool {
	if apiKey.AllowedIPs == nil || len(apiKey.AllowedIPs) == 0 {
		return true
	}

	// Check each allowed IP/CIDR (stored as map keys)
	for allowedIP := range apiKey.AllowedIPs {
		// Check for exact match
		if allowedIP == clientIP {
			return true
		}

		// Check for CIDR match
		if strings.Contains(allowedIP, "/") {
			if ipInCIDR(clientIP, allowedIP) {
				return true
			}
		}
	}

	return false
}

// isValidIPOrCIDR returns true if s is a valid bare IP address or CIDR range.
// Bare IPs are normalized to /32 (IPv4) or /128 (IPv6) so storage is consistent.
func isValidIPOrCIDR(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if strings.Contains(s, "/") {
		_, _, err := net.ParseCIDR(s)
		return err == nil
	}
	return net.ParseIP(s) != nil
}

// ipInCIDR checks if an IP address is within a CIDR range
func ipInCIDR(ipAddress, cidrRange string) bool {
	_, ipNet, err := net.ParseCIDR(cidrRange)
	if err != nil {
		// Invalid CIDR format
		return false
	}

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		// Invalid IP address
		return false
	}

	return ipNet.Contains(ip)
}

// IncrementKeyUsage increments the use count for a Trust API key.
func (r *Repository) IncrementKeyUsage(keyID uuid.UUID) error {
	result := r.db.Model(&APIKey{}).
		Where("id = ?", keyID).
		UpdateColumn("use_count", gorm.Expr("use_count + 1"))
	if result.Error != nil {
		return fmt.Errorf("failed to increment key usage: %w", result.Error)
	}
	return nil
}

// RevokeAPIKey revokes a Trust API key.
func (r *Repository) RevokeAPIKey(keyID uuid.UUID, reason string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"is_revoked":     true,
		"revoked_at":     now,
		"revoked_reason": reason,
		"is_active":      false,
		"updated_at":     now,
	}
	result := r.db.Model(&APIKey{}).Where("id = ?", keyID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to revoke API key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("API key not found: %s", keyID)
	}
	return nil
}

// ListAPIKeysForPartner lists all API keys for a Trust partner.
func (r *Repository) ListAPIKeysForPartner(ctx context.Context, partnerID uuid.UUID, includeRevoked bool) ([]*APIKey, error) {
	var keys []*APIKey
	query := r.db.WithContext(ctx).Where("partner_id = ?", partnerID)
	if !includeRevoked {
		query = query.Where("is_revoked = ?", false)
	}
	if err := query.Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, fmt.Errorf("failed to list API keys for partner: %w", err)
	}
	return keys, nil
}

// GetAPIKeyByKeyID retrieves a Trust API key by its public key ID (key_id column).
func (r *Repository) GetAPIKeyByKeyID(ctx context.Context, keyID string) (*APIKey, error) {
	var apiKey APIKey
	// For new format keys, key_id is the full key (fft_v1_xxx_xx)
	// For legacy keys, key_id is fft_xxx (just the raw key without hash)
	err := r.db.WithContext(ctx).Where("key_id = ?", keyID).First(&apiKey).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}
	return &apiKey, nil
}
