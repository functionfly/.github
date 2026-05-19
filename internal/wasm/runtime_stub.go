//go:build !cgo

package wasm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

var errWasmNotAvailable = errors.New("WASM runtime not available in this build (requires CGO_ENABLED=1); rebuild with CGO for Python/WASM execution")

const (
	DefaultChunkSize   = 64 * 1024
	MaxStreamingChunks = 100
)

type PythonRuntime struct {
	mu         sync.RWMutex
	closed     bool
	wasmPath   string
	config     *WASMSecurityConfig
	execCount  int64
	createdAt  time.Time
}

func NewPythonRuntime(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler) (*PythonRuntime, error) {
	return nil, errWasmNotAvailable
}

func NewPythonRuntimeWithDebug(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler, debug bool) (*PythonRuntime, error) {
	return nil, errWasmNotAvailable
}

func NewPythonRuntimeWithConfig(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler, config *WASMSecurityConfig) (*PythonRuntime, error) {
	return nil, errWasmNotAvailable
}

func NewPythonRuntimeWithConfigAndDebug(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler, config *WASMSecurityConfig, debug bool) (*PythonRuntime, error) {
	return nil, errWasmNotAvailable
}

func (r *PythonRuntime) Init() error {
	return errWasmNotAvailable
}

func (r *PythonRuntime) LoadCode(code string) error {
	return errWasmNotAvailable
}

func (r *PythonRuntime) Execute(input []byte) ([]byte, error) {
	return nil, errWasmNotAvailable
}

func (r *PythonRuntime) ExecuteWithContext(ctx context.Context, input []byte) ([]byte, error) {
	return nil, errWasmNotAvailable
}

func (r *PythonRuntime) GetMemoryUsage() uint64 {
	return 0
}

func (r *PythonRuntime) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	return nil
}

type StreamingConfig struct {
	ChunkSize         int
	MaxTotalSize      int
	OutputBufferSize  int
	EnableBackpressure bool
}

func DefaultStreamingConfig() *StreamingConfig {
	return &StreamingConfig{
		ChunkSize:         DefaultChunkSize,
		MaxTotalSize:      10 * 1024 * 1024,
		OutputBufferSize:  1024 * 1024,
		EnableBackpressure: true,
	}
}

type StreamingExecutionResult struct {
	Output      []byte
	TotalChunks int
	TotalBytes  int
	Duration    time.Duration
	Error       error
}

func (r *PythonRuntime) ExecuteLargeInput(ctx context.Context, input []byte, config *StreamingConfig) (*StreamingExecutionResult, error) {
	return nil, errWasmNotAvailable
}

type StreamingState struct {
	mu       sync.RWMutex
	chunks   [][]byte
	active   bool
	metadata map[string]interface{}
}

func NewStreamingState() *StreamingState {
	return &StreamingState{
		chunks:   make([][]byte, 0),
		metadata: make(map[string]interface{}),
	}
}

func (s *StreamingState) Init() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chunks = make([][]byte, 0)
	s.active = true
}

func (s *StreamingState) AddInputChunk(chunkID int32, data []byte, isLast bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	chunkCopy := make([]byte, len(data))
	copy(chunkCopy, data)
	s.chunks = append(s.chunks, chunkCopy)
	if isLast {
		s.active = false
	}
}

func (s *StreamingState) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

func (s *StreamingState) GetChunks() [][]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([][]byte, len(s.chunks))
	copy(result, s.chunks)
	return result
}

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

type StreamInput struct {
	ChunkID int
	Data    []byte
	IsLast  bool
}

type StreamOutput struct {
	ChunkID int
	Data    []byte
	IsLast  bool
	Error   error
}

func (s *StreamingExecutor) ExecuteStreaming(ctx context.Context, input io.Reader) (<-chan StreamOutput, error) {
	return nil, errWasmNotAvailable
}

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

type StreamingReader struct {
	data      []byte
	chunkSize int
	pos       int
}

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

type StreamingWriter struct {
	buffer []byte
	mu     sync.Mutex
	closed bool
}

func NewStreamingWriter() *StreamingWriter {
	return &StreamingWriter{
		buffer: make([]byte, 0, 4096),
	}
}

func (w *StreamingWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return 0, fmt.Errorf("writer closed")
	}
	w.buffer = append(w.buffer, p...)
	return len(p), nil
}

func (w *StreamingWriter) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	result := make([]byte, len(w.buffer))
	copy(result, w.buffer)
	return result
}

func (w *StreamingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.closed = true
	return nil
}

func (r *PythonRuntime) AddFuel(fuel uint64) error {
	return errWasmNotAvailable
}

func (r *PythonRuntime) GetFuelRemaining() (uint64, error) {
	return 0, errWasmNotAvailable
}

type WASMRuntimeWithDeterminism struct {
	runtime            *PythonRuntime
	config             *DeterministicConfig
	instructionCounter uint64
	mu                 sync.Mutex
	randomState        uint64
}

func NewWASMRuntimeWithDeterminism(wasmPath string, stdout, stderr io.Writer, handler HostFunctionHandler, detConfig *DeterministicConfig) (*WASMRuntimeWithDeterminism, error) {
	return nil, errWasmNotAvailable
}

func (r *WASMRuntimeWithDeterminism) ExecuteDeterministic(ctx context.Context, input []byte, detConfig *DeterministicConfig) (*DeterministicResult, error) {
	return nil, errWasmNotAvailable
}

func (r *WASMRuntimeWithDeterminism) Close() error {
	return nil
}