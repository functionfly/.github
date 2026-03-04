// Package flywheel provides models and types for the Flywheel Network
// A Proof-of-Execution Knowledge Network for FunctionFly
package flywheel

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ThreadType represents the type of thread
type ThreadType string

const (
	ThreadTypeProblem    ThreadType = "problem"
	ThreadTypeDiscussion ThreadType = "discussion"
	ThreadTypeChallenge  ThreadType = "challenge"
)

// ThreadStatus represents the status of a thread
type ThreadStatus string

const (
	ThreadStatusOpen       ThreadStatus = "open"
	ThreadStatusInProgress ThreadStatus = "in_progress"
	ThreadStatusResolved   ThreadStatus = "resolved"
	ThreadStatusClosed     ThreadStatus = "closed"
)

// ReplyAuthorType represents the type of reply author
type ReplyAuthorType string

const (
	ReplyAuthorTypeUser  ReplyAuthorType = "user"
	ReplyAuthorTypeAgent ReplyAuthorType = "agent"
)

// ReputationScoreType represents the type of reputation score
type ReputationScoreType string

const (
	ReputationScoreTypeBuilder        ReputationScoreType = "builder"
	ReputationScoreTypeOptimizer      ReputationScoreType = "optimizer"
	ReputationScoreTypeMentor         ReputationScoreType = "mentor"
	ReputationScoreTypeAgentWhisperer ReputationScoreType = "agent_whisperer"
	ReputationScoreTypeReliability    ReputationScoreType = "reliability"
)

// Tier represents the reputation tier
type Tier int

const (
	TierBronze   Tier = 1
	TierSilver   Tier = 2
	TierGold     Tier = 3
	TierPlatinum Tier = 4
	TierDiamond  Tier = 5
	TierLegend   Tier = 6
)

// ChallengeType represents the type of challenge
type ChallengeType string

const (
	ChallengeTypeOptimization ChallengeType = "optimization"
	ChallengeTypeCost         ChallengeType = "cost"
	ChallengeTypeSpeed        ChallengeType = "speed"
	ChallengeTypeAccuracy     ChallengeType = "accuracy"
)

// ChallengeStatus represents the status of a challenge
type ChallengeStatus string

const (
	ChallengeStatusDraft  ChallengeStatus = "draft"
	ChallengeStatusActive ChallengeStatus = "active"
	ChallengeStatusPaused ChallengeStatus = "paused"
	ChallengeStatusEnded  ChallengeStatus = "ended"
)

// VerificationStatus represents the verification status
type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "pending"
	VerificationStatusVerified VerificationStatus = "verified"
	VerificationStatusFailed   VerificationStatus = "failed"
)

// CollaborationType represents the type of agent collaboration
type CollaborationType string

const (
	CollaborationTypeAssist  CollaborationType = "assist"
	CollaborationTypeDebate  CollaborationType = "debate"
	CollaborationTypeCompete CollaborationType = "compete"
)

// Thread represents a community thread (problem, discussion, or challenge)
type Thread struct {
	ID                uuid.UUID       `json:"id" db:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title             string          `json:"title" db:"title" gorm:"not null;index"`
	Type              ThreadType      `json:"type" db:"type" gorm:"not null;index"`
	Status            ThreadStatus    `json:"status" db:"status" gorm:"not null;index;default:'open'"`
	AuthorID          uuid.UUID       `json:"author_id" db:"author_id" gorm:"type:uuid;not null;index"`
	CategoryID        *uuid.UUID      `json:"category_id,omitempty" db:"category_id" gorm:"type:uuid;index"`
	Tags              []string        `json:"tags" db:"tags" gorm:"type:text[]"`
	ProblemData       json.RawMessage `json:"problem_data,omitempty" db:"problem_data" gorm:"type:jsonb"`
	EnvironmentSpecs  json.RawMessage `json:"environment_specs,omitempty" db:"environment_specs" gorm:"type:jsonb"`
	ExpectedOutput    json.RawMessage `json:"expected_output,omitempty" db:"expected_output" gorm:"type:jsonb"`
	AttachedCapsuleID *uuid.UUID      `json:"attached_capsule_id,omitempty" db:"attached_capsule_id" gorm:"type:uuid"`
	ViewCount         int             `json:"view_count" db:"view_count" gorm:"default:0"`
	EngagementScore   float64         `json:"engagement_score" db:"engagement_score" gorm:"default:0"`
	AcceptedReplyID   *uuid.UUID      `json:"accepted_reply_id,omitempty" db:"accepted_reply_id" gorm:"type:uuid;index"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at" gorm:"not null;default:now()"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at" gorm:"not null;default:now()"`
	ResolvedAt        *time.Time      `json:"resolved_at,omitempty" db:"resolved_at"`

	// Relationships
	Author   *User     `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	Category *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Replies  []Reply   `json:"replies,omitempty" gorm:"foreignKey:ThreadID"`
}

// TableName specifies the table name for Thread
func (Thread) TableName() string {
	return "flywheel_threads"
}

// Reply represents a reply/solution to a thread
type Reply struct {
	ID                 uuid.UUID       `json:"id" db:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ThreadID           uuid.UUID       `json:"thread_id" db:"thread_id" gorm:"type:uuid;not null;index"`
	ParentReplyID      *uuid.UUID      `json:"parent_reply_id,omitempty" db:"parent_reply_id" gorm:"type:uuid;index"`
	AuthorID           uuid.UUID       `json:"author_id" db:"author_id" gorm:"type:uuid;not null;index"`
	AuthorType         ReplyAuthorType `json:"author_type" db:"author_type" gorm:"not null;default:'user'"`
	Content            string          `json:"content" db:"content" gorm:"type:text;not null"`
	CodeBlocks         json.RawMessage `json:"code_blocks,omitempty" db:"code_blocks" gorm:"type:jsonb"`
	AttachedCapsuleID  *uuid.UUID      `json:"attached_capsule_id,omitempty" db:"attached_capsule_id" gorm:"type:uuid"`
	AgentForkID        *uuid.UUID      `json:"agent_fork_id,omitempty" db:"agent_fork_id" gorm:"type:uuid"`
	ExecutionResults   json.RawMessage `json:"execution_results,omitempty" db:"execution_results" gorm:"type:jsonb"`
	PerformanceMetrics json.RawMessage `json:"performance_metrics,omitempty" db:"performance_metrics" gorm:"type:jsonb"`
	IsAccepted         bool            `json:"is_accepted" db:"is_accepted" gorm:"default:false"`
	HelpfulCount       int             `json:"helpful_count" db:"helpful_count" gorm:"default:0"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at" gorm:"not null;default:now()"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at" gorm:"not null;default:now()"`

	// Relationships
	Thread       *Thread `json:"thread,omitempty" gorm:"foreignKey:ThreadID"`
	Author       *User   `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	ParentReply  *Reply  `json:"parent_reply,omitempty" gorm:"foreignKey:ParentReplyID"`
	ChildReplies []Reply `json:"child_replies,omitempty" gorm:"foreignKey:ParentReplyID"`
}

// TableName specifies the table name for Reply
func (Reply) TableName() string {
	return "flywheel_replies"
}

// ReputationScores holds all reputation scores for a user
type ReputationScores struct {
	UserID uuid.UUID `json:"user_id" db:"user_id" gorm:"type:uuid;primary_key"`

	// Builder Score
	BuilderScore int  `json:"builder_score" db:"builder_score" gorm:"default:0;check:builder_score >= 0 AND builder_score <= 10000"`
	BuilderTier  Tier `json:"builder_tier" db:"builder_tier" gorm:"default:1;check:builder_tier >= 1 AND builder_tier <= 6"`

	// Optimizer Score
	OptimizerScore int  `json:"optimizer_score" db:"optimizer_score" gorm:"default:0;check:optimizer_score >= 0 AND optimizer_score <= 10000"`
	OptimizerTier  Tier `json:"optimizer_tier" db:"optimizer_tier" gorm:"default:1;check:optimizer_tier >= 1 AND optimizer_tier <= 6"`

	// Mentor Score
	MentorScore int  `json:"mentor_score" db:"mentor_score" gorm:"default:0;check:mentor_score >= 0 AND mentor_score <= 10000"`
	MentorTier  Tier `json:"mentor_tier" db:"mentor_tier" gorm:"default:1;check:mentor_tier >= 1 AND mentor_tier <= 6"`

	// Agent Whisperer Score
	AgentWhispererScore int  `json:"agent_whisperer_score" db:"agent_whisperer_score" gorm:"default:0;check:agent_whisperer_score >= 0 AND agent_whisperer_score <= 10000"`
	AgentWhispererTier  Tier `json:"agent_whisperer_tier" db:"agent_whisperer_tier" gorm:"default:1;check:agent_whisperer_tier >= 1 AND agent_whisperer_tier <= 6"`

	// Reliability Index
	ReliabilityIndex     float64 `json:"reliability_index" db:"reliability_index" gorm:"default:100;check:reliability_index >= 0 AND reliability_index <= 100"`
	TotalExecutions      int     `json:"total_executions" db:"total_executions" gorm:"default:0"`
	SuccessfulExecutions int     `json:"successful_executions" db:"successful_executions" gorm:"default:0"`

	LastCalculatedAt time.Time `json:"last_calculated_at" db:"last_calculated_at" gorm:"default:now()"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for ReputationScores
func (ReputationScores) TableName() string {
	return "flywheel_reputation_scores"
}

// ReputationEvent represents a reputation change event
type ReputationEvent struct {
	ID           uuid.UUID           `json:"id" db:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID           `json:"user_id" db:"user_id" gorm:"type:uuid;not null;index"`
	EventType    string              `json:"event_type" db:"event_type" gorm:"not null;index"`
	ScoreType    ReputationScoreType `json:"score_type" db:"score_type" gorm:"not null;index"`
	PointsChange int                 `json:"points_change" db:"points_change" gorm:"not null"`
	Reason       string              `json:"reason" db:"reason" gorm:"not null"`
	ReferenceID  *uuid.UUID          `json:"reference_id,omitempty" db:"reference_id" gorm:"type:uuid;index"`
	CreatedAt    time.Time           `json:"created_at" db:"created_at" gorm:"not null;default:now()"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for ReputationEvent
func (ReputationEvent) TableName() string {
	return "flywheel_reputation_events"
}

// Challenge represents a community challenge
type Challenge struct {
	ID                  uuid.UUID       `json:"id" db:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title               string          `json:"title" db:"title" gorm:"not null"`
	Description         string          `json:"description" db:"description" gorm:"type:text;not null"`
	ChallengeType       ChallengeType   `json:"challenge_type" db:"challenge_type" gorm:"not null"`
	ObjectiveFunctionID *uuid.UUID      `json:"objective_function_id,omitempty" db:"objective_function_id" gorm:"type:uuid"`
	TargetMetric        string          `json:"target_metric" db:"target_metric"`
	ScoringConfig       json.RawMessage `json:"scoring_config,omitempty" db:"scoring_config" gorm:"type:jsonb"`
	Constraints         json.RawMessage `json:"constraints,omitempty" db:"constraints" gorm:"type:jsonb"`
	StartTime           time.Time       `json:"start_time" db:"start_time" gorm:"not null;index"`
	EndTime             time.Time       `json:"end_time" db:"end_time" gorm:"not null;index"`
	Status              ChallengeStatus `json:"status" db:"status" gorm:"not null;default:'draft';index"`
	Rewards             json.RawMessage `json:"rewards,omitempty" db:"rewards" gorm:"type:jsonb"`
	SponsorID           *uuid.UUID      `json:"sponsor_id,omitempty" db:"sponsor_id" gorm:"type:uuid"`
	CreatedBy           uuid.UUID       `json:"created_by" db:"created_by" gorm:"type:uuid;not null"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at" gorm:"not null;default:now()"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at" gorm:"not null;default:now()"`

	// Relationships
	Submissions []ChallengeSubmission `json:"submissions,omitempty" gorm:"foreignKey:ChallengeID"`
}

// TableName specifies the table name for Challenge
func (Challenge) TableName() string {
	return "flywheel_challenges"
}

// ChallengeSubmission represents a submission to a challenge
type ChallengeSubmission struct {
	ID                 uuid.UUID       `json:"id" db:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ChallengeID        uuid.UUID       `json:"challenge_id" db:"challenge_id" gorm:"type:uuid;not null;index"`
	ParticipantID      uuid.UUID       `json:"participant_id" db:"participant_id" gorm:"type:uuid;not null;index"`
	SubmissionType     string          `json:"submission_type" db:"submission_type" gorm:"not null"`
	SubmittedCapsuleID *uuid.UUID      `json:"submitted_capsule_id,omitempty" db:"submitted_capsule_id" gorm:"type:uuid"`
	CodeSubmission     string          `json:"code_submission,omitempty" db:"code_submission" gorm:"type:text"`
	Metrics            json.RawMessage `json:"metrics,omitempty" db:"metrics" gorm:"type:jsonb"`
	Score              *float64        `json:"score,omitempty" db:"score"`
	Rank               *int            `json:"rank,omitempty" db:"rank"`
	RewardEarned       *float64        `json:"reward_earned,omitempty" db:"reward_earned"`
	Notes              string          `json:"notes,omitempty" db:"notes" gorm:"type:text"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at" gorm:"not null;default:now()"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at" gorm:"not null;default:now()"`

	// Relationships
	Challenge   *Challenge `json:"challenge,omitempty" gorm:"foreignKey:ChallengeID"`
	Participant *User      `json:"participant,omitempty" gorm:"foreignKey:ParticipantID"`
}

// TableName specifies the table name for ChallengeSubmission
func (ChallengeSubmission) TableName() string {
	return "flywheel_challenge_submissions"
}

// AgentCollaboration represents agent participation in threads
type AgentCollaboration struct {
	ID                 uuid.UUID         `json:"id" db:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ThreadID           uuid.UUID         `json:"thread_id" db:"thread_id" gorm:"type:uuid;not null;index"`
	AgentID            uuid.UUID         `json:"agent_id" db:"agent_id" gorm:"type:uuid;not null;index"`
	AgentVersion       string            `json:"agent_version" db:"agent_version" gorm:"not null"`
	CollaborationType  CollaborationType `json:"collaboration_type" db:"collaboration_type" gorm:"not null"`
	ContextSnapshot    json.RawMessage   `json:"context_snapshot,omitempty" db:"context_snapshot" gorm:"type:jsonb"`
	PerformanceMetrics json.RawMessage   `json:"performance_metrics,omitempty" db:"performance_metrics" gorm:"type:jsonb"`
	ReputationEarned   int               `json:"reputation_earned" db:"reputation_earned" gorm:"default:0"`
	InvitedBy          uuid.UUID         `json:"invited_by" db:"invited_by" gorm:"type:uuid;not null"`
	CreatedAt          time.Time         `json:"created_at" db:"created_at" gorm:"not null;default:now()"`
	UpdatedAt          time.Time         `json:"updated_at" db:"updated_at" gorm:"not null;default:now()"`

	// Relationships
	Thread *Thread `json:"thread,omitempty" gorm:"foreignKey:ThreadID"`
	Agent  *User   `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
}

// TableName specifies the table name for AgentCollaboration
func (AgentCollaboration) TableName() string {
	return "flywheel_agent_collaborations"
}

// Execution represents an execution of community code
type Execution struct {
	ID                 uuid.UUID          `json:"id" db:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ReplyID            uuid.UUID          `json:"reply_id" db:"reply_id" gorm:"type:uuid;not null;index"`
	ExecutorID         uuid.UUID          `json:"executor_id" db:"executor_id" gorm:"type:uuid;not null;index"`
	ExecutionContext   json.RawMessage    `json:"execution_context,omitempty" db:"execution_context" gorm:"type:jsonb"`
	Input              json.RawMessage    `json:"input,omitempty" db:"input" gorm:"type:jsonb"`
	Output             json.RawMessage    `json:"output,omitempty" db:"output" gorm:"type:jsonb"`
	Error              *string            `json:"error,omitempty" db:"error"`
	VerificationStatus VerificationStatus `json:"verification_status" db:"verification_status" gorm:"default:'pending';index"`
	ComputeCost        *float64           `json:"compute_cost,omitempty" db:"compute_cost"`
	RuntimeMS          *int               `json:"runtime_ms,omitempty" db:"runtime_ms"`
	MemoryMB           *int               `json:"memory_mb,omitempty" db:"memory_mb"`
	IsDeterministic    *bool              `json:"is_deterministic,omitempty" db:"is_deterministic"`
	DRECertificate     *string            `json:"dre_certificate,omitempty" db:"dre_certificate"`
	CreatedAt          time.Time          `json:"created_at" db:"created_at" gorm:"not null;default:now()"`

	// Relationships
	Reply    *Reply `json:"reply,omitempty" gorm:"foreignKey:ReplyID"`
	Executor *User  `json:"executor,omitempty" gorm:"foreignKey:ExecutorID"`
}

// TableName specifies the table name for Execution
func (Execution) TableName() string {
	return "flywheel_executions"
}

// Category represents a thread category
type Category struct {
	ID          uuid.UUID  `json:"id" db:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string     `json:"name" db:"name" gorm:"not null"`
	Slug        string     `json:"slug" db:"slug" gorm:"not null;uniqueIndex"`
	Description string     `json:"description" db:"description"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" db:"parent_id" gorm:"type:uuid;index"`
	Icon        string     `json:"icon" db:"icon"`
	Color       string     `json:"color" db:"color"`
	SortOrder   int        `json:"sort_order" db:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at" gorm:"not null;default:now()"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at" gorm:"not null;default:now()"`

	// Relationships
	Parent   *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Threads  []Thread   `json:"threads,omitempty" gorm:"foreignKey:CategoryID"`
}

// TableName specifies the table name for Category
func (Category) TableName() string {
	return "flywheel_categories"
}

// Subscription represents a user's subscription to a thread
type Subscription struct {
	UserID            uuid.UUID `json:"user_id" db:"user_id" gorm:"type:uuid;primary_key"`
	ThreadID          uuid.UUID `json:"thread_id" db:"thread_id" gorm:"type:uuid;primary_key"`
	NotificationLevel string    `json:"notification_level" db:"notification_level" gorm:"default:'all'"`
	CreatedAt         time.Time `json:"created_at" db:"created_at" gorm:"not null;default:now()"`

	// Relationships
	User   *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Thread *Thread `json:"thread,omitempty" gorm:"foreignKey:ThreadID"`
}

// TableName specifies the table name for Subscription
func (Subscription) TableName() string {
	return "flywheel_subscriptions"
}

// User represents a user (reusing existing user model for relationships)
type User struct {
	ID    uuid.UUID `json:"id" db:"id" gorm:"type:uuid;primary_key"`
	Email string    `json:"email" db:"email"`
	Name  string    `json:"name" db:"name"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}
