package tracing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TraceContext holds distributed tracing information
type TraceContext struct {
	// TraceID is the unique identifier for the entire trace
	TraceID string
	// SpanID is the unique identifier for this span
	SpanID string
	// ParentSpanID is the parent span ID (empty for root spans)
	ParentSpanID string
	// Operation is the name of the operation being traced
	Operation string
	// StartTime is when the span started
	StartTime time.Time
	// Attributes are key-value pairs for the span
	Attributes map[string]interface{}
	// Events are timestamped events within the span
	Events []SpanEvent
}

// SpanEvent represents a timestamped event within a span
type SpanEvent struct {
	Name       string
	Timestamp  time.Time
	Attributes map[string]interface{}
}

// contextKey is a typed key for trace context
type contextKey string

const traceContextKey contextKey = "trace_context"

// StartSpan starts a new trace span
func StartSpan(ctx context.Context, operation string) (context.Context, *TraceContext) {
	// Check if there's already a trace context
	if existing := FromContext(ctx); existing != nil {
		// Create child span
		span := &TraceContext{
			TraceID:      existing.TraceID,
			SpanID:       uuid.New().String(),
			ParentSpanID: existing.SpanID,
			Operation:    operation,
			StartTime:    time.Now(),
			Attributes:   make(map[string]interface{}),
			Events:       make([]SpanEvent, 0),
		}
		return context.WithValue(ctx, traceContextKey, span), span
	}

	// Create root span
	span := &TraceContext{
		TraceID:    uuid.New().String(),
		SpanID:     uuid.New().String(),
		Operation:  operation,
		StartTime:  time.Now(),
		Attributes: make(map[string]interface{}),
		Events:     make([]SpanEvent, 0),
	}

	return context.WithValue(ctx, traceContextKey, span), span
}

// FromContext extracts trace context from context
func FromContext(ctx context.Context) *TraceContext {
	if span, ok := ctx.Value(traceContextKey).(*TraceContext); ok {
		return span
	}
	return nil
}

// SetAttribute sets an attribute on the current span
func SetAttribute(ctx context.Context, key string, value interface{}) {
	if span := FromContext(ctx); span != nil {
		span.Attributes[key] = value
	}
}

// AddEvent adds an event to the current span
func AddEvent(ctx context.Context, name string, attributes map[string]interface{}) {
	if span := FromContext(ctx); span != nil {
		event := SpanEvent{
			Name:       name,
			Timestamp:  time.Now(),
			Attributes: attributes,
		}
		span.Events = append(span.Events, event)
	}
}

// RecordError records an error on the current span
func RecordError(ctx context.Context, err error) {
	if span := FromContext(ctx); span != nil {
		span.Attributes["error"] = true
		span.Attributes["error_message"] = err.Error()
		AddEvent(ctx, "error", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// Finish finishes the current span and logs it
func Finish(ctx context.Context) {
	if span := FromContext(ctx); span != nil {
		duration := time.Since(span.StartTime)

		fields := logrus.Fields{
			"trace_id":       span.TraceID,
			"span_id":        span.SpanID,
			"parent_span_id": span.ParentSpanID,
			"operation":      span.Operation,
			"duration_ms":    duration.Milliseconds(),
		}

		// Add attributes
		for k, v := range span.Attributes {
			fields[k] = v
		}

		// Add events count
		if len(span.Events) > 0 {
			fields["event_count"] = len(span.Events)
		}

		// Log based on error status
		if hasError, ok := span.Attributes["error"].(bool); ok && hasError {
			logrus.WithFields(fields).Error("Span completed with error")
		} else {
			logrus.WithFields(fields).Info("Span completed")
		}
	}
}

// InjectHeaders injects trace context into HTTP headers
func InjectHeaders(ctx context.Context) map[string]string {
	headers := make(map[string]string)

	if span := FromContext(ctx); span != nil {
		headers["X-Trace-ID"] = span.TraceID
		headers["X-Span-ID"] = span.SpanID
		if span.ParentSpanID != "" {
			headers["X-Parent-Span-ID"] = span.ParentSpanID
		}
	}

	return headers
}

// ExtractHeaders extracts trace context from HTTP headers
func ExtractHeaders(ctx context.Context, headers map[string]string) context.Context {
	traceID := headers["X-Trace-ID"]
	spanID := headers["X-Span-ID"]
	parentSpanID := headers["X-Parent-Span-ID"]

	if traceID == "" {
		return ctx
	}

	if spanID == "" {
		spanID = uuid.New().String()
	}

	span := &TraceContext{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
		StartTime:    time.Now(),
		Attributes:   make(map[string]interface{}),
		Events:       make([]SpanEvent, 0),
	}

	return context.WithValue(ctx, traceContextKey, span)
}

// GetTraceID returns the trace ID from context
func GetTraceID(ctx context.Context) string {
	if span := FromContext(ctx); span != nil {
		return span.TraceID
	}
	return ""
}

// GetSpanID returns the span ID from context
func GetSpanID(ctx context.Context) string {
	if span := FromContext(ctx); span != nil {
		return span.SpanID
	}
	return ""
}

// WithTraceContext creates a new context with trace information
func WithTraceContext(ctx context.Context, traceID, spanID, parentSpanID string) context.Context {
	span := &TraceContext{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
		StartTime:    time.Now(),
		Attributes:   make(map[string]interface{}),
		Events:       make([]SpanEvent, 0),
	}

	return context.WithValue(ctx, traceContextKey, span)
}

// FormatTraceID formats a trace ID for logging
func FormatTraceID(traceID string) string {
	if traceID == "" {
		return "no-trace"
	}
	return fmt.Sprintf("trace:%s", traceID)
}
