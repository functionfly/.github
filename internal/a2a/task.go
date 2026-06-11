package a2a

import (
	"encoding/json"
	"time"
)

// TaskState represents the A2A task lifecycle state machine.
type TaskState string

const (
	StateSubmitted     TaskState = "submitted"
	StateWorking       TaskState = "working"
	StateInputRequired TaskState = "input-required"
	StateCompleted     TaskState = "completed"
	StateFailed        TaskState = "failed"
	StateCanceled      TaskState = "canceled"
)

// ValidTransitions defines the allowed state transitions for an A2A task.
// The key is the current state; the value is the set of states it can move to.
var ValidTransitions = map[TaskState][]TaskState{
	StateSubmitted:     {StateWorking, StateFailed, StateCanceled},
	StateWorking:       {StateCompleted, StateFailed, StateInputRequired, StateCanceled},
	StateInputRequired: {StateWorking, StateFailed, StateCanceled},
	StateCompleted:     {},
	StateFailed:        {},
	StateCanceled:      {},
}

// Task represents an A2A task. The execution data lives in
// registry_executions_public (same table as MCP receipts); this
// struct holds the A2A-specific envelope.
type Task struct {
	ID           string          `json:"id"`
	State        TaskState       `json:"state"`
	Artifacts    []Artifact      `json:"artifacts,omitempty"`
	Messages     []Message       `json:"messages,omitempty"`
	UpdatedAt    time.Time       `json:"updatedAt"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
}

// Artifact is a piece of output produced by the task.
type Artifact struct {
	Name  string          `json:"name,omitempty"`
	Parts []Part          `json:"parts"`
}

// Part is a unit of content in a message or artifact.
type Part struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
	MimeType string          `json:"mimeType,omitempty"`
}

// Message is a turn in the task conversation.
type Message struct {
	Role     string `json:"role"` // "user" | "agent"
	Parts    []Part `json:"parts"`
}

// TaskStateTransition records a state change for observability.
type TaskStateTransition struct {
	From      TaskState `json:"from"`
	To        TaskState `json:"to"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// ValidateTransition checks whether a state transition is allowed.
func ValidateTransition(from, to TaskState) bool {
	allowed, ok := ValidTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}
