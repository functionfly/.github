package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	"github.com/functionfly/functionfly/internal/apierror"
)

type MessageRateLimiter struct {
	mu       sync.RWMutex
	requests map[string][]time.Time
	window   time.Duration
	limit    int
}

func NewMessageRateLimiter() *MessageRateLimiter {
	return &MessageRateLimiter{
		requests: make(map[string][]time.Time),
		window:   time.Minute,
		limit:    100,
	}
}

func (r *MessageRateLimiter) Allow(key string) bool {
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

func (r *MessageRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		key := req.RemoteAddr
		if !r.Allow(key) {
			apierror.WriteError(w, apierror.NewRateLimited("Rate limit exceeded"))
			return
		}
		next.ServeHTTP(w, req)
	})
}

func (r *MessageRateLimiter) LimitCreate(next http.HandlerFunc) http.HandlerFunc {
	return r.limitByAction("create", next)
}

func (r *MessageRateLimiter) LimitEdit(next http.HandlerFunc) http.HandlerFunc {
	return r.limitByAction("edit", next)
}

func (r *MessageRateLimiter) LimitDelete(next http.HandlerFunc) http.HandlerFunc {
	return r.limitByAction("delete", next)
}

func (r *MessageRateLimiter) LimitAttachment(next http.HandlerFunc) http.HandlerFunc {
	return r.limitByAction("attachment", next)
}

func (r *MessageRateLimiter) LimitReact(next http.HandlerFunc) http.HandlerFunc {
	return r.limitByAction("react", next)
}

func (r *MessageRateLimiter) limitByAction(action string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		key := fmt.Sprintf("%s:%s:%s", req.RemoteAddr, req.URL.Path, action)
		if !r.Allow(key) {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = fmt.Fprintf(w, `{"error":"rate_limit_exceeded","code":"MESSAGE_RATE_LIMIT","message":"Too many %s requests. Please try again later."}`, action)
			return
		}
		next.ServeHTTP(w, req)
	}
}