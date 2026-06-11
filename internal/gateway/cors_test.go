package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetCORSHeaders_Defaults(t *testing.T) {
	rec := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	SetCORSHeaders(rec, r, CORSOptions{})

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Allow-Origin = %q, want %q", got, "*")
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, OPTIONS" {
		t.Errorf("Allow-Methods = %q, want %q", got, "GET, OPTIONS")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, Authorization" {
		t.Errorf("Allow-Headers = %q, want default", got)
	}
	if got := rec.Header().Get("Access-Control-Max-Age"); got != "86400" {
		t.Errorf("Max-Age = %q, want %q", got, "86400")
	}
	// ExposeHeaders is opt-in; default should be absent.
	if got := rec.Header().Get("Access-Control-Expose-Headers"); got != "" {
		t.Errorf("Expose-Headers = %q, want empty (opt-in)", got)
	}
}

func TestSetCORSHeaders_MCP(t *testing.T) {
	rec := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/v1/mcp", nil)
	SetCORSHeaders(rec, r, CORSOptions{
		AllowMethods:  "GET, POST, OPTIONS",
		AllowHeaders:  "Content-Type, Authorization, Mcp-Session-Id, Mcp-Protocol-Version",
		ExposeHeaders: "Mcp-Session-Id, X-Request-Id",
	})

	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, OPTIONS" {
		t.Errorf("Allow-Methods = %q, want MCP shape", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, Authorization, Mcp-Session-Id, Mcp-Protocol-Version" {
		t.Errorf("Allow-Headers = %q, want MCP shape", got)
	}
	if got := rec.Header().Get("Access-Control-Expose-Headers"); got != "Mcp-Session-Id, X-Request-Id" {
		t.Errorf("Expose-Headers = %q, want MCP shape", got)
	}
}

func TestSetCORSHeaders_WellKnown(t *testing.T) {
	rec := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/.well-known/functionfly.json", nil)
	SetCORSHeaders(rec, r, CORSOptions{AllowMethods: "GET, OPTIONS"})

	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, OPTIONS" {
		t.Errorf("Allow-Methods = %q, want %q", got, "GET, OPTIONS")
	}
}

func TestSetCORSHeaders_NilRequestIsOK(t *testing.T) {
	// Some MCP write paths (e.g. writeJSONRPCResponse in error paths) pass
	// nil for the request. The implementation must not panic.
	rec := httptest.NewRecorder()
	SetCORSHeaders(rec, nil, CORSOptions{AllowMethods: "GET, POST, OPTIONS"})
	if rec.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("expected headers to be set even with nil request")
	}
}

// Ensure the test file does not lose the http import if a future test is
// removed (defensive).
var _ = http.MethodGet
