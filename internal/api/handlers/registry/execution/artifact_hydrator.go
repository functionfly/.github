package execution

import (
	"context"
	"database/sql"
	"io"

	"github.com/functionfly/functionfly/internal/artifacts"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// ArtifactHydrator hydrates fnVersion.WasmBinary and fnVersion.SourceCode from
// the configured artifact store. When the store is nil (or no storage keys are
// set on the row) the legacy DB columns are used as-is, preserving the
// pre-cutover behaviour.
//
// Hydration is in-place: the engine code path is untouched.
type ArtifactHydrator struct {
	resolver *artifacts.Resolver
}

// NewArtifactHydrator builds a hydrator. A nil resolver is safe and yields a
// no-op hydrator (preserves the legacy DB-column read path).
func NewArtifactHydrator(resolver *artifacts.Resolver) *ArtifactHydrator {
	if resolver == nil || resolver.Store == nil {
		return &ArtifactHydrator{}
	}
	return &ArtifactHydrator{resolver: resolver}
}

// Hydrate fills fnVersion.WasmBinary / SourceCode from the artifact store when
// the storage backend is not "db". When the resolver is nil, this is a no-op.
func (h *ArtifactHydrator) Hydrate(ctx context.Context, fnVersion *storage.RegistryFunctionVersion) {
	if h == nil || h.resolver == nil || fnVersion == nil {
		return
	}

	// When the row says bytes live in the DB, do nothing — legacy path.
	if fnVersion.StorageBackend == "" || fnVersion.StorageBackend == string(artifacts.BackendDB) {
		return
	}

	// New rows store on R2 (or local). Fetch via resolver; on miss, fall back
	// to the row's legacy columns.
	key := ""
	if fnVersion.StorageKey.Valid {
		key = fnVersion.StorageKey.String
	}
	if key != "" && len(fnVersion.WasmBinary) == 0 {
		data, err := h.resolver.Wasm(ctx, key, fnVersion.StorageBackend, nil)
		if err != nil {
			logrus.WithError(err).WithField("version_id", fnVersion.ID).Debug("artifact: wasm fetch failed, falling back to legacy column")
		} else {
			fnVersion.WasmBinary = data
		}
	}

	if fnVersion.SourceStorageKey.Valid && fnVersion.SourceStorageKey.String != "" && !fnVersion.SourceCode.Valid {
		rc, err := h.resolver.Store.Get(ctx, fnVersion.SourceStorageKey.String)
		if err != nil {
			logrus.WithError(err).WithField("version_id", fnVersion.ID).Debug("artifact: source fetch failed")
			return
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			logrus.WithError(err).Debug("artifact: source read failed")
			return
		}
		fnVersion.SourceCode = sql.NullString{String: string(data), Valid: true}
	}
}