package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xitongsys/parquet-go/writer"
)

// ArchiveConfig holds configuration for the archive service
type ArchiveConfig struct {
	S3Endpoint  string
	S3Region    string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	TablesToArchive []string
	BatchSize       int
	RetentionDays   int
}

// CostAllocationEntry represents a row from the cost_allocation_entries table
type CostAllocationEntry struct {
	ID                 string  `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	TenantID           string  `parquet:"name=tenant_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	APIKeyID           *string `parquet:"name=api_key_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	FunctionID         string  `parquet:"name=function_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	FunctionName       string  `parquet:"name=function_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	FunctionAuthor     string  `parquet:"name=function_author, type=BYTE_ARRAY, convertedtype=UTF8"`
	ExecutionID        string  `parquet:"name=execution_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	ExecutionOutcome   string  `parquet:"name=execution_outcome, type=BYTE_ARRAY, convertedtype=UTF8"`
	Cached             bool    `parquet:"name=cached, type=BOOLEAN"`
	DurationMs         int64   `parquet:"name=duration_ms, type=INT64"`
	ExecutionCostCents int64   `parquet:"name=execution_cost_cents, type=INT64"`
	TotalCostCents     int64   `parquet:"name=total_cost_cents, type=INT64"`
	Region             *string `parquet:"name=region, type=BYTE_ARRAY, convertedtype=UTF8"`
	Timestamp          int64   `parquet:"name=timestamp, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
	TagsJSON           string  `parquet:"name=tags_json, type=BYTE_ARRAY, convertedtype=UTF8"`
	MetadataJSON       string  `parquet:"name=metadata_json, type=BYTE_ARRAY, convertedtype=UTF8"`
}

// ArchiveBatch represents an archived batch of data
type ArchiveBatch struct {
	ID             uuid.UUID
	TableName      string
	MinTimestamp   time.Time
	MaxTimestamp   time.Time
	RowCount       int64
	FileSizeBytes  int64
	ChecksumSHA256 string
	S3Path         string
}

// ArchiveService handles archiving data to S3 in Parquet format
type ArchiveService struct {
	db       *pgx.Conn
	s3Client *s3.Client
	config   ArchiveConfig
}

// NewArchiveService creates a new archive service
func NewArchiveService(cfg ArchiveConfig) (*ArchiveService, error) {
	ctx := context.Background()

	// Initialize S3 client
	var awsCfg aws.Config
	var err error

	if cfg.S3Endpoint != "" {
		// Custom S3-compatible endpoint (MinIO, DigitalOcean Spaces, etc.)
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           cfg.S3Endpoint,
				SigningRegion: cfg.S3Region,
			}, nil
		})

		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithEndpointResolverWithOptions(customResolver),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.S3AccessKey, cfg.S3SecretKey, "",
			)),
			config.WithRegion(cfg.S3Region),
		)
	} else {
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.S3AccessKey, cfg.S3SecretKey, "",
			)),
			config.WithRegion(cfg.S3Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg)

	// Initialize database connection
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &ArchiveService{
		db:       db,
		s3Client: s3Client,
		config:   cfg,
	}, nil
}

// Close closes the service connections
func (s *ArchiveService) Close() error {
	if s.db != nil {
		return s.db.Close(context.Background())
	}
	return nil
}

// ArchiveCostAllocationEntries archives cost allocation entries older than retention days
func (s *ArchiveService) ArchiveCostAllocationEntries(ctx context.Context) (*ArchiveBatch, error) {
	cutoffDate := time.Now().AddDate(0, 0, -s.config.RetentionDays)

	log.Printf("Archiving cost_allocation_entries older than %s", cutoffDate.Format("2006-01-02"))

	// Query data to archive
	query := `
		SELECT 
			id, tenant_id, api_key_id, function_id, function_name, function_author,
			execution_id, execution_outcome, cached, duration_ms, execution_cost_cents, 
			total_cost_cents, region, timestamp, tags::text, metadata::text
		FROM cost_allocation_entries
		WHERE timestamp < $1
		ORDER BY timestamp
		LIMIT $2
	`

	rows, err := s.db.Query(ctx, query, cutoffDate, s.config.BatchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to query data: %w", err)
	}
	defer rows.Close()

	// Create temp parquet file
	tempFile := fmt.Sprintf("/tmp/cost_allocation_%s.parquet", time.Now().Format("20060102_150405"))
	file, err := os.Create(tempFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile)

	// Create parquet writer
	pw, err := writer.NewParquetWriterFromWriter(file, new(CostAllocationEntry), 4)
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet writer: %w", err)
	}

	var minTs, maxTs time.Time
	var count int64

	// Write rows to parquet
	for rows.Next() {
		var entry CostAllocationEntry
		var ts time.Time

		err := rows.Scan(
			&entry.ID, &entry.TenantID, &entry.APIKeyID, &entry.FunctionID,
			&entry.FunctionName, &entry.FunctionAuthor, &entry.ExecutionID,
			&entry.ExecutionOutcome, &entry.Cached, &entry.DurationMs,
			&entry.ExecutionCostCents, &entry.TotalCostCents, &entry.Region,
			&ts, &entry.TagsJSON, &entry.MetadataJSON,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		entry.Timestamp = ts.UnixMilli()

		if count == 0 || ts.Before(minTs) {
			minTs = ts
		}
		if ts.After(maxTs) {
			maxTs = ts
		}

		if err := pw.Write(entry); err != nil {
			log.Printf("Error writing parquet row: %v", err)
			continue
		}

		count++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	if count == 0 {
		log.Println("No data to archive")
		return nil, nil
	}

	// Close parquet writer
	if err := pw.WriteStop(); err != nil {
		return nil, fmt.Errorf("failed to finalize parquet: %w", err)
	}
	file.Close()

	// Calculate checksum
	fileData, err := os.ReadFile(tempFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read temp file: %w", err)
	}

	hash := sha256.Sum256(fileData)
	checksum := hex.EncodeToString(hash[:])

	// Upload to S3
	s3Path := fmt.Sprintf("data-retention/cost-allocation/%s/%s.parquet",
		minTs.Format("2006/01"), uuid.New().String())

	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.S3Bucket),
		Key:         aws.String(s3Path),
		Body:        file,
		ContentType: aws.String("application/octet-stream"),
		Metadata: map[string]string{
			"x-amz-meta-checksum-sha256": checksum,
			"x-amz-meta-row-count":       fmt.Sprintf("%d", count),
			"x-amz-meta-min-timestamp":   minTs.Format(time.RFC3339),
			"x-amz-meta-max-timestamp":   maxTs.Format(time.RFC3339),
			"x-amz-meta-table-name":      "cost_allocation_entries",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Register archive batch in database
	batchID := uuid.New()
	_, err = s.db.Exec(ctx, `
		SELECT prepare_archive_batch(
			$1, $2, $3, $4, $5, $6, $7, 'zstd', 'parquet'
		)
	`, "cost_allocation_entries", minTs, maxTs, count, len(fileData), checksum,
		fmt.Sprintf("s3://%s/%s", s.config.S3Bucket, s3Path))

	if err != nil {
		return nil, fmt.Errorf("failed to register archive batch: %w", err)
	}

	log.Printf("Successfully archived %d rows to s3://%s/%s", count, s.config.S3Bucket, s3Path)

	return &ArchiveBatch{
		ID:             batchID,
		TableName:      "cost_allocation_entries",
		MinTimestamp:   minTs,
		MaxTimestamp:   maxTs,
		RowCount:       count,
		FileSizeBytes:  int64(len(fileData)),
		ChecksumSHA256: checksum,
		S3Path:         s3Path,
	}, nil
}

// VerifyArchiveBatch verifies an archived batch by downloading and checking checksum
func (s *ArchiveService) VerifyArchiveBatch(ctx context.Context, batchID uuid.UUID) error {
	// Get batch info from database
	var s3Path, expectedChecksum string
	err := s.db.QueryRow(ctx, `
		SELECT archive_path, checksum_sha256 
		FROM archive_batches 
		WHERE id = $1
	`, batchID).Scan(&s3Path, &expectedChecksum)

	if err != nil {
		return fmt.Errorf("failed to get batch info: %w", err)
	}

	// Download from S3
	resp, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.S3Bucket),
		Key:    aws.String(s3Path),
	})
	if err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}
	defer resp.Body.Close()

	// Calculate checksum
	hash := sha256.New()
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			hash.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))

	// Verify checksum
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	// Mark as verified
	_, err = s.db.Exec(ctx, `SELECT verify_archive_batch($1, true, NULL)`, batchID)
	if err != nil {
		return fmt.Errorf("failed to mark batch as verified: %w", err)
	}

	log.Printf("Archive batch %s verified successfully", batchID)
	return nil
}

// DeleteArchivedSourceData deletes source data after verification
func (s *ArchiveService) DeleteArchivedSourceData(ctx context.Context, batchID uuid.UUID) error {
	var tableName string
	var minTs, maxTs time.Time

	err := s.db.QueryRow(ctx, `
		SELECT table_name, min_timestamp, max_timestamp 
		FROM archive_batches 
		WHERE id = $1 AND verification_status = 'verified'
	`, batchID).Scan(&tableName, &minTs, &maxTs)

	if err != nil {
		return fmt.Errorf("batch not found or not verified: %w", err)
	}

	// Check legal holds
	var hasHold bool
	err = s.db.QueryRow(ctx, `SELECT is_under_legal_hold($1)`, tableName).Scan(&hasHold)
	if err != nil {
		return fmt.Errorf("failed to check legal holds: %w", err)
	}

	if hasHold {
		return fmt.Errorf("cannot delete: table is under legal hold")
	}

	// Delete data in batches
	query := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE timestamp >= $1 AND timestamp <= $2
	`, tableName)

	_, err = s.db.Exec(ctx, query, minTs, maxTs)
	if err != nil {
		return fmt.Errorf("failed to delete source data: %w", err)
	}

	// Mark as deleted
	_, err = s.db.Exec(ctx, `SELECT confirm_source_deleted($1)`, batchID)
	if err != nil {
		return fmt.Errorf("failed to mark as deleted: %w", err)
	}

	log.Printf("Deleted archived source data for batch %s", batchID)
	return nil
}

func main() {
	// Load configuration from environment
	config := ArchiveConfig{
		S3Endpoint:  getEnv("S3_ENDPOINT", ""),
		S3Region:    getEnv("S3_REGION", "us-east-1"),
		S3Bucket:    getEnv("S3_BUCKET", "functionfly-archives"),
		S3AccessKey: getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey: getEnv("S3_SECRET_KEY", ""),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "functionfly"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		TablesToArchive: []string{"cost_allocation_entries"},
		BatchSize:       100000,
		RetentionDays:   90,
	}

	service, err := NewArchiveService(config)
	if err != nil {
		log.Fatalf("Failed to create archive service: %v", err)
	}
	defer service.Close()

	ctx := context.Background()

	// Run archive process
	batch, err := service.ArchiveCostAllocationEntries(ctx)
	if err != nil {
		log.Fatalf("Archive failed: %v", err)
	}

	if batch == nil {
		log.Println("No data to archive")
		return
	}

	log.Printf("Archived %d rows (%d bytes) to S3", batch.RowCount, batch.FileSizeBytes)

	// Get batch ID from database
	var batchID uuid.UUID
	err = service.db.QueryRow(ctx, `
		SELECT id FROM archive_batches 
		WHERE table_name = $1 AND min_timestamp = $2 AND max_timestamp = $3
		ORDER BY created_at DESC LIMIT 1
	`, batch.TableName, batch.MinTimestamp, batch.MaxTimestamp).Scan(&batchID)

	if err != nil {
		log.Fatalf("Failed to get batch ID: %v", err)
	}

	// Verify the archive
	if err := service.VerifyArchiveBatch(ctx, batchID); err != nil {
		log.Fatalf("Verification failed: %v", err)
	}

	log.Println("Archive verified successfully")

	// Optionally delete source data (use with caution!)
	// Uncomment only after thorough testing:
	// if err := service.DeleteArchivedSourceData(ctx, batchID); err != nil {
	//     log.Fatalf("Failed to delete source data: %v", err)
	// }
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
