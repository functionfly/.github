package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// RequireInternalSecret restricts routes to callers presenting X-Internal-Webhook-Secret
// matching INTERNAL_WEBHOOK_SECRET. When the secret is unset, only loopback clients are allowed.
func RequireInternalSecret(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			next(w, r)
			return
		}

		internalSecret := os.Getenv("INTERNAL_WEBHOOK_SECRET")
		if internalSecret != "" {
			if r.Header.Get("X-Internal-Webhook-Secret") != internalSecret {
				logrus.Warn("internal route: invalid or missing X-Internal-Webhook-Secret")
				apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
				return
			}
			next(w, r)
			return
		}

		host := r.Host
		remoteAddr := r.RemoteAddr
		if strings.Contains(host, "localhost") ||
			strings.Contains(remoteAddr, "127.0.0.1") ||
			strings.Contains(remoteAddr, "[::1]") {
			next(w, r)
			return
		}

		logrus.Warn("internal route: rejected non-localhost request without INTERNAL_WEBHOOK_SECRET")
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
	}
}
