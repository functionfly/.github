package middleware

import (
	"net/http"
	"sync"
	"time"
	"github.com/functionfly/functionfly/internal/apierror"
)

type HealthRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewHealthRateLimiter(redisClient interface{}) *HealthRateLimiter {
	return &HealthRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    100,
		window:   time.Minute,
	}
}

func (rl *HealthRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	requests := rl.requests[ip]
	var validRequests []time.Time
	for _, t := range requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	if len(validRequests) >= rl.limit {
		rl.requests[ip] = validRequests
		return false
	}

	rl.requests[ip] = append(validRequests, now)
	return true
}

func (rl *HealthRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if ip != "" {
			if !rl.Allow(ip) {
				apierror.WriteError(w, apierror.NewRateLimited("rate limit exceeded"))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *HealthRateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	for ip, requests := range rl.requests {
		var validRequests []time.Time
		for _, t := range requests {
			if t.After(windowStart) {
				validRequests = append(validRequests, t)
			}
		}
		if len(validRequests) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = validRequests
		}
	}
}