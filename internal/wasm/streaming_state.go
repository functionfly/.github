//go:build cgo

package wasm

import (
	"sync"
	"sync/atomic"
)

// StreamingState tracks the state of streaming execution
// This is shared between the host functions and the WASM module
type StreamingState struct {
	mu sync.RWMutex

	// Input chunks
	inputChunks   [][]byte
	inputChunkIDs []int32
	isLastChunk   []bool
	inputCount    atomic.Int32

	// Output chunks
	outputChunks   [][]byte
	outputChunkIDs []int32
	outputReady    []bool
	outputCount    atomic.Int32

	// Streaming status
	isActive    atomic.Bool
	isComplete  atomic.Bool
	totalInput  atomic.Int32
	totalOutput atomic.Int32
}

// Memory layout constants for streaming
const (
	// InputChunkBaseOffset is where input chunks start in WASM memory
	InputChunkBaseOffset = 8192
	// OutputChunkBaseOffset is where output chunks start in WASM memory
	OutputChunkBaseOffset = 65536
	// ChunkMetaSize is the size of chunk metadata (4 fields * 4 bytes)
	ChunkMetaSize = 16
)

// NewStreamingState creates a new streaming state
func NewStreamingState() *StreamingState {
	return &StreamingState{
		inputChunks:   make([][]byte, 0, MaxStreamingChunks),
		inputChunkIDs: make([]int32, 0, MaxStreamingChunks),
		isLastChunk:   make([]bool, 0, MaxStreamingChunks),
		outputChunks:  make([][]byte, 0, MaxStreamingChunks),
		outputChunkIDs: make([]int32, 0, MaxStreamingChunks),
		outputReady:   make([]bool, 0, MaxStreamingChunks),
	}
}

// Init initializes the streaming state for a new execution
func (s *StreamingState) Init() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.inputChunks = s.inputChunks[:0]
	s.inputChunkIDs = s.inputChunkIDs[:0]
	s.isLastChunk = s.isLastChunk[:0]
	s.outputChunks = s.outputChunks[:0]
	s.outputChunkIDs = s.outputChunkIDs[:0]
	s.outputReady = s.outputReady[:0]

	s.inputCount.Store(0)
	s.outputCount.Store(0)
	s.totalInput.Store(0)
	s.totalOutput.Store(0)
	s.isActive.Store(false)
	s.isComplete.Store(false)
}

// AddInputChunk adds an input chunk to the streaming state
func (s *StreamingState) AddInputChunk(chunkID int32, data []byte, isLast bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Grow slices if needed
	if chunkID >= int32(cap(s.inputChunks)) {
		// Expand capacity
		newChunks := make([][]byte, len(s.inputChunks), int(chunkID)+1)
		newIDs := make([]int32, len(s.inputChunkIDs), int(chunkID)+1)
		newLast := make([]bool, len(s.isLastChunk), int(chunkID)+1)
		copy(newChunks, s.inputChunks)
		copy(newIDs, s.inputChunkIDs)
		copy(newLast, s.isLastChunk)
		s.inputChunks = newChunks
		s.inputChunkIDs = newIDs
		s.isLastChunk = newLast
	}

	// Ensure index exists
	for int32(len(s.inputChunks)) <= chunkID {
		s.inputChunks = append(s.inputChunks, nil)
		s.inputChunkIDs = append(s.inputChunkIDs, 0)
		s.isLastChunk = append(s.isLastChunk, false)
	}

	// Store chunk
	s.inputChunks[chunkID] = data
	s.inputChunkIDs[chunkID] = chunkID
	s.isLastChunk[chunkID] = isLast
	s.inputCount.Add(1)
	s.totalInput.Add(int32(len(data)))

	if isLast {
		s.isActive.Store(true)
	}
}

// GetInputChunk retrieves an input chunk by ID
func (s *StreamingState) GetInputChunk(chunkID int32) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if chunkID < 0 || int(chunkID) >= len(s.inputChunks) {
		return nil, false
	}
	return s.inputChunks[chunkID], s.isLastChunk[chunkID]
}

// GetInputChunkPtr returns the memory offset for a given chunk ID
func (s *StreamingState) GetInputChunkPtr(chunkID int32) int32 {
	if chunkID < 0 {
		return 0
	}
	return InputChunkBaseOffset + (chunkID * int32(MaxStreamingChunks))
}

// AddOutputChunk adds an output chunk to the streaming state
func (s *StreamingState) AddOutputChunk(chunkID int32, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure capacity
	for int32(len(s.outputChunks)) <= chunkID {
		s.outputChunks = append(s.outputChunks, nil)
		s.outputChunkIDs = append(s.outputChunkIDs, 0)
		s.outputReady = append(s.outputReady, false)
	}

	s.outputChunks[chunkID] = data
	s.outputChunkIDs[chunkID] = chunkID
	s.outputReady[chunkID] = true
	s.outputCount.Add(1)
	s.totalOutput.Add(int32(len(data)))
}

// GetOutputChunk retrieves an output chunk by ID
func (s *StreamingState) GetOutputChunk(chunkID int32) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if chunkID < 0 || int(chunkID) >= len(s.outputChunks) {
		return nil, false
	}
	return s.outputChunks[chunkID], s.outputReady[chunkID]
}

// GetOutputChunkPtr returns the memory offset for an output chunk
func (s *StreamingState) GetOutputChunkPtr(chunkID int32) int32 {
	if chunkID < 0 {
		return 0
	}
	return OutputChunkBaseOffset + (chunkID * int32(MaxStreamingChunks))
}

// SetOutputReady marks an output chunk as ready with its size
func (s *StreamingState) SetOutputReady(chunkID int32, ptr int32, chunkLen int32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure capacity
	for int32(len(s.outputChunks)) <= chunkID {
		s.outputChunks = append(s.outputChunks, nil)
		s.outputReady = append(s.outputReady, false)
	}

	s.outputReady[chunkID] = true
	_ = ptr   // ptr not used in this implementation
	_ = chunkLen
}

// Complete marks the streaming as complete
func (s *StreamingState) Complete() {
	s.isComplete.Store(true)
}

// IsComplete returns whether streaming is complete
func (s *StreamingState) IsComplete() bool {
	return s.isComplete.Load()
}

// IsActive returns whether streaming is active
func (s *StreamingState) IsActive() bool {
	return s.isActive.Load()
}

// GetInputCount returns the number of input chunks
func (s *StreamingState) GetInputCount() int32 {
	return s.inputCount.Load()
}

// GetOutputCount returns the number of output chunks
func (s *StreamingState) GetOutputCount() int32 {
	return s.outputCount.Load()
}

// GetTotalInput returns total bytes of input
func (s *StreamingState) GetTotalInput() int32 {
	return s.totalInput.Load()
}

// GetTotalOutput returns total bytes of output
func (s *StreamingState) GetTotalOutput() int32 {
	return s.totalOutput.Load()
}

// ReadChunkInto reads chunk data into the provided memory location
func (s *StreamingState) ReadChunkInto(chunkID int32, destPtr int32, maxLen int32, memory []byte) int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if chunkID < 0 || int(chunkID) >= len(s.inputChunks) {
		return -1 // Error: chunk not found
	}

	chunk := s.inputChunks[chunkID]
	if chunk == nil {
		return -1
	}

	toCopy := int32(len(chunk))
	if toCopy > maxLen {
		toCopy = maxLen
	}

	if int(destPtr)+int(toCopy) > len(memory) {
		return -1 // Error: out of bounds
	}

	copy(memory[destPtr:destPtr+toCopy], chunk)
	return toCopy
}
