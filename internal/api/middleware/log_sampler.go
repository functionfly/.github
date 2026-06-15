package middleware

import (
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// LogSampler provides configurable log sampling for high-volume endpoints.
// It uses a combination of token bucket and probabilistic sampling to
// reduce log volume while ensuring errors and slow requests are always logged.
type LogSampler struct {
	// Default sampling rate (0.0 to 1.0) for successful requests
	successRate float64
	// Default sampling rate for 4xx client errors
	clientErrorRate float64
	// 5xx errors and slow requests are always logged
	alwaysLogErrors bool
	// Slow request threshold - requests slower than this are always logged
	slowThreshold time.Duration

	// Token bucket for rate-limited sampling
	tokens      int64
	maxTokens   int64
	refillRate  int64 // tokens per second
	lastRefillN int64 // nanoseconds since last refill
	enabled     bool

	// Counters for stats
	totalSeen  atomic.Uint64
	sampledOut atomic.Uint64
	mu         sync.Mutex
}

// LogSamplerConfig configures the log sampler
type LogSamplerConfig struct {
	// SuccessSampleRate is the sampling rate for 2xx responses (0.0 to 1.0)
	SuccessSampleRate float64
	// ClientErrorSampleRate is the sampling rate for 4xx responses (0.0 to 1.0)
	ClientErrorSampleRate float64
	// AlwaysLogErrors ensures 5xx errors are always logged
	AlwaysLogErrors bool
	// SlowThreshold is the duration above which requests are always logged
	SlowThreshold time.Duration
	// RateLimit is the max requests per second to log (0 = no rate limit)
	RateLimit int
	// Burst is the maximum burst size for rate-limited sampling
	Burst int
}

// NewLogSampler creates a new log sampler with the given configuration
func NewLogSampler(config LogSamplerConfig) *LogSampler {
	// Apply defaults
	if config.SuccessSampleRate == 0 {
		config.SuccessSampleRate = 0.1 // Default: 10% of successful requests
	}
	if config.ClientErrorSampleRate == 0 {
		config.ClientErrorSampleRate = 1.0 // Default: log all client errors
	}
	if config.SlowThreshold == 0 {
		config.SlowThreshold = 1 * time.Second
	}
	if config.RateLimit == 0 {
		config.RateLimit = 100 // Default: 100 logs/second
	}
	if config.Burst == 0 {
		config.Burst = 200
	}

	return &LogSampler{
		successRate:     config.SuccessSampleRate,
		clientErrorRate: config.ClientErrorSampleRate,
		alwaysLogErrors: config.AlwaysLogErrors,
		slowThreshold:   config.SlowThreshold,
		tokens:          int64(config.Burst),
		maxTokens:       int64(config.Burst),
		refillRate:      int64(config.RateLimit),
		lastRefillN:     time.Now().UnixNano(),
		enabled:         true,
	}
}

// NewLogSamplerFromEnv creates a log sampler from environment variables
func NewLogSamplerFromEnv() *LogSampler {
	config := LogSamplerConfig{
		SuccessSampleRate:     getEnvFloatFromEnv("LOG_SAMPLE_SUCCESS_RATE", 0.1),
		ClientErrorSampleRate: getEnvFloatFromEnv("LOG_SAMPLE_CLIENT_ERROR_RATE", 1.0),
		AlwaysLogErrors:       getEnvBool("LOG_SAMPLE_ALWAYS_LOG_ERRORS", true),
		SlowThreshold:         getEnvDuration("LOG_SAMPLE_SLOW_THRESHOLD", 1*time.Second),
		RateLimit:             getEnvInt("LOG_SAMPLE_RATE_LIMIT", 100),
		Burst:                 getEnvInt("LOG_SAMPLE_BURST", 200),
	}
	return NewLogSampler(config)
}

// ShouldLog determines whether a log entry should be emitted
func (ls *LogSampler) ShouldLog(statusCode int, duration time.Duration) bool {
	ls.totalSeen.Add(1)

	// 5xx errors are always logged
	if statusCode >= 500 && ls.alwaysLogErrors {
		return true
	}

	// Slow requests are always logged
	if duration >= ls.slowThreshold {
		return true
	}

	// Client errors (4xx) - use client error sampling rate
	if statusCode >= 400 {
		return ls.shouldSample(ls.clientErrorRate)
	}

	// Success - use success sampling rate with rate limiting
	if !ls.takeToken() {
		ls.sampledOut.Add(1)
		return false
	}
	return ls.shouldSample(ls.successRate)
}

// shouldSample returns true based on a sampling rate
func (ls *LogSampler) shouldSample(rate float64) bool {
	if rate >= 1.0 {
		return true
	}
	if rate <= 0.0 {
		return false
	}
	return rand.Float64() < rate
}

// takeToken implements a simple token bucket for rate limiting
func (ls *LogSampler) takeToken() bool {
	now := time.Now().UnixNano()
	ls.mu.Lock()
	defer ls.mu.Unlock()

	// Refill tokens based on elapsed time
	elapsed := now - ls.lastRefillN
	if elapsed > 0 {
		refill := int64(elapsed) * ls.refillRate / int64(time.Second)
		if refill > 0 {
			ls.tokens += refill
			if ls.tokens > ls.maxTokens {
				ls.tokens = ls.maxTokens
			}
			ls.lastRefillN = now
		}
	}

	if ls.tokens > 0 {
		ls.tokens--
		return true
	}
	return false
}

// Stats returns sampling statistics
func (ls *LogSampler) Stats() (seen, sampledOut uint64) {
	return ls.totalSeen.Load(), ls.sampledOut.Load()
}

// LogSamplingMiddleware wraps a handler to apply log sampling
func (ls *LogSampler) LogSamplingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !ls.enabled {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		// Wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		// Only log if sampler says so
		if ls.ShouldLog(rw.statusCode, duration) {
			logger := logrus.WithFields(logrus.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      rw.statusCode,
				"duration_ms": duration.Milliseconds(),
				"ip":          getClientIP(r),
			})
			logger.Info("HTTP request")
		}
	})
}

// getEnvFloatFromEnv reads a float64 from an environment variable, returning the default if unset or invalid.
func getEnvFloatFromEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

// getEnvInt reads an int from an environment variable, returning the default if unset or invalid.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

// getEnvDuration reads a time.Duration from an environment variable (as a string parseable by time.ParseDuration),
// returning the default if unset or invalid.
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
