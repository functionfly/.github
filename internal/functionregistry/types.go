package functionregistry

import (
	"encoding/json"
	"time"
)

// Supported runtimes
const (
	RuntimeNode18    = "node18"
	RuntimeNode20    = "node20"
	RuntimePython311 = "python3.11"
	RuntimePython312 = "python3.12"
	RuntimeGo121     = "go1.21"
	RuntimeRust175   = "rust1.75"
)

// FunctionVisibility types
const (
	VisibilityPublic   = "public"
	VisibilityPrivate  = "private"
	VisibilityUnlisted = "unlisted"
)

// Execution outcome types
const (
	OutcomeSuccess = "success"
	OutcomeError   = "error"
	OutcomeTimeout = "timeout"
)

// Error codes for standardized error responses
const (
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeTimeout          = "TIMEOUT"
	ErrCodeMemoryExceeded   = "MEMORY_EXCEEDED"
	ErrCodeRuntimeError     = "RUNTIME_ERROR"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeRateLimited      = "RATE_LIMITED"
	ErrCodeInternalError    = "INTERNAL_ERROR"
	ErrCodeCapabilityDenied = "CAPABILITY_DENIED"
)

// AllowedCapabilities defines the set of capabilities a function can declare
var AllowedCapabilities = []string{
	"fetch:read",   // HTTP GET requests
	"fetch:write",  // HTTP POST/PUT/PATCH/DELETE requests
	"crypto",       // Cryptographic operations (hash, sign, etc.)
	"cache:read",   // Read from cache
	"cache:write",  // Write to cache
	"kv",           // Key-value store access
	"webhook",      // Webhook triggers
	"email",        // Email sending
	"storage",      // File storage access
	"ai",           // AI/ML inference
	"external_api", // External API access
}

// IsValidCapability checks if a capability is in the allowed list
func IsValidCapability(cap string) bool {
	for _, allowed := range AllowedCapabilities {
		if cap == allowed {
			return true
		}
	}
	return false
}

// HasCapability checks if a capability slice contains a specific capability
func HasCapability(caps []string, cap string) bool {
	for _, c := range caps {
		if c == cap {
			return true
		}
	}
	return false
}

// FunctionManifest represents the functionfly.json manifest
type FunctionManifest struct {
	Name                 string   `json:"name" validate:"required,pattern=^[a-z0-9-]+$"`
	Version              string   `json:"version" validate:"required,semver"`
	Runtime              string   `json:"runtime" validate:"required"`
	Title                string   `json:"title,omitempty"`
	Description          string   `json:"description,omitempty"`
	Input                *IOType  `json:"input,omitempty"`
	Output               *IOType  `json:"output,omitempty"`
	TimeoutMs            int      `json:"timeout_ms,omitempty" validate:"min=100,max=30000"`
	MemoryMB             int      `json:"memory_mb,omitempty" validate:"min=32,max=1024"`
	Deterministic        bool     `json:"deterministic,omitempty"`
	SideEffects          string   `json:"side_effects,omitempty"` // none | network | external_state
	Idempotent           bool     `json:"idempotent,omitempty"`
	CacheTTL             int      `json:"cache_ttl,omitempty" validate:"min=0,max=86400"`
	Public               bool     `json:"public,omitempty"`
	PricePerCall         float64  `json:"price_per_call,omitempty" validate:"min=0"`
	Category             string   `json:"category,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
	PlaygroundVisibility string   `json:"playground_visibility,omitempty"`
	Capabilities         []string `json:"capabilities,omitempty" validate:"dive,oneof=fetch:read fetch:write crypto cache:read cache:write kv webhook email storage ai external_api"`
}

// IOType defines input/output type specifications
type IOType struct {
	Type       string          `json:"type"`
	Properties json.RawMessage `json:"properties,omitempty"`
	Example    json.RawMessage `json:"example,omitempty"`
	Schema     json.RawMessage `json:"schema,omitempty"`
	Required   RequiredField   `json:"required,omitempty"` // bool or JSON Schema "required" array of property names
}

// RequiredField accepts either a bool (body required) or []string (required property names) in JSON
type RequiredField struct {
	Bool  bool
	Array []string
}

// UnmarshalJSON allows "required" to be either a bool or an array of strings (JSON Schema style)
func (r *RequiredField) UnmarshalJSON(data []byte) error {
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		r.Array = arr
		return nil
	}
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		r.Bool = b
		return nil
	}
	return nil // ignore invalid values
}

// IsRequired returns true if the request body / output is required (bool true or any required properties)
func (r *RequiredField) IsRequired() bool {
	if r.Bool {
		return true
	}
	return len(r.Array) > 0
}

// Function represents a function in the registry
type Function struct {
	ID                 string    `json:"id"`
	Author             string    `json:"author"`
	Name               string    `json:"name"`
	LatestVersion      string    `json:"latest_version,omitempty"`
	Title              string    `json:"title,omitempty"`
	Description        string    `json:"description,omitempty"`
	Category           string    `json:"category,omitempty"`
	Tags               []string  `json:"tags,omitempty"`
	Visibility         string    `json:"visibility"`
	PricePerCall       float64   `json:"price_per_call"`
	PopularityScore    int       `json:"popularity_score"`
	ReliabilityScore   float64   `json:"reliability_score"`
	DeterministicScore float64   `json:"deterministic_score"`
	TenantID           *string   `json:"tenant_id,omitempty"`
	OwnerUserID        *string   `json:"owner_user_id,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// FunctionVersion represents a specific version of a function
type FunctionVersion struct {
	ID            string          `json:"id"`
	FunctionID    string          `json:"function_id"`
	Version       string          `json:"version"`
	Manifest      json.RawMessage `json:"manifest"`
	Runtime       string          `json:"runtime"`
	TimeoutMs     int             `json:"timeout_ms"`
	MemoryMB      int             `json:"memory_mb"`
	Deterministic bool            `json:"deterministic"`
	SideEffects   string          `json:"side_effects"` // none | network | external_state
	Idempotent    bool            `json:"idempotent"`
	CacheTTL      int             `json:"cache_ttl"`
	DeploymentID  *string         `json:"deployment_id,omitempty"`
	BackendID     *string         `json:"backend_id,omitempty"`
	ContentHash   string          `json:"content_hash,omitempty"`
	PublishedAt   time.Time       `json:"published_at"`
}

// FunctionExecution represents a single execution record
type FunctionExecution struct {
	ID         string    `json:"id"`
	FunctionID string    `json:"function_id"`
	Version    string    `json:"version"`
	DurationMs int       `json:"duration_ms"`
	StatusCode int       `json:"status_code"`
	Cached     bool      `json:"cached"`
	Outcome    string    `json:"outcome"`
	ErrorCode  *string   `json:"error_code,omitempty"`
	CallerIP   *string   `json:"caller_ip,omitempty"`
	UserAgent  *string   `json:"user_agent,omitempty"`
	GeoCountry *string   `json:"geo_country,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// FunctionRating represents aggregated ratings for a function
type FunctionRating struct {
	ID                 string    `json:"id"`
	FunctionID         string    `json:"function_id"`
	OverallScore       float64   `json:"overall_score"`
	ReliabilityScore   float64   `json:"reliability_score"`
	LatencyScore       float64   `json:"latency_score"`
	DocumentationScore float64   `json:"documentation_score"`
	TotalRatings       int       `json:"total_ratings"`
	SuccessRate        float64   `json:"success_rate"`
	P95LatencyMs       int       `json:"p95_latency_ms"`
	AvgLatencyMs       int       `json:"avg_latency_ms"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ExecutionRequest represents an incoming execution request
type ExecutionRequest struct {
	Author  string          `json:"author"`
	Name    string          `json:"name"`
	Version string          `json:"version,omitempty"` // Empty means latest
	Input   json.RawMessage `json:"input"`
}

// ExecutionResponse is the standard success response format
type ExecutionResponse struct {
	OK          bool            `json:"ok"`
	Data        json.RawMessage `json:"data,omitempty"`
	Cached      bool            `json:"cached"`
	DurationMs  int             `json:"duration_ms"`
	Version     string          `json:"version"`
	ExecutionID *string         `json:"execution_id,omitempty"`
}

// ExecutionError is the standard error response format
type ExecutionError struct {
	OK         bool        `json:"ok"`
	Error      ErrorDetail `json:"error"`
	DurationMs int         `json:"duration_ms"`
	Version    string      `json:"version,omitempty"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PublishRequest represents a request to publish a new function/version
type PublishRequest struct {
	Author     string          `json:"author" validate:"required"`
	Name       string          `json:"name" validate:"required,pattern=^[a-z0-9-]+$"`
	Version    string          `json:"version" validate:"required,semver"`
	Manifest   json.RawMessage `json:"manifest" validate:"required"`
	Source     *FunctionSource `json:"source,omitempty"` // Source code for sandbox execution
	Deployment *DeploymentRef  `json:"deployment,omitempty"`
	TrustLevel string          `json:"trust_level,omitempty"` // "standard", "high", "enterprise"
}

// FunctionSource represents the source code for a function
type FunctionSource struct {
	Code       string            `json:"code" validate:"required"`    // Main source code
	Files      map[string]string `json:"files,omitempty"`             // Additional files (package.json, etc.)
	Runtime    string            `json:"runtime" validate:"required"` // Runtime type (node18, python3.11, wasm, etc.)
	WasmBinary string            `json:"wasm_binary,omitempty"`       // Base64-encoded pre-compiled WASM (for "wasm" runtime)
}

// DeploymentRef references an existing deployment
type DeploymentRef struct {
	DeploymentID string `json:"deployment_id"`
	BackendID    string `json:"backend_id"`
}

// PublishResponse represents the response from a publish operation
type PublishResponse struct {
	OK                 bool   `json:"ok"`
	Function           string `json:"function"` // author/name
	Version            string `json:"version"`
	Message            string `json:"message,omitempty"`
	VerificationStatus string `json:"verification_status,omitempty"`
	// Full function details
	FunctionID    string   `json:"function_id,omitempty"`
	Runtime       string   `json:"runtime,omitempty"`
	TimeoutMs     int      `json:"timeout_ms,omitempty"`
	MemoryMB      int      `json:"memory_mb,omitempty"`
	Capabilities  []string `json:"capabilities,omitempty"`
	Deterministic bool     `json:"deterministic"`
	SideEffects   string   `json:"side_effects,omitempty"`
	Idempotent    bool     `json:"idempotent"`
	CacheTTL      int      `json:"cache_ttl,omitempty"`
	BundleSize    int      `json:"bundle_size,omitempty"`
}

// FunctionInfo represents public function information
type FunctionInfo struct {
	Author        string          `json:"author"`
	Name          string          `json:"name"`
	Version       string          `json:"version"`
	Title         string          `json:"title,omitempty"`
	Description   string          `json:"description,omitempty"`
	Runtime       string          `json:"runtime"`
	Category      string          `json:"category,omitempty"`
	Tags          []string        `json:"tags,omitempty"`
	PricePerCall  float64         `json:"price_per_call"`
	Reliability   float64         `json:"reliability"`
	Deterministic bool            `json:"deterministic"`
	SideEffects   string          `json:"side_effects"` // none | network | external_state
	Idempotent    bool            `json:"idempotent"`
	CacheTTL      int             `json:"cache_ttl"`
	InputType     string          `json:"input_type,omitempty"`
	OutputType    string          `json:"output_type,omitempty"`
	InputExample  json.RawMessage `json:"input_example,omitempty"`
	OutputExample json.RawMessage `json:"output_example,omitempty"`
	// Trust Score fields
	TrustScore        float64 `json:"trust_score"`
	TrustLevel        string  `json:"trust_level"`
	SuccessRate       float64 `json:"success_rate"`
	P50LatencyMs      int     `json:"p50_latency_ms"`
	P95LatencyMs      int     `json:"p95_latency_ms"`
	TimeoutRate       float64 `json:"timeout_rate"`
	ErrorRate         float64 `json:"error_rate"`
	ConsumerDiversity float64 `json:"consumer_diversity"`
	TenantDiversity   int     `json:"tenant_diversity"`
	UserDiversity     int     `json:"user_diversity"`
	// Additional version metadata
	Capabilities     []string `json:"capabilities,omitempty"`
	TimeoutMs        int      `json:"timeout_ms,omitempty"`
	MemoryMB         int      `json:"memory_mb,omitempty"`
	BundleSize       int      `json:"bundle_size,omitempty"`
	SourceHash       string   `json:"source_hash,omitempty"`
	DeploymentID     string   `json:"deployment_id,omitempty"`
	BackendID        string   `json:"backend_id,omitempty"`
	DocumentationURL string   `json:"documentation_url,omitempty"`
	PlaygroundURL    string   `json:"playground_url,omitempty"`
}

// ListFunctionsRequest represents request to list functions
type ListFunctionsRequest struct {
	Author     string `json:"author,omitempty"`
	Category   string `json:"category,omitempty"`
	Tags       string `json:"tags,omitempty"`       // comma-separated
	Visibility string `json:"visibility,omitempty"` // defaults to public
	Limit      int    `json:"limit,omitempty"`      // defaults to 20
	Offset     int    `json:"offset,omitempty"`     // defaults to 0
}

// ListFunctionsResponse represents response with function list
type ListFunctionsResponse struct {
	Functions []FunctionInfo `json:"functions"`
	Total     int            `json:"total"`
	Limit     int            `json:"limit"`
	Offset    int            `json:"offset"`
}

// SearchFunctionsRequest represents a search request
type SearchFunctionsRequest struct {
	Query     string  `json:"query"`                // search text
	Category  string  `json:"category,omitempty"`   // filter by category
	Runtime   string  `json:"runtime,omitempty"`    // filter by runtime
	Tags      string  `json:"tags,omitempty"`       // comma-separated tags
	MinRating float64 `json:"min_rating,omitempty"` // minimum rating
	Limit     int     `json:"limit,omitempty"`      // defaults to 20
	Offset    int     `json:"offset,omitempty"`     // defaults to 0
}

// GetByAddress parses a function address like "author/name@version"
func GetByAddress(address string) (author, name, version string, isLatest bool) {
	// Handle version suffix
	version = "latest"
	isLatest = true

	// Parse author/name@version format
	var base string
	if idx := len(address) - 1; idx >= 0 && address[idx] == '@' {
		return "", "", "", false // Invalid format
	}

	// Find version separator
	for i := len(address) - 1; i >= 0; i-- {
		if address[i] == '@' {
			version = address[i+1:]
			base = address[:i]
			isLatest = version == "latest"
			break
		}
	}

	if base == "" {
		base = address
	}

	// Find author/name separator
	for i := 0; i < len(base); i++ {
		if base[i] == '/' {
			author = base[:i]
			name = base[i+1:]
			return author, name, version, isLatest
		}
	}

	return "", base, version, isLatest
}
