package modelresolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

type Feature string

const (
	FeatureComposer   Feature = "composer"
	FeatureFRG        Feature = "frg"
	FeatureDNA        Feature = "dna"
	FeatureChat       Feature = "chat"
	FeatureSupport    Feature = "support"
	FeatureEmbeddings Feature = "embeddings"
	FeatureAgent      Feature = "agent"
)

type Resolver struct {
	prefsRepo *storage.AIModelPreferencesRepository
	repo      storage.Repository
}

func New(prefsRepo *storage.AIModelPreferencesRepository, repo storage.Repository) *Resolver {
	return &Resolver{
		prefsRepo: prefsRepo,
		repo:      repo,
	}
}

func (r *Resolver) Resolve(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	feature Feature,
	requestOverride *storage.ModelSelection,
) (*storage.ModelSelection, error) {
	prefs, err := r.prefsRepo.GetTenantAIPreferences(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if requestOverride != nil && requestOverride.Provider != "" && requestOverride.ModelID != "" {
		if !isEnabledModel(prefs.EnabledModels, requestOverride) {
			return nil, fmt.Errorf("requested model is not enabled for this tenant")
		}
		return requestOverride, nil
	}

	if prefs.AllowUserOverrides {
		settings, err := r.repo.GetUserSettings(ctx, userID)
		if err == nil {
			if sel := getUserOverride(settings, string(feature)); sel != nil && sel.Provider != "" && sel.ModelID != "" {
				if isEnabledModel(prefs.EnabledModels, sel) {
					return sel, nil
				}
			}
		}
	}

	if prefs.UseSameModelEverywhere && prefs.GlobalDefault != nil && prefs.GlobalDefault.ModelID != "" {
		return prefs.GlobalDefault, nil
	}

	if featureSel, ok := prefs.Defaults[string(feature)]; ok && featureSel.ModelID != "" {
		return &featureSel, nil
	}

	return nil, nil
}

func getUserOverride(settings map[string]interface{}, feature string) *storage.ModelSelection {
	aiRaw, ok := settings["ai"].(map[string]interface{})
	if !ok {
		return nil
	}
	overrides, ok := aiRaw["overrides"].(map[string]interface{})
	if !ok {
		return nil
	}
	featureRaw, ok := overrides[feature].(map[string]interface{})
	if !ok {
		return nil
	}
	provider, _ := featureRaw["provider"].(string)
	modelID, _ := featureRaw["model_id"].(string)
	if provider == "" || modelID == "" {
		return nil
	}
	return &storage.ModelSelection{
		Provider: strings.TrimSpace(provider),
		ModelID:  strings.TrimSpace(modelID),
	}
}

func isEnabledModel(enabled []storage.ModelSelection, model *storage.ModelSelection) bool {
	if len(enabled) == 0 {
		return true
	}
	for _, item := range enabled {
		if item.Provider == model.Provider && item.ModelID == model.ModelID {
			return true
		}
	}
	return false
}
