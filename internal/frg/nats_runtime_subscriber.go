package frg

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

// RuntimeRegistration represents a runtime cell/agent registration message
type RuntimeRegistration struct {
	CellID       string   `json:"cell_id"`
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities"`
	Timestamp    string   `json:"timestamp,omitempty"`
}

// RuntimeHeartbeat represents a heartbeat from a runtime
type RuntimeHeartbeat struct {
	CellID           string `json:"cell_id"`
	Status           string `json:"status"`
	ActiveExecutions uint32 `json:"active_executions"`
	Timestamp        string `json:"timestamp,omitempty"`
}

// RuntimeExecutionResult represents an execution result from a runtime
type RuntimeExecutionResult struct {
	ExecutionID string `json:"execution_id"`
	CellID      string `json:"cell_id"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}

// RuntimeStatusReport represents a runtime status report
type RuntimeStatusReport struct {
	Healthy      bool   `json:"healthy"`
	ActiveCells  uint32 `json:"active_cells"`
	ActiveSwarms uint32 `json:"active_swarms"`
}

// RuntimeSubscriberConfig configures the NATS runtime subscriber
type RuntimeSubscriberConfig struct {
	// Subjects to subscribe to for cell/agent registration
	RegisterSubjects []string
	// Subjects for heartbeats
	HeartbeatSubjects []string
	// Subjects for execution results
	ExecutionResultSubjects []string
	// Heartbeat timeout — how long before a runtime is considered stale
	HeartbeatTimeout time.Duration
}

// DefaultRuntimeSubscriberConfig returns default configuration
func DefaultRuntimeSubscriberConfig() *RuntimeSubscriberConfig {
	return &RuntimeSubscriberConfig{
		RegisterSubjects: []string{
			"prism.cell.registered",
			"orchestrator.agent.registered",
		},
		HeartbeatSubjects: []string{
			"prism.cell.heartbeat",
			"orchestrator.agent.heartbeat",
		},
		ExecutionResultSubjects: []string{
			"prism.execution.result",
		},
		HeartbeatTimeout: 90 * time.Second,
	}
}

// RuntimeEventHandlers contains callbacks for runtime events
type RuntimeEventHandlers struct {
	OnRegistration     func(RuntimeRegistration)
	OnHeartbeat        func(RuntimeHeartbeat)
	OnExecutionResult  func(RuntimeExecutionResult)
	OnRuntimeStatus    func(RuntimeStatusReport)
}

// RuntimeSubscriber subscribes to NATS messages from Prism/SAR/Kotlin runtimes
// and dispatches them to registered handlers.
type RuntimeSubscriber struct {
	nc            *nats.Conn
	config        *RuntimeSubscriberConfig
	handlers      RuntimeEventHandlers
	subscriptions []*nats.Subscription
	mu            sync.Mutex
	running       bool
}

// NewRuntimeSubscriber creates a new runtime subscriber
func NewRuntimeSubscriber(nc *nats.Conn, config *RuntimeSubscriberConfig, handlers RuntimeEventHandlers) *RuntimeSubscriber {
	if config == nil {
		config = DefaultRuntimeSubscriberConfig()
	}
	return &RuntimeSubscriber{
		nc:       nc,
		config:   config,
		handlers: handlers,
	}
}

// Start begins subscribing to all configured NATS subjects
func (rs *RuntimeSubscriber) Start() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.running {
		return nil
	}

	// Subscribe to registration messages
	for _, subject := range rs.config.RegisterSubjects {
		sub, err := rs.nc.Subscribe(subject, rs.handleRegistration)
		if err != nil {
			logrus.WithError(err).WithField("subject", subject).Error("Failed to subscribe to registration subject")
			continue
		}
		rs.subscriptions = append(rs.subscriptions, sub)
		logrus.WithField("subject", subject).Info("Subscribed to runtime registration")
	}

	// Subscribe to heartbeat messages
	for _, subject := range rs.config.HeartbeatSubjects {
		sub, err := rs.nc.Subscribe(subject, rs.handleHeartbeat)
		if err != nil {
			logrus.WithError(err).WithField("subject", subject).Error("Failed to subscribe to heartbeat subject")
			continue
		}
		rs.subscriptions = append(rs.subscriptions, sub)
		logrus.WithField("subject", subject).Info("Subscribed to runtime heartbeat")
	}

	// Subscribe to execution results
	for _, subject := range rs.config.ExecutionResultSubjects {
		sub, err := rs.nc.Subscribe(subject, rs.handleExecutionResult)
		if err != nil {
			logrus.WithError(err).WithField("subject", subject).Error("Failed to subscribe to execution result subject")
			continue
		}
		rs.subscriptions = append(rs.subscriptions, sub)
		logrus.WithField("subject", subject).Info("Subscribed to runtime execution results")
	}

	// Subscribe to runtime status
	sub, err := rs.nc.Subscribe("prism.runtime.status", rs.handleRuntimeStatus)
	if err != nil {
		logrus.WithError(err).Error("Failed to subscribe to runtime status")
	} else {
		rs.subscriptions = append(rs.subscriptions, sub)
		logrus.Info("Subscribed to runtime status")
	}

	rs.running = true
	logrus.WithField("subscriptions", len(rs.subscriptions)).Info("NATS runtime subscriber started")
	return nil
}

// Stop unsubscribes from all subjects
func (rs *RuntimeSubscriber) Stop() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	for _, sub := range rs.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			logrus.WithError(err).Warn("Failed to unsubscribe")
		}
	}
	rs.subscriptions = nil
	rs.running = false
	logrus.Info("NATS runtime subscriber stopped")
}

// handleRegistration processes runtime registration messages
func (rs *RuntimeSubscriber) handleRegistration(msg *nats.Msg) {
	// Messages may be CBOR-encoded (Prism) or JSON (SAR/Kotlin).
	// Try JSON first (most common), then fall back to raw parsing.
	var reg RuntimeRegistration
	if err := json.Unmarshal(msg.Data, &reg); err != nil {
		logrus.WithError(err).WithField("subject", msg.Subject).Debug("Failed to parse registration message")
		return
	}

	logrus.WithFields(logrus.Fields{
		"cell_id": reg.CellID,
		"name":    reg.Name,
		"subject": msg.Subject,
	}).Info("Runtime registered via NATS")

	if rs.handlers.OnRegistration != nil {
		rs.handlers.OnRegistration(reg)
	}
}

// handleHeartbeat processes heartbeat messages
func (rs *RuntimeSubscriber) handleHeartbeat(msg *nats.Msg) {
	var hb RuntimeHeartbeat
	if err := json.Unmarshal(msg.Data, &hb); err != nil {
		logrus.WithError(err).WithField("subject", msg.Subject).Debug("Failed to parse heartbeat message")
		return
	}

	if rs.handlers.OnHeartbeat != nil {
		rs.handlers.OnHeartbeat(hb)
	}
}

// handleExecutionResult processes execution result messages
func (rs *RuntimeSubscriber) handleExecutionResult(msg *nats.Msg) {
	var result RuntimeExecutionResult
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		logrus.WithError(err).WithField("subject", msg.Subject).Debug("Failed to parse execution result message")
		return
	}

	logrus.WithFields(logrus.Fields{
		"execution_id": result.ExecutionID,
		"cell_id":      result.CellID,
		"status":       result.Status,
	}).Debug("Execution result received via NATS")

	if rs.handlers.OnExecutionResult != nil {
		rs.handlers.OnExecutionResult(result)
	}
}

// handleRuntimeStatus processes runtime status messages
func (rs *RuntimeSubscriber) handleRuntimeStatus(msg *nats.Msg) {
	var status RuntimeStatusReport
	if err := json.Unmarshal(msg.Data, &status); err != nil {
		logrus.WithError(err).WithField("subject", msg.Subject).Debug("Failed to parse runtime status message")
		return
	}

	logrus.WithFields(logrus.Fields{
		"healthy":       status.Healthy,
		"active_cells":  status.ActiveCells,
		"active_swarms": status.ActiveSwarms,
	}).Debug("Runtime status received via NATS")

	if rs.handlers.OnRuntimeStatus != nil {
		rs.handlers.OnRuntimeStatus(status)
	}
}
