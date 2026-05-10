package bundler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/manifest"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// CompiledBundle holds compiled WASM or JS bundle bytes plus metadata.
type CompiledBundle struct {
	Hash                 string    // SHA-256 of source_code
	Bytes                []byte    // WASM binary or JS bundle
	Runtime              string    // e.g. python3.12, node20
	CompiledSizeBytes    int
	CompilationDurationMs int
	CompiledAt           time.Time
	IsValid              bool
}

// RuntimeCompiler turns source code + manifest into a compiled bundle.
type RuntimeCompiler interface {
	Compile(ctx context.Context, sourceCode string, manifest *manifest.Manifest) (*CompiledBundle, error)
}

// BundleService moves bundling from lazy (at-execution) to eager (at-publish).
type BundleService struct {
	cache     *redis.Client
	compilers map[string]RuntimeCompiler
}

// NewBundleService creates a BundleService backed by Redis.
func NewBundleService(redisAddr string) (*BundleService, error) {
	var rdb *redis.Client
	if redisAddr != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr: redisAddr,
		})
	}
	return &BundleService{
		cache:     rdb,
		compilers: make(map[string]RuntimeCompiler),
	}, nil
}

// RegisterCompiler adds a compiler for a runtime.
func (s *BundleService) RegisterCompiler(runtime string, c RuntimeCompiler) {
	s.compilers[runtime] = c
}

// cacheKey returns the Redis key for a bundle.
func cacheKey(fnVersionID, runtime, hash string) string {
	return fmt.Sprintf("bundle:%s:%s:%s", fnVersionID, runtime, hash)
}

// sourceHash computes SHA-256 of source code.
func sourceHash(source string) string {
	h := sha256.Sum256([]byte(source))
	return hex.EncodeToString(h[:])
}

// Bundle compiles a function version if not already cached.
func (s *BundleService) Bundle(ctx context.Context, fn *storage.RegistryFunctionVersion) (*CompiledBundle, error) {
	src := fn.SourceCode.String
	if src == "" {
		return nil, fmt.Errorf("function version has no source code")
	}

	hash := sourceHash(src)
	key := cacheKey(fn.ID.String(), fn.Runtime, hash)

	// 1. Check Redis L2 cache.
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, key).Bytes()
		if err == nil && len(cached) > 0 {
			logrus.WithFields(logrus.Fields{
				"function_version_id": fn.ID,
				"runtime":             fn.Runtime,
			}).Debug("Bundle cache hit (Redis L2)")
			return &CompiledBundle{
				Hash:   hash,
				Bytes:  cached,
				Runtime: fn.Runtime,
				IsValid: true,
			}, nil
		}
	}

	// 2. Compile.
	compiler, ok := s.compilers[fn.Runtime]
	if !ok {
		return nil, fmt.Errorf("no compiler registered for runtime %s", fn.Runtime)
	}

	var m manifest.Manifest
	if len(fn.Manifest) > 0 {
		_ = json.Unmarshal(fn.Manifest, &m) // best-effort
	}

	start := time.Now()
	bundle, err := compiler.Compile(ctx, src, &m)
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %w", err)
	}
	bundle.CompilationDurationMs = int(time.Since(start).Milliseconds())
	bundle.Hash = hash
	bundle.Runtime = fn.Runtime
	bundle.CompiledAt = time.Now()
	bundle.CompiledSizeBytes = len(bundle.Bytes)
	bundle.IsValid = true

	// 3. Store in Redis L2 cache.
	if s.cache != nil {
		if err := s.cache.Set(ctx, key, bundle.Bytes, 24*time.Hour).Err(); err != nil {
			logrus.WithError(err).Warn("Failed to cache bundle in Redis")
		}
	}

	logrus.WithFields(logrus.Fields{
		"function_version_id": fn.ID,
		"runtime":             fn.Runtime,
		"compiled_size":       bundle.CompiledSizeBytes,
		"compilation_ms":      bundle.CompilationDurationMs,
	}).Info("Bundle compiled and cached")

	return bundle, nil
}

// Warm pre-compiles a function version into the cache.
func (s *BundleService) Warm(ctx context.Context, fn *storage.RegistryFunctionVersion) error {
	if fn == nil {
		return nil
	}
	_, err := s.Bundle(ctx, fn)
	return err
}

// Get returns a cached bundle without compiling.
func (s *BundleService) Get(ctx context.Context, fn *storage.RegistryFunctionVersion) (*CompiledBundle, error) {
	src := fn.SourceCode.String
	if src == "" {
		return nil, fmt.Errorf("function version has no source code")
	}
	hash := sourceHash(src)
	key := cacheKey(fn.ID.String(), fn.Runtime, hash)

	if s.cache == nil {
		return nil, fmt.Errorf("no cache configured")
	}

	cached, err := s.cache.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return &CompiledBundle{
		Hash:    hash,
		Bytes:   cached,
		Runtime: fn.Runtime,
		IsValid: true,
	}, nil
}

// Close closes the Redis connection.
func (s *BundleService) Close() error {
	if s.cache != nil {
		return s.cache.Close()
	}
	return nil
}
