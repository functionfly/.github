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

func (db *PostgresDB) ListTenants() ([]*Tenant, error) {
	return db.tenantRepository.ListTenants()
}

func (db *PostgresDB) UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) (*Tenant, error) {
	return db.tenantRepository.UpdateTenant(ctx, tenantID, updates)
}

func (db *PostgresDB) DeleteTenant(ctx context.Context, tenantID uuid.UUID) error {
	return db.tenantRepository.DeleteTenant(ctx, tenantID)
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
