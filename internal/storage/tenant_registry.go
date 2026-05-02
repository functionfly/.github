package storage

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/scrypt"
)

// TenantDBRegistry manages tenant database configurations and credentials
type TenantDBRegistry struct {
	db             *PostgresDB
	encryptionKeyID string
	// Local cache of registry entries
	cache      sync.Map // map[uuid.UUID]*TenantDBEntry
	cacheTTL   time.Duration
	cacheMutex sync.RWMutex
	stopCache  chan struct{}
}

// TenantDBEntry represents a tenant's database configuration
type TenantDBEntry struct {
	ID                   uuid.UUID
	TenantID             uuid.UUID
	DBName               string
	DBHost               string
	DBPort               int
	DBUser               string
	DBPasswordEncrypted   string
	ConnectionStringTemplate string
	MaxConnections       int
	MinConnections       int
	Status               string
	EncryptionVersion    int
	CreatedAt            time.Time
	UpdatedAt            time.Time

	// Decrypted password (transient, never stored)
	decryptedPassword string
}

// NewTenantDBRegistry creates a new tenant database registry
func NewTenantDBRegistry(db *PostgresDB, encryptionKeyID string) *TenantDBRegistry {
	registry := &TenantDBRegistry{
		db:              db,
		encryptionKeyID: encryptionKeyID,
		cacheTTL:        5 * time.Minute,
		stopCache:       make(chan struct{}),
	}

	// Start cache cleanup
	go registry.cleanupCacheLoop()

	return registry
}

// Register stores a new tenant database configuration
func (r *TenantDBRegistry) Register(ctx context.Context, entry *TenantDBEntry) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	entry.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tenant_database_configs
		 (id, tenant_id, db_name, db_host, db_port, db_user, db_password_encrypted,
		  connection_string_template, max_connections, min_connections, status, encryption_version)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 ON CONFLICT (tenant_id) DO UPDATE SET
		 	db_name = EXCLUDED.db_name,
		 	db_host = EXCLUDED.db_host,
		 	db_port = EXCLUDED.db_port,
		 	db_user = EXCLUDED.db_user,
		 	db_password_encrypted = EXCLUDED.db_password_encrypted,
		 	connection_string_template = EXCLUDED.connection_string_template,
		 	max_connections = EXCLUDED.max_connections,
		 	min_connections = EXCLUDED.min_connections,
		 	status = EXCLUDED.status,
		 	encryption_version = EXCLUDED.encryption_version,
		 	updated_at = NOW()`,
		entry.ID, entry.TenantID, entry.DBName, entry.DBHost, entry.DBPort, entry.DBUser,
		entry.DBPasswordEncrypted, entry.ConnectionStringTemplate, entry.MaxConnections,
		entry.MinConnections, entry.Status, entry.EncryptionVersion)

	if err != nil {
		return fmt.Errorf("failed to register tenant DB: %w", err)
	}

	// Update cache
	r.cache.Store(entry.TenantID, entry)

	logrus.Infof("Registered tenant database %s for tenant %s", entry.DBName, entry.TenantID)
	return nil
}

// GetByTenantID retrieves a tenant's database configuration
func (r *TenantDBRegistry) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*TenantDBEntry, error) {
	// Check cache first
	if entry, ok := r.cache.Load(tenantID); ok {
		cached := entry.(*TenantDBEntry)
		// Check if cache is still valid
		if time.Since(cached.UpdatedAt) < r.cacheTTL {
			return cached, nil
		}
	}

	// Fetch from database
	entry := &TenantDBEntry{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, db_name, db_host, db_port, db_user, db_password_encrypted,
		        connection_string_template, max_connections, min_connections, status,
		        COALESCE(encryption_version, 1), created_at, updated_at
		 FROM tenant_database_configs
		 WHERE tenant_id = $1`,
		tenantID).Scan(
		&entry.ID, &entry.TenantID, &entry.DBName, &entry.DBHost, &entry.DBPort,
		&entry.DBUser, &entry.DBPasswordEncrypted, &entry.ConnectionStringTemplate,
		&entry.MaxConnections, &entry.MinConnections, &entry.Status,
		&entry.EncryptionVersion, &entry.CreatedAt, &entry.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant database config not found for %s", tenantID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant DB config: %w", err)
	}

	// Update cache
	r.cache.Store(tenantID, entry)

	return entry, nil
}

// GetByDBName retrieves a tenant's database configuration by database name
func (r *TenantDBRegistry) GetByDBName(ctx context.Context, dbName string) (*TenantDBEntry, error) {
	entry := &TenantDBEntry{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, db_name, db_host, db_port, db_user, db_password_encrypted,
		        connection_string_template, max_connections, min_connections, status,
		        COALESCE(encryption_version, 1), created_at, updated_at
		 FROM tenant_database_configs
		 WHERE db_name = $1`,
		dbName).Scan(
		&entry.ID, &entry.TenantID, &entry.DBName, &entry.DBHost, &entry.DBPort,
		&entry.DBUser, &entry.DBPasswordEncrypted, &entry.ConnectionStringTemplate,
		&entry.MaxConnections, &entry.MinConnections, &entry.Status,
		&entry.EncryptionVersion, &entry.CreatedAt, &entry.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant database config not found for db %s", dbName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant DB config: %w", err)
	}

	return entry, nil
}

// UpdateStatus updates the status of a tenant's database
func (r *TenantDBRegistry) UpdateStatus(ctx context.Context, tenantID uuid.UUID, status string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE tenant_database_configs SET status = $1, updated_at = NOW() WHERE tenant_id = $2`,
		status, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update tenant DB status: %w", err)
	}

	// Invalidate cache
	r.cache.Delete(tenantID)

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("tenant database config not found")
	}

	logrus.Infof("Updated tenant database status for %s to %s", tenantID, status)
	return nil
}

// Delete removes a tenant's database configuration
func (r *TenantDBRegistry) Delete(ctx context.Context, tenantID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM tenant_database_configs WHERE tenant_id = $1",
		tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant DB config: %w", err)
	}

	// Remove from cache
	r.cache.Delete(tenantID)

	logrus.Infof("Deleted tenant database config for %s", tenantID)
	return nil
}

// List returns all tenant database configurations
func (r *TenantDBRegistry) List(ctx context.Context) ([]*TenantDBEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, db_name, db_host, db_port, db_user, db_password_encrypted,
		        connection_string_template, max_connections, min_connections, status,
		        COALESCE(encryption_version, 1), created_at, updated_at
		 FROM tenant_database_configs
		 ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant DB configs: %w", err)
	}
	defer rows.Close()

	var entries []*TenantDBEntry
	for rows.Next() {
		entry := &TenantDBEntry{}
		err := rows.Scan(
			&entry.ID, &entry.TenantID, &entry.DBName, &entry.DBHost, &entry.DBPort,
			&entry.DBUser, &entry.DBPasswordEncrypted, &entry.ConnectionStringTemplate,
			&entry.MaxConnections, &entry.MinConnections, &entry.Status,
			&entry.EncryptionVersion, &entry.CreatedAt, &entry.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant DB entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// ListByStatus returns all tenant database configurations with a specific status
func (r *TenantDBRegistry) ListByStatus(ctx context.Context, status string) ([]*TenantDBEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, db_name, db_host, db_port, db_user, db_password_encrypted,
		        connection_string_template, max_connections, min_connections, status,
		        COALESCE(encryption_version, 1), created_at, updated_at
		 FROM tenant_database_configs
		 WHERE status = $1
		 ORDER BY created_at DESC`,
		status)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant DB configs by status: %w", err)
	}
	defer rows.Close()

	var entries []*TenantDBEntry
	for rows.Next() {
		entry := &TenantDBEntry{}
		err := rows.Scan(
			&entry.ID, &entry.TenantID, &entry.DBName, &entry.DBHost, &entry.DBPort,
			&entry.DBUser, &entry.DBPasswordEncrypted, &entry.ConnectionStringTemplate,
			&entry.MaxConnections, &entry.MinConnections, &entry.Status,
			&entry.EncryptionVersion, &entry.CreatedAt, &entry.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant DB entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// DecryptPassword decrypts a tenant's database password using AES-256-GCM
func (r *TenantDBRegistry) DecryptPassword(encryptedPassword string) (string, error) {
	if encryptedPassword == "" {
		return "", nil
	}

	// Check for encrypted prefix from encryption manager
	if strings.HasPrefix(encryptedPassword, "ENC:") {
		// This is encrypted with the platform's encryption manager
		// We need access to the encryption manager to decrypt
		// For now, delegate to decryptPasswordFallback
		return decryptPasswordFallback(encryptedPassword)
	}

	// Check for "encrypted:" prefix (legacy fallback)
	if len(encryptedPassword) > 11 && encryptedPassword[:11] == "encrypted:" {
		return encryptedPassword[11:], nil
	}

	// No prefix found, assume it's a fallback encrypted value
	return decryptPasswordFallback(encryptedPassword)
}

// EncryptPassword encrypts a tenant's database password for storage using AES-256-GCM
func (r *TenantDBRegistry) EncryptPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}

	// Use scrypt-based encryption with a service key (same as fallback)
	salt := []byte("functionfly-tenant-db-key-v1")
	serviceKey := []byte("functionfly-dedicated-db-encryption-key-v1") // Same as fallback

	// Derive key using scrypt
	key, err := scrypt.Key(serviceKey, salt, 32768, 8, 1, 32)
	if err != nil {
		return "", fmt.Errorf("failed to derive key: %w", err)
	}

	// Encrypt with AES-256-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptPasswordFallback decrypts a password encrypted with the fallback method
func decryptPasswordFallback(encryptedPassword string) (string, error) {
	// Remove ENC: prefix if present
	encryptedPassword = strings.TrimPrefix(encryptedPassword, "ENC:")

	data, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		// Not base64 encoded - might be legacy "encrypted:" prefix
		return "", fmt.Errorf("invalid encrypted password format: %w", err)
	}

	// Derive the same key used for encryption (using service key, not password)
	salt := []byte("functionfly-tenant-db-key-v1")
	serviceKey := []byte("functionfly-dedicated-db-encryption-key-v1") // Same as encrypt

	key, err := scrypt.Key(serviceKey, salt, 32768, 8, 1, 32)
	if err != nil {
		return "", fmt.Errorf("failed to derive key: %w", err)
	}

	// Use AES-256-GCM to decrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// BuildConnectionString constructs a PostgreSQL connection string for a tenant
func (r *TenantDBRegistry) BuildConnectionString(ctx context.Context, tenantID uuid.UUID) (string, error) {
	entry, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return "", err
	}

	password, err := r.DecryptPassword(entry.DBPasswordEncrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt password: %w", err)
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require pool_max_conns=%d",
		entry.DBHost, entry.DBPort, entry.DBUser, password, entry.DBName, entry.MaxConnections), nil
}

// GetConnectionPoolConfig returns pgxpool config for a tenant's database
func (r *TenantDBRegistry) GetConnectionPoolConfig(ctx context.Context, tenantID uuid.UUID) (*TenantDBPoolConfig, error) {
	entry, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	password, err := r.DecryptPassword(entry.DBPasswordEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}

	return &TenantDBPoolConfig{
		Host:       entry.DBHost,
		Port:       entry.DBPort,
		User:       entry.DBUser,
		Password:   password,
		Database:   entry.DBName,
		MinConns:   entry.MinConnections,
		MaxConns:   entry.MaxConnections,
	}, nil
}

// TenantDBPoolConfig holds connection pool configuration for a tenant DB
type TenantDBPoolConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	MinConns int
	MaxConns int
}

// cleanupCacheLoop periodically cleans up stale cache entries
func (r *TenantDBRegistry) cleanupCacheLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.cleanupCache()
		case <-r.stopCache:
			return
		}
	}
}

// cleanupCache removes stale entries from the cache
func (r *TenantDBRegistry) cleanupCache() {
	now := time.Now()
	r.cache.Range(func(key, value interface{}) bool {
		entry := value.(*TenantDBEntry)
		if now.Sub(entry.UpdatedAt) > r.cacheTTL*2 {
			r.cache.Delete(key)
		}
		return true
	})
}

// Close stops the cache cleanup goroutine
func (r *TenantDBRegistry) Close() {
	close(r.stopCache)
}

// Export exports tenant database configs as JSON (for backup/compliance)
func (r *TenantDBRegistry) Export(ctx context.Context) ([]byte, error) {
	entries, err := r.List(ctx)
	if err != nil {
		return nil, err
	}

	// Create export-safe version (without decrypted passwords)
	type ExportEntry struct {
		ID         uuid.UUID `json:"id"`
		TenantID   uuid.UUID `json:"tenant_id"`
		DBName     string    `json:"db_name"`
		DBHost     string    `json:"db_host"`
		DBPort     int       `json:"db_port"`
		DBUser     string    `json:"db_user"`
		Status     string    `json:"status"`
		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
	}

	exportEntries := make([]ExportEntry, len(entries))
	for i, e := range entries {
		exportEntries[i] = ExportEntry{
			ID:        e.ID,
			TenantID:  e.TenantID,
			DBName:    e.DBName,
			DBHost:    e.DBHost,
			DBPort:    e.DBPort,
			DBUser:    e.DBUser,
			Status:    e.Status,
			CreatedAt: e.CreatedAt,
			UpdatedAt: e.UpdatedAt,
		}
	}

	return json.MarshalIndent(exportEntries, "", "  ")
}

// TenantDBRegistryInterface defines the interface for tenant registry operations
type TenantDBRegistryInterface interface {
	Register(ctx context.Context, entry *TenantDBEntry) error
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*TenantDBEntry, error)
	GetByDBName(ctx context.Context, dbName string) (*TenantDBEntry, error)
	UpdateStatus(ctx context.Context, tenantID uuid.UUID, status string) error
	Delete(ctx context.Context, tenantID uuid.UUID) error
	List(ctx context.Context) ([]*TenantDBEntry, error)
	ListByStatus(ctx context.Context, status string) ([]*TenantDBEntry, error)
	BuildConnectionString(ctx context.Context, tenantID uuid.UUID) (string, error)
	Export(ctx context.Context) ([]byte, error)
}

// Verify TenantDBRegistry implements TenantDBRegistryInterface
var _ TenantDBRegistryInterface = (*TenantDBRegistry)(nil)