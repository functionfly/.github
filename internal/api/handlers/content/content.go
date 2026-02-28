package content

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler contains content management handlers
type Handler struct {
	repo          storage.Repository
	githubService *services.GitHubService
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// NewHandler creates a new content handler
func NewHandler(repo storage.Repository) *Handler {
	// Initialize with proper config from environment
	githubOwner := getEnvOrDefault("GITHUB_OWNER", "functionfly")
	githubRepo := getEnvOrDefault("GITHUB_REPO", "functionfly")
	githubToken := getEnvOrDefault("GITHUB_TOKEN", "")

	githubService := services.NewGitHubService(githubOwner, githubRepo, githubToken)

	return &Handler{
		repo:          repo,
		githubService: githubService,
	}
}

// Changelog Handlers

// HandleListChangelogEntries lists changelog entries
func (h *Handler) HandleListChangelogEntries(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	publishedOnlyStr := r.URL.Query().Get("published_only")

	limit := 50 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	publishedOnly := false
	if publishedOnlyStr == "true" {
		publishedOnly = true
	}

	entries, err := h.repo.ListChangelogEntries(limit, offset, publishedOnly)
	if err != nil {
		logrus.WithError(err).Error("Failed to list changelog entries")
		http.Error(w, "Failed to list changelog entries", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
		"limit":   limit,
		"offset":  offset,
	})
}

// HandleGetChangelogEntry gets a specific changelog entry
func (h *Handler) HandleGetChangelogEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entryIDStr := vars["entryId"]

	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		http.Error(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	entry, err := h.repo.GetChangelogEntryByID(entryID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get changelog entry")
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Changelog entry not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get changelog entry", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// HandleCreateChangelogEntry creates a new changelog entry
func (h *Handler) HandleCreateChangelogEntry(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Version     string                    `json:"version"`
		Date        time.Time                 `json:"date"`
		Type        string                    `json:"type"`
		Title       string                    `json:"title"`
		Description string                    `json:"description"`
		ReleaseURL  *string                   `json:"release_url,omitempty"`
		GitHubID    *string                   `json:"github_id,omitempty"`
		IsPublished bool                      `json:"is_published"`
		Changes     []storage.ChangelogChange `json:"changes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Version == "" || req.Title == "" || req.Type == "" {
		http.Error(w, "Version, title, and type are required", http.StatusBadRequest)
		return
	}

	// Validate type
	if req.Type != "major" && req.Type != "minor" && req.Type != "patch" {
		http.Error(w, "Type must be major, minor, or patch", http.StatusBadRequest)
		return
	}

	entry := &storage.ChangelogEntry{
		Version:     req.Version,
		Date:        req.Date,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		ReleaseURL:  req.ReleaseURL,
		GitHubID:    req.GitHubID,
		IsPublished: req.IsPublished,
		Changes:     req.Changes,
	}

	created, err := h.repo.CreateChangelogEntry(r.Context(), entry)
	if err != nil {
		logrus.WithError(err).Error("Failed to create changelog entry")
		http.Error(w, "Failed to create changelog entry", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// HandleUpdateChangelogEntry updates a changelog entry
func (h *Handler) HandleUpdateChangelogEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entryIDStr := vars["entryId"]

	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		http.Error(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate type if provided
	if typeVal, ok := updates["type"]; ok {
		if typeStr, ok := typeVal.(string); ok {
			if typeStr != "major" && typeStr != "minor" && typeStr != "patch" {
				http.Error(w, "Type must be major, minor, or patch", http.StatusBadRequest)
				return
			}
		}
	}

	updated, err := h.repo.UpdateChangelogEntry(r.Context(), entryID, updates)
	if err != nil {
		logrus.WithError(err).Error("Failed to update changelog entry")
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Changelog entry not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update changelog entry", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// HandleDeleteChangelogEntry deletes a changelog entry
func (h *Handler) HandleDeleteChangelogEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entryIDStr := vars["entryId"]

	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		http.Error(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteChangelogEntry(r.Context(), entryID); err != nil {
		logrus.WithError(err).Error("Failed to delete changelog entry")
		http.Error(w, "Failed to delete changelog entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleCreateChangelogChange creates a new changelog change
func (h *Handler) HandleCreateChangelogChange(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entryIDStr := vars["entryId"]

	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		http.Error(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Category string   `json:"category"`
		Icon     string   `json:"icon"`
		Items    []string `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Category == "" || req.Icon == "" || len(req.Items) == 0 {
		http.Error(w, "Category, icon, and items are required", http.StatusBadRequest)
		return
	}

	change := &storage.ChangelogChange{
		EntryID:  entryID,
		Category: req.Category,
		Icon:     req.Icon,
		Items:    req.Items,
	}

	created, err := h.repo.CreateChangelogChange(r.Context(), change)
	if err != nil {
		logrus.WithError(err).Error("Failed to create changelog change")
		http.Error(w, "Failed to create changelog change", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// HandleUpdateChangelogChange updates a changelog change
func (h *Handler) HandleUpdateChangelogChange(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	changeIDStr := vars["changeId"]

	changeID, err := uuid.Parse(changeIDStr)
	if err != nil {
		http.Error(w, "Invalid change ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updated, err := h.repo.UpdateChangelogChange(r.Context(), changeID, updates)
	if err != nil {
		logrus.WithError(err).Error("Failed to update changelog change")
		http.Error(w, "Failed to update changelog change", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// HandleDeleteChangelogChange deletes a changelog change
func (h *Handler) HandleDeleteChangelogChange(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	changeIDStr := vars["changeId"]

	changeID, err := uuid.Parse(changeIDStr)
	if err != nil {
		http.Error(w, "Invalid change ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteChangelogChange(r.Context(), changeID); err != nil {
		logrus.WithError(err).Error("Failed to delete changelog change")
		http.Error(w, "Failed to delete changelog change", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Blog Handlers

// HandleListBlogPosts lists blog posts
func (h *Handler) HandleListBlogPosts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	publishedOnlyStr := r.URL.Query().Get("published_only")
	tagStr := r.URL.Query().Get("tags")

	limit := 50 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	publishedOnly := false
	if publishedOnlyStr == "true" {
		publishedOnly = true
	}

	var tagFilter []string
	if tagStr != "" {
		tagFilter = strings.Split(tagStr, ",")
		// Trim spaces
		for i, tag := range tagFilter {
			tagFilter[i] = strings.TrimSpace(tag)
		}
	}

	posts, err := h.repo.ListBlogPosts(limit, offset, publishedOnly, tagFilter)
	if err != nil {
		logrus.WithError(err).Error("Failed to list blog posts")
		http.Error(w, "Failed to list blog posts", http.StatusInternalServerError)
		return
	}

	// Ensure posts is never nil for JSON encoding
	if posts == nil {
		posts = []*storage.BlogPost{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts":  posts,
		"limit":  limit,
		"offset": offset,
	})
}

// HandleGetBlogPost gets a specific blog post
func (h *Handler) HandleGetBlogPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr := vars["postId"]

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := h.repo.GetBlogPostByID(postID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get blog post")
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Blog post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get blog post", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// HandleGetBlogPostBySlug gets a blog post by slug
func (h *Handler) HandleGetBlogPostBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		http.Error(w, "Slug is required", http.StatusBadRequest)
		return
	}

	post, err := h.repo.GetBlogPostBySlug(slug)
	if err != nil {
		logrus.WithError(err).Error("Failed to get blog post by slug")
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Blog post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get blog post", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// HandleCreateBlogPost creates a new blog post
func (h *Handler) HandleCreateBlogPost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title         string     `json:"title"`
		Slug          string     `json:"slug"`
		Content       string     `json:"content"`
		Excerpt       string     `json:"excerpt"`
		Author        string     `json:"author"`
		Tags          []string   `json:"tags"`
		FeaturedImage *string    `json:"featured_image,omitempty"`
		SanityID      *string    `json:"sanity_id,omitempty"`
		IsPublished   bool       `json:"is_published"`
		PublishedAt   *time.Time `json:"published_at,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" || req.Slug == "" || req.Content == "" || req.Author == "" {
		http.Error(w, "Title, slug, content, and author are required", http.StatusBadRequest)
		return
	}

	post := &storage.BlogPost{
		Title:         req.Title,
		Slug:          req.Slug,
		Content:       req.Content,
		Excerpt:       req.Excerpt,
		Author:        req.Author,
		Tags:          req.Tags,
		FeaturedImage: req.FeaturedImage,
		SanityID:      req.SanityID,
		IsPublished:   req.IsPublished,
		PublishedAt:   req.PublishedAt,
	}

	created, err := h.repo.CreateBlogPost(r.Context(), post)
	if err != nil {
		logrus.WithError(err).Error("Failed to create blog post")
		http.Error(w, "Failed to create blog post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// HandleUpdateBlogPost updates a blog post
func (h *Handler) HandleUpdateBlogPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr := vars["postId"]

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updated, err := h.repo.UpdateBlogPost(r.Context(), postID, updates)
	if err != nil {
		logrus.WithError(err).Error("Failed to update blog post")
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Blog post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update blog post", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// HandleDeleteBlogPost deletes a blog post
func (h *Handler) HandleDeleteBlogPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr := vars["postId"]

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteBlogPost(r.Context(), postID); err != nil {
		logrus.WithError(err).Error("Failed to delete blog post")
		http.Error(w, "Failed to delete blog post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Public endpoints for frontend consumption

// HandleGetPublishedChangelogEntries returns published changelog entries for frontend
func (h *Handler) HandleGetPublishedChangelogEntries(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default for frontend
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	entries, err := h.repo.ListChangelogEntries(limit, 0, true)
	if err != nil {
		logrus.WithError(err).Error("Failed to get published changelog entries")
		http.Error(w, "Failed to get changelog entries", http.StatusInternalServerError)
		return
	}

	// Ensure entries is never nil for JSON encoding
	if entries == nil {
		entries = []*storage.ChangelogEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
	})
}

// HandleGetPublishedBlogPosts returns published blog posts for frontend
func (h *Handler) HandleGetPublishedBlogPosts(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	tagStr := r.URL.Query().Get("tags")

	limit := 10 // default for frontend
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var tagFilter []string
	if tagStr != "" {
		for _, t := range strings.Split(tagStr, ",") {
			if s := strings.TrimSpace(t); s != "" {
				tagFilter = append(tagFilter, s)
			}
		}
	}

	posts, err := h.repo.ListBlogPosts(limit, offset, true, tagFilter)
	if err != nil {
		logrus.WithError(err).Error("Failed to list published blog posts")
		http.Error(w, "Failed to load blog posts", http.StatusInternalServerError)
		return
	}
	if posts == nil {
		posts = []*storage.BlogPost{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts":  posts,
		"limit":  limit,
		"offset": offset,
	})
}

// HandleGetPublishedBlogPostBySlug returns a published blog post by slug for frontend
func (h *Handler) HandleGetPublishedBlogPostBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		http.Error(w, "Slug is required", http.StatusBadRequest)
		return
	}

	post, err := h.repo.GetBlogPostBySlug(slug)
	if err != nil {
		logrus.WithError(err).Error("Failed to get blog post by slug")
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Blog post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get blog post", http.StatusInternalServerError)
		}
		return
	}

	// Only return published posts
	if !post.IsPublished {
		http.Error(w, "Blog post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// HandleGetBlogCategories returns unique categories (tags) from published blog posts
func (h *Handler) HandleGetBlogCategories(w http.ResponseWriter, r *http.Request) {
	// TEMPORARY: Return mock categories
	categories := []string{"test", "blog", "tutorial", "news"}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories": categories,
	})
}

// HandleGetBlogAuthors returns unique authors from published blog posts
func (h *Handler) HandleGetBlogAuthors(w http.ResponseWriter, r *http.Request) {
	// TEMPORARY: Return mock authors
	authors := []string{"Test Author", "John Doe", "Jane Smith"}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"authors": authors,
	})
}

// HandleSyncGitHubReleases syncs GitHub releases with changelog entries
func (h *Handler) HandleSyncGitHubReleases(w http.ResponseWriter, r *http.Request) {
	if err := h.githubService.SyncReleases(r.Context(), h.repo); err != nil {
		logrus.WithError(err).Error("Failed to sync GitHub releases")
		http.Error(w, "Failed to sync GitHub releases", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "GitHub releases synced successfully",
	})
}
