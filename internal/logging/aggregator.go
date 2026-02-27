package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time         `json:"timestamp"`
	Level       string            `json:"level"`
	Message     string            `json:"message"`
	Service     string            `json:"service"`
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	RequestID   string            `json:"request_id,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
	TenantID    string            `json:"tenant_id,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Error       string            `json:"error,omitempty"`
	StackTrace  string            `json:"stack_trace,omitempty"`
}

// Aggregator handles log aggregation and forwarding
type Aggregator struct {
	buffer     []*LogEntry
	bufferSize int
	mutex      sync.Mutex
	client     *http.Client
	endpoints  []string
	flushInterval time.Duration
	stopCh     chan struct{}
}

// NewAggregator creates a new log aggregator
func NewAggregator(bufferSize int, flushInterval time.Duration, endpoints []string) *Aggregator {
	return &Aggregator{
		buffer:        make([]*LogEntry, 0, bufferSize),
		bufferSize:    bufferSize,
		client:        &http.Client{Timeout: 10 * time.Second},
		endpoints:     endpoints,
		flushInterval: flushInterval,
		stopCh:        make(chan struct{}),
	}
}

// Start begins the log aggregation process
func (a *Aggregator) Start(ctx context.Context) {
	ticker := time.NewTicker(a.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.flush()
			return
		case <-ticker.C:
			a.flush()
		case <-a.stopCh:
			a.flush()
			return
		}
	}
}

// Stop stops the aggregator and flushes remaining logs
func (a *Aggregator) Stop() {
	close(a.stopCh)
}

// AddLogEntry adds a log entry to the buffer
func (a *Aggregator) AddLogEntry(entry *LogEntry) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.buffer = append(a.buffer, entry)

	// Flush if buffer is full
	if len(a.buffer) >= a.bufferSize {
		a.flushLocked()
	}
}

// flush flushes the current buffer to all endpoints
func (a *Aggregator) flush() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.flushLocked()
}

// flushLocked flushes the buffer (must be called with mutex held)
func (a *Aggregator) flushLocked() {
	if len(a.buffer) == 0 {
		return
	}

	// Send to all configured endpoints
	for _, endpoint := range a.endpoints {
		if err := a.sendBatch(endpoint, a.buffer); err != nil {
			logrus.WithError(err).WithField("endpoint", endpoint).Error("Failed to send log batch")
		}
	}

	// Clear buffer
	a.buffer = a.buffer[:0]
}

// sendBatch sends a batch of log entries to an endpoint
func (a *Aggregator) sendBatch(endpoint string, entries []*LogEntry) error {
	payload, err := json.Marshal(map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
		"timestamp": time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal log batch: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", getLogToken()))

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("endpoint returned status %d", resp.StatusCode)
	}

	return nil
}

// Hook is a logrus hook that forwards logs to the aggregator
type Hook struct {
	aggregator *Aggregator
	service    string
	version    string
	environment string
}

// NewHook creates a new logrus hook for the aggregator
func NewHook(aggregator *Aggregator, service, version, environment string) *Hook {
	return &Hook{
		aggregator:  aggregator,
		service:     service,
		version:     version,
		environment: environment,
	}
}

// Fire sends the log entry to the aggregator
func (h *Hook) Fire(entry *logrus.Entry) error {
	logEntry := &LogEntry{
		Timestamp:   entry.Time,
		Level:       entry.Level.String(),
		Message:     entry.Message,
		Service:     h.service,
		Version:     h.version,
		Environment: h.environment,
		Fields:      make(map[string]interface{}),
	}

	// Extract known fields
	if requestID, ok := entry.Data["request_id"]; ok {
		if rid, ok := requestID.(string); ok {
			logEntry.RequestID = rid
		}
	}

	if userID, ok := entry.Data["user_id"]; ok {
		if uid, ok := userID.(string); ok {
			logEntry.UserID = uid
		}
	}

	if tenantID, ok := entry.Data["tenant_id"]; ok {
		if tid, ok := tenantID.(string); ok {
			logEntry.TenantID = tid
		}
	}

	// Add error information
	if err, ok := entry.Data[logrus.ErrorKey]; ok {
		if errStr, ok := err.(string); ok {
			logEntry.Error = errStr
		}
	}

	// Add remaining fields
	for key, value := range entry.Data {
		if key != "request_id" && key != "user_id" && key != "tenant_id" && key != logrus.ErrorKey {
			logEntry.Fields[key] = value
		}
	}

	h.aggregator.AddLogEntry(logEntry)
	return nil
}

// Levels returns the log levels to hook into
func (h *Hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
	}
}

// getLogToken retrieves the log aggregation token from environment
func getLogToken() string {
	// Try environment variable first
	if token := os.Getenv("LOG_AGGREGATION_TOKEN"); token != "" {
		return token
	}

	// Try alternative environment variable names
	if token := os.Getenv("LOG_TOKEN"); token != "" {
		return token
	}

	// In development/testing, use a placeholder (but log a warning)
	logrus.Warn("LOG_AGGREGATION_TOKEN environment variable not set, using development placeholder")
	return "dev-log-aggregation-token"
}