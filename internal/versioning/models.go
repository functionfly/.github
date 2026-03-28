// Package versioning provides models for API and function version management.
// This is distinct from the build version in the version package.
package versioning

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// API version status constants
const (
	APIVersionStatusActive     = "active"
	APIVersionStatusDeprecated = "deprecated"
	APIVersionStatusSunset     = "sunset"
	APIVersionStatusArchived   = "archived"
)

// Function version state constants
const (
	FunctionVersionStateDraft      = "draft"
	FunctionVersionStatePublished  = "published"
	FunctionVersionStateDeprecated = "deprecated"
	FunctionVersionStateArchived   = "archived"
)

// Change type constants
const (
	ChangeTypeMajor       = "major"
	ChangeTypeMinor       = "minor"
	ChangeTypePatch       = "patch"
	ChangeTypeBreaking    = "breaking"
	ChangeTypeFeature     = "feature"
	ChangeTypeFix         = "fix"
	ChangeTypeSecurity    = "security"
	ChangeTypeDeprecation = "deprecation"
)

// Deployment version status constants
const (
	DeploymentVersionStatusPending    = "pending"
	DeploymentVersionStatusBuilding   = "building"
	DeploymentVersionStatusDeploying  = "deploying"
	DeploymentVersionStatusSuccess    = "success"
	DeploymentVersionStatusFailed     = "failed"
	DeploymentVersionStatusRolledBack = "rolled_back"
)

// APIVersion represents a platform API version
type APIVersion struct {
	ID             uuid.UUID       `json:"id"`
	Version        string          `json:"version"`
	PathPrefix     string          `json:"path_prefix"`
	Status         string          `json:"status"`
	IsDefault      bool            `json:"is_default"`
	ReleasedAt     time.Time       `json:"released_at"`
	DeprecatedAt   *time.Time      `json:"deprecated_at,omitempty"`
	SunsetAt       *time.Time      `json:"sunset_at,omitempty"`
	SunsetMessage  string          `json:"sunset_message,omitempty"`
	Metadata       json.RawMessage `json:"metadata"`
	OpenAPISpecURL string          `json:"openapi_spec_url,omitempty"`
	ChangelogURL   string          `json:"changelog_url,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// APIVersionResponse represents the API response format for API versions
type APIVersionResponse struct {
	Version     string           `json:"version"`
	Status      string           `json:"status"`
	ReleasedAt  time.Time        `json:"releasedAt"`
	Features    []string         `json:"features,omitempty"`
	Deprecation *DeprecationInfo `json:"deprecation,omitempty"`
}

// DeprecationInfo contains deprecation details for API versions
type DeprecationInfo struct {
	DeprecatedAt   time.Time `json:"deprecatedAt"`
	SunsetAt       time.Time `json:"sunsetAt,omitempty"`
	MigrationGuide string    `json:"migrationGuide,omitempty"`
}

// FunctionVersion represents a function version with state tracking
// This extends the registry_function_versions table
type FunctionVersion struct {
	ID                uuid.UUID  `json:"id"`
	FunctionID        uuid.UUID  `json:"function_id"`
	Version           string     `json:"version"`
	VersionState      string     `json:"version_state"`
	DeprecationReason string     `json:"deprecation_reason,omitempty"`
	ReplacedByVersion string     `json:"replaced_by_version,omitempty"`
	MigrationGuide    string     `json:"migration_guide,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	PublishedAt       *time.Time `json:"published_at,omitempty"`
	ArchivedAt        *time.Time `json:"archived_at,omitempty"`
}

// FunctionVersionResponse represents the API response format for function versions
type FunctionVersionResponse struct {
	ID          uuid.UUID               `json:"id"`
	FunctionID  uuid.UUID               `json:"functionId"`
	Version     string                  `json:"version"`
	Status      string                  `json:"status"`
	PublishedAt *time.Time              `json:"publishedAt,omitempty"`
	IsLatest    bool                    `json:"isLatest"`
	IsStable    bool                    `json:"isStable"`
	Deprecation *VersionDeprecationInfo `json:"deprecation,omitempty"`
}

// VersionDeprecationInfo contains deprecation details for function versions
type VersionDeprecationInfo struct {
	DeprecatedAt   time.Time  `json:"deprecatedAt,omitempty"`
	SunsetAt       *time.Time `json:"sunsetAt,omitempty"`
	Reason         string     `json:"reason,omitempty"`
	MigrationGuide string     `json:"migrationGuide,omitempty"`
	ReplacedBy     string     `json:"replacedBy,omitempty"`
}

// FunctionVersionChangelog represents a changelog entry for a function version
type FunctionVersionChangelog struct {
	ID              uuid.UUID       `json:"id"`
	FunctionID      uuid.UUID       `json:"function_id"`
	Version         string          `json:"version"`
	ChangeType      string          `json:"change_type"`
	ChangeCategory  string          `json:"change_category"`
	Description     string          `json:"description"`
	BreakingChanges json.RawMessage `json:"breaking_changes"`
	MigrationSteps  json.RawMessage `json:"migration_steps"`
	CreatedBy       *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

// ChangelogEntry represents a single changelog entry
type ChangelogEntry struct {
	Type        string `json:"type"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// DeploymentVersion represents a deployment of a specific function version
type DeploymentVersion struct {
	ID           uuid.UUID       `json:"id"`
	FunctionID   uuid.UUID       `json:"function_id"`
	Version      string          `json:"function_version"`
	DeploymentID *uuid.UUID      `json:"deployment_id,omitempty"`
	Provider     string          `json:"provider"`
	Region       string          `json:"region,omitempty"`
	Status       string          `json:"status"`
	ArtifactURI  string          `json:"artifact_uri,omitempty"`
	Checksum     string          `json:"checksum,omitempty"`
	RollbackID   *uuid.UUID      `json:"rollback_id,omitempty"`
	Metadata     json.RawMessage `json:"metadata"`
	CreatedAt    time.Time       `json:"created_at"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
}

// ServiceContract represents an internal service contract for versioning
type ServiceContract struct {
	ID                  uuid.UUID       `json:"id"`
	ServiceName         string          `json:"service_name"`
	ContractVersion     string          `json:"contract_version"`
	ContractType        string          `json:"contract_type"`
	Schema              json.RawMessage `json:"schema"`
	Status              string          `json:"status"`
	IntroducedInRelease string          `json:"introduced_in_release,omitempty"`
	DeprecatedInRelease string          `json:"deprecated_in_release,omitempty"`
	RemovedInRelease    string          `json:"removed_in_release,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

// ListAPIVersionsParams contains parameters for listing API versions
type ListAPIVersionsParams struct {
	Status string
	Limit  int
	Cursor string
}

// ListFunctionVersionsParams contains parameters for listing function versions
type ListFunctionVersionsParams struct {
	FunctionID uuid.UUID
	Status     string
	Limit      int
	Cursor     string
}

// CreateChangelogParams contains parameters for creating a changelog entry
type CreateChangelogParams struct {
	FunctionID      uuid.UUID
	Version         string
	ChangeType      string
	ChangeCategory  string
	Description     string
	BreakingChanges []string
	MigrationSteps  []string
	CreatedBy       *uuid.UUID
}

// VersionAlias represents a symbolic version reference
type VersionAlias string

const (
	VersionAliasLatest VersionAlias = "latest"
	VersionAliasStable VersionAlias = "stable"
	VersionAliasDraft  VersionAlias = "draft"
)

// ResolveVersionAlias resolves a version alias to a concrete version
func ResolveVersionAlias(alias VersionAlias, versions []string) string {
	switch alias {
	case VersionAliasLatest:
		// Return the highest version (last in sorted list)
		if len(versions) > 0 {
			return versions[len(versions)-1]
		}
	case VersionAliasStable:
		// Return the highest non-prerelease version
		for i := len(versions) - 1; i >= 0; i-- {
			if !isPrerelease(versions[i]) {
				return versions[i]
			}
		}
	case VersionAliasDraft:
		// Return "draft" for unpublished work
		return string(VersionAliasDraft)
	}
	return ""
}

// isPrerelease checks if a version string is a prerelease
func isPrerelease(version string) bool {
	for _, c := range []string{"-alpha", "-beta", "-rc", "-dev"} {
		if len(version) > len(c) && version[len(version)-len(c):] == c {
			return true
		}
	}
	return false
}

// ==================== Phase 2: Publishing and Rollback Models ====================

// RollbackStrategy defines the type of rollback strategy
type RollbackStrategy string

const (
	RollbackStrategyImmediate RollbackStrategy = "immediate"
	RollbackStrategyGradual   RollbackStrategy = "gradual"
	RollbackStrategyCanary    RollbackStrategy = "canary"
)

// VersionAliasRecord represents a version alias (latest, stable)
type VersionAliasRecord struct {
	ID         uuid.UUID `json:"id"`
	FunctionID uuid.UUID `json:"function_id"`
	Alias      string    `json:"alias"`
	VersionID  uuid.UUID `json:"version_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// RollbackRecord represents a rollback history entry
type RollbackRecord struct {
	ID          uuid.UUID        `json:"id"`
	FunctionID  uuid.UUID        `json:"function_id"`
	FromVersion string           `json:"from_version"`
	ToVersion   string           `json:"to_version"`
	Strategy    RollbackStrategy `json:"strategy"`
	Status      string           `json:"status"`
	InitiatedBy *uuid.UUID       `json:"initiated_by,omitempty"`
	InitiatedAt time.Time        `json:"initiated_at"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
	Metadata    json.RawMessage  `json:"metadata"`
}

// PublishVersionRequest represents a request to publish a function version
type PublishVersionRequest struct {
	Version     string `json:"version"`
	SetAsLatest bool   `json:"setAsLatest"`
	SetAsStable bool   `json:"setAsStable"`
}

// PublishVersionResponse represents the response after publishing a version
type PublishVersionResponse struct {
	ID          uuid.UUID  `json:"id"`
	FunctionID  uuid.UUID  `json:"functionId"`
	Version     string     `json:"version"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
	IsLatest    bool       `json:"isLatest"`
	IsStable    bool       `json:"isStable"`
}

// ArchiveVersionRequest represents a request to archive a function version
type ArchiveVersionRequest struct {
	Reason string `json:"reason"`
}

// ArchiveVersionResponse represents the response after archiving a version
type ArchiveVersionResponse struct {
	Version    string     `json:"version"`
	Status     string     `json:"status"`
	ArchivedAt *time.Time `json:"archivedAt,omitempty"`
}

// RollbackVersionRequest represents a request to rollback to a version
type RollbackVersionRequest struct {
	ToVersion string           `json:"toVersion"`
	Strategy  RollbackStrategy `json:"strategy"`
}

// RollbackVersionResponse represents the response after a rollback
type RollbackVersionResponse struct {
	RollbackID  uuid.UUID        `json:"rollbackId"`
	FunctionID  uuid.UUID        `json:"functionId"`
	FromVersion string           `json:"fromVersion"`
	ToVersion   string           `json:"toVersion"`
	Strategy    RollbackStrategy `json:"strategy"`
	Status      string           `json:"status"`
	InitiatedAt time.Time        `json:"initiatedAt"`
	CompletedAt *time.Time       `json:"completedAt,omitempty"`
}

// DeprecateVersionRequest represents a request to deprecate a function version
type DeprecateVersionRequest struct {
	Reason          string     `json:"reason"`
	ReplacedBy      string     `json:"replacedBy"`
	MigrationGuide  string     `json:"migrationGuide"`
	EffectiveAt     *time.Time `json:"effectiveAt"`
	GracePeriodDays int        `json:"gracePeriodDays"`
}

// SetAliasRequest represents a request to set a version alias
type SetAliasRequest struct {
	Version string `json:"version"`
}

// SetAliasResponse represents the response after setting an alias
type SetAliasResponse struct {
	Alias   string `json:"alias"`
	Version string `json:"version"`
}

// CreateAPIVersionRequest represents a request to create an API version
type CreateAPIVersionRequest struct {
	Version        string          `json:"version"`
	PathPrefix     string          `json:"pathPrefix"`
	Status         string          `json:"status"`
	ReleasedAt     time.Time       `json:"releasedAt"`
	OpenAPISpecURL string          `json:"openapiSpecUrl"`
	ChangelogURL   string          `json:"changelogUrl"`
	Metadata       json.RawMessage `json:"metadata"`
}

// UpdateAPIVersionRequest represents a request to update an API version
type UpdateAPIVersionRequest struct {
	PathPrefix     *string         `json:"pathPrefix"`
	Status         *string         `json:"status"`
	OpenAPISpecURL *string         `json:"openapiSpecUrl"`
	ChangelogURL   *string         `json:"changelogUrl"`
	Metadata       json.RawMessage `json:"metadata"`
}

// ==================== Phase 3: Deployment and Contract Models ====================

// DeploymentMetadata contains additional deployment information
type DeploymentMetadata struct {
	Runtime      string            `json:"runtime,omitempty"`
	Config       map[string]string `json:"config,omitempty"`
	Environment  string            `json:"environment,omitempty"`
	BuildTime    int64             `json:"buildTime,omitempty"`
	DeployTime   int64             `json:"deployTime,omitempty"`
	InstanceType string            `json:"instanceType,omitempty"`
	MemoryMB     int               `json:"memoryMB,omitempty"`
	TimeoutSec   int               `json:"timeoutSec,omitempty"`
}

// DeploymentVersionResponse represents the API response for a deployment
type DeploymentVersionResponse struct {
	ID          uuid.UUID           `json:"id"`
	FunctionID  uuid.UUID           `json:"functionId"`
	Version     string              `json:"version"`
	Provider    string              `json:"provider"`
	Region      string              `json:"region,omitempty"`
	Status      string              `json:"status"`
	ArtifactURI string              `json:"artifactUri,omitempty"`
	Checksum    string              `json:"checksum,omitempty"`
	Metadata    *DeploymentMetadata `json:"metadata,omitempty"`
	CreatedAt   time.Time           `json:"createdAt"`
	CompletedAt *time.Time          `json:"completedAt,omitempty"`
}

// ContractSchema represents a service contract schema
type ContractSchema struct {
	Endpoints []EndpointSchema `json:"endpoints"`
	DataTypes []DataTypeSchema `json:"dataTypes"`
	Version   string           `json:"version"`
	Format    string           `json:"format"` // openapi3, json-schema
}

// EndpointSchema represents an endpoint in a service contract
type EndpointSchema struct {
	Path       string          `json:"path"`
	Method     string          `json:"method"`
	Summary    string          `json:"summary,omitempty"`
	Deprecated bool            `json:"deprecated"`
	Request    *RequestSchema  `json:"request,omitempty"`
	Response   *ResponseSchema `json:"response,omitempty"`
}

// RequestSchema represents a request schema
type RequestSchema struct {
	ContentType string `json:"contentType"`
	Schema      any    `json:"schema"`
}

// ResponseSchema represents a response schema
type ResponseSchema struct {
	ContentType string `json:"contentType"`
	Schema      any    `json:"schema"`
}

// DataTypeSchema represents a data type in the contract
type DataTypeSchema struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ServiceContractResponse represents the API response for a service contract
type ServiceContractResponse struct {
	ID                  uuid.UUID       `json:"id"`
	ServiceName         string          `json:"serviceName"`
	ContractVersion     string          `json:"contractVersion"`
	ContractType        string          `json:"contractType"`
	Schema              *ContractSchema `json:"schema,omitempty"`
	Status              string          `json:"status"`
	IntroducedInRelease string          `json:"introducedInRelease,omitempty"`
	DeprecatedInRelease string          `json:"deprecatedInRelease,omitempty"`
	CreatedAt           time.Time       `json:"createdAt"`
}

// ContractListResponse represents a list of service contracts
type ContractListResponse struct {
	Services  []string                  `json:"services"`
	Contracts []ServiceContractResponse `json:"contracts,omitempty"`
}

// ContractVersionNegotiationRequest represents a contract version negotiation request
type ContractVersionNegotiationRequest struct {
	ConsumerService   string   `json:"consumerService"`
	ProviderService   string   `json:"providerService"`
	SupportedVersions []string `json:"supportedVersions"`
}

// ContractVersionNegotiationResponse represents the response for contract version negotiation
type ContractVersionNegotiationResponse struct {
	ConsumerService string `json:"consumerService"`
	ProviderService string `json:"providerService"`
	AgreedVersion   string `json:"agreedVersion"`
	Compatible      bool   `json:"compatible"`
	Reason          string `json:"reason,omitempty"`
}

// SchemaCompatibilityResult represents schema compatibility check result
type SchemaCompatibilityResult struct {
	Compatible      bool     `json:"compatible"`
	BreakingChanges []string `json:"breakingChanges"`
	Warnings        []string `json:"warnings"`
}

// ==================== Phase 3: Version Lineage Models ====================

// VersionLineageEntry represents a single entry in version history
type VersionLineageEntry struct {
	ID            uuid.UUID  `json:"id"`
	Version       string     `json:"version"`
	State         string     `json:"state"`
	ParentVersion string     `json:"parentVersion,omitempty"`
	ChangeType    string     `json:"changeType,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	PublishedAt   *time.Time `json:"publishedAt,omitempty"`
	ArchivedAt    *time.Time `json:"archivedAt,omitempty"`
}

// VersionLineageResponse represents the version history/lineage
type VersionLineageResponse struct {
	FunctionID uuid.UUID             `json:"functionId"`
	Entries    []VersionLineageEntry `json:"entries"`
	TotalCount int                   `json:"totalCount"`
}

// VersionDiffEntry represents a change between two versions
type VersionDiffEntry struct {
	Field      string `json:"field"`
	FromValue  string `json:"fromValue"`
	ToValue    string `json:"toValue"`
	ChangeType string `json:"changeType"` // added, removed, modified
}

// VersionDiffResponse represents the diff between two versions
type VersionDiffResponse struct {
	FunctionID uuid.UUID          `json:"functionId"`
	Version1   string             `json:"version1"`
	Version2   string             `json:"version2"`
	Changes    []VersionDiffEntry `json:"changes"`
	Summary    VersionDiffSummary `json:"summary"`
}

// VersionDiffSummary contains a summary of changes
type VersionDiffSummary struct {
	TotalChanges    int      `json:"totalChanges"`
	Added           int      `json:"added"`
	Removed         int      `json:"removed"`
	Modified        int      `json:"modified"`
	Breaking        bool     `json:"breaking"`
	BreakingChanges []string `json:"breakingChanges"`
}

// VersionCompareQuery represents query parameters for version comparison
type VersionCompareQuery struct {
	V1 string `json:"v1"` // first version to compare
	V2 string `json:"v2"` // second version to compare
}

// ListDeploymentsParams contains parameters for listing deployments
type ListDeploymentsParams struct {
	FunctionID uuid.UUID
	Version    string
	Status     string
	Limit      int
	Cursor     string
}

// ListServiceContractsParams contains parameters for listing service contracts
type ListServiceContractsParams struct {
	ServiceName string
	Status      string
	Limit       int
}
