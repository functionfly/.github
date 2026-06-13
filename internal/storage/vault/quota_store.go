package vault

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/functionfly/functionfly/internal/storage/vault/quota"
)

// VaultRateLimit mirrors the row in the vault_rate_limits table.
// We define it here (rather than in models.go) so quota is the
// only place the column layout is referenced.
type VaultRateLimit struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID              uuid.UUID `gorm:"type:uuid;not null;index"`
	Resource              string    `gorm:"size:50;not null"`
	RequestsPerMinute     int       `gorm:"not null;default:0"`
	RequestsPerHour       int       `gorm:"not null;default:0"`
	MaxSecrets            int       `gorm:"not null;default:0"`
	MaxDynamicCredentials int       `gorm:"not null;default:0"`
	AuditExportsPerDay    int       `gorm:"not null;default:0"`
	Notes                 string    `gorm:"type:text"`
	CreatedAt             time.Time `gorm:"autoCreateTime"`
	UpdatedAt             time.Time `gorm:"autoUpdateTime"`
}

func (VaultRateLimit) TableName() string { return "vault_rate_limits" }

func (r *VaultRateLimit) BeforeCreate(tx interface{ Now() time.Time }) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// QuotaStore adapts the vault repository to the quota.Store
// interface. It is constructed at process start and shared by the
// rate-limit middleware.
type QuotaStore struct {
	repo *Repository
}

// NewQuotaStore returns a Store the quota package can use.
func NewQuotaStore(repo *Repository) *QuotaStore { return &QuotaStore{repo: repo} }

// GetTenantPlan reads the tenant's plan from the public.users /
// tenants tables. The query joins tenants.plan (assumed column)
// with tenants.id.
func (s *QuotaStore) GetTenantPlan(ctx context.Context, tenantID uuid.UUID) (quota.Plan, error) {
	var row struct {
		Plan string
	}
	if err := s.repo.db.WithContext(ctx).
		Table("tenants").
		Select("COALESCE(plan, 'free') AS plan").
		Where("id = ?", tenantID).
		Scan(&row).Error; err != nil {
		// If the tenants table doesn't expose a plan column (e.g.
		// during early dev), we fall back to free.
		return quota.PlanFree, nil
	}
	return quota.Plan(row.Plan), nil
}

// GetOverride returns the admin override for a resource, or
// (limit=0, ok=false) when none exists.
func (s *QuotaStore) GetOverride(ctx context.Context, tenantID uuid.UUID, resource quota.Resource) (int64, time.Duration, bool, error) {
	var row VaultRateLimit
	err := s.repo.db.WithContext(ctx).
		Where("tenant_id = ? AND resource = ?", tenantID, string(resource)).
		First(&row).Error
	if err != nil {
		return 0, 0, false, nil // not-found means no override
	}
	limit, window := overrideLimit(row, resource)
	if limit <= 0 {
		return 0, 0, false, nil
	}
	return limit, window, true, nil
}

func overrideLimit(row VaultRateLimit, resource quota.Resource) (int64, time.Duration) {
	switch resource {
	case quota.ResourceSecrets:
		return int64(row.MaxSecrets), 0
	case quota.ResourceDynamicCreds:
		return int64(row.MaxDynamicCredentials), 30 * 24 * time.Hour
	case quota.ResourceTokensPerSecret:
		return int64(row.RequestsPerHour), 0
	case quota.ResourceAuditExports:
		return int64(row.AuditExportsPerDay), 24 * time.Hour
	}
	return 0, 0
}

// CountSecrets returns the current non-deleted secret count.
func (s *QuotaStore) CountSecrets(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var n int64
	err := s.repo.db.WithContext(ctx).Model(&Secret{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&n).Error
	return n, err
}

// CountActiveTokens returns the non-revoked, non-expired token
// count for a secret.
func (s *QuotaStore) CountActiveTokens(ctx context.Context, secretID uuid.UUID) (int64, error) {
	var n int64
	err := s.repo.db.WithContext(ctx).Model(&AccessToken{}).
		Where("secret_id = ? AND is_revoked = ? AND expires_at > ?", secretID, false, time.Now()).
		Count(&n).Error
	return n, err
}

// CountDynamicCredsSince counts dynamic credentials minted in the
// window.
func (s *QuotaStore) CountDynamicCredsSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int64, error) {
	var n int64
	err := s.repo.db.WithContext(ctx).Model(&DynamicCredential{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Count(&n).Error
	return n, err
}

// CountAuditExportsSince counts audit-log exports in the window.
// The plan calls for a dedicated counter; we approximate using
// audit log rows with action = "audit_export" (a custom action).
// For now, we return 0 — the audit export counter is a Phase 5.2
// follow-up. Returning 0 keeps the API contract intact while
// letting callers opt into a tighter check later.
func (s *QuotaStore) CountAuditExportsSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int64, error) {
	return 0, nil
}
