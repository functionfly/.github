package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// nullIfEmpty returns nil for empty string so DB stores NULL; otherwise returns s.
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
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

	if dob, ok := updates["date_of_birth"].(*time.Time); ok {
		setParts = append(setParts, fmt.Sprintf("date_of_birth = $%d", argIndex))
		args = append(args, dob)
		argIndex++
	}

	if passwordHash, ok := updates["password_hash"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("password_hash = $%d", argIndex))
		args = append(args, passwordHash)
		argIndex++
	}

	if verificationToken, ok := updates["verification_token"]; ok {
		setParts = append(setParts, fmt.Sprintf("verification_token = $%d", argIndex))
		args = append(args, verificationToken) // nil clears it
		argIndex++
	}

	if verificationExpiresAt, ok := updates["verification_expires_at"]; ok {
		setParts = append(setParts, fmt.Sprintf("verification_expires_at = $%d", argIndex))
		args = append(args, verificationExpiresAt) // nil clears it
		argIndex++
	}

	if provider, ok := updates["provider"].(*string); ok {
		setParts = append(setParts, fmt.Sprintf("provider = $%d", argIndex))
		args = append(args, provider)
		argIndex++
	}

	if providerID, ok := updates["provider_id"].(*string); ok {
		setParts = append(setParts, fmt.Sprintf("provider_id = $%d", argIndex))
		args = append(args, providerID)
		argIndex++
	}

	if providerData, ok := updates["provider_data"].(map[string]interface{}); ok {
		dataJSON, err := json.Marshal(providerData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal provider data: %w", err)
		}
		setParts = append(setParts, fmt.Sprintf("provider_data = $%d", argIndex))
		args = append(args, dataJSON)
		argIndex++
	}

	if len(setParts) == 0 {
		return current, nil // No updates
	}

	setParts = append(setParts, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d RETURNING id, tenant_id, username, email, password_hash, role, company_name, name, bio, location, website, job_title, twitter_url, github_url, linkedin_url, date_of_birth, created_at, updated_at",
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
	var dateOfBirthNull sql.NullTime
	err = r.db.QueryRow(query, args...).Scan(
		&updated.ID, &updated.TenantID, &usernameNull, &updated.Email, &updated.PasswordHash, &updated.Role, &companyNameNull, &nameNull, &bioNull,
		&locationNull, &websiteNull, &jobTitleNull, &twitterURLNull, &githubURLNull, &linkedinURLNull,
		&dateOfBirthNull,
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
	if err == nil && dateOfBirthNull.Valid {
		t := dateOfBirthNull.Time
		updated.DateOfBirth = &t
	}

	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updated, nil
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
