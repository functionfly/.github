package domain

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
)

// DashboardRepository handles dashboard and analytics operations
type DashboardRepository interface {
	// Dashboard configurations
	CreateDashboardConfig(ctx context.Context, config *types.DashboardConfig) (*types.DashboardConfig, error)
	GetDashboardConfigsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*types.DashboardConfig, error)
	GetDashboardConfigsByUser(ctx context.Context, userID uuid.UUID) ([]*types.DashboardConfig, error)
	GetDashboardConfigByID(ctx context.Context, configID uuid.UUID) (*types.DashboardConfig, error)
	UpdateDashboardConfig(ctx context.Context, configID uuid.UUID, updates map[string]interface{}) (*types.DashboardConfig, error)
	DeleteDashboardConfig(ctx context.Context, configID uuid.UUID) error

	// Aggregations and metrics
	GetUsageByDay(ctx context.Context, tenantID uuid.UUID, days int) ([]types.UsageByDay, error)
	GetExecutionRateByHour(ctx context.Context, tenantID uuid.UUID, hours int) ([]types.ExecutionRateByHour, error)
	GetRecentActivityForTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]types.DashboardActivityItem, error)
	GetDashboardMetrics(ctx context.Context, tenantID uuid.UUID) (*types.DashboardMetrics, error)

	// Tenant analytics
	InitializeTenantAnalytics(ctx context.Context, tenantID uuid.UUID) error
	GetUserExecutionStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	GetUserProfileStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	GetUserTrustBreakdown(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	GetUserPopularFunctions(ctx context.Context, userID uuid.UUID, limit int) ([]map[string]interface{}, error)
	GetUserGeographicStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	GetUserDeviceStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
}

// UserProfileRepository handles user profile, achievements, and social features
type UserProfileRepository interface {
	// User skills
	GetUserSkills(ctx context.Context, userID uuid.UUID) ([]*types.UserSkill, error)
	AddUserSkill(ctx context.Context, skill *types.UserSkill) error
	RemoveUserSkill(ctx context.Context, skillID uuid.UUID) error

	// Achievements
	GetUserAchievements(ctx context.Context, userID uuid.UUID) ([]*types.UserAchievement, error)
	GetAchievementBySlug(ctx context.Context, slug string) (*types.Achievement, error)
	ListAchievements(ctx context.Context) ([]*types.Achievement, error)
	AwardAchievement(ctx context.Context, userID, achievementID uuid.UUID, metadata map[string]interface{}) error
	UpdateAchievementProgress(ctx context.Context, userAchievementID uuid.UUID, progress int, isCompleted bool) error

	// Activity
	GetUserActivity(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*types.UserActivity, error)
	CreateUserActivity(ctx context.Context, activity *types.UserActivity) error
	GetUserContributionDailyCounts(ctx context.Context, userID uuid.UUID, since time.Time) (map[string]int64, error)

	// Follows - Users
	FollowUser(ctx context.Context, followerID, followedUserID uuid.UUID, reason *string, notifyOnNewFunction, notifyOnFunctionUpdate, notifyOnNewVersion bool) (*types.UserFollow, error)
	UnfollowUser(ctx context.Context, followerID, followedUserID uuid.UUID) error
	IsFollowingUser(ctx context.Context, followerID, followedUserID uuid.UUID) (bool, error)
	GetUserFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*types.UserFollow, int, error)
	GetUserFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*types.UserFollow, int, error)
	GetUserFollowerCount(ctx context.Context, userID uuid.UUID) (int, error)
	GetUserFollowingCount(ctx context.Context, userID uuid.UUID) (int, error)

	// Follows - Functions
	FollowFunction(ctx context.Context, userID, functionID uuid.UUID, reason *string, notifyOnNewVersion, notifyOnRatingChange, notifyOnTrustChange, notifyOnVerification bool) (*types.FunctionFollow, error)
	UnfollowFunction(ctx context.Context, userID, functionID uuid.UUID) error
	IsFollowingFunction(ctx context.Context, userID, functionID uuid.UUID) (bool, error)
	GetFunctionFollowers(ctx context.Context, functionID uuid.UUID, limit, offset int) ([]*types.FunctionFollow, int, error)
	GetUserFunctionFollows(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*types.FunctionFollow, int, error)
	GetFunctionFollowerCount(ctx context.Context, functionID uuid.UUID) (int, error)

	// Favorites
	AddFavorite(ctx context.Context, userID, functionID uuid.UUID, position int) (*types.FunctionFavorite, error)
	RemoveFavorite(ctx context.Context, userID, functionID uuid.UUID) error
	IsFavorite(ctx context.Context, userID, functionID uuid.UUID) (bool, error)
	GetUserFavorites(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*types.FunctionFavorite, int, error)
	GetFavoriteCount(ctx context.Context, functionID uuid.UUID) (int, error)
	UpdateFavoritePosition(ctx context.Context, userID, functionID uuid.UUID, position int) error
	GetFavoriteByFunction(ctx context.Context, userID, functionID uuid.UUID) (*types.FunctionFavorite, error)
}

// IncidentRepository handles incident management
type IncidentRepository interface {
	CreateIncident(ctx context.Context, incident *types.Incident) (*types.Incident, error)
	GetIncidentByID(ctx context.Context, incidentID uuid.UUID) (*types.Incident, error)
	ListIncidents(ctx context.Context, limit int, offset int, status *string) ([]*types.Incident, error)
	ListIncidentsSince(ctx context.Context, since time.Time, limit int) ([]*types.Incident, error)
	CountIncidentsSince(ctx context.Context, since time.Time) (int, error)
	CountIncidentsGroupedByDay(ctx context.Context, since time.Time) ([]types.DailyIncidentCount, error)
	GetTotalDowntimeMinutesSince(ctx context.Context, since time.Time) (int, error)
	UpdateIncident(ctx context.Context, incidentID uuid.UUID, updates map[string]interface{}) (*types.Incident, error)
	ResolveIncident(ctx context.Context, incidentID uuid.UUID) (*types.Incident, error)
}

// PlatformFeatureRepository handles platform configuration
type PlatformFeatureRepository interface {
	ListFeatureMeasures(ctx context.Context) ([]*types.FeatureMeasure, error)
	UpdateFeatureMeasureEnabled(ctx context.Context, id uuid.UUID, enabled bool) error
}

// EncryptionRepository handles encryption/decryption operations
type EncryptionRepository interface {
	EncryptField(ctx context.Context, value string) (string, error)
	DecryptField(ctx context.Context, value string) (string, error)
}

// PostgreSQLRepository handles PostgreSQL-specific operations
type PostgreSQLRepository interface {
	PgNotify(channel, payload string) error
	PgListen(ctx context.Context, channel string) error
	PgWaitForNotification(ctx context.Context) (*types.PgNotification, error)
}

// TeamInviteRepository handles team invitation management
type TeamInviteRepository interface {
	CreateTeamInvite(ctx context.Context, invite *types.TeamInvite) error
	GetTeamInviteByToken(ctx context.Context, token string) (*types.TeamInvite, error)
	GetTeamInvitesByTeam(ctx context.Context, teamID uuid.UUID) ([]*types.TeamInvite, error)
	UpdateTeamInviteStatus(ctx context.Context, inviteID uuid.UUID, status string) error
	GetTeamByUserID(ctx context.Context, userID uuid.UUID) (*types.Team, error)
	IsTeamAdmin(ctx context.Context, userID uuid.UUID, teamID string) (bool, error)
}
