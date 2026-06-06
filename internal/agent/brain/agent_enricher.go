package brain

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

type AgentBrainEnricher struct {
	repo   *storage.BrainRepository
	scorer *Scorer
}

func NewAgentBrainEnricher(repo *storage.BrainRepository) *AgentBrainEnricher {
	return &AgentBrainEnricher{
		repo:   repo,
		scorer: NewScorer(repo),
	}
}

type AgentExecutionContext struct {
	BrainSignals      []*storage.BrainSignal `json:"brain_signals,omitempty"`
	BrainConnectorIDs []string               `json:"brain_connector_ids,omitempty"`
	BrainMemoryUsed   int                    `json:"brain_memory_used,omitempty"`
	BrainEnabled      bool                   `json:"brain_enabled"`
}

// EnrichExecution injects brain context into agent execution
func (e *AgentBrainEnricher) EnrichExecution(
	ctx context.Context,
	tenantID uuid.UUID,
	agentID uuid.UUID,
	query string,
	planTier string,
) *AgentExecutionContext {
	limits := plans.GetBrainConnectorLimits(planTier)
	if !limits.AgentBrainEnabled {
		return &AgentExecutionContext{BrainEnabled: false}
	}

	// Determine brain scope
	scopeResolver := NewScopeResolver()
	scope, scopeID := scopeResolver.GetBrainScope(ctx, agentID, tenantID)
	_ = scope

	// Get recent signals
	signals, err := e.repo.GetRecentSignals(ctx, scopeID, 7, limits.ContextHints)
	if err != nil || len(signals) == 0 {
		return &AgentExecutionContext{BrainEnabled: true}
	}

	// Score signals
	scored := e.scorer.ScoreSignals(ctx, signals, signal_now())
	if len(scored) > limits.ContextHints {
		scored = scored[:limits.ContextHints]
	}

	result := &AgentExecutionContext{
		BrainEnabled:  true,
		BrainSignals:  make([]*storage.BrainSignal, len(scored)),
		BrainMemoryUsed: len(signals),
	}

	connSet := make(map[string]bool)
	for i, ss := range scored {
		result.BrainSignals[i] = ss.Signal
		if !connSet[ss.Signal.ConnectorSlug] {
			connSet[ss.Signal.ConnectorSlug] = true
			result.BrainConnectorIDs = append(result.BrainConnectorIDs, ss.Signal.ConnectorSlug)
		}
	}

	return result
}

// EnrichedContextMap returns the brain context as a map for injection into request context
func (aec *AgentExecutionContext) EnrichedContextMap() map[string]interface{} {
	if !aec.BrainEnabled || len(aec.BrainSignals) == 0 {
		return nil
	}
	return map[string]interface{}{
		"brain_signals":      aec.BrainSignals,
		"brain_connectors":   aec.BrainConnectorIDs,
		"brain_memory_used":  aec.BrainMemoryUsed,
	}
}

// FormatBrainContextPrompt formats brain signals for injection into LLM system prompts
func FormatBrainContextPrompt(signals []*storage.BrainSignal) string {
	if len(signals) == 0 {
		return ""
	}
	result := "## Brain Context (from connected accounts):\n"
	for _, s := range signals {
		result += fmt.Sprintf("- [%s/%s] %s\n", s.ConnectorSlug, s.SignalType, s.Fact)
	}
	return result
}

func signal_now() time.Time {
	return time.Now().UTC()
}
