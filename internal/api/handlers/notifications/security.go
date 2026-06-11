package notifications

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type NotificationRateLimiter struct {
	logger *logrus.Logger

	userLimits map[uuid.UUID]*userNotificationLimit

	requestsPerMinute int
	windowDuration    time.Duration

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

func (r *NotificationRateLimiter) AddRateLimitHeaders(w http.ResponseWriter, result *RateLimitResult) {
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(r.requestsPerMinute))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))
}

func (r *NotificationRateLimiter) LogRateLimitExceeded(userID uuid.UUID, result *RateLimitResult) {
	r.logger.WithFields(logrus.Fields{
		"user_id":     userID.String(),
		"retry_after": int(result.RetryAfter.Seconds()),
	}).Warn("Notification API rate limit exceeded")
}

type NotificationAuditLogger struct {
	logger *logrus.Logger
}

func NewNotificationAuditLogger(logger *logrus.Logger) *NotificationAuditLogger {
	if logger == nil {
		logger = logrus.New()
	}
	return &NotificationAuditLogger{logger: logger}
}

type AuditEvent struct {
	Timestamp     time.Time
	UserID        uuid.UUID
	Action        string
	NotificationID uuid.UUID
	IPAddress     string
	UserAgent     string
	Success       bool
	ErrorMessage  string
}

func (l *NotificationAuditLogger) LogNotificationAccess(ctx context.Context, userID uuid.UUID, action string, notificationID uuid.UUID, ipAddress, userAgent string, success bool, err error) {
	event := AuditEvent{
		Timestamp:      time.Now().UTC(),
		UserID:         userID,
		Action:         action,
		NotificationID: notificationID,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		Success:       success,
	}
	if err != nil {
		event.ErrorMessage = err.Error()
	}

	l.logger.WithFields(logrus.Fields{
		"event":            "notification_access",
		"user_id":          userID.String(),
		"action":           action,
		"notification_id":  notificationID.String(),
		"ip_address":       ipAddress,
		"user_agent":       userAgent,
		"success":          success,
		"error":            event.ErrorMessage,
		"timestamp":        event.Timestamp.Format(time.RFC3339),
	}).Info("Notification audit event")
}

func (l *NotificationAuditLogger) LogPreferenceChange(ctx context.Context, userID uuid.UUID, channel, category string, enabled bool, ipAddress, userAgent string) {
	l.logger.WithFields(logrus.Fields{
		"event":      "preference_change",
		"user_id":    userID.String(),
		"channel":    channel,
		"category":   category,
		"enabled":   enabled,
		"ip_address": ipAddress,
		"user_agent": userAgent,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}).Info("Notification preference changed")
}
