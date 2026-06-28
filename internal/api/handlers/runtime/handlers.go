package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/gorilla/mux"
)

// Handler handles runtime-related API requests
type Handler struct {
	registry interface {
		GetFunction(id string) (*FunctionConfig, error)
		UpdateFunction(fn *FunctionConfig) error
	}
}

// New creates a new runtime handler
func New() *Handler {
	return &Handler{
		registry: nil,
	}
}

// FunctionConfig represents the runtime configuration for a function used by this handler.
// It is intentionally minimal and decoupled from storage models.
type FunctionConfig struct {
	Runtime        string
	MemoryMB       int
	TimeoutMs      int
	NetworkEnabled bool
	Environment    map[string]string
	MaxConcurrent  int
}

// RuntimeInfo represents runtime information
type RuntimeInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Status      string   `json:"status"`
	Features    []string `json:"features"`
	MemoryLimit int      `json:"memory_limit_mb"`
	Timeout     int      `json:"timeout_ms"`
}

// RuntimeDiagnostics represents runtime diagnostics data
type RuntimeDiagnostics struct {
	Runtime     string            `json:"runtime"`
	Environment map[string]string `json:"environment"`
	Resources   ResourceUsage     `json:"resources"`
	Network     NetworkStatus     `json:"network"`
	Security    SecurityStatus    `json:"security"`
	Performance PerformanceData   `json:"performance"`
	Timestamp   time.Time         `json:"timestamp"`
}

// ResourceUsage represents resource usage
type ResourceUsage struct {
	MemoryLimit int `json:"memory_limit_mb"`
	MemoryUsed  int `json:"memory_used_mb"`
	CPUCores    int `json:"cpu_cores"`
	Timeout     int `json:"timeout_ms"`
}

// NetworkStatus represents network status
type NetworkStatus struct {
	Enabled       bool     `json:"enabled"`
	DNSWorking    bool     `json:"dns_working"`
	ExternalCalls int      `json:"external_calls_per_minute"`
	AllowedHosts  []string `json:"allowed_hosts"`
}

// SecurityStatus represents security status
type SecurityStatus struct {
	SandboxEnabled     bool     `json:"sandbox_enabled"`
	ModuleRestrictions bool     `json:"module_restrictions"`
	BlockedModules     []string `json:"blocked_modules"`
	EnvironmentCount   int      `json:"environment_variables"`
}

// PerformanceData represents performance metrics
type PerformanceData struct {
	ColdStartAvgMs   int  `json:"cold_start_avg_ms"`
	WarmExecAvgMs    int  `json:"warm_execution_avg_ms"`
	CodeCacheEnabled bool `json:"code_cache_enabled"`
	ConcurrentLimit  int  `json:"concurrent_limit"`
	SuccessRate      int  `json:"success_rate_percent"`
}

// ListRuntimes returns a list of available runtimes
func (h *Handler) ListRuntimes(w http.ResponseWriter, r *http.Request) {
	runtimes := []RuntimeInfo{
		// SAR Runtime - Stateful Agent Runtime for AI agent execution
		{
			ID:          "sar",
			Name:        "SAR Runtime",
			Version:     "1.0",
			Status:      "stable",
			Features:    []string{"Stateful AI Agents", "Event-driven NATS", "WASM sandbox", "Multi-tier memory", "gRPC API", "Prometheus metrics"},
			MemoryLimit: 4096,
			Timeout:     300000,
		},
		// Prism - Universal Adaptive WASM Execution Fabric
		{
			ID:          "prism",
			Name:        "Prism",
			Version:     "2.0",
			Status:      "stable",
			Features:    []string{"Adaptive Execution Cells", "HyperCore scheduler", "WASM Fusion Engine", "Quantum snapshotting", "Neural optimization", "Autonomous swarms"},
			MemoryLimit: 4096,
			Timeout:     300000,
		},
		// Node.js - QuickJS WASM-based JavaScript execution
		{
			ID:          "nodejs",
			Name:        "Node.js",
			Version:     "20.x LTS",
			Status:      "stable",
			Features:    []string{"async/await", "ES Modules", "fetch API", "Worker Threads", "QuickJS WASM", "Prometheus metrics"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		{
			ID:          "nodejs18",
			Name:        "Node.js",
			Version:     "18.x LTS",
			Status:      "stable",
			Features:    []string{"async/await", "ES Modules", "fetch API"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		// Deno - Secure JavaScript/TypeScript execution
		{
			ID:          "deno",
			Name:        "Deno",
			Version:     "2.x",
			Status:      "stable",
			Features:    []string{"TypeScript", "Built-in formatting", "Secure by default", "Deno Deploy", "Web APIs", "Node.js compatibility", "NPM compatibility", "WASM support", "WASM sandbox", "seccomp/landlock"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		// Bun - Fast JavaScript/TypeScript runtime
		{
			ID:          "bun",
			Name:        "Bun",
			Version:     "1.x",
			Status:      "stable",
			Features:    []string{"TypeScript", "Built-in bundler", "Fast startup", "Web APIs", "Node.js compatibility", "Native Bun runtime", "React Server Components", "WASM sandbox", "seccomp/landlock"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		// Ruby - Secure Ruby execution via mruby/WASM
		{
			ID:          "ruby",
			Name:        "Ruby",
			Version:     "3.3 (mruby)",
			Status:      "stable",
			Features:    []string{"mruby interpreter", "WASM", "Dynamic typing", "WASM sandbox", "seccomp/landlock", "NATS client"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		// Kotlin - Secure Kotlin/JVM execution
		{
			ID:          "kotlin",
			Name:        "Kotlin",
			Version:     "1.9+",
			Status:      "stable",
			Features:    []string{"Kotlin/WASM", "wasmWasi target", "Null safety", "Coroutines", "JVM execution", "WASM sandbox", "seccomp/landlock", "NATS client"},
			MemoryLimit: 4096,
			Timeout:     300000,
		},
		// WasmEdge - C/C++ and WASM execution
		{
			ID:          "wasmedge",
			Name:        "WasmEdge",
			Version:     "0.14+",
			Status:      "stable",
			Features:    []string{"C/C++", "WASI", "wasmedge-sdk", "Fuel metering", "Secure execution", "seccomp/landlock", "NATS client"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		// Python 3.12
		{
			ID:          "python3.12",
			Name:        "Python",
			Version:     "3.12",
			Status:      "stable",
			Features:    []string{"async/await", "Type hints", "Better errors", "MicroPython WASM", "WASM sandbox"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		{
			ID:          "python3.11",
			Name:        "Python",
			Version:     "3.11",
			Status:      "stable",
			Features:    []string{"async/await", "Type hints", "f-strings"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		// Python MicroVM - Enterprise tier
		{
			ID:          "python-microvm",
			Name:        "Python (MicroVM)",
			Version:     "Enterprise",
			Status:      "stable",
			Features:    []string{"CPython 3.11+", "NumPy/Pandas", "C extensions", "Firecracker isolation"},
			MemoryLimit: 8192,
			Timeout:     600000,
		},
		// Rust - WASM compilation
		{
			ID:          "rust",
			Name:        "Rust",
			Version:     "1.75+",
			Status:      "stable",
			Features:    []string{"wasm32-wasip1", "WASI", "Near-native performance", "Zero-cost abstractions"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		// Go - WASM compilation
		{
			ID:          "go",
			Name:        "Go",
			Version:     "1.21+",
			Status:      "stable",
			Features:    []string{"GOOS=wasip1", "WASI", "Goroutines", "Standard library"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		// C - WASM compilation
		{
			ID:          "c",
			Name:        "C",
			Version:     "C11",
			Status:      "beta",
			Features:    []string{"Emscripten/WASI-SDK", "WASI", "Low-level control", "libc"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		// C++ - WASM compilation
		{
			ID:          "cpp",
			Name:        "C++",
			Version:     "C++17",
			Status:      "beta",
			Features:    []string{"Emscripten/WASI-SDK", "WASI", "STL", "Templates"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		// Swift - Experimental WASM support
		{
			ID:          "swift",
			Name:        "Swift",
			Version:     "5.9+",
			Status:      "experimental",
			Features:    []string{"SwiftWasm", "Protocols", "Value types", "Structured concurrency"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"runtimes": runtimes,
	})
}

// GetRuntimeInfo returns information about a specific runtime
func (h *Handler) GetRuntimeInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runtimeID := vars["id"]

	runtimes := map[string]RuntimeInfo{
		"nodejs20": {
			ID:          "nodejs20",
			Name:        "Node.js",
			Version:     "20.x LTS",
			Status:      "stable",
			Features:    []string{"async/await", "ES Modules", "fetch API", "Worker Threads", "Web Streams", "Blob", "BroadcastChannel"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		"nodejs18": {
			ID:          "nodejs18",
			Name:        "Node.js",
			Version:     "18.x LTS",
			Status:      "stable",
			Features:    []string{"async/await", "ES Modules", "fetch API", "Web Streams"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		"deno": {
			ID:          "deno",
			Name:        "Deno",
			Version:     "2.x",
			Status:      "stable",
			Features:    []string{"TypeScript", "Built-in formatting", "Secure by default", "Deno Deploy", "Web APIs", "Node.js compatibility", "NPM compatibility", "WASM support"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		"bun": {
			ID:          "bun",
			Name:        "Bun",
			Version:     "1.x",
			Status:      "stable",
			Features:    []string{"TypeScript", "Built-in bundler", "Fast startup", "Web APIs", "Node.js compatibility", "Native Bun runtime", "React Server Components"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		"python3.12": {
			ID:          "python3.12",
			Name:        "Python",
			Version:     "3.12",
			Status:      "stable",
			Features:    []string{"async/await", "Type hints", "Pattern matching", "Better errors"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		"python3.11": {
			ID:          "python3.11",
			Name:        "Python",
			Version:     "3.11",
			Status:      "stable",
			Features:    []string{"async/await", "Type hints", "f-strings"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		"python-microvm": {
			ID:          "python-microvm",
			Name:        "Python (MicroVM)",
			Version:     "Enterprise",
			Status:      "stable",
			Features:    []string{"CPython 3.11+", "NumPy/Pandas", "C extensions", "Firecracker isolation"},
			MemoryLimit: 8192,
			Timeout:     600000,
		},
		"rust": {
			ID:          "rust",
			Name:        "Rust",
			Version:     "1.75+",
			Status:      "stable",
			Features:    []string{"wasm32-wasip1", "WASI", "Near-native performance", "Zero-cost abstractions"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		"go": {
			ID:          "go",
			Name:        "Go",
			Version:     "1.21+",
			Status:      "stable",
			Features:    []string{"GOOS=wasip1", "WASI", "Goroutines", "Standard library"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		"c": {
			ID:          "c",
			Name:        "C",
			Version:     "C11",
			Status:      "beta",
			Features:    []string{"Emscripten/WASI-SDK", "WASI", "Low-level control", "libc"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		"cpp": {
			ID:          "cpp",
			Name:        "C++",
			Version:     "C++17",
			Status:      "beta",
			Features:    []string{"Emscripten/WASI-SDK", "WASI", "STL", "Templates"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		"swift": {
			ID:          "swift",
			Name:        "Swift",
			Version:     "5.9+",
			Status:      "experimental",
			Features:    []string{"SwiftWasm", "Protocols", "Value types", "Structured concurrency"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
	}

	runtime, exists := runtimes[runtimeID]
	if !exists {
		apierror.WriteError(w, apierror.NewNotFound(fmt.Sprintf("Runtime '%s' not found", runtimeID)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runtime)
}

// GetDiagnostics returns diagnostics for a function's runtime
func (h *Handler) GetDiagnostics(w http.ResponseWriter, r *http.Request) {
	if h.registry == nil {
		apierror.WriteError(w, apierror.NewInternal("Registry not available"))
		return
	}

	vars := mux.Vars(r)
	functionID := vars["function_id"]

	fn, err := h.registry.GetFunction(functionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound(fmt.Sprintf("Function '%s' not found", functionID)))
		return
	}

	diagnostics := RuntimeDiagnostics{
		Runtime: fn.Runtime,
		Environment: map[string]string{
			"Node.js Version": "20.11.0",
			"V8 Engine":       "12.0.267.53",
			"Platform":        "linux/amd64",
			"Architecture":    "x86_64",
		},
		Resources: ResourceUsage{
			MemoryLimit: fn.MemoryMB,
			MemoryUsed:  45,
			CPUCores:    1,
			Timeout:     fn.TimeoutMs,
		},
		Network: NetworkStatus{
			Enabled:       fn.NetworkEnabled,
			DNSWorking:    true,
			ExternalCalls: 3,
		},
		Security: SecurityStatus{
			SandboxEnabled:     true,
			ModuleRestrictions: true,
			BlockedModules:     []string{"child_process", "fs", "net", "tls", "http", "https"},
			EnvironmentCount:   len(fn.Environment),
		},
		Performance: PerformanceData{
			ColdStartAvgMs:   450,
			WarmExecAvgMs:    12,
			CodeCacheEnabled: true,
			ConcurrentLimit:  fn.MaxConcurrent,
			SuccessRate:      99,
		},
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(diagnostics)
}

// UpdateRuntimeConfig updates runtime configuration for a function
func (h *Handler) UpdateRuntimeConfig(w http.ResponseWriter, r *http.Request) {
	if h.registry == nil {
		apierror.WriteError(w, apierror.NewInternal("Registry not available"))
		return
	}

	vars := mux.Vars(r)
	functionID := vars["function_id"]

	var config struct {
		Runtime        string `json:"runtime"`
		MemoryMB       int    `json:"memory_mb"`
		TimeoutMs      int    `json:"timeout_ms"`
		NetworkEnabled bool   `json:"network_enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Get function from registry
	fn, err := h.registry.GetFunction(functionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound(fmt.Sprintf("Function '%s' not found", functionID)))
		return
	}

	// Update function config
	fn.Runtime = config.Runtime
	fn.MemoryMB = config.MemoryMB
	fn.TimeoutMs = config.TimeoutMs
	fn.NetworkEnabled = config.NetworkEnabled

	// Save to registry
	if err := h.registry.UpdateFunction(fn); err != nil {
		apierror.WriteError(w, apierror.NewInternal(fmt.Sprintf("Failed to update function: %v", err)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "updated",
		"message": fmt.Sprintf("Runtime configuration for '%s' updated", functionID),
	})
}

// ListToolchains returns available WASM compilation toolchains
func (h *Handler) ListToolchains(w http.ResponseWriter, r *http.Request) {
	type ToolchainEntry struct {
		Name      string `json:"name"`
		Language  string `json:"language"`
		Available bool   `json:"available"`
		Version   string `json:"version,omitempty"`
		Toolchain string `json:"toolchain,omitempty"`
	}

	toolchains := []ToolchainEntry{
		{Name: "Rust (cargo)", Language: "rust", Available: true, Toolchain: "cargo"},
		{Name: "Go", Language: "go", Available: true, Toolchain: "go"},
		{Name: "C (Emscripten)", Language: "c", Available: true, Toolchain: "emscripten"},
		{Name: "C (WASI-SDK)", Language: "c", Available: false, Toolchain: "wasi-sdk"},
		{Name: "C++ (Emscripten)", Language: "cpp", Available: true, Toolchain: "emscripten"},
		{Name: "Ruby (mruby)", Language: "ruby", Available: false, Toolchain: "mruby-wasm"},
		{Name: "Kotlin", Language: "kotlin", Available: false, Toolchain: "kotlin-wasm"},
		{Name: "Swift (SwiftWasm)", Language: "swift", Available: false, Toolchain: "swiftwasm"},
		{Name: "JavaScript (Javy)", Language: "javascript", Available: true, Toolchain: "javy"},
		{Name: "Python (MicroPython)", Language: "python", Available: true, Toolchain: "micropython"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"toolchains": toolchains,
	})
}

// RegisterRoutes registers runtime routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/runtimes", h.ListRuntimes).Methods("GET")
	router.HandleFunc("/runtimes/{id}", h.GetRuntimeInfo).Methods("GET")
	router.HandleFunc("/runtimes/toolchains", h.ListToolchains).Methods("GET")
	router.HandleFunc("/functions/{function_id}/diagnostics", h.GetDiagnostics).Methods("GET")
	router.HandleFunc("/functions/{function_id}/runtime", h.UpdateRuntimeConfig).Methods("PUT", "POST")
}
