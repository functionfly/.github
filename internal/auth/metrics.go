package auth

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// AuthCallbacksTotal counts authentication callback attempts by provider and outcome
	AuthCallbacksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_callbacks_total",
			Help: "Total number of authentication callbacks by provider and outcome",
		},
		[]string{"provider", "outcome"},
	)

	// AuthCallbackDuration tracks the latency of authentication callbacks
	AuthCallbackDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_callback_duration_seconds",
			Help:    "Duration of authentication callback processing in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider"},
	)

	// AuthLoginTotal counts login attempts by method and outcome
	AuthLoginTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_logins_total",
			Help: "Total number of login attempts by method and outcome",
		},
		[]string{"method", "outcome"},
	)

	// AuthTokenRefreshesTotal counts token refresh attempts
	AuthTokenRefreshesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_token_refreshes_total",
			Help: "Total number of token refresh attempts by outcome",
		},
		[]string{"outcome"},
	)

	// AuthAccountLinksTotal counts account linking attempts
	AuthAccountLinksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_account_links_total",
			Help: "Total number of account linking attempts by outcome",
		},
		[]string{"outcome"},
	)
)

// RecordAuthCallback records metrics for an authentication callback
// Usage: defer RecordAuthCallback("oauth", "success")()
func RecordAuthCallback(provider, outcome string) func() {
	start := time.Now()
	return func() {
		AuthCallbacksTotal.WithLabelValues(provider, outcome).Inc()
		AuthCallbackDuration.WithLabelValues(provider).Observe(time.Since(start).Seconds())
	}
}

// RecordAuthCallbackOutcome records just the outcome (for error cases where you can't use defer)
func RecordAuthCallbackOutcome(provider, outcome string, duration time.Duration) {
	AuthCallbacksTotal.WithLabelValues(provider, outcome).Inc()
	AuthCallbackDuration.WithLabelValues(provider).Observe(duration.Seconds())
}

// RecordLogin records a login attempt
func RecordLogin(method, outcome string) {
	AuthLoginTotal.WithLabelValues(method, outcome).Inc()
}

// RecordTokenRefresh records a token refresh attempt
func RecordTokenRefresh(outcome string) {
	AuthTokenRefreshesTotal.WithLabelValues(outcome).Inc()
}

// RecordAccountLink records an account linking attempt
func RecordAccountLink(outcome string) {
	AuthAccountLinksTotal.WithLabelValues(outcome).Inc()
}
