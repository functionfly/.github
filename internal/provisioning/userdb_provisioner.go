package provisioning

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// UserDBProvisioner extends the existing TenantDBProvisioner to create
// the dedicated tenant database and apply the SaaS Starter schema migration.
//
// This is the FIRST step in the provisioning pipeline — all other provisioners
// depend on the tenant database existing.
//
// What it provisions:
//   - Dedicated PostgreSQL database for the tenant (via TenantDBProvisioner)
//   - Base tenant schema (tenant_users, tenant_configs, tenant_api_keys, etc.)
//   - SaaS Starter extension schema (auth keys, sessions, payments, email, analytics)
//   - Platform reference record linking tenant DB to platform
//   - Default admin user for the tenant
type UserDBProvisioner struct {
	platformRepo  storage.Repository
	dbProvisioner *storage.TenantDBProvisioner
}

// NewUserDBProvisioner creates a new UserDB provisioner
func NewUserDBProvisioner(platformRepo storage.Repository, dbProvisioner *storage.TenantDBProvisioner) *UserDBProvisioner {
	return &UserDBProvisioner{
		platformRepo:  platformRepo,
		dbProvisioner: dbProvisioner,
	}
}

// Provision creates the dedicated tenant database and applies all migrations.
func (up *UserDBProvisioner) Provision(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (*ComponentState, error) {
	startTime := time.Now()
	log := logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"component": "user_db",
	})

	state := &ComponentState{
		Status:    StatusProvisioning,
		Timestamp: startTime,
	}

	// 1. Create dedicated tenant database (idempotent — returns nil if already exists)
	if err := up.dbProvisioner.CreateTenantDB(ctx, tenantID); err != nil {
		return state, fmt.Errorf("failed to create tenant database: %w", err)
	}
	log.Info("Dedicated tenant database created/verified")

	// 2. Get connection pool to the tenant's database
	pool, err := up.dbProvisioner.GetTenantPool(ctx, tenantID)
	if err != nil {
		return state, fmt.Errorf("failed to get tenant DB pool: %w", err)
	}

	// 3. Create platform reference record in tenant DB (for cross-DB lookups)
	// Get the platform tenant to find the creator
	tenant, err := up.platformRepo.GetTenantByID(tenantID)
	if err != nil || tenant == nil {
		log.WithError(err).Warn("Could not get platform tenant (non-fatal)")
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_platform_refs (id, tenant_id, platform_tenant_id, platform_user_id, plan, created_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())
		 ON CONFLICT (tenant_id) DO UPDATE SET plan = EXCLUDED.plan`,
		uuid.New(), tenantID, tenantID, tenantID, bundleSlug)
	if err != nil {
		log.WithError(err).Warn("Failed to create platform ref (non-fatal)")
	}

	// 4. Initialize tenant config in the dedicated DB
	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_configs (id, tenant_id, settings, feature_flags, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NOW(), NOW())
		 ON CONFLICT (tenant_id) DO NOTHING`,
		uuid.New(), tenantID,
		`{"app_name": "My SaaS", "default_currency": "usd", "timezone": "UTC", "date_format": "YYYY-MM-DD"}`,
		`{"analytics_enabled": true, "email_workflows_enabled": true, "self_signup_enabled": true}`)
	if err != nil {
		log.WithError(err).Warn("Failed to create tenant config (non-fatal)")
	}

	// 5. Create the first admin user from the platform user
	// Get the platform user who created this tenant
	platformUsers, err := up.platformRepo.ListActiveUsersByTenant(ctx, tenantID)
	if err == nil && len(platformUsers) > 0 {
		adminUser := platformUsers[0]
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_users (id, tenant_id, email, password_hash, display_name, avatar_url, metadata, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, '{"role":"admin","source":"platform_sync"}', true, NOW(), NOW())
			 ON CONFLICT (tenant_id, email) DO NOTHING`,
			adminUser.ID, tenantID, adminUser.Email, adminUser.PasswordHash, adminUser.Name, "")
		if err != nil {
			log.WithError(err).Warn("Failed to sync admin user (non-fatal)")
		} else {
			log.WithField("email", adminUser.Email).Info("Admin user synced to tenant DB")
		}
	}

	state.Status = StatusActive
	state.ResourceID = fmt.Sprintf("db:functionfly_tenant_%s", tenantID.String()[:8])
	log.Info("User DB provisioning complete")
	return state, nil
}
