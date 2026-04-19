package sebg

import (
	"time"

	"github.com/google/uuid"
)

// AutonomyTier represents the level of SEBG autonomy for a tenant.
type AutonomyTier string

const (
	AutonomyTierManual      AutonomyTier = "manual"       // observe and recommend; human approves all
	AutonomyTierAssisted    AutonomyTier = "assisted"     // auto-apply low-risk; human approves high-risk
	AutonomyTierAutonomous  AutonomyTier = "fully_autonomous" // SEBG operates without human intervention
)

// TenantConfig holds SEBG configuration for a tenant.
type TenantConfig struct {
	ID                    uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID              uuid.UUID    `json:"tenant_id" gorm:"type:uuid;not null;uniqueIndex"`
	AutonomyTier          AutonomyTier `json:"autonomy_tier" gorm:"type:text;not null;default:'manual'"`
	AutonomyTierUpdatedAt  time.Time    `json:"autonomy_tier_updated_at" gorm:"autoUpdateTime"`
	RevenueShareFeePct    float64      `json:"revenue_share_fee_pct" gorm:"type:decimal(5,2);default:10.00"`
	MaxRiskScoreAutoApply float64      `json:"max_risk_score_auto_apply" gorm:"type:decimal(3,2);default:0.20"`
	VerticalTag           string       `json:"vertical_tag" gorm:"type:text;default:'e-commerce'"`
	IsActive              bool         `json:"is_active" gorm:"not null;default:true"`
	CreatedAt             time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name.
func (TenantConfig) TableName() string { return "sebg_tenant_configs" }

// RevenueImprovement tracks the measured revenue improvement for a tenant.
type RevenueImprovement struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID        uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	VerticalTag     string    `json:"vertical_tag" gorm:"type:text"`
	Period          string    `json:"period" gorm:"not null;index"` // "2026-04"
	BaselineRevenue int64     `json:"baseline_revenue_cents" gorm:"default:0"`
	CurrentRevenue  int64     `json:"current_revenue_cents" gorm:"default:0"`
	ImprovementPct  float64   `json:"improvement_pct" gorm:"type:decimal(6,2);default:0"`
	FeePct          float64   `json:"fee_pct" gorm:"type:decimal(5,2);default:10.00"`
	FeeCents        int64      `json:"fee_cents" gorm:"default:0"`
	VerifiedAt      *time.Time `json:"verified_at"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the GORM table name.
func (RevenueImprovement) TableName() string { return "sebg_revenue_improvements" }

// CanAutoApply returns true if a proposal with the given risk score can be auto-applied.
func (c *TenantConfig) CanAutoApply(riskScore float64) bool {
	if c.AutonomyTier == AutonomyTierAutonomous {
		return true
	}
	if c.AutonomyTier == AutonomyTierAssisted {
		return riskScore <= c.MaxRiskScoreAutoApply
	}
	return false
}
