package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// FeatureMeasureRepository handles platform feature measure persistence.
type FeatureMeasureRepository struct {
	db *sql.DB
}

// NewFeatureMeasureRepository creates a new feature measure repository.
func NewFeatureMeasureRepository(db *sql.DB) *FeatureMeasureRepository {
	return &FeatureMeasureRepository{db: db}
}

// ListFeatureMeasures returns all platform feature measures ordered by category and sort_order.
func (r *FeatureMeasureRepository) ListFeatureMeasures(ctx context.Context) ([]*FeatureMeasure, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, key, name, description, category, icon, enabled, sort_order, created_at, updated_at
		FROM platform_feature_measures
		ORDER BY category, sort_order, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*FeatureMeasure
	for rows.Next() {
		var m FeatureMeasure
		err := rows.Scan(
			&m.ID, &m.Key, &m.Name, &m.Description, &m.Category, &m.Icon,
			&m.Enabled, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}

// UpdateFeatureMeasureEnabled sets the enabled flag for a measure by ID.
func (r *FeatureMeasureRepository) UpdateFeatureMeasureEnabled(ctx context.Context, id uuid.UUID, enabled bool) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE platform_feature_measures SET enabled = $1, updated_at = $2 WHERE id = $3`,
		enabled, time.Now().UTC(), id)
	return err
}
