package flywheel

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/services"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service provides business logic for Flywheel Network
type Service struct {
	repo        *Repository
	execService ExecutionService
	logger      *logrus.Logger
}

// ExecutionService interface for code execution
type ExecutionService interface {
	ExecuteCode(ctx context.Context, code string, language string, input json.RawMessage) (*ExecutionResult, error)
	VerifyOutput(ctx context.Context, actual, expected json.RawMessage) (bool, string, error)
}

// ExecutionResult represents the result of code execution
type ExecutionResult struct {
	Output          json.RawMessage `json:"output"`
	Error           string          `json:"error,omitempty"`
	RuntimeMS       int             `json:"runtime_ms"`
	MemoryMB        int             `json:"memory_mb"`
	ComputeCost     float64         `json:"compute_cost"`
	IsDeterministic bool            `json:"is_deterministic"`
}

// NewService creates a new Flywheel service
func NewService(repo *Repository, execService ExecutionService, logger *logrus.Logger) *Service {
	return &Service{
		repo:        repo,
		execService: execService,
		logger:      logger,
	}
}

// ThreadService methods

// CreateThread creates a new thread with validation
func (s *Service) CreateThread(ctx context.Context, thread *Thread) error {
	// Validate thread type
	if thread.Type == "" {
		return fmt.Errorf("thread type is required")
	}
	if thread.Type != ThreadTypeProblem && thread.Type != ThreadTypeDiscussion && thread.Type != ThreadTypeChallenge {
		return fmt.Errorf("invalid thread type: %s", thread.Type)
	}

	// Validate problem data for problem threads
	if thread.Type == ThreadTypeProblem {
		if thread.ProblemData == nil || len(thread.ProblemData) == 0 {
			return fmt.Errorf("problem_data is required for problem threads")
		}
	}

	// Set default status
	if thread.Status == "" {
		thread.Status = ThreadStatusOpen
	}

	thread.CreatedAt = time.Now()
	thread.UpdatedAt = time.Now()

	return s.repo.CreateThread(ctx, thread)
}

// GetThread retrieves a thread and increments view count
func (s *Service) GetThread(ctx context.Context, id uuid.UUID) (*Thread, error) {
	thread, err := s.repo.GetThreadByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Increment view count asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.repo.IncrementThreadViewCount(ctx, id); err != nil {
			s.logger.WithError(err).Warn("Failed to increment thread view count")
		}
	}()

	return thread, nil
}

// ListThreads lists threads with filtering
func (s *Service) ListThreads(ctx context.Context, filter ThreadFilter) ([]Thread, int64, error) {
	return s.repo.ListThreads(ctx, filter)
}

// IsModeratorRole returns true for platform roles that can moderate threads (update/delete others' content).
func IsModeratorRole(role string) bool {
	switch role {
	case "super_admin", "admin", "support":
		return true
	default:
		return false
	}
}

// UpdateThread updates a thread with authorization check
func (s *Service) UpdateThread(ctx context.Context, threadID uuid.UUID, updates map[string]interface{}, userID uuid.UUID, callerCanModerate bool) error {
	thread, err := s.repo.GetThreadByID(ctx, threadID)
	if err != nil {
		return err
	}

	// Check authorization: author or moderator
	if thread.AuthorID != userID && !callerCanModerate {
		return fmt.Errorf("unauthorized: only the author or a moderator can update this thread")
	}

	// Apply updates
	if title, ok := updates["title"].(string); ok {
		thread.Title = title
	}
	if content, ok := updates["content"].(string); ok {
		// For problem threads, update problem_data
		if thread.Type == ThreadTypeProblem {
			thread.ProblemData = json.RawMessage(content)
		}
	}
	if status, ok := updates["status"].(string); ok {
		thread.Status = ThreadStatus(status)
	}
	if categoryID, ok := updates["category_id"].(string); ok {
		id, err := uuid.Parse(categoryID)
		if err != nil {
			return fmt.Errorf("invalid category_id: %w", err)
		}
		thread.CategoryID = &id
	}
	if tags, ok := updates["tags"].([]string); ok {
		thread.Tags = tags
	}

	thread.UpdatedAt = time.Now()
	return s.repo.UpdateThread(ctx, thread)
}

// ResolveThread marks a thread as resolved with an accepted solution
func (s *Service) ResolveThread(ctx context.Context, threadID, replyID, userID uuid.UUID) error {
	thread, err := s.repo.GetThreadByID(ctx, threadID)
	if err != nil {
		return err
	}

	// Verify user is the author
	if thread.AuthorID != userID {
		return fmt.Errorf("unauthorized: only the author can resolve this thread")
	}

	// Verify reply exists and belongs to this thread
	reply, err := s.repo.GetReplyByID(ctx, replyID)
	if err != nil {
		return fmt.Errorf("reply not found: %w", err)
	}
	if reply.ThreadID != threadID {
		return fmt.Errorf("reply does not belong to this thread")
	}

	// Mark thread as resolved
	if err := s.repo.MarkThreadAsResolved(ctx, threadID, replyID); err != nil {
		return err
	}

	// Mark reply as accepted
	if err := s.repo.MarkReplyAsAccepted(ctx, replyID); err != nil {
		return err
	}

	// Award reputation to the solution author
	reputationEvent := &ReputationEvent{
		UserID:       reply.AuthorID,
		EventType:    "solution_accepted",
		ScoreType:    ReputationScoreTypeBuilder,
		PointsChange: 100,
		Reason:       fmt.Sprintf("Solution accepted for thread: %s", thread.Title),
		ReferenceID:  &replyID,
	}
	if err := s.repo.CreateReputationEvent(ctx, reputationEvent); err != nil {
		s.logger.WithError(err).Warn("Failed to award reputation for accepted solution")
	}

	return nil
}

// ReplyService methods

// CreateReply creates a new reply
func (s *Service) CreateReply(ctx context.Context, reply *Reply) error {
	// Verify thread exists and is open
	thread, err := s.repo.GetThreadByID(ctx, reply.ThreadID)
	if err != nil {
		return fmt.Errorf("thread not found: %w", err)
	}

	if thread.Status == ThreadStatusClosed {
		return fmt.Errorf("cannot reply to a closed thread")
	}

	reply.CreatedAt = time.Now()
	reply.UpdatedAt = time.Now()
	reply.HelpfulCount = 0
	reply.IsAccepted = false

	if reply.AuthorType == "" {
		reply.AuthorType = ReplyAuthorTypeUser
	}

	if err := s.repo.CreateReply(ctx, reply); err != nil {
		return err
	}

	// Update thread engagement score
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Simple engagement score calculation
		engagementUpdate := map[string]interface{}{
			"engagement_score": gorm.Expr("engagement_score + 1"),
		}
		if err := s.repo.UpdateThread(ctx, &Thread{ID: reply.ThreadID}); err != nil {
			s.logger.WithError(err).Warn("Failed to update thread engagement")
		}
		_ = engagementUpdate
	}()

	// Award reputation for contributing
	if reply.AuthorType == ReplyAuthorTypeUser {
		reputationEvent := &ReputationEvent{
			UserID:       reply.AuthorID,
			EventType:    "reply_created",
			ScoreType:    ReputationScoreTypeMentor,
			PointsChange: 5,
			Reason:       "Contributed to community discussion",
			ReferenceID:  &reply.ID,
		}
		if err := s.repo.CreateReputationEvent(ctx, reputationEvent); err != nil {
			s.logger.WithError(err).Warn("Failed to award reputation for reply")
		}
	}

	return nil
}

// GetReply retrieves a reply by ID
func (s *Service) GetReply(ctx context.Context, id uuid.UUID) (*Reply, error) {
	return s.repo.GetReplyByID(ctx, id)
}

// ListReplies lists replies for a thread
func (s *Service) ListReplies(ctx context.Context, threadID uuid.UUID, filter ReplyFilter) ([]Reply, int64, error) {
	return s.repo.ListRepliesByThread(ctx, threadID, filter)
}

// ExecuteReply executes the code in a reply
func (s *Service) ExecuteReply(ctx context.Context, replyID, executorID uuid.UUID, input json.RawMessage) (*Execution, error) {
	reply, err := s.repo.GetReplyByID(ctx, replyID)
	if err != nil {
		return nil, fmt.Errorf("reply not found: %w", err)
	}

	// Extract code from code_blocks
	var codeBlocks []struct {
		Language string `json:"language"`
		Code     string `json:"code"`
	}
	if err := json.Unmarshal(reply.CodeBlocks, &codeBlocks); err != nil {
		return nil, fmt.Errorf("failed to parse code blocks: %w", err)
	}

	if len(codeBlocks) == 0 {
		return nil, fmt.Errorf("no code blocks found in reply")
	}

	// Use the first code block for execution
	code := codeBlocks[0].Code
	language := codeBlocks[0].Language

	// Execute the code
	execResult, err := s.execService.ExecuteCode(ctx, code, language, input)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	// Create execution record
	execution := &Execution{
		ReplyID:            replyID,
		ExecutorID:         executorID,
		ExecutionContext:   json.RawMessage(`{}`),
		Input:              input,
		Output:             execResult.Output,
		VerificationStatus: VerificationStatusPending,
		ComputeCost:        &execResult.ComputeCost,
		RuntimeMS:          &execResult.RuntimeMS,
		MemoryMB:           &execResult.MemoryMB,
		IsDeterministic:    &execResult.IsDeterministic,
		CreatedAt:          time.Now(),
	}

	if execResult.Error != "" {
		execution.Error = &execResult.Error
		execution.VerificationStatus = VerificationStatusFailed
	}

	if err := s.repo.CreateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to save execution: %w", err)
	}

	// Update reliability stats
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		scores, err := s.repo.GetReputationScores(ctx, reply.AuthorID)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to get reputation scores")
			return
		}

		scores.TotalExecutions++
		if execResult.Error == "" {
			scores.SuccessfulExecutions++
		}

		// Recalculate reliability index
		if scores.TotalExecutions > 0 {
			scores.ReliabilityIndex = float64(scores.SuccessfulExecutions) / float64(scores.TotalExecutions) * 100
		}

		if err := s.repo.UpdateReputationScores(ctx, scores); err != nil {
			s.logger.WithError(err).Warn("Failed to update reliability index")
		}
	}()

	return execution, nil
}

// VerifyReply verifies a reply against expected output
func (s *Service) VerifyReply(ctx context.Context, replyID uuid.UUID) error {
	reply, err := s.repo.GetReplyByID(ctx, replyID)
	if err != nil {
		return fmt.Errorf("reply not found: %w", err)
	}

	thread, err := s.repo.GetThreadByID(ctx, reply.ThreadID)
	if err != nil {
		return fmt.Errorf("thread not found: %w", err)
	}

	// Get the latest execution
	executions, err := s.repo.ListExecutionsByReply(ctx, replyID, 1)
	if err != nil {
		return fmt.Errorf("failed to get executions: %w", err)
	}

	if len(executions) == 0 {
		return fmt.Errorf("no executions found for this reply")
	}

	execution := executions[0]

	// Verify output against expected
	if thread.ExpectedOutput != nil && len(thread.ExpectedOutput) > 0 {
		verified, reason, err := s.execService.VerifyOutput(ctx, execution.Output, thread.ExpectedOutput)
		if err != nil {
			return fmt.Errorf("verification failed: %w", err)
		}

		if verified {
			execution.VerificationStatus = VerificationStatusVerified
		} else {
			execution.VerificationStatus = VerificationStatusFailed
			if reason != "" {
				execution.Error = &reason
			}
		}

		if err := s.repo.UpdateExecutionVerificationStatus(ctx, execution.ID, execution.VerificationStatus); err != nil {
			return fmt.Errorf("failed to update verification status: %w", err)
		}

		// Award reputation for verified solution
		if execution.VerificationStatus == VerificationStatusVerified {
			reputationEvent := &ReputationEvent{
				UserID:       reply.AuthorID,
				EventType:    "solution_verified",
				ScoreType:    ReputationScoreTypeBuilder,
				PointsChange: 50,
				Reason:       "Solution verified with expected output",
				ReferenceID:  &replyID,
			}
			if err := s.repo.CreateReputationEvent(ctx, reputationEvent); err != nil {
				s.logger.WithError(err).Warn("Failed to award reputation for verified solution")
			}
		}
	}

	return nil
}

// ReputationService methods

// GetUserReputation retrieves a user's reputation scores
func (s *Service) GetUserReputation(ctx context.Context, userID uuid.UUID) (*ReputationScores, error) {
	return s.repo.GetReputationScores(ctx, userID)
}

// CountSharedThreads returns the number of flywheel threads where both users participated.
func (s *Service) CountSharedThreads(ctx context.Context, userID1, userID2 uuid.UUID) (int64, error) {
	return s.repo.CountSharedThreads(ctx, userID1, userID2)
}

// ListReputationEvents lists reputation events for a user
func (s *Service) ListReputationEvents(ctx context.Context, userID uuid.UUID, limit, offset int) ([]ReputationEvent, int64, error) {
	return s.repo.ListReputationEvents(ctx, userID, limit, offset)
}

// GetLeaderboard retrieves the leaderboard
func (s *Service) GetLeaderboard(ctx context.Context, scoreType ReputationScoreType, limit, offset int) ([]ReputationScores, int64, error) {
	return s.repo.GetLeaderboard(ctx, scoreType, limit, offset)
}

// ChallengeService methods

// CreateChallenge creates a new challenge (admin only)
func (s *Service) CreateChallenge(ctx context.Context, challenge *Challenge) error {
	if challenge.Status == "" {
		challenge.Status = ChallengeStatusDraft
	}

	challenge.CreatedAt = time.Now()
	challenge.UpdatedAt = time.Now()

	return s.repo.CreateChallenge(ctx, challenge)
}

// GetChallenge retrieves a challenge
func (s *Service) GetChallenge(ctx context.Context, id uuid.UUID) (*Challenge, error) {
	return s.repo.GetChallengeByID(ctx, id)
}

// ListChallenges lists challenges
func (s *Service) ListChallenges(ctx context.Context, filter ChallengeFilter) ([]Challenge, int64, error) {
	return s.repo.ListChallenges(ctx, filter)
}

// SubmitChallengeEntry submits an entry to a challenge
func (s *Service) SubmitChallengeEntry(ctx context.Context, submission *ChallengeSubmission) error {
	// Verify challenge exists and is active
	challenge, err := s.repo.GetChallengeByID(ctx, submission.ChallengeID)
	if err != nil {
		return fmt.Errorf("challenge not found: %w", err)
	}

	if challenge.Status != ChallengeStatusActive {
		return fmt.Errorf("challenge is not active")
	}

	now := time.Now()
	if now.Before(challenge.StartTime) || now.After(challenge.EndTime) {
		return fmt.Errorf("challenge is not accepting submissions at this time")
	}

	// Check if user already has a submission
	existing, _ := s.repo.GetUserChallengeSubmission(ctx, submission.ChallengeID, submission.ParticipantID)
	if existing != nil {
		// Update existing submission
		submission.ID = existing.ID
		submission.CreatedAt = existing.CreatedAt
		submission.UpdatedAt = time.Now()
		// Update logic would go here
		return fmt.Errorf("updating submissions not yet implemented")
	}

	submission.CreatedAt = time.Now()
	submission.UpdatedAt = time.Now()

	return s.repo.CreateChallengeSubmission(ctx, submission)
}

// GetChallengeLeaderboard retrieves the leaderboard for a challenge
func (s *Service) GetChallengeLeaderboard(ctx context.Context, challengeID uuid.UUID, limit int) ([]ChallengeSubmission, error) {
	return s.repo.GetChallengeLeaderboard(ctx, challengeID, limit)
}

// CategoryService methods

// ListCategories lists all categories
func (s *Service) ListCategories(ctx context.Context) ([]Category, error) {
	return s.repo.ListCategories(ctx)
}

// SubscriptionService methods

// SubscribeToThread subscribes a user to a thread
func (s *Service) SubscribeToThread(ctx context.Context, userID, threadID uuid.UUID, notificationLevel string) error {
	if notificationLevel == "" {
		notificationLevel = "all"
	}

	subscription := &Subscription{
		UserID:            userID,
		ThreadID:          threadID,
		NotificationLevel: notificationLevel,
		CreatedAt:         time.Now(),
	}

	return s.repo.CreateSubscription(ctx, subscription)
}

// UnsubscribeFromThread unsubscribes a user from a thread
func (s *Service) UnsubscribeFromThread(ctx context.Context, userID, threadID uuid.UUID) error {
	return s.repo.DeleteSubscription(ctx, userID, threadID)
}

// ListUserSubscriptions lists a user's subscriptions
func (s *Service) ListUserSubscriptions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Subscription, int64, error) {
	return s.repo.ListUserSubscriptions(ctx, userID, limit, offset)
}

// Helper to suppress unused import warning for services package
type _ = services.StorageService
