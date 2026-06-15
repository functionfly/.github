package middleware

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultRequestTimeout = 30 * time.Second
	envRequestTimeout     = "REQUEST_TIMEOUT"
)

func RequestTimeoutMiddleware(next http.Handler) http.Handler {
	timeout := getRequestTimeout()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		done := make(chan struct{})
		go func() {
			next.ServeHTTP(w, r.WithContext(ctx))
			close(done)
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			logrus.WithFields(logrus.Fields{
				"method":  r.Method,
				"path":    r.URL.Path,
				"timeout": timeout.String(),
			}).Warn("Request timed out")
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
		}
	})
}

func getRequestTimeout() time.Duration {
	if val := os.Getenv(envRequestTimeout); val != "" {
		if duration, err := time.ParseDuration(val); err == nil && duration > 0 {
			return duration
		}
	}
	return defaultRequestTimeout
}
