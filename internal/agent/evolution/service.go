package evolution

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service handles agent evolution and learning
type Service struct {
	db *gorm.DB
}

// NewService creates a new evolution service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// AnalyzePerformance analyzes agent performance metrics
func (s *Service) AnalyzePerformance(ctx context.Context, agentID string, timeWindow time.Duration) (*PerformanceAnalysis, error) {
	since := time.Now().Add(-timeWindow)

	// Get execution records
	var records []identity.AgentExecutionRecord
	err := s.db.WithContext(ctx).
		Where("agent_id = ? AND timestamp >= ?", agentID, since).
		Order("timestamp DESC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return &PerformanceAnalysis{
			AgentID:          agentID,
			AnalysisWindow:   timeWindow,
			TotalExecutions:  0,
			SuccessRate:      0,
			AvgLatencyMs:     0,
			AvgCostUSD:       0,
			FailureCategories: map[string]int{},
		}, nil
	}

	// Calculate metrics
	var successCount, totalCost, totalLatency int
	failureCategories := map[string]int{}

	for _, r := range records {
		switch r.Outcome {
		case "success":
			successCount++
		case "error":
			if r.ErrorCode != nil {
				failureCategories[*r.ErrorCode]++
			} else {
				failureCategories["unknown"]++
			}
		case "timeout":
			failureCategories["timeout"]++
		case "policy_violation":
			failureCategories["policy_violation"]++
		}
		totalCost += int(r.CostUSD * 1000000) // Convert to micro-units
		totalLatency += r.LatencyMs
	}

	successRate := float64(successCount) / float64(len(records)) * 100
	avgLatency := float64(totalLatency) / float64(len(records))
	avgCost := float64(totalCost) / float64(len(records)) / 1000000

	return &PerformanceAnalysis{
		AgentID:           agentID,
		AnalysisWindow:    timeWindow,
		TotalExecutions:   len(records),
		SuccessRate:       successRate,
		AvgLatencyMs:      avgLatency,
		AvgCostUSD:        avgCost,
		FailureCategories: failureCategories,
	}, nil
}

// PerformanceAnalysis represents the analysis of agent performance
type PerformanceAnalysis struct {
	AgentID           string
	AnalysisWindow    time.Duration
	TotalExecutions   int
	SuccessRate       float64
	AvgLatencyMs      float64
	AvgCostUSD        float64
	FailureCategories map[string]int
}

// ProposeEvolution creates an evolution proposal based on performance analysis
func (s *Service) ProposeEvolution(ctx context.Context, agentID string, analysis *PerformanceAnalysis) (*identity.EvolutionProposal, error) {
	var proposalType string
	var proposalData map[string]any

	// Decision logic based on analysis
	if analysis.SuccessRate < 80 {
		// Low success rate - propose spawning a specialist helper
		proposalType = identity.EvolutionTypeSpawnSpecialist
		proposalData = map[string]any{
			"reason":            "low_success_rate",
			"current_success":   analysis.SuccessRate,
			"target_success":    80.0,
			"specialist_role":   "error_handler",
			"capabilities":      []string{"error_recovery", "retry_logic"},
		}
	} else if analysis.AvgLatencyMs > 10000 {
		// High latency - propose optimization
		proposalType = identity.EvolutionTypeModifyPolicy
		proposalData = map[string]any{
			"reason":         "high_latency",
			"current_latency": analysis.AvgLatencyMs,
			"target_latency": 5000.0,
			"policy_changes": map[string]any{
				"timeout_ms": analysis.AvgLatencyMs * 2,
			},
		}
	} else if analysis.TotalExecutions > 100 && analysis.SuccessRate > 95 {
		// High volume and success - safe to generate new function
		proposalType = identity.EvolutionTypeGenerateFunction
		proposalData = map[string]any{
			"reason":           "high_volume_success",
			"executions":       analysis.TotalExecutions,
			"success_rate":     analysis.SuccessRate,
			"capability_focus": "common_task_automation",
		}
	} else {
		// No clear evolution path
		return nil, fmt.Errorf("no clear evolution path identified")
	}

	// Get parent agent to check if parent approval is required
	var parentID *string
	var rel identity.AgentRelationship
	if err := s.db.WithContext(ctx).Where("child_agent_id = ?", agentID).First(&rel).Error; err == nil {
		parentID = &rel.ParentAgentID
	}

	proposal := &identity.EvolutionProposal{
		ID:                      uuid.New(),
		AgentID:                 agentID,
		ProposalType:            proposalType,
		ProposalData:            proposalData,
		Status:                  "pending",
		ParentApprovalRequired:  parentID != nil,
		SimulatedOutcome:        nil,
		ApprovedBy:              nil,
		ImplementedAt:           nil,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(proposal).Error; err != nil {
		return nil, err
	}

	return proposal, nil
}

// SubmitProposal submits an evolution proposal
func (s *Service) SubmitProposal(ctx context.Context, proposal *identity.EvolutionProposal) error {
	proposal.ID = uuid.New()
	proposal.Status = "pending"
	proposal.CreatedAt = time.Now()
	proposal.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Create(proposal).Error
}

// ApproveProposal approves an evolution proposal
func (s *Service) ApproveProposal(ctx context.Context, proposalID uuid.UUID, approverAgentID string) (*identity.EvolutionProposal, error) {
	var proposal identity.EvolutionProposal
	if err := s.db.WithContext(ctx).Where("id = ?", proposalID).First(&proposal).Error; err != nil {
		return nil, err
	}

	if proposal.Status != "pending" {
		return nil, fmt.Errorf("proposal is not pending")
	}

	// Verify approver is the parent or has authority
	if proposal.ParentApprovalRequired {
		var rel identity.AgentRelationship
		if err := s.db.WithContext(ctx).Where("child_agent_id = ?", proposal.AgentID).First(&rel).Error; err != nil {
			return nil, fmt.Errorf("cannot verify approval authority")
		}
		if rel.ParentAgentID != approverAgentID {
			return nil, fmt.Errorf("only parent agent can approve this proposal")
		}
	}

	proposal.Status = "approved"
	proposal.ApprovedBy = &approverAgentID
	proposal.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Save(&proposal).Error; err != nil {
		return nil, err
	}

	return &proposal, nil
}

// ImplementProposal implements an approved evolution proposal
func (s *Service) ImplementProposal(ctx context.Context, proposalID uuid.UUID) error {
	var proposal identity.EvolutionProposal
	if err := s.db.WithContext(ctx).Where("id = ?", proposalID).First(&proposal).Error; err != nil {
		return err
	}

	if proposal.Status != "approved" {
		return fmt.Errorf("proposal must be approved before implementation")
	}

	now := time.Now()
	proposal.Status = "implemented"
	proposal.ImplementedAt = &now
	proposal.UpdatedAt = now

	if err := s.db.WithContext(ctx).Save(&proposal).Error; err != nil {
		return err
	}

	// Execute the actual evolution based on type
	switch proposal.ProposalType {
	case identity.EvolutionTypeSpawnSpecialist:
		// This would trigger the swarm service to spawn a new agent
		return s.implementSpawnSpecialist(ctx, proposal)
	case identity.EvolutionTypeModifyPolicy:
		return s.implementPolicyModification(ctx, proposal)
	case identity.EvolutionTypeGenerateFunction:
		return s.implementFunctionGeneration(ctx, proposal)
	// Add other types as needed
	}

	return nil
}

func (s *Service) implementSpawnSpecialist(ctx context.Context, proposal identity.EvolutionProposal) error {
	// Extract specialist config from proposal data
	data, _ := json.Marshal(proposal.ProposalData)
	var config map[string]any
	json.Unmarshal(data, &config)

	// Log the implementation
	fmt.Printf("Implementing spawn specialist for agent %s: %v\n", proposal.AgentID, config)
	return nil
}

func (s *Service) implementPolicyModification(ctx context.Context, proposal identity.EvolutionProposal) error {
	data, _ := json.Marshal(proposal.ProposalData)
	var config map[string]any
	json.Unmarshal(data, &config)

	fmt.Printf("Implementing policy modification for agent %s: %v\n", proposal.AgentID, config)
	return nil
}

func (s *Service) implementFunctionGeneration(ctx context.Context, proposal identity.EvolutionProposal) error {
	data, _ := json.Marshal(proposal.ProposalData)
	var config map[string]any
	json.Unmarshal(data, &config)

	fmt.Printf("Implementing function generation for agent %s: %v\n", proposal.AgentID, config)
	return nil
}

// GetProposals gets evolution proposals for an agent
func (s *Service) GetProposals(ctx context.Context, agentID string, status string) ([]identity.EvolutionProposal, error) {
	query := s.db.WithContext(ctx).Where("agent_id = ?", agentID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var proposals []identity.EvolutionProposal
	err := query.Order("created_at DESC").Find(&proposals).Error
	return proposals, err
}

// GetProposal gets a specific proposal
func (s *Service) GetProposal(ctx context.Context, proposalID uuid.UUID) (*identity.EvolutionProposal, error) {
	var proposal identity.EvolutionProposal
	err := s.db.WithContext(ctx).Where("id = ?", proposalID).First(&proposal).Error
	return &proposal, err
}

// RetireChild proposes retiring a low-performing child agent
func (s *Service) ProposeRetireChild(ctx context.Context, parentAgentID, childAgentID, reason string) (*identity.EvolutionProposal, error) {
	proposal := &identity.EvolutionProposal{
		ID:                      uuid.New(),
		AgentID:                 parentAgentID,
		ProposalType:            identity.EvolutionTypeRetireChild,
		ProposalData: map[string]any{
			"child_agent_id": childAgentID,
			"reason":         reason,
		},
		Status:                  "pending",
		ParentApprovalRequired:  false, // Self-approval allowed
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(proposal).Error; err != nil {
		return nil, err
	}

	return proposal, nil
}
