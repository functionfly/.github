package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// GetUserForPublicProfile resolves a user for GET /users/{login}: by username first, then by
// unambiguous email local-part (before @). Used when some accounts have no username set (e.g. admin).
func (r *UserRepository) GetUserForPublicProfile(login string) (*User, error) {
	login = strings.TrimSpace(login)
	if login == "" {
		return nil, nil
	}
	u, err := r.GetUserByUsername(login)
	if err != nil {
		return nil, err
	}
	if u != nil {
		return u, nil
	}
	rows, err := r.db.QueryContext(context.Background(), `
		SELECT id FROM users
		WHERE deactivated_at IS NULL
		  AND LOWER(SPLIT_PART(email::text, '@', 1)) = LOWER($1)
		LIMIT 2`, login)
	if err != nil {
		return nil, fmt.Errorf("get user by email local part: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan user id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) != 1 {
		return nil, nil
	}
	return r.GetUserByID(ids[0])
}

func escapeILikePrefix(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// SearchUsersByUsernamePrefix returns active users matching a typed fragment on username, display name, or email.
func (r *UserRepository) SearchUsersByUsernamePrefix(ctx context.Context, prefix string, limit int) ([]UserSearchHit, error) {
	if limit < 1 {
		limit = 8
	}
	if limit > 20 {
		limit = 20
	}
	prefix = strings.TrimSpace(prefix)
	if len(prefix) < 2 {
		return nil, nil
	}
	pattern := "%" + escapeILikePrefix(prefix) + "%"

	rows, err := r.db.QueryContext(ctx, `
		SELECT id,
			COALESCE(NULLIF(TRIM(username::text), ''), SPLIT_PART(email::text, '@', 1)) AS handle,
			COALESCE(NULLIF(name, ''), '')
		FROM users
		WHERE deactivated_at IS NULL
		  AND (
			(username IS NOT NULL AND TRIM(username::text) <> '' AND username::text ILIKE $1 ESCAPE '\')
			OR COALESCE(NULLIF(TRIM(name::text), ''), '') ILIKE $1 ESCAPE '\'
			OR email::text ILIKE $1 ESCAPE '\'
		  )
		ORDER BY handle ASC
		LIMIT $2`, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search users by username prefix: %w", err)
	}
	defer rows.Close()

	var out []UserSearchHit
	for rows.Next() {
		var hit UserSearchHit
		if err := rows.Scan(&hit.ID, &hit.Username, &hit.Name); err != nil {
			return nil, fmt.Errorf("scan user search hit: %w", err)
		}
		if strings.TrimSpace(hit.Username) == "" {
			continue
		}
		out = append(out, hit)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user search: %w", err)
	}
	return out, nil
}

// IsUsernameReserved checks if a username is reserved in the database
func (r *UserRepository) IsUsernameReserved(username string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(context.Background(), `
		SELECT EXISTS(SELECT 1 FROM reserved_usernames WHERE LOWER(username) = LOWER($1))`,
		username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check reserved username: %w", err)
	}
	return exists, nil
}
