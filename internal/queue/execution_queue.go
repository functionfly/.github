package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ExecutionPriority represents the priority level of an execution
type ExecutionPriority int

const (
	PriorityLow    ExecutionPriority = 1
	PriorityNormal ExecutionPriority = 2
	PriorityHigh   ExecutionPriority = 3
	PriorityCritical ExecutionPriority = 4
)

// String returns the string representation of the priority
func (p ExecutionPriority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// QueuedExecution represents an execution waiting in the queue
type QueuedExecution struct {
	ID            string                  `json:"id"`
	FunctionID    uuid.UUID               `json:"function_id"`
	Version       string                  `json:"version"`
	Input         json.RawMessage         `json:"input"`
	Priority      ExecutionPriority       `json:"priority"`
	UserID        *uuid.UUID              `json:"user_id,omitempty"`
	TenantID      *uuid.UUID              `json:"tenant_id,omitempty"`
	RequestID     string                  `json:"request_id"`
	QueuedAt      time.Time               `json:"queued_at"`
	TimeoutMs     int                     `json:"timeout_ms"`
	MaxRetries    int                     `json:"max_retries"`
	RetryCount    int                     `json:"retry_count"`
	LastAttemptAt *time.Time              `json:"last_attempt_at,omitempty"`
	ErrorMessage  string                  `json:"error_message,omitempty"`
	IPAddress     string                  `json:"ip_address,omitempty"`
	UserAgent     string                  `json:"user_agent,omitempty"`
}

// ExecutionQueue manages queued function executions
type ExecutionQueue struct {
	queue        []*QueuedExecution
	mu           sync.RWMutex
	workers      []*ExecutionWorker
	maxQueueSize int
	maxWorkers   int
	processing   map[string]bool // Track currently processing executions

	// Callbacks
	onExecution func(*QueuedExecution) error
}

// ExecutionWorker represents a worker that processes queued executions
type ExecutionWorker struct {
	id       int
	queue    *ExecutionQueue
	stopChan chan bool
}

// NewExecutionQueue creates a new execution queue
func NewExecutionQueue(maxQueueSize, maxWorkers int) *ExecutionQueue {
	if maxQueueSize <= 0 {
		maxQueueSize = 1000
	}
	if maxWorkers <= 0 {
		maxWorkers = 10
	}

	return &ExecutionQueue{
		queue:        make([]*QueuedExecution, 0, maxQueueSize),
		maxQueueSize: maxQueueSize,
		maxWorkers:   maxWorkers,
		processing:   make(map[string]bool),
	}
}

// Start begins processing the execution queue with workers
func (q *ExecutionQueue) Start(onExecution func(*QueuedExecution) error) {
	q.onExecution = onExecution

	// Start workers
	q.workers = make([]*ExecutionWorker, q.maxWorkers)
	for i := 0; i < q.maxWorkers; i++ {
		worker := &ExecutionWorker{
			id:       i + 1,
			queue:    q,
			stopChan: make(chan bool),
		}
		q.workers[i] = worker
		go worker.start()
	}

	logrus.WithFields(logrus.Fields{
		"max_workers": q.maxWorkers,
		"max_queue_size": q.maxQueueSize,
	}).Info("Started execution queue")
}

// Stop gracefully stops the execution queue
func (q *ExecutionQueue) Stop() {
	logrus.Info("Stopping execution queue...")

	// Stop all workers
	for _, worker := range q.workers {
		worker.stopChan <- true
	}

	// Wait a bit for workers to finish
	time.Sleep(2 * time.Second)

	logrus.Info("Execution queue stopped")
}

// Enqueue adds an execution to the queue
func (q *ExecutionQueue) Enqueue(ctx context.Context, exec *QueuedExecution) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check queue size limit
	if len(q.queue) >= q.maxQueueSize {
		return fmt.Errorf("queue is full (max size: %d)", q.maxQueueSize)
	}

	// Set defaults
	if exec.ID == "" {
		exec.ID = uuid.New().String()
	}
	if exec.QueuedAt.IsZero() {
		exec.QueuedAt = time.Now()
	}
	if exec.Priority == 0 {
		exec.Priority = PriorityNormal
	}
	if exec.MaxRetries == 0 {
		exec.MaxRetries = 3
	}

	// Insert based on priority (higher priority = lower index)
	insertIndex := len(q.queue)
	for i, queued := range q.queue {
		if exec.Priority > queued.Priority {
			insertIndex = i
			break
		}
	}

	// Insert at the correct position
	q.queue = append(q.queue, nil)
	copy(q.queue[insertIndex+1:], q.queue[insertIndex:])
	q.queue[insertIndex] = exec

	logrus.WithFields(logrus.Fields{
		"execution_id": exec.ID,
		"function_id":  exec.FunctionID,
		"priority":     exec.Priority.String(),
		"queue_size":   len(q.queue),
	}).Debug("Enqueued execution")

	return nil
}

// Dequeue removes and returns the highest priority execution
func (q *ExecutionQueue) dequeue() *QueuedExecution {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	// Check if this execution is already being processed
	exec := q.queue[0]
	if q.processing[exec.ID] {
		// Remove it from queue since it's being processed
		q.queue = q.queue[1:]
		return nil
	}

	// Mark as processing
	q.processing[exec.ID] = true
	q.queue = q.queue[1:]

	return exec
}

// Requeue puts an execution back in the queue for retry
func (q *ExecutionQueue) Requeue(exec *QueuedExecution, errorMsg string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	exec.LastAttemptAt = &now
	exec.RetryCount++
	exec.ErrorMessage = errorMsg

	// Remove from processing
	delete(q.processing, exec.ID)

	// Check if we've exceeded max retries
	if exec.RetryCount >= exec.MaxRetries {
		logrus.WithFields(logrus.Fields{
			"execution_id": exec.ID,
			"retry_count":  exec.RetryCount,
			"max_retries":  exec.MaxRetries,
			"error":        errorMsg,
		}).Warn("Execution exceeded max retries, dropping")
		return
	}

	// Re-insert into queue with lower priority for retry
	exec.Priority = PriorityLow // Lower priority for retries

	insertIndex := len(q.queue)
	for i, queued := range q.queue {
		if exec.Priority > queued.Priority {
			insertIndex = i
			break
		}
	}

	q.queue = append(q.queue, nil)
	copy(q.queue[insertIndex+1:], q.queue[insertIndex:])
	q.queue[insertIndex] = exec

	logrus.WithFields(logrus.Fields{
		"execution_id": exec.ID,
		"retry_count":  exec.RetryCount,
		"priority":     exec.Priority.String(),
		"queue_size":   len(q.queue),
	}).Info("Requeued execution for retry")
}

// MarkCompleted marks an execution as completed
func (q *ExecutionQueue) MarkCompleted(executionID string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.processing, executionID)
}

// GetStats returns queue statistics
func (q *ExecutionQueue) GetStats() map[string]interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	// Count executions by priority
	priorityCounts := make(map[string]int)
	for _, exec := range q.queue {
		priorityCounts[exec.Priority.String()]++
	}

	return map[string]interface{}{
		"queue_size":     len(q.queue),
		"max_queue_size": q.maxQueueSize,
		"processing":     len(q.processing),
		"workers":        q.maxWorkers,
		"priority_counts": priorityCounts,
	}
}

// start begins the worker's processing loop
func (w *ExecutionWorker) start() {
	logrus.WithField("worker_id", w.id).Info("Started execution worker")

	for {
		select {
		case <-w.stopChan:
			logrus.WithField("worker_id", w.id).Info("Stopping execution worker")
			return
		default:
			// Try to get an execution from the queue
			if exec := w.queue.dequeue(); exec != nil {
				w.processExecution(exec)
			} else {
				// No executions available, wait a bit
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// processExecution processes a single queued execution
func (w *ExecutionWorker) processExecution(exec *QueuedExecution) {
	logrus.WithFields(logrus.Fields{
		"worker_id":    w.id,
		"execution_id": exec.ID,
		"function_id":  exec.FunctionID,
	}).Debug("Processing queued execution")

	// Execute the function
	if w.queue.onExecution != nil {
		if err := w.queue.onExecution(exec); err != nil {
			logrus.WithFields(logrus.Fields{
				"worker_id":    w.id,
				"execution_id": exec.ID,
				"error":        err.Error(),
			}).Error("Failed to execute queued function")

			// Requeue for retry
			w.queue.Requeue(exec, err.Error())
			return
		}
	}

	// Mark as completed
	w.queue.MarkCompleted(exec.ID)

	logrus.WithFields(logrus.Fields{
		"worker_id":    w.id,
		"execution_id": exec.ID,
	}).Debug("Successfully processed queued execution")
}

// CreateQueuedExecution creates a new queued execution from execution request data
func CreateQueuedExecution(functionID uuid.UUID, version string, input json.RawMessage, priority ExecutionPriority, userID, tenantID *uuid.UUID, requestID, ipAddress, userAgent string, timeoutMs int) *QueuedExecution {
	return &QueuedExecution{
		FunctionID: functionID,
		Version:    version,
		Input:      input,
		Priority:   priority,
		UserID:     userID,
		TenantID:   tenantID,
		RequestID:  requestID,
		TimeoutMs:  timeoutMs,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}
}
