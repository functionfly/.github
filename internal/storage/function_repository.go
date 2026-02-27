package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// FunctionRepository handles function-related database operations
type FunctionRepository struct {
	db *sql.DB
}

// NewFunctionRepository creates a new function repository
func NewFunctionRepository(db *sql.DB) *FunctionRepository {
	return &FunctionRepository{db: db}
}

// CreateFunction creates a new function
func (r *FunctionRepository) CreateFunction(ctx context.Context, function *FunctionConfig) (*FunctionConfig, error) {
	if function.ID == uuid.Nil {
		function.ID = uuid.New()
	}
	if function.Version == "" {
		function.Version = "1.0.0"
	}
	if function.Status == "" {
		function.Status = "draft"
	}
	function.CreatedAt = time.Now()
	function.UpdatedAt = time.Now()

	envVarsJSON, err := json.Marshal(function.EnvVars)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal env vars: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO functions (id, tenant_id, name, providers, region, code, env_vars, version, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		function.ID, function.TenantID, function.Name, pq.Array(function.Providers),
		function.Region, function.Code, envVarsJSON, function.Version, function.Status,
		function.CreatedAt, function.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create function: %w", err)
	}

	return function, nil
}

// GetFunctionByID retrieves a function by ID
func (r *FunctionRepository) GetFunctionByID(ctx context.Context, functionID uuid.UUID) (*FunctionConfig, error) {
	function := &FunctionConfig{}
	var envVarsJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, providers, region, code, env_vars, version, status, created_at, updated_at
		FROM functions WHERE id = $1`, functionID).Scan(
		&function.ID, &function.TenantID, &function.Name, pq.Array(&function.Providers),
		&function.Region, &function.Code, &envVarsJSON, &function.Version, &function.Status,
		&function.CreatedAt, &function.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("function not found")
		}
		return nil, fmt.Errorf("failed to get function: %w", err)
	}

	if err := json.Unmarshal(envVarsJSON, &function.EnvVars); err != nil {
		return nil, fmt.Errorf("failed to unmarshal env vars: %w", err)
	}

	return function, nil
}

// ListFunctionsByTenant retrieves all functions for a tenant
func (r *FunctionRepository) ListFunctionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*FunctionConfig, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, name, providers, region, code, env_vars, version, status, created_at, updated_at
		FROM functions WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)

	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}
	defer rows.Close()

	var functions []*FunctionConfig
	for rows.Next() {
		function := &FunctionConfig{}
		var envVarsJSON []byte

		err := rows.Scan(
			&function.ID, &function.TenantID, &function.Name, pq.Array(&function.Providers),
			&function.Region, &function.Code, &envVarsJSON, &function.Version, &function.Status,
			&function.CreatedAt, &function.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan function: %w", err)
		}

		if err := json.Unmarshal(envVarsJSON, &function.EnvVars); err != nil {
			return nil, fmt.Errorf("failed to unmarshal env vars: %w", err)
		}

		functions = append(functions, function)
	}

	return functions, nil
}

// ListAllFunctions retrieves all functions across tenants for admin (with optional filters, limit, offset)
func (r *FunctionRepository) ListAllFunctions(ctx context.Context, limit, offset int, tenantID *uuid.UUID, status *string) ([]*FunctionConfig, int, error) {
	// Count total matching
	countQuery := `SELECT COUNT(*) FROM functions WHERE 1=1`
	countArgs := []interface{}{}
	argIdx := 1
	if tenantID != nil {
		countQuery += fmt.Sprintf(" AND tenant_id = $%d", argIdx)
		countArgs = append(countArgs, *tenantID)
		argIdx++
	}
	if status != nil && *status != "" {
		countQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		countArgs = append(countArgs, *status)
		argIdx++
	}
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count functions: %w", err)
	}

	// List with limit/offset
	query := `SELECT id, tenant_id, name, providers, region, code, env_vars, version, status, created_at, updated_at
		FROM functions WHERE 1=1`
	args := []interface{}{}
	argIdx = 1
	if tenantID != nil {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIdx)
		args = append(args, *tenantID)
		argIdx++
	}
	if status != nil && *status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *status)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list functions: %w", err)
	}
	defer rows.Close()

	var functions []*FunctionConfig
	for rows.Next() {
		function := &FunctionConfig{}
		var envVarsJSON []byte
		err := rows.Scan(
			&function.ID, &function.TenantID, &function.Name, pq.Array(&function.Providers),
			&function.Region, &function.Code, &envVarsJSON, &function.Version, &function.Status,
			&function.CreatedAt, &function.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan function: %w", err)
		}
		if err := json.Unmarshal(envVarsJSON, &function.EnvVars); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal env vars: %w", err)
		}
		functions = append(functions, function)
	}
	return functions, total, nil
}

// UpdateFunction updates a function
func (r *FunctionRepository) UpdateFunction(ctx context.Context, functionID uuid.UUID, updates map[string]interface{}) (*FunctionConfig, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		switch key {
		case "name", "region", "code", "version", "status":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
			args = append(args, value)
			argIndex++
		case "providers":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
			args = append(args, pq.Array(value))
			argIndex++
		case "env_vars":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
			envVarsJSON, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal env vars: %w", err)
			}
			args = append(args, envVarsJSON)
			argIndex++
		}
	}

	if len(setParts) == 0 {
		return r.GetFunctionByID(ctx, functionID)
	}

	setParts = append(setParts, "updated_at = $"+fmt.Sprintf("%d", argIndex))
	args = append(args, time.Now())
	args = append(args, functionID)

	query := fmt.Sprintf("UPDATE functions SET %s WHERE id = $%d",
		fmt.Sprintf("%s", setParts[0]), argIndex)

	for _, part := range setParts[1:] {
		query = fmt.Sprintf("%s, %s", query, part)
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update function: %w", err)
	}

	return r.GetFunctionByID(ctx, functionID)
}

// DeleteFunction deletes a function
func (r *FunctionRepository) DeleteFunction(ctx context.Context, functionID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM functions WHERE id = $1", functionID)
	if err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}
	return nil
}

// GetFunctionByAppIDAndName retrieves a function by app ID and name
func (r *FunctionRepository) GetFunctionByAppIDAndName(ctx context.Context, appID uuid.UUID, name string) (*FunctionConfig, error) {
	function := &FunctionConfig{}
	var envVarsJSON []byte
	var playgroundConfigJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, app_id, name, providers, region, code, env_vars, version, status,
		       playground_enabled, playground_config, created_at, updated_at
		FROM functions
		WHERE app_id = $1 AND name = $2 AND playground_enabled = true`,
		appID, name).Scan(
		&function.ID, &function.TenantID, &function.AppID, &function.Name, pq.Array(&function.Providers),
		&function.Region, &function.Code, &envVarsJSON, &function.Version, &function.Status,
		&function.PlaygroundEnabled, &playgroundConfigJSON, &function.CreatedAt, &function.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("function not found")
		}
		return nil, fmt.Errorf("failed to get function: %w", err)
	}

	if err := json.Unmarshal(envVarsJSON, &function.EnvVars); err != nil {
		return nil, fmt.Errorf("failed to unmarshal env vars: %w", err)
	}

	if err := json.Unmarshal(playgroundConfigJSON, &function.PlaygroundConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal playground config: %w", err)
	}

	return function, nil
}

// GetActiveDeploymentForFunction retrieves the active deployment for a function
func (r *FunctionRepository) GetActiveDeploymentForFunction(ctx context.Context, functionID uuid.UUID) (*FunctionDeployment, error) {
	deployment := &FunctionDeployment{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, function_id, version, status, provider, region, deployed_url, error_message, created_at, updated_at
		FROM function_deployments
		WHERE function_id = $1 AND status = 'success'
		ORDER BY created_at DESC
		LIMIT 1`, functionID).Scan(
		&deployment.ID, &deployment.FunctionID, &deployment.Version, &deployment.Status,
		&deployment.Provider, &deployment.Region, &deployment.DeployedURL, &deployment.ErrorMessage,
		&deployment.CreatedAt, &deployment.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active deployment found")
		}
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	return deployment, nil
}

// CreateFunctionDeployment creates a new function deployment
func (r *FunctionRepository) CreateFunctionDeployment(ctx context.Context, deployment *FunctionDeployment) (*FunctionDeployment, error) {
	if deployment.ID == uuid.Nil {
		deployment.ID = uuid.New()
	}
	deployment.CreatedAt = time.Now()
	deployment.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO function_deployments (id, function_id, version, status, provider, region, deployed_url, error_message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		deployment.ID, deployment.FunctionID, deployment.Version, deployment.Status,
		deployment.Provider, deployment.Region, deployment.DeployedURL, deployment.ErrorMessage,
		deployment.CreatedAt, deployment.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create function deployment: %w", err)
	}

	return deployment, nil
}

// GetFunctionDeploymentByID retrieves a function deployment by ID
func (r *FunctionRepository) GetFunctionDeploymentByID(ctx context.Context, deploymentID uuid.UUID) (*FunctionDeployment, error) {
	deployment := &FunctionDeployment{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, function_id, version, status, provider, region, deployed_url, error_message, created_at, updated_at
		FROM function_deployments WHERE id = $1`, deploymentID).Scan(
		&deployment.ID, &deployment.FunctionID, &deployment.Version, &deployment.Status,
		&deployment.Provider, &deployment.Region, &deployment.DeployedURL, &deployment.ErrorMessage,
		&deployment.CreatedAt, &deployment.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("function deployment not found")
		}
		return nil, fmt.Errorf("failed to get function deployment: %w", err)
	}

	return deployment, nil
}

// ListFunctionDeployments retrieves deployments for a function
func (r *FunctionRepository) ListFunctionDeployments(ctx context.Context, functionID uuid.UUID, limit int) ([]*FunctionDeployment, error) {
	query := `
		SELECT id, function_id, version, status, provider, region, deployed_url, error_message, created_at, updated_at
		FROM function_deployments WHERE function_id = $1 ORDER BY created_at DESC`

	args := []interface{}{functionID}
	if limit > 0 {
		query += " LIMIT $2"
		args = append(args, limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list function deployments: %w", err)
	}
	defer rows.Close()

	var deployments []*FunctionDeployment
	for rows.Next() {
		deployment := &FunctionDeployment{}

		err := rows.Scan(
			&deployment.ID, &deployment.FunctionID, &deployment.Version, &deployment.Status,
			&deployment.Provider, &deployment.Region, &deployment.DeployedURL, &deployment.ErrorMessage,
			&deployment.CreatedAt, &deployment.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan function deployment: %w", err)
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// UpdateFunctionDeploymentStatus updates a deployment status
func (r *FunctionRepository) UpdateFunctionDeploymentStatus(ctx context.Context, deploymentID uuid.UUID, status string, deployedURL, errorMessage *string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE function_deployments
		SET status = $1, deployed_url = $2, error_message = $3, updated_at = $4
		WHERE id = $5`,
		status, deployedURL, errorMessage, time.Now(), deploymentID)

	if err != nil {
		return fmt.Errorf("failed to update function deployment status: %w", err)
	}

	return nil
}

// CreateFunctionLog creates a new function log entry
func (r *FunctionRepository) CreateFunctionLog(ctx context.Context, log *FunctionLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	log.Timestamp = time.Now()

	metadataJSON, err := json.Marshal(log.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO function_logs (id, function_id, deployment_id, level, message, timestamp, source, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		log.ID, log.FunctionID, log.DeploymentID, log.Level, log.Message, log.Timestamp, log.Source, metadataJSON)

	if err != nil {
		return fmt.Errorf("failed to create function log: %w", err)
	}

	return nil
}

// GetFunctionLogs retrieves function logs with optional filters
func (r *FunctionRepository) GetFunctionLogs(ctx context.Context, functionID *uuid.UUID, deploymentID *uuid.UUID, limit int, since *time.Time, level *string) ([]*FunctionLog, error) {
	whereParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if functionID != nil {
		whereParts = append(whereParts, fmt.Sprintf("function_id = $%d", argIndex))
		args = append(args, *functionID)
		argIndex++
	}

	if deploymentID != nil {
		whereParts = append(whereParts, fmt.Sprintf("deployment_id = $%d", argIndex))
		args = append(args, *deploymentID)
		argIndex++
	}

	if since != nil {
		whereParts = append(whereParts, fmt.Sprintf("timestamp >= $%d", argIndex))
		args = append(args, *since)
		argIndex++
	}

	if level != nil {
		whereParts = append(whereParts, fmt.Sprintf("level = $%d", argIndex))
		args = append(args, *level)
		argIndex++
	}

	query := `SELECT id, function_id, deployment_id, level, message, timestamp, source, metadata FROM function_logs`
	if len(whereParts) > 0 {
		query += " WHERE " + fmt.Sprintf("%s", whereParts[0])
		for _, part := range whereParts[1:] {
			query += " AND " + part
		}
	}
	query += " ORDER BY timestamp DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get function logs: %w", err)
	}
	defer rows.Close()

	var logs []*FunctionLog
	for rows.Next() {
		log := &FunctionLog{}
		var metadataJSON []byte

		err := rows.Scan(
			&log.ID, &log.FunctionID, &log.DeploymentID, &log.Level, &log.Message,
			&log.Timestamp, &log.Source, &metadataJSON)

		if err != nil {
			return nil, fmt.Errorf("failed to scan function log: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &log.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		logs = append(logs, log)
	}

	return logs, nil
}
