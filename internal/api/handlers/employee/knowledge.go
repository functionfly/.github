package employee

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListArticles(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListKnowledgeOpts{
		Limit:  20,
		Offset: 0,
	}
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			opts.Limit = n
		}
	}
	if o := q.Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			opts.Offset = n
		}
	}
	if s := q.Get("category"); s != "" {
		opts.Category = &s
	}
	if s := q.Get("status"); s != "" {
		opts.Status = &s
	}

	articles, total, err := h.repo.ListKnowledgeArticles(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list articles")
		apierror.WriteError(w, apierror.NewInternal("Failed to list articles"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"articles": articles,
		"total":    total,
		"limit":    opts.Limit,
		"offset":   opts.Offset,
	})
}

func (h *Handler) HandleGetArticle(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	slug := mux.Vars(r)["slug"]
	if slug == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Article slug is required"))
		return
	}

	article, err := h.repo.GetKnowledgeArticleBySlug(r.Context(), claims.TenantID, slug)
	if err != nil {
		h.log.WithError(err).Error("Failed to get article")
		apierror.WriteError(w, apierror.NewInternal("Failed to get article"))
		return
	}
	if article == nil {
		apierror.WriteError(w, apierror.NewNotFound("Article not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"article": article,
	})
}

type createArticleRequest struct {
	Title    string   `json:"title"`
	Body     string   `json:"body"`
	Slug     string   `json:"slug,omitempty"`
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Status   string   `json:"status,omitempty"`
}

func (h *Handler) HandleCreateArticle(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req createArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	if req.Title == "" || req.Body == "" {
		apierror.WriteError(w, apierror.NewBadRequest("title and body are required"))
		return
	}

	now := time.Now()
	slug := req.Slug
	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(req.Title, " ", "-"))
		slug = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
				return r
			}
			return -1
		}, slug) + "-" + uuid.New().String()[:8]
	}

	// Resolve employee ID from user ID
	emp, _ := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	authorID := claims.UserID
	if emp != nil {
		authorID = emp.ID
	}

	status := req.Status
	if status == "" {
		status = "published"
	}

	article := &types.KnowledgeArticle{
		ID:        uuid.New(),
		TenantID:  claims.TenantID,
		Title:     req.Title,
		Slug:      slug,
		Body:      req.Body,
		AuthorID:  authorID,
		Status:    status,
		Tags:      types.JSONMap{"tags": req.Tags},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if req.Category != "" {
		article.Category = &req.Category
	}
	if status == "published" {
		article.PublishedAt = &now
	}

	created, err := h.repo.CreateKnowledgeArticle(r.Context(), article)
	if err != nil {
		h.log.WithError(err).Error("Failed to create article")
		apierror.WriteError(w, apierror.NewInternal("Failed to create article"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"article": created,
	})
}

type updateArticleRequest struct {
	Title *string  `json:"title,omitempty"`
	Body  *string  `json:"body,omitempty"`
	Tags  []string `json:"tags,omitempty"`
	Draft *bool    `json:"draft,omitempty"`
}

func (h *Handler) HandleUpdateArticle(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid article ID"))
		return
	}

	var req updateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Body != nil {
		updates["body"] = *req.Body
	}
	if req.Tags != nil {
		updates["tags"] = types.JSONMap{"tags": req.Tags}
	}
	if req.Draft != nil {
		if *req.Draft {
			updates["status"] = "draft"
		} else {
			updates["status"] = "published"
			now := time.Now()
			updates["published_at"] = now
		}
	}
	updates["updated_at"] = time.Now()

	if len(updates) == 1 {
		apierror.WriteError(w, apierror.NewBadRequest("No fields to update"))
		return
	}

	if err := h.repo.UpdateKnowledgeArticle(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to update article")
		apierror.WriteError(w, apierror.NewInternal("Failed to update article"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) HandleSearchKnowledge(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Search query 'q' is required"))
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	articles, err := h.repo.SearchKnowledgeArticles(r.Context(), claims.TenantID, q, limit)
	if err != nil {
		h.log.WithError(err).Error("Failed to search knowledge articles")
		apierror.WriteError(w, apierror.NewInternal("Failed to search knowledge articles"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": articles,
		"query":   q,
	})
}
