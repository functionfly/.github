// Package wasm provides WebAssembly runtime support for FunctionFly
// This file contains browser-specific security configuration
package wasm

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Browser WASM defaults - designed for client-side execution
const (
	DefaultBrowserMaxWasmModuleSize uint32 = 10 * 1024 * 1024 // 10MB max WASM binary
	DefaultBrowserMaxInputSize      uint32 = 1 * 1024 * 1024  // 1MB max input
	DefaultBrowserMaxOutputSize    uint32 = 1 * 1024 * 1024 // 1MB max output
	DefaultBrowserExecutionTimeout         = 30 * time.Second // 30s browser timeout
	DefaultBrowserMaxMemory         uint32 = 256 * 1024 * 1024 // 256MB (browser heap)
	DefaultBrowserEnableNetwork     bool   = true              // Allow fetch calls
	DefaultBrowserEnableSharedArrayBuffer bool = false         // Disable SharedArrayBuffer (Spectre)
)

// BrowserWasmSecurityConfig defines security limits for Browser Native WASM execution
// Unlike server-side WASM, browser execution relies on the browser's sandbox for isolation
type BrowserWasmSecurityConfig struct {
	// MaxWasmModuleSize is the maximum WASM binary size in bytes (default: 10MB)
	MaxWasmModuleSize uint32 `json:"max_wasm_module_size"`

	// MaxInputSize is the maximum input size in bytes (default: 1MB)
	MaxInputSize uint32 `json:"max_input_size"`

	// MaxOutputSize is the maximum output size in bytes (default: 1MB)
	MaxOutputSize uint32 `json:"max_output_size"`

	// ExecutionTimeout is the maximum execution time (default: 30s)
	// Note: Browsers typically kill WASM after 30-60 seconds
	ExecutionTimeout time.Duration `json:"execution_timeout"`

	// MaxMemory is the maximum memory allocation in bytes (default: 256MB)
	// Note: This is a soft limit; browsers control actual memory
	MaxMemory uint32 `json:"max_memory"`

	// EnableNetworkAccess enables fetch() calls from WASM (default: true)
	// When false, WASM can only compute locally
	EnableNetworkAccess bool `json:"enable_network_access"`

	// AllowedOrigins restricts CORS for network requests (default: all)
	// Only applies when EnableNetworkAccess is true
	AllowedOrigins []string `json:"allowed_origins"`

	// EnableSharedArrayBuffer enables SharedArrayBuffer (default: false)
	// WARNING: Enabling this has Spectre/variant 1 security implications
	// Only enable if you understand the security implications
	EnableSharedArrayBuffer bool `json:"enable_shared_array_buffer"`

	// ValidateModules enables WASM module validation before execution (default: true)
	ValidateModules bool `json:"validate_modules"`

	// EnableStackTracking enables stack trace collection on errors (default: true)
	EnableStackTracking bool `json:"enable_stack_tracking"`
}

// NewDefaultBrowserWasmSecurityConfig returns a browser WASM security config with sensible defaults
func NewDefaultBrowserWasmSecurityConfig() *BrowserWasmSecurityConfig {
	return &BrowserWasmSecurityConfig{
		MaxWasmModuleSize:       DefaultBrowserMaxWasmModuleSize,
		MaxInputSize:            DefaultBrowserMaxInputSize,
		MaxOutputSize:           DefaultBrowserMaxOutputSize,
		ExecutionTimeout:        DefaultBrowserExecutionTimeout,
		MaxMemory:               DefaultBrowserMaxMemory,
		EnableNetworkAccess:    DefaultBrowserEnableNetwork,
		AllowedOrigins:         []string{}, // Empty means all
		EnableSharedArrayBuffer: DefaultBrowserEnableSharedArrayBuffer,
		ValidateModules:         true,
		EnableStackTracking:     true,
	}
}

// NewBrowserWasmSecurityConfigFromEnv creates a browser WASM security config from environment variables
func NewBrowserWasmSecurityConfigFromEnv() *BrowserWasmSecurityConfig {
	config := NewDefaultBrowserWasmSecurityConfig()

	// BROWSER_WASM_MAX_MODULE_SIZE - max WASM module size in MB
	if v := os.Getenv("BROWSER_WASM_MAX_MODULE_SIZE"); v != "" {
		if mb, err := strconv.ParseUint(v, 10, 32); err == nil {
			config.MaxWasmModuleSize = uint32(mb) * 1024 * 1024
		}
	}

	// BROWSER_WASM_MAX_INPUT_SIZE - max input size in MB
	if v := os.Getenv("BROWSER_WASM_MAX_INPUT_SIZE"); v != "" {
		if mb, err := strconv.ParseUint(v, 10, 32); err == nil {
			config.MaxInputSize = uint32(mb) * 1024 * 1024
		}
	}

	// BROWSER_WASM_MAX_OUTPUT_SIZE - max output size in MB
	if v := os.Getenv("BROWSER_WASM_MAX_OUTPUT_SIZE"); v != "" {
		if mb, err := strconv.ParseUint(v, 10, 32); err == nil {
			config.MaxOutputSize = uint32(mb) * 1024 * 1024
		}
	}

	// BROWSER_WASM_TIMEOUT - execution timeout (e.g., "30s", "1m")
	if v := os.Getenv("BROWSER_WASM_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.ExecutionTimeout = d
		}
	}

	// BROWSER_WASM_ENABLE_NETWORK - enable/disable network access
	if v := os.Getenv("BROWSER_WASM_ENABLE_NETWORK"); v != "" {
		config.EnableNetworkAccess = strings.ToLower(v) == "true" || v == "1"
	}

	// BROWSER_WASM_ALLOWED_ORIGINS - comma-separated list of allowed origins
	if v := os.Getenv("BROWSER_WASM_ALLOWED_ORIGINS"); v != "" {
		config.AllowedOrigins = strings.Split(v, ",")
		for i := range config.AllowedOrigins {
			config.AllowedOrigins[i] = strings.TrimSpace(config.AllowedOrigins[i])
		}
	}

	// BROWSER_WASM_ENABLE_SAB - enable SharedArrayBuffer (security risk)
	if v := os.Getenv("BROWSER_WASM_ENABLE_SAB"); v != "" {
		config.EnableSharedArrayBuffer = strings.ToLower(v) == "true" || v == "1"
	}

	return config
}

// ValidateWasmModuleSize checks if the WASM module size is within limits
func (c *BrowserWasmSecurityConfig) ValidateWasmModuleSize(size uint32) bool {
	return size <= c.MaxWasmModuleSize
}

// ValidateInputSize checks if the input size is within limits
func (c *BrowserWasmSecurityConfig) ValidateInputSize(size uint32) bool {
	return size <= c.MaxInputSize
}

// ValidateOutputSize checks if the output size is within limits
func (c *BrowserWasmSecurityConfig) ValidateOutputSize(size uint32) bool {
	return size <= c.MaxOutputSize
}

// IsOriginAllowed checks if an origin is allowed for CORS
// Returns true if AllowedOrigins is empty (no restriction) or origin is in the list
func (c *BrowserWasmSecurityConfig) IsOriginAllowed(origin string) bool {
	if len(c.AllowedOrigins) == 0 {
		return true
	}
	origin = strings.ToLower(origin)
	for _, allowed := range c.AllowedOrigins {
		allowed = strings.ToLower(allowed)
		if origin == allowed || strings.HasSuffix(origin, "."+allowed) {
			return true
		}
	}
	return false
}

// GetCORSHeaders returns appropriate CORS headers based on the security config
func (c *BrowserWasmSecurityConfig) GetCORSHeaders() map[string]string {
	headers := map[string]string{
		"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Requested-With",
		"Access-Control-Max-Age": "86400", // 24 hours
	}

	if len(c.AllowedOrigins) == 0 {
		headers["Access-Control-Allow-Origin"] = "*"
	} else {
		// For multiple origins, we cannot use "*" - must use a specific origin
		// or implement CORS proxy logic. Here we return the first allowed origin.
		headers["Access-Control-Allow-Origin"] = c.AllowedOrigins[0]
	}

	return headers
}

// Clone creates a copy of the browser WASM security config
func (c *BrowserWasmSecurityConfig) Clone() *BrowserWasmSecurityConfig {
	origins := make([]string, len(c.AllowedOrigins))
	copy(origins, c.AllowedOrigins)

	return &BrowserWasmSecurityConfig{
		MaxWasmModuleSize:       c.MaxWasmModuleSize,
		MaxInputSize:            c.MaxInputSize,
		MaxOutputSize:           c.MaxOutputSize,
		ExecutionTimeout:        c.ExecutionTimeout,
		MaxMemory:               c.MaxMemory,
		EnableNetworkAccess:     c.EnableNetworkAccess,
		AllowedOrigins:          origins,
		EnableSharedArrayBuffer: c.EnableSharedArrayBuffer,
		ValidateModules:         c.ValidateModules,
		EnableStackTracking:     c.EnableStackTracking,
	}
}

// SecurityWarnings returns a list of security warnings for this config
func (c *BrowserWasmSecurityConfig) SecurityWarnings() []string {
	var warnings []string

	if c.EnableSharedArrayBuffer {
		warnings = append(warnings, "WARNING: SharedArrayBuffer is enabled. This can be exploited by Spectre/variant 1 attacks. Only enable if absolutely necessary and understand the security implications.")
	}

	if len(c.AllowedOrigins) == 0 && c.EnableNetworkAccess {
		warnings = append(warnings, "INFO: Network access is enabled with no origin restrictions. All origins can make fetch() calls from WASM.")
	}

	if c.MaxWasmModuleSize > 50*1024*1024 {
		warnings = append(warnings, "INFO: Large WASM module size limit (>50MB) may indicate a misconfiguration.")
	}

	if c.ExecutionTimeout > 60*time.Second {
		warnings = append(warnings, "INFO: Execution timeout exceeds browser recommendations (>60s). Browser may kill the worker before timeout.")
	}

	return warnings
}
