//go:build !cgo

// go_runtime_wazero.go implements the production Go runtime on top of
// wazero (pure Go, no CGO) for builds without CGO.
//
// File name does NOT end in `_wasm.go` because the Go toolchain treats
// that suffix as GOARCH=wasm-only sources.
package wasm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WazeroGoRuntime implements GoRuntimeIfc on wazero.
type WazeroGoRuntime struct {
	mu        sync.Mutex
	wasmBytes []byte
	cfg       GoRuntimeConfig
	handler   HostFunctionHandler
	scanner   *RuntimeScanner
	workDir   string

	wazeroRT wazero.Runtime
	module   api.Module

	createdAt time.Time
	execCount int64
}

// NewGoRuntime (no-CGO build) compiles and instantiates a Go WASM module
// via wazero.
func NewGoRuntime(wasmBytes []byte, handler HostFunctionHandler, cfg GoRuntimeConfig) (*WazeroGoRuntime, error) {
	if len(wasmBytes) == 0 {
		return nil, fmt.Errorf("go runtime: empty wasm bytes")
	}
	if cfg.MaxMemoryMB <= 0 {
		cfg = NewDefaultGoRuntimeConfig()
	}
	if cfg.BaseWorkDir == "" {
		cfg.BaseWorkDir = "/var/lib/functionfly/go-instances"
	}
	if err := os.MkdirAll(cfg.BaseWorkDir, 0o755); err != nil {
		return nil, fmt.Errorf("go runtime: mkdir base workdir: %w", err)
	}
	workDir, err := os.MkdirTemp(cfg.BaseWorkDir, "instance-")
	if err != nil {
		return nil, fmt.Errorf("go runtime: mkdir workdir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(workDir, "functionfly"), 0o755); err != nil {
		_ = os.RemoveAll(workDir)
		return nil, fmt.Errorf("go runtime: mkdir functionfly: %w", err)
	}

	if handler == nil {
		handler = NewDefaultHostHandler(nil)
	}
	scanner := NewRuntimeScanner()

	rt := wazero.NewRuntimeWithConfig(context.Background(),
		wazero.NewRuntimeConfig().
			WithCloseOnContextDone(true).
			WithMemoryLimitPages(cfg.MaxMemoryPages()))

	if _, err := wasi_snapshot_preview1.Instantiate(context.Background(), rt); err != nil {
		_ = os.RemoveAll(workDir)
		rt.Close(context.Background())
		return nil, fmt.Errorf("go runtime: instantiate wasi: %w", err)
	}

	if _, err := buildGoHostModule(rt, handler); err != nil {
		_ = os.RemoveAll(workDir)
		rt.Close(context.Background())
		return nil, fmt.Errorf("go runtime: build host module: %w", err)
	}

	return &WazeroGoRuntime{
		wasmBytes: wasmBytes,
		cfg:       cfg,
		handler:   handler,
		scanner:   scanner,
		workDir:   workDir,
		wazeroRT:  rt,
		createdAt: time.Now(),
	}, nil
}

func (r *WazeroGoRuntime) WorkDir() string { return r.workDir }
func (r *WazeroGoRuntime) CreatedAt() time.Time {
	return r.createdAt
}

func (r *WazeroGoRuntime) Healthy(_ context.Context) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.wazeroRT != nil
}

func (r *WazeroGoRuntime) ExecuteWithContext(ctx context.Context, input []byte) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.wazeroRT == nil {
		return nil, ErrGoRuntimeClosed
	}
	if len(input) > r.cfg.MaxInputBytes {
		return nil, fmt.Errorf("%w: %d > %d", ErrGoInputTooLarge, len(input), r.cfg.MaxInputBytes)
	}

	execCtx, cancel := context.WithTimeout(ctx, r.cfg.Timeout)
	defer cancel()

	inputPath := filepath.Join(r.workDir, "input.json")
	outputPath := filepath.Join(r.workDir, "output.json")
	_ = os.Remove(outputPath)

	if err := os.WriteFile(inputPath, input, 0o600); err != nil {
		return nil, fmt.Errorf("go runtime: write input: %w", err)
	}

	compiled, err := r.wazeroRT.CompileModule(execCtx, r.wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("go runtime: compile: %w", err)
	}

	modCfg := wazero.NewModuleConfig().
		WithName("ff_go").
		WithStartFunctions("_start").
		WithFSConfig(wazero.NewFSConfig().WithDirMount(r.workDir, "/functionfly"))

	mod, err := r.wazeroRT.InstantiateModule(execCtx, compiled, modCfg)
	if err != nil {
		return nil, fmt.Errorf("go runtime: instantiate: %w", err)
	}
	defer mod.Close(execCtx)
	r.module = mod

	r.execCount++

	data, err := os.ReadFile(outputPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrGoNoOutput
		}
		return nil, fmt.Errorf("go runtime: read output: %w", err)
	}
	if len(data) > r.cfg.MaxOutputBytes {
		return nil, fmt.Errorf("%w: %d > %d", ErrGoOutputTooLarge, len(data), r.cfg.MaxOutputBytes)
	}

	envelope := struct {
		Success bool            `json:"success"`
		Result  json.RawMessage `json:"result"`
		Error   string          `json:"error"`
	}{}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGoBadEnvelope, err)
	}
	if !envelope.Success {
		return data, fmt.Errorf("%w: %s", ErrGoHandlerError, envelope.Error)
	}
	return data, nil
}

func (r *WazeroGoRuntime) Execute(input []byte) ([]byte, error) {
	return r.ExecuteWithContext(context.Background(), input)
}

func (r *WazeroGoRuntime) LoadCode(_ []byte) error { return nil }
func (r *WazeroGoRuntime) Init() error             { return nil }

func (r *WazeroGoRuntime) Stats() GoRuntimeStats {
	r.mu.Lock()
	defer r.mu.Unlock()
	return GoRuntimeStats{
		WorkDir:   r.workDir,
		ExecCount: r.execCount,
		CreatedAt: r.createdAt,
		Uptime:    time.Since(r.createdAt),
	}
}

func (r *WazeroGoRuntime) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.module = nil
	if r.wazeroRT != nil {
		_ = r.wazeroRT.Close(context.Background())
		r.wazeroRT = nil
	}
	if r.workDir != "" {
		_ = os.RemoveAll(r.workDir)
		r.workDir = ""
	}
	return nil
}

var _ GoRuntimeIfc = (*WazeroGoRuntime)(nil)

func buildGoHostModule(rt wazero.Runtime, handler HostFunctionHandler) (api.Module, error) {
	b := rt.NewHostModuleBuilder("functionfly")
	b.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, ptr, size uint32) {
		mem := m.Memory()
		if mem == nil {
			return
		}
		data, ok := mem.Read(ptr, size)
		if !ok {
			return
		}
		handler.Log(string(data))
	}).Export("log")
	mod, err := b.Instantiate(context.Background())
	if err != nil {
		return nil, err
	}
	return mod, nil
}
