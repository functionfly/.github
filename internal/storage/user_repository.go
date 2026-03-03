package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository handles user-related database operations
type UserRepository struct {
	db *PostgresDB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *PostgresDB) *UserRepository {
	return &UserRepository{db: db}
}

// nullIfEmpty returns nil for empty string so DB stores NULL; otherwise returns s.
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(email, passwordHash string, tenantID uuid.UUID) (*User, error) {
	user := &User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Email:         email,
		PasswordHash:  passwordHash,
		EmailVerified: true, // Auto-verify for setup
		MFAEnabled:    false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Serialize ProviderData and MFABackupCodes to JSON for SQL
	providerDataJSON, _ := json.Marshal(user.ProviderData)
	mfaBackupCodesJSON, _ := json.Marshal(user.MFABackupCodes)

	_, err := r.db.Exec(`
		INSERT INTO users (id, tenant_id, username, email, password_hash, role, email_verified, company_name, verification_token, verification_expires_at, provider, provider_id, provider_data, mfa_secret, mfa_enabled, mfa_backup_codes, mfa_last_used, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		user.ID, user.TenantID, user.Username, user.Email, user.PasswordHash, user.Role, user.EmailVerified, user.CompanyName, user.VerificationToken, user.VerificationExpiresAt, user.Provider, user.ProviderID, providerDataJSON, user.MFASecret, user.MFAEnabled, mfaBackupCodesJSON, user.MFALastUsed, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// CreateUserWithSocialAuth creates a new user with social authentication
func (r *UserRepository) CreateUserWithSocialAuth(email string, tenantID uuid.UUID, provider, providerID string, providerData map[string]interface{}) (*User, error) {
	user := &User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Email:         email,
		EmailVerified: true, // Social auth users are pre-verified
		Provider:      &provider,
		ProviderID:    &providerID,
		ProviderData:  providerData,
		MFAEnabled:    false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Serialize ProviderData and MFABackupCodes to JSON for SQL
	providerDataJSON, _ := json.Marshal(user.ProviderData)
	mfaBackupCodesJSON, _ := json.Marshal(user.MFABackupCodes)

	_, err := r.db.Exec(`
		INSERT INTO users (id, tenant_id, username, email, password_hash, role, email_verified, company_name, verification_token, verification_expires_at, provider, provider_id, provider_data, mfa_secret, mfa_enabled, mfa_backup_codes, mfa_last_used, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		user.ID, user.TenantID, user.Username, user.Email, nil, user.Role, user.EmailVerified, user.CompanyName, user.VerificationToken, user.VerificationExpiresAt, user.Provider, user.ProviderID, providerDataJSON, user.MFASecret, user.MFAEnabled, mfaBackupCodesJSON, user.MFALastUsed, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user with social auth: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*User, error) {
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
	var nameNull sql.NullString
	var bioNull sql.NullString

	var row *sql.Row
	if stmt, _ := r.db.GetPreparedStatement("getUserByEmail"); stmt != nil {
		row = stmt.QueryRow(email)
	} else {
		row = r.db.QueryRowContext(context.Background(), r.db.GetStatementQuery("getUserByEmail"), email)
	}
	err := row.Scan(
		&user.ID, &user.TenantID, &username, &user.Email, &user.PasswordHash, &role, &user.EmailVerified, &companyName,
		&verificationToken, &verificationExpiresAt, &provider, &providerID, &providerData,
		&mfaSecret, &user.MFAEnabled, &mfaBackupCodes, &mfaLastUsed, &user.CreatedAt, &user.UpdatedAt,
		&nameNull, &bioNull)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if username.Valid {
		user.Username = &username.String
	}
	if companyName.Valid {
		user.CompanyName = &companyName.String
	}
	if role.Valid {
		user.Role = role.String
	}
	if verificationToken.Valid {
		user.VerificationToken = &verificationToken.String
	}
	if verificationExpiresAt.Valid {
		user.VerificationExpiresAt = &verificationExpiresAt.Time
	}
	if provider.Valid {
		user.Provider = &provider.String
	}
	if providerID.Valid {
		user.ProviderID = &providerID.String
	}
	if len(providerData) > 0 {
		// Parse JSON data
		if err := json.Unmarshal(providerData, &user.ProviderData); err != nil {
			return nil, fmt.Errorf("failed to parse provider data: %w", err)
		}
	}
	if mfaSecret.Valid {
		user.MFASecret = &mfaSecret.String
	}
	if len(mfaBackupCodes) > 0 {
		if err := json.Unmarshal(mfaBackupCodes, &user.MFABackupCodes); err != nil {
			return nil, fmt.Errorf("failed to parse MFA backup codes: %w", err)
		}
	}
	if mfaLastUsed.Valid {
		user.MFALastUsed = &mfaLastUsed.Time
	}
	if nameNull.Valid {
		user.Name = nameNull.String
	}
	if bioNull.Valid {
		user.Bio = &bioNull.String
	}

	return user, nil
}

// GetUserByUsername retrieves a user by their username (case-insensitive)
func (r *UserRepository) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	var usernameNull sql.NullString
	var companyName sql.NullString
	var role sql.NullString
	var provider sql.NullString
	var providerID sql.NullString
	var providerData []byte

	var nameNull sql.NullString
	var bioNull sql.NullString
	var locationNull sql.NullString
	var websiteNull sql.NullString
	var jobTitleNull sql.NullString
	var twitterURLNull sql.NullString
	var githubURLNull sql.NullString
	var linkedinURLNull sql.NullString

	err := r.db.QueryRowContext(context.Background(), `
		SELECT id, tenant_id, username, email, password_hash, role, email_verified, company_name,
		       provider, provider_id, provider_data, created_at, updated_at, name, bio,
		       location, website, job_title, twitter_url, github_url, linkedin_url
		FROM users WHERE LOWER(username) = LOWER($1)`, username).Scan(
		&user.ID, &user.TenantID, &usernameNull, &user.Email, &user.PasswordHash, &role,
		&user.EmailVerified, &companyName, &provider, &providerID, &providerData,
		&user.CreatedAt, &user.UpdatedAt, &nameNull, &bioNull,
		&locationNull, &websiteNull, &jobTitleNull, &twitterURLNull, &githubURLNull, &linkedinURLNull)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	if usernameNull.Valid {
		user.Username = &usernameNull.String
	}
	if companyName.Valid {
		user.CompanyName = &companyName.String
	}
	if role.Valid {
		user.Role = role.String
	}
	if provider.Valid {
		user.Provider = &provider.String
	}
	if providerID.Valid {
		user.ProviderID = &providerID.String
	}
	if len(providerData) > 0 {
		if err := json.Unmarshal(providerData, &user.ProviderData); err != nil {
			return nil, fmt.Errorf("failed to parse provider data: %w", err)
		}
	}
	if nameNull.Valid {
		user.Name = nameNull.String
	}
	if bioNull.Valid {
		user.Bio = &bioNull.String
	}
	if locationNull.Valid {
		user.Location = &locationNull.String
	}
	if websiteNull.Valid {
		user.Website = &websiteNull.String
	}
	if jobTitleNull.Valid {
		user.JobTitle = &jobTitleNull.String
	}
	if twitterURLNull.Valid {
		user.TwitterURL = &twitterURLNull.String
	}
	if githubURLNull.Valid {
		user.GithubURL = &githubURLNull.String
	}
	if linkedinURLNull.Valid {
		user.LinkedInURL = &linkedinURLNull.String
	}

	return user, nil
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

// GetUserByVerificationToken retrieves a user by verification token
func (r *UserRepository) GetUserByVerificationToken(token string) (*User, error) {
	user := &User{}
	var username sql.NullString
	var companyName sql.NullString
	var role sql.NullString
	var verificationToken sql.NullString
	var verificationExpiresAt sql.NullTime
	var provider sql.NullString
	var providerID sql.NullString
	var providerData []byte
	err := r.db.QueryRow(`
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

	if username.Valid {
		user.Username = &username.String
	}
	if companyName.Valid {
		user.CompanyName = &companyName.String
	}
	if role.Valid {
		user.Role = role.String
	}
	if verificationToken.Valid {
		user.VerificationToken = &verificationToken.String
	}
	if verificationExpiresAt.Valid {
		user.VerificationExpiresAt = &verificationExpiresAt.Time
	}
	if provider.Valid {
		user.Provider = &provider.String
	}
	if providerID.Valid {
		user.ProviderID = &providerID.String
	}
	if len(providerData) > 0 {
		if err := json.Unmarshal(providerData, &user.ProviderData); err != nil {
			return nil, fmt.Errorf("failed to parse provider data: %w", err)
		}
	}

	return user, nil
}

// UpdateUserEmailVerification updates user email verification status
func (r *UserRepository) UpdateUserEmailVerification(ctx context.Context, userID uuid.UUID, verified bool, token *string, expiresAt *time.Time) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET email_verified = $1, verification_token = $2, verification_expires_at = $3, updated_at = $4
		WHERE id = $5`,
		verified, token, expiresAt, time.Now(), userID)

	if err != nil {
		return fmt.Errorf("failed to update user email verification: %w", err)
	}

	return nil
}

// CreateUserWithRole creates a user with a specific role (e.g. admin users).
// EmailVerified is taken from the user struct so admins can be created pre-verified and log in immediately.
func (r *UserRepository) CreateUserWithRole(ctx context.Context, user *User) (*User, error) {
	query := `
		INSERT INTO users (id, tenant_id, username, email, password_hash, role, email_verified, company_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, tenant_id, username, email, password_hash, role, company_name, created_at, updated_at`

	var username sql.NullString
	var companyName sql.NullString
	err := r.db.QueryRow(query, user.ID, user.TenantID, user.Username, user.Email, user.PasswordHash, user.Role, user.EmailVerified, user.CompanyName).Scan(
		&user.ID, &user.TenantID, &username, &user.Email, &user.PasswordHash, &user.Role, &companyName, &user.CreatedAt, &user.UpdatedAt)
	if err == nil && username.Valid {
		user.Username = &username.String
	}
	if err == nil && companyName.Valid {
		user.CompanyName = &companyName.String
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(userID uuid.UUID) (*User, error) {
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
	var nameNull sql.NullString
	var bioNull sql.NullString

	var row *sql.Row
	if stmt, _ := r.db.GetPreparedStatement("getUserByID"); stmt != nil {
		row = stmt.QueryRow(userID)
	} else {
		row = r.db.QueryRowContext(context.Background(), r.db.GetStatementQuery("getUserByID"), userID)
	}
	err := row.Scan(
		&user.ID, &user.TenantID, &username, &user.Email, &user.PasswordHash, &role, &user.EmailVerified, &companyName,
		&verificationToken, &verificationExpiresAt, &provider, &providerID, &providerData,
		&mfaSecret, &user.MFAEnabled, &mfaBackupCodes, &mfaLastUsed, &user.CreatedAt, &user.UpdatedAt,
		&nameNull, &bioNull)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	if username.Valid {
		user.Username = &username.String
	}
	if companyName.Valid {
		user.CompanyName = &companyName.String
	}
	if role.Valid {
		user.Role = role.String
	}
	if verificationToken.Valid {
		user.VerificationToken = &verificationToken.String
	}
	if verificationExpiresAt.Valid {
		user.VerificationExpiresAt = &verificationExpiresAt.Time
	}
	if provider.Valid {
		user.Provider = &provider.String
	}
	if providerID.Valid {
		user.ProviderID = &providerID.String
	}
	if len(providerData) > 0 {
		if err := json.Unmarshal(providerData, &user.ProviderData); err != nil {
			return nil, fmt.Errorf("failed to parse provider data: %w", err)
		}
	}
	if mfaSecret.Valid {
		user.MFASecret = &mfaSecret.String
	}
	if len(mfaBackupCodes) > 0 {
		if err := json.Unmarshal(mfaBackupCodes, &user.MFABackupCodes); err != nil {
			return nil, fmt.Errorf("failed to parse MFA backup codes: %w", err)
		}
	}
	if mfaLastUsed.Valid {
		user.MFALastUsed = &mfaLastUsed.Time
	}
	if nameNull.Valid {
		user.Name = nameNull.String
	}
	if bioNull.Valid {
		user.Bio = &bioNull.String
	}

	return user, nil
}

// ListUsers lists all users
func (r *UserRepository) ListUsers() ([]*User, error) {
	query := `SELECT id, tenant_id, username, email, password_hash, role, company_name, created_at, updated_at FROM users ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		var username sql.NullString
		var companyName sql.NullString
		var role sql.NullString
		err := rows.Scan(&user.ID, &user.TenantID, &username, &user.Email, &user.PasswordHash, &role, &companyName, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		if username.Valid {
			user.Username = &username.String
		}
		if companyName.Valid {
			user.CompanyName = &companyName.String
		}
		if role.Valid {
			user.Role = role.String
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdateUser updates user fields dynamically
func (r *UserRepository) UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*User, error) {
	// Get current user
	current, err := r.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	if current == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if email, ok := updates["email"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, email)
		argIndex++
	}

	if username, ok := updates["username"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, username)
		argIndex++
	}

	if companyName, ok := updates["company_name"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("company_name = $%d", argIndex))
		args = append(args, companyName)
		argIndex++
	}

	if role, ok := updates["role"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, role)
		argIndex++
	}

	if name, ok := updates["name"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, name)
		argIndex++
	}
	if bio, ok := updates["bio"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("bio = $%d", argIndex))
		args = append(args, bio)
		argIndex++
	}
	if location, ok := updates["location"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("location = $%d", argIndex))
		args = append(args, nullIfEmpty(location))
		argIndex++
	}
	if website, ok := updates["website"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("website = $%d", argIndex))
		args = append(args, nullIfEmpty(website))
		argIndex++
	}
	if jobTitle, ok := updates["job_title"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("job_title = $%d", argIndex))
		args = append(args, nullIfEmpty(jobTitle))
		argIndex++
	}
	if twitterURL, ok := updates["twitter_url"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("twitter_url = $%d", argIndex))
		args = append(args, nullIfEmpty(twitterURL))
		argIndex++
	}
	if githubURL, ok := updates["github_url"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("github_url = $%d", argIndex))
		args = append(args, nullIfEmpty(githubURL))
		argIndex++
	}
	if linkedinURL, ok := updates["linkedin_url"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("linkedin_url = $%d", argIndex))
		args = append(args, nullIfEmpty(linkedinURL))
		argIndex++
	}

	if len(setParts) == 0 {
		return current, nil // No updates
	}

	setParts = append(setParts, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d RETURNING id, tenant_id, username, email, password_hash, role, company_name, name, bio, location, website, job_title, twitter_url, github_url, linkedin_url, created_at, updated_at",
		strings.Join(setParts, ", "), argIndex)

	args = append(args, userID)

	updated := &User{}
	var usernameNull sql.NullString
	var companyNameNull sql.NullString
	var nameNull sql.NullString
	var bioNull sql.NullString
	var locationNull sql.NullString
	var websiteNull sql.NullString
	var jobTitleNull sql.NullString
	var twitterURLNull sql.NullString
	var githubURLNull sql.NullString
	var linkedinURLNull sql.NullString
	err = r.db.QueryRow(query, args...).Scan(
		&updated.ID, &updated.TenantID, &usernameNull, &updated.Email, &updated.PasswordHash, &updated.Role, &companyNameNull, &nameNull, &bioNull,
		&locationNull, &websiteNull, &jobTitleNull, &twitterURLNull, &githubURLNull, &linkedinURLNull,
		&updated.CreatedAt, &updated.UpdatedAt)
	if err == nil && usernameNull.Valid {
		updated.Username = &usernameNull.String
	}
	if err == nil && companyNameNull.Valid {
		updated.CompanyName = &companyNameNull.String
	}
	if err == nil && nameNull.Valid {
		updated.Name = nameNull.String
	}
	if err == nil && bioNull.Valid {
		updated.Bio = &bioNull.String
	}
	if err == nil && locationNull.Valid {
		updated.Location = &locationNull.String
	}
	if err == nil && websiteNull.Valid {
		updated.Website = &websiteNull.String
	}
	if err == nil && jobTitleNull.Valid {
		updated.JobTitle = &jobTitleNull.String
	}
	if err == nil && twitterURLNull.Valid {
		updated.TwitterURL = &twitterURLNull.String
	}
	if err == nil && githubURLNull.Valid {
		updated.GithubURL = &githubURLNull.String
	}
	if err == nil && linkedinURLNull.Valid {
		updated.LinkedInURL = &linkedinURLNull.String
	}

	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updated, nil
}

// GetUserBySocialProvider retrieves a user by social provider and provider ID
func (r *UserRepository) GetUserBySocialProvider(provider, providerID string) (*User, error) {
	user := &User{}
	var username sql.NullString
	var companyName sql.NullString
	var role sql.NullString
	var verificationToken sql.NullString
	var verificationExpiresAt sql.NullTime
	var providerNull sql.NullString
	var providerIDNull sql.NullString
	var providerData []byte
	err := r.db.QueryRow(`
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

	if username.Valid {
		user.Username = &username.String
	}
	if companyName.Valid {
		user.CompanyName = &companyName.String
	}
	if role.Valid {
		user.Role = role.String
	}
	if verificationToken.Valid {
		user.VerificationToken = &verificationToken.String
	}
	if verificationExpiresAt.Valid {
		user.VerificationExpiresAt = &verificationExpiresAt.Time
	}
	if providerNull.Valid {
		user.Provider = &providerNull.String
	}
	if providerIDNull.Valid {
		user.ProviderID = &providerIDNull.String
	}
	if len(providerData) > 0 {
		if err := json.Unmarshal(providerData, &user.ProviderData); err != nil {
			return nil, fmt.Errorf("failed to parse provider data: %w", err)
		}
	}

	return user, nil
}

// UpdateUserProviderData updates the provider data for a user
func (r *UserRepository) UpdateUserProviderData(userID uuid.UUID, providerData map[string]interface{}) error {
	dataJSON, err := json.Marshal(providerData)
	if err != nil {
		return fmt.Errorf("failed to marshal provider data: %w", err)
	}

	_, err = r.db.Exec(`
		UPDATE users SET provider_data = $1, updated_at = NOW() WHERE id = $2`,
		dataJSON, userID)

	if err != nil {
		return fmt.Errorf("failed to update user provider data: %w", err)
	}

	return nil
}

// MFA operations
func (r *UserRepository) UpdateUserMFA(userID uuid.UUID, secret *string, enabled bool, backupCodes []string, lastUsed *time.Time) error {
	backupCodesJSON, err := json.Marshal(backupCodes)
	if err != nil {
		return fmt.Errorf("failed to marshal backup codes: %w", err)
	}

	_, err = r.db.Exec(`
		UPDATE users SET mfa_secret = $1, mfa_enabled = $2, mfa_backup_codes = $3, mfa_last_used = $4, updated_at = NOW() WHERE id = $5`,
		secret, enabled, backupCodesJSON, lastUsed, userID)

	if err != nil {
		return fmt.Errorf("failed to update user MFA: %w", err)
	}

	return nil
}

func (r *UserRepository) UpdateUserMFAEnabled(userID uuid.UUID, enabled bool) error {
	_, err := r.db.Exec(`
		UPDATE users SET mfa_enabled = $1, updated_at = NOW() WHERE id = $2`,
		enabled, userID)

	if err != nil {
		return fmt.Errorf("failed to update user MFA enabled: %w", err)
	}

	return nil
}

func (r *UserRepository) UpdateUserMFABackupCodes(userID uuid.UUID, backupCodes []string) error {
	backupCodesJSON, err := json.Marshal(backupCodes)
	if err != nil {
		return fmt.Errorf("failed to marshal backup codes: %w", err)
	}

	_, err = r.db.Exec(`
		UPDATE users SET mfa_backup_codes = $1, updated_at = NOW() WHERE id = $2`,
		backupCodesJSON, userID)

	if err != nil {
		return fmt.Errorf("failed to update user MFA backup codes: %w", err)
	}

	return nil
}

func (r *UserRepository) UpdateUserMFALastUsed(userID uuid.UUID, lastUsed *time.Time) error {
	_, err := r.db.Exec(`
		UPDATE users SET mfa_last_used = $1, updated_at = NOW() WHERE id = $2`,
		lastUsed, userID)

	if err != nil {
		return fmt.Errorf("failed to update user MFA last used: %w", err)
	}

	return nil
}

// UpdateUserSettings updates the settings JSONB field for a user
func (r *UserRepository) UpdateUserSettings(userID uuid.UUID, settings map[string]interface{}) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal user settings: %w", err)
	}

	_, err = r.db.Exec(`
		UPDATE users SET settings = $1, updated_at = NOW() WHERE id = $2`,
		settingsJSON, userID)

	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	return nil
}

// GetUserSettings retrieves the settings JSONB field for a user
func (r *UserRepository) GetUserSettings(userID uuid.UUID) (map[string]interface{}, error) {
	user, err := r.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.Settings == nil {
		return make(map[string]interface{}), nil
	}

	return user.Settings, nil
}

// HashPassword securely hashes a password using bcrypt
func (r *UserRepository) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against the stored hash
func (r *UserRepository) VerifyPassword(userID uuid.UUID, password string) (bool, error) {
	var storedHash string
	err := r.db.QueryRow(`
		SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&storedHash)

	if err == sql.ErrNoRows {
		return false, fmt.Errorf("user not found")
	}
	if err != nil {
		return false, fmt.Errorf("failed to get password hash: %w", err)
	}

	// Verify the password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to verify password: %w", err)
	}

	return true, nil
}
