package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type PublicRateLimiter struct {
	mu       sync.RWMutex
	requests map[string][]time.Time
	window   time.Duration
	limit    int
}

func NewPublicRateLimiter(limit int, window time.Duration) *PublicRateLimiter {
	return &PublicRateLimiter{
		requests: make(map[string][]time.Time),
		window:   window,
		limit:    limit,
	}
}

func (r *PublicRateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	if times, ok := r.requests[key]; ok {
		valid := make([]time.Time, 0)
		for _, t := range times {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		r.requests[key] = valid
	}

	if len(r.requests[key]) >= r.limit {
		return false
	}

	r.requests[key] = append(r.requests[key], now)
	return true
}

func (r *PublicRateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		key := fmt.Sprintf("%s:%s", req.RemoteAddr, req.URL.Path)
		if !r.Allow(key) {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(r.window.Seconds())))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = fmt.Fprintf(w, `{"message":"Too many requests. Please wait before trying again."}`)
			return
		}
		next.ServeHTTP(w, req)
	}
}