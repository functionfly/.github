package verification

import (
	"errors"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VerificationService orchestrates all function content verification
type VerificationService struct {
	repo           *registry.RegistryRepository
	signatureSvc   *SignatureService
	malwareScanner *MalwareScanner
	approvalSvc    *ApprovalService
}

// NewVerificationService creates a new verification service
func NewVerificationService(repo *registry.RegistryRepository, clamAVURL, yaraRulesURL string) *VerificationService {
	return &VerificationService{
		repo:           repo,
		signatureSvc:   NewSignatureService(repo),
		malwareScanner: NewMalwareScanner(repo, clamAVURL, yaraRulesURL),
		approvalSvc:    NewApprovalService(repo),
	}
}

// VerifyFunction performs complete verification of a function version
func (v *VerificationService) VerifyFunction(functionVersionID uuid.UUID, sourceCode string, wasmBinary []byte, trustLevel string) (*storage.RegistryFunctionVerificationStatus, error) {
	status := &storage.RegistryFunctionVerificationStatus{
		FunctionVersionID: functionVersionID,
		OverallStatus:     "verifying",
	}

	// 1. Content hash verification
	contentHashVerified := v.verifyContentHash(functionVersionID)
	status.ContentHashVerified = contentHashVerified

	// 2. Malware scanning
	malwareScan, err := v.malwareScanner.ScanFunction(functionVersionID, sourceCode, wasmBinary)
	if err != nil {
		return nil, fmt.Errorf("malware scan failed: %w", err)
	}

	status.MalwareScanned = true
	status.MalwareStatus = malwareScan.Status
	status.MalwareRiskScore = malwareScan.RiskScore

	// 3. Check if approval is required
	status.ApprovalRequired = v.isApprovalRequired(trustLevel)

	// 4. Determine overall status
	status.OverallStatus = v.calculateOverallStatus(status)

	// 5. Set next verification time (for periodic re-verification)
	nextVerification := time.Now().Add(7 * 24 * time.Hour) // Re-verify weekly
	status.NextVerificationAt = &nextVerification

	// Save verification status
	if err := v.repo.CreateOrUpdateVerificationStatus(status); err != nil {
		return nil, fmt.Errorf("failed to save verification status: %w", err)
	}

	// 6. Auto-request approvals if required
	if status.ApprovalRequired && status.OverallStatus != "blocked" {
		if err := v.requestApprovals(functionVersionID, trustLevel); err != nil {
			// Log error but don't fail verification
			fmt.Printf("Failed to request approvals: %v\n", err)
		}
	}

	return status, nil
}

// SignFunction signs a function version
func (v *VerificationService) SignFunction(functionVersionID uuid.UUID, signerID, privateKeyPEM, algorithm string) (*storage.RegistryFunctionSignature, error) {
	return v.signatureSvc.SignFunction(functionVersionID, signerID, privateKeyPEM, algorithm)
}

// VerifySignature verifies a function signature
func (v *VerificationService) VerifySignature(signatureID uuid.UUID, publicKeyPEM string) error {
	return v.signatureSvc.VerifySignature(signatureID, publicKeyPEM)
}

// RequestApproval requests approval for a function version
func (v *VerificationService) RequestApproval(req ApprovalRequest) (*storage.RegistryFunctionApproval, error) {
	return v.approvalSvc.RequestApproval(req)
}

// MakeApprovalDecision processes an approval decision
func (v *VerificationService) MakeApprovalDecision(decision ApprovalDecision) error {
	return v.approvalSvc.MakeDecision(decision)
}

// CheckExecutionAllowed checks if a function version is allowed to execute
func (v *VerificationService) CheckExecutionAllowed(functionVersionID uuid.UUID, author string) (bool, string, error) {
	// Allow trusted authors to bypass verification
	if v.isTrustedAuthor(author) {
		return true, "", nil
	}

	status, err := v.repo.GetVerificationStatus(functionVersionID)
	if err != nil {
		// No verification record (e.g. newly published function) -> allow execution
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, "", nil
		}
		return false, "verification status not found", err
	}

	// Check malware status
	if status.MalwareStatus == "malicious" {
		return false, "function contains malicious code", nil
	}

	// Check approval status
	if status.ApprovalRequired {
		if status.ApprovalStatus == "rejected" {
			return false, "function approval was rejected", nil
		}
		if status.ApprovalStatus == "pending" {
			return false, "function approval is pending", nil
		}
	}

	// Check overall status
	if status.OverallStatus == "blocked" || status.OverallStatus == "failed" {
		return false, fmt.Sprintf("function verification status: %s", status.OverallStatus), nil
	}

	return true, "", nil
}

// GetVerificationStatus gets the verification status for a function version
func (v *VerificationService) GetVerificationStatus(functionVersionID uuid.UUID) (*storage.RegistryFunctionVerificationStatus, error) {
	return v.repo.GetVerificationStatus(functionVersionID)
}

// Helper methods

func (v *VerificationService) verifyContentHash(functionVersionID uuid.UUID) bool {
	// Get function version
	version, err := v.repo.GetFunctionVersion(functionVersionID, "")
	if err != nil {
		return false
	}

	// For now, just check if we have a source hash
	// In a real implementation, we'd verify against a known good hash
	return version.SourceHash.Valid && version.SourceHash.String != ""
}

func (v *VerificationService) isApprovalRequired(trustLevel string) bool {
	switch trustLevel {
	case "high", "enterprise":
		return true
	default:
		return false
	}
}

func (v *VerificationService) calculateOverallStatus(status *storage.RegistryFunctionVerificationStatus) string {
	// Blocked if malware detected
	if status.MalwareStatus == "malicious" {
		return "blocked"
	}

	// Failed if malware scan failed
	if status.MalwareScanned && status.MalwareStatus == "error" {
		return "failed"
	}

	// Pending if approval required but not completed
	if status.ApprovalRequired && status.ApprovalStatus != "approved" {
		return "pending"
	}

	// Verified if all checks pass
	if status.ContentHashVerified && status.MalwareScanned && (!status.ApprovalRequired || status.ApprovalStatus == "approved") {
		return "verified"
	}

	// Unverified if no verification completed
	return "unverified"
}

func (v *VerificationService) requestApprovals(functionVersionID uuid.UUID, trustLevel string) error {
	// Get the function to find the owner
	version, err := v.repo.GetFunctionVersion(functionVersionID, "")
	if err != nil {
		return err
	}

	function, err := v.repo.GetFunctionByID(version.FunctionID)
	if err != nil {
		return err
	}

	var requestedBy uuid.UUID
	if function.OwnerUserID != nil {
		requestedBy = *function.OwnerUserID
	} else {
		// Use tenant ID as fallback - this should be improved
		return fmt.Errorf("no owner user ID found for function")
	}

	requiredApprovals := v.getRequiredApprovals(trustLevel)

	for _, approvalType := range requiredApprovals {
		req := ApprovalRequest{
			FunctionVersionID: functionVersionID,
			ApprovalType:      approvalType,
			TrustLevel:        trustLevel,
			Priority:          v.getApprovalPriority(trustLevel),
			RequestedBy:       requestedBy,
		}

		if _, err := v.approvalSvc.RequestApproval(req); err != nil {
			return err
		}
	}

	return nil
}

func (v *VerificationService) getRequiredApprovals(trustLevel string) []string {
	switch trustLevel {
	case "enterprise":
		return []string{"security_review", "code_review", "compliance"}
	case "high":
		return []string{"security_review", "code_review"}
	default:
		return []string{}
	}
}

func (v *VerificationService) getApprovalPriority(trustLevel string) string {
	switch trustLevel {
	case "enterprise":
		return "high"
	case "high":
		return "medium"
	default:
		return "low"
	}
}

// isTrustedAuthor checks if an author is trusted and can bypass verification
func (v *VerificationService) isTrustedAuthor(author string) bool {
	trustedAuthors := []string{
		"functionfly",      // Official FunctionFly account
		"functionfly-team", // Official team account
	}

	for _, trusted := range trustedAuthors {
		if author == trusted {
			return true
		}
	}

	return false
}
