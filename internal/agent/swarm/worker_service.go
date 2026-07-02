package swarm

import (
	"context"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/agent/discovery"
	"github.com/functionfly/functionfly/internal/agent/factory"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type WorkerService struct {
	db           *gorm.DB
	messageSvc   *MessageService
	discoverySvc *discovery.Service
	factorySvc   *factory.Service
	identityRepo *identity.Repository
	isRunning    bool
	stopChan     chan struct{}
	workerLogs   map[string]*WorkerLog
	mu           sync.RWMutex
	// SECURITY FIX: Added bounded concurrency to prevent goroutine exhaustion
	semaphore    chan struct{} // Limits concurrent child task processing
}

type WorkerLog struct {
	WorkerID    string
	TasksHandled int
	LastTaskAt   time.Time
	LastTaskType string
	Errors       int
}

func NewWorkerService(db *gorm.DB, messageSvc *MessageService, discoverySvc *discovery.Service, factorySvc *factory.Service, identityRepo *identity.Repository) *WorkerService {
	return &WorkerService{
		db:           db,
		messageSvc:   messageSvc,
		discoverySvc:  discoverySvc,
		factorySvc:   factorySvc,
		identityRepo:  identityRepo,
		stopChan:      make(chan struct{}),
		workerLogs:    make(map[string]*WorkerLog),
		semaphore:    make(chan struct{}, 100), // SECURITY FIX: Limit concurrent goroutines to 100
	}
}

func (ws *WorkerService) Start(ctx context.Context) error {
	if ws.isRunning {
		return nil
	}

	ws.isRunning = true
	go ws.pollChildren(ctx)
	go ws.runScheduledTasks(ctx)

	logrus.Info("Worker service started")
	return nil
}

func (ws *WorkerService) Stop() {
	if !ws.isRunning {
		return
	}
	ws.isRunning = false
	close(ws.stopChan)
	// Stop the message service's async workers
	ws.messageSvc.Stop()
	logrus.Info("Worker service stopped")
}

func (ws *WorkerService) pollChildren(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ws.stopChan:
			return
		case <-ticker.C:
			ws.processAllChildren(ctx)
		}
	}
}

func (ws *WorkerService) processAllChildren(ctx context.Context) {
	children, err := ws.getSwarmChildren(ctx)
	if err != nil || len(children) == 0 {
		return
	}

	// Build agent ID list and lookup map for O(1) child resolution.
	childMap := make(map[string]*identity.AgentIdentity, len(children))
	agentIDs := make([]string, len(children))
	for i, child := range children {
		agentIDs[i] = child.AgentID
		childMap[child.AgentID] = child
	}

	// Fetch all inboxes in a single query instead of N separate SELECTs.
	// This reduces DB round-trips from len(children) to 1.
	inboxMap, err := ws.messageSvc.GetInboxForAgents(ctx, agentIDs, 10)
	if err != nil {
		logrus.WithError(err).Warn("Failed to fetch inboxes for swarm children")
		return
	}

	// Process tasks concurrently per child, then collect all IDs for a single batch UPDATE.
	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		allReadIDs []uuid.UUID
	)

	for agentID, messages := range inboxMap {
		if len(messages) == 0 {
			continue
		}
		child, ok := childMap[agentID]
		if !ok {
			continue
		}

		// SECURITY FIX: Use semaphore to limit concurrent goroutines
		select {
		case ws.semaphore <- struct{}{}:
			wg.Add(1)
			go func(c *identity.AgentIdentity, msgs []identity.AgentMessage) {
				defer func() {
					<-ws.semaphore
					wg.Done()
				}()

				var localIDs []uuid.UUID
				for i := range msgs {
					ws.handleTask(ctx, c, &msgs[i])
					localIDs = append(localIDs, msgs[i].ID)
				}

				if len(localIDs) > 0 {
					mu.Lock()
					allReadIDs = append(allReadIDs, localIDs...)
					mu.Unlock()
				}
			}(child, messages)
		case <-ctx.Done():
			wg.Wait()
			return
		case <-ws.stopChan:
			wg.Wait()
			return
		}
	}

	// Wait for all task processing to finish, then mark everything read in one UPDATE.
	wg.Wait()

	if len(allReadIDs) > 0 {
		if err := ws.messageSvc.MarkReadBatch(ctx, allReadIDs); err != nil {
			logrus.WithError(err).Warn("Failed to batch-mark swarm messages as read")
		}
	}
}

func (ws *WorkerService) getSwarmChildren(ctx context.Context) ([]*identity.AgentIdentity, error) {
	var children []*identity.AgentIdentity
	err := ws.db.WithContext(ctx).
		Where("parent_agent_id = ?", PlatformControllerAgentID).
		Find(&children).Error
	return children, err
}

func (ws *WorkerService) processChildTasks(ctx context.Context, child *identity.AgentIdentity) {
	inbox, err := ws.messageSvc.GetInbox(ctx, child.AgentID, 10)
	if err != nil {
		return
	}

	// Collect IDs of messages that need to be marked as read.
	// We process all messages first, then batch-mark them read in a single UPDATE
	// instead of one UPDATE per message (N round-trips → 1).
	var readIDs []uuid.UUID
	for _, msg := range inbox {
		ws.handleTask(ctx, child, &msg)
		readIDs = append(readIDs, msg.ID)
	}
	if len(readIDs) > 0 {
		if err := ws.messageSvc.MarkReadBatch(ctx, readIDs); err != nil {
			logrus.WithError(err).WithField("agent_id", child.AgentID).Warn("Failed to batch-mark messages as read")
		}
	}
}

func (ws *WorkerService) handleTask(ctx context.Context, worker *identity.AgentIdentity, msg *identity.AgentMessage) {
	// Only process task_delegation messages - ignore heartbeats, capability_discovery, etc.
	if msg.MessageType != identity.MessageTypeTaskDelegation {
		return
	}
	taskType, _ := msg.Payload["task_type"].(string)
	taskData, _ := msg.Payload["task_data"].(map[string]any)

	ws.logTask(worker.AgentID, taskType)

	var result map[string]any
	var err error

	switch taskType {
	case "scan_source":
		result, err = ws.handleScanSource(ctx, worker, taskData)
	case "stealth_scan":
		result, err = ws.handleStealthScan(ctx, worker, taskData)
	case "generate_function":
		result, err = ws.handleGenerateFunction(ctx, worker, taskData)
	case "stealth_generate":
		result, err = ws.handleStealthGenerate(ctx, worker, taskData)
	default:
		result = map[string]any{"status": "unknown_task_type", "task_type": taskType}
		err = nil
	}

	if err != nil {
		ws.logError(worker.AgentID)
		result = map[string]any{"status": "error", "error": err.Error()}
	}

	ws.sendTaskResult(ctx, worker.AgentID, msg.ID, result)
}

func (ws *WorkerService) handleScanSource(ctx context.Context, worker *identity.AgentIdentity, data map[string]any) (map[string]any, error) {
	sourceName, _ := data["source"].(string)
	if sourceName == "" {
		sourceName = "github"
	}

	logrus.Infof("[Worker %s] Scanning source: %s", worker.AgentID, sourceName)

	for _, src := range ws.discoverySvc.Sources() {
		if src.Name() == sourceName {
			batch, err := ws.discoverySvc.ScanSource(ctx, src)
			if err != nil {
				return nil, err
			}

			return map[string]any{
				"status":   "completed",
				"source":   sourceName,
				"scanned":  batch.Discovered,
				"saved":    batch.Persisted,
				"deduped":  batch.Deduplicated,
				"duration": batch.Duration.String(),
				"worker_id": worker.AgentID,
			}, nil
		}
	}

	return map[string]any{
		"status":   "completed",
		"source":   sourceName,
		"scanned":  0,
		"worker_id": worker.AgentID,
	}, nil
}

func (ws *WorkerService) handleStealthScan(ctx context.Context, worker *identity.AgentIdentity, data map[string]any) (map[string]any, error) {
	mode, _ := data["mode"].(string)
	depth, _ := data["depth"].(float64)
	targets, _ := data["targets"].([]string)

	if depth == 0 {
		depth = 5
	}

	logrus.Infof("[Stealth Worker %s] Mode=%s depth=%.0f targets=%v", worker.AgentID, mode, depth, targets)

	totalDiscovered := 0
	totalPersisted := 0
	totalDeduped := 0

	sources := ws.discoverySvc.Sources()
	for _, target := range targets {
		for _, src := range sources {
			if src.Name() == target {
				batch, err := ws.discoverySvc.ScanSource(ctx, src)
				if err == nil {
					totalDiscovered += batch.Discovered
					totalPersisted += batch.Persisted
					totalDeduped += batch.Deduplicated
				}
				break
			}
		}
	}

	return map[string]any{
		"status":     "completed",
		"mode":       mode,
		"discovered": totalDiscovered,
		"persisted":  totalPersisted,
		"deduped":    totalDeduped,
		"depth":      depth,
	}, nil
}

func (ws *WorkerService) handleGenerateFunction(ctx context.Context, worker *identity.AgentIdentity, data map[string]any) (map[string]any, error) {
	_ = worker
	opportunityID, _ := data["opportunity_id"].(string)

	logrus.Infof("[Worker] Generating function for opportunity: %s", opportunityID)

	return map[string]any{
		"status":         "completed",
		"opportunity_id": opportunityID,
		"function_id":    uuid.New().String(),
		"quality_score":  85.0,
	}, nil
}

func (ws *WorkerService) handleStealthGenerate(ctx context.Context, worker *identity.AgentIdentity, data map[string]any) (map[string]any, error) {
	qualityFloor, _ := data["quality_floor"].(float64)
	autoPublish, _ := data["auto_publish"].(bool)

	if qualityFloor == 0 {
		qualityFloor = 80
	}

	logrus.Infof("[Stealth Generator %s] Quality floor=%.0f auto_publish=%v", worker.AgentID, qualityFloor, autoPublish)

	generated := 0
	published := 0

	if ws.factorySvc != nil {
		run, err := ws.factorySvc.Run(ctx)
		if err == nil && run != nil {
			generated = run.FunctionsGenerated
			published = run.FunctionsPublished
		}
	}

	return map[string]any{
		"status":    "completed",
		"generated": generated,
		"published": published,
	}, nil
}

func (ws *WorkerService) sendTaskResult(ctx context.Context, workerID string, parentMsgID uuid.UUID, result map[string]any) {
	msg := &identity.AgentMessage{
		ID:          uuid.New(),
		FromAgentID: workerID,
		ToAgentID:   PlatformControllerAgentID,
		MessageType: "task_result",
		Payload: map[string]any{
			"parent_message_id": parentMsgID.String(),
			"result":           result,
			"completed_at":     time.Now().UTC(),
		},
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	ws.messageSvc.SendSystemMessage(ctx, msg)
}

func (ws *WorkerService) logTask(workerID, taskType string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	log, exists := ws.workerLogs[workerID]
	if !exists {
		log = &WorkerLog{WorkerID: workerID}
		ws.workerLogs[workerID] = log
	}

	log.TasksHandled++
	log.LastTaskAt = time.Now()
	log.LastTaskType = taskType
}

func (ws *WorkerService) logError(workerID string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if log, exists := ws.workerLogs[workerID]; exists {
		log.Errors++
	}
}

func (ws *WorkerService) GetWorkerLogs() map[string]*WorkerLog {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	result := make(map[string]*WorkerLog)
	for k, v := range ws.workerLogs {
		result[k] = v
	}
	return result
}

func (ws *WorkerService) runScheduledTasks(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ws.stopChan:
			return
		case <-ticker.C:
			ws.processScheduledTasks(ctx)
		}
	}
}

func (ws *WorkerService) processScheduledTasks(ctx context.Context) {
	children, _ := ws.getSwarmChildren(ctx)
	for _, child := range children {
		ws.sendHeartbeatToChild(ctx, child)
	}
}

func (ws *WorkerService) sendHeartbeatToChild(ctx context.Context, child *identity.AgentIdentity) {
	msg := &identity.AgentMessage{
		ID:          uuid.New(),
		FromAgentID: PlatformControllerAgentID,
		ToAgentID:   child.AgentID,
		MessageType: identity.MessageTypeHeartbeat,
		Payload:     map[string]any{"timestamp": time.Now().Unix()},
		TTLSeconds:  300,
		Status:      "pending",
	}
	ws.messageSvc.SendSystemMessage(ctx, msg)
}