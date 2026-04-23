package middleware

import (
	"context"
	"net/http"
)

// environmentContextKey is the typed key for storing environment in context
type environmentContextKey int

const (
	contextKeyEnvironment environmentContextKey = iota
)

// ValidEnvironmentValues represents the allowed environment values
var ValidEnvironmentValues = map[string]bool{
	"production":  true,
	"staging":     true,
	"development": true,
}

// EnvironmentMiddleware extracts the X-Environment header from requests and adds it to context.
// This allows API handlers to filter data based on the user's selected environment.
// The header is optional - if not present, "production" is used as the default.
func EnvironmentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract environment from header
		env := r.Header.Get("X-Environment")

		// Validate and normalize
		if env == "" || !ValidEnvironmentValues[env] {
			env = "production" // Default to production
		}

		// Add to context
		ctx := context.WithValue(r.Context(), contextKeyEnvironment, env)

		// Also set response header to indicate which environment was used
		w.Header().Set("X-Active-Environment", env)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetEnvironmentFromContext extracts the environment value from request context.
// Returns the environment string (production, staging, or development) or "production" if not set.
func GetEnvironmentFromContext(r *http.Request) string {
	return GetEnvironmentFromContextValue(r.Context())
}

// GetEnvironmentFromContextValue extracts environment from a context value.
func GetEnvironmentFromContextValue(ctx context.Context) string {
	if env, ok := ctx.Value(contextKeyEnvironment).(string); ok {
		return env
	}
	return "production" // Default
}
