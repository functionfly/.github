package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CreateEmailWorkflowConfig creates a new email workflow configuration
func (db *PostgresDB) CreateEmailWorkflowConfig(ctx context.Context, config *EmailWorkflowConfig) error {
	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	_, err := db.ExecContext(ctx, `
		INSERT INTO email_workflow_configs (id, tenant_id, bundle_slug, name, description, trigger, category, delay_days, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (tenant_id, bundle_slug, name) DO NOTHING`,
		config.ID, config.TenantID, config.BundleSlug, config.Name, config.Description,
		config.Trigger, config.Category, config.DelayDays, config.Active,
		config.CreatedAt, config.UpdatedAt)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"workflow_id": config.ID,
		"tenant_id":   config.TenantID,
		"name":        config.Name,
	}).Debug("Created email workflow config")

	return nil
}

// GetEmailWorkflowConfigsByTenant returns all email workflow configs for a tenant
func (db *PostgresDB) GetEmailWorkflowConfigsByTenant(ctx context.Context, tenantID uuid.UUID) ([]EmailWorkflowConfig, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, tenant_id, bundle_slug, name, description, trigger, category, delay_days, active, created_at, updated_at
		FROM email_workflow_configs
		WHERE tenant_id = $1
		ORDER BY category, name`,
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []EmailWorkflowConfig
	for rows.Next() {
		var c EmailWorkflowConfig
		err := rows.Scan(&c.ID, &c.TenantID, &c.BundleSlug, &c.Name, &c.Description,
			&c.Trigger, &c.Category, &c.DelayDays, &c.Active, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, rows.Err()
}

// GetEmailWorkflowConfigsByBundle returns all email workflow configs for a tenant and bundle
func (db *PostgresDB) GetEmailWorkflowConfigsByBundle(ctx context.Context, tenantID uuid.UUID, bundleSlug string) ([]EmailWorkflowConfig, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, tenant_id, bundle_slug, name, description, trigger, category, delay_days, active, created_at, updated_at
		FROM email_workflow_configs
		WHERE tenant_id = $1 AND bundle_slug = $2
		ORDER BY category, name`,
		tenantID, bundleSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []EmailWorkflowConfig
	for rows.Next() {
		var c EmailWorkflowConfig
		err := rows.Scan(&c.ID, &c.TenantID, &c.BundleSlug, &c.Name, &c.Description,
			&c.Trigger, &c.Category, &c.DelayDays, &c.Active, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, rows.Err()
}

// GetActiveEmailWorkflowConfigsByTenant returns all active email workflow configs for a tenant
func (db *PostgresDB) GetActiveEmailWorkflowConfigsByTenant(ctx context.Context, tenantID uuid.UUID) ([]EmailWorkflowConfig, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, tenant_id, bundle_slug, name, description, trigger, category, delay_days, active, created_at, updated_at
		FROM email_workflow_configs
		WHERE tenant_id = $1 AND active = true
		ORDER BY category, name`,
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []EmailWorkflowConfig
	for rows.Next() {
		var c EmailWorkflowConfig
		err := rows.Scan(&c.ID, &c.TenantID, &c.BundleSlug, &c.Name, &c.Description,
			&c.Trigger, &c.Category, &c.DelayDays, &c.Active, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, rows.Err()
}

// GetEmailWorkflowConfigByID returns a single email workflow config by ID
func (db *PostgresDB) GetEmailWorkflowConfigByID(ctx context.Context, id uuid.UUID) (*EmailWorkflowConfig, error) {
	var c EmailWorkflowConfig
	err := db.QueryRowContext(ctx, `
		SELECT id, tenant_id, bundle_slug, name, description, trigger, category, delay_days, active, created_at, updated_at
		FROM email_workflow_configs
		WHERE id = $1`,
		id).Scan(&c.ID, &c.TenantID, &c.BundleSlug, &c.Name, &c.Description,
		&c.Trigger, &c.Category, &c.DelayDays, &c.Active, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateEmailWorkflowConfig updates an email workflow configuration
func (db *PostgresDB) UpdateEmailWorkflowConfig(ctx context.Context, config *EmailWorkflowConfig) error {
	config.UpdatedAt = time.Now()
	_, err := db.ExecContext(ctx, `
		UPDATE email_workflow_configs
		SET name = $2, description = $3, trigger = $4, category = $5, delay_days = $6, active = $7, updated_at = $8
		WHERE id = $1`,
		config.ID, config.Name, config.Description, config.Trigger, config.Category,
		config.DelayDays, config.Active, config.UpdatedAt)
	return err
}

// DeleteEmailWorkflowConfig deletes an email workflow configuration
func (db *PostgresDB) DeleteEmailWorkflowConfig(ctx context.Context, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, `DELETE FROM email_workflow_configs WHERE id = $1`, id)
	return err
}

// CreateEmailWorkflowExecution creates a new email workflow execution record
func (db *PostgresDB) CreateEmailWorkflowExecution(ctx context.Context, exec *EmailWorkflowExecution) error {
	if exec.ID == uuid.Nil {
		exec.ID = uuid.New()
	}
	exec.CreatedAt = time.Now()
	exec.UpdatedAt = time.Now()

	_, err := db.ExecContext(ctx, `
		INSERT INTO email_workflow_executions (id, tenant_id, workflow_id, recipient, status, scheduled_at, email_subject, email_template, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		exec.ID, exec.TenantID, exec.WorkflowID, exec.Recipient, exec.Status, exec.ScheduledAt,
		exec.EmailSubject, exec.EmailTemplate, exec.CreatedAt, exec.UpdatedAt)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"execution_id": exec.ID,
		"workflow_id":  exec.WorkflowID,
		"tenant_id":   exec.TenantID,
		"status":      exec.Status,
	}).Debug("Created email workflow execution")

	return nil
}

// GetPendingEmailWorkflowExecutions returns pending executions ready to be processed
func (db *PostgresDB) GetPendingEmailWorkflowExecutions(ctx context.Context, limit int) ([]EmailWorkflowExecution, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, tenant_id, workflow_id, recipient, status, scheduled_at, sent_at, error, retry_count, last_retry_at, email_subject, email_template, created_at, updated_at
		FROM email_workflow_executions
		WHERE status = 'pending' AND scheduled_at <= NOW()
		ORDER BY scheduled_at ASC
		LIMIT $1`,
		limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var execs []EmailWorkflowExecution
	for rows.Next() {
		var e EmailWorkflowExecution
		err := rows.Scan(&e.ID, &e.TenantID, &e.WorkflowID, &e.Recipient, &e.Status, &e.ScheduledAt,
			&e.SentAt, &e.Error, &e.RetryCount, &e.LastRetryAt, &e.EmailSubject, &e.EmailTemplate,
			&e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		execs = append(execs, e)
	}
	return execs, rows.Err()
}

// GetEmailWorkflowExecutionsByWorkflow returns all executions for a workflow
func (db *PostgresDB) GetEmailWorkflowExecutionsByWorkflow(ctx context.Context, workflowID uuid.UUID, limit int) ([]EmailWorkflowExecution, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, tenant_id, workflow_id, recipient, status, scheduled_at, sent_at, error, retry_count, last_retry_at, email_subject, email_template, created_at, updated_at
		FROM email_workflow_executions
		WHERE workflow_id = $1
		ORDER BY created_at DESC
		LIMIT $2`,
		workflowID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var execs []EmailWorkflowExecution
	for rows.Next() {
		var e EmailWorkflowExecution
		err := rows.Scan(&e.ID, &e.TenantID, &e.WorkflowID, &e.Recipient, &e.Status, &e.ScheduledAt,
			&e.SentAt, &e.Error, &e.RetryCount, &e.LastRetryAt, &e.EmailSubject, &e.EmailTemplate,
			&e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		execs = append(execs, e)
	}
	return execs, rows.Err()
}

// GetEmailWorkflowExecutionsByTenant returns all executions for a tenant
func (db *PostgresDB) GetEmailWorkflowExecutionsByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]EmailWorkflowExecution, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, tenant_id, workflow_id, recipient, status, scheduled_at, sent_at, error, retry_count, last_retry_at, email_subject, email_template, created_at, updated_at
		FROM email_workflow_executions
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2`,
		tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var execs []EmailWorkflowExecution
	for rows.Next() {
		var e EmailWorkflowExecution
		err := rows.Scan(&e.ID, &e.TenantID, &e.WorkflowID, &e.Recipient, &e.Status, &e.ScheduledAt,
			&e.SentAt, &e.Error, &e.RetryCount, &e.LastRetryAt, &e.EmailSubject, &e.EmailTemplate,
			&e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		execs = append(execs, e)
	}
	return execs, rows.Err()
}

// UpdateEmailWorkflowExecution updates an email workflow execution
func (db *PostgresDB) UpdateEmailWorkflowExecution(ctx context.Context, exec *EmailWorkflowExecution) error {
	exec.UpdatedAt = time.Now()
	_, err := db.ExecContext(ctx, `
		UPDATE email_workflow_executions
		SET status = $2, error = $3, retry_count = $4, last_retry_at = $5, updated_at = $6
		WHERE id = $1`,
		exec.ID, exec.Status, exec.Error, exec.RetryCount, exec.LastRetryAt, exec.UpdatedAt)
	return err
}

// MarkEmailWorkflowExecutionSent marks an execution as successfully sent
func (db *PostgresDB) MarkEmailWorkflowExecutionSent(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := db.ExecContext(ctx, `
		UPDATE email_workflow_executions
		SET status = 'sent', sent_at = $2, updated_at = $2
		WHERE id = $1`,
		id, now)
	return err
}

// MarkEmailWorkflowExecutionFailed marks an execution as failed
func (db *PostgresDB) MarkEmailWorkflowExecutionFailed(ctx context.Context, id uuid.UUID, errorMsg string) error {
	now := time.Now()
	_, err := db.ExecContext(ctx, `
		UPDATE email_workflow_executions
		SET status = 'failed', error = $2, retry_count = retry_count + 1, last_retry_at = $3, updated_at = $3
		WHERE id = $1`,
		id, errorMsg, now)
	return err
}

// RetryFailedEmailWorkflowExecutions retries failed executions that are under the retry limit
func (db *PostgresDB) RetryFailedEmailWorkflowExecutions(ctx context.Context, maxRetries int) ([]EmailWorkflowExecution, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, tenant_id, workflow_id, recipient, status, scheduled_at, sent_at, error, retry_count, last_retry_at, email_subject, email_template, created_at, updated_at
		FROM email_workflow_executions
		WHERE status = 'failed' AND retry_count < $1
		ORDER BY last_retry_at ASC`,
		maxRetries)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var execs []EmailWorkflowExecution
	for rows.Next() {
		var e EmailWorkflowExecution
		err := rows.Scan(&e.ID, &e.TenantID, &e.WorkflowID, &e.Recipient, &e.Status, &e.ScheduledAt,
			&e.SentAt, &e.Error, &e.RetryCount, &e.LastRetryAt, &e.EmailSubject, &e.EmailTemplate,
			&e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		execs = append(execs, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	now := time.Now()
	for _, e := range execs {
		_, err := db.ExecContext(ctx, `
			UPDATE email_workflow_executions
			SET status = 'pending', updated_at = $2
			WHERE id = $1`,
			e.ID, now)
		if err != nil {
			return nil, err
		}
	}

	return execs, nil
}

// CleanupOldEmailWorkflowExecutions removes old completed executions
func (db *PostgresDB) CleanupOldEmailWorkflowExecutions(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result, err := db.ExecContext(ctx, `
		DELETE FROM email_workflow_executions
		WHERE status IN ('sent', 'failed', 'cancelled') AND created_at < $1`,
		cutoff)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if rowsAffected > 0 {
		logrus.WithFields(logrus.Fields{
			"deleted_count": rowsAffected,
			"cutoff_date":   cutoff,
		}).Info("Cleaned up old email workflow executions")
	}

	return rowsAffected, nil
}
