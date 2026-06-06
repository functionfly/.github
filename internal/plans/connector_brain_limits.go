package plans

// ==================== Connector + Brain Feature Limits ====================

// Connector limits per plan
const (
	FreeMaxConnectors       = 1
	StarterMaxConnectors    = 3
	ProMaxConnectors        = 10
	EnterpriseMaxConnectors = -1 // Unlimited
)

// Brain signal limits per plan
const (
	FreeMaxBrainSignals       = 50
	StarterMaxBrainSignals    = 200
	ProMaxBrainSignals        = 500
	EnterpriseMaxBrainSignals = 2000
)

// Brain context hints (signals injected per query)
const (
	FreeBrainContextHints       = 3
	StarterBrainContextHints    = 10
	ProBrainContextHints        = 25
	EnterpriseBrainContextHints = 50
)

// Brain signal retention in days
const (
	FreeBrainSignalRetentionDays       = 7
	StarterBrainSignalRetentionDays    = 30
	ProBrainSignalRetentionDays        = 30
	EnterpriseBrainSignalRetentionDays = 90
)

// Brain sync frequency strings
const (
	FreeBrainSyncFrequency       = "6h"
	StarterBrainSyncFrequency    = "1h"
	ProBrainSyncFrequency        = "15m"
	EnterpriseBrainSyncFrequency = "5m"
)

// Connector availability by plan
var ConnectorAvailability = map[string]map[string]bool{
	"github":  {PlanFree: true, PlanStarter: true, PlanPro: true, PlanEnterprise: true},
	"notion":  {PlanFree: false, PlanStarter: true, PlanPro: true, PlanEnterprise: true},
	"slack":   {PlanFree: false, PlanStarter: true, PlanPro: true, PlanEnterprise: true},
	"gmail":   {PlanFree: false, PlanStarter: false, PlanPro: true, PlanEnterprise: true},
	"linear":  {PlanFree: false, PlanStarter: false, PlanPro: true, PlanEnterprise: true},
}

// BrainConnectorLimits provides full brain + connector limits for a plan
type BrainConnectorLimits struct {
	MaxConnectors        int    `json:"max_connectors"`
	MaxSignals           int    `json:"max_signals"`
	ContextHints         int    `json:"context_hints"`
	SyncFrequency        string `json:"sync_frequency"`
	SignalRetentionDays  int    `json:"signal_retention_days"`
	SemanticSearch       bool   `json:"semantic_search"`
	ImportanceLearning   bool   `json:"importance_learning"`
	BrainComposer        bool   `json:"brain_composer"`
	TrustScore           bool   `json:"trust_score"`
	AgentBrainEnabled    bool   `json:"agent_brain_enabled"`
	AgentBrainScope      string `json:"agent_brain_scope"`
	BrainTriggerDaemon   bool   `json:"brain_trigger_daemon"`
	AnomalyAlerts        bool   `json:"anomaly_alerts"`
	FlyMindBrainContext  bool   `json:"flymind_brain_context"`
	FlyMindRAGSignals    bool   `json:"flymind_rag_brain_signals"`
	FlyMindMemoryToBrain bool   `json:"flymind_memory_to_brain"`
	FlyMindEconomicBrain bool   `json:"flymind_economic_brain_enrich"`
	AnalyticsModelFeedback bool `json:"analytics_model_feedback"`
	AnalyticsExport      bool   `json:"analytics_export"`
	ConnectorGraph       bool   `json:"connector_graph"`
}

// GetBrainConnectorLimits returns the limits for a given plan
func GetBrainConnectorLimits(plan string) BrainConnectorLimits {
	switch plan {
	case PlanEnterprise:
		return BrainConnectorLimits{
			MaxConnectors:        EnterpriseMaxConnectors,
			MaxSignals:           EnterpriseMaxBrainSignals,
			ContextHints:         EnterpriseBrainContextHints,
			SyncFrequency:        EnterpriseBrainSyncFrequency,
			SignalRetentionDays:  EnterpriseBrainSignalRetentionDays,
			SemanticSearch:       true,
			ImportanceLearning:   true,
			BrainComposer:        true,
			TrustScore:           true,
			AgentBrainEnabled:    true,
			AgentBrainScope:      "both",
			BrainTriggerDaemon:   true,
			AnomalyAlerts:        true,
			FlyMindBrainContext:  true,
			FlyMindRAGSignals:    true,
			FlyMindMemoryToBrain: true,
			FlyMindEconomicBrain: true,
			AnalyticsModelFeedback: true,
			AnalyticsExport:      true,
			ConnectorGraph:       true,
		}
	case PlanPro:
		return BrainConnectorLimits{
			MaxConnectors:        ProMaxConnectors,
			MaxSignals:           ProMaxBrainSignals,
			ContextHints:         ProBrainContextHints,
			SyncFrequency:        ProBrainSyncFrequency,
			SignalRetentionDays:  ProBrainSignalRetentionDays,
			SemanticSearch:       true,
			ImportanceLearning:   true,
			BrainComposer:        true,
			TrustScore:           true,
			AgentBrainEnabled:    true,
			AgentBrainScope:      "both",
			BrainTriggerDaemon:   true,
			AnomalyAlerts:        true,
			FlyMindBrainContext:  true,
			FlyMindRAGSignals:    true,
			FlyMindMemoryToBrain: true,
			FlyMindEconomicBrain: true,
			AnalyticsModelFeedback: true,
			AnalyticsExport:      false,
			ConnectorGraph:       false,
		}
	case PlanStarter:
		return BrainConnectorLimits{
			MaxConnectors:        StarterMaxConnectors,
			MaxSignals:           StarterMaxBrainSignals,
			ContextHints:         StarterBrainContextHints,
			SyncFrequency:        StarterBrainSyncFrequency,
			SignalRetentionDays:  StarterBrainSignalRetentionDays,
			SemanticSearch:       false,
			ImportanceLearning:   false,
			BrainComposer:        false,
			TrustScore:           false,
			AgentBrainEnabled:    true,
			AgentBrainScope:      "tenant",
			BrainTriggerDaemon:   false,
			AnomalyAlerts:        false,
			FlyMindBrainContext:  true,
			FlyMindRAGSignals:    false,
			FlyMindMemoryToBrain: false,
			FlyMindEconomicBrain: false,
			AnalyticsModelFeedback: false,
			AnalyticsExport:      false,
			ConnectorGraph:       false,
		}
	default: // Free
		return BrainConnectorLimits{
			MaxConnectors:        FreeMaxConnectors,
			MaxSignals:           FreeMaxBrainSignals,
			ContextHints:         FreeBrainContextHints,
			SyncFrequency:        FreeBrainSyncFrequency,
			SignalRetentionDays:  FreeBrainSignalRetentionDays,
			SemanticSearch:       false,
			ImportanceLearning:   false,
			BrainComposer:        false,
			TrustScore:           false,
			AgentBrainEnabled:    false,
			AgentBrainScope:      "tenant",
			BrainTriggerDaemon:   false,
			AnomalyAlerts:        false,
			FlyMindBrainContext:  false,
			FlyMindRAGSignals:    false,
			FlyMindMemoryToBrain: false,
			FlyMindEconomicBrain: false,
			AnalyticsModelFeedback: false,
			AnalyticsExport:      false,
			ConnectorGraph:       false,
		}
	}
}

// IsConnectorAvailableForPlan checks if a connector slug is available for the plan
func IsConnectorAvailableForPlan(slug, plan string) bool {
	connector, ok := ConnectorAvailability[slug]
	if !ok {
		return false
	}
	available, ok := connector[plan]
	if !ok {
		return connector[PlanFree]
	}
	return available
}

// GetMaxConnectors returns the max connectors for a plan
func GetMaxConnectors(plan string) int {
	limits := GetBrainConnectorLimits(plan)
	return limits.MaxConnectors
}

// GetMaxBrainSignals returns the max brain signals for a plan
func GetMaxBrainSignals(plan string) int {
	limits := GetBrainConnectorLimits(plan)
	return limits.MaxSignals
}

// GetBrainContextHints returns the number of context hints for a plan
func GetBrainContextHints(plan string) int {
	limits := GetBrainConnectorLimits(plan)
	return limits.ContextHints
}
