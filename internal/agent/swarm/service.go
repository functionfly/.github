package swarm

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/economy"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service handles agent swarm orchestration
type Service struct {
	db             *gorm.DB
	identityRepo   *identity.Repository
	messageService *MessageService
	walletService  *economy.Service
}

// NewService creates a new swarm service
func NewService(db *gorm.DB, identityRepo *identity.Repository, walletService *economy.Service) *Service {
	return &Service{
		db:            db,
		identityRepo:  identityRepo,
		walletService: walletService,
	}
}

// SpawnChildRequest represents a request to spawn a child agent
type SpawnChildRequest struct {
	ParentAgentID    string
	ChildAgentID     string
	ChildName        string
	ChildDescription string
	SwarmRole        string // worker | manager | infrastructure
	MaxChildAgents   int
	Capabilities     map[string]any
	InitialBudgetUSD float64
	PolicyConfig     *PolicyConfig
}

// PolicyConfig defines the policy configuration for a child agent
type PolicyConfig struct {
	MaxExecutionDepth   int
	MaxRecursionDepth   int
	MaxWallTimeMs       int
	MaxMemoryGrowthMB   int
	ForbiddenFunctions  []string
	DeterministicOnly   bool
	AllowedCapabilities []string
}

// SpawnChild creates a new child agent under the parent
func (s *Service) SpawnChild(ctx context.Context, req *SpawnChildRequest) (*identity.AgentIdentity, string, error) {
	// Validate parent exists and has capacity
	parent, err := s.identityRepo.GetAgent(ctx, req.ParentAgentID)
	if err != nil {
		return nil, "", fmt.Errorf("parent agent not found: %w", err)
	}

	// Check parent's child capacity
	childCount, err := s.CountChildren(ctx, req.ParentAgentID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to count children: %w", err)
	}

	if parent.MaxChildAgents > 0 && childCount >= int64(parent.MaxChildAgents) {
		return nil, "", fmt.Errorf("parent agent has reached max child capacity (%d)", parent.MaxChildAgents)
	}

	// Validate role
	if req.SwarmRole == "" {
		req.SwarmRole = identity.SwarmRoleWorker
	}
	if req.SwarmRole != identity.SwarmRoleWorker && req.SwarmRole != identity.SwarmRoleManager &&
		req.SwarmRole != identity.SwarmRoleInfrastructure {
		return nil, "", fmt.Errorf("invalid swarm role: %s", req.SwarmRole)
	}

	// Create the child agent
	childReq := &identity.RegisterAgentRequest{
		AgentID:     req.ChildAgentID,
		Name:        req.ChildName,
		Description: req.ChildDescription,
	}

	child, apiKey, err := s.identityRepo.CreateAgent(ctx, parent.TenantID, childReq)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create child agent: %w", err)
	}

	// Update swarm fields
	updates := map[string]interface{}{
		"parent_agent_id":    req.ParentAgentID,
		"swarm_role":         req.SwarmRole,
		"max_child_agents":   req.MaxChildAgents,
		"capabilities":       req.Capabilities,
		"autonomous_enabled": false, // Disabled by default for spawned agents
		"evolution_enabled":  false,
	}

	if err := s.db.WithContext(ctx).Model(&identity.AgentIdentity{}).
		Where("agent_id = ?", req.ChildAgentID).
		Updates(updates).Error; err != nil {
		return nil, "", fmt.Errorf("failed to update child agent swarm fields: %w", err)
	}

	// Create relationship record
	relationship := &identity.AgentRelationship{
		ID:                 uuid.New(),
		ParentAgentID:      req.ParentAgentID,
		ChildAgentID:       req.ChildAgentID,
		RelationshipType:   "parent",
		MaxDelegationDepth: 5,
		CreatedAt:          time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(relationship).Error; err != nil {
		return nil, "", fmt.Errorf("failed to create relationship: %w", err)
	}

	// Initialize wallet with initial budget if provided
	if req.InitialBudgetUSD > 0 && s.walletService != nil {
		if _, err := s.walletService.GetOrCreateWallet(ctx, req.ChildAgentID); err != nil {
			return nil, "", fmt.Errorf("failed to create wallet: %w", err)
		}
		if _, err := s.walletService.Credit(ctx, req.ChildAgentID, req.InitialBudgetUSD, "initial_budget", map[string]any{"parent_agent_id": req.ParentAgentID}); err != nil {
			return nil, "", fmt.Errorf("failed to credit initial budget: %w", err)
		}
	} else if req.InitialBudgetUSD > 0 {
		wallet := &identity.AgentWallet{
			ID:               uuid.New(),
			AgentID:          req.ChildAgentID,
			BalanceUSD:       req.InitialBudgetUSD,
			EscrowBalanceUSD: 0,
			TotalEarnedUSD:   0,
			TotalSpentUSD:    0,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := s.db.WithContext(ctx).Create(wallet).Error; err != nil {
			return nil, "", fmt.Errorf("failed to create wallet: %w", err)
		}
	}

	return child, apiKey, nil
}

// CountChildren returns the number of child agents for a parent
func (s *Service) CountChildren(ctx context.Context, parentAgentID string) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&identity.AgentRelationship{}).
		Where("parent_agent_id = ?", parentAgentID).
		Count(&count).Error
	return count, err
}

// GetChildren returns all child agents for a parent
func (s *Service) GetChildren(ctx context.Context, parentAgentID string) ([]*identity.AgentIdentity, error) {
	var relationships []identity.AgentRelationship
	err := s.db.WithContext(ctx).
		Where("parent_agent_id = ?", parentAgentID).
		Find(&relationships).Error
	if err != nil {
		return nil, err
	}

	childIDs := make([]string, len(relationships))
	for i, rel := range relationships {
		childIDs[i] = rel.ChildAgentID
	}

	var children []*identity.AgentIdentity
	if len(childIDs) > 0 {
		err = s.db.WithContext(ctx).
			Where("agent_id IN ?", childIDs).
			Find(&children).Error
	}
	return children, err
}

// GetParent returns the parent agent for a given child
func (s *Service) GetParent(ctx context.Context, childAgentID string) (*identity.AgentIdentity, error) {
	var rel identity.AgentRelationship
	err := s.db.WithContext(ctx).
		Where("child_agent_id = ?", childAgentID).
		First(&rel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No parent
		}
		return nil, err
	}

	return s.identityRepo.GetAgent(ctx, rel.ParentAgentID)
}

// TerminateChild terminates a child agent relationship
func (s *Service) TerminateChild(ctx context.Context, parentAgentID, childAgentID string) error {
	// Verify relationship exists
	var rel identity.AgentRelationship
	err := s.db.WithContext(ctx).
		Where("parent_agent_id = ? AND child_agent_id = ?", parentAgentID, childAgentID).
		First(&rel).Error
	if err != nil {
		return fmt.Errorf("relationship not found: %w", err)
	}

	// Delete relationship
	if err := s.db.WithContext(ctx).Delete(&rel).Error; err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	// Update child to remove parent reference
	if err := s.db.WithContext(ctx).Model(&identity.AgentIdentity{}).
		Where("agent_id = ?", childAgentID).
		Update("parent_agent_id", nil).Error; err != nil {
		return fmt.Errorf("failed to update child agent: %w", err)
	}

	return nil
}

// GetDelegationDepth calculates the current delegation depth for an agent
func (s *Service) GetDelegationDepth(ctx context.Context, agentID string) (int, error) {
	depth := 0
	currentAgentID := agentID

	for depth < 20 { // Safety cap
		var rel identity.AgentRelationship
		err := s.db.WithContext(ctx).
			Where("child_agent_id = ?", currentAgentID).
			First(&rel).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				break
			}
			return depth, err
		}
		currentAgentID = rel.ParentAgentID
		depth++
	}

	return depth, nil
}
