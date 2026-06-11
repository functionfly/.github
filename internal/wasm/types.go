// Package wasm: shared types and interfaces (built with and without CGO).
package wasm

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// FetchRequest represents an HTTP request from WASM
type FetchRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// FetchResponse represents an HTTP response to WASM
type FetchResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body"`
}

// HostFunctionHandler defines the interface for host function implementations
type HostFunctionHandler interface {
	Log(message string)
	Fetch(request string) (string, error)
	KVGet(key string) (string, error)
	KVSet(key, value string) error
	GetEnv(name string) (string, error)
	// AIInference performs AI inference via the AI Gateway
	// model: the AI model to use
	// input: the input data for inference
	// params: JSON string of model parameters
	// Returns: JSON response string with output, latency, cost
	AIInference(model string, input []byte, params string) (string, error)
	// StateFabric operations for edge functions
	// StateGet retrieves a value from StateFabric
	// path: full state path (tenant/fabric/key)
	// Returns: JSON string with value and metadata
	StateGet(path string) (string, error)
	// StateSet stores a value in StateFabric
	// path: full state path (tenant/fabric/key)
	// value: JSON string of the value to store
	// Returns: error if operation fails
	StateSet(path string, value string) error
	// StateDelete removes a value from StateFabric
	// path: full state path (tenant/fabric/key)
	// Returns: error if operation fails
	StateDelete(path string) error
	// StateGetFabric retrieves fabric metadata and stores
	// fabricID: the fabric identifier
	// Returns: JSON string with fabric configuration
	StateGetFabric(fabricID string) (string, error)
	// StateCreateSnapshot creates a snapshot of current state
	// path: state path to snapshot
	// label: optional snapshot label
	// Returns: JSON string with snapshot metadata
	StateCreateSnapshot(path string, label string) (string, error)
}

// DefaultHostHandler provides default implementations of host functions
type DefaultHostHandler struct {
	logger  *log.Logger
	kvStore map[string]string
	kvMutex sync.RWMutex
}

// NewDefaultHostHandler creates a new default host handler
func NewDefaultHostHandler(logger *log.Logger) *DefaultHostHandler {
	return &DefaultHostHandler{
		logger:  logger,
		kvStore: make(map[string]string),
	}
}

// Log logs a message from the WASM module
func (h *DefaultHostHandler) Log(message string) {
	if h.logger != nil {
		h.logger.Println("[WASM]", message)
	} else {
		log.Println("[WASM]", message)
	}
}

// Fetch performs an HTTP request
func (h *DefaultHostHandler) Fetch(request string) (string, error) {
	var req FetchRequest
	if err := json.Unmarshal([]byte(request), &req); err != nil {
		return "", fmt.Errorf("failed to parse request JSON: %w", err)
	}
	if req.Method == "" {
		req.Method = "GET"
	}
	if req.URL == "" {
		return "", fmt.Errorf("URL is required")
	}
	var body io.Reader
	if req.Body != "" {
		body = strings.NewReader(req.Body)
	}
	httpReq, err := http.NewRequest(req.Method, req.URL, body)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	fetchResp := FetchResponse{
		Status: resp.StatusCode,
		Body:   string(respBody),
	}
	fetchResp.Headers = make(map[string]string)
	for key, values := range resp.Header {
		fetchResp.Headers[key] = strings.Join(values, ", ")
	}
	respJSON, err := json.Marshal(fetchResp)
	if err != nil {
		return "", fmt.Errorf("failed to serialize response: %w", err)
	}
	return string(respJSON), nil
}

// KVGet retrieves a value from key-value storage
func (h *DefaultHostHandler) KVGet(key string) (string, error) {
	h.kvMutex.RLock()
	defer h.kvMutex.RUnlock()
	value, exists := h.kvStore[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return value, nil
}

// KVSet stores a value in key-value storage
func (h *DefaultHostHandler) KVSet(key, value string) error {
	h.kvMutex.Lock()
	defer h.kvMutex.Unlock()
	h.kvStore[key] = value
	return nil
}

// GetEnv retrieves an environment variable
func (h *DefaultHostHandler) GetEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("environment variable not set: %s", name)
	}
	return value, nil
}

// AIInference performs AI inference via the AI Gateway
// DefaultHostHandler returns an error since AI inference is not configured
func (h *DefaultHostHandler) AIInference(model string, input []byte, params string) (string, error) {
	return "", fmt.Errorf("ai inference not configured: use a handler with AI inference enabled")
}

// StateGet retrieves a value from StateFabric
// DefaultHostHandler returns an error since StateFabric is not configured
func (h *DefaultHostHandler) StateGet(path string) (string, error) {
	return "", fmt.Errorf("state fabric not configured: use a handler with StateFabric enabled")
}

// StateSet stores a value in StateFabric
// DefaultHostHandler returns an error since StateFabric is not configured
func (h *DefaultHostHandler) StateSet(path string, value string) error {
	return fmt.Errorf("state fabric not configured: use a handler with StateFabric enabled")
}

// StateDelete removes a value from StateFabric
// DefaultHostHandler returns an error since StateFabric is not configured
func (h *DefaultHostHandler) StateDelete(path string) error {
	return fmt.Errorf("state fabric not configured: use a handler with StateFabric enabled")
}

// StateGetFabric retrieves fabric metadata
// DefaultHostHandler returns an error since StateFabric is not configured
func (h *DefaultHostHandler) StateGetFabric(fabricID string) (string, error) {
	return "", fmt.Errorf("state fabric not configured: use a handler with StateFabric enabled")
}

// StateCreateSnapshot creates a snapshot of state
// DefaultHostHandler returns an error since StateFabric is not configured
func (h *DefaultHostHandler) StateCreateSnapshot(path string, label string) (string, error) {
	return "", fmt.Errorf("state fabric not configured: use a handler with StateFabric enabled")
}
