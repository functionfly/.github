package api

import (
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/api/utils"
	"github.com/functionfly/functionfly/internal/tracing"
)

// reservedSubdomains are subdomains handled by dedicated services (not app slugs).
var reservedSubdomains = map[string]bool{
	"www": true, "api": true, "app": true, "admin": true,
	"auth": true, "docs": true, "cdn": true, "edge": true,
	"run": true, "registry": true, "status": true, "blog": true,
}

// knownPathPrefixes are URL path prefixes that belong to the API, auth, or
// internal services and should never be rewritten by the edge middleware.
var knownPathPrefixes = []string{
	"/v1/", "/v2/", "/auth/", "/api/", "/admin/", "/content/",
	"/health", "/healthz", "/health/", "/metrics",
	"/swagger/", "/swagger", "/.well-known/",
	"/ws/", "/webhook/", "/marketplace/",
	"/frg/", "/gx/",
}

// EdgeSlugMiddleware rewrites incoming requests that carry an X-FF-Slug header
// (set by the Cloudflare Worker or local Caddy edge proxy) into the internal
// /{appSlug}/... path format that the public route matcher in routes.go expects.
//
// Because the middleware runs before gorilla/mux route matching, the rewritten
// path is what the router sees — so mux.Vars(r)["appSlug"] works correctly
// in handlePublicRoute and ProxyToBackend.
func EdgeSlugMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := r.Header.Get("X-FF-Slug")
		if slug == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Start tracing span for edge middleware
		ctx, _ := tracing.StartSpan(r.Context(), "edge-middleware")
		defer tracing.Finish(ctx)
		tracing.SetAttribute(ctx, "app_slug", slug)
		tracing.SetAttribute(ctx, "original_path", r.URL.Path)
		r = r.WithContext(ctx)

		// Defense-in-depth: skip reserved subdomains (CF Worker already filters).
		if reservedSubdomains[strings.ToLower(slug)] {
			tracing.SetAttribute(ctx, "blocked", "reserved_subdomain")
			http.NotFound(w, r)
			return
		}

		// Don't rewrite paths that already target a known API/internal prefix.
		for _, prefix := range knownPathPrefixes {
			if strings.HasPrefix(r.URL.Path, prefix) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Rewrite path so the public route matcher picks it up.
		origPath := r.URL.Path
		ctx = utils.SetOriginalPath(ctx, origPath)
		r = r.WithContext(ctx)
		if origPath == "/" || origPath == "" {
			r.URL.Path = "/" + slug + "/index"
		} else {
			r.URL.Path = "/" + slug + origPath
		}

		tracing.SetAttribute(ctx, "rewritten_path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
