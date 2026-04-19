package statefabric

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/functionfly/functionfly/internal/storage/state"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// R2StorageBackend provides Cloudflare R2 storage for state fabric data types
// including event logs, snapshots, memory blobs, and replay data.
type R2StorageBackend struct {
	client        *s3.Client
	accountID     string
	buckets       R2StorageBuckets
	primaryRegion string
	compression   bool
}

// R2StorageBuckets holds the configured bucket names for different data types
type R2StorageBuckets struct {
	Events    string // Event logs storage
	Snapshots string // Snapshot data storage
	Memory    string // Memory blob storage
	Replay    string // Replay data storage
	General   string // Fallback/general purpose storage
}

// R2StorageConfig holds configuration for R2 state fabric storage
type R2StorageConfig struct {
	AccountID     string
	AccessKeyID   string
	SecretKey     string
	Buckets       R2StorageBuckets
	PrimaryRegion string
	Compression   bool // Enable gzip compression for storage
}

// R2StoragePath prefixes for different data types
const (
	R2PathEvents    = "statefabric/events"
	R2PathSnapshots = "statefabric/snapshots"
	R2PathMemory    = "statefabric/memory"
	R2PathReplay    = "statefabric/replays"
)

// R2StorageObject represents an object stored in R2 with metadata
type R2StorageObject struct {
	Key         string    `json:"key"`
	Bucket      string    `json:"bucket"`
	Size        int64     `json:"size"`
	ContentHash string    `json:"content_hash"`
	ContentType string    `json:"content_type"`
	Compressed  bool      `json:"compressed"`
	TenantID    uuid.UUID `json:"tenant_id"`
	EntityID    uuid.UUID `json:"entity_id"`   // State/Fabric ID
	EntityType  string    `json:"entity_type"` // "event", "snapshot", "memory", "replay"
	CreatedAt   time.Time `json:"created_at"`
	Metadata    JSONMap   `json:"metadata"`
}

// NewR2StorageBackend creates a new R2 storage backend from environment variables
func NewR2StorageBackend() (*R2StorageBackend, error) {
	cfg := loadR2StorageConfigFromEnv()
	if cfg == nil {
		return nil, fmt.Errorf("R2 storage configuration not available: missing required environment variables")
	}
	return NewR2StorageBackendWithConfig(cfg)
}

// NewR2StorageBackendWithConfig creates a new R2 storage backend with explicit configuration
func NewR2StorageBackendWithConfig(cfg *R2StorageConfig) (*R2StorageBackend, error) {
	ctx := context.Background()

	// Build R2 endpoint URL
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	// Create custom endpoint resolver for R2
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			HostnameImmutable: true,
		}, nil
	})

	region := cfg.PrimaryRegion
	if region == "" {
		region = "auto"
	}

	// Load AWS config with R2-specific settings
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 AWS config: %w", err)
	}

	// Create S3 client with path-style addressing (required for R2)
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	backend := &R2StorageBackend{
		client:        client,
		accountID:     cfg.AccountID,
		buckets:       cfg.Buckets,
		primaryRegion: region,
		compression:   cfg.Compression,
	}

	// Verify buckets exist
	if err := backend.verifyBuckets(ctx); err != nil {
		return nil, err
	}

	logrus.Infof("R2StorageBackend initialized: events=%s, snapshots=%s, memory=%s, replay=%s",
		cfg.Buckets.Events, cfg.Buckets.Snapshots, cfg.Buckets.Memory, cfg.Buckets.Replay)

	return backend, nil
}

// loadR2StorageConfigFromEnv loads R2 storage configuration from environment variables
func loadR2StorageConfigFromEnv() *R2StorageConfig {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKeyID := os.Getenv("R2_ACCESS_KEY_ID")
	if accessKeyID == "" {
		accessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	if secretKey == "" {
		secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}

	// Check required values
	if accountID == "" || accessKeyID == "" || secretKey == "" {
		return nil
	}

	// Load bucket names - can be separate buckets or a single bucket with prefixes
	cfg := &R2StorageConfig{
		AccountID:     accountID,
		AccessKeyID:   accessKeyID,
		SecretKey:     secretKey,
		PrimaryRegion: os.Getenv("R2_STATEFABRIC_REGION"),
		Compression:   os.Getenv("R2_STATEFABRIC_COMPRESSION") != "false", // Default to true
		Buckets: R2StorageBuckets{
			Events:    os.Getenv("R2_STATEFABRIC_EVENTS_BUCKET"),
			Snapshots: os.Getenv("R2_STATEFABRIC_SNAPSHOTS_BUCKET"),
			Memory:    os.Getenv("R2_STATEFABRIC_MEMORY_BUCKET"),
			Replay:    os.Getenv("R2_STATEFABRIC_REPLAY_BUCKET"),
			General:   os.Getenv("R2_STATEFABRIC_BUCKET"), // Fallback
		},
	}

	// Use general bucket as fallback for any unspecified buckets
	if cfg.Buckets.General != "" {
		if cfg.Buckets.Events == "" {
			cfg.Buckets.Events = cfg.Buckets.General
		}
		if cfg.Buckets.Snapshots == "" {
			cfg.Buckets.Snapshots = cfg.Buckets.General
		}
		if cfg.Buckets.Memory == "" {
			cfg.Buckets.Memory = cfg.Buckets.General
		}
		if cfg.Buckets.Replay == "" {
			cfg.Buckets.Replay = cfg.Buckets.General
		}
	}

	// Ensure at least one bucket is configured
	if cfg.Buckets.Events == "" && cfg.Buckets.General == "" {
		return nil
	}

	return cfg
}

// verifyBuckets checks that all configured buckets exist and are accessible
func (r *R2StorageBackend) verifyBuckets(ctx context.Context) error {
	buckets := []string{
		r.buckets.Events,
		r.buckets.Snapshots,
		r.buckets.Memory,
		r.buckets.Replay,
	}

	for _, bucket := range buckets {
		if bucket == "" {
			continue
		}
		_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: &bucket,
		})
		if err != nil {
			return fmt.Errorf("R2 bucket %s not accessible: %w", bucket, err)
		}
	}

	return nil
}

// ==================== Event Log Storage ====================

// StoreEventLogs stores event logs to R2, typically as compressed JSON batches
func (r *R2StorageBackend) StoreEventLogs(ctx context.Context, tenantID, stateID uuid.UUID, events []state.StateEvent) (*R2StorageObject, error) {
	if len(events) == 0 {
		return nil, nil
	}

	// Serialize events
	eventData, err := json.Marshal(events)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal events: %w", err)
	}

	// Build key: statefabric/events/{tenant}/{state}/{timestamp}_{firstSeq}-{lastSeq}.json.gz
	firstSeq := events[0].SequenceNum
	lastSeq := events[len(events)-1].SequenceNum
	timestamp := time.Now().UTC().Format("20060102_150405")
	key := fmt.Sprintf("%s/%s/%s/%s/%s_%d-%d.json",
		R2PathEvents,
		tenantID.String(),
		stateID.String(),
		timestamp[:8],
		timestamp,
		firstSeq,
		lastSeq,
	)

	contentType := "application/json"
	data := eventData
	compressed := false

	// Compress if enabled
	if r.compression {
		data, err = compressData(eventData)
		if err == nil {
			key += ".gz"
			contentType = "application/gzip"
			compressed = true
		}
	}

	return r.storeObject(ctx, r.buckets.Events, key, data, contentType, tenantID, stateID, "event", compressed, JSONMap{
		"event_count":    len(events),
		"first_sequence": firstSeq,
		"last_sequence":  lastSeq,
	})
}

// GetEventLogs retrieves event logs from R2
func (r *R2StorageBackend) GetEventLogs(ctx context.Context, key string) ([]state.StateEvent, error) {
	data, err := r.getObject(ctx, r.buckets.Events, key)
	if err != nil {
		return nil, err
	}

	// Decompress if needed
	if strings.HasSuffix(key, ".gz") {
		data, err = decompressData(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress event logs: %w", err)
		}
	}

	var events []state.StateEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}

	return events, nil
}

// ListEventLogBatches lists stored event log batches for a state
func (r *R2StorageBackend) ListEventLogBatches(ctx context.Context, tenantID, stateID uuid.UUID) ([]R2StorageObject, error) {
	prefix := fmt.Sprintf("%s/%s/%s/", R2PathEvents, tenantID.String(), stateID.String())
	return r.listObjects(ctx, r.buckets.Events, prefix)
}

// ==================== Snapshot Storage ====================

// StoreSnapshotData stores snapshot data to R2
func (r *R2StorageBackend) StoreSnapshotData(ctx context.Context, tenantID, stateID, snapshotID uuid.UUID, data JSONMap, metadata JSONMap) (*R2StorageObject, error) {
	// Serialize snapshot data
	snapshotData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal snapshot data: %w", err)
	}

	// Build key: statefabric/snapshots/{tenant}/{state}/{snapshot}.json.gz
	key := fmt.Sprintf("%s/%s/%s/%s.json",
		R2PathSnapshots,
		tenantID.String(),
		stateID.String(),
		snapshotID.String(),
	)

	contentType := "application/json"
	storedData := snapshotData
	compressed := false

	// Compress if enabled (snapshots are usually large, so compression helps)
	if r.compression && len(snapshotData) > 1024 {
		storedData, err = compressData(snapshotData)
		if err == nil {
			key += ".gz"
			contentType = "application/gzip"
			compressed = true
		}
	}

	return r.storeObject(ctx, r.buckets.Snapshots, key, storedData, contentType, tenantID, snapshotID, "snapshot", compressed, metadata)
}

// GetSnapshotData retrieves snapshot data from R2
func (r *R2StorageBackend) GetSnapshotData(ctx context.Context, key string) (JSONMap, error) {
	data, err := r.getObject(ctx, r.buckets.Snapshots, key)
	if err != nil {
		return nil, err
	}

	// Decompress if needed
	if strings.HasSuffix(key, ".gz") {
		data, err = decompressData(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress snapshot: %w", err)
		}
	}

	var snapshotData JSONMap
	if err := json.Unmarshal(data, &snapshotData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot data: %w", err)
	}

	return snapshotData, nil
}

// DeleteSnapshotData removes snapshot data from R2
func (r *R2StorageBackend) DeleteSnapshotData(ctx context.Context, key string) error {
	return r.deleteObject(ctx, r.buckets.Snapshots, key)
}

// ==================== Memory Blob Storage ====================

// StoreMemoryBlob stores agent memory content to R2
func (r *R2StorageBackend) StoreMemoryBlob(ctx context.Context, tenantID, memoryID uuid.UUID, content []byte, memoryType string, metadata JSONMap) (*R2StorageObject, error) {
	// Build key: statefabric/memory/{tenant}/{type}/{memory_id}
	key := fmt.Sprintf("%s/%s/%s/%s.bin",
		R2PathMemory,
		tenantID.String(),
		memoryType,
		memoryID.String(),
	)

	// Memory content could be text or binary
	contentType := "application/octet-stream"
	if isTextContent(content) {
		contentType = "text/plain; charset=utf-8"
	}

	return r.storeObject(ctx, r.buckets.Memory, key, content, contentType, tenantID, memoryID, "memory", false, metadata)
}

// GetMemoryBlob retrieves memory content from R2
func (r *R2StorageBackend) GetMemoryBlob(ctx context.Context, key string) ([]byte, error) {
	return r.getObject(ctx, r.buckets.Memory, key)
}

// DeleteMemoryBlob removes memory content from R2
func (r *R2StorageBackend) DeleteMemoryBlob(ctx context.Context, key string) error {
	return r.deleteObject(ctx, r.buckets.Memory, key)
}

// ListMemoryBlobs lists memory blobs for a tenant
func (r *R2StorageBackend) ListMemoryBlobs(ctx context.Context, tenantID uuid.UUID, memoryType string) ([]R2StorageObject, error) {
	prefix := fmt.Sprintf("%s/%s/%s/", R2PathMemory, tenantID.String(), memoryType)
	return r.listObjects(ctx, r.buckets.Memory, prefix)
}

// ==================== Replay Data Storage ====================

// StoreReplayData stores replay session data to R2
func (r *R2StorageBackend) StoreReplayData(ctx context.Context, tenantID, replayID uuid.UUID, events []state.StateEvent, metadata JSONMap) (*R2StorageObject, error) {
	if len(events) == 0 {
		return nil, nil
	}

	// Create replay data structure
	replayData := ReplayData{
		ID:        replayID,
		TenantID:  tenantID,
		Events:    events,
		CreatedAt: time.Now().UTC(),
		Metadata:  metadata,
	}

	data, err := json.Marshal(replayData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal replay data: %w", err)
	}

	// Build key: statefabric/replays/{tenant}/{replay}.json.gz
	key := fmt.Sprintf("%s/%s/%s.json",
		R2PathReplay,
		tenantID.String(),
		replayID.String(),
	)

	contentType := "application/json"
	storedData := data
	compressed := false

	// Compress if enabled
	if r.compression {
		storedData, err = compressData(data)
		if err == nil {
			key += ".gz"
			contentType = "application/gzip"
			compressed = true
		}
	}

	return r.storeObject(ctx, r.buckets.Replay, key, storedData, contentType, tenantID, replayID, "replay", compressed, JSONMap{
		"event_count": len(events),
	})
}

// GetReplayData retrieves replay data from R2
func (r *R2StorageBackend) GetReplayData(ctx context.Context, key string) (*ReplayData, error) {
	data, err := r.getObject(ctx, r.buckets.Replay, key)
	if err != nil {
		return nil, err
	}

	// Decompress if needed
	if strings.HasSuffix(key, ".gz") {
		data, err = decompressData(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress replay data: %w", err)
		}
	}

	var replayData ReplayData
	if err := json.Unmarshal(data, &replayData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal replay data: %w", err)
	}

	return &replayData, nil
}

// DeleteReplayData removes replay data from R2
func (r *R2StorageBackend) DeleteReplayData(ctx context.Context, key string) error {
	return r.deleteObject(ctx, r.buckets.Replay, key)
}

// ReplayData represents a replay session stored in R2
type ReplayData struct {
	ID        uuid.UUID          `json:"id"`
	TenantID  uuid.UUID          `json:"tenant_id"`
	Events    []state.StateEvent `json:"events"`
	CreatedAt time.Time          `json:"created_at"`
	Metadata  JSONMap            `json:"metadata"`
}

// ==================== Generic R2 Operations ====================

// storeObject stores an object in R2 and returns metadata
func (r *R2StorageBackend) storeObject(ctx context.Context, bucket, key string, data []byte, contentType string,
	tenantID, entityID uuid.UUID, entityType string, compressed bool, metadata JSONMap) (*R2StorageObject, error) {

	if bucket == "" {
		bucket = r.buckets.General
	}

	// Calculate content hash
	hash := sha256.Sum256(data)
	contentHash := fmt.Sprintf("%x", hash[:])

	// Store in R2
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
		ACL:         awsTypes.ObjectCannedACLPrivate,
		Metadata: map[string]string{
			"content-hash": contentHash,
			"entity-type":  entityType,
			"tenant-id":    tenantID.String(),
			"entity-id":    entityID.String(),
			"compressed":   fmt.Sprintf("%v", compressed),
			"stored-at":    time.Now().UTC().Format(time.RFC3339),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store object in R2: %w", err)
	}

	logrus.Debugf("Stored R2 object: bucket=%s, key=%s, size=%d, type=%s", bucket, key, len(data), entityType)

	return &R2StorageObject{
		Key:         key,
		Bucket:      bucket,
		Size:        int64(len(data)),
		ContentHash: contentHash,
		ContentType: contentType,
		Compressed:  compressed,
		TenantID:    tenantID,
		EntityID:    entityID,
		EntityType:  entityType,
		CreatedAt:   time.Now().UTC(),
		Metadata:    metadata,
	}, nil
}

// getObject retrieves an object from R2
func (r *R2StorageBackend) getObject(ctx context.Context, bucket, key string) ([]byte, error) {
	if bucket == "" {
		bucket = r.buckets.General
	}

	resp, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "NotFound") {
			return nil, fmt.Errorf("object not found: %s/%s", bucket, key)
		}
		return nil, fmt.Errorf("failed to get object from R2: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	return data, nil
}

// deleteObject removes an object from R2
func (r *R2StorageBackend) deleteObject(ctx context.Context, bucket, key string) error {
	if bucket == "" {
		bucket = r.buckets.General
	}

	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from R2: %w", err)
	}

	logrus.Debugf("Deleted R2 object: bucket=%s, key=%s", bucket, key)
	return nil
}

// listObjects lists objects with a given prefix
func (r *R2StorageBackend) listObjects(ctx context.Context, bucket, prefix string) ([]R2StorageObject, error) {
	if bucket == "" {
		bucket = r.buckets.General
	}

	var objects []R2StorageObject
	paginator := s3.NewListObjectsV2Paginator(r.client, &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			if obj.Key == nil || obj.Size == nil {
				continue
			}

			objects = append(objects, R2StorageObject{
				Key:       *obj.Key,
				Bucket:    bucket,
				Size:      *obj.Size,
				CreatedAt: *obj.LastModified,
			})
		}
	}

	return objects, nil
}

// GetPresignedURL generates a presigned URL for direct download
func (r *R2StorageBackend) GetPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	if bucket == "" {
		bucket = r.buckets.General
	}

	presignClient := s3.NewPresignClient(r.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}

// HealthCheck verifies the R2 connection and bucket accessibility
func (r *R2StorageBackend) HealthCheck(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("R2 client not initialized")
	}

	buckets := []string{r.buckets.Events, r.buckets.Snapshots, r.buckets.Memory, r.buckets.Replay}
	for _, bucket := range buckets {
		if bucket == "" {
			continue
		}
		_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: &bucket,
		})
		if err != nil {
			return fmt.Errorf("R2 bucket %s health check failed: %w", bucket, err)
		}
	}

	return nil
}

// GetPublicURL generates a public URL for R2 object (when using R2 public access or custom domain)
func (r *R2StorageBackend) GetPublicURL(bucket, key string) string {
	r2PublicURL := os.Getenv("R2_PUBLIC_URL")
	if r2PublicURL != "" {
		return fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(r2PublicURL, "/"), bucket, key)
	}
	// Default R2 URL format
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", r.accountID, bucket, key)
}

// ==================== Helper Functions ====================

// compressData compresses data using gzip
func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decompressData decompresses gzip data
func decompressData(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	return io.ReadAll(gz)
}

// isTextContent tries to determine if content is text
func isTextContent(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	// Check for null bytes (indicates binary)
	for _, b := range data {
		if b == 0 {
			return false
		}
	}
	return true
}

// IsConfigured returns true if R2 storage is properly configured
func IsR2StorageConfigured() bool {
	return os.Getenv("R2_ACCOUNT_ID") != "" &&
		(os.Getenv("R2_ACCESS_KEY_ID") != "" || os.Getenv("AWS_ACCESS_KEY_ID") != "") &&
		(os.Getenv("R2_SECRET_ACCESS_KEY") != "" || os.Getenv("AWS_SECRET_ACCESS_KEY") != "") &&
		(os.Getenv("R2_STATEFABRIC_BUCKET") != "" ||
			os.Getenv("R2_STATEFABRIC_EVENTS_BUCKET") != "" ||
			os.Getenv("R2_STATEFABRIC_SNAPSHOTS_BUCKET") != "")
}
