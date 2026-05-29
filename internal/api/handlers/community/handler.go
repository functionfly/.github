package community

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	comm "github.com/functionfly/functionfly/internal/community"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler serves community forum API endpoints.
type Handler struct {
	repo   *storage.CommunityRepository
	logger *logrus.Logger
}

// NewHandler creates a community handler.
func NewHandler(repo *storage.CommunityRepository, logger *logrus.Logger) *Handler {
	if logger == nil {
		logger = logrus.New()
	}
	return &Handler{repo: repo, logger: logger}
}

// ListCategories handles GET /community/categories
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.repo.ListCategories(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("list community categories failed")
		http.Error(w, `{"error":"Failed to list categories"}`, http.StatusInternalServerError)
		return
	}
	if cats == nil {
		cats = []comm.Category{}
	}
	writeJSON(w, map[string]any{"categories": cats})
}

// ListPosts handles GET /community/posts
func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	opts := storage.ListPostsOptions{
		CategorySlug: r.URL.Query().Get("category"),
		Sort:         r.URL.Query().Get("sort"),
		Query:        strings.TrimSpace(r.URL.Query().Get("q")),
	}
	if opts.Sort == "" {
		opts.Sort = "hot"
	}
	opts.Limit, opts.Offset = parsePagination(r)

	if user := middleware.GetUserFromContext(r); user != nil {
		opts.ViewerID = &user.UserID
	}

	posts, err := h.repo.ListPosts(r.Context(), opts)
	if err != nil {
		h.logger.WithError(err).Error("list community posts failed")
		http.Error(w, `{"error":"Failed to list posts"}`, http.StatusInternalServerError)
		return
	}
	if posts == nil {
		posts = []comm.PostListItem{}
	}
	writeJSON(w, map[string]any{"posts": posts, "limit": opts.Limit, "offset": opts.Offset})
}

// GetPost handles GET /community/posts/{id}
func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid post ID"}`, http.StatusBadRequest)
		return
	}

	var viewerID *uuid.UUID
	if user := middleware.GetUserFromContext(r); user != nil {
		viewerID = &user.UserID
	}

	post, err := h.repo.GetPostByID(r.Context(), id, viewerID)
	if err != nil {
		h.logger.WithError(err).Error("get community post failed")
		http.Error(w, `{"error":"Failed to load post"}`, http.StatusInternalServerError)
		return
	}
	if post == nil {
		http.Error(w, `{"error":"Post not found"}`, http.StatusNotFound)
		return
	}

	_ = h.repo.IncrementPostViews(r.Context(), id)

	comments, err := h.repo.ListCommentsForPost(r.Context(), id, viewerID)
	if err != nil {
		h.logger.WithError(err).Error("list community comments failed")
		http.Error(w, `{"error":"Failed to load comments"}`, http.StatusInternalServerError)
		return
	}
	if comments == nil {
		comments = []comm.CommentWithAuthor{}
	}

	writeJSON(w, map[string]any{"post": post, "comments": comments})
}

type createPostRequest struct {
	CategorySlug string   `json:"category_slug"`
	Title        string   `json:"title"`
	Body         string   `json:"body"`
	Tags         []string `json:"tags"`
}

// CreatePost handles POST /community/posts
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	req.CategorySlug = strings.TrimSpace(req.CategorySlug)

	if req.Title == "" || req.Body == "" || req.CategorySlug == "" {
		http.Error(w, `{"error":"category_slug, title, and body are required"}`, http.StatusBadRequest)
		return
	}
	if len(req.Title) > 300 {
		http.Error(w, `{"error":"Title must be 300 characters or less"}`, http.StatusBadRequest)
		return
	}
	if len(req.Body) > 10000 {
		http.Error(w, `{"error":"Body must be 10000 characters or less"}`, http.StatusBadRequest)
		return
	}

	cat, err := h.repo.GetCategoryBySlug(r.Context(), req.CategorySlug)
	if err != nil || cat == nil {
		http.Error(w, `{"error":"Invalid category"}`, http.StatusBadRequest)
		return
	}

	post := &comm.Post{
		CategoryID: cat.ID,
		AuthorID:   user.UserID,
		Title:      req.Title,
		Body:       req.Body,
		Tags:       req.Tags,
		Status:     comm.StatusOpen,
	}
	if post.Tags == nil {
		post.Tags = []string{}
	}

	if err := h.repo.CreatePost(r.Context(), post); err != nil {
		h.logger.WithError(err).Error("create community post failed")
		http.Error(w, `{"error":"Failed to create post"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

type createCommentRequest struct {
	Body     string     `json:"body"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

// CreateComment handles POST /community/posts/{id}/comments
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	postID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid post ID"}`, http.StatusBadRequest)
		return
	}

	post, err := h.repo.GetPostByID(r.Context(), postID, nil)
	if err != nil || post == nil {
		http.Error(w, `{"error":"Post not found"}`, http.StatusNotFound)
		return
	}
	if post.Status == comm.StatusLocked {
		http.Error(w, `{"error":"This thread is locked"}`, http.StatusForbidden)
		return
	}

	var req createCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	req.Body = strings.TrimSpace(req.Body)
	if req.Body == "" {
		http.Error(w, `{"error":"body is required"}`, http.StatusBadRequest)
		return
	}
	if len(req.Body) > 8000 {
		http.Error(w, `{"error":"Comment must be 8000 characters or less"}`, http.StatusBadRequest)
		return
	}

	if req.ParentID != nil {
		parent, err := h.repo.GetCommentByID(r.Context(), *req.ParentID)
		if err != nil || parent == nil || parent.PostID != postID {
			http.Error(w, `{"error":"Invalid parent comment"}`, http.StatusBadRequest)
			return
		}
	}

	comment := &comm.Comment{
		PostID:   postID,
		ParentID: req.ParentID,
		AuthorID: user.UserID,
		Body:     req.Body,
	}

	if err := h.repo.CreateComment(r.Context(), comment); err != nil {
		h.logger.WithError(err).Error("create community comment failed")
		http.Error(w, `{"error":"Failed to create comment"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

type voteRequest struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Value      int    `json:"value"`
}

// Vote handles POST /community/votes
func (h *Handler) Vote(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	var req voteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	targetType := comm.VoteTargetType(req.TargetType)
	if targetType != comm.VoteTargetPost && targetType != comm.VoteTargetComment {
		http.Error(w, `{"error":"target_type must be post or comment"}`, http.StatusBadRequest)
		return
	}

	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		http.Error(w, `{"error":"Invalid target_id"}`, http.StatusBadRequest)
		return
	}

	score, err := h.repo.UpsertVote(r.Context(), user.UserID, targetType, targetID, req.Value)
	if err != nil {
		h.logger.WithError(err).Error("community vote failed")
		http.Error(w, `{"error":"Failed to record vote"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{"vote_score": score, "user_vote": req.Value})
}

// AcceptComment handles POST /community/posts/{id}/accept/{comment_id}
func (h *Handler) AcceptComment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	postID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid post ID"}`, http.StatusBadRequest)
		return
	}
	commentID, err := parseUUID(mux.Vars(r)["comment_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid comment ID"}`, http.StatusBadRequest)
		return
	}

	post, err := h.repo.GetPostByID(r.Context(), postID, nil)
	if err != nil || post == nil {
		http.Error(w, `{"error":"Post not found"}`, http.StatusNotFound)
		return
	}
	if post.AuthorID != user.UserID {
		http.Error(w, `{"error":"Only the thread author can accept an answer"}`, http.StatusForbidden)
		return
	}

	comment, err := h.repo.GetCommentByID(r.Context(), commentID)
	if err != nil || comment == nil || comment.PostID != postID {
		http.Error(w, `{"error":"Comment not found"}`, http.StatusNotFound)
		return
	}

	if err := h.repo.AcceptComment(r.Context(), postID, commentID); err != nil {
		h.logger.WithError(err).Error("accept community comment failed")
		http.Error(w, `{"error":"Failed to accept answer"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "solved"})
}

// RegisterRoutes registers community forum routes on the API router.
func (h *Handler) RegisterRoutes(api *mux.Router, auth *middleware.AuthMiddleware) {
	api.HandleFunc("/community/categories", h.ListCategories).Methods("GET", "OPTIONS")
	api.HandleFunc("/community/posts", h.ListPosts).Methods("GET", "OPTIONS")
	api.HandleFunc("/community/posts/{id}", h.GetPost).Methods("GET", "OPTIONS")
	api.HandleFunc("/community/posts", auth.RequireAuth(h.CreatePost)).Methods("POST", "OPTIONS")
	api.HandleFunc("/community/posts/{id}/comments", auth.RequireAuth(h.CreateComment)).Methods("POST", "OPTIONS")
	api.HandleFunc("/community/votes", auth.RequireAuth(h.Vote)).Methods("POST", "OPTIONS")
	api.HandleFunc("/community/posts/{id}/accept/{comment_id}", auth.RequireAuth(h.AcceptComment)).Methods("POST", "OPTIONS")
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func parsePagination(r *http.Request) (limit, offset int) {
	limit = 20
	offset = 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
