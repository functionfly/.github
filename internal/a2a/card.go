// Package a2a implements the Agent-to-Agent Protocol (A2A) as specified
// in the A2A 0.3.0 spec (LF-governed, post-ACP merge).
//
// Surface area:
//
//	GET  /.well-known/agent.json           — gateway Agent Card (public)
//	GET  /v1/a2a/agents/{agent_id}/card    — per-agent Agent Card
//	POST /v1/a2a/{agent_id}/tasks/send     — send a task
//	GET  /v1/a2a/tasks/{task_id}           — poll task status
//	POST /v1/a2a/tasks/{task_id}/cancel    — cancel a task
//	POST /v1/a2a/tasks/{task_id}/subscribe — SSE stream for task updates
//	POST /v1/a2a/{agent_id}/message/send   — short message
//
// Design notes:
//   - A2A Tasks ARE long-lived Receipts. They live in the same
//     registry_executions_public table with protocol='a2a' and a state
//     column for the A2A state machine. There is NO a2a_tasks table.
//   - The A2A handler is ONLY A2A plumbing (JSON shaping, task state
//     transitions, SSE). It calls GatewayCore.Call for actual execution.
//   - Zero execution logic lives in this package.
package a2a

import (
	"encoding/json"
	"time"

	"github.com/functionfly/functionfly/internal/gateway"
)

// CanTransition returns true if the transition from → to is allowed.
func CanTransition(from, to TaskState) bool {
	return ValidateTransition(from, to)
}

// SendTaskRequest is the wire shape for POST /v1/a2a/{agent_id}/tasks/send.
type SendTaskRequest struct {
	Message  TaskMessage       `json:"message"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// TaskMessage is a message within a task.
type TaskMessage struct {
	Role  string `json:"role"` // "user" | "agent"
	Parts []Part `json:"parts"`
}

// GetTaskResponse is the wire shape for GET /v1/a2a/tasks/{task_id}.
type GetTaskResponse struct {
	ID        string            `json:"id"`
	Status    TaskStatusInfo    `json:"status"`
	Artifacts []Artifact        `json:"artifacts,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// TaskStatusInfo contains the current state and optional status message.
type TaskStatusInfo struct {
	State     TaskState `json:"state"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// CancelTaskResponse is the wire shape for POST /v1/a2a/tasks/{task_id}/cancel.
type CancelTaskResponse struct {
	ID     string         `json:"id"`
	Status TaskStatusInfo `json:"status"`
}

// SendMessageRequest is the wire shape for POST /v1/a2a/{agent_id}/message/send.
type SendMessageRequest struct {
	Message TaskMessage `json:"message"`
}

// SSEEvent is a Server-Sent Event for task state changes.
type SSEEvent struct {
	Event string      `json:"event"` // "task_status_change" | "task_message" | "task_artifact"
	Data  interface{} `json:"data"`
}

// AgentCardInfo is the per-agent card stored in the agent_cards table.
type AgentCardInfo struct {
	ID              string          `json:"id" gorm:"primaryKey;type:text"`
	Version         string          `json:"version" gorm:"type:text;not null;default:'1.0'"`
	Name            string          `json:"name" gorm:"type:text;not null"`
	Description     string          `json:"description" gorm:"type:text"`
	URL             string          `json:"url" gorm:"type:text"`
	ProtocolVersion string          `json:"protocol_version" gorm:"type:text;not null;default:'0.3.0'"`
	Capabilities    []string        `json:"capabilities" gorm:"type:text[];not null;default:'{}'"`
	Skills          json.RawMessage `json:"skills" gorm:"type:jsonb;not null;default:'[]'"`
	AuthSchemes     []string        `json:"auth_schemes" gorm:"type:text[];not null;default:'{}'"`
	InputModes      []string        `json:"input_modes" gorm:"type:text[];not null;default:'{application/json}'"`
	OutputModes     []string        `json:"output_modes" gorm:"type:text[];not null;default:'{application/json}'"`
	TrustScore      float64         `json:"trust_score" gorm:"not null;default:0"`
	PeerJWKSURL     string          `json:"peer_jwks_url" gorm:"type:text"`
	PublishedAt     time.Time       `json:"published_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name for AgentCardInfo.
func (AgentCardInfo) TableName() string {
	return "agent_cards"
}

// ToAgentCard converts an AgentCardInfo to the wire-format AgentCard.
func (a *AgentCardInfo) ToAgentCard() gateway.AgentCard {
	var skills []gateway.AgentSkill
	if len(a.Skills) > 0 {
		_ = json.Unmarshal(a.Skills, &skills)
	}
	return gateway.AgentCard{
		Name:               a.Name,
		Description:        a.Description,
		URL:                a.URL,
		Version:            a.Version,
		ProtocolVersion:    a.ProtocolVersion,
		Capabilities:       a.Capabilities,
		Skills:             skills,
		Authentication:     gateway.AgentAuth{Schemes: a.AuthSchemes},
		DefaultInputModes:  a.InputModes,
		DefaultOutputModes: a.OutputModes,
	}
}

// MessagePart is kept as an alias for Part for backward compatibility.
type MessagePart = Part
