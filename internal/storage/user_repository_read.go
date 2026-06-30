package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	var username sql.NullString
	var companyName sql.NullString
	var role sql.NullString
	var verificationToken sql.NullString
	var verificationExpiresAt sql.NullTime
	var provider sql.NullString
	var providerID sql.NullString
	var providerData []byte
	var mfaSecret sql.NullString
	var mfaBackupCodes []byte
	var mfaLastUsed sql.NullTime
	var lastActiveNull sql.NullTime
	var nameNull sql.NullString
	var bioNull sql.NullString
	var profileNumberNull sql.NullInt64

	var row *sql.Row
	if stmt, _ := r.db.GetPreparedStatement("getUserByEmail"); stmt != nil {
		row = stmt.QueryRowContext(ctx, email)
	} else {
		row = r.db.QueryRowContext(ctx, r.db.GetStatementQuery("getUserByEmail"), email)
	}
	err := row.Scan(
		&user.ID, &user.TenantID, &username, &user.Email, &user.PasswordHash, &role, &user.EmailVerified, &companyName,
		&verificationToken, &verificationExpiresAt, &provider, &providerID, &providerData,
		&mfaSecret, &user.MFAEnabled, &mfaBackupCodes, &mfaLastUsed, &user.CreatedAt, &user.UpdatedAt,
		&nameNull, &bioNull, &lastActiveNull, &profileNumberNull)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	populateUserCoreIdent(user, username, companyName, role)
	populateUserVerification(user, verificationToken, verificationExpiresAt)
	if err := populateUserOAuth(user, provider, providerID, providerData); err != nil {
		return nil, err
	}
	if err := populateUserMFA(user, mfaSecret, mfaBackupCodes, mfaLastUsed); err != nil {
		return nil, err
	}
	populateUserNameBioLastActiveProfile(user, nameNull, bioNull, lastActiveNull, profileNumberNull)
	return user, nil
}

// GetUserByUsername retrieves a user by their username (case-insensitive)
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	var usernameNull sql.NullString
	var companyName sql.NullString
	var role sql.NullString
	var provider sql.NullString
	var providerID sql.NullString
	var providerData []byte
	var lastActiveNull sql.NullTime
	var profileNumberNull sql.NullInt64

	var nameNull sql.NullString
	var bioNull sql.NullString
	var locationNull sql.NullString
	var websiteNull sql.NullString
	var jobTitleNull sql.NullString
	var twitterURLNull sql.NullString
	var githubURLNull sql.NullString
	var linkedinURLNull sql.NullString

	var isFounderNull sql.NullBool
	var founderNumberNull sql.NullInt64
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, username, email, password_hash, role, email_verified, company_name,
		       provider, provider_id, provider_data, created_at, updated_at, name, bio,
		       location, website, job_title, twitter_url, github_url, linkedin_url, last_active_at, profile_number,
		       is_founder, founder_number
		FROM users WHERE LOWER(username) = LOWER($1)`, username).Scan(
		&user.ID, &user.TenantID, &usernameNull, &user.Email, &user.PasswordHash, &role,
		&user.EmailVerified, &companyName, &provider, &providerID, &providerData,
		&user.CreatedAt, &user.UpdatedAt, &nameNull, &bioNull,
		&locationNull, &websiteNull, &jobTitleNull, &twitterURLNull, &githubURLNull, &linkedinURLNull,
		&lastActiveNull, &profileNumberNull, &isFounderNull, &founderNumberNull)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	populateUserCoreIdent(user, usernameNull, companyName, role)
	if err := populateUserOAuth(user, provider, providerID, providerData); err != nil {
		return nil, err
	}
	populateUserNameBioLastActiveProfile(user, nameNull, bioNull, lastActiveNull, profileNumberNull)
	populateUserPublicProfileFields(user, locationNull, websiteNull, jobTitleNull, twitterURLNull, githubURLNull, linkedinURLNull)
	populateUserFounderFields(user, isFounderNull, founderNumberNull)
	return user, nil
}

// GetUserByVerificationToken retrieves a user by verification token
func (r *UserRepository) GetUserByVerificationToken(ctx context.Context, token string) (*User, error) {
	user := &User{}
	var username sql.NullString
	var companyName sql.NullString
	var role sql.NullString
	var verificationToken sql.NullString
	var verificationExpiresAt sql.NullTime
	var provider sql.NullString
	var providerID sql.NullString
	var providerData []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, username, email, password_hash, role, email_verified, company_name, verification_token, verification_expires_at, provider, provider_id, provider_data, created_at, updated_at
		FROM users WHERE verification_token = $1`, token).Scan(
		&user.ID, &user.TenantID, &username, &user.Email, &user.PasswordHash, &role, &user.EmailVerified, &companyName,
		&verificationToken, &verificationExpiresAt, &provider, &providerID, &providerData, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by verification token: %w", err)
	}

	populateUserCoreIdent(user, username, companyName, role)
	populateUserVerification(user, verificationToken, verificationExpiresAt)
	if err := populateUserOAuth(user, provider, providerID, providerData); err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	user := &User{}
	var username sql.NullString
	var companyName sql.NullString
	var role sql.NullString
	var verificationToken sql.NullString
	var verificationExpiresAt sql.NullTime
	var provider sql.NullString
	var providerID sql.NullString
	var providerData []byte
	var mfaSecret sql.NullString
	var mfaBackupCodes []byte
	var mfaLastUsed sql.NullTime
	var lastActiveNull sql.NullTime
	var nameNull sql.NullString
	var bioNull sql.NullString
	var profileNumberNull sql.NullInt64
	var isFounderNull sql.NullBool
	var founderNumberNull sql.NullInt64

	var row *sql.Row
	if stmt, _ := r.db.GetPreparedStatement("getUserByID"); stmt != nil {
		row = stmt.QueryRowContext(ctx, userID)
	} else {
		row = r.db.QueryRowContext(ctx, r.db.GetStatementQuery("getUserByID"), userID)
	}
	err := row.Scan(
		&user.ID, &user.TenantID, &username, &user.Email, &user.PasswordHash, &role, &user.EmailVerified, &companyName,
		&verificationToken, &verificationExpiresAt, &provider, &providerID, &providerData,
		&mfaSecret, &user.MFAEnabled, &mfaBackupCodes, &mfaLastUsed, &user.CreatedAt, &user.UpdatedAt,
		&nameNull, &bioNull, &lastActiveNull, &profileNumberNull, &isFounderNull, &founderNumberNull)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	populateUserCoreIdent(user, username, companyName, role)
	populateUserVerification(user, verificationToken, verificationExpiresAt)
	if err := populateUserOAuth(user, provider, providerID, providerData); err != nil {
		return nil, err
	}
	if err := populateUserMFA(user, mfaSecret, mfaBackupCodes, mfaLastUsed); err != nil {
		return nil, err
	}
	populateUserNameBioLastActiveProfile(user, nameNull, bioNull, lastActiveNull, profileNumberNull)
	populateUserFounderFields(user, isFounderNull, founderNumberNull)
	return user, nil
}

// ListUsers lists all users across all tenants (for admin dashboard).
// Tries with role column first; if that fails (e.g. column missing), falls back to minimal columns.
func (r *UserRepository) ListUsers(ctx context.Context) ([]*User, error) {
	users, err := r.listUsersWithRole(ctx)
	if err == nil {
		return users, nil
	}
	// Fallback: schema may lack "role" (e.g. pre-000004); list with minimal columns
	return r.listUsersMinimal(ctx)
}

func (r *UserRepository) listUsersWithRole(ctx context.Context) ([]*User, error) {
	query := `SELECT id, tenant_id, email, password_hash, role, created_at, updated_at FROM users ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*User
	for rows.Next() {
		user := &User{}
		var role sql.NullString
		if err := rows.Scan(&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		if role.Valid {
			user.Role = role.String
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepository) listUsersMinimal(ctx context.Context) ([]*User, error) {
	query := `SELECT id, tenant_id, email, password_hash, created_at, updated_at FROM users ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()
	var users []*User
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, nil
}

// ListUserIDsByTenant returns all user IDs for a tenant.
func (r *UserRepository) ListUserIDsByTenant(ctx context.Context, tenantID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT id FROM users WHERE tenant_id = $1`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant user ids: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan tenant user id: %w", err)
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating tenant user ids: %w", err)
	}

	return userIDs, nil
}

// GetUserBySocialProvider retrieves a user by social provider and provider ID
func (r *UserRepository) GetUserBySocialProvider(ctx context.Context, provider, providerID string) (*User, error) {
	user := &User{}
	var username sql.NullString
	var companyName sql.NullString
	var role sql.NullString
	var verificationToken sql.NullString
	var verificationExpiresAt sql.NullTime
	var providerNull sql.NullString
	var providerIDNull sql.NullString
	var providerData []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, username, email, password_hash, role, email_verified, company_name, verification_token, verification_expires_at, provider, provider_id, provider_data, created_at, updated_at
		FROM users WHERE provider = $1 AND provider_id = $2`, provider, providerID).Scan(
		&user.ID, &user.TenantID, &username, &user.Email, &user.PasswordHash, &role, &user.EmailVerified, &companyName,
		&verificationToken, &verificationExpiresAt, &providerNull, &providerIDNull, &providerData, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by social provider: %w", err)
	}

	populateUserCoreIdent(user, username, companyName, role)
	populateUserVerification(user, verificationToken, verificationExpiresAt)
	if err := populateUserOAuth(user, providerNull, providerIDNull, providerData); err != nil {
		return nil, err
	}
	return user, nil
}
