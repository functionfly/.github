package swarm

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/economy"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
// SECURITY FIX: Uses transaction with row locking to prevent TOCTOU race on capacity check
func (s *Service) SpawnChild(ctx context.Context, req *SpawnChildRequest) (*identity.AgentIdentity, string, error) {
	// Start transaction for atomic operation
	tx := s.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, "", fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// SECURITY FIX: Lock parent row to prevent race on capacity check
	var parent identity.AgentIdentity
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"})).
		Where("agent_id = ? AND status = ?", req.ParentAgentID, "active").
		First(&parent).Error; err != nil {
		tx.Rollback()
		return nil, "", fmt.Errorf("parent agent not found: %w", err)
	}

	// Check parent's child capacity within transaction
	var childCount int64
	if err := tx.Model(&identity.AgentRelationship{}).
		Where("parent_agent_id = ?", req.ParentAgentID).
		Count(&childCount).Error; err != nil {
		tx.Rollback()
		return nil, "", fmt.Errorf("failed to count children: %w", err)
	}

	if parent.MaxChildAgents > 0 && childCount >= int64(parent.MaxChildAgents) {
		tx.Rollback()
		return nil, "", fmt.Errorf("parent agent has reached max child capacity (%d)", parent.MaxChildAgents)
	}

	// Validate role
	if req.SwarmRole == "" {
		req.SwarmRole = identity.SwarmRoleWorker
	}
	if req.SwarmRole != identity.SwarmRoleWorker && req.SwarmRole != identity.SwarmRoleManager &&
		req.SwarmRole != identity.SwarmRoleInfrastructure {
		tx.Rollback()
		return nil, "", fmt.Errorf("invalid swarm role: %s", req.SwarmRole)
	}

	// Create the child agent (uses its own internal transaction, safe to call)
	childReq := &identity.RegisterAgentRequest{
		AgentID:     req.ChildAgentID,
		Name:        req.ChildName,
		Description: req.ChildDescription,
	}

	child, apiKey, signingKey, err := s.identityRepo.CreateAgent(ctx, parent.TenantID, childReq)
	if err != nil {
		tx.Rollback()
		return nil, "", fmt.Errorf("failed to create child agent: %w", err)
	}

	// Store signing key in daemon config for later use by the parent
	if signingKey != "" {
		daemonConfig := map[string]any{
			"signing_key": signingKey,
		}
		tx.WithContext(ctx).Model(&identity.AgentIdentity{}).
			Where("agent_id = ?", child.AgentID).
			Update("daemon_config", daemonConfig)
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

	if err := tx.WithContext(ctx).Model(&identity.AgentIdentity{}).
		Where("agent_id = ?", req.ChildAgentID).
		Updates(updates).Error; err != nil {
		tx.Rollback()
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

	if err := tx.WithContext(ctx).Create(relationship).Error; err != nil {
		tx.Rollback()
		return nil, "", fmt.Errorf("failed to create relationship: %w", err)
	}

	// Initialize wallet with initial budget if provided
	if req.InitialBudgetUSD > 0 && s.walletService != nil {
		if _, err := s.walletService.GetOrCreateWallet(ctx, req.ChildAgentID); err != nil {
			tx.Rollback()
			return nil, "", fmt.Errorf("failed to create wallet: %w", err)
		}
		if _, err := s.walletService.Credit(ctx, req.ChildAgentID, req.InitialBudgetUSD, "initial_budget", map[string]any{"parent_agent_id": req.ParentAgentID}); err != nil {
			tx.Rollback()
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
		if err := tx.WithContext(ctx).Create(wallet).Error; err != nil {
			tx.Rollback()
			return nil, "", fmt.Errorf("failed to create wallet: %w", err)
		}
	}

	// SECURITY FIX: Commit the transaction to make all changes atomic
	if err := tx.Commit().Error; err != nil {
		return nil, "", fmt.Errorf("failed to commit spawn transaction: %w", err)
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

// ReassignRole updates the swarm role of an agent
func (s *Service) ReassignRole(ctx context.Context, agentID string, newRole string) error {
	validRoles := map[string]bool{
		identity.SwarmRoleWorker:         true,
		identity.SwarmRoleManager:        true,
		identity.SwarmRoleInfrastructure: true,
	}
	if !validRoles[newRole] {
		return fmt.Errorf("invalid swarm role: %s (must be worker, manager, or infrastructure)", newRole)
	}

	agent, err := s.identityRepo.GetAgent(ctx, agentID)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	if agent.SwarmRole == newRole {
		return nil // No change needed
	}

	return s.identityRepo.UpdateSwarmRole(ctx, agentID, newRole)
}

// ReshapeSwarm updates the swarm topology of an agent
func (s *Service) ReshapeSwarm(ctx context.Context, agentID string, newTopology string) error {
	validTopologies := map[string]bool{
		identity.SwarmTopologyChain: true,
		identity.SwarmTopologyStar:  true,
		identity.SwarmTopologyMesh:  true,
		identity.SwarmTopologyTree:  true,
	}
	if !validTopologies[newTopology] {
		return fmt.Errorf("invalid swarm topology: %s (must be chain, star, mesh, or tree)", newTopology)
	}

	agent, err := s.identityRepo.GetAgent(ctx, agentID)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	if agent.SwarmTopology == newTopology {
		return nil // No change needed
	}

	return s.identityRepo.UpdateSwarmTopology(ctx, agentID, newTopology)
}
