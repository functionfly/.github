package content

import (
	"bytes"
	"context"
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
	contentRepo   *storage.ContentRepository
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
func NewHandler(repo storage.Repository, contentRepo *storage.ContentRepository) *Handler {
	// Initialize with proper config from environment
	githubOwner := getEnvOrDefault("GITHUB_OWNER", "functionfly")
	githubRepo := getEnvOrDefault("GITHUB_REPO", "functionfly")
	githubToken := getEnvOrDefault("GITHUB_TOKEN", "")

	githubService := services.NewGitHubService(githubOwner, githubRepo, githubToken)

	return &Handler{
		repo:          repo,
		contentRepo:   contentRepo,
		githubService: githubService,
	}
}

// Changelog Handlers

// HandleListChangelogEntries lists changelog entries
func (h *Handler) HandleListChangelogEntries(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
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

	entries, err := h.repo.ListChangelogEntries(ctx, limit, offset, publishedOnly)
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
	ctx := r.Context()
	vars := mux.Vars(r)
	entryIDStr := vars["entryId"]

	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		http.Error(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	entry, err := h.repo.GetChangelogEntryByID(ctx, entryID)
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
	ctx := r.Context()
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

	posts, err := h.repo.ListBlogPosts(ctx, limit, offset, publishedOnly, tagFilter)
	if err != nil {
		logrus.WithError(err).Error("Failed to list blog posts")
		// Return empty list instead of 500 to prevent UI crashes during auth state changes
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"posts":  []*storage.BlogPost{},
			"limit":  limit,
			"offset": offset,
		})
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
	ctx := r.Context()
	vars := mux.Vars(r)
	postIDStr := vars["postId"]

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := h.repo.GetBlogPostByID(ctx, postID)
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
	ctx := r.Context()
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		http.Error(w, "Slug is required", http.StatusBadRequest)
		return
	}

	post, err := h.repo.GetBlogPostBySlug(ctx, slug)
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
	ctx := r.Context()
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default for frontend
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	entries, err := h.repo.ListChangelogEntries(ctx, limit, 0, true)
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
	ctx := r.Context()
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

	posts, err := h.repo.ListBlogPosts(ctx, limit, offset, true, tagFilter)
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
	ctx := r.Context()
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		http.Error(w, "Slug is required", http.StatusBadRequest)
		return
	}

	post, err := h.repo.GetBlogPostBySlug(ctx, slug)
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

// HandleGetBlogCategories returns blog categories for the public blog page (same data as admin list, read-only).
func (h *Handler) HandleGetBlogCategories(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.ListBlogCategories(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to list blog categories")
		http.Error(w, "Failed to list categories", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []*storage.BlogCategory{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// HandleGetBlogAuthors returns blog authors for the public blog page (same data as admin list, read-only).
func (h *Handler) HandleGetBlogAuthors(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.ListBlogAuthors(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to list blog authors")
		http.Error(w, "Failed to list authors", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []*storage.BlogAuthor{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// Admin categories (CRUD)

// HandleListAdminCategories returns GET /v1/admin/content/categories
func (h *Handler) HandleListAdminCategories(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.ListBlogCategories(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to list blog categories")
		http.Error(w, "Failed to list blog categories", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []*storage.BlogCategory{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// HandleCreateAdminCategory returns POST /v1/admin/content/categories
func (h *Handler) HandleCreateAdminCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		Color       string `json:"color"`
		Icon        string `json:"icon"`
		Order       int    `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		// Derive slug from title: lowercase, spaces to hyphens, drop non-alnum except hyphen
		b := make([]byte, 0, len(req.Title))
		for _, r := range strings.TrimSpace(req.Title) {
			if r >= 'A' && r <= 'Z' {
				b = append(b, byte(r+'a'-'A'))
			} else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				b = append(b, byte(r))
			} else if r == ' ' || r == '-' {
				if len(b) > 0 && b[len(b)-1] != '-' {
					b = append(b, '-')
				}
			}
		}
		slug = strings.Trim(string(b), "-")
		if slug == "" {
			slug = "category"
		}
	}
	c := &storage.BlogCategory{
		Title:       strings.TrimSpace(req.Title),
		Slug:        slug,
		Description: strings.TrimSpace(req.Description),
		Color:       strings.TrimSpace(req.Color),
		Icon:        strings.TrimSpace(req.Icon),
		Order:       req.Order,
	}
	created, err := h.repo.CreateBlogCategory(r.Context(), c)
	if err != nil {
		logrus.WithError(err).Error("Failed to create blog category")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":  "Failed to create blog category",
			"detail": err.Error(),
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// HandleGetAdminCategory returns GET /v1/admin/content/categories/{id}
func (h *Handler) HandleGetAdminCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	c, err := h.repo.GetBlogCategoryByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			logrus.WithError(err).Error("Failed to get blog category")
			http.Error(w, "Failed to get blog category", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// HandleUpdateAdminCategory returns PATCH /v1/admin/content/categories/{id}
func (h *Handler) HandleUpdateAdminCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	updated, err := h.repo.UpdateBlogCategory(r.Context(), id, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			logrus.WithError(err).Error("Failed to update blog category")
			http.Error(w, "Failed to update blog category", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// HandleDeleteAdminCategory returns DELETE /v1/admin/content/categories/{id}
func (h *Handler) HandleDeleteAdminCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	if err := h.repo.DeleteBlogCategory(r.Context(), id); err != nil {
		logrus.WithError(err).Error("Failed to delete blog category")
		http.Error(w, "Failed to delete blog category", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Admin authors (CRUD)

// HandleListAdminAuthors returns GET /v1/admin/content/authors
func (h *Handler) HandleListAdminAuthors(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.ListBlogAuthors(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to list blog authors")
		http.Error(w, "Failed to list blog authors", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []*storage.BlogAuthor{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// HandleCreateAdminAuthor returns POST /v1/admin/content/authors
func (h *Handler) HandleCreateAdminAuthor(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string                 `json:"name"`
		Slug        string                 `json:"slug"`
		Bio         string                 `json:"bio"`
		Photo       map[string]interface{} `json:"photo"`
		Email       string                 `json:"email"`
		Website     string                 `json:"website"`
		SocialLinks map[string]interface{} `json:"social_links"`
		Role        string                 `json:"role"`
		Active      bool                   `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Slug) == "" {
		http.Error(w, "Name and slug are required", http.StatusBadRequest)
		return
	}
	a := &storage.BlogAuthor{
		Name:        strings.TrimSpace(req.Name),
		Slug:        strings.TrimSpace(req.Slug),
		Bio:         strings.TrimSpace(req.Bio),
		Photo:       req.Photo,
		Email:       strings.TrimSpace(req.Email),
		Website:     strings.TrimSpace(req.Website),
		SocialLinks: req.SocialLinks,
		Role:        strings.TrimSpace(req.Role),
		Active:      req.Active,
	}
	created, err := h.repo.CreateBlogAuthor(r.Context(), a)
	if err != nil {
		logrus.WithError(err).Error("Failed to create blog author")
		http.Error(w, "Failed to create blog author", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// HandleGetAdminAuthor returns GET /v1/admin/content/authors/{id}
func (h *Handler) HandleGetAdminAuthor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid author ID", http.StatusBadRequest)
		return
	}
	a, err := h.repo.GetBlogAuthorByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Author not found", http.StatusNotFound)
		} else {
			logrus.WithError(err).Error("Failed to get blog author")
			http.Error(w, "Failed to get blog author", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a)
}

// HandleUpdateAdminAuthor returns PATCH /v1/admin/content/authors/{id}
func (h *Handler) HandleUpdateAdminAuthor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid author ID", http.StatusBadRequest)
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	updated, err := h.repo.UpdateBlogAuthor(r.Context(), id, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Author not found", http.StatusNotFound)
		} else {
			logrus.WithError(err).Error("Failed to update blog author")
			http.Error(w, "Failed to update blog author", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// HandleDeleteAdminAuthor returns DELETE /v1/admin/content/authors/{id}
func (h *Handler) HandleDeleteAdminAuthor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid author ID", http.StatusBadRequest)
		return
	}
	if err := h.repo.DeleteBlogAuthor(r.Context(), id); err != nil {
		logrus.WithError(err).Error("Failed to delete blog author")
		http.Error(w, "Failed to delete blog author", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

const openRouterFreeModel = "arcee-ai/trinity-large-preview:free"
const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"

// callOpenRouter sends a prompt to Open Router and returns the trimmed response content.
func callOpenRouter(ctx context.Context, prompt string, maxTokens int) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return "", nil
	}
	reqBody := map[string]interface{}{
		"model": openRouterFreeModel,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": maxTokens,
	}
	encoded, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", openRouterURL, bytes.NewReader(encoded))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	var openResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&openResp); err != nil {
		return "", err
	}
	if len(openResp.Choices) == 0 {
		return "", nil
	}
	return strings.TrimSpace(openResp.Choices[0].Message.Content), nil
}

// HandleGenerateChangelogContent returns POST /v1/admin/content/generate/changelog
// Uses Open Router to generate title and description from version/type/topic.
func (h *Handler) HandleGenerateChangelogContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		http.Error(w, "Open Router API key not configured (OPENROUTER_API_KEY)", http.StatusServiceUnavailable)
		return
	}

	var body struct {
		Version string `json:"version"`
		Type    string `json:"type"`
		Topic   string `json:"topic"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	body.Version = strings.TrimSpace(body.Version)
	body.Type = strings.TrimSpace(body.Type)
	if body.Type == "" {
		body.Type = "minor"
	}

	prompt := "Generate a changelog entry title and a one- or two-sentence description."
	if body.Version != "" {
		prompt += " Version: " + body.Version + "."
	}
	prompt += " Release type: " + body.Type + "."
	if body.Topic != "" {
		prompt += " Topic or focus: " + body.Topic + "."
	}
	prompt += " Output exactly two lines: line 1 = title only, line 2 = description only. No labels or extra text."

	out, err := callOpenRouter(r.Context(), prompt, 200)
	if err != nil {
		logrus.WithError(err).Error("Open Router request failed for changelog generate")
		http.Error(w, "Open Router request failed", http.StatusBadGateway)
		return
	}
	title, description := "", ""
	lines := strings.SplitN(out, "\n", 2)
	if len(lines) >= 1 {
		title = strings.TrimSpace(lines[0])
	}
	if len(lines) >= 2 {
		description = strings.TrimSpace(lines[1])
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"title": title, "description": description})
}

// HandleGenerateBlogContent returns POST /v1/admin/content/generate/blog
// Uses Open Router to generate title, description, and excerpt from topic.
func (h *Handler) HandleGenerateBlogContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		http.Error(w, "Open Router API key not configured (OPENROUTER_API_KEY)", http.StatusServiceUnavailable)
		return
	}

	var body struct {
		Topic string `json:"topic"`
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	body.Topic = strings.TrimSpace(body.Topic)
	body.Title = strings.TrimSpace(body.Title)
	if body.Topic == "" && body.Title == "" {
		http.Error(w, "topic or title is required", http.StatusBadRequest)
		return
	}

	prompt := "Generate a blog post title, a one-sentence meta description, and a short excerpt (2-3 sentences)."
	if body.Topic != "" {
		prompt += " Topic: " + body.Topic + "."
	}
	if body.Title != "" {
		prompt += " Optional existing title: " + body.Title + "."
	}
	prompt += " Output exactly three lines: line 1 = title only, line 2 = meta description only, line 3 = excerpt only. No labels or extra text."

	out, err := callOpenRouter(r.Context(), prompt, 300)
	if err != nil {
		logrus.WithError(err).Error("Open Router request failed for blog generate")
		http.Error(w, "Open Router request failed", http.StatusBadGateway)
		return
	}
	title, description, excerpt := "", "", ""
	lines := strings.SplitN(out, "\n", 3)
	if len(lines) >= 1 {
		title = strings.TrimSpace(lines[0])
	}
	if len(lines) >= 2 {
		description = strings.TrimSpace(lines[1])
	}
	if len(lines) >= 3 {
		excerpt = strings.TrimSpace(lines[2])
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"title":       title,
		"description": description,
		"excerpt":     excerpt,
	})
}

// HandleGenerateAuthorContent returns POST /v1/admin/content/generate/author
// Uses Open Router to generate a short author bio from name/role.
func (h *Handler) HandleGenerateAuthorContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		http.Error(w, "Open Router API key not configured (OPENROUTER_API_KEY)", http.StatusServiceUnavailable)
		return
	}

	var body struct {
		Name string `json:"name"`
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	body.Role = strings.TrimSpace(body.Role)
	if body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	prompt := "Write a short professional author bio (2-3 sentences) for: " + body.Name + "."
	if body.Role != "" {
		prompt += " Role: " + body.Role + "."
	}
	prompt += " Output only the bio text, no quotes or prefix."

	out, err := callOpenRouter(r.Context(), prompt, 150)
	if err != nil {
		logrus.WithError(err).Error("Open Router request failed for author generate")
		http.Error(w, "Open Router request failed", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"bio": strings.TrimSpace(out)})
}

// HandleGenerateCategoryContent returns POST /v1/admin/content/generate/category
// Uses Open Router to generate a category description from title.
func (h *Handler) HandleGenerateCategoryContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		http.Error(w, "Open Router API key not configured (OPENROUTER_API_KEY)", http.StatusServiceUnavailable)
		return
	}

	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	body.Title = strings.TrimSpace(body.Title)
	if body.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	prompt := "Write a short one- or two-sentence description for a blog category titled: " + body.Title + ". Output only the description, no quotes or prefix."

	out, err := callOpenRouter(r.Context(), prompt, 100)
	if err != nil {
		logrus.WithError(err).Error("Open Router request failed for category generate")
		http.Error(w, "Open Router request failed", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"description": strings.TrimSpace(out)})
}

// HandleGetBlogSettings returns GET /v1/admin/content/blog/settings
func (h *Handler) HandleGetBlogSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.repo.GetBlogSettings(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get blog settings")
		http.Error(w, "Failed to get blog settings", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// HandleUpdateBlogSettings returns PATCH /v1/admin/content/blog/settings
func (h *Handler) HandleUpdateBlogSettings(w http.ResponseWriter, r *http.Request) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	settings, err := h.repo.UpdateBlogSettings(r.Context(), updates)
	if err != nil {
		logrus.WithError(err).Error("Failed to update blog settings")
		http.Error(w, "Failed to update blog settings", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// Blog Analytics Handlers

// HandleRecordBlogView records a page view for a blog post
// POST /v1/content/blog/{postId}/view
func (h *Handler) HandleRecordBlogView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIDStr := vars["postId"]
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var req struct {
		VisitorID  string `json:"visitor_id"`
		Referrer   string `json:"referrer"`
		Country    string `json:"country"`
		City       string `json:"city"`
		DeviceType string `json:"device_type"`
		Browser    string `json:"browser"`
		OS         string `json:"os"`
	}
	// Request body is optional for public tracking
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}

	view := &storage.BlogPageView{
		PostID:     postID,
		VisitorID: req.VisitorID,
		Referrer:   req.Referrer,
		UserAgent: r.UserAgent(),
		IPAddress: r.RemoteAddr,
		Country:    req.Country,
		City:       req.City,
		DeviceType: req.DeviceType,
		Browser:    req.Browser,
		OS:         req.OS,
	}

	if err := h.contentRepo.RecordBlogPageView(r.Context(), view); err != nil {
		logrus.WithError(err).Error("Failed to record blog view")
		http.Error(w, "Failed to record view", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetBlogAnalyticsSummary returns blog analytics summary
// GET /v1/admin/content/analytics/summary
func (h *Handler) HandleGetBlogAnalyticsSummary(w http.ResponseWriter, r *http.Request) {
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	summary, err := h.contentRepo.GetBlogAnalyticsSummary(r.Context(), days)
	if err != nil {
		logrus.WithError(err).Error("Failed to get blog analytics summary")
		http.Error(w, "Failed to get analytics summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// HandleGetBlogViewsTimeSeries returns views over time
// GET /v1/admin/content/analytics/timeseries
func (h *Handler) HandleGetBlogViewsTimeSeries(w http.ResponseWriter, r *http.Request) {
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	series, err := h.contentRepo.GetBlogViewsTimeSeries(r.Context(), days)
	if err != nil {
		logrus.WithError(err).Error("Failed to get blog views time series")
		http.Error(w, "Failed to get time series", http.StatusInternalServerError)
		return
	}

	if series == nil {
		series = []storage.BlogViewsTimeSeries{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(series)
}

// HandleGetTopBlogPosts returns top performing blog posts
// GET /v1/admin/content/analytics/top-posts
func (h *Handler) HandleGetTopBlogPosts(w http.ResponseWriter, r *http.Request) {
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	posts, err := h.contentRepo.GetTopBlogPosts(r.Context(), days, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get top blog posts")
		http.Error(w, "Failed to get top posts", http.StatusInternalServerError)
		return
	}

	if posts == nil {
		posts = []storage.TopBlogPost{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
