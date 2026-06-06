package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AnalyticsEventRepository struct {
	db *sql.DB
}

func NewAnalyticsEventRepository(db *sql.DB) *AnalyticsEventRepository {
	return &AnalyticsEventRepository{db: db}
}

func (r *AnalyticsEventRepository) SaveEvent(ctx context.Context, event *AnalyticsEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Metadata == nil {
		event.Metadata = json.RawMessage("{}")
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO analytics_events (id, event_type, tenant_tier, connector_slug, signal_type, importance, signals_count, fact_length, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		event.ID, event.EventType, event.TenantTier, event.ConnectorSlug, event.SignalType,
		event.Importance, event.SignalsCount, event.FactLength, event.Metadata, event.CreatedAt)
	return err
}

func (r *AnalyticsEventRepository) SaveEventsBatch(ctx context.Context, events []*AnalyticsEvent) (int, error) {
	if len(events) == 0 {
		return 0, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO analytics_events (id, event_type, tenant_tier, connector_slug, signal_type, importance, signals_count, fact_length, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`)
	if err != nil {
		return 0, fmt.Errorf("prepare batch: %w", err)
	}
	defer stmt.Close()

	count := 0
	for _, event := range events {
		if event.ID == uuid.Nil {
			event.ID = uuid.New()
		}
		if event.Metadata == nil {
			event.Metadata = json.RawMessage("{}")
		}
		if _, err := stmt.ExecContext(ctx,
			event.ID, event.EventType, event.TenantTier, event.ConnectorSlug,
			event.SignalType, event.Importance, event.SignalsCount, event.FactLength,
			event.Metadata, event.CreatedAt,
		); err != nil {
			return count, err
		}
		count++
	}

	return count, tx.Commit()
}

func (r *AnalyticsEventRepository) QueryEvents(ctx context.Context, eventType string, since time.Time, limit int) ([]*AnalyticsEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, event_type, tenant_tier, COALESCE(connector_slug,''), COALESCE(signal_type,''),
		       COALESCE(importance,0), COALESCE(signals_count,0), COALESCE(fact_length,0),
		       COALESCE(metadata,'{}'), created_at
		FROM analytics_events
		WHERE event_type = $1 AND created_at > $2
		ORDER BY created_at DESC LIMIT $3`, eventType, since, limit)
	if err != nil {
		return nil, fmt.Errorf("query analytics events: %w", err)
	}
	defer rows.Close()

	var events []*AnalyticsEvent
	for rows.Next() {
		e := &AnalyticsEvent{}
		if err := rows.Scan(&e.ID, &e.EventType, &e.TenantTier, &e.ConnectorSlug, &e.SignalType,
			&e.Importance, &e.SignalsCount, &e.FactLength, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan analytics event: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *AnalyticsEventRepository) GetFeedbackEvents(ctx context.Context, helpful bool, days int) ([]*AnalyticsEvent, error) {
	if days <= 0 {
		days = 7
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT ae.id, ae.event_type, ae.tenant_tier, COALESCE(ae.connector_slug,''), COALESCE(ae.signal_type,''),
		       COALESCE(ae.importance,0), COALESCE(ae.signals_count,0), COALESCE(ae.fact_length,0),
		       COALESCE(ae.metadata,'{}'), ae.created_at
		FROM analytics_events ae
		WHERE ae.event_type = 'brain.feedback'
		  AND ae.created_at > NOW() - ($1 || ' days')::interval
		  AND (ae.metadata->>'helpful')::boolean = $2
		ORDER BY ae.created_at DESC`, days, helpful)
	if err != nil {
		return nil, fmt.Errorf("query feedback events: %w", err)
	}
	defer rows.Close()

	var events []*AnalyticsEvent
	for rows.Next() {
		e := &AnalyticsEvent{}
		if err := rows.Scan(&e.ID, &e.EventType, &e.TenantTier, &e.ConnectorSlug, &e.SignalType,
			&e.Importance, &e.SignalsCount, &e.FactLength, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan feedback event: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}
