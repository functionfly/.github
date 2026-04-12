// Package frg implements the Function Registry + Live Runtime Graph system.
// It provides versioned, composable function graphs with streaming execution,
// DRE (Deterministic Replay Execution) support, and AI-powered optimization.
package frg

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/agent/graph"
	"github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
	frgstate "github.com/functionfly/functionfly/internal/frg/state"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	statestorage "github.com/functionfly/functionfly/internal/storage/state"
	"github.com/google/uuid"
)

// ExecutionEngine orchestrates graph execution with streaming and event-driven support
type ExecutionEngine struct {
	// Data layer
	repo         *Repository
	registryRepo *registry.RegistryRepository
	functionRepo *storage.FunctionRepository

	// State management for graph instances
	stateRepo     *statestorage.StateRepository
	stateManagers map[uuid.UUID]*frgstate.GraphStateManager
	stateMu       sync.RWMutex

	// Execution backends
	sandboxExecutor *execution.SandboxExecutor
	sandboxClient   interface{} // execution.SandboxClient for daemon mode

	// Graph analysis (reuse existing agent/graph)
	graphService *graph.Service

	// Live instances
	instances   map[uuid.UUID]*GraphRuntime
	instancesMu sync.RWMutex

	// Event handling
	eventBus EventStream

	// Background tasks
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	// DRE (Deterministic Replay Execution) configuration
	dreNodeID      string
	dreRegion      string
	dreNodeKey     ed25519.PrivateKey
	drePlatformKey ed25519.PrivateKey
}

// NewExecutionEngine creates a new graph execution engine
func NewExecutionEngine(
	repo *Repository,
	registryRepo *registry.RegistryRepository,
	functionRepo *storage.FunctionRepository,
	stateRepo *statestorage.StateRepository,
	graphService *graph.Service,
	eventBus EventStream,
	dreNodeID, dreRegion string,
	dreNodeKey, drePlatformKey ed25519.PrivateKey,
) (*ExecutionEngine, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize sandbox executor for function execution
	sandboxExecutor, err := execution.NewSandboxExecutor()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create sandbox executor: %w", err)
	}

	return &ExecutionEngine{
		repo:            repo,
		registryRepo:    registryRepo,
		functionRepo:    functionRepo,
		stateRepo:       stateRepo,
		stateManagers:   make(map[uuid.UUID]*frgstate.GraphStateManager),
		graphService:    graphService,
		eventBus:        eventBus,
		sandboxExecutor: sandboxExecutor,
		instances:       make(map[uuid.UUID]*GraphRuntime),
		ctx:             ctx,
		cancel:          cancel,
		dreNodeID:       dreNodeID,
		dreRegion:       dreRegion,
		dreNodeKey:      dreNodeKey,
		drePlatformKey:  drePlatformKey,
	}, nil
}

// GetStateManager gets or creates a state manager for a graph instance
func (e *ExecutionEngine) GetStateManager(ctx context.Context, instanceID uuid.UUID, tenantID uuid.UUID, graphName string) (*frgstate.GraphStateManager, error) {
	e.stateMu.RLock()
	if mgr, ok := e.stateManagers[instanceID]; ok {
		e.stateMu.RUnlock()
		return mgr, nil
	}
	e.stateMu.RUnlock()

	// Create new state manager
	mgr, err := frgstate.NewGraphStateManager(ctx, e.stateRepo, tenantID, instanceID, graphName)
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	e.stateMu.Lock()
	e.stateManagers[instanceID] = mgr
	e.stateMu.Unlock()

	return mgr, nil
}

// Close shuts down the execution engine
func (e *ExecutionEngine) Close() error {
	e.cancel()

	// Stop all running instances
	e.instancesMu.Lock()
	for _, runtime := range e.instances {
		close(runtime.InputChannel)
		close(runtime.OutputChannel)
	}
	e.instances = make(map[uuid.UUID]*GraphRuntime)
	e.instancesMu.Unlock()

	// Cleanup state managers
	e.stateMu.Lock()
	for id, mgr := range e.stateManagers {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		mgr.CleanupState(ctx)
		cancel()
		delete(e.stateManagers, id)
	}
	e.stateMu.Unlock()

	// Wait for background tasks
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Close sandbox executor
		if e.sandboxExecutor != nil {
			e.sandboxExecutor.Close()
		}
		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for engine shutdown")
	}
}

// GetInstanceStatus returns the current status of an instance
func (e *ExecutionEngine) GetInstanceStatus(instanceID uuid.UUID) (*GraphInstance, error) {
	e.instancesMu.RLock()
	runtime, exists := e.instances[instanceID]
	e.instancesMu.RUnlock()

	if !exists {
		// Check database
		return e.repo.GetInstanceByID(context.Background(), instanceID)
	}

	return runtime.Instance, nil
}

// StopInstance stops a running instance
func (e *ExecutionEngine) StopInstance(instanceID uuid.UUID) error {
	e.instancesMu.Lock()
	runtime, exists := e.instances[instanceID]
	if exists {
		delete(e.instances, instanceID)
	}
	e.instancesMu.Unlock()

	if !exists {
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	// Signal shutdown
	close(runtime.InputChannel)

	// Cleanup state
	e.stateMu.Lock()
	if mgr, ok := e.stateManagers[instanceID]; ok {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		mgr.CleanupState(ctx)
		cancel()
		delete(e.stateManagers, instanceID)
	}
	e.stateMu.Unlock()

	// Update status
	return e.repo.UpdateInstanceStatus(context.Background(), instanceID, InstanceStatusCompleted)
}

// Helper function
func timePtr(t time.Time) *time.Time {
	return &t
}
