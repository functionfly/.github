package evolution

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/agent/attribution"
	"github.com/functionfly/functionfly/internal/agent/identity"
	agentpolicy "github.com/functionfly/functionfly/internal/agent/policy"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var evolutionTestDBCounter int64

func TestProposeEvolutionCreatesSpawnSpecialistForLowSuccess(t *testing.T) {
	t.Parallel()

	db := newEvolutionTestDB(t)
	ctx := context.Background()
	svc := NewService(db, nil, nil)

	tenantID := uuid.New()
	agentID := "agent-low-success"
	now := time.Now()

	records := []attribution.AgentExecutionRecord{
		{
			ID:          uuid.New(),
			AgentID:     agentID,
			TenantID:    tenantID,
			FunctionID:  uuid.New(),
			ExecutionID: "exec-1",
			Outcome:     attribution.OutcomeError,
			LatencyMs:   1200,
			CostUSD:     0.001,
			Timestamp:   now,
		},
		{
			ID:          uuid.New(),
			AgentID:     agentID,
			TenantID:    tenantID,
			FunctionID:  uuid.New(),
			ExecutionID: "exec-2",
			Outcome:     attribution.OutcomeError,
			LatencyMs:   1300,
			CostUSD:     0.001,
			Timestamp:   now.Add(-time.Minute),
		},
		{
			ID:          uuid.New(),
			AgentID:     agentID,
			TenantID:    tenantID,
			FunctionID:  uuid.New(),
			ExecutionID: "exec-3",
			Outcome:     attribution.OutcomeError,
			LatencyMs:   900,
			CostUSD:     0.001,
			Timestamp:   now.Add(-2 * time.Minute),
		},
		{
			ID:          uuid.New(),
			AgentID:     agentID,
			TenantID:    tenantID,
			FunctionID:  uuid.New(),
			ExecutionID: "exec-4",
			Outcome:     attribution.OutcomeSuccess,
			LatencyMs:   850,
			CostUSD:     0.001,
			Timestamp:   now.Add(-3 * time.Minute),
		},
		{
			ID:          uuid.New(),
			AgentID:     agentID,
			TenantID:    tenantID,
			FunctionID:  uuid.New(),
			ExecutionID: "exec-5",
			Outcome:     attribution.OutcomeSuccess,
			LatencyMs:   700,
			CostUSD:     0.001,
			Timestamp:   now.Add(-4 * time.Minute),
		},
	}
	for i := range records {
		require.NoError(t, db.Create(&records[i]).Error)
	}

	analysis, err := svc.AnalyzePerformance(ctx, agentID, 24*time.Hour)
	require.NoError(t, err)
	require.InDelta(t, 40.0, analysis.SuccessRate, 0.01)

	proposal, err := svc.ProposeEvolution(ctx, agentID, analysis)
	require.NoError(t, err)
	assert.Equal(t, identity.EvolutionTypeSpawnSpecialist, proposal.ProposalType)
	assert.Equal(t, "pending", proposal.Status)
	assert.False(t, proposal.ParentApprovalRequired)
}

func TestRejectProposalUpdatesStatus(t *testing.T) {
	t.Parallel()

	db := newEvolutionTestDB(t)
	ctx := context.Background()
	svc := NewService(db, nil, nil)

	proposal := identity.EvolutionProposal{
		ID:           uuid.New(),
		AgentID:      "agent-reject",
		ProposalType: identity.EvolutionTypeModifyPolicy,
		Status:       "pending",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, proposal.SetProposalData(map[string]any{"reason": "test"}))
	require.NoError(t, db.Create(&proposal).Error)

	updated, err := svc.RejectProposal(ctx, proposal.ID, "agent-reject")
	require.NoError(t, err)
	assert.Equal(t, "rejected", updated.Status)
	assert.Equal(t, "agent-reject", *updated.ApprovedBy)
}

func TestImplementProposalModifyPolicyAppliesPolicyChanges(t *testing.T) {
	t.Parallel()

	db := newEvolutionTestDB(t)
	ctx := context.Background()
	svc := NewService(db, nil, nil)

	agentID := "agent-policy"
	policyChanges := map[string]any{
		"timeout_ms":           45000,
		"max_execution_depth":  12,
		"max_recursion_depth":  4,
		"max_memory_growth_mb": 640,
		"deterministic_only":   true,
	}
	proposal := identity.EvolutionProposal{
		ID:           uuid.New(),
		AgentID:      agentID,
		ProposalType: identity.EvolutionTypeModifyPolicy,
		Status:       "approved",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, proposal.SetProposalData(map[string]any{"policy_changes": policyChanges}))
	require.NoError(t, db.Create(&proposal).Error)

	require.NoError(t, svc.ImplementProposal(ctx, proposal.ID))

	var persisted identity.EvolutionProposal
	require.NoError(t, db.First(&persisted, "id = ?", proposal.ID).Error)
	assert.Equal(t, "implemented", persisted.Status)
	require.NotNil(t, persisted.ImplementedAt)

	var policy agentpolicy.BehavioralPolicy
	require.NoError(t, db.First(&policy, "agent_id = ?", agentID).Error)
	assert.Equal(t, 45000, policy.MaxWallTimeMs)
	assert.Equal(t, 12, policy.MaxExecutionDepth)
	assert.Equal(t, 4, policy.MaxRecursionDepth)
	assert.Equal(t, 640, policy.MaxMemoryGrowthMB)
	assert.True(t, policy.DeterministicOnly)
}

func TestImplementProposalSpawnSpecialistCreatesChildAndRelationship(t *testing.T) {
	t.Parallel()

	db := newEvolutionTestDB(t)
	ctx := context.Background()
	svc := NewService(db, nil, nil)

	parentTenant := uuid.New()
	parentAgentID := "agent-parent"

	parent := identity.AgentIdentity{
		ID:                uuid.New(),
		TenantID:          parentTenant,
		AgentID:           parentAgentID,
		Name:              "Parent Agent",
		Status:            identity.AgentStatusActive,
		PlanTier:          "agent_starter",
		APIKeyHash:        "test-hash",
		SwarmRole:         identity.SwarmRoleManager,
		MaxChildAgents:    5,
		Capabilities:      identity.JSONBMap{"coordination": true},
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		AutonomousEnabled: false,
		EvolutionEnabled:  true,
	}
	require.NoError(t, db.Create(&parent).Error)
	require.NoError(t, db.Create(&identity.AgentQuotaConfig{
		ID:                uuid.New(),
		AgentID:           parentAgentID,
		MaxCallsPerMinute: 100,
		MaxCallsPerDay:    1000,
	}).Error)

	proposal := identity.EvolutionProposal{
		ID:           uuid.New(),
		AgentID:      parentAgentID,
		ProposalType: identity.EvolutionTypeSpawnSpecialist,
		Status:       "approved",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, proposal.SetProposalData(map[string]any{
		"reason":          "error_recovery",
		"specialist_role": identity.SwarmRoleWorker,
		"capabilities":    []any{"error_recovery", "retry_logic"},
	}))
	require.NoError(t, db.Create(&proposal).Error)

	require.NoError(t, svc.ImplementProposal(ctx, proposal.ID))

	var children []identity.AgentIdentity
	require.NoError(t, db.Where("parent_agent_id = ?", parentAgentID).Find(&children).Error)
	require.Len(t, children, 1, "specialist child should be created")
	assert.Contains(t, children[0].AgentID, parentAgentID+"-specialist-")

	var rels []identity.AgentRelationship
	require.NoError(t, db.Where("parent_agent_id = ?", parentAgentID).Find(&rels).Error)
	require.Len(t, rels, 1, "relationship should be created for specialist child")
	assert.Equal(t, children[0].AgentID, rels[0].ChildAgentID)
}

func newEvolutionTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	n := atomic.AddInt64(&evolutionTestDBCounter, 1)
	dsn := fmt.Sprintf("file:evolutiondb%d?mode=memory&cache=private", n)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS agent_identities (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			agent_id TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			plan_tier TEXT NOT NULL DEFAULT 'agent_starter',
			status TEXT NOT NULL DEFAULT 'active',
			api_key_hash TEXT,
			parent_agent_id TEXT,
			swarm_role TEXT NOT NULL DEFAULT 'worker',
			max_child_agents INTEGER NOT NULL DEFAULT 0,
			capabilities TEXT DEFAULT '{}',
			autonomous_enabled INTEGER NOT NULL DEFAULT 0,
			evolution_enabled INTEGER NOT NULL DEFAULT 0,
			trust_score REAL DEFAULT 0,
			economic_score REAL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS agent_quota_configs (
			id TEXT PRIMARY KEY,
			agent_id TEXT NOT NULL UNIQUE,
			max_calls_per_minute INTEGER NOT NULL DEFAULT 100,
			max_calls_per_day INTEGER NOT NULL DEFAULT 16667,
			max_state_writes_per_hr INTEGER NOT NULL DEFAULT 1000,
			max_cost_per_execution REAL NOT NULL DEFAULT 0.01,
			max_daily_spend_usd REAL NOT NULL DEFAULT 5.00,
			allowed_functions TEXT,
			forbidden_functions TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS agent_relationships (
			id TEXT PRIMARY KEY,
			parent_agent_id TEXT NOT NULL,
			child_agent_id TEXT NOT NULL,
			relationship_type TEXT NOT NULL DEFAULT 'parent',
			max_delegation_depth INTEGER NOT NULL DEFAULT 5,
			created_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS agent_evolution_proposals (
			id TEXT PRIMARY KEY,
			agent_id TEXT NOT NULL,
			proposal_type TEXT NOT NULL,
			proposal_data TEXT NOT NULL DEFAULT '{}',
			status TEXT NOT NULL DEFAULT 'pending',
			parent_approval_required INTEGER NOT NULL DEFAULT 1,
			simulated_outcome TEXT,
			approved_by TEXT,
			implemented_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS agent_behavioral_policies (
			agent_id TEXT PRIMARY KEY,
			max_execution_depth INTEGER NOT NULL DEFAULT 10,
			max_recursion_depth INTEGER NOT NULL DEFAULT 3,
			max_wall_time_ms INTEGER NOT NULL DEFAULT 300000,
			max_memory_growth_mb INTEGER NOT NULL DEFAULT 512,
			forbidden_functions TEXT,
			deterministic_only INTEGER NOT NULL DEFAULT 0,
			allowed_capabilities TEXT
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS agent_execution_records (
			id TEXT PRIMARY KEY,
			agent_id TEXT NOT NULL,
			tenant_id TEXT NOT NULL,
			function_id TEXT NOT NULL,
			function_uri TEXT,
			execution_id TEXT,
			session_id TEXT,
			call_depth INTEGER NOT NULL DEFAULT 0,
			retry_count INTEGER,
			input_hash TEXT,
			output_hash TEXT,
			memory_before_hash TEXT,
			memory_after_hash TEXT,
			cost_usd REAL NOT NULL DEFAULT 0,
			latency_ms INTEGER NOT NULL DEFAULT 0,
			outcome TEXT NOT NULL,
			error_code TEXT,
			policy_violation TEXT,
			object_key TEXT,
			timestamp DATETIME NOT NULL
		)
	`).Error)

	// Keep compile-time references so imports remain validated.
	_ = agentpolicy.BehavioralPolicy{}
	_ = attribution.AgentExecutionRecord{}
	_ = identity.AgentIdentity{}
	return db
}
