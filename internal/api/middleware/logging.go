package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// LoggingMiddleware provides structured logging for HTTP requests
type LoggingMiddleware struct{}

// NewLoggingMiddleware creates a new logging middleware instance
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

// StructuredLogger adds structured logging to HTTP requests
func (lm *LoggingMiddleware) StructuredLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate or extract request ID
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set("X-Request-Id", requestID)
		}

		// Add request ID to response headers
		w.Header().Set("X-Request-Id", requestID)

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Create logger with request context
		logger := logrus.WithFields(logrus.Fields{
			"request_id":    requestID,
			"method":        r.Method,
			"path":          r.URL.Path,
			"query":         r.URL.RawQuery,
			"user_agent":    r.Header.Get("User-Agent"),
			"remote_addr":   getClientIP(r),
			"content_type":  r.Header.Get("Content-Type"),
			"content_length": r.ContentLength,
		})

		// Add tenant/app context if available from URL or headers
		if tenantID := r.Header.Get("X-Tenant-Id"); tenantID != "" {
			logger = logger.WithField("tenant_id", tenantID)
		}
		if appID := r.Header.Get("X-App-Id"); appID != "" {
			logger = logger.WithField("app_id", appID)
		}

		// Log request start
		logger.Info("HTTP request started")

		// Call next handler
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start)

		// Log request completion with response details
		logger.WithFields(logrus.Fields{
			"status_code":    rw.statusCode,
			"duration_ms":    duration.Milliseconds(),
			"duration_human": duration.String(),
			"response_size":  rw.size,
		}).Info("HTTP request completed")
	}
}

// RequestIDMiddleware ensures every request has a request ID
func (lm *LoggingMiddleware) RequestIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to both request and response headers
		r.Header.Set("X-Request-Id", requestID)
		w.Header().Set("X-Request-Id", requestID)

		// Add request ID to request context for use in handlers
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}
}

// GetRequestLogger returns a logger with request context
func GetRequestLogger(r *http.Request) *logrus.Entry {
	fields := logrus.Fields{}

	// Add request ID if available
	if requestID := r.Header.Get("X-Request-Id"); requestID != "" {
		fields["request_id"] = requestID
	} else if ctxRequestID := r.Context().Value("request_id"); ctxRequestID != nil {
		fields["request_id"] = ctxRequestID
	}

	// Add other common fields
	fields["method"] = r.Method
	fields["path"] = r.URL.Path

	return logrus.WithFields(fields)
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.size += int64(size)
	return size, err
}