package trustapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RevocationRepository handles database operations for trust revocations
type RevocationRepository struct {
	db *gorm.DB
}

// NewRevocationRepository creates a new revocation repository
func NewRevocationRepository(db *gorm.DB) *RevocationRepository {
	return &RevocationRepository{db: db}
}

// CreateRevocation creates a new trust revocation record
func (r *RevocationRepository) CreateRevocation(revocation *TrustRevocation) error {
	if revocation.ID == uuid.Nil {
		revocation.ID = uuid.New()
	}

	// Generate public revocation ID
	b := make([]byte, 12)
	rand.Read(b)
	revocation.RevocationID = "rvk_" + hex.EncodeToString(b)

	if revocation.Status == "" {
		revocation.Status = string(RevocationStatusActive)
	}
	if revocation.RevocationType == "" {
		revocation.RevocationType = "full"
	}
	if revocation.RevokedAt.IsZero() {
		revocation.RevokedAt = time.Now()
	}

	return r.db.Create(revocation).Error
}

// GetRevocationByID retrieves a revocation by ID
func (r *RevocationRepository) GetRevocationByID(id uuid.UUID) (*TrustRevocation, error) {
	var revocation TrustRevocation
	err := r.db.First(&revocation, id).Error
	if err != nil {
		return nil, err
	}
	return &revocation, nil
}

// GetRevocationByRevocationID retrieves a revocation by its public ID
func (r *RevocationRepository) GetRevocationByRevocationID(revocationID string) (*TrustRevocation, error) {
	var revocation TrustRevocation
	err := r.db.Where("revocation_id = ?", revocationID).First(&revocation).Error
	if err != nil {
		return nil, err
	}
	return &revocation, nil
}

// GetActiveRevocationForFunction checks if a function has an active revocation
func (r *RevocationRepository) GetActiveRevocationForFunction(functionID uuid.UUID) (*TrustRevocation, error) {
	var revocation TrustRevocation
	err := r.db.Where("function_id = ? AND status = ?", functionID, RevocationStatusActive).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("created_at DESC").
		First(&revocation).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &revocation, nil
}

// ListRevocations lists revocations with filtering options
func (r *RevocationRepository) ListRevocations(
	functionID *uuid.UUID,
	status string,
	reason string,
	severity string,
	revokedBy *uuid.UUID,
	limit, offset int,
) ([]TrustRevocation, int64, error) {
	var revocations []TrustRevocation
	var total int64

	query := r.db.Model(&TrustRevocation{})

	if functionID != nil {
		query = query.Where("function_id = ?", *functionID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if reason != "" {
		query = query.Where("reason = ?", reason)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if revokedBy != nil {
		query = query.Where("revoked_by = ?", *revokedBy)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&revocations).Error; err != nil {
		return nil, 0, err
	}

	return revocations, total, nil
}

// ListActiveRevocations lists all currently active revocations
func (r *RevocationRepository) ListActiveRevocations(limit, offset int) ([]TrustRevocation, int64, error) {
	return r.ListRevocations(nil, string(RevocationStatusActive), "", "", nil, limit, offset)
}

// LiftRevocation lifts (reverses) a trust revocation
func (r *RevocationRepository) LiftRevocation(revocationID uuid.UUID, liftedBy uuid.UUID, reason string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":      RevocationStatusLifted,
		"lifted_at":   &now,
		"lifted_by":   liftedBy,
		"lift_reason": reason,
	}

	return r.db.Model(&TrustRevocation{}).
		Where("id = ? AND status = ?", revocationID, RevocationStatusActive).
		Updates(updates).Error
}

// ExpireRevocation marks a revocation as expired (called by scheduled job)
func (r *RevocationRepository) ExpireRevocation(revocationID uuid.UUID) error {
	return r.db.Model(&TrustRevocation{}).
		Where("id = ?", revocationID).
		Update("status", RevocationStatusExpired).Error
}

// UpdateAppealStatus updates the appeal status of a revocation
func (r *RevocationRepository) UpdateAppealStatus(revocationID uuid.UUID, status string, appealReason string) error {
	updates := map[string]interface{}{
		"appeal_status": status,
	}

	if appealReason != "" {
		updates["appeal_reason"] = appealReason
		now := time.Now()
		updates["appeal_submitted_at"] = &now
	}

	return r.db.Model(&TrustRevocation{}).
		Where("id = ?", revocationID).
		Updates(updates).Error
}

// MarkAsNotified marks that users have been notified of the revocation
func (r *RevocationRepository) MarkAsNotified(revocationID uuid.UUID) error {
	return r.db.Model(&TrustRevocation{}).
		Where("id = ?", revocationID).
		Update("notified_users", true).Error
}

// GetExpiredRevocations gets all revocations that have passed their expiration time
func (r *RevocationRepository) GetExpiredRevocations() ([]TrustRevocation, error) {
	var revocations []TrustRevocation
	err := r.db.Where("status = ? AND expires_at IS NOT NULL AND expires_at < ?",
		RevocationStatusActive, time.Now()).
		Find(&revocations).Error
	return revocations, err
}

// ============================================
// Attestation Repository Methods
// ============================================

// CreateAttestation creates a new attestation record
func (r *RevocationRepository) CreateAttestation(attestation *TrustAttestation) error {
	if attestation.ID == uuid.Nil {
		attestation.ID = uuid.New()
	}

	// Generate public attestation ID
	b := make([]byte, 12)
	rand.Read(b)
	attestation.AttestationID = "att_" + hex.EncodeToString(b)

	// Calculate proof hash for immutability
	attestation.ProofHash = attestation.CalculateProofHash()

	if attestation.Status == "" {
		attestation.Status = string(AttestationStatusValid)
	}
	if attestation.AttestedAt.IsZero() {
		attestation.AttestedAt = time.Now()
	}

	return r.db.Create(attestation).Error
}

// GetAttestationByID retrieves an attestation by ID
func (r *RevocationRepository) GetAttestationByID(id uuid.UUID) (*TrustAttestation, error) {
	var attestation TrustAttestation
	err := r.db.First(&attestation, id).Error
	if err != nil {
		return nil, err
	}
	return &attestation, nil
}

// GetAttestationByAttestationID retrieves an attestation by its public ID
func (r *RevocationRepository) GetAttestationByAttestationID(attestationID string) (*TrustAttestation, error) {
	var attestation TrustAttestation
	err := r.db.Where("attestation_id = ?", attestationID).First(&attestation).Error
	if err != nil {
		return nil, err
	}
	return &attestation, nil
}

// ListAttestationsForFunction lists all attestations for a function
func (r *RevocationRepository) ListAttestationsForFunction(
	functionID uuid.UUID,
	attestationType string,
	status string,
	includeRevoked bool,
	limit, offset int,
) ([]TrustAttestation, int64, error) {
	var attestations []TrustAttestation
	var total int64

	query := r.db.Model(&TrustAttestation{}).Where("function_id = ?", functionID)

	if attestationType != "" {
		query = query.Where("type = ?", attestationType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if !includeRevoked {
		query = query.Where("status != ?", AttestationStatusRevoked)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	if err := query.Order("attested_at DESC").Limit(limit).Offset(offset).Find(&attestations).Error; err != nil {
		return nil, 0, err
	}

	return attestations, total, nil
}

// GetLatestAttestationForType gets the most recent attestation of a specific type
func (r *RevocationRepository) GetLatestAttestationForType(functionID uuid.UUID, attestationType string) (*TrustAttestation, error) {
	var attestation TrustAttestation
	err := r.db.Where("function_id = ? AND type = ? AND status = ?",
		functionID, attestationType, AttestationStatusValid).
		Order("attested_at DESC").
		First(&attestation).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &attestation, nil
}

// RevokeAttestation revokes an attestation (creates a new attestation record for the revocation)
func (r *RevocationRepository) RevokeAttestation(
	attestationID uuid.UUID,
	revokedBy uuid.UUID,
	reason string,
	revocationID *uuid.UUID,
) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":        AttestationStatusRevoked,
		"revoked_at":    &now,
		"revoked_by":    revokedBy,
		"revoke_reason": reason,
	}

	if revocationID != nil {
		updates["revocation_id"] = *revocationID
	}

	return r.db.Model(&TrustAttestation{}).
		Where("id = ? AND status = ?", attestationID, AttestationStatusValid).
		Updates(updates).Error
}

// GetAttestationChain gets the chain of attestations for a function (for audit)
func (r *RevocationRepository) GetAttestationChain(functionID uuid.UUID) ([]TrustAttestation, error) {
	var attestations []TrustAttestation
	err := r.db.Where("function_id = ?", functionID).
		Order("attested_at ASC").
		Find(&attestations).Error
	return attestations, err
}

// VerifyAttestationIntegrity verifies that an attestation hasn't been tampered with
func (r *RevocationRepository) VerifyAttestationIntegrity(attestationID string) (bool, error) {
	attestation, err := r.GetAttestationByAttestationID(attestationID)
	if err != nil {
		return false, err
	}

	return attestation.VerifyIntegrity(), nil
}

// ============================================
// Policy Repository Methods
// ============================================

// CreatePolicy creates a new trust policy
func (r *RevocationRepository) CreatePolicy(policy *TrustPolicy) error {
	if policy.ID == uuid.Nil {
		policy.ID = uuid.New()
	}

	// Generate public policy ID
	b := make([]byte, 12)
	rand.Read(b)
	policy.PolicyID = "pol_" + hex.EncodeToString(b)

	// Marshal rules to JSON
	rulesJSON, err := json.Marshal(policy.Rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}
	policy.Rules = rulesJSON

	if policy.Status == "" {
		policy.Status = "active"
	}
	if policy.DefaultAction == "" {
		policy.DefaultAction = "deny"
	}
	if policy.ValidFrom.IsZero() {
		policy.ValidFrom = time.Now()
	}

	return r.db.Create(policy).Error
}

// GetPolicyByID retrieves a policy by ID
func (r *RevocationRepository) GetPolicyByID(id uuid.UUID) (*TrustPolicy, error) {
	var policy TrustPolicy
	err := r.db.First(&policy, id).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// GetPolicyByPolicyID retrieves a policy by its public ID
func (r *RevocationRepository) GetPolicyByPolicyID(policyID string) (*TrustPolicy, error) {
	var policy TrustPolicy
	err := r.db.Where("policy_id = ?", policyID).First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// GetDefaultPolicyForOwner gets the default policy for an owner
func (r *RevocationRepository) GetDefaultPolicyForOwner(ownerID uuid.UUID, ownerType string) (*TrustPolicy, error) {
	var policy TrustPolicy
	err := r.db.Where("owner_id = ? AND owner_type = ? AND is_default = ? AND status = ?",
		ownerID, ownerType, true, "active").
		Order("created_at DESC").
		First(&policy).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// ListPoliciesForOwner lists all policies for an owner
func (r *RevocationRepository) ListPoliciesForOwner(
	ownerID uuid.UUID,
	ownerType string,
	status string,
	limit, offset int,
) ([]TrustPolicy, int64, error) {
	var policies []TrustPolicy
	var total int64

	query := r.db.Model(&TrustPolicy{}).Where("owner_id = ? AND owner_type = ?", ownerID, ownerType)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&policies).Error; err != nil {
		return nil, 0, err
	}

	return policies, total, nil
}

// UpdatePolicy updates a policy
func (r *RevocationRepository) UpdatePolicy(policy *TrustPolicy) error {
	// Marshal rules if they changed
	if policy.Rules != nil {
		rulesJSON, err := json.Marshal(policy.Rules)
		if err != nil {
			return fmt.Errorf("failed to marshal rules: %w", err)
		}
		policy.Rules = rulesJSON
	}

	return r.db.Save(policy).Error
}

// SetDefaultPolicy sets a policy as the default for an owner (unsets others)
func (r *RevocationRepository) SetDefaultPolicy(policyID uuid.UUID, ownerID uuid.UUID, ownerType string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Unset existing default
		if err := tx.Model(&TrustPolicy{}).
			Where("owner_id = ? AND owner_type = ? AND is_default = ?", ownerID, ownerType, true).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// Set new default
		return tx.Model(&TrustPolicy{}).
			Where("id = ?", policyID).
			Update("is_default", true).Error
	})
}

// DeprecatePolicy marks a policy as deprecated
func (r *RevocationRepository) DeprecatePolicy(policyID uuid.UUID, deprecatedBy uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&TrustPolicy{}).
		Where("id = ?", policyID).
		Updates(map[string]interface{}{
			"status":        "deprecated",
			"deprecated_at": &now,
			"deprecated_by": deprecatedBy,
		}).Error
}

// IncrementPolicyUseCount increments the use counter for a policy
func (r *RevocationRepository) IncrementPolicyUseCount(policyID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&TrustPolicy{}).
		Where("id = ?", policyID).
		Updates(map[string]interface{}{
			"use_count":    gorm.Expr("use_count + 1"),
			"last_used_at": &now,
		}).Error
}

// ============================================
// Policy Evaluation Repository Methods
// ============================================

// CreateEvaluation creates a new policy evaluation record
func (r *RevocationRepository) CreateEvaluation(eval *TrustPolicyEvaluation) error {
	if eval.ID == uuid.Nil {
		eval.ID = uuid.New()
	}

	// Generate public evaluation ID
	b := make([]byte, 12)
	rand.Read(b)
	eval.EvaluationID = "evl_" + hex.EncodeToString(b)

	if eval.EvaluatedAt.IsZero() {
		eval.EvaluatedAt = time.Now()
	}

	return r.db.Create(eval).Error
}

// GetEvaluationByID retrieves an evaluation by ID
func (r *RevocationRepository) GetEvaluationByID(id uuid.UUID) (*TrustPolicyEvaluation, error) {
	var eval TrustPolicyEvaluation
	err := r.db.First(&eval, id).Error
	if err != nil {
		return nil, err
	}
	return &eval, nil
}

// GetEvaluationByEvaluationID retrieves an evaluation by its public ID
func (r *RevocationRepository) GetEvaluationByEvaluationID(evaluationID string) (*TrustPolicyEvaluation, error) {
	var eval TrustPolicyEvaluation
	err := r.db.Where("evaluation_id = ?", evaluationID).First(&eval).Error
	if err != nil {
		return nil, err
	}
	return &eval, nil
}

// GetCachedEvaluation gets a cached evaluation that's still valid
func (r *RevocationRepository) GetCachedEvaluation(policyID, functionID uuid.UUID) (*TrustPolicyEvaluation, error) {
	var eval TrustPolicyEvaluation
	err := r.db.Where("policy_id = ? AND function_id = ? AND is_cached = ?",
		policyID, functionID, true).
		Where("cache_valid_until IS NULL OR cache_valid_until > ?", time.Now()).
		Order("evaluated_at DESC").
		First(&eval).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &eval, nil
}

// ListEvaluations lists policy evaluations with filtering
func (r *RevocationRepository) ListEvaluations(
	policyID *uuid.UUID,
	functionID *uuid.UUID,
	result string,
	limit, offset int,
) ([]TrustPolicyEvaluation, int64, error) {
	var evaluations []TrustPolicyEvaluation
	var total int64

	query := r.db.Model(&TrustPolicyEvaluation{})

	if policyID != nil {
		query = query.Where("policy_id = ?", *policyID)
	}
	if functionID != nil {
		query = query.Where("function_id = ?", *functionID)
	}
	if result != "" {
		query = query.Where("result = ?", result)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	if err := query.Order("evaluated_at DESC").Limit(limit).Offset(offset).Find(&evaluations).Error; err != nil {
		return nil, 0, err
	}

	return evaluations, total, nil
}

// InvalidateCacheForFunction invalidates cached evaluations for a function
func (r *RevocationRepository) InvalidateCacheForFunction(functionID uuid.UUID) error {
	return r.db.Model(&TrustPolicyEvaluation{}).
		Where("function_id = ? AND is_cached = ?", functionID, true).
		Update("is_cached", false).Error
}

// CleanExpiredEvaluations removes old cached evaluations
func (r *RevocationRepository) CleanExpiredEvalocations(olderThan time.Time) error {
	return r.db.Where("evaluated_at < ? AND is_cached = ?", olderThan, true).
		Delete(&TrustPolicyEvaluation{}).Error
}
