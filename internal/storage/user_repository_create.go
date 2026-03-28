package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

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

	var profileNumber int
	err := r.db.QueryRow(`
		INSERT INTO users (id, tenant_id, username, email, password_hash, role, email_verified, company_name, verification_token, verification_expires_at, provider, provider_id, provider_data, mfa_secret, mfa_enabled, mfa_backup_codes, mfa_last_used, created_at, updated_at, profile_number)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, nextval('users_profile_number_seq'))
		RETURNING profile_number`,
		user.ID, user.TenantID, user.Username, user.Email, user.PasswordHash, user.Role, user.EmailVerified, user.CompanyName, user.VerificationToken, user.VerificationExpiresAt, user.Provider, user.ProviderID, providerDataJSON, user.MFASecret, user.MFAEnabled, mfaBackupCodesJSON, user.MFALastUsed, user.CreatedAt, user.UpdatedAt).Scan(&profileNumber)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.ProfileNumber = &profileNumber
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

	var profileNumber int
	err := r.db.QueryRow(`
		INSERT INTO users (id, tenant_id, username, email, password_hash, role, email_verified, company_name, verification_token, verification_expires_at, provider, provider_id, provider_data, mfa_secret, mfa_enabled, mfa_backup_codes, mfa_last_used, created_at, updated_at, profile_number)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, nextval('users_profile_number_seq'))
		RETURNING profile_number`,
		user.ID, user.TenantID, user.Username, user.Email, nil, user.Role, user.EmailVerified, user.CompanyName, user.VerificationToken, user.VerificationExpiresAt, user.Provider, user.ProviderID, providerDataJSON, user.MFASecret, user.MFAEnabled, mfaBackupCodesJSON, user.MFALastUsed, user.CreatedAt, user.UpdatedAt).Scan(&profileNumber)

	if err != nil {
		return nil, fmt.Errorf("failed to create user with social auth: %w", err)
	}

	user.ProfileNumber = &profileNumber
	return user, nil
}

// CreateUserWithRole creates a user with a specific role (e.g. admin users).
// EmailVerified is taken from the user struct so admins can be created pre-verified and log in immediately.
func (r *UserRepository) CreateUserWithRole(ctx context.Context, user *User) (*User, error) {
	query := `
		INSERT INTO users (id, tenant_id, username, email, password_hash, role, email_verified, company_name, created_at, updated_at, profile_number)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW(), nextval('users_profile_number_seq'))
		RETURNING id, tenant_id, username, email, password_hash, role, company_name, created_at, updated_at, profile_number`

	var username sql.NullString
	var companyName sql.NullString
	var profileNumber int
	err := r.db.QueryRow(query, user.ID, user.TenantID, user.Username, user.Email, user.PasswordHash, user.Role, user.EmailVerified, user.CompanyName).Scan(
		&user.ID, &user.TenantID, &username, &user.Email, &user.PasswordHash, &user.Role, &companyName, &user.CreatedAt, &user.UpdatedAt, &profileNumber)
	if err == nil && username.Valid {
		user.Username = &username.String
	}
	if err == nil && companyName.Valid {
		user.CompanyName = &companyName.String
	}
	if err == nil {
		user.ProfileNumber = &profileNumber
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
