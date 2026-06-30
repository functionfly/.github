// Package wasm provides WebAssembly runtime support for FunctionFly
// This file contains security configuration types (works with and without CGO)
package wasm

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Default values for security configuration
const (
	DefaultMaxMemory       uint32 = 64 * 1024 * 1024       // 64MB
	DefaultMaxExecutionTime        = 30 * time.Second       // 30s
	DefaultMaxInstructions uint64  = 100_000_000          // 100M instructions
	DefaultPoolSize         int    = 10                    // Instances per tenant
	DefaultWASIEnabled       bool   = true
	DefaultAllowRawPointers        = false

	// AI Inference defaults
	DefaultAIInferenceEnabled     = false
	DefaultAIGatewayURL           = "http://localhost:8082"
	DefaultAIMaxModelSizeMB       = 100
	DefaultAIDefaultModel         = "gpt-4"
	DefaultAITimeoutSeconds       = 60
	DefaultAIEnableCaching        = true
)

// WASMSecurityConfig defines security limits for WASM execution
type WASMSecurityConfig struct {
	// MaxMemory is the maximum memory allocation in bytes (default: 64MB)
	MaxMemory uint32 `json:"max_memory"`

	// MaxExecutionTime is the maximum execution time allowed (default: 30s)
	MaxExecutionTime time.Duration `json:"max_execution_time"`

	// MaxInstructions is the maximum number of CPU instructions (default: 100M)
	MaxInstructions uint64 `json:"max_instructions"`

	// EnableWASI enables WASI (WebAssembly System Interface) support (default: true)
	EnableWASI bool `json:"enable_wasi"`

	// AllowRawPointers allows raw pointer access in WASM (default: false)
	AllowRawPointers bool `json:"allow_raw_pointers"`

	// AllowedDomains is an allowlist of domains for fetch() calls
	AllowedDomains []string `json:"allowed_domains"`

	// InstancePoolPerTenant enables per-tenant instance isolation (default: true)
	InstancePoolPerTenant bool `json:"instance_pool_per_tenant"`

	// PoolSize is the maximum number of instances per tenant (default: 10)
	PoolSize int `json:"pool_size"`

	// EnableDeterministic enables deterministic execution mode (default: false)
	EnableDeterministic bool `json:"enable_deterministic"`

	// DisableDeterministic disables deterministic execution mode (default: false)
	// When false, fuel/instruction metering is enabled by default
	DisableDeterministic bool `json:"disable_deterministic"`

	// MaxInputSize is the maximum input size in bytes (default: 1MB)
	MaxInputSize uint32 `json:"max_input_size"`

	// MaxOutputSize is the maximum output size in bytes (default: 1MB)
	MaxOutputSize uint32 `json:"max_output_size"`

	// AIInference configures AI inference capabilities (default: disabled)
	AIInference AIInferenceConfig `json:"ai_inference"`
}

// AIInferenceConfig defines configuration for AI inference via the AI Gateway
type AIInferenceConfig struct {
	// Enabled enables AI inference capability (default: false)
	Enabled bool `json:"enabled"`

	// GatewayURL is the AI Gateway endpoint URL (default: "http://localhost:8082")
	GatewayURL string `json:"gateway_url"`

	// MaxModelSizeMB is the maximum model response size in MB (default: 100)
	MaxModelSizeMB int `json:"max_model_size_mb"`

	// DefaultModel is the default AI model to use (default: "gpt-4")
	DefaultModel string `json:"default_model"`

	// TimeoutSeconds is the inference timeout in seconds (default: 60)
	TimeoutSeconds int `json:"timeout_seconds"`

	// EnableCaching enables response caching for inference calls (default: true)
	EnableCaching bool `json:"enable_caching"`
}

// NewDefaultSecurityConfig returns a security config with default values
func NewDefaultSecurityConfig() *WASMSecurityConfig {
	return &WASMSecurityConfig{
		MaxMemory:             DefaultMaxMemory,
		MaxExecutionTime:      DefaultMaxExecutionTime,
		MaxInstructions:       DefaultMaxInstructions,
		EnableWASI:            DefaultWASIEnabled,
		AllowRawPointers:      DefaultAllowRawPointers,
		InstancePoolPerTenant: true,
		PoolSize:              DefaultPoolSize,
		EnableDeterministic:    false,
		MaxInputSize:          1024 * 1024,  // 1MB
		MaxOutputSize:         1024 * 1024,  // 1MB
		AIInference: AIInferenceConfig{
			Enabled:          DefaultAIInferenceEnabled,
			GatewayURL:       DefaultAIGatewayURL,
			MaxModelSizeMB:   DefaultAIMaxModelSizeMB,
			DefaultModel:     DefaultAIDefaultModel,
			TimeoutSeconds:   DefaultAITimeoutSeconds,
			EnableCaching:    DefaultAIEnableCaching,
		},
	}
}

// ValidateOutputSize checks if the output size is within configured limits
func (c *WASMSecurityConfig) ValidateOutputSize(size uint32) bool {
	return size <= c.MaxOutputSize
}

// NewSecurityConfigFromEnv creates a security config from environment variables
func NewSecurityConfigFromEnv() *WASMSecurityConfig {
	config := NewDefaultSecurityConfig()

	// WASM_MAX_MEMORY - max memory in MB
	if v := os.Getenv("WASM_MAX_MEMORY"); v != "" {
		if mb, err := strconv.ParseUint(v, 10, 32); err == nil {
			config.MaxMemory = uint32(mb) * 1024 * 1024
		}
	}

	// WASM_MAX_TIMEOUT - max execution timeout (e.g., "30s", "1m")
	if v := os.Getenv("WASM_MAX_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.MaxExecutionTime = d
		}
	}

	// WASM_POOL_SIZE - instances per tenant
	if v := os.Getenv("WASM_POOL_SIZE"); v != "" {
		if size, err := strconv.Atoi(v); err == nil && size > 0 {
			config.PoolSize = size
		}
	}

	// WASM_ENABLE_DETERMINISTIC - enable deterministic mode
	if v := os.Getenv("WASM_ENABLE_DETERMINISTIC"); v != "" {
		config.EnableDeterministic = strings.ToLower(v) == "true" || v == "1"
	}

	// WASM_DISABLE_DETERMINISTIC - disable deterministic mode (overrides WASM_ENABLE_DETERMINISTIC)
	if v := os.Getenv("WASM_DISABLE_DETERMINISTIC"); v != "" {
		config.DisableDeterministic = strings.ToLower(v) == "true" || v == "1"
	}

	// WASM_ENABLE_WASI - enable/disable WASI
	if v := os.Getenv("WASM_ENABLE_WASI"); v != "" {
		config.EnableWASI = strings.ToLower(v) == "true" || v == "1"
	}

	// WASM_ALLOWED_DOMAINS - comma-separated list of allowed domains
	if v := os.Getenv("WASM_ALLOWED_DOMAINS"); v != "" {
		config.AllowedDomains = strings.Split(v, ",")
		for i := range config.AllowedDomains {
			config.AllowedDomains[i] = strings.TrimSpace(config.AllowedDomains[i])
		}
	}

	// WASM_MAX_INPUT_SIZE - max input size in MB
	if v := os.Getenv("WASM_MAX_INPUT_SIZE"); v != "" {
		if mb, err := strconv.ParseUint(v, 10, 32); err == nil {
			config.MaxInputSize = uint32(mb) * 1024 * 1024
		}
	}

	// WASM_MAX_OUTPUT_SIZE - max output size in MB
	if v := os.Getenv("WASM_MAX_OUTPUT_SIZE"); v != "" {
		if mb, err := strconv.ParseUint(v, 10, 32); err == nil {
			config.MaxOutputSize = uint32(mb) * 1024 * 1024
		}
	}

	return config
}

// IsDomainAllowed checks if a domain is in the allowed domains list.
// If AllowedDomains is empty, all domains are DENIED (default-deny).
// Use AllowedDomains = ["*"] to explicitly allow all domains.
func (c *WASMSecurityConfig) IsDomainAllowed(domain string) bool {
	if len(c.AllowedDomains) == 0 {
		return false // Default-deny: no domains allowed when list is empty
	}

	domain = strings.ToLower(domain)
	for _, allowed := range c.AllowedDomains {
		allowed = strings.ToLower(allowed)
		if allowed == "*" {
			return true // Explicit wildcard — allow everything
		}
		if domain == allowed || strings.HasSuffix(domain, "."+allowed) {
			return true
		}
	}
	return false
}

// ValidateInputSize checks if the input size is within limits
func (c *WASMSecurityConfig) ValidateInputSize(size uint32) bool {
	return size <= c.MaxInputSize
}

// GetChunkBufferSize returns the recommended chunk buffer size for streaming
// This is based on the MaxInputSize but ensures at least a reasonable chunk size
func (c *WASMSecurityConfig) GetChunkBufferSize() int {
	const minChunkSize = 64 * 1024 // 64KB minimum
	maxInput := int(c.MaxInputSize)
	if maxInput <= 0 {
		return minChunkSize
	}
	// Use 1/10th of max input, but at least the minimum
	chunkSize := maxInput / 10
	if chunkSize < minChunkSize {
		return minChunkSize
	}
	return chunkSize
}

// Clone creates a copy of the security config
func (c *WASMSecurityConfig) Clone() *WASMSecurityConfig {
	domains := make([]string, len(c.AllowedDomains))
	copy(domains, c.AllowedDomains)

	return &WASMSecurityConfig{
		MaxMemory:             c.MaxMemory,
		MaxExecutionTime:     c.MaxExecutionTime,
		MaxInstructions:      c.MaxInstructions,
		EnableWASI:           c.EnableWASI,
		AllowRawPointers:      c.AllowRawPointers,
		AllowedDomains:        domains,
		InstancePoolPerTenant: c.InstancePoolPerTenant,
		PoolSize:              c.PoolSize,
		EnableDeterministic:   c.EnableDeterministic,
		DisableDeterministic:   c.DisableDeterministic,
		MaxInputSize:          c.MaxInputSize,
		MaxOutputSize:         c.MaxOutputSize,
		AIInference:          c.AIInference, // Deep copy is safe for struct values
	}
}
