package strategist

import (
	"context"
	"encoding/json"
	"time"

	"github.com/functionfly/functionfly/internal/agent/analyzer"
	"github.com/functionfly/functionfly/internal/agent/globalpatternlibrary"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service generates graph modification proposals from conversion analysis.
type Service struct {
	db              *gorm.DB
	patternLib      *globalpatternlibrary.Service
}

// NewService creates a new strategist service.
func NewService(db *gorm.DB, patternLib *globalpatternlibrary.Service) *Service {
	return &Service{db: db, patternLib: patternLib}
}

// GenerateProposals generates modification proposals from a conversion analysis.
// It consults the global pattern library to ground proposals in observed improvements.
func (s *Service) GenerateProposals(ctx context.Context, analysis *analyzer.ConversionAnalysis) ([]ModificationProposal, error) {
	var proposals []ModificationProposal

	for _, target := range analysis.ImprovementTargets {
		// Query global patterns for similar optimizations
		patterns, err := s.patternLib.Query(ctx, globalpatternlibrary.QueryParams{
			VerticalTags: []string{"e-commerce", "universal"},
			SharingTiers: []string{globalpatternlibrary.SharingTierUniversal, globalpatternlibrary.SharingTierVertical},
			Limit:        5,
		})
		if err != nil {
			patterns = nil
		}

		var patternRefs []string
		var avgImprovement float64
		for _, p := range patterns {
			patternRefs = append(patternRefs, p.ID.String())
			avgImprovement += p.ObservedImprovementPct
		}
		if len(patterns) > 0 {
			avgImprovement /= float64(len(patterns))
		}

		// Use global pattern improvement if available, else use analyzer estimate
		expectedLift := target.ExpectedLiftPct
		if avgImprovement > 0 {
			expectedLift = avgImprovement
		}

		patternRefsJSON, _ := json.Marshal(patternRefs)

		proposal := ModificationProposal{
			ID:                     uuid.New(),
			TenantID:               analysis.TenantID.String(),
			GraphID:                analysis.GraphID.String(),
			ChangeType:             target.ChangeType,
			TargetNodeID:           target.NodeID.String(),
			TargetNodeName:         target.NodeName,
			ExpectedRevenueLift:    target.ExpectedRevenueCents,
			ExpectedLiftPct:        expectedLift,
			RiskScore:              target.RiskScore,
			RollbackPlan:           buildRollbackPlan(&target),
			GlobalPatternRefs:      string(patternRefsJSON),
			Status:                 StatusPending,
			CreatedAt:              time.Now().UTC(),
		}

		if err := s.db.WithContext(ctx).Create(&proposal).Error; err != nil {
			continue
		}
		proposals = append(proposals, proposal)
	}

	return proposals, nil
}

// ModificationProposal represents a proposed graph change.
type ModificationProposal struct {
	ID                     uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID               string    `json:"tenant_id" gorm:"not null;index"`
	GraphID                string    `json:"graph_id" gorm:"not null;index"`
	ChangeType             string    `json:"change_type" gorm:"not null"` // add_node / remove_node / rewire_edge / modify_policy
	TargetNodeID           string    `json:"target_node_id" gorm:"not null"`
	TargetNodeName         string    `json:"target_node_name"`
	ExpectedRevenueLift    int64     `json:"expected_revenue_lift" gorm:"default:0"`
	ExpectedLiftPct        float64   `json:"expected_lift_pct"`
	RiskScore              float64   `json:"risk_score"` // 0.0–1.0
	RollbackPlan           string    `json:"rollback_plan"`
	GlobalPatternRefs      string    `json:"global_pattern_refs" gorm:"type:text"` // JSON array of pattern IDs
	Status                 string    `json:"status" gorm:"not null;default:'pending'"`
	ApprovedBy             *string   `json:"approved_by"`
	ImplementedAt          *time.Time `json:"implemented_at"`
	CreatedAt              time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt              time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name.
func (ModificationProposal) TableName() string { return "sebg_modification_proposals" }

// Status constants.
const (
	StatusPending   = "pending"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusApplied   = "applied"
	StatusExpired   = "expired"
)

// Approve approves a proposal.
func (s *Service) Approve(ctx context.Context, proposalID uuid.UUID, approverID string) error {
	return s.db.WithContext(ctx).Model(&ModificationProposal{}).
		Where("id = ?", proposalID).
		Updates(map[string]any{
			"status":      StatusApproved,
			"approved_by": approverID,
			"updated_at":  time.Now().UTC(),
		}).Error
}

// Reject rejects a proposal.
func (s *Service) Reject(ctx context.Context, proposalID uuid.UUID, rejectedBy string) error {
	return s.db.WithContext(ctx).Model(&ModificationProposal{}).
		Where("id = ?", proposalID).
		Updates(map[string]any{
			"status":      StatusRejected,
			"approved_by": rejectedBy,
			"updated_at":  time.Now().UTC(),
		}).Error
}

// ListPending returns all pending proposals for a tenant.
func (s *Service) ListPending(ctx context.Context, tenantID uuid.UUID) ([]ModificationProposal, error) {
	var proposals []ModificationProposal
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ?", tenantID.String(), StatusPending).
		Order("expected_revenue_lift DESC").
		Find(&proposals).Error
	return proposals, err
}

func buildRollbackPlan(target *analyzer.ImprovementTarget) string {
	switch target.ChangeType {
	case "add_retry":
		return "Remove the added retry node and restore original edge routing"
	case "add_specialist":
		return "Remove the specialist node and restore original graph topology"
	case "optimize":
		return "Revert the optimized node to its previous implementation"
	case "rewire_edge":
		return "Restore the original edge mapping from graph snapshot"
	}
	return "Restore graph from last known good snapshot"
}
