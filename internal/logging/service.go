package logging

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// Service manages the logging infrastructure
type Service struct {
	config     *Config
	aggregator *Aggregator
	logger     *logrus.Logger
}

// NewService creates a new logging service
func NewService() *Service {
	config := LoadConfig()

	// Configure global logger
	if err := ConfigureLogger(config); err != nil {
		logrus.WithError(err).Fatal("Failed to configure logger")
	}

	service := &Service{
		config: config,
		logger: logrus.New(),
	}

	// Set up log aggregation if endpoints are configured
	endpoints := getLogEndpoints()
	if len(endpoints) > 0 {
		service.aggregator = NewAggregator(100, 30*time.Second, endpoints)

		// Add aggregation hook
		hook := NewHook(service.aggregator, config.Service, config.Version, config.Environment)
		logrus.AddHook(hook)
		service.logger.AddHook(hook)
	}

	return service
}

// Start starts the logging service
func (s *Service) Start(ctx context.Context) error {
	logrus.WithFields(logrus.Fields{
		"service":     s.config.Service,
		"version":     s.config.Version,
		"environment": s.config.Environment,
		"level":       s.config.Level,
		"format":      s.config.Format,
	}).Info("Logging service started")

	// Start aggregator if configured
	if s.aggregator != nil {
		go s.aggregator.Start(ctx)

		// Handle graceful shutdown
		go func() {
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh

			logrus.Info("Shutting down logging service...")
			s.aggregator.Stop()
		}()
	}

	return nil
}

// Stop stops the logging service
func (s *Service) Stop() {
	if s.aggregator != nil {
		s.aggregator.Stop()
	}
	logrus.Info("Logging service stopped")
}

// GetLogger returns a logger instance
func (s *Service) GetLogger() *logrus.Logger {
	return s.logger
}

// CreateRequestLogger creates a logger for HTTP requests
func (s *Service) CreateRequestLogger(requestID, method, path string) *logrus.Entry {
	return RequestLogger(requestID, method, path)
}

// CreateErrorLogger creates a logger for errors
func (s *Service) CreateErrorLogger(err error, fields map[string]interface{}) *logrus.Entry {
	return ErrorLogger(err, fields)
}

// CreateAuditLogger creates a logger for audit events
func (s *Service) CreateAuditLogger(action, resource, userID string) *logrus.Entry {
	return AuditLogger(action, resource, userID)
}

// LogSystemEvent logs system-level events
func (s *Service) LogSystemEvent(event string, fields map[string]interface{}) {
	entry := logrus.WithFields(logrus.Fields{
		"event_type": "system",
		"event":      event,
	})

	if fields != nil {
		entry = entry.WithFields(fields)
	}

	entry.Info("System event")
}

// LogPerformanceEvent logs performance-related events
func (s *Service) LogPerformanceEvent(operation string, duration time.Duration, fields map[string]interface{}) {
	entry := logrus.WithFields(logrus.Fields{
		"event_type":   "performance",
		"operation":    operation,
		"duration_ms":  duration.Milliseconds(),
		"duration":     duration.String(),
	})

	if fields != nil {
		entry = entry.WithFields(fields)
	}

	entry.Info("Performance event")
}

// getLogEndpoints retrieves log aggregation endpoints from environment
func getLogEndpoints() []string {
	endpoints := []string{}

	// Check for comma-separated endpoints
	if endpointStr := os.Getenv("LOG_ENDPOINTS"); endpointStr != "" {
		// Simple split by comma (could be enhanced)
		endpoints = []string{endpointStr}
	}

	// Check for individual endpoints
	if lokiURL := os.Getenv("LOKI_URL"); lokiURL != "" {
		endpoints = append(endpoints, lokiURL+"/loki/api/v1/push")
	}

	if elasticURL := os.Getenv("ELASTICSEARCH_URL"); elasticURL != "" {
		endpoints = append(endpoints, elasticURL+"/_bulk")
	}

	return endpoints
}