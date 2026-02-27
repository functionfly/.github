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
		INSERT INTO users (id, tenant_id, email, password_hash, role, email_verified, verification_token, verification_expires_at, provider, provider_id, provider_data, mfa_secret, mfa_enabled, mfa_backup_codes, mfa_last_used, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		user.ID, user.TenantID, user.Email, user.PasswordHash, user.Role, user.EmailVerified, user.VerificationToken, user.VerificationExpiresAt, user.Provider, user.ProviderID, providerDataJSON, user.MFASecret, user.MFAEnabled, mfaBackupCodesJSON, user.MFALastUsed, user.CreatedAt, user.UpdatedAt)

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
		INSERT INTO users (id, tenant_id, email, password_hash, role, email_verified, verification_token, verification_expires_at, provider, provider_id, provider_data, mfa_secret, mfa_enabled, mfa_backup_codes, mfa_last_used, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		user.ID, user.TenantID, user.Email, nil, user.Role, user.EmailVerified, user.VerificationToken, user.VerificationExpiresAt, user.Provider, user.ProviderID, providerDataJSON, user.MFASecret, user.MFAEnabled, mfaBackupCodesJSON, user.MFALastUsed, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user with social auth: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	var role sql.NullString
	var verificationToken sql.NullString
	var verificationExpiresAt sql.NullTime
	var provider sql.NullString
	var providerID sql.NullString
	var providerData []byte
	var mfaSecret sql.NullString
	var mfaBackupCodes []byte
	var mfaLastUsed sql.NullTime

	var row *sql.Row
	if stmt, _ := r.db.GetPreparedStatement("getUserByEmail"); stmt != nil {
		row = stmt.QueryRow(email)
	} else {
		row = r.db.QueryRowContext(context.Background(), r.db.GetStatementQuery("getUserByEmail"), email)
	}
	err := row.Scan(
		&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &role, &user.EmailVerified,
		&verificationToken, &verificationExpiresAt, &provider, &providerID, &providerData,
		&mfaSecret, &user.MFAEnabled, &mfaBackupCodes, &mfaLastUsed, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
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

	return user, nil
}

// GetUserByVerificationToken retrieves a user by verification token
func (r *UserRepository) GetUserByVerificationToken(token string) (*User, error) {
	user := &User{}
	var role sql.NullString
	var verificationToken sql.NullString
	var verificationExpiresAt sql.NullTime
	var provider sql.NullString
	var providerID sql.NullString
	var providerData []byte
	err := r.db.QueryRow(`
		SELECT id, tenant_id, email, password_hash, role, email_verified, verification_token, verification_expires_at, provider, provider_id, provider_data, created_at, updated_at
		FROM users WHERE verification_token = $1`, token).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &role, &user.EmailVerified,
		&verificationToken, &verificationExpiresAt, &provider, &providerID, &providerData, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by verification token: %w", err)
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
		INSERT INTO users (id, tenant_id, email, password_hash, role, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, tenant_id, email, password_hash, role, created_at, updated_at`

	err := r.db.QueryRow(query, user.ID, user.TenantID, user.Email, user.PasswordHash, user.Role, user.EmailVerified).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(userID uuid.UUID) (*User, error) {
	user := &User{}
	var role sql.NullString
	var verificationToken sql.NullString
	var verificationExpiresAt sql.NullTime
	var provider sql.NullString
	var providerID sql.NullString
	var providerData []byte
	var mfaSecret sql.NullString
	var mfaBackupCodes []byte
	var mfaLastUsed sql.NullTime

	var row *sql.Row
	if stmt, _ := r.db.GetPreparedStatement("getUserByID"); stmt != nil {
		row = stmt.QueryRow(userID)
	} else {
		row = r.db.QueryRowContext(context.Background(), r.db.GetStatementQuery("getUserByID"), userID)
	}
	err := row.Scan(
		&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &role, &user.EmailVerified,
		&verificationToken, &verificationExpiresAt, &provider, &providerID, &providerData,
		&mfaSecret, &user.MFAEnabled, &mfaBackupCodes, &mfaLastUsed, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
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

	return user, nil
}

// ListUsers lists all users
func (r *UserRepository) ListUsers() ([]*User, error) {
	query := `SELECT id, tenant_id, email, password_hash, role, created_at, updated_at FROM users ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		var role sql.NullString
		err := rows.Scan(&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
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

	if role, ok := updates["role"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, role)
		argIndex++
	}

	if len(setParts) == 0 {
		return current, nil // No updates
	}

	setParts = append(setParts, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d RETURNING id, tenant_id, email, password_hash, role, created_at, updated_at",
		strings.Join(setParts, ", "), argIndex)

	args = append(args, userID)

	updated := &User{}
	err = r.db.QueryRow(query, args...).Scan(
		&updated.ID, &updated.TenantID, &updated.Email, &updated.PasswordHash, &updated.Role, &updated.CreatedAt, &updated.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updated, nil
}

// GetUserBySocialProvider retrieves a user by social provider and provider ID
func (r *UserRepository) GetUserBySocialProvider(provider, providerID string) (*User, error) {
	user := &User{}
	var role sql.NullString
	var verificationToken sql.NullString
	var verificationExpiresAt sql.NullTime
	var providerNull sql.NullString
	var providerIDNull sql.NullString
	var providerData []byte
	err := r.db.QueryRow(`
		SELECT id, tenant_id, email, password_hash, role, email_verified, verification_token, verification_expires_at, provider, provider_id, provider_data, created_at, updated_at
		FROM users WHERE provider = $1 AND provider_id = $2`, provider, providerID).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &role, &user.EmailVerified,
		&verificationToken, &verificationExpiresAt, &providerNull, &providerIDNull, &providerData, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by social provider: %w", err)
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
