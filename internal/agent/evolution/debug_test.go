package evolution

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/agent/attribution"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDebugParentApproval(t *testing.T) {
	n := atomic.AddInt64(&evolutionTestDBCounter, 1)
	dsn := fmt.Sprintf("file:debug%d?mode=memory&cache=private", n)
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

	svc := NewService(db, nil, nil)
	tenantID := uuid.New()
	agentID := "agent-debug"
	now := time.Now()

	records := []attribution.AgentExecutionRecord{
		{ID: uuid.New(), AgentID: agentID, TenantID: tenantID, FunctionID: uuid.New(), ExecutionID: "exec-1", Outcome: attribution.OutcomeError, LatencyMs: 1200, CostUSD: 0.001, Timestamp: now},
		{ID: uuid.New(), AgentID: agentID, TenantID: tenantID, FunctionID: uuid.New(), ExecutionID: "exec-2", Outcome: attribution.OutcomeError, LatencyMs: 1300, CostUSD: 0.001, Timestamp: now.Add(-time.Minute)},
		{ID: uuid.New(), AgentID: agentID, TenantID: tenantID, FunctionID: uuid.New(), ExecutionID: "exec-3", Outcome: attribution.OutcomeError, LatencyMs: 900, CostUSD: 0.001, Timestamp: now.Add(-2*time.Minute)},
		{ID: uuid.New(), AgentID: agentID, TenantID: tenantID, FunctionID: uuid.New(), ExecutionID: "exec-4", Outcome: attribution.OutcomeSuccess, LatencyMs: 850, CostUSD: 0.001, Timestamp: now.Add(-3*time.Minute)},
		{ID: uuid.New(), AgentID: agentID, TenantID: tenantID, FunctionID: uuid.New(), ExecutionID: "exec-5", Outcome: attribution.OutcomeSuccess, LatencyMs: 700, CostUSD: 0.001, Timestamp: now.Add(-4*time.Minute)},
	}
	for i := range records {
		require.NoError(t, db.Create(&records[i]).Error)
	}

	analysis, err := svc.AnalyzePerformance(context.Background(), agentID, 24*time.Hour)
	require.NoError(t, err)
	t.Logf("SuccessRate=%.2f, Total=%d", analysis.SuccessRate, analysis.TotalExecutions)

	proposal, err := svc.ProposeEvolution(context.Background(), agentID, analysis)
	require.NoError(t, err)
	t.Logf("After ProposeEvolution: ProposalType=%s, ParentApprovalRequired=%v", proposal.ProposalType, proposal.ParentApprovalRequired)

	// Read from DB
	var fromDB identity.EvolutionProposal
	require.NoError(t, db.First(&fromDB, "id = ?", proposal.ID).Error)
	t.Logf("FromDB: ParentApprovalRequired=%v", fromDB.ParentApprovalRequired)
}
