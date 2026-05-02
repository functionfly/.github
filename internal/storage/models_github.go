package storage

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// GitHubConnection represents a linked GitHub account with encrypted OAuth tokens.
type GitHubConnection struct {
	ID               uuid.UUID       `json:"id"`
	UserID           uuid.UUID       `json:"user_id"`
	TenantID         uuid.UUID       `json:"tenant_id"`
	GithubUserID     int64           `json:"github_user_id"`
	GithubUsername    string          `json:"github_username"`
	GithubAvatarURL  *string         `json:"github_avatar_url,omitempty"`
	GithubProfileURL *string         `json:"github_profile_url,omitempty"`
	EncryptedToken   string          `json:"encrypted_token"`
	TokenIV          string          `json:"token_iv"`
	TokenTag         string          `json:"token_tag"`
	EncryptedRefresh *string         `json:"encrypted_refresh,omitempty"`
	RefreshIV        *string         `json:"refresh_iv,omitempty"`
	RefreshTag       *string         `json:"refresh_tag,omitempty"`
	TokenScope       *string         `json:"token_scope,omitempty"`
	TokenExpiresAt   *time.Time      `json:"token_expires_at,omitempty"`
	GithubAppInstall bool            `json:"github_app_install"`
	GithubInstallID  *int64          `json:"github_install_id,omitempty"`
	Status           string          `json:"status"`
	LastSyncedAt     *time.Time      `json:"last_synced_at,omitempty"`
	Metadata         json.RawMessage `json:"metadata"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// GitHubRepo represents cached repository metadata with detected functions.
type GitHubRepo struct {
	ID                 uuid.UUID       `json:"id"`
	ConnectionID       uuid.UUID       `json:"connection_id"`
	GithubRepoID       int64           `json:"github_repo_id"`
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
	SizeKB             int             `json:"size_kb"`
	PushedAt           *time.Time      `json:"pushed_at,omitempty"`
	HtmlURL            string          `json:"html_url"`
	CloneURL           string          `json:"clone_url"`
	SSHURL             string          `json:"ssh_url"`
	DetectedFunctions  json.RawMessage `json:"detected_functions"`
	DetectedRuntime    *string         `json:"detected_runtime,omitempty"`
	HasFunctionflyJSON bool            `json:"has_functionfly_json"`
	ImportStatus       string          `json:"import_status"`
	Metadata           json.RawMessage `json:"metadata"`
	LastScannedAt      *time.Time      `json:"last_scanned_at,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// GitHubImport represents an import job with progress tracking.
type GitHubImport struct {
	ID                  uuid.UUID       `json:"id"`
	UserID              uuid.UUID       `json:"user_id"`
	TenantID            uuid.UUID       `json:"tenant_id"`
	ConnectionID        uuid.UUID       `json:"connection_id"`
	RepoID              uuid.UUID       `json:"repo_id"`
	SourceBranch        string          `json:"source_branch"`
	SourcePath          *string         `json:"source_path,omitempty"`
	FunctionName        string          `json:"function_name"`
	FunctionID          *uuid.UUID      `json:"function_id,omitempty"`
	FunctionVersionID   *uuid.UUID      `json:"function_version_id,omitempty"`
	Visibility          string          `json:"visibility"`
	RuntimeOverride     *string         `json:"runtime_override,omitempty"`
	ManifestOverrides   json.RawMessage `json:"manifest_overrides"`
	AutoSyncEnabled     bool            `json:"auto_sync_enabled"`
	SyncBranches        json.RawMessage `json:"sync_branches"`
	EnvironmentMappings json.RawMessage `json:"environment_mappings"`
	Status              string          `json:"status"`
	Progress            int             `json:"progress"`
	ErrorMessage        *string         `json:"error_message,omitempty"`
	ErrorDetails        json.RawMessage `json:"error_details"`
	ContentHash         *string         `json:"content_hash,omitempty"`
	CommitSHA           *string         `json:"commit_sha,omitempty"`
	FilesImported       int             `json:"files_imported"`
	TotalSizeBytes      int64           `json:"total_size_bytes"`
	Metadata            json.RawMessage `json:"metadata"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	CompletedAt         *time.Time      `json:"completed_at,omitempty"`
}

// GitHubWebhook represents a registered webhook for auto-sync.
type GitHubWebhook struct {
	ID             uuid.UUID       `json:"id"`
	ConnectionID   uuid.UUID       `json:"connection_id"`
	RepoID         uuid.UUID       `json:"repo_id"`
	GithubWebhookID *int64         `json:"github_webhook_id,omitempty"`
	WebhookSecret  string          `json:"webhook_secret"`
	Events         json.RawMessage `json:"events"`
	IsActive       bool            `json:"is_active"`
	LastDeliveryAt *time.Time      `json:"last_delivery_at,omitempty"`
	LastEventType  *string         `json:"last_event_type,omitempty"`
	DeliveryCount  int             `json:"delivery_count"`
	ErrorCount     int             `json:"error_count"`
	LastError      *string         `json:"last_error,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// GitHubSyncLog represents sync/deploy history.
type GitHubSyncLog struct {
	ID               uuid.UUID       `json:"id"`
	ImportID         uuid.UUID       `json:"import_id"`
	FunctionID       *uuid.UUID      `json:"function_id,omitempty"`
	TriggerType      string          `json:"trigger_type"`
	TriggerBranch    *string         `json:"trigger_branch,omitempty"`
	TriggerCommitSHA *string         `json:"trigger_commit_sha,omitempty"`
	TriggerPRNumber  *int            `json:"trigger_pr_number,omitempty"`
	Status           string          `json:"status"`
	VersionPublished *string         `json:"version_published,omitempty"`
	StatusCheckURL   *string         `json:"status_check_url,omitempty"`
	DurationMs       *int            `json:"duration_ms,omitempty"`
	ErrorMessage     *string         `json:"error_message,omitempty"`
	Metadata         json.RawMessage `json:"metadata"`
	CreatedAt        time.Time       `json:"created_at"`
	CompletedAt      *time.Time      `json:"completed_at,omitempty"`
}

// GitHubImportTemplate represents a reusable import configuration.
type GitHubImportTemplate struct {
	ID             uuid.UUID       `json:"id"`
	TenantID       uuid.UUID       `json:"tenant_id"`
	UserID         uuid.UUID       `json:"user_id"`
	Name           string          `json:"name"`
	Description    *string         `json:"description,omitempty"`
	Config         json.RawMessage `json:"config"`
	DetectionRules json.RawMessage `json:"detection_rules"`
	IsDefault      bool            `json:"is_default"`
	UsageCount     int             `json:"usage_count"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// ListReposParams contains filtering/pagination parameters for listing repos.
type ListReposParams struct {
	Page       int
	PerPage    int
	Sort       string
	Direction  string
	Language   string
	Visibility string
	Search     string
}

// ListImportsParams contains filtering/pagination parameters for listing imports.
type ListImportsParams struct {
	Page    int
	PerPage int
	Status  string
	RepoID  *uuid.UUID
}

// ListSyncLogsParams contains filtering/pagination parameters for listing sync logs.
type ListSyncLogsParams struct {
	Page    int
	PerPage int
	Status  string
}
