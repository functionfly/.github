package storage

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

// PostgresArchiveStorage implements ArchiveStorage using PostgreSQL
type PostgresArchiveStorage struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewPostgresArchiveStorage creates a new PostgreSQL archive storage instance
func NewPostgresArchiveStorage(db *sql.DB) *PostgresArchiveStorage {
	return &PostgresArchiveStorage{
		db:     db,
		logger: logrus.New(),
	}
}

// StoreArchive stores compressed and encrypted archive data in PostgreSQL
func (p *PostgresArchiveStorage) StoreArchive(ctx context.Context, key string, data io.Reader, metadata *ArchiveMetadata) error {
	// Read and compress data
	dataBytes, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("failed to read archive data: %w", err)
	}

	// Compress data
	compressedData, err := p.compressData(dataBytes)
	if err != nil {
		return fmt.Errorf("failed to compress archive data: %w", err)
	}

	// Encrypt compressed data
	encryptedData, encryptionKey, err := p.encryptData(compressedData)
	if err != nil {
		return fmt.Errorf("failed to encrypt archive data: %w", err)
	}

	// Calculate checksum of original data
	checksum := fmt.Sprintf("%x", sha256.Sum256(dataBytes))

	// Update metadata
	metadata.Checksum = checksum
	metadata.CompressedSize = int64(len(compressedData))
	metadata.OriginalSize = int64(len(dataBytes))
	if metadata.OriginalSize > 0 {
		metadata.CompressionRatio = float64(metadata.CompressedSize) / float64(metadata.OriginalSize)
	}
	metadata.CreatedAt = time.Now()
	metadata.Status = "completed"
	metadata.StorageKey = key

	// Store encryption key ID (in production, this should be stored securely)
	metadata.EncryptionKeyID = "postgres-archive-key-v1"

	// Convert metadata to JSON for storage
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Insert into database
	query := `
		INSERT INTO archive_data (
			id, storage_key, archive_type, compressed_data, encryption_key,
			metadata_json, created_at, checksum, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (storage_key) DO UPDATE SET
			compressed_data = EXCLUDED.compressed_data,
			encryption_key = EXCLUDED.encryption_key,
			metadata_json = EXCLUDED.metadata_json,
			updated_at = NOW(),
			checksum = EXCLUDED.checksum,
			status = EXCLUDED.status
	`

	_, err = p.db.ExecContext(ctx, query,
		metadata.ID,
		key,
		metadata.ArchiveType,
		encryptedData,
		encryptionKey,
		metadataJSON,
		metadata.CreatedAt,
		checksum,
		metadata.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to store archive in database: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"key":              key,
		"original_size":    metadata.OriginalSize,
		"compressed_size":  metadata.CompressedSize,
		"compression_ratio": fmt.Sprintf("%.2f", metadata.CompressionRatio),
		"record_count":     metadata.RecordCount,
	}).Info("Archive stored successfully in PostgreSQL")

	return nil
}

// RetrieveArchive retrieves archive data from PostgreSQL
func (p *PostgresArchiveStorage) RetrieveArchive(ctx context.Context, key string) (io.ReadCloser, error) {
	query := `
		SELECT compressed_data, encryption_key, metadata_json
		FROM archive_data
		WHERE storage_key = $1 AND status = 'completed'
	`

	var compressedData, encryptionKey []byte
	var metadataJSON []byte

	err := p.db.QueryRowContext(ctx, query, key).Scan(&compressedData, &encryptionKey, &metadataJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("archive not found: %s", key)
		}
		return nil, fmt.Errorf("failed to retrieve archive: %w", err)
	}

	// Decrypt data
	decryptedData, err := p.decryptData(compressedData, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt archive data: %w", err)
	}

	// Decompress data
	originalData, err := p.decompressData(decryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress archive data: %w", err)
	}

	return io.NopCloser(bytes.NewReader(originalData)), nil
}

// DeleteArchive marks an archive as deleted (soft delete for compliance)
func (p *PostgresArchiveStorage) DeleteArchive(ctx context.Context, key string) error {
	query := `
		UPDATE archive_data
		SET status = 'deleted', updated_at = NOW()
		WHERE storage_key = $1 AND status = 'completed'
	`

	result, err := p.db.ExecContext(ctx, query, key)
	if err != nil {
		return fmt.Errorf("failed to delete archive: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("archive not found or already deleted: %s", key)
	}

	p.logger.WithField("key", key).Info("Archive marked as deleted")
	return nil
}

// ListArchives lists archives with optional prefix filtering
func (p *PostgresArchiveStorage) ListArchives(ctx context.Context, prefix string) ([]*ArchiveMetadata, error) {
	query := `
		SELECT metadata_json
		FROM archive_data
		WHERE status = 'completed'
	`
	args := []interface{}{}

	if prefix != "" {
		query += " AND storage_key LIKE $1"
		args = append(args, prefix+"%")
	}

	query += " ORDER BY created_at DESC"

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list archives: %w", err)
	}
	defer rows.Close()

	var archives []*ArchiveMetadata
	for rows.Next() {
		var metadataJSON []byte
		if err := rows.Scan(&metadataJSON); err != nil {
			return nil, fmt.Errorf("failed to scan archive metadata: %w", err)
		}

		var metadata ArchiveMetadata
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		archives = append(archives, &metadata)
	}

	return archives, rows.Err()
}

// GetArchiveMetadata retrieves metadata for a specific archive
func (p *PostgresArchiveStorage) GetArchiveMetadata(ctx context.Context, key string) (*ArchiveMetadata, error) {
	query := `
		SELECT metadata_json
		FROM archive_data
		WHERE storage_key = $1 AND status = 'completed'
	`

	var metadataJSON []byte
	err := p.db.QueryRowContext(ctx, query, key).Scan(&metadataJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("archive metadata not found: %s", key)
		}
		return nil, fmt.Errorf("failed to retrieve archive metadata: %w", err)
	}

	var metadata ArchiveMetadata
	if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// HealthCheck performs a health check on the PostgreSQL storage
func (p *PostgresArchiveStorage) HealthCheck(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// compressData compresses data using gzip
func (p *PostgresArchiveStorage) compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	if _, err := gzipWriter.Write(data); err != nil {
		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompressData decompresses gzip data
func (p *PostgresArchiveStorage) decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// encryptData encrypts data using AES-GCM
func (p *PostgresArchiveStorage) encryptData(data []byte) ([]byte, []byte, error) {
	// Generate a random 32-byte key for AES-256
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext, key, nil
}

// decryptData decrypts AES-GCM encrypted data
func (p *PostgresArchiveStorage) decryptData(encryptedData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}