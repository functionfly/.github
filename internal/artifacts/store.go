// Package artifacts provides content-addressed object storage for user-uploaded
// function artifacts (source code, compiled WASM, generated READMEs).
//
// Postgres keeps only metadata + content hash; the bytes live in R2 (or any
// S3-compatible backend), keyed by SHA-256 so identical uploads from different
// authors dedupe automatically.
package artifacts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"
)

// Kind identifies the type of artifact stored in object storage.
type Kind string

const (
	KindWASM   Kind = "wasm"
	KindSource Kind = "source"
	KindReadme Kind = "readme"
	KindCode   Kind = "code" // tenant "paste-code" path
)

// Backend identifies which object store implementation backs a row.
type Backend string

const (
	BackendDB   Backend = "db"   // legacy: artifact bytes still live in the DB row
	BackendR2   Backend = "r2"   // Cloudflare R2 / S3-compatible
	BackendLocal Backend = "local" // local filesystem (dev / --skip-migrations)
)

// DefaultMaxBytes is the per-artifact cap applied when ARTIFACT_MAX_BYTES is
// unset. 25 MiB comfortably fits compiled WASM + a generous Python stdlib stub.
const DefaultMaxBytes int64 = 25 * 1024 * 1024

// PresignTTL is the lifetime for presigned PUT URLs minted for direct browser
// upload. Short to limit the blast radius of a leaked URL.
const PresignTTL = 5 * time.Minute

// PresignSmallThreshold is the payload size below which the publish path uses
// server-proxied upload instead of presigned direct upload. Direct upload adds
// a round-trip and dashboard complexity that isn't worth it for tiny payloads.
const PresignSmallThreshold int64 = 256 * 1024

// Meta describes a stored artifact.
type Meta struct {
	Key         string // object key in the backend
	Backend     Backend
	ContentHash string // sha256 hex
	Size        int64
	ContentType string
}

// Store hides the concrete object-store implementation behind an interface so
// the publish / execute paths don't depend on R2 specifics.
type Store interface {
	// Put uploads bytes server-side and returns the content hash and size.
	Put(ctx context.Context, kind Kind, body io.Reader, contentType string) (Meta, error)

	// Get fetches the bytes for a previously stored key.
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Head returns metadata without downloading the body.
	Head(ctx context.Context, key string) (Meta, error)

	// Delete removes an object.
	Delete(ctx context.Context, key string) error

	// Exists reports whether the object is present.
	Exists(ctx context.Context, key string) (bool, error)

	// PresignPut returns a short-lived URL the browser can PUT directly to.
	// maxBytes is enforced by the backend (R2 rejects oversized uploads when
	// the client sends Content-Length).
	PresignPut(ctx context.Context, key, contentType string, maxBytes int64) (string, error)

	// PresignGet returns a short-lived URL the browser (or another client)
	// can GET the object from. Used for source/WASM preview without making
	// the dashboard proxy the bytes.
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)

	// Backend returns the backend name (r2, local, ...).
	Backend() Backend
}

// PresignRequest is the payload returned to the dashboard for direct upload.
type PresignRequest struct {
	Key            string `json:"key"`
	URL            string `json:"url"`
	Method         string `json:"method"`
	ContentType    string `json:"content_type"`
	MaxBytes       int64  `json:"max_bytes"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// ContentHash returns the lowercase hex SHA-256 of data.
func ContentHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// KeyFor builds a content-addressed key of the form:
//
//	<kind>/<sha[0:2]>/<sha>.<ext>
//
// ext is optional and may be empty. The 2-char prefix shards the keyspace so
// R2 list operations can be scoped to a hot subset without scanning the full
// bucket.
func KeyFor(kind Kind, sha, ext string) string {
	if len(sha) < 4 {
		sha = fmt.Sprintf("%064s", sha)
	}
	sha = strings.ToLower(sha)
	prefix := sha[:2]
	if ext == "" {
		return fmt.Sprintf("%s/%s/%s", kind, prefix, sha)
	}
	return fmt.Sprintf("%s/%s/%s.%s", kind, prefix, sha, strings.TrimPrefix(ext, "."))
}