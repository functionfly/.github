package storage

import (
	"time"

	"github.com/google/uuid"
)

// ReputationTier represents the tier level for a user's reputation
type ReputationTier string

const (
	ReputationTierNovice     ReputationTier = "novice"
	ReputationTierContributor ReputationTier = "contributor"
	ReputationTierExpert     ReputationTier = "expert"
	ReputationTierMaster     ReputationTier = "master"
	ReputationTierLegend     ReputationTier = "legend"
)

// ReputationBadge represents a badge earned by a user
type ReputationBadge struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IconURL     string    `json:"icon_url"`
	EarnedAt    time.Time `json:"earned_at"`
	Category    string    `json:"category"` // builder, optimizer, mentor, agent_whisperer
}

// ReputationStats holds detailed statistics for a user's reputation
type ReputationStats struct {
	TotalFunctions        int `json:"total_functions"`
	VerifiedFunctions     int `json:"verified_functions"`
	TotalExecutions      int `json:"total_executions"`
	SuccessfulExecutions int `json:"successful_executions"`
	TotalMentorships     int `json:"total_mentorships"`
	ActiveMentorships    int `json:"active_mentorships"`
	SolutionsProvided    int `json:"solutions_provided"`
	HelpfulVotes         int `json:"helpful_votes"`
	BuilderPointsEarned int `json:"builder_points_earned"`
	OptimizerPointsEarned int `json:"optimizer_points_earned"`
	MentorPointsEarned   int `json:"mentor_points_earned"`
	AgentPointsEarned    int `json:"agent_points_earned"`
}

// ReputationScoreHistoryEntry represents a single entry in the score history
type ReputationScoreHistoryEntry struct {
	Timestamp          time.Time `json:"timestamp"`
	OverallScore       int       `json:"overall_score"`
	BuilderScore       int       `json:"builder_score"`
	OptimizerScore     int       `json:"optimizer_score"`
	MentorScore        int       `json:"mentor_score"`
	AgentWhispererScore int      `json:"agent_whisperer_score"`
	Tier               string    `json:"tier"`
}

// ReputationProfile represents a user's reputation profile in the system
type ReputationProfile struct {
	UserID              uuid.UUID                  `json:"user_id" gorm:"type:uuid;primaryKey"`
	TenantID            uuid.UUID                  `json:"tenant_id" gorm:"type:uuid;not null"`
	BuilderScore       int                       `json:"builder_score" gorm:"default:0"`
	OptimizerScore     int                       `json:"optimizer_score" gorm:"default:0"`
	MentorScore        int                       `json:"mentor_score" gorm:"default:0"`
	AgentWhispererScore int                      `json:"agent_whisperer_score" gorm:"default:0"`
	ReliabilityIndex    float64                   `json:"reliability_index" gorm:"type:numeric(5,4);default:1.0"`
	ConsistencyScore   float64                   `json:"consistency_score" gorm:"type:numeric(5,4);default:1.0"`
	OverallScore       int                      `json:"overall_score" gorm:"default:0"`
	Tier               ReputationTier            `json:"tier" gorm:"size:20;default:'novice'"`
	Badges             []ReputationBadge        `json:"badges" gorm:"type:jsonb;default:'[]'"`
	Stats              ReputationStats          `json:"stats" gorm:"type:jsonb;default:'{}'"`
	ScoreHistory       []ReputationScoreHistoryEntry `json:"score_history" gorm:"type:jsonb;default:'[]'"`
	UpdatedAt          time.Time                `json:"updated_at" gorm:"type:timestamptz"`

	// Cached computed values for fast reads
	TrustScore   float64 `json:"trust_score" gorm:"-"`
	FunctionCount int    `json:"function_count" gorm:"-"`
}

// TableName returns the database table name for ReputationProfile.
func (ReputationProfile) TableName() string {
	return "reputation_profiles"
}

// GetBadges returns the badges as a slice of ReputationBadge
func (r *ReputationProfile) GetBadges() []ReputationBadge {
	if r.Badges == nil {
		return []ReputationBadge{}
	}
	return r.Badges
}

// GetStats returns the stats as ReputationStats
func (r *ReputationProfile) GetStats() ReputationStats {
	return r.Stats
}

// GetScoreHistory returns the score history as a slice
func (r *ReputationProfile) GetScoreHistory() []ReputationScoreHistoryEntry {
	if r.ScoreHistory == nil {
		return []ReputationScoreHistoryEntry{}
	}
	return r.ScoreHistory
}

// CalculateOverallScore calculates the overall score from component scores
func (r *ReputationProfile) CalculateOverallScore() int {
	// Weighted average: builder 30%, optimizer 25%, mentor 25%, agent_whisperer 20%
	overall := (r.BuilderScore * 30 / 100) +
		(r.OptimizerScore * 25 / 100) +
		(r.MentorScore * 25 / 100) +
		(r.AgentWhispererScore * 20 / 100)
	return overall
}

// DetermineTier determines the tier based on overall score
func (r *ReputationProfile) DetermineTier() ReputationTier {
	score := r.CalculateOverallScore()
	switch {
	case score >= 10000:
		return ReputationTierLegend
	case score >= 5000:
		return ReputationTierMaster
	case score >= 2000:
		return ReputationTierExpert
	case score >= 500:
		return ReputationTierContributor
	default:
		return ReputationTierNovice
	}
}

// ReputationFarmingAlert represents an alert for potential reputation farming
type ReputationFarmingAlert struct {
	ID                uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type              string                `json:"type" gorm:"size:50;not null"` // self_rating, cross_account, sybil_attack, wash_trading
	Description      string                `json:"description" gorm:"type:text"`
	AffectedFunctions []uuid.UUID           `json:"affected_functions" gorm:"type:jsonb;default:'[]'"`
	AffectedUsers    []uuid.UUID           `json:"affected_users" gorm:"type:jsonb;default:'[]'"`
	Severity         string                `json:"severity" gorm:"size:20;default:'low'"` // low, medium, high, critical
	Status           string                `json:"status" gorm:"size:20;default:'open'"` // open, investigating, resolved, dismissed
	DetectedAt       time.Time             `json:"detected_at" gorm:"type:timestamptz"`
	ResolvedAt       *time.Time            `json:"resolved_at" gorm:"type:timestamptz"`
	ResolvedBy       *uuid.UUID            `json:"resolved_by" gorm:"type:uuid"`
	Notes            string                `json:"notes" gorm:"type:text"`
	Details          map[string]interface{} `json:"details" gorm:"type:jsonb;default:'{}'"`
}

// TableName returns the database table name for ReputationFarmingAlert.
func (ReputationFarmingAlert) TableName() string {
	return "reputation_farming_alerts"
}

// TrustScoreWeightsConfigV2 represents configurable trust score weights (v2 for per-component weights)
type TrustScoreWeightsConfigV2 struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string    `json:"name" gorm:"size:100;uniqueIndex"`
	Description  string    `json:"description" gorm:"type:text"`
	Reliability  float64   `json:"reliability" gorm:"type:numeric(5,4);default:0.30"`
	Latency      float64   `json:"latency" gorm:"type:numeric(5,4);default:0.20"`
	ErrorRate    float64   `json:"error_rate" gorm:"type:numeric(5,4);default:0.20"`
	UserRating   float64   `json:"user_rating" gorm:"type:numeric(5,4);default:0.15"`
	Verification float64   `json:"verification" gorm:"type:numeric(5,4);default:0.15"`
	IsActive     bool      `json:"is_active" gorm:"default:false"`
	CreatedAt    time.Time `json:"created_at" gorm:"type:timestamptz"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"type:timestamptz"`
}

// TableName returns the database table name for TrustScoreWeightsConfigV2.
func (TrustScoreWeightsConfigV2) TableName() string {
	return "trust_score_weights_config"
}

// Validate validates that weights sum to 1.0
func (t *TrustScoreWeightsConfigV2) Validate() bool {
	total := t.Reliability + t.Latency + t.ErrorRate + t.UserRating + t.Verification
	return total > 0.99 && total < 1.01
}

// ReputationEvent represents a reputation-changing event for audit trail
type ReputationEvent struct {
	ID           uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID              `json:"user_id" gorm:"type:uuid;not null;index"`
	EventType    string                `json:"event_type" gorm:"size:50;not null"` // function_published, function_updated, mentorship_started, mentorship_completed, badge_earned, etc.
	ScoreChange  int                   `json:"score_change"`
	Component    string                `json:"component" gorm:"size:50"` // builder, optimizer, mentor, agent_whisperer
	ReferenceID  *uuid.UUID            `json:"reference_id" gorm:"type:uuid"` // function_id, mentorship_id, etc.
	Description string                `json:"description" gorm:"type:text"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt   time.Time             `json:"created_at" gorm:"type:timestamptz"`
}

// TableName returns the database table name for ReputationEvent.
func (ReputationEvent) TableName() string {
	return "reputation_events"
}

// ReputationLeaderboardEntry represents a user on the leaderboard
type ReputationLeaderboardEntry struct {
	Rank               int       `json:"rank"`
	UserID             uuid.UUID `json:"user_id"`
	Username           string    `json:"username"`
	DisplayName        string    `json:"display_name"`
	OverallScore       int       `json:"overall_score"`
	Tier               string    `json:"tier"`
	BuilderScore       int       `json:"builder_score"`
	OptimizerScore     int       `json:"optimizer_score"`
	MentorScore        int       `json:"mentor_score"`
	AgentWhispererScore int      `json:"agent_whisperer_score"`
	FunctionCount      int       `json:"function_count"`
}
