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

	envVars := function.EnvVars
	if envVars == nil {
		envVars = []EnvironmentVariable{}
	}
	envVarsJSON, err := json.Marshal(envVars)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal env vars: %w", err)
	}
	// Bind JSONB as string: pgx database/sql encodes []byte as BYTEA, which Postgres rejects for json/jsonb.
	envVarsText := string(envVarsJSON)

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO functions (id, tenant_id, name, providers, region, code, env_vars, version, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10, $11)`,
		function.ID, function.TenantID, function.Name, pq.Array(function.Providers),
		function.Region, function.Code, envVarsText, function.Version, function.Status,
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
		SELECT id, tenant_id, app_id, name, providers, region, code, env_vars, version, status, created_at, updated_at
		FROM functions WHERE id = $1`, functionID).Scan(
		&function.ID, &function.TenantID, &function.AppID, &function.Name, pq.Array(&function.Providers),
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
		SELECT DISTINCT ON (id) id, tenant_id, app_id, name, providers, region, code, env_vars, version, status, created_at, updated_at
		FROM functions
		WHERE tenant_id = $1
		  AND name NOT ILIKE '%demo%'
		  AND name NOT ILIKE '%test%'
		  AND name NOT ILIKE 'my-demo%'
		  AND name NOT ILIKE 'my_demo%'
		ORDER BY id, created_at DESC`, tenantID)

	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}
	defer rows.Close()

	var functions []*FunctionConfig
	for rows.Next() {
		function := &FunctionConfig{}
		var envVarsJSON []byte

		err := rows.Scan(
			&function.ID, &function.TenantID, &function.AppID, &function.Name, pq.Array(&function.Providers),
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
	countQuery := `SELECT COUNT(*) FROM functions WHERE 1=1 
		  AND (name NOT ILIKE '%demo%' AND name NOT ILIKE '%test%' AND name NOT ILIKE 'my-demo%' AND name NOT ILIKE 'my_demo%')`
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
	query := `SELECT id, tenant_id, app_id, name, providers, region, code, env_vars, version, status, created_at, updated_at
		FROM functions WHERE 1=1 
		  AND (name NOT ILIKE '%demo%' AND name NOT ILIKE '%test%' AND name NOT ILIKE 'my-demo%' AND name NOT ILIKE 'my_demo%')`
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
			&function.ID, &function.TenantID, &function.AppID, &function.Name, pq.Array(&function.Providers),
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
			setParts = append(setParts, fmt.Sprintf("%s = $%d::jsonb", key, argIndex))
			envVarsJSON, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal env vars: %w", err)
			}
			args = append(args, string(envVarsJSON))
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

	meta := log.Metadata
	if meta == nil {
		meta = map[string]interface{}{}
	}
	metadataJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	metadataText := string(metadataJSON)

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO function_logs (id, function_id, deployment_id, level, message, timestamp, source, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb)`,
		log.ID, log.FunctionID, log.DeploymentID, log.Level, log.Message, log.Timestamp, log.Source, metadataText)

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

// DeleteFunctionLogsOlderThan deletes function log entries older than the cutoff time and returns the number deleted.
func (r *FunctionRepository) DeleteFunctionLogsOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM function_logs WHERE timestamp < $1`,
		cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old function logs: %w", err)
	}
	return result.RowsAffected()
}

// UsageByDay is a single day's usage count for dashboard
type UsageByDay struct {
	Time  string `json:"time"` // date as YYYY-MM-DD or formatted for display
	Value int64  `json:"value"`
}

// GetUsageByDay returns daily log counts for the tenant's functions (last N days).
func (r *FunctionRepository) GetUsageByDay(ctx context.Context, tenantID uuid.UUID, days int) ([]UsageByDay, error) {
	if days <= 0 {
		days = 14
	}
	since := time.Now().AddDate(0, 0, -days)

	query := `
		SELECT date_trunc('day', fl.timestamp)::date AS day, COUNT(*)::bigint
		FROM function_logs fl
		INNER JOIN functions f ON f.id = fl.function_id
		WHERE f.tenant_id = $1 AND fl.timestamp >= $2
		GROUP BY date_trunc('day', fl.timestamp)::date
		ORDER BY day`
	rows, err := r.db.QueryContext(ctx, query, tenantID, since)
	if err != nil {
		return nil, fmt.Errorf("get usage by day: %w", err)
	}
	defer rows.Close()

	var result []UsageByDay
	for rows.Next() {
		var day time.Time
		var count int64
		if err := rows.Scan(&day, &count); err != nil {
			return nil, fmt.Errorf("scan usage row: %w", err)
		}
		result = append(result, UsageByDay{Time: day.Format("2006-01-02"), Value: count})
	}
	return result, rows.Err()
}

// DashboardMetrics holds aggregated metrics for the dashboard (tenant-scoped).
type DashboardMetrics struct {
	RequestsThisMonth int64     `json:"requests_this_month"`
	RequestsPrevMonth int64     `json:"requests_prev_month"`
	AvgLatencyMs      *float64  `json:"avg_latency_ms,omitempty"`
	UptimePct         *float64  `json:"uptime_pct,omitempty"`         // last 7 days success rate
	UptimePrevPct     *float64  `json:"uptime_prev_pct,omitempty"`    // previous 7 days for comparison
	UptimeSparkline   []float64 `json:"uptime_sparkline,omitempty"`   // 7 daily success rates (newest last)
	RequestsSparkline []int64   `json:"requests_sparkline,omitempty"` // last 7 days daily counts (newest last)
}

// ExecutionRateByHour is one hour's execution count for dashboard
type ExecutionRateByHour struct {
	Time string `json:"time"`
	Rate int64  `json:"rate"`
}

// GetExecutionRateByHour returns hourly log counts for the tenant's functions (last N hours).
func (r *FunctionRepository) GetExecutionRateByHour(ctx context.Context, tenantID uuid.UUID, hours int) ([]ExecutionRateByHour, error) {
	if hours <= 0 {
		hours = 24
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	query := `
		SELECT date_trunc('hour', fl.timestamp) AS hour, COUNT(*)::bigint
		FROM function_logs fl
		INNER JOIN functions f ON f.id = fl.function_id
		WHERE f.tenant_id = $1 AND fl.timestamp >= $2
		GROUP BY date_trunc('hour', fl.timestamp)
		ORDER BY hour`
	rows, err := r.db.QueryContext(ctx, query, tenantID, since)
	if err != nil {
		return nil, fmt.Errorf("get execution rate by hour: %w", err)
	}
	defer rows.Close()

	var result []ExecutionRateByHour
	for rows.Next() {
		var hour time.Time
		var count int64
		if err := rows.Scan(&hour, &count); err != nil {
			return nil, fmt.Errorf("scan execution rate row: %w", err)
		}
		result = append(result, ExecutionRateByHour{
			Time: hour.Format("15:04"),
			Rate: count,
		})
	}
	return result, rows.Err()
}

// DashboardActivityItem represents one item in the dashboard activity feed (log or deployment).
type DashboardActivityItem struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"` // "deployment", "success", "error", "info", "invocation", "timeout"
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	FunctionID   string    `json:"function_id,omitempty"`
	FunctionName string    `json:"function_name,omitempty"`
}

// GetRecentActivityForTenant returns merged recent deployments and logs for the tenant, sorted by time desc.
func (r *FunctionRepository) GetRecentActivityForTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]DashboardActivityItem, error) {
	if limit <= 0 {
		limit = 20
	}

	// Recent deployments: id, function_id, status, created_at; join functions for name
	deployQuery := `
		SELECT fd.id::text, fd.function_id::text, f.name, fd.status, fd.created_at
		FROM function_deployments fd
		INNER JOIN functions f ON f.id = fd.function_id
		WHERE f.tenant_id = $1
		ORDER BY fd.created_at DESC
		LIMIT $2`
	deployRows, err := r.db.QueryContext(ctx, deployQuery, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("get recent deployments: %w", err)
	}
	defer deployRows.Close()

	var items []DashboardActivityItem
	for deployRows.Next() {
		var id, fnID, fnName, status string
		var ts time.Time
		if err := deployRows.Scan(&id, &fnID, &fnName, &status, &ts); err != nil {
			return nil, fmt.Errorf("scan deployment row: %w", err)
		}
		title := "Deployment " + status
		if status == "success" {
			title = "Deployment completed"
		} else if status == "failed" {
			title = "Deployment failed"
		}
		items = append(items, DashboardActivityItem{
			ID:           id,
			Type:         mapDeploymentStatusToActivityType(status),
			Title:        title,
			Timestamp:    ts,
			FunctionID:   fnID,
			FunctionName: fnName,
		})
	}
	if err := deployRows.Err(); err != nil {
		return nil, err
	}

	// Recent logs: id, function_id, level, message, timestamp; join functions for name
	logQuery := `
		SELECT fl.id::text, fl.function_id, f.name, fl.level, fl.message, fl.timestamp
		FROM function_logs fl
		INNER JOIN functions f ON f.id = fl.function_id
		WHERE f.tenant_id = $1
		ORDER BY fl.timestamp DESC
		LIMIT $2`
	logRows, err := r.db.QueryContext(ctx, logQuery, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("get recent logs: %w", err)
	}
	defer logRows.Close()

	for logRows.Next() {
		var id, level, message string
		var fnID *uuid.UUID
		var fnName *string
		var ts time.Time
		if err := logRows.Scan(&id, &fnID, &fnName, &level, &message, &ts); err != nil {
			return nil, fmt.Errorf("scan log row: %w", err)
		}
		fnIDStr := ""
		if fnID != nil {
			fnIDStr = fnID.String()
		}
		fnNameStr := ""
		if fnName != nil {
			fnNameStr = *fnName
		}
		items = append(items, DashboardActivityItem{
			ID:           id,
			Type:         mapLogLevelToActivityType(level),
			Title:        message,
			Description:  message,
			Timestamp:    ts,
			FunctionID:   fnIDStr,
			FunctionName: fnNameStr,
		})
	}
	if err := logRows.Err(); err != nil {
		return nil, err
	}

	// Sort by timestamp desc and take up to limit
	return sortDashboardActivitiesByTime(items, limit), nil
}

// GetDashboardMetrics returns aggregated metrics for the tenant dashboard.
func (r *FunctionRepository) GetDashboardMetrics(ctx context.Context, tenantID uuid.UUID) (*DashboardMetrics, error) {
	now := time.Now()
	startThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startPrevMonth := startThisMonth.AddDate(0, -1, 0)
	endPrevMonth := startThisMonth.Add(-time.Nanosecond)
	sevenDaysAgo := now.AddDate(0, 0, -7)
	fourteenDaysAgo := now.AddDate(0, 0, -14)

	out := &DashboardMetrics{
		UptimeSparkline:   make([]float64, 7),
		RequestsSparkline: make([]int64, 7),
	}

	// Requests this month
	var thisMonth, prevMonth int64
	q1 := `SELECT COUNT(*)::bigint FROM function_logs fl INNER JOIN functions f ON f.id = fl.function_id WHERE f.tenant_id = $1 AND fl.timestamp >= $2`
	if err := r.db.QueryRowContext(ctx, q1, tenantID, startThisMonth).Scan(&thisMonth); err != nil {
		return nil, fmt.Errorf("requests this month: %w", err)
	}
	out.RequestsThisMonth = thisMonth

	q2 := `SELECT COUNT(*)::bigint FROM function_logs fl INNER JOIN functions f ON f.id = fl.function_id WHERE f.tenant_id = $1 AND fl.timestamp >= $2 AND fl.timestamp <= $3`
	if err := r.db.QueryRowContext(ctx, q2, tenantID, startPrevMonth, endPrevMonth).Scan(&prevMonth); err != nil {
		return nil, fmt.Errorf("requests prev month: %w", err)
	}
	out.RequestsPrevMonth = prevMonth

	// Avg latency from metadata->>'duration_ms' (json/jsonb)
	var avgLat *float64
	latQuery := `
		SELECT AVG((fl.metadata->>'duration_ms')::double precision)
		FROM function_logs fl
		INNER JOIN functions f ON f.id = fl.function_id
		WHERE f.tenant_id = $1 AND fl.timestamp >= $2
		  AND fl.metadata IS NOT NULL
		  AND fl.metadata->>'duration_ms' IS NOT NULL
		  AND (fl.metadata->>'duration_ms') ~ '^[0-9.eE+-]+$'`
	err := r.db.QueryRowContext(ctx, latQuery, tenantID, sevenDaysAgo).Scan(&avgLat)
	if err != nil {
		// non-fatal: no duration data
		avgLat = nil
	}
	out.AvgLatencyMs = avgLat

	// Uptime last 7d: success rate = (total - errors) / total * 100
	var total7, errors7 int64
	uptimeQuery := `
		SELECT COUNT(*)::bigint,
		       COUNT(*) FILTER (WHERE fl.level = 'error')::bigint
		FROM function_logs fl
		INNER JOIN functions f ON f.id = fl.function_id
		WHERE f.tenant_id = $1 AND fl.timestamp >= $2`
	if err := r.db.QueryRowContext(ctx, uptimeQuery, tenantID, sevenDaysAgo).Scan(&total7, &errors7); err != nil {
		return nil, fmt.Errorf("uptime 7d: %w", err)
	}
	if total7 > 0 {
		pct := 100.0 * (float64(total7-errors7) / float64(total7))
		out.UptimePct = &pct
	}

	// Uptime previous 7d for comparison
	var total14, errors14 int64
	if err := r.db.QueryRowContext(ctx, uptimeQuery, tenantID, fourteenDaysAgo).Scan(&total14, &errors14); err != nil {
		return nil, fmt.Errorf("uptime 14d: %w", err)
	}
	// Prev 7d = last 14d total - last 7d total
	prevTotal := total14 - total7
	prevErrors := errors14 - errors7
	if prevTotal > 0 {
		pct := 100.0 * (float64(prevTotal-prevErrors) / float64(prevTotal))
		out.UptimePrevPct = &pct
	}

	// Uptime sparkline: 7 days, each day success rate
	for i := 0; i < 7; i++ {
		dayStart := now.AddDate(0, 0, -7+i).Truncate(24 * time.Hour)
		dayEnd := dayStart.Add(24*time.Hour - time.Nanosecond)
		var dayTotal, dayErrors int64
		dayQuery := `
			SELECT COUNT(*)::bigint, COUNT(*) FILTER (WHERE fl.level = 'error')::bigint
			FROM function_logs fl INNER JOIN functions f ON f.id = fl.function_id
			WHERE f.tenant_id = $1 AND fl.timestamp >= $2 AND fl.timestamp <= $3`
		if err := r.db.QueryRowContext(ctx, dayQuery, tenantID, dayStart, dayEnd).Scan(&dayTotal, &dayErrors); err != nil {
			out.UptimeSparkline[i] = 100
			continue
		}
		if dayTotal == 0 {
			out.UptimeSparkline[i] = 100
		} else {
			out.UptimeSparkline[i] = 100.0 * (float64(dayTotal-dayErrors) / float64(dayTotal))
		}
	}

	// Requests sparkline: last 7 days daily counts
	for i := 0; i < 7; i++ {
		dayStart := now.AddDate(0, 0, -7+i).Truncate(24 * time.Hour)
		dayEnd := dayStart.Add(24*time.Hour - time.Nanosecond)
		var c int64
		sparkQuery := `SELECT COUNT(*)::bigint FROM function_logs fl INNER JOIN functions f ON f.id = fl.function_id WHERE f.tenant_id = $1 AND fl.timestamp >= $2 AND fl.timestamp <= $3`
		if err := r.db.QueryRowContext(ctx, sparkQuery, tenantID, dayStart, dayEnd).Scan(&c); err != nil {
			continue
		}
		out.RequestsSparkline[i] = c
	}

	return out, nil
}

func mapDeploymentStatusToActivityType(status string) string {
	switch status {
	case "success":
		return "success"
	case "failed":
		return "error"
	case "deploying", "pending":
		return "deploy"
	default:
		return "info"
	}
}

func mapLogLevelToActivityType(level string) string {
	switch level {
	case "error":
		return "error"
	case "warn", "warning":
		return "timeout"
	default:
		return "invocation"
	}
}

func sortDashboardActivitiesByTime(items []DashboardActivityItem, limit int) []DashboardActivityItem {
	// Sort by timestamp descending
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Timestamp.After(items[i].Timestamp) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	if len(items) > limit {
		return items[:limit]
	}
	return items
}
