package notifications

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

type NotificationRateLimiter struct {
	logger *logrus.Logger

	userLimits map[uuid.UUID]*userNotificationLimit

	requestsPerMinute int
	windowDuration     time.Duration

	mu sync.RWMutex
}

type userNotificationLimit struct {
	count       int
	windowStart time.Time
}

func NewNotificationRateLimiter(requestsPerMinute int, logger *logrus.Logger) *NotificationRateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60
	}
	if logger == nil {
		logger = logrus.New()
	}
	return &NotificationRateLimiter{
		logger:            logger,
		userLimits:        make(map[uuid.UUID]*userNotificationLimit),
		requestsPerMinute: requestsPerMinute,
		windowDuration:    time.Minute,
	}
}

type RateLimitResult struct {
	Allowed    bool
	Remaining  int
	ResetAt    time.Time
	RetryAfter time.Duration
}

func (r *NotificationRateLimiter) CheckRateLimit(userID uuid.UUID) *RateLimitResult {
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	limit, exists := r.userLimits[userID]
	if !exists || now.Sub(limit.windowStart) >= r.windowDuration {
		r.userLimits[userID] = &userNotificationLimit{
			count:       1,
			windowStart: now,
		}
		return &RateLimitResult{
			Allowed:   true,
			Remaining: r.requestsPerMinute - 1,
			ResetAt:   now.Add(r.windowDuration),
		}
	}

	if limit.count >= r.requestsPerMinute {
		retryAfter := r.windowDuration - now.Sub(limit.windowStart)
		return &RateLimitResult{
			Allowed:     false,
			Remaining:   0,
			ResetAt:     limit.windowStart.Add(r.windowDuration),
			RetryAfter: retryAfter,
		}
	}

	limit.count++
	return &RateLimitResult{
		Allowed:   true,
		Remaining: r.requestsPerMinute - limit.count,
		ResetAt:   limit.windowStart.Add(r.windowDuration),
	}
}

func (r *NotificationRateLimiter) CleanupExpired() {
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()

	for userID, limit := range r.userLimits {
		if now.Sub(limit.windowStart) >= r.windowDuration {
			delete(r.userLimits, userID)
		}
	}
}

func (r *NotificationRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		userID := getUserIDFromContext(req)
		if userID == uuid.Nil {
			next.ServeHTTP(w, req)
			return
		}

		result := r.CheckRateLimit(userID)

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(r.requestsPerMinute))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

		if !result.Allowed {
			retryAfterSec := int(result.RetryAfter.Seconds())
			if retryAfterSec < 1 {
				retryAfterSec = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfterSec))
			apierror.WriteError(w, apierror.NewRateLimited("Rate limit exceeded. Please try again later."))
			r.logger.WithFields(logrus.Fields{
				"user_id":     userID.String(),
				"retry_after": retryAfterSec,
			}).Warn("Notification API rate limit exceeded")
			return
		}

		next.ServeHTTP(w, req)
	})
}

func getUserIDFromContext(req *http.Request) uuid.UUID {
	if user, ok := req.Context().Value(userContextKey{}).(*UserClaims); ok && user != nil {
		return user.UserID
	}
	return uuid.Nil
}

type UserClaims struct {
	UserID uuid.UUID
}

type userContextKey struct{}

func init() {
	_ = fmt.Sprintf("")
}
