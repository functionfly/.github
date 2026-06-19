package middleware

import (
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// RecoveryMiddleware recovers from panics in HTTP handlers and returns a 500
// response instead of crashing the server. It logs the stack trace via logrus
// for post-mortem debugging.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logrus.WithFields(logrus.Fields{
					"panic":       rec,
					"method":      r.Method,
					"path":        r.URL.Path,
					"stack":       string(debug.Stack()),
					"remote_addr": r.RemoteAddr,
				}).Error("Panic recovered in HTTP handler")

				apierror.WriteError(w, apierror.NewInternal("Internal Server Error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequireAuthInProduction gates internal endpoints (Swagger UI, /metrics)
// behind a valid Authorization header when not in development mode.
// In development, requests pass through unimpeded.
func RequireAuthInProduction(authMiddleware *AuthMiddleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isDev := os.Getenv("DEVELOPMENT") == "true"
			if isDev {
				next.ServeHTTP(w, r)
				return
			}

			// Production: require a valid JWT
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
				return
			}
			// Delegate actual token validation to the existing auth middleware
			authMiddleware.RequireAuth(http.HandlerFunc(next.ServeHTTP)).ServeHTTP(w, r)
		})
	}
}
