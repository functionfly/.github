package a2a

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TaskStore is the storage surface the task engine needs. It operates
// on the registry_executions_public table (A2A tasks are receipts).
type TaskStore interface {
	// UpdateTaskState updates the state column of a receipt.
	UpdateTaskState(ctx context.Context, publicID string, fromState, to TaskState) error
	// GetTaskState returns the current state of a task receipt.
	GetTaskState(ctx context.Context, publicID string) (TaskState, error)
	// SetTaskOutput sets the output column when a task completes.
	SetTaskOutput(ctx context.Context, publicID string, output []byte) error
}

// TaskEngine manages A2A task lifecycle. It is the ONLY place that
// writes to registry_executions_public.state for A2A tasks.
type TaskEngine struct {
	store  TaskStore
	logger *logrus.Logger

	// subscribers tracks active SSE subscribers per task ID.
	mu          sync.RWMutex
	subscribers map[string][]chan SSEEvent
}

// NewTaskEngine creates a task engine.
func NewTaskEngine(store TaskStore, logger *logrus.Logger) *TaskEngine {
	if logger == nil {
		logger = logrus.New()
	}
	return &TaskEngine{
		store:       store,
		logger:      logger,
		subscribers: make(map[string][]chan SSEEvent),
	}
}

// Transition attempts a state transition on a task. Returns an error
// if the transition is not allowed or the DB update fails.
func (e *TaskEngine) Transition(ctx context.Context, taskID string, from, to TaskState) error {
	if !CanTransition(from, to) {
		return fmt.Errorf("a2a: invalid transition %s → %s", from, to)
	}
	if e.store == nil {
		return fmt.Errorf("a2a: task store not configured")
	}
	if err := e.store.UpdateTaskState(ctx, taskID, from, to); err != nil {
		return fmt.Errorf("a2a: state update failed: %w", err)
	}

	// Notify SSE subscribers.
	e.notifySubscribers(taskID, SSEEvent{
		Event: "task_status_change",
		Data: map[string]interface{}{
			"task_id":    taskID,
			"from_state": string(from),
			"to_state":   string(to),
		},
	})

	e.logger.WithFields(logrus.Fields{
		"task_id": taskID,
		"from":    string(from),
		"to":      string(to),
	}).Info("a2a: task state transition")
	return nil
}

// Complete marks a task as completed with the given output.
func (e *TaskEngine) Complete(ctx context.Context, taskID string, output []byte) error {
	if e.store != nil {
		if err := e.store.SetTaskOutput(ctx, taskID, output); err != nil {
			return err
		}
	}
	return e.Transition(ctx, taskID, StateWorking, StateCompleted)
}

// Fail marks a task as failed with an error message.
func (e *TaskEngine) Fail(ctx context.Context, taskID, errMsg string) error {
	return e.Transition(ctx, taskID, StateWorking, StateFailed)
}

// Cancel marks a task as canceled.
func (e *TaskEngine) Cancel(ctx context.Context, taskID string) error {
	// Cancel is allowed from submitted, working, or input-required.
	currentState, err := e.store.GetTaskState(ctx, taskID)
	if err != nil {
		return err
	}
	return e.Transition(ctx, taskID, currentState, StateCanceled)
}

// Subscribe registers an SSE channel for task state changes.
// Returns an unsubscribe function.
func (e *TaskEngine) Subscribe(taskID string) (<-chan SSEEvent, func()) {
	ch := make(chan SSEEvent, 16)
	e.mu.Lock()
	e.subscribers[taskID] = append(e.subscribers[taskID], ch)
	e.mu.Unlock()

	unsubscribe := func() {
		e.mu.Lock()
		defer e.mu.Unlock()
		subs := e.subscribers[taskID]
		for i, sub := range subs {
			if sub == ch {
				e.subscribers[taskID] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
	}
	return ch, unsubscribe
}

// notifySubscribers sends an event to all active subscribers for a task.
func (e *TaskEngine) notifySubscribers(taskID string, event SSEEvent) {
	e.mu.RLock()
	subs := e.subscribers[taskID]
	e.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
			// Subscriber is slow; drop the event rather than blocking.
		}
	}
}

// GenerateTaskID generates a new task ID (nanoid).
func GenerateTaskID() string {
	return uuid.New().String()
}

// ParseTaskMessage parses a message into inputs for the executor.
func ParseTaskMessage(msg TaskMessage) ([]byte, error) {
	if len(msg.Parts) == 0 {
		return []byte("{}"), nil
	}
	// If there's a single data part, use it directly.
	for _, part := range msg.Parts {
		if part.Type == "data" && len(part.Data) > 0 {
			return part.Data, nil
		}
	}
	// Otherwise, wrap text parts into a JSON object.
	texts := make([]string, 0, len(msg.Parts))
	for _, part := range msg.Parts {
		if part.Type == "text" && part.Text != "" {
			texts = append(texts, part.Text)
		}
	}
	if len(texts) == 1 {
		return []byte(fmt.Sprintf(`{"text":%q}`, texts[0])), nil
	}
	return []byte(fmt.Sprintf(`{"texts":%q}`, texts)), nil
}
