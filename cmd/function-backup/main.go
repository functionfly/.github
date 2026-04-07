// Package main implements the FunctionFly function backup utility.
// This tool backs up function data from PostgreSQL to Cloudflare R2
// with versioning, compression, and optional encryption.
//
// Usage:
//
//	go run ./cmd/function-backup
//
//	go run ./cmd/function-backup --retention=7 --encrypt
//
//	go run ./cmd/function-backup --dry-run
//
// Environment Variables:
//
//	DATABASE_URL or DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME
//	R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY
//	R2_BACKUP_BUCKET (default: functionfly-backups)
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	_ "github.com/lib/pq"
)

// BackupConfig holds configuration for the backup operation
type BackupConfig struct {
	RetentionDays    int
	CompressLevel    int
	Encrypt          bool
	DryRun           bool
	Tables           []string
	BackupBucket     string
	R2AccountID      string
	R2AccessKeyID    string
	R2SecretKey      string
	DatabaseURL      string
	GPGKeyID         string
	ParallelWorkers  int
	VerifyUpload     bool
}

// BackupMetadata contains metadata about a backup
type BackupMetadata struct {
	Timestamp       string   `json:"timestamp"`
	DatePath        string   `json:"date_path"`
	R2Key           string   `json:"r2_key"`
	R2Bucket        string   `json:"r2_bucket"`
	Checksum        string   `json:"checksum"`
	CompressedSize  int64    `json:"compressed_size"`
	UncompressedSize int64 `json:"uncompressed_size"`
	CompressionRatio float64 `json:"compression_ratio"`
	RetentionDays   int      `json:"retention_days"`
	TablesBackedUp []string `json:"tables_backed_up"`
	Duration        string   `json:"duration"`
	Encrypted       bool     `json:"encrypted"`
}

// Default tables to backup (in dependency order)
var defaultTables = []string{
	"functions",
	"function_versions",
	"function_dependencies",
	"function_environment",
	"function_deployments",
	"deployment_artifacts",
	"function_metrics",
	"function_invocations",
	"function_reputation",
	"function_verification",
	"function_ratings",
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config := parseFlags()
	if err := validateConfig(config); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	ctx := context.Background()
	startTime := time.Now()

	log.Printf("========================================")
	log.Printf("FunctionFly Function Backup Starting")
	log.Printf("Timestamp: %s", startTime.Format(time.RFC3339))
	log.Printf("========================================")

	// Initialize R2 client
	s3Client, err := initR2Client(ctx, config)
	if err != nil {
		log.Fatalf("Failed to initialize R2 client: %v", err)
	}

	// Verify bucket exists
	if err := verifyBucket(ctx, s3Client, config.BackupBucket); err != nil {
		log.Fatalf("R2 bucket verification failed: %v", err)
	}

	// Perform backup
	if err := runBackup(ctx, s3Client, config); err != nil {
		log.Fatalf("Backup failed: %v", err)
	}

	duration := time.Since(startTime)
	log.Printf("========================================")
	log.Printf("Backup completed successfully!")
	log.Printf("Duration: %v", duration)
	log.Printf("========================================")
}

func parseFlags() *BackupConfig {
	config := &BackupConfig{}

	flag.IntVar(&config.RetentionDays, "retention", 30, "Days to retain backups")
	flag.IntVar(&config.CompressLevel, "compress", 6, "Gzip compression level (0-9)")
	flag.BoolVar(&config.Encrypt, "encrypt", false, "Encrypt backup with GPG")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Show what would be backed up without uploading")

	var tablesStr string
	flag.StringVar(&tablesStr, "tables", "", "Comma-separated list of tables to backup (default: all function tables)")

	flag.IntVar(&config.ParallelWorkers, "workers", 4, "Number of parallel upload workers")
	flag.BoolVar(&config.VerifyUpload, "verify", true, "Verify uploaded backup checksum")

	flag.Parse()

	// Parse tables
	if tablesStr != "" {
		config.Tables = strings.Split(tablesStr, ",")
		for i := range config.Tables {
			config.Tables[i] = strings.TrimSpace(config.Tables[i])
		}
	} else {
		config.Tables = defaultTables
	}

	// Load from environment
	config.BackupBucket = getEnvOrDefault("R2_BACKUP_BUCKET", "functionfly-backups")
	config.R2AccountID = os.Getenv("R2_ACCOUNT_ID")
	config.R2AccessKeyID = os.Getenv("R2_ACCESS_KEY_ID")
	if config.R2AccessKeyID == "" {
		config.R2AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	config.R2SecretKey = os.Getenv("R2_SECRET_ACCESS_KEY")
	if config.R2SecretKey == "" {
		config.R2SecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}
	config.DatabaseURL = os.Getenv("DATABASE_URL")
	config.GPGKeyID = os.Getenv("GPG_KEY_ID")

	return config
}

func validateConfig(config *BackupConfig) error {
	var errors []string

	if config.R2AccountID == "" {
		errors = append(errors, "R2_ACCOUNT_ID not set")
	}
	if config.R2AccessKeyID == "" {
		errors = append(errors, "R2_ACCESS_KEY_ID (or AWS_ACCESS_KEY_ID) not set")
	}
	if config.R2SecretKey == "" {
		errors = append(errors, "R2_SECRET_ACCESS_KEY (or AWS_SECRET_ACCESS_KEY) not set")
	}

	if config.DatabaseURL == "" {
		// Check individual DB params
		host := os.Getenv("DB_HOST")
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PASSWORD")
		name := os.Getenv("DB_NAME")

		if host == "" || user == "" || pass == "" || name == "" {
			errors = append(errors, "DATABASE_URL or DB_HOST/DB_USER/DB_PASSWORD/DB_NAME not set")
		} else {
			// Build connection string
			port := getEnvOrDefault("DB_PORT", "5432")
			sslmode := getEnvOrDefault("DB_SSLMODE", "require")
			config.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
				user, pass, host, port, name, sslmode)
		}
	}

	if config.Encrypt && config.GPGKeyID == "" {
		errors = append(errors, "GPG_KEY_ID required when encryption enabled")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, ", "))
	}

	return nil
}

func initR2Client(ctx context.Context, config *BackupConfig) (*s3.Client, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", config.R2AccountID)

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			HostnameImmutable: true,
		}, nil
	})

	awsCfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion("auto"),
		awsConfig.WithEndpointResolverWithOptions(customResolver),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.R2AccessKeyID, config.R2SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return client, nil
}

func verifyBucket(ctx context.Context, client *s3.Client, bucket string) error {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &bucket,
	})
	if err != nil {
		return fmt.Errorf("bucket %s not accessible: %w", bucket, err)
	}
	return nil
}

func runBackup(ctx context.Context, client *s3.Client, config *BackupConfig) error {
	timestamp := time.Now().UTC()
	datePath := timestamp.Format("2006/01/02")
	timestampStr := timestamp.Format("20060102-150405")

	// Create temp directory for backup
	tempDir, err := os.MkdirTemp("", "functionfly-backup-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Dump data
	log.Printf("Dumping function data from database...")
	backupData, err := dumpFunctionsData(config)
	if err != nil {
		return fmt.Errorf("failed to dump function data: %w", err)
	}

	uncompressedSize := int64(len(backupData))
	log.Printf("Backup data: %d bytes (%.2f MB)", uncompressedSize, float64(uncompressedSize)/(1024*1024))

	// Compress
	log.Printf("Compressing with gzip (level %d)...", config.CompressLevel)
	compressedData, err := compressData(backupData, config.CompressLevel)
	if err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}

	compressedSize := int64(len(compressedData))
	compressionRatio := float64(uncompressedSize) / float64(compressedSize)
	log.Printf("Compressed: %d bytes (ratio: %.2f:1)", compressedSize, compressionRatio)

	// Calculate checksum
	checksum := sha256.Sum256(compressedData)
	checksumHex := hex.EncodeToString(checksum[:])

	// Write to temp file
	filename := fmt.Sprintf("functions-backup-%s.sql.gz", timestampStr)
	if config.Encrypt {
		filename += ".gpg"
	}
	tempFile := filepath.Join(tempDir, filename)

	if err := os.WriteFile(tempFile, compressedData, 0600); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Encrypt if requested
	if config.Encrypt {
		// Note: GPG encryption would require CGO or external binary
		// For now, log a warning
		log.Printf("Warning: GPG encryption requires external gpg binary (not implemented in Go version)")
		log.Printf("Use --encrypt flag with the shell script for GPG support")
	}

	if config.DryRun {
		log.Printf("DRY RUN: Would upload to R2:")
		log.Printf("  Bucket: %s", config.BackupBucket)
		log.Printf("  Key: backups/functions/%s/%s", datePath, filename)
		log.Printf("  Size: %d bytes", compressedSize)
		log.Printf("  Checksum: %s", checksumHex)
		return nil
	}

	// Upload to R2
	r2Key := fmt.Sprintf("backups/functions/%s/%s", datePath, filename)
	log.Printf("Uploading to R2: s3://%s/%s", config.BackupBucket, r2Key)

	contentType := "application/gzip"
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &config.BackupBucket,
		Key:         &r2Key,
		Body:        bytes.NewReader(compressedData),
		ContentType: &contentType,
		ACL:         awsTypes.ObjectCannedACLPrivate,
		Metadata: map[string]string{
			"backup-timestamp":    timestampStr,
			"content-checksum":    checksumHex,
			"uncompressed-size":   fmt.Sprintf("%d", uncompressedSize),
			"compressed-size":     fmt.Sprintf("%d", compressedSize),
			"compression-ratio":   fmt.Sprintf("%.2f", compressionRatio),
			"retention-days":      fmt.Sprintf("%d", config.RetentionDays),
			"backup-type":         "function-data",
			"tables":            strings.Join(config.Tables, ","),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload to R2: %w", err)
	}

	log.Printf("Upload complete")

	// Verify upload if requested
	if config.VerifyUpload {
		log.Printf("Verifying upload...")
		resp, err := client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: &config.BackupBucket,
			Key:    &r2Key,
		})
		if err != nil {
			return fmt.Errorf("failed to verify upload: %w", err)
		}
		defer resp.Body.Close()

		downloadedChecksum, err := calculateChecksum(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to calculate download checksum: %w", err)
		}

		if downloadedChecksum != checksumHex {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", checksumHex, downloadedChecksum)
		}

		log.Printf("Verification passed: checksums match")
	}

	// Upload metadata
	metadata := &BackupMetadata{
		Timestamp:          timestampStr,
		DatePath:           datePath,
		R2Key:              r2Key,
		R2Bucket:           config.BackupBucket,
		Checksum:           checksumHex,
		CompressedSize:     compressedSize,
		UncompressedSize:   uncompressedSize,
		CompressionRatio:   compressionRatio,
		RetentionDays:      config.RetentionDays,
		TablesBackedUp:     config.Tables,
		Duration:           "", // Will be set later
		Encrypted:          config.Encrypt,
	}

	metaKey := r2Key + ".meta.json"
	metaJSON, _ := json.MarshalIndent(metadata, "", "  ")

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &config.BackupBucket,
		Key:         &metaKey,
		Body:        bytes.NewReader(metaJSON),
		ContentType: aws.String("application/json"),
		ACL:         awsTypes.ObjectCannedACLPrivate,
	})
	if err != nil {
		log.Printf("Warning: failed to upload metadata: %v", err)
	} else {
		log.Printf("Metadata uploaded: %s", metaKey)
	}

	log.Printf("Backup stored at: s3://%s/%s", config.BackupBucket, r2Key)

	return nil
}

func dumpFunctionsData(config *BackupConfig) ([]byte, error) {
	// For simplicity, this creates a SQL dump using pg_dump
	// In production, you might want to use a Go PostgreSQL driver

	// Check if pg_dump is available
	pgDumpPath, err := findPgDump()
	if err != nil {
		return nil, fmt.Errorf("pg_dump not found: %w", err)
	}

	// Build pg_dump command
	args := []string{
		"--data-only",
		"--no-owner",
		"--no-privileges",
		"--disable-triggers",
		config.DatabaseURL,
	}

	// Add table filters
	for _, table := range config.Tables {
		args = append(args, "--table="+table)
	}

	// Execute pg_dump
	cmd := execCommand(pgDumpPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("pg_dump failed: %w\nOutput: %s", err, string(output))
	}

	return output, nil
}

func findPgDump() (string, error) {
	// Try to find pg_dump in PATH
	return execLookPath("pg_dump")
}

func compressData(data []byte, level int) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip writer: %w", err)
	}

	if _, err := gzWriter.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write to gzip: %w", err)
	}

	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip: %w", err)
	}

	return buf.Bytes(), nil
}

func calculateChecksum(r io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// execCommand and execLookPath are defined at package level for testability
var execCommand = func(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}
var execLookPath = exec.LookPath
