package brain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

type TriggerEvaluator struct {
	repo *storage.BrainRepository
}

func NewTriggerEvaluator(repo *storage.BrainRepository) *TriggerEvaluator {
	return &TriggerEvaluator{repo: repo}
}

type TriggerMatch struct {
	Trigger *storage.BrainTrigger
	Signal  *storage.BrainSignal
}

// EvaluateSignal checks all active triggers against a new signal
func (te *TriggerEvaluator) EvaluateSignal(ctx context.Context, signal *storage.BrainSignal) ([]*TriggerMatch, error) {
	triggers, err := te.repo.GetActiveTriggers(ctx)
	if err != nil {
		return nil, fmt.Errorf("get active triggers: %w", err)
	}

	var matches []*TriggerMatch
	for _, trigger := range triggers {
		if te.matchesTrigger(trigger, signal) {
			matches = append(matches, &TriggerMatch{
				Trigger: trigger,
				Signal:  signal,
			})
			te.repo.UpdateTriggerLastFired(ctx, trigger.ID)
		}
	}

	return matches, nil
}

func (te *TriggerEvaluator) matchesTrigger(t *storage.BrainTrigger, s *storage.BrainSignal) bool {
	// Check tenant scope
	if t.TenantID != s.TenantID {
		return false
	}

	// Check importance
	if s.Importance < t.MinImportance {
		return false
	}

	// Check signal types
	if len(t.SignalTypes) > 0 {
		matched := false
		for _, st := range t.SignalTypes {
			if st == s.SignalType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check connector slugs
	if len(t.ConnectorSlugs) > 0 {
		matched := false
		for _, cs := range t.ConnectorSlugs {
			if cs == s.ConnectorSlug {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// EvaluateAnomaly detects anomalous signal patterns (10x normal in 1 hour)
func (te *TriggerEvaluator) EvaluateAnomaly(ctx context.Context, tenantID uuid.UUID, signal *storage.BrainSignal) (bool, error) {
	// Count signals of this type in the last hour
	signals, _, err := te.repo.ListSignals(ctx, storage.SignalListParams{
		TenantID:   tenantID,
		SignalType: signal.SignalType,
		Limit:      100,
	})
	if err != nil {
		return false, err
	}

	oneHourAgo := time.Now().UTC().Add(-1 * time.Hour)
	recentCount := 0
	for _, s := range signals {
		if s.CreatedAt.After(oneHourAgo) {
			recentCount++
		}
	}

	// If more than 10 signals of the same type in 1 hour, it's anomalous
	return recentCount >= 10, nil
}

// ComposerBriefing generates a briefing from a Brain Composer config
func (te *TriggerEvaluator) ComposerBriefing(ctx context.Context, composer *storage.BrainComposer) (string, error) {
	var filters []storage.SignalFilter
	if err := json.Unmarshal(composer.SignalFilters, &filters); err != nil {
		return "", fmt.Errorf("unmarshal signal filters: %w", err)
	}

	var allSignals []*storage.BrainSignal
	for _, f := range filters {
		for _, slug := range f.ConnectorSlugs {
			signals, _, err := te.repo.ListSignals(ctx, storage.SignalListParams{
				TenantID:      composer.TenantID,
				ConnectorSlug: slug,
				Limit:         50,
			})
			if err != nil {
				continue
			}
			for _, s := range signals {
				if f.ImportanceMin > 0 && s.Importance < f.ImportanceMin {
					continue
				}
				if len(f.SignalTypes) > 0 {
					match := false
					for _, st := range f.SignalTypes {
						if st == s.SignalType {
							match = true
							break
						}
					}
					if !match {
						continue
					}
				}
				allSignals = append(allSignals, s)
			}
		}
	}

	// Format briefing
	briefing := fmt.Sprintf("# %s\n\nGenerated at %s\n\n", composer.Name, time.Now().UTC().Format(time.RFC3339))
	for _, s := range allSignals {
		briefing += fmt.Sprintf("- **[%s]** %s\n", s.ConnectorSlug, s.Fact)
	}

	te.repo.UpdateComposerLastRun(ctx, composer.ID)
	return briefing, nil
}
