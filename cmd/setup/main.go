package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// defaultTenantID is a fixed UUID so setup is idempotent (same tenant every run).
var defaultTenantID = uuid.MustParse("a5eb0001-0000-4000-8000-000000000001")

const (
	adminEmail    = "admin@example.com"
	adminUsername = "functionfly"
)

func getAdminPassword() string {
	if pw, ok := os.LookupEnv("SETUP_ADMIN_PASSWORD"); ok && pw != "" {
		return pw
	}
	logrus.Fatal("SETUP_ADMIN_PASSWORD environment variable must be set - using weak defaults is not allowed")
	return ""
}

var jwtSecretFromEnv = os.Getenv("JWT_SECRET")
var setupJWTSecret   string

func getSetupJWTSecret() string {
	if setupJWTSecret != "" {
		return setupJWTSecret
	}
	setupJWTSecret = jwtSecretFromEnv
	if setupJWTSecret == "" {
		logrus.Fatal("JWT_SECRET environment variable must be set - using weak defaults is not allowed")
	}
	return setupJWTSecret
}

func main() {
	ctx := context.Background()

	// Initialize database
	db, err := storage.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := db.Repository()

	// Upsert default tenant (enterprise plan)
	fmt.Printf("Ensuring default tenant (ID: %s)\n", defaultTenantID)
	_, err = db.Exec(`
		INSERT INTO tenants (id, name, plan) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET plan = $3, name = $2`,
		defaultTenantID, "Default Tenant", plans.PlanEnterprise)
	if err != nil {
		log.Fatalf("Failed to create/update tenant: %v", err)
	}

	// Get or create admin user (idempotent)
	user, err := repo.GetUserByEmail(ctx, adminEmail)
	if err == nil {
		// User exists: ensure username and tenant plan
		_, err = repo.UpdateUser(ctx, user.ID, map[string]interface{}{"username": adminUsername})
		if err != nil {
			log.Fatalf("Failed to update admin username: %v", err)
		}
		if user.TenantID != defaultTenantID {
			_, _ = repo.UpdateTenant(ctx, user.TenantID, map[string]interface{}{"plan": plans.PlanEnterprise})
		}
		uname := adminUsername
		user.Username = &uname
		fmt.Printf("Updated existing admin user:\n")
	} else {
		// Create new admin user
		authSvc, err := auth.NewAuthService(repo, getSetupJWTSecret())
		if err != nil {
			log.Fatalf("Failed to create auth service: %v", err)
		}
		adminPassword := getAdminPassword()
		hash, err := authSvc.HashPassword(adminPassword)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}
		uname := adminUsername
		adminUser := &storage.User{
			ID:            uuid.New(),
			TenantID:      defaultTenantID,
			Username:      &uname,
			Email:         adminEmail,
			PasswordHash:  hash,
			Role:          "admin",
			EmailVerified: true,
		}
		user, err = repo.CreateUserWithRole(ctx, adminUser)
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
		fmt.Printf("Created default user:\n")
	}

	fmt.Printf("  Username: %s\n", adminUsername)
	fmt.Printf("  Email: %s\n", user.Email)
	fmt.Printf("  Tenant ID: %s\n", user.TenantID)
	fmt.Println("\nSetup complete! You can now login.")
}
