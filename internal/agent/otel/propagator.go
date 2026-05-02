package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Propagator struct {
	tracer trace.Tracer
}

func NewPropagator(serviceName string) *Propagator {
	return &Propagator{
		tracer: otel.Tracer(serviceName),
	}
}

func (p *Propagator) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return p.tracer.Start(ctx, name, opts...)
}

func (p *Propagator) InjectTraceContext(ctx context.Context, carrier map[string]string) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(carrier))
}

func (p *Propagator) ExtractTraceContext(ctx context.Context, carrier map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(carrier))
}

func (p *Propagator) AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

func (p *Propagator) RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}

func (p *Propagator) EndSpan(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.End()
}

func (p *Propagator) GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

func (p *Propagator) GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasSpanID() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

func FormatTraceParent(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	if !sc.IsValid() {
		return ""
	}
	return fmt.Sprintf("%s-%s-%s-%s",
		"00",
		sc.TraceID().String(),
		sc.SpanID().String(),
		"01",
	)
}

func InjectW3CTraceParent(ctx context.Context, carrier map[string]string) {
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.MapCarrier(carrier))
}

func ExtractW3CTraceParent(ctx context.Context, carrier map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(carrier))
}