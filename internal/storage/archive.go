package storage

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

// ArchiveMetadata represents metadata for an archived data batch
type ArchiveMetadata struct {
	ID              uuid.UUID              `json:"id"`
	ArchiveType     string                 `json:"archive_type"`     // e.g., "audit_logs", "user_data"
	RecordCount     int                    `json:"record_count"`
	DateRange       ArchiveDateRange       `json:"date_range"`
	StorageKey      string                 `json:"storage_key"`      // S3 key or equivalent
	CompressedSize  int64                  `json:"compressed_size"`
	OriginalSize    int64                  `json:"original_size"`
	CompressionRatio float64               `json:"compression_ratio"`
	EncryptionKeyID string                 `json:"encryption_key_id,omitempty"`
	Checksum        string                 `json:"checksum"`         // SHA256 of archived data
	CreatedAt       time.Time              `json:"created_at"`
	Status          string                 `json:"status"`           // "completed", "failed", "pending"
	ErrorMessage    string                 `json:"error_message,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ArchiveDateRange represents the date range of archived records
type ArchiveDateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ArchiveStorage defines the interface for archive storage operations
type ArchiveStorage interface {
	// StoreArchive stores compressed and encrypted archive data
	StoreArchive(ctx context.Context, key string, data io.Reader, metadata *ArchiveMetadata) error

	// RetrieveArchive retrieves archive data by key
	RetrieveArchive(ctx context.Context, key string) (io.ReadCloser, error)

	// DeleteArchive deletes an archive by key
	DeleteArchive(ctx context.Context, key string) error

	// ListArchives lists archives with optional prefix filtering
	ListArchives(ctx context.Context, prefix string) ([]*ArchiveMetadata, error)

	// GetArchiveMetadata retrieves metadata for a specific archive
	GetArchiveMetadata(ctx context.Context, key string) (*ArchiveMetadata, error)

	// Health check for storage backend
	HealthCheck(ctx context.Context) error
}