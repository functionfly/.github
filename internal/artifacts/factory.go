package artifacts

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// FromEnv builds the right Store backend from environment variables.
//
// ARTIFACT_STORE: r2 (default in prod) | local (default in dev when R2 isn't
// configured).
//
// ARTIFACT_MAX_BYTES: per-artifact cap (default DefaultMaxBytes).
//
// ARTIFACT_LRU_DIR: where the disk LRU caches hot objects (default
// $TMPDIR/functionfly-artifacts-cache).
//
// ARTIFACT_LRU_BYTES: max LRU size in bytes (default 512 MiB).
//
// ARTIFACT_REQUIRE: when set to "true", FromEnv returns an error if no store
// could be built (e.g. R2 env present but unreachable). Otherwise it falls
// back to local FS and returns a nil store on failure so the orchestrator can
// keep starting in degraded mode.
func FromEnv(ctx context.Context) (Store, *DiskLRU, error) {
	require := strings.EqualFold(strings.TrimSpace(os.Getenv("ARTIFACT_REQUIRE")), "true")
	if _, err := parseInt64(os.Getenv("ARTIFACT_MAX_BYTES"), DefaultMaxBytes); err != nil {
		return nil, nil, fmt.Errorf("artifacts: invalid ARTIFACT_MAX_BYTES: %w", err)
	}

	lruDir := os.Getenv("ARTIFACT_LRU_DIR")
	if lruDir == "" {
		lruDir = os.TempDir() + "/functionfly-artifacts-cache"
	}
	lruBytes, err := parseInt64(os.Getenv("ARTIFACT_LRU_BYTES"), 512*1024*1024)
	if err != nil {
		return nil, nil, fmt.Errorf("artifacts: invalid ARTIFACT_LRU_BYTES: %w", err)
	}
	lru, err := NewDiskLRU(lruDir, lruBytes)
	if err != nil {
		return nil, nil, err
	}

	mode := strings.ToLower(strings.TrimSpace(os.Getenv("ARTIFACT_STORE")))
	r2Cfg := LoadR2ConfigFromEnv()
	switch mode {
	case "", "r2":
		if r2Cfg == nil {
			if mode == "r2" {
				if require {
					return nil, nil, fmt.Errorf("artifacts: ARTIFACT_STORE=r2 but R2 credentials/bucket are not configured (ARTIFACT_REQUIRE=true)")
				}
				return nil, lru, nil
			}
			// Default mode: no R2 → use local FS. Surface clearly.
			local, lerr := NewLocalStore("")
			if lerr != nil {
				if require {
					return nil, nil, lerr
				}
				return nil, lru, nil
			}
			logrus.WithFields(logrus.Fields{
				"reason":         "no R2 credentials/bucket configured",
				"fallback":       "local FS",
				"local_root":     localLocalRoot(),
			}).Warn("artifacts: using local-filesystem backend; not for production")
			return local, lru, nil
		}
		store, rerr := NewR2Store(ctx, r2Cfg)
		if rerr != nil {
			if require {
				return nil, nil, rerr
			}
			logrus.WithError(rerr).Warn("artifacts: R2 init failed; falling back to legacy DB storage")
			return nil, lru, nil
		}
		logrus.WithFields(logrus.Fields{
			"backend":   "r2",
			"bucket":    r2Cfg.Bucket,
			"presign_ttl": PresignTTL,
		}).Info("artifacts: R2 store ready")
		return store, lru, nil
	case "local":
		root := os.Getenv("ARTIFACT_LOCAL_DIR")
		store, err := NewLocalStore(root)
		if err != nil {
			if require {
				return nil, nil, err
			}
			return nil, lru, nil
		}
		logrus.WithFields(logrus.Fields{
			"backend":   "local",
			"local_dir": store.root,
		}).Warn("artifacts: using local-filesystem backend; not for production")
		return store, lru, nil
	default:
		return nil, nil, fmt.Errorf("artifacts: unknown ARTIFACT_STORE=%q", mode)
	}
}

func localLocalRoot() string {
	// Re-export for log message; mirrors the default in NewLocalStore.
	dir := os.Getenv("ARTIFACT_LOCAL_DIR")
	if dir == "" {
		dir = os.TempDir() + "/functionfly-artifacts"
	}
	return dir
}

func parseInt64(s string, def int64) (int64, error) {
	if s == "" {
		return def, nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}