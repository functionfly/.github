package services

import (
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ReputationHooker provides helper functions to award reputation points
// from various platform actions. Includes anti-spam daily caps.
type ReputationHooker struct {
	db     *storage.PostgresDB
	logger *logrus.Logger
}

// NewReputationHooker creates a new reputation hooker
func NewReputationHooker(db *storage.PostgresDB, logger *logrus.Logger) *ReputationHooker {
	return &ReputationHooker{db: db, logger: logger}
}

// ReputationAction defines a reputation-awarding action
type ReputationAction string

const (
	ActionFunctionPublished  ReputationAction = "function_published"
	ActionFunctionUpdated    ReputationAction = "function_updated"
	ActionReviewSubmitted    ReputationAction = "review_submitted"
	ActionReviewReceived     ReputationAction = "review_received"
	ActionCommunityPost      ReputationAction = "community_post"
	ActionCommunityComment   ReputationAction = "community_comment"
	ActionAnswerAccepted     ReputationAction = "answer_accepted"
	ActionCommunityUpvote    ReputationAction = "community_upvote_received"
	ActionDailyLogin         ReputationAction = "daily_login"
	ActionMilestoneReached   ReputationAction = "milestone_reached"
)

// Point values per action
var actionPoints = map[ReputationAction]int{
	ActionFunctionPublished:  25,
	ActionFunctionUpdated:    5,
	ActionReviewSubmitted:    15,
	ActionReviewReceived:     10,
	ActionCommunityPost:      10,
	ActionCommunityComment:   5,
	ActionAnswerAccepted:     50,
	ActionCommunityUpvote:    2,
	ActionDailyLogin:         2,
	ActionMilestoneReached:   0, // variable
}

// Component mapping per action
var actionComponent = map[ReputationAction]string{
	ActionFunctionPublished:  "builder",
	ActionFunctionUpdated:    "builder",
	ActionReviewSubmitted:    "mentor",
	ActionReviewReceived:     "builder",
	ActionCommunityPost:      "mentor",
	ActionCommunityComment:   "mentor",
	ActionAnswerAccepted:     "mentor",
	ActionCommunityUpvote:    "mentor",
	ActionDailyLogin:         "agent_whisperer",
	ActionMilestoneReached:   "optimizer",
}

// Daily caps per action type (0 = no cap)
var dailyCaps = map[ReputationAction]int{
	ActionReviewSubmitted:  5,
	ActionCommunityPost:    3,
	ActionCommunityComment: 10,
	ActionCommunityUpvote:  20,
	ActionDailyLogin:       1,
	ActionFunctionUpdated:  1, // per function, handled separately
}

// Award awards reputation points for an action if daily cap not exceeded.
// Returns true if points were awarded, false if capped.
func (h *ReputationHooker) Award(userID, tenantID uuid.UUID, action ReputationAction, description string, referenceID uuid.UUID) bool {
	return h.AwardCustom(userID, tenantID, action, 0, description, referenceID)
}

// AwardCustom awards custom point value (for milestones etc).
// Pass points=0 to use default for the action.
func (h *ReputationHooker) AwardCustom(userID, tenantID uuid.UUID, action ReputationAction, points int, description string, referenceID uuid.UUID) bool {
	if userID == uuid.Nil {
		return false
	}

	// Check daily cap
	if cap, ok := dailyCaps[action]; ok && cap > 0 {
		count := h.countTodayEvents(userID, action)
		if count >= cap {
			return false
		}
	}

	// Use custom points or default
	if points == 0 {
		points = actionPoints[action]
	}
	if points == 0 {
		return false
	}

	component := actionComponent[action]

	// Record score change
	repo := storage.NewReputationRepository(h.db.GORM, h.logger)
	if err := repo.RecordScoreChange(userID, tenantID, component, points); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"action":  action,
		}).Warn("Failed to record reputation score change")
		return false
	}

	// Record event
	var refID *uuid.UUID
	if referenceID != uuid.Nil {
		refID = &referenceID
	}
	event := &storage.ReputationEvent{
		UserID:      userID,
		EventType:   string(action),
		ScoreChange: points,
		Component:   component,
		Description: description,
		ReferenceID: refID,
	}
	if err := repo.RecordReputationEvent(event); err != nil {
		h.logger.WithError(err).Warn("Failed to record reputation event")
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"action":   action,
		"points":   points,
		"component": component,
	}).Debug("Reputation points awarded")

	return true
}

// countTodayEvents counts how many events of this type the user has today
func (h *ReputationHooker) countTodayEvents(userID uuid.UUID, action ReputationAction) int {
	var count int64
	today := time.Now().UTC().Truncate(24 * time.Hour)
	h.db.GORM.Model(&storage.ReputationEvent{}).
		Where("user_id = ? AND event_type = ? AND created_at >= ?", userID, string(action), today).
		Count(&count)
	return int(count)
}
