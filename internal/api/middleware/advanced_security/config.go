package advanced_security

import (
	"time"
)

// AdvancedSecurityConfig holds configuration for advanced security features
type AdvancedSecurityConfig struct {
	// Rate limiting
	SlidingWindowLimit int
	SlidingWindowWindow time.Duration
	TokenBucketRate    float64
	TokenBucketBurst   int

	// DDoS protection
	EnableBotDetection      bool
	EnableTrafficAnalysis   bool
	SuspiciousThreshold     int
	BlockDuration           time.Duration

	// Traffic management
	CircuitBreakerThreshold float64
	CircuitBreakerTimeout   time.Duration
	QueueSize               int
	QueueTimeout            time.Duration

	// Geo-blocking
	BlockedCountries        []string
	BlockedIPs             []string
	AllowedIPs             []string

	// Advanced filtering
	EnableSQLInjectionFilter bool
	EnableXSSFilter         bool
	EnablePathTraversalFilter bool

	// Monitoring
	MetricsEnabled         bool
}