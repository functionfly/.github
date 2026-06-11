package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/functionregistry"
	agenttools "github.com/functionfly/functionfly/internal/agent/tools"
	"github.com/functionfly/functionfly/internal/gateway"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// =============================================================================
// Public interfaces — satisfied by the storage layer + a small adapter
// (see routes.go for the concrete wiring).
// =============================================================================

// FunctionStore is the read-side surface the MCP handler needs to build tool
// definitions. It is defined here as an interface so handlers are easy to
// unit-test.
type FunctionStore interface {
	// GetFunctionByID returns the function row by primary key.
	GetFunctionByID(ctx context.Context, id uuid.UUID) (*registry.RegistryFunction, error)
	// GetFunctionByAuthorName returns the function row. nil + nil means "not found".
	GetFunctionByAuthorName(ctx context.Context, author, name string) (*registry.RegistryFunction, error)
	// GetLatestFunctionVersion returns the version row containing the manifest.
	GetLatestFunctionVersion(ctx context.Context, functionID uuid.UUID) (*registry.RegistryFunctionVersion, error)
	// GetMCPSettings returns the per-function MCP settings (or nil if not configured).
	GetMCPSettings(ctx context.Context, functionID uuid.UUID) (*registry.MCPSettings, error)
	// SearchFunctions returns paginated public functions matching the filter.
	SearchFunctions(ctx context.Context, query, category, runtime string, minRating float64, limit, offset int) ([]registry.RegistryFunction, int, error)
	// ListEnabledMCPSettings returns the slice of MCP-enabled functions.
	ListEnabledMCPSettings(ctx context.Context, category, runtime string, minTrust float64, limit, offset int) ([]registry.MCPSettings, int, error)
	// IncrementMCPInvocationCount bumps the per-function counter.
	IncrementMCPInvocationCount(ctx context.Context, functionID uuid.UUID) error
	// RecordMCPInvocation writes an observability row.
	RecordMCPInvocation(ctx context.Context, rec registry.MCPInvocationRecord) error
}

// Executor is the surface used to actually run a tool call. In production
// it wraps the existing SecureExecution middleware + HandleExecute from the
// registry package. The interface accepts a `raw` JSON body so the executor
// can re-use the REST contract ({"input": ...}) without MCP handlers having
// to know about it.
type Executor interface {
	// ExecuteFunction calls the named function with the given input. The
	// returned bytes are the raw response body (already JSON). StatusCode is
	// the HTTP status the handler should mirror back to the MCP client.
	ExecuteFunction(ctx context.Context, w http.ResponseWriter, r *http.Request, author, name, version string, input json.RawMessage) (statusCode int, responseBody []byte)
}

// =============================================================================
// Handler
// =============================================================================

// Handler is the MCP entry point. Routes:
//
//	GET  /v1/mcp/manifest       — public, cacheable, server identity
//	GET  /v1/mcp/tools          — public, cacheable, tool index (SEO)
//	POST /v1/mcp                — JSON-RPC 2.0 transport (streamable-HTTP)
type Handler struct {
	Store         FunctionStore
	Executor      Executor
	ToolRegistry  *agenttools.Registry
	ServerPublicURL string
	Disabled      bool
	Now           func() time.Time
}

// NewHandler creates a Handler with sensible defaults.
func NewHandler(store FunctionStore, executor Executor) *Handler {
	return &Handler{
		Store:    store,
		Executor: executor,
		Now:      time.Now,
	}
}

// ServerManifest is the response shape for GET /v1/mcp/manifest.
type ServerManifest struct {
	Name            string         `json:"name"`
	Title           string         `json:"title"`
	Version         string         `json:"version"`
	ProtocolVersion string         `json:"protocol_version"`
	Description     string         `json:"description"`
	Homepage        string         `json:"homepage"`
	Icons           []ManifestIcon `json:"icons,omitempty"`
	Capabilities    ManifestCaps   `json:"capabilities"`
	Transport       []string       `json:"transport"`
	Endpoints       ManifestEPs    `json:"endpoints"`
	Stats           ManifestStats  `json:"stats"`
}

type ManifestIcon struct {
	Src     string `json:"src"`
	MimeType string `json:"mimeType,omitempty"`
}

type ManifestCaps struct {
	Tools     *ToolsCap     `json:"tools,omitempty"`
	Resources *ResourcesCap `json:"resources,omitempty"`
	Prompts   bool          `json:"prompts"`
	Logging   *LoggingCap   `json:"logging,omitempty"`
}

type ToolsCap struct {
	ListChanged bool `json:"listChanged"`
}

type ResourcesCap struct {
	Subscribe bool `json:"subscribe"`
}

type LoggingCap struct {
	Level string `json:"level"`
}

type ManifestEPs struct {
	StreamableHTTP string `json:"streamable_http,omitempty"`
	StdioPackage   string `json:"stdio_package,omitempty"`
}

type ManifestStats struct {
	TotalFunctions        int    `json:"total_functions"`
	MCPEnabledFunctions   int    `json:"mcp_enabled_functions"`
	LastUpdated           string `json:"last_updated"`
}

// HandleManifest serves GET /v1/mcp/manifest. Public, cacheable.
func (h *Handler) HandleManifest(w http.ResponseWriter, r *http.Request) {
	if h.Disabled {
		http.Error(w, "MCP registry is temporarily unavailable", http.StatusServiceUnavailable)
		return
	}

	enabled, total, _ := h.Store.ListEnabledMCPSettings(r.Context(), "", "", 0, 1, 0)

	base := h.publicURL()
	m := ServerManifest{
		Name:            ServerName,
		Title:           ServerName,
		Version:         ServerVersion,
		ProtocolVersion: ProtocolVersion,
		Description:     "Searchable, trust-scored directory of MCP-compatible functions. The npm for AI agent functions.",
		Homepage:        base + "/registry",
		Icons: []ManifestIcon{
			{Src: base + "/og-registry.png", MimeType: "image/png"},
		},
		Capabilities: ManifestCaps{
			Tools:     &ToolsCap{ListChanged: true},
			Resources: &ResourcesCap{Subscribe: false},
			Prompts:   false,
			Logging:   &LoggingCap{Level: "info"},
		},
		Transport: []string{"streamable-http", "stdio"},
		Endpoints: ManifestEPs{
			StreamableHTTP: base + "/v1/mcp",
			StdioPackage:   "@functionfly/mcp-server",
		},
		Stats: ManifestStats{
			TotalFunctions:      total,
			MCPEnabledFunctions: len(enabled),
			LastUpdated:         h.Now().UTC().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", ManifestCacheMaxAge))
	setCORSHeaders(w, r)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(m)
}

// HandleToolsIndex serves GET /v1/mcp/tools. Public, cacheable, SEO-friendly.
// Returns a JSON list of tool definitions (NOT a JSON-RPC envelope) so that
// crawlers and the /registry page can index it without parsing JSON-RPC.
func (h *Handler) HandleToolsIndex(w http.ResponseWriter, r *http.Request) {
	if h.Disabled {
		http.Error(w, "MCP registry is temporarily unavailable", http.StatusServiceUnavailable)
		return
	}
	if r.Method == http.MethodOptions {
		setCORSHeaders(w, r)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET, OPTIONS")
		writeHTTPError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only GET is allowed")
		return
	}

	filter := parseToolsListFilter(r)
	rows, total, err := h.Store.ListEnabledMCPSettings(
		r.Context(), filter.Category, filter.Runtime, filter.MinTrust, filter.Limit, filter.Offset,
	)
	if err != nil {
		logrus.WithError(err).Error("mcp: list enabled settings failed")
		writeHTTPError(w, r, http.StatusInternalServerError, "LIST_FAILED", "failed to list tools")
		return
	}

	tools := make([]ToolDefinition, 0, len(rows))
	for _, row := range rows {
		td, ok := h.buildToolDefinitionFromSettings(r.Context(), row)
		if !ok {
			continue
		}
		tools = append(tools, td)
	}

	result := ToolsListResult{
		Tools:       tools,
		Total:       total,
		GeneratedAt: h.Now().Unix(),
	}
	if len(rows) == filter.Limit && len(rows) > 0 {
		// Heuristic: if we filled the page, suggest a next cursor.
		result.NextCursor = EncodeCursor(rows[len(rows)-1].FunctionID.String())
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", ToolsIndexCacheMaxAge))
	setCORSHeaders(w, r)
	_ = json.NewEncoder(w).Encode(result)
}

// ToolDefinition is the wire shape we return for a single MCP tool. It is
// the union of (a) the MCP `tools/list` response item and (b) the items we
// return in our /v1/mcp/tools index, so the public index IS a strict
// superset of what `tools/list` would return (minus pagination).
type ToolDefinition struct {
	Name        string          `json:"name"`
	Title       string          `json:"title,omitempty"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema,omitempty"`
	Annotations ToolAnnotations `json:"annotations,omitempty"`
	Meta        ToolMeta        `json:"_meta,omitempty"`
}

type ToolAnnotations struct {
	ReadOnlyHint  bool   `json:"readOnlyHint,omitempty"`
	Destructive   bool   `json:"destructiveHint,omitempty"`
	OpenWorldHint bool   `json:"openWorldHint,omitempty"`
	Category      string `json:"category,omitempty"`
}

type ToolMeta struct {
	FunctionFly FunctionFlyMeta `json:"functionfly,omitempty"`
}

type FunctionFlyMeta struct {
	Author      string  `json:"author"`
	Name        string  `json:"name"`
	Version     string  `json:"version,omitempty"`
	TrustScore  float64 `json:"trust_score,omitempty"`
	TrustTier   string  `json:"trust_tier,omitempty"`
	VerifiedMCP bool    `json:"verified_mcp,omitempty"`
	Homepage    string  `json:"homepage,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Runtime     string  `json:"runtime,omitempty"`
}

// buildToolDefinitionFromSettings is the single source of truth for turning
// a DB row into a wire-format tool definition. Both `tools/list` and the
// public /v1/mcp/tools index call this.
func (h *Handler) buildToolDefinitionFromSettings(ctx context.Context, row registry.MCPSettings) (ToolDefinition, bool) {
	fnRow, err := h.Store.GetFunctionByID(ctx, row.FunctionID)
	if err != nil || fnRow == nil {
		return ToolDefinition{}, false
	}
	v, err := h.Store.GetLatestFunctionVersion(ctx, fnRow.ID)
	if err != nil || v == nil {
		return ToolDefinition{}, false
	}

	// settings from row + function fields merged into a single view.
	settings := ToMCPSettings{
		FunctionID:         row.FunctionID.String(),
		Author:             fnRow.Author,
		Name:               fnRow.Name,
		Version:            fnRow.LatestVersion.String,
		Enabled:            row.Enabled,
		Transports:         []string(row.Transports),
		ExposeInputSchema:  row.ExposeInputSchema,
		ExposeOutputSchema: row.ExposeOutputSchema,
		ToolNameOverride:   row.ToolNameOverride.String,
		RateLimitPerMin:    row.RateLimitPerMin,
		AllowlistOrigins:   []string(row.AllowlistOrigins),
		VerifiedMCP:        row.VerifiedMCP,
		TrustScore:         fnRow.TrustScore,
		TrustTier:          string(fnRow.TrustTier),
		Title:              fnRow.Title.String,
		Description:        fnRow.Description.String,
		Category:           fnRow.Category.String,
		Tags:               decodeTags(fnRow.Tags),
		Runtime:            v.Runtime,
		Manifest:           v.Manifest,
	}

	td := h.buildToolDefinitionWithHomepage(settings)
	return td, true
}

// buildToolDefinition is the pure (no I/O) core of tool-definition building.
// It is exported as BuildToolDefinition for use in tests and the npm proxy.
func BuildToolDefinition(s ToMCPSettings) ToolDefinition {
	toolName := SanitizeToolName(s.ToolNameOverride)
	if toolName == "" || toolName == "fn" {
		toolName = SanitizeToolName(s.Author + "__" + s.Name)
	}

	title := s.Title
	if title == "" {
		title = s.Name
	}
	desc := s.Description
	if desc == "" {
		desc = fmt.Sprintf("Function %s/%s", s.Author, s.Name)
	}

	// Build the MCP input schema from the manifest. If the publisher
	// disabled input schema exposure we return an empty object.
	inputSchema := json.RawMessage(`{"type":"object","properties":{}}`)
	if s.ExposeInputSchema && len(s.Manifest) > 0 {
		if got := extractInputSchema(s.Manifest); got != nil {
			inputSchema = got
		}
	}

	tags := s.Tags
	if tags == nil {
		tags = []string{}
	}

	base := ""
	// The home page link is best-effort: if ServerPublicURL is empty the
	// client can still reach the function via the registry search endpoint.
	if s.Author != "" && s.Name != "" {
		// We can't know h.ServerPublicURL here (it's a method receiver),
		// so callers should set Meta.Homepage after calling this builder.
	}

	return ToolDefinition{
		Name:        toolName,
		Title:       title,
		Description: desc,
		InputSchema: inputSchema,
		Annotations: ToolAnnotations{
			ReadOnlyHint:  s.VerifiedMCP, // curated tools are read-only by default
			OpenWorldHint: true,
			Category:      s.Category,
		},
		Meta: ToolMeta{
			FunctionFly: FunctionFlyMeta{
				Author:      s.Author,
				Name:        s.Name,
				Version:     s.Version,
				TrustScore:  s.TrustScore,
				TrustTier:   s.TrustTier,
				VerifiedMCP: s.VerifiedMCP,
				Homepage:    base,
				Tags:        tags,
				Runtime:     s.Runtime,
			},
		},
	}
}

// buildToolDefinitionWithHomepage calls BuildToolDefinition and patches the
// homepage with the handler's ServerPublicURL.
func (h *Handler) buildToolDefinitionWithHomepage(s ToMCPSettings) ToolDefinition {
	td := BuildToolDefinition(s)
	if h.ServerPublicURL != "" && s.Author != "" && s.Name != "" {
		td.Meta.FunctionFly.Homepage = fmt.Sprintf("%s/@%s/v1/fx/%s", h.ServerPublicURL, s.Author, s.Name)
	}
	return td
}

// extractInputSchema pulls the JSON-Schema "input" block out of a stored
// FunctionManifest. It is shared with wellknown/handler.go's buildOpenAIToolSchema
// to keep the on-the-wire shape consistent.
func extractInputSchema(manifestRaw []byte) json.RawMessage {
	if len(manifestRaw) == 0 {
		return nil
	}
	var m functionregistry.FunctionManifest
	if err := json.Unmarshal(manifestRaw, &m); err != nil || m.Input == nil {
		return nil
	}
	if m.Input.Schema != nil {
		return m.Input.Schema
	}
	// Fall back to the legacy {type, properties, required} shape.
	obj := map[string]interface{}{"type": "object"}
	if m.Input.Type != "" {
		obj["type"] = m.Input.Type
	}
	if m.Input.Properties != nil {
		var p map[string]interface{}
		if err := json.Unmarshal(m.Input.Properties, &p); err == nil {
			obj["properties"] = p
		} else {
			obj["properties"] = map[string]interface{}{}
		}
	} else {
		obj["properties"] = map[string]interface{}{}
	}
	if len(m.Input.Required.Array) > 0 {
		obj["required"] = m.Input.Required.Array
	}
	out, _ := json.Marshal(obj)
	return out
}

// =============================================================================
// JSON-RPC transport
// =============================================================================

// HandleJSONRPC serves POST /v1/mcp. Implements the JSON-RPC 2.0 transport
// (streamable-HTTP variant per MCP spec 2025-03-26).
//
// The Accept header must include both `application/json` and
// `text/event-stream` (per spec) for sessionful behavior. In v1 we always
// respond with `application/json` (the spec allows this for non-streaming
// responses) and rely on session-stickiness at the load balancer for any
// future per-session state.
func (h *Handler) HandleJSONRPC(w http.ResponseWriter, r *http.Request) {
	if h.Disabled {
		writeJSONRPCError(w, nil, ErrCodeInternalError, MsgInternalError)
		return
	}
	if r.Method == http.MethodOptions {
		setCORSHeaders(w, r)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST, OPTIONS")
		writeJSONRPCError(w, nil, ErrCodeInvalidRequest, MsgInvalidRequest)
		return
	}

	// Per-stream CORS / origin check.
	if !h.checkOriginAllowed(r) {
		writeJSONRPCError(w, nil, ErrCodeInternalError, "Origin not allowed")
		return
	}

	// Cap body size to defend against denial-of-service.
	r.Body = http.MaxBytesReader(w, r.Body, int64(ToolsCallMaxArgsBytes*4+8*1024))
	defer r.Body.Close()

	// Parse the JSON-RPC frame.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		writeJSONRPCError(w, nil, ErrCodeParseError, MsgParseError)
		return
	}
	reqs, errResp := ParseFrame(raw)
	if errResp != nil {
		writeJSONRPCResponse(w, errResp)
		return
	}

	// Fan out: each request gets its own response. We collect them into a
	// batch array only if the input was a batch.
	responses := make([]*JSONRPCResponse, 0, len(reqs))
	for _, req := range reqs {
		responses = append(responses, h.dispatch(r, req))
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Mcp-Session-Id", r.Header.Get("Mcp-Session-Id"))
	setCORSHeaders(w, r)
	if len(responses) == 1 {
		_ = json.NewEncoder(w).Encode(responses[0])
		return
	}
	_ = json.NewEncoder(w).Encode(responses)
}

// dispatch routes one request to the right method handler.
func (h *Handler) dispatch(r *http.Request, req *JSONRPCRequest) *JSONRPCResponse {
	start := h.Now()
	ctx := r.Context()
	method := req.Method

	switch method {
	case "initialize":
		return h.handleInitialize(req)
	case "notifications/initialized":
		// Client signals it has completed init. No response needed.
		return nil
	case "ping":
		return MakeResult(req.ID, map[string]interface{}{})
	case "tools/list":
		return h.handleToolsList(ctx, req)
	case "tools/call":
		return h.handleToolsCall(ctx, r, req, start)
	case "resources/list":
		// We do not expose MCP resources in v1. Return empty list.
		return MakeResult(req.ID, map[string]interface{}{"resources": []interface{}{}})
	case "prompts/list":
		return MakeResult(req.ID, map[string]interface{}{"prompts": []interface{}{}})
	case "logging/setLevel":
		return MakeResult(req.ID, map[string]interface{}{})
	default:
		return MakeError(req.ID, ErrCodeMethodNotFound, MsgMethodNotFound, method)
	}
}

// handleInitialize responds with server info + capabilities.
func (h *Handler) handleInitialize(req *JSONRPCRequest) *JSONRPCResponse {
	result := map[string]interface{}{
		"protocolVersion": ProtocolVersion,
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{"listChanged": true},
		},
		"serverInfo": map[string]interface{}{
			"name":    ServerName,
			"version": ServerVersion,
		},
	}
	if h.ServerPublicURL != "" {
		result["serverInfo"].(map[string]interface{})["homepage"] = h.ServerPublicURL + "/registry"
	}
	return MakeResult(req.ID, result)
}

// handleToolsList returns a paginated list of tool definitions.
func (h *Handler) handleToolsList(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	filter := ToolsListFilter{Limit: ToolsListDefaultLimit}

	if len(req.Params) > 0 {
		var p struct {
			Category string  `json:"category"`
			Runtime  string  `json:"runtime"`
			MinTrust float64 `json:"minTrust"`
			Search   string  `json:"search"`
			Cursor   string  `json:"cursor"`
			Limit    int     `json:"limit"`
		}
		_ = json.Unmarshal(req.Params, &p)
		filter.Category = trim(p.Category, 128)
		filter.Runtime = trim(p.Runtime, 64)
		filter.MinTrust = p.MinTrust
		filter.Search = trim(p.Search, 256)
		filter.Cursor = p.Cursor
		filter.Limit = p.Limit
	}
	filter = clampToolsListFilter(filter)

	rows, _, err := h.Store.ListEnabledMCPSettings(ctx, filter.Category, filter.Runtime, filter.MinTrust, filter.Limit, 0)
	if err != nil {
		return MakeError(req.ID, ErrCodeInternalError, MsgInternalError, "list failed")
	}

	// Apply cursor.
	cutoff := DecodeCursor(filter.Cursor)
	tools := make([]ToolDefinition, 0, len(rows))
	lastID := ""
	for _, row := range rows {
		if cutoff != "" && row.FunctionID.String() <= cutoff {
			continue
		}
		td, ok := h.buildToolDefinitionFromSettings(ctx, row)
		if !ok {
			continue
		}
		tools = append(tools, td)
		lastID = row.FunctionID.String()
		if len(tools) >= filter.Limit {
			break
		}
	}

	resp := map[string]interface{}{"tools": tools}
	if lastID != "" && len(rows) >= filter.Limit {
		resp["nextCursor"] = EncodeCursor(lastID)
	}
	return MakeResult(req.ID, resp)
}

// =============================================================================
// helpers
// =============================================================================

// parseToolsListFilter parses the public /v1/mcp/tools query string.
func parseToolsListFilter(r *http.Request) ToolsListFilter {
	q := r.URL.Query()
	f := ToolsListFilter{
		Category: trim(q.Get("category"), 128),
		Runtime:  trim(q.Get("runtime"), 64),
		Search:   trim(q.Get("q"), 256),
	}
	if v, err := strconv.ParseFloat(q.Get("min_trust"), 64); err == nil {
		f.MinTrust = v
	}
	if v, err := strconv.Atoi(q.Get("limit")); err == nil && v > 0 {
		f.Limit = v
	}
	if v, err := strconv.Atoi(q.Get("offset")); err == nil && v >= 0 {
		f.Offset = v
	}
	return clampToolsListFilter(f)
}

func clampToolsListFilter(f ToolsListFilter) ToolsListFilter {
	if f.Limit <= 0 {
		f.Limit = ToolsListDefaultLimit
	}
	if f.Limit > ToolsListMaxLimit {
		f.Limit = ToolsListMaxLimit
	}
	return f
}

func (h *Handler) publicURL() string {
	if h.ServerPublicURL != "" {
		return strings.TrimRight(h.ServerPublicURL, "/")
	}
	if v := os.Getenv("PUBLIC_SITE_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "https://functionfly.com"
}

func (h *Handler) checkOriginAllowed(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// No Origin header (typical for stdio + same-origin); allow.
		return true
	}
	// Public CORS: any origin is allowed for the streamable-HTTP transport
	// (the bearer-token auth gates execution). Per-function allowlist
	// origins are enforced in the per-function rate limiter.
	return true
}

// getFunctionByID is a thin adapter that calls the store's GetFunctionByID
// (declared in the FunctionStore interface). Kept as a method for symmetry
// with the other helpers.
func (h *Handler) getFunctionByID(ctx context.Context, id uuid.UUID) (*registry.RegistryFunction, error) {
	return h.Store.GetFunctionByID(ctx, id)
}

func (h *Handler) getCallerID(r *http.Request) string {
	if c := CallerFromContext(r.Context()); c != nil {
		return c.TokenHash
	}
	return "anonymous"
}

func (h *Handler) recordInvocation(r *http.Request, functionID uuid.UUID, method string, status int, errorCode string, dur time.Duration, reqBytes, respBytes int) {
	caller := CallerFromContext(r.Context())
	callerID := ""
	callerOrigin := ""
	if caller != nil {
		callerID = caller.TokenHash
	}
	callerOrigin = r.Header.Get("Origin")

	rec := registry.MCPInvocationRecord{
		FunctionID:    functionID,
		CallerID:      callerID,
		CallerOrigin:  callerOrigin,
		Transport:     TransportStreamableHTTP,
		Method:        method,
		DurationMs:    int(dur.Milliseconds()),
		StatusCode:    status,
		ErrorCode:     errorCode,
		RequestBytes:  reqBytes,
		ResponseBytes: respBytes,
		Timestamp:     h.Now().UTC(),
	}
	if err := h.Store.RecordMCPInvocation(r.Context(), rec); err != nil {
		logrus.WithError(err).Debug("mcp: failed to record invocation")
	}
	// Best-effort counter bump; ignore failure.
	_ = h.Store.IncrementMCPInvocationCount(r.Context(), functionID)
}

// =============================================================================
// public http helpers (shared with jsonrpc.go)
// =============================================================================

// setCORSHeaders writes the MCP transport CORS headers. It is a thin
// wrapper over gateway.SetCORSHeaders — the actual header set is the
// single source of truth in the gateway package (see P0 of the
// Two-Protocol Gateway plan).
func setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	gateway.SetCORSHeaders(w, r, gateway.CORSOptions{
		AllowMethods:  "GET, POST, OPTIONS",
		AllowHeaders:  "Content-Type, Authorization, Mcp-Session-Id, Mcp-Protocol-Version",
		ExposeHeaders: "Mcp-Session-Id, X-Request-Id",
	})
}

func writeJSONRPCError(w http.ResponseWriter, id json.RawMessage, code int, msg string) {
	writeJSONRPCResponse(w, MakeError(id, code, msg, nil))
}

func writeJSONRPCResponse(w http.ResponseWriter, resp *JSONRPCResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	setCORSHeaders(w, nil)
	_ = json.NewEncoder(w).Encode(resp)
}

func writeHTTPError(w http.ResponseWriter, r *http.Request, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	setCORSHeaders(w, r)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":    false,
		"error": map[string]string{"code": code, "message": msg},
	})
}

func writeAuthRequiredHTTP(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("WWW-Authenticate", `Bearer realm="mcp", error="invalid_token"`)
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"ok":false,"error":{"code":"UNAUTHORIZED","message":"Authentication required"}}`))
}

// nullStr extracts a string from a sql.NullString. Defined as a function
// (not a method) so callers can pass it any nullable string-shaped value.
func nullStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// decodeTags unmarshals the JSONB tags column into a []string.
func decodeTags(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func trim(s string, max int) string {
	s = strings.TrimSpace(s)
	if max > 0 && len(s) > max {
		return s[:max]
	}
	return s
}

// _ = mux.Vars is referenced to avoid an unused-import error if the handler
// later needs to read path vars.
var _ = mux.Vars
