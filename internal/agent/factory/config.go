package factory

import (
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
	ScheduleEnabled  bool           `json:"schedule_enabled" gorm:"not null;default:false"`
	ScheduleCron     string         `json:"schedule_cron" gorm:"type:text"`
	ScheduleTimezone string         `json:"schedule_timezone" gorm:"type:text;default:'UTC'"`
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

func (FactoryConfig) TableName() string { return "factory_config" }

// ToConfig converts FactoryConfig to the Config struct used by the service.
func (fc *FactoryConfig) ToConfig() Config {
	return Config{
		AgentID:                fc.AgentID,
		DiscoveryBatchSize:     fc.DiscoveryBatchSize,
		MinimumQualityScore:    fc.MinimumQualityScore,
		MinimumTestScore:       fc.MinimumTestScore,
		RequireAllTestsPass:    fc.RequireAllTestsPass,
		AutoPublish:            fc.AutoPublish,
		MaxOpportunitiesPerRun: fc.MaxOpportunitiesPerRun,
		RetryAttempts:          fc.RetryAttempts,
		RetryBackoff:           time.Duration(fc.RetryBackoffMs) * time.Millisecond,
		ScheduleEnabled:        fc.ScheduleEnabled,
		ScheduleCron:           fc.ScheduleCron,
		ScheduleTimezone:       fc.ScheduleTimezone,
	}
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
}

type Config struct {
	AgentID                string
	DiscoveryBatchSize     int
	MinimumQualityScore    float64
	MinimumTestScore       float64
	RequireAllTestsPass    bool
	AutoPublish            bool
	MaxOpportunitiesPerRun int
	RetryAttempts          int
	RetryBackoff           time.Duration
	// Scheduling configuration
	ScheduleEnabled  bool   `json:"schedule_enabled"`
	ScheduleCron     string `json:"schedule_cron"`
	ScheduleTimezone string `json:"schedule_timezone"`
}

func DefaultConfig(agentID string) Config {
	return Config{
		AgentID:                agentID,
		DiscoveryBatchSize:     10,
		MinimumQualityScore:    70,
		MinimumTestScore:       80,
		RequireAllTestsPass:    true,
		AutoPublish:            true,
		MaxOpportunitiesPerRun: 3,
		RetryAttempts:          1,
		RetryBackoff:           500 * time.Millisecond,
		// Default scheduling: disabled
		ScheduleEnabled:  false,
		ScheduleCron:     "0 0 * * *", // Daily at midnight
		ScheduleTimezone: "UTC",
	}
}
