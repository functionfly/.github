package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ============================================================================
// User Profile & Activity Operations
// ============================================================================

// GetUserSkills retrieves all skills for a user
func (db *PostgresDB) GetUserSkills(ctx context.Context, userID uuid.UUID) ([]*UserSkill, error) {
	var skills []*UserSkill
	if err := db.GORM.WithContext(ctx).Where("user_id = ?", userID).Find(&skills).Error; err != nil {
		return nil, fmt.Errorf("failed to get user skills: %w", err)
	}
	return skills, nil
}

// AddUserSkill adds a new skill for a user
func (db *PostgresDB) AddUserSkill(skill *UserSkill) error {
	if err := db.GORM.Create(skill).Error; err != nil {
		return fmt.Errorf("failed to add user skill: %w", err)
	}
	return nil
}

// RemoveUserSkill removes a skill by ID
func (db *PostgresDB) RemoveUserSkill(skillID uuid.UUID) error {
	if err := db.GORM.Delete(&UserSkill{}, skillID).Error; err != nil {
		return fmt.Errorf("failed to remove user skill: %w", err)
	}
	return nil
}

// GetUserAchievements retrieves all achievements for a user
func (db *PostgresDB) GetUserAchievements(ctx context.Context, userID uuid.UUID) ([]*UserAchievement, error) {
	var achievements []*UserAchievement
	if err := db.GORM.WithContext(ctx).Where("user_id = ?", userID).
		Preload("Achievement").
		Order("earned_at DESC").
		Find(&achievements).Error; err != nil {
		return nil, fmt.Errorf("failed to get user achievements: %w", err)
	}
	return achievements, nil
}

// GetAchievementBySlug retrieves an achievement by its slug
func (db *PostgresDB) GetAchievementBySlug(ctx context.Context, slug string) (*Achievement, error) {
	var achievement Achievement
	if err := db.GORM.WithContext(ctx).Where("slug = ?", slug).First(&achievement).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get achievement by slug: %w", err)
	}
	return &achievement, nil
}

// ListAchievements retrieves all achievement definitions
func (db *PostgresDB) ListAchievements(ctx context.Context) ([]*Achievement, error) {
	var achievements []*Achievement
	if err := db.GORM.WithContext(ctx).Order("category, name").Find(&achievements).Error; err != nil {
		return nil, fmt.Errorf("failed to list achievements: %w", err)
	}
	return achievements, nil
}

// AwardAchievement awards an achievement to a user
func (db *PostgresDB) AwardAchievement(userID, achievementID uuid.UUID, metadata map[string]interface{}) error {
	ua := &UserAchievement{
		UserID:        userID,
		AchievementID: achievementID,
		EarnedAt:      time.Now(),
		Progress:      100,
		IsCompleted:   true,
		Metadata:      metadata,
	}
	if err := db.GORM.Create(ua).Error; err != nil {
		return fmt.Errorf("failed to award achievement: %w", err)
	}
	return nil
}

// UpdateAchievementProgress updates the progress of a user achievement
func (db *PostgresDB) UpdateAchievementProgress(userAchievementID uuid.UUID, progress int, isCompleted bool) error {
	updates := map[string]interface{}{
		"progress": progress,
	}
	if isCompleted {
		updates["is_completed"] = true
		updates["earned_at"] = time.Now()
	}
	if err := db.GORM.Model(&UserAchievement{}).Where("id = ?", userAchievementID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update achievement progress: %w", err)
	}
	return nil
}

// GetUserActivity retrieves activity feed for a user
func (db *PostgresDB) GetUserActivity(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserActivity, error) {
	var activities []*UserActivity
	if err := db.GORM.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get user activity: %w", err)
	}
	return activities, nil
}

// CreateUserActivity creates a new activity feed item
func (db *PostgresDB) CreateUserActivity(ctx context.Context, activity *UserActivity) error {
	if err := db.GORM.WithContext(ctx).Create(activity).Error; err != nil {
		return fmt.Errorf("failed to create user activity: %w", err)
	}
	return nil
}

// GetUserContributionDailyCounts aggregates contribution events per UTC calendar day:
// each user_activity row and each registry function created by the user counts as one.
func (db *PostgresDB) GetUserContributionDailyCounts(ctx context.Context, userID uuid.UUID, since time.Time) (map[string]int64, error) {
	var rows []struct {
		Day   string `gorm:"column:day"`
		Count int64  `gorm:"column:cnt"`
	}
	q := `
WITH per_day AS (
	SELECT (timezone('UTC', created_at))::date AS d FROM user_activity
	WHERE user_id = ? AND created_at >= ?
	UNION ALL
	SELECT (timezone('UTC', created_at))::date FROM registry_functions
	WHERE owner_user_id IS NOT NULL AND owner_user_id = ? AND created_at >= ?
)
SELECT d::text AS day, COUNT(*)::bigint AS cnt FROM per_day GROUP BY d ORDER BY d`
	if err := db.GORM.WithContext(ctx).Raw(q, userID, since, userID, since).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to get contribution daily counts: %w", err)
	}
	out := make(map[string]int64, len(rows))
	for _, r := range rows {
		out[r.Day] = r.Count
	}
	return out, nil
}

// GetUserExecutionStats retrieves execution statistics for a user
func (db *PostgresDB) GetUserExecutionStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	// Get user's published functions from registry
	var functions []struct {
		ID             uuid.UUID
		ExecutionCount int64
		UniqueUsers    int64
	}

	if err := db.GORM.WithContext(ctx).Raw(`
		SELECT rf.id,
			COUNT(re.id)::bigint AS execution_count,
			COUNT(DISTINCT re.caller_ip)::bigint AS unique_users
		FROM registry_functions rf
		LEFT JOIN registry_function_executions re ON re.function_id = rf.id
		WHERE rf.owner_user_id = ?
		GROUP BY rf.id
	`, userID).Scan(&functions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user functions: %w", err)
	}

	var totalExecutions, totalUniqueUsers int64
	for _, fn := range functions {
		totalExecutions += fn.ExecutionCount
		totalUniqueUsers += fn.UniqueUsers
	}

	// Get execution history for last 30 days
	var history []struct {
		Date       string `json:"date"`
		Executions int64  `json:"executions"`
	}

	if err := db.GORM.WithContext(ctx).Raw(`
		SELECT DATE(re.timestamp) as date, COUNT(*) as executions
		FROM registry_function_executions re
		INNER JOIN registry_functions rf ON rf.id = re.function_id
		WHERE rf.owner_user_id = ? AND re.timestamp >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(re.timestamp)
		ORDER BY date
	`, userID).Scan(&history).Error; err != nil {
		// Continue even if no data
		history = []struct {
			Date       string `json:"date"`
			Executions int64  `json:"executions"`
		}{}
	}

	return map[string]interface{}{
		"totalExecutions":  totalExecutions,
		"totalUniqueUsers": totalUniqueUsers,
		"functionCount":    len(functions),
		"executionHistory": history,
	}, nil
}

// GetUserProfileStats returns aggregate stats for public profile cards (functions count, executions, trust, follows).
func (db *PostgresDB) GetUserProfileStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	execStats, err := db.GetUserExecutionStats(ctx, userID)
	if err != nil {
		return nil, err
	}
	fc := 0
	if v, ok := execStats["functionCount"].(int); ok {
		fc = v
	}
	te, _ := execStats["totalExecutions"].(int64)

	followers, err := db.GetUserFollowerCount(ctx, userID)
	if err != nil {
		logrus.WithError(err).WithField("userID", userID).Warn("GetUserFollowerCount failed")
		followers = 0
	}
	following, err := db.GetUserFollowingCount(ctx, userID)
	if err != nil {
		logrus.WithError(err).WithField("userID", userID).Warn("GetUserFollowingCount failed")
		following = 0
	}

	var avgOverall float64
	if err := db.GORM.Raw(`
		SELECT COALESCE(AVG(rat.overall_score), 0)
		FROM registry_functions rf
		LEFT JOIN registry_function_ratings rat ON rat.function_id = rf.id
		WHERE rf.owner_user_id = ?`, userID).Scan(&avgOverall).Error; err != nil {
		logrus.WithError(err).WithField("userID", userID).Warn("avg overall_score for profile trust failed")
		avgOverall = 0
	}

	trustScore := int(avgOverall*100.0 + 0.5)
	if trustScore < 0 {
		trustScore = 0
	}
	if trustScore > 100 {
		trustScore = 100
	}

	// Fetch reputation profile data
	var repProfile struct {
		BuilderScore        int
		OptimizerScore      int
		MentorScore         int
		AgentWhispererScore int
		OverallScore        int
		Tier                string
	}
	hasReputationProfile := true
	if err := db.GORM.Raw(`
		SELECT
			builder_score,
			optimizer_score,
			mentor_score,
			agent_whisperer_score,
			overall_score,
			tier
		FROM reputation_profiles
		WHERE user_id = ?`, userID).Scan(&repProfile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			hasReputationProfile = false
		} else {
			logrus.WithError(err).WithField("userID", userID).Warn("Failed to get reputation profile")
		}
	}

	stats := map[string]interface{}{
		"functionsCount":  fc,
		"totalExecutions": te,
		"trustScore":      trustScore,
		"followersCount":  followers,
		"followingCount": following,
	}

	// Add reputation profile data if available
	if hasReputationProfile {
		stats["builderScore"] = repProfile.BuilderScore
		stats["optimizerScore"] = repProfile.OptimizerScore
		stats["mentorScore"] = repProfile.MentorScore
		stats["agentWhispererScore"] = repProfile.AgentWhispererScore
		stats["overallReputationScore"] = repProfile.OverallScore
		stats["reputationTier"] = repProfile.Tier
	}

	return stats, nil
}

// GetUserTrustBreakdown returns per-component trust metrics aggregated across all of a user's functions.
func (db *PostgresDB) GetUserTrustBreakdown(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	var result struct {
		Reliability       float64
		Latency          float64
		ErrorRate        float64
		UserRating       float64
		Verification     float64
		TotalCalls       int64
		SuccessRate      float64
		AvgP50LatencyMs  float64
		AvgP95LatencyMs  float64
		FunctionsWithTrust int
	}

	// Aggregate reliability_score, success_rate, latency, and verification across user's functions
	if err := db.GORM.WithContext(ctx).Raw(`
		SELECT
			COALESCE(AVG(rat.reliability_score), 0) AS reliability,
			COALESCE(AVG(1 - rat.error_rate), 0) AS error_rate,
			COALESCE(AVG(rat.success_rate), 0) AS success_rate,
			COALESCE(AVG(rat.trust_score), 0) AS user_rating,
			COALESCE(SUM(CASE WHEN rat.trust_score > 0 THEN 1 ELSE 0 END), 0) AS functions_with_trust
		FROM registry_functions rf
		LEFT JOIN registry_function_ratings rat ON rat.function_id = rf.id
		WHERE rf.owner_user_id = ?
	`, userID).Scan(&result).Error; err != nil {
		return nil, err
	}

	// Aggregate execution-level stats (latency) from function_executions
	latencyRow := struct {
		AvgP50 float64
		AvgP95 float64
	}{}
	if err := db.GORM.WithContext(ctx).Raw(`
		SELECT
			COALESCE(AVG(fe.duration_ms FILTER (WHERE fe.percentile = 50)), 0) AS avg_p50,
			COALESCE(AVG(fe.duration_ms FILTER (WHERE fe.percentile = 95)), 0) AS avg_p95
		FROM registry_functions rf
		LEFT JOIN function_executions fe ON fe.function_id = rf.id
		WHERE rf.owner_user_id = ?
	`, userID).Scan(&latencyRow).Error; err != nil {
		logrus.WithError(err).Warn("GetUserTrustBreakdown: failed to get latency stats")
	}

	// Verification: ratio of verified functions
	var verifiedRatio float64
	if err := db.GORM.WithContext(ctx).Raw(`
		SELECT
			COALESCE(
				AVG(CASE WHEN rat.trust_score >= 0.8 THEN 1.0 ELSE 0.0 END),
				0
			) AS verification
		FROM registry_functions rf
		LEFT JOIN registry_function_ratings rat ON rat.function_id = rf.id
		WHERE rf.owner_user_id = ?
	`, userID).Scan(&verifiedRatio).Error; err != nil {
		logrus.WithError(err).Warn("GetUserTrustBreakdown: failed to get verification ratio")
	}

	// Total calls from executions
	var totalCalls int64
	if err := db.GORM.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(fe.call_count), 0)
		FROM registry_functions rf
		LEFT JOIN function_execution_stats fe ON fe.function_id = rf.id
		WHERE rf.owner_user_id = ?
	`, userID).Scan(&totalCalls).Error; err != nil {
		logrus.WithError(err).Warn("GetUserTrustBreakdown: failed to get total calls")
	}

	// Scale reliability and verification from 0-1 to 0-100
	reliabilityScaled := result.Reliability * 100
	if reliabilityScaled > 100 {
		reliabilityScaled = 100
	}
	verificationScaled := verifiedRatio * 100
	if verificationScaled > 100 {
		verificationScaled = 100
	}

	return map[string]interface{}{
		"reliability":          reliabilityScaled,
		"latency":              latencyRow.AvgP50,        // already in ms
		"error_rate":           result.ErrorRate * 100,   // scale 0-1 to 0-100
		"user_rating":          result.UserRating * 100,  // scale 0-1 to 0-100
		"verification":         verificationScaled,
		"total_calls":          totalCalls,
		"success_rate":         result.SuccessRate * 100, // scale 0-1 to 0-100
		"avg_p50_latency_ms":   latencyRow.AvgP50,
		"avg_p95_latency_ms":   latencyRow.AvgP95,
		"functions_with_trust": result.FunctionsWithTrust,
	}, nil
}

// GetUserPopularFunctions retrieves most popular functions for a user
func (db *PostgresDB) GetUserPopularFunctions(ctx context.Context, userID uuid.UUID, limit int) ([]map[string]interface{}, error) {
	var functions []map[string]interface{}

	if err := db.GORM.WithContext(ctx).Raw(`
		SELECT
			rf.id,
			rf.name,
			rf.description,
			COALESCE(exec.cnt, 0)::bigint AS execution_count,
			COALESCE(ratings.overall_score, 0) AS rating,
			COALESCE(ratings.total_ratings, 0)::bigint AS total_ratings
		FROM registry_functions rf
		LEFT JOIN (
			SELECT function_id, COUNT(*)::bigint AS cnt
			FROM registry_function_executions
			GROUP BY function_id
		) exec ON exec.function_id = rf.id
		LEFT JOIN registry_function_ratings ratings ON ratings.function_id = rf.id
		WHERE rf.owner_user_id = ?
		ORDER BY exec.cnt DESC NULLS LAST
		LIMIT ?
	`, userID, limit).Scan(&functions).Error; err != nil {
		return nil, fmt.Errorf("failed to get popular functions: %w", err)
	}

	return functions, nil
}

// GetUserGeographicStats retrieves geographic distribution of executions
func (db *PostgresDB) GetUserGeographicStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	var regions []struct {
		Region     string `json:"region"`
		Executions int64  `json:"executions"`
	}

	if err := db.GORM.WithContext(ctx).Raw(`
		SELECT
			COALESCE(re.geo_country, 'unknown') as region,
			COUNT(*) as executions
		FROM registry_function_executions re
		INNER JOIN registry_functions rf ON rf.id = re.function_id
		WHERE rf.owner_user_id = ? AND re.timestamp >= NOW() - INTERVAL '30 days'
		GROUP BY re.geo_country
		ORDER BY executions DESC
	`, userID).Scan(&regions).Error; err != nil {
		regions = []struct {
			Region     string `json:"region"`
			Executions int64  `json:"executions"`
		}{}
	}

	return map[string]interface{}{
		"regions": regions,
	}, nil
}

// GetUserDeviceStats retrieves device/browser statistics
func (db *PostgresDB) GetUserDeviceStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	var devices []struct {
		Device     string `json:"device"`
		Executions int64  `json:"executions"`
	}

	if err := db.GORM.WithContext(ctx).Raw(`
		SELECT
			COALESCE(re.user_agent, 'unknown') as device,
			COUNT(*) as executions
		FROM registry_function_executions re
		INNER JOIN registry_functions rf ON rf.id = re.function_id
		WHERE rf.owner_user_id = ? AND re.timestamp >= NOW() - INTERVAL '30 days'
		GROUP BY re.user_agent
		ORDER BY executions DESC
		LIMIT 5
	`, userID).Scan(&devices).Error; err != nil {
		devices = []struct {
			Device     string `json:"device"`
			Executions int64  `json:"executions"`
		}{}
	}

	return map[string]interface{}{
		"devices": devices,
	}, nil
}
