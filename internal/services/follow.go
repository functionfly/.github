package services

import (
	"context"
	"errors"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

var (
	// ErrAlreadyFollowingUser is returned when trying to follow a user already being followed
	ErrAlreadyFollowingUser = errors.New("already following this user")
	// ErrNotFollowingUser is returned when trying to unfollow a user not being followed
	ErrNotFollowingUser = errors.New("not following this user")
	// ErrAlreadyFollowingFunction is returned when trying to follow a function already being followed
	ErrAlreadyFollowingFunction = errors.New("already following this function")
	// ErrNotFollowingFunction is returned when trying to unfollow a function not being followed
	ErrNotFollowingFunction = errors.New("not following this function")
	// ErrCannotFollowSelf is returned when trying to follow yourself
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
	// ErrUserNotFound is returned when the target user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrFunctionNotFound is returned when the target function is not found
	ErrFunctionNotFound = errors.New("function not found")
)

// FollowService handles follow business logic
type FollowService struct {
	followRepo    storage.FollowRepositoryInterface
	userRepo     storage.Repository
	functionRepo interface {
		GetFunctionByID(ctx context.Context, id uuid.UUID) (*storage.RegistryFunction, error)
	}
}

// FollowRepositoryInterface defines the follow repository interface
type FollowRepositoryInterface interface {
	// User Follows
	FollowUser(ctx context.Context, followerID, followedUserID uuid.UUID, reason *string, notifyOnNewFunction, notifyOnFunctionUpdate, notifyOnNewVersion bool) (*storage.UserFollow, error)
	UnfollowUser(ctx context.Context, followerID, followedUserID uuid.UUID) error
	IsFollowingUser(ctx context.Context, followerID, followedUserID uuid.UUID) (bool, error)
	GetUserFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*storage.UserFollow, int, error)
	GetUserFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*storage.UserFollow, int, error)
	GetUserFollowerCount(ctx context.Context, userID uuid.UUID) (int, error)
	GetUserFollowingCount(ctx context.Context, userID uuid.UUID) (int, error)

	// Function Follows
	FollowFunction(ctx context.Context, userID, functionID uuid.UUID, reason *string, notifyOnNewVersion, notifyOnRatingChange, notifyOnTrustChange, notifyOnVerification bool) (*storage.FunctionFollow, error)
	UnfollowFunction(ctx context.Context, userID, functionID uuid.UUID) error
	IsFollowingFunction(ctx context.Context, userID, functionID uuid.UUID) (bool, error)
	GetFunctionFollowers(ctx context.Context, functionID uuid.UUID, limit, offset int) ([]*storage.FunctionFollow, int, error)
	GetUserFunctionFollows(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*storage.FunctionFollow, int, error)
	GetFunctionFollowerCount(ctx context.Context, functionID uuid.UUID) (int, error)
}

// NewFollowService creates a new follow service
func NewFollowService(followRepo storage.FollowRepositoryInterface, userRepo storage.Repository) *FollowService {
	return &FollowService{
		followRepo: followRepo,
		userRepo:   userRepo,
	}
}

// FollowUserRequest represents a request to follow a user
type FollowUserRequest struct {
	FollowerID            uuid.UUID
	FollowedUserID        uuid.UUID
	FollowedUsername      string // Optional: can use username instead of ID
	Reason                *string
	NotifyOnNewFunction   *bool
	NotifyOnFunctionUpdate *bool
	NotifyOnNewVersion    *bool
}

// FollowUser allows a user to follow another user
func (s *FollowService) FollowUser(ctx context.Context, req FollowUserRequest) (*storage.UserFollow, error) {
	var targetUserID uuid.UUID
	var err error

	// Resolve target user ID from username if provided
	if req.FollowedUsername != "" {
		targetUser, err := s.userRepo.GetUserByUsername(req.FollowedUsername)
		if err != nil {
			return nil, ErrUserNotFound
		}
		targetUserID = targetUser.ID
	} else {
		targetUserID = req.FollowedUserID
	}

	// Validation
	if req.FollowerID == targetUserID {
		return nil, ErrCannotFollowSelf
	}

	// Check if already following
	exists, err := s.followRepo.IsFollowingUser(ctx, req.FollowerID, targetUserID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyFollowingUser
	}

	// Check if target user exists
	if _, err := s.userRepo.GetUserByID(targetUserID); err != nil {
		return nil, ErrUserNotFound
	}

	// Set default notification preferences
	notifyOnNewFunction := true
	notifyOnFunctionUpdate := true
	notifyOnNewVersion := true
	if req.NotifyOnNewFunction != nil {
		notifyOnNewFunction = *req.NotifyOnNewFunction
	}
	if req.NotifyOnFunctionUpdate != nil {
		notifyOnFunctionUpdate = *req.NotifyOnFunctionUpdate
	}
	if req.NotifyOnNewVersion != nil {
		notifyOnNewVersion = *req.NotifyOnNewVersion
	}

	// Create follow relationship
	follow, err := s.followRepo.FollowUser(ctx, req.FollowerID, targetUserID, req.Reason, notifyOnNewFunction, notifyOnFunctionUpdate, notifyOnNewVersion)
	if err != nil {
		return nil, err
	}

	return follow, nil
}

// UnfollowUser allows a user to unfollow another user
func (s *FollowService) UnfollowUser(ctx context.Context, followerID, followedUserID uuid.UUID) error {
	exists, err := s.followRepo.IsFollowingUser(ctx, followerID, followedUserID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFollowingUser
	}

	return s.followRepo.UnfollowUser(ctx, followerID, followedUserID)
}

// FollowFunctionRequest represents a request to follow a function
type FollowFunctionRequest struct {
	UserID                 uuid.UUID
	FunctionID             uuid.UUID
	Reason                 *string
	NotifyOnNewVersion     *bool
	NotifyOnRatingChange   *bool
	NotifyOnTrustChange   *bool
	NotifyOnVerification  *bool
}

// FollowFunction allows a user to follow a function
func (s *FollowService) FollowFunction(ctx context.Context, req FollowFunctionRequest) (*storage.FunctionFollow, error) {
	exists, err := s.followRepo.IsFollowingFunction(ctx, req.UserID, req.FunctionID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyFollowingFunction
	}

	// Set default notification preferences
	notifyOnNewVersion := true
	notifyOnRatingChange := false
	notifyOnTrustChange := true
	notifyOnVerification := true
	if req.NotifyOnNewVersion != nil {
		notifyOnNewVersion = *req.NotifyOnNewVersion
	}
	if req.NotifyOnRatingChange != nil {
		notifyOnRatingChange = *req.NotifyOnRatingChange
	}
	if req.NotifyOnTrustChange != nil {
		notifyOnTrustChange = *req.NotifyOnTrustChange
	}
	if req.NotifyOnVerification != nil {
		notifyOnVerification = *req.NotifyOnVerification
	}

	return s.followRepo.FollowFunction(ctx, req.UserID, req.FunctionID, req.Reason, notifyOnNewVersion, notifyOnRatingChange, notifyOnTrustChange, notifyOnVerification)
}

// UnfollowFunction allows a user to unfollow a function
func (s *FollowService) UnfollowFunction(ctx context.Context, userID, functionID uuid.UUID) error {
	exists, err := s.followRepo.IsFollowingFunction(ctx, userID, functionID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFollowingFunction
	}

	return s.followRepo.UnfollowFunction(ctx, userID, functionID)
}

// GetUserFollowers returns followers of a user
func (s *FollowService) GetUserFollowers(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*storage.UserFollow, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.followRepo.GetUserFollowers(ctx, userID, pageSize, offset)
}

// GetUserFollowing returns users that a user is following
func (s *FollowService) GetUserFollowing(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*storage.UserFollow, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.followRepo.GetUserFollowing(ctx, userID, pageSize, offset)
}

// GetFunctionFollowers returns followers of a function
func (s *FollowService) GetFunctionFollowers(ctx context.Context, functionID uuid.UUID, page, pageSize int) ([]*storage.FunctionFollow, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.followRepo.GetFunctionFollowers(ctx, functionID, pageSize, offset)
}

// GetUserFunctionFollows returns functions that a user is following
func (s *FollowService) GetUserFunctionFollows(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*storage.FunctionFollow, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.followRepo.GetUserFunctionFollows(ctx, userID, pageSize, offset)
}

// IsFollowingUser checks if a user is following another user
func (s *FollowService) IsFollowingUser(ctx context.Context, followerID, followedUserID uuid.UUID) (bool, error) {
	return s.followRepo.IsFollowingUser(ctx, followerID, followedUserID)
}

// IsFollowingFunction checks if a user is following a function
func (s *FollowService) IsFollowingFunction(ctx context.Context, userID, functionID uuid.UUID) (bool, error) {
	return s.followRepo.IsFollowingFunction(ctx, userID, functionID)
}

// GetUserFollowerCount returns the follower count for a user
func (s *FollowService) GetUserFollowerCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.followRepo.GetUserFollowerCount(ctx, userID)
}

// GetUserFollowingCount returns the following count for a user
func (s *FollowService) GetUserFollowingCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.followRepo.GetUserFollowingCount(ctx, userID)
}

// GetFunctionFollowerCount returns the follower count for a function
func (s *FollowService) GetFunctionFollowerCount(ctx context.Context, functionID uuid.UUID) (int, error) {
	return s.followRepo.GetFunctionFollowerCount(ctx, functionID)
}
