// Package gateway provides the shared "Two-Protocol Gateway" core used
// by both the MCP and (future) A2A protocol adapters. See
// plans/TWO_PROTOCOL_GATEWAY_PLAN.md for the architecture.
//
// The DRY principle: protocol adapters are thin shells that translate
// their wire shape (JSON-RPC for MCP, REST/JSON for A2A) into a
// CallRequest and back. The actual execution path, auth resolution,
// capability check, rate-limit brokering, fallback chain, and receipt
// emission all live here, behind a single GatewayCore.Call.
package gateway

import "net/http"

// CORSOptions customises SetCORSHeaders. The zero value is valid and
// produces sensible defaults for a public read endpoint.
type CORSOptions struct {
	// AllowOrigin is the value of Access-Control-Allow-Origin. Default "*".
	AllowOrigin string
	// AllowMethods is the value of Access-Control-Allow-Methods. Default
	// "GET, OPTIONS" — appropriate for a public read endpoint.
	AllowMethods string
	// AllowHeaders is the value of Access-Control-Allow-Headers. Default
	// "Content-Type, Authorization".
	AllowHeaders string
	// ExposeHeaders is the value of Access-Control-Expose-Headers. Default "".
	ExposeHeaders string
	// MaxAge is the value of Access-Control-Max-Age. Default "86400" (1 day).
	MaxAge string
}

// SetCORSHeaders writes the configured CORS headers. Idempotent.
//
// The request argument is ignored — CORS headers are identical for
// every request on the same route. The parameter is kept so callers
// can pass the live *http.Request without a stub.
func SetCORSHeaders(w http.ResponseWriter, _ *http.Request, opts CORSOptions) {
	if opts.AllowOrigin == "" {
		opts.AllowOrigin = "*"
	}
	if opts.AllowMethods == "" {
		opts.AllowMethods = "GET, OPTIONS"
	}
	if opts.AllowHeaders == "" {
		opts.AllowHeaders = "Content-Type, Authorization"
	}
	if opts.MaxAge == "" {
		opts.MaxAge = "86400"
	}
	h := w.Header()
	h.Set("Access-Control-Allow-Origin", opts.AllowOrigin)
	h.Set("Access-Control-Allow-Methods", opts.AllowMethods)
	h.Set("Access-Control-Allow-Headers", opts.AllowHeaders)
	if opts.ExposeHeaders != "" {
		h.Set("Access-Control-Expose-Headers", opts.ExposeHeaders)
	}
	h.Set("Access-Control-Max-Age", opts.MaxAge)
}
