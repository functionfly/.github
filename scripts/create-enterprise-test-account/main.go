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

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultEmail    = "enterprise-test@functionfly.local"
	defaultPassword = "enterprise123"
	defaultTenant   = "Enterprise Test"

	defaultDBHost = "localhost"
	defaultDBPort = 5432
	defaultDBUser = "postgres"
	defaultDBName = "functionfly"
)

func main() {
	ctx := context.Background()

	var (
		email = flag.String("email", defaultEmail, "User email")
		pass  = flag.String("password", defaultPassword, "User password")

		tenantID = flag.String("tenant-id", "", "Use existing tenant (plan will be set to enterprise)")

		dbURL  = flag.String("db-url", "", "Postgres connection string; if empty, DATABASE_URL is used")
		dbHost = flag.String("db-host", defaultDBHost, "Database host")
		dbPort = flag.Int("db-port", defaultDBPort, "Database port")
		dbUser = flag.String("db-user", defaultDBUser, "Database user")
		dbName = flag.String("db-name", defaultDBName, "Database name")
	)
	flag.Parse()

	if strings.TrimSpace(*email) == "" {
		log.Fatal("Missing --email")
	}
	if *pass == "" {
		log.Fatal("Missing --password")
	}

	// Configure DB connection for the shared storage layer.
	// For Neon pooled URLs, use the direct endpoint so parameterized queries
	// (and lib/pq's unnamed prepared statements) work; the pooler would otherwise
	// return "unnamed prepared statement does not exist".
	if raw := strings.TrimSpace(*dbURL); raw != "" {
		if strings.Contains(raw, "-pooler.") {
			raw = strings.Replace(raw, "-pooler.", ".", 1)
		}
		_ = os.Setenv("DATABASE_URL", raw)
	}
	if os.Getenv("DATABASE_URL") == "" {
		_ = os.Setenv("DB_HOST", *dbHost)
		_ = os.Setenv("DB_PORT", fmt.Sprintf("%d", *dbPort))
		_ = os.Setenv("DB_USER", *dbUser)
		_ = os.Setenv("DB_NAME", *dbName)
		// DB_PASSWORD/PGPASSWORD are expected to already be set by the caller (.env / script wrapper).
	}

	pg, err := storage.NewPostgresDBWithOptions(true /* skipPreparedStatements */)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pg.Close()
	db := pg.DB

	// Password hashing: bcrypt cost 10 (bcrypt.DefaultCost).
	hash, err := bcrypt.GenerateFromPassword([]byte(*pass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	userExists, err := countUsersByEmail(ctx, db, *email)
	if err != nil {
		log.Fatalf("Failed to check user existence: %v", err)
	}
	if userExists > 0 {
		log.Fatalf("User already exists: %s", *email)
	}

	tenantUUID := strings.TrimSpace(*tenantID)
	if tenantUUID != "" {
		// Validate tenant exists, then update plan.
		ok, err := tenantExists(ctx, db, tenantUUID)
		if err != nil {
			log.Fatalf("Failed to check tenant existence: %v", err)
		}
		if !ok {
			log.Fatalf("Tenant not found: %s", tenantUUID)
		}

		if _, err := db.ExecContext(ctx, `
			UPDATE tenants SET plan = 'enterprise', updated_at = NOW()
			WHERE id = $1
		`, tenantUUID); err != nil {
			log.Fatalf("Failed to set tenant plan: %v", err)
		}
	} else {
		tenantUUID = uuid.NewString()

		if _, err := db.ExecContext(ctx, `
			INSERT INTO tenants (id, name, plan, status, created_at, updated_at)
			VALUES ($1, $2, 'enterprise', 'active', NOW(), NOW())
			ON CONFLICT (id) DO UPDATE
			SET plan = 'enterprise', updated_at = NOW();
		`, tenantUUID, defaultTenant); err != nil {
			log.Fatalf("Failed to create/update tenant: %v", err)
		}
	}

	username := usernameFromEmail(*email)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO users (id, tenant_id, email, username, password_hash, role, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'owner', true, NOW(), NOW());
	`, uuid.NewString(), tenantUUID, *email, username, string(hash)); err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	// Print the created user row.
	var (
		createdID   string
		outEmail    string
		outUsername string
		outRole     string
		outTenantID string
		createdAt   time.Time
	)

	err = db.QueryRowContext(ctx, `
		SELECT id, email, username, role, tenant_id, created_at
		FROM users
		WHERE email = $1;
	`, *email).Scan(&createdID, &outEmail, &outUsername, &outRole, &outTenantID, &createdAt)
	if err != nil {
		log.Fatalf("Failed to fetch created user: %v", err)
	}

	fmt.Println("Enterprise test account created.")
	fmt.Println()
	fmt.Println("User:")
	fmt.Printf("  id:          %s\n", createdID)
	fmt.Printf("  email:       %s\n", outEmail)
	fmt.Printf("  username:    %s\n", outUsername)
	fmt.Printf("  role:        %s\n", outRole)
	fmt.Printf("  tenant_id:   %s\n", outTenantID)
	fmt.Printf("  created_at:  %s\n", createdAt.Format(time.RFC3339))

	fmt.Println()
	fmt.Println("Login:")
	fmt.Printf("  Email:    %s\n", *email)
	fmt.Printf("  Password: %s\n", *pass)
	fmt.Println()
	fmt.Println("Enterprise features to test:")
	fmt.Println("  SLA:      /enterprise/sla")
	fmt.Println("  Audit:    /enterprise/audit")
	fmt.Println("  Support:  /enterprise/support")
}

func usernameFromEmail(email string) string {
	if at := strings.IndexByte(email, '@'); at > 0 {
		return email[:at]
	}
	return email
}

func countUsersByEmail(ctx context.Context, db *sql.DB, email string) (int, error) {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE email = $1;`, email).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func tenantExists(ctx context.Context, db *sql.DB, tenantID string) (bool, error) {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tenants WHERE id = $1;`, tenantID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}
