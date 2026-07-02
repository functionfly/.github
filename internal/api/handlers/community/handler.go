package community

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	comm "github.com/functionfly/functionfly/internal/community"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/lib/pq"
)

// Handler serves community forum API endpoints.
type Handler struct {
	repo    *storage.CommunityRepository
	logger  *logrus.Logger
	reputer *services.ReputationHooker
}

// NewHandler creates a community handler.
func NewHandler(repo *storage.CommunityRepository, logger *logrus.Logger, reputer *services.ReputationHooker) *Handler {
	if logger == nil {
		logger = logrus.New()
	}
	return &Handler{repo: repo, logger: logger, reputer: reputer}
}

// ListCategories handles GET /community/categories
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.repo.ListCategoriesWithCounts(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("list community categories failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to list categories"))
		return
	}
	if cats == nil {
		cats = []comm.CategoryWithCount{}
	}
	writeJSON(w, map[string]any{"categories": cats})
}

// ListPosts handles GET /community/posts
func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	opts := storage.ListPostsOptions{
		CategorySlug: r.URL.Query().Get("category"),
		Sort:         r.URL.Query().Get("sort"),
		Query:        strings.TrimSpace(r.URL.Query().Get("q")),
		TagFilter:    r.URL.Query().Get("tag"),
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
		apierror.WriteError(w, apierror.NewInternal("Failed to list posts"))
		return
	}
	if posts == nil {
		posts = []comm.PostListItem{}
	}

	total, err := h.repo.CountPosts(r.Context(), opts)
	if err != nil {
		h.logger.WithError(err).Warn("count community posts failed")
		total = len(posts)
	}

	writeJSON(w, map[string]any{"posts": posts, "total": total, "limit": opts.Limit, "offset": opts.Offset})
}

// GetPost handles GET /community/posts/{idOrSlug}
func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	idOrSlug := mux.Vars(r)["id"]

	var viewerID *uuid.UUID
	if user := middleware.GetUserFromContext(r); user != nil {
		viewerID = &user.UserID
	}

	var post *comm.PostDetail
	var postID uuid.UUID
	var err error

	// Try UUID first, fall back to slug
	if id, parseErr := parseUUID(idOrSlug); parseErr == nil {
		postID = id
		post, err = h.repo.GetPostByID(r.Context(), id, viewerID)
	} else {
		post, err = h.repo.GetPostBySlug(r.Context(), idOrSlug, viewerID)
		if post != nil {
			postID = post.ID
		}
	}
	if err != nil {
		h.logger.WithError(err).Error("get community post failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to load post"))
		return
	}
	if post == nil {
		apierror.WriteError(w, apierror.NewNotFound("Post not found"))
		return
	}

	_ = h.repo.IncrementPostViews(r.Context(), postID)

	comments, err := h.repo.ListCommentsForPost(r.Context(), postID, viewerID)
	if err != nil {
		h.logger.WithError(err).Error("list community comments failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to load comments"))
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
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	req.CategorySlug = strings.TrimSpace(req.CategorySlug)

	if req.Title == "" || req.Body == "" || req.CategorySlug == "" {
		apierror.WriteError(w, apierror.NewBadRequest("category_slug, title, and body are required"))
		return
	}
	if len(req.Title) > 300 {
		apierror.WriteError(w, apierror.NewBadRequest("Title must be 300 characters or less"))
		return
	}
	if len(req.Body) > 10000 {
		apierror.WriteError(w, apierror.NewBadRequest("Body must be 10000 characters or less"))
		return
	}

	cat, err := h.repo.GetCategoryBySlug(r.Context(), req.CategorySlug)
	if err != nil || cat == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid category"))
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
		post.Tags = pq.StringArray{}
	}

	if err := h.repo.CreatePost(r.Context(), post); err != nil {
		h.logger.WithError(err).Error("create community post failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to create post"))
		return
	}

	// Award reputation points asynchronously — don't block the response
	if h.reputer != nil {
		go h.reputer.Award(user.UserID, user.TenantID, services.ActionCommunityPost,
			"Created community post: "+req.Title, post.ID)
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
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	postID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid post ID"))
		return
	}

	post, err := h.repo.GetPostByID(r.Context(), postID, nil)
	if err != nil || post == nil {
		apierror.WriteError(w, apierror.NewNotFound("Post not found"))
		return
	}
	if post.Status == comm.StatusLocked {
		apierror.WriteError(w, apierror.NewForbidden("This thread is locked"))
		return
	}

	var req createCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	req.Body = strings.TrimSpace(req.Body)
	if req.Body == "" {
		apierror.WriteError(w, apierror.NewBadRequest("body is required"))
		return
	}
	if len(req.Body) > 8000 {
		apierror.WriteError(w, apierror.NewBadRequest("Comment must be 8000 characters or less"))
		return
	}

	if req.ParentID != nil {
		parent, err := h.repo.GetCommentByID(r.Context(), *req.ParentID)
		if err != nil || parent == nil || parent.PostID != postID {
			apierror.WriteError(w, apierror.NewBadRequest("Invalid parent comment"))
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
		apierror.WriteError(w, apierror.NewInternal("Failed to create comment"))
		return
	}

	// Award reputation points asynchronously
	if h.reputer != nil {
		go h.reputer.Award(user.UserID, user.TenantID, services.ActionCommunityComment,
			"Commented on community post", comment.ID)
	}

	// Notify post author about the reply (don't notify yourself)
	if post.AuthorID != user.UserID {
		notif := &comm.Notification{
			UserID:    post.AuthorID,
			ActorID:   user.UserID,
			Type:      "reply",
			PostID:    &postID,
			CommentID: &comment.ID,
		}
		if err := h.repo.CreateNotification(r.Context(), notif); err != nil {
			h.logger.WithError(err).Warn("failed to create reply notification")
		}
	}
	// If replying to a specific comment, also notify that comment's author
	if req.ParentID != nil {
		parent, _ := h.repo.GetCommentByID(r.Context(), *req.ParentID)
		if parent != nil && parent.AuthorID != user.UserID && parent.AuthorID != post.AuthorID {
			notif := &comm.Notification{
				UserID:    parent.AuthorID,
				ActorID:   user.UserID,
				Type:      "reply",
				PostID:    &postID,
				CommentID: &comment.ID,
			}
			if err := h.repo.CreateNotification(r.Context(), notif); err != nil {
				h.logger.WithError(err).Warn("failed to create nested reply notification")
			}
		}
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
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	var req voteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	targetType := comm.VoteTargetType(req.TargetType)
	if targetType != comm.VoteTargetPost && targetType != comm.VoteTargetComment {
		apierror.WriteError(w, apierror.NewBadRequest("target_type must be post or comment"))
		return
	}

	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid target_id"))
		return
	}

	score, err := h.repo.UpsertVote(r.Context(), user.UserID, targetType, targetID, req.Value)
	if err != nil {
		h.logger.WithError(err).Error("community vote failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to record vote"))
		return
	}

	// Award reputation points to the content author for receiving an upvote
	if h.reputer != nil && req.Value > 0 {
		// Look up the content author to award them points
		var authorID uuid.UUID
		if targetType == comm.VoteTargetPost {
			if post, err := h.repo.GetPostByID(r.Context(), targetID, nil); err == nil && post != nil {
				authorID = post.AuthorID
			}
		} else {
			if comment, err := h.repo.GetCommentByID(r.Context(), targetID); err == nil && comment != nil {
				authorID = comment.AuthorID
			}
		}
		if authorID != uuid.Nil && authorID != user.UserID {
			go h.reputer.Award(authorID, user.TenantID, services.ActionCommunityUpvote,
				"Received upvote on community content", targetID)
		}
	}

	writeJSON(w, map[string]any{"vote_score": score, "user_vote": req.Value})
}

// AcceptComment handles POST /community/posts/{id}/accept/{comment_id}
func (h *Handler) AcceptComment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	postID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid post ID"))
		return
	}
	commentID, err := parseUUID(mux.Vars(r)["comment_id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid comment ID"))
		return
	}

	post, err := h.repo.GetPostByID(r.Context(), postID, nil)
	if err != nil || post == nil {
		apierror.WriteError(w, apierror.NewNotFound("Post not found"))
		return
	}
	if post.AuthorID != user.UserID {
		apierror.WriteError(w, apierror.NewForbidden("Only the thread author can accept an answer"))
		return
	}

	comment, err := h.repo.GetCommentByID(r.Context(), commentID)
	if err != nil || comment == nil || comment.PostID != postID {
		apierror.WriteError(w, apierror.NewNotFound("Comment not found"))
		return
	}

	if err := h.repo.AcceptComment(r.Context(), postID, commentID); err != nil {
		h.logger.WithError(err).Error("accept community comment failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to accept answer"))
		return
	}

	// Award reputation points asynchronously
	if h.reputer != nil && comment.AuthorID != user.UserID {
		go h.reputer.Award(comment.AuthorID, user.TenantID, services.ActionAnswerAccepted,
			"Answer accepted on community post", commentID)
	}

	// Notify the comment author their answer was accepted
	if comment.AuthorID != user.UserID {
		notif := &comm.Notification{
			UserID:  comment.AuthorID,
			ActorID: user.UserID,
			Type:    "accepted",
			PostID:  &postID,
		}
		if err := h.repo.CreateNotification(r.Context(), notif); err != nil {
			h.logger.WithError(err).Warn("failed to create accept notification")
		}
	}

	writeJSON(w, map[string]string{"status": "solved"})
}

// RegisterRoutes registers community forum routes on the API router.
func (h *Handler) RegisterRoutes(api *mux.Router, auth *middleware.AuthMiddleware) {
	api.HandleFunc("/community/categories", h.ListCategories).Methods("GET", "OPTIONS")
	api.HandleFunc("/community/posts", h.ListPosts).Methods("GET", "OPTIONS")
	api.HandleFunc("/community/posts/{id}", h.GetPost).Methods("GET", "OPTIONS")
	api.HandleFunc("/community/posts", auth.RequireAuth(h.CreatePost)).Methods("POST", "OPTIONS")
	api.HandleFunc("/community/posts/{id}", auth.RequireAuth(h.UpdatePost)).Methods("PUT", "OPTIONS")
	api.HandleFunc("/community/posts/{id}", auth.RequireAuth(h.DeletePost)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/community/posts/{id}/comments", auth.RequireAuth(h.CreateComment)).Methods("POST", "OPTIONS")
	api.HandleFunc("/community/comments/{id}", auth.RequireAuth(h.UpdateComment)).Methods("PUT", "OPTIONS")
	api.HandleFunc("/community/comments/{id}", auth.RequireAuth(h.DeleteComment)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/community/votes", auth.RequireAuth(h.Vote)).Methods("POST", "OPTIONS")
	api.HandleFunc("/community/posts/{id}/accept/{comment_id}", auth.RequireAuth(h.AcceptComment)).Methods("POST", "OPTIONS")
	api.HandleFunc("/community/posts/{id}/bookmark", auth.RequireAuth(h.BookmarkPost)).Methods("POST", "OPTIONS")
	api.HandleFunc("/community/posts/{id}/bookmark", auth.RequireAuth(h.UnbookmarkPost)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/community/bookmarks", auth.RequireAuth(h.ListBookmarks)).Methods("GET", "OPTIONS")
	api.HandleFunc("/community/notifications", auth.RequireAuth(h.ListNotifications)).Methods("GET", "OPTIONS")
	api.HandleFunc("/community/notifications/unread-count", auth.RequireAuth(h.UnreadNotificationsCount)).Methods("GET", "OPTIONS")
	api.HandleFunc("/community/notifications/read", auth.RequireAuth(h.MarkNotificationsRead)).Methods("POST", "OPTIONS")
	api.HandleFunc("/community/users/{userId}/posts", h.ListPostsByAuthor).Methods("GET", "OPTIONS")
	api.HandleFunc("/community/rules", h.ListRules).Methods("GET", "OPTIONS")
}

// UpdatePost handles PUT /community/posts/{id}
func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	postID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid post ID"))
		return
	}

	type updatePostRequest struct {
		Title string   `json:"title"`
		Body  string   `json:"body"`
		Tags  []string `json:"tags"`
	}
	var req updatePostRequest
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
	if len(req.Title) > 300 {
		apierror.WriteError(w, apierror.NewBadRequest("Title must be 300 characters or less"))
		return
	}
	if len(req.Body) > 10000 {
		apierror.WriteError(w, apierror.NewBadRequest("Body must be 10000 characters or less"))
		return
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}

	if err := h.repo.UpdatePost(r.Context(), postID, user.UserID, req.Title, req.Body, req.Tags); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not owned") {
			apierror.WriteError(w, apierror.NewForbidden("Cannot edit this post"))
			return
		}
		h.logger.WithError(err).Error("update community post failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to update post"))
		return
	}
	writeJSON(w, map[string]string{"status": "updated"})
}

// DeletePost handles DELETE /community/posts/{id}
func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	postID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid post ID"))
		return
	}

	if err := h.repo.DeletePost(r.Context(), postID, user.UserID); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not owned") {
			apierror.WriteError(w, apierror.NewForbidden("Cannot delete this post"))
			return
		}
		h.logger.WithError(err).Error("delete community post failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete post"))
		return
	}
	writeJSON(w, map[string]string{"status": "deleted"})
}

// UpdateComment handles PUT /community/comments/{id}
func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	commentID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid comment ID"))
		return
	}

	type updateCommentRequest struct {
		Body string `json:"body"`
	}
	var req updateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	req.Body = strings.TrimSpace(req.Body)
	if req.Body == "" {
		apierror.WriteError(w, apierror.NewBadRequest("body is required"))
		return
	}
	if len(req.Body) > 8000 {
		apierror.WriteError(w, apierror.NewBadRequest("Comment must be 8000 characters or less"))
		return
	}

	if err := h.repo.UpdateComment(r.Context(), commentID, user.UserID, req.Body); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not owned") {
			apierror.WriteError(w, apierror.NewForbidden("Cannot edit this comment"))
			return
		}
		h.logger.WithError(err).Error("update community comment failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to update comment"))
		return
	}
	writeJSON(w, map[string]string{"status": "updated"})
}

// DeleteComment handles DELETE /community/comments/{id}
func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	commentID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid comment ID"))
		return
	}

	if err := h.repo.DeleteComment(r.Context(), commentID, user.UserID); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not owned") {
			apierror.WriteError(w, apierror.NewForbidden("Cannot delete this comment"))
			return
		}
		h.logger.WithError(err).Error("delete community comment failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete comment"))
		return
	}
	writeJSON(w, map[string]string{"status": "deleted"})
}

// BookmarkPost handles POST /community/posts/{id}/bookmark
func (h *Handler) BookmarkPost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	postID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid post ID"))
		return
	}
	if err := h.repo.BookmarkPost(r.Context(), user.UserID, postID); err != nil {
		h.logger.WithError(err).Error("bookmark post failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to bookmark"))
		return
	}
	writeJSON(w, map[string]string{"status": "bookmarked"})
}

// UnbookmarkPost handles DELETE /community/posts/{id}/bookmark
func (h *Handler) UnbookmarkPost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	postID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid post ID"))
		return
	}
	if err := h.repo.UnbookmarkPost(r.Context(), user.UserID, postID); err != nil {
		h.logger.WithError(err).Error("unbookmark post failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to unbookmark"))
		return
	}
	writeJSON(w, map[string]string{"status": "unbookmarked"})
}

// ListBookmarks handles GET /community/bookmarks
func (h *Handler) ListBookmarks(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	limit, offset := parsePagination(r)
	posts, total, err := h.repo.ListBookmarks(r.Context(), user.UserID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("list bookmarks failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to list bookmarks"))
		return
	}
	if posts == nil {
		posts = []comm.PostListItem{}
	}
	writeJSON(w, map[string]any{"posts": posts, "total": total, "limit": limit, "offset": offset})
}

// ListNotifications handles GET /community/notifications
func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	limit, offset := parsePagination(r)
	notifs, total, err := h.repo.ListNotifications(r.Context(), user.UserID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("list notifications failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to list notifications"))
		return
	}
	if notifs == nil {
		notifs = []comm.NotificationWithActor{}
	}
	writeJSON(w, map[string]any{"notifications": notifs, "total": total, "limit": limit, "offset": offset})
}

// UnreadNotificationsCount handles GET /community/notifications/unread-count
func (h *Handler) UnreadNotificationsCount(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	count, err := h.repo.CountUnreadNotifications(r.Context(), user.UserID)
	if err != nil {
		h.logger.WithError(err).Error("count unread notifications failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to count notifications"))
		return
	}
	writeJSON(w, map[string]any{"count": count})
}

// MarkNotificationsRead handles POST /community/notifications/read
func (h *Handler) MarkNotificationsRead(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	if err := h.repo.MarkNotificationsRead(r.Context(), user.UserID); err != nil {
		h.logger.WithError(err).Error("mark notifications read failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to mark notifications"))
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// ListPostsByAuthor handles GET /community/users/{userId}/posts
func (h *Handler) ListPostsByAuthor(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUUID(mux.Vars(r)["userId"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid user ID"))
		return
	}
	limit, offset := parsePagination(r)
	posts, total, err := h.repo.ListPostsByAuthor(r.Context(), userID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("list posts by author failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to list posts"))
		return
	}
	if posts == nil {
		posts = []comm.PostListItem{}
	}
	writeJSON(w, map[string]any{"posts": posts, "total": total, "limit": limit, "offset": offset})
}

// ListRules handles GET /community/rules — returns active rules for public display.
func (h *Handler) ListRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.repo.ListRules(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("list community rules failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to list rules"))
		return
	}
	if rules == nil {
		rules = []comm.Rule{}
	}
	writeJSON(w, map[string]any{"rules": rules})
}

// RegisterAdminRoutes registers admin-only community management routes.
func (h *Handler) RegisterAdminRoutes(api *mux.Router, auth *middleware.AuthMiddleware) {
	api.HandleFunc("/admin/community/rules", auth.RequireAuth(h.AdminListRules)).Methods("GET", "OPTIONS")
	api.HandleFunc("/admin/community/rules", auth.RequireAuth(h.AdminCreateRule)).Methods("POST", "OPTIONS")
	api.HandleFunc("/admin/community/rules/{id}", auth.RequireAuth(h.AdminUpdateRule)).Methods("PUT", "OPTIONS")
	api.HandleFunc("/admin/community/rules/{id}", auth.RequireAuth(h.AdminDeleteRule)).Methods("DELETE", "OPTIONS")
}

// AdminListRules handles GET /admin/community/rules — returns all rules including inactive.
func (h *Handler) AdminListRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.repo.ListAllRules(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("admin list rules failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to list rules"))
		return
	}
	if rules == nil {
		rules = []comm.Rule{}
	}
	writeJSON(w, map[string]any{"rules": rules})
}

type createRuleRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Enforcement string `json:"enforcement"`
	SortOrder   int    `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
}

// AdminCreateRule handles POST /admin/community/rules
func (h *Handler) AdminCreateRule(w http.ResponseWriter, r *http.Request) {
	var req createRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		apierror.WriteError(w, apierror.NewBadRequest("title is required"))
		return
	}
	if req.Category == "" {
		req.Category = "conduct"
	}
	if req.Enforcement == "" {
		req.Enforcement = "warning"
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	rule := &comm.Rule{
		Title:       req.Title,
		Description: req.Description,
		Category:    comm.RuleCategory(req.Category),
		Enforcement: comm.RuleEnforcement(req.Enforcement),
		SortOrder:   req.SortOrder,
		IsActive:    isActive,
	}
	if err := h.repo.CreateRule(r.Context(), rule); err != nil {
		h.logger.WithError(err).Error("create rule failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to create rule"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

type updateRuleRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Enforcement *string `json:"enforcement"`
	SortOrder   *int    `json:"sort_order"`
	IsActive    *bool   `json:"is_active"`
}

// AdminUpdateRule handles PUT /admin/community/rules/{id}
func (h *Handler) AdminUpdateRule(w http.ResponseWriter, r *http.Request) {
	ruleID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid rule ID"))
		return
	}
	existing, err := h.repo.GetRule(r.Context(), ruleID)
	if err != nil || existing == nil {
		apierror.WriteError(w, apierror.NewNotFound("Rule not found"))
		return
	}

	var req updateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Title != nil {
		existing.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Category != nil {
		existing.Category = comm.RuleCategory(*req.Category)
	}
	if req.Enforcement != nil {
		existing.Enforcement = comm.RuleEnforcement(*req.Enforcement)
	}
	if req.SortOrder != nil {
		existing.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := h.repo.UpdateRule(r.Context(), existing); err != nil {
		h.logger.WithError(err).Error("update rule failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to update rule"))
		return
	}
	writeJSON(w, existing)
}

// AdminDeleteRule handles DELETE /admin/community/rules/{id}
func (h *Handler) AdminDeleteRule(w http.ResponseWriter, r *http.Request) {
	ruleID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid rule ID"))
		return
	}
	if err := h.repo.DeleteRule(r.Context(), ruleID); err != nil {
		h.logger.WithError(err).Error("delete rule failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete rule"))
		return
	}
	writeJSON(w, map[string]string{"status": "deleted"})
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
