// Package network implements network mode handling for the DCC Protocol.
// It provides Record, Stub, and Disabled modes for deterministic network behavior.
package network

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	mode           Mode
	stubs          map[string][]byte
	recordings     map[string][]byte
	dependencyHash string
	mu             sync.RWMutex
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

// doHTTPRequest performs an HTTP request and returns the response body.
// Uses POST with requestBody when non-nil, GET otherwise.
func doHTTPRequest(url string, requestBody []byte) ([]byte, error) {
	var req *http.Request
	var err error
	if len(requestBody) > 0 {
		req, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(requestBody))
	} else {
		req, err = http.NewRequest(http.MethodGet, url, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if len(requestBody) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// handleRecord handles a request in record mode.
// First execution: allow outbound, record request/response
// Replay: return recorded response
func (h *NetworkHandler) handleRecord(url string, requestBody []byte) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if we have a recording for this URL (replay)
	if recording, ok := h.recordings[url]; ok {
		return recording, nil
	}

	// First execution: perform the actual network request and record the response
	respBody, err := doHTTPRequest(url, requestBody)
	if err != nil {
		return nil, fmt.Errorf("record mode: request failed for %s: %w", url, err)
	}
	h.recordings[url] = respBody
	return respBody, nil
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

// stubEntry is one stub in the manifest (body and optional status).
type stubEntry struct {
	Body   string `json:"body"`
	Status int    `json:"status"`
}

// manifestFile is the JSON structure of the stub manifest file.
type manifestFile struct {
	Stubs map[string]stubEntry `json:"stubs"`
}

// LoadManifest loads stub responses from a manifest file.
// Manifest format: { "stubs": { "url": { "body": "...", "status": 200 }, ... } }
// Body is the raw response body (string); status is optional and currently unused.
func (h *NetworkHandler) LoadManifest(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}
	var manifest manifestFile
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("parse manifest: %w", err)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.stubs == nil {
		h.stubs = make(map[string][]byte)
	}
	for url, entry := range manifest.Stubs {
		h.stubs[url] = []byte(entry.Body)
	}
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

	hasher := sha256.New()
	for url, body := range h.recordings {
		hasher.Write([]byte(url))
		hasher.Write(body)
	}
	for url, body := range h.stubs {
		hasher.Write([]byte("stub:"))
		hasher.Write([]byte(url))
		hasher.Write(body)
	}
	h.dependencyHash = hex.EncodeToString(hasher.Sum(nil))
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
	Body    string            `json:"body"`
	Status  int               `json:"status"`
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
