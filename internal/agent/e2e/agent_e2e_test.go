package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/agent/autonomy"
	"github.com/functionfly/functionfly/internal/agent/economy"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/agent/lifecycle"
	"github.com/functionfly/functionfly/internal/agent/swarm"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Test fixtures and helpers

type e2eTestEnv struct {
	db              *gorm.DB
	identityRepo    *identity.Repository
	swarmService    *swarm.Service
	messageService  *swarm.MessageService
	walletService   *economy.Service
	lifecycleSvc    *lifecycle.Service
	autonomySvc     *autonomy.Service
	tenantID        uuid.UUID
	logger          *logrus.Logger
}

func newE2ETestEnv(t *testing.T) *e2eTestEnv {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate all required models
	err = db.AutoMigrate(
		&identity.AgentIdentity{},
		&identity.AgentRelationship{},
		&identity.AgentMessage{},
		&identity.AgentWallet{},
		&identity.AgentQuotaConfig{},
		&identity.RevenueTransaction{},
		&identity.AutonomySchedule{},
		&identity.EvolutionProposal{},
		&identity.AgentHiring{},
		&identity.FunctionPurchase{},
		&lifecycle.AgentLifecycleEvent{},
	)
	require.NoError(t, err)

	tenantID := uuid.New()
	identityRepo := identity.NewRepository(db)

	// Create economy service
	walletService := economy.NewService(db)

	// Create message service (without Redis for testing)
	messageService := swarm.NewMessageService(db, nil)

	// Create swarm service
	swarmService := swarm.NewService(db, identityRepo, walletService)

	// Create lifecycle service
	lifecycleSvc := lifecycle.NewService(db, nil, logger)

	// Create autonomy service
	autonomySvc := autonomy.NewService(db)

	return &e2eTestEnv{
		db:              db,
		identityRepo:    identityRepo,
		swarmService:    swarmService,
		messageService:  messageService,
		walletService:   walletService,
		lifecycleSvc:    lifecycleSvc,
		autonomySvc:     autonomySvc,
		tenantID:        tenantID,
		logger:          logger,
	}
}

// ============================================================
// FEATURE 1: Agent Registration & Identity
// ============================================================

func TestAgentRegistration(t *testing.T) {
	t.Run("should register a new agent with valid credentials", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		req := &identity.RegisterAgentRequest{
			AgentID:     "test-org/test-agent",
			Name:        "Test Agent",
			Description: "A test agent for e2e testing",
			PlanTier:    "agent_starter",
		}

		agent, apiKey, signingKey, err := env.identityRepo.CreateAgent(ctx, env.tenantID, req)
		require.NoError(t, err)

		assert.NotEmpty(t, apiKey)
		assert.NotEmpty(t, signingKey)
		assert.Equal(t, "test-org/test-agent", agent.AgentID)
		assert.Equal(t, "Test Agent", agent.Name)
		assert.Equal(t, identity.AgentStatusActive, agent.Status)
		assert.Equal(t, identity.SwarmRoleWorker, agent.SwarmRole)
	})

	t.Run("should reject duplicate agent IDs", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		req := &identity.RegisterAgentRequest{
			AgentID:  "test-org/duplicate-agent",
			Name:     "First Agent",
			PlanTier: "agent_starter",
		}

		_, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, req)
		require.NoError(t, err)

		// Try to register again
		req2 := &identity.RegisterAgentRequest{
			AgentID:  "test-org/duplicate-agent",
			Name:     "Second Agent",
			PlanTier: "agent_starter",
		}
		_, _, _, err = env.identityRepo.CreateAgent(ctx, env.tenantID, req2)
		assert.Error(t, err)
	})

	t.Run("should create default quota config on registration", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		req := &identity.RegisterAgentRequest{
			AgentID:  "test-org/quota-agent",
			Name:     "Quota Agent",
			PlanTier: "agent_starter",
		}

		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, req)
		require.NoError(t, err)

		quota, err := env.identityRepo.GetQuotaConfig(ctx, agent.AgentID)
		require.NoError(t, err)

		assert.Equal(t, agent.AgentID, quota.AgentID)
		assert.Greater(t, quota.MaxCallsPerMinute, 0)
		assert.Greater(t, quota.MaxCallsPerDay, 0)
	})

	t.Run("should list agents for a tenant", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		// Create multiple agents
		for i := 0; i < 3; i++ {
			req := &identity.RegisterAgentRequest{
				AgentID:  fmt.Sprintf("test-org/list-agent-%d", i),
				Name:     fmt.Sprintf("List Agent %d", i),
				PlanTier: "agent_starter",
			}
			_, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, req)
			require.NoError(t, err)
		}

		agents, total, err := env.identityRepo.ListAgents(ctx, env.tenantID, 10, 0)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(agents), 3)
		assert.GreaterOrEqual(t, total, int64(3))
	})

	t.Run("should rotate API key", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		req := &identity.RegisterAgentRequest{
			AgentID:  "test-org/rotate-agent",
			Name:     "Rotate Agent",
			PlanTier: "agent_starter",
		}

		agent, originalKey, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, req)
		require.NoError(t, err)

		newKey, err := env.identityRepo.RotateAPIKey(ctx, agent.AgentID)
		require.NoError(t, err)

		assert.NotEqual(t, originalKey, newKey)
		assert.True(t, len(newKey) > 10)
	})
}

// ============================================================
// FEATURE 2: Swarm Management (Parent-Child Hierarchy)
// ============================================================

func TestSwarmSpawnChild(t *testing.T) {
	t.Run("should spawn a child agent with correct parent relationship", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		// Create parent agent
		parentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/parent-agent",
			Name:     "Parent Agent",
			PlanTier: "agent_starter",
		}
		parent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, parentReq)
		require.NoError(t, err)

		// Spawn child
		spawnReq := &swarm.SpawnChildRequest{
			ParentAgentID:    parent.AgentID,
			ChildAgentID:     "test-org/child-agent",
			ChildName:        "Child Agent",
			ChildDescription: "A child agent",
			SwarmRole:        identity.SwarmRoleWorker,
			MaxChildAgents:   5,
			InitialBudgetUSD: 10.00,
		}

		child, apiKey, err := env.swarmService.SpawnChild(ctx, spawnReq)
		require.NoError(t, err)
		require.NotEmpty(t, apiKey)

		assert.Equal(t, "test-org/child-agent", child.AgentID)
		assert.Equal(t, parent.AgentID, *child.ParentAgentID)
		assert.Equal(t, identity.SwarmRoleWorker, child.SwarmRole)
	})

	t.Run("should track children count correctly", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		// Create parent with max children
		parentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/max-children-parent",
			Name:     "Max Children Parent",
			PlanTier: "agent_enterprise",
		}
		parent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, parentReq)
		require.NoError(t, err)

		// Update max children
		env.db.Model(&identity.AgentIdentity{}).Where("agent_id = ?", parent.AgentID).
			Update("max_child_agents", 3)

		// Spawn 2 children
		for i := 0; i < 2; i++ {
			spawnReq := &swarm.SpawnChildRequest{
				ParentAgentID: parent.AgentID,
				ChildAgentID:  fmt.Sprintf("test-org/child-%d", i),
				ChildName:     fmt.Sprintf("Child %d", i),
				SwarmRole:     identity.SwarmRoleWorker,
			}
			_, _, err := env.swarmService.SpawnChild(ctx, spawnReq)
			require.NoError(t, err)
		}

		count, err := env.swarmService.CountChildren(ctx, parent.AgentID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("should reject spawning when at max capacity", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		parentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/at-capacity-parent",
			Name:     "At Capacity Parent",
			PlanTier: "agent_starter",
		}
		parent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, parentReq)
		require.NoError(t, err)

		// Set max children to 1
		env.db.Model(&identity.AgentIdentity{}).Where("agent_id = ?", parent.AgentID).
			Update("max_child_agents", 1)

		// Spawn first child
		spawnReq1 := &swarm.SpawnChildRequest{
			ParentAgentID: parent.AgentID,
			ChildAgentID:  "test-org/child-1",
			ChildName:     "Child 1",
			SwarmRole:     identity.SwarmRoleWorker,
		}
		_, _, err = env.swarmService.SpawnChild(ctx, spawnReq1)
		require.NoError(t, err)

		// Try to spawn second child - should fail
		spawnReq2 := &swarm.SpawnChildRequest{
			ParentAgentID: parent.AgentID,
			ChildAgentID:  "test-org/child-2",
			ChildName:     "Child 2",
			SwarmRole:     identity.SwarmRoleWorker,
		}
		_, _, err = env.swarmService.SpawnChild(ctx, spawnReq2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max child capacity")
	})

	t.Run("should terminate child and remove relationship", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		parentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/terminate-parent",
			Name:     "Terminate Parent",
			PlanTier: "agent_starter",
		}
		parent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, parentReq)
		require.NoError(t, err)

		spawnReq := &swarm.SpawnChildRequest{
			ParentAgentID: parent.AgentID,
			ChildAgentID:  "test-org/terminate-child",
			ChildName:     "Terminate Child",
			SwarmRole:     identity.SwarmRoleWorker,
		}
		_, _, err = env.swarmService.SpawnChild(ctx, spawnReq)
		require.NoError(t, err)

		err = env.swarmService.TerminateChild(ctx, parent.AgentID, "test-org/terminate-child")
		require.NoError(t, err)

		count, err := env.swarmService.CountChildren(ctx, parent.AgentID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("should get parent for child agent", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		parentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/parent-lookup",
			Name:     "Parent Lookup",
			PlanTier: "agent_starter",
		}
		parent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, parentReq)
		require.NoError(t, err)

		spawnReq := &swarm.SpawnChildRequest{
			ParentAgentID: parent.AgentID,
			ChildAgentID:  "test-org/child-lookup",
			ChildName:     "Child Lookup",
			SwarmRole:     identity.SwarmRoleWorker,
		}
		_, _, err = env.swarmService.SpawnChild(ctx, spawnReq)
		require.NoError(t, err)

		foundParent, err := env.swarmService.GetParent(ctx, "test-org/child-lookup")
		require.NoError(t, err)
		assert.NotNil(t, foundParent)
		assert.Equal(t, parent.AgentID, foundParent.AgentID)
	})

	t.Run("should reassign swarm role", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		req := &identity.RegisterAgentRequest{
			AgentID:  "test-org/role-assign",
			Name:     "Role Assign",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, req)
		require.NoError(t, err)

		err = env.swarmService.ReassignRole(ctx, agent.AgentID, identity.SwarmRoleManager)
		require.NoError(t, err)

		updated, err := env.identityRepo.GetAgent(ctx, agent.AgentID)
		require.NoError(t, err)
		assert.Equal(t, identity.SwarmRoleManager, updated.SwarmRole)
	})

	t.Run("should reshape swarm topology", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		req := &identity.RegisterAgentRequest{
			AgentID:  "test-org/topology-change",
			Name:     "Topology Change",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, req)
		require.NoError(t, err)

		err = env.swarmService.ReshapeSwarm(ctx, agent.AgentID, identity.SwarmTopologyStar)
		require.NoError(t, err)

		updated, err := env.identityRepo.GetAgent(ctx, agent.AgentID)
		require.NoError(t, err)
		assert.Equal(t, identity.SwarmTopologyStar, updated.SwarmTopology)
	})

	t.Run("should calculate delegation depth", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		// Create grandparent
		gpReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/grandparent",
			Name:     "Grandparent",
			PlanTier: "agent_enterprise",
		}
		gp, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, gpReq)
		require.NoError(t, err)

		// Create parent
		pReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/depth-parent",
			Name:     "Depth Parent",
			PlanTier: "agent_starter",
		}
		p, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, pReq)
		require.NoError(t, err)

		// Create child
		cReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/depth-child",
			Name:     "Depth Child",
			PlanTier: "agent_starter",
		}
		c, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, cReq)
		require.NoError(t, err)

		// Create relationships (parent -> child)
		spawnReq := &swarm.SpawnChildRequest{
			ParentAgentID: gp.AgentID,
			ChildAgentID:  p.AgentID,
			ChildName:     "Parent",
			SwarmRole:     identity.SwarmRoleManager,
		}
		_, _, err = env.swarmService.SpawnChild(ctx, spawnReq)
		require.NoError(t, err)

		spawnReq2 := &swarm.SpawnChildRequest{
			ParentAgentID: p.AgentID,
			ChildAgentID:  c.AgentID,
			ChildName:     "Child",
			SwarmRole:     identity.SwarmRoleWorker,
		}
		_, _, err = env.swarmService.SpawnChild(ctx, spawnReq2)
		require.NoError(t, err)

		depth, err := env.swarmService.GetDelegationDepth(ctx, c.AgentID)
		require.NoError(t, err)
		assert.Equal(t, 2, depth)
	})
}

// ============================================================
// FEATURE 3: Agent-to-Agent (A2A) Communication
// ============================================================

func TestAgentMessaging(t *testing.T) {
	t.Run("should send and receive signed task delegation messages", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		// Create two agents
		senderReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/sender",
			Name:     "Sender Agent",
			PlanTier: "agent_starter",
		}
		sender, senderKey, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, senderReq)
		require.NoError(t, err)

		receiverReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/receiver",
			Name:     "Receiver Agent",
			PlanTier: "agent_starter",
		}
		receiver, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, receiverReq)
		require.NoError(t, err)

		// Send task delegation
		taskData := map[string]any{
			"task_type": "scan_source",
			"task_id":   uuid.New().String(),
			"payload":   map[string]any{"query": "test"},
		}
		sessionID := "sess_" + uuid.New().String()

		err = env.messageService.SendTaskDelegation(ctx, sender.AgentID, receiver.AgentID, taskData, sessionID, senderKey)
		require.NoError(t, err)

		// Verify message was stored
		inbox, err := env.messageService.GetInbox(ctx, receiver.AgentID, 10)
		require.NoError(t, err)

		found := false
		for _, msg := range inbox {
			if msg.MessageType == identity.MessageTypeTaskDelegation && msg.FromAgentID == sender.AgentID {
				found = true
				assert.NotEmpty(t, msg.Signature)
				assert.NotEmpty(t, msg.Nonce)
				assert.Equal(t, sessionID, *msg.SessionID)
				break
			}
		}
		assert.True(t, found, "task delegation message not found in receiver's inbox")
	})

	t.Run("should mark messages as delivered", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		// Create agents
		senderReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/deliver-sender",
			Name:     "Deliver Sender",
			PlanTier: "agent_starter",
		}
		sender, senderKey, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, senderReq)
		require.NoError(t, err)

		receiverReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/deliver-receiver",
			Name:     "Deliver Receiver",
			PlanTier: "agent_starter",
		}
		receiver, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, receiverReq)
		require.NoError(t, err)

		// Send message
		err = env.messageService.SendTaskDelegation(ctx, sender.AgentID, receiver.AgentID,
			map[string]any{"test": "data"}, "sess_123", senderKey)
		require.NoError(t, err)

		// Get message from inbox
		inbox, err := env.messageService.GetInbox(ctx, receiver.AgentID, 10)
		require.NoError(t, err)
		require.NotEmpty(t, inbox)

		// Mark as delivered
		err = env.messageService.MarkDelivered(ctx, inbox[0].ID)
		require.NoError(t, err)

		// Verify status
		updatedInbox, err := env.messageService.GetInbox(ctx, receiver.AgentID, 10)
		require.NoError(t, err)
		assert.Equal(t, "delivered", updatedInbox[0].Status)
	})

	t.Run("should send heartbeat messages", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		senderReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/heartbeat-sender",
			Name:     "Heartbeat Sender",
			PlanTier: "agent_starter",
		}
		sender, senderKey, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, senderReq)
		require.NoError(t, err)

		receiverReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/heartbeat-receiver",
			Name:     "Heartbeat Receiver",
			PlanTier: "agent_starter",
		}
		receiver, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, receiverReq)
		require.NoError(t, err)

		err = env.messageService.SendHeartbeat(ctx, sender.AgentID, receiver.AgentID, senderKey)
		require.NoError(t, err)

		// Verify in inbox
		inbox, err := env.messageService.GetInbox(ctx, receiver.AgentID, 10)
		require.NoError(t, err)

		found := false
		for _, msg := range inbox {
			if msg.MessageType == identity.MessageTypeHeartbeat {
				found = true
				break
			}
		}
		assert.True(t, found, "heartbeat message not found")
	})

	t.Run("should get outbox for agent", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		senderReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/outbox-sender",
			Name:     "Outbox Sender",
			PlanTier: "agent_starter",
		}
		sender, senderKey, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, senderReq)
		require.NoError(t, err)

		receiverReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/outbox-receiver",
			Name:     "Outbox Receiver",
			PlanTier: "agent_starter",
		}
		receiver, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, receiverReq)
		require.NoError(t, err)

		// Send multiple messages
		for i := 0; i < 3; i++ {
			err = env.messageService.SendTaskDelegation(ctx, sender.AgentID, receiver.AgentID,
				map[string]any{"index": i}, fmt.Sprintf("sess_%d", i), senderKey)
			require.NoError(t, err)
		}

		outbox, err := env.messageService.GetOutbox(ctx, sender.AgentID, 10)
		require.NoError(t, err)
		assert.Len(t, outbox, 3)
	})

	t.Run("should cleanup expired messages", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		// Create and send a message
		senderReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/cleanup-sender",
			Name:     "Cleanup Sender",
			PlanTier: "agent_starter",
		}
		sender, senderKey, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, senderReq)
		require.NoError(t, err)

		receiverReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/cleanup-receiver",
			Name:     "Cleanup Receiver",
			PlanTier: "agent_starter",
		}
		receiver, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, receiverReq)
		require.NoError(t, err)

		err = env.messageService.SendTaskDelegation(ctx, sender.AgentID, receiver.AgentID,
			map[string]any{"test": "cleanup"}, "sess_cleanup", senderKey)
		require.NoError(t, err)

		// Mark as pending (already is by default)
		deleted, err := env.messageService.CleanupExpired(ctx)
		require.NoError(t, err)
		// Should not delete recent messages
		assert.Equal(t, int64(0), deleted)
	})
}

// ============================================================
// FEATURE 4: Lifecycle Management
// ============================================================

func TestAgentLifecycle(t *testing.T) {
	t.Run("should register agent for lifecycle management", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		req := &identity.RegisterAgentRequest{
			AgentID:  "test-org/lifecycle-agent",
			Name:     "Lifecycle Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, req)
		require.NoError(t, err)

		err = env.lifecycleSvc.RegisterAgent(ctx, agent.AgentID)
		require.NoError(t, err)

		state, err := env.lifecycleSvc.GetLifecycleStatus(ctx, agent.AgentID)
		require.NoError(t, err)
		assert.Equal(t, string(lifecycle.LifecycleStatusActive), state.LifecycleStatus)
	})

	t.Run("should record heartbeat and update status", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		req := &identity.RegisterAgentRequest{
			AgentID:  "test-org/heartbeat-agent",
			Name:     "Heartbeat Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, req)
		require.NoError(t, err)

		err = env.lifecycleSvc.RegisterAgent(ctx, agent.AgentID)
		require.NoError(t, err)

		stateSnapshot := lifecycle.JSONMap{"tasks_completed": 5, "errors": 0}
		err = env.lifecycleSvc.RecordHeartbeat(ctx, agent.AgentID, stateSnapshot)
		require.NoError(t, err)

		isAlive := env.lifecycleSvc.IsAgentAlive(ctx, agent.AgentID)
		assert.True(t, isAlive)
	})

	t.Run("should initiate and complete graceful shutdown", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		req := &identity.RegisterAgentRequest{
			AgentID:  "test-org/shutdown-agent",
			Name:     "Shutdown Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, req)
		require.NoError(t, err)

		err = env.lifecycleSvc.RegisterAgent(ctx, agent.AgentID)
		require.NoError(t, err)

		err = env.lifecycleSvc.InitiateGracefulShutdown(ctx, agent.AgentID)
		require.NoError(t, err)

		state, err := env.lifecycleSvc.GetLifecycleStatus(ctx, agent.AgentID)
		require.NoError(t, err)
		assert.Equal(t, string(lifecycle.LifecycleStatusGracefulShutdownStart), state.LifecycleStatus)

		err = env.lifecycleSvc.CompleteGracefulShutdown(ctx, agent.AgentID)
		require.NoError(t, err)

		state, err = env.lifecycleSvc.GetLifecycleStatus(ctx, agent.AgentID)
		require.NoError(t, err)
		assert.Equal(t, string(lifecycle.LifecycleStatusGracefulShutdownDone), state.LifecycleStatus)
	})

	t.Run("should record orphan detection", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		req := &identity.RegisterAgentRequest{
			AgentID:  "test-org/orphan-agent",
			Name:     "Orphan Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, req)
		require.NoError(t, err)

		err = env.lifecycleSvc.RegisterAgent(ctx, agent.AgentID)
		require.NoError(t, err)

		err = env.lifecycleSvc.RecordOrphanDetection(ctx, agent.AgentID)
		require.NoError(t, err)

		state, err := env.lifecycleSvc.GetLifecycleStatus(ctx, agent.AgentID)
		require.NoError(t, err)
		assert.Equal(t, string(lifecycle.LifecycleStatusOrphaned), state.LifecycleStatus)
	})

	t.Run("should save and retrieve state snapshots", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		req := &identity.RegisterAgentRequest{
			AgentID:  "test-org/snapshot-agent",
			Name:     "Snapshot Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, req)
		require.NoError(t, err)

		err = env.lifecycleSvc.RegisterAgent(ctx, agent.AgentID)
		require.NoError(t, err)

		snapshot := lifecycle.JSONMap{
			"memory":        map[string]any{"key1": "value1"},
			"task_history":  []any{"task1", "task2"},
			"last_position": 42,
		}

		err = env.lifecycleSvc.SaveStateSnapshot(ctx, agent.AgentID, snapshot)
		require.NoError(t, err)

		retrieved, err := env.lifecycleSvc.GetStateSnapshot(ctx, agent.AgentID)
		require.NoError(t, err)
		assert.Equal(t, 42, retrieved["last_position"])
	})

	t.Run("should track lifecycle events", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		req := &identity.RegisterAgentRequest{
			AgentID:  "test-org/events-agent",
			Name:     "Events Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, req)
		require.NoError(t, err)

		err = env.lifecycleSvc.RegisterAgent(ctx, agent.AgentID)
		require.NoError(t, err)

		err = env.lifecycleSvc.RecordHeartbeat(ctx, agent.AgentID, nil)
		require.NoError(t, err)

		events, err := env.lifecycleSvc.GetEvents(ctx, agent.AgentID, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(events), 1)

		// Should contain heartbeat event
		found := false
		for _, e := range events {
			if e.EventType == lifecycle.EventTypeHeartbeat {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}

// ============================================================
// FEATURE 5: Agent Economy (Wallets & Transactions)
// ============================================================

func TestAgentEconomy(t *testing.T) {
	t.Run("should create wallet on first access", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/wallet-agent",
			Name:     "Wallet Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		wallet, err := env.walletService.GetOrCreateWallet(ctx, agent.AgentID)
		require.NoError(t, err)

		assert.Equal(t, agent.AgentID, wallet.AgentID)
		assert.Equal(t, 0.0, wallet.BalanceUSD)
	})

	t.Run("should credit wallet and track earnings", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/credit-agent",
			Name:     "Credit Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		tx, err := env.walletService.Credit(ctx, agent.AgentID, 100.00, identity.TransactionTypeDelegationPayment, nil)
		require.NoError(t, err)

		assert.Equal(t, 100.00, tx.AmountUSD)
		assert.Equal(t, agent.AgentID, tx.ToAgentID)

		wallet, err := env.walletService.GetWallet(ctx, agent.AgentID)
		require.NoError(t, err)
		assert.Equal(t, 100.00, wallet.BalanceUSD)
		assert.Equal(t, 100.00, wallet.TotalEarnedUSD)
	})

	t.Run("should debit wallet and track spending", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/debit-agent",
			Name:     "Debit Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		// Credit first
		_, err = env.walletService.Credit(ctx, agent.AgentID, 100.00, identity.TransactionTypeDelegationPayment, nil)
		require.NoError(t, err)

		// Debit
		tx, err := env.walletService.Debit(ctx, agent.AgentID, 30.00, identity.TransactionTypeFunctionCall, nil)
		require.NoError(t, err)

		assert.Equal(t, 30.00, tx.AmountUSD)

		wallet, err := env.walletService.GetWallet(ctx, agent.AgentID)
		require.NoError(t, err)
		assert.Equal(t, 70.00, wallet.BalanceUSD)
		assert.Equal(t, 30.00, wallet.TotalSpentUSD)
	})

	t.Run("should reject debit with insufficient funds", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/insufficient-agent",
			Name:     "Insufficient Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		_, err = env.walletService.Credit(ctx, agent.AgentID, 50.00, identity.TransactionTypeDelegationPayment, nil)
		require.NoError(t, err)

		_, err = env.walletService.Debit(ctx, agent.AgentID, 100.00, identity.TransactionTypeFunctionCall, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
	})

	t.Run("should transfer funds between agents", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		// Create two agents
		agent1Req := &identity.RegisterAgentRequest{
			AgentID:  "test-org/transfer-from",
			Name:     "Transfer From",
			PlanTier: "agent_starter",
		}
		agent1, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agent1Req)
		require.NoError(t, err)

		agent2Req := &identity.RegisterAgentRequest{
			AgentID:  "test-org/transfer-to",
			Name:     "Transfer To",
			PlanTier: "agent_starter",
		}
		agent2, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agent2Req)
		require.NoError(t, err)

		// Credit agent1
		_, err = env.walletService.Credit(ctx, agent1.AgentID, 100.00, identity.TransactionTypeDelegationPayment, nil)
		require.NoError(t, err)

		// Transfer to agent2
		tx, err := env.walletService.Transfer(ctx, agent1.AgentID, agent2.AgentID, 40.00, identity.TransactionTypeDelegationPayment, nil)
		require.NoError(t, err)
		assert.Equal(t, 40.00, tx.AmountUSD)

		wallet1, _ := env.walletService.GetWallet(ctx, agent1.AgentID)
		wallet2, _ := env.walletService.GetWallet(ctx, agent2.AgentID)

		assert.Equal(t, 60.00, wallet1.BalanceUSD)
		assert.Equal(t, 40.00, wallet2.BalanceUSD)
	})

	t.Run("should get transaction history", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/history-agent",
			Name:     "History Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		// Create multiple transactions
		for i := 0; i < 5; i++ {
			_, err = env.walletService.Credit(ctx, agent.AgentID, float64(i+1)*10, identity.TransactionTypeDelegationPayment, nil)
			require.NoError(t, err)
		}

		txs, total, err := env.walletService.GetTransactions(ctx, agent.AgentID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, txs, 5)
		assert.Equal(t, int64(5), total)
	})

	t.Run("should process delegation payment with revenue split", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		parentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/split-parent",
			Name:     "Split Parent",
			PlanTier: "agent_starter",
		}
		parent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, parentReq)
		require.NoError(t, err)

		childReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/split-child",
			Name:     "Split Child",
			PlanTier: "agent_starter",
		}
		child, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, childReq)
		require.NoError(t, err)

		// Process delegation payment with 70/30 split
		err = env.walletService.ProcessDelegationPayment(ctx, parent.AgentID, child.AgentID, 100.00, 70.0, nil)
		require.NoError(t, err)

		parentWallet, _ := env.walletService.GetWallet(ctx, parent.AgentID)
		childWallet, _ := env.walletService.GetWallet(ctx, child.AgentID)

		// Parent gets 70, child gets 30
		assert.Equal(t, 70.00, parentWallet.BalanceUSD)
		assert.Equal(t, 30.00, childWallet.BalanceUSD)
	})

	t.Run("should add and release from escrow", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/escrow-agent",
			Name:     "Escrow Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		// Credit wallet first
		_, err = env.walletService.Credit(ctx, agent.AgentID, 100.00, identity.TransactionTypeDelegationPayment, nil)
		require.NoError(t, err)

		// Add to escrow
		err = env.walletService.AddToEscrow(ctx, agent.AgentID, 50.00)
		require.NoError(t, err)

		wallet, _ := env.walletService.GetWallet(ctx, agent.AgentID)
		assert.Equal(t, 50.00, wallet.BalanceUSD)
		assert.Equal(t, 50.00, wallet.EscrowBalanceUSD)

		// Release from escrow
		err = env.walletService.ReleaseFromEscrow(ctx, agent.AgentID, 30.00, "")
		require.NoError(t, err)

		wallet, _ = env.walletService.GetWallet(ctx, agent.AgentID)
		assert.Equal(t, 80.00, wallet.BalanceUSD)
		assert.Equal(t, 20.00, wallet.EscrowBalanceUSD)
	})
}

// ============================================================
// FEATURE 6: Autonomy & Scheduled Execution
// ============================================================

func TestAgentAutonomy(t *testing.T) {
	t.Run("should create recurring schedule", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/schedule-agent",
			Name:     "Schedule Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		schedule := &identity.AutonomySchedule{
			AgentID:        agent.AgentID,
			ScheduleType:   identity.AutonomyScheduleRecurring,
			CronExpression: strPtr("0 * * * *"), // Every hour
			ActionType:     identity.AutonomyActionExecuteFunction,
			ActionPayload:  map[string]any{"function": "test-func"},
		}

		err = env.autonomySvc.CreateSchedule(ctx, schedule)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, schedule.ID)
		assert.True(t, schedule.IsActive)
		assert.NotNil(t, schedule.NextRunAt)
	})

	t.Run("should create one-time schedule", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/onetime-agent",
			Name:     "OneTime Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		nextRun := time.Now().Add(1 * time.Hour)
		schedule := &identity.AutonomySchedule{
			AgentID:      agent.AgentID,
			ScheduleType: identity.AutonomyScheduleOneTime,
			NextRunAt:    &nextRun,
			ActionType:   identity.AutonomyActionSpawnAgent,
			ActionPayload: map[string]any{"name": "spawned-agent"},
		}

		err = env.autonomySvc.CreateSchedule(ctx, schedule)
		require.NoError(t, err)
		assert.False(t, schedule.IsActive) // One-time schedules deactivate after run
	})

	t.Run("should create trigger-based schedule", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/trigger-agent",
			Name:     "Trigger Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		schedule := &identity.AutonomySchedule{
			AgentID:        agent.AgentID,
			ScheduleType:   identity.AutonomyScheduleTriggerBased,
			TriggerEvent:   strPtr("task_completed"),
			TriggerCondition: map[string]any{"status": "success"},
			ActionType:     identity.AutonomyActionSendMessage,
			ActionPayload:  map[string]any{"recipient": "alerts-channel"},
		}

		err = env.autonomySvc.CreateSchedule(ctx, schedule)
		require.NoError(t, err)
	})

	t.Run("should list schedules for agent", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/list-schedules-agent",
			Name:     "List Schedules Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		// Create multiple schedules
		for i := 0; i < 3; i++ {
			schedule := &identity.AutonomySchedule{
				AgentID:      agent.AgentID,
				ScheduleType: identity.AutonomyScheduleRecurring,
				CronExpression: strPtr("0 * * * *"),
				ActionType:   identity.AutonomyActionExecuteFunction,
				ActionPayload: map[string]any{"index": i},
			}
			err = env.autonomySvc.CreateSchedule(ctx, schedule)
			require.NoError(t, err)
		}

		schedules, err := env.autonomySvc.GetSchedules(ctx, agent.AgentID)
		require.NoError(t, err)
		assert.Len(t, schedules, 3)
	})

	t.Run("should execute schedule and update last run time", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/execute-schedule-agent",
			Name:     "Execute Schedule Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		schedule := &identity.AutonomySchedule{
			AgentID:        agent.AgentID,
			ScheduleType:   identity.AutonomyScheduleRecurring,
			CronExpression: strPtr("*/5 * * * *"),
			ActionType:     identity.AutonomyActionUpdateState,
			ActionPayload:  map[string]any{"key": "value"},
		}
		err = env.autonomySvc.CreateSchedule(ctx, schedule)
		require.NoError(t, err)

		result, err := env.autonomySvc.ExecuteSchedule(ctx, schedule.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify last run was updated
		updatedSchedules, err := env.autonomySvc.GetSchedules(ctx, agent.AgentID)
		require.NoError(t, err)
		require.NotEmpty(t, updatedSchedules)
		assert.NotNil(t, updatedSchedules[0].LastRunAt)
	})

	t.Run("should deactivate and activate schedule", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/toggle-schedule-agent",
			Name:     "Toggle Schedule Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		schedule := &identity.AutonomySchedule{
			AgentID:        agent.AgentID,
			ScheduleType:   identity.AutonomyScheduleRecurring,
			CronExpression: strPtr("0 * * * *"),
			ActionType:     identity.AutonomyActionEvolve,
		}
		err = env.autonomySvc.CreateSchedule(ctx, schedule)
		require.NoError(t, err)

		// Deactivate
		err = env.autonomySvc.DeactivateSchedule(ctx, schedule.ID)
		require.NoError(t, err)

		// Activate
		err = env.autonomySvc.ActivateSchedule(ctx, schedule.ID)
		require.NoError(t, err)
	})

	t.Run("should delete schedule", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/delete-schedule-agent",
			Name:     "Delete Schedule Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		schedule := &identity.AutonomySchedule{
			AgentID:        agent.AgentID,
			ScheduleType:   identity.AutonomyScheduleRecurring,
			CronExpression: strPtr("0 * * * *"),
			ActionType:     identity.AutonomyActionExecuteFunction,
		}
		err = env.autonomySvc.CreateSchedule(ctx, schedule)
		require.NoError(t, err)

		err = env.autonomySvc.DeleteSchedule(ctx, schedule.ID)
		require.NoError(t, err)

		schedules, _ := env.autonomySvc.GetSchedules(ctx, agent.AgentID)
		assert.Len(t, schedules, 0)
	})

	t.Run("should process triggered schedules based on event data", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/process-trigger-agent",
			Name:     "Process Trigger Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		schedule := &identity.AutonomySchedule{
			AgentID:           agent.AgentID,
			ScheduleType:      identity.AutonomyScheduleTriggerBased,
			TriggerEvent:      strPtr("task_completed"),
			TriggerCondition:  map[string]any{"status": "success"},
			ActionType:        identity.AutonomyActionSendMessage,
			ActionPayload:     map[string]any{"message": "Task completed!"},
		}
		err = env.autonomySvc.CreateSchedule(ctx, schedule)
		require.NoError(t, err)

		// Process with matching event data
		executed, err := env.autonomySvc.ProcessTriggeredSchedules(ctx, "task_completed",
			map[string]any{"status": "success", "task_id": "123"})
		require.NoError(t, err)
		assert.Len(t, executed, 1)

		// Process with non-matching event data
		executed, err = env.autonomySvc.ProcessTriggeredSchedules(ctx, "task_completed",
			map[string]any{"status": "failed", "task_id": "456"})
		require.NoError(t, err)
		assert.Len(t, executed, 0) // Should not trigger
	})
}

// ============================================================
// FEATURE 7: Agent Hiring & Function Purchasing
// ============================================================

func TestAgentMarketplace(t *testing.T) {
	t.Run("should create agent hiring record", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/hiring-agent",
			Name:     "Hiring Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		hiring := &identity.AgentHiring{
			ID:          uuid.New(),
			AgentID:     agent.AgentID,
			HirerID:     "user-123",
			TenantID:    env.tenantID.String(),
			TaskType:    "code_generation",
			TaskPayload: map[string]any{"spec": "generate login function"},
			BudgetUSD:   50.00,
			Status:      "pending",
		}

		err = env.identityRepo.CreateAgentHiring(ctx, hiring)
		require.NoError(t, err)

		// Verify hiring count
		count, err := env.identityRepo.CountAgentHiring(ctx, agent.AgentID)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("should create function purchase record", func(t *testing.T) {
		env := newE2ETestEnv(t)
		ctx := context.Background()

		agentReq := &identity.RegisterAgentRequest{
			AgentID:  "test-org/purchase-agent",
			Name:     "Purchase Agent",
			PlanTier: "agent_starter",
		}
		agent, _, _, err := env.identityRepo.CreateAgent(ctx, env.tenantID, agentReq)
		require.NoError(t, err)

		purchase := &identity.FunctionPurchase{
			ID:             uuid.New(),
			AgentID:        agent.AgentID,
			FunctionAuthor: "author123",
			FunctionName:   "login_function",
			PublishedID:    uuid.New(),
			PricePaidUSD:   25.00,
			Status:         "completed",
		}

		err = env.identityRepo.CreateFunctionPurchase(ctx, purchase)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, purchase.ID)
	})
}

// ============================================================
// Helper Functions
// ============================================================

func strPtr(s string) *string {
	return &s
}

// mockRedisClient creates a minimal mock for Redis operations needed in tests
type mockRedisClient struct{}

func (m *mockRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	cmd.SetErr(nil)
	return cmd
}
