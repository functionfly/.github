package statefabricaddons

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entitlement is a persisted tenant add-on row.
type Entitlement struct {
	ID                       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID                 uuid.UUID `gorm:"type:uuid;not null;index"`
	AddonID                  string    `gorm:"size:64;not null"`
	Status                   string    `gorm:"size:32;not null"`
	StripeSubscriptionID     *string   `gorm:"size:128"`
	StripeSubscriptionItemID *string   `gorm:"size:128"`
	CreatedAt                time.Time `gorm:"not null"`
	UpdatedAt                time.Time `gorm:"not null"`
}

func (Entitlement) TableName() string {
	return "state_fabric_addon_entitlements"
}

// Repository reads/writes state_fabric_addon_entitlements.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a repository backed by GORM.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ListActiveAddonIDs returns active add-on IDs for a tenant.
func (r *Repository) ListActiveAddonIDs(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Model(&Entitlement{}).
		Where("tenant_id = ? AND status = ?", tenantID, "active").
		Pluck("addon_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// ListEntitlementsByTenant returns all entitlement rows for tenant for admin/support tools.
func (r *Repository) ListEntitlementsByTenant(ctx context.Context, tenantID uuid.UUID) ([]Entitlement, error) {
	var out []Entitlement
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("addon_id ASC").
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetEntitlement returns the entitlement row for tenant+addon, or nil if none.
func (r *Repository) GetEntitlement(ctx context.Context, tenantID uuid.UUID, addonID string) (*Entitlement, error) {
	var e Entitlement
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND addon_id = ?", tenantID, addonID).
		First(&e).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

// HasActiveAddon reports whether a tenant has an active entitlement for the add-on.
func (r *Repository) HasActiveAddon(ctx context.Context, tenantID uuid.UUID, addonID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Entitlement{}).
		Where("tenant_id = ? AND addon_id = ? AND status = ?", tenantID, addonID, "active").
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpsertEntitlement creates or updates an entitlement row for tenant+addon.
func (r *Repository) UpsertEntitlement(
	ctx context.Context,
	tenantID uuid.UUID,
	addonID string,
	status string,
	stripeSubscriptionID *string,
	stripeSubscriptionItemID *string,
) error {
	updates := map[string]interface{}{
		"status":                      status,
		"stripe_subscription_id":      stripeSubscriptionID,
		"stripe_subscription_item_id": stripeSubscriptionItemID,
		"updated_at":                  time.Now().UTC(),
	}
	res := r.db.WithContext(ctx).Model(&Entitlement{}).
		Where("tenant_id = ? AND addon_id = ?", tenantID, addonID).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&Entitlement{
		TenantID:                 tenantID,
		AddonID:                  addonID,
		Status:                   status,
		StripeSubscriptionID:     stripeSubscriptionID,
		StripeSubscriptionItemID: stripeSubscriptionItemID,
	}).Error
}

// SetEntitlementStatusBySubscription updates row status by Stripe subscription ID.
func (r *Repository) SetEntitlementStatusBySubscription(ctx context.Context, stripeSubscriptionID string, status string) error {
	return r.db.WithContext(ctx).Model(&Entitlement{}).
		Where("stripe_subscription_id = ?", stripeSubscriptionID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now().UTC(),
		}).Error
}
