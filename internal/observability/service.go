// Package observability provides production-ready observability services
package observability

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Service provides observability services
type Service struct {
	config *Config
	logger *logrus.Logger
}

// NewService creates a new observability service
func NewService(config *Config) *Service {
	logger := logrus.New()

	// Configure logger based on config
	if config.StructuredLoggingEnabled {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	}

	// Set log level
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return &Service{
		config: config,
		logger: logger,
	}
}

// GetLogger returns the configured logger
func (s *Service) GetLogger() *logrus.Logger {
	return s.logger
}

// LogHTTPRequest logs an HTTP request with structured fields
func (s *Service) LogHTTPRequest(r *http.Request, statusCode int, duration time.Duration, requestID string) {
	fields := logrus.Fields{
		"request_id":    requestID,
		"method":        r.Method,
		"path":          r.URL.Path,
		"query":         r.URL.RawQuery,
		"status_code":   statusCode,
		"duration_ms":   duration.Milliseconds(),
		"duration_human": duration.String(),
		"remote_addr":   r.RemoteAddr,
		"user_agent":    r.UserAgent(),
		"content_type":  r.Header.Get("Content-Type"),
	}

	// Add tenant/app context if available
	if tenantID := r.Header.Get("X-Tenant-Id"); tenantID != "" {
		fields["tenant_id"] = tenantID
	}
	if appID := r.Header.Get("X-App-Id"); appID != "" {
		fields["app_id"] = appID
	}

	// Log based on status code
	if statusCode >= 500 {
		s.logger.WithFields(fields).Error("HTTP request completed with server error")
	} else if statusCode >= 400 {
		s.logger.WithFields(fields).Warn("HTTP request completed with client error")
	} else {
		s.logger.WithFields(fields).Info("HTTP request completed")
	}
}

// LogRateLimitExceeded logs a rate limit exceeded event
func (s *Service) LogRateLimitExceeded(r *http.Request, limitType string) {
	s.logger.WithFields(logrus.Fields{
		"request_id":  r.Header.Get("X-Request-Id"),
		"method":      r.Method,
		"path":        r.URL.Path,
		"remote_addr": r.RemoteAddr,
		"limit_type":  limitType,
	}).Warn("Rate limit exceeded")
}

// LogSecurityEvent logs a security event
func (s *Service) LogSecurityEvent(eventType string, r *http.Request, details string) {
	s.logger.WithFields(logrus.Fields{
		"event_type":  eventType,
		"request_id":  r.Header.Get("X-Request-Id"),
		"method":      r.Method,
		"path":        r.URL.Path,
		"remote_addr": r.RemoteAddr,
		"user_agent":  r.UserAgent(),
		"details":     details,
	}).Warn("Security event")
}

// LogHealthCheck logs a health check event
func (s *Service) LogHealthCheck(checkName string, healthy bool, duration time.Duration) {
	fields := logrus.Fields{
		"check_name":    checkName,
		"healthy":       healthy,
		"duration_ms":   duration.Milliseconds(),
		"duration_human": duration.String(),
	}

	if healthy {
		s.logger.WithFields(fields).Info("Health check passed")
	} else {
		s.logger.WithFields(fields).Error("Health check failed")
	}
}

// LogCircuitBreakerTransition logs a circuit breaker state transition
func (s *Service) LogCircuitBreakerTransition(backendID, fromState, toState string) {
	s.logger.WithFields(logrus.Fields{
		"backend_id": backendID,
		"from_state": fromState,
		"to_state":   toState,
		"component":  "circuit_breaker",
	}).Info("Circuit breaker state transition")
}

// LogRoutingDecision logs a routing decision
func (s *Service) LogRoutingDecision(appID, backendID, reason string, latency time.Duration) {
	s.logger.WithFields(logrus.Fields{
		"app_id":      appID,
		"backend_id":  backendID,
		"reason":      reason,
		"latency_ms":  latency.Milliseconds(),
		"component":   "routing",
	}).Info("Routing decision")
}

// WithContext returns a logger with context fields
func (s *Service) WithContext(ctx context.Context) *logrus.Entry {
	fields := logrus.Fields{}

	// Add request ID from context if available
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}

	// Add trace ID from context if available
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields["trace_id"] = traceID
	}

	// Add span ID from context if available
	if spanID := ctx.Value("span_id"); spanID != nil {
		fields["span_id"] = spanID
	}

	return s.logger.WithFields(fields)
}

// WithFields returns a logger with additional fields
func (s *Service) WithFields(fields logrus.Fields) *logrus.Entry {
	return s.logger.WithFields(fields)
}
