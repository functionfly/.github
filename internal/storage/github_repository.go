package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GitHubRepository handles all GitHub integration database operations.
type GitHubRepository struct {
	db *sql.DB
}

// NewGitHubRepository creates a new GitHub repository.
func NewGitHubRepository(db *sql.DB) *GitHubRepository {
	return &GitHubRepository{db: db}
}

// ─── Connections ────────────────────────────────────────────────────────────

// CreateConnection inserts a new GitHub connection and returns the created record.
func (r *GitHubRepository) CreateConnection(ctx context.Context, conn *GitHubConnection) (*GitHubConnection, error) {
	if conn.ID == uuid.Nil {
		conn.ID = uuid.New()
	}
	now := time.Now().UTC()
	conn.CreatedAt = now
	conn.UpdatedAt = now

	if conn.Metadata == nil {
		conn.Metadata = json.RawMessage(`{}`)
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO github_connections (
			id, user_id, tenant_id, github_user_id, github_username,
			github_avatar_url, github_profile_url,
			encrypted_token, token_iv, token_tag,
			encrypted_refresh, refresh_iv, refresh_tag,
			token_scope, token_expires_at,
			github_app_install, github_install_id,
			status, last_synced_at, metadata,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7,
			$8, $9, $10,
			$11, $12, $13,
			$14, $15,
			$16, $17,
			$18, $19, $20::jsonb,
			$21, $22
		)`,
		conn.ID, conn.UserID, conn.TenantID, conn.GithubUserID, conn.GithubUsername,
		conn.GithubAvatarURL, conn.GithubProfileURL,
		conn.EncryptedToken, conn.TokenIV, conn.TokenTag,
		conn.EncryptedRefresh, conn.RefreshIV, conn.RefreshTag,
		conn.TokenScope, conn.TokenExpiresAt,
		conn.GithubAppInstall, conn.GithubInstallID,
		conn.Status, conn.LastSyncedAt, string(conn.Metadata),
		conn.CreatedAt, conn.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create github connection: %w", err)
	}

	return conn, nil
}

// GetConnectionByID retrieves a GitHub connection by its UUID.
func (r *GitHubRepository) GetConnectionByID(ctx context.Context, id uuid.UUID) (*GitHubConnection, error) {
	conn := &GitHubConnection{}
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, tenant_id, github_user_id, github_username,
			github_avatar_url, github_profile_url,
			encrypted_token, token_iv, token_tag,
			encrypted_refresh, refresh_iv, refresh_tag,
			token_scope, token_expires_at,
			github_app_install, github_install_id,
			status, last_synced_at, metadata,
			created_at, updated_at
		FROM github_connections WHERE id = $1`, id).Scan(
		&conn.ID, &conn.UserID, &conn.TenantID, &conn.GithubUserID, &conn.GithubUsername,
		&conn.GithubAvatarURL, &conn.GithubProfileURL,
		&conn.EncryptedToken, &conn.TokenIV, &conn.TokenTag,
		&conn.EncryptedRefresh, &conn.RefreshIV, &conn.RefreshTag,
		&conn.TokenScope, &conn.TokenExpiresAt,
		&conn.GithubAppInstall, &conn.GithubInstallID,
		&conn.Status, &conn.LastSyncedAt, &metadataJSON,
		&conn.CreatedAt, &conn.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get github connection: %w", err)
	}
	conn.Metadata = metadataJSON

	return conn, nil
}

// GetConnectionByUserAndGitHubID retrieves a connection by user ID and GitHub user ID.
func (r *GitHubRepository) GetConnectionByUserAndGitHubID(ctx context.Context, userID uuid.UUID, githubUserID int64) (*GitHubConnection, error) {
	conn := &GitHubConnection{}
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, tenant_id, github_user_id, github_username,
			github_avatar_url, github_profile_url,
			encrypted_token, token_iv, token_tag,
			encrypted_refresh, refresh_iv, refresh_tag,
			token_scope, token_expires_at,
			github_app_install, github_install_id,
			status, last_synced_at, metadata,
			created_at, updated_at
		FROM github_connections WHERE user_id = $1 AND github_user_id = $2`,
		userID, githubUserID).Scan(
		&conn.ID, &conn.UserID, &conn.TenantID, &conn.GithubUserID, &conn.GithubUsername,
		&conn.GithubAvatarURL, &conn.GithubProfileURL,
		&conn.EncryptedToken, &conn.TokenIV, &conn.TokenTag,
		&conn.EncryptedRefresh, &conn.RefreshIV, &conn.RefreshTag,
		&conn.TokenScope, &conn.TokenExpiresAt,
		&conn.GithubAppInstall, &conn.GithubInstallID,
		&conn.Status, &conn.LastSyncedAt, &metadataJSON,
		&conn.CreatedAt, &conn.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get github connection by user and github id: %w", err)
	}
	conn.Metadata = metadataJSON

	return conn, nil
}

// GetConnectionByUserID retrieves the first active connection for a user.
func (r *GitHubRepository) GetConnectionByUserID(ctx context.Context, userID uuid.UUID) (*GitHubConnection, error) {
	conn := &GitHubConnection{}
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, tenant_id, github_user_id, github_username,
			github_avatar_url, github_profile_url,
			encrypted_token, token_iv, token_tag,
			encrypted_refresh, refresh_iv, refresh_tag,
			token_scope, token_expires_at,
			github_app_install, github_install_id,
			status, last_synced_at, metadata,
			created_at, updated_at
		FROM github_connections WHERE user_id = $1 AND status = 'active'
		ORDER BY created_at DESC LIMIT 1`, userID).Scan(
		&conn.ID, &conn.UserID, &conn.TenantID, &conn.GithubUserID, &conn.GithubUsername,
		&conn.GithubAvatarURL, &conn.GithubProfileURL,
		&conn.EncryptedToken, &conn.TokenIV, &conn.TokenTag,
		&conn.EncryptedRefresh, &conn.RefreshIV, &conn.RefreshTag,
		&conn.TokenScope, &conn.TokenExpiresAt,
		&conn.GithubAppInstall, &conn.GithubInstallID,
		&conn.Status, &conn.LastSyncedAt, &metadataJSON,
		&conn.CreatedAt, &conn.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get github connection by user id: %w", err)
	}
	conn.Metadata = metadataJSON

	return conn, nil
}

// UpdateConnection applies a dynamic set of fields to a GitHub connection.
func (r *GitHubRepository) UpdateConnection(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates)+1)
	args := make([]interface{}, 0, len(updates)+1)
	argIdx := 1

	allowed := map[string]bool{
		"github_username": true, "github_avatar_url": true, "github_profile_url": true,
		"encrypted_token": true, "token_iv": true, "token_tag": true,
		"encrypted_refresh": true, "refresh_iv": true, "refresh_tag": true,
		"token_scope": true, "token_expires_at": true,
		"github_app_install": true, "github_install_id": true,
		"status": true, "last_synced_at": true, "metadata": true,
	}

	for key, value := range updates {
		if !allowed[key] {
			return fmt.Errorf("unsupported update field: %s", key)
		}
		if key == "metadata" {
			setParts = append(setParts, fmt.Sprintf("%s = $%d::jsonb", key, argIdx))
			switch v := value.(type) {
			case json.RawMessage:
				args = append(args, string(v))
			case []byte:
				args = append(args, string(v))
			case string:
				args = append(args, v)
			default:
				b, err := json.Marshal(value)
				if err != nil {
					return fmt.Errorf("failed to marshal metadata: %w", err)
				}
				args = append(args, string(b))
			}
		} else {
			setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIdx))
			args = append(args, value)
		}
		argIdx++
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now().UTC())
	argIdx++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE github_connections SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIdx)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update github connection: %w", err)
	}
	return nil
}

// DeleteConnection removes a GitHub connection by ID.
func (r *GitHubRepository) DeleteConnection(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM github_connections WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete github connection: %w", err)
	}
	return nil
}

// ─── Repos ──────────────────────────────────────────────────────────────────

// UpsertRepo inserts or updates a GitHub repo record by (connection_id, github_repo_id).
func (r *GitHubRepository) UpsertRepo(ctx context.Context, repo *GitHubRepo) (*GitHubRepo, error) {
	if repo.ID == uuid.Nil {
		repo.ID = uuid.New()
	}
	now := time.Now().UTC()
	repo.CreatedAt = now
	repo.UpdatedAt = now

	if repo.Languages == nil {
		repo.Languages = json.RawMessage(`{}`)
	}
	if repo.Topics == nil {
		repo.Topics = json.RawMessage(`[]`)
	}
	if repo.DetectedFunctions == nil {
		repo.DetectedFunctions = json.RawMessage(`[]`)
	}
	if repo.Metadata == nil {
		repo.Metadata = json.RawMessage(`{}`)
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO github_repos (
			id, connection_id, github_repo_id, full_name, name, owner,
			description, default_branch, language, languages,
			is_private, is_fork, is_archived, topics,
			stars_count, forks_count, size_kb, pushed_at,
			html_url, clone_url, ssh_url,
			detected_functions, detected_runtime, has_functionfly_json,
			import_status, metadata, last_scanned_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10::jsonb,
			$11, $12, $13, $14::jsonb,
			$15, $16, $17, $18,
			$19, $20, $21,
			$22::jsonb, $23, $24,
			$25, $26::jsonb, $27,
			$28, $29
		)
		ON CONFLICT (connection_id, github_repo_id) DO UPDATE SET
			full_name = EXCLUDED.full_name,
			name = EXCLUDED.name,
			owner = EXCLUDED.owner,
			description = EXCLUDED.description,
			default_branch = EXCLUDED.default_branch,
			language = EXCLUDED.language,
			languages = EXCLUDED.languages,
			is_private = EXCLUDED.is_private,
			is_fork = EXCLUDED.is_fork,
			is_archived = EXCLUDED.is_archived,
			topics = EXCLUDED.topics,
			stars_count = EXCLUDED.stars_count,
			forks_count = EXCLUDED.forks_count,
			size_kb = EXCLUDED.size_kb,
			pushed_at = EXCLUDED.pushed_at,
			html_url = EXCLUDED.html_url,
			clone_url = EXCLUDED.clone_url,
			ssh_url = EXCLUDED.ssh_url,
			updated_at = NOW()`,
		repo.ID, repo.ConnectionID, repo.GithubRepoID, repo.FullName, repo.Name, repo.Owner,
		repo.Description, repo.DefaultBranch, repo.Language, string(repo.Languages),
		repo.IsPrivate, repo.IsFork, repo.IsArchived, string(repo.Topics),
		repo.StarsCount, repo.ForksCount, repo.SizeKB, repo.PushedAt,
		repo.HtmlURL, repo.CloneURL, repo.SSHURL,
		string(repo.DetectedFunctions), repo.DetectedRuntime, repo.HasFunctionflyJSON,
		repo.ImportStatus, string(repo.Metadata), repo.LastScannedAt,
		repo.CreatedAt, repo.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert github repo: %w", err)
	}

	return repo, nil
}

// GetRepoByID retrieves a GitHub repo by its UUID.
func (r *GitHubRepository) GetRepoByID(ctx context.Context, id uuid.UUID) (*GitHubRepo, error) {
	repo := &GitHubRepo{}
	var languagesJSON, topicsJSON, detectedJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, connection_id, github_repo_id, full_name, name, owner,
			description, default_branch, language, languages,
			is_private, is_fork, is_archived, topics,
			stars_count, forks_count, size_kb, pushed_at,
			html_url, clone_url, ssh_url,
			detected_functions, detected_runtime, has_functionfly_json,
			import_status, metadata, last_scanned_at,
			created_at, updated_at
		FROM github_repos WHERE id = $1`, id).Scan(
		&repo.ID, &repo.ConnectionID, &repo.GithubRepoID, &repo.FullName, &repo.Name, &repo.Owner,
		&repo.Description, &repo.DefaultBranch, &repo.Language, &languagesJSON,
		&repo.IsPrivate, &repo.IsFork, &repo.IsArchived, &topicsJSON,
		&repo.StarsCount, &repo.ForksCount, &repo.SizeKB, &repo.PushedAt,
		&repo.HtmlURL, &repo.CloneURL, &repo.SSHURL,
		&detectedJSON, &repo.DetectedRuntime, &repo.HasFunctionflyJSON,
		&repo.ImportStatus, &metadataJSON, &repo.LastScannedAt,
		&repo.CreatedAt, &repo.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get github repo: %w", err)
	}
	repo.Languages = languagesJSON
	repo.Topics = topicsJSON
	repo.DetectedFunctions = detectedJSON
	repo.Metadata = metadataJSON

	return repo, nil
}

// GetRepoByGitHubRepoID retrieves a GitHub repo by its GitHub-side repo ID.
func (r *GitHubRepository) GetRepoByGitHubRepoID(ctx context.Context, githubRepoID int64) (*GitHubRepo, error) {
	repo := &GitHubRepo{}
	var languagesJSON, topicsJSON, detectedJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, connection_id, github_repo_id, full_name, name, owner,
			description, default_branch, language, languages,
			is_private, is_fork, is_archived, topics,
			stars_count, forks_count, size_kb, pushed_at,
			html_url, clone_url, ssh_url,
			detected_functions, detected_runtime, has_functionfly_json,
			import_status, metadata, last_scanned_at,
			created_at, updated_at
		FROM github_repos WHERE github_repo_id = $1`, githubRepoID).Scan(
		&repo.ID, &repo.ConnectionID, &repo.GithubRepoID, &repo.FullName, &repo.Name, &repo.Owner,
		&repo.Description, &repo.DefaultBranch, &repo.Language, &languagesJSON,
		&repo.IsPrivate, &repo.IsFork, &repo.IsArchived, &topicsJSON,
		&repo.StarsCount, &repo.ForksCount, &repo.SizeKB, &repo.PushedAt,
		&repo.HtmlURL, &repo.CloneURL, &repo.SSHURL,
		&detectedJSON, &repo.DetectedRuntime, &repo.HasFunctionflyJSON,
		&repo.ImportStatus, &metadataJSON, &repo.LastScannedAt,
		&repo.CreatedAt, &repo.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get github repo by github repo id: %w", err)
	}
	repo.Languages = languagesJSON
	repo.Topics = topicsJSON
	repo.DetectedFunctions = detectedJSON
	repo.Metadata = metadataJSON

	return repo, nil
}

// ListReposByConnection retrieves repos for a connection with filtering, sorting, and pagination.
// Returns the page of repos and the total count.
func (r *GitHubRepository) ListReposByConnection(ctx context.Context, connectionID uuid.UUID, params ListReposParams) ([]*GitHubRepo, int, error) {
	if params.PerPage <= 0 {
		params.PerPage = 20
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	offset := (params.Page - 1) * params.PerPage

	where := []string{"connection_id = $1"}
	args := []interface{}{connectionID}
	argIdx := 2

	if params.Language != "" {
		where = append(where, fmt.Sprintf("language = $%d", argIdx))
		args = append(args, params.Language)
		argIdx++
	}
	if params.Visibility == "public" {
		where = append(where, "is_private = false")
	} else if params.Visibility == "private" {
		where = append(where, "is_private = true")
	}
	if params.Search != "" {
		where = append(where, fmt.Sprintf("(full_name ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	sortCol := "full_name"
	switch params.Sort {
	case "stars":
		sortCol = "stars_count"
	case "updated", "updated_at":
		sortCol = "updated_at"
	case "pushed", "pushed_at":
		sortCol = "pushed_at"
	case "created", "created_at":
		sortCol = "created_at"
	case "name", "full_name":
		sortCol = "full_name"
	}
	sortDir := "ASC"
	if strings.ToLower(params.Direction) == "desc" {
		sortDir = "DESC"
	}

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM github_repos WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count github repos: %w", err)
	}

	// Fetch page
	query := fmt.Sprintf(`
		SELECT id, connection_id, github_repo_id, full_name, name, owner,
			description, default_branch, language, languages,
			is_private, is_fork, is_archived, topics,
			stars_count, forks_count, size_kb, pushed_at,
			html_url, clone_url, ssh_url,
			detected_functions, detected_runtime, has_functionfly_json,
			import_status, metadata, last_scanned_at,
			created_at, updated_at
		FROM github_repos WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		whereClause, sortCol, sortDir, argIdx, argIdx+1)
	args = append(args, params.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list github repos: %w", err)
	}
	defer rows.Close()

	var repos []*GitHubRepo
	for rows.Next() {
		repo := &GitHubRepo{}
		var languagesJSON, topicsJSON, detectedJSON, metadataJSON []byte

		if err := rows.Scan(
			&repo.ID, &repo.ConnectionID, &repo.GithubRepoID, &repo.FullName, &repo.Name, &repo.Owner,
			&repo.Description, &repo.DefaultBranch, &repo.Language, &languagesJSON,
			&repo.IsPrivate, &repo.IsFork, &repo.IsArchived, &topicsJSON,
			&repo.StarsCount, &repo.ForksCount, &repo.SizeKB, &repo.PushedAt,
			&repo.HtmlURL, &repo.CloneURL, &repo.SSHURL,
			&detectedJSON, &repo.DetectedRuntime, &repo.HasFunctionflyJSON,
			&repo.ImportStatus, &metadataJSON, &repo.LastScannedAt,
			&repo.CreatedAt, &repo.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan github repo: %w", err)
		}
		repo.Languages = languagesJSON
		repo.Topics = topicsJSON
		repo.DetectedFunctions = detectedJSON
		repo.Metadata = metadataJSON
		repos = append(repos, repo)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating github repos: %w", err)
	}

	return repos, total, nil
}

// UpdateRepoScanResults sets detected functions and runtime for a repo.
func (r *GitHubRepository) UpdateRepoScanResults(ctx context.Context, id uuid.UUID, functions json.RawMessage, runtime string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE github_repos
		SET detected_functions = $1::jsonb, detected_runtime = $2, last_scanned_at = $3, updated_at = $3
		WHERE id = $4`,
		string(functions), runtime, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update repo scan results: %w", err)
	}
	return nil
}

// UpdateRepoImportStatus sets the import_status field for a repo.
func (r *GitHubRepository) UpdateRepoImportStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE github_repos SET import_status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update repo import status: %w", err)
	}
	return nil
}

// UpdateRepoFullName updates the full_name field for a repo (e.g., after rename).
func (r *GitHubRepository) UpdateRepoFullName(ctx context.Context, id uuid.UUID, fullName string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE github_repos SET full_name = $1, updated_at = $2 WHERE id = $3`,
		fullName, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update repo full name: %w", err)
	}
	return nil
}

// ─── Imports ────────────────────────────────────────────────────────────────

// CreateImport inserts a new GitHub import job.
func (r *GitHubRepository) CreateImport(ctx context.Context, imp *GitHubImport) (*GitHubImport, error) {
	if imp.ID == uuid.Nil {
		imp.ID = uuid.New()
	}
	now := time.Now().UTC()
	imp.CreatedAt = now
	imp.UpdatedAt = now

	if imp.ManifestOverrides == nil {
		imp.ManifestOverrides = json.RawMessage(`{}`)
	}
	if imp.SyncBranches == nil {
		imp.SyncBranches = json.RawMessage(`["main"]`)
	}
	if imp.EnvironmentMappings == nil {
		imp.EnvironmentMappings = json.RawMessage(`{}`)
	}
	if imp.Metadata == nil {
		imp.Metadata = json.RawMessage(`{}`)
	}
	if imp.ErrorDetails == nil {
		imp.ErrorDetails = json.RawMessage(`{}`)
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO github_imports (
			id, user_id, tenant_id, connection_id, repo_id,
			source_branch, source_path, function_name,
			function_id, function_version_id,
			visibility, runtime_override,
			manifest_overrides, auto_sync_enabled, sync_branches, environment_mappings,
			status, progress, error_message, error_details,
			content_hash, commit_sha, files_imported, total_size_bytes,
			metadata, created_at, updated_at, completed_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10,
			$11, $12,
			$13::jsonb, $14, $15::jsonb, $16::jsonb,
			$17, $18, $19, $20::jsonb,
			$21, $22, $23, $24,
			$25::jsonb, $26, $27, $28
		)`,
		imp.ID, imp.UserID, imp.TenantID, imp.ConnectionID, imp.RepoID,
		imp.SourceBranch, imp.SourcePath, imp.FunctionName,
		imp.FunctionID, imp.FunctionVersionID,
		imp.Visibility, imp.RuntimeOverride,
		string(imp.ManifestOverrides), imp.AutoSyncEnabled, string(imp.SyncBranches), string(imp.EnvironmentMappings),
		imp.Status, imp.Progress, imp.ErrorMessage, string(imp.ErrorDetails),
		imp.ContentHash, imp.CommitSHA, imp.FilesImported, imp.TotalSizeBytes,
		string(imp.Metadata), imp.CreatedAt, imp.UpdatedAt, imp.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create github import: %w", err)
	}

	return imp, nil
}

// GetImportByID retrieves a GitHub import by its UUID.
func (r *GitHubRepository) GetImportByID(ctx context.Context, id uuid.UUID) (*GitHubImport, error) {
	imp := &GitHubImport{}
	var manifestJSON, syncBranchesJSON, envMapJSON, errorDetailsJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, tenant_id, connection_id, repo_id,
			source_branch, source_path, function_name,
			function_id, function_version_id,
			visibility, runtime_override,
			manifest_overrides, auto_sync_enabled, sync_branches, environment_mappings,
			status, progress, error_message, error_details,
			content_hash, commit_sha, files_imported, total_size_bytes,
			metadata, created_at, updated_at, completed_at
		FROM github_imports WHERE id = $1`, id).Scan(
		&imp.ID, &imp.UserID, &imp.TenantID, &imp.ConnectionID, &imp.RepoID,
		&imp.SourceBranch, &imp.SourcePath, &imp.FunctionName,
		&imp.FunctionID, &imp.FunctionVersionID,
		&imp.Visibility, &imp.RuntimeOverride,
		&manifestJSON, &imp.AutoSyncEnabled, &syncBranchesJSON, &envMapJSON,
		&imp.Status, &imp.Progress, &imp.ErrorMessage, &errorDetailsJSON,
		&imp.ContentHash, &imp.CommitSHA, &imp.FilesImported, &imp.TotalSizeBytes,
		&metadataJSON, &imp.CreatedAt, &imp.UpdatedAt, &imp.CompletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get github import: %w", err)
	}
	imp.ManifestOverrides = manifestJSON
	imp.SyncBranches = syncBranchesJSON
	imp.EnvironmentMappings = envMapJSON
	imp.ErrorDetails = errorDetailsJSON
	imp.Metadata = metadataJSON

	return imp, nil
}

// ListImportsByUser retrieves imports for a user with filtering and pagination.
func (r *GitHubRepository) ListImportsByUser(ctx context.Context, userID uuid.UUID, params ListImportsParams) ([]*GitHubImport, int, error) {
	if params.PerPage <= 0 {
		params.PerPage = 20
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	offset := (params.Page - 1) * params.PerPage

	where := []string{"user_id = $1"}
	args := []interface{}{userID}
	argIdx := 2

	if params.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, params.Status)
		argIdx++
	}
	if params.RepoID != nil {
		where = append(where, fmt.Sprintf("repo_id = $%d", argIdx))
		args = append(args, *params.RepoID)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM github_imports WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count github imports: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, tenant_id, connection_id, repo_id,
			source_branch, source_path, function_name,
			function_id, function_version_id,
			visibility, runtime_override,
			manifest_overrides, auto_sync_enabled, sync_branches, environment_mappings,
			status, progress, error_message, error_details,
			content_hash, commit_sha, files_imported, total_size_bytes,
			metadata, created_at, updated_at, completed_at
		FROM github_imports WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)
	args = append(args, params.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list github imports: %w", err)
	}
	defer rows.Close()

	var imports []*GitHubImport
	for rows.Next() {
		imp := &GitHubImport{}
		var manifestJSON, syncBranchesJSON, envMapJSON, errorDetailsJSON, metadataJSON []byte

		if err := rows.Scan(
			&imp.ID, &imp.UserID, &imp.TenantID, &imp.ConnectionID, &imp.RepoID,
			&imp.SourceBranch, &imp.SourcePath, &imp.FunctionName,
			&imp.FunctionID, &imp.FunctionVersionID,
			&imp.Visibility, &imp.RuntimeOverride,
			&manifestJSON, &imp.AutoSyncEnabled, &syncBranchesJSON, &envMapJSON,
			&imp.Status, &imp.Progress, &imp.ErrorMessage, &errorDetailsJSON,
			&imp.ContentHash, &imp.CommitSHA, &imp.FilesImported, &imp.TotalSizeBytes,
			&metadataJSON, &imp.CreatedAt, &imp.UpdatedAt, &imp.CompletedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan github import: %w", err)
		}
		imp.ManifestOverrides = manifestJSON
		imp.SyncBranches = syncBranchesJSON
		imp.EnvironmentMappings = envMapJSON
		imp.ErrorDetails = errorDetailsJSON
		imp.Metadata = metadataJSON
		imports = append(imports, imp)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating github imports: %w", err)
	}

	return imports, total, nil
}

// UpdateImportStatus sets status, progress, and error message for an import.
func (r *GitHubRepository) UpdateImportStatus(ctx context.Context, id uuid.UUID, status string, progress int, errMsg string) error {
	var completedAt *time.Time
	if status == "completed" || status == "failed" {
		now := time.Now().UTC()
		completedAt = &now
	}

	_, err := r.db.ExecContext(ctx, `
		UPDATE github_imports
		SET status = $1, progress = $2, error_message = $3, completed_at = $4, updated_at = $5
		WHERE id = $6`,
		status, progress, nullIfEmpty(errMsg), completedAt, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update github import status: %w", err)
	}
	return nil
}

// UpdateImportResult records the final result of a successful import.
func (r *GitHubRepository) UpdateImportResult(ctx context.Context, id uuid.UUID, functionID, versionID uuid.UUID, commitSHA, contentHash string, filesImported int, totalSize int64) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, `
		UPDATE github_imports
		SET function_id = $1, function_version_id = $2,
			commit_sha = $3, content_hash = $4,
			files_imported = $5, total_size_bytes = $6,
			status = 'completed', progress = 100,
			completed_at = $7, updated_at = $7
		WHERE id = $8`,
		functionID, versionID, commitSHA, contentHash,
		filesImported, totalSize, now, id)
	if err != nil {
		return fmt.Errorf("failed to update github import result: %w", err)
	}
	return nil
}

// DeleteImport removes an import by ID.
func (r *GitHubRepository) DeleteImport(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM github_imports WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete github import: %w", err)
	}
	return nil
}

// UpdateImportSyncSettings persists auto_sync_enabled and sync_branches for an import.
func (r *GitHubRepository) UpdateImportSyncSettings(ctx context.Context, id uuid.UUID, autoSyncEnabled bool, syncBranches json.RawMessage) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE github_imports
		SET auto_sync_enabled = $1, sync_branches = $2, updated_at = $3
		WHERE id = $4`,
		autoSyncEnabled, syncBranches, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update import sync settings: %w", err)
	}
	return nil
}

// ─── Webhooks ───────────────────────────────────────────────────────────────

// CreateWebhook inserts a new GitHub webhook.
func (r *GitHubRepository) CreateWebhook(ctx context.Context, wh *GitHubWebhook) (*GitHubWebhook, error) {
	if wh.ID == uuid.Nil {
		wh.ID = uuid.New()
	}
	now := time.Now().UTC()
	wh.CreatedAt = now
	wh.UpdatedAt = now

	if wh.Events == nil {
		wh.Events = json.RawMessage(`["push"]`)
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO github_webhooks (
			id, connection_id, repo_id, github_webhook_id,
			webhook_secret, events, is_active,
			last_delivery_at, last_event_type,
			delivery_count, error_count, last_error,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6::jsonb, $7,
			$8, $9,
			$10, $11, $12,
			$13, $14
		)`,
		wh.ID, wh.ConnectionID, wh.RepoID, wh.GithubWebhookID,
		wh.WebhookSecret, string(wh.Events), wh.IsActive,
		wh.LastDeliveryAt, wh.LastEventType,
		wh.DeliveryCount, wh.ErrorCount, wh.LastError,
		wh.CreatedAt, wh.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create github webhook: %w", err)
	}

	return wh, nil
}

// GetWebhookByRepoID retrieves the first webhook for a repo.
func (r *GitHubRepository) GetWebhookByRepoID(ctx context.Context, repoID uuid.UUID) (*GitHubWebhook, error) {
	wh := &GitHubWebhook{}
	var eventsJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, connection_id, repo_id, github_webhook_id,
			webhook_secret, events, is_active,
			last_delivery_at, last_event_type,
			delivery_count, error_count, last_error,
			created_at, updated_at
		FROM github_webhooks WHERE repo_id = $1 LIMIT 1`, repoID).Scan(
		&wh.ID, &wh.ConnectionID, &wh.RepoID, &wh.GithubWebhookID,
		&wh.WebhookSecret, &eventsJSON, &wh.IsActive,
		&wh.LastDeliveryAt, &wh.LastEventType,
		&wh.DeliveryCount, &wh.ErrorCount, &wh.LastError,
		&wh.CreatedAt, &wh.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get github webhook: %w", err)
	}
	wh.Events = eventsJSON

	return wh, nil
}

// GetActiveWebhooksByRepoID retrieves all active webhooks for a repo.
func (r *GitHubRepository) GetActiveWebhooksByRepoID(ctx context.Context, repoID uuid.UUID) ([]*GitHubWebhook, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, connection_id, repo_id, github_webhook_id,
			webhook_secret, events, is_active,
			last_delivery_at, last_event_type,
			delivery_count, error_count, last_error,
			created_at, updated_at
		FROM github_webhooks WHERE repo_id = $1 AND is_active = true
		ORDER BY created_at DESC`, repoID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active github webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []*GitHubWebhook
	for rows.Next() {
		wh := &GitHubWebhook{}
		var eventsJSON []byte

		if err := rows.Scan(
			&wh.ID, &wh.ConnectionID, &wh.RepoID, &wh.GithubWebhookID,
			&wh.WebhookSecret, &eventsJSON, &wh.IsActive,
			&wh.LastDeliveryAt, &wh.LastEventType,
			&wh.DeliveryCount, &wh.ErrorCount, &wh.LastError,
			&wh.CreatedAt, &wh.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan github webhook: %w", err)
		}
		wh.Events = eventsJSON
		webhooks = append(webhooks, wh)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating github webhooks: %w", err)
	}

	return webhooks, nil
}

// UpdateWebhookDelivery records a delivery attempt result.
func (r *GitHubRepository) UpdateWebhookDelivery(ctx context.Context, id uuid.UUID, eventType string, success bool, errMsg string) error {
	if success {
		_, err := r.db.ExecContext(ctx, `
			UPDATE github_webhooks
			SET last_delivery_at = $1, last_event_type = $2,
				delivery_count = delivery_count + 1,
				updated_at = $1
			WHERE id = $3`,
			time.Now().UTC(), eventType, id)
		if err != nil {
			return fmt.Errorf("failed to update webhook delivery: %w", err)
		}
	} else {
		_, err := r.db.ExecContext(ctx, `
			UPDATE github_webhooks
			SET last_delivery_at = $1, last_event_type = $2,
				delivery_count = delivery_count + 1,
				error_count = error_count + 1, last_error = $3,
				updated_at = $1
			WHERE id = $4`,
			time.Now().UTC(), eventType, errMsg, id)
		if err != nil {
			return fmt.Errorf("failed to update webhook delivery: %w", err)
		}
	}
	return nil
}

// DeleteWebhook removes a webhook by ID.
func (r *GitHubRepository) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM github_webhooks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete github webhook: %w", err)
	}
	return nil
}

// ─── Sync Logs ──────────────────────────────────────────────────────────────

// CreateSyncLog inserts a new sync log entry.
func (r *GitHubRepository) CreateSyncLog(ctx context.Context, log *GitHubSyncLog) (*GitHubSyncLog, error) {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	log.CreatedAt = time.Now().UTC()

	if log.Metadata == nil {
		log.Metadata = json.RawMessage(`{}`)
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO github_sync_logs (
			id, import_id, function_id,
			trigger_type, trigger_branch, trigger_commit_sha, trigger_pr_number,
			status, version_published, status_check_url,
			duration_ms, error_message, metadata,
			created_at, completed_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6, $7,
			$8, $9, $10,
			$11, $12, $13::jsonb,
			$14, $15
		)`,
		log.ID, log.ImportID, log.FunctionID,
		log.TriggerType, log.TriggerBranch, log.TriggerCommitSHA, log.TriggerPRNumber,
		log.Status, log.VersionPublished, log.StatusCheckURL,
		log.DurationMs, log.ErrorMessage, string(log.Metadata),
		log.CreatedAt, log.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create github sync log: %w", err)
	}

	return log, nil
}

// GetSyncLogByID retrieves a sync log by its UUID.
func (r *GitHubRepository) GetSyncLogByID(ctx context.Context, id uuid.UUID) (*GitHubSyncLog, error) {
	log := &GitHubSyncLog{}
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, import_id, function_id,
			trigger_type, trigger_branch, trigger_commit_sha, trigger_pr_number,
			status, version_published, status_check_url,
			duration_ms, error_message, metadata,
			created_at, completed_at
		FROM github_sync_logs WHERE id = $1`, id).Scan(
		&log.ID, &log.ImportID, &log.FunctionID,
		&log.TriggerType, &log.TriggerBranch, &log.TriggerCommitSHA, &log.TriggerPRNumber,
		&log.Status, &log.VersionPublished, &log.StatusCheckURL,
		&log.DurationMs, &log.ErrorMessage, &metadataJSON,
		&log.CreatedAt, &log.CompletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get github sync log: %w", err)
	}
	log.Metadata = metadataJSON

	return log, nil
}

// ListSyncLogsByImport retrieves sync logs for an import with filtering and pagination.
func (r *GitHubRepository) ListSyncLogsByImport(ctx context.Context, importID uuid.UUID, params ListSyncLogsParams) ([]*GitHubSyncLog, int, error) {
	if params.PerPage <= 0 {
		params.PerPage = 20
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	offset := (params.Page - 1) * params.PerPage

	where := []string{"import_id = $1"}
	args := []interface{}{importID}
	argIdx := 2

	if params.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, params.Status)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM github_sync_logs WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count github sync logs: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, import_id, function_id,
			trigger_type, trigger_branch, trigger_commit_sha, trigger_pr_number,
			status, version_published, status_check_url,
			duration_ms, error_message, metadata,
			created_at, completed_at
		FROM github_sync_logs WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)
	args = append(args, params.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list github sync logs: %w", err)
	}
	defer rows.Close()

	var logs []*GitHubSyncLog
	for rows.Next() {
		l := &GitHubSyncLog{}
		var metadataJSON []byte

		if err := rows.Scan(
			&l.ID, &l.ImportID, &l.FunctionID,
			&l.TriggerType, &l.TriggerBranch, &l.TriggerCommitSHA, &l.TriggerPRNumber,
			&l.Status, &l.VersionPublished, &l.StatusCheckURL,
			&l.DurationMs, &l.ErrorMessage, &metadataJSON,
			&l.CreatedAt, &l.CompletedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan github sync log: %w", err)
		}
		l.Metadata = metadataJSON
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating github sync logs: %w", err)
	}

	return logs, total, nil
}

// UpdateSyncLogStatus sets status, version, error, and duration for a sync log.
func (r *GitHubRepository) UpdateSyncLogStatus(ctx context.Context, id uuid.UUID, status, versionPublished, errMsg string, durationMs int) error {
	var completedAt *time.Time
	if status == "success" || status == "failed" {
		now := time.Now().UTC()
		completedAt = &now
	}

	_, err := r.db.ExecContext(ctx, `
		UPDATE github_sync_logs
		SET status = $1, version_published = $2, error_message = $3,
			duration_ms = $4, completed_at = $5
		WHERE id = $6`,
		status, nullIfEmpty(versionPublished), nullIfEmpty(errMsg),
		durationMs, completedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update github sync log status: %w", err)
	}
	return nil
}

// ─── Templates ──────────────────────────────────────────────────────────────

// CreateTemplate inserts a new import template.
func (r *GitHubRepository) CreateTemplate(ctx context.Context, tmpl *GitHubImportTemplate) (*GitHubImportTemplate, error) {
	if tmpl.ID == uuid.Nil {
		tmpl.ID = uuid.New()
	}
	now := time.Now().UTC()
	tmpl.CreatedAt = now
	tmpl.UpdatedAt = now

	if tmpl.DetectionRules == nil {
		tmpl.DetectionRules = json.RawMessage(`{}`)
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO github_import_templates (
			id, tenant_id, user_id, name, description,
			config, detection_rules, is_default, usage_count,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6::jsonb, $7::jsonb, $8, $9,
			$10, $11
		)`,
		tmpl.ID, tmpl.TenantID, tmpl.UserID, tmpl.Name, tmpl.Description,
		string(tmpl.Config), string(tmpl.DetectionRules), tmpl.IsDefault, tmpl.UsageCount,
		tmpl.CreatedAt, tmpl.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create github import template: %w", err)
	}

	return tmpl, nil
}

// ListTemplatesByTenant retrieves all templates for a tenant.
func (r *GitHubRepository) ListTemplatesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*GitHubImportTemplate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, user_id, name, description,
			config, detection_rules, is_default, usage_count,
			created_at, updated_at
		FROM github_import_templates
		WHERE tenant_id = $1
		ORDER BY is_default DESC, name ASC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list github import templates: %w", err)
	}
	defer rows.Close()

	var templates []*GitHubImportTemplate
	for rows.Next() {
		tmpl := &GitHubImportTemplate{}
		var configJSON, rulesJSON []byte

		if err := rows.Scan(
			&tmpl.ID, &tmpl.TenantID, &tmpl.UserID, &tmpl.Name, &tmpl.Description,
			&configJSON, &rulesJSON, &tmpl.IsDefault, &tmpl.UsageCount,
			&tmpl.CreatedAt, &tmpl.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan github import template: %w", err)
		}
		tmpl.Config = configJSON
		tmpl.DetectionRules = rulesJSON
		templates = append(templates, tmpl)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating github import templates: %w", err)
	}

	return templates, nil
}

// GetTemplateByID retrieves a template by its UUID.
func (r *GitHubRepository) GetTemplateByID(ctx context.Context, id uuid.UUID) (*GitHubImportTemplate, error) {
	tmpl := &GitHubImportTemplate{}
	var configJSON, rulesJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, user_id, name, description,
			config, detection_rules, is_default, usage_count,
			created_at, updated_at
		FROM github_import_templates WHERE id = $1`, id).Scan(
		&tmpl.ID, &tmpl.TenantID, &tmpl.UserID, &tmpl.Name, &tmpl.Description,
		&configJSON, &rulesJSON, &tmpl.IsDefault, &tmpl.UsageCount,
		&tmpl.CreatedAt, &tmpl.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get github import template: %w", err)
	}
	tmpl.Config = configJSON
	tmpl.DetectionRules = rulesJSON

	return tmpl, nil
}

// UpdateTemplate applies a dynamic set of fields to a template.
func (r *GitHubRepository) UpdateTemplate(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates)+1)
	args := make([]interface{}, 0, len(updates)+1)
	argIdx := 1

	allowed := map[string]bool{
		"name": true, "description": true, "config": true,
		"detection_rules": true, "is_default": true,
	}

	for key, value := range updates {
		if !allowed[key] {
			return fmt.Errorf("unsupported update field: %s", key)
		}
		if key == "config" || key == "detection_rules" {
			setParts = append(setParts, fmt.Sprintf("%s = $%d::jsonb", key, argIdx))
			switch v := value.(type) {
			case json.RawMessage:
				args = append(args, string(v))
			case []byte:
				args = append(args, string(v))
			case string:
				args = append(args, v)
			default:
				b, err := json.Marshal(value)
				if err != nil {
					return fmt.Errorf("failed to marshal %s: %w", key, err)
				}
				args = append(args, string(b))
			}
		} else {
			setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIdx))
			args = append(args, value)
		}
		argIdx++
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now().UTC())
	argIdx++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE github_import_templates SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIdx)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update github import template: %w", err)
	}
	return nil
}

// DeleteTemplate removes a template by ID.
func (r *GitHubRepository) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM github_import_templates WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete github import template: %w", err)
	}
	return nil
}

// IncrementTemplateUsage atomically increments the usage_count for a template.
func (r *GitHubRepository) IncrementTemplateUsage(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE github_import_templates
		SET usage_count = usage_count + 1, updated_at = $1
		WHERE id = $2`,
		time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to increment template usage: %w", err)
	}
	return nil
}

// ─── OAuth State (for GitHub OAuth CSRF protection) ──────────────────────────

// CreateOAuthState inserts a new OAuth state for CSRF protection.
func (r *GitHubRepository) CreateOAuthState(ctx context.Context, state *OAuthState) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO oauth_states (state, user_id, tenant_id, provider, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (state) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			tenant_id = EXCLUDED.tenant_id,
			provider = EXCLUDED.provider,
			expires_at = EXCLUDED.expires_at`,
		state.State, state.UserID, state.TenantID, state.Provider, state.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to create oauth state: %w", err)
	}
	return nil
}

// GetOAuthState retrieves an OAuth state and validates it's not expired.
func (r *GitHubRepository) GetOAuthState(ctx context.Context, state string) (*OAuthState, error) {
	oauthState := &OAuthState{}
	var expiresAt time.Time
	err := r.db.QueryRowContext(ctx, `
		SELECT state, user_id, tenant_id, provider, expires_at
		FROM oauth_states
		WHERE state = $1 AND expires_at > $2`,
		state, time.Now().UTC()).Scan(
		&oauthState.State, &oauthState.UserID, &oauthState.TenantID, &oauthState.Provider, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get oauth state: %w", err)
	}
	oauthState.ExpiresAt = expiresAt
	return oauthState, nil
}

// ConsumeOAuthState deletes the OAuth state after it's been used (single-use).
func (r *GitHubRepository) ConsumeOAuthState(ctx context.Context, state string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM oauth_states WHERE state = $1`,
		state)
	if err != nil {
		return fmt.Errorf("failed to consume oauth state: %w", err)
	}
	return nil
}

// CleanupExpiredOAuthStates removes expired OAuth states from the database.
func (r *GitHubRepository) CleanupExpiredOAuthStates(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM oauth_states WHERE expires_at <= $1`,
		time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired oauth states: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	return affected, nil
}
