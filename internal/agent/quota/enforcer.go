package quota

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Enforcer checks and consumes agent quotas using Redis atomic counters
// with Postgres as the source of truth for configuration.
type Enforcer struct {
	redis    *redis.Client
	db       *gorm.DB
	identityRepo *identity.Repository
}

// NewEnforcer creates a new quota enforcer
func NewEnforcer(db *gorm.DB, redisClient *redis.Client) *Enforcer {
	return &Enforcer{
		redis:        redisClient,
		db:           db,
		identityRepo: identity.NewRepository(db),
	}
}

// CheckAndConsume validates quota and atomically increments counters.
// Returns a QuotaResult indicating whether the call is allowed.
// estimatedCost is the expected USD cost of this execution (may be 0 for free functions).
func (e *Enforcer) CheckAndConsume(ctx context.Context, agentID string, functionURI string, estimatedCost float64) (*QuotaResult, error) {
	// Load quota config from DB
	quotaCfg, err := e.identityRepo.GetQuotaConfig(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load quota config: %w", err)
	}

	// Check function allow/deny lists
	if err := checkFunctionAccess(functionURI, quotaCfg.AllowedFunctions, quotaCfg.ForbiddenFunctions); err != nil {
		return &QuotaResult{
			Allowed: false,
			Reason:  err.Error(),
		}, err
	}

	// Check per-minute rate limit
	minuteKey := fmt.Sprintf("quota:calls:minute:%s:%s", agentID, minuteBucket())
	minuteCount, err := e.incrementCounter(ctx, minuteKey, 60)
	if err != nil {
		return nil, fmt.Errorf("failed to check minute quota: %w", err)
	}
	if int(minuteCount) > quotaCfg.MaxCallsPerMinute {
		// Decrement since we over-counted
		e.redis.Decr(ctx, minuteKey)
		retryAfter := 60 - time.Now().Second()
		return &QuotaResult{
			Allowed:        false,
			RetryAfterSecs: retryAfter,
			Reason:         string(ViolationRateLimitMinute),
		}, &QuotaViolationError{
			Code:           ViolationRateLimitMinute,
			Message:        fmt.Sprintf("rate limit exceeded: %d calls/min (limit: %d)", minuteCount, quotaCfg.MaxCallsPerMinute),
			RetryAfterSecs: retryAfter,
		}
	}

	// Check per-day rate limit
	dayKey := fmt.Sprintf("quota:calls:day:%s:%s", agentID, dayBucket())
	dayCount, err := e.incrementCounter(ctx, dayKey, 86400)
	if err != nil {
		e.redis.Decr(ctx, minuteKey)
		return nil, fmt.Errorf("failed to check daily quota: %w", err)
	}
	if int(dayCount) > quotaCfg.MaxCallsPerDay {
		e.redis.Decr(ctx, minuteKey)
		e.redis.Decr(ctx, dayKey)
		secondsUntilMidnight := secondsUntilMidnightUTC()
		return &QuotaResult{
			Allowed:        false,
			RetryAfterSecs: secondsUntilMidnight,
			Reason:         string(ViolationRateLimitDay),
		}, &QuotaViolationError{
			Code:           ViolationRateLimitDay,
			Message:        fmt.Sprintf("daily call limit exceeded: %d calls/day (limit: %d)", dayCount, quotaCfg.MaxCallsPerDay),
			RetryAfterSecs: secondsUntilMidnight,
		}
	}

	// Check daily spend cap (if estimatedCost > 0)
	if estimatedCost > 0 && quotaCfg.MaxDailySpendUSD > 0 {
		spendKey := fmt.Sprintf("quota:spend:day:%s:%s", agentID, dayBucket())
		currentSpend, err := e.getSpendCounter(ctx, spendKey)
		if err != nil {
			// Non-fatal: log and continue
			currentSpend = 0
		}
		if currentSpend+estimatedCost > quotaCfg.MaxDailySpendUSD {
			e.redis.Decr(ctx, minuteKey)
			e.redis.Decr(ctx, dayKey)
			return &QuotaResult{
				Allowed:        false,
				RemainingSpend: quotaCfg.MaxDailySpendUSD - currentSpend,
				Reason:         string(ViolationSpendCapDaily),
			}, &QuotaViolationError{
				Code:    ViolationSpendCapDaily,
				Message: fmt.Sprintf("daily spend cap exceeded: $%.4f spent (cap: $%.2f)", currentSpend, quotaCfg.MaxDailySpendUSD),
			}
		}
	}

	remaining := quotaCfg.MaxCallsPerDay - int(dayCount)
	if remaining < 0 {
		remaining = 0
	}

	return &QuotaResult{
		Allowed:        true,
		RemainingCalls: remaining,
	}, nil
}

// RecordSpend records the actual cost of an execution after it completes
func (e *Enforcer) RecordSpend(ctx context.Context, agentID string, actualCostUSD float64) error {
	if actualCostUSD <= 0 {
		return nil
	}
	spendKey := fmt.Sprintf("quota:spend:day:%s:%s", agentID, dayBucket())
	// Use INCRBYFLOAT for spend tracking
	return e.redis.IncrByFloat(ctx, spendKey, actualCostUSD).Err()
}

// GetCurrentUsage returns the current usage counters for an agent
func (e *Enforcer) GetCurrentUsage(ctx context.Context, agentID string) (*AgentUsage, error) {
	minuteKey := fmt.Sprintf("quota:calls:minute:%s:%s", agentID, minuteBucket())
	dayKey := fmt.Sprintf("quota:calls:day:%s:%s", agentID, dayBucket())
	spendKey := fmt.Sprintf("quota:spend:day:%s:%s", agentID, dayBucket())

	pipe := e.redis.Pipeline()
	minuteCmd := pipe.Get(ctx, minuteKey)
	dayCmd := pipe.Get(ctx, dayKey)
	spendCmd := pipe.Get(ctx, spendKey)
	pipe.Exec(ctx)

	usage := &AgentUsage{
		AgentID:     agentID,
		LastUpdated: time.Now(),
	}

	if v, err := minuteCmd.Int64(); err == nil {
		usage.CallsThisMinute = v
	}
	if v, err := dayCmd.Int64(); err == nil {
		usage.CallsToday = v
	}
	if v, err := spendCmd.Float64(); err == nil {
		usage.SpendTodayUSD = v
	}

	return usage, nil
}

// ResetDailyCounters resets daily counters for all agents (called by a cron job at midnight UTC)
func (e *Enforcer) ResetDailyCounters(ctx context.Context) error {
	// Redis TTL handles expiry automatically; this is a no-op for Redis-based counters.
	// For Postgres-based aggregates, a separate job would reset them.
	return nil
}

// incrementCounter atomically increments a Redis counter and sets TTL on first use
func (e *Enforcer) incrementCounter(ctx context.Context, key string, ttlSeconds int) (int64, error) {
	pipe := e.redis.TxPipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Duration(ttlSeconds)*time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incrCmd.Val(), nil
}

// getSpendCounter reads the current spend value from Redis
func (e *Enforcer) getSpendCounter(ctx context.Context, key string) (float64, error) {
	val, err := e.redis.Get(ctx, key).Float64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// checkFunctionAccess validates that a function URI is allowed by the quota config
func checkFunctionAccess(functionURI string, allowed, forbidden []string) error {
	// Check forbidden list first
	for _, pattern := range forbidden {
		if matchFunctionPattern(functionURI, pattern) {
			return &QuotaViolationError{
				Code:    ViolationFunctionForbidden,
				Message: fmt.Sprintf("function %s is forbidden by agent policy", functionURI),
			}
		}
	}

	// If allowed list is set, function must match at least one pattern
	if len(allowed) > 0 {
		for _, pattern := range allowed {
			if matchFunctionPattern(functionURI, pattern) {
				return nil
			}
		}
		return &QuotaViolationError{
			Code:    ViolationFunctionNotAllowed,
			Message: fmt.Sprintf("function %s is not in the allowed list", functionURI),
		}
	}

	return nil
}

// matchFunctionPattern matches a function URI against a pattern (supports * wildcard)
func matchFunctionPattern(uri, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(uri, prefix+"/")
	}
	return uri == pattern
}

// minuteBucket returns a string key for the current minute (UTC)
func minuteBucket() string {
	now := time.Now().UTC()
	return fmt.Sprintf("%d%02d%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())
}

// dayBucket returns a string key for the current day (UTC)
func dayBucket() string {
	now := time.Now().UTC()
	return fmt.Sprintf("%d%02d%02d", now.Year(), now.Month(), now.Day())
}

// secondsUntilMidnightUTC returns the number of seconds until midnight UTC
func secondsUntilMidnightUTC() int {
	now := time.Now().UTC()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	return int(midnight.Sub(now).Seconds())
}
