package governor

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/globalpatternlibrary"
	"github.com/functionfly/functionfly/internal/agent/strategist"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service reviews specialist proposals against global patterns and safety guardrails.
type Service struct {
	db           *gorm.DB
	patternLib   *globalpatternlibrary.Service
	strategistSvc *strategist.Service
}

// NewService creates a new governor service.
func NewService(db *gorm.DB, patternLib *globalpatternlibrary.Service, strategistSvc *strategist.Service) *Service {
	return &Service{db: db, patternLib: patternLib, strategistSvc: strategistSvc}
}

// ReviewDecision represents the outcome of a governor review.
type ReviewDecision struct {
	ProposalID   uuid.UUID `json:"proposal_id"`
	Decision    string    `json:"decision"` // approved / rejected / escalated
	Reason      string    `json:"reason"`
	AutoApproved bool    `json:"auto_approved"`
	RiskScore   float64   `json:"risk_score"`
}

// ReviewProposal reviews a single proposal.
// Low-risk proposals (risk_score < 0.2) are auto-approved.
// High-risk proposals are escalated to human review.
func (s *Service) ReviewProposal(ctx context.Context, proposalID uuid.UUID) (*ReviewDecision, error) {
	var proposal strategist.ModificationProposal
	if err := s.db.WithContext(ctx).Where("id = ?", proposalID).First(&proposal).Error; err != nil {
		return nil, fmt.Errorf("proposal not found: %w", err)
	}

	// Independently calculate risk score - do not trust proposal.RiskScore
	calculatedRiskScore := s.calculateRiskScore(&proposal)
	decision := &ReviewDecision{
		ProposalID: proposalID,
		RiskScore:  calculatedRiskScore,
	}

	// Auto-approve low-risk changes
	if calculatedRiskScore < 0.2 {
		decision.Decision = "approved"
		decision.Reason = fmt.Sprintf("auto-approved: risk score %.2f below threshold", calculatedRiskScore)
		decision.AutoApproved = true
		if err := s.strategistSvc.Approve(ctx, proposalID, "governor_auto"); err != nil {
			return nil, err
		}
		return decision, nil
	}

	// Check global pattern library for supporting evidence
	patterns, err := s.patternLib.Query(ctx, globalpatternlibrary.QueryParams{
		TenantID:    uuid.UUID{},
		VerticalTags: []string{"e-commerce", "universal"},
		SharingTiers: []string{globalpatternlibrary.SharingTierUniversal, globalpatternlibrary.SharingTierVertical},
		Limit:        5,
	})
	if err == nil && len(patterns) > 0 {
		var totalConf float64
		for _, p := range patterns {
			totalConf += p.ConfidenceScore
		}
		avgConf := totalConf / float64(len(patterns))
		if avgConf > 0.7 && calculatedRiskScore < 0.4 {
			decision.Decision = "approved"
			decision.Reason = fmt.Sprintf("approved with pattern support: %.0f%% avg confidence from %d patterns", avgConf*100, len(patterns))
			decision.AutoApproved = true
			if err := s.strategistSvc.Approve(ctx, proposalID, "governor_pattern"); err != nil {
				return nil, err
			}
			return decision, nil
		}
	}

	// Safety check: never auto-approve node removal on critical path
	if proposal.ChangeType == "remove_node" && calculatedRiskScore >= 0.2 {
		decision.Decision = "rejected"
		decision.Reason = "safety guardrail: cannot auto-approve critical path node removal"
		decision.AutoApproved = false
		if err := s.strategistSvc.Reject(ctx, proposalID, "governor_safety"); err != nil {
			return nil, err
		}
		return decision, nil
	}

	// Default: escalate to human review
	decision.Decision = "escalated"
	decision.Reason = fmt.Sprintf("manual review required: risk score %.2f above auto-approve threshold", calculatedRiskScore)
	decision.AutoApproved = false
	return decision, nil
}

// ReviewBatch reviews all pending proposals for a tenant.
func (s *Service) ReviewBatch(ctx context.Context, tenantID uuid.UUID) ([]ReviewDecision, error) {
	pending, err := s.strategistSvc.ListPending(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var decisions []ReviewDecision
	for _, p := range pending {
		d, err := s.ReviewProposal(ctx, p.ID)
		if err != nil {
			continue
		}
		decisions = append(decisions, *d)
	}
	return decisions, nil
}

// calculateRiskScore independently calculates risk score from proposal attributes.
// This prevents bypass via risk score manipulation in untrusted proposal data.
func (s *Service) calculateRiskScore(proposal *strategist.ModificationProposal) float64 {
	score := 0.0

	switch proposal.ChangeType {
	case "remove_node":
		score += 0.5
	case "modify_policy":
		score += 0.4
	case "rewire_edge":
		score += 0.3
	case "add_node":
		score += 0.1
	}

	// Higher expected revenue lift indicates more impact and potentially higher risk
	if proposal.ExpectedRevenueLift > 10000 {
		score += 0.2
	} else if proposal.ExpectedRevenueLift > 1000 {
		score += 0.1
	}

	// Higher expected lift percentage indicates more aggressive change
	if proposal.ExpectedLiftPct > 0.2 {
		score += 0.15
	} else if proposal.ExpectedLiftPct > 0.1 {
		score += 0.1
	}

	// Lack of rollback plan increases risk
	if proposal.RollbackPlan == "" {
		score += 0.1
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// GovernorLog records all review decisions for audit.
type GovernorLog struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProposalID  uuid.UUID  `json:"proposal_id" gorm:"type:uuid;not null;index"`
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Decision   string     `json:"decision" gorm:"not null"`
	Reason     string     `json:"reason" gorm:"type:text"`
	RiskScore  float64    `json:"risk_score"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the GORM table name.
func (GovernorLog) TableName() string { return "sebg_governor_log" }

// AutoMigrate runs database migrations for governor components.
func (s *Service) AutoMigrate(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&GovernorLog{})
}