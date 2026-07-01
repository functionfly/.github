package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ============================================================================
// Prepared Statement Management
// ============================================================================

// InitPreparedStatements initializes prepared statements. Call after migrations when the DB
// was created with skipPreparedStatements=true. No-op if already initialized.
func (db *PostgresDB) InitPreparedStatements(ctx context.Context) error {
	db.stmtMutex.Lock()
	already := len(db.preparedStatements) > 0
	db.stmtMutex.Unlock()
	if already {
		return nil
	}
	return db.initPreparedStatements(ctx)
}

// initPreparedStatements initializes commonly used prepared statements with enhanced management
func (db *PostgresDB) initPreparedStatements(ctx context.Context) error {
	db.stmtMutex.Lock()
	defer db.stmtMutex.Unlock()

	// Define prepared statements with their queries
	statements := map[string]string{
		"getUserByEmail": `
			SELECT id, tenant_id, username, email, password_hash, role, email_verified, company_name, verification_token, verification_expires_at,
			       provider, provider_id, provider_data, mfa_secret, mfa_enabled, mfa_backup_codes, mfa_last_used,
			       created_at, updated_at, name, bio, last_active_at, profile_number
			FROM users WHERE email = $1`,
		"getUserByID": `
			SELECT id, tenant_id, username, email, password_hash, role, email_verified, company_name, verification_token, verification_expires_at,
			       provider, provider_id, provider_data, mfa_secret, mfa_enabled, mfa_backup_codes, mfa_last_used,
			       created_at, updated_at, name, bio, last_active_at, profile_number,
			       location, website, job_title, twitter_url, github_url, linkedin_url,
			       is_founder, founder_number
			FROM users WHERE id = $1`,
		"getTenantByID": `
			SELECT id, name, plan, status, created_at, updated_at
			FROM tenants WHERE id = $1`,
		"getAppByID": `
			SELECT id, tenant_id, name, slug, created_at, updated_at
			FROM apps WHERE id = $1`,
		"getAppBySlug": `
			SELECT id, tenant_id, name, slug, created_at, updated_at
			FROM apps WHERE slug = $1`,
		"listAppsByTenant": `
			SELECT id, tenant_id, name, slug, created_at, updated_at
			FROM apps WHERE tenant_id = $1 ORDER BY created_at DESC`,
		"getBackendByID": `
			SELECT id, app_id, provider, region, url, shared_secret, enabled, priority, created_at, updated_at
			FROM backends WHERE id = $1`,
		"listBackendsByApp": `
			SELECT id, app_id, provider, region, url, shared_secret, enabled, priority, created_at, updated_at
			FROM backends WHERE app_id = $1 ORDER BY priority ASC NULLS LAST, created_at DESC`,
		"getDeploymentByID": `
			SELECT id, app_id, provider, region, deployment_id, status, artifact_key, routes, message, metadata, created_at, updated_at
			FROM deployments WHERE id = $1`,
		"listDeploymentsByApp": `
			SELECT id, app_id, provider, region, deployment_id, status, artifact_key, routes, message, metadata, created_at, updated_at
			FROM deployments WHERE app_id = $1 ORDER BY created_at DESC LIMIT $2`,
		// Registry function queries
		"getRegistryFunctionByID": `
			SELECT id, author, name, latest_version, title, description, category, tags, visibility,
			       price_per_call, popularity_score, reliability_score, deterministic_score, capabilities,
			       tenant_id, owner_user_id, created_at, updated_at
			FROM registry_functions WHERE id = $1`,
		"getRegistryFunctionByAuthorName": `
			SELECT id, author, name, latest_version, title, description, category, tags, visibility,
			       price_per_call, popularity_score, reliability_score, deterministic_score, capabilities,
			       tenant_id, owner_user_id, created_at, updated_at
			FROM registry_functions WHERE author = $1 AND name = $2`,
		"searchRegistryFunctions": `
			SELECT id, author, name, description, popularity_score, COALESCE(reliability_score, 0) AS trust_score, created_at
			FROM registry_functions
			WHERE visibility = 'public' AND (name ILIKE $1 OR description ILIKE $1)
			ORDER BY COALESCE(reliability_score, 0) DESC, popularity_score DESC
			LIMIT $2 OFFSET $3`,
	}

	// Prepare all statements
	for name, query := range statements {
		stmt, err := db.PrepareContext(ctx, query)
		if err != nil {
			db.closePreparedStatementsLocked() // Clean up on error; already hold stmtMutex
			return fmt.Errorf("failed to prepare statement %s: %w", name, err)
		}
		db.preparedStatements[name] = stmt

		// Set legacy statement pointers for backward compatibility
		switch name {
		case "getUserByEmail":
			db.stmtGetUserByEmail = stmt
		case "getUserByID":
			db.stmtGetUserByID = stmt
		case "getTenantByID":
			db.stmtGetTenantByID = stmt
		case "getAppByID":
			db.stmtGetAppByID = stmt
		case "getAppBySlug":
			db.stmtGetAppBySlug = stmt
		case "listAppsByTenant":
			db.stmtListAppsByTenant = stmt
		case "getBackendByID":
			db.stmtGetBackendByID = stmt
		case "listBackendsByApp":
			db.stmtListBackendsByApp = stmt
		case "getDeploymentByID":
			db.stmtGetDeploymentByID = stmt
		case "listDeploymentsByApp":
			db.stmtListDeploymentsByApp = stmt
		}
	}

	// Start prepared statement re-preparation ticker
	db.stmtReprepareTicker = time.NewTicker(1 * time.Hour) // Reprepare every hour
	db.lastStmtReprepare = time.Now()

	go db.reprepareStatementsRoutine()

	// Mark initialization as successful before releasing lock
	// This ensures the goroutine only runs when initialization completed successfully
	db.stmtInitialized = true

	logrus.Info("Initialized enhanced prepared statement management")
	return nil
}

// reprepareStatementsRoutine periodically re-prepares statements to prevent staleness
func (db *PostgresDB) reprepareStatementsRoutine() {
	for range db.stmtReprepareTicker.C {
		// Guard against orphaned goroutine: if initialization didn't complete,
		// the stmtInitialized flag will be false and we exit gracefully
		db.stmtMutex.RLock()
		initialized := db.stmtInitialized
		db.stmtMutex.RUnlock()

		if !initialized {
			logrus.Debug("Prepared statement re-prepare goroutine exiting: initialization did not complete")
			return
		}

		if err := db.reprepareStatements(); err != nil {
			logrus.WithError(err).Warn("Failed to re-prepare statements")
		} else {
			logrus.Debug("Successfully re-prepared database statements")
		}
	}
}

// reprepareStatements re-prepares all cached statements
func (db *PostgresDB) reprepareStatements() error {
	db.stmtMutex.Lock()
	defer db.stmtMutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Re-prepare each statement
	for name, oldStmt := range db.preparedStatements {
		newStmt, err := db.PrepareContext(ctx, db.getStatementQuery(name))
		if err != nil {
			return fmt.Errorf("failed to re-prepare statement %s: %w", name, err)
		}

		// Close old statement and replace
		oldStmt.Close()
		db.preparedStatements[name] = newStmt

		// Update legacy pointers
		switch name {
		case "getUserByEmail":
			db.stmtGetUserByEmail = newStmt
		case "getUserByID":
			db.stmtGetUserByID = newStmt
		case "getTenantByID":
			db.stmtGetTenantByID = newStmt
		case "getAppByID":
			db.stmtGetAppByID = newStmt
		case "getAppBySlug":
			db.stmtGetAppBySlug = newStmt
		case "listAppsByTenant":
			db.stmtListAppsByTenant = newStmt
		case "getBackendByID":
			db.stmtGetBackendByID = newStmt
		case "listBackendsByApp":
			db.stmtListBackendsByApp = newStmt
		case "getDeploymentByID":
			db.stmtGetDeploymentByID = newStmt
		case "listDeploymentsByApp":
			db.stmtListDeploymentsByApp = newStmt
		}
	}

	db.lastStmtReprepare = time.Now()
	return nil
}

// GetStatementQuery returns the SQL query for a given statement name (for fallback when prepared stmt is nil)
func (db *PostgresDB) GetStatementQuery(name string) string {
	return db.getStatementQuery(name)
}

	// getStatementQuery returns the SQL query for a given statement name
func (db *PostgresDB) getStatementQuery(name string) string {
	queries := map[string]string{
		"getUserByEmail": `
			SELECT id, tenant_id, username, email, password_hash, role, email_verified, company_name, verification_token, verification_expires_at,
			       provider, provider_id, provider_data, mfa_secret, mfa_enabled, mfa_backup_codes, mfa_last_used,
			       created_at, updated_at, name, bio, last_active_at, profile_number
			FROM users WHERE email = $1`,
		"getUserByID": `
			SELECT id, tenant_id, username, email, password_hash, role, email_verified, company_name, verification_token, verification_expires_at,
			       provider, provider_id, provider_data, mfa_secret, mfa_enabled, mfa_backup_codes, mfa_last_used,
			       created_at, updated_at, name, bio, last_active_at, profile_number,
			       location, website, job_title, twitter_url, github_url, linkedin_url,
			       is_founder, founder_number
			FROM users WHERE id = $1`,
		"getTenantByID": `
			SELECT id, name, plan, status, created_at, updated_at
			FROM tenants WHERE id = $1`,
		"getAppByID": `
			SELECT id, tenant_id, name, slug, created_at, updated_at
			FROM apps WHERE id = $1`,
		"getAppBySlug": `
			SELECT id, tenant_id, name, slug, created_at, updated_at
			FROM apps WHERE slug = $1`,
		"listAppsByTenant": `
			SELECT id, tenant_id, name, slug, created_at, updated_at
			FROM apps WHERE tenant_id = $1 ORDER BY created_at DESC`,
		"getBackendByID": `
			SELECT id, app_id, provider, region, url, shared_secret, enabled, priority, created_at, updated_at
			FROM backends WHERE id = $1`,
		"listBackendsByApp": `
			SELECT id, app_id, provider, region, url, shared_secret, enabled, priority, created_at, updated_at
			FROM backends WHERE app_id = $1 ORDER BY priority ASC NULLS LAST, created_at DESC`,
		"getDeploymentByID": `
			SELECT id, app_id, provider, region, deployment_id, status, artifact_key, routes, message, metadata, created_at, updated_at
			FROM deployments WHERE id = $1`,
		"listDeploymentsByApp": `
			SELECT id, app_id, provider, region, deployment_id, status, artifact_key, routes, message, metadata, created_at, updated_at
			FROM deployments WHERE app_id = $1 ORDER BY created_at DESC LIMIT $2`,
		"getRegistryFunctionByID": `
			SELECT id, author, name, latest_version, title, description, category, tags, visibility,
			       price_per_call, popularity_score, reliability_score, deterministic_score, capabilities,
			       tenant_id, owner_user_id, created_at, updated_at
			FROM registry_functions WHERE id = $1`,
		"getRegistryFunctionByAuthorName": `
			SELECT id, author, name, latest_version, title, description, category, tags, visibility,
			       price_per_call, popularity_score, reliability_score, deterministic_score, capabilities,
			       tenant_id, owner_user_id, created_at, updated_at
			FROM registry_functions WHERE author = $1 AND name = $2`,
		"searchRegistryFunctions": `
			SELECT id, author, name, description, popularity_score, COALESCE(reliability_score, 0) AS trust_score, created_at
			FROM registry_functions
			WHERE visibility = 'public' AND (name ILIKE $1 OR description ILIKE $1)
			ORDER BY COALESCE(reliability_score, 0) DESC, popularity_score DESC
			LIMIT $2 OFFSET $3`,
	}

	return queries[name]
}

// GetPreparedStatement returns a prepared statement by name (thread-safe)
func (db *PostgresDB) GetPreparedStatement(name string) (*sql.Stmt, error) {
	db.stmtMutex.RLock()
	stmt, exists := db.preparedStatements[name]
	db.stmtMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("prepared statement %s not found", name)
	}

	return stmt, nil
}

// closePreparedStatements closes all prepared statements (takes stmtMutex).
func (db *PostgresDB) closePreparedStatements() {
	db.stmtMutex.Lock()
	defer db.stmtMutex.Unlock()
	db.closePreparedStatementsLocked()
}

// closePreparedStatementsLocked does the actual close; caller must hold db.stmtMutex.
// Used from initPreparedStatements on error to avoid deadlock (init already holds the lock).
func (db *PostgresDB) closePreparedStatementsLocked() {
	// Stop re-preparation ticker
	if db.stmtReprepareTicker != nil {
		db.stmtReprepareTicker.Stop()
	}

	// Mark as not initialized so orphaned goroutine exits
	db.stmtInitialized = false

	// Close all prepared statements
	for name, stmt := range db.preparedStatements {
		if stmt != nil {
			if err := stmt.Close(); err != nil {
				logrus.WithError(err).Warnf("Failed to close prepared statement: %s", name)
			}
		}
	}

	// Clear the map
	db.preparedStatements = make(map[string]*sql.Stmt)

	// Clear legacy pointers
	db.stmtGetUserByEmail = nil
	db.stmtGetUserByID = nil
	db.stmtGetTenantByID = nil
	db.stmtGetAppByID = nil
	db.stmtGetAppBySlug = nil
	db.stmtListAppsByTenant = nil
	db.stmtGetBackendByID = nil
	db.stmtListBackendsByApp = nil
	db.stmtGetDeploymentByID = nil
	db.stmtListDeploymentsByApp = nil
}
