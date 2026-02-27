package verification

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// ApprovalService handles function approval workflows
type ApprovalService struct {
	repo *registry.RegistryRepository
}

// ApprovalRequest represents a request to create an approval
type ApprovalRequest struct {
	FunctionVersionID uuid.UUID  `json:"function_version_id"`
	ApprovalType      string     `json:"approval_type"`
	TrustLevel        string     `json:"trust_level"`
	Priority          string     `json:"priority"`
	RequestedBy       uuid.UUID  `json:"requested_by"`
	Comments          string     `json:"comments,omitempty"`
	ReviewDeadline    *time.Time `json:"review_deadline,omitempty"`
}

// ApprovalDecision represents a decision on an approval
type ApprovalDecision struct {
	ApprovalID      uuid.UUID        `json:"approval_id"`
	Decision        string           `json:"decision"` // "approve", "reject", "request_changes"
	ReviewerID      uuid.UUID        `json:"reviewer_id"`
	Comments        string           `json:"comments,omitempty"`
	RequiredActions []ApprovalAction `json:"required_actions,omitempty"`
}

// ApprovalAction represents an action required for approval
type ApprovalAction struct {
	Type        string                 `json:"type"` // "code_change", "documentation", "testing", "security_review"
	Description string                 `json:"description"`
	Status      string                 `json:"status"` // "pending", "completed", "cancelled"
	DueDate     *time.Time             `json:"due_date,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewApprovalService creates a new approval service
func NewApprovalService(repo *registry.RegistryRepository) *ApprovalService {
	return &ApprovalService{repo: repo}
}

// RequestApproval creates a new approval request
func (a *ApprovalService) RequestApproval(req ApprovalRequest) (*registry.RegistryFunctionApproval, error) {
	// Validate trust level requirements
	if !a.isValidTrustLevel(req.TrustLevel) {
		return nil, fmt.Errorf("invalid trust level: %s", req.TrustLevel)
	}

	// Check if approval already exists for this function version and type
	existingApprovals, err := a.repo.GetFunctionApprovals(req.FunctionVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing approvals: %w", err)
	}

	for _, approval := range existingApprovals {
		if approval.ApprovalType == req.ApprovalType && approval.Status == "pending" {
			return nil, fmt.Errorf("approval of type %s already pending for this function version", req.ApprovalType)
		}
	}

	// Set default priority if not specified
	if req.Priority == "" {
		req.Priority = a.getDefaultPriority(req.TrustLevel)
	}

	// Set default deadline if not specified
	if req.ReviewDeadline == nil {
		deadline := time.Now().Add(a.getDefaultDeadline(req.TrustLevel))
		req.ReviewDeadline = &deadline
	}

	// Create approval
	approval := &registry.RegistryFunctionApproval{
		FunctionVersionID: req.FunctionVersionID,
		ApprovalType:      req.ApprovalType,
		RequestedBy:       req.RequestedBy,
		Status:            "pending",
		Priority:          req.Priority,
		TrustLevel:        req.TrustLevel,
		ReviewDeadline:    req.ReviewDeadline,
		ReviewComments:    req.Comments,
	}

	if err := a.repo.CreateFunctionApproval(approval); err != nil {
		return nil, fmt.Errorf("failed to create approval: %w", err)
	}

	return approval, nil
}

// MakeDecision processes an approval decision
func (a *ApprovalService) MakeDecision(decision ApprovalDecision) error {
	// Get the approval
	approval, err := a.getApprovalByID(decision.ApprovalID)
	if err != nil {
		return fmt.Errorf("failed to get approval: %w", err)
	}

	if approval.Status != "pending" {
		return fmt.Errorf("approval is not in pending status")
	}

	// Validate decision
	if !a.isValidDecision(decision.Decision) {
		return fmt.Errorf("invalid decision: %s", decision.Decision)
	}

	updates := map[string]interface{}{
		"approved_by": decision.ReviewerID,
		"approved_at": time.Now(),
		"comments":    decision.Comments,
	}

	switch decision.Decision {
	case "approve":
		updates["status"] = "approved"
		// Mark required actions as completed if any
		if len(decision.RequiredActions) > 0 {
			actionsJSON, _ := json.Marshal(decision.RequiredActions)
			updates["completed_actions"] = actionsJSON
		}

	case "reject":
		updates["status"] = "rejected"

	case "request_changes":
		updates["status"] = "requires_changes"
		if len(decision.RequiredActions) > 0 {
			actionsJSON, _ := json.Marshal(decision.RequiredActions)
			updates["required_actions"] = actionsJSON
		}
	}

	if err := a.repo.UpdateFunctionApproval(decision.ApprovalID, updates); err != nil {
		return fmt.Errorf("failed to update approval: %w", err)
	}

	// Add comment if provided
	if decision.Comments != "" {
		comment := &registry.RegistryFunctionApprovalComment{
			ApprovalID: decision.ApprovalID,
			UserID:     decision.ReviewerID,
			Comment:    decision.Comments,
			IsInternal: false, // Reviewer comments are public
		}

		if err := a.repo.CreateApprovalComment(comment); err != nil {
			// Log error but don't fail the decision
			fmt.Printf("Failed to create approval comment: %v\n", err)
		}
	}

	return nil
}

// AssignReviewer assigns a reviewer to an approval
func (a *ApprovalService) AssignReviewer(approvalID, reviewerID uuid.UUID) error {
	updates := map[string]interface{}{
		"assigned_to": &reviewerID,
	}

	return a.repo.UpdateFunctionApproval(approvalID, updates)
}

// AddComment adds a comment to an approval
func (a *ApprovalService) AddComment(approvalID, userID uuid.UUID, comment string, isInternal bool) error {
	commentRecord := &registry.RegistryFunctionApprovalComment{
		ApprovalID: approvalID,
		UserID:     userID,
		Comment:    comment,
		IsInternal: isInternal,
	}

	return a.repo.CreateApprovalComment(commentRecord)
}

// GetApprovalStatus gets the current status of approvals for a function version
func (a *ApprovalService) GetApprovalStatus(functionVersionID uuid.UUID) (map[string]string, error) {
	approvals, err := a.repo.GetFunctionApprovals(functionVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get approvals: %w", err)
	}

	status := make(map[string]string)
	for _, approval := range approvals {
		status[approval.ApprovalType] = approval.Status
	}

	return status, nil
}

// CheckApprovalRequirements checks if a function version meets approval requirements for a trust level
func (a *ApprovalService) CheckApprovalRequirements(functionVersionID uuid.UUID, trustLevel string) (bool, []string, error) {
	requiredApprovals := a.getRequiredApprovals(trustLevel)
	missingApprovals := []string{}

	for _, approvalType := range requiredApprovals {
		approvals, err := a.repo.GetFunctionApprovals(functionVersionID)
		if err != nil {
			return false, nil, fmt.Errorf("failed to get approvals: %w", err)
		}

		found := false
		for _, approval := range approvals {
			if approval.ApprovalType == approvalType && approval.Status == "approved" {
				found = true
				break
			}
		}

		if !found {
			missingApprovals = append(missingApprovals, approvalType)
		}
	}

	isApproved := len(missingApprovals) == 0
	return isApproved, missingApprovals, nil
}

// GetOverdueApprovals gets approvals that are past their deadline
func (a *ApprovalService) GetOverdueApprovals() ([]registry.RegistryFunctionApproval, error) {
	// This would need a custom query - for now return empty
	return []registry.RegistryFunctionApproval{}, nil
}

// EscalateApproval escalates an approval to higher priority
func (a *ApprovalService) EscalateApproval(approvalID uuid.UUID, newPriority string) error {
	if !a.isValidPriority(newPriority) {
		return fmt.Errorf("invalid priority: %s", newPriority)
	}

	updates := map[string]interface{}{
		"priority": newPriority,
	}

	return a.repo.UpdateFunctionApproval(approvalID, updates)
}

// Helper methods

func (a *ApprovalService) isValidTrustLevel(trustLevel string) bool {
	validLevels := []string{"standard", "high", "enterprise"}
	for _, level := range validLevels {
		if level == trustLevel {
			return true
		}
	}
	return false
}

func (a *ApprovalService) isValidDecision(decision string) bool {
	validDecisions := []string{"approve", "reject", "request_changes"}
	for _, d := range validDecisions {
		if d == decision {
			return true
		}
	}
	return false
}

func (a *ApprovalService) isValidPriority(priority string) bool {
	validPriorities := []string{"low", "medium", "high", "critical"}
	for _, p := range validPriorities {
		if p == priority {
			return true
		}
	}
	return false
}

func (a *ApprovalService) getDefaultPriority(trustLevel string) string {
	switch trustLevel {
	case "enterprise":
		return "high"
	case "high":
		return "medium"
	default:
		return "low"
	}
}

func (a *ApprovalService) getDefaultDeadline(trustLevel string) time.Duration {
	switch trustLevel {
	case "enterprise":
		return 24 * time.Hour // 1 day
	case "high":
		return 7 * 24 * time.Hour // 1 week
	default:
		return 30 * 24 * time.Hour // 30 days
	}
}

func (a *ApprovalService) getRequiredApprovals(trustLevel string) []string {
	switch trustLevel {
	case "enterprise":
		return []string{"security_review", "code_review", "compliance"}
	case "high":
		return []string{"security_review", "code_review"}
	default:
		return []string{}
	}
}

func (a *ApprovalService) getApprovalByID(approvalID uuid.UUID) (*registry.RegistryFunctionApproval, error) {
	return a.repo.GetApprovalByID(approvalID)
}
