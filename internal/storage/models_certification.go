package storage

import (
	"time"

	"github.com/google/uuid"
)

// CertTier represents a certification level (Associate, Professional, Architect)
type CertTier struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Slug             string    `json:"slug" gorm:"uniqueIndex;size:50;not null"`
	Name             string    `json:"name" gorm:"size:255;not null"`
	Description      string    `json:"description" gorm:"type:text;not null"`
	Icon             string    `json:"icon" gorm:"size:100"`
	Color            string    `json:"color" gorm:"size:20"`
	SortOrder        int       `json:"sort_order" gorm:"default:0"`
	PriceCents       int       `json:"price_cents" gorm:"default:0"`
	Currency         string    `json:"currency" gorm:"size:3;default:'USD'"`
	PassThreshold    float64   `json:"pass_threshold" gorm:"type:numeric(5,2);default:70"`
	TimeLimitMinutes int       `json:"time_limit_minutes" gorm:"default:90"`
	QuestionCount    int       `json:"question_count" gorm:"default:50"`
	PracticalCount   int       `json:"practical_count" gorm:"default:3"`
	ValidityMonths   int       `json:"validity_months" gorm:"default:24"`
	IsActive         bool      `json:"is_active" gorm:"default:true"`
	IsComingSoon     bool      `json:"is_coming_soon" gorm:"default:false"`
	Metadata         JSONMap   `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CertTier) TableName() string { return "cert_tiers" }

// CertQuestion represents a knowledge question in the bank
type CertQuestion struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TierID         uuid.UUID  `json:"tier_id" gorm:"type:uuid;not null;index"`
	Tier           *CertTier  `json:"tier,omitempty" gorm:"foreignKey:TierID"`
	Category       string     `json:"category" gorm:"size:100;not null"`
	Difficulty     string     `json:"difficulty" gorm:"size:20;default:'medium'"`
	QuestionText   string     `json:"question_text" gorm:"type:text;not null"`
	QuestionFormat string     `json:"question_format" gorm:"size:20;default:'multiple_choice'"`
	Options        interface{} `json:"options" gorm:"type:jsonb;not null"`
	CorrectAnswers JSONMap    `json:"correct_answers,omitempty" gorm:"type:jsonb;not null"`
	Explanation    string     `json:"explanation,omitempty" gorm:"type:text"`
	Points         int        `json:"points" gorm:"default:1"`
	IsActive       bool       `json:"is_active" gorm:"default:true"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty" gorm:"type:uuid"`
	Metadata       JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CertQuestion) TableName() string { return "cert_questions" }

// CertQuestionPublic is the question without answers (sent to exam takers)
type CertQuestionPublic struct {
	ID             uuid.UUID `json:"id"`
	Category       string    `json:"category"`
	Difficulty     string    `json:"difficulty"`
	QuestionText   string    `json:"question_text"`
	QuestionFormat string    `json:"question_format"`
	Options        interface{} `json:"options"`
	Points         int        `json:"points"`
}

// CertPracticalChallenge represents a hands-on grading challenge
type CertPracticalChallenge struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TierID            uuid.UUID  `json:"tier_id" gorm:"type:uuid;not null;index"`
	Tier              *CertTier  `json:"tier,omitempty" gorm:"foreignKey:TierID"`
	Slug              string     `json:"slug" gorm:"uniqueIndex;size:100;not null"`
	Name              string     `json:"name" gorm:"size:255;not null"`
	Description       string     `json:"description" gorm:"type:text;not null"`
	Category          string     `json:"category" gorm:"size:100;not null"`
	Difficulty        string     `json:"difficulty" gorm:"size:20;default:'medium'"`
	Points            int        `json:"points" gorm:"default:10"`
	TimeLimitMinutes  int        `json:"time_limit_minutes" gorm:"default:30"`
	GradingConfig     JSONMap    `json:"grading_config" gorm:"type:jsonb;not null"`
	ValidatorFuncID   *uuid.UUID `json:"validator_function_id,omitempty" gorm:"type:uuid"`
	EnvironmentURL    string     `json:"environment_url" gorm:"size:500"`
	IsActive          bool       `json:"is_active" gorm:"default:true"`
	Metadata          JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CertPracticalChallenge) TableName() string { return "cert_practical_challenges" }

// CertExam represents a single exam session
type CertExam struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID           uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	User             *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	TierID           uuid.UUID  `json:"tier_id" gorm:"type:uuid;not null"`
	Tier             *CertTier  `json:"tier,omitempty" gorm:"foreignKey:TierID"`
	Status           string     `json:"status" gorm:"size:20;default:'in_progress';not null"`
	StripePaymentID  *string    `json:"stripe_payment_id,omitempty" gorm:"size:255"`
	AmountCents      int        `json:"amount_cents" gorm:"default:0"`
	Currency         string     `json:"currency" gorm:"size:3;default:'USD'"`
	StartedAt        time.Time  `json:"started_at" gorm:"not null;default:now()"`
	SubmittedAt      *time.Time `json:"submitted_at,omitempty"`
	GradedAt         *time.Time `json:"graded_at,omitempty"`
	ExpiresAt        time.Time  `json:"expires_at" gorm:"not null"`
	KnowledgeScore   *float64   `json:"knowledge_score,omitempty" gorm:"type:numeric(5,2)"`
	PracticalScore   *float64   `json:"practical_score,omitempty" gorm:"type:numeric(5,2)"`
	TotalScore       *float64   `json:"total_score,omitempty" gorm:"type:numeric(5,2)"`
	Passed           *bool      `json:"passed,omitempty"`
	QuestionIDs      []uuid.UUID `json:"question_ids" gorm:"-"`
	PracticalIDs     []uuid.UUID `json:"practical_ids" gorm:"-"`
	Answers          JSONMap     `json:"answers,omitempty" gorm:"type:jsonb;default:'{}'"`
	PracticalResults JSONMap     `json:"practical_results,omitempty" gorm:"type:jsonb;default:'{}'"`
	IPAddress        *string    `json:"ip_address,omitempty" gorm:"type:inet"`
	UserAgent        *string    `json:"user_agent,omitempty" gorm:"size:500"`
	Metadata         JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CertExam) TableName() string { return "cert_exams" }

// Exam status constants
const (
	CertExamStatusInProgress     = "in_progress"
	CertExamStatusPendingPayment = "pending_payment"
	CertExamStatusSubmitted      = "submitted"
	CertExamStatusGrading        = "grading"
	CertExamStatusPassed         = "passed"
	CertExamStatusFailed         = "failed"
	CertExamStatusExpired        = "expired"
	CertExamStatusAbandoned      = "abandoned"
)

// CertCredential represents an earned certification
type CertCredential struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID           uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	User             *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	TierID           uuid.UUID  `json:"tier_id" gorm:"type:uuid;not null"`
	Tier             *CertTier  `json:"tier,omitempty" gorm:"foreignKey:TierID"`
	ExamID           uuid.UUID  `json:"exam_id" gorm:"type:uuid;not null"`
	Exam             *CertExam  `json:"exam,omitempty" gorm:"foreignKey:ExamID"`
	CredentialNumber string     `json:"credential_number" gorm:"uniqueIndex;size:20;not null"`
	Status           string     `json:"status" gorm:"size:20;default:'active';not null"`
	IssuedAt         time.Time  `json:"issued_at" gorm:"not null;default:now()"`
	ExpiresAt        time.Time  `json:"expires_at" gorm:"not null"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	RevokedReason    *string    `json:"revoked_reason,omitempty" gorm:"type:text"`
	OBACredential    JSONMap    `json:"oba_credential,omitempty" gorm:"type:jsonb"`
	VerificationHash string     `json:"verification_hash" gorm:"size:64;not null"`
	VerificationURL  *string    `json:"verification_url,omitempty" gorm:"size:500"`
	RenewalExamID    *uuid.UUID `json:"renewal_exam_id,omitempty" gorm:"type:uuid"`
	Metadata         JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CertCredential) TableName() string { return "cert_credentials" }

// Credential status constants
const (
	CertCredentialStatusActive   = "active"
	CertCredentialStatusExpired  = "expired"
	CertCredentialStatusRevoked  = "revoked"
	CertCredentialStatusSuspended = "suspended"
)

// CertSubscription represents a credential renewal subscription
type CertSubscription struct {
	ID                   uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID               uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	TierID               uuid.UUID  `json:"tier_id" gorm:"type:uuid;not null"`
	Tier                 *CertTier  `json:"tier,omitempty" gorm:"foreignKey:TierID"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty" gorm:"size:255"`
	Status               string     `json:"status" gorm:"size:20;default:'active';not null"`
	RenewalPriceCents    int        `json:"renewal_price_cents" gorm:"default:4900"`
	Currency             string     `json:"currency" gorm:"size:3;default:'USD'"`
	CurrentPeriodStart   *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd     *time.Time `json:"current_period_end,omitempty"`
	CanceledAt           *time.Time `json:"canceled_at,omitempty"`
	Metadata             JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt            time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CertSubscription) TableName() string { return "cert_subscriptions" }

// CertTeamBadge represents an enterprise team certification
type CertTeamBadge struct {
	ID                   uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID             uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	TierID               uuid.UUID  `json:"tier_id" gorm:"type:uuid;not null"`
	Tier                 *CertTier  `json:"tier,omitempty" gorm:"foreignKey:TierID"`
	BadgeName            string     `json:"badge_name" gorm:"size:255;not null"`
	MinCertified         int        `json:"min_certified" gorm:"default:5"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty" gorm:"size:255"`
	AnnualPriceCents     int        `json:"annual_price_cents" gorm:"default:49900"`
	Currency             string     `json:"currency" gorm:"size:3;default:'USD'"`
	Status               string     `json:"status" gorm:"size:20;default:'active';not null"`
	IssuedAt             time.Time  `json:"issued_at" gorm:"not null;default:now()"`
	ExpiresAt            time.Time  `json:"expires_at" gorm:"not null"`
	VerifiedCount        int        `json:"verified_count" gorm:"default:0"`
	Metadata             JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt            time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CertTeamBadge) TableName() string { return "cert_team_badges" }

// CertGradingQueueItem represents a practical challenge grading task
type CertGradingQueueItem struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExamID       uuid.UUID  `json:"exam_id" gorm:"type:uuid;not null;index"`
	ChallengeID  uuid.UUID  `json:"challenge_id" gorm:"type:uuid;not null"`
	Status       string     `json:"status" gorm:"size:20;default:'pending';not null"`
	Attempts     int        `json:"attempts" gorm:"default:0"`
	MaxAttempts  int        `json:"max_attempts" gorm:"default:3"`
	Result       JSONMap    `json:"result,omitempty" gorm:"type:jsonb"`
	ErrorMessage *string    `json:"error_message,omitempty" gorm:"type:text"`
	LockedAt     *time.Time `json:"locked_at,omitempty"`
	LockedBy     *string    `json:"locked_by,omitempty" gorm:"size:255"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CertGradingQueueItem) TableName() string { return "cert_grading_queue" }

// Grading queue status constants
const (
	CertGradingStatusPending    = "pending"
	CertGradingStatusProcessing = "processing"
	CertGradingStatusCompleted  = "completed"
	CertGradingStatusFailed     = "failed"
)

// GradingConfig defines how a practical challenge is auto-graded.
// Stored as JSONB in cert_practical_challenges.grading_config.
type GradingConfig struct {
	Type           string                 `json:"type"`
	Endpoint       string                 `json:"endpoint"`
	Method         string                 `json:"method"`
	Headers        map[string]string      `json:"headers"`
	Body           string                 `json:"body"`
	ExpectedStatus int                    `json:"expected_status"`
	ExpectedBody   *string                `json:"expected_body"`
	ExpectedJSON   map[string]interface{} `json:"expected_json"`
	StateChecks    []StateCheck           `json:"state_checks"`
	TimeoutSeconds int                    `json:"timeout_seconds"`
}

// StateCheck defines a verification rule for state stores
type StateCheck struct {
	Store    string      `json:"store"`
	Key      string      `json:"key"`
	Value    interface{} `json:"value"`
	Operator string      `json:"operator"` // eq, contains, exists, gt, lt
}

// Difficulty and format constants
const (
	CertDifficultyEasy   = "easy"
	CertDifficultyMedium = "medium"
	CertDifficultyHard   = "hard"

	CertFormatMultipleChoice = "multiple_choice"
	CertFormatTrueFalse      = "true_false"
	CertFormatMultiSelect    = "multi_select"
)
