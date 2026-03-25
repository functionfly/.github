// Package observability provides production-ready observability configuration
package observability

import (
	"os"
	"strconv"
	"time"
)

// Config holds observability configuration
type Config struct {
	// Logging configuration
	LogLevel  string
	LogFormat string

	// Metrics configuration
	MetricsEnabled bool
	MetricsPort    int

	// Tracing configuration
	TracingEnabled      bool
	TracingSampleRate   float64
	TracingServiceName  string

	// Health check configuration
	HealthCheckEnabled  bool
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration

	// Request tracing
	RequestTracingEnabled bool

	// Structured logging
	StructuredLoggingEnabled bool

	// Metrics collection
	MetricsCollectionEnabled bool

	// Performance monitoring
	PerformanceMonitoringEnabled bool

	// Error tracking
	ErrorTrackingEnabled bool

	// Audit logging
	AuditLoggingEnabled bool
}

// LoadConfig loads observability configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		LogLevel:  getEnvOrDefault("LOG_LEVEL", "info"),
		LogFormat: getEnvOrDefault("LOG_FORMAT", "json"),

		MetricsEnabled: getEnvBoolOrDefault("METRICS_ENABLED", true),
		MetricsPort:    getEnvIntOrDefault("METRICS_PORT", 9090),

		TracingEnabled:      getEnvBoolOrDefault("REQUEST_TRACING_ENABLED", true),
		TracingSampleRate:   getEnvFloatOrDefault("TRACING_SAMPLE_RATE", 1.0),
		TracingServiceName:  getEnvOrDefault("SERVICE_NAME", "functionfly-api"),

		HealthCheckEnabled:  getEnvBoolOrDefault("HEALTH_CHECKS_ENABLED", true),
		HealthCheckInterval: getEnvDurationOrDefault("HEALTH_CHECK_INTERVAL", 30*time.Second),
		HealthCheckTimeout:  getEnvDurationOrDefault("HEALTH_CHECK_TIMEOUT", 10*time.Second),

		RequestTracingEnabled: getEnvBoolOrDefault("REQUEST_TRACING_ENABLED", true),

		StructuredLoggingEnabled: getEnvBoolOrDefault("STRUCTURED_LOGGING_ENABLED", true),

		MetricsCollectionEnabled: getEnvBoolOrDefault("METRICS_COLLECTION_ENABLED", true),

		PerformanceMonitoringEnabled: getEnvBoolOrDefault("PERFORMANCE_MONITORING_ENABLED", true),

		ErrorTrackingEnabled: getEnvBoolOrDefault("ERROR_TRACKING_ENABLED", true),

		AuditLoggingEnabled: getEnvBoolOrDefault("AUDIT_LOGGING_ENABLED", true),
	}
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBoolOrDefault returns environment variable as bool or default
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvIntOrDefault returns environment variable as int or default
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvFloatOrDefault returns environment variable as float64 or default
func getEnvFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvDurationOrDefault returns environment variable as duration or default
func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
