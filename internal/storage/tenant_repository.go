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
func (r *TenantRepository) GetTenantByID(ctx context.Context, tenantID uuid.UUID) (*Tenant, error) {
	tenant := &Tenant{}
	var plan sql.NullString
	var stripeCustomerID sql.NullString
	err := r.db.QueryRowContext(ctx, `
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

// GetTenantPlan retrieves just the plan for a tenant
func (r *TenantRepository) GetTenantPlan(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var plan sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT plan FROM tenants WHERE id = $1`, tenantID).Scan(&plan)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get tenant plan: %w", err)
	}
	if plan.Valid {
		return plan.String, nil
	}
	return "", nil
}

// GetTenantByStripeCustomerID retrieves a tenant by Stripe customer ID
func (r *TenantRepository) GetTenantByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*Tenant, error) {
	if stripeCustomerID == "" {
		return nil, nil
	}

	tenant := &Tenant{}
	var plan sql.NullString
	var stripeCID sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, plan, status, stripe_customer_id, created_at, updated_at
		FROM tenants WHERE stripe_customer_id = $1`, stripeCustomerID).Scan(
		&tenant.ID, &tenant.Name, &plan, &tenant.Status,
		&stripeCID, &tenant.CreatedAt, &tenant.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by stripe customer id: %w", err)
	}

	if plan.Valid {
		tenant.Plan = plan.String
	}
	if stripeCID.Valid && stripeCID.String != "" {
		tenant.StripeCustomerID = &stripeCID.String
	}
	if tenant.Status == "" {
		tenant.Status = "active"
	}

	return tenant, nil
}

// ListTenantsWithStripeCustomerID retrieves all tenants that have a Stripe customer ID
// This is used for syncing payment methods from Stripe
func (r *TenantRepository) ListTenantsWithStripeCustomerID(ctx context.Context) ([]*Tenant, error) {
	query := `
		SELECT id, name, plan, status, stripe_customer_id, created_at, updated_at
		FROM tenants
		WHERE stripe_customer_id IS NOT NULL AND stripe_customer_id != ''
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants with stripe customer id: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		tenant := &Tenant{}
		var plan sql.NullString
		var stripeCID sql.NullString

		err := rows.Scan(
			&tenant.ID, &tenant.Name, &plan, &tenant.Status,
			&stripeCID, &tenant.CreatedAt, &tenant.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}

		if plan.Valid {
			tenant.Plan = plan.String
		}
		if stripeCID.Valid && stripeCID.String != "" {
			tenant.StripeCustomerID = &stripeCID.String
		}
		if tenant.Status == "" {
			tenant.Status = "active"
		}

		tenants = append(tenants, tenant)
	}

	return tenants, nil
}

// CountRoutingEventsForTenantSince counts routing events for a tenant since a given time
func (r *TenantRepository) CountRoutingEventsForTenantSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
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
func (r *TenantRepository) ListTenants(ctx context.Context) ([]*Tenant, error) {
	query := `SELECT id, name, plan, status, stripe_customer_id, created_at, updated_at FROM tenants ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
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
	current, err := r.GetTenantByID(ctx, tenantID)
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
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&updated.ID, &updated.Name, &plan, &updated.Status, &stripeCustomerID, &updated.CreatedAt, &updated.UpdatedAt)

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
	err := r.db.QueryRowContext(ctx, query, tenant.ID, tenant.Name, tenant.Plan, tenant.Status).Scan(
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
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE tenant_id = $1", tenantID).Scan(&userCount)
	if err != nil {
		return fmt.Errorf("failed to check tenant users: %w", err)
	}

	if userCount > 0 {
		return fmt.Errorf("cannot delete tenant with existing users")
	}

	// Delete the tenant
	result, err := r.db.ExecContext(ctx, "DELETE FROM tenants WHERE id = $1", tenantID)
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

// CountUsersByTenant returns the number of users in a tenant
func (r *TenantRepository) CountUsersByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE tenant_id = $1", tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tenant users: %w", err)
	}
	return count, nil
}

// IsUserInTenant checks if a user has access to a specific tenant (either as primary tenant or via membership)
func (r *TenantRepository) IsUserInTenant(ctx context.Context, userID, tenantID uuid.UUID) (bool, error) {
	// First check if this is the user's primary tenant
	var primaryTenantID uuid.UUID
	err := r.db.QueryRowContext(ctx, "SELECT tenant_id FROM users WHERE id = $1", userID).Scan(&primaryTenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to get user primary tenant: %w", err)
	}
	if primaryTenantID == tenantID {
		return true, nil
	}

	// Check if user has a tenant membership (and has accepted the invitation)
	var exists bool
	err = r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM tenant_memberships
			WHERE user_id = $1
			AND tenant_id = $2
			AND accepted_at IS NOT NULL
		)`, userID, tenantID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check tenant membership: %w", err)
	}
	return exists, nil
}

// GetUserTenants returns all tenant IDs that a user has access to (primary + memberships)
func (r *TenantRepository) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var primaryTenantID uuid.UUID
	err := r.db.QueryRowContext(ctx, "SELECT tenant_id FROM users WHERE id = $1", userID).Scan(&primaryTenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user primary tenant: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT tenant_id FROM tenant_memberships
		WHERE user_id = $1 AND accepted_at IS NOT NULL`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant memberships: %w", err)
	}
	defer rows.Close()

	// Use a map to deduplicate (shouldn't happen but defensive)
	tenantMap := make(map[uuid.UUID]bool)
	tenantMap[primaryTenantID] = true

	for rows.Next() {
		var tid uuid.UUID
		if err := rows.Scan(&tid); err != nil {
			return nil, fmt.Errorf("failed to scan tenant membership: %w", err)
		}
		tenantMap[tid] = true
	}

	tenants := make([]uuid.UUID, 0, len(tenantMap))
	for tid := range tenantMap {
		tenants = append(tenants, tid)
	}
	return tenants, nil
}

// AddTenantMember adds a user as a member to a tenant
func (r *TenantRepository) AddTenantMember(ctx context.Context, userID, tenantID, invitedBy uuid.UUID, role string) error {
	if role == "" {
		role = "member"
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tenant_memberships (user_id, tenant_id, role, invited_by, invited_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, tenant_id) DO UPDATE SET
			role = EXCLUDED.role,
			updated_at = NOW()
	`, userID, tenantID, role, invitedBy)
	if err != nil {
		return fmt.Errorf("failed to add tenant member: %w", err)
	}
	return nil
}

// AcceptTenantMembership marks a tenant membership as accepted
func (r *TenantRepository) AcceptTenantMembership(ctx context.Context, userID, tenantID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tenant_memberships
		SET accepted_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND tenant_id = $2 AND accepted_at IS NULL
	`, userID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to accept tenant membership: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no pending membership found")
	}
	return nil
}

// RemoveTenantMember removes a user's membership from a tenant
func (r *TenantRepository) RemoveTenantMember(ctx context.Context, userID, tenantID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM tenant_memberships
		WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to remove tenant member: %w", err)
	}
	return nil
}

// UpdateTenantStatus updates a tenant's status (e.g., active, suspended)
// Used by billing suspension workflows to restrict/restore service
func (r *TenantRepository) UpdateTenantStatus(ctx context.Context, tenantID uuid.UUID, status string) error {
	validStatuses := map[string]bool{"active": true, "suspended": true, "inactive": true}
	if !validStatuses[status] {
		return fmt.Errorf("invalid tenant status: %s", status)
	}

	_, err := r.db.ExecContext(ctx, `
		UPDATE tenants
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, status, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update tenant status: %w", err)
	}
	return nil
}

// SetTenantDegradedMode marks a tenant as running in degraded mode (shared-DB fallback)
// or clears the degraded flag when isolated provisioning recovers.
func (r *TenantRepository) SetTenantDegradedMode(ctx context.Context, tenantID uuid.UUID, degraded bool, reason string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tenants
		SET degraded_mode = $1,
		    degradation_reason = $2,
		    degradation_updated_at = NOW(),
		    updated_at = NOW()
		WHERE id = $3
	`, degraded, reason, tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant degraded mode: %w", err)
	}
	return nil
}

// UpdateTenantTaxSettings updates a tenant's tax-related fields including billing
// location and tax ID information.
func (r *TenantRepository) UpdateTenantTaxSettings(ctx context.Context, tenantID uuid.UUID, settings *TaxSettings) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tenants
		SET billing_country = $1,
		    billing_state = $2,
		    billing_postal_code = $3,
		    tax_id = $4,
		    tax_id_type = $5,
		    tax_exempt = $6,
		    updated_at = NOW()
		WHERE id = $7
	`, settings.BillingCountry,
		settings.BillingState,
		settings.BillingPostalCode,
		settings.TaxID,
		settings.TaxIDType,
		settings.TaxExempt,
		tenantID)
	if err != nil {
		return fmt.Errorf("failed to update tenant tax settings: %w", err)
	}
	return nil
}

// ============================================================================
// Dedicated Tenant Database Management
// ============================================================================

// TenantDBConfig represents tenant-specific database configuration
type TenantDBConfig struct {
	TenantID        uuid.UUID
	DBName          string
	Status          string
	ConnectionString string
	MaxConnections  int
	CreatedAt       time.Time
}

// ShouldHaveDedicatedDB checks if a tenant qualifies for a dedicated database
// based on their plan. Starter pack and above get dedicated databases.
func (r *TenantRepository) ShouldHaveDedicatedDB(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	tenant, err := r.GetTenantByID(ctx, tenantID)
	if err != nil {
		return false, err
	}
	if tenant == nil {
		return false, fmt.Errorf("tenant not found")
	}

	// Plans that qualify for dedicated databases
	dedicatedPlans := map[string]bool{
		plans.PlanStarter: true,
		plans.PlanPro:     true,
		plans.PlanEnterprise: true,
	}

	return dedicatedPlans[tenant.Plan], nil
}

// GetTenantDBConfig retrieves the dedicated database configuration for a tenant
func (r *TenantRepository) GetTenantDBConfig(ctx context.Context, tenantID uuid.UUID) (*TenantDBConfig, error) {
	cfg := &TenantDBConfig{}
	err := r.db.QueryRowContext(ctx, `
		SELECT tenant_id, db_name, status, connection_string_template, max_connections, created_at
		FROM tenant_database_configs
		WHERE tenant_id = $1
	`, tenantID).Scan(
		&cfg.TenantID, &cfg.DBName, &cfg.Status, &cfg.ConnectionString,
		&cfg.MaxConnections, &cfg.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // No dedicated DB for this tenant
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant DB config: %w", err)
	}

	return cfg, nil
}

// HasDedicatedDB checks if a tenant already has a dedicated database configured
func (r *TenantRepository) HasDedicatedDB(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM tenant_database_configs
			WHERE tenant_id = $1 AND status IN ('active', 'provisioning')
		)
	`, tenantID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check dedicated DB: %w", err)
	}
	return exists, nil
}

// AssignDedicatedDB assigns a dedicated database to a tenant (called after provisioning)
func (r *TenantRepository) AssignDedicatedDB(ctx context.Context, tenantID uuid.UUID, dbName, connectionTemplate string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tenant_database_configs (tenant_id, db_name, connection_string_template, status)
		VALUES ($1, $2, $3, 'active')
		ON CONFLICT (tenant_id) DO UPDATE SET
			db_name = EXCLUDED.db_name,
			connection_string_template = EXCLUDED.connection_string_template,
			status = 'active',
			updated_at = NOW()
	`, tenantID, dbName, connectionTemplate)
	if err != nil {
		return fmt.Errorf("failed to assign dedicated DB: %w", err)
	}
	return nil
}

// RemoveDedicatedDB removes the dedicated database from a tenant (e.g., when downgrading)
func (r *TenantRepository) RemoveDedicatedDB(ctx context.Context, tenantID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tenant_database_configs SET status = 'deleting', updated_at = NOW()
		WHERE tenant_id = $1
	`, tenantID)
	if err != nil {
		return fmt.Errorf("failed to remove dedicated DB: %w", err)
	}
	return nil
}

// ListTenantsWithDedicatedDBs returns all tenants that have dedicated databases
func (r *TenantRepository) ListTenantsWithDedicatedDBs(ctx context.Context) ([]*TenantDBConfig, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tenant_id, db_name, status, connection_string_template, max_connections, created_at
		FROM tenant_database_configs
		WHERE status IN ('active', 'provisioning', 'suspended')
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list dedicated DBs: %w", err)
	}
	defer rows.Close()

	var configs []*TenantDBConfig
	for rows.Next() {
		cfg := &TenantDBConfig{}
		err := rows.Scan(&cfg.TenantID, &cfg.DBName, &cfg.Status, &cfg.ConnectionString, &cfg.MaxConnections, &cfg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant DB config: %w", err)
		}
		configs = append(configs, cfg)
	}

	return configs, rows.Err()
}

// GetOrCreateTenantDBProvisioner returns a configured tenant DB provisioner
// This integrates the provisioner with the existing tenant repository
func (r *TenantRepository) GetOrCreateTenantDBProvisioner(provisioner *TenantDBProvisioner, poolManager *TenantPoolManager) (*TenantDBProvisioner, *TenantPoolManager, error) {
	// Return the injected provisioner and pool manager
	// In production, these would be injected at startup
	return provisioner, poolManager, nil
}
