package registry

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CanaryConfig represents a canary deployment configuration for a function version
type CanaryConfig struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID        uuid.UUID  `json:"function_id" gorm:"type:uuid;not null;index"`
	FunctionVersionID *uuid.UUID `json:"function_version_id,omitempty" gorm:"type:uuid"`
	Version           string     `json:"version" gorm:"not null"`
	TrafficPercent    int        `json:"traffic_percent" gorm:"default:10"`
	AutoPromote       bool       `json:"auto_promote" gorm:"default:false"`
	PromoteThreshold  float64    `json:"promote_threshold" gorm:"default:0.01"` // 1% error rate threshold
	PromoteWindow     int        `json:"promote_window" gorm:"default:300"`     // seconds
	Status            string     `json:"status" gorm:"default:'active'"`        // active, promoted, rolled_back, cancelled
	ErrorRate         float64    `json:"error_rate" gorm:"default:0"`
	RequestCount      int        `json:"request_count" gorm:"default:0"`
	SuccessCount      int        `json:"success_count" gorm:"default:0"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	PromotedAt        *time.Time `json:"promoted_at,omitempty"`
	RolledBackAt      *time.Time `json:"rolled_back_at,omitempty"`
}

// CanaryMetrics stores canary deployment metrics
type CanaryMetrics struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CanaryConfigID uuid.UUID `json:"canary_config_id" gorm:"type:uuid;not null;index"`
	Timestamp      time.Time `json:"timestamp" gorm:"index"`
	ErrorRate      float64   `json:"error_rate"`
	LatencyP50     float64   `json:"latency_p50"`
	LatencyP95     float64   `json:"latency_p95"`
	LatencyP99     float64   `json:"latency_p99"`
	RequestCount   int       `json:"request_count"`
	SuccessCount   int       `json:"success_count"`
	ErrorCount     int       `json:"error_count"`
}

// BeforeCreate hook for CanaryConfig
func (c *CanaryConfig) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// CanaryConfigRepository provides methods for canary config storage
type CanaryConfigRepository struct {
	db *gorm.DB
}

// NewCanaryConfigRepository creates a new canary config repository
func NewCanaryConfigRepository(db *gorm.DB) *CanaryConfigRepository {
	return &CanaryConfigRepository{db: db}
}

// Create creates a new canary configuration
func (r *CanaryConfigRepository) Create(config *CanaryConfig) error {
	return r.db.Create(config).Error
}

// GetByID returns a canary config by ID
func (r *CanaryConfigRepository) GetByID(id uuid.UUID) (*CanaryConfig, error) {
	var config CanaryConfig
	err := r.db.First(&config, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetByFunctionID returns active canary config for a function
func (r *CanaryConfigRepository) GetByFunctionID(functionID uuid.UUID) (*CanaryConfig, error) {
	var config CanaryConfig
	err := r.db.Where("function_id = ? AND status = ?", functionID, "active").First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetAllByFunctionID returns all canary configs for a function
func (r *CanaryConfigRepository) GetAllByFunctionID(functionID uuid.UUID) ([]*CanaryConfig, error) {
	var configs []*CanaryConfig
	err := r.db.Where("function_id = ?", functionID).Order("created_at DESC").Find(&configs).Error
	return configs, err
}

// Update updates a canary configuration
func (r *CanaryConfigRepository) Update(config *CanaryConfig) error {
	return r.db.Save(config).Error
}

// UpdateStatus updates the status of a canary configuration
func (r *CanaryConfigRepository) UpdateStatus(id uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == "promoted" {
		now := time.Now()
		updates["promoted_at"] = &now
	} else if status == "rolled_back" {
		now := time.Now()
		updates["rolled_back_at"] = &now
	}

	return r.db.Model(&CanaryConfig{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateMetrics updates the metrics for a canary configuration
func (r *CanaryConfigRepository) UpdateMetrics(id uuid.UUID, metrics *CanaryMetrics) error {
	metrics.CanaryConfigID = id
	return r.db.Create(metrics).Error
}

// Delete deletes a canary configuration
func (r *CanaryConfigRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&CanaryConfig{}, "id = ?", id).Error
}

// GetMetricsHistory returns metrics for a canary config within a time window
func (r *CanaryConfigRepository) GetMetricsHistory(canaryID uuid.UUID, since time.Time) ([]*CanaryMetrics, error) {
	var metrics []*CanaryMetrics
	err := r.db.Where("canary_config_id = ? AND timestamp > ?", canaryID, since).
		Order("timestamp ASC").Find(&metrics).Error
	return metrics, err
}

// RecordCanaryRequest records a request for a canary deployment
func (r *CanaryConfigRepository) RecordCanaryRequest(canaryID uuid.UUID, success bool) error {
	updates := map[string]interface{}{
		"request_count": gorm.Expr("request_count + 1"),
	}
	if success {
		updates["success_count"] = gorm.Expr("success_count + 1")
	}

	return r.db.Model(&CanaryConfig{}).Where("id = ?", canaryID).Updates(updates).Error
}

// CalculateErrorRate calculates the current error rate for a canary
func (r *CanaryConfigRepository) CalculateErrorRate(canaryID uuid.UUID) (float64, error) {
	var config CanaryConfig
	err := r.db.First(&config, "id = ?", canaryID).Error
	if err != nil {
		return 0, err
	}

	if config.RequestCount == 0 {
		return 0, nil
	}

	return float64(config.RequestCount-config.SuccessCount) / float64(config.RequestCount), nil
}

// AutoPromoteCheck checks if a canary should be auto-promoted
func (r *CanaryConfigRepository) AutoPromoteCheck(canaryID uuid.UUID) (bool, error) {
	config, err := r.GetByID(canaryID)
	if err != nil {
		return false, err
	}

	// Check if auto-promote is enabled
	if !config.AutoPromote {
		return false, nil
	}

	// Check if enough time has passed since creation
	windowStart := config.CreatedAt.Add(time.Duration(config.PromoteWindow) * time.Second)
	if time.Now().Before(windowStart) {
		return false, nil
	}

	// Check error rate threshold
	if config.ErrorRate <= config.PromoteThreshold {
		return true, nil
	}

	return false, nil
}
