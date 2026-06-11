package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// IPAllowlist represents an IP allowlist configuration for a tenant
type IPAllowlist struct {
	ID                      uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID                uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	Name                    string    `json:"name" gorm:"size:255;not nil"`
	Description             *string   `json:"description" gorm:"type:text"`
	DefaultPolicy           string    `json:"default_policy" gorm:"size:20;default:deny"` // 'allow' or 'deny'
	MFARequiredForUnknownIP bool      `json:"mfa_required_for_unknown_ip" gorm:"default:true"`
	CreatedAt               time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt               time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// IPAllowlistEntry represents a single IP address or CIDR range in an allowlist
type IPAllowlistEntry struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AllowlistID uuid.UUID `json:"allowlist_id" gorm:"type:uuid;not null"`
	Type        string    `json:"type" gorm:"size:20;not null"` // 'ip' or 'cidr'
	Value       string    `json:"value" gorm:"size:100;not null"`
	Description *string   `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// IPAllowlistRepository handles IP allowlist database operations
type IPAllowlistRepository struct {
	db *PostgresDB
}

// NewIPAllowlistRepository creates a new IP allowlist repository
func NewIPAllowlistRepository(db *PostgresDB) *IPAllowlistRepository {
	return &IPAllowlistRepository{db: db}
}

// CreateAllowlist creates a new IP allowlist
func (r *IPAllowlistRepository) CreateAllowlist(ctx context.Context, allowlist *IPAllowlist) error {
	query := `
		INSERT INTO ip_allowlists (id, tenant_id, name, description, default_policy, mfa_required_for_unknown_ip, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, tenant_id, name, description, default_policy, mfa_required_for_unknown_ip, created_at, updated_at`

	var description sql.NullString
	if allowlist.Description != nil {
		description.String = *allowlist.Description
		description.Valid = true
	}

	err := r.db.QueryRow(query,
		allowlist.ID,
		allowlist.TenantID,
		allowlist.Name,
		description,
		allowlist.DefaultPolicy,
		allowlist.MFARequiredForUnknownIP,
	).Scan(
		&allowlist.ID,
		&allowlist.TenantID,
		&allowlist.Name,
		&description,
		&allowlist.DefaultPolicy,
		&allowlist.MFARequiredForUnknownIP,
		&allowlist.CreatedAt,
		&allowlist.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create IP allowlist: %w", err)
	}

	if description.Valid {
		allowlist.Description = &description.String
	}

	return nil
}

// GetAllowlistByID retrieves an IP allowlist by ID
func (r *IPAllowlistRepository) GetAllowlistByID(ctx context.Context, allowlistID uuid.UUID) (*IPAllowlist, error) {
	query := `
		SELECT id, tenant_id, name, description, default_policy, mfa_required_for_unknown_ip, created_at, updated_at
		FROM ip_allowlists WHERE id = $1`

	allowlist := &IPAllowlist{}
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, allowlistID).Scan(
		&allowlist.ID,
		&allowlist.TenantID,
		&allowlist.Name,
		&description,
		&allowlist.DefaultPolicy,
		&allowlist.MFARequiredForUnknownIP,
		&allowlist.CreatedAt,
		&allowlist.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get IP allowlist: %w", err)
	}

	if description.Valid {
		allowlist.Description = &description.String
	}

	return allowlist, nil
}

// GetAllowlistByTenantID retrieves the active IP allowlist for a tenant
func (r *IPAllowlistRepository) GetAllowlistByTenantID(ctx context.Context, tenantID uuid.UUID) (*IPAllowlist, error) {
	// For now, return the most recently created allowlist for the tenant
	// In the future, this could be extended to support multiple allowlists per tenant
	query := `
		SELECT id, tenant_id, name, description, default_policy, mfa_required_for_unknown_ip, created_at, updated_at
		FROM ip_allowlists WHERE tenant_id = $1
		ORDER BY created_at DESC LIMIT 1`

	allowlist := &IPAllowlist{}
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&allowlist.ID,
		&allowlist.TenantID,
		&allowlist.Name,
		&description,
		&allowlist.DefaultPolicy,
		&allowlist.MFARequiredForUnknownIP,
		&allowlist.CreatedAt,
		&allowlist.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get IP allowlist for tenant: %w", err)
	}

	if description.Valid {
		allowlist.Description = &description.String
	}

	return allowlist, nil
}

// ListAllowlistsByTenantID lists all IP allowlists for a tenant
func (r *IPAllowlistRepository) ListAllowlistsByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*IPAllowlist, error) {
	query := `
		SELECT id, tenant_id, name, description, default_policy, mfa_required_for_unknown_ip, created_at, updated_at
		FROM ip_allowlists WHERE tenant_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list IP allowlists: %w", err)
	}
	defer rows.Close()

	var allowlists []*IPAllowlist
	for rows.Next() {
		allowlist := &IPAllowlist{}
		var description sql.NullString

		err := rows.Scan(
			&allowlist.ID,
			&allowlist.TenantID,
			&allowlist.Name,
			&description,
			&allowlist.DefaultPolicy,
			&allowlist.MFARequiredForUnknownIP,
			&allowlist.CreatedAt,
			&allowlist.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan IP allowlist: %w", err)
		}

		if description.Valid {
			allowlist.Description = &description.String
		}

		allowlists = append(allowlists, allowlist)
	}

	return allowlists, nil
}

// UpdateAllowlist updates an IP allowlist
func (r *IPAllowlistRepository) UpdateAllowlist(ctx context.Context, allowlist *IPAllowlist) error {
	query := `
		UPDATE ip_allowlists 
		SET name = $1, description = $2, default_policy = $3, mfa_required_for_unknown_ip = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING id, tenant_id, name, description, default_policy, mfa_required_for_unknown_ip, created_at, updated_at`

	var description sql.NullString
	if allowlist.Description != nil {
		description.String = *allowlist.Description
		description.Valid = true
	}

	err := r.db.QueryRowContext(ctx, query,
		allowlist.Name,
		description,
		allowlist.DefaultPolicy,
		allowlist.MFARequiredForUnknownIP,
		allowlist.ID,
	).Scan(
		&allowlist.ID,
		&allowlist.TenantID,
		&allowlist.Name,
		&description,
		&allowlist.DefaultPolicy,
		&allowlist.MFARequiredForUnknownIP,
		&allowlist.CreatedAt,
		&allowlist.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update IP allowlist: %w", err)
	}

	if description.Valid {
		allowlist.Description = &description.String
	}

	return nil
}

// DeleteAllowlist deletes an IP allowlist and all its entries
func (r *IPAllowlistRepository) DeleteAllowlist(ctx context.Context, allowlistID uuid.UUID) error {
	// Entries are deleted automatically due to CASCADE constraint
	result, err := r.db.ExecContext(ctx, "DELETE FROM ip_allowlists WHERE id = $1", allowlistID)
	if err != nil {
		return fmt.Errorf("failed to delete IP allowlist: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("IP allowlist not found")
	}

	return nil
}

// CreateEntry adds an IP entry to an allowlist
func (r *IPAllowlistRepository) CreateEntry(ctx context.Context, entry *IPAllowlistEntry) error {
	query := `
		INSERT INTO ip_allowlist_entries (id, allowlist_id, type, value, description, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, allowlist_id, type, value, description, created_at`

	var description sql.NullString
	if entry.Description != nil {
		description.String = *entry.Description
		description.Valid = true
	}

	err := r.db.QueryRowContext(ctx, query,
		entry.ID,
		entry.AllowlistID,
		entry.Type,
		entry.Value,
		description,
	).Scan(
		&entry.ID,
		&entry.AllowlistID,
		&entry.Type,
		&entry.Value,
		&description,
		&entry.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create IP allowlist entry: %w", err)
	}

	if description.Valid {
		entry.Description = &description.String
	}

	return nil
}

// GetEntriesByAllowlistID retrieves all entries for an allowlist
func (r *IPAllowlistRepository) GetEntriesByAllowlistID(ctx context.Context, allowlistID uuid.UUID) ([]*IPAllowlistEntry, error) {
	query := `
		SELECT id, allowlist_id, type, value, description, created_at
		FROM ip_allowlist_entries WHERE allowlist_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, allowlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get IP allowlist entries: %w", err)
	}
	defer rows.Close()

	var entries []*IPAllowlistEntry
	for rows.Next() {
		entry := &IPAllowlistEntry{}
		var description sql.NullString

		err := rows.Scan(
			&entry.ID,
			&entry.AllowlistID,
			&entry.Type,
			&entry.Value,
			&description,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan IP allowlist entry: %w", err)
		}

		if description.Valid {
			entry.Description = &description.String
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// GetEntryByID retrieves a single entry by ID
func (r *IPAllowlistRepository) GetEntryByID(ctx context.Context, entryID uuid.UUID) (*IPAllowlistEntry, error) {
	query := `
		SELECT id, allowlist_id, type, value, description, created_at
		FROM ip_allowlist_entries WHERE id = $1`

	entry := &IPAllowlistEntry{}
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, entryID).Scan(
		&entry.ID,
		&entry.AllowlistID,
		&entry.Type,
		&entry.Value,
		&description,
		&entry.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get IP allowlist entry: %w", err)
	}

	if description.Valid {
		entry.Description = &description.String
	}

	return entry, nil
}

// UpdateEntry updates an IP allowlist entry
func (r *IPAllowlistRepository) UpdateEntry(ctx context.Context, entry *IPAllowlistEntry) error {
	query := `
		UPDATE ip_allowlist_entries 
		SET type = $1, value = $2, description = $3
		WHERE id = $4
		RETURNING id, allowlist_id, type, value, description, created_at`

	var description sql.NullString
	if entry.Description != nil {
		description.String = *entry.Description
		description.Valid = true
	}

	err := r.db.QueryRowContext(ctx, query,
		entry.Type,
		entry.Value,
		description,
		entry.ID,
	).Scan(
		&entry.ID,
		&entry.AllowlistID,
		&entry.Type,
		&entry.Value,
		&description,
		&entry.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update IP allowlist entry: %w", err)
	}

	if description.Valid {
		entry.Description = &description.String
	}

	return nil
}

// DeleteEntry deletes an IP allowlist entry
func (r *IPAllowlistRepository) DeleteEntry(ctx context.Context, entryID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM ip_allowlist_entries WHERE id = $1", entryID)
	if err != nil {
		return fmt.Errorf("failed to delete IP allowlist entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("IP allowlist entry not found")
	}

	return nil
}

// DeleteAllEntriesForAllowlist deletes all entries for an allowlist
func (r *IPAllowlistRepository) DeleteAllEntriesForAllowlist(ctx context.Context, allowlistID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM ip_allowlist_entries WHERE allowlist_id = $1", allowlistID)
	if err != nil {
		return fmt.Errorf("failed to delete IP allowlist entries: %w", err)
	}

	return nil
}
