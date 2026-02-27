package registry

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GetFunctionEmbedConfig retrieves the embed configuration for a function.
// Returns a default EmbedConfig (enabled=false) if no config is stored.
func (r *RegistryRepository) GetFunctionEmbedConfig(functionID uuid.UUID) (*EmbedConfig, error) {
	var fn RegistryFunction
	if err := r.db.Select("embed_config").Where("id = ?", functionID).First(&fn).Error; err != nil {
		return nil, fmt.Errorf("failed to get function embed config: %w", err)
	}

	if fn.EmbedConfig == nil {
		// Return sensible defaults when no config is set
		return &EmbedConfig{
			Enabled:          false,
			AllowedOrigins:   []string{"*"},
			RequireAPIKey:    false,
			UIEnabled:        true,
			UITheme:          "auto",
			RateLimitPerHour: 1000,
		}, nil
	}

	var cfg EmbedConfig
	if err := json.Unmarshal(fn.EmbedConfig, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse embed config: %w", err)
	}

	return &cfg, nil
}

// UpdateFunctionEmbedConfig persists the embed configuration for a function.
func (r *RegistryRepository) UpdateFunctionEmbedConfig(functionID uuid.UUID, cfg *EmbedConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal embed config: %w", err)
	}

	if err := r.db.Model(&RegistryFunction{}).
		Where("id = ?", functionID).
		Update("embed_config", data).Error; err != nil {
		return fmt.Errorf("failed to update embed config: %w", err)
	}

	return nil
}

// EmbedOriginStat holds execution count for a single embed origin domain.
type EmbedOriginStat struct {
	Origin string `json:"origin"`
	Count  int64  `json:"count"`
}

// GetEmbedAnalytics returns the top embed origin domains for a function
// within the given time window, ordered by execution count descending.
func (r *RegistryRepository) GetEmbedAnalytics(functionID uuid.UUID, since time.Time, limit int) ([]EmbedOriginStat, error) {
	type row struct {
		EmbedOrigin string
		Count       int64
	}

	var rows []row
	err := r.db.Model(&RegistryFunctionExecution{}).
		Select("embed_origin, COUNT(*) as count").
		Where("function_id = ? AND embed_origin IS NOT NULL AND timestamp >= ?", functionID, since).
		Group("embed_origin").
		Order("count DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get embed analytics: %w", err)
	}

	stats := make([]EmbedOriginStat, len(rows))
	for i, r := range rows {
		stats[i] = EmbedOriginStat{Origin: r.EmbedOrigin, Count: r.Count}
	}
	return stats, nil
}

// GetEmbedExecutionCountByOrigin returns the number of embed executions from a
// specific origin domain within the given time window (for rate limiting).
func (r *RegistryRepository) GetEmbedExecutionCountByOrigin(functionID uuid.UUID, origin string, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&RegistryFunctionExecution{}).
		Where("function_id = ? AND embed_origin = ? AND timestamp >= ?", functionID, origin, since).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count embed executions by origin: %w", err)
	}
	return count, nil
}
