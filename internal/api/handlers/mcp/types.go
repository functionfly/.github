// Package mcp implements the MCP (Model Context Protocol) Function Registry.
//
// Surface area:
//   GET  /v1/mcp/manifest            public, cacheable server identity
//   GET  /v1/mcp/tools               public, cacheable tool index (SEO anchor)
//   POST /v1/mcp                     JSON-RPC 2.0 transport (streamable-HTTP)
//
// Reference: https://modelcontextprotocol.io specification 2025-03-26.
//
// Design notes:
//   - We re-use the existing registry search/storage layer (registry.RegistryRepository
//     .SearchFunctions) so /v1/mcp/tools and /.well-known/functionfly.json can never
//     disagree about which functions exist.
//   - Per-function MCP gating lives in registry_function_mcp_settings, NOT in
//     registry_functions.settings, to avoid JSONB drift.
//   - We re-use the existing SecureExecution middleware to enforce tenant/visibility/
//     rate-limit/verification invariants on tools/call. No new trust boundary.
//   - The JSON-RPC envelope is implemented in a small, dependency-free package so it
//     is easy to unit-test and easy to vendor into the @functionfly/mcp-server npm
//     package (which proxies this same wire format over stdio).
package mcp

import (
	"github.com/functionfly/functionfly/internal/storage/registry"
)

// Server identity — exposed via /v1/mcp/manifest.
const (
	ServerName             = "FunctionFly MCP Function Registry"
	ServerVersion          = "1.0.0"
	ProtocolVersion        = "2025-03-26" // MCP spec version we implement
	ProviderName           = "functionfly"
	ToolsListDefaultLimit  = 100
	ToolsListMaxLimit      = 500
	ToolsCallMaxArgsBytes  = 256 * 1024 // 256 KiB — defend against huge JSON args
	ToolsCallMaxDepth      = 10
	ToolNameMaxLength      = 64
	ManifestCacheMaxAge    = 300
	ToolsIndexCacheMaxAge  = 60
)

// Errors — JSON-RPC 2.0 reserved range [-32768, -32000].
// Server-defined range starts at -32000 per spec.
const (
	ErrCodeParseError     = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternalError  = -32603

	// Server-defined (per MCP spec, application errors fall in -32000 to -32099).
	ErrCodeToolNotFound      = -32001
	ErrCodeToolDisabled      = -32002
	ErrCodeAuthRequired      = -32003
	ErrCodeRateLimited       = -32004
	ErrCodeExecutionFailed   = -32005
	ErrCodeFunctionPrivate   = -32006
	ErrCodeMalwareBlocked    = -32007
	ErrCodeOriginNotAllowed  = -32008
	ErrCodePayloadTooLarge   = -32009
)

// Standard JSON-RPC 2.0 error messages.
const (
	MsgParseError          = "Parse error"
	MsgInvalidRequest      = "Invalid Request"
	MsgMethodNotFound      = "Method not found"
	MsgInvalidParams       = "Invalid params"
	MsgInternalError       = "Internal error"
	MsgToolNotFound        = "Tool not found"
	MsgToolDisabled        = "Tool is not MCP-enabled"
	MsgAuthRequired        = "Authentication required"
	MsgRateLimited         = "Rate limit exceeded"
	MsgExecutionFailed     = "Tool execution failed"
	MsgFunctionPrivate     = "Function is not public"
	MsgMalwareBlocked      = "Function failed security scan"
	MsgOriginNotAllowed    = "Origin not allowed"
	MsgPayloadTooLarge     = "Request payload too large"
)

// Transport identifiers (stored in registry_function_mcp_settings.transports).
const (
	TransportStreamableHTTP = "streamable-http"
	TransportStdio          = "stdio"
)

// CallerIdentity is what the auth middleware attaches to the request context
// after a successful Bearer-token check. The MCP handlers read this to record
// who invoked a tool.
type CallerIdentity struct {
	UserID    string
	TenantID  string
	APIKeyID  string
	AuthType  string // "apikey" | "session" | "anonymous"
	TokenHash string // SHA-256 prefix of the bearer (used as caller_id in invocations log)
}

// ToolsListFilter is a function-local filter struct used by the public
// /v1/mcp/tools index endpoint (and the JSON-RPC tools/list method).
type ToolsListFilter struct {
	Category string
	Runtime  string
	MinTrust float64
	Search   string
	Cursor   string // base64(function_id) — exclusive lower bound
	Limit    int
	Offset   int
}

// ToolsListResult is the response shape for /v1/mcp/tools (no JSON-RPC envelope).
type ToolsListResult struct {
	Tools       []ToolDefinition `json:"tools"`
	NextCursor  string           `json:"next_cursor,omitempty"`
	Total       int              `json:"total"`
	GeneratedAt int64            `json:"generated_at"`
}

// ToMCPSettings exposes the per-function settings that the JSON-RPC handlers
// need without forcing them to import the storage package. This keeps the
// handler interface narrow and testable.
type ToMCPSettings struct {
	FunctionID         string
	Author             string
	Name               string
	Version            string
	Enabled            bool
	Transports         []string
	ExposeInputSchema  bool
	ExposeOutputSchema bool
	ToolNameOverride   string
	RateLimitPerMin    int
	AllowlistOrigins   []string
	VerifiedMCP        bool
	TrustScore         float64
	TrustTier          string
	Title              string
	Description        string
	Category           string
	Tags               []string
	Runtime            string
	Manifest           []byte
}

// ToToolName returns the MCP-safe tool name (double-underscore delimiter).
// If the publisher set a tool_name_override AND it is still unique, it is
// preferred. Otherwise we fall back to "{author}__{name}".
//
// We always apply the same sanitization as wellknown/handler.go to keep
// names consistent across the platform. The sanitized form is what we put
// in the wire.
func (s *ToMCPSettings) ToToolName() string {
	if s.ToolNameOverride != "" {
		return SanitizeToolName(s.ToolNameOverride)
	}
	return SanitizeToolName(s.Author + "__" + s.Name)
}

// SettingsProvider is the minimal surface the MCP handlers need from storage.
// Defined as an interface so tests can mock it.
type SettingsProvider interface {
	GetMCPSettings(ctx interface{ Done() <-chan struct{} }, functionID interface{}) (*registry.MCPSettings, error)
}

// (SettingsProvider is declared but the concrete adapter lives in routes.go —
// we keep this file free of storage imports for testability.)
