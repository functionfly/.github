// Package membership provides services for handling membership-related events
// like plan upgrades, achievement awards, and activity feed creation.
package membership

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Service handles membership-related operations like plan upgrades,
// awarding achievements, and creating activity feed items.
type Service struct {
	repo storage.Repository
}

// NewService creates a new membership service.
func NewService(repo storage.Repository) *Service {
	return &Service{repo: repo}
}

// PlanUpgradeData contains information about a plan upgrade.
type PlanUpgradeData struct {
	UserID       uuid.UUID
	TenantID     uuid.UUID
	OldPlan      string
	NewPlan      string
	UpgradedAt   time.Time
	UpgradedBy   uuid.UUID // User who performed the upgrade (may be same as UserID for self-upgrades)
}

// HandlePlanUpgrade processes a plan upgrade, awarding achievements and creating activity feed items.
// This should be called whenever a user's plan is upgraded.
func (s *Service) HandlePlanUpgrade(ctx context.Context, data PlanUpgradeData) error {
	// Create activity feed item for the upgrade
	if err := s.createMembershipUpgradeActivity(ctx, data); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id":   data.UserID,
			"tenant_id": data.TenantID,
			"old_plan":  data.OldPlan,
			"new_plan":  data.NewPlan,
		}).Warn("Failed to create membership upgrade activity")
		// Don't return error - activity creation is best-effort
	}

	// Award enterprise achievement if upgrading to enterprise
	if isEnterprisePlan(data.NewPlan) && !isEnterprisePlan(data.OldPlan) {
		if err := s.awardEnterpriseAchievement(ctx, data.UserID); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"user_id":   data.UserID,
				"tenant_id": data.TenantID,
			}).Warn("Failed to award enterprise achievement")
			// Don't return error - achievement awarding is best-effort
		}
	}

	logrus.WithFields(logrus.Fields{
		"user_id":   data.UserID,
		"tenant_id": data.TenantID,
		"old_plan":  data.OldPlan,
		"new_plan":  data.NewPlan,
	}).Info("Processed plan upgrade")

	return nil
}

// createMembershipUpgradeActivity creates an activity feed item for a plan upgrade.
func (s *Service) createMembershipUpgradeActivity(ctx context.Context, data PlanUpgradeData) error {
	// Format plan names for display
	newPlanDisplay := formatPlanName(data.NewPlan)

	// Create activity title and description
	title := fmt.Sprintf("Upgraded to %s", newPlanDisplay)
	description := s.getUpgradeDescription(data.NewPlan)

	// Create metadata
	metadata := map[string]interface{}{
		"plan":         data.NewPlan,
		"previousPlan": data.OldPlan,
		"upgradedAt":   data.UpgradedAt.Format(time.RFC3339),
	}

	// Create the activity
	activity := &storage.UserActivity{
		UserID:       data.UserID,
		ActivityType: "membership_upgraded",
		Title:        title,
		Description:  description,
		Metadata:     metadata,
		IsPublic:     true,
	}

	if err := s.repo.CreateUserActivity(ctx, activity); err != nil {
		return fmt.Errorf("failed to create user activity: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    data.UserID,
		"activity_id": activity.ID,
		"new_plan":   data.NewPlan,
	}).Debug("Created membership upgrade activity")

	return nil
}

// awardEnterpriseAchievement awards the "Enterprise Pioneer" achievement to a user.
func (s *Service) awardEnterpriseAchievement(ctx context.Context, userID uuid.UUID) error {
	// Get the achievement by slug
	achievement, err := s.repo.GetAchievementBySlug(ctx, "enterprise_pioneer")
	if err != nil {
		return fmt.Errorf("failed to get enterprise pioneer achievement: %w", err)
	}

	if achievement == nil {
		// Achievement doesn't exist yet - log warning but don't fail
		logrus.Warn("Enterprise Pioneer achievement not found in database - skipping award")
		return nil
	}

	// Check if user already has this achievement
	existingAchievements, err := s.repo.GetUserAchievements(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check existing achievements: %w", err)
	}

	for _, ea := range existingAchievements {
		if ea.AchievementID == achievement.ID {
			// User already has this achievement
			logrus.WithFields(logrus.Fields{
				"user_id":        userID,
				"achievement_id": achievement.ID,
			}).Debug("User already has Enterprise Pioneer achievement")
			return nil
		}
	}

	// Award the achievement
	metadata := map[string]interface{}{
		"awarded_at": time.Now().Format(time.RFC3339),
		"reason":     "Upgraded to Enterprise plan",
	}

	if err := s.repo.AwardAchievement(userID, achievement.ID, metadata); err != nil {
		return fmt.Errorf("failed to award achievement: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":        userID,
		"achievement_id": achievement.ID,
	}).Info("Awarded Enterprise Pioneer achievement")

	return nil
}

// getUpgradeDescription returns a description based on the plan tier.
func (s *Service) getUpgradeDescription(plan string) string {
	switch plan {
	case "enterprise":
		return "Unlimited functions, dedicated support, and premium enterprise features"
	case "professional", "pro":
		return "Advanced features, priority support, and increased limits"
	case "starter":
		return "Expanded features and higher execution limits"
	default:
		return "Membership upgraded with new features and benefits"
	}
}

// isEnterprisePlan checks if a plan name represents an enterprise tier.
func isEnterprisePlan(plan string) bool {
	return plan == "enterprise" || plan == "Enterprise"
}

// formatPlanName formats a plan name for display.
func formatPlanName(plan string) string {
	switch plan {
	case "enterprise", "Enterprise":
		return "Enterprise"
	case "professional", "pro", "Pro":
		return "Professional"
	case "starter", "Starter":
		return "Starter"
	case "free", "Free":
		return "Free"
	default:
		// Capitalize first letter
		if len(plan) > 0 {
			return string(plan[0]-32) + plan[1:] // -32 converts lowercase to uppercase
		}
		return plan
	}
}
