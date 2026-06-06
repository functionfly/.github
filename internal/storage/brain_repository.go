package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	brainSignalKeyPrefix = "brain:tenant:%s:signals"
	brainIdxTypePrefix   = "brain:tenant:%s:idx:%s"
	brainIdxConnPrefix   = "brain:tenant:%s:connector:%s"
	brainEntitiesPrefix  = "brain:tenant:%s:entities"
	brainAgentKeyPrefix  = "brain:agent:%s:signals"
	defaultSignalTTL     = 30 * 24 * time.Hour // 30 days
)

type BrainRepository struct {
	db  *sql.DB
	rdb *redis.Client
}

func NewBrainRepository(db *sql.DB, rdb *redis.Client) *BrainRepository {
	return &BrainRepository{db: db, rdb: rdb}
}

func (r *BrainRepository) GetDB() *sql.DB {
	return r.db
}

// ---- Redis Operations (Hot Storage) ----

func (r *BrainRepository) redisKey(tenantID uuid.UUID) string {
	return fmt.Sprintf(brainSignalKeyPrefix, tenantID.String())
}

func (r *BrainRepository) redisIdxTypeKey(tenantID uuid.UUID, signalType string) string {
	return fmt.Sprintf(brainIdxTypePrefix, tenantID.String(), signalType)
}

func (r *BrainRepository) redisIdxConnKey(tenantID uuid.UUID, connectorSlug string) string {
	return fmt.Sprintf(brainIdxConnPrefix, tenantID.String(), connectorSlug)
}

func (r *BrainRepository) redisEntitiesKey(tenantID uuid.UUID) string {
	return fmt.Sprintf(brainEntitiesPrefix, tenantID.String())
}

func (r *BrainRepository) redisAgentKey(agentID uuid.UUID) string {
	return fmt.Sprintf(brainAgentKeyPrefix, agentID.String())
}

// SaveSignal writes to both Redis (hot) and PostgreSQL (warm/durable)
func (r *BrainRepository) SaveSignal(ctx context.Context, signal *BrainSignal) (*BrainSignal, error) {
	if signal.ID == uuid.Nil {
		signal.ID = uuid.New()
	}
	now := time.Now().UTC()
	signal.CreatedAt = now
	signal.LastSeenAt = now
	if signal.TTLHours <= 0 {
		signal.TTLHours = 720 // 30 days default
	}
	if signal.Metadata == nil {
		signal.Metadata = json.RawMessage("{}")
	}

	// Write to Redis
	if r.rdb != nil {
		ttl := time.Duration(signal.TTLHours) * time.Hour
		signalJSON, err := json.Marshal(signal)
		if err != nil {
			return nil, fmt.Errorf("marshal signal for redis: %w", err)
		}

		score := float64(signal.CreatedAt.Unix())
		pipe := r.rdb.Pipeline()
		pipe.ZAdd(ctx, r.redisKey(signal.TenantID), redis.Z{Score: score, Member: signalJSON})
		pipe.ZAdd(ctx, r.redisIdxTypeKey(signal.TenantID, signal.SignalType), redis.Z{Score: score, Member: signal.ID.String()})
		pipe.ZAdd(ctx, r.redisIdxConnKey(signal.TenantID, signal.ConnectorSlug), redis.Z{Score: score, Member: signal.ID.String()})
		pipe.HSet(ctx, r.redisEntitiesKey(signal.TenantID), signal.EntityID, signal.LastSeenAt.Unix())
		pipe.Expire(ctx, r.redisKey(signal.TenantID), ttl)
		pipe.Expire(ctx, r.redisIdxTypeKey(signal.TenantID, signal.SignalType), ttl)
		pipe.Expire(ctx, r.redisIdxConnKey(signal.TenantID, signal.ConnectorSlug), ttl)
		pipe.Expire(ctx, r.redisEntitiesKey(signal.TenantID), ttl)
		if _, err := pipe.Exec(ctx); err != nil {
			// Log but don't fail — PostgreSQL write is the durable one
			fmt.Printf("brain: redis write failed for signal %s: %v\n", signal.ID, err)
		}
	}

	// Write to PostgreSQL (TimescaleDB)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO signals (id, tenant_id, connector_slug, signal_type, entity_id, entity_name, fact, importance, source_url, created_at, last_seen_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		signal.ID, signal.TenantID, signal.ConnectorSlug, signal.SignalType,
		signal.EntityID, signal.EntityName, signal.Fact, signal.Importance,
		signal.SourceURL, signal.CreatedAt, signal.LastSeenAt, signal.Metadata)
	if err != nil {
		return nil, fmt.Errorf("save signal to postgres: %w", err)
	}

	return signal, nil
}

// SaveSignalsBatch writes multiple signals efficiently
func (r *BrainRepository) SaveSignalsBatch(ctx context.Context, signals []*BrainSignal) (int, error) {
	if len(signals) == 0 {
		return 0, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO signals (id, tenant_id, connector_slug, signal_type, entity_id, entity_name, fact, importance, source_url, created_at, last_seen_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`)
	if err != nil {
		return 0, fmt.Errorf("prepare batch insert: %w", err)
	}
	defer stmt.Close()

	count := 0
	for _, signal := range signals {
		if signal.ID == uuid.Nil {
			signal.ID = uuid.New()
		}
		now := time.Now().UTC()
		signal.CreatedAt = now
		signal.LastSeenAt = now
		if signal.Metadata == nil {
			signal.Metadata = json.RawMessage("{}")
		}

		if _, err := stmt.ExecContext(ctx,
			signal.ID, signal.TenantID, signal.ConnectorSlug, signal.SignalType,
			signal.EntityID, signal.EntityName, signal.Fact, signal.Importance,
			signal.SourceURL, signal.CreatedAt, signal.LastSeenAt, signal.Metadata,
		); err != nil {
			return count, fmt.Errorf("insert signal %s: %w", signal.ID, err)
		}
		count++

		// Also write to Redis
		if r.rdb != nil {
			signalJSON, _ := json.Marshal(signal)
			score := float64(signal.CreatedAt.Unix())
			r.rdb.ZAdd(ctx, r.redisKey(signal.TenantID), redis.Z{Score: score, Member: signalJSON})
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit batch insert: %w", err)
	}
	return count, nil
}

// ListSignals returns signals for a tenant with optional filters
func (r *BrainRepository) ListSignals(ctx context.Context, params SignalListParams) ([]*BrainSignal, int, error) {
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 500 {
		params.Limit = 500
	}

	// Try Redis first for recent signals
	if r.rdb != nil && params.ConnectorSlug == "" && params.SignalType == "" {
		key := r.redisKey(params.TenantID)
		total, err := r.rdb.ZCard(ctx, key).Result()
		if err == nil && total > 0 {
			start := int64(params.Offset)
			end := start + int64(params.Limit) - 1
			results, err := r.rdb.ZRevRange(ctx, key, start, end).Result()
			if err == nil {
				var signals []*BrainSignal
				for _, raw := range results {
					s := &BrainSignal{}
					if json.Unmarshal([]byte(raw), s) == nil {
						signals = append(signals, s)
					}
				}
				if len(signals) > 0 {
					return signals, int(total), nil
				}
			}
		}
	}

	// Fallback to PostgreSQL
	where := "WHERE tenant_id = $1"
	args := []interface{}{params.TenantID}
	argIdx := 2

	if params.ConnectorSlug != "" {
		where += fmt.Sprintf(" AND connector_slug = $%d", argIdx)
		args = append(args, params.ConnectorSlug)
		argIdx++
	}
	if params.SignalType != "" {
		where += fmt.Sprintf(" AND signal_type = $%d", argIdx)
		args = append(args, params.SignalType)
		argIdx++
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM signals %s", where)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count signals: %w", err)
	}

	order := "created_at DESC"
	if params.SortBy == "importance" {
		order = "importance DESC, created_at DESC"
	}

	query := fmt.Sprintf(`
		SELECT id, tenant_id, connector_slug, signal_type, COALESCE(entity_id,''), COALESCE(entity_name,''),
		       fact, importance, COALESCE(source_url,''), created_at, last_seen_at, COALESCE(metadata,'{}')
		FROM signals %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		where, order, argIdx, argIdx+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list signals: %w", err)
	}
	defer rows.Close()

	var signals []*BrainSignal
	for rows.Next() {
		s := &BrainSignal{}
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.ConnectorSlug, &s.SignalType,
			&s.EntityID, &s.EntityName, &s.Fact, &s.Importance,
			&s.SourceURL, &s.CreatedAt, &s.LastSeenAt, &s.Metadata,
		); err != nil {
			return nil, 0, fmt.Errorf("scan signal: %w", err)
		}
		signals = append(signals, s)
	}
	return signals, total, nil
}

// GetSignal returns a single signal
func (r *BrainRepository) GetSignal(ctx context.Context, tenantID, signalID uuid.UUID) (*BrainSignal, error) {
	s := &BrainSignal{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, connector_slug, signal_type, COALESCE(entity_id,''), COALESCE(entity_name,''),
		       fact, importance, COALESCE(source_url,''), created_at, last_seen_at, COALESCE(metadata,'{}')
		FROM signals WHERE tenant_id = $1 AND id = $2`, tenantID, signalID).Scan(
		&s.ID, &s.TenantID, &s.ConnectorSlug, &s.SignalType,
		&s.EntityID, &s.EntityName, &s.Fact, &s.Importance,
		&s.SourceURL, &s.CreatedAt, &s.LastSeenAt, &s.Metadata)
	if err != nil {
		return nil, fmt.Errorf("get signal: %w", err)
	}
	return s, nil
}

// DeleteSignal removes a signal from both Redis and PostgreSQL
func (r *BrainRepository) DeleteSignal(ctx context.Context, tenantID, signalID uuid.UUID) error {
	// Remove from PostgreSQL
	result, err := r.db.ExecContext(ctx, `DELETE FROM signals WHERE tenant_id = $1 AND id = $2`, tenantID, signalID)
	if err != nil {
		return fmt.Errorf("delete signal: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("signal not found")
	}

	// Remove from Redis (best-effort)
	if r.rdb != nil {
		members, _ := r.rdb.ZRange(ctx, r.redisKey(tenantID), 0, -1).Result()
		for _, m := range members {
			s := &BrainSignal{}
			if json.Unmarshal([]byte(m), s) == nil && s.ID == signalID {
				r.rdb.ZRem(ctx, r.redisKey(tenantID), m)
				break
			}
		}
	}

	return nil
}

// PurgeSignals removes all signals for a tenant
func (r *BrainRepository) PurgeSignals(ctx context.Context, tenantID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM signals WHERE tenant_id = $1`, tenantID)
	if err != nil {
		return fmt.Errorf("purge signals: %w", err)
	}

	// Clear Redis keys
	if r.rdb != nil {
		r.rdb.Del(ctx, r.redisKey(tenantID))
		r.rdb.Del(ctx, r.redisEntitiesKey(tenantID))
	}

	return nil
}

// GetBrainStats returns brain usage statistics
func (r *BrainRepository) GetBrainStats(ctx context.Context, tenantID uuid.UUID) (*BrainStats, error) {
	stats := &BrainStats{
		SignalsByType:      make(map[string]int),
		SignalsByConnector: make(map[string]int),
	}

	// Total signals
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM signals WHERE tenant_id = $1`, tenantID).Scan(&stats.TotalSignals)

	// Signals by type
	rows, err := r.db.QueryContext(ctx, `
		SELECT signal_type, COUNT(*) FROM signals WHERE tenant_id = $1 GROUP BY signal_type`, tenantID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var st string
			var count int
			rows.Scan(&st, &count)
			stats.SignalsByType[st] = count
		}
	}

	// Signals by connector
	rows2, err := r.db.QueryContext(ctx, `
		SELECT connector_slug, COUNT(*) FROM signals WHERE tenant_id = $1 GROUP BY connector_slug`, tenantID)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var cs string
			var count int
			rows2.Scan(&cs, &count)
			stats.SignalsByConnector[cs] = count
		}
	}

	// Oldest and newest
	r.db.QueryRowContext(ctx, `
		SELECT MIN(created_at), MAX(created_at) FROM signals WHERE tenant_id = $1`, tenantID).Scan(
		&stats.OldestSignal, &stats.NewestSignal)

	// Memory used from Redis
	if r.rdb != nil {
		card, _ := r.rdb.ZCard(ctx, r.redisKey(tenantID)).Result()
		stats.MemoryUsed = int(card)
	}

	return stats, nil
}

// SemanticSearch performs pgvector similarity search (Pro+)
func (r *BrainRepository) SemanticSearch(ctx context.Context, tenantID uuid.UUID, embedding []float32, limit int) ([]*BrainSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	embeddingStr := vectorToString(embedding)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, connector_slug, signal_type, COALESCE(entity_id,''), COALESCE(entity_name,''),
		       fact, importance, COALESCE(source_url,''), created_at, last_seen_at, COALESCE(metadata,'{}'),
		       1 - (embedding <=> $2::vector) as similarity
		FROM signals
		WHERE tenant_id = $1 AND embedding IS NOT NULL
		ORDER BY embedding <=> $2::vector
		LIMIT $3`, tenantID, embeddingStr, limit)
	if err != nil {
		return nil, fmt.Errorf("semantic search: %w", err)
	}
	defer rows.Close()

	var results []*BrainSearchResult
	for rows.Next() {
		s := &BrainSignal{}
		r := &BrainSearchResult{Signal: s}
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.ConnectorSlug, &s.SignalType,
			&s.EntityID, &s.EntityName, &s.Fact, &s.Importance,
			&s.SourceURL, &s.CreatedAt, &s.LastSeenAt, &s.Metadata,
			&r.Distance,
		); err != nil {
			return nil, fmt.Errorf("scan semantic result: %w", err)
		}
		r.Score = 1.0 - r.Distance
		results = append(results, r)
	}
	return results, nil
}

// UpdateSignalEmbedding stores the pgvector embedding for a signal
func (r *BrainRepository) UpdateSignalEmbedding(ctx context.Context, signalID uuid.UUID, embedding []float32) error {
	embeddingStr := vectorToString(embedding)
	_, err := r.db.ExecContext(ctx, `UPDATE signals SET embedding = $1::vector WHERE id = $2`, embeddingStr, signalID)
	return err
}

// GetRecentSignals returns signals from the last N days for context assembly
func (r *BrainRepository) GetRecentSignals(ctx context.Context, tenantID uuid.UUID, days int, limit int) ([]*BrainSignal, error) {
	if days <= 0 {
		days = 7
	}
	if limit <= 0 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, connector_slug, signal_type, COALESCE(entity_id,''), COALESCE(entity_name,''),
		       fact, importance, COALESCE(source_url,''), created_at, last_seen_at, COALESCE(metadata,'{}')
		FROM signals
		WHERE tenant_id = $1 AND created_at > NOW() - ($2 || ' days')::interval
		ORDER BY importance DESC, created_at DESC
		LIMIT $3`, tenantID, days, limit)
	if err != nil {
		return nil, fmt.Errorf("get recent signals: %w", err)
	}
	defer rows.Close()

	var signals []*BrainSignal
	for rows.Next() {
		s := &BrainSignal{}
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.ConnectorSlug, &s.SignalType,
			&s.EntityID, &s.EntityName, &s.Fact, &s.Importance,
			&s.SourceURL, &s.CreatedAt, &s.LastSeenAt, &s.Metadata,
		); err != nil {
			return nil, fmt.Errorf("scan recent signal: %w", err)
		}
		signals = append(signals, s)
	}
	return signals, nil
}

func vectorToString(v []float32) string {
	if len(v) == 0 {
		return "[]"
	}
	s := "["
	for i, f := range v {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf("%f", f)
	}
	s += "]"
	return s
}

// ---- Brain Composer Operations ----

func (r *BrainRepository) CreateComposer(ctx context.Context, c *BrainComposer) (*BrainComposer, error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO brain_composers (id, tenant_id, name, schedule, signal_filters, output_format, actions, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		c.ID, c.TenantID, c.Name, c.Schedule, c.SignalFilters, c.OutputFormat, c.Actions, c.IsActive, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create composer: %w", err)
	}
	return c, nil
}

func (r *BrainRepository) ListComposers(ctx context.Context, tenantID uuid.UUID) ([]*BrainComposer, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, name, schedule, signal_filters, output_format, actions, is_active, last_run_at, created_at, updated_at
		FROM brain_composers WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list composers: %w", err)
	}
	defer rows.Close()

	var composers []*BrainComposer
	for rows.Next() {
		c := &BrainComposer{}
		var lastRun sql.NullTime
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Name, &c.Schedule, &c.SignalFilters, &c.OutputFormat, &c.Actions, &c.IsActive, &lastRun, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan composer: %w", err)
		}
		if lastRun.Valid {
			c.LastRunAt = &lastRun.Time
		}
		composers = append(composers, c)
	}
	return composers, nil
}

func (r *BrainRepository) DeleteComposer(ctx context.Context, tenantID, composerID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM brain_composers WHERE tenant_id = $1 AND id = $2`, tenantID, composerID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("composer not found")
	}
	return nil
}

func (r *BrainRepository) UpdateComposerLastRun(ctx context.Context, composerID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE brain_composers SET last_run_at = NOW(), updated_at = NOW() WHERE id = $1`, composerID)
	return err
}

// ---- Brain Feedback Operations ----

func (r *BrainRepository) SaveFeedback(ctx context.Context, tenantID uuid.UUID, fb *BrainFeedbackRequest) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO brain_feedback (id, tenant_id, signal_id, helpful, context, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())`,
		uuid.New(), tenantID, fb.SignalID, fb.Helpful, fb.Context)
	return err
}

// ---- Brain Trigger Operations ----

func (r *BrainRepository) CreateTrigger(ctx context.Context, t *BrainTrigger) (*BrainTrigger, error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO brain_triggers (id, tenant_id, agent_id, name, signal_types, connector_slugs, min_importance, schedule, action, action_config, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		t.ID, t.TenantID, t.AgentID, t.Name, t.SignalTypes, t.ConnectorSlugs,
		t.MinImportance, t.Schedule, t.Action, t.ActionConfig, t.IsActive, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create trigger: %w", err)
	}
	return t, nil
}

func (r *BrainRepository) ListTriggers(ctx context.Context, tenantID uuid.UUID) ([]*BrainTrigger, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, agent_id, name, signal_types, connector_slugs, min_importance, schedule, action, action_config, is_active, last_fired_at, created_at, updated_at
		FROM brain_triggers WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list triggers: %w", err)
	}
	defer rows.Close()

	var triggers []*BrainTrigger
	for rows.Next() {
		t := &BrainTrigger{}
		var lastFired sql.NullTime
		if err := rows.Scan(&t.ID, &t.TenantID, &t.AgentID, &t.Name, &t.SignalTypes, &t.ConnectorSlugs,
			&t.MinImportance, &t.Schedule, &t.Action, &t.ActionConfig, &t.IsActive, &lastFired, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan trigger: %w", err)
		}
		if lastFired.Valid {
			t.LastFiredAt = &lastFired.Time
		}
		triggers = append(triggers, t)
	}
	return triggers, nil
}

func (r *BrainRepository) GetActiveTriggers(ctx context.Context) ([]*BrainTrigger, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, agent_id, name, signal_types, connector_slugs, min_importance, schedule, action, action_config, is_active, last_fired_at, created_at, updated_at
		FROM brain_triggers WHERE is_active = true ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("get active triggers: %w", err)
	}
	defer rows.Close()

	var triggers []*BrainTrigger
	for rows.Next() {
		t := &BrainTrigger{}
		var lastFired sql.NullTime
		if err := rows.Scan(&t.ID, &t.TenantID, &t.AgentID, &t.Name, &t.SignalTypes, &t.ConnectorSlugs,
			&t.MinImportance, &t.Schedule, &t.Action, &t.ActionConfig, &t.IsActive, &lastFired, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan trigger: %w", err)
		}
		if lastFired.Valid {
			t.LastFiredAt = &lastFired.Time
		}
		triggers = append(triggers, t)
	}
	return triggers, nil
}

func (r *BrainRepository) UpdateTriggerLastFired(ctx context.Context, triggerID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE brain_triggers SET last_fired_at = NOW(), updated_at = NOW() WHERE id = $1`, triggerID)
	return err
}

func (r *BrainRepository) DeleteTrigger(ctx context.Context, tenantID, triggerID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM brain_triggers WHERE tenant_id = $1 AND id = $2`, tenantID, triggerID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("trigger not found")
	}
	return nil
}
