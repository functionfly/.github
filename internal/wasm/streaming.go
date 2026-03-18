//go:build cgo

// Package wasm provides WebAssembly runtime support for FunctionFly
// This file contains streaming execution implementation
package wasm

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

// DefaultChunkSize is the default size for streaming chunks
const DefaultChunkSize = 64 * 1024 // 64KB

// MaxStreamingChunks is the maximum number of chunks for a single streaming execution
const MaxStreamingChunks = 100

// StreamingConfig contains configuration for streaming execution
type StreamingConfig struct {
	// ChunkSize is the size of each chunk in bytes
	ChunkSize int

	// MaxTotalSize is the maximum total input size in bytes
	MaxTotalSize int

	// OutputBufferSize is the size of the output buffer
	OutputBufferSize int

	// EnableBackpressure enables backpressure when output buffer is full
	EnableBackpressure bool
}

// DefaultStreamingConfig returns default streaming configuration
func DefaultStreamingConfig() *StreamingConfig {
	return &StreamingConfig{
		ChunkSize:        DefaultChunkSize,
		MaxTotalSize:     10 * 1024 * 1024, // 10MB
		OutputBufferSize: 1024 * 1024,       // 1MB
		EnableBackpressure: true,
	}
}

// StreamingExecutor handles streaming execution of WASM functions
type StreamingExecutor struct {
	runtime    *PythonRuntime
	config     *StreamingConfig
	outputChan chan []byte
	doneChan   chan struct{}
	mu         sync.Mutex
	closed     bool
}

// NewStreamingExecutor creates a new streaming executor
func NewStreamingExecutor(runtime *PythonRuntime, config *StreamingConfig) *StreamingExecutor {
	if config == nil {
		config = DefaultStreamingConfig()
	}

	return &StreamingExecutor{
		runtime:    runtime,
		config:     config,
		outputChan: make(chan []byte, config.OutputBufferSize),
		doneChan:   make(chan struct{}),
	}
}

// StreamInput represents a chunk of input data
type StreamInput struct {
	ChunkID int
	Data    []byte
	IsLast  bool
}

// StreamOutput represents a chunk of output data
type StreamOutput struct {
	ChunkID int
	Data    []byte
	IsLast  bool
	Error   error
}

// ExecuteStreaming executes a WASM function with streaming input/output
func (s *StreamingExecutor) ExecuteStreaming(ctx context.Context, input io.Reader) (<-chan StreamOutput, error) {
	if s == nil || s.runtime == nil {
		return nil, fmt.Errorf("executor not initialized")
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, fmt.Errorf("executor already closed")
	}
	s.mu.Unlock()

	// Reset channels
	s.outputChan = make(chan []byte, s.config.OutputBufferSize)
	s.doneChan = make(chan struct{})

	// Start the streaming execution
	go s.processStreamingInput(ctx, input)

	// Create output channel
	outputChan := make(chan StreamOutput, s.config.OutputBufferSize)

	// Start output reader
	go s.streamOutputToChan(outputChan)

	return outputChan, nil
}

// processStreamingInput processes input in chunks and sends to WASM
func (s *StreamingExecutor) processStreamingInput(ctx context.Context, input io.Reader) {
	defer close(s.doneChan)

	chunkID := 0
	totalSize := 0
	buffer := make([]byte, s.config.ChunkSize)

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Read chunk
		n, err := io.ReadFull(input, buffer)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			if n > 0 {
				// Send last chunk
				chunk := make([]byte, n)
				copy(chunk, buffer[:n])
				s.sendChunk(chunkID, chunk, true)
			}
			break
		}
		if err != nil {
			return
		}

		totalSize += n

		// Check max total size
		if totalSize > s.config.MaxTotalSize {
			return
		}

		// Create chunk
		chunk := make([]byte, n)
		copy(chunk, buffer[:n])
		isLast := false

		// Send chunk
		s.sendChunk(chunkID, chunk, isLast)
		chunkID++

		// Check max chunks
		if chunkID >= MaxStreamingChunks {
			return
		}

		// Small delay to prevent overwhelming the runtime
		time.Sleep(time.Millisecond)
	}
}

// sendChunk sends a chunk to the WASM runtime
func (s *StreamingExecutor) sendChunk(chunkID int, data []byte, isLast bool) {
	// In a real implementation, this would send to WASM
	// For now, we'll simulate the streaming behavior
	select {
	case s.outputChan <- data:
	case <-time.After(time.Second):
		// Timeout handling
	}
}

// streamOutputToChan streams output to the output channel
func (s *StreamingExecutor) streamOutputToChan(outputChan chan<- StreamOutput) {
	defer close(outputChan)

	chunkID := 0

	for {
		select {
		case data, ok := <-s.outputChan:
			if !ok {
				return
			}
			outputChan <- StreamOutput{
				ChunkID: chunkID,
				Data:    data,
				IsLast:  false,
			}
			chunkID++
		case <-s.doneChan:
			return
		}
	}
}

// Close closes the streaming executor
func (s *StreamingExecutor) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	if s.outputChan != nil {
		close(s.outputChan)
	}
	if s.doneChan != nil {
		close(s.doneChan)
	}

	return nil
}

// StreamingExecutionResult contains the result of a streaming execution
type StreamingExecutionResult struct {
	Output     []byte
	TotalChunks int
	TotalBytes  int
	Duration   time.Duration
	Error      error
}

// ExecuteLargeInput executes a WASM function with large input using streaming
func (r *PythonRuntime) ExecuteLargeInput(ctx context.Context, input []byte, config *StreamingConfig) (*StreamingExecutionResult, error) {
	if config == nil {
		config = DefaultStreamingConfig()
	}

	startTime := time.Now()

	// For large inputs, we process in chunks
	// This is useful for memory-efficient handling

	chunkSize := config.ChunkSize
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}

	// If input is small enough, execute normally
	if len(input) <= chunkSize {
		output, err := r.ExecuteWithContext(ctx, input)
		return &StreamingExecutionResult{
			Output:     output,
			TotalChunks: 1,
			TotalBytes:  len(input),
			Duration:   time.Since(startTime),
			Error:      err,
		}, err
	}

	// For large inputs, we simulate chunked processing
	// In a real implementation, this would send chunks to WASM

	totalChunks := (len(input) + chunkSize - 1) / chunkSize
	if totalChunks > MaxStreamingChunks {
		return nil, fmt.Errorf("input too large: %d chunks exceeds maximum of %d", totalChunks, MaxStreamingChunks)
	}

	for i := 0; i < totalChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(input) {
			end = len(input)
		}

		chunk := input[start:end]
		isLast := (i == totalChunks-1)

		// Process chunk (in real implementation, send to WASM)
		// For now, we'll just track progress
		_ = chunk
		_ = isLast
	}

	// Execute the full input in the runtime
	output, err := r.ExecuteWithContext(ctx, input)

	return &StreamingExecutionResult{
		Output:     output,
		TotalChunks: totalChunks,
		TotalBytes:  len(input),
		Duration:   time.Since(startTime),
		Error:      err,
	}, err
}

// CreateStreamingReader creates an io.Reader for streaming input
func CreateStreamingReader(data []byte, chunkSize int) *StreamingReader {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	return &StreamingReader{
		data:      data,
		chunkSize: chunkSize,
		pos:       0,
	}
}

// StreamingReader implements io.Reader for chunked reading
type StreamingReader struct {
	data      []byte
	chunkSize int
	pos       int
}

// Read implements io.Reader
func (r *StreamingReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	remaining := len(r.data) - r.pos
	toRead := r.chunkSize
	if toRead > len(p) {
		toRead = len(p)
	}
	if toRead > remaining {
		toRead = remaining
	}

	copy(p[:toRead], r.data[r.pos:r.pos+toRead])
	r.pos += toRead

	return toRead, nil
}

// StreamingWriter collects streamed output into a buffer
type StreamingWriter struct {
	buffer   []byte
	mu       sync.Mutex
	closed   bool
}

// NewStreamingWriter creates a new streaming writer
func NewStreamingWriter() *StreamingWriter {
	return &StreamingWriter{
		buffer: make([]byte, 0, 4096),
	}
}

// Write implements io.Writer
func (w *StreamingWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, fmt.Errorf("writer closed")
	}

	w.buffer = append(w.buffer, p...)
	return len(p), nil
}

// Bytes returns the accumulated output
func (w *StreamingWriter) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()

	result := make([]byte, len(w.buffer))
	copy(result, w.buffer)
	return result
}

// Close marks the writer as closed
func (w *StreamingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.closed = true
	return nil
}
