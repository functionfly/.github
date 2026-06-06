package brain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

type ContextBuilder struct {
	repo   *storage.BrainRepository
	scorer *Scorer
	memory *Memory
}

func NewContextBuilder(repo *storage.BrainRepository) *ContextBuilder {
	return &ContextBuilder{
		repo:   repo,
		scorer: NewScorer(repo),
		memory: NewMemory(repo),
	}
}

type BrainContext struct {
	Signals       []*storage.BrainSignal `json:"signals"`
	ConnectorIDs  []string               `json:"connector_ids"`
	MemoryUsed    int                    `json:"memory_used"`
	ScoreSummary  string                 `json:"score_summary"`
}

// GetContextForQuery assembles the best brain signals for an agent or chat query
func (cb *ContextBuilder) GetContextForQuery(ctx context.Context, tenantID uuid.UUID, query string, intent string, maxHints int) (*BrainContext, error) {
	if maxHints <= 0 {
		maxHints = plans.FreeBrainContextHints
	}

	// Get recent signals (last 7 days default)
	signals, err := cb.repo.GetRecentSignals(ctx, tenantID, 7, maxHints*3)
	if err != nil {
		return nil, fmt.Errorf("get recent signals: %w", err)
	}

	if len(signals) == 0 {
		return &BrainContext{
			Signals:      []*storage.BrainSignal{},
			ConnectorIDs: []string{},
			MemoryUsed:   0,
		}, nil
	}

	// Score signals
	queryTime := time.Now().UTC()
	scored := cb.scorer.ScoreSignals(ctx, signals, queryTime)

	// Take top N
	if len(scored) > maxHints {
		scored = scored[:maxHints]
	}

	// Build context
	result := &BrainContext{
		Signals:      make([]*storage.BrainSignal, len(scored)),
		ConnectorIDs: []string{},
		MemoryUsed:   len(signals),
	}

	connectorSet := make(map[string]bool)
	for i, ss := range scored {
		result.Signals[i] = ss.Signal
		if !connectorSet[ss.Signal.ConnectorSlug] {
			connectorSet[ss.Signal.ConnectorSlug] = true
			result.ConnectorIDs = append(result.ConnectorIDs, ss.Signal.ConnectorSlug)
		}
	}

	// Build summary
	var facts []string
	for _, s := range result.Signals {
		facts = append(facts, fmt.Sprintf("- [%s] %s", s.ConnectorSlug, s.Fact))
	}
	result.ScoreSummary = strings.Join(facts, "\n")

	return result, nil
}

// FormatContextForPrompt formats brain context for injection into LLM prompts
func (cb *ContextBuilder) FormatContextForPrompt(bc *BrainContext) string {
	if bc == nil || len(bc.Signals) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("## Context from connected accounts:\n\n")
	for _, s := range bc.Signals {
		b.WriteString(fmt.Sprintf("- [%s/%s] %s", s.ConnectorSlug, s.SignalType, s.Fact))
		if s.SourceURL != "" {
			b.WriteString(fmt.Sprintf(" (%s)", s.SourceURL))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// EnrichAgentContext provides brain context for agent execution
func (cb *ContextBuilder) EnrichAgentContext(ctx context.Context, tenantID uuid.UUID, query string, planTier string) map[string]interface{} {
	limits := plans.GetBrainConnectorLimits(planTier)
	if !limits.AgentBrainEnabled {
		return nil
	}

	bc, err := cb.GetContextForQuery(ctx, tenantID, query, "general", limits.ContextHints)
	if err != nil || bc == nil {
		return nil
	}

	return map[string]interface{}{
		"brain_signals":     bc.Signals,
		"brain_connectors":  bc.ConnectorIDs,
		"brain_memory_used": bc.MemoryUsed,
	}
}
