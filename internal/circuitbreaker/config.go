package circuitbreaker

import (
	"os"
	"strconv"
	"time"
)

// Config holds circuit breaker configuration.
type Config struct {
	// FailureThreshold is the number of consecutive failures before opening the circuit.
	FailureThreshold int
	// SuccessThreshold is the number of successes in half-open state before closing.
	SuccessThreshold int
	// BaseCooldown is the base duration to wait in OPEN before transitioning to HALF_OPEN.
	BaseCooldown time.Duration
	// MaxCooldown is the maximum cooldown with exponential backoff.
	MaxCooldown time.Duration
	// BackoffMultiplier is the multiplier for exponential backoff on repeated reopenings.
	BackoffMultiplier float64
	// HalfOpenMaxRequests is the maximum number of requests allowed in HALF_OPEN state.
	HalfOpenMaxRequests int
	// OnStateChange is called (asynchronously) when the state changes.
	OnStateChange func(key string, from, to State)
	// Persistence is optional; when set, circuit state is synced to a store.
	Persistence Persistence
}

// DefaultConfig returns production-ready defaults.
func DefaultConfig() Config {
	return Config{
		FailureThreshold:    3,
		SuccessThreshold:    2,
		BaseCooldown:        30 * time.Second,
		MaxCooldown:         5 * time.Minute,
		BackoffMultiplier:   2.0,
		HalfOpenMaxRequests: 3,
	}
}

// ConfigFromEnv loads configuration from environment variables with fallback to defaults.
func ConfigFromEnv() Config {
	cfg := DefaultConfig()

	if v := envInt("CIRCUIT_BREAKER_FAILURE_THRESHOLD"); v > 0 {
		cfg.FailureThreshold = v
	}
	if v := envInt("CIRCUIT_BREAKER_SUCCESS_THRESHOLD"); v > 0 {
		cfg.SuccessThreshold = v
	}
	if v := envDuration("CIRCUIT_BREAKER_BASE_COOLDOWN"); v > 0 {
		cfg.BaseCooldown = v
	}
	if v := envDuration("CIRCUIT_BREAKER_MAX_COOLDOWN"); v > 0 {
		cfg.MaxCooldown = v
	}
	if v := envFloat("CIRCUIT_BREAKER_BACKOFF_MULTIPLIER"); v > 0 {
		cfg.BackoffMultiplier = v
	}
	if v := envInt("CIRCUIT_BREAKER_HALF_OPEN_MAX"); v > 0 {
		cfg.HalfOpenMaxRequests = v
	}

	return cfg
}

func envInt(key string) int {
	v := os.Getenv(key)
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}

func envDuration(key string) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return 0
	}
	d, _ := time.ParseDuration(v)
	return d
}

func envFloat(key string) float64 {
	v := os.Getenv(key)
	if v == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(v, 64)
	return f
}
