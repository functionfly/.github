package types

import (
	"time"

	"github.com/google/uuid"
)

// Department represents an organizational unit.
type Department struct {
	ID          int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	Name        string     `json:"name" gorm:"not null"`
	Slug        string     `json:"slug" gorm:"not null"`
	Description *string    `json:"description,omitempty"`
	ParentID    *int64     `json:"parent_id,omitempty"`
	Parent      *Department `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	HeadID      *uuid.UUID `json:"head_id,omitempty"`
	Head        *Employee  `json:"head,omitempty" gorm:"foreignKey:HeadID"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// Employee represents an employee profile extending a user account.
type Employee struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID           uuid.UUID  `json:"user_id" gorm:"type:uuid;uniqueIndex;not null"`
	User             *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	TenantID         uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	EmployeeNumber   string     `json:"employee_number" gorm:"uniqueIndex;not null"`
	FFID             string     `json:"ffid" gorm:"uniqueIndex;not null"`
	DepartmentID     *int64     `json:"department_id,omitempty"`
	Department       *Department `json:"department,omitempty" gorm:"foreignKey:DepartmentID"`
	ManagerID        *uuid.UUID `json:"manager_id,omitempty"`
	Manager          *Employee  `json:"manager,omitempty" gorm:"foreignKey:ManagerID"`
	HireDate         *time.Time `json:"hire_date,omitempty" gorm:"column:hire_date;type:date"`
	EmploymentType   string     `json:"employment_type" gorm:"default:'full_time'"`
	ClearanceLevel   string     `json:"clearance_level" gorm:"default:'standard'"`
	WorkLocation     *string    `json:"work_location,omitempty"`
	OfficeLocation   *string    `json:"office_location,omitempty"`
	Timezone         *string    `json:"timezone,omitempty"`
	Bio              *string    `json:"bio,omitempty"`
	Pronouns         *string    `json:"pronouns,omitempty"`
	EmergencyContact JSONMap    `json:"emergency_contact,omitempty" gorm:"type:jsonb"`
	Status           string     `json:"status" gorm:"default:'active'"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// EmployeeDepartment represents a many-to-many employee-department membership.
type EmployeeDepartment struct {
	ID           int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	EmployeeID   uuid.UUID  `json:"employee_id" gorm:"type:uuid;not null"`
	DepartmentID int64      `json:"department_id" gorm:"not null"`
	RoleInDept   string     `json:"role_in_dept" gorm:"default:'member'"`
	IsPrimary    bool       `json:"is_primary" gorm:"default:true"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// EmployeeSkill represents a skill for an employee.
type EmployeeSkill struct {
	ID          int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	EmployeeID  uuid.UUID  `json:"employee_id" gorm:"type:uuid;not null"`
	SkillName   string     `json:"skill_name" gorm:"not null"`
	Category    *string    `json:"category,omitempty"`
	Proficiency string     `json:"proficiency" gorm:"default:'intermediate'"`
	YearsExp    *float64   `json:"years_exp,omitempty"`
	Endorsements int       `json:"endorsements" gorm:"default:0"`
	Verified    bool       `json:"verified" gorm:"default:false"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// EmployeeCertification represents a professional certification.
type EmployeeCertification struct {
	ID             int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	EmployeeID     uuid.UUID  `json:"employee_id" gorm:"type:uuid;not null"`
	Name           string     `json:"name" gorm:"not null"`
	Issuer         string     `json:"issuer" gorm:"not null"`
	CredentialID   *string    `json:"credential_id,omitempty"`
	CredentialURL  *string    `json:"credential_url,omitempty"`
	IssuedDate     *time.Time `json:"issued_date,omitempty" gorm:"type:date"`
	ExpiryDate     *time.Time `json:"expiry_date,omitempty" gorm:"type:date"`
	Verified       bool       `json:"verified" gorm:"default:false"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// EmployeeAchievement represents recognition or an award.
type EmployeeAchievement struct {
	ID          int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	EmployeeID  uuid.UUID  `json:"employee_id" gorm:"type:uuid;not null"`
	Title       string     `json:"title" gorm:"not null"`
	Description *string    `json:"description,omitempty"`
	Type        string     `json:"type" gorm:"default:'recognition'"`
	AwardedBy   *uuid.UUID `json:"awarded_by,omitempty"`
	Points      int        `json:"points" gorm:"default:0"`
	BadgeURL    *string    `json:"badge_url,omitempty"`
	AwardedAt   time.Time  `json:"awarded_at" gorm:"autoCreateTime"`
}

// Project represents an internal company project.
type Project struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	Name        string     `json:"name" gorm:"not null"`
	Slug        string     `json:"slug" gorm:"not null"`
	Description *string    `json:"description,omitempty"`
	Status      string     `json:"status" gorm:"default:'active'"`
	Priority    string     `json:"priority" gorm:"default:'medium'"`
	OwnerID     uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null"`
	Owner       *Employee  `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	StartDate   *time.Time `json:"start_date,omitempty" gorm:"type:date"`
	TargetDate  *time.Time `json:"target_date,omitempty" gorm:"type:date"`
	Tags        JSONMap    `json:"tags,omitempty" gorm:"type:jsonb;default:'[]'"`
	Metadata    JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// Task represents a project task.
type Task struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProjectID      uuid.UUID  `json:"project_id" gorm:"type:uuid;not null"`
	Project        *Project   `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	TenantID       uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	Title          string     `json:"title" gorm:"not null"`
	Description    *string    `json:"description,omitempty"`
	Status         string     `json:"status" gorm:"default:'todo'"`
	Priority       string     `json:"priority" gorm:"default:'medium'"`
	AssigneeID     *uuid.UUID `json:"assignee_id,omitempty"`
	Assignee       *Employee  `json:"assignee,omitempty" gorm:"foreignKey:AssigneeID"`
	ReporterID     uuid.UUID  `json:"reporter_id" gorm:"type:uuid;not null"`
	Reporter       *Employee  `json:"reporter,omitempty" gorm:"foreignKey:ReporterID"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	DueDate        *time.Time `json:"due_date,omitempty" gorm:"type:date"`
	EstimatedHours *float64   `json:"estimated_hours,omitempty"`
	ActualHours    *float64   `json:"actual_hours,omitempty"`
	Tags           JSONMap    `json:"tags,omitempty" gorm:"type:jsonb;default:'[]'"`
	Position       int        `json:"position" gorm:"default:0"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TaskComment represents a comment on a task.
type TaskComment struct {
	ID        int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID    uuid.UUID  `json:"task_id" gorm:"type:uuid;not null"`
	AuthorID  uuid.UUID  `json:"author_id" gorm:"type:uuid;not null"`
	Author    *Employee  `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	Body      string     `json:"body" gorm:"not null"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// LearningCourse represents a course in the learning catalog.
type LearningCourse struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	Title        string     `json:"title" gorm:"not null"`
	Description  *string    `json:"description,omitempty"`
	Category     *string    `json:"category,omitempty"`
	Difficulty   *string    `json:"difficulty,omitempty"`
	DurationMin  *int       `json:"duration_min,omitempty"`
	ContentURL   *string    `json:"content_url,omitempty"`
	ThumbnailURL *string    `json:"thumbnail_url,omitempty"`
	IsMandatory  bool       `json:"is_mandatory" gorm:"default:false"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// EmployeeLearning represents an employee's enrollment and progress in a course.
type EmployeeLearning struct {
	ID          int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	EmployeeID  uuid.UUID  `json:"employee_id" gorm:"type:uuid;not null"`
	CourseID    uuid.UUID  `json:"course_id" gorm:"type:uuid;not null"`
	Course      *LearningCourse `json:"course,omitempty" gorm:"foreignKey:CourseID"`
	Status      string     `json:"status" gorm:"default:'not_started'"`
	ProgressPct int        `json:"progress_pct" gorm:"default:0"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Score       *float64   `json:"score,omitempty"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// KnowledgeArticle represents an article in the company knowledge base.
type KnowledgeArticle struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	Title       string     `json:"title" gorm:"not null"`
	Slug        string     `json:"slug" gorm:"not null"`
	Body        string     `json:"body" gorm:"not null"`
	Category    *string    `json:"category,omitempty"`
	Tags        JSONMap    `json:"tags,omitempty" gorm:"type:jsonb;default:'[]'"`
	AuthorID    uuid.UUID  `json:"author_id" gorm:"type:uuid;not null"`
	Author      *Employee  `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	Status      string     `json:"status" gorm:"default:'draft'"`
	ViewCount   int        `json:"view_count" gorm:"default:0"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// CompensationRecord represents an employee's salary information.
type CompensationRecord struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID      uuid.UUID  `json:"employee_id" gorm:"type:uuid;not null"`
	TenantID        uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	BaseSalaryCents int64      `json:"base_salary_cents" gorm:"not null"`
	Currency        string     `json:"currency" gorm:"default:'USD'"`
	PayFrequency    string     `json:"pay_frequency" gorm:"default:'biweekly'"`
	EffectiveDate   time.Time  `json:"effective_date" gorm:"type:date;not null"`
	EndDate         *time.Time `json:"end_date,omitempty" gorm:"type:date"`
	ReviewDate      *time.Time `json:"review_date,omitempty" gorm:"type:date"`
	Notes           *string    `json:"notes,omitempty"`
	CreatedBy       uuid.UUID  `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// EquityGrant represents a stock/RSU grant with vesting schedule.
type EquityGrant struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID        uuid.UUID  `json:"employee_id" gorm:"type:uuid;not null"`
	TenantID          uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	GrantType         string     `json:"grant_type" gorm:"not null"`
	Shares            int        `json:"shares" gorm:"not null"`
	StrikePriceCents  *int64     `json:"strike_price_cents,omitempty"`
	VestingStart      time.Time  `json:"vesting_start" gorm:"type:date;not null"`
	VestingEnd        time.Time  `json:"vesting_end" gorm:"type:date;not null"`
	CliffDate         *time.Time `json:"cliff_date,omitempty" gorm:"type:date"`
	VestedShares      int        `json:"vested_shares" gorm:"default:0"`
	Status            string     `json:"status" gorm:"default:'active'"`
	GrantDate         time.Time  `json:"grant_date" gorm:"type:date;not null"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// CompensationAccessLog records every access to compensation data.
type CompensationAccessLog struct {
	ID          int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	AccessorID  uuid.UUID  `json:"accessor_id" gorm:"type:uuid;not null"`
	TargetID    uuid.UUID  `json:"target_id" gorm:"type:uuid;not null"`
	Action      string     `json:"action" gorm:"not null"`
	IPAddress   *string    `json:"ip_address,omitempty"`
	UserAgent   *string    `json:"user_agent,omitempty"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// FWOSNotification represents an employee notification.
type FWOSNotification struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	TenantID  uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	Type      string     `json:"type" gorm:"not null"`
	Title     string     `json:"title" gorm:"not null"`
	Body      *string    `json:"body,omitempty"`
	ActionURL *string    `json:"action_url,omitempty"`
	IsRead    bool       `json:"is_read" gorm:"default:false"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// ListEmployeesOpts contains options for listing employees.
type ListEmployeesOpts struct {
	DepartmentID *int64
	Status       *string
	Search       *string
	Limit        int
	Offset       int
}

// ListProjectsOpts contains options for listing projects.
type ListProjectsOpts struct {
	Status *string
	OwnerID *uuid.UUID
	Search  *string
	Limit   int
	Offset  int
}

// ListTasksOpts contains options for listing tasks.
type ListTasksOpts struct {
	ProjectID  *uuid.UUID
	AssigneeID *uuid.UUID
	Status     *string
	Priority   *string
	Search     *string
	Limit      int
	Offset     int
}

// ListKnowledgeOpts contains options for listing knowledge articles.
type ListKnowledgeOpts struct {
	Category *string
	Status   *string
	Search   *string
	Limit    int
	Offset   int
}

// AIChatSession represents an AI chat conversation.
type AIChatSession struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	Title           string     `json:"title"`
	ContextType     string     `json:"context_type"`
	ContextReference *uuid.UUID `json:"context_reference,omitempty"`
	IsActive        bool       `json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// AIChatMessage represents a single message in an AI chat session.
type AIChatMessage struct {
	ID         uuid.UUID       `json:"id"`
	SessionID  uuid.UUID       `json:"session_id"`
	Role       string          `json:"role"`
	Content    string          `json:"content"`
	TokensUsed int             `json:"tokens_used"`
	Model      *string         `json:"model,omitempty"`
	Metadata   JSONMap         `json:"metadata,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// PerformanceGoal represents an employee performance goal.
type PerformanceGoal struct {
	ID          uuid.UUID  `json:"id"`
	EmployeeID  uuid.UUID  `json:"employee_id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Category    string     `json:"category"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	TargetDate  *time.Time `json:"target_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ProgressPct int        `json:"progress_pct"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PerformanceReview represents a performance review.
type PerformanceReview struct {
	ID                  uuid.UUID  `json:"id"`
	EmployeeID          uuid.UUID  `json:"employee_id"`
	ReviewerID          uuid.UUID  `json:"reviewer_id"`
	TenantID            uuid.UUID  `json:"tenant_id"`
	ReviewPeriod        string     `json:"review_period"`
	ReviewType          string     `json:"review_type"`
	Status              string     `json:"status"`
	Strengths           *string    `json:"strengths,omitempty"`
	AreasForImprovement *string    `json:"areas_for_improvement,omitempty"`
	OverallRating       *int       `json:"overall_rating,omitempty"`
	Comments            *string    `json:"comments,omitempty"`
	SubmittedAt         *time.Time `json:"submitted_at,omitempty"`
	AcknowledgedAt      *time.Time `json:"acknowledged_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// PeerFeedback represents peer feedback for a review cycle.
type PeerFeedback struct {
	ID             uuid.UUID  `json:"id"`
	ReviewID       *uuid.UUID `json:"review_id,omitempty"`
	FromEmployeeID uuid.UUID  `json:"from_employee_id"`
	ToEmployeeID   uuid.UUID  `json:"to_employee_id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	FeedbackText   string     `json:"feedback_text"`
	Rating         *int       `json:"rating,omitempty"`
	IsAnonymous    bool       `json:"is_anonymous"`
	CreatedAt      time.Time  `json:"created_at"`
}

// TimeEntry represents a logged time entry.
type TimeEntry struct {
	ID          uuid.UUID  `json:"id"`
	EmployeeID  uuid.UUID  `json:"employee_id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	TaskID      *uuid.UUID `json:"task_id,omitempty"`
	Date        time.Time  `json:"date"`
	Hours       float64    `json:"hours"`
	Description *string    `json:"description,omitempty"`
	EntryType   string     `json:"entry_type"`
	IsBillable  bool       `json:"is_billable"`
	ApprovedBy  *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PTORequest represents a paid time off request.
type PTORequest struct {
	ID         uuid.UUID  `json:"id"`
	EmployeeID uuid.UUID  `json:"employee_id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	PTOType    string     `json:"pto_type"`
	StartDate  time.Time  `json:"start_date"`
	EndDate    time.Time  `json:"end_date"`
	Days       float64    `json:"days"`
	Reason     *string    `json:"reason,omitempty"`
	Status     string     `json:"status"`
	ApprovedBy *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedAt *time.Time `json:"approved_at,omitempty"`
	Notes      *string    `json:"notes,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ListAIChatSessionsOpts contains options for listing chat sessions.
type ListAIChatSessionsOpts struct {
	ContextType *string
	Limit       int
	Offset      int
}

// ListPerformanceGoalsOpts contains options for listing performance goals.
type ListPerformanceGoalsOpts struct {
	Status   *string
	Category *string
	Priority *string
	Limit    int
	Offset   int
}

// ListPerformanceReviewsOpts contains options for listing performance reviews.
type ListPerformanceReviewsOpts struct {
	ReviewType *string
	Status     *string
	Period     *string
	Limit      int
	Offset     int
}

// ListTimeEntriesOpts contains options for listing time entries.
type ListTimeEntriesOpts struct {
	ProjectID *uuid.UUID
	EntryType *string
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
}

// ListPTORequestsOpts contains options for listing PTO requests.
type ListPTORequestsOpts struct {
	PTOType *string
	Status  *string
	Limit   int
	Offset  int
}

// ── FWOS Phase 3 Types ─────────────────────────────────────────────────────

// InnovationGrant represents an innovation grant proposal.
type InnovationGrant struct {
	ID                   uuid.UUID  `json:"id"`
	TenantID             uuid.UUID  `json:"tenant_id"`
	ProposerID           uuid.UUID  `json:"proposer_id"`
	Title                string     `json:"title"`
	Description          string     `json:"description"`
	Category             string     `json:"category"`
	RequestedAmountCents *int64     `json:"requested_amount_cents,omitempty"`
	Status               string     `json:"status"`
	FeasibilityScore     *float64   `json:"feasibility_score,omitempty"`
	VotesFor             int        `json:"votes_for"`
	VotesAgainst         int        `json:"votes_against"`
	ReviewedBy           *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt           *time.Time `json:"reviewed_at,omitempty"`
	RejectionReason      *string    `json:"rejection_reason,omitempty"`
	FundedAt             *time.Time `json:"funded_at,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// InnovationGrantVote represents a vote on an innovation grant.
type InnovationGrantVote struct {
	ID        int64      `json:"id"`
	GrantID   uuid.UUID  `json:"grant_id"`
	VoterID   uuid.UUID  `json:"voter_id"`
	Vote      bool       `json:"vote"`
	Comment   *string    `json:"comment,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// MarketplaceOpportunity represents a talent marketplace opportunity.
type MarketplaceOpportunity struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	PostedBy        uuid.UUID  `json:"posted_by"`
	DepartmentID    *int64     `json:"department_id,omitempty"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	OpportunityType string     `json:"opportunity_type"`
	SkillsRequired  JSONMap    `json:"skills_required"`
	HoursPerWeek    *float64   `json:"hours_per_week,omitempty"`
	DurationWeeks   *int       `json:"duration_weeks,omitempty"`
	IsRemote        bool       `json:"is_remote"`
	Status          string     `json:"status"`
	MaxApplicants   *int       `json:"max_applicants,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// MarketplaceApplication represents an application to a marketplace opportunity.
type MarketplaceApplication struct {
	ID            uuid.UUID  `json:"id"`
	OpportunityID uuid.UUID  `json:"opportunity_id"`
	ApplicantID   uuid.UUID  `json:"applicant_id"`
	Message       *string    `json:"message,omitempty"`
	Status        string     `json:"status"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// CareerPath represents a career path definition.
type CareerPath struct {
	ID                  uuid.UUID  `json:"id"`
	TenantID            uuid.UUID  `json:"tenant_id"`
	Title               string     `json:"title"`
	Track               string     `json:"track"`
	Level               int        `json:"level"`
	Description         *string    `json:"description,omitempty"`
	Requirements        JSONMap    `json:"requirements"`
	SalaryRangeMinCents *int64     `json:"salary_range_min_cents,omitempty"`
	SalaryRangeMaxCents *int64     `json:"salary_range_max_cents,omitempty"`
	NextPathID          *uuid.UUID `json:"next_path_id,omitempty"`
	IsActive            bool       `json:"is_active"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// EmployeeCareerProgress represents an employee's progress on a career path.
type EmployeeCareerProgress struct {
	ID          int64      `json:"id"`
	EmployeeID  uuid.UUID  `json:"employee_id"`
	CareerPathID uuid.UUID `json:"career_path_id"`
	Status      string     `json:"status"`
	GapAnalysis JSONMap    `json:"gap_analysis"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	TargetDate  *time.Time `json:"target_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// MentorshipMatch represents a mentorship pairing.
type MentorshipMatch struct {
	ID               uuid.UUID  `json:"id"`
	TenantID         uuid.UUID  `json:"tenant_id"`
	MentorID         uuid.UUID  `json:"mentor_id"`
	MenteeID         uuid.UUID  `json:"mentee_id"`
	FocusArea        *string    `json:"focus_area,omitempty"`
	Status           string     `json:"status"`
	StartedAt        time.Time  `json:"started_at"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
	MeetingFrequency *string    `json:"meeting_frequency,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Document represents a company document.
type Document struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	AuthorID   uuid.UUID  `json:"author_id"`
	Title      string     `json:"title"`
	Body       *string    `json:"body,omitempty"`
	DocType    string     `json:"doc_type"`
	Category   *string    `json:"category,omitempty"`
	Tags       JSONMap    `json:"tags"`
	IsTemplate bool       `json:"is_template"`
	Status     string     `json:"status"`
	ViewCount  int        `json:"view_count"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// DocumentShare represents a document sharing record.
type DocumentShare struct {
	ID         int64      `json:"id"`
	DocumentID uuid.UUID  `json:"document_id"`
	SharedWith uuid.UUID  `json:"shared_with"`
	Permission string     `json:"permission"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ListInnovationGrantsOpts contains options for listing innovation grants.
type ListInnovationGrantsOpts struct {
	Status *string
	Category *string
	Limit  int
	Offset int
}

// ListMarketplaceOpportunitiesOpts contains options for listing marketplace opportunities.
type ListMarketplaceOpportunitiesOpts struct {
	Type   *string
	Status *string
	Limit  int
	Offset int
}

// ListCareerPathsOpts contains options for listing career paths.
type ListCareerPathsOpts struct {
	Track *string
	Level *int
	Limit  int
	Offset int
}

// ListMentorshipMatchesOpts contains options for listing mentorship matches.
type ListMentorshipMatchesOpts struct {
	Status *string
	Limit  int
	Offset int
}

// ListDocumentsOpts contains options for listing documents.
type ListDocumentsOpts struct {
	DocType *string
	Status  *string
	Search  *string
	Limit   int
	Offset  int
}

// ── FWOS Phase 4 Types ─────────────────────────────────────────────────────

// TeamHealthMetric represents aggregated team health data.
type TeamHealthMetric struct {
	ID                    uuid.UUID  `json:"id"`
	TenantID              uuid.UUID  `json:"tenant_id"`
	DepartmentID          *int64     `json:"department_id,omitempty"`
	TeamID                *uuid.UUID `json:"team_id,omitempty"`
	MetricDate            time.Time  `json:"metric_date"`
	WorkloadScore         *float64   `json:"workload_score,omitempty"`
	BurnoutRisk           *float64   `json:"burnout_risk,omitempty"`
	VelocityScore         *float64   `json:"velocity_score,omitempty"`
	CollaborationScore    *float64   `json:"collaboration_score,omitempty"`
	KnowledgeSharingScore *float64   `json:"knowledge_sharing_score,omitempty"`
	PTOUtilizationPct     *float64   `json:"pto_utilization_pct,omitempty"`
	AvgOvertimeHours      *float64   `json:"avg_overtime_hours,omitempty"`
	Headcount             *int       `json:"headcount,omitempty"`
	Metadata              JSONMap    `json:"metadata,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
}

// SkillsGraph represents an aggregated skill in the skills graph.
type SkillsGraph struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	SkillName      string     `json:"skill_name"`
	Category       *string    `json:"category,omitempty"`
	TotalEmployees int        `json:"total_employees"`
	AvgProficiency *float64   `json:"avg_proficiency,omitempty"`
	DemandScore    *float64   `json:"demand_score,omitempty"`
	SupplyScore    *float64   `json:"supply_score,omitempty"`
	GapScore       *float64   `json:"gap_score,omitempty"`
	Trending       bool       `json:"trending"`
	LastCalculated *time.Time `json:"last_calculated,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ReputationScore represents an employee's reputation in a category.
type ReputationScore struct {
	ID             uuid.UUID  `json:"id"`
	EmployeeID     uuid.UUID  `json:"employee_id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	Category       string     `json:"category"`
	Score          float64    `json:"score"`
	Rank           *int       `json:"rank,omitempty"`
	Percentile     *float64   `json:"percentile,omitempty"`
	Components     JSONMap    `json:"components,omitempty"`
	LastCalculated *time.Time `json:"last_calculated,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// DigitalBadge represents a badge definition.
type DigitalBadge struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Slug        string     `json:"slug"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	IconURL     *string    `json:"icon_url,omitempty"`
	Category    string     `json:"category"`
	Criteria    JSONMap    `json:"criteria,omitempty"`
	Points      int        `json:"points"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

// EmployeeBadge represents a badge awarded to an employee.
type EmployeeBadge struct {
	ID         int64      `json:"id"`
	EmployeeID uuid.UUID  `json:"employee_id"`
	BadgeID    uuid.UUID  `json:"badge_id"`
	AwardedBy   *uuid.UUID `json:"awarded_by,omitempty"`
	AwardedAt  time.Time  `json:"awarded_at"`
}

// LivingMemoryEntry represents a living memory entry.
type LivingMemoryEntry struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	AuthorID       uuid.UUID  `json:"author_id"`
	Title          string     `json:"title"`
	Body           string     `json:"body"`
	MemoryType     string     `json:"memory_type"`
	ProjectID      *uuid.UUID `json:"project_id,omitempty"`
	Tags           JSONMap    `json:"tags,omitempty"`
	Participants   JSONMap    `json:"participants,omitempty"`
	Importance     string     `json:"importance"`
	SearchableText *string    `json:"searchable_text,omitempty"`
	ViewCount      int        `json:"view_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// MissionControlSnapshot represents a cached executive metrics snapshot.
type MissionControlSnapshot struct {
	ID                       uuid.UUID  `json:"id"`
	TenantID                 uuid.UUID  `json:"tenant_id"`
	SnapshotDate             time.Time  `json:"snapshot_date"`
	TotalEmployees           *int       `json:"total_employees,omitempty"`
	ActiveEmployees          *int       `json:"active_employees,omitempty"`
	NewHires30d              *int       `json:"new_hires_30d,omitempty"`
	Departures30d            *int       `json:"departures_30d,omitempty"`
	TotalProjects            *int       `json:"total_projects,omitempty"`
	ActiveProjects           *int       `json:"active_projects,omitempty"`
	CompletedProjects30d     *int       `json:"completed_projects_30d,omitempty"`
	TotalTasks               *int       `json:"total_tasks,omitempty"`
	CompletedTasks30d        *int       `json:"completed_tasks_30d,omitempty"`
	AvgTaskCompletionDays    *float64   `json:"avg_task_completion_days,omitempty"`
	TotalLearningHours       *float64   `json:"total_learning_hours,omitempty"`
	AvgSkillProficiency      *float64   `json:"avg_skill_proficiency,omitempty"`
	InnovationGrantsSubmitted *int      `json:"innovation_grants_submitted,omitempty"`
	InnovationGrantsFunded   *int       `json:"innovation_grants_funded,omitempty"`
	PTODaysUsed30d           *float64   `json:"pto_days_used_30d,omitempty"`
	AvgBurnoutRisk           *float64   `json:"avg_burnout_risk,omitempty"`
	Metadata                 JSONMap    `json:"metadata,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
}

// ListTeamHealthOpts contains options for listing team health metrics.
type ListTeamHealthOpts struct {
	DepartmentID *int64
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}

// ListReputationOpts contains options for listing reputation scores.
type ListReputationOpts struct {
	Category *string
	Limit    int
	Offset   int
}

// ListBadgesOpts contains options for listing badges.
type ListBadgesOpts struct {
	Category *string
	IsActive *bool
	Limit    int
	Offset   int
}

// ListLivingMemoryOpts contains options for listing living memory entries.
type ListLivingMemoryOpts struct {
	MemoryType *string
	ProjectID  *uuid.UUID
	Importance *string
	Limit      int
	Offset     int
}

// SearchLivingMemoryOpts contains options for full-text search of living memory.
type SearchLivingMemoryOpts struct {
	Query      string
	MemoryType *string
	ProjectID  *uuid.UUID
	Limit      int
}

// ── FWOS Phase 5 Types ─────────────────────────────────────────────────────

// FWOSIncident represents a production incident.
type FWOSIncident struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	Title          string     `json:"title"`
	Description    *string    `json:"description,omitempty"`
	Severity       string     `json:"severity"`
	Status         string     `json:"status"`
	CommanderID    *uuid.UUID `json:"commander_id,omitempty"`
	ProjectID      *uuid.UUID `json:"project_id,omitempty"`
	DetectedAt     time.Time  `json:"detected_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
	RootCause      *string    `json:"root_cause,omitempty"`
	Impact         *string    `json:"impact,omitempty"`
	DurationMinutes *int      `json:"duration_minutes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// IncidentEvent represents a timeline event in an incident.
type IncidentEvent struct {
	ID         int64      `json:"id"`
	IncidentID uuid.UUID  `json:"incident_id"`
	AuthorID   uuid.UUID  `json:"author_id"`
	EventType  string     `json:"event_type"`
	Body       string     `json:"body"`
	Metadata   JSONMap    `json:"metadata,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// IncidentResponder represents a responder assigned to an incident.
type IncidentResponder struct {
	ID         int64      `json:"id"`
	IncidentID uuid.UUID  `json:"incident_id"`
	EmployeeID uuid.UUID  `json:"employee_id"`
	Role       string     `json:"role"`
	JoinedAt   time.Time  `json:"joined_at"`
	LeftAt     *time.Time `json:"left_at,omitempty"`
}

// Postmortem represents a post-incident review.
type Postmortem struct {
	ID                 uuid.UUID  `json:"id"`
	IncidentID         uuid.UUID  `json:"incident_id"`
	TenantID           uuid.UUID  `json:"tenant_id"`
	AuthorID           uuid.UUID  `json:"author_id"`
	Summary            string     `json:"summary"`
	RootCause          string     `json:"root_cause"`
	ContributingFactors *string   `json:"contributing_factors,omitempty"`
	WhatWentWell       *string    `json:"what_went_well,omitempty"`
	WhatWentWrong      *string    `json:"what_went_wrong,omitempty"`
	ActionItems        JSONMap    `json:"action_items,omitempty"`
	LessonsLearned     *string    `json:"lessons_learned,omitempty"`
	Status             string     `json:"status"`
	PublishedAt        *time.Time `json:"published_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// LifecycleEvent represents an employee lifecycle event.
type LifecycleEvent struct {
	ID          uuid.UUID  `json:"id"`
	EmployeeID  uuid.UUID  `json:"employee_id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	EventType   string     `json:"event_type"`
	Payload     JSONMap    `json:"payload,omitempty"`
	TriggeredBy *uuid.UUID `json:"triggered_by,omitempty"`
	Notes       *string    `json:"notes,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// LifecycleWorkflow represents a lifecycle workflow template.
type LifecycleWorkflow struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	Name         string     `json:"name"`
	Description  *string    `json:"description,omitempty"`
	TriggerEvent string     `json:"trigger_event"`
	Steps        JSONMap    `json:"steps"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// LifecycleWorkflowInstance represents a running instance of a lifecycle workflow.
type LifecycleWorkflowInstance struct {
	ID          uuid.UUID  `json:"id"`
	WorkflowID  uuid.UUID  `json:"workflow_id"`
	EmployeeID  uuid.UUID  `json:"employee_id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Status      string     `json:"status"`
	CurrentStep int        `json:"current_step"`
	StepsStatus JSONMap    `json:"steps_status,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// FeatureFlag represents a feature flag configuration.
type FeatureFlag struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	Key            string     `json:"key"`
	Name           string     `json:"name"`
	Description    *string    `json:"description,omitempty"`
	FlagType       string     `json:"flag_type"`
	IsEnabled      bool       `json:"is_enabled"`
	RolloutPct     int        `json:"rollout_pct"`
	Variants       JSONMap    `json:"variants,omitempty"`
	TargetAudience JSONMap    `json:"target_audience,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// DataClassification represents a data classification label on a resource.
type DataClassification struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	ResourceType   string     `json:"resource_type"`
	ResourceID     uuid.UUID  `json:"resource_id"`
	Classification string     `json:"classification"`
	ClassifiedBy   *uuid.UUID `json:"classified_by,omitempty"`
	Reason         *string    `json:"reason,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// EmployeeCertificate represents an employee certificate (FF-CERT).
type EmployeeCertificate struct {
	ID              uuid.UUID  `json:"id"`
	EmployeeID      uuid.UUID  `json:"employee_id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	CertificateSerial string   `json:"certificate_serial"`
	CertificateType string     `json:"certificate_type"`
	Subject         string     `json:"subject"`
	Issuer          string     `json:"issuer"`
	PublicKeyPEM    string     `json:"public_key_pem"`
	Fingerprint     string     `json:"fingerprint"`
	DeviceID        *string    `json:"device_id,omitempty"`
	DeviceName      *string    `json:"device_name,omitempty"`
	IssuedAt        time.Time  `json:"issued_at"`
	ExpiresAt       time.Time  `json:"expires_at"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	RevokeReason    *string    `json:"revoke_reason,omitempty"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
}

// FWOSEvent represents an event in the FWOS event log.
type FWOSEvent struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	EventType    string     `json:"event_type"`
	Source       string     `json:"source"`
	ActorID      *uuid.UUID `json:"actor_id,omitempty"`
	ResourceType *string    `json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	Payload      JSONMap    `json:"payload,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ListIncidentsOpts contains options for listing incidents.
type ListIncidentsOpts struct {
	Status   *string
	Severity *string
	Limit    int
	Offset   int
}

// ListLifecycleEventsOpts contains options for listing lifecycle events.
type ListLifecycleEventsOpts struct {
	EventType *string
	Limit     int
	Offset    int
}

// ListFeatureFlagsOpts contains options for listing feature flags.
type ListFeatureFlagsOpts struct {
	IsEnabled *bool
	Limit     int
	Offset    int
}

// ListDataClassificationsOpts contains options for listing data classifications.
type ListDataClassificationsOpts struct {
	ResourceType   *string
	Classification *string
	Limit          int
	Offset         int
}

// ListCertificatesOpts contains options for listing employee certificates.
type ListCertificatesOpts struct {
	Status *string
	Limit  int
	Offset int
}

// ListFWOSEventsOpts contains options for listing FWOS events.
type ListFWOSEventsOpts struct {
	EventType    *string
	Source       *string
	ResourceType *string
	Limit        int
	Offset       int
}

// ── FWOS Phase 6 Types ─────────────────────────────────────────────────────

// EmailAccount represents a provisioned @functionfly.com email address.
type EmailAccount struct {
	ID                 uuid.UUID  `json:"id"`
	EmployeeID         uuid.UUID  `json:"employee_id"`
	TenantID           uuid.UUID  `json:"tenant_id"`
	Email              string     `json:"email"`
	DisplayName        *string    `json:"display_name,omitempty"`
	Provider           string     `json:"provider"`
	ProviderAccountID  *string    `json:"provider_account_id,omitempty"`
	Aliases            []string   `json:"aliases,omitempty"`
	Groups             []string   `json:"groups,omitempty"`
	Status             string     `json:"status"`
	ProvisionedAt      *time.Time `json:"provisioned_at,omitempty"`
	LastSyncAt         *time.Time `json:"last_sync_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// Device represents a registered device.
type Device struct {
	ID               uuid.UUID  `json:"id"`
	EmployeeID       *uuid.UUID `json:"employee_id,omitempty"`
	TenantID         uuid.UUID  `json:"tenant_id"`
	DeviceName       string     `json:"device_name"`
	DeviceType       string     `json:"device_type"`
	SerialNumber     *string    `json:"serial_number,omitempty"`
	OS               *string    `json:"os,omitempty"`
	OSVersion        *string    `json:"os_version,omitempty"`
	Manufacturer     *string    `json:"manufacturer,omitempty"`
	Model            *string    `json:"model,omitempty"`
	LastSeenAt       *time.Time `json:"last_seen_at,omitempty"`
	ComplianceStatus string     `json:"compliance_status"`
	CertificateID    *uuid.UUID `json:"certificate_id,omitempty"`
	EnrolledAt       *time.Time `json:"enrolled_at,omitempty"`
	Metadata         JSONMap    `json:"metadata,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// SSOProvisioningConfig represents an SSO provisioning configuration.
type SSOProvisioningConfig struct {
	ID                   uuid.UUID  `json:"id"`
	TenantID             uuid.UUID  `json:"tenant_id"`
	Provider             string     `json:"provider"`
	ProviderURL          *string    `json:"provider_url,omitempty"`
	ClientID             *string    `json:"client_id,omitempty"`
	ClientSecretEncrypted *string   `json:"client_secret_encrypted,omitempty"`
	SCIMEndpoint         *string    `json:"scim_endpoint,omitempty"`
	SCIMTokenEncrypted   *string    `json:"scim_token_encrypted,omitempty"`
	AutoCreateEmployee   bool       `json:"auto_create_employee"`
	AutoUpdateEmployee   bool       `json:"auto_update_employee"`
	AutoDeactivate       bool       `json:"auto_deactivate"`
	DefaultDepartmentID  *int64     `json:"default_department_id,omitempty"`
	DefaultClearance     string     `json:"default_clearance"`
	FieldMappings        JSONMap    `json:"field_mappings,omitempty"`
	IsActive             bool       `json:"is_active"`
	LastSyncAt           *time.Time `json:"last_sync_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// SSOProvisioningLog represents a log entry for SSO provisioning actions.
type SSOProvisioningLog struct {
	ID             int64      `json:"id"`
	ConfigID       uuid.UUID  `json:"config_id"`
	ExternalUserID *string    `json:"external_user_id,omitempty"`
	EmployeeID     *uuid.UUID `json:"employee_id,omitempty"`
	Action         string     `json:"action"`
	Details        JSONMap    `json:"details,omitempty"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// WalletPass represents a digital wallet pass (Apple/Google Wallet).
type WalletPass struct {
	ID               uuid.UUID  `json:"id"`
	EmployeeID       uuid.UUID  `json:"employee_id"`
	TenantID         uuid.UUID  `json:"tenant_id"`
	PassType         string     `json:"pass_type"`
	Platform         string     `json:"platform"`
	PassID           string     `json:"pass_id"`
	QRToken          string     `json:"qr_token"`
	QRExpiresAt      time.Time  `json:"qr_expires_at"`
	DeviceID         *string    `json:"device_id,omitempty"`
	InstalledAt      *time.Time `json:"installed_at,omitempty"`
	LastPresentedAt  *time.Time `json:"last_presented_at,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// PushSubscription represents a Web Push notification subscription.
type PushSubscription struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	Endpoint   string     `json:"endpoint"`
	P256DH     string     `json:"p256dh"`
	Auth       string     `json:"auth"`
	UserAgent  *string    `json:"user_agent,omitempty"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// NotificationPreference represents a user's notification preference for a channel+event pair.
type NotificationPreference struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	Channel         string     `json:"channel"`
	EventType       string     `json:"event_type"`
	IsEnabled       bool       `json:"is_enabled"`
	QuietHoursStart *string    `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd   *string    `json:"quiet_hours_end,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ListEmailAccountsOpts contains options for listing email accounts.
type ListEmailAccountsOpts struct {
	EmployeeID *uuid.UUID
	Status     *string
	Limit      int
	Offset     int
}

// ListDevicesOpts contains options for listing devices.
type ListDevicesOpts struct {
	EmployeeID       *uuid.UUID
	DeviceType       *string
	ComplianceStatus *string
	Status           *string
	Limit            int
	Offset           int
}

// ListSSOProvisioningConfigsOpts contains options for listing SSO configs.
type ListSSOProvisioningConfigsOpts struct {
	Provider *string
	IsActive *bool
	Limit    int
	Offset   int
}

// ListSSOProvisioningLogsOpts contains options for listing SSO provisioning logs.
type ListSSOProvisioningLogsOpts struct {
	Action *string
	Limit  int
	Offset int
}

// ListWalletPassesOpts contains options for listing wallet passes.
type ListWalletPassesOpts struct {
	Status *string
	Limit  int
	Offset int
}

// ListNotificationPreferencesOpts contains options for listing notification preferences.
type ListNotificationPreferencesOpts struct {
	Channel *string
	Limit   int
	Offset  int
}

// ── FWOS Remaining Types ─────────────────────────────────────────────────────

// FeedbackRound represents a 360 feedback round.
type FeedbackRound struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	Name         string     `json:"name"`
	Description  *string    `json:"description,omitempty"`
	ReviewPeriod string     `json:"review_period"`
	RoundType    string     `json:"round_type"`
	Status       string     `json:"status"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      time.Time  `json:"end_date"`
	Questions    JSONMap    `json:"questions,omitempty"`
	CreatedBy    uuid.UUID  `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// FeedbackRoundAssignment represents a reviewer-reviewee pair in a feedback round.
type FeedbackRoundAssignment struct {
	ID          int64      `json:"id"`
	RoundID     uuid.UUID  `json:"round_id"`
	ReviewerID  uuid.UUID  `json:"reviewer_id"`
	RevieweeID  uuid.UUID  `json:"reviewee_id"`
	Status      string     `json:"status"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// FeedbackRoundResponse represents a response to a feedback question.
type FeedbackRoundResponse struct {
	ID             int64      `json:"id"`
	AssignmentID   int64      `json:"assignment_id"`
	QuestionIndex  int        `json:"question_index"`
	ResponseText   *string    `json:"response_text,omitempty"`
	ResponseRating *int       `json:"response_rating,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// DocumentSignature represents a signature request on a document.
type DocumentSignature struct {
	ID            uuid.UUID  `json:"id"`
	DocumentID    uuid.UUID  `json:"document_id"`
	SignerID      uuid.UUID  `json:"signer_id"`
	SignerName    string     `json:"signer_name"`
	SignerEmail   *string    `json:"signer_email,omitempty"`
	SignatureData *string    `json:"signature_data,omitempty"`
	SignedAt      *time.Time `json:"signed_at,omitempty"`
	Status        string     `json:"status"`
	DeclineReason *string    `json:"decline_reason,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// CertificateKey represents a PKI key pair for an employee certificate.
type CertificateKey struct {
	ID             uuid.UUID  `json:"id"`
	CertificateID  uuid.UUID  `json:"certificate_id"`
	PrivateKeyPEM  *string    `json:"private_key_pem,omitempty"`
	PublicKeyPEM   string     `json:"public_key_pem"`
	KeyType        string     `json:"key_type"`
	KeySize        int        `json:"key_size"`
	CreatedAt      time.Time  `json:"created_at"`
}

// WalletPassTemplate represents a wallet pass template.
type WalletPassTemplate struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	Name         string     `json:"name"`
	PassType     string     `json:"pass_type"`
	Platform     string     `json:"platform"`
	TemplateData JSONMap    `json:"template_data"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// OrgChartImport represents an org chart import job.
type OrgChartImport struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	UploadedBy    uuid.UUID  `json:"uploaded_by"`
	FileName      string     `json:"file_name"`
	FileType      string     `json:"file_type"`
	Status        string     `json:"status"`
	TotalRows     int        `json:"total_rows"`
	ProcessedRows int        `json:"processed_rows"`
	ErrorRows     int        `json:"error_rows"`
	Errors        JSONMap    `json:"errors,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// PackageRegistry represents an internal package.
type PackageRegistry struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	Name           string     `json:"name"`
	Scope          *string    `json:"scope,omitempty"`
	Description    *string    `json:"description,omitempty"`
	RegistryType   string     `json:"registry_type"`
	LatestVersion  *string    `json:"latest_version,omitempty"`
	TotalDownloads int        `json:"total_downloads"`
	IsInternal     bool       `json:"is_internal"`
	RepositoryURL  *string    `json:"repository_url,omitempty"`
	PublishedBy    *uuid.UUID `json:"published_by,omitempty"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PackageVersion represents a version of an internal package.
type PackageVersion struct {
	ID           uuid.UUID  `json:"id"`
	PackageID    uuid.UUID  `json:"package_id"`
	Version      string     `json:"version"`
	Description  *string    `json:"description,omitempty"`
	Dependencies JSONMap    `json:"dependencies,omitempty"`
	PublishedBy  *uuid.UUID `json:"published_by,omitempty"`
	Downloads    int        `json:"downloads"`
	TarballURL   *string    `json:"tarball_url,omitempty"`
	PublishedAt  time.Time  `json:"published_at"`
}

// ListFeedbackRoundsOpts contains options for listing feedback rounds.
type ListFeedbackRoundsOpts struct {
	Status *string
	Limit  int
	Offset int
}

// ListDocumentSignaturesOpts contains options for listing document signatures.
type ListDocumentSignaturesOpts struct {
	Status *string
	Limit  int
	Offset int
}

// ListWalletPassTemplatesOpts contains options for listing wallet pass templates.
type ListWalletPassTemplatesOpts struct {
	Platform *string
	IsActive *bool
	Limit    int
	Offset   int
}

// ListOrgChartImportsOpts contains options for listing org chart imports.
type ListOrgChartImportsOpts struct {
	Status *string
	Limit  int
	Offset int
}

// ListPackageRegistryOpts contains options for listing packages.
type ListPackageRegistryOpts struct {
	RegistryType *string
	Search       *string
	Limit        int
	Offset       int
}

// ListPackageVersionsOpts contains options for listing package versions.
type ListPackageVersionsOpts struct {
	Limit  int
	Offset int
}

// ── FFID Identity System Types ──────────────────────────────────────────────

// Clearance level constants
const (
	ClearanceL0 = 0 // Public
	ClearanceL1 = 1 // Standard
	ClearanceL2 = 2 // Department
	ClearanceL3 = 3 // Cross-Department
	ClearanceL4 = 4 // Executive
	ClearanceL5 = 5 // Founder
)

// ClearanceLevelFromText maps text clearance to numeric level.
func ClearanceLevelFromText(level string) int {
	switch level {
	case "standard":
		return ClearanceL1
	case "elevated":
		return ClearanceL2
	case "confidential":
		return ClearanceL3
	case "top_secret":
		return ClearanceL4
	default:
		return ClearanceL1
	}
}

// AchievementDefinition represents a pre-defined badge/achievement.
type AchievementDefinition struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	Slug              string     `json:"slug"`
	Name              string     `json:"name"`
	Description       *string    `json:"description,omitempty"`
	Icon              *string    `json:"icon,omitempty"`
	Category          string     `json:"category"`
	CriteriaType      string     `json:"criteria_type"`
	CriteriaThreshold int        `json:"criteria_threshold"`
	Points            int        `json:"points"`
	Tier              int        `json:"tier"`
	IsActive          bool       `json:"is_active"`
	CreatedAt         time.Time  `json:"created_at"`
}

// AchievementProgress tracks an employee's progress toward an achievement.
type AchievementProgress struct {
	ID            int64      `json:"id"`
	EmployeeID    uuid.UUID  `json:"employee_id"`
	AchievementID uuid.UUID  `json:"achievement_id"`
	CurrentValue  int        `json:"current_value"`
	Awarded       bool       `json:"awarded"`
	AwardedAt     *time.Time `json:"awarded_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CareerTimelineEvent represents a career milestone event.
type CareerTimelineEvent struct {
	ID          uuid.UUID  `json:"id"`
	EmployeeID  uuid.UUID  `json:"employee_id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	EventType   string     `json:"event_type"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Metadata    JSONMap    `json:"metadata,omitempty"`
	EventDate   time.Time  `json:"event_date"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ReputationHistory represents a historical reputation data point.
type ReputationHistory struct {
	ID         int64       `json:"id"`
	EmployeeID uuid.UUID   `json:"employee_id"`
	TenantID   uuid.UUID   `json:"tenant_id"`
	Category   string      `json:"category"`
	Score      float64     `json:"score"`
	RecordedAt time.Time   `json:"recorded_at"`
}

// IdentityCard represents the full identity card view for an employee.
type IdentityCard struct {
	Employee           *Employee              `json:"employee"`
	ClearanceLevel     int                    `json:"clearance_level"`
	IdentitySignature  string                 `json:"identity_signature"`
	ReputationTotal    int                    `json:"reputation_total"`
	TrustScore         float64                `json:"trust_score"`
	Achievements       []*AchievementProgress `json:"achievements"`
	Timeline           []*CareerTimelineEvent `json:"timeline"`
	ReputationByCategory map[string]float64   `json:"reputation_by_category"`
}
