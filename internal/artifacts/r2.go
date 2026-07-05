package artifacts

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
)

// R2Config configures a content-addressed artifact store backed by Cloudflare
// R2 (or any S3-compatible bucket).
type R2Config struct {
	AccountID     string
	AccessKeyID   string
	SecretKey     string
	Bucket        string
	Region        string
	Endpoint      string // optional override; otherwise auto from AccountID
	PresignExpiry time.Duration
}

// LoadR2ConfigFromEnv builds an R2Config from environment variables, returning
// nil when required values are missing.
func LoadR2ConfigFromEnv() *R2Config {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKeyID := os.Getenv("R2_ACCESS_KEY_ID")
	if accessKeyID == "" {
		accessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	if secretKey == "" {
		secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}
	bucket := os.Getenv("ARTIFACT_R2_BUCKET")
	if bucket == "" {
		bucket = os.Getenv("R2_STATEFABRIC_BUCKET")
	}
	if accountID == "" || accessKeyID == "" || secretKey == "" || bucket == "" {
		return nil
	}
	return &R2Config{
		AccountID:     accountID,
		AccessKeyID:   accessKeyID,
		SecretKey:     secretKey,
		Bucket:        bucket,
		Region:        os.Getenv("ARTIFACT_R2_REGION"),
		Endpoint:      os.Getenv("ARTIFACT_R2_ENDPOINT"),
		PresignExpiry: PresignTTL,
	}
}

// R2Store implements Store backed by Cloudflare R2.
type R2Store struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
	expiry  time.Duration
}

// NewR2Store builds an R2Store from the given config, verifying the bucket
// exists.
func NewR2Store(ctx context.Context, cfg *R2Config) (*R2Store, error) {
	if cfg == nil {
		return nil, errors.New("artifacts: nil R2Config")
	}
	if cfg.AccountID == "" || cfg.AccessKeyID == "" || cfg.SecretKey == "" {
		return nil, errors.New("artifacts: R2Config missing credentials")
	}
	if cfg.Bucket == "" {
		return nil, errors.New("artifacts: R2Config bucket is empty")
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)
	}

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			HostnameImmutable: true,
		}, nil
	})

	region := cfg.Region
	if region == "" {
		region = "auto"
	}

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("artifacts: load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	store := &R2Store{
		client:  client,
		presign: s3.NewPresignClient(client),
		bucket:  cfg.Bucket,
		expiry:  cfg.PresignExpiry,
	}
	if store.expiry <= 0 {
		store.expiry = PresignTTL
	}

	if _, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(cfg.Bucket)}); err != nil {
		return nil, fmt.Errorf("artifacts: R2 bucket %q unreachable: %w", cfg.Bucket, err)
	}
	logrus.WithField("bucket", cfg.Bucket).Info("artifacts: R2 store initialised")
	return store, nil
}

// Backend implements Store.
func (s *R2Store) Backend() Backend { return BackendR2 }

// Put implements Store. Bytes are streamed through an SHA-256 hasher so the
// returned hash is authoritative.
func (s *R2Store) Put(ctx context.Context, kind Kind, body io.Reader, contentType string) (Meta, error) {
	if body == nil {
		return Meta{}, errors.New("artifacts: nil body")
	}
	hasher := sha256.New()
	buf := &bytes.Buffer{}
	if _, err := io.Copy(io.MultiWriter(hasher, buf), body); err != nil {
		return Meta{}, fmt.Errorf("artifacts: read body: %w", err)
	}
	data := buf.Bytes()
	sha := hex.EncodeToString(hasher.Sum(nil))
	if kind == "" {
		kind = kindFromContentType(contentType)
	}
	ext := extFromContentType(contentType)
	key := KeyFor(kind, sha, ext)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"content-hash": sha,
			"stored-at":    time.Now().UTC().Format(time.RFC3339),
		},
	})
	if err != nil {
		return Meta{}, fmt.Errorf("artifacts: r2 put %s: %w", key, err)
	}
	return Meta{
		Key:         key,
		Backend:     BackendR2,
		ContentHash: sha,
		Size:        int64(len(data)),
		ContentType: contentType,
	}, nil
}

// Get implements Store.
func (s *R2Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("artifacts: r2 get %s: %w", key, err)
	}
	return out.Body, nil
}

// Head implements Store.
func (s *R2Store) Head(ctx context.Context, key string) (Meta, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return Meta{}, fmt.Errorf("artifacts: r2 head %s: %w", key, err)
	}
	meta := Meta{
		Key:         key,
		Backend:     BackendR2,
		Size:        aws.ToInt64(out.ContentLength),
		ContentType: aws.ToString(out.ContentType),
	}
	if out.Metadata != nil {
		if h, ok := out.Metadata["content-hash"]; ok {
			meta.ContentHash = h
		}
	}
	return meta, nil
}

// Delete implements Store.
func (s *R2Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("artifacts: r2 delete %s: %w", key, err)
	}
	return nil
}

// Exists implements Store.
func (s *R2Store) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return true, nil
	}
	if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "NoSuchKey") {
		return false, nil
	}
	return false, err
}

// PresignPut implements Store.
func (s *R2Store) PresignPut(ctx context.Context, key, contentType string, maxBytes int64) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}
	if maxBytes > 0 {
		input.ContentLength = aws.Int64(maxBytes)
	}
	req, err := s.presign.PresignPutObject(ctx, input, s3.WithPresignExpires(s.expiry))
	if err != nil {
		return "", fmt.Errorf("artifacts: presign put %s: %w", key, err)
	}
	return req.URL, nil
}

// PresignGet implements Store.
func (s *R2Store) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = s.expiry
	}
	if ttl > 24*time.Hour {
		ttl = 24 * time.Hour
	}
	req, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("artifacts: presign get %s: %w", key, err)
	}
	return req.URL, nil
}

func kindFromContentType(ct string) Kind {
	switch {
	case strings.HasPrefix(ct, "application/wasm"):
		return KindWASM
	case strings.HasPrefix(ct, "text/markdown"):
		return KindReadme
	case strings.HasPrefix(ct, "text/x-python"),
		strings.HasPrefix(ct, "application/x-python"),
		strings.HasPrefix(ct, "text/javascript"),
		strings.HasPrefix(ct, "application/javascript"),
		strings.HasPrefix(ct, "text/typescript"):
		return KindSource
	default:
		return KindCode
	}
}

func extFromContentType(ct string) string {
	switch ct {
	case "application/wasm":
		return "wasm"
	case "text/x-python", "application/x-python":
		return "py"
	case "text/javascript", "application/javascript":
		return "js"
	case "text/typescript":
		return "ts"
	case "text/markdown":
		return "md"
	default:
		return ""
	}
}