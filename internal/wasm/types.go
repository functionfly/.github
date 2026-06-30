// Package wasm: shared types used by adapters and runtime plumbing.
//
// The Fabric / FabricStore / Snapshot / StateFabricRepo types live here
// (not in the storage package) because the wasm SDK exposes them as the
// minimal interface adapters must implement to participate in
// per-tenant state caching, regardless of which backing store the
// orchestrator uses. Keeping them here means the SDK stays decoupled
// from the orchestrator's storage internals.
package wasm

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Fabric represents a state fabric that the pool can use for caching.
type Fabric struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description string
	Type        string
	Status      string
	Settings    map[string]interface{}
	Throughput  int64
	Latency     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FabricStore is a state store within a fabric.
type FabricStore struct {
	ID        uuid.UUID
	FabricID  uuid.UUID
	Name      string
	Type      string
	Status    string
	MaxSize   int64
	Size      int64
	Region    string
	CreatedAt time.Time
}

// Snapshot is a point-in-time snapshot of a fabric.
type Snapshot struct {
	ID         uuid.UUID
	FabricID   uuid.UUID
	Name       string
	EventCount int64
	Size       int64
	SizeBytes  int64
	CreatedAt  time.Time
}

// StateFabricRepo is the interface the pool uses to read/write
// state-fabric data.
type StateFabricRepo interface {
	GetFabric(ctx context.Context, tenantID, fabricID uuid.UUID) (*Fabric, error)
	ListStores(ctx context.Context, tenantID, fabricID uuid.UUID) ([]FabricStore, error)
	CreateSnapshot(ctx context.Context, tenantID, fabricID uuid.UUID, name string) (*Snapshot, error)
	UpdateFabric(ctx context.Context, tenantID, fabricID uuid.UUID, updates map[string]interface{}) (*Fabric, error)
}

// HostFunctionHandler is the interface WASM runtimes use to expose
// host functions (KV, HTTP, logging, etc.) to guest modules.
type HostFunctionHandler interface {
	// Log records a message from the guest module.
	Log(message string)
	// Fetch performs an outbound HTTP fetch.
	Fetch(request string) (string, error)
	// KVGet reads a key from the tenant KV store.
	KVGet(key string) (string, error)
	// KVSet writes a key to the tenant KV store.
	KVSet(key string, value string) error
	// GetEnv returns an environment variable visible to the module.
	GetEnv(name string) string
	// AIInference runs an AI inference call.
	AIInference(model string, input []byte, params string) (string, error)
	// StateGet reads from the per-tenant state fabric.
	StateGet(path string) (string, error)
	// StateSet writes to the per-tenant state fabric.
	StateSet(path string, value string) error
	// StateDelete removes a key from the per-tenant state fabric.
	StateDelete(path string) error
	// StateGetFabric returns a fabric-scoped value.
	StateGetFabric(fabricID string) (string, error)
	// StateCreateSnapshot creates a fabric snapshot.
	StateCreateSnapshot(path string, label string) (string, error)
	// GetAttestation retrieves an attestation by ID for the current function's context.
	// Returns the attestation as a JSON string.
	GetAttestation(attestationID string) (string, error)
	// Delegate delegates execution to another function with trust-aware routing.
	// targetFunctionID is the function to call, input is the JSON-encoded input,
	// and options is an optional JSON string with delegation options (min_trust_score, timeout_ms, etc.).
	Delegate(targetFunctionID string, input string, options string) (string, error)
	// Call invokes a host function by name with the given arguments
	// and returns its result. The function implementation is responsible
	// for validating the name and argument types.
	Call(ctx context.Context, name string, args ...interface{}) (interface{}, error)
}

// FetchRequest is a minimal outbound HTTP request descriptor.
type FetchRequest struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    []byte
}

// FetchResponse is the matching response.
type FetchResponse struct {
	Status  int
	Headers map[string]string
	Body    []byte
}

// DefaultHostHandler is the no-op handler used when no host functions
// are configured. All calls return safe zero values.
type DefaultHostHandler struct {
	logger interface{ Warnf(string, ...interface{}) }
}

// NewDefaultHostHandler constructs a DefaultHostHandler.
func NewDefaultHostHandler(logger interface{ Warnf(string, ...interface{}) }) *DefaultHostHandler {
	return &DefaultHostHandler{logger: logger}
}

// Log is a no-op.
func (d *DefaultHostHandler) Log(_ string) {}

// Fetch returns an error indicating fetch is not available.
func (d *DefaultHostHandler) Fetch(_ string) (string, error) {
	return "", ErrFetchNotAvailable
}

// KVGet returns ErrKVNotAvailable.
func (d *DefaultHostHandler) KVGet(_ string) (string, error) {
	return "", ErrKVNotAvailable
}

// KVSet returns ErrKVNotAvailable.
func (d *DefaultHostHandler) KVSet(_ string, _ string) error {
	return ErrKVNotAvailable
}

// GetEnv returns "".
func (d *DefaultHostHandler) GetEnv(_ string) string { return "" }

// AIInference returns ErrAINotAvailable.
func (d *DefaultHostHandler) AIInference(_ string, _ []byte, _ string) (string, error) {
	return "", ErrAINotAvailable
}

// StateGet returns ErrStateNotAvailable.
func (d *DefaultHostHandler) StateGet(_ string) (string, error) {
	return "", ErrStateNotAvailable
}

// StateSet returns ErrStateNotAvailable.
func (d *DefaultHostHandler) StateSet(_ string, _ string) error {
	return ErrStateNotAvailable
}

// StateDelete returns ErrStateNotAvailable.
func (d *DefaultHostHandler) StateDelete(_ string) error {
	return ErrStateNotAvailable
}

// StateGetFabric returns ErrStateNotAvailable.
func (d *DefaultHostHandler) StateGetFabric(_ string) (string, error) {
	return "", ErrStateNotAvailable
}

// StateCreateSnapshot returns ErrStateNotAvailable.
func (d *DefaultHostHandler) StateCreateSnapshot(_, _ string) (string, error) {
	return "", ErrStateNotAvailable
}

// GetAttestation returns ErrAttestationNotAvailable.
func (d *DefaultHostHandler) GetAttestation(_ string) (string, error) {
	return "", ErrAttestationNotAvailable
}

// Delegate returns ErrDelegateNotAvailable.
func (d *DefaultHostHandler) Delegate(_, _, _ string) (string, error) {
	return "", ErrDelegateNotAvailable
}

// Call returns (nil, nil) for every invocation.
func (d *DefaultHostHandler) Call(_ context.Context, _ string, _ ...interface{}) (interface{}, error) {
	return nil, nil
}

// Sentinel errors returned by the no-op DefaultHostHandler.
var (
	ErrFetchNotAvailable      = errorString("fetch not available in no-op host handler")
	ErrKVNotAvailable         = errorString("kv store not available in no-op host handler")
	ErrAINotAvailable         = errorString("ai inference not available in no-op host handler")
	ErrStateNotAvailable      = errorString("state fabric not available in no-op host handler")
	ErrAttestationNotAvailable = errorString("attestation not available in no-op host handler")
	ErrDelegateNotAvailable   = errorString("delegate not available in no-op host handler")
)

type errorString string

func (e errorString) Error() string { return string(e) }
