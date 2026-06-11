package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	HeaderRateLimitLimit     = "X-RateLimit-Limit"
	HeaderRateLimitRemaining = "X-RateLimit-Remaining"
	HeaderRateLimitReset     = "X-RateLimit-Reset"
	HeaderRetryAfter         = "Retry-After"
)

type RateLimitInfo struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

type RateLimitHeaders struct {
	mu       sync.RWMutex
	counters map[string]*rateLimitCounter
	cleanup  time.Duration
}

type rateLimitCounter struct {
	count    int
	resetAt  time.Time
	limit    int
	window   time.Duration
	key      string
}

func NewRateLimitHeaders(cleanupDuration time.Duration) *RateLimitHeaders {
	rlh := &RateLimitHeaders{
		counters: make(map[string]*rateLimitCounter),
		cleanup: cleanupDuration,
	}
	go rlh.cleanupLoop()
	return rlh
}

func (rlh *RateLimitHeaders) cleanupLoop() {
	ticker := time.NewTicker(rlh.cleanup)
	defer ticker.Stop()
	for range ticker.C {
		rlh.mu.Lock()
		now := time.Now()
		for key, counter := range rlh.counters {
			if now.After(counter.resetAt) {
				delete(rlh.counters, key)
			}
		}
		rlh.mu.Unlock()
	}
}

func (rlh *RateLimitHeaders) GetRateLimitInfo(key string, limit int, window time.Duration) RateLimitInfo {
	rlh.mu.Lock()
	defer rlh.mu.Unlock()

	now := time.Now()
	counter, exists := rlh.counters[key]

	if !exists || now.After(counter.resetAt) {
		counter = &rateLimitCounter{
			count:   0,
			resetAt: now.Add(window),
			limit:   limit,
			window:  window,
			key:     key,
		}
		rlh.counters[key] = counter
	}

	counter.count++
	remaining := limit - counter.count
	if remaining < 0 {
		remaining = 0
	}

	return RateLimitInfo{
		Limit:     limit,
		Remaining: remaining,
		Reset:     counter.resetAt,
	}
}

func (rlh *RateLimitHeaders) RecordRequest(key string, limit int, window time.Duration) RateLimitInfo {
	info := rlh.GetRateLimitInfo(key, limit, window)
	return info
}

func WriteRateLimitHeaders(w http.ResponseWriter, info RateLimitInfo) {
	w.Header().Set(HeaderRateLimitLimit, strconv.Itoa(info.Limit))
	w.Header().Set(HeaderRateLimitRemaining, strconv.Itoa(info.Remaining))
	w.Header().Set(HeaderRateLimitReset, strconv.FormatInt(info.Reset.Unix(), 10))
}

func WriteRetryAfter(w http.ResponseWriter, retryAfter time.Duration) {
	w.Header().Set(HeaderRetryAfter, strconv.Itoa(int(retryAfter.Seconds())))
}

type RateLimitHeaderMiddleware struct {
	headers *RateLimitHeaders
	limiter RateLimiterProvider
}

type RateLimiterProvider interface {
	Allow(key string) bool
	GetLimit() int
	GetWindow() time.Duration
}

func NewRateLimitHeaderMiddleware(limiter RateLimiterProvider) *RateLimitHeaderMiddleware {
	return &RateLimitHeaderMiddleware{
		headers: NewRateLimitHeaders(5 * time.Minute),
		limiter:  limiter,
	}
}

func (m *RateLimitHeaderMiddleware) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := m.getClientKey(r)
		info := m.headers.RecordRequest(key, m.limiter.GetLimit(), m.limiter.GetWindow())
		WriteRateLimitHeaders(w, info)

		if !m.limiter.Allow(key) {
			retryAfter := time.Until(info.Reset)
			WriteRetryAfter(w, retryAfter)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (m *RateLimitHeaderMiddleware) getClientKey(r *http.Request) string {
	if claims := GetUserFromContext(r); claims != nil {
		return fmt.Sprintf("user:%s", claims.UserID)
	}
	return fmt.Sprintf("ip:%s", GetClientIP(r))
}

type GlobalRateLimitInfo struct {
	Limit     int
	Remaining int
	Reset     time.Time
	RetryAfter time.Duration
}

type GlobalRateLimiter struct {
	limit         int
	window        time.Duration
	requests      map[string][]time.Time
	mu            sync.Mutex
	enabled       bool
}

func NewGlobalRateLimiter(limit int, window time.Duration) *GlobalRateLimiter {
	return &GlobalRateLimiter{
		limit:    limit,
		window:   window,
		requests: make(map[string][]time.Time),
		enabled:  true,
	}
}

func (g *GlobalRateLimiter) Allow(key string) bool {
	if !g.enabled {
		return true
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-g.window)

	requests := g.requests[key]
	var validRequests []time.Time
	for _, t := range requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	if len(validRequests) >= g.limit {
		g.requests[key] = validRequests
		return false
	}

	validRequests = append(validRequests, now)
	g.requests[key] = validRequests
	return true
}

func (g *GlobalRateLimiter) GetInfo(key string) GlobalRateLimitInfo {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-g.window)

	requests := g.requests[key]
	var validRequests []time.Time
	for _, t := range requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	remaining := g.limit - len(validRequests)
	if remaining < 0 {
		remaining = 0
	}

	var reset time.Time
	var retryAfter time.Duration
	if len(validRequests) >= g.limit {
		if len(validRequests) > 0 {
			reset = validRequests[0].Add(g.window)
			retryAfter = time.Until(reset)
		} else {
			retryAfter = g.window
		}
	}

	return GlobalRateLimitInfo{
		Limit:      g.limit,
		Remaining:  remaining,
		Reset:      reset,
		RetryAfter: retryAfter,
	}
}

func (g *GlobalRateLimiter) GetLimit() int {
	return g.limit
}

func (g *GlobalRateLimiter) GetWindow() time.Duration {
	return g.window
}

func (g *GlobalRateLimiter) SetEnabled(enabled bool) {
	g.enabled = enabled
}

type RateLimitStatusWriter struct {
	http.ResponseWriter
	mu            sync.MWMutex
	headers       *RateLimitHeaders
	limiter       RateLimiterProvider
	wroteStatus   bool
	statusCode    int
}

func NewRateLimitStatusWriter(w http.ResponseWriter, limiter RateLimiterProvider) *RateLimitStatusWriter {
	return &RateLimitStatusWriter{
		ResponseWriter: w,
		limiter:        limiter,
	}
}

func (w *RateLimitStatusWriter) WriteHeader(statusCode int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.wroteStatus {
		w.wroteStatus = true
		w.statusCode = statusCode
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *RateLimitStatusWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.wroteStatus {
		w.wroteStatus = true
		w.statusCode = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

type ResponseStatusAnnotator struct {
	delegate http.Handler
}

func NewResponseStatusAnnotator(delegate http.Handler) *ResponseStatusAnnotator {
	return &ResponseStatusAnnotator{delegate: delegate}
}

func (a *ResponseStatusAnnotator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.delegate.ServeHTTP(w, r)
}

type RateLimitResponseWriter struct {
	http.ResponseWriter
	rlh     *RateLimitHeaders
	limiter RateLimiterProvider
	key     string
}

func NewRateLimitResponseWriter(w http.ResponseWriter, r *http.Request, limiter RateLimiterProvider, rlh *RateLimitHeaders) *RateLimitResponseWriter {
	key := getClientKeyForRateLimiter(r, limiter)
	return &RateLimitResponseWriter{
		ResponseWriter: w,
		rlh:            rlh,
		limiter:        limiter,
		key:            key,
	}
}

func getClientKeyForRateLimiter(r *http.Request, limiter RateLimiterProvider) string {
	if claims := GetUserFromContext(r); claims != nil {
		return fmt.Sprintf("user:%s", claims.UserID)
	}
	return fmt.Sprintf("ip:%s", GetClientIP(r))
}

func (w *RateLimitResponseWriter) WriteHeader(statusCode int) {
	if statusCode == http.StatusTooManyRequests {
		info := w.rlh.GetRateLimitInfo(w.key, w.limiter.GetLimit(), w.limiter.GetWindow())
		WriteRateLimitHeaders(w.ResponseWriter, info)
		WriteRetryAfter(w.ResponseWriter, info.Reset.Sub(time.Now()))
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

type RateLimitLogger struct {
	logger *logrus.Logger
}

func NewRateLimitLogger(logger *logrus.Logger) *RateLimitLogger {
	if logger == nil {
		logger = logrus.New()
	}
	return &RateLimitLogger{logger: logger}
}

func (l *RateLimitLogger) LogRateLimitHit(key string, limit int) {
	l.logger.WithFields(logrus.Fields{
		"key":   key,
		"limit": limit,
	}).Warn("Rate limit exceeded")
}
