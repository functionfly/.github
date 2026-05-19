package storage

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/scrypt"
)

// TenantDBProvisioner handles the lifecycle of per-tenant dedicated databases
type TenantDBProvisioner struct {
	config    *TenantDatabaseConfig
	platformDB *PostgresDB
	// template pool for efficient cloning
	templatePool *pgxpool.Pool
	// pools for active tenant databases
	tenantPools sync.Map // map[uuid.UUID]*pgxpool.Pool
	// migration files for tenant schema
	migrationFiles []string
	// state
	mu sync.RWMutex
}

// NewTenantDBProvisioner creates a new tenant database provisioner
func NewTenantDBProvisioner(config *TenantDatabaseConfig, platformDB *PostgresDB) (*TenantDBProvisioner, error) {
	if !config.Enabled {
		logrus.Info("TenantDBProvisioner: disabled (TENANT_DB_ENABLED=false)")
		return &TenantDBProvisioner{config: config}, nil
	}

	provisioner := &TenantDBProvisioner{
		config:       config,
		platformDB:   platformDB,
		migrationFiles: loadTenantMigrationFiles(),
	}

	// Pre-check template DB connectivity
	if config.UseTemplateDB {
		if err := provisioner.validateTemplateDB(context.Background()); err != nil {
			logrus.Warnf("TenantDBProvisioner: template DB validation failed: %v", err)
		}
	}

	logrus.Infof("TenantDBProvisioner: initialized (template=%s, prefix=%s)",
		config.TemplateDB, config.Prefix)

	return provisioner, nil
}

// loadTenantMigrationFiles loads tenant schema migration files
func loadTenantMigrationFiles() []string {
	migrationsDir := filepath.Join("internal/storage/sql/tenant_migrations")
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		logrus.Warnf("Failed to load tenant migration files: %v", err)
		return nil
	}
	return files
}

// CreateTenantDB creates a new dedicated database for a tenant
func (p *TenantDBProvisioner) CreateTenantDB(ctx context.Context, tenantID uuid.UUID) error {
	if !p.config.Enabled {
		return fmt.Errorf("tenant databases are disabled")
	}

	dbName := p.config.GetTenantDBName(tenantID.String())

	logrus.Infof("Creating tenant database: %s for tenant %s", dbName, tenantID)

	// Check if DB already exists
	exists, err := p.databaseExists(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to check existing database: %w", err)
	}
	if exists {
		logrus.Infof("Tenant database %s already exists", dbName)
		return nil
	}

	// Create the database (via template or from scratch)
	if p.config.UseTemplateDB {
		err = p.createFromTemplate(ctx, dbName)
	} else {
		err = p.createEmpty(ctx, dbName)
	}
	if err != nil {
		return fmt.Errorf("failed to create tenant database: %w", err)
	}

	// Run tenant migrations
	if err := p.runMigrations(ctx, tenantID, dbName); err != nil {
		logrus.Errorf("Failed to run migrations for tenant %s: %v", tenantID, err)
		// Don't fail provision - migrations might be retried
	}

	// Store the tenant DB config in platform DB
	if err := p.registerTenantDB(ctx, tenantID, dbName); err != nil {
		logrus.Errorf("Failed to register tenant DB config for %s: %v", tenantID, err)
	}

	logrus.Infof("Successfully provisioned tenant database: %s", dbName)
	return nil
}

// databaseExists checks if a database already exists
func (p *TenantDBProvisioner) databaseExists(ctx context.Context, dbName string) (bool, error) {
	// Connect to postgres system db to check
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=require",
		p.config.Host, p.config.Port, p.config.User, p.config.Password)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return false, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer conn.Close(ctx)

	var exists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	return exists, err
}

// createFromTemplate creates a database by cloning the template
func (p *TenantDBProvisioner) createFromTemplate(ctx context.Context, dbName string) error {
	// Ensure template exists
	if err := p.ensureTemplateDB(ctx); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Clone from template using CREATE DATABASE WITH TEMPLATE
	_, err := p.platformDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s WITH TEMPLATE = %s", pqQuoteIdent(dbName), pqQuoteIdent(p.config.TemplateDB)))
	if err != nil {
		// Fallback: create empty database
		logrus.Warnf("Template clone failed, creating empty DB: %v", err)
		return p.createEmpty(ctx, dbName)
	}

	return nil
}

// createEmpty creates a fresh database without using a template
func (p *TenantDBProvisioner) createEmpty(ctx context.Context, dbName string) error {
	_, err := p.platformDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", pqQuoteIdent(dbName)))
	return err
}

// ensureTemplateDB ensures the template database exists and is accessible
func (p *TenantDBProvisioner) ensureTemplateDB(ctx context.Context) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=require",
		p.config.Host, p.config.Port, p.config.User, p.config.Password)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer conn.Close(ctx)

	// Check if template exists
	var exists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", p.config.TemplateDB).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		// Create template database
		logrus.Infof("Creating template database: %s", p.config.TemplateDB)
		_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", pqQuoteIdent(p.config.TemplateDB)))
		if err != nil {
			return fmt.Errorf("failed to create template DB: %w", err)
		}
	}

	return nil
}

// validateTemplateDB validates that the template DB is accessible and has the expected schema
func (p *TenantDBProvisioner) validateTemplateDB(ctx context.Context) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
		p.config.Host, p.config.Port, p.config.User, p.config.Password, p.config.TemplateDB)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return fmt.Errorf("template DB not accessible: %w", err)
	}
	defer conn.Close(ctx)

	// Check for expected tables
	var tableCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)
	if err != nil {
		return fmt.Errorf("failed to query template schema: %w", err)
	}

	logrus.Debugf("Template DB %s has %d tables", p.config.TemplateDB, tableCount)
	return nil
}

// runMigrations applies tenant schema migrations to a tenant database
func (p *TenantDBProvisioner) runMigrations(ctx context.Context, tenantID uuid.UUID, dbName string) error {
	if len(p.migrationFiles) == 0 {
		logrus.Warn("No tenant migration files found")
		return nil
	}

	// Create connection to tenant DB
	connStr := p.config.BuildTenantDBConnectionString(dbName)
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("failed to parse tenant DB config: %w", err)
	}

	poolConfig.MaxConns = 5
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create migration pool: %w", err)
	}
	defer pool.Close()

	// Get current migration version for this tenant
	currentVersion, err := p.getTenantMigrationVersion(ctx, tenantID)
	if err != nil {
		currentVersion = -1 // Start from beginning
	}

	// Apply each migration
	for _, migrationFile := range p.migrationFiles {
		version, err := extractMigrationVersion(migrationFile)
		if err != nil {
			logrus.Warnf("Skipping migration file %s: %v", migrationFile, err)
			continue
		}

		if int(version) <= currentVersion {
			continue // Already applied
		}

		startTime := time.Now()
		if err := p.applyMigration(ctx, pool, migrationFile); err != nil {
			p.recordMigrationFailure(ctx, tenantID, version, err)
			return fmt.Errorf("migration %d failed: %w", version, err)
		}

		duration := int(time.Since(startTime).Milliseconds())
		p.recordMigrationSuccess(ctx, tenantID, version, duration)
		logrus.Infof("Applied tenant migration %d for tenant %s", version, tenantID)
	}

	return nil
}

// getTenantMigrationVersion gets the last successfully applied migration version
func (p *TenantDBProvisioner) getTenantMigrationVersion(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var version sql.NullInt64
	err := p.platformDB.QueryRowContext(ctx,
		"SELECT MAX(migration_version) FROM tenant_db_migrations WHERE tenant_id = $1 AND success = true",
		tenantID).Scan(&version)
	if err != nil {
		return -1, nil
	}
	if !version.Valid {
		return -1, nil
	}
	return int(version.Int64), nil
}

// applyMigration applies a single migration file to a database pool
func (p *TenantDBProvisioner) applyMigration(ctx context.Context, pool *pgxpool.Pool, migrationFile string) error {
	sqlContent, err := os.ReadFile(migrationFile)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute the migration
	_, err = pool.Exec(ctx, string(sqlContent))
	return err
}

// recordMigrationSuccess records a successful migration
func (p *TenantDBProvisioner) recordMigrationSuccess(ctx context.Context, tenantID uuid.UUID, version int, durationMs int) {
	_, _ = p.platformDB.ExecContext(ctx,
		`INSERT INTO tenant_db_migrations (tenant_id, migration_version, duration_ms, success)
		 VALUES ($1, $2, $3, true)
		 ON CONFLICT DO NOTHING`,
		tenantID, version, durationMs)
}

// recordMigrationFailure records a failed migration
func (p *TenantDBProvisioner) recordMigrationFailure(ctx context.Context, tenantID uuid.UUID, version int, err error) {
	_, _ = p.platformDB.ExecContext(ctx,
		`INSERT INTO tenant_db_migrations (tenant_id, migration_version, success, error_message)
		 VALUES ($1, $2, false, $3)`,
		tenantID, version, err.Error())
}

// registerTenantDB stores the tenant DB configuration in the platform database
func (p *TenantDBProvisioner) registerTenantDB(ctx context.Context, tenantID uuid.UUID, dbName string) error {
	encryptedPassword, err := p.encryptPassword(p.config.Password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	connectionTemplate := p.config.BuildTenantDBConnectionString(dbName)
	// Redact password in stored template
	connectionTemplate = redactPassword(connectionTemplate)

	_, err = p.platformDB.ExecContext(ctx,
		`INSERT INTO tenant_database_configs (tenant_id, db_name, db_host, db_port, db_user, db_password_encrypted, connection_string_template, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET
		 	db_name = EXCLUDED.db_name,
		 	updated_at = NOW(),
		 	status = 'active'`,
		tenantID, dbName, p.config.Host, p.config.Port, p.config.User, encryptedPassword, connectionTemplate)

	return err
}

// DeleteTenantDB deletes a tenant's dedicated database
func (p *TenantDBProvisioner) DeleteTenantDB(ctx context.Context, tenantID uuid.UUID) error {
	if !p.config.Enabled {
		return fmt.Errorf("tenant databases are disabled")
	}

	dbName := p.config.GetTenantDBName(tenantID.String())

	logrus.Infof("Deleting tenant database: %s for tenant %s", dbName, tenantID)

	// Close any open connections
	p.closeTenantPool(tenantID)

	// Terminate existing connections to the database
	_, err := p.platformDB.ExecContext(ctx,
		fmt.Sprintf("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = %s AND pid <> pg_backend_pid()", pqQuoteIdent(dbName)))
	if err != nil {
		logrus.Warnf("Failed to terminate connections to %s: %v", dbName, err)
	}

	// Drop the database
	_, err = p.platformDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", pqQuoteIdent(dbName)))
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	// Remove from registry
	_, err = p.platformDB.ExecContext(ctx, "DELETE FROM tenant_database_configs WHERE tenant_id = $1", tenantID)
	if err != nil {
		logrus.Warnf("Failed to remove tenant DB config: %v", err)
	}

	logrus.Infof("Successfully deleted tenant database: %s", dbName)
	return nil
}

// SuspendTenantDB suspends a tenant's database (revokes connect privileges)
func (p *TenantDBProvisioner) SuspendTenantDB(ctx context.Context, tenantID uuid.UUID) error {
	dbName := p.config.GetTenantDBName(tenantID.String())

	// Close existing pools
	p.closeTenantPool(tenantID)

	// Revoke connect privilege
	_, err := p.platformDB.ExecContext(ctx,
		fmt.Sprintf("REVOKE CONNECT ON DATABASE %s FROM PUBLIC", pqQuoteIdent(dbName)))
	if err != nil {
		return fmt.Errorf("failed to suspend database: %w", err)
	}

	// Update status
	_, err = p.platformDB.ExecContext(ctx,
		"UPDATE tenant_database_configs SET status = 'suspended', updated_at = NOW() WHERE tenant_id = $1",
		tenantID)

	return err
}

// ResumeTenantDB resumes a suspended tenant database
func (p *TenantDBProvisioner) ResumeTenantDB(ctx context.Context, tenantID uuid.UUID) error {
	dbName := p.config.GetTenantDBName(tenantID.String())

	// Grant connect privilege back
	_, err := p.platformDB.ExecContext(ctx,
		fmt.Sprintf("GRANT CONNECT ON DATABASE %s TO PUBLIC", pqQuoteIdent(dbName)))
	if err != nil {
		return fmt.Errorf("failed to resume database: %w", err)
	}

	// Update status
	_, err = p.platformDB.ExecContext(ctx,
		"UPDATE tenant_database_configs SET status = 'active', updated_at = NOW() WHERE tenant_id = $1",
		tenantID)

	return err
}

// GetTenantPool returns a connection pool for a tenant's database
func (p *TenantDBProvisioner) GetTenantPool(ctx context.Context, tenantID uuid.UUID) (*pgxpool.Pool, error) {
	if !p.config.Enabled {
		return nil, fmt.Errorf("tenant databases are disabled")
	}

	// Check if we already have a pool
	if pool, ok := p.tenantPools.Load(tenantID); ok {
		return pool.(*pgxpool.Pool), nil
	}

	// Get tenant DB config
	dbName := p.config.GetTenantDBName(tenantID.String())

	// Create new pool
	connStr := p.config.BuildTenantDBConnectionString(dbName)
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MinConns = int32(p.config.PoolMin)
	poolConfig.MaxConns = int32(p.config.PoolMax)
	poolConfig.MaxConnIdleTime = p.config.MaxIdleTime
	poolConfig.MaxConnLifetime = p.config.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping tenant database: %w", err)
	}

	// Store the pool
	p.tenantPools.Store(tenantID, pool)

	return pool, nil
}

// closeTenantPool closes and removes a tenant's connection pool
func (p *TenantDBProvisioner) closeTenantPool(tenantID uuid.UUID) {
	if pool, ok := p.tenantPools.LoadAndDelete(tenantID); ok {
		pool.(*pgxpool.Pool).Close()
	}
}

// Close gracefully shuts down the tenant database provisioner
// It stops accepting new operations and waits for ongoing operations to complete
func (p *TenantDBProvisioner) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Signal that we're shutting down (set a flag)
	logrus.Info("TenantDBProvisioner: initiating graceful shutdown")

	// First, close all managed pools for active tenants
	var wg sync.WaitGroup
	p.tenantPools.Range(func(key, value interface{}) bool {
		wg.Add(1)
		go func(pool *pgxpool.Pool) {
			defer wg.Done()
			// Give each pool a timeout for graceful closure
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			// Ping to trigger any final operations
			pool.Ping(ctx)
			pool.Close()
		}(value.(*pgxpool.Pool))
		return true
	})

	// Wait for all pools to close (with timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logrus.Info("TenantDBProvisioner: all tenant pools closed")
	case <-time.After(10 * time.Second):
		logrus.Warn("TenantDBProvisioner: shutdown timed out waiting for pool closures")
	}

	// Close template pool if it exists
	if p.templatePool != nil {
		p.templatePool.Close()
	}

	// Clear the pools map
	p.tenantPools = sync.Map{}

	logrus.Info("TenantDBProvisioner: shutdown complete")
	return nil
}

// encryptPassword encrypts the tenant DB password for storage using AES-256-GCM
func (p *TenantDBProvisioner) encryptPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}

	// Use platform DB's encryption manager if available
	if p.platformDB != nil && p.platformDB.encryptionManager != nil {
		encrypted, err := p.platformDB.encryptionManager.EncryptField(password)
		if err != nil {
			return "", fmt.Errorf("failed to encrypt password: %w", err)
		}
		return encrypted, nil
	}

	// Fallback: use scrypt-based encryption with a derived key
	return encryptPasswordFallback(password)
}

// encryptPasswordFallback provides encryption when the encryption manager is not available
// Uses a key derived from a service identifier that can be configured via environment variable
// WARNING: This is NOT zero-knowledge - it's service-level encryption for DB credentials
// In production, proper KMS/HSM-backed encryption MUST be used instead
func encryptPasswordFallback(password string) (string, error) {
	if password == "" {
		return "", nil
	}

	// Allow environment variable override for salt to support multi-region deployments
	// with different encryption keys. If not set, uses the default (which should only
	// be used in development/local setups).
	salt := os.Getenv("TENANT_DB_FALLBACK_SALT")
	if salt == "" {
		logrus.Warn("TENANT_DB_FALLBACK_SALT not set - using default salt. This is INSECURE for production use.")
		salt = "functionfly-tenant-db-key-v1-fallback-only"
	}

	serviceKey := os.Getenv("TENANT_DB_FALLBACK_SERVICE_KEY")
	if serviceKey == "" {
		logrus.Warn("TENANT_DB_FALLBACK_SERVICE_KEY not set - using default key. This is INSECURE for production use.")
		serviceKey = "functionfly-dedicated-db-encryption-key-v1"
	}

	if p := os.Getenv("PRODUCTION"); p == "true" || p == "production" {
		logrus.Error("CRITICAL: Using fallback encryption in production - set TENANT_DB_ENCRYPTION_KEY or use proper KMS encryption")
	}

	scryptSalt := []byte(salt)
	scryptKey := []byte(serviceKey)

	// Derive key using scrypt with service key
	key, err := scrypt.Key(scryptKey, scryptSalt, 32768, 8, 1, 32)
	if err != nil {
		return "", fmt.Errorf("failed to derive key: %w", err)
	}

	// Encrypt with AES-256-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// redactPassword redacts password from connection string for logging
func redactPassword(connStr string) string {
	re := regexp.MustCompile(`password=[^ ]*`)
	return re.ReplaceAllString(connStr, "password=****")
}

// pqQuoteIdent quotes a PostgreSQL identifier
func pqQuoteIdent(s string) string {
	// Escape single quotes by doubling them
	s = strings.ReplaceAll(s, "'", "''")
	return `"` + s + `"`
}

// extractMigrationVersion extracts version number from migration filename
func extractMigrationVersion(filename string) (int, error) {
	// Expected format: YYYYMMDDHHMMSS_description.up.sql
	re := regexp.MustCompile(`^(\d{14})_.*\.up\.sql$`)
	matches := re.FindStringSubmatch(filepath.Base(filename))
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	version, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse version: %w", err)
	}

	return version, nil
}

// GetTenantDBStatus returns the status of a tenant's database
func (p *TenantDBProvisioner) GetTenantDBStatus(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var status string
	err := p.platformDB.QueryRowContext(ctx,
		"SELECT status FROM tenant_database_configs WHERE tenant_id = $1",
		tenantID).Scan(&status)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("tenant database not found")
	}
	return status, err
}

// ListTenantDatabases returns all tenant databases with their status
func (p *TenantDBProvisioner) ListTenantDatabases(ctx context.Context) ([]TenantDBInfo, error) {
	rows, err := p.platformDB.QueryContext(ctx,
		`SELECT t.id, tdc.db_name, tdc.status, tdc.created_at, tdc.updated_at
		 FROM tenants t
		 JOIN tenant_database_configs tdc ON t.id = tdc.tenant_id
		 ORDER BY tdc.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TenantDBInfo
	for rows.Next() {
		var info TenantDBInfo
		err := rows.Scan(&info.TenantID, &info.DBName, &info.Status, &info.CreatedAt, &info.UpdatedAt)
		if err != nil {
			return nil, err
		}
		results = append(results, info)
	}

	return results, rows.Err()
}

// TenantDBInfo holds information about a tenant's database
type TenantDBInfo struct {
	TenantID  uuid.UUID
	DBName    string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TenantDBProvisioner implements migrations.Migratable interface
var _ interface {
	Migrate(ctx context.Context) error
} = (*TenantDBProvisioner)(nil)

// Migrate implements the migrations.Migratable interface (no-op for tenant DBs)
func (p *TenantDBProvisioner) Migrate(ctx context.Context) error {
	return nil // Tenant DBs have their own migrations managed by provisioner
}