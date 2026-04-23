package storage

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeployKeysRepository struct {
	db *gorm.DB
}

func NewDeployKeysRepository(db *gorm.DB) *DeployKeysRepository {
	return &DeployKeysRepository{db: db}
}

func ComputeSSHKeyFingerprint(publicKey string) (string, error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		parts := strings.Fields(publicKey)
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid SSH public key format")
		}
		keyData, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return "", fmt.Errorf("failed to decode key data: %w", err)
		}
		hash := sha256.Sum256(keyData)
		return "SHA256:" + base64.RawStdEncoding.EncodeToString(hash[:]), nil
	}

	hash := sha256.Sum256(block.Bytes)
	return "SHA256:" + base64.RawStdEncoding.EncodeToString(hash[:]), nil
}

func ParseAndValidateSSHKey(publicKey string) (string, error) {
	parts := strings.Fields(publicKey)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid SSH public key format: expected 'type base64-key'")
	}

	keyType := strings.ToLower(parts[0])
	switch keyType {
	case "ssh-ed25519", "ssh-rsa", "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521":
	default:
		return "", fmt.Errorf("unsupported key type: %s (supported: ssh-ed25519, ssh-rsa, ecdsa-sha2-nistp*)", keyType)
	}

	_, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode key: %w", err)
	}

	return publicKey, nil
}

func (r *DeployKeysRepository) Create(ctx context.Context, tenantID uuid.UUID, name, publicKey string, createdBy *uuid.UUID) (*DeployKey, error) {
	parsedKey, err := ParseAndValidateSSHKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	fingerprint, err := ComputeSSHKeyFingerprint(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to compute fingerprint: %w", err)
	}

	deployKey := &DeployKey{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		PublicKey:   parsedKey,
		Fingerprint: fingerprint,
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
	}

	if err := r.db.WithContext(ctx).Create(deployKey).Error; err != nil {
		return nil, err
	}

	return deployKey, nil
}

func (r *DeployKeysRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*DeployKey, error) {
	var key DeployKey
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *DeployKeysRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*DeployKey, int64, error) {
	var keys []*DeployKey
	var total int64

	query := r.db.WithContext(ctx).Model(&DeployKey{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&keys).Error; err != nil {
		return nil, 0, err
	}

	return keys, total, nil
}

func (r *DeployKeysRepository) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&DeployKey{}).Error
}

func (r *DeployKeysRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&DeployKey{}).Where("id = ?", id).Update("last_used_at", &now).Error
}

func (r *DeployKeysRepository) GetByFingerprint(ctx context.Context, fingerprint string) (*DeployKey, error) {
	var key DeployKey
	if err := r.db.WithContext(ctx).Where("fingerprint = ?", fingerprint).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *DeployKeysRepository) VerifyKey(ctx context.Context, id, tenantID uuid.UUID) (*DeployKey, error) {
	key, err := r.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("deploy key has expired")
	}

	if err := r.UpdateLastUsed(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to update last used: %w", err)
	}

	return key, nil
}
