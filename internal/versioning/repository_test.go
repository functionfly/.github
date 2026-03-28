// Package versioning provides tests for version repository.
package versioning

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Repository Creation Tests ====================

func TestNewRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.db)
}

func TestNewRepository_NilDB(t *testing.T) {
	repo := NewRepository(nil)
	assert.NotNil(t, repo)
	assert.Nil(t, repo.db)
}

// ==================== API Version Repository Tests ====================

func TestCreateAPIVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	now := time.Now()
	apiVersion := &APIVersion{
		ID:             uuid.New(),
		Version:        "v1",
		PathPrefix:     "/v1",
		Status:         APIVersionStatusActive,
		ReleasedAt:     now,
		DeprecatedAt:   nil,
		SunsetAt:       nil,
		SunsetMessage:  "",
		Metadata:       nil,
		OpenAPISpecURL: "",
		ChangelogURL:   "",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	mock.ExpectExec("INSERT INTO api_versions").
		WithArgs(
			apiVersion.ID, apiVersion.Version, apiVersion.PathPrefix, apiVersion.Status,
			apiVersion.ReleasedAt, apiVersion.DeprecatedAt, apiVersion.SunsetAt, apiVersion.SunsetMessage,
			apiVersion.Metadata, apiVersion.OpenAPISpecURL, apiVersion.ChangelogURL,
			apiVersion.CreatedAt, apiVersion.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateAPIVersion(context.Background(), apiVersion)
	assert.NoError(t, err)
}

func TestCreateAPIVersion_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	now := time.Now()
	apiVersion := &APIVersion{
		ID:         uuid.New(),
		Version:    "v1",
		PathPrefix: "/v1",
		Status:     APIVersionStatusActive,
		ReleasedAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	mock.ExpectExec("INSERT INTO api_versions").
		WillReturnError(assert.AnError)

	err = repo.CreateAPIVersion(context.Background(), apiVersion)
	assert.Error(t, err)
}

func TestGetAPIVersionByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	versionID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "version", "path_prefix", "status", "is_default", "released_at",
		"deprecated_at", "sunset_at", "sunset_message", "metadata",
		"openapi_spec_url", "changelog_url", "created_at", "updated_at",
	}).AddRow(
		versionID, "v1", "/v1", "active", true, now,
		nil, nil, "", []byte(`{}`), "", "", now, now,
	)

	mock.ExpectQuery("SELECT .* FROM api_versions WHERE id = \\$1").
		WithArgs(versionID).
		WillReturnRows(rows)

	result, err := repo.GetAPIVersionByID(context.Background(), versionID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "v1", result.Version)
}

func TestGetAPIVersionByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	versionID := uuid.New()

	mock.ExpectQuery("SELECT .* FROM api_versions WHERE id = \\$1").
		WithArgs(versionID).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetAPIVersionByID(context.Background(), versionID)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestGetAPIVersionByVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	versionID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "version", "path_prefix", "status", "is_default", "released_at",
		"deprecated_at", "sunset_at", "sunset_message", "metadata",
		"openapi_spec_url", "changelog_url", "created_at", "updated_at",
	}).AddRow(versionID, "v1", "/v1", "active", false, now, nil, nil, "", []byte(`{}`), "", "", now, now)

	mock.ExpectQuery("SELECT .* FROM api_versions WHERE version = \\$1").
		WithArgs("v1").
		WillReturnRows(rows)

	result, err := repo.GetAPIVersionByVersion(context.Background(), "v1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "v1", result.Version)
}

func TestListAPIVersions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "version", "path_prefix", "status", "is_default", "released_at",
		"deprecated_at", "sunset_at", "sunset_message", "metadata",
		"openapi_spec_url", "changelog_url", "created_at", "updated_at",
	}).AddRow(uuid.New(), "v2", "/v2", "active", false, now, nil, nil, "", []byte(`{}`), "", "", now, now).
		AddRow(uuid.New(), "v1", "/v1", "deprecated", true, now, nil, nil, "", []byte(`{}`), "", "", now, now)

	mock.ExpectQuery("SELECT .* FROM api_versions").
		WillReturnRows(rows)

	result, err := repo.ListAPIVersions(context.Background(), ListAPIVersionsParams{
		Status: "",
		Limit:  20,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestListAPIVersions_WithStatusFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "version", "path_prefix", "status", "is_default", "released_at",
		"deprecated_at", "sunset_at", "sunset_message", "metadata",
		"openapi_spec_url", "changelog_url", "created_at", "updated_at",
	}).AddRow(uuid.New(), "v1", "/v1", "deprecated", false, now, nil, nil, "", []byte(`{}`), "", "", now, now)

	mock.ExpectQuery("SELECT .* FROM api_versions WHERE \\(\\$1::text IS NULL OR status = \\$1\\)").
		WithArgs("deprecated", 20).
		WillReturnRows(rows)

	result, err := repo.ListAPIVersions(context.Background(), ListAPIVersionsParams{
		Status: "deprecated",
		Limit:  20,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestUpdateAPIVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	versionID := uuid.New()
	now := time.Now()

	apiVersion := &APIVersion{
		ID:            versionID,
		PathPrefix:    "/v1",
		Status:        APIVersionStatusDeprecated,
		DeprecatedAt:  &now,
		SunsetAt:      &now,
		SunsetMessage: "Use v2",
		UpdatedAt:     now,
	}

	mock.ExpectExec("UPDATE api_versions SET").
		WithArgs(
			apiVersion.PathPrefix, apiVersion.Status, apiVersion.DeprecatedAt,
			apiVersion.SunsetAt, apiVersion.SunsetMessage, apiVersion.Metadata,
			apiVersion.OpenAPISpecURL, apiVersion.ChangelogURL, apiVersion.UpdatedAt,
			apiVersion.ID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateAPIVersion(context.Background(), apiVersion)
	assert.NoError(t, err)
}

func TestDeprecateAPIVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	versionID := uuid.New()
	sunsetAt := time.Now().Add(30 * 24 * time.Hour)

	mock.ExpectExec("UPDATE api_versions SET status = \\$2, deprecated_at = NOW(), sunset_at = \\$3, sunset_message = \\$4, updated_at = NOW()").
		WithArgs(versionID, APIVersionStatusDeprecated, &sunsetAt, "Use v2 instead").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.DeprecateAPIVersion(context.Background(), versionID, &sunsetAt, "Use v2 instead")
	assert.NoError(t, err)
}

// ==================== Function Version Repository Tests ====================

func TestGetLatestFunctionVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	functionID := uuid.New()
	versionID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "function_id", "version", "version_state", "deprecation_reason",
		"replaced_by_version", "migration_guide", "created_at", "published_at", "archived_at",
	}).AddRow(
		versionID, functionID, "v1.0.0", "published", "",
		"", "", now, &now, nil,
	)

	mock.ExpectQuery("SELECT .* FROM registry_function_versions WHERE function_id = \\$1 AND version_state = 'published'").
		WithArgs(functionID).
		WillReturnRows(rows)

	result, err := repo.GetLatestFunctionVersion(context.Background(), functionID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "v1.0.0", result.Version)
}

func TestGetLatestFunctionVersion_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	functionID := uuid.New()

	mock.ExpectQuery("SELECT .* FROM registry_function_versions WHERE function_id = \\$1 AND version_state = 'published'").
		WithArgs(functionID).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetLatestFunctionVersion(context.Background(), functionID)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestListFunctionVersions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	functionID := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "function_id", "version", "version_state",
	}).AddRow(uuid.New(), functionID, "v1.0.0", "published").
		AddRow(uuid.New(), functionID, "v0.9.0", "draft")

	mock.ExpectQuery("SELECT .* FROM registry_function_versions WHERE function_id = \\$1").
		WillReturnRows(rows)

	result, err := repo.ListFunctionVersions(context.Background(), ListFunctionVersionsParams{
		FunctionID: functionID,
		Limit:      20,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestUpdateFunctionVersionState(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	versionID := uuid.New()

	mock.ExpectExec("UPDATE registry_function_versions SET version_state = \\$2").
		WithArgs(versionID, FunctionVersionStateDeprecated).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateFunctionVersionState(context.Background(), versionID, FunctionVersionStateDeprecated)
	assert.NoError(t, err)
}

func TestDeprecateFunctionVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	versionID := uuid.New()

	mock.ExpectExec("UPDATE registry_function_versions SET version_state = \\$2, deprecation_reason = \\$3, replaced_by_version = \\$4, migration_guide = \\$5").
		WithArgs(versionID, FunctionVersionStateDeprecated, "Use v2.0.0", "v2.0.0", "See migration guide").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.DeprecateFunctionVersion(context.Background(), versionID, "Use v2.0.0", "v2.0.0", "See migration guide")
	assert.NoError(t, err)
}

// ==================== Publish and Archive Tests ====================

func TestPublishFunctionVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	versionID := uuid.New()
	functionID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "function_id", "version", "version_state", "deprecation_reason",
		"replaced_by_version", "migration_guide", "created_at", "published_at", "archived_at",
	}).AddRow(
		versionID, functionID, "v1.0.0", "published", "",
		"", "", now, &now, nil,
	)

	mock.ExpectQuery("UPDATE registry_function_versions SET version_state = \\$2, published_at = \\$3").
		WithArgs(versionID, FunctionVersionStatePublished, now).
		WillReturnRows(rows)

	result, err := repo.PublishFunctionVersion(context.Background(), versionID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, FunctionVersionStatePublished, result.VersionState)
}

func TestPublishFunctionVersion_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	versionID := uuid.New()

	mock.ExpectQuery("UPDATE registry_function_versions SET version_state = \\$2, published_at = \\$3").
		WillReturnError(sql.ErrNoRows)

	result, err := repo.PublishFunctionVersion(context.Background(), versionID)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestArchiveFunctionVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	versionID := uuid.New()
	functionID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "function_id", "version", "version_state",
	}).AddRow(versionID, functionID, "v1.0.0", "archived")

	mock.ExpectQuery("UPDATE registry_function_versions SET version_state = \\$2, archived_at = \\$3").
		WithArgs(versionID, FunctionVersionStateArchived, now).
		WillReturnRows(rows)

	result, err := repo.ArchiveFunctionVersion(context.Background(), versionID, "No longer needed")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, FunctionVersionStateArchived, result.VersionState)
}

func TestGetFunctionVersionByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	versionID := uuid.New()
	functionID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "function_id", "version", "version_state", "deprecation_reason", "replaced_by_version", "migration_guide", "created_at", "published_at", "archived_at",
	}).AddRow(versionID, functionID, "v1.0.0", "published", "", "", "", now, &now, nil)

	mock.ExpectQuery("SELECT .* FROM registry_function_versions WHERE id = \\$1").
		WithArgs(versionID).
		WillReturnRows(rows)

	result, err := repo.GetFunctionVersionByID(context.Background(), versionID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "v1.0.0", result.Version)
	assert.Equal(t, now.Truncate(time.Second), result.CreatedAt.Truncate(time.Second))
}

func TestGetFunctionVersionByVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	versionID := uuid.New()
	functionID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "function_id", "version", "version_state", "deprecation_reason", "replaced_by_version", "migration_guide", "created_at", "published_at", "archived_at",
	}).AddRow(versionID, functionID, "v1.0.0", "published", "", "", "", now, &now, nil)

	mock.ExpectQuery("SELECT .* FROM registry_function_versions WHERE function_id = \\$1 AND version = \\$2").
		WithArgs(functionID, "v1.0.0").
		WillReturnRows(rows)

	result, err := repo.GetFunctionVersionByVersion(context.Background(), functionID, "v1.0.0")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "v1.0.0", result.Version)
	assert.Equal(t, now.Truncate(time.Second), result.CreatedAt.Truncate(time.Second))
}

// ==================== Alias Tests ====================

func TestSetVersionAlias(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	functionID := uuid.New()
	aliasID := uuid.New()
	versionID := uuid.New()
	now := time.Now()

	mock.ExpectExec("INSERT INTO version_aliases").
		WithArgs(aliasID, functionID, "latest", versionID, now, now).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SetVersionAlias(context.Background(), functionID, "latest", versionID)
	assert.NoError(t, err)
}

func TestGetVersionAlias(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	functionID := uuid.New()
	versionID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "function_id", "version", "version_state", "deprecation_reason", "replaced_by_version", "migration_guide", "created_at", "published_at", "archived_at",
	}).AddRow(versionID, functionID, "v1.0.0", "published", "", "", "", now, &now, nil)

	mock.ExpectQuery("SELECT .* FROM registry_function_versions v").
		WithArgs(functionID, "latest").
		WillReturnRows(rows)

	result, err := repo.GetVersionAlias(context.Background(), functionID, "latest")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "v1.0.0", result.Version)
	assert.Equal(t, versionID, result.ID)
}

func TestGetVersionAlias_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	functionID := uuid.New()

	mock.ExpectQuery("SELECT .* FROM version_aliases WHERE function_id = \\$1 AND alias = \\$2").
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetVersionAlias(context.Background(), functionID, "latest")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

// ==================== Changelog Tests ====================

func TestCreateChangelog(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	functionID := uuid.New()
	changelogID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "function_id", "version", "change_type", "change_category",
		"description", "breaking_changes", "migration_steps", "created_by", "created_at",
	}).AddRow(changelogID, functionID, "v2.0.0", "major", "api", "Major update", "[]", "[]", nil, now)

	mock.ExpectQuery("INSERT INTO function_version_changelog").
		WillReturnRows(rows)

	params := CreateChangelogParams{
		FunctionID:     functionID,
		Version:        "v2.0.0",
		ChangeType:     ChangeTypeMajor,
		ChangeCategory: "api",
		Description:    "Major update",
	}

	result, err := repo.CreateChangelog(context.Background(), params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGetChangelogByFunctionID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	functionID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "function_id", "version", "change_type", "change_category", "description", "breaking_changes", "migration_steps", "created_by", "created_at",
	}).AddRow(uuid.New(), functionID, "v2.0.0", "major", "", "", nil, nil, nil, now).
		AddRow(uuid.New(), functionID, "v1.0.0", "minor", "", "", nil, nil, nil, now)

	mock.ExpectQuery("SELECT .* FROM function_version_changelog WHERE function_id = \\$1").
		WithArgs(functionID).
		WillReturnRows(rows)

	result, err := repo.GetChangelogByFunctionID(context.Background(), functionID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// ==================== Deployment Version Tests ====================

func TestCreateDeploymentVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	dv := &DeploymentVersion{
		ID:         uuid.New(),
		FunctionID: uuid.New(),
		Version:    "v1.0.0",
		Provider:   "vercel",
		Region:     "iad1",
		Status:     DeploymentVersionStatusSuccess,
		CreatedAt:  time.Now(),
	}

	mock.ExpectExec("INSERT INTO deployment_versions").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateDeploymentVersion(context.Background(), dv)
	assert.NoError(t, err)
}

func TestGetDeploymentVersionsByFunctionID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	functionID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "function_id", "function_version", "deployment_id", "provider", "region", "status", "artifact_uri", "checksum", "rollback_id", "metadata", "created_at", "completed_at",
	}).AddRow(uuid.New(), functionID, "v1.0.0", nil, "vercel", "", "success", "", "", nil, nil, now, &now).
		AddRow(uuid.New(), functionID, "v0.9.0", nil, "cloudflare", "", "success", "", "", nil, nil, now, &now)

	mock.ExpectQuery("SELECT .* FROM deployment_versions WHERE function_id = \\$1").
		WithArgs(functionID).
		WillReturnRows(rows)

	result, err := repo.GetDeploymentVersionsByFunctionID(context.Background(), functionID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, now.Truncate(time.Second), result[0].CreatedAt.Truncate(time.Second))
}

func TestUpdateDeploymentVersionStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	deploymentID := uuid.New()

	mock.ExpectExec("UPDATE deployment_versions SET status = \\$2, completed_at = CASE").
		WithArgs(deploymentID, DeploymentVersionStatusSuccess).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateDeploymentVersionStatus(context.Background(), deploymentID, DeploymentVersionStatusSuccess)
	assert.NoError(t, err)
}

// ==================== Service Contract Tests ====================

func TestCreateServiceContract(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	sc := &ServiceContract{
		ID:              uuid.New(),
		ServiceName:     "user-service",
		ContractVersion: "1.0.0",
		ContractType:    "rest",
		Status:          "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	mock.ExpectExec("INSERT INTO service_contracts").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateServiceContract(context.Background(), sc)
	assert.NoError(t, err)
}

func TestGetServiceContractsByServiceName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "service_name", "contract_version", "contract_type", "schema", "status", "introduced_in_release", "deprecated_in_release", "removed_in_release", "created_at", "updated_at",
	}).AddRow(uuid.New(), "user-service", "1.0.0", "rest", nil, "active", "", "", "", now, now).
		AddRow(uuid.New(), "user-service", "0.9.0", "rest", nil, "deprecated", "", "", "", now, now)

	mock.ExpectQuery("SELECT .* FROM service_contracts WHERE service_name = \\$1").
		WithArgs("user-service").
		WillReturnRows(rows)

	result, err := repo.GetServiceContractsByServiceName(context.Background(), "user-service")
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, now.Truncate(time.Second), result[0].CreatedAt.Truncate(time.Second))
}
