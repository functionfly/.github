package studio

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// SearchResult represents a unified Studio search hit.
type SearchResult struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Path        string  `json:"path,omitempty"`
	Relevance   float64 `json:"relevance"`
	Recent      bool    `json:"recent,omitempty"`
}

// SearchRepository runs cross-entity Studio search queries.
type SearchRepository struct {
	db *sql.DB
}

// NewSearchRepository creates a search repository.
func NewSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// SearchParams scopes a search request.
type SearchParams struct {
	TenantID    string
	UserID      string
	Environment string
	Query       string
	Type        string
	Limit       int
}

// Search returns matching projects, files, workflows, and nodes.
func (r *SearchRepository) Search(ctx context.Context, params SearchParams) ([]SearchResult, error) {
	query := strings.TrimSpace(params.Query)
	if query == "" {
		return []SearchResult{}, nil
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	pattern := "%" + query + "%"
	var results []SearchResult

	if params.Type == "" || params.Type == "project" || params.Type == "graph" {
		projectResults, err := r.searchProjects(ctx, params, pattern)
		if err != nil {
			return nil, err
		}
		results = append(results, projectResults...)
	}

	if params.Type == "" || params.Type == "file" || params.Type == "doc" {
		fileResults, err := r.searchFiles(ctx, params, pattern)
		if err != nil {
			return nil, err
		}
		results = append(results, fileResults...)
	}

	if params.Type == "" || params.Type == "graph" || params.Type == "workflow" {
		workflowResults, err := r.searchWorkflows(ctx, params, pattern)
		if err != nil {
			return nil, err
		}
		results = append(results, workflowResults...)
	}

	if params.Type == "" || params.Type == "node" {
		nodeResults, err := r.searchWorkflowNodes(ctx, params, pattern)
		if err != nil {
			return nil, err
		}
		results = append(results, nodeResults...)
	}

	if params.Type == "" || params.Type == "plugin" {
		pluginResults, err := r.searchExtensions(ctx, params, pattern)
		if err != nil {
			return nil, err
		}
		results = append(results, pluginResults...)
	}

	if len(results) > params.Limit {
		results = results[:params.Limit]
	}
	return results, nil
}

func (r *SearchRepository) searchProjects(ctx context.Context, params SearchParams, pattern string) ([]SearchResult, error) {
	query := `
		SELECT id, name, updated_at > NOW() - INTERVAL '7 days' AS recent
		FROM studio_projects
		WHERE tenant_id = $1 AND user_id = $2 AND environment = $3
		  AND name ILIKE $4
		ORDER BY updated_at DESC
		LIMIT $5`
	rows, err := r.db.QueryContext(ctx, query, params.TenantID, params.UserID, params.Environment, pattern, params.Limit)
	if err != nil {
		return nil, fmt.Errorf("search projects: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, name string
		var recent bool
		if err := rows.Scan(&id, &name, &recent); err != nil {
			return nil, fmt.Errorf("scan project search result: %w", err)
		}
		results = append(results, SearchResult{
			ID:          id,
			Type:        "graph",
			Title:       name,
			Description: "Studio project",
			Path:        "Projects/" + name,
			Relevance:   0.9,
			Recent:      recent,
		})
	}
	return results, rows.Err()
}

func (r *SearchRepository) searchFiles(ctx context.Context, params SearchParams, pattern string) ([]SearchResult, error) {
	query := `
		SELECT f.id, f.name, f.path, p.name, f.updated_at > NOW() - INTERVAL '7 days' AS recent
		FROM studio_project_files f
		INNER JOIN studio_projects p ON p.id = f.project_id
		WHERE f.tenant_id = $1 AND p.user_id = $2 AND p.environment = $3
		  AND (f.name ILIKE $4 OR f.path ILIKE $4)
		ORDER BY f.updated_at DESC
		LIMIT $5`
	rows, err := r.db.QueryContext(ctx, query, params.TenantID, params.UserID, params.Environment, pattern, params.Limit)
	if err != nil {
		return nil, fmt.Errorf("search files: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, name, path, projectName string
		var recent bool
		if err := rows.Scan(&id, &name, &path, &projectName, &recent); err != nil {
			return nil, fmt.Errorf("scan file search result: %w", err)
		}
		results = append(results, SearchResult{
			ID:          id,
			Type:        "doc",
			Title:       name,
			Description: fmt.Sprintf("File in %s", projectName),
			Path:        path,
			Relevance:   0.85,
			Recent:      recent,
		})
	}
	return results, rows.Err()
}

func (r *SearchRepository) searchWorkflows(ctx context.Context, params SearchParams, pattern string) ([]SearchResult, error) {
	query := `
		SELECT id, name, updated_at > NOW() - INTERVAL '7 days' AS recent
		FROM studio_workflows
		WHERE tenant_id = $1 AND name ILIKE $2
		ORDER BY updated_at DESC
		LIMIT $3`
	rows, err := r.db.QueryContext(ctx, query, params.TenantID, pattern, params.Limit)
	if err != nil {
		return nil, fmt.Errorf("search workflows: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, name string
		var recent bool
		if err := rows.Scan(&id, &name, &recent); err != nil {
			return nil, fmt.Errorf("scan workflow search result: %w", err)
		}
		results = append(results, SearchResult{
			ID:          id,
			Type:        "graph",
			Title:       name,
			Description: "Workflow graph",
			Path:        "Graphs/" + name,
			Relevance:   0.88,
			Recent:      recent,
		})
	}
	return results, rows.Err()
}

func (r *SearchRepository) searchWorkflowNodes(ctx context.Context, params SearchParams, pattern string) ([]SearchResult, error) {
	query := `
		SELECT n.id, n.name, n.type, w.name, w.id
		FROM studio_workflow_nodes n
		INNER JOIN studio_workflows w ON w.id = n.graph_id
		WHERE w.tenant_id = $1
		  AND (n.name ILIKE $2 OR n.type ILIKE $2)
		ORDER BY n.created_at DESC
		LIMIT $3`
	rows, err := r.db.QueryContext(ctx, query, params.TenantID, pattern, params.Limit)
	if err != nil {
		return nil, fmt.Errorf("search workflow nodes: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, name, nodeType, workflowName, workflowID string
		if err := rows.Scan(&id, &name, &nodeType, &workflowName, &workflowID); err != nil {
			return nil, fmt.Errorf("scan node search result: %w", err)
		}
		results = append(results, SearchResult{
			ID:          id,
			Type:        "node",
			Title:       name,
			Description: fmt.Sprintf("%s node in %s", nodeType, workflowName),
			Path:        fmt.Sprintf("Graphs/%s/nodes/%s", workflowName, name),
			Relevance:   0.8,
		})
	}
	return results, rows.Err()
}

func (r *SearchRepository) searchExtensions(ctx context.Context, params SearchParams, pattern string) ([]SearchResult, error) {
	query := `
		SELECT id, name, COALESCE(description, '')
		FROM studio_extensions
		WHERE tenant_id = $1
		  AND (name ILIKE $2 OR COALESCE(description, '') ILIKE $2)
		ORDER BY updated_at DESC
		LIMIT $3`
	rows, err := r.db.QueryContext(ctx, query, params.TenantID, pattern, params.Limit)
	if err != nil {
		// Table may not exist in all environments; treat as empty.
		return []SearchResult{}, nil
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, name, description string
		if err := rows.Scan(&id, &name, &description); err != nil {
			return nil, fmt.Errorf("scan extension search result: %w", err)
		}
		results = append(results, SearchResult{
			ID:          id,
			Type:        "plugin",
			Title:       name,
			Description: description,
			Relevance:   0.75,
		})
	}
	return results, rows.Err()
}
