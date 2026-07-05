package artifacts

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

// Resolver fetches function artifacts using the configured Store, transparently
// falling back to bytes still living in the legacy DB columns during the
// dual-read cutover window.
type Resolver struct {
	Store Store
	LRU   *DiskLRU

	// Metrics (atomic so callers can poll without locks).
	hitsLRU  atomic.Int64
	hitsR2   atomic.Int64
	hitsDB   atomic.Int64
	misses   atomic.Int64
}

// NewResolver wires a Resolver over the given Store. lru may be nil to disable
// on-disk caching.
func NewResolver(s Store, lru *DiskLRU) *Resolver {
	return &Resolver{Store: s, LRU: lru}
}

// Stats returns lightweight counters suitable for /healthz or metrics
// endpoints.
func (r *Resolver) Stats() map[string]int64 {
	if r == nil {
		return map[string]int64{}
	}
	return map[string]int64{
		"lru_hits":  r.hitsLRU.Load(),
		"r2_hits":   r.hitsR2.Load(),
		"db_hits":   r.hitsDB.Load(),
		"misses":    r.misses.Load(),
	}
}

// FetchBytes returns the body for key. It tries LRU, then Store, and finally —
// if both fail and a legacy DB fallback reader is provided — uses the fallback.
//
// dbFallback may be nil when the caller is the migration job and has no DB row
// to fall back to.
func (r *Resolver) FetchBytes(ctx context.Context, key string, dbFallback func() ([]byte, error)) ([]byte, error) {
	if r == nil || r.Store == nil {
		// No store configured — must rely on DB fallback.
		if dbFallback == nil {
			return nil, errors.New("artifacts: resolver not configured and no DB fallback")
		}
		return dbFallback()
	}

	// 1. LRU
	if r.LRU != nil {
		f, err := r.LRU.Get(key)
		if err == nil {
			defer func() { _ = f.Close() }()
			data, err := io.ReadAll(f)
			if err == nil {
				r.hitsLRU.Add(1)
				return data, nil
			}
		}
	}

	// 2. Object store
	rc, err := r.Store.Get(ctx, key)
	if err == nil {
		defer func() { _ = rc.Close() }()
		data, err := io.ReadAll(rc)
		if err != nil {
			return nil, err
		}
		r.hitsR2.Add(1)
		if r.LRU != nil {
			if err := r.LRU.Put(key, bytes.NewReader(data)); err != nil {
				logrus.WithError(err).WithField("key", key).Debug("artifacts: lru put failed")
			}
		}
		return data, nil
	}
	if !errors.Is(err, ErrMiss) {
		// Real error (network etc.); still try DB as last resort.
		logrus.WithError(err).WithField("key", key).Debug("artifacts: store get failed, attempting DB fallback")
	}

	// 3. DB fallback (cutover only).
	if dbFallback == nil {
		r.misses.Add(1)
		return nil, err
	}
	data, dbErr := dbFallback()
	if dbErr != nil {
		r.misses.Add(1)
		return nil, dbErr
	}
	r.hitsDB.Add(1)
	if r.LRU != nil && len(data) > 0 {
		if err := r.LRU.Put(key, bytes.NewReader(data)); err != nil {
			logrus.WithError(err).WithField("key", key).Debug("artifacts: lru put from DB fallback failed")
		}
	}
	return data, nil
}

// Wasm returns compiled WASM bytes for a function version using the dual-read
// contract described on the package doc.
//
// wasmKey / sourceKey / readmeKey are the storage_key values on the version row;
// when empty or the row's storage_backend is "db", the resolver falls back to
// the supplied closure which reads the legacy DB column.
func (r *Resolver) Wasm(ctx context.Context, wasmKey, storageBackend string, dbFallback func() ([]byte, error)) ([]byte, error) {
	if storageBackend == string(BackendDB) || wasmKey == "" {
		if dbFallback == nil {
			return nil, errors.New("artifacts: no wasm available for version")
		}
		return dbFallback()
	}
	return r.FetchBytes(ctx, wasmKey, dbFallback)
}

// Source returns the original source code as a UTF-8 string. Behaviour mirrors
// Wasm.
func (r *Resolver) Source(ctx context.Context, sourceKey, storageBackend string, dbFallback func() (string, error)) (string, error) {
	if storageBackend == string(BackendDB) || sourceKey == "" {
		if dbFallback == nil {
			return "", errors.New("artifacts: no source available for version")
		}
		return dbFallback()
	}
	data, err := r.FetchBytes(ctx, sourceKey, nil)
	if err != nil {
		// Last resort: DB fallback even when a key is set, in case the
		// object vanished during the cutover.
		if dbFallback == nil {
			return "", err
		}
		return dbFallback()
	}
	return string(data), nil
}

// MaybeOpen wraps an io.ReadCloser returned from disk to return a nil error
// when the underlying file vanished between stat and open.
func MaybeOpen(f *os.File, err error) (io.ReadCloser, error) {
	if err != nil {
		return nil, err
	}
	return f, nil
}