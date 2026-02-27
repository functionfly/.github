package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StorageService handles file uploads to local filesystem
// TODO: Replace with cloud storage service (S3, Cloudflare R2, etc.) for production
type StorageService struct {
	baseDir string
	baseURL string
}

// NewStorageService creates a new local filesystem storage service
func NewStorageService(baseURL string) *StorageService {
	baseDir := "./uploads" // Local uploads directory

	// Create uploads directory if it doesn't exist
	os.MkdirAll(baseDir, 0755)

	return &StorageService{
		baseDir: baseDir,
		baseURL: baseURL,
	}
}

// UploadFile uploads a file to local filesystem storage
func (s *StorageService) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, path string) (string, error) {
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

// DeleteFile deletes a file from local filesystem storage
func (s *StorageService) DeleteFile(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.baseDir, path)
	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	return nil
}

// getPublicURL generates a public URL for a file
func (s *StorageService) getPublicURL(path string) string {
	// For local development, return a local URL
	// In production, this should be replaced with cloud storage URLs
	return fmt.Sprintf("%s/uploads/%s", strings.TrimSuffix(s.baseURL, "/"), path)
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