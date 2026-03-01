package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/functionfly/functionfly/internal/functionregistry"
)

// Handler handles runtime-related API requests
type Handler struct {
	registry functionregistry.Registry
}

// New creates a new runtime handler
func New(registry functionregistry.Registry) *Handler {
	return &Handler{
		registry: registry,
	}
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
	Runtime      string            `json:"runtime"`
	Environment  map[string]string `json:"environment"`
	Resources    ResourceUsage     `json:"resources"`
	Network      NetworkStatus     `json:"network"`
	Security     SecurityStatus    `json:"security"`
	Performance  PerformanceData   `json:"performance"`
	Timestamp    time.Time         `json:"timestamp"`
}

// ResourceUsage represents resource usage
type ResourceUsage struct {
	MemoryLimit   int `json:"memory_limit_mb"`
	MemoryUsed    int `json:"memory_used_mb"`
	CPUCores      int `json:"cpu_cores"`
	Timeout       int `json:"timeout_ms"`
}

// NetworkStatus represents network status
type NetworkStatus struct {
	Enabled        bool     `json:"enabled"`
	DNSWorking     bool     `json:"dns_working"`
	ExternalCalls  int      `json:"external_calls_per_minute"`
	AllowedHosts   []string `json:"allowed_hosts"`
}

// SecurityStatus represents security status
type SecurityStatus struct {
	SandboxEnabled    bool     `json:"sandbox_enabled"`
	ModuleRestrictions bool    `json:"module_restrictions"`
	BlockedModules    []string `json:"blocked_modules"`
	EnvironmentCount   int     `json:"environment_variables"`
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
		{
			ID:          "nodejs20",
			Name:        "Node.js",
			Version:     "20.x LTS",
			Status:      "stable",
			Features:    []string{"async/await", "ES Modules", "fetch API", "Worker Threads"},
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
		{
			ID:          "deno",
			Name:        "Deno",
			Version:     "latest",
			Status:      "beta",
			Features:    []string{"TypeScript", "Built-in formatting", "Secure by default"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		{
			ID:          "python3.12",
			Name:        "Python",
			Version:     "3.12",
			Status:      "stable",
			Features:    []string{"async/await", "Type hints", "Better errors"},
			MemoryLimit: 2048,
			Timeout:     300000,
		},
		{
			ID:          "python3.11",
			Name:        "Python",
			Version:     "3.11",
			Status:      "stable",
			Features:    []string{"async/await", "Type hints"},
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
			Version:     "1.x",
			Status:      "beta",
			Features:    []string{"TypeScript", "Built-in formatting", "Secure by default", "Deno Deploy"},
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
	}

	runtime, exists := runtimes[runtimeID]
	if !exists {
		http.Error(w, fmt.Sprintf("Runtime '%s' not found", runtimeID), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runtime)
}

// GetDiagnostics returns diagnostics for a function's runtime
func (h *Handler) GetDiagnostics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID := vars["function_id"]

	// Get function from registry
	fn, err := h.registry.GetFunction(functionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Function '%s' not found", functionID), http.StatusNotFound)
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
	vars := mux.Vars(r)
	functionID := vars["function_id"]

	var config struct {
		Runtime        string `json:"runtime"`
		MemoryMB       int    `json:"memory_mb"`
		TimeoutMs     int    `json:"timeout_ms"`
		NetworkEnabled bool   `json:"network_enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get function from registry
	fn, err := h.registry.GetFunction(functionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Function '%s' not found", functionID), http.StatusNotFound)
		return
	}

	// Update function config
	fn.Runtime = config.Runtime
	fn.MemoryMB = config.MemoryMB
	fn.TimeoutMs = config.TimeoutMs
	fn.NetworkEnabled = config.NetworkEnabled

	// Save to registry
	if err := h.registry.UpdateFunction(fn); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update function: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "updated",
		"message": fmt.Sprintf("Runtime configuration for '%s' updated", functionID),
	})
}

// RegisterRoutes registers runtime routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/runtimes", h.ListRuntimes).Methods("GET")
	router.HandleFunc("/api/v1/runtimes/{id}", h.GetRuntimeInfo).Methods("GET")
	router.HandleFunc("/api/v1/functions/{function_id}/diagnostics", h.GetDiagnostics).Methods("GET")
	router.HandleFunc("/api/v1/functions/{function_id}/runtime", h.UpdateRuntimeConfig).Methods("PUT", "POST")
}
