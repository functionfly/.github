// Package flywheel provides HTTP handlers for the Flywheel Network
package flywheel

import (
	"encoding/json"

	"github.com/functionfly/functionfly/internal/flywheel"
)

// CreateThreadRequest represents a request to create a thread
type CreateThreadRequest struct {
	Title             string          `json:"title"`
	Type              string          `json:"type"`
	Content           string          `json:"content"`
	CategoryID        *string         `json:"category_id,omitempty"`
	Tags              []string        `json:"tags,omitempty"`
	ProblemData       json.RawMessage `json:"problem_data,omitempty"`
	EnvironmentSpecs  json.RawMessage `json:"environment_specs,omitempty"`
	ExpectedOutput    json.RawMessage `json:"expected_output,omitempty"`
	AttachedCapsuleID *string         `json:"attached_capsule_id,omitempty"`
}

// CreateReplyRequest represents a request to create a reply
type CreateReplyRequest struct {
	Content           string          `json:"content"`
	ParentReplyID     *string         `json:"parent_reply_id,omitempty"`
	CodeBlocks        json.RawMessage `json:"code_blocks,omitempty"`
	AttachedCapsuleID *string         `json:"attached_capsule_id,omitempty"`
}

// ChallengeSubmissionRequest represents a request to submit to a challenge
type ChallengeSubmissionRequest struct {
	SubmissionType     string  `json:"submission_type"`
	CodeSubmission     string  `json:"code_submission,omitempty"`
	SubmittedCapsuleID *string `json:"submitted_capsule_id,omitempty"`
	Notes              string  `json:"notes,omitempty"`
}

// AgentResponseRequest represents a request for an agent to respond
type AgentResponseRequest struct {
	Content string `json:"content"`
}

// MarketplacePublishRequest represents a request to publish to marketplace
type MarketplacePublishRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Visibility  string   `json:"visibility"`
	Price       *float64 `json:"price,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// ResolveThreadRequest represents a request to resolve a thread
type ResolveThreadRequest struct {
	ReplyID string `json:"reply_id"`
}

// ExecuteReplyRequest represents a request to execute a reply
type ExecuteReplyRequest struct {
	Input json.RawMessage `json:"input"`
}

// NotificationLevelRequest represents a request to set notification level
type NotificationLevelRequest struct {
	NotificationLevel string `json:"notification_level"`
}

// Type aliases for flywheel package types
type ThreadFilter = flywheel.ThreadFilter
type ReplyFilter = flywheel.ReplyFilter
type ChallengeFilter = flywheel.ChallengeFilter
