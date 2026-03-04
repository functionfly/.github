package follow

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler contains follow-related HTTP handlers
type Handler struct {
	followService *services.FollowService
	repo          storage.Repository
	authSvc       *auth.AuthService
}

// NewHandler creates a new follow handler
func NewHandler(followService *services.FollowService, repo storage.Repository, authSvc *auth.AuthService) *Handler {
	return &Handler{
		followService: followService,
		repo:          repo,
		authSvc:       authSvc,
	}
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// getCurrentUserID gets the current user ID from the context
func getCurrentUserID(r *http.Request) (uuid.UUID, error) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		return uuid.Nil, http.ErrNoCookie
	}
	return claims.UserID, nil
}

// FollowUserRequest represents a request to follow a user
type FollowUserRequest struct {
	Reason                 *string `json:"reason,omitempty"`
	NotifyOnNewFunction    *bool   `json:"notify_on_new_function,omitempty"`
	NotifyOnFunctionUpdate *bool   `json:"notify_on_function_update,omitempty"`
	NotifyOnNewVersion     *bool   `json:"notify_on_new_version,omitempty"`
}

// FollowUserResponse represents the response from following a user
type FollowUserResponse struct {
	OK       bool   `json:"ok"`
	FollowID string `json:"follow_id"`
}

// HandleFollowUser handles POST /v1/follow/users/{username}/follow
func (h *Handler) HandleFollowUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "username is required")
		return
	}

	currentUserID, err := getCurrentUserID(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request body
	var req FollowUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is allowed - use defaults
		req = FollowUserRequest{}
	}

	// Follow the user
	follow, err := h.followService.FollowUser(r.Context(), services.FollowUserRequest{
		FollowerID:             currentUserID,
		FollowedUsername:       username,
		Reason:                 req.Reason,
		NotifyOnNewFunction:    req.NotifyOnNewFunction,
		NotifyOnFunctionUpdate: req.NotifyOnFunctionUpdate,
		NotifyOnNewVersion:     req.NotifyOnNewVersion,
	})

	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			writeJSONError(w, http.StatusNotFound, "user not found")
		case services.ErrCannotFollowSelf:
			writeJSONError(w, http.StatusBadRequest, "cannot follow yourself")
		case services.ErrAlreadyFollowingUser:
			writeJSONError(w, http.StatusConflict, "already following this user")
		default:
			logrus.WithError(err).Error("Failed to follow user")
			writeJSONError(w, http.StatusInternalServerError, "failed to follow user")
		}
		return
	}

	writeJSON(w, http.StatusCreated, FollowUserResponse{
		OK:       true,
		FollowID: follow.ID.String(),
	})
}

// HandleUnfollowUser handles DELETE /v1/follow/users/{username}/follow
func (h *Handler) HandleUnfollowUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "username is required")
		return
	}

	currentUserID, err := getCurrentUserID(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get target user by username
	targetUser, err := h.repo.GetUserByUsername(username)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Unfollow the user
	err = h.followService.UnfollowUser(r.Context(), currentUserID, targetUser.ID)
	if err != nil {
		switch err {
		case services.ErrNotFollowingUser:
			writeJSONError(w, http.StatusNotFound, "not following this user")
		default:
			logrus.WithError(err).Error("Failed to unfollow user")
			writeJSONError(w, http.StatusInternalServerError, "failed to unfollow user")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// UserFollowResponse represents a user follow relationship in the API
type UserFollowResponse struct {
	ID             string `json:"id"`
	FollowerID     string `json:"follower_id"`
	FollowerName   string `json:"follower_name,omitempty"`
	FollowerAvatar string `json:"follower_avatar,omitempty"`
	FollowedID     string `json:"followed_id,omitempty"`
	FollowedName   string `json:"followed_name,omitempty"`
	FollowedAvatar string `json:"followed_avatar,omitempty"`
	Reason         string `json:"reason,omitempty"`
	CreatedAt      string `json:"created_at"`
}

// PaginatedUserFollowResponse represents a paginated list of user follows
type PaginatedUserFollowResponse struct {
	Data       []UserFollowResponse `json:"data"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// HandleGetUserFollowers handles GET /v1/follow/users/{username}/followers
func (h *Handler) HandleGetUserFollowers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "username is required")
		return
	}

	// Get user by username
	user, err := h.repo.GetUserByUsername(username)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get followers
	follows, total, err := h.followService.GetUserFollowers(r.Context(), user.ID, page, pageSize)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user followers")
		writeJSONError(w, http.StatusInternalServerError, "failed to get followers")
		return
	}

	// Convert to response
	response := PaginatedUserFollowResponse{
		Data:       make([]UserFollowResponse, len(follows)),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + pageSize - 1) / pageSize,
	}

	for i, follow := range follows {
		resp := UserFollowResponse{
			ID:        follow.ID.String(),
			FollowerID: follow.FollowerID.String(),
			FollowedID: follow.FollowedUserID.String(),
			CreatedAt: follow.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if follow.FollowReason != nil {
			resp.Reason = *follow.FollowReason
		}
		if follow.Follower != nil {
			resp.FollowerName = follow.Follower.Name
			if follow.Follower.Username != nil {
				resp.FollowerName = *follow.Follower.Username
			}
		}
		if follow.FollowedUser != nil {
			resp.FollowedName = follow.FollowedUser.Name
			if follow.FollowedUser.Username != nil {
				resp.FollowedName = *follow.FollowedUser.Username
			}
		}
		response.Data[i] = resp
	}

	writeJSON(w, http.StatusOK, response)
}

// HandleGetUserFollowing handles GET /v1/follow/users/{username}/following
func (h *Handler) HandleGetUserFollowing(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "username is required")
		return
	}

	// Get user by username
	user, err := h.repo.GetUserByUsername(username)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get following
	follows, total, err := h.followService.GetUserFollowing(r.Context(), user.ID, page, pageSize)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user following")
		writeJSONError(w, http.StatusInternalServerError, "failed to get following")
		return
	}

	// Convert to response
	response := PaginatedUserFollowResponse{
		Data:       make([]UserFollowResponse, len(follows)),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + pageSize - 1) / pageSize,
	}

	for i, follow := range follows {
		resp := UserFollowResponse{
			ID:         follow.ID.String(),
			FollowerID: follow.FollowerID.String(),
			FollowedID: follow.FollowedUserID.String(),
			CreatedAt:  follow.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if follow.FollowReason != nil {
			resp.Reason = *follow.FollowReason
		}
		if follow.Follower != nil {
			resp.FollowerName = follow.Follower.Name
			if follow.Follower.Username != nil {
				resp.FollowerName = *follow.Follower.Username
			}
		}
		if follow.FollowedUser != nil {
			resp.FollowedName = follow.FollowedUser.Name
			if follow.FollowedUser.Username != nil {
				resp.FollowedName = *follow.FollowedUser.Username
			}
		}
		response.Data[i] = resp
	}

	writeJSON(w, http.StatusOK, response)
}

// HandleCheckFollowingStatus handles GET /v1/follow/users/{username}/status
func (h *Handler) HandleCheckFollowingStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "username is required")
		return
	}

	currentUserID, err := getCurrentUserID(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get target user by username
	targetUser, err := h.repo.GetUserByUsername(username)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	// Check following status
	isFollowing, err := h.followService.IsFollowingUser(r.Context(), currentUserID, targetUser.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check following status")
		writeJSONError(w, http.StatusInternalServerError, "failed to check following status")
		return
	}

	// Get follower counts
	followerCount, _ := h.followService.GetUserFollowerCount(r.Context(), targetUser.ID)
	followingCount, _ := h.followService.GetUserFollowingCount(r.Context(), targetUser.ID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"is_following":   isFollowing,
		"follower_count": followerCount,
		"following_count": followingCount,
	})
}

// FollowFunctionRequest represents a request to follow a function
type FollowFunctionRequest struct {
	Reason               *string `json:"reason,omitempty"`
	NotifyOnNewVersion   *bool   `json:"notify_on_new_version,omitempty"`
	NotifyOnRatingChange *bool   `json:"notify_on_rating_change,omitempty"`
	NotifyOnTrustChange  *bool   `json:"notify_on_trust_change,omitempty"`
	NotifyOnVerification *bool   `json:"notify_on_verification,omitempty"`
}

// FollowFunctionResponse represents the response from following a function
type FollowFunctionResponse struct {
	OK       bool   `json:"ok"`
	FollowID string `json:"follow_id"`
}

// HandleFollowFunction handles POST /v1/follow/functions/{functionID}/follow
func (h *Handler) HandleFollowFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionID"]
	if functionIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "function ID is required")
		return
	}

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid function ID")
		return
	}

	currentUserID, err := getCurrentUserID(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse request body
	var req FollowFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req = FollowFunctionRequest{}
	}

	// Follow the function
	follow, err := h.followService.FollowFunction(r.Context(), services.FollowFunctionRequest{
		UserID:               currentUserID,
		FunctionID:           functionID,
		Reason:               req.Reason,
		NotifyOnNewVersion:   req.NotifyOnNewVersion,
		NotifyOnRatingChange: req.NotifyOnRatingChange,
		NotifyOnTrustChange:  req.NotifyOnTrustChange,
		NotifyOnVerification: req.NotifyOnVerification,
	})

	if err != nil {
		switch err {
		case services.ErrAlreadyFollowingFunction:
			writeJSONError(w, http.StatusConflict, "already following this function")
		default:
			logrus.WithError(err).Error("Failed to follow function")
			writeJSONError(w, http.StatusInternalServerError, "failed to follow function")
		}
		return
	}

	writeJSON(w, http.StatusCreated, FollowFunctionResponse{
		OK:       true,
		FollowID: follow.ID.String(),
	})
}

// HandleUnfollowFunction handles DELETE /v1/follow/functions/{functionID}/follow
func (h *Handler) HandleUnfollowFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionID"]
	if functionIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "function ID is required")
		return
	}

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid function ID")
		return
	}

	currentUserID, err := getCurrentUserID(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Unfollow the function
	err = h.followService.UnfollowFunction(r.Context(), currentUserID, functionID)
	if err != nil {
		switch err {
		case services.ErrNotFollowingFunction:
			writeJSONError(w, http.StatusNotFound, "not following this function")
		default:
			logrus.WithError(err).Error("Failed to unfollow function")
			writeJSONError(w, http.StatusInternalServerError, "failed to unfollow function")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// FunctionFollowResponse represents a function follow relationship in the API
type FunctionFollowResponse struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name,omitempty"`
	FunctionID   string `json:"function_id"`
	FunctionName string `json:"function_name,omitempty"`
	Reason       string `json:"reason,omitempty"`
	CreatedAt    string `json:"created_at"`
}

// PaginatedFunctionFollowResponse represents a paginated list of function follows
type PaginatedFunctionFollowResponse struct {
	Data       []FunctionFollowResponse `json:"data"`
	Total      int                      `json:"total"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
}

// HandleGetFunctionFollowers handles GET /v1/follow/functions/{functionID}/followers
func (h *Handler) HandleGetFunctionFollowers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionID"]
	if functionIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "function ID is required")
		return
	}

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid function ID")
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get followers
	follows, total, err := h.followService.GetFunctionFollowers(r.Context(), functionID, page, pageSize)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function followers")
		writeJSONError(w, http.StatusInternalServerError, "failed to get followers")
		return
	}

	// Convert to response
	response := PaginatedFunctionFollowResponse{
		Data:       make([]FunctionFollowResponse, len(follows)),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + pageSize - 1) / pageSize,
	}

	for i, follow := range follows {
		resp := FunctionFollowResponse{
			ID:         follow.ID.String(),
			UserID:     follow.UserID.String(),
			FunctionID: follow.FunctionID.String(),
			CreatedAt:  follow.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if follow.FollowReason != nil {
			resp.Reason = *follow.FollowReason
		}
		if follow.User != nil {
			resp.UserName = follow.User.Name
			if follow.User.Username != nil {
				resp.UserName = *follow.User.Username
			}
		}
		if follow.Function != nil {
			resp.FunctionName = follow.Function.Name
		}
		response.Data[i] = resp
	}

	writeJSON(w, http.StatusOK, response)
}

// HandleGetMyFollowedFunctions handles GET /v1/follow/me/functions
func (h *Handler) HandleGetMyFollowedFunctions(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := getCurrentUserID(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get followed functions
	follows, total, err := h.followService.GetUserFunctionFollows(r.Context(), currentUserID, page, pageSize)
	if err != nil {
		logrus.WithError(err).Error("Failed to get followed functions")
		writeJSONError(w, http.StatusInternalServerError, "failed to get followed functions")
		return
	}

	// Convert to response
	response := PaginatedFunctionFollowResponse{
		Data:       make([]FunctionFollowResponse, len(follows)),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + pageSize - 1) / pageSize,
	}

	for i, follow := range follows {
		resp := FunctionFollowResponse{
			ID:         follow.ID.String(),
			UserID:     follow.UserID.String(),
			FunctionID: follow.FunctionID.String(),
			CreatedAt:  follow.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if follow.FollowReason != nil {
			resp.Reason = *follow.FollowReason
		}
		if follow.User != nil {
			resp.UserName = follow.User.Name
			if follow.User.Username != nil {
				resp.UserName = *follow.User.Username
			}
		}
		if follow.Function != nil {
			resp.FunctionName = follow.Function.Name
		}
		response.Data[i] = resp
	}

	writeJSON(w, http.StatusOK, response)
}

// HandleCheckFunctionFollowingStatus handles GET /v1/follow/functions/{functionID}/status
func (h *Handler) HandleCheckFunctionFollowingStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionID"]
	if functionIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "function ID is required")
		return
	}

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid function ID")
		return
	}

	currentUserID, err := getCurrentUserID(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Check following status
	isFollowing, err := h.followService.IsFollowingFunction(r.Context(), currentUserID, functionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check function following status")
		writeJSONError(w, http.StatusInternalServerError, "failed to check following status")
		return
	}

	// Get follower count
	followerCount, _ := h.followService.GetFunctionFollowerCount(r.Context(), functionID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"is_following":   isFollowing,
		"follower_count": followerCount,
	})
}

// HandleGetMyFollowStats handles GET /v1/follow/me/stats
func (h *Handler) HandleGetMyFollowStats(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := getCurrentUserID(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get stats
	followerCount, _ := h.followService.GetUserFollowerCount(r.Context(), currentUserID)
	followingCount, _ := h.followService.GetUserFollowingCount(r.Context(), currentUserID)
	functionFollows, _ := h.followService.GetUserFunctionFollows(r.Context(), currentUserID, 1, 1)
	_, functionFollowCount, _ := h.followService.GetUserFunctionFollows(r.Context(), currentUserID, 1, 1000)

	_ = functionFollows // Suppress unused variable warning

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"followers":        followerCount,
		"following":        followingCount,
		"functions_followed": functionFollowCount,
	})
}
