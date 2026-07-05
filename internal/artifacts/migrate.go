package artifacts

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// MigrationConfig configures the migration worker.
type MigrationConfig struct {
	BatchSize    int
	DryRun       bool
	ProgressLog  time.Duration
}

// DefaultMigrationConfig returns a config suited to the periodic watch mode.
func DefaultMigrationConfig() MigrationConfig {
	return MigrationConfig{
		BatchSize:   100,
		ProgressLog: 10 * time.Second,
	}
}

// Migrator drains legacy function artifact bytes from Postgres into the
// configured Store. After successfully uploading, the bulky columns are
// nullified and the storage_backend / storage_key columns are populated.
//
// Designed to be resumable: on each iteration it picks up where the
// previous run left off using the function_artifact_migration_cursor table.
type Migrator struct {
	DB    *gorm.DB
	Store Store
	Cfg   MigrationConfig
}

// NewMigrator builds a Migrator.
func NewMigrator(db *gorm.DB, store Store, cfg MigrationConfig) *Migrator {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	return &Migrator{DB: db, Store: store, Cfg: cfg}
}

// RunOnce processes up to one batch and returns the number of rows migrated.
// Returns context.Canceled / context.DeadlineExceeded when ctx is done.
func (m *Migrator) RunOnce(ctx context.Context) (int, error) {
	if m.Store == nil {
		return 0, errors.New("artifacts: migrator needs a Store")
	}

	count, err := m.migrateVersions(ctx)
	if err != nil {
		return count, fmt.Errorf("migrate versions: %w", err)
	}
	fnCount, err := m.migrateFunctions(ctx)
	if err != nil {
		return count + fnCount, fmt.Errorf("migrate functions: %w", err)
	}
	return count + fnCount, nil
}

// RunWatch repeatedly runs RunOnce at the configured interval until ctx is
// cancelled. Logs progress at the configured cadence.
func (m *Migrator) RunWatch(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = time.Hour
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	lastLog := time.Now()
	for {
		n, err := m.RunOnce(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			logrus.WithError(err).Warn("artifact migration batch failed")
		}
		if n > 0 || time.Since(lastLog) >= m.Cfg.ProgressLog {
			logrus.WithField("rows", n).Info("artifact migration batch")
			lastLog = time.Now()
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// migrateVersions handles registry_function_versions rows.
func (m *Migrator) migrateVersions(ctx context.Context) (int, error) {
	type row struct {
		ID              string
		WasmBinary      []byte
		SourceCode      []byte
		Readme          []byte
		StorageBackend  string
		StorageKey      string
		SourceStorageKey string
		ReadmeStorageKey string
	}

	cursor, err := m.loadCursor(ctx)
	if err != nil {
		return 0, err
	}

	var rows []row
	q := m.DB.WithContext(ctx).
		Raw(`SELECT id::text, wasm_binary, source_code, readme FROM registry_function_versions
		     WHERE storage_backend = 'db'
		       AND (wasm_binary IS NOT NULL OR source_code IS NOT NULL OR readme IS NOT NULL)
		       AND id > ?
		     ORDER BY id ASC LIMIT ?`, cursor.LastVersionID, m.Cfg.BatchSize)
	if err := q.Scan(&rows).Error; err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}

	migrated := 0
	for _, r := range rows {
		select {
		case <-ctx.Done():
			return migrated, ctx.Err()
		default:
		}

		newBackend := string(m.Store.Backend())
		updates := map[string]any{
			"storage_backend": newBackend,
		}
		if !m.Cfg.DryRun {
			if len(r.WasmBinary) > 0 {
				key, err := m.upload(ctx, KindWASM, r.WasmBinary, "application/wasm")
				if err != nil {
					logrus.WithError(err).Warn("migrator: wasm upload failed; skipping row")
					continue
				}
				updates["storage_key"] = key
				updates["artifact_hash"] = ContentHash(r.WasmBinary)
				updates["bundle_size"] = len(r.WasmBinary)
			}
			if len(r.SourceCode) > 0 {
				key, err := m.upload(ctx, KindSource, r.SourceCode, "text/plain; charset=utf-8")
				if err != nil {
					logrus.WithError(err).Warn("migrator: source upload failed; skipping row")
					continue
				}
				updates["source_storage_key"] = key
			}
			if len(r.Readme) > 0 {
				key, err := m.upload(ctx, KindReadme, r.Readme, "text/markdown; charset=utf-8")
				if err != nil {
					logrus.WithError(err).Warn("migrator: readme upload failed; skipping row")
					continue
				}
				updates["readme_storage_key"] = key
			}
			updates["wasm_binary"] = nil
			updates["source_code"] = nil
			updates["readme"] = nil

			if err := m.DB.WithContext(ctx).
				Table("registry_function_versions").
				Where("id = ?", r.ID).
				Updates(updates).Error; err != nil {
				logrus.WithError(err).WithField("version_id", r.ID).Warn("migrator: update failed")
				continue
			}
			if err := m.saveCursor(ctx, r.ID, cursor.LastFunctionID); err != nil {
				logrus.WithError(err).Debug("migrator: cursor save failed")
			}
		}
		migrated++
	}
	return migrated, nil
}

// migrateFunctions handles registry_functions.code (paste-code path).
func (m *Migrator) migrateFunctions(ctx context.Context) (int, error) {
	type row struct {
		ID   string
		Code string
	}
	cursor, err := m.loadCursor(ctx)
	if err != nil {
		return 0, err
	}
	var rows []row
	q := m.DB.WithContext(ctx).
		Raw(`SELECT id::text, code FROM registry_functions
		     WHERE code_storage_backend = 'db'
		       AND code IS NOT NULL AND length(code) > 0
		       AND id > ?
		     ORDER BY id ASC LIMIT ?`, cursor.LastFunctionID, m.Cfg.BatchSize)
	if err := q.Scan(&rows).Error; err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}

	migrated := 0
	for _, r := range rows {
		select {
		case <-ctx.Done():
			return migrated, ctx.Err()
		default:
		}

		if !m.Cfg.DryRun {
			payload := []byte(r.Code)
			key, err := m.upload(ctx, KindCode, payload, "text/plain; charset=utf-8")
			if err != nil {
				logrus.WithError(err).WithField("function_id", r.ID).Warn("migrator: code upload failed; skipping row")
				continue
			}
			updates := map[string]any{
				"code_storage_backend": string(m.Store.Backend()),
				"code_storage_key":     key,
				"code_content_hash":    ContentHash(payload),
				"code":                 "",
			}
			if err := m.DB.WithContext(ctx).
				Table("registry_functions").
				Where("id = ?", r.ID).
				Updates(updates).Error; err != nil {
				logrus.WithError(err).WithField("function_id", r.ID).Warn("migrator: code update failed")
				continue
			}
			if err := m.saveCursor(ctx, cursor.LastVersionID, r.ID); err != nil {
				logrus.WithError(err).Debug("migrator: cursor save failed")
			}
		}
		migrated++
	}
	return migrated, nil
}

func (m *Migrator) upload(ctx context.Context, kind Kind, payload []byte, ct string) (string, error) {
	meta, err := m.Store.Put(ctx, kind, bytes.NewReader(payload), ct)
	if err != nil {
		return "", err
	}
	return meta.Key, nil
}

type cursorState struct {
	LastVersionID  string
	LastFunctionID string
}

func (m *Migrator) loadCursor(ctx context.Context) (*cursorState, error) {
	var row struct {
		LastVersionID  *string
		LastFunctionID *string
	}
	err := m.DB.WithContext(ctx).
		Raw(`SELECT last_version_id::text, last_function_id::text FROM function_artifact_migration_cursor WHERE id = 1`).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	out := &cursorState{}
	if row.LastVersionID != nil {
		out.LastVersionID = *row.LastVersionID
	}
	if row.LastFunctionID != nil {
		out.LastFunctionID = *row.LastFunctionID
	}
	return out, nil
}

func (m *Migrator) saveCursor(ctx context.Context, versionID, functionID string) error {
	return m.DB.WithContext(ctx).
		Exec(`UPDATE function_artifact_migration_cursor
		      SET last_version_id = COALESCE(NULLIF(?, '')::uuid, last_version_id),
		          last_function_id = COALESCE(NULLIF(?, '')::uuid, last_function_id),
		          updated_at = NOW()
		      WHERE id = 1`, versionID, functionID).Error
}

// _ unused import guard
var _ = log.Default