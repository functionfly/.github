package studio

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const defaultStarterFileContent = `// New FunctionFly file
export function handler(input: unknown) {
  return { ok: true, input };
}
`

// StudioProjectFile represents a source file in a Studio project.
type StudioProjectFile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Content   string    `json:"content"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StudioProject represents a Studio code project with nested files.
type StudioProject struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Files     []StudioProjectFile `json:"files"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// StudioWorkspaceState is the full workspace bootstrap payload.
type StudioWorkspaceState struct {
	Projects          []StudioProject `json:"projects"`
	ActiveProjectID   *string         `json:"active_project_id"`
	ActiveFileID      *string         `json:"active_file_id"`
}

type workspaceScope struct {
	TenantID    string
	UserID      string
	Environment string
}

// ProjectRepository handles Studio project persistence.
type ProjectRepository struct {
	db *sql.DB
}

// NewProjectRepository creates a project repository.
func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func inferLanguageFromName(name string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	switch ext {
	case "ts", "tsx":
		return "typescript"
	case "js", "jsx":
		return "javascript"
	case "py":
		return "python"
	case "go":
		return "go"
	case "rs":
		return "rust"
	case "json":
		return "json"
	case "md":
		return "markdown"
	default:
		return "plaintext"
	}
}

func normalizeFilePath(dir, name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "untitled.ts"
	}
	dir = strings.Trim(strings.TrimSpace(dir), "/")
	if dir == "" {
		return name
	}
	return dir + "/" + name
}

// GetWorkspace returns all projects with files and the active session selection.
// Creates a default project when none exist.
func (r *ProjectRepository) GetWorkspace(ctx context.Context, scope workspaceScope) (*StudioWorkspaceState, error) {
	projects, err := r.listProjectsWithFiles(ctx, scope)
	if err != nil {
		return nil, err
	}

	session, err := r.getSession(ctx, scope)
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		project, err := r.createProjectWithDefaultFile(ctx, scope, "Untitled Project", defaultStarterFileContent)
		if err != nil {
			return nil, err
		}
		projects = []StudioProject{*project}
		if err := r.saveSession(ctx, scope, &project.ID, fileIDPtr(project)); err != nil {
			return nil, err
		}
		return &StudioWorkspaceState{
			Projects:        projects,
			ActiveProjectID: &project.ID,
			ActiveFileID:    fileIDPtr(project),
		}, nil
	}

	activeProjectID, activeFileID := r.resolveSessionTargets(session, projects)
	if activeProjectID != nil {
		if err := r.saveSession(ctx, scope, activeProjectID, activeFileID); err != nil {
			return nil, err
		}
	}

	return &StudioWorkspaceState{
		Projects:        projects,
		ActiveProjectID: activeProjectID,
		ActiveFileID:    activeFileID,
	}, nil
}

func fileIDPtr(project *StudioProject) *string {
	if project == nil || len(project.Files) == 0 {
		return nil
	}
	id := project.Files[0].ID
	return &id
}

func (r *ProjectRepository) resolveSessionTargets(session *studioSession, projects []StudioProject) (*string, *string) {
	if len(projects) == 0 {
		return nil, nil
	}

	var activeProject *StudioProject
	if session != nil && session.ActiveProjectID != nil {
		for i := range projects {
			if projects[i].ID == *session.ActiveProjectID {
				activeProject = &projects[i]
				break
			}
		}
	}
	if activeProject == nil {
		activeProject = &projects[0]
	}

	var activeFileID *string
	if session != nil && session.ActiveFileID != nil {
		for _, f := range activeProject.Files {
			if f.ID == *session.ActiveFileID {
				id := f.ID
				activeFileID = &id
				break
			}
		}
	}
	if activeFileID == nil && len(activeProject.Files) > 0 {
		id := activeProject.Files[0].ID
		activeFileID = &id
	}

	id := activeProject.ID
	return &id, activeFileID
}

type studioSession struct {
	ActiveProjectID *string
	ActiveFileID    *string
}

func (r *ProjectRepository) getSession(ctx context.Context, scope workspaceScope) (*studioSession, error) {
	query := `
		SELECT active_project_id, active_file_id
		FROM studio_project_sessions
		WHERE tenant_id = $1 AND user_id = $2 AND environment = $3
	`
	var projectID, fileID sql.NullString
	err := r.db.QueryRowContext(ctx, query, scope.TenantID, scope.UserID, scope.Environment).Scan(&projectID, &fileID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get studio session: %w", err)
	}
	session := &studioSession{}
	if projectID.Valid {
		session.ActiveProjectID = &projectID.String
	}
	if fileID.Valid {
		session.ActiveFileID = &fileID.String
	}
	return session, nil
}

func (r *ProjectRepository) saveSession(ctx context.Context, scope workspaceScope, projectID, fileID *string) error {
	query := `
		INSERT INTO studio_project_sessions (id, tenant_id, user_id, environment, active_project_id, active_file_id, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (tenant_id, user_id, environment)
		DO UPDATE SET
			active_project_id = EXCLUDED.active_project_id,
			active_file_id = EXCLUDED.active_file_id,
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, uuid.New().String(), scope.TenantID, scope.UserID, scope.Environment, projectID, fileID)
	if err != nil {
		return fmt.Errorf("save studio session: %w", err)
	}
	return nil
}

func (r *ProjectRepository) listProjectsWithFiles(ctx context.Context, scope workspaceScope) ([]StudioProject, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM studio_projects
		WHERE tenant_id = $1 AND user_id = $2 AND environment = $3
		ORDER BY updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, scope.TenantID, scope.UserID, scope.Environment)
	if err != nil {
		return nil, fmt.Errorf("list studio projects: %w", err)
	}
	defer rows.Close()

	var projects []StudioProject
	for rows.Next() {
		var p StudioProject
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan studio project: %w", err)
		}
		files, err := r.listFilesForProject(ctx, scope.TenantID, p.ID)
		if err != nil {
			return nil, err
		}
		p.Files = files
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *ProjectRepository) listFilesForProject(ctx context.Context, tenantID, projectID string) ([]StudioProjectFile, error) {
	query := `
		SELECT id, name, path, content, language, created_at, updated_at
		FROM studio_project_files
		WHERE tenant_id = $1 AND project_id = $2
		ORDER BY path ASC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, projectID)
	if err != nil {
		return nil, fmt.Errorf("list studio project files: %w", err)
	}
	defer rows.Close()

	var files []StudioProjectFile
	for rows.Next() {
		var f StudioProjectFile
		if err := rows.Scan(&f.ID, &f.Name, &f.Path, &f.Content, &f.Language, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan studio project file: %w", err)
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

func (r *ProjectRepository) createProjectWithDefaultFile(ctx context.Context, scope workspaceScope, name, starterContent string) (*StudioProject, error) {
	if strings.TrimSpace(name) == "" {
		name = "Untitled Project"
	}
	projectID := uuid.New().String()
	now := time.Now()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO studio_projects (id, tenant_id, user_id, environment, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
	`, projectID, scope.TenantID, scope.UserID, scope.Environment, name, now)
	if err != nil {
		return nil, fmt.Errorf("insert studio project: %w", err)
	}

	file, err := r.insertFileTx(ctx, tx, scope.TenantID, projectID, "main.ts", "src/main.ts", starterContent, "typescript", now)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create project: %w", err)
	}

	return &StudioProject{
		ID:        projectID,
		Name:      name,
		Files:     []StudioProjectFile{*file},
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (r *ProjectRepository) insertFileTx(ctx context.Context, tx *sql.Tx, tenantID, projectID, name, path, content, language string, now time.Time) (*StudioProjectFile, error) {
	fileID := uuid.New().String()
	if language == "" {
		language = inferLanguageFromName(name)
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO studio_project_files (id, project_id, tenant_id, name, path, content, language, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
	`, fileID, projectID, tenantID, name, path, content, language, now)
	if err != nil {
		return nil, fmt.Errorf("insert studio project file: %w", err)
	}
	return &StudioProjectFile{
		ID:        fileID,
		Name:      name,
		Path:      path,
		Content:   content,
		Language:  language,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// CreateProject creates a project with a default main.ts file.
func (r *ProjectRepository) CreateProject(ctx context.Context, scope workspaceScope, name, starterContent string) (*StudioProject, error) {
	if starterContent == "" {
		starterContent = defaultStarterFileContent
	}
	return r.createProjectWithDefaultFile(ctx, scope, name, starterContent)
}

// GetProject returns one project with files.
func (r *ProjectRepository) GetProject(ctx context.Context, scope workspaceScope, projectID string) (*StudioProject, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM studio_projects
		WHERE tenant_id = $1 AND user_id = $2 AND environment = $3 AND id = $4
	`
	var p StudioProject
	err := r.db.QueryRowContext(ctx, query, scope.TenantID, scope.UserID, scope.Environment, projectID).
		Scan(&p.ID, &p.Name, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get studio project: %w", err)
	}
	files, err := r.listFilesForProject(ctx, scope.TenantID, p.ID)
	if err != nil {
		return nil, err
	}
	p.Files = files
	return &p, nil
}

// UpdateProject renames a project.
func (r *ProjectRepository) UpdateProject(ctx context.Context, scope workspaceScope, projectID, name string) (*StudioProject, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	query := `
		UPDATE studio_projects
		SET name = $1, updated_at = NOW()
		WHERE tenant_id = $2 AND user_id = $3 AND environment = $4 AND id = $5
		RETURNING id, name, created_at, updated_at
	`
	var p StudioProject
	err := r.db.QueryRowContext(ctx, query, name, scope.TenantID, scope.UserID, scope.Environment, projectID).
		Scan(&p.ID, &p.Name, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update studio project: %w", err)
	}
	files, err := r.listFilesForProject(ctx, scope.TenantID, p.ID)
	if err != nil {
		return nil, err
	}
	p.Files = files
	return &p, nil
}

// DeleteProject removes a project and its files.
func (r *ProjectRepository) DeleteProject(ctx context.Context, scope workspaceScope, projectID string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM studio_projects
		WHERE tenant_id = $1 AND user_id = $2 AND environment = $3 AND id = $4
	`, scope.TenantID, scope.UserID, scope.Environment, projectID)
	if err != nil {
		return fmt.Errorf("delete studio project: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete studio project rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("project not found")
	}
	return nil
}

// DuplicateProject clones a project and all files.
func (r *ProjectRepository) DuplicateProject(ctx context.Context, scope workspaceScope, projectID string) (*StudioProject, error) {
	source, err := r.GetProject(ctx, scope, projectID)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, nil
	}

	newID := uuid.New().String()
	now := time.Now()
	name := source.Name + " (copy)"

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO studio_projects (id, tenant_id, user_id, environment, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
	`, newID, scope.TenantID, scope.UserID, scope.Environment, name, now)
	if err != nil {
		return nil, fmt.Errorf("insert duplicated project: %w", err)
	}

	var files []StudioProjectFile
	for _, f := range source.Files {
		copied, err := r.insertFileTx(ctx, tx, scope.TenantID, newID, f.Name, f.Path, f.Content, f.Language, now)
		if err != nil {
			return nil, err
		}
		files = append(files, *copied)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit duplicate project: %w", err)
	}

	return &StudioProject{
		ID:        newID,
		Name:      name,
		Files:     files,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// CreateFile adds a file to a project.
func (r *ProjectRepository) CreateFile(ctx context.Context, scope workspaceScope, projectID, fileName, dir, content string) (*StudioProjectFile, error) {
	project, err := r.GetProject(ctx, scope, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, nil
	}

	path := normalizeFilePath(dir, fileName)
	name := filepath.Base(path)
	if content == "" {
		content = defaultStarterFileContent
	}
	language := inferLanguageFromName(name)
	now := time.Now()
	fileID := uuid.New().String()

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO studio_project_files (id, project_id, tenant_id, name, path, content, language, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
	`, fileID, projectID, scope.TenantID, name, path, content, language, now)
	if err != nil {
		return nil, fmt.Errorf("create studio file: %w", err)
	}

	_, _ = r.db.ExecContext(ctx, `
		UPDATE studio_projects SET updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`, projectID, scope.TenantID)

	return &StudioProjectFile{
		ID:        fileID,
		Name:      name,
		Path:      path,
		Content:   content,
		Language:  language,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateFile updates file metadata and/or content.
func (r *ProjectRepository) UpdateFile(ctx context.Context, scope workspaceScope, projectID, fileID string, updates map[string]interface{}) (*StudioProjectFile, error) {
	if _, err := r.GetProject(ctx, scope, projectID); err != nil {
		return nil, err
	}

	sets := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if v, ok := updates["content"]; ok {
		sets = append(sets, fmt.Sprintf("content = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["name"]; ok {
		sets = append(sets, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["path"]; ok {
		sets = append(sets, fmt.Sprintf("path = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["language"]; ok {
		sets = append(sets, fmt.Sprintf("language = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}

	if len(sets) == 1 {
		return nil, fmt.Errorf("no updates provided")
	}

	query := fmt.Sprintf(`
		UPDATE studio_project_files
		SET %s
		WHERE tenant_id = $%d AND project_id = $%d AND id = $%d
		RETURNING id, name, path, content, language, created_at, updated_at
	`, strings.Join(sets, ", "), argIdx, argIdx+1, argIdx+2)
	args = append(args, scope.TenantID, projectID, fileID)

	var f StudioProjectFile
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&f.ID, &f.Name, &f.Path, &f.Content, &f.Language, &f.CreatedAt, &f.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update studio file: %w", err)
	}

	_, _ = r.db.ExecContext(ctx, `
		UPDATE studio_projects SET updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`, projectID, scope.TenantID)

	return &f, nil
}

// DeleteFile removes a file; keeps at least one file in the project.
func (r *ProjectRepository) DeleteFile(ctx context.Context, scope workspaceScope, projectID, fileID string) error {
	project, err := r.GetProject(ctx, scope, projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return fmt.Errorf("project not found")
	}
	if len(project.Files) <= 1 {
		return fmt.Errorf("cannot delete the last file in a project")
	}

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM studio_project_files
		WHERE tenant_id = $1 AND project_id = $2 AND id = $3
	`, scope.TenantID, projectID, fileID)
	if err != nil {
		return fmt.Errorf("delete studio file: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("file not found")
	}
	return nil
}

// SaveWorkspaceSession stores active project/file selection.
func (r *ProjectRepository) SaveWorkspaceSession(ctx context.Context, scope workspaceScope, projectID, fileID *string) error {
	if projectID != nil {
		project, err := r.GetProject(ctx, scope, *projectID)
		if err != nil {
			return err
		}
		if project == nil {
			return fmt.Errorf("project not found")
		}
		if fileID != nil {
			found := false
			for _, f := range project.Files {
				if f.ID == *fileID {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("file not found in project")
			}
		}
	}
	return r.saveSession(ctx, scope, projectID, fileID)
}
