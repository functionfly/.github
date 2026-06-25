package storage

import (
	"context"
	"fmt"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
)

// SeedAchievementDefinitions creates the default achievement definitions for a tenant.
func (r *FFIDRepository) SeedAchievementDefinitions(ctx context.Context, tenantID uuid.UUID) error {
	defs := []types.AchievementDefinition{
		{Slug: "builder-i", Name: "Builder I", Description: stringPtr("Ship your first project"), Icon: stringPtr("hammer"), Category: "engineering", CriteriaType: "projects_shipped", CriteriaThreshold: 1, Points: 100, Tier: 1},
		{Slug: "builder-ii", Name: "Builder II", Description: stringPtr("Ship 5 projects"), Icon: stringPtr("hammer"), Category: "engineering", CriteriaType: "projects_shipped", CriteriaThreshold: 5, Points: 250, Tier: 2},
		{Slug: "builder-iii", Name: "Builder III", Description: stringPtr("Ship 20 projects"), Icon: stringPtr("hammer"), Category: "engineering", CriteriaType: "projects_shipped", CriteriaThreshold: 20, Points: 500, Tier: 3},
		{Slug: "launch-architect", Name: "Launch Architect", Description: stringPtr("Lead a production launch"), Icon: stringPtr("rocket"), Category: "engineering", CriteriaType: "projects_shipped", CriteriaThreshold: 1, Points: 300, Tier: 2},
		{Slug: "guardian", Name: "Guardian", Description: stringPtr("Resolve a critical incident"), Icon: stringPtr("shield"), Category: "operations", CriteriaType: "incidents_resolved", CriteriaThreshold: 1, Points: 200, Tier: 1},
		{Slug: "mentor-i", Name: "Mentor I", Description: stringPtr("Complete your first mentorship"), Icon: stringPtr("users"), Category: "leadership", CriteriaType: "mentorships", CriteriaThreshold: 1, Points: 150, Tier: 1},
		{Slug: "mentor-ii", Name: "Mentor II", Description: stringPtr("Complete 5 mentorships"), Icon: stringPtr("users"), Category: "leadership", CriteriaType: "mentorships", CriteriaThreshold: 5, Points: 400, Tier: 2},
		{Slug: "innovator", Name: "Innovator", Description: stringPtr("Get a grant funded"), Icon: stringPtr("lightbulb"), Category: "innovation", CriteriaType: "grants_funded", CriteriaThreshold: 1, Points: 250, Tier: 2},
		{Slug: "security-champion", Name: "Security Champion", Description: stringPtr("100% compliance for 6 months"), Icon: stringPtr("lock"), Category: "security", CriteriaType: "compliance_streak", CriteriaThreshold: 1, Points: 300, Tier: 2},
		{Slug: "knowledge-keeper", Name: "Knowledge Keeper", Description: stringPtr("Publish 10 articles"), Icon: stringPtr("book"), Category: "knowledge", CriteriaType: "articles_published", CriteriaThreshold: 10, Points: 200, Tier: 1},
		{Slug: "collaborator", Name: "Collaborator", Description: stringPtr("Receive 50 peer feedbacks"), Icon: stringPtr("people-arrows"), Category: "culture", CriteriaType: "peer_feedbacks", CriteriaThreshold: 50, Points: 200, Tier: 1},
		{Slug: "reliable", Name: "Reliable", Description: stringPtr("Maintain 95%+ on-time delivery"), Icon: stringPtr("clock"), Category: "performance", CriteriaType: "on_time_pct", CriteriaThreshold: 1, Points: 250, Tier: 2},
	}

	for _, d := range defs {
		d.TenantID = tenantID
		d.IsActive = true
		_, err := r.CreateAchievementDefinition(ctx, &d)
		if err != nil {
			return fmt.Errorf("failed to seed achievement %s: %w", d.Slug, err)
		}
	}
	return nil
}
