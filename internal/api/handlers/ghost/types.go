package ghost

import (
	"os"
	"time"
)

type GhostPhase string

const (
	PhasePlanning     GhostPhase = "planning"
	PhaseProvisioning GhostPhase = "provisioning"
	PhaseBuilding     GhostPhase = "building"
	PhaseDeploying    GhostPhase = "deploying"
	PhaseMonitoring   GhostPhase = "monitoring"
	PhaseComplete     GhostPhase = "complete"
	PhaseError        GhostPhase = "error"
	PhasePaused       GhostPhase = "paused"
)

const (
	FeatureFlagGhostModeBeta = "GHOST_MODE_BETA"
)

func IsGhostModeBeta() bool {
	return os.Getenv(FeatureFlagGhostModeBeta) == "true"
}

func IsGhostModeProduction() bool {
	if os.Getenv("GHOST_MODE_ENABLED") == "false" {
		return false
	}
	return os.Getenv("GHOST_MODE_ENV") == "production" || os.Getenv("ENVIRONMENT") == "production"
}

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
	StatusSkipped    TaskStatus = "skipped"
)

type BuildState struct {
	ID                    string     `json:"id"`
	Goal                  string     `json:"goal"`
	Description           string     `json:"description"`
	Phase                 GhostPhase `json:"phase"`
	Progress              float64    `json:"progress"`
	Tasks                 []TaskState `json:"tasks"`
	StartedAt             time.Time  `json:"started_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	EstimatedCompletion   *time.Time `json:"estimated_completion,omitempty"`
	CurrentTaskID         string     `json:"current_task_id,omitempty"`
	HumanApprovalRequired bool       `json:"human_approval_required"`
	ApprovalType          string     `json:"approval_type,omitempty"`
	Error                 string     `json:"error,omitempty"`
	Artifacts             []Artifact `json:"artifacts,omitempty"`
}

type TaskState struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Status       TaskStatus `json:"status"`
	Phase        GhostPhase `json:"phase"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	DurationMs   int        `json:"duration_ms,omitempty"`
	Logs         []LogEntry `json:"logs"`
	Artifacts    []Artifact `json:"artifacts,omitempty"`
	AgentID      string     `json:"agent_id,omitempty"`
	Confidence   float64    `json:"confidence,omitempty"`
	Dependencies []string   `json:"dependencies,omitempty"`
	LLMOutput    string     `json:"llm_output,omitempty"`
}

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

type Artifact struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
	Size int    `json:"size,omitempty"`
}

type CreateBuildRequest struct {
	Goal        string `json:"goal"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AgentID     string `json:"agent_id,omitempty"`
}

type CreateBuildResponse struct {
	BuildID string `json:"build_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type GetBuildResponse struct {
	OK    bool        `json:"ok"`
	Build *BuildState `json:"build,omitempty"`
}

type ApprovalRequest struct {
	BuildID      string `json:"build_id"`
	ApprovalType string `json:"approval_type"`
	Decision     string `json:"decision"`
	Notes        string `json:"notes,omitempty"`
}

type LogEntryRequest struct {
	BuildID string `json:"build_id"`
	TaskID  string `json:"task_id"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

type PlanArchitectureRequest struct {
	Goal   string `json:"goal"`
	Domain string `json:"domain,omitempty"`
}

type ArchitecturePlan struct {
	Components    []ComponentSpec `json:"components"`
	DataModel       []EntitySpec    `json:"data_model"`
	APIDesign       []EndpointSpec  `json:"api_design"`
	TechStack       []string        `json:"tech_stack"`
	Dependencies    []string        `json:"dependencies"`
	EstimatedCost   string          `json:"estimated_cost"`
	RiskFactors     []string        `json:"risk_factors"`
}

type ComponentSpec struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Technology  string `json:"technology"`
}

type EntitySpec struct {
	Name      string         `json:"name"`
	Fields    []FieldSpec    `json:"fields"`
	Indexes   []string       `json:"indexes,omitempty"`
	Relations []RelationSpec `json:"relations,omitempty"`
}

type FieldSpec struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Default  string `json:"default,omitempty"`
}

type RelationSpec struct {
	Entity   string `json:"entity"`
	Type     string `json:"type"`
	ViaField string `json:"via_field,omitempty"`
}

type EndpointSpec struct {
	Method       string         `json:"method"`
	Path         string         `json:"path"`
	Handler      string         `json:"handler"`
	Auth         bool           `json:"auth"`
	InputSchema  map[string]any `json:"input_schema,omitempty"`
	OutputSchema map[string]any `json:"output_schema,omitempty"`
}

type ProvisionDatabaseRequest struct {
	Schema string `json:"schema"`
}

type ProvisionAPIRequest struct {
	Spec string `json:"spec"`
}

type ProvisionDockerRequest struct {
	Services []DockerServiceSpec `json:"services"`
}

type DockerServiceSpec struct {
	Name      string            `json:"name"`
	Image     string            `json:"image"`
	Ports     []string          `json:"ports"`
	Env       map[string]string `json:"env"`
	Volumes   []string          `json:"volumes"`
	DependsOn []string          `json:"depends_on,omitempty"`
}

type GenerateSchemaRequest struct {
	Entities []EntitySpec `json:"entities"`
}

type GenerateBackendRequest struct {
	Spec string `json:"spec"`
	Lang string `json:"language"`
}

type GenerateFrontendRequest struct {
	Spec      string `json:"spec"`
	Framework string `json:"framework"`
}

type GenerateTestsRequest struct {
	Code     string `json:"code"`
	Coverage string `json:"coverage_target"`
}

type DeployRequest struct {
	BuildID    string `json:"build_id"`
	CommitSHA  string `json:"commit_sha,omitempty"`
	RollbackID string `json:"rollback_id,omitempty"`
}

type SetupMonitoringRequest struct {
	Services   []string `json:"services"`
	Dashboards []string `json:"dashboards"`
}

type SecureContextRequest struct {
	TenantID    string            `json:"tenant_id"`
	UserID      string            `json:"user_id"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type UpdateBuildSecureRequest struct {
	BuildID      string  `json:"build_id"`
	Phase        string  `json:"phase,omitempty"`
	Progress     float64 `json:"progress,omitempty"`
	Error        string  `json:"error,omitempty"`
	AgentID      string  `json:"agent_id,omitempty"`
	SecureContext SecureContextRequest `json:"secure_context"`
}
