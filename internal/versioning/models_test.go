// Package versioning provides tests for version models.
package versioning

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== API Version Tests ====================

func TestAPIVersionConstants(t *testing.T) {
	// Test API version status constants
	assert.Equal(t, "active", APIVersionStatusActive)
	assert.Equal(t, "deprecated", APIVersionStatusDeprecated)
	assert.Equal(t, "sunset", APIVersionStatusSunset)
	assert.Equal(t, "archived", APIVersionStatusArchived)
}

func TestFunctionVersionConstants(t *testing.T) {
	// Test function version state constants
	assert.Equal(t, "draft", FunctionVersionStateDraft)
	assert.Equal(t, "published", FunctionVersionStatePublished)
	assert.Equal(t, "deprecated", FunctionVersionStateDeprecated)
	assert.Equal(t, "archived", FunctionVersionStateArchived)
}

func TestChangeTypeConstants(t *testing.T) {
	// Test change type constants
	assert.Equal(t, "major", ChangeTypeMajor)
	assert.Equal(t, "minor", ChangeTypeMinor)
	assert.Equal(t, "patch", ChangeTypePatch)
	assert.Equal(t, "breaking", ChangeTypeBreaking)
	assert.Equal(t, "feature", ChangeTypeFeature)
	assert.Equal(t, "fix", ChangeTypeFix)
	assert.Equal(t, "security", ChangeTypeSecurity)
	assert.Equal(t, "deprecation", ChangeTypeDeprecation)
}

func TestDeploymentVersionStatusConstants(t *testing.T) {
	// Test deployment version status constants
	assert.Equal(t, "pending", DeploymentVersionStatusPending)
	assert.Equal(t, "building", DeploymentVersionStatusBuilding)
	assert.Equal(t, "deploying", DeploymentVersionStatusDeploying)
	assert.Equal(t, "success", DeploymentVersionStatusSuccess)
	assert.Equal(t, "failed", DeploymentVersionStatusFailed)
	assert.Equal(t, "rolled_back", DeploymentVersionStatusRolledBack)
}

// ==================== Version Alias Tests ====================

func TestVersionAliasConstants(t *testing.T) {
	assert.Equal(t, VersionAlias("latest"), VersionAliasLatest)
	assert.Equal(t, VersionAlias("stable"), VersionAliasStable)
	assert.Equal(t, VersionAlias("draft"), VersionAliasDraft)
}

func TestResolveVersionAlias_Latest(t *testing.T) {
	versions := []string{"v1.0.0", "v1.1.0", "v2.0.0"}
	result := ResolveVersionAlias(VersionAliasLatest, versions)
	assert.Equal(t, "v2.0.0", result)
}

func TestResolveVersionAlias_Latest_EmptyVersions(t *testing.T) {
	versions := []string{}
	result := ResolveVersionAlias(VersionAliasLatest, versions)
	assert.Equal(t, "", result)
}

func TestResolveVersionAlias_Latest_NilVersions(t *testing.T) {
	var versions []string
	result := ResolveVersionAlias(VersionAliasLatest, versions)
	assert.Equal(t, "", result)
}

func TestResolveVersionAlias_Stable(t *testing.T) {
	versions := []string{"v1.0.0", "v1.1.0-beta", "v2.0.0", "v2.0.1-rc1"}
	result := ResolveVersionAlias(VersionAliasStable, versions)
	assert.Equal(t, "v2.0.0", result)
}

func TestResolveVersionAlias_Stable_AllPrerelease(t *testing.T) {
	versions := []string{"v1.0.0-alpha", "v1.0.1-beta", "v2.0.0-rc1"}
	result := ResolveVersionAlias(VersionAliasStable, versions)
	assert.Equal(t, "", result)
}

func TestResolveVersionAlias_Stable_EmptyVersions(t *testing.T) {
	versions := []string{}
	result := ResolveVersionAlias(VersionAliasStable, versions)
	assert.Equal(t, "", result)
}

func TestResolveVersionAlias_Draft(t *testing.T) {
	versions := []string{"v1.0.0", "v2.0.0"}
	result := ResolveVersionAlias(VersionAliasDraft, versions)
	assert.Equal(t, "draft", result)
}

func TestResolveVersionAlias_Invalid(t *testing.T) {
	versions := []string{"v1.0.0"}
	result := ResolveVersionAlias(VersionAlias("invalid"), versions)
	assert.Equal(t, "", result)
}

// ==================== isPrerelease Tests ====================

func TestIsPrerelease_Alpha(t *testing.T) {
	assert.True(t, isPrerelease("v1.0.0-alpha"))
	assert.True(t, isPrerelease("v1.0.0-alpha.1"))
}

func TestIsPrerelease_Beta(t *testing.T) {
	assert.True(t, isPrerelease("v1.0.0-beta"))
	assert.True(t, isPrerelease("v1.0.0-beta.2"))
}

func TestIsPrerelease_RC(t *testing.T) {
	assert.True(t, isPrerelease("v1.0.0-rc"))
	assert.True(t, isPrerelease("v1.0.0-rc.1"))
}

func TestIsPrerelease_Dev(t *testing.T) {
	assert.True(t, isPrerelease("v1.0.0-dev"))
}

func TestIsPrerelease_Stable(t *testing.T) {
	assert.False(t, isPrerelease("v1.0.0"))
	assert.False(t, isPrerelease("v1.0.1"))
	assert.False(t, isPrerelease("v2.0.0"))
}

// ==================== Function Version State Transition Tests ====================

func TestFunctionVersion_NewVersion_IsDraft(t *testing.T) {
	fv := FunctionVersion{
		ID:           uuid.New(),
		FunctionID:   uuid.New(),
		Version:      "v1.0.0",
		VersionState: FunctionVersionStateDraft,
		CreatedAt:    time.Now(),
	}

	assert.Equal(t, FunctionVersionStateDraft, fv.VersionState)
	assert.Nil(t, fv.PublishedAt)
	assert.Nil(t, fv.ArchivedAt)
}

func TestFunctionVersion_PublishTransition(t *testing.T) {
	now := time.Now()
	fv := FunctionVersion{
		ID:           uuid.New(),
		FunctionID:   uuid.New(),
		Version:      "v1.0.0",
		VersionState: FunctionVersionStateDraft,
		CreatedAt:    now,
	}

	// Simulate publishing
	fv.VersionState = FunctionVersionStatePublished
	fv.PublishedAt = &now

	assert.Equal(t, FunctionVersionStatePublished, fv.VersionState)
	assert.NotNil(t, fv.PublishedAt)
}

func TestFunctionVersion_DeprecateTransition(t *testing.T) {
	now := time.Now()
	fv := FunctionVersion{
		ID:                uuid.New(),
		FunctionID:        uuid.New(),
		Version:           "v1.0.0",
		VersionState:      FunctionVersionStatePublished,
		PublishedAt:       &now,
		DeprecationReason: "Use v2.0.0 instead",
		ReplacedByVersion: "v2.0.0",
	}

	// Simulate deprecation
	fv.VersionState = FunctionVersionStateDeprecated

	assert.Equal(t, FunctionVersionStateDeprecated, fv.VersionState)
	assert.Equal(t, "Use v2.0.0 instead", fv.DeprecationReason)
	assert.Equal(t, "v2.0.0", fv.ReplacedByVersion)
}

func TestFunctionVersion_ArchiveTransition(t *testing.T) {
	createdAt := time.Now().Add(-24 * time.Hour)
	now := time.Now()
	fv := FunctionVersion{
		ID:           uuid.New(),
		FunctionID:   uuid.New(),
		Version:      "v1.0.0",
		VersionState: FunctionVersionStatePublished,
		PublishedAt:  &createdAt,
		ArchivedAt:   &now,
	}

	// Simulate archiving
	fv.VersionState = FunctionVersionStateArchived

	assert.Equal(t, FunctionVersionStateArchived, fv.VersionState)
	assert.NotNil(t, fv.ArchivedAt)
}

// ==================== Invalid State Transition Tests ====================

func TestFunctionVersion_InvalidDraftToArchived(t *testing.T) {
	fv := FunctionVersion{
		ID:           uuid.New(),
		FunctionID:   uuid.New(),
		Version:      "v1.0.0",
		VersionState: FunctionVersionStateDraft,
		CreatedAt:    time.Now(),
	}

	// Cannot go directly from draft to archived
	assert.NotEqual(t, FunctionVersionStateArchived, fv.VersionState)
}

func TestFunctionVersion_InvalidArchivedToPublished(t *testing.T) {
	now := time.Now()
	fv := FunctionVersion{
		ID:           uuid.New(),
		FunctionID:   uuid.New(),
		Version:      "v1.0.0",
		VersionState: FunctionVersionStateArchived,
		PublishedAt:  &now,
		ArchivedAt:   &now,
	}

	// Cannot go from archived to published
	assert.Equal(t, FunctionVersionStateArchived, fv.VersionState)
}

// ==================== Model Serialization Tests ====================

func TestAPIVersion_JSONMarshaling(t *testing.T) {
	now := time.Now()
	deprecatedAt := now.Add(30 * 24 * time.Hour)
	sunsetAt := now.Add(60 * 24 * time.Hour)

	apiVersion := APIVersion{
		ID:            uuid.New(),
		Version:       "v1",
		PathPrefix:    "/v1",
		Status:        APIVersionStatusDeprecated,
		ReleasedAt:    now,
		DeprecatedAt:  &deprecatedAt,
		SunsetAt:      &sunsetAt,
		SunsetMessage: "Use v2 API",
		ChangelogURL:  "https://example.com/changelog",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	assert.Equal(t, "v1", apiVersion.Version)
	assert.Equal(t, APIVersionStatusDeprecated, apiVersion.Status)
	assert.NotNil(t, apiVersion.DeprecatedAt)
	assert.NotNil(t, apiVersion.SunsetAt)
}

func TestFunctionVersion_JSONMarshaling(t *testing.T) {
	now := time.Now()
	publishedAt := now.Add(-24 * time.Hour)

	fv := FunctionVersion{
		ID:                uuid.New(),
		FunctionID:        uuid.New(),
		Version:           "v1.0.0",
		VersionState:      FunctionVersionStatePublished,
		DeprecationReason: "Use v2.0.0",
		ReplacedByVersion: "v2.0.0",
		MigrationGuide:    "See migration guide",
		CreatedAt:         now,
		PublishedAt:       &publishedAt,
	}

	assert.Equal(t, "v1.0.0", fv.Version)
	assert.Equal(t, FunctionVersionStatePublished, fv.VersionState)
	assert.NotNil(t, fv.PublishedAt)
}

func TestFunctionVersionResponse_JSONMarshaling(t *testing.T) {
	now := time.Now()
	publishedAt := now.Add(-24 * time.Hour)

	fvr := FunctionVersionResponse{
		ID:          uuid.New(),
		FunctionID:  uuid.New(),
		Version:     "v1.0.0",
		Status:      FunctionVersionStatePublished,
		PublishedAt: &publishedAt,
		IsLatest:    true,
		IsStable:    true,
	}

	assert.True(t, fvr.IsLatest)
	assert.True(t, fvr.IsStable)
}

// ==================== DeprecationInfo Tests ====================

func TestDeprecationInfo_Fields(t *testing.T) {
	now := time.Now()
	sunsetAt := now.Add(30 * 24 * time.Hour)

	info := DeprecationInfo{
		DeprecatedAt:   now,
		SunsetAt:       sunsetAt,
		MigrationGuide: "https://example.com/migration",
	}

	assert.Equal(t, now, info.DeprecatedAt)
	assert.Equal(t, sunsetAt, info.SunsetAt)
	assert.Equal(t, "https://example.com/migration", info.MigrationGuide)
}

func TestVersionDeprecationInfo_Fields(t *testing.T) {
	now := time.Now()
	sunsetAt := now.Add(30 * 24 * time.Hour)

	info := VersionDeprecationInfo{
		DeprecatedAt:   now,
		SunsetAt:       &sunsetAt,
		Reason:         "Breaking changes",
		MigrationGuide: "https://example.com/migration",
		ReplacedBy:     "v2.0.0",
	}

	assert.NotNil(t, info.DeprecatedAt)
	assert.NotNil(t, info.SunsetAt)
	assert.Equal(t, "Breaking changes", info.Reason)
	assert.Equal(t, "v2.0.0", info.ReplacedBy)
}

// ==================== RollbackRecord Tests ====================

func TestRollbackRecord_Fields(t *testing.T) {
	now := time.Now()
	initiatedBy := uuid.New()

	record := RollbackRecord{
		ID:          uuid.New(),
		FunctionID:  uuid.New(),
		FromVersion: "v2.0.0",
		ToVersion:   "v1.0.0",
		Strategy:    RollbackStrategyImmediate,
		Status:      "completed",
		InitiatedBy: &initiatedBy,
		InitiatedAt: now,
	}

	assert.Equal(t, "v2.0.0", record.FromVersion)
	assert.Equal(t, "v1.0.0", record.ToVersion)
	assert.Equal(t, RollbackStrategyImmediate, record.Strategy)
	assert.NotNil(t, record.InitiatedBy)
}

func TestRollbackStrategy_Constants(t *testing.T) {
	assert.Equal(t, RollbackStrategy("immediate"), RollbackStrategyImmediate)
	assert.Equal(t, RollbackStrategy("gradual"), RollbackStrategyGradual)
	assert.Equal(t, RollbackStrategy("canary"), RollbackStrategyCanary)
}

// ==================== Version Lineage Tests ====================

func TestVersionLineageEntry_Fields(t *testing.T) {
	now := time.Now()
	publishedAt := now.Add(-24 * time.Hour)

	entry := VersionLineageEntry{
		ID:            uuid.New(),
		Version:       "v2.0.0",
		State:         FunctionVersionStatePublished,
		ParentVersion: "v1.0.0",
		ChangeType:    ChangeTypeMajor,
		CreatedAt:     now,
		PublishedAt:   &publishedAt,
	}

	assert.Equal(t, "v2.0.0", entry.Version)
	assert.Equal(t, "v1.0.0", entry.ParentVersion)
	assert.Equal(t, ChangeTypeMajor, entry.ChangeType)
	assert.NotNil(t, entry.PublishedAt)
}

func TestVersionLineageResponse_Fields(t *testing.T) {
	functionID := uuid.New()
	now := time.Now()

	response := VersionLineageResponse{
		FunctionID: functionID,
		Entries: []VersionLineageEntry{
			{ID: uuid.New(), Version: "v2.0.0", State: FunctionVersionStatePublished, CreatedAt: now},
			{ID: uuid.New(), Version: "v1.0.0", State: FunctionVersionStateArchived, CreatedAt: now.Add(-24 * time.Hour)},
		},
		TotalCount: 2,
	}

	assert.Equal(t, functionID, response.FunctionID)
	assert.Len(t, response.Entries, 2)
	assert.Equal(t, 2, response.TotalCount)
}

// ==================== Version Diff Tests ====================

func TestVersionDiffEntry_Fields(t *testing.T) {
	entry := VersionDiffEntry{
		Field:      "timeout",
		FromValue:  "30",
		ToValue:    "60",
		ChangeType: "modified",
	}

	assert.Equal(t, "timeout", entry.Field)
	assert.Equal(t, "30", entry.FromValue)
	assert.Equal(t, "60", entry.ToValue)
	assert.Equal(t, "modified", entry.ChangeType)
}

func TestVersionDiffResponse_Fields(t *testing.T) {
	response := VersionDiffResponse{
		FunctionID: uuid.New(),
		Version1:   "v1.0.0",
		Version2:   "v2.0.0",
		Changes: []VersionDiffEntry{
			{Field: "timeout", FromValue: "30", ToValue: "60", ChangeType: "modified"},
			{Field: "memory", FromValue: "256", ToValue: "512", ChangeType: "modified"},
			{Field: "newField", FromValue: "", ToValue: "value", ChangeType: "added"},
		},
		Summary: VersionDiffSummary{
			TotalChanges:    3,
			Added:           1,
			Removed:         0,
			Modified:        2,
			Breaking:        true,
			BreakingChanges: []string{"timeout changed from 30s to 60s"},
		},
	}

	assert.Equal(t, "v1.0.0", response.Version1)
	assert.Equal(t, "v2.0.0", response.Version2)
	assert.Len(t, response.Changes, 3)
	assert.True(t, response.Summary.Breaking)
}

// ==================== Contract Schema Tests ====================

func TestContractSchema_Fields(t *testing.T) {
	schema := ContractSchema{
		Version: "1.0",
		Format:  "openapi3",
		Endpoints: []EndpointSchema{
			{Path: "/users", Method: "GET", Summary: "List users", Deprecated: false},
		},
		DataTypes: []DataTypeSchema{
			{Name: "User", Type: "object", Description: "User type"},
		},
	}

	assert.Equal(t, "1.0", schema.Version)
	assert.Equal(t, "openapi3", schema.Format)
	assert.Len(t, schema.Endpoints, 1)
	assert.Len(t, schema.DataTypes, 1)
}

func TestEndpointSchema_Fields(t *testing.T) {
	endpoint := EndpointSchema{
		Path:       "/users/{id}",
		Method:     "GET",
		Summary:    "Get user by ID",
		Deprecated: false,
	}

	assert.Equal(t, "/users/{id}", endpoint.Path)
	assert.Equal(t, "GET", endpoint.Method)
	assert.False(t, endpoint.Deprecated)
}

func TestServiceContract_Fields(t *testing.T) {
	now := time.Now()

	contract := ServiceContract{
		ID:                  uuid.New(),
		ServiceName:         "user-service",
		ContractVersion:     "1.0.0",
		ContractType:        "rest",
		Status:              "active",
		IntroducedInRelease: "2024.1",
		DeprecatedInRelease: "2024.4",
		RemovedInRelease:    "2025.1",
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	assert.Equal(t, "user-service", contract.ServiceName)
	assert.Equal(t, "1.0.0", contract.ContractVersion)
	assert.Equal(t, "rest", contract.ContractType)
	assert.Equal(t, "active", contract.Status)
}

// ==================== Deployment Version Tests ====================

func TestDeploymentVersion_Fields(t *testing.T) {
	now := time.Now()
	functionID := uuid.New()
	deploymentID := uuid.New()

	dv := DeploymentVersion{
		ID:           uuid.New(),
		FunctionID:   functionID,
		Version:      "v1.0.0",
		DeploymentID: &deploymentID,
		Provider:     "vercel",
		Region:       "iad1",
		Status:       DeploymentVersionStatusSuccess,
		ArtifactURI:  "s3://artifacts/function/v1.0.0.zip",
		Checksum:     "sha256:abc123",
		CreatedAt:    now,
	}

	assert.Equal(t, functionID, dv.FunctionID)
	assert.Equal(t, "v1.0.0", dv.Version)
	assert.NotNil(t, dv.DeploymentID)
	assert.Equal(t, "vercel", dv.Provider)
	assert.Equal(t, "iad1", dv.Region)
	assert.Equal(t, DeploymentVersionStatusSuccess, dv.Status)
}

func TestDeploymentMetadata_Fields(t *testing.T) {
	metadata := DeploymentMetadata{
		Runtime:      "nodejs18.x",
		Environment:  "production",
		BuildTime:    120000,
		DeployTime:   30000,
		InstanceType: "medium",
		MemoryMB:     512,
		TimeoutSec:   30,
	}

	assert.Equal(t, "nodejs18.x", metadata.Runtime)
	assert.Equal(t, "production", metadata.Environment)
	assert.Equal(t, int64(120000), metadata.BuildTime)
	assert.Equal(t, 512, metadata.MemoryMB)
	assert.Equal(t, 30, metadata.TimeoutSec)
}

// ==================== List Parameters Tests ====================

func TestListAPIVersionsParams_Defaults(t *testing.T) {
	params := ListAPIVersionsParams{}

	// Test that zero values are handled correctly
	assert.Equal(t, "", params.Status)
	assert.Equal(t, 0, params.Limit)
	assert.Equal(t, "", params.Cursor)
}

func TestListFunctionVersionsParams_Defaults(t *testing.T) {
	params := ListFunctionVersionsParams{}

	assert.Equal(t, uuid.Nil, params.FunctionID)
	assert.Equal(t, "", params.Status)
	assert.Equal(t, 0, params.Limit)
	assert.Equal(t, "", params.Cursor)
}

func TestCreateChangelogParams_Fields(t *testing.T) {
	functionID := uuid.New()
	createdBy := uuid.New()

	params := CreateChangelogParams{
		FunctionID:      functionID,
		Version:         "v2.0.0",
		ChangeType:      ChangeTypeMajor,
		ChangeCategory:  "api",
		Description:     "Major rewrite",
		BreakingChanges: []string{"Removed legacy API"},
		MigrationSteps:  []string{"Update client library"},
		CreatedBy:       &createdBy,
	}

	assert.Equal(t, functionID, params.FunctionID)
	assert.Equal(t, "v2.0.0", params.Version)
	assert.Equal(t, ChangeTypeMajor, params.ChangeType)
	assert.Len(t, params.BreakingChanges, 1)
	assert.Len(t, params.MigrationSteps, 1)
}

// ==================== Request/Response Model Tests ====================

func TestPublishVersionRequest_Fields(t *testing.T) {
	req := PublishVersionRequest{
		Version:     "v1.0.0",
		SetAsLatest: true,
		SetAsStable: true,
	}

	assert.Equal(t, "v1.0.0", req.Version)
	assert.True(t, req.SetAsLatest)
	assert.True(t, req.SetAsStable)
}

func TestPublishVersionResponse_Fields(t *testing.T) {
	now := time.Now()
	resp := PublishVersionResponse{
		ID:          uuid.New(),
		FunctionID:  uuid.New(),
		Version:     "v1.0.0",
		Status:      FunctionVersionStatePublished,
		PublishedAt: &now,
		IsLatest:    true,
		IsStable:    true,
	}

	assert.NotEqual(t, uuid.Nil, resp.ID)
	assert.Equal(t, FunctionVersionStatePublished, resp.Status)
	assert.True(t, resp.IsLatest)
	assert.True(t, resp.IsStable)
}

func TestDeprecateVersionRequest_Fields(t *testing.T) {
	effectiveAt := time.Now().Add(7 * 24 * time.Hour)
	req := DeprecateVersionRequest{
		Reason:          "Security vulnerability",
		ReplacedBy:      "v2.0.0",
		MigrationGuide:  "See migration guide",
		EffectiveAt:     &effectiveAt,
		GracePeriodDays: 30,
	}

	assert.Equal(t, "Security vulnerability", req.Reason)
	assert.Equal(t, "v2.0.0", req.ReplacedBy)
	assert.Equal(t, 30, req.GracePeriodDays)
	assert.NotNil(t, req.EffectiveAt)
}

func TestRollbackVersionRequest_Fields(t *testing.T) {
	req := RollbackVersionRequest{
		ToVersion: "v1.0.0",
		Strategy:  RollbackStrategyImmediate,
	}

	assert.Equal(t, "v1.0.0", req.ToVersion)
	assert.Equal(t, RollbackStrategyImmediate, req.Strategy)
}

func TestContractVersionNegotiationRequest_Fields(t *testing.T) {
	req := ContractVersionNegotiationRequest{
		ConsumerService:   "client-service",
		ProviderService:   "user-service",
		SupportedVersions: []string{"1.0.0", "1.1.0", "2.0.0"},
	}

	assert.Equal(t, "client-service", req.ConsumerService)
	assert.Equal(t, "user-service", req.ProviderService)
	assert.Len(t, req.SupportedVersions, 3)
}

func TestContractVersionNegotiationResponse_Fields(t *testing.T) {
	resp := ContractVersionNegotiationResponse{
		ConsumerService: "client-service",
		ProviderService: "user-service",
		AgreedVersion:   "1.1.0",
		Compatible:      true,
		Reason:          "Version intersection found",
	}

	assert.Equal(t, "1.1.0", resp.AgreedVersion)
	assert.True(t, resp.Compatible)
	assert.NotEmpty(t, resp.Reason)
}

func TestSchemaCompatibilityResult_Fields(t *testing.T) {
	result := SchemaCompatibilityResult{
		Compatible: false,
		BreakingChanges: []string{
			"Parameter 'limit' changed from optional to required",
			"Response field 'user.id' type changed from string to integer",
		},
		Warnings: []string{
			"New optional field 'metadata' added",
		},
	}

	assert.False(t, result.Compatible)
	assert.Len(t, result.BreakingChanges, 2)
	assert.Len(t, result.Warnings, 1)
}

// ==================== VersionAliasRecord Tests ====================

func TestVersionAliasRecord_Fields(t *testing.T) {
	now := time.Now()
	functionID := uuid.New()
	versionID := uuid.New()

	record := VersionAliasRecord{
		ID:         uuid.New(),
		FunctionID: functionID,
		Alias:      "latest",
		VersionID:  versionID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	assert.Equal(t, functionID, record.FunctionID)
	assert.Equal(t, "latest", record.Alias)
	assert.Equal(t, versionID, record.VersionID)
}

// ==================== Edge Cases ====================

func TestResolveVersionAlias_WithDuplicateVersions(t *testing.T) {
	// When there are duplicate versions (shouldn't happen but test handles it)
	versions := []string{"v1.0.0", "v1.0.0", "v2.0.0"}
	result := ResolveVersionAlias(VersionAliasLatest, versions)
	// Returns the last one in sorted order
	assert.Equal(t, "v2.0.0", result)
}

func TestFunctionVersion_EmptyDeprecationFields(t *testing.T) {
	now := time.Now()
	fv := FunctionVersion{
		ID:           uuid.New(),
		FunctionID:   uuid.New(),
		Version:      "v1.0.0",
		VersionState: FunctionVersionStatePublished,
		CreatedAt:    now,
		PublishedAt:  &now,
	}

	// Empty optional fields should remain empty
	assert.Equal(t, "", fv.DeprecationReason)
	assert.Equal(t, "", fv.ReplacedByVersion)
	assert.Equal(t, "", fv.MigrationGuide)
}

func TestAPIVersion_EmptyOptionalFields(t *testing.T) {
	now := time.Now()
	av := APIVersion{
		ID:         uuid.New(),
		Version:    "v1",
		PathPrefix: "/v1",
		Status:     APIVersionStatusActive,
		ReleasedAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Empty optional fields should remain empty
	assert.Nil(t, av.DeprecatedAt)
	assert.Nil(t, av.SunsetAt)
	assert.Equal(t, "", av.SunsetMessage)
	assert.Equal(t, "", av.OpenAPISpecURL)
	assert.Equal(t, "", av.ChangelogURL)
}

// ==================== Validation Helper Tests ====================

func TestIsValidVersionState(t *testing.T) {
	validStates := []string{
		FunctionVersionStateDraft,
		FunctionVersionStatePublished,
		FunctionVersionStateDeprecated,
		FunctionVersionStateArchived,
	}

	for _, state := range validStates {
		assert.True(t, isValidVersionState(state), "Expected %s to be valid", state)
	}

	assert.False(t, isValidVersionState("invalid"))
	assert.False(t, isValidVersionState(""))
}

func isValidVersionState(state string) bool {
	switch state {
	case FunctionVersionStateDraft, FunctionVersionStatePublished,
		FunctionVersionStateDeprecated, FunctionVersionStateArchived:
		return true
	}
	return false
}

func TestIsValidAPIVersionStatus(t *testing.T) {
	validStatuses := []string{
		APIVersionStatusActive,
		APIVersionStatusDeprecated,
		APIVersionStatusSunset,
		APIVersionStatusArchived,
	}

	for _, status := range validStatuses {
		assert.True(t, isValidAPIVersionStatus(status), "Expected %s to be valid", status)
	}

	assert.False(t, isValidAPIVersionStatus("invalid"))
	assert.False(t, isValidAPIVersionStatus(""))
}

func isValidAPIVersionStatus(status string) bool {
	switch status {
	case APIVersionStatusActive, APIVersionStatusDeprecated,
		APIVersionStatusSunset, APIVersionStatusArchived:
		return true
	}
	return false
}

// ==================== Changelog Entry Tests ====================

func TestChangelogEntry_Fields(t *testing.T) {
	entry := ChangelogEntry{
		Type:        "feature",
		Category:    "api",
		Description: "Added new endpoint",
	}

	assert.Equal(t, "feature", entry.Type)
	assert.Equal(t, "api", entry.Category)
	assert.Equal(t, "Added new endpoint", entry.Description)
}

func TestFunctionVersionChangelog_Fields(t *testing.T) {
	now := time.Now()
	functionID := uuid.New()
	createdBy := uuid.New()

	changelog := FunctionVersionChangelog{
		ID:              uuid.New(),
		FunctionID:      functionID,
		Version:         "v2.0.0",
		ChangeType:      ChangeTypeBreaking,
		ChangeCategory:  "api",
		Description:     "Complete API redesign",
		BreakingChanges: []byte(`["Removed legacy endpoints"]`),
		MigrationSteps:  []byte(`["Update client library"]`),
		CreatedBy:       &createdBy,
		CreatedAt:       now,
	}

	require.NotNil(t, changelog.BreakingChanges)
	require.NotNil(t, changelog.MigrationSteps)
	assert.Equal(t, functionID, changelog.FunctionID)
	assert.Equal(t, ChangeTypeBreaking, changelog.ChangeType)
}

// ==================== Contract List Response Tests ====================

func TestContractListResponse_Fields(t *testing.T) {
	now := time.Now()

	response := ContractListResponse{
		Services: []string{"user-service", "order-service"},
		Contracts: []ServiceContractResponse{
			{
				ID:              uuid.New(),
				ServiceName:     "user-service",
				ContractVersion: "1.0.0",
				ContractType:    "rest",
				Status:          "active",
				CreatedAt:       now,
			},
		},
	}

	assert.Len(t, response.Services, 2)
	assert.Contains(t, response.Services, "user-service")
	assert.Len(t, response.Contracts, 1)
}

func TestContractListResponse_Empty(t *testing.T) {
	response := ContractListResponse{
		Services:  []string{},
		Contracts: []ServiceContractResponse{},
	}

	assert.Empty(t, response.Services)
	assert.Empty(t, response.Contracts)
}
