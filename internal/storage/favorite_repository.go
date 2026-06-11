package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FavoriteRepository handles favorite-related database operations
type FavoriteRepository struct {
	db *PostgresDB
}

// NewFavoriteRepository creates a new favorite repository
func NewFavoriteRepository(db *PostgresDB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

// AddFavorite adds a function to user's favorites
func (r *FavoriteRepository) AddFavorite(ctx context.Context, userID, functionID uuid.UUID, position int) (*FunctionFavorite, error) {
	favorite := &FunctionFavorite{
		ID:         uuid.New(),
		UserID:     userID,
		FunctionID: functionID,
		Position:   position,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO function_favorites (id, user_id, function_id, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, function_id) DO NOTHING`,
		favorite.ID, favorite.UserID, favorite.FunctionID, favorite.Position, favorite.CreatedAt, favorite.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to add favorite: %w", err)
	}

	return favorite, nil
}

// RemoveFavorite removes a function from user's favorites
func (r *FavoriteRepository) RemoveFavorite(ctx context.Context, userID, functionID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM function_favorites WHERE user_id = $1 AND function_id = $2`,
		userID, functionID)

	if err != nil {
		return fmt.Errorf("failed to remove favorite: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("favorite not found")
	}

	return nil
}

// IsFavorite checks if a function is in user's favorites
func (r *FavoriteRepository) IsFavorite(ctx context.Context, userID, functionID uuid.UUID) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM function_favorites WHERE user_id = $1 AND function_id = $2`,
		userID, functionID).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check favorite: %w", err)
	}

	return count > 0, nil
}

// GetUserFavorites returns all favorited functions for a user
func (r *FavoriteRepository) GetUserFavorites(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*FunctionFavorite, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM function_favorites WHERE user_id = $1`,
		userID).Scan(&total)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get favorites count: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, function_id, position, created_at, updated_at
		FROM function_favorites
		WHERE user_id = $1
		ORDER BY position ASC, created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get favorites: %w", err)
	}
	defer rows.Close()

	var favorites []*FunctionFavorite
	for rows.Next() {
		fav := &FunctionFavorite{}
		err := rows.Scan(
			&fav.ID, &fav.UserID, &fav.FunctionID, &fav.Position, &fav.CreatedAt, &fav.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan favorite: %w", err)
		}
		favorites = append(favorites, fav)
	}

	return favorites, total, nil
}

// GetFavoriteCount returns the favorite count for a function
func (r *FavoriteRepository) GetFavoriteCount(ctx context.Context, functionID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM function_favorites WHERE function_id = $1`,
		functionID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to get favorite count: %w", err)
	}

	return count, nil
}

// UpdateFavoritePosition updates the position of a favorite in user's list
func (r *FavoriteRepository) UpdateFavoritePosition(ctx context.Context, userID, functionID uuid.UUID, position int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE function_favorites SET position = $1, updated_at = NOW()
		WHERE user_id = $2 AND function_id = $3`,
		position, userID, functionID)

	if err != nil {
		return fmt.Errorf("failed to update favorite position: %w", err)
	}

	return nil
}

// GetFavoriteByFunction returns a user's favorite entry for a specific function
func (r *FavoriteRepository) GetFavoriteByFunction(ctx context.Context, userID, functionID uuid.UUID) (*FunctionFavorite, error) {
	var fav FunctionFavorite
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, function_id, position, created_at, updated_at
		FROM function_favorites
		WHERE user_id = $1 AND function_id = $2`,
		userID, functionID).Scan(
		&fav.ID, &fav.UserID, &fav.FunctionID, &fav.Position, &fav.CreatedAt, &fav.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get favorite: %w", err)
	}

	return &fav, nil
}
