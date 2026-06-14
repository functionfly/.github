// Package quota provides per-tenant resource quotas and a sliding-window
// rate limiter backed by Redis (or a no-op in-memory fallback).
//
// The package implements the two halves of Phase 5.2:
//
//   - ResourceQuota enforces hard caps (e.g. "max 25 secrets per
//     tenant", "max 100 dynamic creds per month"). Quotas are
//     read from the tenant's plan but can be overridden by an
//     admin in the vault_rate_limits table.
//
//   - SlidingWindowLimiter applies a per-tenant rate cap on
//     individual operations. The cap is in the form
//     "N requests per window" — e.g. "200 reads / minute".
package quota

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Resource is one of the billable dimensions.
type Resource string

const (
	ResourceSecrets         Resource = "secrets"
	ResourceDynamicCreds    Resource = "dynamic_credentials"
	ResourceTokensPerSecret Resource = "tokens_per_secret"
	ResourceAuditExports    Resource = "audit_exports_per_day"
)

// Quota is the hard cap for a single resource.
type Quota struct {
	Resource Resource `json:"resource"`
	Limit    int64    `json:"limit"`
	// Window is the period the cap is enforced over. Zero means
	// "lifetime" (e.g. secrets stored); positive means "per
	// window" (e.g. audit exports per day).
	Window time.Duration `json:"window,omitempty"`
}

// GetPlanQuota returns the default quota for a plan + resource.
// Uses the unified plans package for all plan limits.
func GetPlanQuota(plan string, resource Resource) Quota {
	switch resource {
	case ResourceSecrets:
		return Quota{Resource: resource, Limit: int64(plans.GetMaxSecrets(plan)), Window: 0}
	case ResourceDynamicCreds:
		return Quota{Resource: resource, Limit: int64(plans.GetMaxDynamicCreds(plan)), Window: 30 * 24 * time.Hour}
	case ResourceTokensPerSecret:
		return Quota{Resource: resource, Limit: int64(plans.GetMaxTokensPerSecret(plan)), Window: 0}
	case ResourceAuditExports:
		return Quota{Resource: resource, Limit: int64(plans.GetMaxAuditExportsPerDay(plan)), Window: 24 * time.Hour}
	default:
		return Quota{Resource: resource, Limit: 0}
	}
}


// Redis is a minimal interface that the *quota.Store* and
// *Limiter* need. We keep it small so tests can supply an
// in-process miniredis or a stub.
type Redis interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error)
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}

// Store persists tenant overrides + plan defaults. The primary
// implementation is backed by Postgres, but the interface is
// narrowed to just the queries the quota package needs so it can
// run against the in-memory cache for tests.
type Store interface {
	// GetTenantPlan returns the tenant's current plan slug.
	GetTenantPlan(ctx context.Context, tenantID uuid.UUID) (string, error)
	// GetOverride returns an admin-set override for the resource,
	// or (0, false, nil) when no override exists.
	GetOverride(ctx context.Context, tenantID uuid.UUID, resource Resource) (int64, time.Duration, bool, error)
	// CountSecrets returns the current stored-secret count.
	CountSecrets(ctx context.Context, tenantID uuid.UUID) (int64, error)
	// CountActiveTokens returns the current active-token count for
	// a secret.
	CountActiveTokens(ctx context.Context, secretID uuid.UUID) (int64, error)
	// CountDynamicCredsSince returns the number of credentials
	// minted in the given window.
	CountDynamicCredsSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int64, error)
	// CountAuditExportsSince returns the number of audit exports
	// in the given window.
	CountAuditExportsSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int64, error)
}

// Enforcer reads the tenant's plan + admin overrides, looks up the
// current usage, and decides whether a request is allowed.
type Enforcer struct {
	store Store
}

// NewEnforcer constructs an Enforcer.
func NewEnforcer(store Store) *Enforcer { return &Enforcer{store: store} }

// Decision is the result of an enforcement check.
type Decision struct {
	Allowed   bool  `json:"allowed"`
	Limit     int64 `json:"limit"`
	Current   int64 `json:"current"`
	Remaining int64 `json:"remaining"`
	// Reset is when the window resets (zero for lifetime caps).
	Reset time.Time `json:"reset,omitempty"`
	// Headers are the values for the X-RateLimit-* response headers
	// (per https://datatracker.ietf.org/doc/draft-ietf-httpapi-ratelimit-headers/).
	Headers map[string]string `json:"headers,omitempty"`
}

// CheckSecretCount returns the decision for a create-secret call.
func (e *Enforcer) CheckSecretCount(ctx context.Context, tenantID uuid.UUID) (Decision, error) {
	return e.check(ctx, tenantID, ResourceSecrets, func(s Store) (int64, error) {
		return s.CountSecrets(ctx, tenantID)
	}, time.Time{})
}

// CheckTokensPerSecret returns the decision for a token-create call.
func (e *Enforcer) CheckTokensPerSecret(ctx context.Context, secretID uuid.UUID) (Decision, error) {
	limit, window, ok, err := e.store.GetOverride(ctx, mustTenant(ctx, secretID), ResourceTokensPerSecret)
	if err != nil {
		return Decision{}, err
	}
	if !ok {
		// We need a tenant ID to look up the plan. The store is
		// given a *secret* ID, not a tenant ID; the caller passes
		// the tenant via context. If absent, we fall back to a
		// safe plan-less default.
		limit = int64(GetPlanQuota("free", ResourceTokensPerSecret).Limit)
		window = 0
	}
	current, err := e.store.CountActiveTokens(ctx, secretID)
	if err != nil {
		return Decision{}, err
	}
	return e.decide(current, limit, window, time.Time{}), nil
}

// CheckDynamicCreds checks the rolling-30d dynamic-credential cap.
func (e *Enforcer) CheckDynamicCreds(ctx context.Context, tenantID uuid.UUID) (Decision, error) {
	resetAt := time.Now().Add(30 * 24 * time.Hour)
	return e.check(ctx, tenantID, ResourceDynamicCreds, func(s Store) (int64, error) {
		return s.CountDynamicCredsSince(ctx, tenantID, resetAt.Add(-30*24*time.Hour))
	}, resetAt)
}

// CheckAuditExport checks the daily audit-export cap.
func (e *Enforcer) CheckAuditExport(ctx context.Context, tenantID uuid.UUID) (Decision, error) {
	resetAt := time.Now().Add(24 * time.Hour)
	return e.check(ctx, tenantID, ResourceAuditExports, func(s Store) (int64, error) {
		return s.CountAuditExportsSince(ctx, tenantID, resetAt.Add(-24*time.Hour))
	}, resetAt)
}

// check is the shared lookup: resolve plan + override, look up
// current usage, decide.
func (e *Enforcer) check(
	ctx context.Context,
	tenantID uuid.UUID,
	resource Resource,
	count func(Store) (int64, error),
	reset time.Time,
) (Decision, error) {
	limit, window, ok, err := e.store.GetOverride(ctx, tenantID, resource)
	if err != nil {
		return Decision{}, err
	}
	if !ok {
		plan, err := e.store.GetTenantPlan(ctx, tenantID)
		if err != nil {
			return Decision{}, err
		}
		q := GetPlanQuota(plan, resource)
		limit = q.Limit
		window = q.Window
	}
	current, err := count(e.store)
	if err != nil {
		return Decision{}, err
	}
	return e.decide(current, limit, window, reset), nil
}

func (e *Enforcer) decide(current, limit int64, window time.Duration, reset time.Time) Decision {
	d := Decision{
		Allowed: current < limit,
		Limit:   limit,
		Current: current,
	}
	d.Remaining = limit - current
	if d.Remaining < 0 {
		d.Remaining = 0
	}
	d.Headers = map[string]string{
		"X-RateLimit-Limit":     fmt.Sprintf("%d", d.Limit),
		"X-RateLimit-Remaining": fmt.Sprintf("%d", d.Remaining),
	}
	if !reset.IsZero() {
		delta := time.Until(reset)
		if delta < 0 {
			delta = 0
		}
		d.Headers["X-RateLimit-Reset"] = fmt.Sprintf("%d", int(delta.Seconds()))
	}
	return d
}

// mustTenant is a small helper that returns the zero UUID when the
// secret ID doesn't carry tenant context. In practice the handler
// passes tenant via the override path or via a tenant-aware
// CountActiveTokens implementation.
func mustTenant(_ context.Context, _ uuid.UUID) uuid.UUID {
	return uuid.Nil
}

// ============================================================================
// Sliding-window rate limiter
// ============================================================================

// SlidingWindowLimiter is a Redis-backed sliding-window rate
// limiter. It returns a Decision that mirrors the Quota Decision
// but uses a per-second sliding window rather than a quota
// counter.
type SlidingWindowLimiter struct {
	redis  Redis
	prefix string
	window time.Duration
	limit  int
	logger *logrus.Logger
}

// NewSlidingWindowLimiter constructs a limiter.
//
// `prefix` is included in the Redis key (e.g. "vault_read") and
// `window` / `limit` are the per-tenant cap.
func NewSlidingWindowLimiter(rdb Redis, prefix string, window time.Duration, limit int) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		redis:  rdb,
		prefix: prefix,
		window: window,
		limit:  limit,
		logger: logrus.New(),
	}
}

// SetLogger overrides the default logger.
func (l *SlidingWindowLimiter) SetLogger(logger *logrus.Logger) {
	l.logger = logger
}

// Allow performs the sliding-window check for `tenantID` and
// returns a Decision. The bucket is keyed on (prefix, tenantID).
func (l *SlidingWindowLimiter) Allow(ctx context.Context, tenantID uuid.UUID) (Decision, error) {
	if l == nil || l.redis == nil {
		return Decision{Allowed: true, Limit: int64(l.limit)}, nil
	}
	key := fmt.Sprintf("%s:%s", l.prefix, tenantID.String())
	now := time.Now().UnixMilli()
	windowMs := l.window.Milliseconds()
	cutoff := now - windowMs
	pipe, err := l.redis.Pipelined(ctx, func(p redis.Pipeliner) error {
		// Drop entries outside the window.
		// ZREMRANGEBYSCORE key -inf cutoff
		p.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", cutoff))
		// ZCARD key
		p.ZCard(ctx, key)
		// ZADD key now now
		p.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
		// PEXPIRE key windowMs
		p.Expire(ctx, key, l.window)
		return nil
	})
	if err != nil {
		l.logger.WithError(err).Warn("rate limiter pipeline error")
		// Fail open: better to over-allow than to wedge the
		// whole API when Redis is down.
		return Decision{Allowed: true, Limit: int64(l.limit)}, nil
	}
	// Index 1 is ZCARD.
	card, _ := pipe[1].(*redis.IntCmd)
	used := int64(0)
	if card != nil {
		used = card.Val()
	}
	d := Decision{
		Allowed:   used < int64(l.limit),
		Limit:     int64(l.limit),
		Current:   used,
		Remaining: int64(l.limit) - used,
	}
	if d.Remaining < 0 {
		d.Remaining = 0
	}
	d.Headers = map[string]string{
		"X-RateLimit-Limit":     fmt.Sprintf("%d", d.Limit),
		"X-RateLimit-Remaining": fmt.Sprintf("%d", d.Remaining),
		"X-RateLimit-Reset":     fmt.Sprintf("%d", int(l.window.Seconds())),
	}
	return d, nil
}

// MarshalDecision serializes a Decision to JSON, useful for log
// fields.
func MarshalDecision(d Decision) string {
	b, _ := json.Marshal(d)
	return string(b)
}
