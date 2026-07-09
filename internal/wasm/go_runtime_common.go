// go_runtime_common.go holds the shared types and constants used by both
// the wasmtime-backed GoRuntime (go_runtime.go, //go:build cgo) and the
// wazero-backed WazeroGoRuntime (go_runtime_wazero.go, //go:build !cgo).
package wasm

import (
	"context"
	"errors"
	"os"
	"time"
)

// GoRuntimeConfig captures the security and resource limits for a Go WASM
// runtime instance.
type GoRuntimeConfig struct {
	MaxMemoryMB     int
	Timeout         time.Duration
	MaxInputBytes   int
	MaxOutputBytes  int
	MaxInstructions uint64
	EnableFuel      bool
	WorkDir         string
	BaseWorkDir     string
}

func NewDefaultGoRuntimeConfig() GoRuntimeConfig {
	return GoRuntimeConfig{
		MaxMemoryMB:     256,
		Timeout:         60 * time.Second,
		MaxInputBytes:   10 * 1024 * 1024,
		MaxOutputBytes:  10 * 1024 * 1024,
		MaxInstructions: 5_000_000_000,
		EnableFuel:      true,
		BaseWorkDir:     "/var/lib/functionfly/go-instances",
	}
}

func NewGoRuntimeConfigFromEnv() GoRuntimeConfig {
	cfg := NewDefaultGoRuntimeConfig()
	if v := os.Getenv("GO_RUNTIME_MAX_MEMORY_MB"); v != "" {
		if n, err := parsePositiveInt(v); err == nil {
			cfg.MaxMemoryMB = n
		}
	}
	if v := os.Getenv("GO_RUNTIME_TIMEOUT_MS"); v != "" {
		if n, err := parsePositiveInt(v); err == nil {
			cfg.Timeout = time.Duration(n) * time.Millisecond
		}
	}
	if v := os.Getenv("GO_RUNTIME_MAX_INPUT_BYTES"); v != "" {
		if n, err := parsePositiveInt(v); err == nil {
			cfg.MaxInputBytes = n
		}
	}
	if v := os.Getenv("GO_RUNTIME_MAX_OUTPUT_BYTES"); v != "" {
		if n, err := parsePositiveInt(v); err == nil {
			cfg.MaxOutputBytes = n
		}
	}
	if v := os.Getenv("GO_RUNTIME_MAX_INSTRUCTIONS"); v != "" {
		if n, err := parsePositiveUint64(v); err == nil {
			cfg.MaxInstructions = n
		}
	}
	if v := os.Getenv("GO_RUNTIME_ENABLE_FUEL"); v == "false" || v == "0" {
		cfg.EnableFuel = false
	}
	if v := os.Getenv("GO_RUNTIME_BASE_WORK_DIR"); v != "" {
		cfg.BaseWorkDir = v
	}
	return cfg
}

func (c GoRuntimeConfig) MaxMemoryBytes() uint64 {
	return uint64(c.MaxMemoryMB) * 1024 * 1024
}

func (c GoRuntimeConfig) MaxMemoryPages() uint32 {
	pages := uint32(c.MaxMemoryMB) * 16
	if pages < 1 {
		pages = 1
	}
	return pages
}

// GoRuntimeStats is the observability record for a GoRuntime instance.
type GoRuntimeStats struct {
	WorkDir   string        `json:"work_dir"`
	ExecCount int64         `json:"exec_count"`
	CreatedAt time.Time     `json:"created_at"`
	Uptime    time.Duration `json:"uptime"`
}

// GoRuntimeIfc is the interface both backend implementations satisfy. The
// pool uses it so it can hold instances from either backend.
type GoRuntimeIfc interface {
	ExecuteWithContext(ctx context.Context, input []byte) ([]byte, error)
	Execute(input []byte) ([]byte, error)
	Healthy(ctx context.Context) bool
	Stats() GoRuntimeStats
	Close() error
	WorkDir() string
	CreatedAt() time.Time
}

// Sentinel errors.
var (
	ErrGoRuntimeClosed   = errors.New("go runtime: instance is closed")
	ErrGoInputTooLarge   = errors.New("go runtime: input too large")
	ErrGoOutputTooLarge  = errors.New("go runtime: output too large")
	ErrGoNoOutput        = errors.New("go runtime: handler produced no output.json")
	ErrGoBadEnvelope     = errors.New("go runtime: handler produced malformed envelope")
	ErrGoHandlerError    = errors.New("go runtime: handler returned error")
	ErrGoTimeout         = errors.New("go runtime: execution timed out")
	ErrGoFuelExhausted   = errors.New("go runtime: instruction limit exceeded")
	ErrGoPanic           = errors.New("go runtime: handler panicked")
	ErrGoExecutionFailed = errors.New("go runtime: execution failed")
)

func parsePositiveInt(s string) (int, error) {
	if s == "" {
		return 0, errBadInt
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errBadInt
		}
		n = n*10 + int(c-'0')
	}
	if n <= 0 {
		return 0, errBadInt
	}
	return n, nil
}

func parsePositiveUint64(s string) (uint64, error) {
	if s == "" {
		return 0, errBadInt
	}
	var n uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errBadInt
		}
		n = n*10 + uint64(c-'0')
	}
	if n == 0 {
		return 0, errBadInt
	}
	return n, nil
}

var errBadInt = errors.New("invalid integer")
