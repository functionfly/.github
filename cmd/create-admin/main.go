package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

const opTimeout = 20 * time.Second

// blockedExact are common whole-password defaults rejected in -production mode.
var blockedExact = map[string]struct{}{
	"admin123": {}, "password": {}, "password123": {}, "letmein": {},
	"12345678": {}, "qwerty": {}, "functionfly": {},
}

func validateProductionPassword(p string) error {
	if strings.Contains(p, " ") || strings.Contains(p, "\t") || strings.Contains(p, "\n") {
		return fmt.Errorf("password must not contain whitespace")
	}
	if len(p) < 16 {
		return fmt.Errorf("password must be at least 16 characters")
	}
	var upper, lower, digit, special bool
	for _, r := range p {
		switch {
		case r >= 'A' && r <= 'Z':
			upper = true
		case r >= 'a' && r <= 'z':
			lower = true
		case r >= '0' && r <= '9':
			digit = true
		default:
			if unicode.IsPunct(r) || unicode.IsSymbol(r) {
				special = true
			}
		}
	}
	if !upper || !lower || !digit || !special {
		return fmt.Errorf("password must include uppercase, lowercase, digit, and symbol (punctuation)")
	}
	if _, bad := blockedExact[strings.ToLower(p)]; bad {
		return fmt.Errorf("password matches a known weak default; choose a unique passphrase")
	}
	return nil
}

func isWeakDevPassword(p string) bool {
	return len(p) < 12 || strings.ToLower(p) == "admin123"
}

func main() {
	var (
		email        = flag.String("email", "", "Admin user email (required)")
		passwordFlag = flag.String("password", "", "Admin password, or set ADMIN_CREATE_PASSWORD (not echoed with -production)")
		role         = flag.String("role", "super_admin", "Admin role (super_admin, admin, support, billing_admin, developer_admin)")
		tenantID     = flag.String("tenant-id", "", "Tenant ID (optional, will create default tenant if not provided)")
		production   = flag.Bool("production", false, "Enforce strong password, do not print password to stdout (use for real deployments)")
	)
	flag.Parse()

	password := strings.TrimSpace(*passwordFlag)
	if password == "" {
		password = strings.TrimSpace(os.Getenv("ADMIN_CREATE_PASSWORD"))
	}

	if *email == "" || password == "" {
		fmt.Println("Usage: create-admin -email <email> -password <password> [-role <role>] [-tenant-id <uuid>] [-production]")
		fmt.Println("   Password may be set via ADMIN_CREATE_PASSWORD instead of -password (avoids shell history).")
		fmt.Println("   With -production: strong password required, password not printed.")
		fmt.Println("Example (dev):  create-admin -email admin@example.com -password '...' -role super_admin")
		fmt.Println("Example (prod): ADMIN_CREATE_PASSWORD='...' create-admin -email admin@corp.com -production -role super_admin")
		os.Exit(1)
	}

	if *production {
		if err := validateProductionPassword(password); err != nil {
			log.Fatalf("Production password rejected: %v", err)
		}
	} else if isWeakDevPassword(password) {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: weak password — use a long random password (or -production for enforced policy).\n")
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

	// For Neon pooled URLs, swap to the direct endpoint so pgx simple-protocol queries work.
	if raw := strings.TrimSpace(os.Getenv("DATABASE_URL")); strings.Contains(raw, "-pooler.") {
		_ = os.Setenv("DATABASE_URL", strings.Replace(raw, "-pooler.", ".", 1))
	}

	// Use pgx with QueryExecModeSimpleProtocol so parameterised queries work against
	// Neon / PgBouncer pooled connections (lib/pq's unnamed prepared statements do not).
	pg, err := storage.NewPostgresDBWithOptions(true)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pg.Close()
	db := pg.DB

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

	// Ensure tenant has enterprise plan for admin access
	_, err = db.ExecContext(ctx, `UPDATE tenants SET plan = 'enterprise', updated_at = NOW() WHERE id = $1`, tenantUUID)
	if err != nil {
		log.Fatalf("Failed to set tenant plan to enterprise: %v", err)
	}

	var n int
	err = db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE email = $1`, *email).Scan(&n)
	if err == nil {
		log.Fatalf("User with email %s already exists", *email)
	}
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("Failed to check existing user: %v", err)
	}

	authSvc, err := auth.NewAuthService(nil, "default-secret-key-change-in-production")
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}
	hashedPassword, err := authSvc.HashPassword(password)
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
	fmt.Printf("\n🔐 Login:\n")
	fmt.Printf("  Email: %s\n", *email)
	if *production {
		fmt.Printf("  Password: (not shown — use the value you provided; it was not logged)\n")
	} else {
		fmt.Printf("  Password: %s\n", password)
		fmt.Printf("\n⚠️  Do not use weak passwords in production. Prefer: go run ./cmd/create-admin ... -production\n")
	}
	fmt.Printf("\n🌐 Admin dashboard (standalone SPA): /auth/login on your admin host (e.g. https://admin.functionfly.com/auth/login)\n")
	fmt.Printf("   API must have CORS_ALLOW for that origin. See docs/ADMIN_SETUP_README.md\n")
	if *production {
		fmt.Printf("\n✓ Production mode: strong password enforced; password not echoed.\n")
	}
}
