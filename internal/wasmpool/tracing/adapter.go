// Package tracing: adapter that wraps the orchestrator's internal/tracing
// package to satisfy the wasm.Tracer interface. This keeps the wasm module
// decoupled from the orchestrator's tracing implementation.
package tracing

import (
	"context"

	wasmpool "github.com/functionfly/functionfly/internal/wasm"
	"github.com/functionfly/functionfly/internal/tracing"
)

// TracerAdapter wraps internal/tracing to satisfy wasmpool.Tracer.
type TracerAdapter struct{}

// New returns a wasmpool.Tracer backed by internal/tracing.
func New() wasmpool.Tracer { return TracerAdapter{} }

// StartSpan opens a tracing span and returns the augmented context plus
// a wasmpool.Span that calls Finish/SetAttribute/RecordError on the
// underlying internal/tracing primitives.
func (TracerAdapter) StartSpan(ctx context.Context, name string) (context.Context, wasmpool.Span) {
	newCtx, _ := tracing.StartSpan(ctx, name)
	return newCtx, &spanAdapter{ctx: newCtx}
}

type spanAdapter struct {
	ctx context.Context
}

func (s *spanAdapter) Finish() {
	tracing.Finish(s.ctx)
}

func (s *spanAdapter) SetAttribute(key string, value any) {
	tracing.SetAttribute(s.ctx, key, value)
}

func (s *spanAdapter) RecordError(err error) {
	tracing.RecordError(s.ctx, err)
}
