package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/functionfly/functionfly/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate <command>")
		fmt.Println("Commands:")
		fmt.Println("  up     - Run all pending migrations")
		fmt.Println("  down   - Rollback the last migration")
		fmt.Println("  force  - Force a migration version (clears dirty state)")
		fmt.Println("  status - Show migration status")
		fmt.Println("  version - Show current migration version")
		os.Exit(1)
	}

	command := os.Args[1]

	// Initialize database connection (skip prepared statements for migrations)
	db, err := storage.NewPostgresDBWithOptions(true)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	switch command {
	case "up":
		fmt.Println("Running migrations...")
		if err := storage.RunMigrations(db); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		fmt.Println("Migrations completed successfully")

	case "down":
		if len(os.Args) < 3 {
			fmt.Println("Usage: migrate down <steps>")
			fmt.Println("  steps - Number of migrations to rollback (default: 1)")
			os.Exit(1)
		}

		steps := 1
		if len(os.Args) >= 3 {
			steps, err = strconv.Atoi(os.Args[2])
			if err != nil {
				log.Fatalf("Invalid steps value: %v", err)
			}
		}

		fmt.Printf("Rolling back %d migration(s)...\n", steps)
		for i := 0; i < steps; i++ {
			migrations, err := storage.GetAvailableMigrations()
			if err != nil {
				log.Fatalf("Failed to get available migrations: %v", err)
			}

			if len(migrations) == 0 {
				fmt.Println("No migrations to rollback")
				break
			}

			// Get the latest applied migration
			applied, err := storage.GetAppliedMigrations(db.DB)
			if err != nil {
				log.Fatalf("Failed to get applied migrations: %v", err)
			}

			if len(applied) == 0 {
				fmt.Println("No migrations have been applied")
				break
			}

			// Find the latest applied migration
			var latestMigration storage.Migration
			for _, migration := range migrations {
				if _, exists := applied[migration.Version]; exists {
					latestMigration = migration
					break
				}
			}

			if latestMigration.Version == "" {
				fmt.Println("No migrations to rollback")
				break
			}

			if err := storage.RollbackMigration(db.DB, latestMigration); err != nil {
				log.Fatalf("Failed to rollback migration %s: %v", latestMigration.Version, err)
			}
			fmt.Printf("Rolled back migration %s\n", latestMigration.Version)
		}
		fmt.Println("Rollback completed")

	case "force":
		if len(os.Args) < 3 {
			fmt.Println("Usage: migrate force <version>")
			fmt.Println("  version - Migration version to force as clean")
			os.Exit(1)
		}

		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Invalid version: %v", err)
		}

		// Get migration path (duplicating logic from storage package)
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get working directory: %v", err)
		}
		migrationPath := filepath.Join(wd, "migrations")
		if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
			migrationPath = filepath.Join(wd, "internal", "storage", "sql", "migrations")
			if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
				log.Fatalf("Migrations directory not found")
			}
		}
		migrationPath = fmt.Sprintf("file://%s", migrationPath)

		// Create migrate instance (duplicating logic from storage package)
		driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
		if err != nil {
			log.Fatalf("Failed to create postgres driver: %v", err)
		}
		m, err := migrate.NewWithDatabaseInstance(migrationPath, "postgres", driver)
		if err != nil {
			log.Fatalf("Failed to create migrate instance: %v", err)
		}
		defer m.Close()

		fmt.Printf("Forcing migration version %d to clean...\n", version)
		if err := m.Force(version); err != nil {
			log.Fatalf("Failed to force migration version: %v", err)
		}
		fmt.Printf("Migration version %d marked as clean\n", version)

	case "status":
		fmt.Println("Migration Status:")
		fmt.Println("=================")

		available, err := storage.GetAvailableMigrations()
		if err != nil {
			log.Fatalf("Failed to get available migrations: %v", err)
		}

		applied, err := storage.GetAppliedMigrations(db.DB)
		if err != nil {
			log.Fatalf("Failed to get applied migrations: %v", err)
		}

		fmt.Printf("Available migrations: %d\n", len(available))
		fmt.Printf("Applied migrations: %d\n", len(applied))

		if len(available) > 0 {
			fmt.Println("\nAvailable migrations:")
			for _, migration := range available {
				status := "pending"
				if _, exists := applied[migration.Version]; exists {
					status = "applied"
				}
				fmt.Printf("  %s - %s (%s)\n", migration.Version, migration.Filename, status)
			}
		}

	case "version":
		// Try golang-migrate version first
		m, err := storage.GetAppliedMigrations(db.DB)
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}

		if len(m) == 0 {
			fmt.Println("No migrations have been applied")
		} else {
			// Get the highest version
			var latestVersion string
			for version := range m {
				if version > latestVersion {
					latestVersion = version
				}
			}
			fmt.Printf("Current migration version: %s\n", latestVersion)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: up, down, status, version")
		os.Exit(1)
	}
}