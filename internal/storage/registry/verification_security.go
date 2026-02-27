package registry

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateFunctionSignature creates a new function signature
func (r *RegistryRepository) CreateFunctionSignature(sig *RegistryFunctionSignature) error {
	sig.ID = uuid.New()
	sig.CreatedAt = time.Now()
	sig.UpdatedAt = time.Now()

	if err := r.db.Create(sig).Error; err != nil {
		return fmt.Errorf("failed to create function signature: %w", err)
	}

	return nil
}

// GetFunctionSignatures retrieves all signatures for a function version
func (r *RegistryRepository) GetFunctionSignatures(functionVersionID uuid.UUID) ([]RegistryFunctionSignature, error) {
	var signatures []RegistryFunctionSignature
	if err := r.db.Where("function_version_id = ?", functionVersionID).Find(&signatures).Error; err != nil {
		return nil, fmt.Errorf("failed to get function signatures: %w", err)
	}

	return signatures, nil
}

// CreateMalwareScan creates a new malware scan result
func (r *RegistryRepository) CreateMalwareScan(scan *RegistryFunctionMalwareScan) error {
	scan.ID = uuid.New()
	scan.CreatedAt = time.Now()
	scan.UpdatedAt = time.Now()

	if err := r.db.Create(scan).Error; err != nil {
		return fmt.Errorf("failed to create malware scan: %w", err)
	}

	return nil
}

// GetLatestMalwareScan retrieves the latest malware scan for a function version
func (r *RegistryRepository) GetLatestMalwareScan(functionVersionID uuid.UUID) (*RegistryFunctionMalwareScan, error) {
	var scan RegistryFunctionMalwareScan
	if err := r.db.Where("function_version_id = ?", functionVersionID).
		Order("scanned_at DESC").First(&scan).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest malware scan: %w", err)
	}

	return &scan, nil
}

// CreateFunctionApproval creates a new approval request
func (r *RegistryRepository) CreateFunctionApproval(approval *RegistryFunctionApproval) error {
	approval.ID = uuid.New()
	approval.CreatedAt = time.Now()
	approval.UpdatedAt = time.Now()

	if err := r.db.Create(approval).Error; err != nil {
		return fmt.Errorf("failed to create function approval: %w", err)
	}

	return nil
}

// GetFunctionApprovals retrieves all approvals for a function version
func (r *RegistryRepository) GetFunctionApprovals(functionVersionID uuid.UUID) ([]RegistryFunctionApproval, error) {
	var approvals []RegistryFunctionApproval
	if err := r.db.Where("function_version_id = ?", functionVersionID).
		Preload("RequestedBy").
		Preload("AssignedTo").
		Preload("ApprovedBy").
		Find(&approvals).Error; err != nil {
		return nil, fmt.Errorf("failed to get function approvals: %w", err)
	}

	return approvals, nil
}

// UpdateFunctionApproval updates an approval status
func (r *RegistryRepository) UpdateFunctionApproval(approvalID uuid.UUID, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	if err := r.db.Model(&RegistryFunctionApproval{}).
		Where("id = ?", approvalID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update function approval: %w", err)
	}

	return nil
}

// CreateApprovalComment creates a new approval comment
func (r *RegistryRepository) CreateApprovalComment(comment *RegistryFunctionApprovalComment) error {
	comment.ID = uuid.New()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	if err := r.db.Create(comment).Error; err != nil {
		return fmt.Errorf("failed to create approval comment: %w", err)
	}

	return nil
}

// GetApprovalComments retrieves all comments for an approval
func (r *RegistryRepository) GetApprovalComments(approvalID uuid.UUID) ([]RegistryFunctionApprovalComment, error) {
	var comments []RegistryFunctionApprovalComment
	if err := r.db.Where("approval_id = ?", approvalID).
		Preload("User").
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("failed to get approval comments: %w", err)
	}

	return comments, nil
}

// CreateOrUpdateVerificationStatus creates or updates the verification status for a function version
func (r *RegistryRepository) CreateOrUpdateVerificationStatus(status *RegistryFunctionVerificationStatus) error {
	status.UpdatedAt = time.Now()

	// Try to update existing status
	result := r.db.Model(&RegistryFunctionVerificationStatus{}).
		Where("function_version_id = ?", status.FunctionVersionID).
		Updates(status)

	if result.Error != nil {
		return fmt.Errorf("failed to update verification status: %w", result.Error)
	}

	// If no rows were affected, create new status
	if result.RowsAffected == 0 {
		status.ID = uuid.New()
		status.CreatedAt = time.Now()
		if err := r.db.Create(status).Error; err != nil {
			return fmt.Errorf("failed to create verification status: %w", err)
		}
	}

	return nil
}

// GetVerificationStatus retrieves the verification status for a function version
func (r *RegistryRepository) GetVerificationStatus(functionVersionID uuid.UUID) (*RegistryFunctionVerificationStatus, error) {
	var status RegistryFunctionVerificationStatus
	if err := r.db.Where("function_version_id = ?", functionVersionID).First(&status).Error; err != nil {
		return nil, fmt.Errorf("failed to get verification status: %w", err)
	}

	return &status, nil
}

// GetPendingApprovals retrieves approvals that need review
func (r *RegistryRepository) GetPendingApprovals(limit, offset int) ([]RegistryFunctionApproval, error) {
	var approvals []RegistryFunctionApproval
	if err := r.db.Where("status = ?", "pending").
		Preload("RequestedBy").
		Preload("AssignedTo").
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&approvals).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending approvals: %w", err)
	}

	return approvals, nil
}

// GetApprovalsByTrustLevel retrieves approvals by trust level
func (r *RegistryRepository) GetApprovalsByTrustLevel(trustLevel string, limit, offset int) ([]RegistryFunctionApproval, error) {
	var approvals []RegistryFunctionApproval
	if err := r.db.Where("trust_level = ? AND status = ?", trustLevel, "pending").
		Preload("RequestedBy").
		Preload("AssignedTo").
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&approvals).Error; err != nil {
		return nil, fmt.Errorf("failed to get approvals by trust level: %w", err)
	}

	return approvals, nil
}

// GetSignatureByID retrieves a signature by ID
func (r *RegistryRepository) GetSignatureByID(signatureID uuid.UUID) (*RegistryFunctionSignature, error) {
	var sig RegistryFunctionSignature
	if err := r.db.First(&sig, signatureID).Error; err != nil {
		return nil, fmt.Errorf("failed to get signature by ID: %w", err)
	}

	return &sig, nil
}

// UpdateSignature updates a signature record
func (r *RegistryRepository) UpdateSignature(signatureID uuid.UUID, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	if err := r.db.Model(&RegistryFunctionSignature{}).
		Where("id = ?", signatureID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update signature: %w", err)
	}

	return nil
}

// GetApprovalByID retrieves an approval by ID
func (r *RegistryRepository) GetApprovalByID(approvalID uuid.UUID) (*RegistryFunctionApproval, error) {
	var approval RegistryFunctionApproval
	if err := r.db.Preload("RequestedBy").
		Preload("AssignedTo").
		Preload("ApprovedBy").
		First(&approval, approvalID).Error; err != nil {
		return nil, fmt.Errorf("failed to get approval by ID: %w", err)
	}

	return &approval, nil
}