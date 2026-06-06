package studio

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DevOpsRepository handles database operations for studio DevOps
type DevOpsRepository struct {
	db *sql.DB
}

// NewDevOpsRepository creates a new DevOps repository
func NewDevOpsRepository(db *sql.DB) *DevOpsRepository {
	return &DevOpsRepository{db: db}
}

// InitSchema creates the required tables for DevOps data
func (r *DevOpsRepository) InitSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS studio_devops_pipelines (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			version TEXT DEFAULT '',
			status TEXT DEFAULT 'active',
			stages JSONB DEFAULT '[]',
			current_stage_id TEXT,
			triggered_by TEXT DEFAULT '',
			triggered_at BIGINT DEFAULT 0,
			branch TEXT DEFAULT '',
			commit_sha TEXT DEFAULT '',
			source TEXT DEFAULT 'manual',
			created_at BIGINT DEFAULT 0,
			updated_at BIGINT DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS studio_devops_environments (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT DEFAULT 'development',
			color TEXT DEFAULT '#06b6d4',
			variables JSONB DEFAULT '{}',
			secrets JSONB DEFAULT '[]',
			replicas INT DEFAULT 1,
			auto_scale BOOLEAN DEFAULT false,
			region TEXT DEFAULT '',
			created_at BIGINT DEFAULT 0,
			updated_at BIGINT DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS studio_devops_regions (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			provider TEXT DEFAULT 'aws',
			zone TEXT DEFAULT '',
			zone_name TEXT DEFAULT '',
			location TEXT DEFAULT '',
			country TEXT DEFAULT '',
			coordinates JSONB DEFAULT '{}',
			is_available BOOLEAN DEFAULT true,
			is_recommended BOOLEAN DEFAULT false,
			specs JSONB,
			created_at BIGINT DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_devops_pipelines_tenant ON studio_devops_pipelines(tenant_id)`,
		`CREATE INDEX IF NOT EXISTS idx_devops_environments_tenant ON studio_devops_environments(tenant_id)`,
		`CREATE INDEX IF NOT EXISTS idx_devops_regions_tenant ON studio_devops_regions(tenant_id)`,
	}

	for _, q := range queries {
		if _, err := r.db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("init devops schema: %w", err)
		}
	}
	return nil
}

// ============================================================================
// Pipeline Operations
// ============================================================================

// ListPipelines returns pipelines filtered by tenant and optional status
func (r *DevOpsRepository) ListPipelines(ctx context.Context, tenantID, status string, limit, offset int) ([]Pipeline, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIdx))
	args = append(args, tenantID)
	argIdx++

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT id, tenant_id, name, version, status, stages, current_stage_id,
		       triggered_by, triggered_at, branch, commit_sha, source, created_at, updated_at
		FROM studio_devops_pipelines
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(conditions, " AND "), argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list pipelines: %w", err)
	}
	defer rows.Close()

	var pipelines []Pipeline
	for rows.Next() {
		var p Pipeline
		var stagesJSON []byte
		err := rows.Scan(
			&p.ID, &p.TenantID, &p.Name, &p.Version, &p.Status, &stagesJSON,
			&p.CurrentStageID, &p.TriggeredBy, &p.TriggeredAt,
			&p.Branch, &p.CommitSha, &p.Source, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan pipeline: %w", err)
		}
		if stagesJSON != nil {
			json.Unmarshal(stagesJSON, &p.Stages)
		}
		pipelines = append(pipelines, p)
	}

	return pipelines, rows.Err()
}

// GetPipeline returns a single pipeline by ID
func (r *DevOpsRepository) GetPipeline(ctx context.Context, tenantID, pipelineID string) (*Pipeline, error) {
	var p Pipeline
	var stagesJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, version, status, stages, current_stage_id,
		       triggered_by, triggered_at, branch, commit_sha, source, created_at, updated_at
		FROM studio_devops_pipelines
		WHERE id = $1 AND tenant_id = $2
	`, pipelineID, tenantID).Scan(
		&p.ID, &p.TenantID, &p.Name, &p.Version, &p.Status, &stagesJSON,
		&p.CurrentStageID, &p.TriggeredBy, &p.TriggeredAt,
		&p.Branch, &p.CommitSha, &p.Source, &p.CreatedAt, &p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get pipeline: %w", err)
	}
	if stagesJSON != nil {
		json.Unmarshal(stagesJSON, &p.Stages)
	}
	return &p, nil
}

// CreatePipeline creates a new pipeline
func (r *DevOpsRepository) CreatePipeline(ctx context.Context, pipeline *Pipeline) error {
	stagesJSON, _ := json.Marshal(pipeline.Stages)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO studio_devops_pipelines
		(id, tenant_id, name, version, status, stages, current_stage_id, triggered_by, triggered_at,
		 branch, commit_sha, source, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, pipeline.ID, pipeline.TenantID, pipeline.Name, pipeline.Version, pipeline.Status,
		stagesJSON, pipeline.CurrentStageID, pipeline.TriggeredBy, pipeline.TriggeredAt,
		pipeline.Branch, pipeline.CommitSha, pipeline.Source, pipeline.CreatedAt, pipeline.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create pipeline: %w", err)
	}
	return nil
}

// UpdatePipelineStage updates a specific stage in a pipeline
func (r *DevOpsRepository) UpdatePipelineStage(ctx context.Context, tenantID, pipelineID, stageID string, updates map[string]interface{}) (*Pipeline, error) {
	pipeline, err := r.GetPipeline(ctx, tenantID, pipelineID)
	if err != nil || pipeline == nil {
		return nil, err
	}

	// Find and update the stage
	found := false
	for i := range pipeline.Stages {
		if pipeline.Stages[i].ID == stageID {
			found = true
			if status, ok := updates["status"].(string); ok {
				pipeline.Stages[i].Status = PipelineStageStatus(status)
			}
			if duration, ok := updates["duration"].(int64); ok {
				pipeline.Stages[i].Duration = duration
			}
			if startedAt, ok := updates["started_at"].(int64); ok {
				pipeline.Stages[i].StartedAt = startedAt
			}
			if completedAt, ok := updates["completed_at"].(int64); ok {
				pipeline.Stages[i].CompletedAt = completedAt
			}
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("stage not found")
	}

	pipeline.UpdatedAt = time.Now().UnixMilli()
	stagesJSON, _ := json.Marshal(pipeline.Stages)

	_, err = r.db.ExecContext(ctx, `
		UPDATE studio_devops_pipelines
		SET stages = $1, updated_at = $2
		WHERE id = $3 AND tenant_id = $4
	`, stagesJSON, pipeline.UpdatedAt, pipelineID, tenantID)

	if err != nil {
		return nil, fmt.Errorf("update pipeline stage: %w", err)
	}
	return pipeline, nil
}

// ============================================================================
// Environment Operations
// ============================================================================

// ListEnvironments returns environments filtered by tenant and optional type
func (r *DevOpsRepository) ListEnvironments(ctx context.Context, tenantID, envType string) ([]Environment, error) {
	query := `
		SELECT id, tenant_id, name, type, color, variables, secrets, replicas, auto_scale, region, created_at, updated_at
		FROM studio_devops_environments
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}

	if envType != "" {
		query += " AND type = $2"
		args = append(args, envType)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list environments: %w", err)
	}
	defer rows.Close()

	var environments []Environment
	for rows.Next() {
		var e Environment
		var variablesJSON, secretsJSON []byte
		err := rows.Scan(
			&e.ID, &e.TenantID, &e.Name, &e.Type, &e.Color,
			&variablesJSON, &secretsJSON, &e.Replicas, &e.AutoScale,
			&e.Region, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan environment: %w", err)
		}
		if variablesJSON != nil {
			json.Unmarshal(variablesJSON, &e.Variables)
		}
		if secretsJSON != nil {
			json.Unmarshal(secretsJSON, &e.Secrets)
		}
		environments = append(environments, e)
	}

	return environments, rows.Err()
}

// GetEnvironment returns a single environment by ID
func (r *DevOpsRepository) GetEnvironment(ctx context.Context, tenantID, envID string) (*Environment, error) {
	var e Environment
	var variablesJSON, secretsJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, type, color, variables, secrets, replicas, auto_scale, region, created_at, updated_at
		FROM studio_devops_environments
		WHERE id = $1 AND tenant_id = $2
	`, envID, tenantID).Scan(
		&e.ID, &e.TenantID, &e.Name, &e.Type, &e.Color,
		&variablesJSON, &secretsJSON, &e.Replicas, &e.AutoScale,
		&e.Region, &e.CreatedAt, &e.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get environment: %w", err)
	}
	if variablesJSON != nil {
		json.Unmarshal(variablesJSON, &e.Variables)
	}
	if secretsJSON != nil {
		json.Unmarshal(secretsJSON, &e.Secrets)
	}
	return &e, nil
}

// CreateEnvironment creates a new environment
func (r *DevOpsRepository) CreateEnvironment(ctx context.Context, env *Environment) error {
	variablesJSON, _ := json.Marshal(env.Variables)
	secretsJSON, _ := json.Marshal(env.Secrets)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO studio_devops_environments
		(id, tenant_id, name, type, color, variables, secrets, replicas, auto_scale, region, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, env.ID, env.TenantID, env.Name, env.Type, env.Color,
		variablesJSON, secretsJSON, env.Replicas, env.AutoScale,
		env.Region, env.CreatedAt, env.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create environment: %w", err)
	}
	return nil
}

// UpdateEnvironment updates an environment
func (r *DevOpsRepository) UpdateEnvironment(ctx context.Context, tenantID, envID string, updates map[string]interface{}) (*Environment, error) {
	env, err := r.GetEnvironment(ctx, tenantID, envID)
	if err != nil || env == nil {
		return nil, err
	}

	if name, ok := updates["name"].(string); ok {
		env.Name = name
	}
	if envType, ok := updates["type"].(string); ok {
		env.Type = EnvironmentType(envType)
	}
	if color, ok := updates["color"].(string); ok {
		env.Color = color
	}
	if replicas, ok := updates["replicas"].(int); ok {
		env.Replicas = replicas
	}
	if autoScale, ok := updates["auto_scale"].(bool); ok {
		env.AutoScale = autoScale
	}
	if region, ok := updates["region"].(string); ok {
		env.Region = region
	}

	env.UpdatedAt = time.Now().UnixMilli()
	variablesJSON, _ := json.Marshal(env.Variables)
	secretsJSON, _ := json.Marshal(env.Secrets)

	_, err = r.db.ExecContext(ctx, `
		UPDATE studio_devops_environments
		SET name = $1, type = $2, color = $3, variables = $4, secrets = $5,
		    replicas = $6, auto_scale = $7, region = $8, updated_at = $9
		WHERE id = $10 AND tenant_id = $11
	`, env.Name, env.Type, env.Color, variablesJSON, secretsJSON,
		env.Replicas, env.AutoScale, env.Region, env.UpdatedAt, envID, tenantID)

	if err != nil {
		return nil, fmt.Errorf("update environment: %w", err)
	}
	return env, nil
}

// DeleteEnvironment deletes an environment
func (r *DevOpsRepository) DeleteEnvironment(ctx context.Context, tenantID, envID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM studio_devops_environments WHERE id = $1 AND tenant_id = $2
	`, envID, tenantID)

	if err != nil {
		return fmt.Errorf("delete environment: %w", err)
	}
	return nil
}

// AddEnvironmentVariable adds or updates a variable in an environment
func (r *DevOpsRepository) AddEnvironmentVariable(ctx context.Context, tenantID, envID, key, value string) (*Environment, error) {
	env, err := r.GetEnvironment(ctx, tenantID, envID)
	if err != nil || env == nil {
		return nil, err
	}

	if env.Variables == nil {
		env.Variables = make(map[string]string)
	}
	env.Variables[key] = value
	env.UpdatedAt = time.Now().UnixMilli()

	variablesJSON, _ := json.Marshal(env.Variables)
	_, err = r.db.ExecContext(ctx, `
		UPDATE studio_devops_environments SET variables = $1, updated_at = $2 WHERE id = $3 AND tenant_id = $4
	`, variablesJSON, env.UpdatedAt, envID, tenantID)

	if err != nil {
		return nil, fmt.Errorf("add variable: %w", err)
	}
	return env, nil
}

// AddEnvironmentSecret adds a secret to an environment
func (r *DevOpsRepository) AddEnvironmentSecret(ctx context.Context, tenantID, envID, key string) (*Environment, error) {
	env, err := r.GetEnvironment(ctx, tenantID, envID)
	if err != nil || env == nil {
		return nil, err
	}

	env.Secrets = append(env.Secrets, EnvironmentSecret{
		Key:         key,
		Masked:      true,
		LastUpdated: time.Now().UnixMilli(),
	})
	env.UpdatedAt = time.Now().UnixMilli()

	secretsJSON, _ := json.Marshal(env.Secrets)
	_, err = r.db.ExecContext(ctx, `
		UPDATE studio_devops_environments SET secrets = $1, updated_at = $2 WHERE id = $3 AND tenant_id = $4
	`, secretsJSON, env.UpdatedAt, envID, tenantID)

	if err != nil {
		return nil, fmt.Errorf("add secret: %w", err)
	}
	return env, nil
}

// ============================================================================
// Cloud Region Operations
// ============================================================================

// ListCloudRegions returns cloud regions filtered by tenant and optional provider
func (r *DevOpsRepository) ListCloudRegions(ctx context.Context, tenantID, provider string) ([]CloudRegion, error) {
	query := `
		SELECT id, tenant_id, name, provider, zone, zone_name, location, country,
		       coordinates, is_available, is_recommended, specs, created_at
		FROM studio_devops_regions
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}

	if provider != "" {
		query += " AND provider = $2"
		args = append(args, provider)
	}
	query += " ORDER BY is_recommended DESC, name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list regions: %w", err)
	}
	defer rows.Close()

	var regions []CloudRegion
	for rows.Next() {
		var reg CloudRegion
		var coordsJSON, specsJSON []byte
		err := rows.Scan(
			&reg.ID, &reg.TenantID, &reg.Name, &reg.Provider, &reg.Zone, &reg.ZoneName,
			&reg.Location, &reg.Country, &coordsJSON, &reg.IsAvailable, &reg.IsRecommended,
			&specsJSON, &reg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan region: %w", err)
		}
		if coordsJSON != nil {
			json.Unmarshal(coordsJSON, &reg.Coordinates)
		}
		if specsJSON != nil {
			json.Unmarshal(specsJSON, &reg.Specs)
		}
		regions = append(regions, reg)
	}

	return regions, rows.Err()
}

// GetCloudRegion returns a single region by ID
func (r *DevOpsRepository) GetCloudRegion(ctx context.Context, tenantID, regionID string) (*CloudRegion, error) {
	var reg CloudRegion
	var coordsJSON, specsJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, provider, zone, zone_name, location, country,
		       coordinates, is_available, is_recommended, specs, created_at
		FROM studio_devops_regions
		WHERE id = $1 AND tenant_id = $2
	`, regionID, tenantID).Scan(
		&reg.ID, &reg.TenantID, &reg.Name, &reg.Provider, &reg.Zone, &reg.ZoneName,
		&reg.Location, &reg.Country, &coordsJSON, &reg.IsAvailable, &reg.IsRecommended,
		&specsJSON, &reg.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get region: %w", err)
	}
	if coordsJSON != nil {
		json.Unmarshal(coordsJSON, &reg.Coordinates)
	}
	if specsJSON != nil {
		json.Unmarshal(specsJSON, &reg.Specs)
	}
	return &reg, nil
}

// CreateCloudRegion creates a new cloud region
func (r *DevOpsRepository) CreateCloudRegion(ctx context.Context, region *CloudRegion) error {
	coordsJSON, _ := json.Marshal(region.Coordinates)
	specsJSON, _ := json.Marshal(region.Specs)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO studio_devops_regions
		(id, tenant_id, name, provider, zone, zone_name, location, country,
		 coordinates, is_available, is_recommended, specs, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, region.ID, region.TenantID, region.Name, region.Provider, region.Zone, region.ZoneName,
		region.Location, region.Country, coordsJSON, region.IsAvailable, region.IsRecommended,
		specsJSON, region.CreatedAt)

	if err != nil {
		return fmt.Errorf("create region: %w", err)
	}
	return nil
}