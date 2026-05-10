package timemachine

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	tmstorage "github.com/functionfly/functionfly/internal/storage/timemachine"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ReconciliationEngine struct {
	tmRepo  *tmstorage.Repository
	regRepo *registry.RegistryRepository
}

func NewReconciliationEngine(tmRepo *tmstorage.Repository) *ReconciliationEngine {
	return &ReconciliationEngine{tmRepo: tmRepo}
}

func (e *ReconciliationEngine) SetRegistryRepository(regRepo *registry.RegistryRepository) {
	e.regRepo = regRepo
}

type ReconciliationAction struct {
	Type           string          `json:"type"`
	TargetResource string          `json:"target_resource"`
	Description    string          `json:"description"`
	OldValue       json.RawMessage `json:"old_value,omitempty"`
	NewValue       json.RawMessage `json:"new_value,omitempty"`
	RiskLevel      string          `json:"risk_level"`
	Reversible     bool            `json:"reversible"`
}

type ReconciliationPlan struct {
	TotalItems    int                   `json:"total_items"`
	ChangedItems  int                   `json:"changed_items"`
	Actions       []ReconciliationAction `json:"actions"`
	RiskSummary   map[string]int        `json:"risk_summary"`
	DryRun        bool                  `json:"dry_run"`
}

func (e *ReconciliationEngine) GeneratePlan(replayID uuid.UUID, dryRun bool) (*ReconciliationPlan, error) {
	changedItems, _, err := e.tmRepo.ListChangedItems(replayID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("list changed items: %w", err)
	}

	allItems, _, err := e.tmRepo.ListReplayItems(replayID, 10000, 0, "")
	if err != nil {
		return nil, fmt.Errorf("list all items: %w", err)
	}

	plan := &ReconciliationPlan{
		TotalItems:   len(allItems),
		ChangedItems: len(changedItems),
		DryRun:       dryRun,
		RiskSummary:  map[string]int{"low": 0, "medium": 0, "high": 0},
	}

	actions := make([]ReconciliationAction, 0, len(changedItems))
	reconciliations := make([]tmstorage.Reconciliation, 0, len(changedItems))

	for _, item := range changedItems {
		diffType := ""
		if item.DiffType.Valid {
			diffType = item.DiffType.String
		}

		action := e.buildAction(item, diffType)
		actions = append(actions, action)
		plan.RiskSummary[action.RiskLevel]++

		rec := tmstorage.Reconciliation{
			ID:             uuid.New(),
			ReplayID:       replayID,
			ReplayItemID:   item.ID,
			ActionType:     action.Type,
			TargetResource: action.TargetResource,
			OldValue:       action.OldValue,
			NewValue:       action.NewValue,
			DryRun:         dryRun,
			Reversible:     action.Reversible,
			Status:         "pending",
		}

		if dryRun {
			rec.Status = "dry_run"
		}

		reconciliations = append(reconciliations, rec)
	}

	plan.Actions = actions

	if len(reconciliations) > 0 {
		if err := e.tmRepo.CreateReconciliations(reconciliations); err != nil {
			return nil, fmt.Errorf("create reconciliations: %w", err)
		}

		if !dryRun {
			e.applyReconciliations(reconciliations)
		}
	}

	return plan, nil
}

func (e *ReconciliationEngine) buildAction(item tmstorage.ReplayItem, diffType string) ReconciliationAction {
	action := ReconciliationAction{
		TargetResource: fmt.Sprintf("execution:%s", item.OriginalExecutionID),
		OldValue:       item.OriginalOutput,
		NewValue:       item.NewOutput,
		Reversible:     true,
	}

	switch ClassifyDiffTypeFromDB(diffType) {
	case DiffTypeMinor:
		action.Type = "update_output"
		action.Description = "Minor output change detected - safe to reconcile"
		action.RiskLevel = "low"
	case DiffTypeMajor:
		action.Type = "update_output"
		action.Description = "Major output change detected - review before reconciling"
		action.RiskLevel = "medium"
	case DiffTypeBreaking:
		action.Type = "update_output_with_review"
		action.Description = "Breaking change detected - manual review required"
		action.RiskLevel = "high"
	case DiffTypeError:
		action.Type = "flag_error"
		action.Description = "Execution error during replay - investigate before reconciling"
		action.RiskLevel = "high"
		action.Reversible = false
	default:
		action.Type = "no_action"
		action.Description = "No changes detected"
		action.RiskLevel = "low"
	}

	return action
}

func (e *ReconciliationEngine) applyReconciliations(recs []tmstorage.Reconciliation) {
	for i := range recs {
		rec := &recs[i]
		if rec.DryRun {
			continue
		}

		if err := e.applySingle(rec); err != nil {
			logrus.WithError(err).WithField("reconciliation_id", rec.ID).Error("Failed to apply reconciliation")
			_ = e.tmRepo.UpdateReconciliationStatus(rec.ID, "failed", err.Error())
			continue
		}

		if err := e.tmRepo.UpdateReconciliationStatus(rec.ID, "applied", ""); err != nil {
			logrus.WithError(err).WithField("reconciliation_id", rec.ID).Error("Failed to mark reconciliation as applied")
		}

		logrus.WithFields(logrus.Fields{
			"reconciliation_id": rec.ID,
			"action_type":       rec.ActionType,
			"target":            rec.TargetResource,
		}).Info("Reconciliation action applied")
	}
}

func (e *ReconciliationEngine) applySingle(rec *tmstorage.Reconciliation) error {
	switch rec.ActionType {
	case "update_output", "update_output_with_review":
		return e.applyOutputUpdate(rec)
	case "flag_error":
		logrus.WithFields(logrus.Fields{
			"replay_item_id": rec.ReplayItemID,
			"target":         rec.TargetResource,
		}).Warn("Reconciliation flagged error for manual review")
		return nil
	case "no_action":
		return nil
	default:
		return fmt.Errorf("unknown action type: %s", rec.ActionType)
	}
}

func (e *ReconciliationEngine) applyOutputUpdate(rec *tmstorage.Reconciliation) error {
	item, err := e.tmRepo.GetReplayItem(rec.ReplayItemID)
	if err != nil {
		return fmt.Errorf("get replay item: %w", err)
	}
	if item == nil {
		return fmt.Errorf("replay item %s not found", rec.ReplayItemID)
	}

	item.ReconciliationStatus = "reconciled"
	now := time.Now()
	item.ReconciledAt = &now
	item.ReconciliationActions = rec.OldValue

	if err := e.tmRepo.UpdateReplayItem(item); err != nil {
		return fmt.Errorf("update replay item: %w", err)
	}

	if e.regRepo != nil && item.NewOutput != nil && len(item.NewOutput) > 0 {
		if updateErr := e.regRepo.UpdateExecutionPublicOutput(item.OriginalExecutionID, item.NewOutput); updateErr != nil {
			logrus.WithError(updateErr).WithField("execution_id", item.OriginalExecutionID).Warn("Failed to update registry execution output — replay item still marked reconciled")
		}
	}

	return nil
}

func (e *ReconciliationEngine) GetPlan(replayID uuid.UUID) (*ReconciliationPlan, error) {
	recs, total, err := e.tmRepo.ListReconciliations(replayID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("list reconciliations: %w", err)
	}

	plan := &ReconciliationPlan{
		TotalItems:  int(total),
		RiskSummary: map[string]int{"low": 0, "medium": 0, "high": 0},
	}

	actions := make([]ReconciliationAction, 0, len(recs))
	isDryRun := true

	for _, rec := range recs {
		if !rec.DryRun {
			isDryRun = false
		}

		action := ReconciliationAction{
			Type:           rec.ActionType,
			TargetResource: rec.TargetResource,
			OldValue:       rec.OldValue,
			NewValue:       rec.NewValue,
			Reversible:     rec.Reversible,
		}

		switch rec.ActionType {
		case "update_output":
			action.RiskLevel = "low"
			action.Description = "Output update"
		case "update_output_with_review":
			action.RiskLevel = "high"
			action.Description = "Output update requiring review"
		case "flag_error":
			action.RiskLevel = "high"
			action.Description = "Error flagged for investigation"
		default:
			action.RiskLevel = "low"
		}

		plan.RiskSummary[action.RiskLevel]++
		actions = append(actions, action)
	}

	plan.Actions = actions
	plan.ChangedItems = len(actions)
	plan.DryRun = isDryRun

	return plan, nil
}
