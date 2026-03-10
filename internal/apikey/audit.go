// Package apikey provides API key generation, hashing, validation, and audit logging functionality.
package apikey

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

/*
 * Audit Event Types
 * =================
 * These event types track all API key operations for security and compliance.
 */
const (
	// API_KEY_CREATED - When a new API key is created
	API_KEY_CREATED = "api_key_created"

	// API_KEY_VIEWED - When key details are viewed (masked)
	API_KEY_VIEWED = "api_key_viewed"

	// API_KEY_UPDATED - When key properties are updated
	API_KEY_UPDATED = "api_key_updated"

	// API_KEY_DELETED - When key is revoked/deleted
	API_KEY_DELETED = "api_key_deleted"

	// API_KEY_ROTATED - When key is rotated
	API_KEY_ROTATED = "api_key_rotated"

	// API_KEY_AUTHENTICATED - When key is used for authentication
	API_KEY_AUTHENTICATED = "api_key_authenticated"

	// API_KEY_RATE_LIMITED - When key hits rate limit
	API_KEY_RATE_LIMITED = "api_key_rate_limited"
)

/*
 * Prometheus Metrics
 * ==================
 * Track API key audit events for monitoring and alerting.
 */
var (
	apiKeyAuditEvents = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_api_key_audit_events_total",
			Help: "Total number of API key audit events by event type",
		},
		[]string{"event_type", "tenant_id", "success"},
	)

	apiKeyAuditDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_api_key_audit_duration_seconds",
			Help:    "Duration of API key audit log operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"event_type"},
	)
)

/*
 * APIKeyAudit struct
 * ==================
 * Represents an API key audit event for logging and compliance.
 */
type APIKeyAudit struct {
	// Event type (one of the constants above)
	EventType string `json:"event_type"`

	// Tenant ID associated with the API key
	TenantID *uuid.UUID `json:"tenant_id,omitempty"`

	// User ID who performed the action
	UserID *uuid.UUID `json:"user_id,omitempty"`

	// API Key ID
	APIKeyID uuid.UUID `json:"api_key_id"`

	// API Key Name (masked for privacy)
	APIKeyName string `json:"api_key_name"`

	// API Key Hash (first 8 chars for identification)
	KeyHash string `json:"key_hash"`

	// IP Address of the request
	IPAddress string `json:"ip_address,omitempty"`

	// User Agent of the request
	UserAgent string `json:"user_agent,omitempty"`

	// Additional metadata (JSON)
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Success status
	Success bool `json:"success"`

	// Failure reason if unsuccessful
	FailureReason string `json:"failure_reason,omitempty"`

	// Timestamp of the event
	Timestamp time.Time `json:"timestamp"`
}

// ToMap converts the audit struct to a map for storage
func (a *APIKeyAudit) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"event_type":     a.EventType,
		"tenant_id":      a.TenantID,
		"user_id":        a.UserID,
		"api_key_id":     a.APIKeyID,
		"api_key_name":   a.APIKeyName,
		"key_hash":       a.KeyHash,
		"ip_address":     a.IPAddress,
		"user_agent":     a.UserAgent,
		"metadata":       a.Metadata,
		"success":        a.Success,
		"failure_reason": a.FailureReason,
		"timestamp":      a.Timestamp,
	}
}

// MaskKeyName returns a masked version of the key name for privacy
func MaskKeyName(name string) string {
	if len(name) <= 4 {
		return "****"
	}
	return name[:2] + "**" + name[len(name)-2:]
}

// TruncateHash returns the first n characters of the key hash
func TruncateHash(hash string, n int) string {
	if len(hash) <= n {
		return hash
	}
	return hash[:n]
}

/*
 * Logger Interface
 * ================
 * Defines the contract for logging API key audit events.
 */
type Logger interface {
	// LogAPIKeyEvent logs an API key audit event
	LogAPIKeyEvent(ctx context.Context, event *APIKeyAudit) error

	// LogAPIKeyCreated logs a key creation event
	LogAPIKeyCreated(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, metadata map[string]interface{}) error

	// LogAPIKeyViewed logs a key view event
	LogAPIKeyViewed(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, ipAddress, userAgent string) error

	// LogAPIKeyUpdated logs a key update event
	LogAPIKeyUpdated(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, metadata map[string]interface{}) error

	// LogAPIKeyDeleted logs a key deletion/revocation event
	LogAPIKeyDeleted(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, metadata map[string]interface{}) error

	// LogAPIKeyRotated logs a key rotation event
	LogAPIKeyRotated(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, reason string, metadata map[string]interface{}) error

	// LogAPIKeyAuthenticated logs a key authentication event
	LogAPIKeyAuthenticated(ctx context.Context, tenantID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, success bool, failureReason string) error

	// LogAPIKeyRateLimited logs a rate limiting event
	LogAPIKeyRateLimited(ctx context.Context, tenantID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, metadata map[string]interface{}) error
}

/*
 * AuditLogger
 * ===========
 * Default implementation of the Logger interface.
 */
type AuditLogger struct {
	logger logrus.FieldLogger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger logrus.FieldLogger) *AuditLogger {
	return &AuditLogger{
		logger: logger,
	}
}

// LogAPIKeyEvent logs an API key audit event
func (l *AuditLogger) LogAPIKeyEvent(ctx context.Context, event *APIKeyAudit) error {
	startTime := time.Now()

	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Log to structured logger
	metadataJSON, _ := json.Marshal(event.Metadata)
	l.logger.WithFields(logrus.Fields{
		"audit_event_type": event.EventType,
		"tenant_id":        event.TenantID,
		"user_id":          event.UserID,
		"api_key_id":       event.APIKeyID,
		"api_key_name":     event.APIKeyName,
		"key_hash":         event.KeyHash,
		"ip_address":       event.IPAddress,
		"user_agent":       event.UserAgent,
		"metadata":         string(metadataJSON),
		"success":          event.Success,
		"failure_reason":   event.FailureReason,
		"timestamp":        event.Timestamp,
	}).Info("API Key Audit Event")

	// Update metrics
	l.recordMetrics(event)

	// Record duration
	duration := time.Since(startTime)
	apiKeyAuditDuration.WithLabelValues(event.EventType).Observe(duration.Seconds())

	return nil
}

// recordMetrics updates Prometheus metrics for the audit event
func (l *AuditLogger) recordMetrics(event *APIKeyAudit) {
	tenantLabel := "unknown"
	if event.TenantID != nil {
		tenantLabel = event.TenantID.String()[:8]
	}

	successLabel := "true"
	if !event.Success {
		successLabel = "false"
	}

	apiKeyAuditEvents.WithLabelValues(event.EventType, tenantLabel, successLabel).Inc()
}

// LogAPIKeyCreated logs a key creation event
func (l *AuditLogger) LogAPIKeyCreated(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, metadata map[string]interface{}) error {
	event := &APIKeyAudit{
		EventType:  API_KEY_CREATED,
		TenantID:   tenantID,
		UserID:     userID,
		APIKeyID:   keyID,
		APIKeyName: MaskKeyName(keyName),
		KeyHash:    TruncateHash(keyHash, 8),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Metadata:   metadata,
		Success:    true,
		Timestamp:  time.Now(),
	}

	return l.LogAPIKeyEvent(ctx, event)
}

// LogAPIKeyViewed logs a key view event
func (l *AuditLogger) LogAPIKeyViewed(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, ipAddress, userAgent string) error {
	event := &APIKeyAudit{
		EventType:  API_KEY_VIEWED,
		TenantID:   tenantID,
		UserID:     userID,
		APIKeyID:   keyID,
		APIKeyName: MaskKeyName(keyName),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Success:    true,
		Timestamp:  time.Now(),
	}

	return l.LogAPIKeyEvent(ctx, event)
}

// LogAPIKeyUpdated logs a key update event
func (l *AuditLogger) LogAPIKeyUpdated(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, metadata map[string]interface{}) error {
	event := &APIKeyAudit{
		EventType:  API_KEY_UPDATED,
		TenantID:   tenantID,
		UserID:     userID,
		APIKeyID:   keyID,
		APIKeyName: MaskKeyName(keyName),
		KeyHash:    TruncateHash(keyHash, 8),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Metadata:   metadata,
		Success:    true,
		Timestamp:  time.Now(),
	}

	return l.LogAPIKeyEvent(ctx, event)
}

// LogAPIKeyDeleted logs a key deletion/revocation event
func (l *AuditLogger) LogAPIKeyDeleted(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, metadata map[string]interface{}) error {
	event := &APIKeyAudit{
		EventType:  API_KEY_DELETED,
		TenantID:   tenantID,
		UserID:     userID,
		APIKeyID:   keyID,
		APIKeyName: MaskKeyName(keyName),
		KeyHash:    TruncateHash(keyHash, 8),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Metadata:   metadata,
		Success:    true,
		Timestamp:  time.Now(),
	}

	return l.LogAPIKeyEvent(ctx, event)
}

// LogAPIKeyRotated logs a key rotation event
func (l *AuditLogger) LogAPIKeyRotated(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, reason string, metadata map[string]interface{}) error {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["rotation_reason"] = reason

	event := &APIKeyAudit{
		EventType:  API_KEY_ROTATED,
		TenantID:   tenantID,
		UserID:     userID,
		APIKeyID:   keyID,
		APIKeyName: MaskKeyName(keyName),
		KeyHash:    TruncateHash(keyHash, 8),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Metadata:   metadata,
		Success:    true,
		Timestamp:  time.Now(),
	}

	return l.LogAPIKeyEvent(ctx, event)
}

// LogAPIKeyAuthenticated logs a key authentication event
func (l *AuditLogger) LogAPIKeyAuthenticated(ctx context.Context, tenantID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, success bool, failureReason string) error {
	event := &APIKeyAudit{
		EventType:     API_KEY_AUTHENTICATED,
		TenantID:      tenantID,
		APIKeyID:      keyID,
		APIKeyName:    MaskKeyName(keyName),
		KeyHash:       TruncateHash(keyHash, 8),
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		Success:       success,
		FailureReason: failureReason,
		Timestamp:     time.Now(),
	}

	return l.LogAPIKeyEvent(ctx, event)
}

// LogAPIKeyRateLimited logs a rate limiting event
func (l *AuditLogger) LogAPIKeyRateLimited(ctx context.Context, tenantID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, metadata map[string]interface{}) error {
	event := &APIKeyAudit{
		EventType:     API_KEY_RATE_LIMITED,
		TenantID:      tenantID,
		APIKeyID:      keyID,
		APIKeyName:    MaskKeyName(keyName),
		KeyHash:       TruncateHash(keyHash, 8),
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		Metadata:      metadata,
		Success:       false,
		FailureReason: "rate_limit_exceeded",
		Timestamp:     time.Now(),
	}

	return l.LogAPIKeyEvent(ctx, event)
}

/*
 * NoOpAuditLogger
 * ===============
 * A no-op implementation for cases where audit logging is not needed.
 */
type NoOpAuditLogger struct{}

// NewNoOpAuditLogger creates a no-op audit logger
func NewNoOpAuditLogger() *NoOpAuditLogger {
	return &NoOpAuditLogger{}
}

// LogAPIKeyEvent is a no-op
func (l *NoOpAuditLogger) LogAPIKeyEvent(ctx context.Context, event *APIKeyAudit) error {
	return nil
}

// LogAPIKeyCreated is a no-op
func (l *NoOpAuditLogger) LogAPIKeyCreated(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, metadata map[string]interface{}) error {
	return nil
}

// LogAPIKeyViewed is a no-op
func (l *NoOpAuditLogger) LogAPIKeyViewed(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, ipAddress, userAgent string) error {
	return nil
}

// LogAPIKeyUpdated is a no-op
func (l *NoOpAuditLogger) LogAPIKeyUpdated(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, metadata map[string]interface{}) error {
	return nil
}

// LogAPIKeyDeleted is a no-op
func (l *NoOpAuditLogger) LogAPIKeyDeleted(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, metadata map[string]interface{}) error {
	return nil
}

// LogAPIKeyRotated is a no-op
func (l *NoOpAuditLogger) LogAPIKeyRotated(ctx context.Context, tenantID, userID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, reason string, metadata map[string]interface{}) error {
	return nil
}

// LogAPIKeyAuthenticated is a no-op
func (l *NoOpAuditLogger) LogAPIKeyAuthenticated(ctx context.Context, tenantID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, success bool, failureReason string) error {
	return nil
}

// LogAPIKeyRateLimited is a no-op
func (l *NoOpAuditLogger) LogAPIKeyRateLimited(ctx context.Context, tenantID *uuid.UUID, keyID uuid.UUID, keyName, keyHash, ipAddress, userAgent string, metadata map[string]interface{}) error {
	return nil
}

/*
 * Audit Logger Factory
 * ===================
 * Helper functions for creating audit loggers.
 */

// NewAuditLoggerFromContext creates an audit logger from a context
// Falls back to no-op if no logger is found in context
func NewAuditLoggerFromContext(ctx context.Context) Logger {
	// Try to get logger from context first
	if logger, ok := ctx.Value(AuditLoggerContextKey).(Logger); ok {
		return logger
	}

	// Default to no-op
	return NewNoOpAuditLogger()
}

// AuditLoggerContextKey is the context key for storing the audit logger
type AuditLoggerContextKeyType string

const (
	// AuditLoggerContextKey is the context key for the audit logger
	AuditLoggerContextKey AuditLoggerContextKeyType = "audit_logger"
)

// WithAuditLogger returns a context with the audit logger set
func WithAuditLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, AuditLoggerContextKey, logger)
}

/*
 * Helper Functions
 * ================
 */

// GetClientIP extracts the client IP from the request headers
func GetClientIP(xForwardedFor, xRealIP, remoteAddr string) string {
	// Check X-Forwarded-For first (for proxies)
	if xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		// Format: "client, proxy1, proxy2"
		for _, ip := range splitIPs(xForwardedFor) {
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP
	if xRealIP != "" {
		return xRealIP
	}

	// Fall back to remote addr
	return remoteAddr
}

// splitIPs splits a comma-separated list of IPs
func splitIPs(s string) []string {
	var result []string
	var current string

	for _, c := range s {
		if c == ',' {
			result = append(result, trimSpace(current))
			current = ""
		} else {
			current += string(c)
		}
	}

	result = append(result, trimSpace(current))
	return result
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}

	return s[start:end]
}

// FormatMetadataForAudit formats metadata for audit logging
func FormatMetadataForAudit(data map[string]interface{}) string {
	if data == nil {
		return "{}"
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}

	return string(jsonBytes)
}

// AuditMetadataBuilder helps build audit metadata
type AuditMetadataBuilder struct {
	data map[string]interface{}
}

// NewAuditMetadataBuilder creates a new metadata builder
func NewAuditMetadataBuilder() *AuditMetadataBuilder {
	return &AuditMetadataBuilder{
		data: make(map[string]interface{}),
	}
}

// With adds a key-value pair to the metadata
func (b *AuditMetadataBuilder) With(key string, value interface{}) *AuditMetadataBuilder {
	b.data[key] = value
	return b
}

// WithDescription adds a description to the metadata
func (b *AuditMetadataBuilder) WithDescription(desc string) *AuditMetadataBuilder {
	b.data["description"] = desc
	return b
}

// WithKeyType adds the key type to the metadata
func (b *AuditMetadataBuilder) WithKeyType(keyType KeyType) *AuditMetadataBuilder {
	b.data["key_type"] = keyType
	return b
}

// WithPermissions adds permissions to the metadata
func (b *AuditMetadataBuilder) WithPermissions(perms []Permission) *AuditMetadataBuilder {
	permStrs := make([]string, len(perms))
	for i, p := range perms {
		permStrs[i] = string(p)
	}
	b.data["permissions"] = permStrs
	return b
}

// WithExpiry adds expiry info to the metadata
func (b *AuditMetadataBuilder) WithExpiry(expiresAt *time.Time) *AuditMetadataBuilder {
	if expiresAt != nil {
		b.data["expires_at"] = expiresAt.Format(time.RFC3339)
		b.data["expires_in_days"] = time.Until(*expiresAt).Hours() / 24
	}
	return b
}

// WithRateLimit adds rate limit info to the metadata
func (b *AuditMetadataBuilder) WithRateLimit(rpm, rph, rpd int) *AuditMetadataBuilder {
	b.data["rate_limit_rpm"] = rpm
	b.data["rate_limit_rph"] = rph
	b.data["rate_limit_rpd"] = rpd
	return b
}

// Build returns the metadata map
func (b *AuditMetadataBuilder) Build() map[string]interface{} {
	return b.data
}

// String returns a JSON string representation
func (b *AuditMetadataBuilder) String() string {
	return FormatMetadataForAudit(b.data)
}

// StringMap converts metadata to a string map for logging
func StringMap(data map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range data {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}
