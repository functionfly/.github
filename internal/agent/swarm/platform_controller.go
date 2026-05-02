package swarm

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/discovery"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	PlatformControllerAgentID = "functionfly/platform-controller"
	ScannerAgentPrefix        = "functionfly/scanner-"
	GeneratorWorkerPrefix      = "functionfly/generator-"
	QAWorkerPrefix            = "functionfly/qa-"
	PublisherAgentID          = "functionfly/publisher"
)

type PlatformController struct {
	db              *gorm.DB
	swarmSvc        *Service
	identityRepo    *identity.Repository
	messageService  *MessageService
	discoverySvc    *discovery.Service
	controllerAgent *identity.AgentIdentity
	initialized     bool
}

func NewPlatformController(db *gorm.DB, swarmSvc *Service, identityRepo *identity.Repository, messageSvc *MessageService, discoverySvc *discovery.Service) *PlatformController {
	return &PlatformController{
		db:            db,
		swarmSvc:      swarmSvc,
		identityRepo:  identityRepo,
		messageService: messageSvc,
		discoverySvc:  discoverySvc,
	}
}

func (pc *PlatformController) Initialize(ctx context.Context) error {
	if pc.initialized {
		return nil
	}

	controller, err := pc.identityRepo.GetAgent(ctx, PlatformControllerAgentID)
	if err != nil {
		controller = pc.createPlatformAgent(ctx)
	}

	pc.controllerAgent = controller
	pc.initialized = true

	if err := pc.syncChildAgents(ctx); err != nil {
		return fmt.Errorf("failed to sync child agents: %w", err)
	}

	return nil
}

func (pc *PlatformController) createPlatformAgent(ctx context.Context) *identity.AgentIdentity {
	agent := &identity.AgentIdentity{
		ID:                uuid.New(),
		TenantID:          uuid.Nil,
		AgentID:           PlatformControllerAgentID,
		Name:              "FunctionFly Platform Controller",
		Description:       "Orchestrates the autonomous function generation swarm",
		PlanTier:          "agent_enterprise",
		Status:            identity.AgentStatusActive,
		SwarmRole:         identity.SwarmRoleManager,
		MaxChildAgents:    20,
		Capabilities:      identity.JSONBMap{"orchestration": true, "swarm_control": true, "metrics_collection": true},
		AutonomousEnabled: true,
		EvolutionEnabled:  true,
	}

	if err := pc.db.WithContext(ctx).Create(agent).Error; err == nil {
		return agent
	}

	existing, _ := pc.identityRepo.GetAgent(ctx, PlatformControllerAgentID)
	if existing != nil {
		return existing
	}

	return agent
}

func (pc *PlatformController) syncChildAgents(ctx context.Context) error {
	children, err := pc.swarmSvc.GetChildren(ctx, PlatformControllerAgentID)
	if err != nil {
		return err
	}

	requiredRoles := map[string]int{
		"github-scanner":     1,
		"stackoverflow-scanner": 1,
		"reddit-scanner":    1,
		"generator-worker":  3,
		"qa-worker":         1,
	}

	currentChildren := make(map[string]int)
	for _, child := range children {
		role := pc.extractRoleFromCapabilities(child.Capabilities)
		currentChildren[role]++
	}

	for role, desiredCount := range requiredRoles {
		current := currentChildren[role]
		for current < desiredCount {
			if err := pc.spawnWorker(ctx, role); err != nil {
				return fmt.Errorf("failed to spawn %s: %w", role, err)
			}
			current++
		}
	}

	return nil
}

func (pc *PlatformController) extractRoleFromCapabilities(capabilities identity.JSONBMap) string {
	if role, ok := capabilities["role"].(string); ok {
		return role
	}
	return "unknown"
}

func (pc *PlatformController) spawnWorker(ctx context.Context, role string) error {
	var childAgentID, childName, description string
	var capabilities map[string]any

	switch role {
	case "github-scanner":
		childAgentID = ScannerAgentPrefix + "github-" + uuid.New().String()[:8]
		childName = "GitHub Scanner Worker"
		description = "Scans GitHub issues for function opportunities"
		capabilities = map[string]any{"role": "github-scanner", "source": "github", "swarm_enabled": true}
	case "stackoverflow-scanner":
		childAgentID = ScannerAgentPrefix + "stackoverflow-" + uuid.New().String()[:8]
		childName = "StackOverflow Scanner Worker"
		description = "Scans StackOverflow for function opportunities"
		capabilities = map[string]any{"role": "stackoverflow-scanner", "source": "stackoverflow", "swarm_enabled": true}
	case "reddit-scanner":
		childAgentID = ScannerAgentPrefix + "reddit-" + uuid.New().String()[:8]
		childName = "Reddit Scanner Worker"
		description = "Scans Reddit for function opportunities"
		capabilities = map[string]any{"role": "reddit-scanner", "source": "reddit", "swarm_enabled": true}
	case "generator-worker":
		childAgentID = GeneratorWorkerPrefix + uuid.New().String()[:8]
		childName = "Code Generation Worker"
		description = "Generates function code from opportunities"
		capabilities = map[string]any{"role": "generator-worker", "swarm_enabled": true, "autonomous": true}
	case "qa-worker":
		childAgentID = QAWorkerPrefix + uuid.New().String()[:8]
		childName = "Quality Assurance Worker"
		description = "Reviews and validates generated functions"
		capabilities = map[string]any{"role": "qa-worker", "swarm_enabled": true}
	}

	req := &SpawnChildRequest{
		ParentAgentID:    PlatformControllerAgentID,
		ChildAgentID:     childAgentID,
		ChildName:        childName,
		ChildDescription: description,
		SwarmRole:        identity.SwarmRoleWorker,
		MaxChildAgents:   0,
		Capabilities:     capabilities,
		InitialBudgetUSD: 0,
		PolicyConfig: &PolicyConfig{
			MaxExecutionDepth:   3,
			MaxRecursionDepth:   2,
			MaxWallTimeMs:       30000,
			MaxMemoryGrowthMB:   512,
			ForbiddenFunctions:  []string{"rm", "format", "drop"},
			DeterministicOnly:   true,
			AllowedCapabilities: []string{},
		},
	}

	_, _, err := pc.swarmSvc.SpawnChild(ctx, req)
	return err
}

func (pc *PlatformController) GetStatus(ctx context.Context) (*PlatformStatus, error) {
	children, err := pc.swarmSvc.GetChildren(ctx, PlatformControllerAgentID)
	if err != nil {
		return nil, err
	}

	childrenCount, _ := pc.swarmSvc.CountChildren(ctx, PlatformControllerAgentID)

	status := &PlatformStatus{
		ControllerAgentID: PlatformControllerAgentID,
		IsInitialized:     pc.initialized,
		TotalChildren:    int(childrenCount),
		Children:          make([]ChildAgentStatus, 0, len(children)),
		LastActivityAt:   time.Now(),
	}

	for _, child := range children {
		capabilities := make(map[string]any)
		if child.Capabilities != nil {
			capabilities = child.Capabilities
		}

		childStatus := ChildAgentStatus{
			AgentID:      child.AgentID,
			Name:         child.Name,
			SwarmRole:    child.SwarmRole,
			Status:       child.Status,
			Capabilities: capabilities,
			CreatedAt:    child.CreatedAt,
		}

		if inbox, err := pc.messageService.GetInbox(ctx, child.AgentID, 10); err == nil {
			childStatus.PendingMessages = len(inbox)
			for _, msg := range inbox {
				if msg.Status == "pending" {
					childStatus.PendingMessages++
				}
			}
		}

		if outbox, err := pc.messageService.GetOutbox(ctx, child.AgentID, 10); err == nil {
			childStatus.RecentActivity = len(outbox)
		}

		status.Children = append(status.Children, childStatus)
	}

	var totalRuns, totalVersions, totalPublished int64
	pc.db.WithContext(ctx).Model(&struct{}{}).Table("factory_runs").Count(&totalRuns)
	pc.db.WithContext(ctx).Model(&struct{}{}).Table("factory_versions").Count(&totalVersions)

	var opps []discovery.Opportunity
	pc.db.WithContext(ctx).Where("status = ?", discovery.OpportunityStatusQualified).Find(&opps)
	status.QualifiedOpportunities = len(opps)

	pc.db.WithContext(ctx).Where("status = ?", discovery.OpportunityStatusPublished).Find(&opps)
	status.PublishedOpportunities = len(opps)

	status.Metrics = SwarmMetrics{
		TotalRuns:         totalRuns,
		TotalVersions:     totalVersions,
		TotalPublished:   totalPublished,
		ActiveChildren:    int(childrenCount),
		PendingTasks:      status.TotalPendingMessages(),
	}

	return status, nil
}

func (pc *PlatformController) DispatchTask(ctx context.Context, targetAgentID string, taskType string, taskData map[string]any) error {
	sessionID := uuid.New().String()
	msg := &identity.AgentMessage{
		ID:          uuid.New(),
		FromAgentID: PlatformControllerAgentID,
		ToAgentID:   targetAgentID,
		MessageType: identity.MessageTypeTaskDelegation,
		Payload: map[string]any{
			"task_type":    taskType,
			"task_data":    taskData,
			"dispatched_at": time.Now().UTC(),
		},
		SessionID:  &sessionID,
		TTLSeconds: 3600,
		Status:     "pending",
	}
	return pc.messageService.SendSystemMessage(ctx, msg)
}

func (pc *PlatformController) TriggerDiscoveryScan(ctx context.Context) error {
	for _, child := range pc.getChildrenByRole(ctx, "github-scanner") {
		if err := pc.DispatchTask(ctx, child.AgentID, "scan_source", map[string]any{"source": "github"}); err != nil {
			return err
		}
	}
	for _, child := range pc.getChildrenByRole(ctx, "stackoverflow-scanner") {
		if err := pc.DispatchTask(ctx, child.AgentID, "scan_source", map[string]any{"source": "stackoverflow"}); err != nil {
			return err
		}
	}
	for _, child := range pc.getChildrenByRole(ctx, "reddit-scanner") {
		if err := pc.DispatchTask(ctx, child.AgentID, "scan_source", map[string]any{"source": "reddit"}); err != nil {
			return err
		}
	}
	return nil
}

func (pc *PlatformController) getChildrenByRole(ctx context.Context, role string) []*identity.AgentIdentity {
	children, _ := pc.swarmSvc.GetChildren(ctx, PlatformControllerAgentID)
	var filtered []*identity.AgentIdentity
	for _, child := range children {
		if cap, ok := child.Capabilities["role"].(string); ok && cap == role {
			filtered = append(filtered, child)
		}
	}
	return filtered
}

func (pc *PlatformController) TriggerGeneration(ctx context.Context) error {
	for _, worker := range pc.getChildrenByRole(ctx, "generator-worker") {
		if err := pc.DispatchTask(ctx, worker.AgentID, "generate_function", nil); err != nil {
			return err
		}
	}
	return nil
}

func (pc *PlatformController) SendHeartbeat(ctx context.Context, agentID string) error {
	msg := &identity.AgentMessage{
		ID:          uuid.New(),
		FromAgentID: PlatformControllerAgentID,
		ToAgentID:   agentID,
		MessageType: identity.MessageTypeHeartbeat,
		Payload:     map[string]any{"timestamp": time.Now().Unix()},
		TTLSeconds:  300,
		Status:      "pending",
	}
	return pc.messageService.SendSystemMessage(ctx, msg)
}

type PlatformStatus struct {
	ControllerAgentID      string             `json:"controller_agent_id"`
	IsInitialized          bool               `json:"is_initialized"`
	TotalChildren          int                `json:"total_children"`
	Children               []ChildAgentStatus `json:"children"`
	LastActivityAt         time.Time          `json:"last_activity_at"`
	QualifiedOpportunities int                `json:"qualified_opportunities"`
	PublishedOpportunities  int                `json:"published_opportunities"`
	Metrics                SwarmMetrics       `json:"metrics"`
}

type ChildAgentStatus struct {
	AgentID          string         `json:"agent_id"`
	Name             string         `json:"name"`
	SwarmRole        string         `json:"swarm_role"`
	Status           string         `json:"status"`
	Capabilities     map[string]any `json:"capabilities"`
	PendingMessages  int            `json:"pending_messages"`
	RecentActivity   int            `json:"recent_activity"`
	CreatedAt        time.Time      `json:"created_at"`
}

type SwarmMetrics struct {
	TotalRuns         int64 `json:"total_runs"`
	TotalVersions     int64 `json:"total_versions"`
	TotalPublished    int64 `json:"total_published"`
	ActiveChildren    int   `json:"active_children"`
	PendingTasks      int   `json:"pending_tasks"`
}

func (s *PlatformStatus) TotalPendingMessages() int {
	total := 0
	for _, child := range s.Children {
		total += child.PendingMessages
	}
	return total
}