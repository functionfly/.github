package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// EventType represents the type of audit event
type EventType string

const (
	// EventExecute is logged when an agent executes a function
	EventExecute EventType = "execute"
	// EventSpawn is logged when an agent spawns a child
	EventSpawn EventType = "spawn"
	// EventMessage is logged when an agent sends a message
	EventMessage EventType = "message"
	// EventPolicyViolation is logged when a policy is violated
	EventPolicyViolation EventType = "policy_violation"
	// EventQuotaViolation is logged when a quota is exceeded
	EventQuotaViolation EventType = "quota_violation"
	// EventAuthFailure is logged when authentication fails
	EventAuthFailure EventType = "auth_failure"
	// EventCapabilityViolation is logged when a capability is violated
	EventCapabilityViolation EventType = "capability_violation"
	// EventCircuitOpen is logged when circuit breaker opens
	EventCircuitOpen EventType = "circuit_open"
	// EventRetryExhausted is logged when retries are exhausted
	EventRetryExhausted EventType = "retry_exhausted"
)

// AuditEvent represents an audit log entry
type AuditEvent struct {
	ID              uuid.UUID              `json:"id"`
	Timestamp       time.Time              `json:"timestamp"`
	EventType       EventType              `json:"event_type"`
	AgentID         string                 `json:"agent_id"`
	TenantID        uuid.UUID              `json:"tenant_id"`
	FunctionURI     string                 `json:"function_uri,omitempty"`
	IPAddress       string                 `json:"ip_address,omitempty"`
	UserAgent       string                 `json:"user_agent,omitempty"`
	Outcome         string                 `json:"outcome"`
	Details         map[string]interface{} `json:"details,omitempty"`
	RiskScore       float64                `json:"risk_score"`
	ExecutionID     string                 `json:"execution_id,omitempty"`
	SessionID       string                 `json:"session_id,omitempty"`
	CallDepth       int                    `json:"call_depth,omitempty"`
	LatencyMs       int                    `json:"latency_ms,omitempty"`
	CostUSD         float64                `json:"cost_usd,omitempty"`
	ErrorCode       string                 `json:"error_code,omitempty"`
	PolicyViolation string                 `json:"policy_violation,omitempty"`
}

// Repository defines the interface for audit log storage
type Repository interface {
	Create(ctx context.Context, event *AuditEvent) error
	List(ctx context.Context, agentID string, eventType string, since string, limit, offset int) ([]*AuditEvent, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AuditEvent, error)
	DeleteOlderThan(ctx context.Context, olderThan time.Time) error
}

// Logger handles audit logging
type Logger struct {
	repo           Repository
	logger         *logrus.Logger
	fallbackPath   string
	fallbackMu     sync.Mutex
}

// NewLogger creates a new audit logger
func NewLogger(repo Repository) *Logger {
	return &Logger{
		repo:   repo,
		logger: logrus.New(),
	}
}

// NewLoggerWithFallback creates a new audit logger with file-based fallback
func NewLoggerWithFallback(repo Repository, fallbackPath string) *Logger {
	return &Logger{
		repo:         repo,
		logger:       logrus.New(),
		fallbackPath: fallbackPath,
	}
}

// Log logs an audit event
func (l *Logger) Log(ctx context.Context, event *AuditEvent) {
	// Set defaults
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Calculate risk score
	event.RiskScore = l.calculateRiskScore(event)

	// Log to database
	if l.repo != nil {
		if err := l.repo.Create(ctx, event); err != nil {
			l.logger.WithError(err).Error("failed to write audit log to database")
			l.writeFallback(event)
		}
	} else if l.fallbackPath != "" {
		l.writeFallback(event)
	}

	// Log high-risk events to stdout
	if event.RiskScore > 0.7 {
		l.logger.WithFields(logrus.Fields{
			"event_type":   event.EventType,
			"agent_id":     event.AgentID,
			"risk_score":   event.RiskScore,
			"function_uri": event.FunctionURI,
			"outcome":      event.Outcome,
			"details":      event.Details,
		}).Warn("high-risk audit event")
	}
}

// LogExecute logs an execution event
func (l *Logger) LogExecute(ctx context.Context, agentID string, tenantID uuid.UUID, functionURI string, executionID string, outcome string, latencyMs int, costUSD float64, errorCode string) {
	l.Log(ctx, &AuditEvent{
		EventType:   EventExecute,
		AgentID:     agentID,
		TenantID:    tenantID,
		FunctionURI: functionURI,
		ExecutionID: executionID,
		Outcome:     outcome,
		LatencyMs:   latencyMs,
		CostUSD:     costUSD,
		ErrorCode:   errorCode,
		Details: map[string]interface{}{
			"execution_id": executionID,
			"latency_ms":   latencyMs,
			"cost_usd":     costUSD,
		},
	})
}

// LogPolicyViolation logs a policy violation event
func (l *Logger) LogPolicyViolation(ctx context.Context, agentID string, tenantID uuid.UUID, functionURI string, violationCode string, violationMessage string) {
	l.Log(ctx, &AuditEvent{
		EventType:       EventPolicyViolation,
		AgentID:         agentID,
		TenantID:        tenantID,
		FunctionURI:     functionURI,
		Outcome:         "blocked",
		PolicyViolation: violationCode,
		Details: map[string]interface{}{
			"violation_code":    violationCode,
			"violation_message": violationMessage,
		},
	})
}

// LogQuotaViolation logs a quota violation event
func (l *Logger) LogQuotaViolation(ctx context.Context, agentID string, tenantID uuid.UUID, functionURI string, quotaType string, limit float64, current float64) {
	l.Log(ctx, &AuditEvent{
		EventType:   EventQuotaViolation,
		AgentID:     agentID,
		TenantID:    tenantID,
		FunctionURI: functionURI,
		Outcome:     "blocked",
		Details: map[string]interface{}{
			"quota_type": quotaType,
			"limit":      limit,
			"current":    current,
		},
	})
}

// LogAuthFailure logs an authentication failure event
func (l *Logger) LogAuthFailure(ctx context.Context, agentID string, tenantID uuid.UUID, ipAddress string, userAgent string, reason string) {
	l.Log(ctx, &AuditEvent{
		EventType: EventAuthFailure,
		AgentID:   agentID,
		TenantID:  tenantID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Outcome:   "failed",
		Details: map[string]interface{}{
			"reason": reason,
		},
	})
}

// LogCapabilityViolation logs a capability violation event
func (l *Logger) LogCapabilityViolation(ctx context.Context, agentID string, tenantID uuid.UUID, functionURI string, capability string) {
	l.Log(ctx, &AuditEvent{
		EventType:   EventCapabilityViolation,
		AgentID:     agentID,
		TenantID:    tenantID,
		FunctionURI: functionURI,
		Outcome:     "blocked",
		Details: map[string]interface{}{
			"capability": capability,
		},
	})
}

// LogCircuitOpen logs a circuit breaker open event
func (l *Logger) LogCircuitOpen(ctx context.Context, agentID string, tenantID uuid.UUID, functionURI string) {
	l.Log(ctx, &AuditEvent{
		EventType:   EventCircuitOpen,
		AgentID:     agentID,
		TenantID:    tenantID,
		FunctionURI: functionURI,
		Outcome:     "blocked",
		Details: map[string]interface{}{
			"reason": "circuit breaker is open",
		},
	})
}

// LogRetryExhausted logs a retry exhausted event
func (l *Logger) LogRetryExhausted(ctx context.Context, agentID string, tenantID uuid.UUID, functionURI string, attempts int, lastError string) {
	l.Log(ctx, &AuditEvent{
		EventType:   EventRetryExhausted,
		AgentID:     agentID,
		TenantID:    tenantID,
		FunctionURI: functionURI,
		Outcome:     "failed",
		Details: map[string]interface{}{
			"attempts":   attempts,
			"last_error": lastError,
		},
	})
}

func (l *Logger) calculateRiskScore(event *AuditEvent) float64 {
	score := 0.0

	switch event.EventType {
	case EventPolicyViolation:
		score += 0.5
	case EventQuotaViolation:
		score += 0.3
	case EventAuthFailure:
		score += 0.7
	case EventCapabilityViolation:
		score += 0.6
	case EventCircuitOpen:
		score += 0.4
	case EventRetryExhausted:
		score += 0.2
	}

	// Increase score for errors
	if event.Outcome == "failed" || event.Outcome == "blocked" {
		score += 0.1
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (l *Logger) writeFallback(event *AuditEvent) {
	if l.fallbackPath == "" {
		return
	}

	l.fallbackMu.Lock()
	defer l.fallbackMu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		l.logger.WithError(err).Error("failed to marshal audit event for fallback")
		return
	}

	filename := filepath.Join(l.fallbackPath, "audit-"+event.ID.String()+".json")
	if err := os.WriteFile(filename, data, 0600); err != nil {
		l.logger.WithError(err).Error("failed to write audit event to fallback file")
		return
	}

	l.logger.WithFields(logrus.Fields{
		"event_id": event.ID,
		"path":     filename,
	}).Warn("audit event written to fallback storage")
}