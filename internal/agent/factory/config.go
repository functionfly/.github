package factory

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

const (
	RunStatusPending   = "pending"
	RunStatusRunning   = "running"
	RunStatusSucceeded = "succeeded"
	RunStatusFailed    = "failed"
	RunStatusReview    = "needs_review"
)

// FactoryConfig stores the factory configuration in the database.
type FactoryConfig struct {
	ID                     string  `json:"id" gorm:"type:text;primaryKey"`
	AgentID                string  `json:"agent_id" gorm:"not null"`
	DiscoveryBatchSize     int     `json:"discovery_batch_size" gorm:"not null;default:10"`
	MinimumQualityScore    float64 `json:"minimum_quality_score" gorm:"type:decimal(5,2);not null;default:70"`
	MinimumTestScore       float64 `json:"minimum_test_score" gorm:"type:decimal(5,2);not null;default:80"`
	RequireAllTestsPass    bool    `json:"require_all_tests_pass" gorm:"not null;default:true"`
	AutoPublish            bool    `json:"auto_publish" gorm:"not null;default:true"`
	MaxOpportunitiesPerRun int     `json:"max_opportunities_per_run" gorm:"not null;default:3"`
	RetryAttempts          int     `json:"retry_attempts" gorm:"not null;default:1"`
	RetryBackoffMs         int     `json:"retry_backoff_ms" gorm:"not null;default:500"`
	// Scheduling configuration
	ScheduleEnabled  bool   `json:"schedule_enabled" gorm:"not null;default:false"`
	ScheduleCron     string `json:"schedule_cron" gorm:"type:text"`
	ScheduleTimezone string `json:"schedule_timezone" gorm:"type:text;default:'UTC'"`
	// Extended settings
	NotificationWebhookURL     string          `json:"notification_webhook_url" gorm:"type:text"`
	RateLimitPerHour           int             `json:"rate_limit_per_hour" gorm:"not null;default:10"`
	MaxConcurrentRuns         int             `json:"max_concurrent_runs" gorm:"not null;default:1"`
	DryRunMode                bool            `json:"dry_run_mode" gorm:"not null;default:false"`
	DiscoverySources          json.RawMessage `json:"discovery_sources" gorm:"type:jsonb;default:'[]'"`
	FeatureFlags              json.RawMessage `json:"feature_flags" gorm:"type:jsonb;default:'{}'"`
	ApprovalRequiredAboveQuality int         `json:"approval_required_above_quality" gorm:"not null;default:0"`
	ApprovalRequiredAboveTest int            `json:"approval_required_above_test" gorm:"not null;default:0"`
	LogLevel                  string          `json:"log_level" gorm:"type:text;default:'info'"`
	NotifyOnFailure           bool            `json:"notify_on_failure" gorm:"not null;default:true"`
	NotifyOnReviewRequired    bool            `json:"notify_on_review_required" gorm:"not null;default:true"`
	DiscoveryCooldownMinutes  int             `json:"discovery_cooldown_minutes" gorm:"not null;default:60"`
	MaxVersionsPerFunction    int             `json:"max_versions_per_function" gorm:"not null;default:5"`
	CreatedAt                 time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                 time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt                 gorm.DeletedAt  `json:"-" gorm:"index"`
}

func (FactoryConfig) TableName() string { return "factory_config" }

// ToConfig converts FactoryConfig to the Config struct used by the service.
func (fc *FactoryConfig) ToConfig() Config {
	cfg := Config{
		AgentID:                     fc.AgentID,
		DiscoveryBatchSize:          fc.DiscoveryBatchSize,
		MinimumQualityScore:         fc.MinimumQualityScore,
		MinimumTestScore:            fc.MinimumTestScore,
		RequireAllTestsPass:         fc.RequireAllTestsPass,
		AutoPublish:                 fc.AutoPublish,
		MaxOpportunitiesPerRun:      fc.MaxOpportunitiesPerRun,
		RetryAttempts:               fc.RetryAttempts,
		RetryBackoff:                time.Duration(fc.RetryBackoffMs) * time.Millisecond,
		RetryBackoffMs:              fc.RetryBackoffMs,
		ScheduleEnabled:             fc.ScheduleEnabled,
		ScheduleCron:                fc.ScheduleCron,
		ScheduleTimezone:            fc.ScheduleTimezone,
		NotificationWebhookURL:      fc.NotificationWebhookURL,
		RateLimitPerHour:            fc.RateLimitPerHour,
		MaxConcurrentRuns:           fc.MaxConcurrentRuns,
		DryRunMode:                  fc.DryRunMode,
		ApprovalRequiredAboveQuality: fc.ApprovalRequiredAboveQuality,
		ApprovalRequiredAboveTest:   fc.ApprovalRequiredAboveTest,
		LogLevel:                    fc.LogLevel,
		NotifyOnFailure:             fc.NotifyOnFailure,
		NotifyOnReviewRequired:      fc.NotifyOnReviewRequired,
		DiscoveryCooldownMinutes:    fc.DiscoveryCooldownMinutes,
		MaxVersionsPerFunction:      fc.MaxVersionsPerFunction,
	}
	if len(fc.DiscoverySources) > 0 {
		_ = json.Unmarshal(fc.DiscoverySources, &cfg.DiscoverySources)
	}
	if len(fc.FeatureFlags) > 0 {
		_ = json.Unmarshal(fc.FeatureFlags, &cfg.FeatureFlags)
	}
	return cfg
}

// UpdateFrom applies values from Config to FactoryConfig.
func (fc *FactoryConfig) UpdateFrom(cfg Config) {
	fc.AgentID = cfg.AgentID
	fc.DiscoveryBatchSize = cfg.DiscoveryBatchSize
	fc.MinimumQualityScore = cfg.MinimumQualityScore
	fc.MinimumTestScore = cfg.MinimumTestScore
	fc.RequireAllTestsPass = cfg.RequireAllTestsPass
	fc.AutoPublish = cfg.AutoPublish
	fc.MaxOpportunitiesPerRun = cfg.MaxOpportunitiesPerRun
	fc.RetryAttempts = cfg.RetryAttempts
	fc.RetryBackoffMs = int(cfg.RetryBackoff.Milliseconds())
	fc.ScheduleEnabled = cfg.ScheduleEnabled
	fc.ScheduleCron = cfg.ScheduleCron
	fc.ScheduleTimezone = cfg.ScheduleTimezone
	fc.NotificationWebhookURL = cfg.NotificationWebhookURL
	fc.RateLimitPerHour = cfg.RateLimitPerHour
	fc.MaxConcurrentRuns = cfg.MaxConcurrentRuns
	fc.DryRunMode = cfg.DryRunMode
	fc.ApprovalRequiredAboveQuality = cfg.ApprovalRequiredAboveQuality
	fc.ApprovalRequiredAboveTest = cfg.ApprovalRequiredAboveTest
	fc.LogLevel = cfg.LogLevel
	fc.NotifyOnFailure = cfg.NotifyOnFailure
	fc.NotifyOnReviewRequired = cfg.NotifyOnReviewRequired
	fc.DiscoveryCooldownMinutes = cfg.DiscoveryCooldownMinutes
	fc.MaxVersionsPerFunction = cfg.MaxVersionsPerFunction
	fc.DiscoverySources, _ = json.Marshal(cfg.DiscoverySources)
	if fc.DiscoverySources == nil {
		fc.DiscoverySources = []byte("[]")
	}
	fc.FeatureFlags, _ = json.Marshal(cfg.FeatureFlags)
	if fc.FeatureFlags == nil {
		fc.FeatureFlags = []byte("{}")
	}
}

type Config struct {
	AgentID                     string            `json:"agent_id"`
	DiscoveryBatchSize          int               `json:"discovery_batch_size"`
	MinimumQualityScore         float64           `json:"minimum_quality_score"`
	MinimumTestScore            float64           `json:"minimum_test_score"`
	RequireAllTestsPass         bool              `json:"require_all_tests_pass"`
	AutoPublish                 bool              `json:"auto_publish"`
	MaxOpportunitiesPerRun      int               `json:"max_opportunities_per_run"`
	RetryAttempts               int               `json:"retry_attempts"`
	RetryBackoff                time.Duration     `json:"-"`
	RetryBackoffMs              int               `json:"retry_backoff_ms"`
	ScheduleEnabled             bool              `json:"schedule_enabled"`
	ScheduleCron                string            `json:"schedule_cron"`
	ScheduleTimezone            string            `json:"schedule_timezone"`
	NotificationWebhookURL      string            `json:"notification_webhook_url"`
	RateLimitPerHour            int               `json:"rate_limit_per_hour"`
	MaxConcurrentRuns           int               `json:"max_concurrent_runs"`
	DryRunMode                  bool              `json:"dry_run_mode"`
	DiscoverySources            []string          `json:"discovery_sources"`
	FeatureFlags                map[string]bool   `json:"feature_flags"`
	ApprovalRequiredAboveQuality int              `json:"approval_required_above_quality"`
	ApprovalRequiredAboveTest   int               `json:"approval_required_above_test"`
	LogLevel                    string            `json:"log_level"`
	NotifyOnFailure             bool              `json:"notify_on_failure"`
	NotifyOnReviewRequired      bool              `json:"notify_on_review_required"`
	DiscoveryCooldownMinutes    int               `json:"discovery_cooldown_minutes"`
	MaxVersionsPerFunction      int               `json:"max_versions_per_function"`
}

func DefaultConfig(agentID string) Config {
	return Config{
		AgentID:                     agentID,
		DiscoveryBatchSize:          10,
		MinimumQualityScore:         70,
		MinimumTestScore:            80,
		RequireAllTestsPass:         true,
		AutoPublish:                 true,
		MaxOpportunitiesPerRun:      3,
		RetryAttempts:               1,
		RetryBackoff:                500 * time.Millisecond,
		RetryBackoffMs:              500,
		ScheduleEnabled:             false,
		ScheduleCron:                "0 0 * * *",
		ScheduleTimezone:            "UTC",
		RateLimitPerHour:            10,
		MaxConcurrentRuns:           1,
		ApprovalRequiredAboveQuality: 0,
		ApprovalRequiredAboveTest:   0,
		LogLevel:                    "info",
		NotifyOnFailure:             true,
		NotifyOnReviewRequired:      true,
		DiscoveryCooldownMinutes:    60,
		MaxVersionsPerFunction:      5,
		DiscoverySources:            []string{},
		FeatureFlags:                map[string]bool{},
	}
}
