// Package client provides the WasmPoolClient SDK used by the orchestrator
// to route WASM execution requests to either the local in-process pool
// (current behavior) or a remote wasm-pool-service (Phase 1+).
//
// The router implements the percentage rollout, per-tenant overrides,
// circuit-breaker fallback, and dry-run mode described in
// .kilo/plans/externalize-wasm-pool-service.md (Phase 2).
package client

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"time"
)

// Request is the canonical input to WasmPoolClient.Execute.
type Request struct {
	TenantID  string
	Runtime   string // "python" | "cpython" | "rust" | ...
	WasmPath  string // for PythonRuntimePool routing
	Input     []byte
	// Code is source code to load into the pool's Python runtime via
	// LoadCode() before ExecuteWithContext(). Used for source-based
	// functions where the pool instance is the interpreter and the user's
	// code is loaded per-request. Leave empty for pre-compiled WASM
	// binaries (the binary is already loaded by the pool factory).
	Code       []byte
	Timeout    time.Duration
	MemoryMB   uint32
	FunctionID string
	Version    string
}

// Response is the canonical output from WasmPoolClient.Execute.
type Response struct {
	Output       []byte
	Error        string
	Latency      time.Duration
	MemoryBytes  uint64
	ColdStarted  bool
}

// WasmPoolClient is the interface every transport (local / external / dry-run)
// implements. Keeping it to a single Execute method makes stubbing in tests
// trivial and avoids leaking transport details.
type WasmPoolClient interface {
	Execute(ctx context.Context, req *Request) (*Response, error)
	Name() string
	Close() error
}

// HashTenant returns a stable uint32 hash of tenantID used by the router's
// percentage rollout. We use FNV-1a over the SHA-256 digest so the result
// is uniformly distributed and doesn't require importing a heavy hash lib
// beyond the standard library.
func HashTenant(tenantID string) uint32 {
	sum := sha256.Sum256([]byte(tenantID))
	return binary.LittleEndian.Uint32(sum[:4])
}
