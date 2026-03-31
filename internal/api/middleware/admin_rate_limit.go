package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	// AdminRateLimitWindow is the time window for rate limiting
	AdminRateLimitWindow = 1 * time.Minute
	// AdminRateLimitMaxRequests is the maximum requests per window for admin routes
	AdminRateLimitMaxRequests = 100
	// AdminSensitiveRateLimitMaxRequests is the maximum requests per window for sensitive operations
	AdminSensitiveRateLimitMaxRequests = 10
	// LoginRateLimitMaxRequests is the maximum login attempts per window
	LoginRateLimitMaxRequests = 5
	// LoginRateLimitWindow is the time window for login rate limiting
	LoginRateLimitWindow = 15 * time.Minute
	// LoginRateLimitBlockDuration is how long to block after exceeding limits
	LoginRateLimitBlockDuration = 1 * time.Hour
)

const (
	// RateLimitKeyPrefix is the Redis key prefix for rate limiting
	RateLimitKeyPrefix = "ratelimit:admin:"
	// LoginRateLimitKeyPrefix is the Redis key prefix for login rate limiting
	LoginRateLimitKeyPrefix = "ratelimit:login:"
	// BlockedKeyPrefix is the Redis key prefix for blocked IPs
	BlockedKeyPrefix = "blocked:ip:"
)

// RateLimitResult contains the result of a rate limit check
type RateLimitResult struct {
	Allowed    bool
	Remaining  int
	RetryAfter int // Seconds until retry is allowed (0 if not blocked)
	TotalLimit int
	ResetAt    time.Time
}

// inMemoryRateEntry tracks requests in a sliding window for in-memory rate limiting.
type inMemoryRateEntry struct {
	mu         sync.Mutex
	timestamps []time.Time
}

// inMemoryStore provides a process-local rate limit fallback when Redis is unavailable.
type inMemoryStore struct {
	mu      sync.RWMutex
	entries map[string]*inMemoryRateEntry
	blocked map[string]time.Time
}

func newInMemoryStore() *inMemoryStore {
	return &inMemoryStore{
		entries: make(map[string]*inMemoryRateEntry),
		blocked: make(map[string]time.Time),
	}
}

func (s *inMemoryStore) getOrCreate(key string) *inMemoryRateEntry {
	s.mu.RLock()
	e, ok := s.entries[key]
	s.mu.RUnlock()
	if ok {
		return e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok = s.entries[key]; ok {
		return e
	}
	e = &inMemoryRateEntry{}
	s.entries[key] = e
	return e
}

func (s *inMemoryStore) isBlocked(ip string) (bool, time.Duration) {
	s.mu.RLock()
	expiry, ok := s.blocked[ip]
	s.mu.RUnlock()
	if !ok {
		return false, 0
	}
	remaining := time.Until(expiry)
	if remaining <= 0 {
		s.mu.Lock()
		delete(s.blocked, ip)
		s.mu.Unlock()
		return false, 0
	}
	return true, remaining
}

func (s *inMemoryStore) block(ip string, duration time.Duration) {
	s.mu.Lock()
	s.blocked[ip] = time.Now().Add(duration)
	s.mu.Unlock()
}

// AdminRateLimiter handles rate limiting for admin endpoints
type AdminRateLimiter struct {
	redisClient *redis.Client
	memStore    *inMemoryStore
	logger      *logrus.Entry
}

// NewAdminRateLimiter creates a new admin rate limiter
func NewAdminRateLimiter(redisClient *redis.Client) *AdminRateLimiter {
	return &AdminRateLimiter{
		redisClient: redisClient,
		memStore:    newInMemoryStore(),
		logger:      logrus.WithField("middleware", "admin_ratelimit"),
	}
}

// isProduction returns true when PRODUCTION_ENV=true.
func isProduction() bool {
	return os.Getenv("PRODUCTION_ENV") == "true"
}

// isBlocked checks if an IP is currently blocked
func (m *AdminRateLimiter) isBlocked(ctx context.Context, ip string) (bool, time.Duration, error) {
	if m.redisClient == nil {
		// In-memory fallback
		blocked, retryAfter := m.memStore.isBlocked(ip)
		return blocked, retryAfter, nil
	}

	key := BlockedKeyPrefix + ip
	ttl, err := m.redisClient.TTL(ctx, key).Result()
	if err != nil {
		m.logger.WithError(err).Error("Failed to check blocked status")
		// Fail closed in production: block on error
		if isProduction() {
			return true, 1 * time.Minute, err
		}
		return false, 0, err
	}

	if ttl > 0 {
		return true, ttl, nil
	}
	return false, 0, nil
}

// blockIP blocks an IP address for the specified duration
func (m *AdminRateLimiter) blockIP(ctx context.Context, ip string, duration time.Duration) error {
	if m.redisClient == nil {
		m.memStore.block(ip, duration)
		return nil
	}

	key := BlockedKeyPrefix + ip
	if err := m.redisClient.Set(ctx, key, "1", duration).Err(); err != nil {
		m.logger.WithError(err).Error("Failed to block IP")
		return err
	}

	m.logger.WithFields(logrus.Fields{
		"ip":       ip,
		"duration": duration,
	}).Warn("IP address blocked due to rate limit exceeded")

	return nil
}

// CheckRateLimit checks if a request should be rate limited
// Returns the result with allowed status and remaining requests
func (m *AdminRateLimiter) CheckRateLimit(ctx context.Context, key string, limit int) (*RateLimitResult, error) {
	if m.redisClient == nil {
		// In-memory sliding window fallback
		entry := m.memStore.getOrCreate(key)
		entry.mu.Lock()
		defer entry.mu.Unlock()

		now := time.Now()
		cutoff := now.Add(-AdminRateLimitWindow)
		// Evict expired timestamps
		valid := entry.timestamps[:0]
		for _, t := range entry.timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		entry.timestamps = valid

		if len(entry.timestamps) >= limit {
			return &RateLimitResult{
				Allowed:    false,
				Remaining:  0,
				TotalLimit: limit,
				ResetAt:    now.Add(AdminRateLimitWindow),
			}, nil
		}

		entry.timestamps = append(entry.timestamps, now)
		return &RateLimitResult{
			Allowed:    true,
			Remaining:  limit - len(entry.timestamps),
			TotalLimit: limit,
			ResetAt:    now.Add(AdminRateLimitWindow),
		}, nil
	}

	now := time.Now()
	windowStart := now.Add(-AdminRateLimitWindow).UnixMilli()
	currentTime := float64(now.UnixNano()) / float64(time.Millisecond)

	// Use a Redis sorted set for sliding window rate limiting
	redisKey := RateLimitKeyPrefix + key

	// Remove old entries outside the window
	if err := m.redisClient.ZRemRangeByScore(ctx, redisKey, "-inf", strconv.FormatInt(windowStart, 10)).Err(); err != nil {
		m.logger.WithError(err).Error("Failed to remove old rate limit entries")
		return nil, err
	}

	// Count current requests in this window
	count, err := m.redisClient.ZCard(ctx, redisKey).Result()
	if err != nil {
		m.logger.WithError(err).Error("Failed to count rate limit entries")
		return nil, err
	}

	remaining := limit - int(count) - 1
	if remaining < 0 {
		remaining = 0
	}

	resetAt := now.Add(AdminRateLimitWindow)

	// Check if limit exceeded
	if count >= int64(limit) {
		// Get oldest entry to calculate retry-after
		oldest, err := m.redisClient.ZRange(ctx, redisKey, 0, 0).Result()
		retryAfter := AdminRateLimitWindow
		if err == nil && len(oldest) > 0 {
			if oldestTime, err := strconv.ParseFloat(oldest[0], 64); err == nil {
				retryAfter = time.Until(time.Unix(0, int64(oldestTime)*int64(time.Millisecond))) + AdminRateLimitWindow
			}
		}

		m.logger.WithFields(logrus.Fields{
			"key":       key,
			"count":     count,
			"limit":     limit,
			"retry_sec": int(retryAfter.Seconds()),
		}).Info("Rate limit exceeded")

		return &RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: int(retryAfter.Seconds()),
			TotalLimit: limit,
			ResetAt:    resetAt,
		}, nil
	}

	// Add current request to the window
	if err := m.redisClient.ZAdd(ctx, redisKey, redis.Z{
		Score:  currentTime,
		Member: fmt.Sprintf("%.0f", currentTime),
	}).Err(); err != nil {
		m.logger.WithError(err).Error("Failed to add rate limit entry")
		return nil, err
	}

	// Set expiry on the key
	if err := m.redisClient.Expire(ctx, redisKey, AdminRateLimitWindow+time.Second).Err(); err != nil {
		m.logger.WithError(err).Error("Failed to set rate limit key expiry")
	}

	return &RateLimitResult{
		Allowed:    true,
		Remaining:  remaining,
		TotalLimit: limit,
		ResetAt:    resetAt,
	}, nil
}

// RecordRequest records a request for rate limiting
func (m *AdminRateLimiter) RecordRequest(ctx context.Context, key string, limit int) error {
	result, err := m.CheckRateLimit(ctx, key, limit)
	if err != nil {
		return err
	}
	if !result.Allowed {
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}

// CheckLoginRateLimit checks login-specific rate limiting
func (m *AdminRateLimiter) CheckLoginRateLimit(ctx context.Context, ip string) (*RateLimitResult, error) {
	if m.redisClient == nil {
		// In-memory sliding window fallback for login rate limiting
		key := LoginRateLimitKeyPrefix + ip
		entry := m.memStore.getOrCreate(key)
		entry.mu.Lock()
		defer entry.mu.Unlock()

		now := time.Now()
		cutoff := now.Add(-LoginRateLimitWindow)
		valid := entry.timestamps[:0]
		for _, t := range entry.timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		entry.timestamps = valid

		if len(entry.timestamps) >= LoginRateLimitMaxRequests {
			m.memStore.block(ip, LoginRateLimitBlockDuration)
			return &RateLimitResult{
				Allowed:    false,
				Remaining:  0,
				RetryAfter: int(LoginRateLimitBlockDuration.Seconds()),
				TotalLimit: LoginRateLimitMaxRequests,
				ResetAt:    now.Add(LoginRateLimitWindow),
			}, nil
		}

		entry.timestamps = append(entry.timestamps, now)
		return &RateLimitResult{
			Allowed:    true,
			Remaining:  LoginRateLimitMaxRequests - len(entry.timestamps),
			TotalLimit: LoginRateLimitMaxRequests,
			ResetAt:    now.Add(LoginRateLimitWindow),
		}, nil
	}

	key := LoginRateLimitKeyPrefix + ip
	now := time.Now()
	windowStart := now.Add(-LoginRateLimitWindow).UnixMilli()
	currentTime := float64(now.UnixNano()) / float64(time.Millisecond)

	// Remove old entries outside the window
	if err := m.redisClient.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart, 10)).Err(); err != nil {
		m.logger.WithError(err).Error("Failed to remove old login rate limit entries")
		return nil, err
	}

	// Count current attempts in this window
	count, err := m.redisClient.ZCard(ctx, key).Result()
	if err != nil {
		m.logger.WithError(err).Error("Failed to count login rate limit entries")
		return nil, err
	}

	remaining := LoginRateLimitMaxRequests - int(count) - 1
	if remaining < 0 {
		remaining = 0
	}

	resetAt := now.Add(LoginRateLimitWindow)

	// Check if limit exceeded
	if count >= int64(LoginRateLimitMaxRequests) {
		// Block this IP
		if err := m.blockIP(ctx, ip, LoginRateLimitBlockDuration); err != nil {
			m.logger.WithError(err).Error("Failed to block IP after login rate limit exceeded")
		}

		return &RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: int(LoginRateLimitBlockDuration.Seconds()),
			TotalLimit: LoginRateLimitMaxRequests,
			ResetAt:    resetAt,
		}, nil
	}

	// Add current attempt to the window
	if err := m.redisClient.ZAdd(ctx, key, redis.Z{
		Score:  currentTime,
		Member: fmt.Sprintf("%.0f", currentTime),
	}).Err(); err != nil {
		m.logger.WithError(err).Error("Failed to add login rate limit entry")
		return nil, err
	}

	// Set expiry on the key
	if err := m.redisClient.Expire(ctx, key, LoginRateLimitWindow+time.Second).Err(); err != nil {
		m.logger.WithError(err).Error("Failed to set login rate limit key expiry")
	}

	return &RateLimitResult{
		Allowed:    true,
		Remaining:  remaining,
		TotalLimit: LoginRateLimitMaxRequests,
		ResetAt:    resetAt,
	}, nil
}

// RequireRateLimit is middleware that enforces rate limiting on admin routes
func (m *AdminRateLimiter) RequireRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract client IP
		clientIP := extractClientIPFromRequest(r)

		// Check if IP is blocked
		blocked, retryAfter, err := m.isBlocked(ctx, clientIP)
		if err != nil {
			m.logger.WithError(err).WithField("ip", clientIP).Error("Failed to check if IP is blocked")
			// On error, fail closed in production
			if isProduction() {
				writeRateLimitError(w, "Rate limit check failed", 60)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		if blocked {
			m.logger.WithFields(logrus.Fields{
				"ip":          clientIP,
				"path":        r.URL.Path,
				"retry_after": retryAfter.Seconds(),
			}).Warn("Blocked IP attempted request")

			writeRateLimitError(w, "Too many requests", int(retryAfter.Seconds()))
			return
		}

		// Use IP-based rate limiting for general admin routes
		key := "ip:" + clientIP
		result, err := m.CheckRateLimit(ctx, key, AdminRateLimitMaxRequests)
		if err != nil {
			m.logger.WithError(err).WithField("ip", clientIP).Error("Failed to check rate limit")
			// On error, fail closed in production
			if isProduction() {
				writeRateLimitError(w, "Rate limit check failed", 60)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.TotalLimit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

		if !result.Allowed {
			m.logger.WithFields(logrus.Fields{
				"ip":          clientIP,
				"path":        r.URL.Path,
				"retry_after": result.RetryAfter,
			}).Warn("Rate limit exceeded for IP")

			w.Header().Set("Retry-After", strconv.Itoa(result.RetryAfter))
			writeRateLimitError(w, "Too many requests", result.RetryAfter)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// RequireSensitiveRateLimit is middleware for sensitive operations (e.g., password changes, session revocation)
func (m *AdminRateLimiter) RequireSensitiveRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract client IP
		clientIP := extractClientIPFromRequest(r)

		// Check if IP is blocked
		blocked, retryAfter, err := m.isBlocked(ctx, clientIP)
		if err != nil {
			m.logger.WithError(err).WithField("ip", clientIP).Error("Failed to check if IP is blocked")
			if isProduction() {
				writeRateLimitError(w, "Rate limit check failed", 60)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		if blocked {
			m.logger.WithFields(logrus.Fields{
				"ip":          clientIP,
				"path":        r.URL.Path,
				"retry_after": retryAfter.Seconds(),
			}).Warn("Blocked IP attempted sensitive operation")

			writeRateLimitError(w, "Too many requests", int(retryAfter.Seconds()))
			return
		}

		// Use IP-based rate limiting with stricter limits
		key := "sensitive:ip:" + clientIP
		result, err := m.CheckRateLimit(ctx, key, AdminSensitiveRateLimitMaxRequests)
		if err != nil {
			m.logger.WithError(err).WithField("ip", clientIP).Error("Failed to check sensitive rate limit")
			if isProduction() {
				writeRateLimitError(w, "Rate limit check failed", 60)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.TotalLimit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

		if !result.Allowed {
			m.logger.WithFields(logrus.Fields{
				"ip":          clientIP,
				"path":        r.URL.Path,
				"retry_after": result.RetryAfter,
			}).Warn("Sensitive operation rate limit exceeded")

			w.Header().Set("Retry-After", strconv.Itoa(result.RetryAfter))
			writeRateLimitError(w, "Too many requests", result.RetryAfter)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// RequireLoginRateLimit is middleware for login endpoints
func (m *AdminRateLimiter) RequireLoginRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract client IP
		clientIP := extractClientIPFromRequest(r)

		// Check if IP is blocked
		blocked, retryAfter, err := m.isBlocked(ctx, clientIP)
		if err != nil {
			m.logger.WithError(err).WithField("ip", clientIP).Error("Failed to check if IP is blocked")
			if isProduction() {
				writeRateLimitError(w, "Rate limit check failed", 60)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		if blocked {
			m.logger.WithFields(logrus.Fields{
				"ip":          clientIP,
				"path":        r.URL.Path,
				"retry_after": retryAfter.Seconds(),
			}).Warn("Blocked IP attempted login")

			writeRateLimitError(w, "Too many requests", int(retryAfter.Seconds()))
			return
		}

		// Check login-specific rate limiting
		result, err := m.CheckLoginRateLimit(ctx, clientIP)
		if err != nil {
			m.logger.WithError(err).WithField("ip", clientIP).Error("Failed to check login rate limit")
			if isProduction() {
				writeRateLimitError(w, "Rate limit check failed", 60)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.TotalLimit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

		if !result.Allowed {
			m.logger.WithFields(logrus.Fields{
				"ip":          clientIP,
				"path":        r.URL.Path,
				"retry_after": result.RetryAfter,
			}).Warn("Login rate limit exceeded, IP blocked")

			w.Header().Set("Retry-After", strconv.Itoa(result.RetryAfter))
			writeRateLimitError(w, "Too many requests", result.RetryAfter)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// writeRateLimitError writes a rate limit error response
func writeRateLimitError(w http.ResponseWriter, message string, retryAfter int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	w.WriteHeader(http.StatusTooManyRequests)
	w.Write([]byte(fmt.Sprintf(`{"error":"rate_limit_exceeded","message":"%s","retry_after":%d}`, message, retryAfter)))
}

// extractClientIPFromRequest extracts the client IP address from the request
func extractClientIPFromRequest(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
