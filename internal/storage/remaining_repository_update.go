package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// UpdateFeedbackRound updates a feedback round.
func (r *RemainingRepository) UpdateFeedbackRound(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = "NOW()"

	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates))
	argIdx := 1

	for k, v := range updates {
		if v == "NOW()" {
			setClauses = append(setClauses, fmt.Sprintf("%s = NOW()", k))
		} else {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", k, argIdx))
			args = append(args, v)
			argIdx++
		}
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE feedback_rounds SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update feedback round: %w", err)
	}
	return nil
}

// UpdateFeedbackRoundAssignment updates a feedback round assignment.
func (r *RemainingRepository) UpdateFeedbackRoundAssignment(ctx context.Context, id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates))
	argIdx := 1

	for k, v := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", k, argIdx))
		args = append(args, v)
		argIdx++
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE feedback_round_assignments SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update feedback round assignment: %w", err)
	}
	return nil
}

// UpdateGoalCascade updates goal cascade fields.
func (r *RemainingRepository) UpdateGoalCascade(ctx context.Context, id uuid.UUID, parentGoalID *uuid.UUID, goalLevel, cascadeVisibility string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE performance_goals SET parent_goal_id = $1, goal_level = $2, cascade_visibility = $3, updated_at = NOW()
		WHERE id = $4`,
		parentGoalID, goalLevel, cascadeVisibility, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update goal cascade: %w", err)
	}
	return nil
}

// UpdateDocumentSignature updates a document signature.
func (r *RemainingRepository) UpdateDocumentSignature(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates))
	argIdx := 1

	for k, v := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", k, argIdx))
		args = append(args, v)
		argIdx++
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE document_signatures SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update document signature: %w", err)
	}
	return nil
}

// UpdateOrgChartImport updates an org chart import.
func (r *RemainingRepository) UpdateOrgChartImport(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates))
	argIdx := 1

	for k, v := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", k, argIdx))
		args = append(args, v)
		argIdx++
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE org_chart_imports SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update org chart import: %w", err)
	}
	return nil
}

// UpdatePackage updates a package in the registry.
func (r *RemainingRepository) UpdatePackage(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = "NOW()"

	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates))
	argIdx := 1

	for k, v := range updates {
		if v == "NOW()" {
			setClauses = append(setClauses, fmt.Sprintf("%s = NOW()", k))
		} else {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", k, argIdx))
			args = append(args, v)
			argIdx++
		}
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE package_registry SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update package: %w", err)
	}
	return nil
}
