package verification

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// Common errors
var (
	ErrReviewNotFound     = errors.New("manual review not found")
	ErrReviewNotEditable  = errors.New("manual review cannot be edited in current state")
	ErrInvalidDecision   = errors.New("invalid decision, must be approved, rejected, or escalated")
	ErrUnauthorizedReview = errors.New("user is not authorized to perform this review action")
)

// ManualReviewRepositoryInterface defines the interface for manual review data access
type ManualReviewRepositoryInterface interface {
	CreateManualReviewQueue(ctx context.Context, review *registry.ManualReviewQueue) error
	GetManualReviewQueueByID(ctx context.Context, id uuid.UUID) (*registry.ManualReviewQueue, error)
	GetManualReviewQueueByFunctionVersion(ctx context.Context, functionVersionID uuid.UUID) (*registry.ManualReviewQueue, error)
	UpdateManualReviewQueue(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	AssignManualReviewQueue(ctx context.Context, id, reviewerID uuid.UUID) error
	CompleteManualReviewQueue(ctx context.Context, id, reviewerID uuid.UUID, decision, reason string) error
	EscalateManualReviewQueue(ctx context.Context, id uuid.UUID, reason string) error
	GetPendingManualReviews(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]registry.ManualReviewQueue, int64, error)
	GetOverdueManualReviews(ctx context.Context, limit int) ([]registry.ManualReviewQueue, error)
	GetManualReviewsByReviewer(ctx context.Context, reviewerID uuid.UUID, statusFilter []string, limit, offset int) ([]registry.ManualReviewQueue, int64, error)
	DeleteManualReviewQueue(ctx context.Context, id uuid.UUID) error
}

// ManualReviewService handles human review queue for Level 3 verification
type ManualReviewService struct {
	repo registry.ManualReviewRepositoryInterface
}

// NewManualReviewService creates a new manual review service
func NewManualReviewService(repo registry.ManualReviewRepositoryInterface) (*ManualReviewService, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository is required")
	}
	return &ManualReviewService{repo: repo}, nil
}

// ManualReview represents a manual review entry (view model for API responses)
type ManualReview struct {
	ID                uuid.UUID  `json:"id"`
	FunctionID        uuid.UUID  `json:"function_id"`
	FunctionVersionID uuid.UUID  `json:"function_version_id"`
	VerificationJobID *uuid.UUID `json:"verification_job_id,omitempty"`
	Status            string     `json:"status"`
	Priority          string     `json:"priority"`
	ReviewType        string     `json:"review_type"`
	AssignedTo        *uuid.UUID `json:"assigned_to,omitempty"`
	ReviewNotes       string     `json:"review_notes,omitempty"`
	ReviewComments    string     `json:"review_comments,omitempty"`
	DecisionAt        *time.Time `json:"decision_at,omitempty"`
	DecisionBy        *uuid.UUID `json:"decision_by,omitempty"`
	DecisionReason    string     `json:"decision_reason,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	DueAt             *time.Time `json:"due_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
}

// toModel converts a database ManualReviewQueue to a ManualReview view model
func toModel(q *registry.ManualReviewQueue) *ManualReview {
	if q == nil {
		return nil
	}
	return &ManualReview{
		ID:                q.ID,
		FunctionID:        q.FunctionID,
		FunctionVersionID: q.FunctionVersionID,
		VerificationJobID: q.VerificationJobID,
		Status:            q.Status,
		Priority:          q.Priority,
		ReviewType:        q.ReviewType,
		AssignedTo:        q.AssignedTo,
		ReviewNotes:       q.ReviewNotes,
		ReviewComments:    q.ReviewComments,
		DecisionAt:        q.DecisionAt,
		DecisionBy:        q.DecisionBy,
		DecisionReason:    q.DecisionReason,
		CreatedAt:         q.CreatedAt,
		DueAt:             q.DueAt,
		CompletedAt:       q.CompletedAt,
	}
}

// CreateReviewRequest contains parameters for creating a new review
type CreateReviewRequest struct {
	FunctionID        uuid.UUID
	FunctionVersionID uuid.UUID
	VerificationJobID *uuid.UUID
	Priority          string
	ReviewType        string
	DueInDays         int
}

// CreateReview creates a new manual review entry
func (s *ManualReviewService) CreateReview(ctx context.Context, req CreateReviewRequest) (*ManualReview, error) {
	priority := req.Priority
	if priority == "" {
		priority = "normal"
	}

	reviewType := req.ReviewType
	if reviewType == "" {
		reviewType = "security"
	}

	dueInDays := req.DueInDays
	if dueInDays <= 0 {
		dueInDays = 7 // Default 7 days
	}

	review := &registry.ManualReviewQueue{
		ID:                uuid.New(),
		FunctionID:        req.FunctionID,
		FunctionVersionID: req.FunctionVersionID,
		VerificationJobID: req.VerificationJobID,
		Status:            "pending",
		Priority:          priority,
		ReviewType:        reviewType,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	dueAt := time.Now().Add(time.Duration(dueInDays) * 24 * time.Hour)
	review.DueAt = &dueAt

	if err := s.repo.CreateManualReviewQueue(ctx, review); err != nil {
		return nil, fmt.Errorf("failed to create manual review: %w", err)
	}

	return toModel(review), nil
}

// GetReview retrieves a review by ID
func (s *ManualReviewService) GetReview(ctx context.Context, reviewID uuid.UUID) (*ManualReview, error) {
	review, err := s.repo.GetManualReviewQueueByID(ctx, reviewID)
	if err != nil {
		if errors.Is(err, registry.ErrReviewNotFound) {
			return nil, ErrReviewNotFound
		}
		return nil, fmt.Errorf("failed to get manual review: %w", err)
	}

	return toModel(review), nil
}

// GetReviewByFunctionVersion retrieves a review by function version
func (s *ManualReviewService) GetReviewByFunctionVersion(ctx context.Context, functionVersionID uuid.UUID) (*ManualReview, error) {
	review, err := s.repo.GetManualReviewQueueByFunctionVersion(ctx, functionVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get manual review by function version: %w", err)
	}

	return toModel(review), nil
}

// AssignReviewRequest contains parameters for assigning a review
type AssignReviewRequest struct {
	ReviewID    uuid.UUID
	ReviewerID  uuid.UUID
}

// AssignReview assigns a review to a reviewer
func (s *ManualReviewService) AssignReview(ctx context.Context, req AssignReviewRequest) error {
	// Verify review exists and is in editable state
	review, err := s.repo.GetManualReviewQueueByID(ctx, req.ReviewID)
	if err != nil {
		return fmt.Errorf("failed to get review for assignment: %w", err)
	}

	if !isReviewEditable(review.Status) {
		return ErrReviewNotEditable
	}

	if err := s.repo.AssignManualReviewQueue(ctx, req.ReviewID, req.ReviewerID); err != nil {
		return fmt.Errorf("failed to assign review: %w", err)
	}

	return nil
}

// DecisionRequest contains parameters for making a review decision
type DecisionRequest struct {
	ReviewID   uuid.UUID
	ReviewerID uuid.UUID
	Decision   string // "approved", "rejected", "escalated"
	Reason     string
}

// MakeDecision makes a decision on a review
func (s *ManualReviewService) MakeDecision(ctx context.Context, req DecisionRequest) error {
	if !isValidDecision(req.Decision) {
		return ErrInvalidDecision
	}

	// Verify review exists
	review, err := s.repo.GetManualReviewQueueByID(ctx, req.ReviewID)
	if err != nil {
		return fmt.Errorf("failed to get review for decision: %w", err)
	}

	// Verify reviewer is assigned (unless they're an admin escalating)
	if review.AssignedTo != nil && *review.AssignedTo != req.ReviewerID {
		if req.Decision != "escalated" {
			return ErrUnauthorizedReview
		}
	}

	// Verify review is in a decider state
	if !isReviewDecidable(review.Status) {
		return ErrReviewNotEditable
	}

	if err := s.repo.CompleteManualReviewQueue(ctx, req.ReviewID, req.ReviewerID, req.Decision, req.Reason); err != nil {
		return fmt.Errorf("failed to complete review: %w", err)
	}

	return nil
}

// ListReviewsRequest contains parameters for listing reviews
type ListReviewsRequest struct {
	Status     string     // filter by status
	Priority   string     // filter by priority
	ReviewType string     // filter by review type
	AssignedTo *uuid.UUID // filter by assigned reviewer
	FunctionID *uuid.UUID // filter by function
	Limit      int
	Offset     int
}

// ListReviewsResponse contains the response for listing reviews
type ListReviewsResponse struct {
	Reviews []ManualReview `json:"reviews"`
	Total   int64          `json:"total"`
}

// GetPendingReviews retrieves all pending reviews with optional filters
func (s *ManualReviewService) GetPendingReviews(ctx context.Context, req ListReviewsRequest) (*ListReviewsResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	filters := make(map[string]interface{})
	if req.Status != "" {
		filters["status"] = req.Status
	}
	if req.Priority != "" {
		filters["priority"] = req.Priority
	}
	if req.ReviewType != "" {
		filters["review_type"] = req.ReviewType
	}
	if req.AssignedTo != nil {
		filters["assigned_to"] = *req.AssignedTo
	}
	if req.FunctionID != nil {
		filters["function_id"] = *req.FunctionID
	}

	reviews, total, err := s.repo.GetPendingManualReviews(ctx, filters, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending reviews: %w", err)
	}

	result := make([]ManualReview, len(reviews))
	for i, r := range reviews {
		result[i] = *toModel(&r)
	}

	return &ListReviewsResponse{
		Reviews: result,
		Total:   total,
	}, nil
}

// GetReviewerReviews retrieves all reviews assigned to a specific reviewer
func (s *ManualReviewService) GetReviewerReviews(ctx context.Context, reviewerID uuid.UUID, statusFilter []string, limit, offset int) (*ListReviewsResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	reviews, total, err := s.repo.GetManualReviewsByReviewer(ctx, reviewerID, statusFilter, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewer reviews: %w", err)
	}

	result := make([]ManualReview, len(reviews))
	for i, r := range reviews {
		result[i] = *toModel(&r)
	}

	return &ListReviewsResponse{
		Reviews: result,
		Total:   total,
	}, nil
}

// EscalateReview escalates a review to higher priority
func (s *ManualReviewService) EscalateReview(ctx context.Context, reviewID uuid.UUID, reason string) error {
	review, err := s.repo.GetManualReviewQueueByID(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("failed to get review for escalation: %w", err)
	}

	if !isReviewEditable(review.Status) {
		return ErrReviewNotEditable
	}

	if err := s.repo.EscalateManualReviewQueue(ctx, reviewID, reason); err != nil {
		return fmt.Errorf("failed to escalate review: %w", err)
	}

	return nil
}

// GetOverdueReviews retrieves reviews past their due date
func (s *ManualReviewService) GetOverdueReviews(ctx context.Context, limit int) ([]ManualReview, error) {
	if limit <= 0 {
		limit = 50
	}

	reviews, err := s.repo.GetOverdueManualReviews(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue reviews: %w", err)
	}

	result := make([]ManualReview, len(reviews))
	for i, r := range reviews {
		result[i] = *toModel(&r)
	}

	return result, nil
}

// UpdateReviewNotes updates the review notes
func (s *ManualReviewService) UpdateReviewNotes(ctx context.Context, reviewID uuid.UUID, reviewerID uuid.UUID, notes string) error {
	review, err := s.repo.GetManualReviewQueueByID(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("failed to get review for update: %w", err)
	}

	// Only assigned reviewer can update notes
	if review.AssignedTo == nil || *review.AssignedTo != reviewerID {
		return ErrUnauthorizedReview
	}

	if !isReviewEditable(review.Status) {
		return ErrReviewNotEditable
	}

	if err := s.repo.UpdateManualReviewQueue(ctx, reviewID, map[string]interface{}{
		"review_notes": notes,
	}); err != nil {
		return fmt.Errorf("failed to update review notes: %w", err)
	}

	return nil
}

// Helper functions

func isReviewEditable(status string) bool {
	return status == "pending" || status == "in_review"
}

func isReviewDecidable(status string) bool {
	return status == "pending" || status == "in_review" || status == "escalated"
}

func isValidDecision(decision string) bool {
	return decision == "approved" || decision == "rejected" || decision == "escalated"
}
