package support

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RateLimiter provides rate limiting for support operations
type RateLimiter struct {
	// Message rate limiting per conversation
	messageLimits map[uuid.UUID]*messageRateLimit
	// User rate limiting for new conversations
	userLimits map[uuid.UUID]*userRateLimit
	// Global rate limiting
	globalLimit *globalRateLimit

	logger *logrus.Logger
}

// messageRateLimit tracks message rate for a conversation
type messageRateLimit struct {
	count     int
	windowStart time.Time
}

// userRateLimit tracks user rate for creating conversations
type userRateLimit struct {
	conversations int
	windowStart   time.Time
}

// globalRateLimit tracks global message rate
type globalRateLimit struct {
	count       int
	windowStart time.Time
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// Messages per conversation per window
	MessagesPerConversation int
	// Conversation creation per user per window
	ConversationsPerUser int
	// Global messages per second
	GlobalMessagesPerSecond int
	// Window duration
	WindowDuration time.Duration
}

// DefaultRateLimitConfig returns the default rate limit configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		MessagesPerConversation: 20,       // 20 messages per conversation per minute
		ConversationsPerUser:   5,        // 5 new conversations per user per minute
		GlobalMessagesPerSecond: 100,      // 100 messages globally per second
		WindowDuration:          time.Minute,
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimitConfig, logger *logrus.Logger) *RateLimiter {
	if logger == nil {
		logger = logrus.New()
	}
	if config == nil {
		config = DefaultRateLimitConfig()
	}
	return &RateLimiter{
		messageLimits: make(map[uuid.UUID]*messageRateLimit),
		userLimits:    make(map[uuid.UUID]*userRateLimit),
		globalLimit:   &globalRateLimit{windowStart: time.Now()},
		logger:        logger,
	}
}

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
	RetryAfter time.Duration
}

// CheckMessageRate checks if a message is allowed under rate limits
func (r *RateLimiter) CheckMessageRate(ctx context.Context, conversationID, userID uuid.UUID) *RateLimitResult {
	config := DefaultRateLimitConfig()
	now := time.Now()

	// Check global limit
	if r.shouldResetGlobal(now, config) {
		r.globalLimit.count = 0
		r.globalLimit.windowStart = now
	}
	if r.globalLimit.count >= config.GlobalMessagesPerSecond {
		return &RateLimitResult{
			Allowed:     false,
			Remaining:   0,
			ResetAt:     r.globalLimit.windowStart.Add(config.WindowDuration),
			RetryAfter:  config.WindowDuration - now.Sub(r.globalLimit.windowStart),
		}
	}
	r.globalLimit.count++

	// Check conversation limit
	convLimit, exists := r.messageLimits[conversationID]
	if !exists || r.shouldResetWindow(now, convLimit.windowStart, config) {
		convLimit = &messageRateLimit{windowStart: now}
		r.messageLimits[conversationID] = convLimit
	}
	if convLimit.count >= config.MessagesPerConversation {
		return &RateLimitResult{
			Allowed:     false,
			Remaining:   0,
			ResetAt:     convLimit.windowStart.Add(config.WindowDuration),
			RetryAfter:  config.WindowDuration - now.Sub(convLimit.windowStart),
		}
	}
	convLimit.count++

	return &RateLimitResult{
		Allowed:   true,
		Remaining: config.MessagesPerConversation - convLimit.count,
		ResetAt:   convLimit.windowStart.Add(config.WindowDuration),
	}
}

// CheckConversationRate checks if a new conversation is allowed
func (r *RateLimiter) CheckConversationRate(ctx context.Context, userID uuid.UUID) *RateLimitResult {
	config := DefaultRateLimitConfig()
	now := time.Now()

	userLimit, exists := r.userLimits[userID]
	if !exists || r.shouldResetWindow(now, userLimit.windowStart, config) {
		userLimit = &userRateLimit{windowStart: now}
		r.userLimits[userID] = userLimit
	}
	if userLimit.conversations >= config.ConversationsPerUser {
		return &RateLimitResult{
			Allowed:     false,
			Remaining:   0,
			ResetAt:     userLimit.windowStart.Add(config.WindowDuration),
			RetryAfter:  config.WindowDuration - now.Sub(userLimit.windowStart),
		}
	}
	userLimit.conversations++

	return &RateLimitResult{
		Allowed:   true,
		Remaining: config.ConversationsPerUser - userLimit.conversations,
		ResetAt:   userLimit.windowStart.Add(config.WindowDuration),
	}
}

// shouldResetWindow checks if a rate limit window has expired
func (r *RateLimiter) shouldResetWindow(now, windowStart time.Time, config *RateLimitConfig) bool {
	return now.Sub(windowStart) >= config.WindowDuration
}

// shouldResetGlobal checks if the global rate limit window has expired
func (r *RateLimiter) shouldResetGlobal(now time.Time, config *RateLimitConfig) bool {
	return now.Sub(r.globalLimit.windowStart) >= config.WindowDuration
}

// Cleanup removes expired rate limit entries
func (r *RateLimiter) Cleanup(ctx context.Context) {
	now := time.Now()
	config := DefaultRateLimitConfig()

	for convID, limit := range r.messageLimits {
		if r.shouldResetWindow(now, limit.windowStart, config) {
			delete(r.messageLimits, convID)
		}
	}
	for userID, limit := range r.userLimits {
		if r.shouldResetWindow(now, limit.windowStart, config) {
			delete(r.userLimits, userID)
		}
	}
}

// AuditLogger provides audit logging for support operations
type AuditLogger struct {
	logger *logrus.Logger
}

// AuditEvent represents an audit event
type AuditEvent struct {
	Timestamp     time.Time
	EventType     string
	ConversationID uuid.UUID
	UserID        uuid.UUID
	StaffID       *uuid.UUID
	Action        string
	Details       map[string]interface{}
	IPAddress     string
	UserAgent     string
	Success       bool
	ErrorMessage  string
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *logrus.Logger) *AuditLogger {
	if logger == nil {
		logger = logrus.New()
	}
	return &AuditLogger{logger: logger}
}

// LogConversationCreated logs a conversation creation event
func (a *AuditLogger) LogConversationCreated(ctx context.Context, conv *SupportConversation, ipAddress, userAgent string) {
	a.logEvent(ctx, &AuditEvent{
		Timestamp:     time.Now(),
		EventType:     "conversation_created",
		ConversationID: conv.ID,
		UserID:        conv.UserID,
		Action:        "create",
		Details: map[string]interface{}{
			"type":      conv.Type,
			"priority":  conv.Priority,
			"is_emergency": conv.IsEmergency,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
	})
}

// LogMessageSent logs a message sent event
func (a *AuditLogger) LogMessageSent(ctx context.Context, msg *SupportMessage, ipAddress, userAgent string) {
	a.logEvent(ctx, &AuditEvent{
		Timestamp:     time.Now(),
		EventType:     "message_sent",
		ConversationID: msg.ConversationID,
		UserID:        msg.AuthorID,
		Action:        "send_message",
		Details: map[string]interface{}{
			"message_id":   msg.ID,
			"message_type": msg.MessageType,
			"author_type":  msg.AuthorType,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
	})
}

// LogEscalation logs an escalation event
func (a *AuditLogger) LogEscalation(ctx context.Context, convID, userID uuid.UUID, staffID *uuid.UUID, reason string) {
	a.logEvent(ctx, &AuditEvent{
		Timestamp:     time.Now(),
		EventType:     "escalation",
		ConversationID: convID,
		UserID:        userID,
		StaffID:       staffID,
		Action:        "escalate",
		Details: map[string]interface{}{
			"reason": reason,
		},
		Success: true,
	})
}

// LogEmergencyRequest logs an emergency fix request
func (a *AuditLogger) LogEmergencyRequest(ctx context.Context, req *EmergencyFixRequest, ipAddress, userAgent string) {
	a.logEvent(ctx, &AuditEvent{
		Timestamp:     time.Now(),
		EventType:     "emergency_request",
		ConversationID: req.ConversationID,
		UserID:        req.UserID,
		StaffID:       req.StaffID,
		Action:        "emergency_request",
		Details: map[string]interface{}{
			"emergency_id": req.ID,
			"status":      req.Status,
			"function_id": req.FunctionID,
			"reason":      req.Reason,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
	})
}

// LogResolution logs a conversation resolution
func (a *AuditLogger) LogResolution(ctx context.Context, convID, resolvedBy uuid.UUID, note string) {
	a.logEvent(ctx, &AuditEvent{
		Timestamp:     time.Now(),
		EventType:     "resolution",
		ConversationID: convID,
		UserID:        resolvedBy,
		Action:        "resolve",
		Details: map[string]interface{}{
			"note": note,
		},
		Success: true,
	})
}

// LogSecurityEvent logs a security-related event
func (a *AuditLogger) LogSecurityEvent(ctx context.Context, eventType string, userID uuid.UUID, details map[string]interface{}, ipAddress string) {
	a.logEvent(ctx, &AuditEvent{
		Timestamp: time.Now(),
		EventType: eventType,
		UserID:   userID,
		Action:   "security_event",
		Details:  details,
		Success:  true,
	})
	a.logger.WithFields(logrus.Fields{
		"security_event": true,
		"event_type":     eventType,
		"user_id":        userID,
		"ip_address":     ipAddress,
	}).Warn("Security event in support system")
}

// LogRateLimitExceeded logs when rate limits are exceeded
func (a *AuditLogger) LogRateLimitExceeded(ctx context.Context, userID uuid.UUID, limitType string, ipAddress string) {
	a.logEvent(ctx, &AuditEvent{
		Timestamp: time.Now(),
		EventType: "rate_limit_exceeded",
		UserID:   userID,
		Action:   "rate_limit_exceeded",
		Details: map[string]interface{}{
			"limit_type": limitType,
		},
		IPAddress: ipAddress,
		Success:   false,
	})
}

// logEvent logs an audit event
func (a *AuditLogger) logEvent(ctx context.Context, event *AuditEvent) {
	fields := logrus.Fields{
		"audit":           true,
		"event_type":      event.EventType,
		"conversation_id":  event.ConversationID,
		"user_id":         event.UserID,
		"action":          event.Action,
		"ip_address":       event.IPAddress,
		"user_agent":      event.UserAgent,
		"success":         event.Success,
	}

	if event.StaffID != nil {
		fields["staff_id"] = *event.StaffID
	}

	for k, v := range event.Details {
		fields[k] = v
	}

	if event.ErrorMessage != "" {
		fields["error_message"] = event.ErrorMessage
	}

	a.logger.WithFields(fields).Info(fmt.Sprintf("Support audit: %s", event.Action))
}
