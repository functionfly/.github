package flywheel

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository provides database operations for Flywheel Network
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new Flywheel repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Thread Operations

// CreateThread creates a new thread
func (r *Repository) CreateThread(ctx context.Context, thread *Thread) error {
	return r.db.WithContext(ctx).Create(thread).Error
}

// GetThreadByID retrieves a thread by ID with related data
func (r *Repository) GetThreadByID(ctx context.Context, id uuid.UUID) (*Thread, error) {
	var thread Thread
	err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Category").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Where("parent_reply_id IS NULL").Order("created_at ASC")
		}).
		Preload("Replies.Author").
		First(&thread, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

// ListThreads retrieves threads with filtering and pagination
func (r *Repository) ListThreads(ctx context.Context, filter ThreadFilter) ([]Thread, int64, error) {
	var threads []Thread
	var count int64

	query := r.db.WithContext(ctx).Model(&Thread{})

	// Apply filters
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
	}
	if filter.AuthorID != nil {
		query = query.Where("author_id = ?", *filter.AuthorID)
	}
	if len(filter.Tags) > 0 {
		query = query.Where("tags && ?", filter.Tags)
	}
	if filter.SearchQuery != "" {
		query = query.Where(
			"title ILIKE ? OR problem_data::text ILIKE ?",
			"%"+filter.SearchQuery+"%",
			"%"+filter.SearchQuery+"%",
		)
	}

	// Get count
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	switch filter.SortBy {
	case "newest":
		query = query.Order("created_at DESC")
	case "oldest":
		query = query.Order("created_at ASC")
	case "trending":
		query = query.Order("engagement_score DESC, created_at DESC")
	case "most_viewed":
		query = query.Order("view_count DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Execute query with preloads
	err := query.
		Preload("Author").
		Preload("Category").
		Find(&threads).Error

	return threads, count, err
}

// UpdateThread updates a thread
func (r *Repository) UpdateThread(ctx context.Context, thread *Thread) error {
	return r.db.WithContext(ctx).Save(thread).Error
}

// DeleteThread soft deletes a thread
func (r *Repository) DeleteThread(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Thread{}, "id = ?", id).Error
}

// CountSharedThreads returns the number of distinct flywheel threads where both userID1 and userID2
// participated (as thread author and/or reply author).
func (r *Repository) CountSharedThreads(ctx context.Context, userID1, userID2 uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Thread{}).
		Where(`
			(author_id = ? OR EXISTS (SELECT 1 FROM flywheel_replies r WHERE r.thread_id = flywheel_threads.id AND r.author_id = ?))
			AND (author_id = ? OR EXISTS (SELECT 1 FROM flywheel_replies r WHERE r.thread_id = flywheel_threads.id AND r.author_id = ?))`,
			userID1, userID1, userID2, userID2).
		Count(&count).Error
	return count, err
}

// MarkThreadAsResolved marks a thread as resolved with an accepted solution
func (r *Repository) MarkThreadAsResolved(ctx context.Context, threadID, replyID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&Thread{}).
		Where("id = ?", threadID).
		Updates(map[string]interface{}{
			"status":            ThreadStatusResolved,
			"accepted_reply_id": replyID,
			"resolved_at":       now,
		}).Error
}

// IncrementThreadViewCount increments the view count for a thread
func (r *Repository) IncrementThreadViewCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Thread{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// Reply Operations

// CreateReply creates a new reply
func (r *Repository) CreateReply(ctx context.Context, reply *Reply) error {
	return r.db.WithContext(ctx).Create(reply).Error
}

// GetReplyByID retrieves a reply by ID
func (r *Repository) GetReplyByID(ctx context.Context, id uuid.UUID) (*Reply, error) {
	var reply Reply
	err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Thread").
		Preload("ChildReplies", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("ChildReplies.Author").
		First(&reply, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &reply, nil
}

// ListRepliesByThread retrieves replies for a thread
func (r *Repository) ListRepliesByThread(ctx context.Context, threadID uuid.UUID, filter ReplyFilter) ([]Reply, int64, error) {
	var replies []Reply
	var count int64

	query := r.db.WithContext(ctx).Model(&Reply{}).Where("thread_id = ?", threadID)

	// Filter by parent (for nested replies)
	if filter.ParentOnly {
		query = query.Where("parent_reply_id IS NULL")
	}
	if filter.ParentReplyID != nil {
		query = query.Where("parent_reply_id = ?", *filter.ParentReplyID)
	}

	// Get count
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	query = query.Order("created_at ASC")

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.
		Preload("Author").
		Preload("ChildReplies", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("ChildReplies.Author").
		Find(&replies).Error

	return replies, count, err
}

// UpdateReply updates a reply
func (r *Repository) UpdateReply(ctx context.Context, reply *Reply) error {
	return r.db.WithContext(ctx).Save(reply).Error
}

// MarkReplyAsAccepted marks a reply as the accepted solution
func (r *Repository) MarkReplyAsAccepted(ctx context.Context, replyID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Reply{}).
		Where("id = ?", replyID).
		Update("is_accepted", true).Error
}

// IncrementHelpfulCount increments the helpful count for a reply
func (r *Repository) IncrementHelpfulCount(ctx context.Context, replyID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Reply{}).
		Where("id = ?", replyID).
		UpdateColumn("helpful_count", gorm.Expr("helpful_count + 1")).Error
}

// Reputation Operations

// GetReputationScores retrieves reputation scores for a user
func (r *Repository) GetReputationScores(ctx context.Context, userID uuid.UUID) (*ReputationScores, error) {
	var scores ReputationScores
	err := r.db.WithContext(ctx).First(&scores, "user_id = ?", userID).Error
	if err == gorm.ErrRecordNotFound {
		// Create default scores for new user
		scores = ReputationScores{
			UserID:             userID,
			BuilderTier:        TierBronze,
			OptimizerTier:      TierBronze,
			MentorTier:         TierBronze,
			AgentWhispererTier: TierBronze,
			ReliabilityIndex:   100,
		}
		if err := r.db.WithContext(ctx).Create(&scores).Error; err != nil {
			return nil, err
		}
		return &scores, nil
	}
	if err != nil {
		return nil, err
	}
	return &scores, nil
}

// UpdateReputationScores updates reputation scores
func (r *Repository) UpdateReputationScores(ctx context.Context, scores *ReputationScores) error {
	scores.LastCalculatedAt = time.Now()
	return r.db.WithContext(ctx).Save(scores).Error
}

// ListReputationEvents retrieves reputation events for a user
func (r *Repository) ListReputationEvents(ctx context.Context, userID uuid.UUID, limit, offset int) ([]ReputationEvent, int64, error) {
	var events []ReputationEvent
	var count int64

	query := r.db.WithContext(ctx).Model(&ReputationEvent{}).Where("user_id = ?", userID)

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error

	return events, count, err
}

// CreateReputationEvent creates a reputation event and updates scores atomically
func (r *Repository) CreateReputationEvent(ctx context.Context, event *ReputationEvent) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the event
		if err := tx.Create(event).Error; err != nil {
			return err
		}

		// Update the appropriate score
		scores, err := r.GetReputationScores(ctx, event.UserID)
		if err != nil {
			return err
		}

		switch event.ScoreType {
		case ReputationScoreTypeBuilder:
			scores.BuilderScore += event.PointsChange
			scores.BuilderTier = CalculateTier(scores.BuilderScore)
		case ReputationScoreTypeOptimizer:
			scores.OptimizerScore += event.PointsChange
			scores.OptimizerTier = CalculateTier(scores.OptimizerScore)
		case ReputationScoreTypeMentor:
			scores.MentorScore += event.PointsChange
			scores.MentorTier = CalculateTier(scores.MentorScore)
		case ReputationScoreTypeAgentWhisperer:
			scores.AgentWhispererScore += event.PointsChange
			scores.AgentWhispererTier = CalculateTier(scores.AgentWhispererScore)
		}

		scores.LastCalculatedAt = time.Now()
		return tx.Save(scores).Error
	})
}

// GetLeaderboard retrieves the leaderboard for a specific score type
func (r *Repository) GetLeaderboard(ctx context.Context, scoreType ReputationScoreType, limit, offset int) ([]ReputationScores, int64, error) {
	var scores []ReputationScores
	var count int64

	query := r.db.WithContext(ctx).Model(&ReputationScores{})

	// Apply sorting based on score type
	switch scoreType {
	case ReputationScoreTypeBuilder:
		query = query.Order("builder_score DESC")
	case ReputationScoreTypeOptimizer:
		query = query.Order("optimizer_score DESC")
	case ReputationScoreTypeMentor:
		query = query.Order("mentor_score DESC")
	case ReputationScoreTypeAgentWhisperer:
		query = query.Order("agent_whisperer_score DESC")
	case ReputationScoreTypeReliability:
		query = query.Order("reliability_index DESC")
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Limit(limit).
		Offset(offset).
		Preload("User").
		Find(&scores).Error

	return scores, count, err
}

// Challenge Operations

// CreateChallenge creates a new challenge
func (r *Repository) CreateChallenge(ctx context.Context, challenge *Challenge) error {
	return r.db.WithContext(ctx).Create(challenge).Error
}

// GetChallengeByID retrieves a challenge by ID
func (r *Repository) GetChallengeByID(ctx context.Context, id uuid.UUID) (*Challenge, error) {
	var challenge Challenge
	err := r.db.WithContext(ctx).First(&challenge, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}

// ListChallenges retrieves challenges with filtering
func (r *Repository) ListChallenges(ctx context.Context, filter ChallengeFilter) ([]Challenge, int64, error) {
	var challenges []Challenge
	var count int64

	query := r.db.WithContext(ctx).Model(&Challenge{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Type != "" {
		query = query.Where("challenge_type = ?", filter.Type)
	}
	if filter.ActiveOnly {
		now := time.Now()
		query = query.Where("start_time <= ? AND end_time >= ?", now, now)
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("created_at DESC")

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Find(&challenges).Error
	return challenges, count, err
}

// UpdateChallenge updates a challenge
func (r *Repository) UpdateChallenge(ctx context.Context, challenge *Challenge) error {
	return r.db.WithContext(ctx).Save(challenge).Error
}

// CreateChallengeSubmission creates a new challenge submission
func (r *Repository) CreateChallengeSubmission(ctx context.Context, submission *ChallengeSubmission) error {
	return r.db.WithContext(ctx).Create(submission).Error
}

// GetChallengeLeaderboard retrieves the leaderboard for a challenge
func (r *Repository) GetChallengeLeaderboard(ctx context.Context, challengeID uuid.UUID, limit int) ([]ChallengeSubmission, error) {
	var submissions []ChallengeSubmission
	err := r.db.WithContext(ctx).
		Where("challenge_id = ?", challengeID).
		Where("rank IS NOT NULL").
		Order("rank ASC").
		Limit(limit).
		Preload("Participant").
		Find(&submissions).Error
	return submissions, err
}

// GetUserChallengeSubmission retrieves a user's submission for a challenge
func (r *Repository) GetUserChallengeSubmission(ctx context.Context, challengeID, userID uuid.UUID) (*ChallengeSubmission, error) {
	var submission ChallengeSubmission
	err := r.db.WithContext(ctx).
		Where("challenge_id = ? AND participant_id = ?", challengeID, userID).
		First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

// Execution Operations

// CreateExecution creates a new execution record
func (r *Repository) CreateExecution(ctx context.Context, execution *Execution) error {
	return r.db.WithContext(ctx).Create(execution).Error
}

// GetExecutionByID retrieves an execution by ID
func (r *Repository) GetExecutionByID(ctx context.Context, id uuid.UUID) (*Execution, error) {
	var execution Execution
	err := r.db.WithContext(ctx).
		Preload("Reply").
		Preload("Executor").
		First(&execution, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

// ListExecutionsByReply retrieves executions for a reply
func (r *Repository) ListExecutionsByReply(ctx context.Context, replyID uuid.UUID, limit int) ([]Execution, error) {
	var executions []Execution
	err := r.db.WithContext(ctx).
		Where("reply_id = ?", replyID).
		Order("created_at DESC").
		Limit(limit).
		Find(&executions).Error
	return executions, err
}

// UpdateExecutionVerificationStatus updates the verification status of an execution
func (r *Repository) UpdateExecutionVerificationStatus(ctx context.Context, executionID uuid.UUID, status VerificationStatus) error {
	return r.db.WithContext(ctx).Model(&Execution{}).
		Where("id = ?", executionID).
		Update("verification_status", status).Error
}

// Category Operations

// ListCategories retrieves all categories
func (r *Repository) ListCategories(ctx context.Context) ([]Category, error) {
	var categories []Category
	err := r.db.WithContext(ctx).
		Where("parent_id IS NULL").
		Order("sort_order ASC").
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Find(&categories).Error
	return categories, err
}

// GetCategoryByID retrieves a category by ID
func (r *Repository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error) {
	var category Category
	err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// Subscription Operations

// CreateSubscription creates a new subscription
func (r *Repository) CreateSubscription(ctx context.Context, subscription *Subscription) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

// DeleteSubscription deletes a subscription
func (r *Repository) DeleteSubscription(ctx context.Context, userID, threadID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Subscription{}, "user_id = ? AND thread_id = ?", userID, threadID).Error
}

// ListUserSubscriptions retrieves a user's subscriptions
func (r *Repository) ListUserSubscriptions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Subscription, int64, error) {
	var subscriptions []Subscription
	var count int64

	query := r.db.WithContext(ctx).Model(&Subscription{}).Where("user_id = ?", userID)

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Thread").
		Find(&subscriptions).Error

	return subscriptions, count, err
}

// Filter structs for query parameters

// ThreadFilter represents filters for listing threads
type ThreadFilter struct {
	Type        ThreadType
	Status      ThreadStatus
	CategoryID  *uuid.UUID
	AuthorID    *uuid.UUID
	Tags        []string
	SearchQuery string
	SortBy      string
	Limit       int
	Offset      int
}

// ReplyFilter represents filters for listing replies
type ReplyFilter struct {
	ParentOnly    bool
	ParentReplyID *uuid.UUID
	Limit         int
	Offset        int
}

// ChallengeFilter represents filters for listing challenges
type ChallengeFilter struct {
	Status     ChallengeStatus
	Type       ChallengeType
	ActiveOnly bool
	Limit      int
	Offset     int
}

// CalculateTier calculates the tier based on score
func CalculateTier(score int) Tier {
	switch {
	case score >= 8000:
		return TierLegend
	case score >= 6000:
		return TierDiamond
	case score >= 4000:
		return TierPlatinum
	case score >= 2000:
		return TierGold
	case score >= 500:
		return TierSilver
	default:
		return TierBronze
	}
}

// GetTierName returns the name of a tier
func GetTierName(tier Tier) string {
	switch tier {
	case TierBronze:
		return "Bronze"
	case TierSilver:
		return "Silver"
	case TierGold:
		return "Gold"
	case TierPlatinum:
		return "Platinum"
	case TierDiamond:
		return "Diamond"
	case TierLegend:
		return "Legend"
	default:
		return "Unknown"
	}
}

// GetTierColor returns the color for a tier
func GetTierColor(tier Tier) string {
	switch tier {
	case TierBronze:
		return "#CD7F32"
	case TierSilver:
		return "#C0C0C0"
	case TierGold:
		return "#FFD700"
	case TierPlatinum:
		return "#E5E4E2"
	case TierDiamond:
		return "#B9F2FF"
	case TierLegend:
		return "#FF6B35"
	default:
		return "#888888"
	}
}
