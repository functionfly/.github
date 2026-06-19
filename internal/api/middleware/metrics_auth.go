package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

func MetricsAuthMiddleware(authMiddleware *AuthMiddleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metricsToken := os.Getenv("METRICS_AUTH_TOKEN")
			metricsUser := os.Getenv("METRICS_USER")
			metricsPassword := os.Getenv("METRICS_PASSWORD")

			hasTokenAuth := metricsToken != ""
			hasBasicAuth := metricsUser != "" && metricsPassword != ""

			if !hasTokenAuth && !hasBasicAuth {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
				return
			}

			if strings.HasPrefix(authHeader, "Basic ") {
				if !validateMetricsBasicAuth(r) {
					apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			if strings.HasPrefix(authHeader, "Bearer ") {
				if hasTokenAuth {
					token := strings.TrimPrefix(authHeader, "Bearer ")
					if subtle.ConstantTimeCompare([]byte(token), []byte(metricsToken)) == 1 {
						next.ServeHTTP(w, r)
						return
					}
					apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
					return
				}
				authMiddleware.RequireAuth(http.HandlerFunc(next.ServeHTTP)).ServeHTTP(w, r)
				return
			}

			apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		})
	}
}

func validateMetricsBasicAuth(r *http.Request) bool {
	metricsUser := os.Getenv("METRICS_USER")
	metricsPassword := os.Getenv("METRICS_PASSWORD")

	if metricsUser == "" || metricsPassword == "" {
		logrus.Warn("METRICS_USER or METRICS_PASSWORD not set, rejecting basic auth for /metrics")
		return false
	}

	authEncoded := r.Header.Get("Authorization")[6:]
	decoded, err := base64.StdEncoding.DecodeString(authEncoded)
	if err != nil {
		return false
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return false
	}

	user := parts[0]
	password := parts[1]

	userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(metricsUser)) == 1
	passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(metricsPassword)) == 1

	return userMatch && passwordMatch
}
