package storage

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGitHubRepositoryModels(t *testing.T) {
	t.Run("GitHubConnection defaults", func(t *testing.T) {
		conn := &GitHubConnection{
			ID:           uuid.New(),
			UserID:       uuid.New(),
			TenantID:     uuid.New(),
			GithubUserID: 12345,
			GithubUsername: "testuser",
			Status:       "active",
		}
		assert.NotEqual(t, uuid.Nil, conn.ID)
		assert.Equal(t, "active", conn.Status)
		assert.Equal(t, int64(12345), conn.GithubUserID)
	})

	t.Run("GitHubRepo JSONB fields", func(t *testing.T) {
		repo := &GitHubRepo{
			ID:                uuid.New(),
			ConnectionID:      uuid.New(),
			GithubRepoID:      67890,
			FullName:          "testuser/testrepo",
			Name:              "testrepo",
			Owner:             "testuser",
			DefaultBranch:     "main",
			Languages:         []byte(`{"Go": 80, "JavaScript": 20}`),
			Topics:            []byte(`["serverless", "go"]`),
			DetectedFunctions: []byte(`[{"name": "handler", "path": "main.go"}]`),
			ImportStatus:      "not_imported",
			Metadata:          []byte(`{}`),
		}
		assert.NotNil(t, repo.Languages)
		assert.NotNil(t, repo.Topics)
		assert.NotNil(t, repo.DetectedFunctions)
	})

	t.Run("GitHubImport status constants", func(t *testing.T) {
		imp := &GitHubImport{
			ID:           uuid.New(),
			Status:       "pending",
			Progress:     0,
			SourceBranch: "main",
			Visibility:   "private",
		}
		assert.Equal(t, "pending", imp.Status)
		assert.Equal(t, 0, imp.Progress)
	})

	t.Run("ListReposParams defaults", func(t *testing.T) {
		params := ListReposParams{}
		assert.Equal(t, 0, params.Page)
		assert.Equal(t, 0, params.PerPage)
	})

	t.Run("ListImportsParams with filters", func(t *testing.T) {
		repoID := uuid.New()
		params := ListImportsParams{
			Page:    1,
			PerPage: 25,
			Status:  "completed",
			RepoID:  &repoID,
		}
		assert.Equal(t, "completed", params.Status)
		assert.NotNil(t, params.RepoID)
	})

	t.Run("nullIfEmpty helper", func(t *testing.T) {
		assert.Nil(t, nullIfEmpty(""))
		s := nullIfEmpty("hello")
		assert.NotNil(t, s)
		assert.Equal(t, "hello", s)
	})
}
