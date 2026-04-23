package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

type DeployKeysService struct {
	repo *DeployKeysRepository
}

func NewDeployKeysService(repo *DeployKeysRepository) *DeployKeysService {
	return &DeployKeysService{repo: repo}
}

func (s *DeployKeysService) CreateDeployKey(ctx context.Context, tenantID uuid.UUID, name, publicKey string, createdBy *uuid.UUID) (*DeployKey, error) {
	parsedKey, err := ParseAndValidateSSHKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	fingerprint, err := ComputeSSHKeyFingerprint(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to compute fingerprint: %w", err)
	}

	existing, err := s.repo.GetByFingerprint(ctx, fingerprint)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("deploy key with this fingerprint already exists")
	}

	return s.repo.Create(ctx, tenantID, name, parsedKey, createdBy)
}

func (s *DeployKeysService) GetDeployKey(ctx context.Context, id, tenantID uuid.UUID) (*DeployKey, error) {
	return s.repo.GetByID(ctx, id, tenantID)
}

func (s *DeployKeysService) ListDeployKeys(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*DeployKey, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.repo.ListByTenant(ctx, tenantID, pageSize, offset)
}

func (s *DeployKeysService) DeleteDeployKey(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.repo.Delete(ctx, id, tenantID)
}

func (s *DeployKeysService) VerifyDeployKey(ctx context.Context, id, tenantID uuid.UUID) (*DeployKey, error) {
	return s.repo.VerifyKey(ctx, id, tenantID)
}

func (s *DeployKeysService) RotateDeployKey(ctx context.Context, id, tenantID uuid.UUID, newPublicKey string) (*DeployKey, error) {
	_, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if _, err := ParseAndValidateSSHKey(newPublicKey); err != nil {
		return nil, fmt.Errorf("invalid new public key: %w", err)
	}

	newFingerprint, err := ComputeSSHKeyFingerprint(newPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to compute fingerprint: %w", err)
	}

	existing, err := s.repo.GetByFingerprint(ctx, newFingerprint)
	if err == nil && existing != nil && existing.ID != id {
		return nil, fmt.Errorf("deploy key with this fingerprint already exists")
	}

	newKey := &DeployKey{
		ID:          id,
		TenantID:    tenantID,
		PublicKey:   newPublicKey,
		Fingerprint: newFingerprint,
	}

	return newKey, nil
}

func (s *DeployKeysService) GenerateRandomSecret() string {
	secretBytes := make([]byte, 32)
	_, _ = rand.Read(secretBytes)
	return hex.EncodeToString(secretBytes)
}
