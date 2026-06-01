package deployment

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	"github.com/sirupsen/logrus"
)

// R2ArtifactStore is a Cloudflare R2-based implementation of ArtifactStore for production use
// with cross-region replication support for disaster recovery.
type R2ArtifactStore struct {
	client *s3.Client
	primaryBucket       string
	replicaBuckets      []string   // Additional buckets for cross-region replication
	primaryRegion       string
	replicaRegions      []string
	syncReplication     bool       // If true, replication is synchronous (blocking)
	versioningEnabled   bool       // If true, uses R2 object versioning for rollback support
}

// R2StoreConfig holds configuration for the R2 artifact store
type R2StoreConfig struct {
	AccountID         string
	AccessKeyID      string
	SecretKey        string
	PrimaryBucket    string
	PrimaryRegion    string
	ReplicaBuckets   []string // Optional: additional buckets for cross-region replication
	ReplicaRegions   []string // Optional: regions corresponding to replica buckets
	SyncReplication  bool     // If true, replication is synchronous (blocking)
	VersioningEnabled bool     // If true, uses R2 object versioning for rollback support
}

// NewR2ArtifactStore creates a new R2-based artifact store from environment variables
func NewR2ArtifactStore() (*R2ArtifactStore, error) {
	config := loadR2ConfigFromEnv()
	if config == nil {
		return nil, fmt.Errorf("R2 configuration not available: missing required environment variables")
	}

	return NewR2ArtifactStoreWithConfig(config)
}

// NewR2ArtifactStoreWithConfig creates a new R2-based artifact store with explicit configuration
func NewR2ArtifactStoreWithConfig(cfg *R2StoreConfig) (*R2ArtifactStore, error) {
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

	// Load AWS config with R2-specific settings
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.PrimaryRegion),
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

	// Verify primary bucket exists and is accessible
	_, err = client.HeadBucket(ctx,&s3.HeadBucketInput{
		Bucket: &cfg.PrimaryBucket,
	})
	if err != nil {
		return nil, fmt.Errorf("R2 primary bucket %s not accessible: %w", cfg.PrimaryBucket, err)
	}

	store := &R2ArtifactStore{
		client:            client,
		primaryBucket:     cfg.PrimaryBucket,
		replicaBuckets:   cfg.ReplicaBuckets,
		primaryRegion:    cfg.PrimaryRegion,
		replicaRegions:    cfg.ReplicaRegions,
		syncReplication:  cfg.SyncReplication,
		versioningEnabled: cfg.VersioningEnabled,
	}

	logrus.Infof("R2ArtifactStore initialized: primary=%s (region=%s), replicas=%d, sync=%v, versioning=%v",
		cfg.PrimaryBucket, cfg.PrimaryRegion, len(cfg.ReplicaBuckets), cfg.SyncReplication, cfg.VersioningEnabled)

	return store, nil
}

// loadR2ConfigFromEnv loads R2 configuration from environment variables
func loadR2ConfigFromEnv() *R2StoreConfig {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKeyID := os.Getenv("R2_ACCESS_KEY_ID")
	if accessKeyID == "" {
		accessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	if secretKey == "" {
		secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}
	primaryBucket := os.Getenv("R2_ARTIFACT_BUCKET_PRIMARY")
	if primaryBucket == "" {
		primaryBucket = os.Getenv("R2_BUCKET") // Fallback to general bucket
	}
	primaryRegion := os.Getenv("R2_PRIMARY_REGION")
	if primaryRegion == "" {
		primaryRegion = "auto"
	}

	// Check required values
	if accountID == "" || accessKeyID == "" || secretKey == "" || primaryBucket == "" {
		return nil
	}

	cfg := &R2StoreConfig{
		AccountID:     accountID,
		AccessKeyID:   accessKeyID,
		SecretKey:     secretKey,
		PrimaryBucket: primaryBucket,
		PrimaryRegion: primaryRegion,
	}

	// Load optional replica configurations (up to 3 replicas)
	for i := 1; i <= 3; i++ {
		replicaBucket := os.Getenv(fmt.Sprintf("R2_ARTIFACT_BUCKET_REPLICA_%d", i))
		if replicaBucket == "" {
			continue
		}
		replicaRegion := os.Getenv(fmt.Sprintf("R2_ARTIFACT_REGION_REPLICA_%d", i))
		if replicaRegion == "" {
			replicaRegion = "auto"
		}
		cfg.ReplicaBuckets = append(cfg.ReplicaBuckets, replicaBucket)
		cfg.ReplicaRegions = append(cfg.ReplicaRegions, replicaRegion)
	}

	// Sync replication: blocks until all replicas confirm write
	cfg.SyncReplication = os.Getenv("R2_ARTIFACT_SYNC_REPLICATION") == "true"

	// Versioning: enables R2 object versioning for rollback support
	cfg.VersioningEnabled = os.Getenv("R2_ARTIFACT_VERSIONING_ENABLED") == "true"

	return cfg
}

// Store stores an artifact with the given key in R2 (primary bucket) and optionally replicates to replica buckets
func (s *R2ArtifactStore) Store(ctx context.Context, key string, data []byte) error {
	if s.client == nil {
		return fmt.Errorf("R2 client not initialized")
	}

	// Validate key format
	if !strings.HasPrefix(key, "deployments/") {
		key = fmt.Sprintf("deployments/%s", key)
	}

	// Store in primary bucket
	contentType := "application/javascript"
	if strings.HasSuffix(key, ".wasm") {
		contentType = "application/wasm"
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.primaryBucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
		ACL:         awsTypes.ObjectCannedACLPrivate,
		Metadata: map[string]string{
			"artifact-type":    "function-deployment",
			"stored-at":        time.Now().UTC().Format(time.RFC3339),
			"content-checksum": sha256Sum(data),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to store artifact in R2 primary bucket: %w", err)
	}

	logrus.Debugf("Artifact stored in R2 primary bucket: %s/%s", s.primaryBucket, key)

	// Replicate to replica buckets (sync or async based on configuration)
	if len(s.replicaBuckets) > 0 {
		if s.syncReplication {
			// Synchronous replication: block until all replicas confirm
			s.replicateToReplicasSync(ctx, key, data, contentType)
		} else {
			// Asynchronous replication: non-blocking
			go s.replicateToReplicas(ctx, key, data, contentType)
		}
	}

	return nil
}

// replicateToReplicas replicates the artifact to configured replica buckets (async)
func (s *R2ArtifactStore) replicateToReplicas(ctx context.Context, key string, data []byte, contentType string) {
	for i, replicaBucket := range s.replicaBuckets {
		region := "auto"
		if i < len(s.replicaRegions) {
			region = s.replicaRegions[i]
		}

		// Create new context with timeout for replication
		replicaCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		_, err := s.client.PutObject(replicaCtx, &s3.PutObjectInput{
			Bucket:      &replicaBucket,
			Key:         &key,
			Body:        bytes.NewReader(data),
			ContentType: &contentType,
			ACL:         awsTypes.ObjectCannedACLPrivate,
			Metadata: map[string]string{
				"artifact-type":    "function-deployment",
				"stored-at":        time.Now().UTC().Format(time.RFC3339),
				"content-checksum": sha256Sum(data),
				"replicated-from":  s.primaryBucket,
				"replica-region":   region,
			},
		})

		cancel()

		if err != nil {
			logrus.Warnf("Failed to replicate artifact to R2 replica bucket %s (region=%s): %v",
				replicaBucket, region, err)
		} else {
			logrus.Debugf("Artifact replicated to R2 replica bucket: %s/%s (region=%s)",
				replicaBucket, key, region)
		}
	}
}

// replicateToReplicasSync replicates the artifact to configured replica buckets synchronously
func (s *R2ArtifactStore) replicateToReplicasSync(ctx context.Context, key string, data []byte, contentType string) {
	for i, replicaBucket := range s.replicaBuckets {
		region := "auto"
		if i < len(s.replicaRegions) {
			region = s.replicaRegions[i]
		}

		// Create new context with timeout for replication
		replicaCtx, cancel := context.WithTimeout(ctx, 60*time.Second)

		_, err := s.client.PutObject(replicaCtx, &s3.PutObjectInput{
			Bucket:      &replicaBucket,
			Key:         &key,
			Body:        bytes.NewReader(data),
			ContentType: &contentType,
			ACL:         awsTypes.ObjectCannedACLPrivate,
			Metadata: map[string]string{
				"artifact-type":    "function-deployment",
				"stored-at":        time.Now().UTC().Format(time.RFC3339),
				"content-checksum": sha256Sum(data),
				"replicated-from":  s.primaryBucket,
				"replica-region":   region,
			},
		})

		cancel()

		if err != nil {
			logrus.Errorf("Failed to synchronously replicate artifact to R2 replica bucket %s (region=%s): %v",
				replicaBucket, region, err)
		} else {
			logrus.Infof("Artifact synchronously replicated to R2 replica bucket: %s/%s (region=%s)",
				replicaBucket, key, region)
		}
	}
}

// Retrieve retrieves an artifact by key from R2 (primary bucket, with fallback to replicas on network error)
func (s *R2ArtifactStore) Retrieve(ctx context.Context, key string) ([]byte, error) {
	if s.client == nil {
		return nil, fmt.Errorf("R2 client not initialized")
	}

	// Validate key format
	if !strings.HasPrefix(key, "deployments/") {
		key = fmt.Sprintf("deployments/%s", key)
	}

	// Try primary bucket first
	data, err := s.retrieveFromBucket(ctx, s.primaryBucket, key)
	if err == nil {
		return data, nil
	}

	// Check if it's a network error (not just "not found") - if so, try replicas in parallel
	if isNetworkError(err) && len(s.replicaBuckets) > 0 {
		logrus.Warnf("Primary bucket network error for %s, checking replicas in parallel: %v", key, err)
		return s.retrieveFromReplicasParallel(ctx, key)
	}

	// If not found in primary and we have replicas, try replicas sequentially
	if len(s.replicaBuckets) > 0 {
		logrus.Warnf("Artifact not found in primary bucket %s, checking replicas: %s", s.primaryBucket, key)

		for _, replicaBucket := range s.replicaBuckets {
			data, err = s.retrieveFromBucket(ctx, replicaBucket, key)
			if err == nil {
				logrus.Infof("Artifact retrieved from replica bucket %s: %s", replicaBucket, key)
				return data, nil
			}
		}
	}

	return nil, fmt.Errorf("artifact not found in R2 (primary or replicas): %s", key)
}

// retrieveFromReplicasParallel tries to retrieve from all replicas in parallel, returning the first successful response
func (s *R2ArtifactStore) retrieveFromReplicasParallel(ctx context.Context, key string) ([]byte, error) {
	type result struct {
		data   []byte
		bucket string
		err    error
	}

	resultCh := make(chan result, len(s.replicaBuckets))

	for _, replicaBucket := range s.replicaBuckets {
		go func(bucket string) {
			data, err := s.retrieveFromBucket(ctx, bucket, key)
			resultCh <- result{data: data, bucket: bucket, err: err}
		}(replicaBucket)
	}

	// Wait for first successful result or all failures
	var lastErr error
	for i := 0; i < len(s.replicaBuckets); i++ {
		select {
		case res := <-resultCh:
			if res.err == nil {
				logrus.Infof("Artifact retrieved from replica bucket %s (fastest response)", res.bucket)
				return res.data, nil
			}
			lastErr = res.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("artifact not found in any replica: %s (last error: %v)", key, lastErr)
}

// isNetworkError returns true if the error is a network/connection error (not a "not found" error)
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Check for common network error patterns
	networkIndicators := []string{
		"connection refused",
		"timeout",
		"request canceled",
		"context deadline",
		"i/o timeout",
		"network",
		"temporary failure",
		"no such host",
		"connection reset",
		"connection aborted",
		"endpoint",
	}
	for _, indicator := range networkIndicators {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(indicator)) {
			return true
		}
	}
	return false
}

// retrieveFromBucket retrieves an artifact from a specific bucket
func (s *R2ArtifactStore) retrieveFromBucket(ctx context.Context, bucket, key string) ([]byte, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "NotFound") {
			return nil, fmt.Errorf("artifact not found")
		}
		return nil, fmt.Errorf("failed to retrieve from bucket %s: %w", bucket, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read artifact body from bucket %s: %w", bucket, err)
	}

	return data, nil
}

// Delete removes an artifact by key from R2 (primary and all replica buckets)
func (s *R2ArtifactStore) Delete(ctx context.Context, key string) error {
	if s.client == nil {
		return fmt.Errorf("R2 client not initialized")
	}

	// Validate key format
	if !strings.HasPrefix(key, "deployments/") {
		key = fmt.Sprintf("deployments/%s", key)
	}

	var lastErr error

	// Delete from primary bucket
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.primaryBucket,
		Key:    &key,
	})
	if err != nil {
		lastErr = fmt.Errorf("failed to delete from primary bucket %s: %w", s.primaryBucket, err)
		logrus.Warnf("Failed to delete artifact from R2 primary bucket: %v", err)
	} else {
		logrus.Debugf("Artifact deleted from R2 primary bucket: %s/%s", s.primaryBucket, key)
	}

	// Delete from all replica buckets
	for _, replicaBucket := range s.replicaBuckets {
		_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: &replicaBucket,
			Key:    &key,
		})
		if err != nil {
			lastErr = fmt.Errorf("failed to delete from replica bucket %s: %w", replicaBucket, err)
			logrus.Warnf("Failed to delete artifact from R2 replica bucket %s: %v", replicaBucket, err)
		} else {
			logrus.Debugf("Artifact deleted from R2 replica bucket: %s/%s", replicaBucket, key)
		}
	}

	return lastErr
}

// HealthCheck verifies the R2 connection and bucket accessibility
func (s *R2ArtifactStore) HealthCheck(ctx context.Context) error {
	if s.client == nil {
		return fmt.Errorf("R2 client not initialized")
	}

	// Check primary bucket
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &s.primaryBucket,
	})
	if err != nil {
		return fmt.Errorf("R2 primary bucket health check failed: %w", err)
	}

	// Check replica buckets
	for _, replicaBucket := range s.replicaBuckets {
		_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: &replicaBucket,
		})
		if err != nil {
			logrus.Warnf("R2 replica bucket health check failed for %s: %v", replicaBucket, err)
			// Don't fail health check for replica issues, just warn
		}
	}

	return nil
}

// GetPresignedURL generates a presigned URL for direct artifact download (valid for 1 hour)
func (s *R2ArtifactStore) GetPresignedURL(ctx context.Context, key string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("R2 client not initialized")
	}

	// Validate key format
	if !strings.HasPrefix(key, "deployments/") {
		key = fmt.Sprintf("deployments/%s", key)
	}

	presignClient := s3.NewPresignClient(s.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.primaryBucket,
		Key:    &key,
	}, s3.WithPresignExpires(1*time.Hour))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}

// ListArtifacts lists artifacts with an optional prefix filter
func (s *R2ArtifactStore) ListArtifacts(ctx context.Context, prefix string) ([]string, error) {
	if s.client == nil {
		return nil, fmt.Errorf("R2 client not initialized")
	}

	if prefix != "" && !strings.HasPrefix(prefix, "deployments/") {
		prefix = fmt.Sprintf("deployments/%s", prefix)
	}

	var keys []string
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: &s.primaryBucket,
		Prefix: &prefix,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list artifacts: %w", err)
		}

		for _, obj := range page.Contents {
			if obj.Key != nil {
				keys = append(keys, *obj.Key)
			}
		}
	}

	return keys, nil
}

// GetReplicationStatus returns the replication status for a given artifact
func (s *R2ArtifactStore) GetReplicationStatus(ctx context.Context, key string) (*ReplicationStatus, error) {
	if !strings.HasPrefix(key, "deployments/") {
		key = fmt.Sprintf("deployments/%s", key)
	}

	status := &ReplicationStatus{
		Key:            key,
		PrimaryBucket:  s.primaryBucket,
		PrimaryRegion:  s.primaryRegion,
		ReplicaStatus:  make(map[string]bool),
	}

	// Check primary
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &s.primaryBucket,
		Key:    &key,
	})
	status.PrimaryExists = err == nil

	// Check replicas
	for i, replicaBucket := range s.replicaBuckets {
		region := "unknown"
		if i < len(s.replicaRegions) {
			region = s.replicaRegions[i]
		}

		_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: &replicaBucket,
			Key:    &key,
		})
		status.ReplicaStatus[fmt.Sprintf("%s(%s)", replicaBucket, region)] = err == nil
	}

	return status, nil
}

// ReplicationStatus represents the replication status of an artifact across buckets
type ReplicationStatus struct {
	Key            string
	PrimaryBucket  string
	PrimaryRegion  string
	PrimaryExists  bool
	ReplicaStatus  map[string]bool // bucket(region) -> exists
}

// AllReplicasExist returns true if all configured replicas have the artifact
func (s *ReplicationStatus) AllReplicasExist() bool {
	if !s.PrimaryExists {
		return false
	}
	for _, exists := range s.ReplicaStatus {
		if !exists {
			return false
		}
	}
	return true
}

// sha256Sum computes SHA256 hash of data and returns hex-encoded string for metadata
func sha256Sum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}