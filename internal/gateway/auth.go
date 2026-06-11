package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// AuthResolver is the unified identity resolver used by GatewayCore.
// It accepts any credential type (JWT, API key, agent signing key,
// A2A peer JWT) and returns a Caller. Protocol adapters extract the
// raw credential from their wire format and pass it here.
type AuthResolver interface {
	// ResolveIdentity attempts to resolve a Caller from the raw bearer
	// token. Returns (Caller, nil) on success, (zero Caller, nil) for
	// anonymous, or (zero Caller, err) for a hard auth failure.
	ResolveIdentity(ctx context.Context, token string) (Caller, error)
}

// CredentialType classifies the raw credential the adapter extracted.
type CredentialType int

const (
	CredentialNone     CredentialType = iota
	CredentialJWT                     // session token
	CredentialAPIKey                   // ffp_/fff_/aep_/... prefixed
	CredentialAgentKey                 // aep_/ags_ prefixed (agent identity)
	CredentialPeerJWT                  // A2A peer JWT (verified via peer_jwks_url)
)

// ClassifyCredential returns the credential type based on the token
// prefix and shape. This is a heuristic — the actual validation is
// delegated to the AuthResolver.
func ClassifyCredential(token string) CredentialType {
	if token == "" {
		return CredentialNone
	}
	if looksLikeJWT(token) {
		return CredentialJWT
	}
	if strings.HasPrefix(token, "aep_") || strings.HasPrefix(token, "ags_") {
		return CredentialAgentKey
	}
	if strings.HasPrefix(token, "ff") {
		return CredentialAPIKey
	}
	// Default: treat unknown shapes as API keys (the resolver will reject
	// if invalid).
	return CredentialAPIKey
}

// looksLikeJWT returns true for tokens of the form xxx.yyy.zzz.
func looksLikeJWT(tok string) bool {
	parts := strings.Split(tok, ".")
	return len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != ""
}

// ShortHash returns a truncated SHA-256 hex string (16 chars) suitable
// for use as a caller identifier in observability rows. It avoids
// logging the raw token.
func ShortHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8])
}

// CallerFromContext is a helper that protocol adapters can use to
// extract a Caller from the context. Returns the zero value if not set.
func CallerFromContext(ctx context.Context) Caller {
	if v, ok := ctx.Value(contextKeyCallerCtx).(Caller); ok {
		return v
	}
	return Caller{}
}

// WithCaller returns a new context with the Caller attached.
func WithCaller(ctx context.Context, c Caller) context.Context {
	return context.WithValue(ctx, contextKeyCallerCtx, c)
}

type callerContextKey int

const contextKeyCallerCtx callerContextKey = iota
