package evolution

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/actuator"
	"github.com/functionfly/functionfly/internal/agent/attribution"
	graphpkg "github.com/functionfly/functionfly/internal/agent/graph"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service handles agent evolution and learning
type Service struct {
	db          *gorm.DB
	graphSvc    *graphpkg.Service
	actuatorSvc *actuator.Service
}

// NewService creates a new evolution service.
func NewService(db *gorm.DB, graphSvc *graphpkg.Service, actuatorSvc *actuator.Service) *Service {
	return &Service{db: db, graphSvc: graphSvc, actuatorSvc: actuatorSvc}
}

// AnalyzePerformance analyzes agent performance metrics
func (s *Service) AnalyzePerformance(ctx context.Context, agentID string, timeWindow time.Duration) (*PerformanceAnalysis, error) {
	since := time.Now().Add(-timeWindow)

	// Get execution records
	var records []attribution.AgentExecutionRecord
	err := s.db.WithContext(ctx).
		Where("agent_id = ? AND timestamp >= ?", agentID, since).
		Order("timestamp DESC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return &PerformanceAnalysis{
			AgentID:           agentID,
			AnalysisWindow:    timeWindow,
			TotalExecutions:   0,
			SuccessRate:       0,
			AvgLatencyMs:      0,
			AvgCostUSD:        0,
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
			"reason":          "low_success_rate",
			"current_success": analysis.SuccessRate,
			"target_success":  80.0,
			"specialist_role": "error_handler",
			"capabilities":    []string{"error_recovery", "retry_logic"},
		}
	} else if analysis.AvgLatencyMs > 10000 {
		// High latency - propose optimization
		proposalType = identity.EvolutionTypeModifyPolicy
		proposalData = map[string]any{
			"reason":          "high_latency",
			"current_latency": analysis.AvgLatencyMs,
			"target_latency":  5000.0,
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
		ID:                     uuid.New(),
		AgentID:                agentID,
		ProposalType:           proposalType,
		ProposalData:           proposalData,
		Status:                 "pending",
		ParentApprovalRequired: parentID != nil,
		SimulatedOutcome:       nil,
		ApprovedBy:             nil,
		ImplementedAt:          nil,
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

// RejectProposal rejects an evolution proposal.
func (s *Service) RejectProposal(ctx context.Context, proposalID uuid.UUID, rejectedBy string) (*identity.EvolutionProposal, error) {
	var proposal identity.EvolutionProposal
	if err := s.db.WithContext(ctx).Where("id = ?", proposalID).First(&proposal).Error; err != nil {
		return nil, err
	}

	if proposal.Status != "pending" {
		return nil, fmt.Errorf("proposal is not pending")
	}

	proposal.Status = "rejected"
	proposal.ApprovedBy = &rejectedBy
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
		return s.implementSpawnSpecialist(ctx, proposal)
	case identity.EvolutionTypeModifyPolicy:
		return s.implementPolicyModification(ctx, proposal)
	case identity.EvolutionTypeGenerateFunction:
		return s.implementFunctionGeneration(ctx, proposal)
	case identity.EvolutionTypeRetireChild:
		return s.implementRetireChild(ctx, proposal)
		// Add other types as needed
	}

	return nil
}

func (s *Service) implementSpawnSpecialist(ctx context.Context, proposal identity.EvolutionProposal) error {
	data, _ := json.Marshal(proposal.ProposalData)
	var config map[string]any
	json.Unmarshal(data, &config)

	role, _ := config["specialist_role"].(string)
	if role == "" {
		role = "error_handler"
	}

	// Extract graph ID from proposal data or agent ID context
	graphIDStr, _ := config["graph_id"].(string)
	var graphID uuid.UUID
	if graphIDStr != "" {
		graphID, _ = uuid.Parse(graphIDStr)
	} else {
		return fmt.Errorf("spawn_specialist requires graph_id in proposal data")
	}

	// Create a new specialist node in the graph
	nodeName := fmt.Sprintf("%s_specialist", role)
	newNode, err := s.graphSvc.AddNode(ctx, graphID, "specialist/"+role, nodeName, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to add specialist node: %w", err)
	}

	// Add edge from the triggering node to the specialist
	triggerNodeIDStr, _ := config["trigger_node_id"].(string)
	if triggerNodeIDStr != "" {
		triggerNodeID, err := uuid.Parse(triggerNodeIDStr)
		if err == nil {
			s.graphSvc.AddEdge(ctx, graphID, triggerNodeID, newNode.ID, "trigger", nil)
		}
	}

	// Record the implementation
	proposal.ProposalData["implemented_node_id"] = newNode.ID.String()
	proposal.ProposalData["implemented_at"] = time.Now().UTC().Format(time.RFC3339)
	s.db.Save(&proposal)

	return nil
}

func (s *Service) implementPolicyModification(ctx context.Context, proposal identity.EvolutionProposal) error {
	data, _ := json.Marshal(proposal.ProposalData)
	var config map[string]any
	json.Unmarshal(data, &config)

	graphIDStr, _ := config["graph_id"].(string)
	if graphIDStr == "" {
		return fmt.Errorf("modify_policy requires graph_id in proposal data")
	}

	graphID, err := uuid.Parse(graphIDStr)
	if err != nil {
		return fmt.Errorf("invalid graph_id: %w", err)
	}

	policyChanges, ok := config["policy_changes"].(map[string]any)
	if !ok || len(policyChanges) == 0 {
		return fmt.Errorf("modify_policy requires policy_changes in proposal data")
	}

	// For policy modifications, we update edge metadata with new routing rules scoped to this graph
	nodeIDStr, _ := config["node_id"].(string)
	if nodeIDStr == "" {
		return fmt.Errorf("modify_policy requires node_id in proposal data")
	}

	nodeID, err := uuid.Parse(nodeIDStr)
	if err != nil {
		return fmt.Errorf("invalid node_id: %w", err)
	}

	var edges []graphpkg.Edge
	if err := s.db.WithContext(ctx).
		Where("source_node_id = ? AND graph_id = ?", nodeID, graphID).
		Find(&edges).Error; err != nil {
		return err
	}

	for _, edge := range edges {
		// Inject policy changes into edge metadata
		var metadata map[string]any
		if edge.Metadata != "" {
			json.Unmarshal([]byte(edge.Metadata), &metadata)
		} else {
			metadata = make(map[string]any)
		}
		for k, v := range policyChanges {
			metadata["policy_"+k] = v
		}
		metadata["policy_modified_at"] = time.Now().UTC().Format(time.RFC3339)
		metadata["policy_proposal_id"] = proposal.ID.String()
		metadataJSON, _ := json.Marshal(metadata)
		edge.Metadata = string(metadataJSON)
		s.db.Save(&edge)
	}

	return nil
}

func (s *Service) implementFunctionGeneration(ctx context.Context, proposal identity.EvolutionProposal) error {
	data, _ := json.Marshal(proposal.ProposalData)
	var config map[string]any
	json.Unmarshal(data, &config)

	graphIDStr, _ := config["graph_id"].(string)
	functionName, _ := config["generated_function_name"].(string)
	if functionName == "" {
		functionName = "generated_handler"
	}

	var graphID uuid.UUID
	if graphIDStr != "" {
		graphID, _ = uuid.Parse(graphIDStr)
	}

	// Create a new function node in the graph
	newNode, err := s.graphSvc.AddNode(ctx, graphID, "generated/"+functionName, functionName, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create generated function node: %w", err)
	}

	// Connect it as a child to the parent agent's graph
	insertAfterNodeIDStr, _ := config["insert_after_node_id"].(string)
	if insertAfterNodeIDStr != "" {
		insertAfterNodeID, err := uuid.Parse(insertAfterNodeIDStr)
		if err == nil {
			s.graphSvc.AddEdge(ctx, graphID, insertAfterNodeID, newNode.ID, "dataflow", map[string]string{
				"output": "input",
			})
		}
	}

	proposal.ProposalData["generated_node_id"] = newNode.ID.String()
	proposal.ProposalData["generated_at"] = time.Now().UTC().Format(time.RFC3339)
	s.db.Save(&proposal)

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
func (s *Service) ProposeRetireChild(ctx context.Context, parentAgentID, childAgentID, graphID, reason string) (*identity.EvolutionProposal, error) {
	proposal := &identity.EvolutionProposal{
		ID:           uuid.New(),
		AgentID:      parentAgentID,
		ProposalType: identity.EvolutionTypeRetireChild,
		ProposalData: map[string]any{
			"child_agent_id": childAgentID,
			"graph_id":       graphID,
			"reason":         reason,
		},
		Status:                 "pending",
		ParentApprovalRequired: false, // Self-approval allowed
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(proposal).Error; err != nil {
		return nil, err
	}

	return proposal, nil
}

func (s *Service) implementRetireChild(ctx context.Context, proposal identity.EvolutionProposal) error {
	data, _ := json.Marshal(proposal.ProposalData)
	var config map[string]any
	json.Unmarshal(data, &config)

	childAgentID, _ := config["child_agent_id"].(string)
	graphIDStr, _ := config["graph_id"].(string)

	if childAgentID == "" {
		return fmt.Errorf("retire_child requires child_agent_id in proposal data")
	}

	// If a graph ID is provided, deactivate the child's nodes in the graph
	if graphIDStr != "" {
		graphID, err := uuid.Parse(graphIDStr)
		if err == nil {
			// Find and deactivate all nodes belonging to this agent in this graph
			result := s.db.WithContext(ctx).Model(&graphpkg.Node{}).
				Where("function_id LIKE ?", childAgentID+"%").
				Where("graph_id = ?", graphID).
				Where("is_active = ?", true).
				Update("is_active", false)
			if result.Error != nil {
				return fmt.Errorf("failed to deactivate child nodes: %w", result.Error)
			}
		}
	}

	// Record the retirement
	proposal.ProposalData["retired_at"] = time.Now().UTC().Format(time.RFC3339)
	s.db.Save(&proposal)

	return nil
}
