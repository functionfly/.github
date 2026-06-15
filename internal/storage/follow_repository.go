package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// FollowRepository handles follow-related database operations
type FollowRepository struct {
	db *PostgresDB
}

// NewFollowRepository creates a new follow repository
func NewFollowRepository(db *PostgresDB) *FollowRepository {
	return &FollowRepository{db: db}
}

// User Follow Operations

// FollowUser creates a new user follow relationship
func (r *FollowRepository) FollowUser(ctx context.Context, followerID, followedUserID uuid.UUID, reason *string, notifyOnNewFunction, notifyOnFunctionUpdate, notifyOnNewVersion bool) (*UserFollow, error) {
	follow := &UserFollow{
		ID:                   uuid.New(),
		FollowerID:           followerID,
		FollowedUserID:       followedUserID,
		FollowReason:         reason,
		NotifyOnNewFunction:  notifyOnNewFunction,
		NotifyOnFunctionUpdate: notifyOnFunctionUpdate,
		NotifyOnNewVersion:   notifyOnNewVersion,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	_, err := r.db.Exec(`
		INSERT INTO user_follows (id, follower_id, followed_user_id, follow_reason, notify_on_new_function, notify_on_function_update, notify_on_new_version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		follow.ID, follow.FollowerID, follow.FollowedUserID, follow.FollowReason, follow.NotifyOnNewFunction, follow.NotifyOnFunctionUpdate, follow.NotifyOnNewVersion, follow.CreatedAt, follow.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user follow: %w", err)
	}

	// Update follower/following counts
	r.updateUserFollowCounts(followerID, followedUserID, 1)

	return follow, nil
}

// UnfollowUser removes a user follow relationship
func (r *FollowRepository) UnfollowUser(ctx context.Context, followerID, followedUserID uuid.UUID) error {
	result, err := r.db.Exec(`
		DELETE FROM user_follows WHERE follower_id = $1 AND followed_user_id = $2`,
		followerID, followedUserID)

	if err != nil {
		return fmt.Errorf("failed to delete user follow: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("follow relationship not found")
	}

	// Update follower/following counts
	r.updateUserFollowCounts(followerID, followedUserID, -1)

	return nil
}

// IsFollowingUser checks if a user is following another user
func (r *FollowRepository) IsFollowingUser(ctx context.Context, followerID, followedUserID uuid.UUID) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_follows WHERE follower_id = $1 AND followed_user_id = $2`,
		followerID, followedUserID).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check follow relationship: %w", err)
	}

	return count > 0, nil
}

// GetUserFollowers returns all followers of a user
func (r *FollowRepository) GetUserFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserFollow, int, error) {
	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_follows WHERE followed_user_id = $1`,
		userID).Scan(&total)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get followers count: %w", err)
	}

	// Get followers with user details
	rows, err := r.db.QueryContext(ctx, `
		SELECT uf.id, uf.follower_id, uf.followed_user_id, uf.follow_reason, uf.notify_on_new_function, uf.notify_on_function_update, uf.notify_on_new_version, uf.created_at, uf.updated_at,
		       u.id, u.username, u.email, u.name
		FROM user_follows uf
		JOIN users u ON uf.follower_id = u.id
		WHERE uf.followed_user_id = $1
		ORDER BY uf.created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get followers: %w", err)
	}
	defer rows.Close()

	var follows []*UserFollow
	for rows.Next() {
		follow := &UserFollow{}
		follower := &User{}
		err := rows.Scan(
			&follow.ID, &follow.FollowerID, &follow.FollowedUserID, &follow.FollowReason, &follow.NotifyOnNewFunction, &follow.NotifyOnFunctionUpdate, &follow.NotifyOnNewVersion, &follow.CreatedAt, &follow.UpdatedAt,
			&follower.ID, &follower.Username, &follower.Email, &follower.Name,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan follower: %w", err)
		}
		follow.Follower = follower
		follows = append(follows, follow)
	}

	return follows, total, nil
}

// GetUserFollowing returns all users that a user is following
func (r *FollowRepository) GetUserFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserFollow, int, error) {
	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_follows WHERE follower_id = $1`,
		userID).Scan(&total)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get following count: %w", err)
	}

	// Get following with user details
	rows, err := r.db.QueryContext(ctx, `
		SELECT uf.id, uf.follower_id, uf.followed_user_id, uf.follow_reason, uf.notify_on_new_function, uf.notify_on_function_update, uf.notify_on_new_version, uf.created_at, uf.updated_at,
		       u.id, u.username, u.email, u.name
		FROM user_follows uf
		JOIN users u ON uf.followed_user_id = u.id
		WHERE uf.follower_id = $1
		ORDER BY uf.created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get following: %w", err)
	}
	defer rows.Close()

	var follows []*UserFollow
	for rows.Next() {
		follow := &UserFollow{}
		followedUser := &User{}
		err := rows.Scan(
			&follow.ID, &follow.FollowerID, &follow.FollowedUserID, &follow.FollowReason, &follow.NotifyOnNewFunction, &follow.NotifyOnFunctionUpdate, &follow.NotifyOnNewVersion, &follow.CreatedAt, &follow.UpdatedAt,
			&followedUser.ID, &followedUser.Username, &followedUser.Email, &followedUser.Name,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan following: %w", err)
		}
		follow.FollowedUser = followedUser
		follows = append(follows, follow)
	}

	return follows, total, nil
}

// GetUserFollowerCount returns the follower count for a user
func (r *FollowRepository) GetUserFollowerCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_follows WHERE followed_user_id = $1`,
		userID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to get follower count: %w", err)
	}

	return count, nil
}

// GetUserFollowingCount returns the following count for a user
func (r *FollowRepository) GetUserFollowingCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_follows WHERE follower_id = $1`,
		userID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to get following count: %w", err)
	}

	return count, nil
}

// Function Follow Operations

// FollowFunction creates a new function follow relationship
func (r *FollowRepository) FollowFunction(ctx context.Context, userID, functionID uuid.UUID, reason *string, notifyOnNewVersion, notifyOnRatingChange, notifyOnTrustChange, notifyOnVerification bool) (*FunctionFollow, error) {
	follow := &FunctionFollow{
		ID:                   uuid.New(),
		UserID:               userID,
		FunctionID:           functionID,
		FollowReason:         reason,
		NotifyOnNewVersion:   notifyOnNewVersion,
		NotifyOnRatingChange: notifyOnRatingChange,
		NotifyOnTrustChange:  notifyOnTrustChange,
		NotifyOnVerification: notifyOnVerification,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	_, err := r.db.Exec(`
		INSERT INTO function_follows (id, user_id, function_id, follow_reason, notify_on_new_version, notify_on_rating_change, notify_on_trust_change, notify_on_verification, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		follow.ID, follow.UserID, follow.FunctionID, follow.FollowReason, follow.NotifyOnNewVersion, follow.NotifyOnRatingChange, follow.NotifyOnTrustChange, follow.NotifyOnVerification, follow.CreatedAt, follow.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create function follow: %w", err)
	}

	// Update function follower count
	r.updateFunctionFollowCount(functionID, 1)

	return follow, nil
}

// UnfollowFunction removes a function follow relationship
func (r *FollowRepository) UnfollowFunction(ctx context.Context, userID, functionID uuid.UUID) error {
	result, err := r.db.Exec(`
		DELETE FROM function_follows WHERE user_id = $1 AND function_id = $2`,
		userID, functionID)

	if err != nil {
		return fmt.Errorf("failed to delete function follow: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("follow relationship not found")
	}

	// Update function follower count
	r.updateFunctionFollowCount(functionID, -1)

	return nil
}

// IsFollowingFunction checks if a user is following a function
func (r *FollowRepository) IsFollowingFunction(ctx context.Context, userID, functionID uuid.UUID) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM function_follows WHERE user_id = $1 AND function_id = $2`,
		userID, functionID).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check function follow relationship: %w", err)
	}

	return count > 0, nil
}

// GetFunctionFollowers returns all followers of a function
func (r *FollowRepository) GetFunctionFollowers(ctx context.Context, functionID uuid.UUID, limit, offset int) ([]*FunctionFollow, int, error) {
	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM function_follows WHERE function_id = $1`,
		functionID).Scan(&total)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get function followers count: %w", err)
	}

	// Get followers with user details
	rows, err := r.db.QueryContext(ctx, `
		SELECT ff.id, ff.user_id, ff.function_id, ff.follow_reason, ff.notify_on_new_version, ff.notify_on_rating_change, ff.notify_on_trust_change, ff.notify_on_verification, ff.created_at, ff.updated_at,
		       u.id, u.username, u.email, u.name
		FROM function_follows ff
		JOIN users u ON ff.user_id = u.id
		WHERE ff.function_id = $1
		ORDER BY ff.created_at DESC
		LIMIT $2 OFFSET $3`,
		functionID, limit, offset)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get function followers: %w", err)
	}
	defer rows.Close()

	var follows []*FunctionFollow
	for rows.Next() {
		follow := &FunctionFollow{}
		user := &User{}
		err := rows.Scan(
			&follow.ID, &follow.UserID, &follow.FunctionID, &follow.FollowReason, &follow.NotifyOnNewVersion, &follow.NotifyOnRatingChange, &follow.NotifyOnTrustChange, &follow.NotifyOnVerification, &follow.CreatedAt, &follow.UpdatedAt,
			&user.ID, &user.Username, &user.Email, &user.Name,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan function follower: %w", err)
		}
		follow.User = user
		follows = append(follows, follow)
	}

	return follows, total, nil
}

// GetUserFunctionFollows returns all functions that a user is following
func (r *FollowRepository) GetUserFunctionFollows(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*FunctionFollow, int, error) {
	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM function_follows WHERE user_id = $1`,
		userID).Scan(&total)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user function follows count: %w", err)
	}

	// Get function follows with function details
	rows, err := r.db.QueryContext(ctx, `
		SELECT ff.id, ff.user_id, ff.function_id, ff.follow_reason, ff.notify_on_new_version, ff.notify_on_rating_change, ff.notify_on_trust_change, ff.notify_on_verification, ff.created_at, ff.updated_at,
		       rf.id, rf.name, rf.description
		FROM function_follows ff
		JOIN registry_functions rf ON ff.function_id = rf.id
		WHERE ff.user_id = $1
		ORDER BY ff.created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user function follows: %w", err)
	}
	defer rows.Close()

	var follows []*FunctionFollow
	for rows.Next() {
		follow := &FunctionFollow{}
		function := &registry.RegistryFunction{}
		err := rows.Scan(
			&follow.ID, &follow.UserID, &follow.FunctionID, &follow.FollowReason, &follow.NotifyOnNewVersion, &follow.NotifyOnRatingChange, &follow.NotifyOnTrustChange, &follow.NotifyOnVerification, &follow.CreatedAt, &follow.UpdatedAt,
			&function.ID, &function.Name, &function.Description,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user function follow: %w", err)
		}
// follow.Function = function (field removed during refactor)
		follows = append(follows, follow)
	}

	return follows, total, nil
}

// GetFunctionFollowerCount returns the follower count for a function
func (r *FollowRepository) GetFunctionFollowerCount(ctx context.Context, functionID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM function_follows WHERE function_id = $1`,
		functionID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to get function follower count: %w", err)
	}

	return count, nil
}

// Helper functions

// updateUserFollowCounts updates the follower and following counts for users
func (r *FollowRepository) updateUserFollowCounts(followerID, followedUserID uuid.UUID, delta int) {
	// Update follower's following count
	r.db.Exec(`
		INSERT INTO user_follow_stats (user_id, following_count, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (user_id) DO UPDATE SET 
			following_count = user_follow_stats.following_count + EXCLUDED.following_count,
			updated_at = NOW()
	`, followerID)

	// Update followed user's follower count
	r.db.Exec(`
		INSERT INTO user_follow_stats (user_id, followers_count, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (user_id) DO UPDATE SET 
			followers_count = user_follow_stats.followers_count + EXCLUDED.followers_count,
			updated_at = NOW()
	`, followedUserID)
}

// updateFunctionFollowCount updates the follower count for a function
func (r *FollowRepository) updateFunctionFollowCount(functionID uuid.UUID, delta int) {
	r.db.Exec(`
		INSERT INTO function_follow_stats (function_id, followers_count, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (function_id) DO UPDATE SET 
			followers_count = function_follow_stats.followers_count + EXCLUDED.followers_count,
			updated_at = NOW()
	`, functionID)
}
