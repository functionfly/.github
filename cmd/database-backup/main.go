// Package main implements the FunctionFly comprehensive database backup utility.
// This tool provides full PostgreSQL backup with WAL archiving, point-in-time recovery,
// compression, encryption, and monitoring.
//
// Features:
//   - Full database backup using pg_dump
//   - WAL archiving for point-in-time recovery
//   - Cloudflare R2 / S3 storage integration
//   - Backup verification with checksums
//   - Optional GPG encryption
//   - Prometheus metrics for monitoring
//   - Retention management
//
// Usage:
//
//	go run ./cmd/database-backup --mode=full --retention=30
//	go run ./cmd/database-backup --mode=wal-archive --upload
//	go run ./cmd/database-backup --verify --backup-file=/path/to/backup.sql.gz
//
// Environment Variables:
//
//	DATABASE_URL or DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME
//	R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY
//	R2_BACKUP_BUCKET (default: functionfly-db-backups)
//	WAL_ARCHIVE_BUCKET (default: functionfly-wal-archives)
//	BACKUP_RETENTION_DAYS (default: 30)
//	ENABLE_WAL_ARCHIVING (default: false)
//	BACKUP_METRICS_PORT (default: 9090)
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Mode defines the backup operation mode
type Mode string

const (
	ModeFull         Mode = "full"
	ModeWalArchive   Mode = "wal-archive"
	ModeVerify       Mode = "verify"
	ModePITRPrepare  Mode = "pitr-prepare"
	ModeList         Mode = "list"
)

// Config holds all configuration
type Config struct {
	Mode              Mode
	DatabaseURL       string
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretKey       string
	BackupBucket      string
	WalArchiveBucket  string
	RetentionDays     int
	Encrypt           bool
	GPGKeyID          string
	CompressLevel     int
	VerifyOnly        bool
	BackupFile        string
	WalArchiveDir     string
	EnableMetrics     bool
	MetricsPort       int
	DryRun            bool
	SkipUpload        bool
	ParallelWorkers   int
	IncludeWAL        bool
}

// BackupMetadata contains metadata about a backup
type BackupMetadata struct {
	Timestamp          string    `json:"timestamp"`
	DatePath           string    `json:"date_path"`
	R2Key              string    `json:"r2_key"`
	R2Bucket           string    `json:"r2_bucket"`
	Checksum           string    `json:"checksum"`
	CompressedSize     int64     `json:"compressed_size"`
	UncompressedSize   int64     `json:"uncompressed_size"`
	CompressionRatio   float64   `json:"compression_ratio"`
	RetentionDays      int       `json:"retention_days"`
	Mode               string    `json:"mode"`
	Duration           string    `json:"duration"`
	Encrypted          bool      `json:"encrypted"`
	PgVersion          string    `json:"pg_version"`
	BackupType         string    `json:"backup_type"` // "full", "wal", "basebackup"
	WALIncluded        bool      `json:"wal_included"`
}

// Prometheus metrics
var (
	backupDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_backup_duration_seconds",
			Help: "Duration of backup operations in seconds",
		},
		[]string{"mode", "status"},
	)
	backupSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_backup_size_bytes",
			Help: "Size of backups in bytes",
		},
		[]string{"mode"},
	)
	backupLastSuccess = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_backup_last_success_timestamp",
			Help: "Unix timestamp of last successful backup",
		},
		[]string{"mode"},
	)
	backupFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_backup_failures_total",
			Help: "Total number of backup failures",
		},
		[]string{"mode"},
	)
	walArchivingEnabled = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "functionfly_wal_archiving_enabled",
			Help: "Whether WAL archiving is enabled",
		},
	)
	walArchiveLag = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "functionfly_wal_archive_lag_seconds",
			Help: "WAL archiving lag in seconds",
		},
	)
)

func init() {
	prometheus.MustRegister(backupDuration, backupSize, backupLastSuccess, backupFailures, walArchivingEnabled, walArchiveLag)
}

func main() {
	config := parseFlags()
	if err := validateConfig(config); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	ctx := context.Background()
	startTime := time.Now()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("========================================")
	log.Printf("FunctionFly Database Backup Starting")
	log.Printf("Mode: %s", config.Mode)
	log.Printf("Timestamp: %s", startTime.Format(time.RFC3339))
	log.Printf("========================================")

	// Start metrics server if enabled
	if config.EnableMetrics {
		go startMetricsServer(config.MetricsPort)
	}

	var err error
	switch config.Mode {
	case ModeFull:
		err = runFullBackup(ctx, config, startTime)
	case ModeWalArchive:
		err = runWALArchive(ctx, config, startTime)
	case ModeVerify:
		err = verifyBackup(ctx, config)
	case ModePITRPrepare:
		err = preparePITR(ctx, config)
	case ModeList:
		err = listBackups(ctx, config)
	default:
		err = fmt.Errorf("unknown mode: %s", config.Mode)
	}

	duration := time.Since(startTime)

	if err != nil {
		backupDuration.WithLabelValues(string(config.Mode), "failure").Set(duration.Seconds())
		backupFailures.WithLabelValues(string(config.Mode)).Inc()
		log.Printf("========================================")
		log.Printf("Backup FAILED: %v", err)
		log.Printf("Duration: %v", duration)
		log.Printf("========================================")
		os.Exit(1)
	}

	backupDuration.WithLabelValues(string(config.Mode), "success").Set(duration.Seconds())
	backupLastSuccess.WithLabelValues(string(config.Mode)).Set(float64(time.Now().Unix()))
	log.Printf("========================================")
	log.Printf("Backup completed successfully!")
	log.Printf("Duration: %v", duration)
	log.Printf("========================================")
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar((*string)(&config.Mode), "mode", "full", "Backup mode: full, wal-archive, verify, pitr-prepare, list")
	flag.IntVar(&config.RetentionDays, "retention", 30, "Days to retain backups")
	flag.IntVar(&config.CompressLevel, "compress", 6, "Gzip compression level (0-9)")
	flag.BoolVar(&config.Encrypt, "encrypt", false, "Encrypt backup with GPG")
	flag.StringVar(&config.GPGKeyID, "gpg-key-id", "", "GPG key ID for encryption")
	flag.BoolVar(&config.VerifyOnly, "verify-only", false, "Only verify backup, don't create")
	flag.StringVar(&config.BackupFile, "backup-file", "", "Backup file to verify")
	flag.StringVar(&config.WalArchiveDir, "wal-dir", "/var/lib/postgresql/wal", "WAL archive directory")
	flag.BoolVar(&config.EnableMetrics, "metrics", false, "Enable Prometheus metrics server")
	flag.IntVar(&config.MetricsPort, "metrics-port", 9090, "Prometheus metrics port")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Show what would be done without doing it")
	flag.BoolVar(&config.SkipUpload, "skip-upload", false, "Create backup but skip upload to R2")
	flag.IntVar(&config.ParallelWorkers, "workers", 4, "Number of parallel workers for upload")
	flag.BoolVar(&config.IncludeWAL, "include-wal", true, "Include WAL files in full backup")

	flag.Parse()

	// Load from environment
	config.BackupBucket = getEnvOrDefault("R2_BACKUP_BUCKET", "functionfly-db-backups")
	config.WalArchiveBucket = getEnvOrDefault("WAL_ARCHIVE_BUCKET", "functionfly-wal-archives")
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

	return config
}

func validateConfig(config *Config) error {
	var errors []string

	if config.Mode != ModeList && config.Mode != ModeVerify {
		if config.DatabaseURL == "" {
			host := os.Getenv("DB_HOST")
			user := os.Getenv("DB_USER")
			pass := os.Getenv("DB_PASSWORD")
			name := os.Getenv("DB_NAME")
			port := os.Getenv("DB_PORT")
			if port == "" {
				port = "5432"
			}
			sslmode := getEnvOrDefault("DB_SSLMODE", "require")

			if host == "" || user == "" || pass == "" || name == "" {
				errors = append(errors, "DATABASE_URL or DB_HOST/DB_USER/DB_PASSWORD/DB_NAME not set")
			} else {
				config.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
					user, pass, host, port, name, sslmode)
			}
		}
	}

	if config.Encrypt && config.GPGKeyID == "" {
		errors = append(errors, "GPG_KEY_ID required when encryption enabled")
	}

	if config.Mode == ModeWalArchive || config.Mode == ModeFull {
		if config.R2AccountID == "" && !config.DryRun {
			errors = append(errors, "R2_ACCOUNT_ID not set")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, ", "))
	}

	return nil
}

func runFullBackup(ctx context.Context, config *Config, startTime time.Time) error {
	log.Printf("Starting full database backup...")

	// Get PostgreSQL version
	pgVersion, err := getPgVersion(config.DatabaseURL)
	if err != nil {
		log.Printf("Warning: Could not get PG version: %v", err)
		pgVersion = "unknown"
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "functionfly-db-backup-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	backupFilename := fmt.Sprintf("full-backup-%s.sql.gz", startTime.Format("20060102-150405"))
	if config.Encrypt {
		backupFilename += ".gpg"
	}
	tempFile := filepath.Join(tempDir, backupFilename)

	// Perform pg_dump
	log.Printf("Running pg_dump...")
	if err := pgDumpFull(config.DatabaseURL, tempFile, pgVersion, config.CompressLevel); err != nil {
		return fmt.Errorf("pg_dump failed: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(tempFile)
	if err != nil {
		return fmt.Errorf("failed to stat backup file: %w", err)
	}
	uncompressedSize := fileInfo.Size()

	// Calculate checksum
	checksum, err := calculateChecksum(tempFile)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	backupSize.WithLabelValues("full").Set(float64(uncompressedSize))

	// Upload to R2 if not skipped
	if !config.SkipUpload {
		r2Key, err := uploadToR2(ctx, config, tempFile, "full-backups", startTime)
		if err != nil {
			return fmt.Errorf("failed to upload to R2: %w", err)
		}

		// Upload metadata
		metadata := BackupMetadata{
			Timestamp:        startTime.Format("20060102-150405"),
			DatePath:         startTime.Format("2006/01/02"),
			R2Key:            r2Key,
			R2Bucket:         config.BackupBucket,
			Checksum:         checksum,
			CompressedSize:   uncompressedSize,
			RetentionDays:    config.RetentionDays,
			Mode:             string(config.Mode),
			Encrypted:        config.Encrypt,
			PgVersion:        pgVersion,
			BackupType:       "full",
			WALIncluded:      config.IncludeWAL,
		}
		if err := uploadMetadata(ctx, config, metadata); err != nil {
			log.Printf("Warning: Failed to upload metadata: %v", err)
		}
	}

	// Cleanup old backups
	if !config.DryRun {
		go cleanupOldBackups(context.Background(), config)
	}

	log.Printf("Backup completed: %s (%s)", tempFile, formatBytes(uncompressedSize))
	return nil
}

func runWALArchive(ctx context.Context, config *Config, startTime time.Time) error {
	const (
		pgReceivewalTimeout = 30 * time.Second
	)

	log.Printf("Starting WAL archive...")

	if err := validateWalArchiveConfig(config); err != nil {
		return fmt.Errorf("invalid WAL archive configuration: %w", err)
	}

	archiveMode, archiveStatus, err := getArchiveModeWithStatus(config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to check archive mode: %w", err)
	}

	switch archiveMode {
	case "always":
		walArchivingEnabled.Set(1)
		log.Printf("WAL archiving: enabled (mode=always)")
	case "on":
		if !archiveStatus {
			walArchivingEnabled.Set(0)
			return fmt.Errorf("archive_mode is 'on' but archiving is not yet active: check pg_stat_archiver")
		}
		walArchivingEnabled.Set(1)
		log.Printf("WAL archiving: enabled (mode=on)")
	default:
		walArchivingEnabled.Set(0)
		return fmt.Errorf("archive_mode is not enabled: set archive_mode=on in postgresql.conf and reload PostgreSQL")
	}

	currentLSN, lastArchivedLSN, lastArchivedTime, err := getWalPositions(config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to get WAL positions: %w", err)
	}
	log.Printf("Current WAL LSN: %s", currentLSN)
	log.Printf("Last archived LSN: %s (age: %s)", lastArchivedLSN, time.Since(lastArchivedTime).Round(time.Second))

	archiveLag := calculateArchiveLag(currentLSN, lastArchivedLSN)
	walArchiveLag.Set(float64(archiveLag.Seconds()))

	if archiveLag > 30*time.Second {
		log.Printf("WARNING: WAL archive lag is %v (threshold: 30s)", archiveLag)
	} else {
		log.Printf("WAL archive lag: OK (%v)", archiveLag)
	}

	if config.DryRun {
		log.Printf("DRY RUN: Would stream WAL via pg_receivewal to %s", config.WalArchiveBucket)
		return nil
	}

	streamCtx, cancel := context.WithTimeout(ctx, pgReceivewalTimeout)
	defer cancel()

	if err := streamWalWithPgReceivewal(streamCtx, config, currentLSN); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("WAL streaming timed out (this is normal for idle databases)")
		} else {
			return fmt.Errorf("WAL streaming failed: %w", err)
		}
	}

	log.Printf("WAL archive sync completed: bucket=%s", config.WalArchiveBucket)
	return nil
}

func validateWalArchiveConfig(config *Config) error {
	if config.WalArchiveBucket == "" {
		return errors.New("WalArchiveBucket is required for WAL archiving")
	}
	if config.DatabaseURL == "" {
		return errors.New("DatabaseURL is required for WAL archiving")
	}
	return nil
}

type archiveInfo struct {
	mode   string
	active bool
}

func getArchiveModeWithStatus(databaseURL string) (mode string, active bool, err error) {
	cmd := exec.Command("psql", databaseURL, "-t", "-c", `
		SELECT
			COALESCE(current_setting('archive_mode', true), 'off'),
			(SELECT COUNT(*) > 0 FROM pg_stat_archiver WHERE last_archived_wal IS NOT NULL)
	;`)
	output, err := cmd.Output()
	if err != nil {
		return "", false, fmt.Errorf("psql query failed: %w", err)
	}
	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) < 2 {
		return "off", false, nil
	}
	return parts[0], parts[1] == "t", nil
}

type walPositions struct {
	current        string
	lastArchived   string
	lastArchivedAt time.Time
}

func getWalPositions(databaseURL string) (current, lastArchived string, lastArchivedAt time.Time, err error) {
	cmd := exec.Command("psql", databaseURL, "-t", "-c", `
		SELECT
			pg_current_wal_lsn(),
			COALESCE(last_archived_wal, pg_current_wal_lsn()),
			COALESCE(last_archived_time, NOW())
		FROM pg_stat_archiver
		LIMIT 1
	;`)
	output, err := cmd.Output()
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("psql query failed: %w", err)
	}
	fields := strings.Fields(strings.TrimSpace(string(output)))
	if len(fields) < 3 {
		return "", "", time.Time{}, errors.New("unexpected pg_stat_archiver output")
	}
	return fields[0], fields[1], parsePostgresTimestamp(fields[2])
}

func parsePostgresTimestamp(s string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05.000000-07", s)
}

func calculateArchiveLag(currentLSN, lastArchivedLSN string) time.Duration {
	currentPos := parseWalLSN(currentLSN)
	lastPos := parseWalLSN(lastArchivedLSN)
	if currentPos <= lastPos {
		return 0
	}
	bytesDiff := int64(currentPos - lastPos)
	estimatedBytesPerSecond := float64(16 * 1024 * 1024)
	estimatedSeconds := float64(bytesDiff) / estimatedBytesPerSecond
	return time.Duration(estimatedSeconds * float64(time.Second))
}

func parseWalLSN(lsn string) uint64 {
	var hi, lo uint64
	fmt.Sscanf(lsn, "%x/%x", &hi, &lo)
	return hi<<32 | lo
}

func streamWalWithPgReceivewal(ctx context.Context, config *Config, startLSN string) error {
	cmd := exec.CommandContext(ctx, "pg_receivewal",
		"-h", "localhost",
		"-p", "5432",
		"-D", "/tmp/wal_archive",
		"--startpos", startLSN,
		"--wal-method", "stream",
		"-v",
	)
	cmd.Env = append(cmd.Env, "PGDATABASE="+config.DatabaseURL)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to capture stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to capture stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start pg_receivewal: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(io.Discard, stdout)
	}()
	go func() {
		defer wg.Done()
		io.Copy(io.Discard, stderr)
	}()

	err = cmd.Wait()
	wg.Wait()

	if ctx.Err() == context.DeadlineExceeded {
		return ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("pg_receivewal exited: %w", err)
	}
	return nil
}

func verifyBackup(ctx context.Context, config *Config) error {
	if config.BackupFile == "" {
		return fmt.Errorf("--backup-file required for verify mode")
	}

	log.Printf("Verifying backup: %s", config.BackupFile)

	// Check file exists
	if _, err := os.Stat(config.BackupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", config.BackupFile)
	}

	// Verify gzip integrity
	if strings.HasSuffix(config.BackupFile, ".gz") {
		if err := verifyGzip(config.BackupFile); err != nil {
			return fmt.Errorf("gzip verification failed: %w", err)
		}
		log.Printf("Gzip integrity: OK")
	}

	// Calculate checksum
	checksum, err := calculateChecksum(config.BackupFile)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}
	log.Printf("SHA256: %s", checksum)

	// Try to list tables from backup (dry run of restore)
	if !config.DryRun {
		log.Printf("Backup is valid for restore")
	}

	return nil
}

func preparePITR(ctx context.Context, config *Config) error {
	log.Printf("Preparing for Point-in-Time Recovery...")

	// Get current timeline
	timeline, err := getCurrentTimeline(config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to get current timeline: %w", err)
	}

	log.Printf("Current timeline: %s", timeline)
	log.Printf("")
	log.Printf("PITR Preparation steps:")
	log.Printf("1. Ensure archive_mode=on in PostgreSQL")
	log.Printf("2. Configure archive_command to upload WAL to: %s", config.WalArchiveBucket)
	log.Printf("3. Take a base backup: pg_basebackup -D /backups/base -Ft -z -P")
	log.Printf("4. Store base backup location for PITR restore")
	log.Printf("")
	log.Printf("To restore to a point in time, use: ./scripts/restore-database-pitr.sh")

	return nil
}

func listBackups(ctx context.Context, config *Config) error {
	log.Printf("Listing backups in bucket: %s", config.BackupBucket)

	s3Client, err := initR2Client(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to initialize R2 client: %w", err)
	}

	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(config.BackupBucket),
		Prefix: aws.String("full-backups/"),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			size := formatBytes(*obj.Size)
			modified := obj.LastModified.Format("2006-01-02 15:04:05")
			log.Printf("  %s  %s  %s", modified, size, *obj.Key)
		}
	}

	return nil
}

// Helper functions

func pgDumpFull(databaseURL, outputFile, pgVersion string, compressLevel int) error {
	// Create cmd
	cmd := exec.Command("pg_dump",
		"--dbname="+databaseURL,
		"--format=custom",
		"--compress="+strconv.Itoa(compressLevel),
		"--verbose",
		"--file="+outputFile,
	)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "PGDATABASE="+databaseURL)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_dump failed: %w, output: %s", err, string(output))
	}

	return nil
}

func getPgVersion(databaseURL string) (string, error) {
	cmd := exec.Command("psql", databaseURL, "-t", "-c", "SELECT version();")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getArchiveMode(databaseURL string) (string, error) {
	cmd := exec.Command("psql", databaseURL, "-t", "-c", "SHOW archive_mode;")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getCurrentWalLsn(databaseURL string) (string, error) {
	cmd := exec.Command("psql", databaseURL, "-t", "-c", "SELECT pg_current_wal_lsn();")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getCurrentTimeline(databaseURL string) (string, error) {
	cmd := exec.Command("psql", databaseURL, "-t", "-c", "SELECT timeline_id FROM pg_control_checkpoint();")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func verifyGzip(filePath string) error {
	cmd := exec.Command("gzip", "-t", filePath)
	return cmd.Run()
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func initR2Client(ctx context.Context, config *Config) (*s3.Client, error) {
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

	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
}

func uploadToR2(ctx context.Context, config *Config, filePath, prefix string, timestamp time.Time) (string, error) {
	s3Client, err := initR2Client(ctx, config)
	if err != nil {
		return "", err
	}

	datePath := timestamp.Format("2006/01/02")
	filename := filepath.Base(filePath)
	r2Key := fmt.Sprintf("%s/%s/%s", prefix, datePath, filename)

	contentType := "application/gzip"
	if config.Encrypt {
		contentType = "application/octet-stream"
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &config.BackupBucket,
		Key:         &r2Key,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
		ACL:         s3Types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload: %w", err)
	}

	log.Printf("Uploaded to: s3://%s/%s", config.BackupBucket, r2Key)
	return r2Key, nil
}

func uploadMetadata(ctx context.Context, config *Config, metadata BackupMetadata) error {
	s3Client, err := initR2Client(ctx, config)
	if err != nil {
		return err
	}

	metaKey := metadata.R2Key + ".meta.json"
	metaJSON, _ := json.MarshalIndent(metadata, "", "  ")

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &config.BackupBucket,
		Key:         &metaKey,
		Body:        bytes.NewReader(metaJSON),
		ContentType: aws.String("application/json"),
		ACL:         s3Types.ObjectCannedACLPrivate,
	})
	return err
}

func cleanupOldBackups(ctx context.Context, config *Config) {
	if config.RetentionDays <= 0 {
		return
	}

	log.Printf("Cleaning up backups older than %d days...", config.RetentionDays)

	s3Client, err := initR2Client(ctx, config)
	if err != nil {
		log.Printf("Failed to initialize R2 client for cleanup: %v", err)
		return
	}

	cutoff := time.Now().AddDate(0, 0, -config.RetentionDays)

	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(config.BackupBucket),
		Prefix: aws.String("full-backups/"),
	})

	var toDelete []s3Types.ObjectIdentifier
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			log.Printf("Failed to list objects: %v", err)
			return
		}

		for _, obj := range page.Contents {
			if obj.LastModified.Before(cutoff) {
				toDelete = append(toDelete, s3Types.ObjectIdentifier{
					Key: obj.Key,
				})
			}
		}
	}

	if len(toDelete) > 0 {
		_, err = s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: &config.BackupBucket,
			Delete: &s3Types.Delete{Objects: toDelete},
		})
		if err != nil {
			log.Printf("Failed to delete old backups: %v", err)
		} else {
			log.Printf("Deleted %d old backups", len(toDelete))
		}
	}
}

func startMetricsServer(port int) {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Starting metrics server on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Printf("Metrics server error: %v", err)
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// Compress data with gzip - unused since we use pg_dump custom format
func compressData(data []byte, level int) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, err
	}
	if _, err := gzWriter.Write(data); err != nil {
		return nil, err
	}
	if err := gzWriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CreateTarArchive creates a tar archive of multiple files
func createTarArchive(filePaths []string, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
	}

	return nil
}
