// cmd/apply-migrations/main.go - Standalone migration tool
// Usage: go run cmd/apply-migrations/main.go
// Or:    ./bin/apply-migrations
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		envFile   = flag.String("env", ".env", "Path to .env file")
		dryRun    = flag.Bool("dry-run", false, "Show what would be applied without running")
		status    = flag.Bool("status", false, "Show migration status only")
		target    = flag.String("target", "all", "Target: local, neon, or all")
		skipVal   = flag.Bool("skip-validation", false, "Skip post-migration validation")
	)
	flag.Parse()

	// Load .env file
	if _, err := os.Stat(*envFile); err == nil {
		if err := godotenv.Load(*envFile); err != nil {
			logrus.WithError(err).Warn("Failed to load .env file")
		}
	}

	// Configure logging
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	// Handle skip validation
	if *skipVal {
		os.Setenv("SKIP_MIGRATION_VALIDATION", "true")
	}

	// Handle different targets
	switch *target {
	case "local":
		// Clear DATABASE_URL to force local connection
		os.Unsetenv("DATABASE_URL")
		applyMigrations(*dryRun, *status)
	case "neon":
		// Check for DATABASE_URL
		if os.Getenv("DATABASE_URL") == "" {
			log.Fatal("DATABASE_URL not set. Cannot connect to Neon.")
		}
		// Clear individual DB vars to force DATABASE_URL usage
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		applyMigrations(*dryRun, *status)
	case "all":
		// Apply to local first
		if hasLocalConfig() {
			os.Unsetenv("DATABASE_URL")
			logrus.Info("========================================")
			logrus.Info("Applying migrations to LOCAL PostgreSQL")
			logrus.Info("========================================")
			applyMigrations(*dryRun, *status)
		} else {
			logrus.Warn("Local PostgreSQL config not found, skipping")
		}

		// Apply to Neon
		if os.Getenv("DATABASE_URL") != "" {
			os.Unsetenv("DB_HOST")
			os.Unsetenv("DB_PORT")
			os.Unsetenv("DB_USER")
			os.Unsetenv("DB_PASSWORD")
			os.Unsetenv("DB_NAME")
			logrus.Info("")
			logrus.Info("========================================")
			logrus.Info("Applying migrations to NEON PostgreSQL")
			logrus.Info("========================================")
			applyMigrations(*dryRun, *status)
		} else {
			logrus.Warn("DATABASE_URL not set, skipping Neon")
		}
	default:
		log.Fatalf("Unknown target: %s. Use 'local', 'neon', or 'all'", *target)
	}
}

func hasLocalConfig() bool {
	return os.Getenv("DB_HOST") != "" ||
		os.Getenv("DB_PORT") != "" ||
		os.Getenv("DB_USER") != "" ||
		os.Getenv("DB_NAME") != ""
}

func applyMigrations(dryRun, statusOnly bool) {
	// Initialize database connection
	db, err := storage.NewPostgresDB()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Check current version
	if statusOnly {
		showStatus(db)
		return
	}

	// Show dry-run info
	if dryRun {
		showDryRun(db)
		return
	}

	// Run migrations
	logrus.Info("Running database migrations...")
	result, err := storage.RunMigrationsWithValidation(db)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to run migrations")
	}

	// Report results
	if result.Success {
		logrus.WithFields(logrus.Fields{
			"migrations_run": len(result.MigrationsRun),
			"duration":       result.Duration,
		}).Info("Migrations completed successfully")
	} else {
		logrus.WithFields(logrus.Fields{
			"migrations_run":    len(result.MigrationsRun),
			"validation_errors": len(result.ValidationErrors),
			"duration":          result.Duration,
		}).Warn("Migrations completed with validation errors")

		for _, err := range result.ValidationErrors {
			logrus.Warn("Validation: " + err)
		}
	}

	// Show applied migrations
	if len(result.MigrationsRun) > 0 {
		logrus.Info("Applied migrations:")
		for _, ver := range result.MigrationsRun {
			logrus.Info("  - " + ver)
		}
	} else {
		logrus.Info("No new migrations to apply (already up to date)")
	}
}

func showStatus(db *storage.PostgresDB) {
	// Get available migrations
	available, err := storage.GetAvailableMigrations()
	if err != nil {
		logrus.WithError(err).Error("Failed to get available migrations")
		return
	}

	logrus.Infof("Available migrations: %d", len(available))

	// Get applied migrations
	applied, err := storage.GetAppliedMigrations(db.DB)
	if err != nil {
		logrus.WithError(err).Error("Failed to get applied migrations")
		return
	}

	logrus.Infof("Applied migrations: %d", len(applied))

	// Show pending
	pending := 0
	for _, m := range available {
		if _, ok := applied[m.Version]; !ok {
			pending++
			logrus.Infof("Pending: %s (%s)", m.Version, m.Filename)
		}
	}

	if pending == 0 {
		logrus.Info("All migrations are up to date!")
	} else {
		logrus.Infof("Total pending: %d", pending)
	}
}

func showDryRun(db *storage.PostgresDB) {
	// Get available migrations
	available, err := storage.GetAvailableMigrations()
	if err != nil {
		logrus.WithError(err).Error("Failed to get available migrations")
		return
	}

	// Get applied migrations
	applied, err := storage.GetAppliedMigrations(db.DB)
	if err != nil {
		logrus.WithError(err).Error("Failed to get applied migrations")
		return
	}

	// Show what would be applied
	logrus.Info("DRY RUN: The following migrations would be applied:")
	count := 0
	for _, m := range available {
		if _, ok := applied[m.Version]; !ok {
			count++
			fmt.Printf("  [%s] %s\n", m.Version, m.Filename)
			// Show first 200 chars of SQL
			sql := m.Content
			if len(sql) > 200 {
				sql = sql[:200] + "..."
			}
			fmt.Printf("       %s\n", sql)
		}
	}

	if count == 0 {
		logrus.Info("No migrations to apply - database is up to date")
	} else {
		logrus.Infof("Total to apply: %d", count)
	}
}
