package vault

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "gorm.io/driver/postgres"
)

// PostgresDynamicManager implements DynamicSecretManager for PostgreSQL.
// It uses the standard database/sql + lib/pq driver.
type PostgresDynamicManager struct{}

// DBType implements DynamicSecretManager.
func (p *PostgresDynamicManager) DBType() DynamicSecretDBType {
	return DynamicSecretDBPostgres
}

// dial opens a short-lived admin connection to the target. The
// connection is closed when the returned function is called.
func dialTarget(ctx context.Context, t *DynamicSecretTarget, adminPassword string) (*sql.DB, error) {
	hostPort := net.JoinHostPort(t.Host, fmt.Sprintf("%d", t.Port))
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=10",
		t.Host, t.Port, t.AdminUsername, adminPassword, t.DatabaseName, t.SSLMode,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping %s: %w", hostPort, err)
	}
	return db, nil
}

// CreateUser creates a temporary Postgres role with a TTL, optionally
// granted membership of one or more existing roles.
func (p *PostgresDynamicManager) CreateUser(ctx context.Context, opts CreateUserOptions) (string, string, error) {
	adminPwd, err := decryptAdminPassword(ctx, opts.Target)
	if err != nil {
		return "", "", err
	}
	db, err := dialTarget(ctx, opts.Target, adminPwd)
	if err != nil {
		return "", "", err
	}
	defer db.Close()

	// CREATE USER is transactional so a GRANT failure rolls back the user.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", "", err
	}
	defer tx.Rollback()

	creUser := fmt.Sprintf(
		"CREATE USER %s WITH PASSWORD %s VALID UNTIL %s",
		quoteIdent(opts.Username),
		quoteLiteral(opts.Password),
		quoteLiteral(time.Now().Add(opts.TTL).Format("2006-01-02 15:04:05-07")),
	)
	if _, err := tx.ExecContext(ctx, creUser); err != nil {
		return "", "", fmt.Errorf("create user: %w", err)
	}

	roles := opts.AllowedRoles
	if opts.RoleTemplate != "" {
		roles = append([]string{opts.RoleTemplate}, roles...)
	}
	for _, role := range roles {
		if role == "" {
			continue
		}
		stmt := fmt.Sprintf("GRANT %s TO %s", quoteIdent(role), quoteIdent(opts.Username))
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return "", "", fmt.Errorf("grant role %s: %w", role, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", "", fmt.Errorf("commit: %w", err)
	}
	return opts.Username, opts.Password, nil
}

// RenewUser extends the VALID UNTIL expiry on an existing user.
func (p *PostgresDynamicManager) RenewUser(ctx context.Context, opts RenewUserOptions) error {
	adminPwd, err := decryptAdminPassword(ctx, opts.Target)
	if err != nil {
		return err
	}
	db, err := dialTarget(ctx, opts.Target, adminPwd)
	if err != nil {
		return err
	}
	defer db.Close()

	stmt := fmt.Sprintf(
		"ALTER USER %s VALID UNTIL %s",
		quoteIdent(opts.Username),
		quoteLiteral(time.Now().Add(opts.NewTTL).Format("2006-01-02 15:04:05-07")),
	)
	if _, err := db.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("alter user: %w", err)
	}
	return nil
}

// DropUser removes the temporary user. Idempotent.
func (p *PostgresDynamicManager) DropUser(ctx context.Context, opts DropUserOptions) error {
	adminPwd, err := decryptAdminPassword(ctx, opts.Target)
	if err != nil {
		return err
	}
	db, err := dialTarget(ctx, opts.Target, adminPwd)
	if err != nil {
		return err
	}
	defer db.Close()

	stmt := fmt.Sprintf("DROP USER IF EXISTS %s", quoteIdent(opts.Username))
	if _, err := db.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("drop user: %w", err)
	}
	return nil
}

// MySQLDynamicManager implements DynamicSecretManager for MySQL.
type MySQLDynamicManager struct{}

// DBType implements DynamicSecretManager.
func (m *MySQLDynamicManager) DBType() DynamicSecretDBType {
	return DynamicSecretDBMySQL
}

func dialMySQLTarget(ctx context.Context, t *DynamicSecretTarget, adminPassword string) (*sql.DB, error) {
	// MySQL uses a DSN form: user:pass@tcp(host:port)/db?...
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true&timeout=10s&readTimeout=10s&writeTimeout=10s",
		t.AdminUsername, adminPassword, t.Host, t.Port, t.DatabaseName,
	)
	if t.SSLMode != "" && t.SSLMode != "disable" {
		dsn += "&tls=true"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql target: %w", err)
	}
	return db, nil
}

// CreateUser creates a temporary MySQL user with the given TTL and
// grants the configured roles. MySQL VALID UNTIL requires DATETIME,
// so we pass an explicit format.
func (m *MySQLDynamicManager) CreateUser(ctx context.Context, opts CreateUserOptions) (string, string, error) {
	adminPwd, err := decryptAdminPassword(ctx, opts.Target)
	if err != nil {
		return "", "", err
	}
	db, err := dialMySQLTarget(ctx, opts.Target, adminPwd)
	if err != nil {
		return "", "", err
	}
	defer db.Close()

	expires := time.Now().Add(opts.TTL).UTC().Format("2006-01-02 15:04:05")
	creUser := fmt.Sprintf(
		"CREATE USER %s@'%%' IDENTIFIED BY %s ACCOUNT EXPIRES %s",
		quoteIdent(opts.Username),
		quoteLiteral(opts.Password),
		quoteLiteral(expires),
	)
	if _, err := db.ExecContext(ctx, creUser); err != nil {
		return "", "", fmt.Errorf("create user: %w", err)
	}

	roles := append([]string{}, opts.AllowedRoles...)
	if opts.RoleTemplate != "" {
		roles = append([]string{opts.RoleTemplate}, roles...)
	}
	for _, role := range roles {
		if role == "" {
			continue
		}
		// MySQL's GRANT ... TO syntax differs from Postgres. We grant
		// the configured privilege pattern. For production deployments
		// the recommended approach is to GRANT specific role names.
		stmt := fmt.Sprintf("GRANT %s TO %s@'%%'", quoteIdent(role), quoteIdent(opts.Username))
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return "", "", fmt.Errorf("grant role %s: %w", role, err)
		}
	}
	return opts.Username, opts.Password, nil
}

// RenewUser extends the ACCOUNT EXPIRES date for a MySQL user.
func (m *MySQLDynamicManager) RenewUser(ctx context.Context, opts RenewUserOptions) error {
	adminPwd, err := decryptAdminPassword(ctx, opts.Target)
	if err != nil {
		return err
	}
	db, err := dialMySQLTarget(ctx, opts.Target, adminPwd)
	if err != nil {
		return err
	}
	defer db.Close()

	expires := time.Now().Add(opts.NewTTL).UTC().Format("2006-01-02 15:04:05")
	stmt := fmt.Sprintf(
		"ALTER USER %s@'%%' ACCOUNT EXPIRES %s",
		quoteIdent(opts.Username),
		quoteLiteral(expires),
	)
	if _, err := db.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("alter user: %w", err)
	}
	return nil
}

// DropUser removes a temporary MySQL user. Idempotent.
func (m *MySQLDynamicManager) DropUser(ctx context.Context, opts DropUserOptions) error {
	adminPwd, err := decryptAdminPassword(ctx, opts.Target)
	if err != nil {
		return err
	}
	db, err := dialMySQLTarget(ctx, opts.Target, adminPwd)
	if err != nil {
		return err
	}
	defer db.Close()

	stmt := fmt.Sprintf("DROP USER IF EXISTS %s@'%%'", quoteIdent(opts.Username))
	if _, err := db.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("drop user: %w", err)
	}
	return nil
}
