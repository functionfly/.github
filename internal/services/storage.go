package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// StorageBackend defines the storage backend type
type StorageBackend string

const (
	StorageBackendLocal StorageBackend = "local"
	StorageBackendS3    StorageBackend = "s3"
	StorageBackendR2    StorageBackend = "r2" // Cloudflare R2 (S3-compatible)
)

// StorageService handles file uploads to local filesystem or cloud storage
type StorageService struct {
	baseDir    string
	baseURL    string
	backend    StorageBackend
	s3Client   *s3.Client
	bucket     string
}

// NewStorageService creates a new storage service with the appropriate backend
func NewStorageService(baseURL string) *StorageService {
	// Determine storage backend from environment
	backend := StorageBackend(os.Getenv("STORAGE_BACKEND"))
	if backend == "" {
		backend = StorageBackendLocal
	}

	svc := &StorageService{
		baseDir: "./uploads",
		baseURL: baseURL,
		backend: backend,
		bucket:  os.Getenv("STORAGE_BUCKET"),
	}

	// Initialize S3/R2 client if needed
	if backend == StorageBackendS3 || backend == StorageBackendR2 {
		svc.initS3Client()
	} else {
		// Create uploads directory for local storage
		os.MkdirAll(svc.baseDir, 0755)
	}

	logrus.Infof("Storage service initialized with backend: %s", backend)
	return svc
}

// initS3Client initializes the S3 client for cloud storage
func (s *StorageService) initS3Client() {
	ctx := context.Background()

	// Get AWS configuration
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "auto"
	}

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(awsRegion),
	}

	// Use custom endpoint for R2
	if s.backend == StorageBackendR2 {
		r2Endpoint := os.Getenv("R2_ENDPOINT")
		r2AccountID := os.Getenv("R2_ACCOUNT_ID")
		if r2Endpoint != "" && r2AccountID != "" {
			customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               fmt.Sprintf("https://%s.r2.cloudflarestorage.com", r2AccountID),
					HostnameImmutable: true,
				}, nil
			})
			opts = append(opts, config.WithEndpointResolverWithOptions(customResolver))
		}
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		logrus.Warnf("Failed to load AWS config, using local storage: %v", err)
		s.backend = StorageBackendLocal
		os.MkdirAll(s.baseDir, 0755)
		return
	}

	// Get credentials
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if awsAccessKey != "" && awsSecretKey != "" {
		cfg.Credentials = credentials.NewStaticCredentialsProvider(awsAccessKey, awsSecretKey, "")
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if s.backend == StorageBackendR2 {
			o.UsePathStyle = true // R2 requires path-style
		}
	})

	// Verify bucket exists
	if s.bucket == "" {
		logrus.Warn("STORAGE_BUCKET not set, falling back to local storage")
		s.backend = StorageBackendLocal
		os.MkdirAll(s.baseDir, 0755)
		return
	}

	_, err = s3Client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: &s.bucket})
	if err != nil {
		logrus.Warnf("Bucket %s not accessible, falling back to local storage: %v", s.bucket, err)
		s.backend = StorageBackendLocal
		os.MkdirAll(s.baseDir, 0755)
		return
	}

	s.s3Client = s3Client
	logrus.Infof("S3/R2 client initialized for bucket: %s", s.bucket)
}

// UploadFile uploads a file to storage (local or cloud)
func (s *StorageService) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, path string) (string, error) {
	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	// Use cloud or local storage
	if s.backend == StorageBackendS3 || s.backend == StorageBackendR2 {
		return s.uploadToS3(ctx, path, content, fileHeader.Header.Get("Content-Type"))
	}
	return s.uploadToLocal(ctx, fileHeader, path)
}

// uploadToS3 uploads a file to S3/R2
func (s *StorageService) uploadToS3(ctx context.Context, path string, content []byte, contentType string) (string, error) {
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &path,
		Body:        bytes.NewReader(content),
		ContentType: &contentType,
		ACL:         awsTypes.ObjectCannedACLPrivate,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Generate public URL based on backend
	publicURL := s.getCloudURL(path)
	return publicURL, nil
}

// uploadToLocal uploads a file to local filesystem
func (s *StorageService) uploadToLocal(ctx context.Context, fileHeader *multipart.FileHeader, path string) (string, error) {
	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Ensure the directory exists
	fullPath := filepath.Join(s.baseDir, path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	// Generate public URL
	publicURL := s.getPublicURL(path)
	return publicURL, nil
}

// DeleteFile deletes a file from storage
func (s *StorageService) DeleteFile(ctx context.Context, path string) error {
	if s.backend == StorageBackendS3 || s.backend == StorageBackendR2 {
		_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: &s.bucket,
			Key:    &path,
		})
		if err != nil {
			return fmt.Errorf("failed to delete from S3: %w", err)
		}
		return nil
	}

	// Local storage
	fullPath := filepath.Join(s.baseDir, path)
	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}
	return nil
}

// GetFile retrieves a file from storage
func (s *StorageService) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	if s.backend == StorageBackendS3 || s.backend == StorageBackendR2 {
		resp, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: &s.bucket,
			Key:    &path,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get from S3: %w", err)
		}
		return resp.Body, nil
	}

	// Local storage
	fullPath := filepath.Join(s.baseDir, path)
	return os.Open(fullPath)
}

// getPublicURL generates a public URL for a local file
func (s *StorageService) getPublicURL(path string) string {
	return fmt.Sprintf("%s/uploads/%s", strings.TrimSuffix(s.baseURL, "/"), path)
}

// getCloudURL generates a public URL for a cloud-stored file
func (s *StorageService) getCloudURL(path string) string {
	if s.backend == StorageBackendR2 {
		r2PublicURL := os.Getenv("R2_PUBLIC_URL")
		if r2PublicURL != "" {
			return fmt.Sprintf("%s/%s", strings.TrimSuffix(r2PublicURL, "/"), path)
		}
		// R2 uses custom domain
		return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", os.Getenv("R2_ACCOUNT_ID"), s.bucket, path)
	}
	// S3
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, os.Getenv("AWS_REGION"), path)
}

// GenerateUniquePath generates a unique path for file storage
func (s *StorageService) GenerateUniquePath(prefix, filename string) string {
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	// Generate unique filename with timestamp and UUID
	timestamp := time.Now().Format("20060102_150405")
	uniqueID := uuid.New().String()[:8]

	return fmt.Sprintf("%s/%s_%s_%s%s", prefix, nameWithoutExt, timestamp, uniqueID, ext)
}
