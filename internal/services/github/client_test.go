package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter(t *testing.T) {
	t.Run("new limiter allows requests", func(t *testing.T) {
		rl := NewRateLimiter()
		assert.True(t, rl.Allow())
		assert.True(t, rl.Allow())
	})

	t.Run("decrements remaining", func(t *testing.T) {
		rl := NewRateLimiter()
		rl.Allow()
		assert.Equal(t, 4999, rl.remaining)
	})

	t.Run("UpdateFromHeaders parses rate limit headers", func(t *testing.T) {
		rl := NewRateLimiter()
		headers := http.Header{}
		headers.Set("X-RateLimit-Remaining", "100")
		headers.Set("X-RateLimit-Limit", "5000")
		headers.Set("X-RateLimit-Reset", "1700000000")

		rl.UpdateFromHeaders(headers)
		assert.Equal(t, 100, rl.remaining)
		assert.Equal(t, 5000, rl.limit)
		assert.Equal(t, time.Unix(1700000000, 0), rl.resetAt)
	})

	t.Run("ignores empty headers", func(t *testing.T) {
		rl := NewRateLimiter()
		rl.remaining = 42
		rl.UpdateFromHeaders(http.Header{})
		assert.Equal(t, 42, rl.remaining)
	})

	t.Run("ignores invalid header values", func(t *testing.T) {
		rl := NewRateLimiter()
		headers := http.Header{}
		headers.Set("X-RateLimit-Remaining", "not-a-number")
		rl.UpdateFromHeaders(headers)
		assert.Equal(t, 5000, rl.remaining) // unchanged
	})
}

func TestNewClient(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		c := NewClient("test-token")
		assert.Equal(t, defaultBaseURL, c.baseURL)
		assert.Equal(t, "test-token", c.token)
		assert.Equal(t, defaultUserAgent, c.userAgent)
		assert.NotNil(t, c.httpClient)
		assert.NotNil(t, c.rateLimiter)
	})

	t.Run("custom base URL", func(t *testing.T) {
		c := NewClient("token", WithBaseURL("http://localhost:8080"))
		assert.Equal(t, "http://localhost:8080", c.baseURL)
	})

	t.Run("custom HTTP client", func(t *testing.T) {
		hc := &http.Client{Timeout: 5 * time.Second}
		c := NewClient("token", WithHTTPClient(hc))
		assert.Equal(t, hc, c.httpClient)
	})

	t.Run("custom logger", func(t *testing.T) {
		logger := logrus.New()
		c := NewClient("token", WithLogger(logger))
		assert.Equal(t, logger, c.logger)
	})
}

func TestClient_GetAuthenticatedUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/user", r.URL.Path)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         12345,
				"login":      "testuser",
				"avatar_url": "https://avatars.example.com/u/12345",
				"html_url":   "https://github.com/testuser",
				"name":       "Test User",
				"email":      "test@example.com",
			})
		}))
		defer server.Close()

		client := NewClient("test-token", WithBaseURL(server.URL))
		user, err := client.GetAuthenticatedUser(context.Background())
		require.NoError(t, err)
		assert.Equal(t, int64(12345), user.ID)
		assert.Equal(t, "testuser", user.Login)
		assert.Equal(t, "Test User", user.Name)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("server error retries", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			if requestCount < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    1,
				"login": "user",
			})
		}))
		defer server.Close()

		client := NewClient("token", WithBaseURL(server.URL))
		user, err := client.GetAuthenticatedUser(context.Background())
		require.NoError(t, err)
		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, 3, requestCount)
	})

	t.Run("max retries exceeded", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetAuthenticatedUser(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max retries exceeded")
	})

	t.Run("4xx error not retried", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "Bad credentials"}`))
		}))
		defer server.Close()

		client := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetAuthenticatedUser(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "401")
		assert.Equal(t, 1, requestCount, "should not retry 4xx errors")
	})
}

func TestClient_ListRepos(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/user/repos", r.URL.Path)
			assert.Equal(t, "2", r.URL.Query().Get("page"))
			assert.Equal(t, "50", r.URL.Query().Get("per_page"))

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"id":               1,
					"name":             "repo1",
					"full_name":        "user/repo1",
					"default_branch":   "main",
					"private":          false,
					"html_url":         "https://github.com/user/repo1",
					"clone_url":        "https://github.com/user/repo1.git",
					"ssh_url":          "git@github.com:user/repo1.git",
					"owner":            map[string]interface{}{"id": 1, "login": "user"},
				},
			})
		}))
		defer server.Close()

		client := NewClient("token", WithBaseURL(server.URL))
		repos, err := client.ListRepos(context.Background(), ListReposOptions{
			Page:    2,
			PerPage: 50,
		})
		require.NoError(t, err)
		require.Len(t, repos, 1)
		assert.Equal(t, "repo1", repos[0].Name)
		assert.Equal(t, "user/repo1", repos[0].FullName)
	})
}

func TestClient_GetRepo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/owner/repo", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":             42,
			"name":           "repo",
			"full_name":      "owner/repo",
			"default_branch": "main",
			"private":        true,
			"owner":          map[string]interface{}{"id": 1, "login": "owner"},
		})
	}))
	defer server.Close()

	client := NewClient("token", WithBaseURL(server.URL))
	repo, err := client.GetRepo(context.Background(), "owner", "repo")
	require.NoError(t, err)
	assert.Equal(t, int64(42), repo.ID)
	assert.Equal(t, "repo", repo.Name)
	assert.True(t, repo.Private)
}

func TestClient_ListBranches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/owner/repo/branches", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"name": "main",
				"commit": map[string]interface{}{
					"sha": "abc123",
				},
				"protected": true,
			},
			{
				"name": "develop",
				"commit": map[string]interface{}{
					"sha": "def456",
				},
				"protected": false,
			},
		})
	}))
	defer server.Close()

	client := NewClient("token", WithBaseURL(server.URL))
	branches, err := client.ListBranches(context.Background(), "owner", "repo")
	require.NoError(t, err)
	require.Len(t, branches, 2)
	assert.Equal(t, "main", branches[0].Name)
	assert.Equal(t, "abc123", branches[0].Commit.SHA)
	assert.True(t, branches[0].Protected)
	assert.Equal(t, "develop", branches[1].Name)
}

func TestClient_GetTree(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/owner/repo/git/trees/abc123", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("recursive"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sha": "abc123",
			"tree": []map[string]interface{}{
				{"path": "main.go", "mode": "100644", "type": "blob", "sha": "aaa", "size": 100},
				{"path": "src", "mode": "040000", "type": "tree", "sha": "bbb"},
			},
			"truncated": false,
		})
	}))
	defer server.Close()

	client := NewClient("token", WithBaseURL(server.URL))
	tree, err := client.GetTree(context.Background(), "owner", "repo", "abc123", true)
	require.NoError(t, err)
	assert.Equal(t, "abc123", tree.SHA)
	assert.Len(t, tree.Tree, 2)
	assert.Equal(t, "main.go", tree.Tree[0].Path)
	assert.Equal(t, "blob", tree.Tree[0].Type)
	assert.False(t, tree.Truncated)
}

func TestClient_CreateWebhook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/owner/repo/hooks", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req GitHubWebhookRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "web", req.Name)
		assert.True(t, req.Active)
		assert.Contains(t, req.Events, "push")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     999,
			"name":   req.Name,
			"active": true,
			"events": req.Events,
		})
	}))
	defer server.Close()

	client := NewClient("token", WithBaseURL(server.URL))
	whReq := &GitHubWebhookRequest{
		Name:   "web",
		Active: true,
		Events: []string{"push", "pull_request"},
	}
	whReq.Config.URL = "https://example.com/webhook"
	whReq.Config.ContentType = "json"
	whReq.Config.Secret = "secret"

	wh, err := client.CreateWebhook(context.Background(), "owner", "repo", whReq)
	require.NoError(t, err)
	assert.Equal(t, int64(999), wh.ID)
	assert.True(t, wh.Active)
}

func TestClient_DeleteWebhook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/owner/repo/hooks/999", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token", WithBaseURL(server.URL))
	err := client.DeleteWebhook(context.Background(), "owner", "repo", 999)
	assert.NoError(t, err)
}

func TestClient_CreateCommitStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/owner/repo/statuses/abc123", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req CommitStatusRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "success", req.State)
		assert.Equal(t, "functionfly/import", req.Context)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 1})
	}))
	defer server.Close()

	client := NewClient("token", WithBaseURL(server.URL))
	err := client.CreateCommitStatus(context.Background(), "owner", "repo", "abc123", &CommitStatusRequest{
		State:       "success",
		Description: "Imported as myFunc",
		Context:     "functionfly/import",
	})
	assert.NoError(t, err)
}

func TestClient_GetCompareDiff(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/owner/repo/compare/main...feature", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":       "ahead",
			"ahead_by":     3,
			"behind_by":    0,
			"total_commits": 3,
		})
	}))
	defer server.Close()

	client := NewClient("token", WithBaseURL(server.URL))
	result, err := client.GetCompareDiff(context.Background(), "owner", "repo", "main", "feature")
	require.NoError(t, err)
	assert.Equal(t, "ahead", result["status"])
}

func TestClient_GetFileContent(t *testing.T) {
	t.Run("base64 content", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/repos/owner/repo/contents/package.json", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name":     "package.json",
				"path":     "package.json",
				"sha":      "abc",
				"size":     20,
				"type":     "file",
				"encoding": "base64",
				"content":  "eyJ0ZXN0IjogdHJ1ZX0=", // {"test": true}
			})
		}))
		defer server.Close()

		client := NewClient("token", WithBaseURL(server.URL))
		content, err := client.GetFileContent(context.Background(), "owner", "repo", "package.json", "")
		require.NoError(t, err)
		assert.Equal(t, `{"test": true}`, string(content))
	})
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	client := NewClient("token", WithBaseURL(server.URL))
	_, err := client.GetAuthenticatedUser(ctx)
	assert.Error(t, err)
}

func TestClient_RateLimitHandling(t *testing.T) {
	t.Run("403 rate limit retries", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			if requestCount == 1 {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message": "API rate limit exceeded"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    1,
				"login": "user",
			})
		}))
		defer server.Close()

		client := NewClient("token", WithBaseURL(server.URL))
		user, err := client.GetAuthenticatedUser(context.Background())
		require.NoError(t, err)
		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, 2, requestCount)
	})

	t.Run("updates rate limiter from response headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-RateLimit-Remaining", "4999")
			w.Header().Set("X-RateLimit-Limit", "5000")
			w.Header().Set("X-RateLimit-Reset", "1700000000")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    1,
				"login": "user",
			})
		}))
		defer server.Close()

		client := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetAuthenticatedUser(context.Background())
		require.NoError(t, err)

		assert.Equal(t, 4999, client.rateLimiter.remaining)
		assert.Equal(t, 5000, client.rateLimiter.limit)
	})
}

func TestClient_GetLanguages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/owner/repo/languages", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{
			"Go":         8000,
			"JavaScript": 2000,
		})
	}))
	defer server.Close()

	client := NewClient("token", WithBaseURL(server.URL))
	languages, err := client.GetLanguages(context.Background(), "owner", "repo")
	require.NoError(t, err)
	assert.InDelta(t, 80.0, languages["Go"], 0.1)
	assert.InDelta(t, 20.0, languages["JavaScript"], 0.1)
}

func TestClient_GetRepoContent(t *testing.T) {
	t.Run("directory listing", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{"name": "file1.go", "path": "src/file1.go", "type": "file"},
				{"name": "file2.go", "path": "src/file2.go", "type": "file"},
			})
		}))
		defer server.Close()

		client := NewClient("token", WithBaseURL(server.URL))
		contents, err := client.GetRepoContent(context.Background(), "owner", "repo", "src", "main")
		require.NoError(t, err)
		assert.Len(t, contents, 2)
	})

	t.Run("single file", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name":     "file.go",
				"path":     "file.go",
				"type":     "file",
				"encoding": "base64",
				"content":  "cGFja2FnZSBtYWluCg==", // "package main\n"
			})
		}))
		defer server.Close()

		client := NewClient("token", WithBaseURL(server.URL))
		contents, err := client.GetRepoContent(context.Background(), "owner", "repo", "file.go", "")
		require.NoError(t, err)
		require.Len(t, contents, 1)
		assert.Equal(t, "file.go", contents[0].Name)
	})
}
