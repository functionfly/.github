//go:build !cgo

// Package wasm: stub for streaming types and DefaultStreamingConfig when CGO is disabled.
package wasm

import (
	"context"
	"time"
)

// DefaultChunkSize stub
const DefaultChunkSize = 64 * 1024

// MaxStreamingChunks stub
const MaxStreamingChunks = 100

// StreamingConfig stub (same shape as cgo version)
type StreamingConfig struct {
	ChunkSize         int
	MaxTotalSize      int
	OutputBufferSize  int
	EnableBackpressure bool
}

// DefaultStreamingConfig returns default config when CGO is disabled
func DefaultStreamingConfig() *StreamingConfig {
	return &StreamingConfig{
		ChunkSize:          DefaultChunkSize,
		MaxTotalSize:       10 * 1024 * 1024,
		OutputBufferSize:   1024 * 1024,
		EnableBackpressure: true,
	}
}

// StreamingExecutionResult stub (same shape as cgo version)
type StreamingExecutionResult struct {
	Output      []byte
	TotalChunks int
	TotalBytes  int
	Duration    time.Duration
	Error       error
}

// ExecuteLargeInput returns an error when CGO is disabled
func (r *PythonRuntime) ExecuteLargeInput(ctx context.Context, input []byte, config *StreamingConfig) (*StreamingExecutionResult, error) {
	return nil, errWasmNotAvailable
}
