package logging

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Config holds logging configuration
type Config struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	Service    string `json:"service"`
	Version    string `json:"version"`
	Environment string `json:"environment"`
}

// LoadConfig loads logging configuration from environment variables
func LoadConfig() *Config {
	config := &Config{
		Level:       getEnvOrDefault("LOG_LEVEL", "info"),
		Format:      getEnvOrDefault("LOG_FORMAT", "json"),
		Output:      getEnvOrDefault("LOG_OUTPUT", "stdout"),
		Service:     getEnvOrDefault("SERVICE_NAME", "functionfly"),
		Version:     getEnvOrDefault("SERVICE_VERSION", "dev"),
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
	}

	return config
}

// ConfigureLogger configures the global logrus logger
func ConfigureLogger(config *Config) error {
	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)

	// Set log format
	switch strings.ToLower(config.Format) {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
		})
	default:
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	// Set output destination
	switch strings.ToLower(config.Output) {
	case "stdout":
		logrus.SetOutput(os.Stdout)
	case "stderr":
		logrus.SetOutput(os.Stderr)
	default:
		// Assume it's a file path
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		logrus.SetOutput(file)
	}

	// Add default fields
	logrus.WithFields(logrus.Fields{
		"service":     config.Service,
		"version":     config.Version,
		"environment": config.Environment,
		"hostname":    getHostname(),
	}).Info("Logger configured")

	return nil
}

// CreateLogger creates a new logger instance with service context
func CreateLogger(service, version, environment string) *logrus.Logger {
	logger := logrus.New()

	// Configure with same settings as global logger
	config := LoadConfig()
	level, _ := logrus.ParseLevel(config.Level)
	logger.SetLevel(level)

	switch strings.ToLower(config.Format) {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	// Add service context
	logger.WithFields(logrus.Fields{
		"service":     service,
		"version":     version,
		"environment": environment,
	})

	return logger
}

// WithContext adds contextual fields to a logger entry
func WithContext(logger *logrus.Entry, fields map[string]interface{}) *logrus.Entry {
	return logger.WithFields(fields)
}

// RequestLogger creates a logger with request context
func RequestLogger(requestID, method, path string) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"request_id": requestID,
		"method":     method,
		"path":       path,
	})
}

// ErrorLogger creates a logger for errors with context
func ErrorLogger(err error, fields map[string]interface{}) *logrus.Entry {
	entry := logrus.WithError(err)
	if fields != nil {
		entry = entry.WithFields(fields)
	}
	return entry
}

// AuditLogger creates a logger for audit events
func AuditLogger(action, resource, userID string) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"event_type": "audit",
		"action":     action,
		"resource":   resource,
		"user_id":    userID,
	})
}

// Logger returns the package-level default logger. It is a thin wrapper
// around logrus.StandardLogger() so call sites can write logging.Logger()
// without importing logrus directly.
func Logger() *logrus.Logger {
	return logrus.StandardLogger()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getHostname() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return "unknown"
}