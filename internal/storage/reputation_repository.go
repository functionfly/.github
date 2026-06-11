package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ReputationRepository handles database operations for reputation profiles
type ReputationRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewReputationRepository creates a new reputation repository
func NewReputationRepository(db *gorm.DB, logger *logrus.Logger) *ReputationRepository {
	if logger == nil {
		logger = logrus.New()
	}
	return &ReputationRepository{
		db:     db,
		logger: logger,
	}
}

// GetOrCreateProfile gets or creates a reputation profile for a user
func (r *ReputationRepository) GetOrCreateProfile(userID, tenantID uuid.UUID) (*ReputationProfile, error) {
	var profile ReputationProfile

	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			profile = ReputationProfile{
				UserID:    userID,
				TenantID:  tenantID,
				Tier:      ReputationTierNovice,
				Badges:    []ReputationBadge{},
				Stats:     ReputationStats{},
				UpdatedAt: time.Now(),
			}
			if err := r.db.Create(&profile).Error; err != nil {
				return nil, fmt.Errorf("failed to create reputation profile: %w", err)
			}
			return &profile, nil
		}
		return nil, fmt.Errorf("failed to get reputation profile: %w", err)
	}

	return &profile, nil
}

// GetProfile gets a reputation profile by user ID
func (r *ReputationRepository) GetProfile(userID uuid.UUID) (*ReputationProfile, error) {
	var profile ReputationProfile

	if err := r.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get reputation profile: %w", err)
	}

	return &profile, nil
}

// GetProfileByTenant gets a reputation profile with tenant info
func (r *ReputationRepository) GetProfileByTenant(userID, tenantID uuid.UUID) (*ReputationProfile, error) {
	var profile ReputationProfile

	if err := r.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get reputation profile: %w", err)
	}

	return &profile, nil
}

// UpdateProfile updates a reputation profile
func (r *ReputationRepository) UpdateProfile(profile *ReputationProfile) error {
	profile.UpdatedAt = time.Now()

	if err := r.db.Model(&ReputationProfile{}).
		Where("user_id = ?", profile.UserID).
		Updates(map[string]interface{}{
			"builder_score":        profile.BuilderScore,
			"optimizer_score":      profile.OptimizerScore,
			"mentor_score":         profile.MentorScore,
			"agent_whisperer_score": profile.AgentWhispererScore,
			"reliability_index":     profile.ReliabilityIndex,
			"consistency_score":     profile.ConsistencyScore,
			"overall_score":        profile.OverallScore,
			"tier":                 profile.Tier,
			"badges":               profile.Badges,
			"stats":                profile.Stats,
			"score_history":        profile.ScoreHistory,
			"updated_at":           profile.UpdatedAt,
		}).Error; err != nil {
		return fmt.Errorf("failed to update reputation profile: %w", err)
	}

	return nil
}

// AddBadge adds a badge to a user's reputation profile
func (r *ReputationRepository) AddBadge(userID uuid.UUID, badge ReputationBadge) error {
	profile, err := r.GetProfile(userID)
	if err != nil {
		return err
	}
	if profile == nil {
		return fmt.Errorf("reputation profile not found")
	}

	badges := profile.GetBadges()
	for _, b := range badges {
		if b.ID == badge.ID {
			return nil // Badge already exists
		}
	}
	badges = append(badges, badge)
	profile.Badges = badges

	return r.UpdateProfile(profile)
}

// RecordScoreChange records a score change and updates the profile
func (r *ReputationRepository) RecordScoreChange(userID, tenantID uuid.UUID, component string, scoreChange int) error {
	profile, err := r.GetOrCreateProfile(userID, tenantID)
	if err != nil {
		return err
	}

	// Update component score
	switch component {
	case "builder":
		profile.BuilderScore += scoreChange
	case "optimizer":
		profile.OptimizerScore += scoreChange
	case "mentor":
		profile.MentorScore += scoreChange
	case "agent_whisperer":
		profile.AgentWhispererScore += scoreChange
	}

	// Recalculate overall score and tier
	profile.OverallScore = profile.CalculateOverallScore()
	profile.Tier = profile.DetermineTier()

	// Add to score history
	historyEntry := ReputationScoreHistoryEntry{
		Timestamp:           time.Now(),
		OverallScore:        profile.OverallScore,
		BuilderScore:         profile.BuilderScore,
		OptimizerScore:       profile.OptimizerScore,
		MentorScore:          profile.MentorScore,
		AgentWhispererScore:  profile.AgentWhispererScore,
		Tier:                string(profile.Tier),
	}
	profile.ScoreHistory = append(profile.ScoreHistory, historyEntry)

	// Keep only last 100 entries
	if len(profile.ScoreHistory) > 100 {
		profile.ScoreHistory = profile.ScoreHistory[len(profile.ScoreHistory)-100:]
	}

	return r.UpdateProfile(profile)
}

// GetLeaderboard returns the top users by reputation score
func (r *ReputationRepository) GetLeaderboard(limit int, offset int) ([]ReputationLeaderboardEntry, error) {
	var profiles []ReputationProfile

	if err := r.db.Order("overall_score DESC").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	entries := make([]ReputationLeaderboardEntry, 0, len(profiles))
	for i, profile := range profiles {
		entries = append(entries, ReputationLeaderboardEntry{
			Rank:                offset + i + 1,
			UserID:              profile.UserID,
			OverallScore:        profile.OverallScore,
			Tier:                string(profile.Tier),
			BuilderScore:        profile.BuilderScore,
			OptimizerScore:      profile.OptimizerScore,
			MentorScore:         profile.MentorScore,
			AgentWhispererScore: profile.AgentWhispererScore,
		})
	}

	return entries, nil
}

// GetUserRank returns a user's rank on the leaderboard
func (r *ReputationRepository) GetUserRank(userID uuid.UUID) (int, error) {
	var rank int64

	err := r.db.Raw(`
		SELECT COUNT(*) + 1
		FROM reputation_profiles
		WHERE overall_score > (SELECT overall_score FROM reputation_profiles WHERE user_id = ?)
	`, userID).Scan(&rank).Error

	if err != nil {
		return 0, fmt.Errorf("failed to get user rank: %w", err)
	}

	return int(rank), nil
}

// GetReputationEvents returns reputation events for a user
func (r *ReputationRepository) GetReputationEvents(userID uuid.UUID, limit int, offset int) ([]ReputationEvent, error) {
	var events []ReputationEvent

	if err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to get reputation events: %w", err)
	}

	return events, nil
}

// RecordReputationEvent records a reputation event
func (r *ReputationRepository) RecordReputationEvent(event *ReputationEvent) error {
	if err := r.db.Create(event).Error; err != nil {
		return fmt.Errorf("failed to record reputation event: %w", err)
	}
	return nil
}

// CreateReputationFarmingAlert creates a new reputation farming alert
func (r *ReputationRepository) CreateReputationFarmingAlert(alert *ReputationFarmingAlert) error {
	alert.ID = uuid.New()
	alert.DetectedAt = time.Now()
	alert.Status = "open"

	if err := r.db.Create(alert).Error; err != nil {
		return fmt.Errorf("failed to create reputation farming alert: %w", err)
	}

	return nil
}

// GetReputationFarmingAlerts returns reputation farming alerts
func (r *ReputationRepository) GetReputationFarmingAlerts(status string, limit int, offset int) ([]ReputationFarmingAlert, int64, error) {
	var alerts []ReputationFarmingAlert
	var total int64

	query := r.db.Model(&ReputationFarmingAlert{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count alerts: %w", err)
	}

	if err := query.Order("detected_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&alerts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get alerts: %w", err)
	}

	return alerts, total, nil
}

// ResolveReputationFarmingAlert resolves an alert
func (r *ReputationRepository) ResolveReputationFarmingAlert(alertID, resolvedBy uuid.UUID, notes string) error {
	now := time.Now()
	if err := r.db.Model(&ReputationFarmingAlert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"status":      "resolved",
			"resolved_at": now,
			"resolved_by": resolvedBy,
			"notes":       notes,
		}).Error; err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	return nil
}

// DismissReputationFarmingAlert dismisses an alert
func (r *ReputationRepository) DismissReputationFarmingAlert(alertID, dismissedBy uuid.UUID, notes string) error {
	if err := r.db.Model(&ReputationFarmingAlert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"status":      "dismissed",
			"resolved_at": time.Now(),
			"resolved_by": dismissedBy,
			"notes":       notes,
		}).Error; err != nil {
		return fmt.Errorf("failed to dismiss alert: %w", err)
	}

	return nil
}

// DetectReputationFarming analyzes patterns and detects potential reputation farming
func (r *ReputationRepository) DetectReputationFarming(ctx context.Context) ([]ReputationFarmingAlert, error) {
	var alerts []ReputationFarmingAlert

	// Detect self-rating patterns: users rating their own functions
	selfRatingPattern := `
		SELECT
			u.id as user_id,
			COUNT(DISTINCT rfr.function_id) as function_count,
			AVG(rfr.overall_score) as avg_score
		FROM users u
		JOIN registry_functions rf ON rf.owner_user_id = u.id
		JOIN registry_function_ratings rfr ON rfr.function_id = rf.id
		WHERE rfr.total_ratings = 1
		AND rfr.overall_score >= 4.5
		GROUP BY u.id
		HAVING COUNT(DISTINCT rfr.function_id) > 5
	`

	type selfRatingResult struct {
		UserID        uuid.UUID
		FunctionCount int
		AvgScore      float64
	}

	var selfRatingResults []selfRatingResult
	if err := r.db.Raw(selfRatingPattern).Scan(&selfRatingResults).Error; err != nil {
		r.logger.WithError(err).Warn("Failed to detect self-rating patterns")
	}

	for _, result := range selfRatingResults {
		details := map[string]interface{}{
			"function_count": result.FunctionCount,
			"avg_score":      result.AvgScore,
			"pattern":        "self_rating",
		}

		alert := &ReputationFarmingAlert{
			Type:        "self_rating",
			Description: fmt.Sprintf("User has %d functions with suspiciously high single ratings", result.FunctionCount),
			Severity:    "medium",
			Details:     details,
		}
		alert.AffectedUsers = []uuid.UUID{result.UserID}

		if err := r.CreateReputationFarmingAlert(alert); err != nil {
			r.logger.WithError(err).WithField("user_id", result.UserID).Error("Failed to create self-rating alert")
		} else {
			alerts = append(alerts, *alert)
		}
	}

	// Detect cross-account rating patterns
	crossAccountPattern := `
		SELECT
			rat.function_id,
			COUNT(DISTINCT rat.user_id) as unique_raters,
			AVG(rat.overall_score) as avg_score,
			STDDEV(rat.overall_score) as score_stddev
		FROM registry_function_ratings rat
		JOIN registry_functions rf ON rf.id = rat.function_id
		WHERE rat.created_at > NOW() - INTERVAL '7 days'
		GROUP BY rat.function_id
		HAVING COUNT(DISTINCT rat.user_id) < 3
		AND AVG(rat.overall_score) >= 4.8
		AND STDDEV(rat.overall_score) < 0.3
	`

	type crossAccountResult struct {
		FunctionID    uuid.UUID
		UniqueRaters int
		AvgScore      float64
		ScoreStddev   float64
	}

	var crossAccountResults []crossAccountResult
	if err := r.db.Raw(crossAccountPattern).Scan(&crossAccountResults).Error; err != nil {
		r.logger.WithError(err).Warn("Failed to detect cross-account rating patterns")
	}

	for _, result := range crossAccountResults {
		details := map[string]interface{}{
			"unique_raters":  result.UniqueRaters,
			"avg_score":      result.AvgScore,
			"score_stddev":   result.ScoreStddev,
			"pattern":       "cross_account",
		}

		alert := &ReputationFarmingAlert{
			Type:        "cross_account",
			Description: fmt.Sprintf("Function has %d unique raters with suspiciously uniform high ratings", result.UniqueRaters),
			Severity:    "high",
			Details:     details,
		}
		alert.AffectedFunctions = []uuid.UUID{result.FunctionID}

		if err := r.CreateReputationFarmingAlert(alert); err != nil {
			r.logger.WithError(err).WithField("function_id", result.FunctionID).Error("Failed to create cross-account alert")
		} else {
			alerts = append(alerts, *alert)
		}
	}

	return alerts, nil
}

// UpdateTrustScoreWeights updates the trust score weights configuration
func (r *ReputationRepository) UpdateTrustScoreWeights(config *TrustScoreWeightsConfigV2) error {
	if !config.Validate() {
		return fmt.Errorf("trust score weights must sum to 1.0")
	}

	// Deactivate all other configs
	if err := r.db.Model(&TrustScoreWeightsConfigV2{}).
		Where("is_active = ?", true).
		Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate existing configs: %w", err)
	}

	config.IsActive = true
	config.UpdatedAt = time.Now()

	if config.ID == uuid.Nil {
		config.ID = uuid.New()
		config.CreatedAt = time.Now()
		if err := r.db.Create(config).Error; err != nil {
			return fmt.Errorf("failed to create trust score weights config: %w", err)
		}
	} else {
		if err := r.db.Save(config).Error; err != nil {
			return fmt.Errorf("failed to update trust score weights config: %w", err)
		}
	}

	return nil
}

// GetActiveTrustScoreWeights returns the active trust score weights
func (r *ReputationRepository) GetActiveTrustScoreWeights() (*TrustScoreWeightsConfigV2, error) {
	var config TrustScoreWeightsConfigV2

	if err := r.db.Where("is_active = ?", true).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default weights
			return &TrustScoreWeightsConfigV2{
				Name:         "default",
				Description: "Default trust score weights",
				Reliability:  0.30,
				Latency:      0.20,
				ErrorRate:    0.20,
				UserRating:   0.15,
				Verification: 0.15,
				IsActive:     true,
			}, nil
		}
		return nil, fmt.Errorf("failed to get trust score weights: %w", err)
	}

	return &config, nil
}

// GetTrustScoreWeightsHistory returns the history of trust score weights configurations
func (r *ReputationRepository) GetTrustScoreWeightsHistory(limit int, offset int) ([]TrustScoreWeightsConfigV2, int64, error) {
	var configs []TrustScoreWeightsConfigV2
	var total int64

	if err := r.db.Model(&TrustScoreWeightsConfigV2{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count configs: %w", err)
	}

	if err := r.db.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get trust score weights history: %w", err)
	}

	return configs, total, nil
}

// CleanupOldTrustHistory removes trust history entries older than the retention period
func (r *ReputationRepository) CleanupOldTrustHistory(ctx context.Context, retentionDays int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	result := r.db.Where("calculated_at < ?", cutoffDate).Delete(&registry.TrustHistory{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old trust history: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// GetReputationStats returns aggregated reputation statistics
func (r *ReputationRepository) GetReputationStats(ctx context.Context) (map[string]interface{}, error) {
	var stats struct {
		TotalProfiles   int64
		AvgScore        float64
		LegendCount    int64
		MasterCount     int64
		ExpertCount     int64
		ContributorCount int64
		NoviceCount     int64
	}

	if err := r.db.Raw(`
		SELECT
			COUNT(*) as total_profiles,
			AVG(overall_score) as avg_score,
			COUNT(CASE WHEN tier = 'legend' THEN 1 END) as legend_count,
			COUNT(CASE WHEN tier = 'master' THEN 1 END) as master_count,
			COUNT(CASE WHEN tier = 'expert' THEN 1 END) as expert_count,
			COUNT(CASE WHEN tier = 'contributor' THEN 1 END) as contributor_count,
			COUNT(CASE WHEN tier = 'novice' THEN 1 END) as novice_count
		FROM reputation_profiles
	`).Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get reputation stats: %w", err)
	}

	return map[string]interface{}{
		"total_profiles":    stats.TotalProfiles,
		"avg_score":         stats.AvgScore,
		"tier_distribution": map[string]int64{
			"legend":      stats.LegendCount,
			"master":      stats.MasterCount,
			"expert":      stats.ExpertCount,
			"contributor": stats.ContributorCount,
			"novice":      stats.NoviceCount,
		},
	}, nil
}
