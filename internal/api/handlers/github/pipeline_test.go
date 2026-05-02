package github

import (
	"testing"

	"github.com/functionfly/functionfly/internal/services/github"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestDetectRuntime(t *testing.T) {
	h := &Handler{}

	t.Run("node18 with package.json", func(t *testing.T) {
		tree := []github.GitHubTreeEntry{
			{Path: "package.json", Type: "blob"},
			{Path: "index.js", Type: "blob"},
		}
		repo := &storage.GitHubRepo{}
		assert.Equal(t, "node18", h.detectRuntime(tree, repo))
	})

	t.Run("node18-typescript with tsconfig.json", func(t *testing.T) {
		tree := []github.GitHubTreeEntry{
			{Path: "package.json", Type: "blob"},
			{Path: "tsconfig.json", Type: "blob"},
			{Path: "index.ts", Type: "blob"},
		}
		repo := &storage.GitHubRepo{}
		assert.Equal(t, "node18-typescript", h.detectRuntime(tree, repo))
	})

	t.Run("python3.11 with requirements.txt", func(t *testing.T) {
		tree := []github.GitHubTreeEntry{
			{Path: "requirements.txt", Type: "blob"},
			{Path: "main.py", Type: "blob"},
		}
		repo := &storage.GitHubRepo{}
		assert.Equal(t, "python3.11", h.detectRuntime(tree, repo))
	})

	t.Run("python3.11 with pyproject.toml", func(t *testing.T) {
		tree := []github.GitHubTreeEntry{
			{Path: "pyproject.toml", Type: "blob"},
			{Path: "main.py", Type: "blob"},
		}
		repo := &storage.GitHubRepo{}
		assert.Equal(t, "python3.11", h.detectRuntime(tree, repo))
	})

	t.Run("python3.11 with setup.py", func(t *testing.T) {
		tree := []github.GitHubTreeEntry{
			{Path: "setup.py", Type: "blob"},
		}
		repo := &storage.GitHubRepo{}
		assert.Equal(t, "python3.11", h.detectRuntime(tree, repo))
	})

	t.Run("python3.11 with Pipfile", func(t *testing.T) {
		tree := []github.GitHubTreeEntry{
			{Path: "Pipfile", Type: "blob"},
		}
		repo := &storage.GitHubRepo{}
		assert.Equal(t, "python3.11", h.detectRuntime(tree, repo))
	})

	t.Run("go1.22 with go.mod", func(t *testing.T) {
		tree := []github.GitHubTreeEntry{
			{Path: "go.mod", Type: "blob"},
			{Path: "main.go", Type: "blob"},
		}
		repo := &storage.GitHubRepo{}
		assert.Equal(t, "go1.22", h.detectRuntime(tree, repo))
	})

	t.Run("rust1.75 with Cargo.toml", func(t *testing.T) {
		tree := []github.GitHubTreeEntry{
			{Path: "Cargo.toml", Type: "blob"},
			{Path: "src/main.rs", Type: "blob"},
		}
		repo := &storage.GitHubRepo{}
		assert.Equal(t, "rust1.75", h.detectRuntime(tree, repo))
	})

	t.Run("defaults to node18 for unknown", func(t *testing.T) {
		tree := []github.GitHubTreeEntry{
			{Path: "README.md", Type: "blob"},
			{Path: "Makefile", Type: "blob"},
		}
		repo := &storage.GitHubRepo{}
		assert.Equal(t, "node18", h.detectRuntime(tree, repo))
	})

	t.Run("empty tree defaults to node18", func(t *testing.T) {
		tree := []github.GitHubTreeEntry{}
		repo := &storage.GitHubRepo{}
		assert.Equal(t, "node18", h.detectRuntime(tree, repo))
	})

	t.Run("skips non-blob entries", func(t *testing.T) {
		tree := []github.GitHubTreeEntry{
			{Path: "src", Type: "tree"},
			{Path: "src/package.json", Type: "blob"},
		}
		repo := &storage.GitHubRepo{}
		// package.json is nested, hasFile checks for "/"+name suffix
		assert.Equal(t, "node18", h.detectRuntime(tree, repo))
	})

	t.Run("detects nested go.mod", func(t *testing.T) {
		tree := []github.GitHubTreeEntry{
			{Path: "backend/go.mod", Type: "blob"},
			{Path: "backend/main.go", Type: "blob"},
		}
		repo := &storage.GitHubRepo{}
		assert.Equal(t, "go1.22", h.detectRuntime(tree, repo))
	})
}

func TestComputeContentHash(t *testing.T) {
	t.Run("consistent hash for same input", func(t *testing.T) {
		tree := []github.GitHubTreeEntry{
			{Path: "main.go", Type: "blob", SHA: "abc123", Size: 100},
			{Path: "go.mod", Type: "blob", SHA: "def456", Size: 50},
		}
		hash1 := computeContentHash(tree)
		hash2 := computeContentHash(tree)
		assert.Equal(t, hash1, hash2)
		assert.NotEmpty(t, hash1)
		assert.Len(t, hash1, 64, "SHA-256 hex should be 64 chars")
	})

	t.Run("different trees produce different hashes", func(t *testing.T) {
		tree1 := []github.GitHubTreeEntry{
			{Path: "main.go", Type: "blob", SHA: "abc123", Size: 100},
		}
		tree2 := []github.GitHubTreeEntry{
			{Path: "main.go", Type: "blob", SHA: "xyz789", Size: 100},
		}
		assert.NotEqual(t, computeContentHash(tree1), computeContentHash(tree2))
	})

	t.Run("skips non-blob entries", func(t *testing.T) {
		treeWithDir := []github.GitHubTreeEntry{
			{Path: "src", Type: "tree", SHA: "aaa", Size: 0},
			{Path: "main.go", Type: "blob", SHA: "abc123", Size: 100},
		}
		treeWithoutDir := []github.GitHubTreeEntry{
			{Path: "main.go", Type: "blob", SHA: "abc123", Size: 100},
		}
		assert.Equal(t, computeContentHash(treeWithDir), computeContentHash(treeWithoutDir))
	})

	t.Run("empty tree", func(t *testing.T) {
		hash := computeContentHash([]github.GitHubTreeEntry{})
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64)
	})
}

func TestMustJSON(t *testing.T) {
	t.Run("serializes map", func(t *testing.T) {
		result := mustJSON(map[string]interface{}{
			"key": "value",
			"num": 42,
		})
		assert.Contains(t, result, `"key":"value"`)
		assert.Contains(t, result, `"num":42`)
	})

	t.Run("serializes nil", func(t *testing.T) {
		result := mustJSON(nil)
		assert.Equal(t, "null", result)
	})

	t.Run("serializes slice", func(t *testing.T) {
		result := mustJSON([]string{"a", "b"})
		assert.Equal(t, `["a","b"]`, result)
	})
}
