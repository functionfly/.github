package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	githubsvc "github.com/functionfly/functionfly/internal/services/github"
)

// HandleListRepos lists cached repos with pagination and filtering.
func (h *Handler) HandleListRepos(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	conn, err := h.githubRepo.GetConnectionByUserID(r.Context(), claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get connection")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to get connection")
		return
	}
	if conn == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "No GitHub connection found")
		return
	}

	params := storage.ListReposParams{
		Page:       1,
		PerPage:    20,
		Sort:       "full_name",
		Direction:  "asc",
		Language:   r.URL.Query().Get("language"),
		Visibility: r.URL.Query().Get("visibility"),
		Search:     r.URL.Query().Get("search"),
	}

	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			params.Page = n
		}
	}
	if v := r.URL.Query().Get("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			params.PerPage = n
		}
	}
	if v := r.URL.Query().Get("sort"); v != "" {
		params.Sort = v
	}
	if v := r.URL.Query().Get("direction"); v != "" {
		params.Direction = v
	}

	repos, total, err := h.githubRepo.ListReposByConnection(r.Context(), conn.ID, params)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list repos")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to list repos")
		return
	}

	respRepos := make([]*RepoResponse, len(repos))
	for i, repo := range repos {
		respRepos[i] = h.mapRepoResponse(repo)
	}

	h.respondJSON(w, http.StatusOK, ListReposResponse{
		Repos:   respRepos,
		Total:   total,
		Page:    params.Page,
		PerPage: params.PerPage,
	})
}

// HandleRefreshRepos re-fetches repos from GitHub API and upserts into cache.
func (h *Handler) HandleRefreshRepos(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	conn, err := h.githubRepo.GetConnectionByUserID(r.Context(), claims.UserID)
	if err != nil || conn == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "No GitHub connection found")
		return
	}

	ghClient, err := h.getGitHubClient(r.Context(), claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create GitHub client")
		h.respondError(w, http.StatusInternalServerError, "client_error", "Failed to connect to GitHub")
		return
	}

	var allRepos []*githubsvc.GitHubRepo
	page := 1
	for {
		repos, err := ghClient.ListRepos(r.Context(), githubsvc.ListReposOptions{
			Page:    page,
			PerPage: 100,
			Sort:    "updated",
		})
		if err != nil {
			h.logger.WithError(err).Error("Failed to list repos from GitHub")
			h.respondError(w, http.StatusBadGateway, "github_error", "Failed to fetch repos from GitHub")
			return
		}
		if len(repos) == 0 {
			break
		}
		allRepos = append(allRepos, repos...)
		if len(repos) < 100 {
			break
		}
		page++
	}

	topicsJSON, _ := json.Marshal([]string{})
	now := time.Now().UTC()

	for _, repo := range allRepos {
		if repo.Topics == nil {
			repo.Topics = []string{}
		}
		topics, _ := json.Marshal(repo.Topics)

		var description *string
		if repo.Description != "" {
			description = &repo.Description
		}

		var pushedAt *time.Time
		if repo.PushedAt != nil {
			pushedAt = repo.PushedAt
		}

		ghRepo := &storage.GitHubRepo{
			ConnectionID:  conn.ID,
			GithubRepoID:  repo.ID,
			FullName:      repo.FullName,
			Name:          repo.Name,
			Owner:         repo.Owner.Login,
			Description:   description,
			DefaultBranch: repo.DefaultBranch,
			Language:      &repo.Language,
			IsPrivate:     repo.Private,
			IsFork:        repo.Fork,
			IsArchived:    repo.Archived,
			Topics:        topics,
			StarsCount:    repo.StargazersCount,
			ForksCount:    repo.ForksCount,
			SizeKB:        repo.Size,
			PushedAt:      pushedAt,
			HtmlURL:       repo.HTMLURL,
			CloneURL:      repo.CloneURL,
			SSHURL:        repo.SSHURL,
			ImportStatus:  "available",
		}
		_ = topicsJSON
		_ = now

		if _, err := h.githubRepo.UpsertRepo(r.Context(), ghRepo); err != nil {
			h.logger.WithError(err).WithField("repo", repo.FullName).Warn("Failed to upsert repo")
		}
	}

	_ = h.githubRepo.UpdateConnection(r.Context(), conn.ID, map[string]interface{}{
		"last_synced_at": time.Now().UTC(),
	})

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"refreshed": len(allRepos),
	})
}

// HandleGetRepo returns a single repo by ID.
func (h *Handler) HandleGetRepo(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	repoID, err := h.parseUUID(r, "repoId")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid repo ID")
		return
	}

	repo, err := h.githubRepo.GetRepoByID(r.Context(), repoID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get repo")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to get repo")
		return
	}
	if repo == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "Repo not found")
		return
	}

	conn, err := h.githubRepo.GetConnectionByUserID(r.Context(), claims.UserID)
	if err != nil || conn == nil || repo.ConnectionID != conn.ID {
		h.respondError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	h.respondJSON(w, http.StatusOK, h.mapRepoResponse(repo))
}

// HandleScanRepo runs the scanner on a repo and stores results.
func (h *Handler) HandleScanRepo(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	repoID, err := h.parseUUID(r, "repoId")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid repo ID")
		return
	}

	repo, err := h.githubRepo.GetRepoByID(r.Context(), repoID)
	if err != nil || repo == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "Repo not found")
		return
	}

	conn, err := h.githubRepo.GetConnectionByUserID(r.Context(), claims.UserID)
	if err != nil || conn == nil || repo.ConnectionID != conn.ID {
		h.respondError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	ghClient, err := h.getGitHubClient(r.Context(), claims.UserID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "client_error", "Failed to connect to GitHub")
		return
	}

	scanner := githubsvc.NewScanner(ghClient, h.logger)
	result, err := scanner.ScanRepo(r.Context(), repo.Owner, repo.Name, repo.DefaultBranch)
	if err != nil {
		h.logger.WithError(err).Error("Failed to scan repo")
		h.respondError(w, http.StatusInternalServerError, "scan_error", "Failed to scan repository")
		return
	}

	functionsJSON, err := json.Marshal(result.Functions)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal scan results")
		h.respondError(w, http.StatusInternalServerError, "marshal_error", "Failed to process scan results")
		return
	}

	runtime := result.PrimaryRuntime
	if err := h.githubRepo.UpdateRepoScanResults(r.Context(), repoID, functionsJSON, runtime); err != nil {
		h.logger.WithError(err).Error("Failed to store scan results")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to store scan results")
		return
	}

	h.respondJSON(w, http.StatusOK, ScanRepoResponse{
		Functions:            toInterfaceSlice(result.Functions),
		PrimaryRuntime:       result.PrimaryRuntime,
		OverallConfidence:    result.OverallConfidence,
		StrategyUsed:         result.StrategyUsed,
		Warnings:             result.Warnings,
		EstimatedImportTimeS: result.EstimatedImportTimeS,
		EstimatedCostUSD:     result.EstimatedCostUSD,
	})
}

// HandleListBranches lists branches from GitHub for a repo.
func (h *Handler) HandleListBranches(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	repoID, err := h.parseUUID(r, "repoId")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid repo ID")
		return
	}

	repo, err := h.githubRepo.GetRepoByID(r.Context(), repoID)
	if err != nil || repo == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "Repo not found")
		return
	}

	conn, err := h.githubRepo.GetConnectionByUserID(r.Context(), claims.UserID)
	if err != nil || conn == nil || repo.ConnectionID != conn.ID {
		h.respondError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	ghClient, err := h.getGitHubClient(r.Context(), claims.UserID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "client_error", "Failed to connect to GitHub")
		return
	}

	branches, err := ghClient.ListBranches(r.Context(), repo.Owner, repo.Name)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list branches")
		h.respondError(w, http.StatusBadGateway, "github_error", "Failed to list branches")
		return
	}

	resp := make([]BranchResponse, len(branches))
	for i, b := range branches {
		resp[i] = BranchResponse{
			Name:      b.Name,
			SHA:       b.Commit.SHA,
			Protected: b.Protected,
		}
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// HandleGetTree returns the file tree of a repo.
func (h *Handler) HandleGetTree(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	repoID, err := h.parseUUID(r, "repoId")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid repo ID")
		return
	}

	repo, err := h.githubRepo.GetRepoByID(r.Context(), repoID)
	if err != nil || repo == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "Repo not found")
		return
	}

	conn, err := h.githubRepo.GetConnectionByUserID(r.Context(), claims.UserID)
	if err != nil || conn == nil || repo.ConnectionID != conn.ID {
		h.respondError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	ghClient, err := h.getGitHubClient(r.Context(), claims.UserID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "client_error", "Failed to connect to GitHub")
		return
	}

	branch := r.URL.Query().Get("branch")
	if branch == "" {
		branch = repo.DefaultBranch
	}

	branches, err := ghClient.ListBranches(r.Context(), repo.Owner, repo.Name)
	if err != nil {
		h.respondError(w, http.StatusBadGateway, "github_error", "Failed to list branches")
		return
	}

	var branchSHA string
	for _, b := range branches {
		if b.Name == branch {
			branchSHA = b.Commit.SHA
			break
		}
	}
	if branchSHA == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_branch", fmt.Sprintf("Branch %q not found", branch))
		return
	}

	recursive := r.URL.Query().Get("recursive") == "true"
	tree, err := ghClient.GetTree(r.Context(), repo.Owner, repo.Name, branchSHA, recursive)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get tree")
		h.respondError(w, http.StatusBadGateway, "github_error", "Failed to get file tree")
		return
	}

	nodes := make([]TreeNodeResponse, len(tree.Tree))
	for i, entry := range tree.Tree {
		nodes[i] = TreeNodeResponse{
			Path: entry.Path,
			Mode: entry.Mode,
			Type: entry.Type,
			SHA:  entry.SHA,
			Size: entry.Size,
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"sha":       tree.SHA,
		"truncated": tree.Truncated,
		"tree":      nodes,
	})
}

func (h *Handler) mapRepoResponse(repo *storage.GitHubRepo) *RepoResponse {
	return &RepoResponse{
		ID:                 repo.ID,
		FullName:           repo.FullName,
		Name:               repo.Name,
		Owner:              repo.Owner,
		Description:        repo.Description,
		DefaultBranch:      repo.DefaultBranch,
		Language:           repo.Language,
		Languages:          repo.Languages,
		IsPrivate:          repo.IsPrivate,
		IsFork:             repo.IsFork,
		IsArchived:         repo.IsArchived,
		Topics:             repo.Topics,
		StarsCount:         repo.StarsCount,
		ForksCount:         repo.ForksCount,
		HtmlURL:            repo.HtmlURL,
		DetectedFunctions:  repo.DetectedFunctions,
		DetectedRuntime:    repo.DetectedRuntime,
		HasFunctionflyJSON: repo.HasFunctionflyJSON,
		ImportStatus:       repo.ImportStatus,
		LastScannedAt:      repo.LastScannedAt,
	}
}

func toInterfaceSlice(v interface{}) []interface{} {
	b, _ := json.Marshal(v)
	var result []interface{}
	json.Unmarshal(b, &result)
	return result
}
