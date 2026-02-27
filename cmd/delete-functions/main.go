package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Get database connection string from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Confirm the operation
	if len(os.Args) > 1 && os.Args[1] == "--yes" {
		// Skip confirmation
	} else {
		fmt.Println("WARNING: This will permanently delete ALL functions from the database!")
		fmt.Println("This includes:")
		fmt.Println("- User-created functions")
		fmt.Println("- Function deployments")
		fmt.Println("- Function logs")
		fmt.Println("- Registry functions and versions")
		fmt.Println("- Function executions and ratings")
		fmt.Println("")
		fmt.Print("Are you sure you want to continue? Type 'yes' to confirm: ")

		var response string
		fmt.Scanln(&response)
		if response != "yes" {
			fmt.Println("Operation cancelled.")
			os.Exit(0)
		}
	}

	fmt.Println("Connecting to database...")

	// Initialize database connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	ctx := context.Background()

	fmt.Println("Deleting all functions...")

	// Delete in order to respect foreign key constraints
	tables := []string{
		"registry_function_approval_comments",
		"registry_function_approvals",
		"registry_function_malware_scans",
		"registry_function_signatures",
		"registry_function_ratings",
		"registry_executions_public",
		"registry_function_executions",
		"registry_function_versions",
		"registry_functions",
		"function_logs",
		"function_deployments",
		"functions",
	}

	for _, table := range tables {
		fmt.Printf("Deleting from %s...\n", table)
		query := fmt.Sprintf("DELETE FROM %s", table)
		result, err := db.ExecContext(ctx, query)
		if err != nil {
			log.Fatalf("Failed to delete from %s: %v", table, err)
		}
		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("  Deleted %d rows from %s\n", rowsAffected, table)
	}

	// Reset sequences
	fmt.Println("Resetting sequences...")
	sequences := []string{
		"functions_id_seq",
		"function_deployments_id_seq",
		"function_logs_id_seq",
		"registry_functions_id_seq",
		"registry_function_versions_id_seq",
		"registry_function_executions_id_seq",
		"registry_executions_public_id_seq",
		"registry_function_ratings_id_seq",
		"registry_function_signatures_id_seq",
		"registry_function_malware_scans_id_seq",
		"registry_function_approvals_id_seq",
		"registry_function_approval_comments_id_seq",
	}

	for _, seq := range sequences {
		query := fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", seq)
		_, err := db.ExecContext(ctx, query)
		if err != nil {
			// Some sequences might not exist, so we'll just warn instead of failing
			fmt.Printf("Warning: Could not reset sequence %s: %v\n", seq, err)
		}
	}

	fmt.Println("All functions have been successfully deleted from the database!")
	fmt.Println("Note: This operation cannot be undone. Make sure to backup your data before running this script.")
}