package blog

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

type Handler struct {
	repo *BlogRepository
}

func NewHandler(repo *BlogRepository) *Handler {
	return &Handler{repo: repo}
}

type BlogPost struct {
	ID            uuid.UUID              `json:"id"`
	Title         string                 `json:"title"`
	Slug          string                 `json:"slug"`
	Description   string                 `json:"description"`
	Body          interface{}            `json:"body"`
	AuthorID      *uuid.UUID            `json:"authorId,omitempty"`
	CategoryID    *uuid.UUID            `json:"categoryId,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	HeroImage     *HeroImage             `json:"heroImage,omitempty"`
	Status        string                 `json:"status"`
	PublishedAt   *time.Time            `json:"publishedAt,omitempty"`
	ScheduledAt   *time.Time            `json:"publishedAt,omitempty"`
	UpdatedAt     *time.Time            `json:"updatedAt,omitempty"`
	CreatedAt     time.Time             `json:"createdAt"`
	SEOTitle      *string               `json:"seoTitle,omitempty"`
	SEODescription *string              `json:"seoDescription,omitempty"`
	Keywords      []string               `json:"keywords,omitempty"`
	CanonicalURL  *string               `json:"canonicalUrl,omitempty"`
	OGImage       *OGImage              `json:"ogImage,omitempty"`
	Campaign      *string               `json:"campaign,omitempty"`
	OwnerID       *uuid.UUID            `json:"ownerId,omitempty"`
	Author        *AuthorSummary        `json:"author,omitempty"`
	Category      *CategorySummary      `json:"category,omitempty"`
}

type HeroImage struct {
	URL     string `json:"url"`
	Alt     string `json:"alt"`
	Caption *string `json:"caption,omitempty"`
}

type OGImage struct {
	URL string `json:"url"`
	Alt string `json:"alt"`
}

type AuthorSummary struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Author struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Bio         string                 `json:"bio,omitempty"`
	Email       string                 `json:"email,omitempty"`
	Website     string                 `json:"website,omitempty"`
	Photo       map[string]interface{} `json:"photo,omitempty"`
	SocialLinks []SocialLink           `json:"socialLinks,omitempty"`
	Role        string                 `json:"role,omitempty"`
	Active      bool                   `json:"active"`
	CreatedAt   time.Time             `json:"createdAt"`
	UpdatedAt   time.Time             `json:"updatedAt"`
}

type SocialLink struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
}

type Category struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color,omitempty"`
	Icon        string    `json:"icon,omitempty"`
	Order       int       `json:"order"`
	PostCount   int       `json:"postCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CategorySummary struct {
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

type BlogSettings struct {
	ID               uuid.UUID `json:"id"`
	BlogTitle        string    `json:"blogTitle"`
	PostsPerPage     int       `json:"postsPerPage"`
	MetaDescription  string    `json:"metaDescription"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type BlogRelatedPost struct {
	ID            uuid.UUID `json:"id"`
	PostID        uuid.UUID `json:"postId"`
	RelatedPostID uuid.UUID `json:"relatedPostId"`
}

type BlogCTABlock struct {
	ID          uuid.UUID `json:"id"`
	PostID      uuid.UUID `json:"postId"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	ButtonText  string    `json:"buttonText"`
	ButtonURL   string    `json:"buttonUrl"`
	Style       string    `json:"style"`
	Order       int       `json:"order"`
}

type BlogListResponse struct {
	Data []BlogPost `json:"data"`
	Meta ListMeta   `json:"meta"`
}

type AdminBlogListResponse struct {
	Data []AdminBlogPost `json:"data"`
	Meta ListMeta        `json:"meta"`
}

type AdminBlogPost struct {
	BlogPost
	Author  *Author  `json:"author,omitempty"`
	Category *Category `json:"category,omitempty"`
}

type ListMeta struct {
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	TotalPages int    `json:"totalPages"`
	Search     string `json:"search,omitempty"`
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	// Public routes
	r.HandleFunc("/blog/posts", h.HandleListPosts).Methods("GET")
	r.HandleFunc("/blog/posts/{slug}", h.HandleGetPostBySlug).Methods("GET")
	r.HandleFunc("/blog/categories", h.HandleGetCategories).Methods("GET")
	r.HandleFunc("/blog/authors", h.HandleGetAuthors).Methods("GET")

	// Admin routes (protected)
	admin := r.PathPrefix("/blog").Subrouter()
	admin.HandleFunc("/posts", h.HandleCreatePost).Methods("POST")
	admin.HandleFunc("/posts/{id}", h.HandleUpdatePost).Methods("PUT")
	admin.HandleFunc("/posts/{id}", h.HandleDeletePost).Methods("DELETE")
	admin.HandleFunc("/categories", h.HandleCreateCategory).Methods("POST")
	admin.HandleFunc("/categories/{id}", h.HandleUpdateCategory).Methods("PUT")
	admin.HandleFunc("/categories/{id}", h.HandleDeleteCategory).Methods("DELETE")
	admin.HandleFunc("/authors", h.HandleCreateAuthor).Methods("POST")
	admin.HandleFunc("/authors/{id}", h.HandleUpdateAuthor).Methods("PUT")
	admin.HandleFunc("/authors/{id}", h.HandleDeleteAuthor).Methods("DELETE")
}

// HandleListPosts returns GET /blog/posts
func (h *Handler) HandleListPosts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 10
	}
	status := query.Get("status")
	category := query.Get("category")
	author := query.Get("author")
	tag := query.Get("tag")
	search := query.Get("search")

	offset := (page - 1) * limit

	posts, total, err := h.repo.ListPosts(BlogPostFilter{
		Status:   status,
		Category: category,
		Author:   author,
		Tag:      tag,
		Search:   search,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		logrus.WithError(err).Error("Failed to list blog posts")
		apierror.WriteError(w, apierror.NewInternal("Failed to list posts"))
		return
	}

	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	response := BlogListResponse{
		Data: posts,
		Meta: ListMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}
	if search != "" {
		response.Meta.Search = search
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetPostBySlug returns GET /blog/posts/{slug}
func (h *Handler) HandleGetPostBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Slug is required"))
		return
	}

	post, err := h.repo.GetPostBySlug(slug)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			apierror.WriteError(w, apierror.NewNotFound("Post not found"))
		} else {
			logrus.WithError(err).Error("Failed to get blog post")
			apierror.WriteError(w, apierror.NewInternal("Failed to get post"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// HandleGetCategories returns GET /blog/categories
func (h *Handler) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.ListCategories()
	if err != nil {
		logrus.WithError(err).Error("Failed to list categories")
		apierror.WriteError(w, apierror.NewInternal("Failed to list categories"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// HandleGetAuthors returns GET /blog/authors
func (h *Handler) HandleGetAuthors(w http.ResponseWriter, r *http.Request) {
	authors, err := h.repo.ListAuthors()
	if err != nil {
		logrus.WithError(err).Error("Failed to list authors")
		apierror.WriteError(w, apierror.NewInternal("Failed to list authors"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authors)
}

// HandleListPostsAdmin returns GET /admin/blog/posts
func (h *Handler) HandleListPostsAdmin(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 20
	}
	status := query.Get("status")
	category := query.Get("category")
	author := query.Get("author")
	tag := query.Get("tag")
	search := query.Get("search")

	offset := (page - 1) * limit

	posts, total, err := h.repo.ListPosts(BlogPostFilter{
		Status:   status,
		Category: category,
		Author:   author,
		Tag:      tag,
		Search:   search,
		Limit:    limit,
		Offset:   offset,
		Admin:    true,
	})
	if err != nil {
		logrus.WithError(err).Error("Failed to list blog posts")
		apierror.WriteError(w, apierror.NewInternal("Failed to list posts"))
		return
	}

	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	response := BlogListResponse{
		Data: posts,
		Meta: ListMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}
	if search != "" {
		response.Meta.Search = search
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleCreatePost returns POST /blog/posts
func (h *Handler) HandleCreatePost(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title         string                 `json:"title"`
		Slug          string                 `json:"slug"`
		Description   string                 `json:"description"`
Body          json.RawMessage       `json:"body"`
		AuthorID      *uuid.UUID            `json:"authorId"`
		CategoryID    *uuid.UUID            `json:"categoryId"`
		Tags          []string               `json:"tags"`
		HeroImage     *HeroImage             `json:"heroImage"`
		Status        string                 `json:"status"`
		PublishedAt   *string               `json:"publishedAt"`
		ScheduledAt   *string               `json:"scheduledAt"`
		SEOTitle      *string               `json:"seoTitle"`
		SEODescription *string              `json:"seoDescription"`
		Keywords      []string               `json:"keywords"`
		CanonicalURL  *string               `json:"canonicalUrl"`
		Campaign      *string               `json:"campaign"`
		OwnerID       *uuid.UUID            `json:"ownerId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	if input.Title == "" || input.Slug == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Title and slug are required"))
		return
	}

	if input.Status == "" {
		input.Status = "draft"
	}

	post := &BlogPost{
		Title:         input.Title,
		Slug:          input.Slug,
		Description:   input.Description,
		AuthorID:      input.AuthorID,
		CategoryID:    input.CategoryID,
		Tags:          input.Tags,
		HeroImage:     input.HeroImage,
		Status:        input.Status,
		SEOTitle:      input.SEOTitle,
		SEODescription: input.SEODescription,
		Keywords:      input.Keywords,
		CanonicalURL:  input.CanonicalURL,
		Campaign:      input.Campaign,
		OwnerID:       input.OwnerID,
	}

	if input.Body != nil {
		if err := json.Unmarshal(input.Body, &post.Body); err != nil {
			apierror.WriteError(w, apierror.NewBadRequest("Invalid body JSON"))
			return
		}
	}

	if input.PublishedAt != nil {
		t, _ := time.Parse(time.RFC3339, *input.PublishedAt)
		post.PublishedAt = &t
	}
	if input.ScheduledAt != nil {
		t, _ := time.Parse(time.RFC3339, *input.ScheduledAt)
		post.ScheduledAt = &t
	}

	created, err := h.repo.CreatePost(post)
	if err != nil {
		logrus.WithError(err).Error("Failed to create blog post")
		apierror.WriteError(w, apierror.NewInternal("Failed to create post"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// HandleUpdatePost returns PUT /blog/posts/{id}
func (h *Handler) HandleUpdatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid post ID"))
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	updated, err := h.repo.UpdatePost(id, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			apierror.WriteError(w, apierror.NewNotFound("Post not found"))
		} else {
			logrus.WithError(err).Error("Failed to update blog post")
			apierror.WriteError(w, apierror.NewInternal("Failed to update post"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// HandleDeletePost returns DELETE /blog/posts/{id}
func (h *Handler) HandleDeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid post ID"))
		return
	}

	if err := h.repo.DeletePost(id); err != nil {
		logrus.WithError(err).Error("Failed to delete blog post")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete post"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleCreateCategory returns POST /blog/categories
func (h *Handler) HandleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       string `json:"title"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		Color       string `json:"color"`
		Icon        string `json:"icon"`
		Order       int    `json:"order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	if input.Title == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Title is required"))
		return
	}

	category := &Category{
		Title:       input.Title,
		Slug:        input.Slug,
		Description: input.Description,
		Color:       input.Color,
		Icon:        input.Icon,
		Order:       input.Order,
	}

	created, err := h.repo.CreateCategory(category)
	if err != nil {
		logrus.WithError(err).Error("Failed to create category")
		apierror.WriteError(w, apierror.NewInternal("Failed to create category"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// HandleUpdateCategory returns PUT /blog/categories/{id}
func (h *Handler) HandleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid category ID"))
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	updated, err := h.repo.UpdateCategory(id, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			apierror.WriteError(w, apierror.NewNotFound("Category not found"))
		} else {
			logrus.WithError(err).Error("Failed to update category")
			apierror.WriteError(w, apierror.NewInternal("Failed to update category"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// HandleDeleteCategory returns DELETE /blog/categories/{id}
func (h *Handler) HandleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid category ID"))
		return
	}

	if err := h.repo.DeleteCategory(id); err != nil {
		logrus.WithError(err).Error("Failed to delete category")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete category"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleCreateAuthor returns POST /blog/authors
func (h *Handler) HandleCreateAuthor(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string                 `json:"name"`
		Slug        string                 `json:"slug"`
		Bio         string                 `json:"bio"`
		Email       string                 `json:"email"`
		Website     string                 `json:"website"`
		Photo       map[string]interface{} `json:"photo"`
		SocialLinks []SocialLink           `json:"socialLinks"`
		Role        string                 `json:"role"`
		Active      bool                   `json:"active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	if input.Name == "" || input.Slug == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Name and slug are required"))
		return
	}

	author := &Author{
		Name:        input.Name,
		Slug:        input.Slug,
		Bio:         input.Bio,
		Email:       input.Email,
		Website:     input.Website,
		Photo:       input.Photo,
		SocialLinks: input.SocialLinks,
		Role:        input.Role,
		Active:      input.Active,
	}

	created, err := h.repo.CreateAuthor(author)
	if err != nil {
		logrus.WithError(err).Error("Failed to create author")
		apierror.WriteError(w, apierror.NewInternal("Failed to create author"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// HandleUpdateAuthor returns PUT /blog/authors/{id}
func (h *Handler) HandleUpdateAuthor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid author ID"))
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	updated, err := h.repo.UpdateAuthor(id, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			apierror.WriteError(w, apierror.NewNotFound("Author not found"))
		} else {
			logrus.WithError(err).Error("Failed to update author")
			apierror.WriteError(w, apierror.NewInternal("Failed to update author"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// HandleDeleteAuthor returns DELETE /blog/authors/{id}
func (h *Handler) HandleDeleteAuthor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid author ID"))
		return
	}

	if err := h.repo.DeleteAuthor(id); err != nil {
		logrus.WithError(err).Error("Failed to delete author")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete author"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetPostAdmin returns GET /admin/blog/posts/{id}
func (h *Handler) HandleGetPostAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid post ID"))
		return
	}

	post, err := h.repo.GetPostByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			apierror.WriteError(w, apierror.NewNotFound("Post not found"))
		} else {
			logrus.WithError(err).Error("Failed to get blog post")
			apierror.WriteError(w, apierror.NewInternal("Failed to get post"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// HandleGetAuthorBySlug returns GET /admin/blog/authors/slug/{slug}
func (h *Handler) HandleGetAuthorBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Slug is required"))
		return
	}

	author, err := h.repo.GetAuthorBySlug(slug)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			apierror.WriteError(w, apierror.NewNotFound("Author not found"))
		} else {
			logrus.WithError(err).Error("Failed to get author")
			apierror.WriteError(w, apierror.NewInternal("Failed to get author"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(author)
}

// HandleGetCategoryBySlug returns GET /admin/blog/categories/slug/{slug}
func (h *Handler) HandleGetCategoryBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Slug is required"))
		return
	}

	category, err := h.repo.GetCategoryBySlug(slug)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			apierror.WriteError(w, apierror.NewNotFound("Category not found"))
		} else {
			logrus.WithError(err).Error("Failed to get category")
			apierror.WriteError(w, apierror.NewInternal("Failed to get category"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// HandleGetSettings returns GET /admin/blog/settings
func (h *Handler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.repo.GetSettings()
	if err != nil {
		logrus.WithError(err).Error("Failed to get blog settings")
		apierror.WriteError(w, apierror.NewInternal("Failed to get settings"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// HandleUpdateSettings returns PUT /admin/blog/settings
func (h *Handler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	settings, err := h.repo.UpdateSettings(updates)
	if err != nil {
		logrus.WithError(err).Error("Failed to update blog settings")
		apierror.WriteError(w, apierror.NewInternal("Failed to update settings"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// HandleGetRelatedPosts returns GET /admin/blog/posts/{id}/related
func (h *Handler) HandleGetRelatedPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid post ID"))
		return
	}

	posts, err := h.repo.GetRelatedPosts(id)
	if err != nil {
		logrus.WithError(err).Error("Failed to get related posts")
		apierror.WriteError(w, apierror.NewInternal("Failed to get related posts"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// HandleSetRelatedPosts returns PUT /admin/blog/posts/{id}/related
func (h *Handler) HandleSetRelatedPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid post ID"))
		return
	}

	var input struct {
		RelatedPostIDs []uuid.UUID `json:"relatedPostIds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	if err := h.repo.SetRelatedPosts(id, input.RelatedPostIDs); err != nil {
		logrus.WithError(err).Error("Failed to set related posts")
		apierror.WriteError(w, apierror.NewInternal("Failed to set related posts"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetCTABlocks returns GET /admin/blog/posts/{id}/cta-blocks
func (h *Handler) HandleGetCTABlocks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid post ID"))
		return
	}

	blocks, err := h.repo.GetCTABlocks(id)
	if err != nil {
		logrus.WithError(err).Error("Failed to get CTA blocks")
		apierror.WriteError(w, apierror.NewInternal("Failed to get CTA blocks"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blocks)
}

// HandleSetCTABlocks returns PUT /admin/blog/posts/{id}/cta-blocks
func (h *Handler) HandleSetCTABlocks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid post ID"))
		return
	}

	var input struct {
		Blocks []BlogCTABlock `json:"blocks"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	if err := h.repo.SetCTABlocks(id, input.Blocks); err != nil {
		logrus.WithError(err).Error("Failed to set CTA blocks")
		apierror.WriteError(w, apierror.NewInternal("Failed to set CTA blocks"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
