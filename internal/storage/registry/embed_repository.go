package registry

import (
	"encoding/json"
	"fmt"

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
