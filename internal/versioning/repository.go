// Package versioning provides repository layer for version management database operations.
package versioning

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Repository handles version management database operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new version repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// ==================== API Versions ====================

// CreateAPIVersion creates a new API version
func (r *Repository) CreateAPIVersion(ctx context.Context, v *APIVersion) error {
	query := `
		INSERT INTO api_versions (id, version, path_prefix, status, released_at, deprecated_at, sunset_at, sunset_message, metadata, openapi_spec_url, changelog_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		v.ID, v.Version, v.PathPrefix, v.Status, v.ReleasedAt, v.DeprecatedAt, v.SunsetAt, v.SunsetMessage, v.Metadata, v.OpenAPISpecURL, v.ChangelogURL, v.CreatedAt, v.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create API version: %w", err)
	}
	return nil
}

// GetAPIVersionByID retrieves an API version by ID
func (r *Repository) GetAPIVersionByID(ctx context.Context, id uuid.UUID) (*APIVersion, error) {
	query := `
		SELECT id, version, path_prefix, status, released_at, deprecated_at, sunset_at, sunset_message, metadata, openapi_spec_url, changelog_url, created_at, updated_at
		FROM api_versions WHERE id = $1
	`
	var v APIVersion
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&v.ID, &v.Version, &v.PathPrefix, &v.Status, &v.ReleasedAt, &v.DeprecatedAt, &v.SunsetAt, &v.SunsetMessage, &v.Metadata, &v.OpenAPISpecURL, &v.ChangelogURL, &v.CreatedAt, &v.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API version: %w", err)
	}
	return &v, nil
}

// GetAPIVersionByVersion retrieves an API version by version string
func (r *Repository) GetAPIVersionByVersion(ctx context.Context, version string) (*APIVersion, error) {
	query := `
		SELECT id, version, path_prefix, status, released_at, deprecated_at, sunset_at, sunset_message, metadata, openapi_spec_url, changelog_url, created_at, updated_at
		FROM api_versions WHERE version = $1
	`
	var v APIVersion
	err := r.db.QueryRowContext(ctx, query, version).Scan(
		&v.ID, &v.Version, &v.PathPrefix, &v.Status, &v.ReleasedAt, &v.DeprecatedAt, &v.SunsetAt, &v.SunsetMessage, &v.Metadata, &v.OpenAPISpecURL, &v.ChangelogURL, &v.CreatedAt, &v.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API version: %w", err)
	}
	return &v, nil
}

// ListAPIVersions retrieves API versions with optional filtering
func (r *Repository) ListAPIVersions(ctx context.Context, params ListAPIVersionsParams) ([]APIVersion, error) {
	query := `
		SELECT id, version, path_prefix, status, released_at, deprecated_at, sunset_at, sunset_message, metadata, openapi_spec_url, changelog_url, created_at, updated_at
		FROM api_versions
		WHERE ($1::text IS NULL OR status = $1)
		ORDER BY released_at DESC
		LIMIT $2
	`
	limit := params.Limit
	if limit == 0 {
		limit = 20
	}

	rows, err := r.db.QueryContext(ctx, query, params.Status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list API versions: %w", err)
	}
	defer rows.Close()

	var versions []APIVersion
	for rows.Next() {
		var v APIVersion
		err := rows.Scan(
			&v.ID, &v.Version, &v.PathPrefix, &v.Status, &v.ReleasedAt, &v.DeprecatedAt, &v.SunsetAt, &v.SunsetMessage, &v.Metadata, &v.OpenAPISpecURL, &v.ChangelogURL, &v.CreatedAt, &v.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API version: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, nil
}

// UpdateAPIVersion updates an existing API version
func (r *Repository) UpdateAPIVersion(ctx context.Context, v *APIVersion) error {
	query := `
		UPDATE api_versions SET
			path_prefix = $2, status = $3, deprecated_at = $4, sunset_at = $5, sunset_message = $6,
			metadata = $7, openapi_spec_url = $8, changelog_url = $9, updated_at = $10
		WHERE id = $1
	`
	v.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query,
		v.ID, v.PathPrefix, v.Status, v.DeprecatedAt, v.SunsetAt, v.SunsetMessage, v.Metadata, v.OpenAPISpecURL, v.ChangelogURL, v.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update API version: %w", err)
	}
	return nil
}

// DeprecateAPIVersion marks an API version as deprecated
func (r *Repository) DeprecateAPIVersion(ctx context.Context, id uuid.UUID, sunsetAt *time.Time, sunsetMessage string) error {
	query := `
		UPDATE api_versions SET status = $2, deprecated_at = NOW(), sunset_at = $3, sunset_message = $4, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, APIVersionStatusDeprecated, sunsetAt, sunsetMessage)
	if err != nil {
		return fmt.Errorf("failed to deprecate API version: %w", err)
	}
	return nil
}

// ==================== Function Versions ====================

// GetLatestFunctionVersion retrieves the latest version for a function
func (r *Repository) GetLatestFunctionVersion(ctx context.Context, functionID uuid.UUID) (*FunctionVersion, error) {
	query := `
		SELECT id, function_id, version, version_state, deprecation_reason, replaced_by_version, migration_guide, created_at, published_at, archived_at
		FROM registry_function_versions
		WHERE function_id = $1 AND version_state = 'published'
		ORDER BY created_at DESC
		LIMIT 1
	`
	var v FunctionVersion
	err := r.db.QueryRowContext(ctx, query, functionID).Scan(
		&v.ID, &v.FunctionID, &v.Version, &v.VersionState, &v.DeprecationReason, &v.ReplacedByVersion, &v.MigrationGuide, &v.CreatedAt, &v.PublishedAt, &v.ArchivedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest function version: %w", err)
	}
	return &v, nil
}

// ListFunctionVersions retrieves versions for a function
func (r *Repository) ListFunctionVersions(ctx context.Context, params ListFunctionVersionsParams) ([]FunctionVersion, error) {
	query := `
		SELECT id, function_id, version, version_state, deprecation_reason, replaced_by_version, migration_guide, created_at, published_at, archived_at
		FROM registry_function_versions
		WHERE function_id = $1 AND ($2::text IS NULL OR version_state = $2)
		ORDER BY created_at DESC
		LIMIT $3
	`
	limit := params.Limit
	if limit == 0 {
		limit = 20
	}

	rows, err := r.db.QueryContext(ctx, query, params.FunctionID, params.Status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list function versions: %w", err)
	}
	defer rows.Close()

	var versions []FunctionVersion
	for rows.Next() {
		var v FunctionVersion
		err := rows.Scan(
			&v.ID, &v.FunctionID, &v.Version, &v.VersionState, &v.DeprecationReason, &v.ReplacedByVersion, &v.MigrationGuide, &v.CreatedAt, &v.PublishedAt, &v.ArchivedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan function version: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, nil
}

// UpdateFunctionVersionState updates the state of a function version
func (r *Repository) UpdateFunctionVersionState(ctx context.Context, id uuid.UUID, state string) error {
	query := `
		UPDATE registry_function_versions SET version_state = $2 WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, state)
	if err != nil {
		return fmt.Errorf("failed to update function version state: %w", err)
	}
	return nil
}

// DeprecateFunctionVersion marks a function version as deprecated
func (r *Repository) DeprecateFunctionVersion(ctx context.Context, id uuid.UUID, reason, replacedByVersion, migrationGuide string) error {
	query := `
		UPDATE registry_function_versions SET
			version_state = $2, deprecation_reason = $3, replaced_by_version = $4, migration_guide = $5
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, FunctionVersionStateDeprecated, reason, replacedByVersion, migrationGuide)
	if err != nil {
		return fmt.Errorf("failed to deprecate function version: %w", err)
	}
	return nil
}

// ==================== Function Version Changelog ====================

// CreateChangelog creates a new changelog entry
func (r *Repository) CreateChangelog(ctx context.Context, params CreateChangelogParams) (*FunctionVersionChangelog, error) {
	id := uuid.New()
	now := time.Now()

	breakingChangesJSON, _ := json.Marshal(params.BreakingChanges)
	migrationStepsJSON, _ := json.Marshal(params.MigrationSteps)

	query := `
		INSERT INTO function_version_changelog (id, function_id, version, change_type, change_category, description, breaking_changes, migration_steps, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (function_id, version) DO UPDATE SET
			change_type = EXCLUDED.change_type,
			change_category = EXCLUDED.change_category,
			description = EXCLUDED.description,
			breaking_changes = EXCLUDED.breaking_changes,
			migration_steps = EXCLUDED.migration_steps,
			created_by = EXCLUDED.created_by,
			created_at = EXCLUDED.created_at
		RETURNING id, function_id, version, change_type, change_category, description, breaking_changes, migration_steps, created_by, created_at
	`

	var c FunctionVersionChangelog
	err := r.db.QueryRowContext(ctx, query,
		id, params.FunctionID, params.Version, params.ChangeType, params.ChangeCategory, params.Description,
		breakingChangesJSON, migrationStepsJSON, params.CreatedBy, now,
	).Scan(
		&c.ID, &c.FunctionID, &c.Version, &c.ChangeType, &c.ChangeCategory, &c.Description,
		&c.BreakingChanges, &c.MigrationSteps, &c.CreatedBy, &c.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create changelog: %w", err)
	}
	return &c, nil
}

// GetChangelogByFunctionID retrieves changelog entries for a function
func (r *Repository) GetChangelogByFunctionID(ctx context.Context, functionID uuid.UUID) ([]FunctionVersionChangelog, error) {
	query := `
		SELECT id, function_id, version, change_type, change_category, description, breaking_changes, migration_steps, created_by, created_at
		FROM function_version_changelog
		WHERE function_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, functionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get changelog: %w", err)
	}
	defer rows.Close()

	var changelogs []FunctionVersionChangelog
	for rows.Next() {
		var c FunctionVersionChangelog
		err := rows.Scan(
			&c.ID, &c.FunctionID, &c.Version, &c.ChangeType, &c.ChangeCategory, &c.Description,
			&c.BreakingChanges, &c.MigrationSteps, &c.CreatedBy, &c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan changelog: %w", err)
		}
		changelogs = append(changelogs, c)
	}
	return changelogs, nil
}

// ==================== Deployment Versions ====================

// CreateDeploymentVersion creates a new deployment version record
func (r *Repository) CreateDeploymentVersion(ctx context.Context, dv *DeploymentVersion) error {
	query := `
		INSERT INTO deployment_versions (id, function_id, function_version, deployment_id, provider, region, status, artifact_uri, checksum, rollback_id, metadata, created_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		dv.ID, dv.FunctionID, dv.Version, dv.DeploymentID, dv.Provider, dv.Region, dv.Status,
		dv.ArtifactURI, dv.Checksum, dv.RollbackID, dv.Metadata, dv.CreatedAt, dv.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create deployment version: %w", err)
	}
	return nil
}

// GetDeploymentVersionsByFunctionID retrieves deployment versions for a function
func (r *Repository) GetDeploymentVersionsByFunctionID(ctx context.Context, functionID uuid.UUID) ([]DeploymentVersion, error) {
	query := `
		SELECT id, function_id, function_version, deployment_id, provider, region, status, artifact_uri, checksum, rollback_id, metadata, created_at, completed_at
		FROM deployment_versions
		WHERE function_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, functionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment versions: %w", err)
	}
	defer rows.Close()

	var versions []DeploymentVersion
	for rows.Next() {
		var dv DeploymentVersion
		err := rows.Scan(
			&dv.ID, &dv.FunctionID, &dv.Version, &dv.DeploymentID, &dv.Provider, &dv.Region, &dv.Status,
			&dv.ArtifactURI, &dv.Checksum, &dv.RollbackID, &dv.Metadata, &dv.CreatedAt, &dv.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment version: %w", err)
		}
		versions = append(versions, dv)
	}
	return versions, nil
}

// UpdateDeploymentVersionStatus updates the status of a deployment version
func (r *Repository) UpdateDeploymentVersionStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE deployment_versions SET status = $2, completed_at = CASE WHEN $2 IN ('success', 'failed', 'rolled_back') THEN NOW() ELSE completed_at END
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update deployment version status: %w", err)
	}
	return nil
}

// ==================== Service Contracts ====================

// CreateServiceContract creates a new service contract
func (r *Repository) CreateServiceContract(ctx context.Context, sc *ServiceContract) error {
	query := `
		INSERT INTO service_contracts (id, service_name, contract_version, contract_type, schema, status, introduced_in_release, deprecated_in_release, removed_in_release, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.ExecContext(ctx, query,
		sc.ID, sc.ServiceName, sc.ContractVersion, sc.ContractType, sc.Schema, sc.Status,
		sc.IntroducedInRelease, sc.DeprecatedInRelease, sc.RemovedInRelease, sc.CreatedAt, sc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create service contract: %w", err)
	}
	return nil
}

// GetServiceContractsByServiceName retrieves service contracts for a service
func (r *Repository) GetServiceContractsByServiceName(ctx context.Context, serviceName string) ([]ServiceContract, error) {
	query := `
		SELECT id, service_name, contract_version, contract_type, schema, status, introduced_in_release, deprecated_in_release, removed_in_release, created_at, updated_at
		FROM service_contracts
		WHERE service_name = $1
		ORDER BY contract_version DESC
	`
	rows, err := r.db.QueryContext(ctx, query, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get service contracts: %w", err)
	}
	defer rows.Close()

	var contracts []ServiceContract
	for rows.Next() {
		var sc ServiceContract
		err := rows.Scan(
			&sc.ID, &sc.ServiceName, &sc.ContractVersion, &sc.ContractType, &sc.Schema, &sc.Status,
			&sc.IntroducedInRelease, &sc.DeprecatedInRelease, &sc.RemovedInRelease, &sc.CreatedAt, &sc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service contract: %w", err)
		}
		contracts = append(contracts, sc)
	}
	return contracts, nil
}

// ==================== Phase 2: Publishing and Aliases ====================

// PublishFunctionVersion publishes a function version (draft -> published)
func (r *Repository) PublishFunctionVersion(ctx context.Context, id uuid.UUID) (*FunctionVersion, error) {
	now := time.Now()
	query := `
		UPDATE registry_function_versions
		SET version_state = $2, published_at = $3
		WHERE id = $1
		RETURNING id, function_id, version, version_state, deprecation_reason, replaced_by_version, migration_guide, created_at, published_at, archived_at
	`
	var v FunctionVersion
	err := r.db.QueryRowContext(ctx, query, id, FunctionVersionStatePublished, now).Scan(
		&v.ID, &v.FunctionID, &v.Version, &v.VersionState, &v.DeprecationReason, &v.ReplacedByVersion, &v.MigrationGuide, &v.CreatedAt, &v.PublishedAt, &v.ArchivedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to publish function version: %w", err)
	}
	return &v, nil
}

// ArchiveFunctionVersion archives a function version (published -> archived)
func (r *Repository) ArchiveFunctionVersion(ctx context.Context, id uuid.UUID, reason string) (*FunctionVersion, error) {
	now := time.Now()
	query := `
		UPDATE registry_function_versions
		SET version_state = $2, archived_at = $3
		WHERE id = $1
		RETURNING id, function_id, version, version_state, deprecation_reason, replaced_by_version, migration_guide, created_at, published_at, archived_at
	`
	var v FunctionVersion
	err := r.db.QueryRowContext(ctx, query, id, FunctionVersionStateArchived, now).Scan(
		&v.ID, &v.FunctionID, &v.Version, &v.VersionState, &v.DeprecationReason, &v.ReplacedByVersion, &v.MigrationGuide, &v.CreatedAt, &v.PublishedAt, &v.ArchivedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to archive function version: %w", err)
	}
	return &v, nil
}

// GetFunctionVersionByID retrieves a function version by ID
func (r *Repository) GetFunctionVersionByID(ctx context.Context, id uuid.UUID) (*FunctionVersion, error) {
	query := `
		SELECT id, function_id, version, version_state, deprecation_reason, replaced_by_version, migration_guide, created_at, published_at, archived_at
		FROM registry_function_versions WHERE id = $1
	`
	var v FunctionVersion
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&v.ID, &v.FunctionID, &v.Version, &v.VersionState, &v.DeprecationReason, &v.ReplacedByVersion, &v.MigrationGuide, &v.CreatedAt, &v.PublishedAt, &v.ArchivedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get function version: %w", err)
	}
	return &v, nil
}

// GetFunctionVersionByVersion retrieves a function version by version string
func (r *Repository) GetFunctionVersionByVersion(ctx context.Context, functionID uuid.UUID, version string) (*FunctionVersion, error) {
	query := `
		SELECT id, function_id, version, version_state, deprecation_reason, replaced_by_version, migration_guide, created_at, published_at, archived_at
		FROM registry_function_versions WHERE function_id = $1 AND version = $2
	`
	var v FunctionVersion
	err := r.db.QueryRowContext(ctx, query, functionID, version).Scan(
		&v.ID, &v.FunctionID, &v.Version, &v.VersionState, &v.DeprecationReason, &v.ReplacedByVersion, &v.MigrationGuide, &v.CreatedAt, &v.PublishedAt, &v.ArchivedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get function version: %w", err)
	}
	return &v, nil
}

// SetVersionAlias sets a version alias (latest, stable)
func (r *Repository) SetVersionAlias(ctx context.Context, functionID uuid.UUID, alias string, versionID uuid.UUID) error {
	query := `
		INSERT INTO version_aliases (id, function_id, alias, version_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (function_id, alias) DO UPDATE SET
			version_id = $4, updated_at = $6
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, uuid.New(), functionID, alias, versionID, now, now)
	if err != nil {
		return fmt.Errorf("failed to set version alias: %w", err)
	}
	return nil
}

// GetVersionAlias retrieves a version by alias
func (r *Repository) GetVersionAlias(ctx context.Context, functionID uuid.UUID, alias string) (*FunctionVersion, error) {
	query := `
		SELECT v.id, v.function_id, v.version, v.version_state, v.deprecation_reason, v.replaced_by_version, v.migration_guide, v.created_at, v.published_at, v.archived_at
		FROM registry_function_versions v
		JOIN version_aliases a ON v.id = a.version_id
		WHERE a.function_id = $1 AND a.alias = $2
	`
	var v FunctionVersion
	err := r.db.QueryRowContext(ctx, query, functionID, alias).Scan(
		&v.ID, &v.FunctionID, &v.Version, &v.VersionState, &v.DeprecationReason, &v.ReplacedByVersion, &v.MigrationGuide, &v.CreatedAt, &v.PublishedAt, &v.ArchivedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get version alias: %w", err)
	}
	return &v, nil
}

// GetStableFunctionVersion retrieves the highest non-prerelease version
func (r *Repository) GetStableFunctionVersion(ctx context.Context, functionID uuid.UUID) (*FunctionVersion, error) {
	query := `
		SELECT id, function_id, version, version_state, deprecation_reason, replaced_by_version, migration_guide, created_at, published_at, archived_at
		FROM registry_function_versions
		WHERE function_id = $1 AND version_state = 'published'
		AND version NOT LIKE '%-alpha%' AND version NOT LIKE '%-beta%' AND version NOT LIKE '%-rc%' AND version NOT LIKE '%-dev%'
		ORDER BY
			split_part(version, '.', 1)::int DESC,
			split_part(version, '.', 2)::int DESC,
			split_part(version, '.', 3)::int DESC
		LIMIT 1
	`
	var v FunctionVersion
	err := r.db.QueryRowContext(ctx, query, functionID).Scan(
		&v.ID, &v.FunctionID, &v.Version, &v.VersionState, &v.DeprecationReason, &v.ReplacedByVersion, &v.MigrationGuide, &v.CreatedAt, &v.PublishedAt, &v.ArchivedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stable function version: %w", err)
	}
	return &v, nil
}

// GetPreviousFunctionVersion retrieves the previous published version
func (r *Repository) GetPreviousFunctionVersion(ctx context.Context, functionID uuid.UUID, currentVersion string) (*FunctionVersion, error) {
	query := `
		SELECT id, function_id, version, version_state, deprecation_reason, replaced_by_version, migration_guide, created_at, published_at, archived_at
		FROM registry_function_versions
		WHERE function_id = $1 AND version_state = 'published' AND version != $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	var v FunctionVersion
	err := r.db.QueryRowContext(ctx, query, functionID, currentVersion).Scan(
		&v.ID, &v.FunctionID, &v.Version, &v.VersionState, &v.DeprecationReason, &v.ReplacedByVersion, &v.MigrationGuide, &v.CreatedAt, &v.PublishedAt, &v.ArchivedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get previous function version: %w", err)
	}
	return &v, nil
}

// ==================== Phase 2: Rollback ====================

// CreateRollbackRecord creates a rollback record
func (r *Repository) CreateRollbackRecord(ctx context.Context, params CreateRollbackParams) (*RollbackRecord, error) {
	id := uuid.New()
	now := time.Now()

	metadataJSON, _ := json.Marshal(params.Metadata)

	query := `
		INSERT INTO rollback_records (id, function_id, from_version, to_version, strategy, status, initiated_by, initiated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, function_id, from_version, to_version, strategy, status, initiated_by, initiated_at, metadata, completed_at
	`

	var record RollbackRecord
	err := r.db.QueryRowContext(ctx, query,
		id, params.FunctionID, params.FromVersion, params.ToVersion, params.Strategy, "initiated", params.InitiatedBy, now, metadataJSON,
	).Scan(
		&record.ID, &record.FunctionID, &record.FromVersion, &record.ToVersion, &record.Strategy, &record.Status, &record.InitiatedBy, &record.InitiatedAt, &record.Metadata, &record.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create rollback record: %w", err)
	}
	return &record, nil
}

// CompleteRollbackRecord marks a rollback as completed
func (r *Repository) CompleteRollbackRecord(ctx context.Context, id uuid.UUID, status string) error {
	now := time.Now()
	query := `
		UPDATE rollback_records SET status = $2, completed_at = $3 WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, status, now)
	if err != nil {
		return fmt.Errorf("failed to complete rollback record: %w", err)
	}
	return nil
}

// GetRollbackHistory retrieves rollback history for a function
func (r *Repository) GetRollbackHistory(ctx context.Context, functionID uuid.UUID, limit int) ([]RollbackRecord, error) {
	if limit == 0 {
		limit = 20
	}
	query := `
		SELECT id, function_id, from_version, to_version, strategy, status, initiated_by, initiated_at, metadata, completed_at
		FROM rollback_records
		WHERE function_id = $1
		ORDER BY initiated_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, functionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get rollback history: %w", err)
	}
	defer rows.Close()

	var records []RollbackRecord
	for rows.Next() {
		var record RollbackRecord
		err := rows.Scan(
			&record.ID, &record.FunctionID, &record.FromVersion, &record.ToVersion, &record.Strategy, &record.Status, &record.InitiatedBy, &record.InitiatedAt, &record.Metadata, &record.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rollback record: %w", err)
		}
		records = append(records, record)
	}
	return records, nil
}

// CreateRollbackParams contains parameters for creating a rollback record
type CreateRollbackParams struct {
	FunctionID  uuid.UUID
	FromVersion string
	ToVersion   string
	Strategy    RollbackStrategy
	InitiatedBy *uuid.UUID
	Metadata    map[string]interface{}
}

// ==================== Helper Methods ====================

// InitDefaultVersions initializes default API versions if they don't exist
func (r *Repository) InitDefaultVersions(ctx context.Context) error {
	// Check if v1 exists
	existing, err := r.GetAPIVersionByVersion(ctx, "v1")
	if err != nil {
		logrus.WithError(err).Error("Failed to check for existing API versions")
		return err
	}
	if existing != nil {
		return nil
	}

	// Insert default versions
	defaultVersions := []APIVersion{
		{
			ID:         uuid.New(),
			Version:    "v1",
			PathPrefix: "/v1",
			Status:     APIVersionStatusActive,
			ReleasedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Metadata:   json.RawMessage(`{"features": ["basic_functions", "deployments", "registry"]}`),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			Version:    "v2",
			PathPrefix: "/v2",
			Status:     APIVersionStatusActive,
			ReleasedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			Metadata:   json.RawMessage(`{"features": ["graphql", "webhooks", "advanced_deployments"]}`),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	for _, v := range defaultVersions {
		if err := r.CreateAPIVersion(ctx, &v); err != nil {
			// Ignore duplicate errors
			if !isDuplicateKeyError(err) {
				logrus.WithError(err).WithField("version", v.Version).Error("Failed to create default API version")
				return err
			}
		}
	}

	logrus.Info("Initialized default API versions")
	return nil
}

// isDuplicateKeyError checks if an error is a duplicate key error
func isDuplicateKeyError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // unique_violation
	}
	return false
}

// ==================== Phase 3: Deployment by Version ====================

// GetDeploymentByID retrieves a deployment by ID
func (r *Repository) GetDeploymentByID(ctx context.Context, id uuid.UUID) (*DeploymentVersion, error) {
	query := `
		SELECT id, function_id, function_version, deployment_id, provider, region, status, artifact_uri, checksum, rollback_id, metadata, created_at, completed_at
		FROM deployment_versions WHERE id = $1
	`
	var dv DeploymentVersion
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&dv.ID, &dv.FunctionID, &dv.Version, &dv.DeploymentID, &dv.Provider, &dv.Region, &dv.Status,
		&dv.ArtifactURI, &dv.Checksum, &dv.RollbackID, &dv.Metadata, &dv.CreatedAt, &dv.CompletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}
	return &dv, nil
}

// GetDeploymentsByFunctionVersion retrieves deployments for a specific function version
func (r *Repository) GetDeploymentsByFunctionVersion(ctx context.Context, params ListDeploymentsParams) ([]DeploymentVersion, error) {
	query := `
		SELECT id, function_id, function_version, deployment_id, provider, region, status, artifact_uri, checksum, rollback_id, metadata, created_at, completed_at
		FROM deployment_versions
		WHERE function_id = $1 AND function_version = $2
		AND ($3::text IS NULL OR status = $3)
		ORDER BY created_at DESC
		LIMIT $4
	`
	limit := params.Limit
	if limit == 0 {
		limit = 20
	}

	rows, err := r.db.QueryContext(ctx, query, params.FunctionID, params.Version, params.Status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments: %w", err)
	}
	defer rows.Close()

	var deployments []DeploymentVersion
	for rows.Next() {
		var dv DeploymentVersion
		err := rows.Scan(
			&dv.ID, &dv.FunctionID, &dv.Version, &dv.DeploymentID, &dv.Provider, &dv.Region, &dv.Status,
			&dv.ArtifactURI, &dv.Checksum, &dv.RollbackID, &dv.Metadata, &dv.CreatedAt, &dv.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment: %w", err)
		}
		deployments = append(deployments, dv)
	}
	return deployments, nil
}

// ==================== Phase 3: Service Contracts ====================

// GetAllServiceContracts retrieves all service contracts
func (r *Repository) GetAllServiceContracts(ctx context.Context, params ListServiceContractsParams) ([]ServiceContract, error) {
	query := `
		SELECT id, service_name, contract_version, contract_type, schema, status, introduced_in_release, deprecated_in_release, removed_in_release, created_at, updated_at
		FROM service_contracts
		WHERE ($1::text IS NULL OR service_name = $1)
		AND ($2::text IS NULL OR status = $2)
		ORDER BY service_name, contract_version DESC
		LIMIT $3
	`
	limit := params.Limit
	if limit == 0 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx, query, params.ServiceName, params.Status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get service contracts: %w", err)
	}
	defer rows.Close()

	var contracts []ServiceContract
	for rows.Next() {
		var sc ServiceContract
		err := rows.Scan(
			&sc.ID, &sc.ServiceName, &sc.ContractVersion, &sc.ContractType, &sc.Schema, &sc.Status,
			&sc.IntroducedInRelease, &sc.DeprecatedInRelease, &sc.RemovedInRelease, &sc.CreatedAt, &sc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service contract: %w", err)
		}
		contracts = append(contracts, sc)
	}
	return contracts, nil
}

// GetAllServiceNames retrieves all unique service names
func (r *Repository) GetAllServiceNames(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT service_name FROM service_contracts ORDER BY service_name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get service names: %w", err)
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan service name: %w", err)
		}
		services = append(services, name)
	}
	return services, nil
}

// GetLatestServiceContract retrieves the latest contract for a service
func (r *Repository) GetLatestServiceContract(ctx context.Context, serviceName string) (*ServiceContract, error) {
	query := `
		SELECT id, service_name, contract_version, contract_type, schema, status, introduced_in_release, deprecated_in_release, removed_in_release, created_at, updated_at
		FROM service_contracts
		WHERE service_name = $1 AND status = 'active'
		ORDER BY contract_version DESC
		LIMIT 1
	`
	var sc ServiceContract
	err := r.db.QueryRowContext(ctx, query, serviceName).Scan(
		&sc.ID, &sc.ServiceName, &sc.ContractVersion, &sc.ContractType, &sc.Schema, &sc.Status,
		&sc.IntroducedInRelease, &sc.DeprecatedInRelease, &sc.RemovedInRelease, &sc.CreatedAt, &sc.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest service contract: %w", err)
	}
	return &sc, nil
}

// GetCompatibleContractVersion finds a compatible contract version
func (r *Repository) GetCompatibleContractVersion(ctx context.Context, serviceName string, supportedVersions []string) (*ServiceContract, error) {
	if len(supportedVersions) == 0 {
		// Return latest active if no versions specified
		return r.GetLatestServiceContract(ctx, serviceName)
	}

	// Try each supported version in order
	for _, version := range supportedVersions {
		query := `
			SELECT id, service_name, contract_version, contract_type, schema, status, introduced_in_release, deprecated_in_release, removed_in_release, created_at, updated_at
			FROM service_contracts
			WHERE service_name = $1 AND contract_version = $2 AND status = 'active'
		`
		var sc ServiceContract
		err := r.db.QueryRowContext(ctx, query, serviceName, version).Scan(
			&sc.ID, &sc.ServiceName, &sc.ContractVersion, &sc.ContractType, &sc.Schema, &sc.Status,
			&sc.IntroducedInRelease, &sc.DeprecatedInRelease, &sc.RemovedInRelease, &sc.CreatedAt, &sc.UpdatedAt,
		)
		if err == nil {
			return &sc, nil
		}
		if err != sql.ErrNoRows {
			logrus.WithError(err).WithField("service", serviceName).WithField("version", version).Warn("Error checking contract version")
		}
	}

	// Fall back to latest
	return r.GetLatestServiceContract(ctx, serviceName)
}

// ==================== Phase 3: Version Lineage ====================

// GetVersionLineage retrieves the version history/lineage for a function
func (r *Repository) GetVersionLineage(ctx context.Context, functionID uuid.UUID, limit int) ([]VersionLineageEntry, error) {
	if limit == 0 {
		limit = 50
	}

	query := `
		SELECT id, version, version_state, created_at, published_at, archived_at
		FROM registry_function_versions
		WHERE function_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, functionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get version lineage: %w", err)
	}
	defer rows.Close()

	var entries []VersionLineageEntry
	var prevVersion string
	for rows.Next() {
		var entry VersionLineageEntry
		var state string
		err := rows.Scan(&entry.ID, &entry.Version, &state, &entry.CreatedAt, &entry.PublishedAt, &entry.ArchivedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		entry.State = state
		entry.ParentVersion = prevVersion
		// Determine change type based on version string
		entry.ChangeType = determineChangeType(prevVersion, entry.Version)
		entries = append(entries, entry)
		prevVersion = entry.Version
	}
	return entries, nil
}

// determineChangeType determines the change type between two version strings
func determineChangeType(parentVersion, currentVersion string) string {
	if parentVersion == "" {
		return ChangeTypeMajor // First version
	}

	// Simple semver comparison
	parentParts := strings.Split(parentVersion, ".")
	currentParts := strings.Split(currentVersion, ".")

	if len(parentParts) >= 2 && len(currentParts) >= 2 {
		parentMajor, _ := strconv.Atoi(parentParts[0])
		currentMajor, _ := strconv.Atoi(currentParts[0])

		if currentMajor > parentMajor {
			return ChangeTypeMajor
		}

		if len(parentParts) >= 2 && len(currentParts) >= 2 {
			parentMinor, _ := strconv.Atoi(parentParts[1])
			currentMinor, _ := strconv.Atoi(currentParts[1])

			if currentMinor > parentMinor {
				return ChangeTypeMinor
			}
		}

		return ChangeTypePatch
	}

	return ChangeTypeFeature
}

// CompareVersions compares two versions and returns the differences
func (r *Repository) CompareVersions(ctx context.Context, functionID uuid.UUID, version1, version2 string) (*VersionDiffResponse, error) {
	v1, err := r.GetFunctionVersionByVersion(ctx, functionID, version1)
	if err != nil {
		return nil, fmt.Errorf("failed to get version %s: %w", version1, err)
	}
	v2, err := r.GetFunctionVersionByVersion(ctx, functionID, version2)
	if err != nil {
		return nil, fmt.Errorf("failed to get version %s: %w", version2, err)
	}

	if v1 == nil {
		return nil, fmt.Errorf("version %s not found", version1)
	}
	if v2 == nil {
		return nil, fmt.Errorf("version %s not found", version2)
	}

	// Get changelogs for both versions
	changelogs, _ := r.GetChangelogByFunctionID(ctx, functionID)

	var changes []VersionDiffEntry
	var breakingChanges []string

	// Compare version strings
	if version1 != version2 {
		changes = append(changes, VersionDiffEntry{
			Field:      "version",
			FromValue:  version1,
			ToValue:    version2,
			ChangeType: "modified",
		})
	}

	// Compare states
	if v1.VersionState != v2.VersionState {
		changes = append(changes, VersionDiffEntry{
			Field:      "state",
			FromValue:  v1.VersionState,
			ToValue:    v2.VersionState,
			ChangeType: "modified",
		})
	}

	// Compare deprecation status
	if v1.DeprecationReason != v2.DeprecationReason {
		changes = append(changes, VersionDiffEntry{
			Field:      "deprecationReason",
			FromValue:  v1.DeprecationReason,
			ToValue:    v2.DeprecationReason,
			ChangeType: "modified",
		})
	}

	// Check changelogs for breaking changes
	for _, c := range changelogs {
		if c.Version == version2 && c.ChangeType == ChangeTypeBreaking {
			var bc []string
			if len(c.BreakingChanges) > 0 {
				_ = json.Unmarshal(c.BreakingChanges, &bc)
			}
			breakingChanges = append(breakingChanges, bc...)
		}
	}

	// Determine if there are breaking changes
	isBreaking := len(breakingChanges) > 0 || determineChangeType(version1, version2) == ChangeTypeBreaking

	// Build summary
	added := 0
	removed := 0
	modified := 0
	for _, c := range changes {
		switch c.ChangeType {
		case "added":
			added++
		case "removed":
			removed++
		case "modified":
			modified++
		}
	}

	resp := &VersionDiffResponse{
		FunctionID: functionID,
		Version1:   version1,
		Version2:   version2,
		Changes:    changes,
		Summary: VersionDiffSummary{
			TotalChanges:    len(changes),
			Added:           added,
			Removed:         removed,
			Modified:        modified,
			Breaking:        isBreaking,
			BreakingChanges: breakingChanges,
		},
	}

	return resp, nil
}
