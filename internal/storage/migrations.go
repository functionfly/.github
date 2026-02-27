package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// MigrationResult represents the result of a migration operation
type MigrationResult struct {
	Success          bool
	MigrationsRun    []string
	Duration         time.Duration
	ValidationErrors []string
	RollbackTested   bool
}

// RunMigrationsWithValidation runs all pending migrations with post-migration validation.
// Uses a dedicated DB connection for the migrate library so the main db is never closed.
func RunMigrationsWithValidation(db *PostgresDB) (*MigrationResult, error) {
	startTime := time.Now()
	result := &MigrationResult{
		MigrationsRun:    make([]string, 0),
		ValidationErrors: make([]string, 0),
	}

	// Ensure schema_migrations table is in golang-migrate format (version, dirty) before using the library
	if err := ensureSchemaMigrationsTableForMigrate(db.DB); err != nil {
		return nil, fmt.Errorf("failed to ensure schema_migrations format: %w", err)
	}

	// Get the migration directory path
	migrationPath, err := getMigrationPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get migration path: %w", err)
	}

	// Use a dedicated connection for migrations so m.Close() does not close the main db
	config, err := loadDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}
	migrateConn, err := sql.Open("postgres", buildConnectionString(config))
	if err != nil {
		return nil, fmt.Errorf("failed to open migration connection: %w", err)
	}
	defer migrateConn.Close()

	m, err := createMigrateInstance(migrateConn, migrationPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// If schema is dirty or at version 0 (bad state from a previous failed repair), reset to a valid version so Up() can run
	if curVer, dirty, verErr := m.Version(); verErr == nil && (dirty || curVer == 0) {
		available, listErr := GetAvailableMigrations()
		if listErr != nil {
			return nil, fmt.Errorf("failed to list migrations for repair: %w", listErr)
		}
		var prevVer uint
		if dirty {
			// Dirty at curVer: set to previous migration so the failed one re-runs
			for _, mig := range available {
				v, _ := strconv.ParseUint(mig.Version, 10, 64)
				if v < uint64(curVer) && uint(v) > prevVer {
					prevVer = uint(v)
				}
			}
		} else {
			// Version 0 with no migration file for 0: set to second-to-last so only last migration runs (or -1 if single migration)
			var versions []uint
			for _, mig := range available {
				v, _ := strconv.ParseUint(mig.Version, 10, 64)
				versions = append(versions, uint(v))
			}
			if len(versions) >= 2 {
				sort.Slice(versions, func(i, j int) bool { return versions[i] > versions[j] })
				prevVer = versions[1]
			}
			// else len==1: prevVer stays 0; we'll Force(-1) below so Up() runs the single migration
		}
		forceVer := int(prevVer)
		if curVer == 0 && prevVer == 0 && len(available) > 0 {
			forceVer = -1 // no migrations applied; Up() will run from first
		}
		if err := m.Force(forceVer); err != nil {
			return nil, fmt.Errorf("failed to repair migration state (force version %d): %w", forceVer, err)
		}
		logrus.WithField("repaired_to_version", forceVer).Info("Repaired migration state; migrations will re-run")
	}

	// Get migrations to be applied
	available, err := GetAvailableMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get available migrations: %w", err)
	}

	// Build appliedBefore from current version (avoid GetAppliedMigrations: it creates a second
	// Migrate that closes the shared DB when it returns)
	curVerBefore, _, verErr := m.Version()
	if verErr != nil && verErr != migrate.ErrNilVersion {
		return nil, fmt.Errorf("failed to get migration version: %w", verErr)
	}
	appliedBefore := make(map[string]int)
	if verErr != migrate.ErrNilVersion {
		for _, migration := range available {
			v, _ := strconv.ParseUint(migration.Version, 10, 64)
			if uint64(curVerBefore) >= v {
				appliedBefore[migration.Version] = 1
			}
		}
	}

	logrus.Info("Starting migration process with validation")

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		result.Success = false
		return result, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Build appliedAfter from current version (do not call GetAppliedMigrations again: it creates
	// a second Migrate that closes the shared DB when it returns)
	curVer, _, verErr := m.Version()
	if verErr != nil && verErr != migrate.ErrNilVersion {
		return nil, fmt.Errorf("failed to get version after migrations: %w", verErr)
	}
	appliedAfter := make(map[string]int)
	for _, migration := range available {
		v, _ := strconv.ParseUint(migration.Version, 10, 64)
		if uint64(curVer) >= v {
			appliedAfter[migration.Version] = 1
		}
	}

	// Determine which migrations were run
	for _, migration := range available {
		if _, wasApplied := appliedBefore[migration.Version]; !wasApplied {
			if _, isApplied := appliedAfter[migration.Version]; isApplied {
				result.MigrationsRun = append(result.MigrationsRun, migration.Version)
			}
		}
	}

	logrus.WithField("migrations_run", len(result.MigrationsRun)).Info("Migrations completed, starting validation")

	// Run post-migration validation
	validationErrors, err := runPostMigrationValidation(db)
	if err != nil {
		result.Success = false
		return result, fmt.Errorf("post-migration validation failed: %w", err)
	}
	result.ValidationErrors = validationErrors

	// Test rollback if enabled and we have migrations to test
	if len(result.MigrationsRun) > 0 && shouldTestRollback() {
		logrus.Info("Testing rollback capability")
		if err := testMigrationRollback(db, migrationPath, result.MigrationsRun); err != nil {
			logrus.WithError(err).Warn("Migration rollback test failed")
			result.RollbackTested = false
		} else {
			result.RollbackTested = true
			logrus.Info("Migration rollback test passed")
		}
	}

	result.Duration = time.Since(startTime)
	result.Success = len(result.ValidationErrors) == 0

	if result.Success {
		logrus.WithFields(logrus.Fields{
			"migrations_run": len(result.MigrationsRun),
			"duration":       result.Duration,
		}).Info("Migration process completed successfully")
	} else {
		logrus.WithFields(logrus.Fields{
			"migrations_run":    len(result.MigrationsRun),
			"validation_errors": len(result.ValidationErrors),
			"duration":          result.Duration,
		}).Warn("Migration process completed with validation errors")
	}

	return result, nil
}

// RunMigrations runs all pending migrations using golang-migrate (backward compatibility).
// If SKIP_MIGRATION_VALIDATION=true (e.g. in development), validation failures are logged but do not fail the run.
func RunMigrations(db *PostgresDB) error {
	result, err := RunMigrationsWithValidation(db)
	if err != nil {
		return err
	}
	if !result.Success {
		detail := strings.Join(result.ValidationErrors, "; ")
		for _, msg := range result.ValidationErrors {
			logrus.Warn("Migration validation: " + msg)
		}
		if os.Getenv("SKIP_MIGRATION_VALIDATION") == "true" {
			logrus.Warn("Continuing despite validation errors (SKIP_MIGRATION_VALIDATION=true)")
			return nil
		}
		return fmt.Errorf("migrations completed with %d validation errors: %s", len(result.ValidationErrors), detail)
	}
	return nil
}

// ensureSchemaMigrationsTableForMigrate ensures schema_migrations has (version, dirty) for golang-migrate.
// If the table has the old format (version, applied_at) or any other format, it is converted to the new format.
func ensureSchemaMigrationsTableForMigrate(db *sql.DB) error {
	// First try the same query the migrate library uses; if it works, we're done.
	var version int64
	var dirty bool
	err := db.QueryRow(`SELECT version, dirty FROM schema_migrations LIMIT 1`).Scan(&version, &dirty)
	if err == nil {
		return nil
	}
	// If table does not exist, let the migrate library create it.
	if err == sql.ErrNoRows {
		return nil
	}
	// Check for undefined_table (table missing)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code.Name() == "undefined_table" {
		return nil
	}
	// Column missing (e.g. dirty or applied_at) or other schema mismatch: convert to (version, dirty).
	logrus.Info("Converting schema_migrations to golang-migrate format (version, dirty)")
	var maxVersion int64
	if err := db.QueryRow(`SELECT COALESCE(MAX(CAST(version AS BIGINT)), 0) FROM schema_migrations`).Scan(&maxVersion); err != nil {
		return fmt.Errorf("failed to read current migration version: %w", err)
	}
	if _, err := db.Exec(`ALTER TABLE schema_migrations RENAME TO schema_migrations_old`); err != nil {
		return fmt.Errorf("failed to rename old schema_migrations: %w", err)
	}
	if _, err := db.Exec(`CREATE TABLE schema_migrations (version BIGINT NOT NULL PRIMARY KEY, dirty BOOLEAN NOT NULL DEFAULT FALSE)`); err != nil {
		return fmt.Errorf("failed to create new schema_migrations: %w", err)
	}
	if _, err := db.Exec(`INSERT INTO schema_migrations (version, dirty) VALUES ($1, FALSE)`, maxVersion); err != nil {
		return fmt.Errorf("failed to copy version: %w", err)
	}
	logrus.WithField("version", maxVersion).Info("schema_migrations converted successfully")
	return nil
}

// GetAppliedMigrations returns a map of applied migration versions (for CLI compatibility)
func GetAppliedMigrations(db *sql.DB) (map[string]int, error) {
	if err := ensureSchemaMigrationsTableForMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to ensure schema_migrations format: %w", err)
	}

	migrationPath, err := getMigrationPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get migration path: %w", err)
	}

	m, err := createMigrateInstance(db, migrationPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return nil, fmt.Errorf("failed to get migration version: %w", err)
	}

	applied := make(map[string]int)
	if err != migrate.ErrNilVersion {
		applied[strconv.FormatUint(uint64(version), 10)] = 1
		if dirty {
			return nil, fmt.Errorf("database schema is dirty at version %d", version)
		}
	}

	return applied, nil
}

// migrationPathToFS returns a filesystem path for Glob/ReadFile (strips file:// prefix).
func migrationPathToFS(migrationPath string) string {
	const prefix = "file://"
	if strings.HasPrefix(migrationPath, prefix) {
		return migrationPath[len(prefix):]
	}
	return migrationPath
}

// GetAvailableMigrations returns list of available migrations (for CLI compatibility)
func GetAvailableMigrations() ([]Migration, error) {
	migrationPath, err := getMigrationPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get migration path: %w", err)
	}
	fsPath := migrationPathToFS(migrationPath)

	// Read migration files (use filesystem path; migrationPath may be file:// URL)
	files, err := filepath.Glob(filepath.Join(fsPath, "*.up.sql"))
	if err != nil {
		return nil, fmt.Errorf("failed to read migration files: %w", err)
	}

	var migrations []Migration
	for _, file := range files {
		filename := filepath.Base(file)
		version, err := parseGolangMigrateFilename(filename)
		if err != nil {
			continue // Skip invalid files
		}

		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		// Check for down migration (file is under fsPath)
		downFile := filepath.Join(fsPath, strings.Replace(filename, ".up.sql", ".down.sql", 1))
		var downContent string
		if downData, err := os.ReadFile(downFile); err == nil {
			downContent = string(downData)
		}

		migrations = append(migrations, Migration{
			Version:     version,
			Filename:    filename,
			Content:     string(content),
			DownContent: downContent,
		})
	}

	return migrations, nil
}

// RollbackMigration rolls back the latest migration (for CLI compatibility)
func RollbackMigration(db *sql.DB, migration Migration) error {
	migrationPath, err := getMigrationPath()
	if err != nil {
		return fmt.Errorf("failed to get migration path: %w", err)
	}

	m, err := createMigrateInstance(db, migrationPath)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Rollback one migration
	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// Migration represents a database migration (for CLI compatibility)
type Migration struct {
	Version     string
	Filename    string
	Content     string
	DownContent string
	AppliedAt   interface{} // Not used in golang-migrate
}

// Helper functions

func getMigrationPath() (string, error) {
	// Assume migrations are in a "migrations" directory relative to the working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	migrationPath := filepath.Join(wd, "migrations")
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		// Fallback to internal/storage/sql/migrations if root migrations directory doesn't exist
		migrationPath = filepath.Join(wd, "internal", "storage", "sql", "migrations")
		if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
			return "", fmt.Errorf("migrations directory not found in either location: %s or %s", filepath.Join(wd, "migrations"), migrationPath)
		}
	}

	return fmt.Sprintf("file://%s", migrationPath), nil
}

func createMigrateInstance(db *sql.DB, migrationPath string) (*migrate.Migrate, error) {
	// Create postgres driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(migrationPath, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m, nil
}

// runPostMigrationValidation performs validation checks after migrations
func runPostMigrationValidation(db *PostgresDB) ([]string, error) {
	var errors []string

	// Test 1: Check for orphaned records
	if err := validateForeignKeys(db); err != nil {
		errors = append(errors, fmt.Sprintf("Foreign key validation failed: %v", err))
	}

	// Test 2: Check for required indexes
	if err := validateIndexes(db); err != nil {
		errors = append(errors, fmt.Sprintf("Index validation failed: %v", err))
	}

	// Test 3: Check table structure consistency
	if err := validateTableStructure(db); err != nil {
		errors = append(errors, fmt.Sprintf("Table structure validation failed: %v", err))
	}

	// Test 4: Check data integrity
	if err := validateDataIntegrity(db); err != nil {
		errors = append(errors, fmt.Sprintf("Data integrity validation failed: %v", err))
	}

	return errors, nil
}

// validateForeignKeys checks for orphaned records
func validateForeignKeys(db *PostgresDB) error {
	// Check for orphaned registry function versions
	var orphanedVersions int
	err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM registry_function_versions rv
		LEFT JOIN registry_functions rf ON rv.function_id = rf.id
		WHERE rf.id IS NULL
	`).Scan(&orphanedVersions)
	if err != nil {
		return fmt.Errorf("failed to check orphaned versions: %w", err)
	}
	if orphanedVersions > 0 {
		return fmt.Errorf("found %d orphaned function versions", orphanedVersions)
	}

	// Add more foreign key checks as needed
	return nil
}

// validateIndexes checks for required indexes (names must match migration files)
func validateIndexes(db *PostgresDB) error {
	requiredIndexes := []string{
		"idx_registry_functions_author",              // create_function_registry.up.sql
		"idx_registry_function_versions_function_id", // create_function_registry.up.sql
		// users.email uses UNIQUE constraint (creates users_email_key), not idx_users_email
	}

	for _, indexName := range requiredIndexes {
		var exists bool
		err := db.DB.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes
				WHERE indexname = $1
			)
		`, indexName).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check index %s: %w", indexName, err)
		}
		if !exists {
			return fmt.Errorf("required index %s is missing", indexName)
		}
	}

	return nil
}

// validateTableStructure checks table structure consistency
func validateTableStructure(db *PostgresDB) error {
	// Check for required tables
	requiredTables := []string{
		"registry_functions",
		"registry_function_versions",
		"users",
		"tenants",
	}

	for _, tableName := range requiredTables {
		var exists bool
		err := db.DB.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)
		`, tableName).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", tableName, err)
		}
		if !exists {
			return fmt.Errorf("required table %s is missing", tableName)
		}
	}

	return nil
}

// validateDataIntegrity checks data integrity constraints
func validateDataIntegrity(db *PostgresDB) error {
	// Check for functions without versions (should have at least one)
	var functionsWithoutVersions int
	err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM registry_functions rf
		LEFT JOIN registry_function_versions rv ON rf.id = rv.function_id
		WHERE rv.id IS NULL
	`).Scan(&functionsWithoutVersions)
	if err != nil {
		return fmt.Errorf("failed to check functions without versions: %w", err)
	}
	if functionsWithoutVersions > 0 {
		logrus.WithField("count", functionsWithoutVersions).Warn("Found functions without versions")
		// This is a warning, not an error
	}

	return nil
}

// testMigrationRollback tests rollback capability for recent migrations
func testMigrationRollback(db *PostgresDB, migrationPath string, migrationsRun []string) error {
	if len(migrationsRun) == 0 {
		return nil // Nothing to test
	}

	// Only test rollback for the last migration to avoid complexity
	lastMigration := migrationsRun[len(migrationsRun)-1]

	// Create a temporary connection for rollback testing
	tempDB, err := createTempDatabaseForTesting(db)
	if err != nil {
		return fmt.Errorf("failed to create temp database for rollback testing: %w", err)
	}
	defer tempDB.Close()

	// Apply the migration to temp database
	tempMigrate, err := createMigrateInstance(tempDB, migrationPath)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance for temp DB: %w", err)
	}
	defer tempMigrate.Close()

	// Run up to the migration we want to test
	if err := tempMigrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations to temp DB: %w", err)
	}

	// Test rollback
	if err := tempMigrate.Steps(-1); err != nil {
		return fmt.Errorf("migration rollback test failed for %s: %w", lastMigration, err)
	}

	logrus.WithField("migration", lastMigration).Info("Migration rollback test passed")
	return nil
}

// createTempDatabaseForTesting creates a temporary database for testing (simplified)
func createTempDatabaseForTesting(originalDB *PostgresDB) (*sql.DB, error) {
	// In a real implementation, you'd create a temporary database
	// For now, we'll just return the original DB with a warning
	logrus.Warn("Using original database for rollback testing - implement temp DB creation for production")
	return originalDB.DB, nil
}

// shouldTestRollback determines if rollback testing should be performed
func shouldTestRollback() bool {
	return os.Getenv("DB_TEST_MIGRATION_ROLLBACK") == "true"
}

func parseGolangMigrateFilename(filename string) (version string, err error) {
	// Expected format: NNNNNN_description.up.sql
	parts := strings.Split(filename, "_")
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid filename format: %s", filename)
	}

	version = parts[0]
	return version, nil
}
