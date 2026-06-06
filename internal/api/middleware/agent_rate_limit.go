package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	AgentRegistrationRateLimitWindow     = 1 * time.Minute
	AgentRegistrationRateLimitMaxRequests = 10
	AgentRegistrationRateLimitKeyPrefix   = "ratelimit:agent:registration:"
)

type AgentRateLimiter struct {
	redisClient *redis.Client
	memStore    *inMemoryStore
	logger      *logrus.Entry
}

func NewAgentRateLimiter(redisClient *redis.Client) *AgentRateLimiter {
	return &AgentRateLimiter{
		redisClient: redisClient,
		memStore:    newInMemoryStore(),
		logger:      logrus.WithField("middleware", "agent_ratelimit"),
	}
}

func (m *AgentRateLimiter) CheckRateLimit(ctx context.Context, tenantID string, limit int) (*RateLimitResult, error) {
	if m.redisClient == nil {
		entry := m.memStore.getOrCreate(AgentRegistrationRateLimitKeyPrefix + tenantID)
		entry.mu.Lock()
		defer entry.mu.Unlock()

		now := time.Now()
		cutoff := now.Add(-AgentRegistrationRateLimitWindow)
		valid := entry.timestamps[:0]
		for _, t := range entry.timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		entry.timestamps = valid

		remaining := limit - len(entry.timestamps)
		if remaining < 0 {
			remaining = 0
		}

		if len(entry.timestamps) >= limit {
			return &RateLimitResult{
				Allowed:    false,
				Remaining:  0,
				TotalLimit: limit,
				ResetAt:    now.Add(AgentRegistrationRateLimitWindow),
			}, nil
		}

		entry.timestamps = append(entry.timestamps, now)
		return &RateLimitResult{
			Allowed:    true,
			Remaining:  remaining - 1,
			TotalLimit: limit,
			ResetAt:    now.Add(AgentRegistrationRateLimitWindow),
		}, nil
	}

	now := time.Now()
	windowStart := now.Add(-AgentRegistrationRateLimitWindow).UnixMilli()
	currentTime := float64(now.UnixNano()) / float64(time.Millisecond)

	redisKey := AgentRegistrationRateLimitKeyPrefix + tenantID

	if err := m.redisClient.ZRemRangeByScore(ctx, redisKey, "-inf", strconv.FormatInt(windowStart, 10)).Err(); err != nil {
		m.logger.WithError(err).Error("Failed to remove old rate limit entries")
		return nil, err
	}

	count, err := m.redisClient.ZCard(ctx, redisKey).Result()
	if err != nil {
		m.logger.WithError(err).Error("Failed to count rate limit entries")
		return nil, err
	}

	remaining := limit - int(count) - 1
	if remaining < 0 {
		remaining = 0
	}

	resetAt := now.Add(AgentRegistrationRateLimitWindow)

	if count >= int64(limit) {
		return &RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: int(AgentRegistrationRateLimitWindow.Seconds()),
			TotalLimit: limit,
			ResetAt:    resetAt,
		}, nil
	}

	if err := m.redisClient.ZAdd(ctx, redisKey, redis.Z{
		Score:  currentTime,
		Member: fmt.Sprintf("%.0f", currentTime),
	}).Err(); err != nil {
		m.logger.WithError(err).Error("Failed to add rate limit entry")
		return nil, err
	}

	if err := m.redisClient.Expire(ctx, redisKey, AgentRegistrationRateLimitWindow+time.Second).Err(); err != nil {
		m.logger.WithError(err).Error("Failed to set rate limit key expiry")
	}

	return &RateLimitResult{
		Allowed:    true,
		Remaining:  remaining,
		TotalLimit: limit,
		ResetAt:    resetAt,
	}, nil
}

func (m *AgentRateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r)
		if claims == nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		tenantID := claims.TenantID.String()
		limit := AgentRegistrationRateLimitMaxRequests

		if v := os.Getenv("AGENT_REGISTRATION_RATE_LIMIT"); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		result, err := m.CheckRateLimit(ctx, tenantID, limit)
		if err != nil {
			m.logger.WithError(err).WithField("tenant_id", tenantID).Error("Rate limit check failed; blocking request for safety")
			writeAgentRateLimitError(w, "Rate limit service unavailable", int(AgentRegistrationRateLimitWindow.Seconds()))
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.TotalLimit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

		if !result.Allowed {
			m.logger.WithFields(logrus.Fields{
				"tenant_id":   tenantID,
				"path":        r.URL.Path,
				"retry_after": result.RetryAfter,
			}).Warn("Agent registration rate limit exceeded")

			w.Header().Set("Retry-After", strconv.Itoa(result.RetryAfter))
			writeAgentRateLimitError(w, "Too many agent registrations", result.RetryAfter)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func writeAgentRateLimitError(w http.ResponseWriter, message string, retryAfter int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	w.WriteHeader(http.StatusTooManyRequests)
	w.Write([]byte(fmt.Sprintf(`{"error":"rate_limit_exceeded","code":"AGENT_REGISTRATION_LIMIT","message":"%s","retry_after":%d}`, message, retryAfter)))
}
