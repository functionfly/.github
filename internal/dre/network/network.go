// Package network implements network mode handling for the DCC Protocol.
// It provides Record, Stub, and Disabled modes for deterministic network behavior.
package network

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
)

// Mode represents the network mode for the capsule.
type Mode string

const (
	// ModeRecord allows outbound network calls and records responses for replay.
	ModeRecord Mode = "record"
	// ModeStub uses a manifest to serve deterministic responses.
	ModeStub Mode = "stub"
	// ModeDisabled aborts all network calls.
	ModeDisabled Mode = "disabled"
)

// NetworkHandler handles network requests based on the configured mode.
type NetworkHandler struct {
	mode          Mode
	stubs         map[string][]byte
	recordings    map[string][]byte
	dependencyHash string
	mu            sync.RWMutex
}

// New creates a new NetworkHandler with the given mode.
func New(mode Mode) *NetworkHandler {
	return &NetworkHandler{
		mode:       mode,
		stubs:      make(map[string][]byte),
		recordings: make(map[string][]byte),
	}
}

// NewWithManifest creates a new NetworkHandler with stub manifest.
func NewWithManifest(mode Mode, manifestPath string) (*NetworkHandler, error) {
	h := New(mode)
	
	if mode == ModeStub && manifestPath != "" {
		if err := h.LoadManifest(manifestPath); err != nil {
			return nil, fmt.Errorf("failed to load manifest: %w", err)
		}
	}
	
	return h, nil
}

// Mode returns the current network mode.
func (h *NetworkHandler) Mode() Mode {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.mode
}

// SetMode sets the network mode.
func (h *NetworkHandler) SetMode(mode Mode) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.mode = mode
}

// HandleRequest handles a network request based on the current mode.
// Returns the response body and an error if the request should be aborted.
func (h *NetworkHandler) HandleRequest(url string, requestBody []byte) ([]byte, error) {
	h.mu.RLock()
	mode := h.mode
	mu := &h.mu
	mu.RUnlock() // Release read lock before acquiring write lock
	
	switch mode {
	case ModeRecord:
		return h.handleRecord(url, requestBody)
	case ModeStub:
		return h.handleStub(url)
	case ModeDisabled:
		return nil, fmt.Errorf("network disabled: cannot make request to %s", url)
	default:
		return nil, fmt.Errorf("unknown network mode: %s", mode)
	}
}

// handleRecord handles a request in record mode.
// First execution: allow outbound, record request/response
// Replay: return recorded response
func (h *NetworkHandler) handleRecord(url string, requestBody []byte) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Check if we have a recording for this URL
	if recording, ok := h.recordings[url]; ok {
		return recording, nil
	}
	
	// In a real implementation, this would make the actual network request
	// For now, return an error indicating recording needed
	return nil, fmt.Errorf("record mode: no recording for %s (first execution should record)", url)
}

// Record records a response for a URL (for first execution in record mode).
func (h *NetworkHandler) Record(url string, responseBody []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.recordings[url] = responseBody
}

// handleStub handles a request in stub mode.
// Loads from stub_manifest.json
func (h *NetworkHandler) handleStub(url string) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if stub, ok := h.stubs[url]; ok {
		return stub, nil
	}
	
	return nil, fmt.Errorf("stub mode: no stub for %s", url)
}

// LoadManifest loads stub responses from a manifest file.
func (h *NetworkHandler) LoadManifest(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// In a real implementation, this would read from the file
	// For now, just set up empty stubs
	// The manifest format would be:
	// {
	//   "stubs": {
	//     "https://api.example.com/data": {"body": "...", "status": 200}
	//   }
	// }
	
	return nil
}

// AddStub adds a stub response for a URL.
func (h *NetworkHandler) AddStub(url string, responseBody []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.stubs[url] = responseBody
}

// DependencyHash returns the hash of all recorded/stubbed responses.
// This is included in the DependencyHash of the MEG.
func (h *NetworkHandler) DependencyHash() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if h.dependencyHash != "" {
		return h.dependencyHash
	}
	
	h := sha256.New()
	
	// Hash all recordings
	for url, body := range h.recordings {
		h.Write([]byte(url))
		h.Write(body)
	}
	
	// Hash all stubs
	for url, body := range h.stubs {
		h.Write([]byte("stub:"))
		h.Write([]byte(url))
		h.Write(body)
	}
	
	h.dependencyHash = hex.EncodeToString(h.Sum(nil))
	return h.dependencyHash
}

// Recordings returns a copy of all recordings.
func (h *NetworkHandler) Recordings() map[string][]byte {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	recordings := make(map[string][]byte, len(h.recordings))
	for k, v := range h.recordings {
		recordings[k] = v
	}
	
	return recordings
}

// Stubs returns a copy of all stubs.
func (h *NetworkHandler) Stubs() map[string][]byte {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	stubs := make(map[string][]byte, len(h.stubs))
	for k, v := range h.stubs {
		stubs[k] = v
	}
	
	return stubs
}

// Manifest represents the stub manifest structure.
type Manifest struct {
	Stubs map[string]StubEntry `json:"stubs"`
}

// StubEntry represents a single stub entry.
type StubEntry struct {
	Body   string            `json:"body"`
	Status int               `json:"status"`
	Headers map[string]string `json:"headers,omitempty"`
}

// JSON returns the manifest as JSON.
func (h *NetworkHandler) JSON() (string, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	manifest := Manifest{
		Stubs: make(map[string]StubEntry),
	}
	
	for url, body := range h.stubs {
		manifest.Stubs[url] = StubEntry{
			Body:   string(body),
			Status: 200,
		}
	}
	
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal manifest: %w", err)
	}
	
	return string(data), nil
}

// IsAllowed returns true if network requests are allowed.
func (h *NetworkHandler) IsAllowed() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	return h.mode != ModeDisabled
}
