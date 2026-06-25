package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UpdateReputationScore updates a reputation score dynamically.
func (r *Phase4Repository) UpdateReputationScore(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if score, ok := updates["score"]; ok {
		setParts = append(setParts, fmt.Sprintf("score = $%d", argIdx))
		args = append(args, score)
		argIdx++
	}
	if rank, ok := updates["rank"]; ok {
		setParts = append(setParts, fmt.Sprintf("rank = $%d", argIdx))
		args = append(args, rank)
		argIdx++
	}
	if percentile, ok := updates["percentile"]; ok {
		setParts = append(setParts, fmt.Sprintf("percentile = $%d", argIdx))
		args = append(args, percentile)
		argIdx++
	}
	if components, ok := updates["components"]; ok {
		setParts = append(setParts, fmt.Sprintf("components = $%d", argIdx))
		args = append(args, components)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()", fmt.Sprintf("last_calculated = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++
	args = append(args, id)

	query := fmt.Sprintf("UPDATE reputation_scores SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update reputation score: %w", err)
	}
	return nil
}

// UpdateDigitalBadge updates a digital badge dynamically.
func (r *Phase4Repository) UpdateDigitalBadge(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if name, ok := updates["name"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, name)
		argIdx++
	}
	if description, ok := updates["description"]; ok {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, description)
		argIdx++
	}
	if iconURL, ok := updates["icon_url"]; ok {
		setParts = append(setParts, fmt.Sprintf("icon_url = $%d", argIdx))
		args = append(args, iconURL)
		argIdx++
	}
	if category, ok := updates["category"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, category)
		argIdx++
	}
	if criteria, ok := updates["criteria"]; ok {
		setParts = append(setParts, fmt.Sprintf("criteria = $%d", argIdx))
		args = append(args, criteria)
		argIdx++
	}
	if points, ok := updates["points"].(int); ok {
		setParts = append(setParts, fmt.Sprintf("points = $%d", argIdx))
		args = append(args, points)
		argIdx++
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, isActive)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE digital_badges SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update digital badge: %w", err)
	}
	return nil
}

// UpdateLivingMemoryEntry updates a living memory entry dynamically.
func (r *Phase4Repository) UpdateLivingMemoryEntry(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if title, ok := updates["title"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, title)
		argIdx++
	}
	if body, ok := updates["body"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("body = $%d", argIdx))
		args = append(args, body)
		argIdx++
	}
	if memoryType, ok := updates["memory_type"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("memory_type = $%d", argIdx))
		args = append(args, memoryType)
		argIdx++
	}
	if tags, ok := updates["tags"]; ok {
		setParts = append(setParts, fmt.Sprintf("tags = $%d", argIdx))
		args = append(args, tags)
		argIdx++
	}
	if participants, ok := updates["participants"]; ok {
		setParts = append(setParts, fmt.Sprintf("participants = $%d", argIdx))
		args = append(args, participants)
		argIdx++
	}
	if importance, ok := updates["importance"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("importance = $%d", argIdx))
		args = append(args, importance)
		argIdx++
	}
	if searchableText, ok := updates["searchable_text"]; ok {
		setParts = append(setParts, fmt.Sprintf("searchable_text = $%d", argIdx))
		args = append(args, searchableText)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE living_memory SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update living memory entry: %w", err)
	}
	return nil
}

// IncrementLivingMemoryViewCount increments the view count for a living memory entry.
func (r *Phase4Repository) IncrementLivingMemoryViewCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE living_memory SET view_count = view_count + 1 WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to increment living memory view count: %w", err)
	}
	return nil
}
