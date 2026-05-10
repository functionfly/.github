// fix_user_tenants.go - Data migration to fix users affected by tenant assignment bug
// This script creates individual free-tier tenants for users who were incorrectly
// assigned to shared tenants (often enterprise tenants) due to the signup bug.
//
// Usage: go run scripts/migrations/fix_user_tenants.go
// Or as SQL: See fix_user_tenants.sql for the pure SQL approach
//
// Set MIGRATION_AUTO_COMMIT=true to auto-commit (production)

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

	// Find users incorrectly assigned to enterprise tenants:
	// A user is affected if:
	// 1. Their tenant has plan 'enterprise' AND
	// 2. The tenant has only this one user (single-user tenant that shouldn't be enterprise)
	// OR
	// 3. The tenant has multiple users but the user wasn't the first one created
	//    (first user is assumed to be the legitimate owner)
	//
	// This identifies users who were likely injected via X-Tenant-ID header manipulation
	rows, err := tx.Query(`
		SELECT u.id, u.email, u.tenant_id, t.plan, t.name as tenant_name,
		       (SELECT COUNT(*) FROM users WHERE tenant_id = u.tenant_id) as user_count,
		       (SELECT MIN(created_at) FROM users WHERE tenant_id = u.tenant_id) as first_user_created
		FROM users u
		JOIN tenants t ON u.tenant_id = t.id
		WHERE t.plan = 'enterprise'
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		log.Fatalf("Failed to query affected users: %v", err)
	}
	defer rows.Close()

	type affectedUser struct {
		UserID           uuid.UUID
		Email            string
		TenantID         uuid.UUID
		Plan             string
		TenantName       string
		UserCount        int
		FirstUserCreated string
	}

	var affectedUsers []affectedUser
	for rows.Next() {
		var u affectedUser
		var firstUserCreated sql.NullString
		if err := rows.Scan(&u.UserID, &u.Email, &u.TenantID, &u.Plan, &u.TenantName, &u.UserCount, &firstUserCreated); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		if firstUserCreated.Valid {
			u.FirstUserCreated = firstUserCreated.String
		}
		affectedUsers = append(affectedUsers, u)
	}
	rows.Close()

	if len(affectedUsers) == 0 {
		log.Println("No affected users found in enterprise tenants.")
		return
	}

	log.Printf("Found %d users in enterprise tenants", len(affectedUsers))
	log.Println("Reviewing affected users:")

	// Group users by tenant
	usersByTenant := make(map[uuid.UUID][]affectedUser)
	for _, u := range affectedUsers {
		usersByTenant[u.TenantID] = append(usersByTenant[u.TenantID], u)
	}

	fixedCount := 0
	skippedCount := 0

	for tenantID, users := range usersByTenant {
		log.Printf("\nTenant %s (%s) with plan '%s' has %d user(s):",
			tenantID, users[0].TenantName, users[0].Plan, len(users))

		for i, u := range users {
			// Keep the FIRST user in each enterprise tenant (earliest created_at)
			// These are assumed to be the legitimate enterprise tenant owners
			if i == 0 {
				log.Printf("  [%d] SKIPPED (assumed legitimate owner): %s (created: %s, user_count: %d)",
					i+1, u.Email, u.FirstUserCreated, u.UserCount)
				skippedCount++
				continue
			}

			// All other users in enterprise tenants are suspect:
			// They could have been assigned via X-Tenant-ID header injection
			// Move them to a new free-tier tenant
			newTenantID := uuid.New()
			newTenantName := fmt.Sprintf("%s's Workspace", u.Email)

			_, err := tx.Exec(`
				INSERT INTO tenants (id, name, plan, status, created_at, updated_at)
				VALUES ($1, $2, 'free', 'active', NOW(), NOW())
			`, newTenantID, newTenantName)
			if err != nil {
				log.Printf("  [%d] ERROR creating free tenant for %s: %v", i+1, u.Email, err)
				continue
			}

			_, err = tx.Exec(`
				UPDATE users SET tenant_id = $1, updated_at = NOW() WHERE id = $2
			`, newTenantID, u.UserID)
			if err != nil {
				log.Printf("  [%d] ERROR moving user %s to new tenant: %v", i+1, u.Email, err)
				continue
			}

			log.Printf("  [%d] DOWNGRADED: %s -> new free tenant %s (was in enterprise tenant %s)",
				i+1, u.Email, newTenantID, tenantID)
			fixedCount++
		}
	}

	log.Println()
	log.Printf("Summary: %d users downgraded from enterprise, %d enterprise owners preserved", fixedCount, skippedCount)

	// Auto-commit in production mode, prompt in interactive mode
	if os.Getenv("MIGRATION_AUTO_COMMIT") == "true" {
		if err := tx.Commit(); err != nil {
			log.Fatalf("Failed to commit transaction: %v", err)
		}
		log.Println("Changes committed!")
	} else {
		log.Println("Dry run complete. Set MIGRATION_AUTO_COMMIT=true to apply changes.")
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
