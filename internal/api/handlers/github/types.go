package github

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ConnectResponse struct {
	URL string `json:"url"`
}

type ConnectionResponse struct {
	ID               uuid.UUID  `json:"id"`
	GithubUsername   string     `json:"github_username"`
	GithubAvatarURL  *string    `json:"github_avatar_url,omitempty"`
	GithubProfileURL *string    `json:"github_profile_url,omitempty"`
	TokenScope       *string    `json:"token_scope,omitempty"`
	TokenExpiresAt   *time.Time `json:"token_expires_at,omitempty"`
	Status           string     `json:"status"`
	ConnectedAt      time.Time  `json:"connected_at"`
	LastSyncedAt     *time.Time `json:"last_synced_at,omitempty"`
}

type RefreshTokenResponse struct {
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type ListReposResponse struct {
	Repos   []*RepoResponse `json:"repos"`
	Total   int             `json:"total"`
	Page    int             `json:"page"`
	PerPage int             `json:"per_page"`
}

type RepoResponse struct {
	ID                 uuid.UUID       `json:"id"`
	FullName           string          `json:"full_name"`
	Name               string          `json:"name"`
	Owner              string          `json:"owner"`
	Description        *string         `json:"description,omitempty"`
	DefaultBranch      string          `json:"default_branch"`
	Language           *string         `json:"language,omitempty"`
	Languages          json.RawMessage `json:"languages"`
	IsPrivate          bool            `json:"is_private"`
	IsFork             bool            `json:"is_fork"`
	IsArchived         bool            `json:"is_archived"`
	Topics             json.RawMessage `json:"topics"`
	StarsCount         int             `json:"stars_count"`
	ForksCount         int             `json:"forks_count"`
	HtmlURL            string          `json:"html_url"`
	DetectedFunctions  json.RawMessage `json:"detected_functions"`
	DetectedRuntime    *string         `json:"detected_runtime,omitempty"`
	HasFunctionflyJSON bool            `json:"has_functionfly_json"`
	ImportStatus       string          `json:"import_status"`
	LastScannedAt      *time.Time      `json:"last_scanned_at,omitempty"`
}

type ScanRepoResponse struct {
	Functions            []interface{} `json:"functions"`
	PrimaryRuntime       string        `json:"primary_runtime"`
	OverallConfidence    float64       `json:"overall_confidence"`
	StrategyUsed         string        `json:"strategy_used"`
	Warnings             []string      `json:"warnings"`
	EstimatedImportTimeS int           `json:"estimated_import_time_seconds"`
	EstimatedCostUSD     float64       `json:"estimated_cost_usd"`
}

type BranchResponse struct {
	Name      string `json:"name"`
	SHA       string `json:"sha"`
	Protected bool   `json:"protected"`
}

type TreeNodeResponse struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
	Size int    `json:"size"`
}

type ImportRequest struct {
	RepoID            uuid.UUID       `json:"repo_id"`
	Branch            string          `json:"branch"`
	SourcePath        *string         `json:"source_path,omitempty"`
	FunctionName      string          `json:"function_name"`
	FunctionNames     []string        `json:"function_names,omitempty"`
	Visibility        string          `json:"visibility"`
	RuntimeOverride   *string         `json:"runtime_override,omitempty"`
	ManifestOverrides json.RawMessage `json:"manifest_overrides,omitempty"`
	AutoSync          bool            `json:"auto_sync"`
	SyncBranches      json.RawMessage `json:"sync_branches,omitempty"`
}

type BulkImportRequest struct {
	Imports []ImportRequest `json:"imports"`
}

type ImportResponse struct {
	ImportID uuid.UUID `json:"import_id"`
	Status   string    `json:"status"`
}

type BulkImportResponse struct {
	Imports []ImportResponse `json:"imports"`
}

type ListImportsResponse struct {
	Imports []*ImportDetailResponse `json:"imports"`
	Total   int                     `json:"total"`
	Page    int                     `json:"page"`
	PerPage int                     `json:"per_page"`
}

type ImportDetailResponse struct {
	ID                uuid.UUID       `json:"id"`
	RepoID            uuid.UUID       `json:"repo_id"`
	FunctionName      string          `json:"function_name"`
	SourceBranch      string          `json:"source_branch"`
	SourcePath        *string         `json:"source_path,omitempty"`
	Visibility        string          `json:"visibility"`
	AutoSyncEnabled   bool            `json:"auto_sync_enabled"`
	SyncBranches      json.RawMessage `json:"sync_branches"`
	Status            string          `json:"status"`
	Progress          int             `json:"progress"`
	ErrorMessage      *string         `json:"error_message,omitempty"`
	FunctionID        *uuid.UUID      `json:"function_id,omitempty"`
	FunctionAuthor    string          `json:"function_author,omitempty"`
	FunctionVersionID *uuid.UUID      `json:"function_version_id,omitempty"`
	CommitSHA         *string         `json:"commit_sha,omitempty"`
	FilesImported     int             `json:"files_imported"`
	TotalSizeBytes    int64           `json:"total_size_bytes"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	CompletedAt       *time.Time      `json:"completed_at,omitempty"`
}

type UpdateSyncRequest struct {
	AutoSyncEnabled bool            `json:"auto_sync_enabled"`
	SyncBranches    json.RawMessage `json:"sync_branches,omitempty"`
}

type SyncLogResponse struct {
	ID               uuid.UUID  `json:"id"`
	TriggerType      string     `json:"trigger_type"`
	TriggerBranch    *string    `json:"trigger_branch,omitempty"`
	TriggerCommitSHA *string    `json:"trigger_commit_sha,omitempty"`
	TriggerPRNumber  *int       `json:"trigger_pr_number,omitempty"`
	Status           string     `json:"status"`
	VersionPublished *string    `json:"version_published,omitempty"`
	DurationMs       *int       `json:"duration_ms,omitempty"`
	ErrorMessage     *string    `json:"error_message,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

type ListSyncLogsResponse struct {
	Logs    []*SyncLogResponse `json:"logs"`
	Total   int                `json:"total"`
	Page    int                `json:"page"`
	PerPage int                `json:"per_page"`
}

type TemplateRequest struct {
	Name           string          `json:"name"`
	Description    *string         `json:"description,omitempty"`
	Config         json.RawMessage `json:"config"`
	DetectionRules json.RawMessage `json:"detection_rules,omitempty"`
	IsDefault      bool            `json:"is_default"`
}

type TemplateResponse struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	Description    *string         `json:"description,omitempty"`
	Config         json.RawMessage `json:"config"`
	DetectionRules json.RawMessage `json:"detection_rules"`
	IsDefault      bool            `json:"is_default"`
	UsageCount     int             `json:"usage_count"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type ProgressEvent struct {
	Stage    string `json:"stage"`
	Progress int    `json:"progress"`
	Message  string `json:"message"`
}
