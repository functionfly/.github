package domain

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// DashboardRepository handles dashboard and analytics operations
type DashboardRepository interface {
	// Dashboard configurations
	CreateDashboardConfig(ctx context.Context, config *storage.DashboardConfig) (*storage.DashboardConfig, error)
	GetDashboardConfigsByTenant(tenantID uuid.UUID) ([]*storage.DashboardConfig, error)
	GetDashboardConfigsByUser(userID uuid.UUID) ([]*storage.DashboardConfig, error)
	GetDashboardConfigByID(configID uuid.UUID) (*storage.DashboardConfig, error)
	UpdateDashboardConfig(ctx context.Context, configID uuid.UUID, updates map[string]interface{}) (*storage.DashboardConfig, error)
	DeleteDashboardConfig(ctx context.Context, configID uuid.UUID) error

	// Aggregations and metrics
	GetUsageByDay(ctx context.Context, tenantID uuid.UUID, days int) ([]storage.UsageByDay, error)
	GetExecutionRateByHour(ctx context.Context, tenantID uuid.UUID, hours int) ([]storage.ExecutionRateByHour, error)
	GetRecentActivityForTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]storage.DashboardActivityItem, error)
	GetDashboardMetrics(ctx context.Context, tenantID uuid.UUID) (*storage.DashboardMetrics, error)

	// Tenant analytics
	InitializeTenantAnalytics(tenantID uuid.UUID) error
	GetUserExecutionStats(userID uuid.UUID) (map[string]interface{}, error)
	GetUserProfileStats(userID uuid.UUID) (map[string]interface{}, error)
	GetUserTrustBreakdown(userID uuid.UUID) (map[string]interface{}, error)
	GetUserPopularFunctions(userID uuid.UUID, limit int) ([]map[string]interface{}, error)
	GetUserGeographicStats(userID uuid.UUID) (map[string]interface{}, error)
	GetUserDeviceStats(userID uuid.UUID) (map[string]interface{}, error)
}

// UserProfileRepository handles user profile, achievements, and social features
type UserProfileRepository interface {
	// User skills
	GetUserSkills(userID uuid.UUID) ([]*storage.UserSkill, error)
	AddUserSkill(skill *storage.UserSkill) error
	RemoveUserSkill(skillID uuid.UUID) error

	// Achievements
	GetUserAchievements(userID uuid.UUID) ([]*storage.UserAchievement, error)
	GetAchievementBySlug(slug string) (*storage.Achievement, error)
	ListAchievements() ([]*storage.Achievement, error)
	AwardAchievement(userID, achievementID uuid.UUID, metadata map[string]interface{}) error
	UpdateAchievementProgress(userAchievementID uuid.UUID, progress int, isCompleted bool) error

	// Activity
	GetUserActivity(userID uuid.UUID, limit, offset int) ([]*storage.UserActivity, error)
	CreateUserActivity(activity *storage.UserActivity) error
	GetUserContributionDailyCounts(userID uuid.UUID, since time.Time) (map[string]int64, error)

	// Follows - Users
	FollowUser(ctx context.Context, followerID, followedUserID uuid.UUID, reason *string, notifyOnNewFunction, notifyOnFunctionUpdate, notifyOnNewVersion bool) (*storage.UserFollow, error)
	UnfollowUser(ctx context.Context, followerID, followedUserID uuid.UUID) error
	IsFollowingUser(ctx context.Context, followerID, followedUserID uuid.UUID) (bool, error)
	GetUserFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*storage.UserFollow, int, error)
	GetUserFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*storage.UserFollow, int, error)
	GetUserFollowerCount(ctx context.Context, userID uuid.UUID) (int, error)
	GetUserFollowingCount(ctx context.Context, userID uuid.UUID) (int, error)

	// Follows - Functions
	FollowFunction(ctx context.Context, userID, functionID uuid.UUID, reason *string, notifyOnNewVersion, notifyOnRatingChange, notifyOnTrustChange, notifyOnVerification bool) (*storage.FunctionFollow, error)
	UnfollowFunction(ctx context.Context, userID, functionID uuid.UUID) error
	IsFollowingFunction(ctx context.Context, userID, functionID uuid.UUID) (bool, error)
	GetFunctionFollowers(ctx context.Context, functionID uuid.UUID, limit, offset int) ([]*storage.FunctionFollow, int, error)
	GetUserFunctionFollows(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*storage.FunctionFollow, int, error)
	GetFunctionFollowerCount(ctx context.Context, functionID uuid.UUID) (int, error)
}

// IncidentRepository handles incident management
type IncidentRepository interface {
	CreateIncident(ctx context.Context, incident *storage.Incident) (*storage.Incident, error)
	GetIncidentByID(ctx context.Context, incidentID uuid.UUID) (*storage.Incident, error)
	ListIncidents(ctx context.Context, limit int, offset int, status *string) ([]*storage.Incident, error)
	ListIncidentsSince(ctx context.Context, since time.Time, limit int) ([]*storage.Incident, error)
	CountIncidentsSince(ctx context.Context, since time.Time) (int, error)
	CountIncidentsGroupedByDay(ctx context.Context, since time.Time) ([]storage.DailyIncidentCount, error)
	UpdateIncident(ctx context.Context, incidentID uuid.UUID, updates map[string]interface{}) (*storage.Incident, error)
	ResolveIncident(ctx context.Context, incidentID uuid.UUID) (*storage.Incident, error)
}

// PlatformFeatureRepository handles platform configuration
type PlatformFeatureRepository interface {
	ListFeatureMeasures(ctx context.Context) ([]*storage.FeatureMeasure, error)
	UpdateFeatureMeasureEnabled(ctx context.Context, id uuid.UUID, enabled bool) error
}

// EncryptionRepository handles encryption/decryption operations
type EncryptionRepository interface {
	EncryptField(value string) (string, error)
	DecryptField(value string) (string, error)
}

// PostgreSQLRepository handles PostgreSQL-specific operations
type PostgreSQLRepository interface {
	PgNotify(channel, payload string) error
	PgListen(ctx context.Context, channel string) error
	PgWaitForNotification(ctx context.Context) (*storage.PgNotification, error)
}

// TeamInviteRepository handles team invitation management
type TeamInviteRepository interface {
	CreateTeamInvite(invite *storage.TeamInvite) error
	GetTeamInviteByToken(token string) (*storage.TeamInvite, error)
	GetTeamInvitesByTeam(teamID uuid.UUID) ([]*storage.TeamInvite, error)
	UpdateTeamInviteStatus(inviteID uuid.UUID, status string) error
	GetTeamByUserID(userID uuid.UUID) (*storage.Team, error)
	IsTeamAdmin(userID uuid.UUID, teamID string) (bool, error)
}
