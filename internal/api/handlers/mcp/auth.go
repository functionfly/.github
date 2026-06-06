package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/sirupsen/logrus"
)

// contextKey is unexported to prevent collisions with other packages.
type contextKey int

const (
	contextKeyCaller contextKey = iota
)

// CallerFromContext returns the authenticated CallerIdentity, or nil if the
// request was anonymous. The MCP handlers treat nil as "must authenticate"
// (for tools/call) or "skip authentication" (for tools/list, manifest, etc.)
// depending on the route.
func CallerFromContext(ctx context.Context) *CallerIdentity {
	if v, ok := ctx.Value(contextKeyCaller).(*CallerIdentity); ok {
		return v
	}
	return nil
}

// bearerAuthenticator is the minimal API-key validator surface the middleware
// needs. The concrete implementation is provided by the orchestrator at
// wiring time (see registerMCPRoutes in routes.go).
type bearerAuthenticator interface {
	ValidateAPIKey(rawKey string) (*apikey.APIKey, error)
}

// jwtAuthenticator is the minimal JWT validator surface.
type jwtAuthenticator interface {
	ValidateToken(token string) (*auth.Claims, error)
}

// AuthConfig wires the MCP auth middleware.
type AuthConfig struct {
	APIKeyRepo  bearerAuthenticator
	AuthService jwtAuthenticator
	// Now is overridable for tests.
	Now func() int64
}

// NewBearerAuthMiddleware returns an http.Handler middleware that extracts
// and validates the `Authorization: Bearer <token>` header.
//
// The middleware tries, in order:
//   1. JWT session token (validated by AuthService).
//   2. FunctionFly API key (ffp_ / fff_ / ffo_ / aep_ / ffe_ / fft_).
//
// On success, a *CallerIdentity is attached to the request context. On
// failure, NO 401 is returned — the request is allowed through anonymously
// so the public read endpoints (manifest, tools index) keep working. The
// per-handler auth check (RequireAuthForToolsCall) is responsible for
// rejecting tools/call without a valid caller.
func NewBearerAuthMiddleware(cfg AuthConfig) func(http.Handler) http.Handler {
	if cfg.Now == nil {
		cfg.Now = func() int64 { return 0 }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// OPTIONS preflight: skip auth
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			raw := bearerToken(r.Header.Get("Authorization"))
			if raw == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Skip the heuristic entirely for non-FF-shaped tokens (e.g. some
			// JWT issuers use non-FF prefixes). We try JWT first because
			// JWTs are the most common session token format.
			if looksLikeJWT(raw) {
				if cfg.AuthService != nil {
					if claims, err := cfg.AuthService.ValidateToken(raw); err == nil && claims != nil {
						caller := &CallerIdentity{
							UserID:    claims.UserID.String(),
							TenantID:  claims.TenantID.String(),
							AuthType:  "session",
							TokenHash: shortHash(raw),
						}
						ctx := context.WithValue(r.Context(), contextKeyCaller, caller)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			// Try API key. We try the platform repo first, then fall back
			// to the Trust API key validator.
			if cfg.APIKeyRepo != nil {
				if apiKey, err := cfg.APIKeyRepo.ValidateAPIKey(raw); err == nil && apiKey != nil {
					caller := &CallerIdentity{
						UserID:    apiKey.UserID.String(),
						TenantID:  apiKey.TenantID.String(),
						APIKeyID:  apiKey.KeyID,
						AuthType:  "apikey",
						TokenHash: shortHash(raw),
					}
					ctx := context.WithValue(r.Context(), contextKeyCaller, caller)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// Token present but invalid: do not silently log in. Let the
			// downstream handler decide whether to reject. We do emit a
			// single structured log for ops.
			logrus.WithFields(logrus.Fields{
				"path":   r.URL.Path,
				"method": r.Method,
				"prefix": truncate(raw, 8),
			}).Debug("mcp: bearer token present but failed validation")
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuthForToolsCall returns 401 if no caller is attached to the request
// context. Use as a wrapper around the JSON-RPC handler. Per MCP spec, only
// tools/call is auth-gated; the public listing endpoints stay open.
func RequireAuthForToolsCall(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if CallerFromContext(r.Context()) == nil {
			writeAuthRequired(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// =============================================================================
// helpers
// =============================================================================

func bearerToken(h string) string {
	h = strings.TrimSpace(h)
	if h == "" {
		return ""
	}
	const p = "Bearer "
	if len(h) <= len(p) || !strings.EqualFold(h[:len(p)], p) {
		return ""
	}
	return strings.TrimSpace(h[len(p):])
}

// looksLikeJWT returns true for tokens of the form `xxx.yyy.zzz`. We do not
// enforce a specific issuer — we just rely on the validator to reject
// foreign tokens.
func looksLikeJWT(tok string) bool {
	if tok == "" {
		return false
	}
	parts := strings.Split(tok, ".")
	return len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != ""
}

func shortHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8])
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// writeAuthRequired emits a JSON-RPC 2.0 -32003 error envelope. We use the
// HTTP status 401 (per the spec this is acceptable for transport errors).
func writeAuthRequired(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("WWW-Authenticate", `Bearer realm="mcp", error="invalid_token"`)
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32003,"message":"Authentication required"},"id":null}`))
}

// isAnonymous returns true if no caller is attached. Equivalent to
// CallerFromContext(c) == nil but more idiomatic at call sites.
func isAnonymous(ctx context.Context) bool {
	return CallerFromContext(ctx) == nil
}

// errAuthMissing is exported so tests can assert on it.
var errAuthMissing = errors.New("mcp: caller identity missing from context")
