// Package trace implements trace recording for the DCC Protocol.
// It records execution traces as chunks that are included in TraceHash for verification.
package trace

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
)

// Trace records the execution trace for the capsule.
type Trace struct {
	chunks    []*Chunk
	chunkSize int
	enabled   bool
	hash      string
	mu        sync.RWMutex
}

// Chunk represents a trace chunk.
type Chunk struct {
	Index     int       `json:"index"`
	Entries   []Entry   `json:"entries"`
	Hash      string    `json:"hash"`
	PrevHash  string    `json:"prev_hash"`
}

// Entry represents a single trace entry.
type Entry struct {
	Type      string `json:"type"`       // "syscall", "memory", "control"
	Timestamp uint64 `json:"timestamp"`  // Virtual time
	Data      []byte `json:"data"`       // Canonicalized data
}

// New creates a new Trace with the given chunk size.
func New(chunkSize int) *Trace {
	if chunkSize <= 0 {
		chunkSize = 1000 // Default chunk size
	}
	
	return &Trace{
		chunks:    make([]*Chunk, 0),
		chunkSize: chunkSize,
		enabled:   true,
	}
}

// Enable enables trace recording.
func (t *Trace) Enable() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enabled = true
}

// Disable disables trace recording.
func (t *Trace) Disable() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enabled = false
}

// IsEnabled returns true if trace recording is enabled.
func (t *Trace) IsEnabled() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.enabled
}

// Record records a trace entry.
func (t *Trace) Record(entryType string, timestamp uint64, data []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if !t.enabled {
		return
	}
	
	entry := Entry{
		Type:      entryType,
		Timestamp: timestamp,
		Data:      data,
	}
	
	// Get or create current chunk
	if len(t.chunks) == 0 {
		t.chunks = append(t.chunks, &Chunk{
			Index:     0,
			Entries:   make([]Entry, 0),
			PrevHash:  "",
		})
	}
	
	currentChunk := t.chunks[len(t.chunks)-1]
	currentChunk.Entries = append(currentChunk.Entries, entry)
	
	// If chunk is full, create new chunk
	if len(currentChunk.Entries) > t.chunkSize {
		// Hash the full chunk
		currentChunk.Hash = t.hashChunk(currentChunk)
		
		// Create new chunk
		newChunk := &Chunk{
			Index:    len(t.chunks),
			Entries:  make([]Entry, 0),
			PrevHash: currentChunk.Hash,
		}
		t.chunks = append(t.chunks, newChunk)
	}
}

// hashChunk computes the hash of a chunk.
func (t *Trace) hashChunk(chunk *Chunk) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("index:%d", chunk.Index)))
	h.Write([]byte(chunk.PrevHash))
	
	for _, e := range chunk.Entries {
		h.Write([]byte(e.Type))
		h.Write(e.Data)
	}
	
	return hex.EncodeToString(h.Sum(nil))
}

// Hash returns the trace hash (TraceHash).
// This is included in the OutputHash of the MEG.
func (t *Trace) Hash() string {
	t.mu.RLock()
	
	if t.hash != "" {
		t.mu.RUnlock()
		return t.hash
	}
	
	chunks := t.chunks
	t.mu.RUnlock()
	
	// If no chunks, return empty hash
	if len(chunks) == 0 {
		return ""
	}
	
	t.mu.Lock()
	defer t.mu.Unlock()
	
	// Double-check after acquiring write lock
	if t.hash != "" {
		return t.hash
	}
	
	h := sha256.New()
	h.Write([]byte("FX_TRACE"))
	
	// Hash all chunks
	for _, chunk := range chunks {
		// Ensure chunk is hashed
		if chunk.Hash == "" {
			chunk.Hash = t.hashChunk(chunk)
		}
		h.Write([]byte(chunk.Hash))
	}
	
	t.hash = hex.EncodeToString(h.Sum(nil))
	return t.hash
}

// Chunks returns a copy of all chunks.
func (t *Trace) Chunks() []*Chunk {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	chunks := make([]*Chunk, len(t.chunks))
	copy(chunks, t.chunks)
	
	return chunks
}

// EntryCount returns the total number of entries.
func (t *Trace) EntryCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	count := 0
	for _, chunk := range t.chunks {
		count += len(chunk.Entries)
	}
	
	return count
}

// Clear clears all trace data.
func (t *Trace) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.chunks = make([]*Chunk, 0)
	t.hash = ""
}

// Reset resets and reinitializes with new chunk size.
func (t *Trace) Reset(chunkSize int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	
	t.chunks = make([]*Chunk, 0)
	t.chunkSize = chunkSize
	t.hash = ""
}

// JSON returns the trace as JSON.
func (t *Trace) JSON() (string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	data, err := json.MarshalIndent(t.chunks, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal trace: %w", err)
	}
	
	return string(data), nil
}

// CompactHash returns a compact hash representation (for verification).
func (t *Trace) CompactHash() string {
	hash := t.Hash()
	
	// Return first 16 bytes as hex (for human-readable verification)
	if len(hash) > 16 {
		return hash[:16]
	}
	
	return hash
}
