package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// MetricsRateLimiter implements rate limiting for metrics scraping
type MetricsRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
	redis    *redis.Client
	useRedis bool
	stopCh   chan struct{}
}

// NewMetricsRateLimiter creates a rate limiter for metrics endpoints
func NewMetricsRateLimiter(redisClient *redis.Client) *MetricsRateLimiter {
	rl := &MetricsRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    60, // 60 requests per minute for metrics scraping (Prometheus typically scrapes every 15-60s)
		window:   time.Minute,
		redis:    redisClient,
		useRedis: redisClient != nil,
		stopCh:   make(chan struct{}),
	}
	// Start background cleanup goroutine for in-memory mode
	if !rl.useRedis {
		go rl.cleanupLoop()
	}
	return rl
}

// cleanupLoop periodically removes expired entries to prevent unbounded memory growth
func (rl *MetricsRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCh:
			return
		}
	}
}

// cleanup removes expired entries from the requests map
func (rl *MetricsRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	for ip, requests := range rl.requests {
		// Filter to only valid requests within the window
		validRequests := make([]time.Time, 0)
		for _, t := range requests {
			if t.After(windowStart) {
				validRequests = append(validRequests, t)
			}
		}

		if len(validRequests) == 0 {
			// Remove entry entirely if no valid requests remain
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = validRequests
		}
	}
}

// Stop stops the background cleanup goroutine
func (rl *MetricsRateLimiter) Stop() {
	if rl.stopCh != nil {
		close(rl.stopCh)
	}
}

// Allow checks if a request should be allowed
func (rl *MetricsRateLimiter) Allow(ip string) bool {
	if rl.useRedis {
		return rl.allowRedis(ip)
	}
	return rl.allowLocal(ip)
}

func (rl *MetricsRateLimiter) allowLocal(ip string) bool {
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

func (rl *MetricsRateLimiter) allowRedis(ip string) bool {
	ctx := context.Background()
	key := "metrics_rate_limit:" + ip

	count, err := rl.redis.Incr(ctx, key).Result()
	if err != nil {
		logrus.WithError(err).Warn("Redis metrics rate limit error, falling back to local")
		return rl.allowLocal(ip)
	}

	// Set expiry on first request
	if count == 1 {
		rl.redis.Expire(ctx, key, rl.window)
	}

	return count <= int64(rl.limit)
}

// Limit wraps an HTTP handler with metrics rate limiting
func (rl *MetricsRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)
		if !rl.Allow(ip) {
			http.Error(w, "rate limit exceeded for metrics", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// MetricsRateLimit creates a rate-limited metrics handler
// This prevents DoS via excessive metrics scraping while allowing normal Prometheus scraping
func MetricsRateLimit(handler http.Handler, redisClient *redis.Client) http.Handler {
	limiter := NewMetricsRateLimiter(redisClient)
	return limiter.Limit(handler)
}

// Prometheus metrics for monitoring the metrics endpoint itself
var (
	metricsRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_metrics_requests_total",
			Help: "Total number of metrics scrape requests",
		},
		[]string{"status"},
	)

	metricsRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_metrics_request_duration_seconds",
			Help:    "Duration of metrics scrape requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{},
	)

	metricsRateLimited = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "functionfly_metrics_rate_limited_total",
			Help: "Total number of rate-limited metrics requests",
		},
	)
)

func init() {
	prometheus.MustRegister(metricsRequestsTotal)
	prometheus.MustRegister(metricsRequestDuration)
	prometheus.MustRegister(metricsRateLimited)
}

// InstrumentedMetricsHandler wraps the metrics handler with instrumentation
func InstrumentedMetricsHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &statusCodeWriter{ResponseWriter: w, statusCode: 200}
		handler.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)

		metricsRequestsTotal.WithLabelValues(status).Inc()
		metricsRequestDuration.WithLabelValues().Observe(duration)

		if wrapped.statusCode == 429 {
			metricsRateLimited.Inc()
		}
	})
}

type statusCodeWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusCodeWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
