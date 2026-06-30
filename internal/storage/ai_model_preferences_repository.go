package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ModelSelection struct {
	Provider string `json:"provider"`
	ModelID  string `json:"model_id"`
}

type TenantAIPreferences struct {
	TenantID                uuid.UUID                  `json:"tenant_id"`
	Profile                 string                     `json:"profile"`
	GlobalDefault           *ModelSelection            `json:"global_default,omitempty"`
	UseSameModelEverywhere  bool                       `json:"use_same_model_everywhere"`
	Defaults                map[string]ModelSelection  `json:"defaults"`
	EnabledModels           []ModelSelection           `json:"enabled_models"`
	EnabledProviders        []string                   `json:"enabled_providers"`
	AllowUserOverrides      bool                       `json:"allow_user_overrides"`
	RoutingStrategy         string                     `json:"routing_strategy"`
	UpdatedBy               *uuid.UUID                 `json:"updated_by,omitempty"`
	CreatedAt               time.Time                  `json:"created_at"`
	UpdatedAt               time.Time                  `json:"updated_at"`
}

type TenantAIPreferencesUpdate struct {
	Profile                string                    `json:"profile"`
	GlobalDefault          *ModelSelection           `json:"global_default,omitempty"`
	UseSameModelEverywhere bool                      `json:"use_same_model_everywhere"`
	Defaults               map[string]ModelSelection `json:"defaults"`
	EnabledModels          []ModelSelection          `json:"enabled_models"`
	EnabledProviders       []string                  `json:"enabled_providers"`
	AllowUserOverrides     bool                      `json:"allow_user_overrides"`
	RoutingStrategy        string                    `json:"routing_strategy"`
}

type AIModelPreferencesRepository struct {
	db *sql.DB
}

func NewAIModelPreferencesRepository(db *sql.DB) *AIModelPreferencesRepository {
	return &AIModelPreferencesRepository{db: db}
}

func DefaultTenantAIPreferences(tenantID uuid.UUID) *TenantAIPreferences {
	return &TenantAIPreferences{
		TenantID:               tenantID,
		Profile:                "balanced",
		GlobalDefault:          nil,
		UseSameModelEverywhere: false,
		Defaults:               map[string]ModelSelection{},
		EnabledModels:          []ModelSelection{},
		EnabledProviders:       []string{},
		AllowUserOverrides:     true,
		RoutingStrategy:        "quality_first",
	}
}

func (r *AIModelPreferencesRepository) GetTenantAIPreferences(ctx context.Context, tenantID uuid.UUID) (*TenantAIPreferences, error) {
	const query = `
		SELECT profile, global_default, use_same_model_everywhere, defaults, enabled_models,
		       enabled_providers, allow_user_overrides, routing_strategy, updated_by, created_at, updated_at
		FROM tenant_ai_preferences
		WHERE tenant_id = $1`

	var (
		prefs                      TenantAIPreferences
		globalDefaultRaw           []byte
		defaultsRaw                []byte
		enabledModelsRaw           []byte
		enabledProvidersRaw        []byte
		updatedBy                  sql.NullString
	)

	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&prefs.Profile,
		&globalDefaultRaw,
		&prefs.UseSameModelEverywhere,
		&defaultsRaw,
		&enabledModelsRaw,
		&enabledProvidersRaw,
		&prefs.AllowUserOverrides,
		&prefs.RoutingStrategy,
		&updatedBy,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return DefaultTenantAIPreferences(tenantID), nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tenant ai preferences: %w", err)
	}

	prefs.TenantID = tenantID
	prefs.Defaults = map[string]ModelSelection{}
	prefs.EnabledModels = []ModelSelection{}
	prefs.EnabledProviders = []string{}

	if len(globalDefaultRaw) > 0 && string(globalDefaultRaw) != "null" && string(globalDefaultRaw) != "{}" {
		var selection ModelSelection
		if err := json.Unmarshal(globalDefaultRaw, &selection); err == nil && selection.ModelID != "" {
			prefs.GlobalDefault = &selection
		}
	}
	if len(defaultsRaw) > 0 {
		_ = json.Unmarshal(defaultsRaw, &prefs.Defaults)
	}
	if len(enabledModelsRaw) > 0 {
		_ = json.Unmarshal(enabledModelsRaw, &prefs.EnabledModels)
	}
	if len(enabledProvidersRaw) > 0 {
		_ = json.Unmarshal(enabledProvidersRaw, &prefs.EnabledProviders)
	}
	if updatedBy.Valid {
		if id, err := uuid.Parse(updatedBy.String); err == nil {
			prefs.UpdatedBy = &id
		}
	}

	return &prefs, nil
}

func (r *AIModelPreferencesRepository) UpsertTenantAIPreferences(ctx context.Context, tenantID uuid.UUID, updatedBy uuid.UUID, update TenantAIPreferencesUpdate) (*TenantAIPreferences, error) {
	globalDefaultRaw, _ := json.Marshal(update.GlobalDefault)
	defaultsRaw, _ := json.Marshal(update.Defaults)
	enabledModelsRaw, _ := json.Marshal(update.EnabledModels)
	enabledProvidersRaw, _ := json.Marshal(update.EnabledProviders)

	const query = `
		INSERT INTO tenant_ai_preferences (
			tenant_id, profile, global_default, use_same_model_everywhere, defaults, enabled_models,
			enabled_providers, allow_user_overrides, routing_strategy, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
		)
		ON CONFLICT (tenant_id) DO UPDATE SET
			profile = EXCLUDED.profile,
			global_default = EXCLUDED.global_default,
			use_same_model_everywhere = EXCLUDED.use_same_model_everywhere,
			defaults = EXCLUDED.defaults,
			enabled_models = EXCLUDED.enabled_models,
			enabled_providers = EXCLUDED.enabled_providers,
			allow_user_overrides = EXCLUDED.allow_user_overrides,
			routing_strategy = EXCLUDED.routing_strategy,
			updated_by = EXCLUDED.updated_by,
			updated_at = NOW()
		RETURNING profile, global_default, use_same_model_everywhere, defaults, enabled_models,
		          enabled_providers, allow_user_overrides, routing_strategy, updated_by, created_at, updated_at`

	var (
		prefs               TenantAIPreferences
		globalDefaultOut    []byte
		defaultsOut         []byte
		enabledModelsOut    []byte
		enabledProvidersOut []byte
		updatedByOut        sql.NullString
	)

	err := r.db.QueryRowContext(ctx, query,
		tenantID, update.Profile, globalDefaultRaw, update.UseSameModelEverywhere, defaultsRaw,
		enabledModelsRaw, enabledProvidersRaw, update.AllowUserOverrides, update.RoutingStrategy, updatedBy,
	).Scan(
		&prefs.Profile,
		&globalDefaultOut,
		&prefs.UseSameModelEverywhere,
		&defaultsOut,
		&enabledModelsOut,
		&enabledProvidersOut,
		&prefs.AllowUserOverrides,
		&prefs.RoutingStrategy,
		&updatedByOut,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert tenant ai preferences: %w", err)
	}

	prefs.TenantID = tenantID
	prefs.Defaults = map[string]ModelSelection{}
	prefs.EnabledModels = []ModelSelection{}
	prefs.EnabledProviders = []string{}

	if len(globalDefaultOut) > 0 && string(globalDefaultOut) != "null" && string(globalDefaultOut) != "{}" {
		var selection ModelSelection
		if err := json.Unmarshal(globalDefaultOut, &selection); err == nil && selection.ModelID != "" {
			prefs.GlobalDefault = &selection
		}
	}
	if len(defaultsOut) > 0 {
		_ = json.Unmarshal(defaultsOut, &prefs.Defaults)
	}
	if len(enabledModelsOut) > 0 {
		_ = json.Unmarshal(enabledModelsOut, &prefs.EnabledModels)
	}
	if len(enabledProvidersOut) > 0 {
		_ = json.Unmarshal(enabledProvidersOut, &prefs.EnabledProviders)
	}
	if updatedByOut.Valid {
		if id, err := uuid.Parse(updatedByOut.String); err == nil {
			prefs.UpdatedBy = &id
		}
	}

	return &prefs, nil
}
