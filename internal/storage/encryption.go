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
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/scrypt"
)

// DatabaseEncryptionManager handles database encryption at rest and key management
type DatabaseEncryptionManager struct {
	db            *PostgresDB
	masterKey     []byte
	dataKey       []byte
	keyVersion    string
	encryptionEnabled bool
	logger        *logrus.Logger
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	Enabled         bool
	KeyRotationDays int
	Algorithm       string
	KeySize         int
}

// EncryptedField represents a field that should be encrypted
type EncryptedField struct {
	TableName  string
	ColumnName string
	FieldType  string
}

// EncryptionKey represents an encryption key with metadata
type EncryptionKey struct {
	ID          string    `json:"id"`
	Version     string    `json:"version"`
	Key         string    `json:"encrypted_key"` // encrypted with master key
	Algorithm   string    `json:"algorithm"`
	Purpose     string    `json:"purpose"` // "master", "data", "field"
	CreatedAt   time.Time `json:"created_at"`
	Active      bool      `json:"active"`
	RotatedAt   *time.Time `json:"rotated_at,omitempty"`
}

// NewDatabaseEncryptionManager creates a new database encryption manager
func NewDatabaseEncryptionManager(db *PostgresDB) (*DatabaseEncryptionManager, error) {
	dem := &DatabaseEncryptionManager{
		db:         db,
		logger:     logrus.New(),
	}

	// Check if encryption is enabled via environment variable
	encryptionEnabled := os.Getenv("DB_ENCRYPTION_ENABLED")
	dem.encryptionEnabled = strings.ToLower(encryptionEnabled) == "true"

	if dem.encryptionEnabled {
		if err := dem.initializeEncryption(); err != nil {
			return nil, fmt.Errorf("failed to initialize encryption: %w", err)
		}
	}

	return dem, nil
}

// initializeEncryption sets up encryption keys and database structures
func (dem *DatabaseEncryptionManager) initializeEncryption() error {
	dem.logger.Info("Initializing database encryption")

	// Create encryption keys table if it doesn't exist
	if err := dem.createEncryptionTables(); err != nil {
		return fmt.Errorf("failed to create encryption tables: %w", err)
	}

	// Load or generate master key
	if err := dem.loadOrGenerateMasterKey(); err != nil {
		return fmt.Errorf("failed to load/generate master key: %w", err)
	}

	// Load or generate data encryption key
	if err := dem.loadOrGenerateDataKey(); err != nil {
		return fmt.Errorf("failed to load/generate data key: %w", err)
	}

	// Set up encrypted fields
	if err := dem.setupEncryptedFields(); err != nil {
		return fmt.Errorf("failed to setup encrypted fields: %w", err)
	}

	dem.logger.Info("Database encryption initialized successfully")
	return nil
}

// createEncryptionTables creates necessary tables for encryption key management
func (dem *DatabaseEncryptionManager) createEncryptionTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS encryption_keys (
			id VARCHAR(36) PRIMARY KEY,
			version VARCHAR(50) NOT NULL,
			encrypted_key TEXT NOT NULL,
			algorithm VARCHAR(50) NOT NULL,
			purpose VARCHAR(20) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			active BOOLEAN DEFAULT TRUE,
			rotated_at TIMESTAMP WITH TIME ZONE,
			UNIQUE(version, purpose)
		)`,

		`CREATE TABLE IF NOT EXISTS encrypted_fields (
			id SERIAL PRIMARY KEY,
			table_name VARCHAR(100) NOT NULL,
			column_name VARCHAR(100) NOT NULL,
			field_type VARCHAR(50) NOT NULL,
			encryption_key_version VARCHAR(50) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(table_name, column_name)
		)`,

		`CREATE INDEX IF NOT EXISTS idx_encryption_keys_version_purpose ON encryption_keys(version, purpose)`,
		`CREATE INDEX IF NOT EXISTS idx_encryption_keys_active ON encryption_keys(active)`,
		`CREATE INDEX IF NOT EXISTS idx_encrypted_fields_table_column ON encrypted_fields(table_name, column_name)`,
	}

	for _, query := range queries {
		if _, err := dem.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute encryption table query: %w", err)
		}
	}

	return nil
}

// loadOrGenerateMasterKey loads existing master key or generates a new one
func (dem *DatabaseEncryptionManager) loadOrGenerateMasterKey() error {
	// Try to load existing master key from database
	var encryptedKey string
	var version string
	query := `SELECT version, encrypted_key FROM encryption_keys WHERE purpose = 'master' AND active = true ORDER BY created_at DESC LIMIT 1`

	err := dem.db.QueryRow(query).Scan(&version, &encryptedKey)
	if err == nil {
		// Decrypt the master key using KEK from environment
		kekPassword := os.Getenv("DB_MASTER_KEY_PASSWORD")
		if kekPassword == "" {
			return fmt.Errorf("DB_MASTER_KEY_PASSWORD environment variable not set")
		}

		kek, err := scrypt.Key([]byte(kekPassword), []byte("db-master-key-salt"), 32768, 8, 1, 32)
		if err != nil {
			return err
		}

		dem.masterKey, err = dem.decryptWithKey(encryptedKey, kek)
		if err != nil {
			return fmt.Errorf("failed to decrypt master key: %w", err)
		}

		dem.keyVersion = version
		dem.logger.WithField("version", version).Info("Loaded existing master key")
		return nil
	}

	if err != sql.ErrNoRows {
		return err
	}

	// Generate new master key
	dem.logger.Info("Generating new master key")
	masterKey, version, err := dem.generateEncryptionKey("master")
	if err != nil {
		return err
	}

	dem.masterKey = masterKey
	dem.keyVersion = version

	// Encrypt and store the master key
	if err := dem.storeMasterKey(masterKey, version); err != nil {
		return err
	}

	return nil
}

// loadOrGenerateDataKey loads existing data key or generates a new one
func (dem *DatabaseEncryptionManager) loadOrGenerateDataKey() error {
	// Try to load existing data key
	var encryptedKey string
	var version string
	query := `SELECT version, encrypted_key FROM encryption_keys WHERE purpose = 'data' AND active = true ORDER BY created_at DESC LIMIT 1`

	err := dem.db.QueryRow(query).Scan(&version, &encryptedKey)
	if err == nil {
		// Decrypt the data key using master key
		dem.dataKey, err = dem.decryptWithKey(encryptedKey, dem.masterKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt data key: %w", err)
		}

		dem.logger.WithField("version", version).Info("Loaded existing data key")
		return nil
	}

	if err != sql.ErrNoRows {
		return err
	}

	// Generate new data key
	dem.logger.Info("Generating new data key")
	dataKey, version, err := dem.generateEncryptionKey("data")
	if err != nil {
		return err
	}

	dem.dataKey = dataKey

	// Encrypt data key with master key and store
	encryptedDataKey, err := dem.encryptWithKey(dataKey, dem.masterKey)
	if err != nil {
		return err
	}

	if err := dem.storeDataKey(encryptedDataKey, version); err != nil {
		return err
	}

	return nil
}

// generateEncryptionKey generates a new encryption key
func (dem *DatabaseEncryptionManager) generateEncryptionKey(purpose string) ([]byte, string, error) {
	key := make([]byte, 32) // 256-bit key
	if _, err := rand.Read(key); err != nil {
		return nil, "", err
	}

	version := fmt.Sprintf("%s_%d", purpose, time.Now().Unix())
	return key, version, nil
}

// storeMasterKey stores the encrypted master key in database
func (dem *DatabaseEncryptionManager) storeMasterKey(key []byte, version string) error {
	// Encrypt master key with KEK from environment
	kekPassword := os.Getenv("DB_MASTER_KEY_PASSWORD")
	if kekPassword == "" {
		return fmt.Errorf("DB_MASTER_KEY_PASSWORD environment variable not set")
	}

	kek, err := scrypt.Key([]byte(kekPassword), []byte("db-master-key-salt"), 32768, 8, 1, 32)
	if err != nil {
		return err
	}

	encryptedKey, err := dem.encryptWithKey(key, kek)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO encryption_keys (id, version, encrypted_key, algorithm, purpose, active)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, true)`

	_, err = dem.db.Exec(query, version, encryptedKey, "AES-256-GCM", "master")
	return err
}

// storeDataKey stores the encrypted data key in database
func (dem *DatabaseEncryptionManager) storeDataKey(encryptedKey, version string) error {
	query := `
		INSERT INTO encryption_keys (id, version, encrypted_key, algorithm, purpose, active)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, true)`

	_, err := dem.db.Exec(query, version, encryptedKey, "AES-256-GCM", "data")
	return err
}

// setupEncryptedFields sets up encrypted fields configuration
func (dem *DatabaseEncryptionManager) setupEncryptedFields() error {
	// Define fields that should be encrypted
	encryptedFields := []EncryptedField{
		{TableName: "users", ColumnName: "email", FieldType: "text"},
		{TableName: "users", ColumnName: "password_hash", FieldType: "text"},
		{TableName: "users", ColumnName: "verification_token", FieldType: "text"},
		{TableName: "tenants", ColumnName: "name", FieldType: "text"},
		{TableName: "backends", ColumnName: "shared_secret", FieldType: "text"},
		{TableName: "feedback", ColumnName: "message", FieldType: "text"},
		{TableName: "feedback", ColumnName: "user_email", FieldType: "text"},
		{TableName: "providers", ColumnName: "token", FieldType: "text"},
		{TableName: "function_dna_mutations", ColumnName: "original_code", FieldType: "text"},
		{TableName: "function_dna_mutations", ColumnName: "mutated_code", FieldType: "text"},
	}

	for _, field := range encryptedFields {
		if err := dem.registerEncryptedField(field); err != nil {
			dem.logger.WithFields(logrus.Fields{
				"table":  field.TableName,
				"column": field.ColumnName,
			}).WithError(err).Warn("Failed to register encrypted field")
		}
	}

	return nil
}

// registerEncryptedField registers a field for encryption
func (dem *DatabaseEncryptionManager) registerEncryptedField(field EncryptedField) error {
	query := `
		INSERT INTO encrypted_fields (table_name, column_name, field_type, encryption_key_version)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (table_name, column_name) DO NOTHING`

	_, err := dem.db.Exec(query, field.TableName, field.ColumnName, field.FieldType, dem.keyVersion)
	return err
}

// EncryptField encrypts a field value before storing in database
func (dem *DatabaseEncryptionManager) EncryptField(value string) (string, error) {
	if !dem.encryptionEnabled {
		return value, nil
	}

	if value == "" {
		return value, nil
	}

	encrypted, err := dem.encryptWithKey([]byte(value), dem.dataKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt field: %w", err)
	}

	// Prefix to identify encrypted values
	return "ENC:" + encrypted, nil
}

// DecryptField decrypts a field value when reading from database
func (dem *DatabaseEncryptionManager) DecryptField(value string) (string, error) {
	if !dem.encryptionEnabled {
		return value, nil
	}

	if value == "" || !strings.HasPrefix(value, "ENC:") {
		return value, nil
	}

	encrypted := strings.TrimPrefix(value, "ENC:")
	decrypted, err := dem.decryptWithKey(encrypted, dem.dataKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt field: %w", err)
	}

	return string(decrypted), nil
}

// RotateKeys performs key rotation for enhanced security
func (dem *DatabaseEncryptionManager) RotateKeys(ctx context.Context) error {
	dem.logger.Info("Starting database encryption key rotation")

	// Generate new data key
	newDataKey, newVersion, err := dem.generateEncryptionKey("data")
	if err != nil {
		return fmt.Errorf("failed to generate new data key: %w", err)
	}

	// Encrypt new data key with master key
	encryptedNewDataKey, err := dem.encryptWithKey(newDataKey, dem.masterKey)
	if err != nil {
		return err
	}

	// Store new data key
	if err := dem.storeDataKey(encryptedNewDataKey, newVersion); err != nil {
		return err
	}

	// Re-encrypt all encrypted fields with new key
	if err := dem.reEncryptAllFields(dem.dataKey, newDataKey, newVersion); err != nil {
		return err
	}

	// Mark old key as rotated
	if err := dem.markKeyRotated(dem.keyVersion); err != nil {
		return err
	}

	// Update current key
	dem.dataKey = newDataKey
	dem.keyVersion = newVersion

	// Log key rotation
	event := &AuditEvent{
		Action:       "security.database_key_rotation",
		ResourceType: "encryption_key",
		ResourceID:   nil, // Key version is stored in metadata
		BeforeState:  map[string]interface{}{"key_version": dem.keyVersion},
		AfterState:   map[string]interface{}{"key_version": newVersion},
		Success:      true,
	}

	if err := dem.db.Repository().LogAuditEvent(ctx, event); err != nil {
		dem.logger.WithError(err).Warn("Failed to log key rotation audit event")
	}

	dem.logger.WithField("new_version", newVersion).Info("Database key rotation completed")
	return nil
}

// reEncryptAllFields re-encrypts all encrypted fields with new key
func (dem *DatabaseEncryptionManager) reEncryptAllFields(oldKey, newKey []byte, newVersion string) error {
	// Get all encrypted fields
	rows, err := dem.db.Query(`
		SELECT table_name, column_name
		FROM encrypted_fields
		WHERE encryption_key_version != $1`, newVersion)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName, columnName string
		if err := rows.Scan(&tableName, &columnName); err != nil {
			continue
		}

		if err := dem.reEncryptField(tableName, columnName, oldKey, newKey); err != nil {
			dem.logger.WithFields(logrus.Fields{
				"table":  tableName,
				"column": columnName,
			}).WithError(err).Warn("Failed to re-encrypt field")
		}
	}

	return nil
}

var validIdentifierChars = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func isValidIdentifier(name string) bool {
	return validIdentifierChars.MatchString(name) && len(name) <= 63
}

// reEncryptField re-encrypts a specific field with new key
func (dem *DatabaseEncryptionManager) reEncryptField(tableName, columnName string, oldKey, newKey []byte) error {
	if !isValidIdentifier(tableName) || !isValidIdentifier(columnName) {
		return fmt.Errorf("invalid table or column name: %s.%s", tableName, columnName)
	}
	selectQuery := fmt.Sprintf("SELECT id, %s FROM %s WHERE %s LIKE 'ENC:%%'", columnName, tableName, columnName)
	rows, err := dem.db.Query(selectQuery)
	if err != nil {
		return fmt.Errorf("failed to query encrypted fields: %w", err)
	}
	defer rows.Close()

	// Process each encrypted value
	for rows.Next() {
		var id interface{}
		var encryptedValue string

		if err := rows.Scan(&id, &encryptedValue); err != nil {
			dem.logger.WithError(err).Warn("Failed to scan encrypted field row")
			continue
		}

		// Decrypt with old key
		decrypted, err := dem.decryptWithKey(strings.TrimPrefix(encryptedValue, "ENC:"), oldKey)
		if err != nil {
			dem.logger.WithFields(logrus.Fields{
				"table":  tableName,
				"column": columnName,
				"id":     id,
			}).WithError(err).Warn("Failed to decrypt field value during re-encryption")
			continue
		}

		// Encrypt with new key
		newEncrypted, err := dem.encryptWithKey(decrypted, newKey)
		if err != nil {
			dem.logger.WithFields(logrus.Fields{
				"table":  tableName,
				"column": columnName,
				"id":     id,
			}).WithError(err).Warn("Failed to encrypt field value with new key")
			continue
		}

		// Update the field with new encrypted value
		updateQuery := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE id = $2", tableName, columnName)
		_, err = dem.db.Exec(updateQuery, "ENC:"+newEncrypted, id)
		if err != nil {
			dem.logger.WithFields(logrus.Fields{
				"table":  tableName,
				"column": columnName,
				"id":     id,
			}).WithError(err).Warn("Failed to update field with re-encrypted value")
			continue
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating through encrypted field rows: %w", err)
	}

	return nil
}

// markKeyRotated marks a key as rotated
func (dem *DatabaseEncryptionManager) markKeyRotated(version string) error {
	query := `UPDATE encryption_keys SET rotated_at = NOW() WHERE version = $1`
	_, err := dem.db.Exec(query, version)
	return err
}

// ShouldRotateKeys checks if keys should be rotated
func (dem *DatabaseEncryptionManager) ShouldRotateKeys() bool {
	if !dem.encryptionEnabled {
		return false
	}

	rotationDays := 90 // Default 90 days
	if daysStr := os.Getenv("DB_KEY_ROTATION_DAYS"); daysStr != "" {
		if parsed, err := strconv.Atoi(daysStr); err == nil && parsed > 0 {
			rotationDays = parsed
		}
	}

	var lastRotation time.Time
	query := `SELECT created_at FROM encryption_keys WHERE purpose = 'data' AND active = true ORDER BY created_at DESC LIMIT 1`

	err := dem.db.QueryRow(query).Scan(&lastRotation)
	if err != nil {
		return false
	}

	return time.Since(lastRotation) > time.Duration(rotationDays)*24*time.Hour
}

// IsEncryptionEnabled returns whether encryption is enabled
func (dem *DatabaseEncryptionManager) IsEncryptionEnabled() bool {
	return dem.encryptionEnabled
}

// GetEncryptionStatus returns encryption status information
func (dem *DatabaseEncryptionManager) GetEncryptionStatus() map[string]interface{} {
	status := map[string]interface{}{
		"enabled": dem.encryptionEnabled,
		"key_version": dem.keyVersion,
		"algorithm": "AES-256-GCM",
	}

	if dem.encryptionEnabled {
		var keyCount int
		dem.db.QueryRow(`SELECT COUNT(*) FROM encryption_keys WHERE active = true`).Scan(&keyCount)
		status["active_keys"] = keyCount

		var fieldCount int
		dem.db.QueryRow(`SELECT COUNT(*) FROM encrypted_fields`).Scan(&fieldCount)
		status["encrypted_fields"] = fieldCount

		status["should_rotate"] = dem.ShouldRotateKeys()
	}

	return status
}

// encryptWithKey encrypts data with a specific key
func (dem *DatabaseEncryptionManager) encryptWithKey(data, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptWithKey decrypts data with a specific key
func (dem *DatabaseEncryptionManager) decryptWithKey(ciphertext string, key []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EnableEncryption enables database encryption
func (dem *DatabaseEncryptionManager) EnableEncryption() error {
	if dem.encryptionEnabled {
		return fmt.Errorf("encryption is already enabled")
	}

	dem.encryptionEnabled = true
	return dem.initializeEncryption()
}

// DisableEncryption disables database encryption (WARNING: This will expose data)
func (dem *DatabaseEncryptionManager) DisableEncryption() error {
	if !dem.encryptionEnabled {
		return fmt.Errorf("encryption is not enabled")
	}

	dem.logger.Warn("Disabling database encryption - data will be exposed")
	dem.encryptionEnabled = false
	return nil
}

// BackupEncryptionKeys backs up encryption keys to secure location
func (dem *DatabaseEncryptionManager) BackupEncryptionKeys(backupPath string) error {
	if err := os.MkdirAll(backupPath, 0700); err != nil {
		return err
	}

	// Export keys to encrypted backup file
	backupFile := filepath.Join(backupPath, fmt.Sprintf("encryption_keys_backup_%d.json", time.Now().Unix()))

	keys, err := dem.exportKeys()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return err
	}

	// Encrypt backup with KEK
	kekPassword := os.Getenv("DB_MASTER_KEY_PASSWORD")
	if kekPassword == "" {
		return fmt.Errorf("DB_MASTER_KEY_PASSWORD environment variable not set")
	}

	kek, err := scrypt.Key([]byte(kekPassword), []byte("backup-key-salt"), 32768, 8, 1, 32)
	if err != nil {
		return err
	}

	encryptedBackup, err := dem.encryptWithKey(data, kek)
	if err != nil {
		return err
	}

	return os.WriteFile(backupFile, []byte(encryptedBackup), 0600)
}

// exportKeys exports all encryption keys for backup
func (dem *DatabaseEncryptionManager) exportKeys() ([]EncryptionKey, error) {
	query := `SELECT id, version, encrypted_key, algorithm, purpose, created_at, active, rotated_at FROM encryption_keys`

	rows, err := dem.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []EncryptionKey
	for rows.Next() {
		var key EncryptionKey
		var rotatedAt *time.Time
		err := rows.Scan(&key.ID, &key.Version, &key.Key, &key.Algorithm, &key.Purpose, &key.CreatedAt, &key.Active, &rotatedAt)
		if err != nil {
			continue
		}
		key.RotatedAt = rotatedAt
		keys = append(keys, key)
	}

	return keys, nil
}