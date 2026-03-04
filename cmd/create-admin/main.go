package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

const opTimeout = 20 * time.Second

func main() {
	var (
		email    = flag.String("email", "", "Admin user email (required)")
		password = flag.String("password", "", "Admin user password (required)")
		role     = flag.String("role", "super_admin", "Admin role (super_admin, admin, support, billing_admin, developer_admin)")
		tenantID = flag.String("tenant-id", "", "Tenant ID (optional, will create default tenant if not provided)")
	)
	flag.Parse()

	if *email == "" || *password == "" {
		fmt.Println("Usage: create-admin -email <email> -password <password> [-role <role>] [-tenant-id <uuid>]")
		fmt.Println("Example: create-admin -email admin@example.com -password mypassword -role super_admin")
		os.Exit(1)
	}

	validRoles := map[string]bool{
		"super_admin":     true,
		"admin":           true,
		"support":         true,
		"billing_admin":   true,
		"developer_admin": true,
	}
	if !validRoles[*role] {
		fmt.Printf("Invalid role: %s. Valid roles: super_admin, admin, support, billing_admin, developer_admin\n", *role)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	db, err := storage.OpenSimpleDB(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	var tenantUUID uuid.UUID
	if *tenantID != "" {
		tenantUUID, err = uuid.Parse(*tenantID)
		if err != nil {
			log.Fatalf("Invalid tenant ID: %v", err)
		}
		var n int
		err := db.QueryRowContext(ctx, `SELECT 1 FROM tenants WHERE id = $1`, tenantUUID).Scan(&n)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Fatalf("Tenant not found: %s", *tenantID)
			}
			log.Fatalf("Failed to verify tenant: %v", err)
		}
	} else {
		err := db.QueryRowContext(ctx, `SELECT id FROM tenants ORDER BY created_at DESC NULLS LAST LIMIT 1`).Scan(&tenantUUID)
		if err == sql.ErrNoRows {
			tenantUUID = uuid.New()
			fmt.Printf("Creating default tenant with ID: %s\n", tenantUUID)
			_, err = db.ExecContext(ctx, `
				INSERT INTO tenants (id, name, plan, status, created_at, updated_at)
				VALUES ($1, $2, $3, $4, NOW(), NOW())
				ON CONFLICT (id) DO NOTHING`,
				tenantUUID, "Default Tenant", "enterprise", "active")
			if err != nil {
				log.Fatalf("Failed to create tenant: %v", err)
			}
		} else if err != nil {
			log.Fatalf("Failed to list tenants: %v", err)
		} else {
			fmt.Printf("Using existing tenant: %s\n", tenantUUID)
		}
	}

	var n int
	err = db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE email = $1`, *email).Scan(&n)
	if err == nil {
		log.Fatalf("User with email %s already exists", *email)
	}
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("Failed to check existing user: %v", err)
	}

	authSvc := auth.NewAuthService(nil, "default-secret-key-change-in-production")
	hashedPassword, err := authSvc.HashPassword(*password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	userID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, tenant_id, email, password_hash, role, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())`,
		userID, tenantUUID, *email, hashedPassword, *role)
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("✅ Admin user created successfully!\n")
	fmt.Printf("  Email: %s\n", *email)
	fmt.Printf("  Role: %s\n", *role)
	fmt.Printf("  Tenant ID: %s\n", tenantUUID)
	fmt.Printf("  User ID: %s\n", userID)
	fmt.Printf("\n🔐 Login credentials:\n")
	fmt.Printf("  Email: %s\n", *email)
	fmt.Printf("  Password: %s\n", *password)
	fmt.Printf("\n🌐 Admin Panel Access:\n")
	fmt.Printf("  1. Login at: http://localhost:8080/login\n")
	fmt.Printf("  2. Navigate to: http://localhost:8080/admin\n")
	fmt.Printf("\n⚠️  Remember to change the default JWT secret in production!\n")
}
