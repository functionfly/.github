package brain

import (
	"github.com/functionfly/functionfly/internal/plans"
)

type BrainQuotaConfig struct {
	ContextHints    int    `json:"context_hints"`
	SyncFrequency   string `json:"sync_frequency"`
	MaxSignals      int    `json:"max_signals"`
	RetentionDays   int    `json:"retention_days"`
	SemanticSearch  bool   `json:"semantic_search"`
	ComposerEnabled bool   `json:"composer_enabled"`
	TriggerEnabled  bool   `json:"trigger_enabled"`
}

// GetBrainQuotaForPlan returns the brain quota configuration for a plan tier
func GetBrainQuotaForPlan(planTier string) BrainQuotaConfig {
	limits := plans.GetBrainConnectorLimits(planTier)
	return BrainQuotaConfig{
		ContextHints:    limits.ContextHints,
		SyncFrequency:   limits.SyncFrequency,
		MaxSignals:      limits.MaxSignals,
		RetentionDays:   limits.SignalRetentionDays,
		SemanticSearch:  limits.SemanticSearch,
		ComposerEnabled: limits.BrainComposer,
		TriggerEnabled:  limits.BrainTriggerDaemon,
	}
}

// IsBrainEnabledForPlan checks if brain is enabled for a plan
func IsBrainEnabledForPlan(planTier string) bool {
	return plans.GetBrainConnectorLimits(planTier).AgentBrainEnabled
}

// GetBrainContextHints returns the number of context hints allowed for a plan
func GetBrainContextHints(planTier string) int {
	return plans.GetBrainContextHints(planTier)
}
