package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type traceIDKeyType string

const (
	TraceIDKey  traceIDKeyType = "trace_id"
	SpanIDKey   traceIDKeyType = "span_id"
	ParentIDKey traceIDKeyType = "parent_span_id"

	traceparentHeader = "traceparent"
	tracestateHeader  = "tracestate"
	ffTraceIDHeader   = "X-FunctionFly-Trace-ID"
	ffSpanIDHeader    = "X-FunctionFly-Span-ID"
)

func generateID(byteLen int) string {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func parseTraceparent(header string) (traceID, parentID string, sampled bool) {
	if len(header) < 55 {
		return "", "", false
	}
	parts := splitBy(header, '-')
	if len(parts) != 4 {
		return "", "", false
	}
	return parts[1], parts[2], parts[3] == "01"
}

func splitBy(s string, sep byte) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	return append(parts, s[start:])
}

func formatTraceparent(traceID, spanID string, sampled bool) string {
	flags := "00"
	if sampled {
		flags = "01"
	}
	return fmt.Sprintf("00-%s-%s-%s", traceID, spanID, flags)
}

// TracingMiddleware adds W3C Trace Context distributed tracing to all requests.
// It extracts or generates trace IDs, propagates them via response headers,
// and stores them in the request context for downstream use.
func TracingMiddleware(next http.Handler) http.Handler {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "functionfly-api"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		var traceID, parentID string
		sampled := true

		if tp := r.Header.Get(traceparentHeader); tp != "" {
			traceID, parentID, sampled = parseTraceparent(tp)
		}
		if traceID == "" {
			if ffTrace := r.Header.Get(ffTraceIDHeader); ffTrace != "" {
				traceID = ffTrace
			}
		}
		if traceID == "" {
			traceID = generateID(16)
		}
		spanID := generateID(8)

		ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
		ctx = context.WithValue(ctx, SpanIDKey, spanID)
		if parentID != "" {
			ctx = context.WithValue(ctx, ParentIDKey, parentID)
		}

		w.Header().Set(traceparentHeader, formatTraceparent(traceID, spanID, sampled))
		w.Header().Set(ffTraceIDHeader, traceID)
		w.Header().Set(ffSpanIDHeader, spanID)
		if ts := r.Header.Get(tracestateHeader); ts != "" {
			w.Header().Set(tracestateHeader, ts)
		}

		wrapped := &tracingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		logrus.WithFields(logrus.Fields{
			"trace_id":    traceID,
			"span_id":     spanID,
			"parent_id":   parentID,
			"service":     serviceName,
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": wrapped.statusCode,
			"duration_ms": time.Since(startTime).Milliseconds(),
		}).Debug("request trace span")
	})
}

type tracingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *tracingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// GetTraceID extracts the trace ID from context
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(TraceIDKey).(string); ok {
		return id
	}
	return ""
}

// GetSpanID extracts the span ID from context
func GetSpanID(ctx context.Context) string {
	if id, ok := ctx.Value(SpanIDKey).(string); ok {
		return id
	}
	return ""
}

// InjectTraceHeaders injects trace context headers into an outgoing HTTP request
func InjectTraceHeaders(ctx context.Context, req *http.Request) {
	traceID := GetTraceID(ctx)
	spanID := GetSpanID(ctx)
	if traceID != "" && spanID != "" {
		req.Header.Set(traceparentHeader, formatTraceparent(traceID, spanID, true))
		req.Header.Set(ffTraceIDHeader, traceID)
		req.Header.Set(ffSpanIDHeader, spanID)
	}
}

// TraceFields returns logrus fields for the current trace context
func TraceFields(ctx context.Context) logrus.Fields {
	return logrus.Fields{
		"trace_id": GetTraceID(ctx),
		"span_id":  GetSpanID(ctx),
	}
}
