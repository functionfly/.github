package sebg

import (
	"time"

	"github.com/google/uuid"
)

// Autonomy tier levels.
const (
	TierManual         = "manual"
	TierAssisted       = "assisted"
	TierFullyAutonomous = "fully_autonomous"
)

// TenantConfig stores per-tenant SEBG settings (autonomy tier, billing, etc.).
type TenantConfig struct {
	ID                    uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID              uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;uniqueIndex"`
	AutonomyTier          string    `json:"autonomy_tier" gorm:"not null;default:'manual'"` // manual / assisted / fully_autonomous
	RevenueShareFeePct    float64   `json:"revenue_share_fee_pct" gorm:"default:0"`         // e.g. 0.10 = 10%
	MaxRiskScoreAutoApply float64   `json:"max_risk_score_auto_apply" gorm:"default:0.2"`     // proposals above this risk require human approval
	IsActive             bool      `json:"is_active" gorm:"default:true"`
	CreatedAt            time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name.
func (TenantConfig) TableName() string { return "sebg_tenant_configs" }

// ROIReport summarises applied proposal impact for a tenant.
type ROIReport struct {
	TenantID        uuid.UUID `json:"tenant_id"`
	AppliedCount    int       `json:"applied_count"`
	PendingCount    int       `json:"pending_count"`
	RevenueLiftCents int64    `json:"revenue_lift_cents"`
}
