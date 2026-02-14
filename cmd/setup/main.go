package main

import (
	"fmt"
	"log"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

func main() {
	// Initialize database
	db, err := storage.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := db.Repository()

	// Create default tenant
	tenantID := uuid.New()
	fmt.Printf("Creating default tenant with ID: %s\n", tenantID)

	// Create default tenant in database (manually for now since we don't have a CreateTenant method)
	_, err = db.Exec(`
		INSERT INTO tenants (id, name) VALUES ($1, $2)
		ON CONFLICT (id) DO NOTHING`,
		tenantID, "Default Tenant")
	if err != nil {
		log.Fatalf("Failed to create tenant: %v", err)
	}

	// Create default user
	authSvc := auth.NewAuthService(repo, "default-secret-key-change-in-production")

	password := "admin123"
	hash, err := authSvc.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	user, err := repo.CreateUser("admin@example.com", hash, tenantID)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("Created default user:\n")
	fmt.Printf("  Email: %s\n", user.Email)
	fmt.Printf("  Password: %s\n", password)
	fmt.Printf("  Tenant ID: %s\n", tenantID)

	fmt.Println("\nSetup complete! You can now login with:")
	fmt.Printf("  curl -X POST http://localhost:8080/v1/auth/login -H 'Content-Type: application/json' -d '{\"email\":\"admin@example.com\",\"password\":\"admin123\"}'\n")
}