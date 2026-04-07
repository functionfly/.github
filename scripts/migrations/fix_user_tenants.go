// fix_user_tenants.go - Data migration to fix users affected by tenant assignment bug
// This script creates individual starter tenants for users who were incorrectly
// assigned to shared tenants (often enterprise tenants) due to the signup bug.
//
// Usage: go run scripts/migrations/fix_user_tenants.go
// Or as SQL: See fix_user_tenants.sql for the pure SQL approach

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Try individual env vars
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "postgres")
		pass := getEnv("DB_PASSWORD", "postgres")
		dbname := getEnv("DB_NAME", "functionfly")
		sslmode := getEnv("DB_SSLMODE", "disable")
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, pass, host, port, dbname, sslmode)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database")

	// Start transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Find users who need fixing:
	// Users in tenants that have multiple users OR users in enterprise tenants
	// who were likely incorrectly assigned due to the bug
	rows, err := tx.Query(`
		SELECT u.id, u.email, u.tenant_id, t.plan, t.name as tenant_name,
		       (SELECT COUNT(*) FROM users WHERE tenant_id = u.tenant_id) as user_count
		FROM users u
		JOIN tenants t ON u.tenant_id = t.id
		WHERE t.plan = 'enterprise'
		   OR (SELECT COUNT(*) FROM users WHERE tenant_id = u.tenant_id) > 1
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		log.Fatalf("Failed to query affected users: %v", err)
	}
	defer rows.Close()

	type affectedUser struct {
		UserID      uuid.UUID
		Email       string
		TenantID    uuid.UUID
		Plan        string
		TenantName  string
		UserCount   int
	}

	var affectedUsers []affectedUser
	for rows.Next() {
		var u affectedUser
		if err := rows.Scan(&u.UserID, &u.Email, &u.TenantID, &u.Plan, &u.TenantName, &u.UserCount); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		affectedUsers = append(affectedUsers, u)
	}
	rows.Close()

	if len(affectedUsers) == 0 {
		log.Println("No affected users found. All users appear to have correct tenant assignments.")
		return
	}

	log.Printf("Found %d users that may need tenant reassignment", len(affectedUsers))
	log.Println("Reviewing affected users:")

	// Skip the first tenant (which is legitimate) and fix the rest
	// We assume the first user in each shared tenant is the legitimate owner
	usersToFix := make(map[uuid.UUID][]affectedUser) // tenant_id -> users
	for _, u := range affectedUsers {
		usersToFix[u.TenantID] = append(usersToFix[u.TenantID], u)
	}

	fixedCount := 0
	skippedCount := 0

	for _, users := range usersToFix {
		// Keep the first user in each tenant (assumed to be the legitimate owner)
		// Fix the rest by creating new tenants for them
		for i, u := range users {
			if i == 0 {
				// Skip the first user in this tenant - they're likely the legitimate tenant owner
				log.Printf("  SKIPPED (assumed owner): %s (tenant: %s, plan: %s, users in tenant: %d)",
					u.Email, u.TenantID, u.Plan, u.UserCount)
				skippedCount++
				continue
			}

			// Create new tenant for this user
			newTenantID := uuid.New()
			newTenantName := fmt.Sprintf("%s's Workspace", u.Email)

			_, err := tx.Exec(`
				INSERT INTO tenants (id, name, plan, status, created_at, updated_at)
				VALUES ($1, $2, 'starter', 'active', NOW(), NOW())
			`, newTenantID, newTenantName)
			if err != nil {
				log.Printf("  ERROR creating tenant for %s: %v", u.Email, err)
				continue
			}

			// Move user to new tenant
			_, err = tx.Exec(`
				UPDATE users SET tenant_id = $1, updated_at = NOW() WHERE id = $2
			`, newTenantID, u.UserID)
			if err != nil {
				log.Printf("  ERROR moving user %s to new tenant: %v", u.Email, err)
				continue
			}

			log.Printf("  FIXED: %s -> new tenant %s (was in %s/%s)",
				u.Email, newTenantID, u.TenantID, u.Plan)
			fixedCount++
		}
	}

	log.Println()
	log.Printf("Summary: %d users fixed, %d users skipped (assumed owners)", fixedCount, skippedCount)
	log.Println("Review the changes above. Commit? (yes/no)")

	// Auto-commit in production mode, prompt in interactive mode
	if os.Getenv("MIGRATION_AUTO_COMMIT") == "true" {
		if err := tx.Commit(); err != nil {
			log.Fatalf("Failed to commit transaction: %v", err)
		}
		log.Println("Changes committed!")
	} else {
		log.Println("Dry run complete. Set MIGRATION_AUTO_COMMIT=true to apply changes.")
		log.Println("Or run with --commit flag")
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
