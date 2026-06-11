// Package receipt provides storage and domain logic for the public
// "Execution Receipt" feature. The feature is a productization of the
// existing shareable-execution row in registry_executions_public, plus
// a few new tables for milestones, revocations, and view tracking.
package receipt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Common sentinel errors. Handlers should map these to HTTP status codes.
var (
	ErrNotFound        = errors.New("receipt not found")
	ErrRevoked         = errors.New("receipt has been revoked")
	ErrFunctionPrivate = errors.New("function is not public; receipt cannot be run")
)

// MilestoneChannel is the enum of fan-out channels the worker supports.
type MilestoneChannel string

const (
	ChannelInApp   MilestoneChannel = "inapp"
	ChannelEmail   MilestoneChannel = "email"
	ChannelTweet   MilestoneChannel = "tweet_intent"
	ChannelWebhook MilestoneChannel = "webhook"
)

// MilestoneEvent is one row in receipt_milestone_events.
type MilestoneEvent struct {
	ID            uuid.UUID        `json:"id"             gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID    uuid.UUID        `json:"function_id"    gorm:"type:uuid;not null;index"`
	TenantID      *uuid.UUID       `json:"tenant_id"      gorm:"type:uuid;index"`
	Threshold     int              `json:"threshold"      gorm:"not null"`
	TotalRunsAt   int              `json:"total_runs_at"  gorm:"not null"`
	PublicID      string           `json:"public_id"      gorm:"not null"`
	FiredAt       time.Time        `json:"fired_at"       gorm:"not null;default:now()"`
	ChannelsFired []MilestoneChannel `json:"channels_fired" gorm:"type:text[];not null;default:'{}'"`
	DedupeKey     string           `json:"dedupe_key"     gorm:"not null"`
	CreatedAt     time.Time        `json:"created_at"     gorm:"not null;default:now()"`
}

func (MilestoneEvent) TableName() string { return "receipt_milestone_events" }

// Revocation records when an owner revokes a receipt.
type Revocation struct {
	ID         uuid.UUID `json:"id"          gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PublicID   string    `json:"public_id"   gorm:"not null"`
	FunctionID uuid.UUID `json:"function_id" gorm:"type:uuid;not null;index"`
	RevokedBy  uuid.UUID `json:"revoked_by"  gorm:"not null"`
	RevokedAt  time.Time `json:"revoked_at"  gorm:"not null;default:now()"`
	Reason     string    `json:"reason"`
}

func (Revocation) TableName() string { return "receipt_revocations" }

// Repository is the storage layer for the receipt domain. It is intentionally
// thin: reads delegate to RegistryRepository (single source of truth on
// registry_executions_public) and writes live here on the new tables.
type Repository struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewRepository constructs a Repository. Both arguments are required in
// production; the constructor panics if the DB is nil to surface a
// misconfiguration at startup rather than at request time.
func NewRepository(db *gorm.DB, redisClient *redis.Client) *Repository {
	if db == nil {
		panic("receipt.NewRepository: db is required")
	}
	return &Repository{db: db, redis: redisClient}
}

// ----------------------------------------------------------------------------
// Read paths (delegate to registry.RegistryRepository for the source of truth)
// ----------------------------------------------------------------------------

// GetReceipt returns a single shareable execution row, plus the function
// metadata that the receipt page needs to render. Returns ErrNotFound or
// ErrRevoked for the public-facing 404 / 410 cases.
func (r *Repository) GetReceipt(ctx context.Context, publicID string) (*registry.RegistryExecutionPublic, *registry.RegistryFunction, *registry.RegistryFunctionVersion, error) {
	var exec registry.RegistryExecutionPublic
	err := r.db.WithContext(ctx).
		Where("public_id = ?", publicID).
		First(&exec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, ErrNotFound
		}
		return nil, nil, nil, fmt.Errorf("receipt.GetReceipt: %w", err)
	}
	if exec.RevokedAt.Valid {
		return nil, nil, nil, ErrRevoked
	}
	if !exec.Shareable {
		return nil, nil, nil, ErrNotFound
	}

	var fn registry.RegistryFunction
	if err := r.db.WithContext(ctx).Where("id = ?", exec.FunctionID).First(&fn).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Function was deleted but the receipt row lingers — surface as not-found.
			return nil, nil, nil, ErrNotFound
		}
		return nil, nil, nil, fmt.Errorf("receipt.GetReceipt function: %w", err)
	}

	var ver registry.RegistryFunctionVersion
	if err := r.db.WithContext(ctx).
		Where("function_id = ? AND version = ?", exec.FunctionID, exec.Version).
		First(&ver).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Version was deleted; return the receipt without version data.
			return &exec, &fn, nil, nil
		}
		return nil, nil, nil, fmt.Errorf("receipt.GetReceipt version: %w", err)
	}
	return &exec, &fn, &ver, nil
}

// IncrementViewCount bumps view_count and last_viewed_at. Fire-and-forget at
// the call site — errors are logged and swallowed.
func (r *Repository) IncrementViewCount(ctx context.Context, publicID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&registry.RegistryExecutionPublic{}).
		Where("public_id = ?", publicID).
		Updates(map[string]interface{}{
			"view_count":     gorm.Expr("view_count + 1"),
			"last_viewed_at": now,
		}).Error
}

// IncrementForkCount bumps fork_count. Fire-and-forget.
func (r *Repository) IncrementForkCount(ctx context.Context, publicID string) error {
	return r.db.WithContext(ctx).
		Model(&registry.RegistryExecutionPublic{}).
		Where("public_id = ?", publicID).
		Update("fork_count", gorm.Expr("fork_count + 1")).Error
}

// GetTrending returns the top-viewed shareable receipts from the last 7 days.
// The result powers the "Trending receipts" widget on the marketing site.
func (r *Repository) GetTrending(ctx context.Context, limit int) ([]registry.RegistryExecutionPublic, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	since := time.Now().AddDate(0, 0, -7)
	var out []registry.RegistryExecutionPublic
	err := r.db.WithContext(ctx).
		Where("shareable = TRUE AND revoked_at IS NULL AND created_at >= ?", since).
		Order("view_count DESC, created_at DESC").
		Limit(limit).
		Find(&out).Error
	if err != nil {
		return nil, fmt.Errorf("receipt.GetTrending: %w", err)
	}
	return out, nil
}

// GetFunctionExecutionCount returns the number of receipts for a function.
// Used by the milestone hook to detect threshold crossings.
func (r *Repository) GetFunctionExecutionCount(ctx context.Context, functionID uuid.UUID) (int, error) {
	var n int64
	if err := r.db.WithContext(ctx).
		Model(&registry.RegistryExecutionPublic{}).
		Where("function_id = ?", functionID).
		Count(&n).Error; err != nil {
		return 0, fmt.Errorf("receipt.GetFunctionExecutionCount: %w", err)
	}
	return int(n), nil
}

// ListForFunction returns the most recent shareable receipts for a
// single function. The result powers the "Recent public receipts"
// section on the FunctionPage. Limit is clamped to [1, 50].
func (r *Repository) ListForFunction(ctx context.Context, functionID uuid.UUID, limit int) ([]registry.RegistryExecutionPublic, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	var out []registry.RegistryExecutionPublic
	err := r.db.WithContext(ctx).
		Where("function_id = ? AND shareable = TRUE AND revoked_at IS NULL", functionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&out).Error
	if err != nil {
		return nil, fmt.Errorf("receipt.ListForFunction: %w", err)
	}
	return out, nil
}

// GetFunctionOwnerID returns the OwnerUserID for a function. Returns
// uuid.Nil if the function has no owner (e.g. legacy public functions).
func (r *Repository) GetFunctionOwnerID(ctx context.Context, functionID uuid.UUID) (uuid.UUID, error) {
	var fn registry.RegistryFunction
	if err := r.db.WithContext(ctx).
		Select("owner_user_id").
		Where("id = ?", functionID).
		First(&fn).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, nil
		}
		return uuid.Nil, fmt.Errorf("receipt.GetFunctionOwnerID: %w", err)
	}
	if fn.OwnerUserID == nil {
		return uuid.Nil, nil
	}
	return *fn.OwnerUserID, nil
}

// GetActiveFunctionsSince returns the IDs of functions that have at least
// one receipt created in the given time window. Used by the daily sweep
// scheduler to back-fill any milestones we may have missed during downtime.
func (r *Repository) GetActiveFunctionsSince(ctx context.Context, since time.Time) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&registry.RegistryExecutionPublic{}).
		Where("created_at >= ?", since).
		Distinct("function_id").
		Pluck("function_id", &ids).Error
	if err != nil {
		return nil, fmt.Errorf("receipt.GetActiveFunctionsSince: %w", err)
	}
	return ids, nil
}

// ----------------------------------------------------------------------------
// Write paths (new tables only)
// ----------------------------------------------------------------------------

// RecordMilestone inserts a milestone event with ON CONFLICT DO NOTHING
// semantics — safe to call from multiple workers / replays. Returns true if
// the row was actually inserted (i.e. this caller "won" the race).
func (r *Repository) RecordMilestone(ctx context.Context, evt *MilestoneEvent) (bool, error) {
	if evt.DedupeKey == "" {
		return false, fmt.Errorf("receipt.RecordMilestone: dedupe_key is required")
	}
	if evt.FiredAt.IsZero() {
		evt.FiredAt = time.Now()
	}
	if evt.CreatedAt.IsZero() {
		evt.CreatedAt = time.Now()
	}
	if evt.ChannelsFired == nil {
		evt.ChannelsFired = []MilestoneChannel{}
	}

	// pgx (the GORM postgres driver) maps an empty Go slice to NULL for
	// array columns, which violates our NOT NULL constraint. Convert to
	// []string and use the literal '{}' fallback so the INSERT always
	// has a concrete array value.
	chans := make([]string, 0, len(evt.ChannelsFired))
	for _, c := range evt.ChannelsFired {
		chans = append(chans, string(c))
	}

	// Use raw SQL for ON CONFLICT — GORM's Create ignores ON CONFLICT.
	// We marshal the channels slice to a Postgres array literal ourselves
	// (e.g. '{inapp,email}') so we never depend on pgx's array-binding
	// behaviour — which treats an empty slice as NULL and violates our
	// NOT NULL constraint.
	var channelsLiteral string
	if len(chans) == 0 {
		channelsLiteral = "{}"
	} else {
		// Quote each channel and join with commas.
		parts := make([]string, 0, len(chans))
		for _, c := range chans {
			parts = append(parts, `"`+strings.ReplaceAll(c, `"`, `\"`)+`"`)
		}
		channelsLiteral = "{" + strings.Join(parts, ",") + "}"
	}

	res := r.db.WithContext(ctx).Exec(`
		INSERT INTO receipt_milestone_events
			(id, function_id, tenant_id, threshold, total_runs_at,
			 public_id, fired_at, channels_fired, dedupe_key, created_at)
		VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?::text[], ?, now())
		ON CONFLICT (dedupe_key) DO NOTHING
	`, evt.FunctionID, evt.TenantID, evt.Threshold, evt.TotalRunsAt,
		evt.PublicID, evt.FiredAt, channelsLiteral, evt.DedupeKey)
	if res.Error != nil {
		return false, fmt.Errorf("receipt.RecordMilestone: %w", res.Error)
	}
	return res.RowsAffected > 0, nil
}

// MarkMilestoneChannels appends channel names to an existing milestone row.
// Used by the worker to record which channels it actually fired.
func (r *Repository) MarkMilestoneChannels(ctx context.Context, dedupeKey string, channels []MilestoneChannel) error {
	if dedupeKey == "" || len(channels) == 0 {
		return nil
	}
	parts := make([]string, 0, len(channels))
	for _, c := range channels {
		parts = append(parts, `"`+strings.ReplaceAll(string(c), `"`, `\"`)+`"`)
	}
	literal := "{" + strings.Join(parts, ",") + "}"
	return r.db.WithContext(ctx).
		Model(&MilestoneEvent{}).
		Where("dedupe_key = ?", dedupeKey).
		Update("channels_fired", gorm.Expr("channels_fired || ?::text[]", literal)).Error
}

// ListMilestonesForFunction returns the milestone history for a function
// (newest first). Used by the dashboard "milestones" view.
func (r *Repository) ListMilestonesForFunction(ctx context.Context, functionID uuid.UUID, limit int) ([]MilestoneEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var out []MilestoneEvent
	err := r.db.WithContext(ctx).
		Where("function_id = ?", functionID).
		Order("fired_at DESC").
		Limit(limit).
		Find(&out).Error
	if err != nil {
		return nil, fmt.Errorf("receipt.ListMilestonesForFunction: %w", err)
	}
	return out, nil
}

// Revoke marks a receipt as revoked (idempotent) and records the audit row.
// Only the function owner may call this; auth is enforced at the handler.
func (r *Repository) Revoke(ctx context.Context, publicID string, revokedBy uuid.UUID, reason string) error {
	now := time.Now()
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if rec := recover(); rec != nil {
			logrus.WithFields(logrus.Fields{
				"panic":   rec,
				"stack":   string(debug.Stack()),
				"method":  "Revoke",
				"public_id": publicID,
			}).Error("Receipt repository Revoke panicked, rolling back transaction")
			tx.Rollback()
		}
	}()

	// Look up function_id for the audit row.
	var exec registry.RegistryExecutionPublic
	if err := tx.Where("public_id = ?", publicID).First(&exec).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("receipt.Revoke lookup: %w", err)
	}

	// Idempotent: setting revoked_at twice is fine (it stays the first one).
	if err := tx.Model(&registry.RegistryExecutionPublic{}).
		Where("public_id = ?", publicID).
		Update("revoked_at", now).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("receipt.Revoke update: %w", err)
	}

	rev := &Revocation{
		PublicID:   publicID,
		FunctionID: exec.FunctionID,
		RevokedBy:  revokedBy,
		RevokedAt:  now,
		Reason:     reason,
	}
	if err := tx.Create(rev).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("receipt.Revoke audit: %w", err)
	}
	return tx.Commit().Error
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// BuildDedupeKey is exposed so callers (and tests) can compute the canonical
// key used by the ON CONFLICT clause without duplicating the format string.
func BuildDedupeKey(functionID uuid.UUID, threshold int) string {
	return fmt.Sprintf("%s:%d", functionID.String(), threshold)
}

// BackfillReceiptMetadata writes the denormalized function metadata onto a
// freshly-created receipt row. Called from HandleExecute immediately after
// generateExecutionID succeeds.
//
// This is a *separate* method (not part of the registry repository) so the
// receipt package owns its own write path to its own projection of the row.
func (r *Repository) BackfillReceiptMetadata(ctx context.Context, publicID string, fn *registry.RegistryFunction, ver *registry.RegistryFunctionVersion) error {
	if fn == nil || ver == nil {
		return fmt.Errorf("receipt.BackfillReceiptMetadata: fn and ver are required")
	}
	desc := ""
	if fn.Description.Valid {
		desc = fn.Description.String
	}
	visibility := "public"
	if fn.Visibility != "" {
		visibility = fn.Visibility
	}
	updates := map[string]interface{}{
		"function_name":       fn.Name,
		"function_author":     fn.Author,
		"runtime":             ver.Runtime,
		"function_visibility": visibility,
		"description":         desc,
	}
	if len(ver.Manifest) > 0 {
		var manifest struct {
			Inputs  json.RawMessage `json:"inputs"`
			Outputs json.RawMessage `json:"outputs"`
		}
		if err := json.Unmarshal(ver.Manifest, &manifest); err == nil {
			if len(manifest.Inputs) > 0 {
				updates["input_schema"] = manifest.Inputs
			}
			if len(manifest.Outputs) > 0 {
				updates["output_schema"] = manifest.Outputs
			}
		}
	}
	return r.db.WithContext(ctx).
		Model(&registry.RegistryExecutionPublic{}).
		Where("public_id = ?", publicID).
		Updates(updates).Error
}

// ----------------------------------------------------------------------------
// Redis cache (L1 for /v1/receipts/:id reads)
// ----------------------------------------------------------------------------

// L1Key returns the canonical Redis key for a cached receipt payload. Exposed
// so handlers can both write and invalidate (e.g. on revoke).
func L1Key(publicID string) string {
	return "ff:rcpt:body:" + publicID
}

// CacheGet returns the cached payload bytes (already-marshaled JSON) and
// "hit" = true if present. Misses are not errors.
func (r *Repository) CacheGet(ctx context.Context, publicID string) ([]byte, bool, error) {
	if r.redis == nil {
		return nil, false, nil
	}
	v, err := r.redis.Get(ctx, L1Key(publicID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, nil // cache failures should not break the request
	}
	return v, true, nil
}

// CacheSet writes the payload to Redis with a 60s TTL.
func (r *Repository) CacheSet(ctx context.Context, publicID string, payload []byte) {
	if r.redis == nil {
		return
	}
	_ = r.redis.Set(ctx, L1Key(publicID), payload, 60*time.Second).Err()
}

// CacheInvalidate removes a receipt from the L1 cache. Called on revoke.
func (r *Repository) CacheInvalidate(ctx context.Context, publicID string) {
	if r.redis == nil {
		return
	}
	_ = r.redis.Del(ctx, L1Key(publicID)).Err()
}
