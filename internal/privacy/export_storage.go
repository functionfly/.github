package privacy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ExportStorage defines the interface for storing GDPR export files
type ExportStorage interface {
	// Upload stores the export data and returns a download URL
	Upload(ctx context.Context, requestID uuid.UUID, data []byte, contentType string) (downloadURL string, err error)

	// GenerateDownloadURL creates a time-limited pre-signed download URL
	GenerateDownloadURL(ctx context.Context, requestID uuid.UUID, expiration time.Duration) (string, error)

	// Delete removes the export file after download or expiration
	Delete(ctx context.Context, requestID uuid.UUID) error

	// Exists checks if an export file exists
	Exists(ctx context.Context, requestID uuid.UUID) (bool, error)
}

// S3Storage implements ExportStorage using S3-compatible services (S3, R2, MinIO, etc.)
type S3Storage struct {
	client     *s3.Client
	bucket     string
	baseURL    string
	pathPrefix string
}

// S3Config holds configuration for S3-compatible storage
type S3Config struct {
	Bucket          string
	Region          string
	Endpoint        string // For S3-compatible services like R2, MinIO
	AccessKeyID     string
	SecretAccessKey string
	BaseURL         string // Public URL prefix for downloads (e.g., CDN or custom domain)
	PathPrefix      string // Optional prefix for all stored files
	UsePathStyle    bool   // Use path-style addressing (needed for MinIO)
}

// DefaultS3Config returns S3 config from environment variables
func DefaultS3Config() *S3Config {
	return &S3Config{
		Bucket:          getEnvOrDefault("PRIVACY_EXPORT_BUCKET", "privacy-exports"),
		Region:          getEnvOrDefault("PRIVACY_EXPORT_REGION", "us-east-1"),
		Endpoint:        os.Getenv("PRIVACY_EXPORT_ENDPOINT"), // For R2/MinIO: https://<account>.r2.cloudflarestorage.com
		AccessKeyID:     os.Getenv("PRIVACY_EXPORT_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("PRIVACY_EXPORT_SECRET_ACCESS_KEY"),
		BaseURL:         os.Getenv("PRIVACY_EXPORT_BASE_URL"), // e.g., https://exports.functionfly.io
		PathPrefix:      getEnvOrDefault("PRIVACY_EXPORT_PATH_PREFIX", "gdpr-exports"),
		UsePathStyle:    os.Getenv("PRIVACY_EXPORT_PATH_STYLE") == "true",
	}
}

// NewS3Storage creates a new S3-compatible storage client
func NewS3Storage(cfg *S3Config) (*S3Storage, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("bucket name is required")
	}

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client options
	var clientOpts []func(*s3.Options)

	// Use static credentials if provided
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		staticCreds := credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.Credentials = staticCreds
		})
	}

	// Use custom endpoint for S3-compatible services (R2, MinIO)
	if cfg.Endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = cfg.UsePathStyle
		})
	}

	client := s3.NewFromConfig(awsCfg, clientOpts...)

	storage := &S3Storage{
		client:     client,
		bucket:     cfg.Bucket,
		baseURL:      cfg.BaseURL,
		pathPrefix: cfg.PathPrefix,
	}

	// Verify bucket exists and we have access
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(cfg.Bucket),
	})
	if err != nil {
		logrus.WithError(err).Warn("Failed to verify S3 bucket access, will retry on first upload")
	}

	return storage, nil
}

// Upload stores the export data to S3 and returns a download URL
func (s *S3Storage) Upload(ctx context.Context, requestID uuid.UUID, data []byte, contentType string) (string, error) {
	if contentType == "" {
		contentType = "application/zip"
	}

	key := s.objectKey(requestID)

	// Upload to S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"request-id":    requestID.String(),
			"uploaded-at":   time.Now().UTC().Format(time.RFC3339),
			"content-hash":  s.hashData(data),
			"privacy-purpose": "gdpr-export",
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload export: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"request_id": requestID,
		"bucket":     s.bucket,
		"key":        key,
		"size":       len(data),
	}).Info("GDPR export uploaded to S3")

	// Return the download URL
	if s.baseURL != "" {
		return fmt.Sprintf("%s/%s", s.baseURL, key), nil
	}

	// Generate pre-signed URL as fallback
	return s.GenerateDownloadURL(ctx, requestID, 7*24*time.Hour)
}

// GenerateDownloadURL creates a time-limited pre-signed download URL
func (s *S3Storage) GenerateDownloadURL(ctx context.Context, requestID uuid.UUID, expiration time.Duration) (string, error) {
	key := s.objectKey(requestID)

	presignClient := s3.NewPresignClient(s.client)

	presignReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))

	if err != nil {
		return "", fmt.Errorf("failed to generate pre-signed URL: %w", err)
	}

	return presignReq.URL, nil
}

// Delete removes the export file
func (s *S3Storage) Delete(ctx context.Context, requestID uuid.UUID) error {
	key := s.objectKey(requestID)

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete export: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"request_id": requestID,
		"key":        key,
	}).Info("GDPR export deleted from S3")

	return nil
}

// Exists checks if an export file exists
func (s *S3Storage) Exists(ctx context.Context, requestID uuid.UUID) (bool, error) {
	key := s.objectKey(requestID)

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		// Check if it's a "not found" error
		var ae smithy.APIError
		if errors.As(err, &ae) {
			if ae.ErrorCode() == "NoSuchKey" || ae.ErrorCode() == "NotFound" {
				return false, nil
			}
		}
		return false, err
	}

	return true, nil
}

// objectKey generates the S3 object key for a request
func (s *S3Storage) objectKey(requestID uuid.UUID) string {
	if s.pathPrefix != "" {
		return fmt.Sprintf("%s/%s/%s.zip", s.pathPrefix, requestID.String()[:8], requestID)
	}
	return fmt.Sprintf("%s/%s.zip", requestID.String()[:8], requestID)
}

// hashData calculates SHA-256 hash of data for integrity verification
func (s *S3Storage) hashData(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// LocalStorage implements ExportStorage using local filesystem (for development)
type LocalStorage struct {
	basePath   string
	baseURL    string
	pathPrefix string
}

// NewLocalStorage creates a new local filesystem storage
func NewLocalStorage(basePath, baseURL string) (*LocalStorage, error) {
	if basePath == "" {
		basePath = "./exports"
	}

	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0750); err != nil {
		return nil, fmt.Errorf("failed to create exports directory: %w", err)
	}

	return &LocalStorage{
		basePath:   basePath,
		baseURL:      baseURL,
		pathPrefix: "gdpr-exports",
	}, nil
}

// Upload stores the export data locally
func (s *LocalStorage) Upload(ctx context.Context, requestID uuid.UUID, data []byte, contentType string) (string, error) {
	dir := filepath.Join(s.basePath, s.pathPrefix, requestID.String()[:8])
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	path := filepath.Join(dir, fmt.Sprintf("%s.zip", requestID))
	if err := os.WriteFile(path, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write export file: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"request_id": requestID,
		"path":       path,
		"size":       len(data),
	}).Info("GDPR export saved locally")

	if s.baseURL != "" {
		return fmt.Sprintf("%s/%s/%s/%s.zip", s.baseURL, s.pathPrefix, requestID.String()[:8], requestID), nil
	}

	return path, nil
}

// GenerateDownloadURL returns the file path (for local storage, no pre-signing needed)
func (s *LocalStorage) GenerateDownloadURL(ctx context.Context, requestID uuid.UUID, expiration time.Duration) (string, error) {
	path := filepath.Join(s.basePath, s.pathPrefix, requestID.String()[:8], fmt.Sprintf("%s.zip", requestID))

	if s.baseURL != "" {
		return fmt.Sprintf("%s/%s/%s/%s.zip", s.baseURL, s.pathPrefix, requestID.String()[:8], requestID), nil
	}

	return path, nil
}

// Delete removes the local export file
func (s *LocalStorage) Delete(ctx context.Context, requestID uuid.UUID) error {
	path := filepath.Join(s.basePath, s.pathPrefix, requestID.String()[:8], fmt.Sprintf("%s.zip", requestID))

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete export: %w", err)
	}

	// Try to remove empty parent directory
	dir := filepath.Join(s.basePath, s.pathPrefix, requestID.String()[:8])
	os.Remove(dir) // Ignore error - may not be empty

	logrus.WithFields(logrus.Fields{
		"request_id": requestID,
		"path":       path,
	}).Info("Local GDPR export deleted")

	return nil
}

// Exists checks if the local export file exists
func (s *LocalStorage) Exists(ctx context.Context, requestID uuid.UUID) (bool, error) {
	path := filepath.Join(s.basePath, s.pathPrefix, requestID.String()[:8], fmt.Sprintf("%s.zip", requestID))
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// NewExportStorage creates the appropriate storage backend based on configuration
func NewExportStorage() (ExportStorage, error) {
	// Check if S3 credentials are configured
	s3Cfg := DefaultS3Config()

	// Fall back to standard AWS env vars if privacy-specific ones aren't set
	if s3Cfg.AccessKeyID == "" {
		s3Cfg.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if s3Cfg.SecretAccessKey == "" {
		s3Cfg.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}

	if s3Cfg.AccessKeyID != "" && s3Cfg.SecretAccessKey != "" {
		logrus.Info("Initializing S3-compatible storage for GDPR exports")
		return NewS3Storage(s3Cfg)
	}

	// Fall back to local storage for development
	basePath := getEnvOrDefault("PRIVACY_EXPORT_LOCAL_PATH", "./exports")
	baseURL := os.Getenv("PRIVACY_EXPORT_LOCAL_URL") // e.g., http://localhost:8080/exports

	logrus.WithField("path", basePath).Info("Using local filesystem storage for GDPR exports (set S3 env vars for production)")
	return NewLocalStorage(basePath, baseURL)
}
