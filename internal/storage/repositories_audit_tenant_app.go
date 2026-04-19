package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PostgresDB methods: audit, tenants, apps.

// Audit operations
func (db *PostgresDB) ListAuditEvents(limit, offset int) ([]*AuditEvent, error) {
	return db.auditRepository.ListAuditEvents(limit, offset)
}

func (db *PostgresDB) LogAuditEvent(ctx context.Context, event *AuditEvent) error {
	return db.auditRepository.LogAuditEvent(ctx, event)
}

func (db *PostgresDB) ListAuditEventsFiltered(limit, offset int, filters map[string]interface{}) ([]*AuditEvent, error) {
	return db.auditRepository.ListAuditEventsFiltered(limit, offset, filters)
}

func (db *PostgresDB) GetAuditEventByID(id uuid.UUID) (*AuditEvent, error) {
	return db.auditRepository.GetAuditEventByID(id)
}

// Tenant operations
func (db *PostgresDB) CountRoutingEventsForTenantSince(tenantID uuid.UUID, since time.Time) (int, error) {
	return db.tenantRepository.CountRoutingEventsForTenantSince(tenantID, since)
}

func (db *PostgresDB) CreateTenant(ctx context.Context, name string) (*Tenant, error) {
	return db.tenantRepository.CreateTenant(ctx, name)
}

func (db *PostgresDB) GetTenantByID(tenantID uuid.UUID) (*Tenant, error) {
	return db.tenantRepository.GetTenantByID(tenantID)
}

func (db *PostgresDB) GetTenantByStripeCustomerID(stripeCustomerID string) (*Tenant, error) {
	return db.tenantRepository.GetTenantByStripeCustomerID(stripeCustomerID)
}

func (db *PostgresDB) ListTenants() ([]*Tenant, error) {
	return db.tenantRepository.ListTenants()
}

func (db *PostgresDB) ListTenantsWithStripeCustomerID() ([]*Tenant, error) {
	return db.tenantRepository.ListTenantsWithStripeCustomerID()
}

func (db *PostgresDB) UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) (*Tenant, error) {
	return db.tenantRepository.UpdateTenant(ctx, tenantID, updates)
}

func (db *PostgresDB) UpdateTenantStatus(ctx context.Context, tenantID uuid.UUID, status string) error {
	return db.tenantRepository.UpdateTenantStatus(ctx, tenantID, status)
}

func (db *PostgresDB) DeleteTenant(ctx context.Context, tenantID uuid.UUID) error {
	return db.tenantRepository.DeleteTenant(ctx, tenantID)
}

func (db *PostgresDB) CountUsersByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	return db.tenantRepository.CountUsersByTenant(ctx, tenantID)
}

// IsUserInTenant checks if a user has access to a specific tenant (primary or via membership)
func (db *PostgresDB) IsUserInTenant(ctx context.Context, userID, tenantID uuid.UUID) (bool, error) {
	return db.tenantRepository.IsUserInTenant(ctx, userID, tenantID)
}

// GetUserTenants returns all tenant IDs that a user has access to (primary + memberships)
func (db *PostgresDB) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return db.tenantRepository.GetUserTenants(ctx, userID)
}

// AddTenantMember adds a user as a member to a tenant
func (db *PostgresDB) AddTenantMember(ctx context.Context, userID, tenantID, invitedBy uuid.UUID, role string) error {
	return db.tenantRepository.AddTenantMember(ctx, userID, tenantID, invitedBy, role)
}

// AcceptTenantMembership marks a tenant membership as accepted
func (db *PostgresDB) AcceptTenantMembership(ctx context.Context, userID, tenantID uuid.UUID) error {
	return db.tenantRepository.AcceptTenantMembership(ctx, userID, tenantID)
}

// RemoveTenantMember removes a user's membership from a tenant
func (db *PostgresDB) RemoveTenantMember(ctx context.Context, userID, tenantID uuid.UUID) error {
	return db.tenantRepository.RemoveTenantMember(ctx, userID, tenantID)
}

// App operations
func (db *PostgresDB) CreateApp(name, slug string, tenantID uuid.UUID) (*App, error) {
	return db.appRepository.CreateApp(name, slug, tenantID)
}

func (db *PostgresDB) GetAppByID(id uuid.UUID) (*App, error) {
	return db.appRepository.GetAppByID(id)
}

func (db *PostgresDB) GetAppBySlug(slug string) (*App, error) {
	return db.appRepository.GetAppBySlug(slug)
}

func (db *PostgresDB) GetAppBySlugAndTenant(slug string, tenantID uuid.UUID) (*App, error) {
	return db.appRepository.GetAppBySlugAndTenant(slug, tenantID)
}

func (db *PostgresDB) ListAppsByTenant(tenantID uuid.UUID) ([]*App, error) {
	return db.appRepository.ListAppsByTenant(tenantID)
}
