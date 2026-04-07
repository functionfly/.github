package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/google/uuid"
)

// TenantRepository handles tenant-related database operations
type TenantRepository struct {
	db *PostgresDB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *PostgresDB) *TenantRepository {
	return &TenantRepository{db: db}
}

// GetTenantByID retrieves a tenant by ID
func (r *TenantRepository) GetTenantByID(tenantID uuid.UUID) (*Tenant, error) {
	tenant := &Tenant{}
	var plan sql.NullString
	var stripeCustomerID sql.NullString
	err := r.db.QueryRow(`
		SELECT id, name, plan, status, stripe_customer_id, created_at, updated_at
		FROM tenants WHERE id = $1`, tenantID).Scan(
		&tenant.ID, &tenant.Name, &plan, &tenant.Status,
		&stripeCustomerID, &tenant.CreatedAt, &tenant.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	if plan.Valid {
		tenant.Plan = plan.String
	}
	if stripeCustomerID.Valid && stripeCustomerID.String != "" {
		tenant.StripeCustomerID = &stripeCustomerID.String
	}
	if tenant.Status == "" {
		tenant.Status = "active" // Default status for backward compatibility
	}

	return tenant, nil
}

// CountRoutingEventsForTenantSince counts routing events for a tenant since a given time
func (r *TenantRepository) CountRoutingEventsForTenantSince(tenantID uuid.UUID, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM routing_events re
		JOIN apps a ON re.app_id = a.id
		WHERE a.tenant_id = $1 AND re.timestamp >= $2`,
		tenantID, since).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count routing events: %w", err)
	}

	return count, nil
}

// ListTenants lists all tenants
func (r *TenantRepository) ListTenants() ([]*Tenant, error) {
	query := `SELECT id, name, plan, status, stripe_customer_id, created_at, updated_at FROM tenants ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		tenant := &Tenant{}
		var plan sql.NullString
		var stripeCustomerID sql.NullString
		err := rows.Scan(&tenant.ID, &tenant.Name, &plan, &tenant.Status, &stripeCustomerID, &tenant.CreatedAt, &tenant.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		if plan.Valid {
			tenant.Plan = plan.String
		}
		if stripeCustomerID.Valid && stripeCustomerID.String != "" {
			tenant.StripeCustomerID = &stripeCustomerID.String
		}
		if tenant.Status == "" {
			tenant.Status = "active" // Default status
		}
		tenants = append(tenants, tenant)
	}

	return tenants, nil
}

// UpdateTenant updates tenant fields dynamically
func (r *TenantRepository) UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) (*Tenant, error) {
	// Get current tenant
	current, err := r.GetTenantByID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current tenant: %w", err)
	}
	if current == nil {
		return nil, fmt.Errorf("tenant not found")
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if name, ok := updates["name"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, name)
		argIndex++
	}

	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if plan, ok := updates["plan"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("plan = $%d", argIndex))
		args = append(args, plan)
		argIndex++
	}

	if stripeCustomerID, ok := updates["stripe_customer_id"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("stripe_customer_id = $%d", argIndex))
		args = append(args, stripeCustomerID)
		argIndex++
	}
	if _, ok := updates["stripe_customer_id"]; ok && updates["stripe_customer_id"] == nil {
		setParts = append(setParts, "stripe_customer_id = NULL")
	}

	if len(setParts) == 0 {
		return current, nil // No updates
	}

	setParts = append(setParts, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE tenants SET %s WHERE id = $%d RETURNING id, name, plan, status, stripe_customer_id, created_at, updated_at",
		strings.Join(setParts, ", "), argIndex)

	args = append(args, tenantID)

	updated := &Tenant{}
	var plan sql.NullString
	var stripeCustomerID sql.NullString
	err = r.db.QueryRow(query, args...).Scan(&updated.ID, &updated.Name, &plan, &updated.Status, &stripeCustomerID, &updated.CreatedAt, &updated.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	if plan.Valid {
		updated.Plan = plan.String
	}
	if stripeCustomerID.Valid && stripeCustomerID.String != "" {
		updated.StripeCustomerID = &stripeCustomerID.String
	}

	return updated, nil
}

// CreateTenant creates a new tenant with the free plan by default
func (r *TenantRepository) CreateTenant(ctx context.Context, name string) (*Tenant, error) {
	tenant := &Tenant{
		ID:     uuid.New(),
		Name:   name,
		Plan:   plans.PlanFree, // Default new signups to free tier
		Status: "active",
	}

	query := `
		INSERT INTO tenants (id, name, plan, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, name, plan, status, stripe_customer_id, created_at, updated_at`

	var plan sql.NullString
	var stripeCustomerID sql.NullString
	err := r.db.QueryRow(query, tenant.ID, tenant.Name, tenant.Plan, tenant.Status).Scan(
		&tenant.ID, &tenant.Name, &plan, &tenant.Status, &stripeCustomerID, &tenant.CreatedAt, &tenant.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	if plan.Valid {
		tenant.Plan = plan.String
	}
	if stripeCustomerID.Valid && stripeCustomerID.String != "" {
		tenant.StripeCustomerID = &stripeCustomerID.String
	}

	return tenant, nil
}

// DeleteTenant deletes a tenant
func (r *TenantRepository) DeleteTenant(ctx context.Context, tenantID uuid.UUID) error {
	// Check if tenant has any users
	var userCount int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE tenant_id = $1", tenantID).Scan(&userCount)
	if err != nil {
		return fmt.Errorf("failed to check tenant users: %w", err)
	}

	if userCount > 0 {
		return fmt.Errorf("cannot delete tenant with existing users")
	}

	// Delete the tenant
	result, err := r.db.Exec("DELETE FROM tenants WHERE id = $1", tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}
