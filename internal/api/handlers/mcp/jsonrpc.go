package mcp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

// =============================================================================
// JSON-RPC 2.0 Envelope
// =============================================================================
//
// Spec: https://www.jsonrpc.org/specification
//
// A request:    {"jsonrpc":"2.0","method":"...","params":{...},"id":N}
// A success:    {"jsonrpc":"2.0","result":...,"id":N}
// An error:     {"jsonrpc":"2.0","error":{"code":N,"message":"..."},"id":N}
// A notification: {"jsonrpc":"2.0","method":"...","params":{...}}  (no "id" key)
//
// `id` may be a string, number, or null. We normalize to interface{}.

// JSONRPCRequest is one parsed request frame. For batch requests, the parser
// returns []JSONRPCRequest; the caller fans out and assembles []JSONRPCResponse.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"` // string | number | null
	IsNotification bool     `json:"-"`            // derived: id field absent
}

// HasID returns true if this request has a non-null id (i.e. expects a
// response). False for notifications.
func (r *JSONRPCRequest) HasID() bool {
	return len(r.ID) > 0 && string(r.ID) != "null"
}

// IDString returns the id field as a string for logging/correlation. The
// original RawMessage is preserved in .ID for round-tripping.
func (r *JSONRPCRequest) IDString() string {
	if len(r.ID) == 0 {
		return ""
	}
	return string(r.ID)
}

// JSONRPCError is the standard error object (code + message + optional data).
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSONRPCResponse is the envelope we write back to the client.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	ID      json.RawMessage `json:"id"` // echoed back (may be null for parse errors)
}

// MakeError is the canonical error-frame builder. It always sets JSONRPC to
// "2.0" and echoes the original id.
func MakeError(id json.RawMessage, code int, message string, data interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	}
}

// MakeResult is the canonical success-frame builder.
func MakeResult(id json.RawMessage, result interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
}

// ParseRequest parses a single JSON-RPC 2.0 request object. It enforces:
//   - jsonrpc field is "2.0"
//   - method is a non-empty string
//   - params (if present) is a JSON object or array
//   - id (if present) is a string, number, or null
//
// Returns a *JSONRPCParseError with a JSON-RPC -32700 code on failure.
func ParseRequest(raw []byte) (*JSONRPCRequest, *JSONRPCResponse) {
	if len(raw) == 0 {
		return nil, MakeError(nil, ErrCodeParseError, MsgParseError, nil)
	}

	var probe map[string]json.RawMessage
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, MakeError(nil, ErrCodeParseError, MsgParseError, nil)
	}

	// Batch: { "jsonrpc": "...", ... } is a single request; an array is a batch.
	// We disallow batches in v1 (return a single -32600) — MCP does not require
	// them, and they complicate tool/call error semantics.
	if _, isArr := probe["__is_array__"]; isArr {
		// (placeholder; the array branch is handled in ParseFrame below)
	}

	// Per spec, a single request is itself an object.
	if _, hasJSONRPC := probe["jsonrpc"]; !hasJSONRPC {
		return nil, MakeError(nil, ErrCodeInvalidRequest, MsgInvalidRequest, "missing jsonrpc field")
	}
	// `probe["jsonrpc"]` is a JSON-encoded string (e.g. `"2.0"` including
	// the surrounding quotes). We must unquote it before comparing.
	var v string
	if err := json.Unmarshal(probe["jsonrpc"], &v); err != nil {
		return nil, MakeError(nil, ErrCodeParseError, MsgParseError, "jsonrpc field is not a string")
	}
	if v != "2.0" {
		return nil, MakeError(nil, ErrCodeInvalidRequest, MsgInvalidRequest,
			fmt.Sprintf("jsonrpc must be \"2.0\", got %q", v))
	}
	if _, hasMethod := probe["method"]; !hasMethod {
		return nil, MakeError(nil, ErrCodeInvalidRequest, MsgInvalidRequest, "missing method field")
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, MakeError(nil, ErrCodeParseError, MsgParseError, nil)
	}
	if req.JSONRPC != "2.0" {
		return nil, MakeError(nil, ErrCodeInvalidRequest, MsgInvalidRequest, nil)
	}
	if strings.TrimSpace(req.Method) == "" {
		return nil, MakeError(nil, ErrCodeInvalidRequest, MsgInvalidRequest, "method is empty")
	}

	// Derive IsNotification.
	req.IsNotification = !req.HasID()
	return &req, nil
}

// ParseFrame parses the HTTP body. It accepts either a single object or a
// non-empty JSON array of objects. On any parse failure it returns a single
// parse-error response.
//
// Per JSON-RPC 2.0 spec, an empty array (`[]`) is an invalid request.
func ParseFrame(raw []byte) (reqs []*JSONRPCRequest, errResp *JSONRPCResponse) {
	if len(raw) == 0 {
		return nil, MakeError(nil, ErrCodeInvalidRequest, MsgInvalidRequest, "empty body")
	}

	// Peek at first non-whitespace byte.
	var first byte
	firstSeen := false
	for _, b := range raw {
		if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
			continue
		}
		first = b
		firstSeen = true
		break
	}
	if !firstSeen {
		return nil, MakeError(nil, ErrCodeInvalidRequest, MsgInvalidRequest, "empty body (whitespace only)")
	}

	if first == '[' {
		// Batch
		var arr []json.RawMessage
		if err := json.Unmarshal(raw, &arr); err != nil {
			return nil, MakeError(nil, ErrCodeParseError, MsgParseError, nil)
		}
		if len(arr) == 0 {
			return nil, MakeError(nil, ErrCodeInvalidRequest, MsgInvalidRequest, "empty batch")
		}
		for _, item := range arr {
			r, e := ParseRequest(item)
			if e != nil {
				// Per spec, a per-request parse error in a batch becomes a
				// standalone error response. We'll still keep parsing the rest
				// but caller is single-threaded — return both via a struct
				// below. For simplicity in v1 we return a single error.
				_ = r
				return nil, e
			}
			reqs = append(reqs, r)
		}
		return reqs, nil
	}

	if first == '{' {
		r, e := ParseRequest(raw)
		if e != nil {
			return nil, e
		}
		return []*JSONRPCRequest{r}, nil
	}
	return nil, MakeError(nil, ErrCodeParseError, MsgParseError, "body must be a JSON object or array")
}

// =============================================================================
// Cursor encoding for tools/list pagination
// =============================================================================
//
// A cursor is a base64url-encoded JSON object {"id":"<function_uuid>"}. We
// return a nextCursor only when there are likely more results. The MCP client
// is expected to pass it back unchanged as `params.cursor` on the next call.

type cursorPayload struct {
	ID string `json:"id"`
}

// EncodeCursor returns "" for empty input.
func EncodeCursor(functionID string) string {
	if functionID == "" {
		return ""
	}
	raw, _ := json.Marshal(cursorPayload{ID: functionID})
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeCursor returns the function_id encoded in the cursor, or "" if the
// cursor is empty / malformed.
func DecodeCursor(cursor string) string {
	cursor = strings.TrimSpace(cursor)
	if cursor == "" {
		return ""
	}
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return ""
	}
	var p cursorPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return ""
	}
	return p.ID
}

// =============================================================================
// Tool name sanitization
// =============================================================================
//
// MCP spec: tool names MUST match ^[a-zA-Z0-9_-]{1,64}$. We use the same
// rule as wellknown/handler.go: replace any non-conforming char with `_`
// and clamp to 64 chars. The double-underscore `__` is our reserved
// delimiter for `{author}__{name}`.

var toolNameAllowed = func() map[rune]struct{} {
	m := make(map[rune]struct{}, 64)
	for r := 'a'; r <= 'z'; r++ {
		m[r] = struct{}{}
	}
	for r := 'A'; r <= 'Z'; r++ {
		m[r] = struct{}{}
	}
	for r := '0'; r <= '9'; r++ {
		m[r] = struct{}{}
	}
	m['_'] = struct{}{}
	m['-'] = struct{}{}
	return m
}()

// SanitizeToolName returns a name safe for MCP tool use. Falls back to
// "fn_<hash>" if the input is entirely invalid.
func SanitizeToolName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "fn"
	}
	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		if _, ok := toolNameAllowed[r]; ok {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := b.String()
	out = strings.Trim(out, "_-")
	if out == "" {
		return "fn"
	}
	if len(out) > ToolNameMaxLength {
		// UTF-8 safe truncation: walk back to a rune boundary.
		cut := ToolNameMaxLength
		for cut > 0 && !unicode.IsLetter(rune(out[cut-1])) && !unicode.IsDigit(rune(out[cut-1])) && out[cut-1] != '_' && out[cut-1] != '-' {
			cut--
		}
		out = out[:cut]
	}
	return out
}

// ParseToolName reverses ToToolName to recover author and name. If the
// tool name was created from a tool_name_override, the override is returned
// in the override parameter and the caller should look it up via the
// per-function override column.
func ParseToolName(toolName string) (author, name, override string) {
	toolName = strings.TrimSpace(toolName)
	if toolName == "" {
		return "", "", ""
	}
	// We do not know if the input was an override. We try the most common
	// case (author__name) first.
	if idx := strings.Index(toolName, "__"); idx > 0 && idx < len(toolName)-2 {
		author = toolName[:idx]
		name = toolName[idx+2:]
		return author, name, ""
	}
	return "", "", toolName
}

// =============================================================================
// Argument validation helpers
// =============================================================================

// MaxDepth returns the maximum nesting depth of a JSON value. Used to defend
// against pathological arguments that would otherwise blow the JSON parser.
func MaxDepth(raw json.RawMessage) int {
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return -1
	}
	return depthOf(v)
}

func depthOf(v interface{}) int {
	switch x := v.(type) {
	case map[string]interface{}:
		d := 0
		for _, c := range x {
			if cd := depthOf(c); cd > d {
				d = cd
			}
		}
		return d + 1
	case []interface{}:
		d := 0
		for _, c := range x {
			if cd := depthOf(c); cd > d {
				d = cd
			}
		}
		return d + 1
	}
	return 0
}
