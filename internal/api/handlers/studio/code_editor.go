package studio

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CodeVersion represents a versioned snapshot of code in the studio editor.
type CodeVersion struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	UserID      string                 `json:"user_id"`
	Environment string                 `json:"environment"`
	FilePath    string                 `json:"file_path"`
	Content     string                 `json:"content"`
	Version     int                    `json:"version"`
	Action      string                 `json:"action"` // "save", "undo", "redo", "format"
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// CodeEditorRepository handles code version storage and retrieval.
type CodeEditorRepository struct {
	db *sql.DB
}

// NewCodeEditorRepository creates a new CodeEditorRepository.
func NewCodeEditorRepository(db *sql.DB) *CodeEditorRepository {
	return &CodeEditorRepository{db: db}
}

// GetLatestVersion returns the most recent code version for a tenant/user/filepath.
func (r *CodeEditorRepository) GetLatestVersion(ctx context.Context, tenantID, userID, environment, filePath string) (*CodeVersion, error) {
	query := `
		SELECT id, tenant_id, COALESCE(user_id, ''), COALESCE(environment, ''),
		       COALESCE(file_path, ''), content, version, action, metadata, created_at
		FROM studio_code_versions
		WHERE tenant_id = $1 AND COALESCE(user_id, '') = $2
		  AND COALESCE(environment, '') = $3 AND file_path = $4
		ORDER BY version DESC
		LIMIT 1`

	var cv CodeVersion
	var contentRaw, metaRaw []byte
	err := r.db.QueryRowContext(ctx, query, tenantID, userID, environment, filePath).Scan(
		&cv.ID, &cv.TenantID, &cv.UserID, &cv.Environment,
		&cv.FilePath, &contentRaw, &cv.Version, &cv.Action, &metaRaw, &cv.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get latest code version: %w", err)
	}
	cv.Content = string(contentRaw)
	if len(metaRaw) > 0 {
		_ = json.Unmarshal(metaRaw, &cv.Metadata)
	}
	return &cv, nil
}

// SaveVersion creates a new code version entry.
func (r *CodeEditorRepository) SaveVersion(ctx context.Context, tenantID, userID, environment, filePath, content, action string, metadata map[string]interface{}) (*CodeVersion, error) {
	// Get next version number
	var maxVersion int
	versionQuery := `
		SELECT COALESCE(MAX(version), 0) FROM studio_code_versions
		WHERE tenant_id = $1 AND COALESCE(user_id, '') = $2
		  AND COALESCE(environment, '') = $3 AND file_path = $4`
	if err := r.db.QueryRowContext(ctx, versionQuery, tenantID, userID, environment, filePath).Scan(&maxVersion); err != nil {
		return nil, fmt.Errorf("get max version: %w", err)
	}

	id := uuid.New().String()
	nextVersion := maxVersion + 1
	metaRaw, _ := json.Marshal(metadata)

	query := `
		INSERT INTO studio_code_versions (id, tenant_id, user_id, environment, file_path, content, version, action, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id, tenant_id, COALESCE(user_id, ''), COALESCE(environment, ''),
		          COALESCE(file_path, ''), content, version, action, metadata, created_at`

	var cv CodeVersion
	var contentOut, metaOut []byte
	err := r.db.QueryRowContext(ctx, query, id, tenantID, userID, environment, filePath, content, nextVersion, action, metaRaw).Scan(
		&cv.ID, &cv.TenantID, &cv.UserID, &cv.Environment,
		&cv.FilePath, &contentOut, &cv.Version, &cv.Action, &metaOut, &cv.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("save code version: %w", err)
	}
	cv.Content = string(contentOut)
	if len(metaOut) > 0 {
		_ = json.Unmarshal(metaOut, &cv.Metadata)
	}
	return &cv, nil
}

// GetVersion retrieves a specific version of code.
func (r *CodeEditorRepository) GetVersion(ctx context.Context, tenantID, userID, environment, filePath string, version int) (*CodeVersion, error) {
	query := `
		SELECT id, tenant_id, COALESCE(user_id, ''), COALESCE(environment, ''),
		       COALESCE(file_path, ''), content, version, action, metadata, created_at
		FROM studio_code_versions
		WHERE tenant_id = $1 AND COALESCE(user_id, '') = $2
		  AND COALESCE(environment, '') = $3 AND file_path = $4 AND version = $5`

	var cv CodeVersion
	var contentRaw, metaRaw []byte
	err := r.db.QueryRowContext(ctx, query, tenantID, userID, environment, filePath, version).Scan(
		&cv.ID, &cv.TenantID, &cv.UserID, &cv.Environment,
		&cv.FilePath, &contentRaw, &cv.Version, &cv.Action, &metaRaw, &cv.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get code version: %w", err)
	}
	cv.Content = string(contentRaw)
	if len(metaRaw) > 0 {
		_ = json.Unmarshal(metaRaw, &cv.Metadata)
	}
	return &cv, nil
}

// ListVersions returns version history for a file (newest first).
func (r *CodeEditorRepository) ListVersions(ctx context.Context, tenantID, userID, environment, filePath string, limit, offset int) ([]CodeVersion, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, tenant_id, COALESCE(user_id, ''), COALESCE(environment, ''),
		       COALESCE(file_path, ''), content, version, action, metadata, created_at
		FROM studio_code_versions
		WHERE tenant_id = $1 AND COALESCE(user_id, '') = $2
		  AND COALESCE(environment, '') = $3 AND file_path = $4
		ORDER BY version DESC
		LIMIT $5 OFFSET $6`

	rows, err := r.db.QueryContext(ctx, query, tenantID, userID, environment, filePath, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list code versions: %w", err)
	}
	defer rows.Close()

	var versions []CodeVersion
	for rows.Next() {
		var cv CodeVersion
		var contentRaw, metaRaw []byte
		if err := rows.Scan(&cv.ID, &cv.TenantID, &cv.UserID, &cv.Environment,
			&cv.FilePath, &contentRaw, &cv.Version, &cv.Action, &metaRaw, &cv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan code version: %w", err)
		}
		cv.Content = string(contentRaw)
		if len(metaRaw) > 0 {
			_ = json.Unmarshal(metaRaw, &cv.Metadata)
		}
		versions = append(versions, cv)
	}
	return versions, rows.Err()
}

// GetUndoRedoVersions returns the previous version (for undo) and next version (for redo).
func (r *CodeEditorRepository) GetUndoRedoVersions(ctx context.Context, tenantID, userID, environment, filePath string, currentVersion int) (prev *CodeVersion, next *CodeVersion, err error) {
	// Get previous version (one less than current)
	if currentVersion > 1 {
		prev, err = r.GetVersion(ctx, tenantID, userID, environment, filePath, currentVersion-1)
		if err != nil {
			return nil, nil, fmt.Errorf("get previous version: %w", err)
		}
	}

	// Get next version (one more than current)
	next, err = r.GetVersion(ctx, tenantID, userID, environment, filePath, currentVersion+1)
	if err != nil {
		return nil, nil, fmt.Errorf("get next version: %w", err)
	}

	return prev, next, nil
}

// CleanupOldVersions removes versions beyond the retention limit.
func (r *CodeEditorRepository) CleanupOldVersions(ctx context.Context, tenantID, userID, environment, filePath string, keepCount int) error {
	query := `
		DELETE FROM studio_code_versions
		WHERE tenant_id = $1 AND COALESCE(user_id, '') = $2
		  AND COALESCE(environment, '') = $3 AND file_path = $4
		  AND version <= (
		      SELECT COALESCE(MAX(version), 0) - $5
		      FROM studio_code_versions
		      WHERE tenant_id = $1 AND COALESCE(user_id, '') = $2
		        AND COALESCE(environment, '') = $3 AND file_path = $4
		  )`
	_, err := r.db.ExecContext(ctx, query, tenantID, userID, environment, filePath, keepCount)
	if err != nil {
		return fmt.Errorf("cleanup old versions: %w", err)
	}
	return nil
}