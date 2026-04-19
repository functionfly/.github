//go:build cgo

// Package wasm provides WebAssembly runtime support for FunctionFly
// This file contains streaming execution implementation
package wasm

import (
	"context"
	"fmt"
	"io"
	"log"
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
	runtime      *PythonRuntime
	config       *StreamingConfig
	outputChan   chan []byte
	doneChan     chan struct{}
	inputBuffer  [][]byte
	chunkCount   int
	mu           sync.Mutex
	closed       bool
	execStarted  bool
	execDone     bool
}

// NewStreamingExecutor creates a new streaming executor
func NewStreamingExecutor(runtime *PythonRuntime, config *StreamingConfig) *StreamingExecutor {
	if config == nil {
		config = DefaultStreamingConfig()
	}

	return &StreamingExecutor{
		runtime:     runtime,
		config:      config,
		outputChan:  make(chan []byte, config.OutputBufferSize),
		doneChan:    make(chan struct{}),
		inputBuffer: make([][]byte, 0, MaxStreamingChunks),
		chunkCount:  0,
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
	// Reset state for new execution
	s.inputBuffer = make([][]byte, 0, MaxStreamingChunks)
	s.chunkCount = 0
	s.execStarted = false
	s.execDone = false
	s.closed = false
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

// sendChunk sends a chunk to the WASM runtime for streaming execution
func (s *StreamingExecutor) sendChunk(chunkID int, data []byte, isLast bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Guard against sending after close
	if s.closed {
		return
	}

	// If execution is already done, reject new chunks
	if s.execDone {
		return
	}

	// Write chunk to WASM memory if runtime is available
	if s.runtime != nil && s.runtime.memory != nil {
		if err := s.writeChunkToWasm(chunkID, data, isLast); err != nil {
			s.debugf("Failed to write chunk to WASM: %v", err)
		}
	}

	// Store chunk in input buffer for later execution
	chunkCopy := make([]byte, len(data))
	copy(chunkCopy, data)
	s.inputBuffer = append(s.inputBuffer, chunkCopy)
	s.chunkCount++

	// If this is the last chunk, trigger execution
	if isLast {
		s.execStarted = true
		go s.executeBuffered()
	}
}

// writeChunkToWasm writes a chunk of data to WASM memory for streaming
func (s *StreamingExecutor) writeChunkToWasm(chunkID int, data []byte, isLast bool) error {
	if s.runtime == nil || s.runtime.memory == nil {
		return fmt.Errorf("runtime or memory not available")
	}

	memoryData := s.runtime.memory.UnsafeData(s.runtime.store)
	if memoryData == nil {
		return fmt.Errorf("memory data not accessible")
	}

	// Calculate offset for this chunk (use chunk ID to avoid overwriting)
	// Memory layout: each chunk at offset = 8192 + (chunkID * chunkSize)
	chunkOffset := 8192 + (chunkID * s.runtime.config.GetChunkBufferSize())

	// Ensure we don't exceed memory bounds
	if chunkOffset+len(data) > len(memoryData) {
		return fmt.Errorf("chunk exceeds memory bounds: offset=%d len=%d mem=%d",
			chunkOffset, len(data), len(memoryData))
	}

	// Write chunk data to WASM memory
	copy(memoryData[chunkOffset:], data)

	s.debugf("Wrote chunk %d (%d bytes) to WASM memory at offset %d, isLast=%v",
		chunkID, len(data), chunkOffset, isLast)

	return nil
}

// executeBuffered executes the buffered input when all chunks are received
func (s *StreamingExecutor) executeBuffered() {
	s.mu.Lock()
	if len(s.inputBuffer) == 0 || s.execDone {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	// Concatenate all chunks
	var totalSize int
	for _, chunk := range s.inputBuffer {
		totalSize += len(chunk)
	}

	input := make([]byte, 0, totalSize)
	for _, chunk := range s.inputBuffer {
		input = append(input, chunk...)
	}

	s.debugf("Executing buffered input: %d bytes across %d chunks",
		len(input), len(s.inputBuffer))

	// Execute via the runtime
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := s.runtime.ExecuteWithContext(ctx, input)
	if err != nil {
		s.debugf("Buffered execution failed: %v", err)
		// Send error as output
		errMsg := []byte(fmt.Sprintf(`{"error":"streaming execution failed: %v"}`, err))
		select {
		case s.outputChan <- errMsg:
		case <-time.After(time.Second):
		}
	} else {
		// Send output chunks (stream from the result)
		s.streamOutputChunks(output)
	}

	s.mu.Lock()
	s.execDone = true
	s.mu.Unlock()

	close(s.doneChan)
}

// streamOutputChunks splits output into chunks and sends to output channel
func (s *StreamingExecutor) streamOutputChunks(output []byte) {
	if len(output) == 0 {
		return
	}

	chunkSize := s.config.ChunkSize
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}

	for i := 0; i < len(output); i += chunkSize {
		end := i + chunkSize
		if end > len(output) {
			end = len(output)
		}

		chunk := output[i:end]
		isLast := end >= len(output)

		select {
		case s.outputChan <- chunk:
		case <-time.After(5 * time.Second):
			s.debugf("Timeout sending output chunk")
			return
		}

		s.debugf("Sent output chunk %d (%d bytes), isLast=%v", i/chunkSize, len(chunk), isLast)
	}
}

// debugf logs debug information if the runtime has debug enabled
func (s *StreamingExecutor) debugf(format string, args ...interface{}) {
	if s.runtime != nil && s.runtime.debug {
		log.Printf("[StreamingExecutor] "+format, args...)
	}
}

// streamOutputToChan streams output to the output channel
func (s *StreamingExecutor) streamOutputToChan(outputChan chan<- StreamOutput) {
	defer close(outputChan)

	chunkID := 0
	isLastEmitted := false

	for {
		select {
		case data, ok := <-s.outputChan:
			if !ok {
				// Check if we need to emit the final IsLast chunk
				if !isLastEmitted {
					select {
					case outputChan <- StreamOutput{
						ChunkID: chunkID,
						Data:    []byte{},
						IsLast:  true,
					}:
						isLastEmitted = true
					default:
					}
				}
				return
			}
			isLastEmitted = false // Reset since we got actual data
			outputChan <- StreamOutput{
				ChunkID: chunkID,
				Data:    data,
				IsLast:  false,
			}
			chunkID++
		case <-s.doneChan:
			// Execution is done, drain remaining output
		drainLoop:
			for {
				select {
				case data, ok := <-s.outputChan:
					if !ok {
						break drainLoop
					}
					outputChan <- StreamOutput{
						ChunkID: chunkID,
						Data:    data,
						IsLast:  false,
					}
					chunkID++
				default:
					break drainLoop
				}
			}
			// Emit final IsLast chunk
			outputChan <- StreamOutput{
				ChunkID: chunkID,
				Data:    []byte{},
				IsLast:  true,
			}
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

	// Write chunks to WASM memory and execute with proper streaming state
	totalChunks := (len(input) + chunkSize - 1) / chunkSize
	if totalChunks > MaxStreamingChunks {
		return nil, fmt.Errorf("input too large: %d chunks exceeds maximum of %d", totalChunks, MaxStreamingChunks)
	}

	// Initialize streaming state for this execution
	if r.streamingState != nil {
		r.streamingState.Init()
	}

	chunksWritten := 0
	// Write each chunk to WASM memory at proper offsets
	for i := 0; i < totalChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(input) {
			end = len(input)
		}

		chunk := input[start:end]
		isLast := i == totalChunks-1

		// Write chunk to WASM memory at the offset computed for this chunk ID
		chunkOffset := 8192 + (i * r.config.GetChunkBufferSize())
		if chunkOffset+len(chunk) > len(r.memory.UnsafeData(r.store)) {
			// Memory not large enough; fall back to concatenated execution
			r.debugf("Chunk %d exceeds memory bounds, falling back to concatenated execution", i)
			break
		}

		memoryData := r.memory.UnsafeData(r.store)
		copy(memoryData[chunkOffset:], chunk)

		// Notify WASM module of the chunk via streaming state if supported
		if r.streamingState != nil {
			r.streamingState.AddInputChunk(int32(i), chunk, isLast)
		}
		chunksWritten++

		r.debugf("Wrote chunk %d (%d bytes) to WASM at offset %d, isLast=%v",
			i, len(chunk), chunkOffset, isLast)
	}

	// Execute via the streaming executor if available, otherwise use standard execute
	var output []byte
	var err error

	if r.streamingState != nil && chunksWritten == totalChunks && r.streamingState.IsActive() {
		// Streaming state is ready; use streaming execution path
		execFunc := r.instance.GetExport(r.store, "execute").Func()
		if execFunc == nil {
			return nil, fmt.Errorf("module does not export execute function")
		}

		// For streaming, pass pointer to first chunk and total chunks as metadata
		firstChunkPtr := int32(8192)
		metadataPtr, allocErr := r.allocate(16) // 4 x int32: ptr, len, total_chunks, flags
		if allocErr != nil {
			return nil, fmt.Errorf("failed to allocate metadata: %w", allocErr)
		}

		metadata := []byte{
			byte(firstChunkPtr), byte(firstChunkPtr >> 8), byte(firstChunkPtr >> 16), byte(firstChunkPtr >> 24),
			byte(len(input)), byte(len(input) >> 8), byte(len(input) >> 16), byte(len(input) >> 24),
			byte(totalChunks), byte(totalChunks >> 8), byte(totalChunks >> 16), byte(totalChunks >> 24),
			0, 0, 0, 0, // flags
		}
		_ = r.writeMemory(metadataPtr, metadata)

		result, execErr := execFunc.Call(r.store, metadataPtr, len(metadata))
		if execErr != nil {
			err = fmt.Errorf("streaming execute call failed: %w", execErr)
		} else if resultPtr, ok := result.(int32); ok && resultPtr != 0 {
			maxOut := int(r.config.MaxOutputSize)
			if maxOut <= 0 {
				maxOut = 1024 * 1024 // fallback 1MB
			}
			output, err = r.extractOutputFromResult(resultPtr, maxOut)
		}
	} else {
		// Fall back: execute concatenated input normally
		output, err = r.ExecuteWithContext(ctx, input)
	}

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
