// Package apikey provides API key generation, hashing, validation, and rate limiting functionality.
package apikey

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Rate limit key prefixes
const (
	rateLimitKeyPrefix = "ratelimit:api:"
	rateLimitMinute    = "minute"
	rateLimitHour      = "hour"
	rateLimitDay       = "day"
)

// Redis key TTLs
const (
	rateLimitTTLMinute = 60 * time.Second
	rateLimitTTLHour   = 3600 * time.Second
	rateLimitTTLDay    = 86400 * time.Second
)

// RateLimitResult holds the result of a rate limit check
type RateLimitResult struct {
	Allowed            bool
	Remaining          int
	ResetTime          time.Time
	RetryAfter         time.Duration
	ServiceUnavailable bool // Set to true when Redis fails (fail-closed)
}

// RateLimiter provides distributed rate limiting using Redis
type RateLimiter struct {
	redis  *redis.Client
	db     *gorm.DB
	hasher *Hasher
	repo   *Repository
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *redis.Client, db *gorm.DB) *RateLimiter {
	return &RateLimiter{
		redis:  redisClient,
		db:     db,
		hasher: NewHasher(),
		repo:   NewRepository(db),
	}
}

// Limit checks and increments the rate limit for the given key hash
// Uses sliding window algorithm with Redis sorted sets
func (r *RateLimiter) Limit(ctx context.Context, keyHash string, rpm, rph, rpd int) (*RateLimitResult, error) {
	now := time.Now()
	currentTime := float64(now.UnixNano())

	// Default limits if not provided
	if rpm <= 0 {
		rpm = DefaultRateLimitRPM
	}
	if rph <= 0 {
		rph = DefaultRateLimitRPH
	}
	if rpd <= 0 {
		rpd = DefaultRateLimitRPD
	}

	// Check all three time windows
	result := r.checkWindow(ctx, keyHash, rateLimitMinute, currentTime, rpm)
	if !result.Allowed {
		return result, nil
	}

	result = r.checkWindow(ctx, keyHash, rateLimitHour, currentTime, rph)
	if !result.Allowed {
		return result, nil
	}

	result = r.checkWindow(ctx, keyHash, rateLimitDay, currentTime, rpd)
	if !result.Allowed {
		return result, nil
	}

	// All limits OK - record the request in all windows
	r.recordRequest(ctx, keyHash, currentTime)

	// Get remaining counts
	remaining := r.GetRemaining(ctx, keyHash, rpm, rph, rpd)

	return &RateLimitResult{
		Allowed:    true,
		Remaining:  remaining,
		ResetTime:  now.Add(rateLimitTTLMinute),
		RetryAfter: 0,
	}, nil
}

// checkWindow checks a specific time window using sliding window algorithm
func (r *RateLimiter) checkWindow(ctx context.Context, keyHash, window string, currentTime float64, limit int) *RateLimitResult {
	now := time.Now()
	var windowTTL time.Duration

	switch window {
	case rateLimitMinute:
		windowTTL = rateLimitTTLMinute
	case rateLimitHour:
		windowTTL = rateLimitTTLHour
	case rateLimitDay:
		windowTTL = rateLimitTTLDay
	default:
		windowTTL = rateLimitTTLMinute
	}

	// Calculate the window start time (for cleanup)
	windowStart := now.Add(-windowTTL)
	windowStartUnix := float64(windowStart.UnixNano())

	// Get the key for this window
	key := r.getRateLimitKey(keyHash, window)

	// Remove old entries outside the window
	r.redis.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatFloat(windowStartUnix, 'f', 0, 64))

	// Count current requests in this window
	count, err := r.redis.ZCard(ctx, key).Result()
	if err != nil {
		logrus.WithError(err).Warn("Failed to get rate limit count from Redis")
		// On error, deny the request (fail closed) to prevent abuse during Redis outages
		return &RateLimitResult{
			Allowed:            false,
			Remaining:          0,
			ResetTime:          now.Add(windowTTL),
			RetryAfter:         windowTTL,
			ServiceUnavailable: true,
		}
	}

	// Check if limit exceeded
	if int(count) >= limit {
		// Get the oldest entry to calculate reset time
		oldest, err := r.redis.ZRange(ctx, key, 0, 0).Result()
		if err == nil && len(oldest) > 0 {
			// Parse the timestamp
			ts, err := strconv.ParseFloat(oldest[0], 64)
			if err == nil {
				resetTime := time.Unix(0, int64(ts))
				retryAfter := resetTime.Sub(now)
				if retryAfter < 0 {
					retryAfter = time.Second
				}
				return &RateLimitResult{
					Allowed:    false,
					Remaining:  0,
					ResetTime:  resetTime,
					RetryAfter: retryAfter,
				}
			}
		}

		// Fallback reset time
		return &RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			ResetTime:  now.Add(windowTTL),
			RetryAfter: windowTTL,
		}
	}

	// Limit not exceeded
	return &RateLimitResult{
		Allowed:    true,
		Remaining:  limit - int(count) - 1, // -1 for the request we're about to add
		ResetTime:  now.Add(windowTTL),
		RetryAfter: 0,
	}
}

// recordRequest records a request in all time windows
func (r *RateLimiter) recordRequest(ctx context.Context, keyHash string, currentTime float64) {
	// Record in minute window
	minuteKey := r.getRateLimitKey(keyHash, rateLimitMinute)
	r.redis.ZAdd(ctx, minuteKey, redis.Z{Score: currentTime, Member: fmt.Sprintf("%.0f", currentTime)})
	r.redis.Expire(ctx, minuteKey, rateLimitTTLMinute)

	// Record in hour window
	hourKey := r.getRateLimitKey(keyHash, rateLimitHour)
	r.redis.ZAdd(ctx, hourKey, redis.Z{Score: currentTime, Member: fmt.Sprintf("%.0f", currentTime)})
	r.redis.Expire(ctx, hourKey, rateLimitTTLHour)

	// Record in day window
	dayKey := r.getRateLimitKey(keyHash, rateLimitDay)
	r.redis.ZAdd(ctx, dayKey, redis.Z{Score: currentTime, Member: fmt.Sprintf("%.0f", currentTime)})
	r.redis.Expire(ctx, dayKey, rateLimitTTLDay)
}

// getRateLimitKey generates the Redis key for rate limiting
func (r *RateLimiter) getRateLimitKey(keyHash, window string) string {
	return fmt.Sprintf("%s%s:%s", rateLimitKeyPrefix, keyHash, window)
}

// GetRemaining returns the minimum remaining requests across all windows
func (r *RateLimiter) GetRemaining(ctx context.Context, keyHash string, rpm, rph, rpd int) int {
	now := time.Now()

	// Default limits
	if rpm <= 0 {
		rpm = DefaultRateLimitRPM
	}
	if rph <= 0 {
		rph = DefaultRateLimitRPH
	}
	if rpd <= 0 {
		rpd = DefaultRateLimitRPD
	}

	// Calculate window start times
	minuteStart := now.Add(-rateLimitTTLMinute)
	hourStart := now.Add(-rateLimitTTLHour)
	dayStart := now.Add(-rateLimitTTLDay)

	// Get counts for each window
	minuteKey := r.getRateLimitKey(keyHash, rateLimitMinute)
	hourKey := r.getRateLimitKey(keyHash, rateLimitHour)
	dayKey := r.getRateLimitKey(keyHash, rateLimitDay)

	// Clean up old entries
	r.redis.ZRemRangeByScore(ctx, minuteKey, "-inf", strconv.FormatFloat(float64(minuteStart.UnixNano()), 'f', 0, 64))
	r.redis.ZRemRangeByScore(ctx, hourKey, "-inf", strconv.FormatFloat(float64(hourStart.UnixNano()), 'f', 0, 64))
	r.redis.ZRemRangeByScore(ctx, dayKey, "-inf", strconv.FormatFloat(float64(dayStart.UnixNano()), 'f', 0, 64))

	// Get counts
	minuteCount, _ := r.redis.ZCard(ctx, minuteKey).Result()
	hourCount, _ := r.redis.ZCard(ctx, hourKey).Result()
	dayCount, _ := r.redis.ZCard(ctx, dayKey).Result()

	// Return minimum remaining
	remaining := rpm - int(minuteCount)
	if rph-int(hourCount) < remaining {
		remaining = rph - int(hourCount)
	}
	if rpd-int(dayCount) < remaining {
		remaining = rpd - int(dayCount)
	}

	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetResetTime returns the earliest reset time across all windows
func (r *RateLimiter) GetResetTime(ctx context.Context, keyHash string) time.Time {
	now := time.Now()

	// Get the oldest entry in each window
	minuteKey := r.getRateLimitKey(keyHash, rateLimitMinute)
	hourKey := r.getRateLimitKey(keyHash, rateLimitHour)
	dayKey := r.getRateLimitKey(keyHash, rateLimitDay)

	var earliestReset time.Time

	// Check minute window
	oldest, err := r.redis.ZRange(ctx, minuteKey, 0, 0).Result()
	if err == nil && len(oldest) > 0 {
		ts, err := strconv.ParseFloat(oldest[0], 64)
		if err == nil {
			resetTime := time.Unix(0, int64(ts)).Add(rateLimitTTLMinute)
			if earliestReset.IsZero() || resetTime.Before(earliestReset) {
				earliestReset = resetTime
			}
		}
	}

	// Check hour window
	oldest, err = r.redis.ZRange(ctx, hourKey, 0, 0).Result()
	if err == nil && len(oldest) > 0 {
		ts, err := strconv.ParseFloat(oldest[0], 64)
		if err == nil {
			resetTime := time.Unix(0, int64(ts)).Add(rateLimitTTLHour)
			if earliestReset.IsZero() || resetTime.Before(earliestReset) {
				earliestReset = resetTime
			}
		}
	}

	// Check day window
	oldest, err = r.redis.ZRange(ctx, dayKey, 0, 0).Result()
	if err == nil && len(oldest) > 0 {
		ts, err := strconv.ParseFloat(oldest[0], 64)
		if err == nil {
			resetTime := time.Unix(0, int64(ts)).Add(rateLimitTTLDay)
			if earliestReset.IsZero() || resetTime.Before(earliestReset) {
				earliestReset = resetTime
			}
		}
	}

	// If no entries found, return default reset time
	if earliestReset.IsZero() {
		return now.Add(rateLimitTTLMinute)
	}

	return earliestReset
}

// APIKeyContextKey is the context key for storing API key information
type APIKeyContextKey string

const (
	// APIKeyIDContextKey is the context key for the API key ID
	APIKeyIDContextKey APIKeyContextKey = "api_key_id"
	// APIKeyHashContextKey is the context key for the API key hash
	APIKeyHashContextKey APIKeyContextKey = "api_key_hash"
	// APIKeyTenantIDContextKey is the context key for the tenant ID
	APIKeyTenantIDContextKey APIKeyContextKey = "api_key_tenant_id"
	// APIKeyUserIDContextKey is the context key for the user ID
	APIKeyUserIDContextKey APIKeyContextKey = "api_key_user_id"
	// APIKeyTypeContextKey is the context key for the API key type
	APIKeyTypeContextKey APIKeyContextKey = "api_key_type"
	// APIKeyNameContextKey is the context key for the API key name
	APIKeyNameContextKey APIKeyContextKey = "api_key_name"
)

// Middleware returns an HTTP middleware for API key rate limiting
func (r *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Extract API key from request
		plaintextKey := r.extractAPIKey(req)
		if plaintextKey == "" {
			http.Error(w, `{"error":{"code":"unauthorized","message":"Missing API key"}}`, http.StatusUnauthorized)
			return
		}

		// Hash the key
		keyHash := r.hasher.Hash(plaintextKey)

		// Look up the API key in the database
		apiKey, err := r.repo.GetByHash(req.Context(), keyHash)
		if err != nil {
			logrus.WithError(err).Warn("API key authentication failed - invalid key")
			http.Error(w, `{"error":{"code":"unauthorized","message":"Invalid API key"}}`, http.StatusUnauthorized)
			return
		}

		// Check if key is active
		if !apiKey.IsActive {
			http.Error(w, `{"error":{"code":"unauthorized","message":"API key is revoked"}}`, http.StatusUnauthorized)
			return
		}

		// Check if key has expired
		if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
			http.Error(w, `{"error":{"code":"unauthorized","message":"API key has expired"}}`, http.StatusUnauthorized)
			return
		}

		// Get rate limits from the API key
		rpm := apiKey.RateLimitRPM
		rph := apiKey.RateLimitRPH
		rpd := apiKey.RateLimitRPD

		// Check rate limits
		result, err := r.Limit(req.Context(), keyHash, rpm, rph, rpd)
		if err != nil {
			logrus.WithError(err).Error("Rate limit check failed")
			// On error, deny the request (fail closed)
			w.Header().Set("Retry-After", "60")
			http.Error(w, `{"error":{"code":"service_unavailable","message":"Rate limit service temporarily unavailable. Please retry later."}}`, http.StatusServiceUnavailable)
			return
		}

		// If rate limit service is unavailable (Redis failure), return 503
		if result != nil && result.ServiceUnavailable {
			logrus.WithError(err).Warn("Rate limit service unavailable - failing closed")
			w.Header().Set("Retry-After", "60")
			http.Error(w, `{"error":{"code":"service_unavailable","message":"Rate limit service temporarily unavailable. Please retry later."}}`, http.StatusServiceUnavailable)
			return
		}

		// If rate limited
		if result != nil && !result.Allowed {
			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rpm))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))
			w.Header().Set("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))

			http.Error(w, fmt.Sprintf(`{"error":{"code":"rate_limit_exceeded","message":"Rate limit exceeded. Retry after %v"}}`, result.RetryAfter), http.StatusTooManyRequests)
			return
		}

		// Add rate limit headers even if not limited
		if result != nil {
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rpm))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))
		}

		// Store API key info in context
		ctx := context.WithValue(req.Context(), APIKeyIDContextKey, apiKey.ID)
		ctx = context.WithValue(ctx, APIKeyHashContextKey, keyHash)
		ctx = context.WithValue(ctx, APIKeyTenantIDContextKey, apiKey.TenantID)
		ctx = context.WithValue(ctx, APIKeyUserIDContextKey, apiKey.UserID)
		ctx = context.WithValue(ctx, APIKeyTypeContextKey, apiKey.KeyType)
		ctx = context.WithValue(ctx, APIKeyNameContextKey, apiKey.Name)

		// Update last used timestamp (async, don't block)
		go func() {
			_ = r.repo.UpdateLastUsed(context.Background(), apiKey.ID)
		}()

		// Call next handler with updated context
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// extractAPIKey extracts the API key from the request
// Supports both "Authorization: ApiKey <key>" and "X-API-Key: <key>" headers
func (r *RateLimiter) extractAPIKey(req *http.Request) string {
	// Check X-API-Key header first
	if key := req.Header.Get("X-API-Key"); key != "" {
		return strings.TrimSpace(key)
	}

	// Check Authorization header
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	// Support "ApiKey <key>" format
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "ApiKey") {
		return strings.TrimSpace(parts[1])
	}

	return ""
}

// GetAPIKeyFromContext retrieves the API key ID from the context
func GetAPIKeyFromContext(ctx context.Context) (uuid.UUID, bool) {
	keyID, ok := ctx.Value(APIKeyIDContextKey).(uuid.UUID)
	return keyID, ok
}

// GetTenantIDFromContext retrieves the tenant ID from the context
func GetTenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value(APIKeyTenantIDContextKey).(uuid.UUID)
	return tenantID, ok
}

// GetUserIDFromContext retrieves the user ID from the context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(APIKeyUserIDContextKey).(uuid.UUID)
	return userID, ok
}

// GetAPIKeyTypeFromContext retrieves the API key type from the context
func GetAPIKeyTypeFromContext(ctx context.Context) (KeyType, bool) {
	keyType, ok := ctx.Value(APIKeyTypeContextKey).(KeyType)
	return keyType, ok
}

// GetAPIKeyNameFromContext retrieves the API key name from the context
func GetAPIKeyNameFromContext(ctx context.Context) (string, bool) {
	keyName, ok := ctx.Value(APIKeyNameContextKey).(string)
	return keyName, ok
}

// GetAPIKeyHashFromContext retrieves the API key hash from the context
func GetAPIKeyHashFromContext(ctx context.Context) (string, bool) {
	keyHash, ok := ctx.Value(APIKeyHashContextKey).(string)
	return keyHash, ok
}
