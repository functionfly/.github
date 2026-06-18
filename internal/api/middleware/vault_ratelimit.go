package middleware

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// VaultRateLimitConfig tunes the per-(user, target, IP) rate limiter
// applied to client-mode dynamic-credential issuance endpoints.
type VaultRateLimitConfig struct {
	// PerTargetPerMinute caps how many times a single user can
	// issue against a single target per minute.
	PerTargetPerMinute int
	// PerIPPerMinute caps the global issuance rate per source IP.
	PerIPPerMinute int
	// PerTargetPerHour caps the per-target rate over a sliding hour.
	PerTargetPerHour int
	// Logger is used for limiter hits.
	Logger *logrus.Logger
	// KeyPrefix is prepended to the rate-limit key (e.g. "vault:dyn:").
	KeyPrefix string
	// Incr is the counter increment callback. It returns the new
	// count + the TTL to set if the key was just created. The
	// returned bool is true if the call should be rejected.
	Incr func(key string, ttl time.Duration) (count int64, rejected bool, err error)
}

// VaultRateLimit returns a middleware that enforces the configured
// per-(user, target, IP) rate limits on client-mode issuance. The
// user/target/IP identifiers are extracted from the request: user
// from JWT context, target from the URL path, IP from RemoteAddr /
// X-Forwarded-For.
func VaultRateLimit(cfg VaultRateLimitConfig) func(http.Handler) http.Handler {
	if cfg.PerTargetPerMinute == 0 {
		cfg.PerTargetPerMinute = 10
	}
	if cfg.PerIPPerMinute == 0 {
		cfg.PerIPPerMinute = 100
	}
	if cfg.PerTargetPerHour == 0 {
		cfg.PerTargetPerHour = 1000
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "vault:dyn:"
	}
	logger := cfg.Logger
	if logger == nil {
		logger = logrus.New()
	}
	_ = logger
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserFromContext(r)
			if claims == nil {
				next.ServeHTTP(w, r)
				return
			}
			target := r.PathValue("id")
			if target == "" {
				target = "n/a"
			}
			ip := vaultClientIP(r)
			uid := claims.UserID.String()

			checks := []struct {
				key      string
				ttl      time.Duration
				limit    int
				scope    string
			}{
				{fmt.Sprintf("%s%s:%s:m", cfg.KeyPrefix, uid, target), time.Minute, cfg.PerTargetPerMinute, "per_target_per_minute"},
				{fmt.Sprintf("%sip:%s:m", cfg.KeyPrefix, ip), time.Minute, cfg.PerIPPerMinute, "per_ip_per_minute"},
				{fmt.Sprintf("%s%s:%s:h", cfg.KeyPrefix, uid, target), time.Hour, cfg.PerTargetPerHour, "per_target_per_hour"},
			}
			if cfg.Incr != nil {
				for _, c := range checks {
					count, rejected, err := cfg.Incr(c.key, c.ttl)
					if err != nil {
						// Best-effort: do not block on a transient
						// rate-limiter failure.
						continue
					}
					if rejected {
						rejectRateLimit(w, c.scope, count, c.limit, int(c.ttl.Seconds()))
						return
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func rejectRateLimit(w http.ResponseWriter, scope string, count int64, limit int, retryAfter int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "rate limit exceeded",
		"code":    "VAULT_RATE_LIMIT",
		"scope":   scope,
		"count":   count,
		"limit":   limit,
		"message": "Too many dynamic-credential operations. Please slow down.",
	})
}

func vaultClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for _, ip := range strings.Split(xff, ",") {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				return ip
			}
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
