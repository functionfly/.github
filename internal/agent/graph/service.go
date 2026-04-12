package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Node represents a function in the composition graph
type Node struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID   string    `json:"function_id" gorm:"uniqueIndex;not null"`
	Name         string    `json:"name" gorm:"not null"`
	Category     string    `json:"category" gorm:"index"`
	InputSchema  string    `json:"input_schema" gorm:"type:jsonb"`
	OutputSchema string    `json:"output_schema" gorm:"type:jsonb"`
	Metadata     string    `json:"metadata" gorm:"type:jsonb"`
	IsActive     bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (Node) TableName() string { return "graph_nodes" }

// Edge represents a directed edge (dependency) in the composition graph
type Edge struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SourceNodeID uuid.UUID `json:"source_node_id" gorm:"type:uuid;not null;index"`
	TargetNodeID uuid.UUID `json:"target_node_id" gorm:"type:uuid;not null;index"`
	SourceNode   *Node     `json:"source_node,omitempty" gorm:"foreignKey:SourceNodeID;references:ID"`
	TargetNode   *Node     `json:"target_node,omitempty" gorm:"foreignKey:TargetNodeID;references:ID"`
	EdgeType     string    `json:"edge_type" gorm:"not null;default:'dataflow'"` // dataflow | trigger | dependency
	Mapping      string    `json:"mapping" gorm:"type:jsonb"`                    // How output maps to input
	Metadata     string    `json:"metadata" gorm:"type:jsonb"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the GORM table name
func (Edge) TableName() string { return "graph_edges" }

// Execution represents a graph execution instance
type Execution struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	GraphID      uuid.UUID  `json:"graph_id" gorm:"type:uuid;not null;index"`
	Graph        *Graph     `json:"graph,omitempty" gorm:"foreignKey:GraphID;references:ID"`
	Status       string     `json:"status" gorm:"not null;default:'pending'"` // pending | running | completed | failed
	InputData    string     `json:"input_data" gorm:"type:jsonb"`
	OutputData   string     `json:"output_data" gorm:"type:jsonb"`
	ErrorMessage *string    `json:"error_message" gorm:"type:text"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the GORM table name
func (Execution) TableName() string { return "graph_executions" }

// Graph represents a composition of functions
type Graph struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	OwnerID     string    `json:"owner_id" gorm:"not null;index"`
	IsPublic    bool      `json:"is_public" gorm:"not null;default:false"`
	Metadata    string    `json:"metadata" gorm:"type:jsonb"`
	IsActive    bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (Graph) TableName() string { return "graphs" }

// Service handles function composition graph operations
type Service struct {
	db       *gorm.DB
	executor *ExecutionService
}

// NewService creates a new graph service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db:       db,
		executor: NewExecutionService(db),
	}
}

// AutoMigrate runs database migrations for graph components
func (s *Service) AutoMigrate(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&Graph{}, &Node{}, &Edge{}, &Execution{})
}

// CreateGraph creates a new function composition graph
func (s *Service) CreateGraph(ctx context.Context, ownerID string, name string, description string) (*Graph, error) {
	graph := &Graph{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		IsPublic:    false,
		Metadata:    "{}",
		IsActive:    true,
	}
	if err := s.db.WithContext(ctx).Create(graph).Error; err != nil {
		return nil, fmt.Errorf("failed to create graph: %w", err)
	}
	return graph, nil
}

// AddNode adds a function to the graph
func (s *Service) AddNode(ctx context.Context, graphID uuid.UUID, functionID string, name string, inputSchema, outputSchema map[string]any) (*Node, error) {
	inputJSON, _ := json.Marshal(inputSchema)
	outputJSON, _ := json.Marshal(outputSchema)

	node := &Node{
		ID:           uuid.New(),
		FunctionID:   functionID,
		Name:         name,
		InputSchema:  string(inputJSON),
		OutputSchema: string(outputJSON),
		Metadata:     "{}",
		IsActive:     true,
	}
	if err := s.db.WithContext(ctx).Create(node).Error; err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	// Link node to graph via an internal edge if needed
	return node, nil
}

// AddEdge creates a directed edge between two nodes
func (s *Service) AddEdge(ctx context.Context, graphID uuid.UUID, sourceNodeID uuid.UUID, targetNodeID uuid.UUID, edgeType string, mapping map[string]string) error {
	mappingJSON, _ := json.Marshal(mapping)

	edge := &Edge{
		ID:           uuid.New(),
		SourceNodeID: sourceNodeID,
		TargetNodeID: targetNodeID,
		EdgeType:     edgeType,
		Mapping:      string(mappingJSON),
		Metadata:     "{}",
	}
	if err := s.db.WithContext(ctx).Create(edge).Error; err != nil {
		return fmt.Errorf("failed to create edge: %w", err)
	}
	return nil
}

// DetectCycle detects if adding an edge would create a cycle
func (s *Service) DetectCycle(ctx context.Context, sourceNodeID uuid.UUID, targetNodeID uuid.UUID) (bool, error) {
	// BFS from target to see if we can reach source
	visited := make(map[uuid.UUID]bool)
	queue := []uuid.UUID{targetNodeID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == sourceNodeID {
			return true, nil // Cycle detected
		}

		if visited[current] {
			continue
		}
		visited[current] = true

		// Find all nodes that this node points to
		var edges []Edge
		if err := s.db.WithContext(ctx).Where("source_node_id = ?", current).Find(&edges).Error; err != nil {
			return false, err
		}
		for _, edge := range edges {
			queue = append(queue, edge.TargetNodeID)
		}
	}

	return false, nil
}

// GetExecutionOrder returns nodes in topological order for execution
func (s *Service) GetExecutionOrder(ctx context.Context, graphID uuid.UUID) ([]Node, error) {
	var nodes []Node
	if err := s.db.WithContext(ctx).Find(&nodes).Error; err != nil {
		return nil, err
	}

	// Build adjacency list and in-degree count
	inDegree := make(map[uuid.UUID]int)
	adjList := make(map[uuid.UUID][]uuid.UUID)

	for _, node := range nodes {
		inDegree[node.ID] = 0
	}

	var edges []Edge
	if err := s.db.WithContext(ctx).Find(&edges).Error; err != nil {
		return nil, err
	}

	for _, edge := range edges {
		adjList[edge.SourceNodeID] = append(adjList[edge.SourceNodeID], edge.TargetNodeID)
		inDegree[edge.TargetNodeID]++
	}

	// Kahn's algorithm for topological sort
	var queue []uuid.UUID
	for _, node := range nodes {
		if inDegree[node.ID] == 0 {
			queue = append(queue, node.ID)
		}
	}

	var result []Node
	nodeMap := make(map[uuid.UUID]Node)
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, nodeMap[current])

		for _, neighbor := range adjList[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycle (if not all nodes are in result)
	if len(result) != len(nodes) {
		return nil, fmt.Errorf("graph contains a cycle")
	}

	return result, nil
}

// ExecuteGraph executes the function composition graph
func (s *Service) ExecuteGraph(ctx context.Context, graphID uuid.UUID, input map[string]any) (*Execution, error) {
	execution := &Execution{
		ID:        uuid.New(),
		GraphID:   graphID,
		Status:    "pending",
		InputData: mustMarshal(input),
	}
	if err := s.db.WithContext(ctx).Create(execution).Error; err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	// Execute asynchronously
	go s.executor.Run(ctx, execution.ID)

	return execution, nil
}

// GetExecution returns an execution by ID
func (s *Service) GetExecution(ctx context.Context, executionID uuid.UUID) (*Execution, error) {
	var execution Execution
	if err := s.db.WithContext(ctx).First(&execution, executionID).Error; err != nil {
		return nil, err
	}
	return &execution, nil
}

// Executor interface for function execution
type Executor interface {
	Execute(ctx context.Context, tenantID, functionID uuid.UUID, runtimeType string, input json.RawMessage) (*execution.ExecuteResult, error)
}

// ExecutionService handles running graph executions
type ExecutionService struct {
	db       *gorm.DB
	graph    *Service
	executor Executor
}

// NewExecutionService creates a new execution service
func NewExecutionService(db *gorm.DB) *ExecutionService {
	return &ExecutionService{
		db:    db,
		graph: NewService(db),
	}
}

// SetExecutor sets the function executor for graph node execution
func (s *ExecutionService) SetExecutor(executor Executor) {
	s.executor = executor
}

// Run executes a graph
func (s *ExecutionService) Run(ctx context.Context, executionID uuid.UUID) {
	var execution Execution
	if err := s.db.First(&execution, executionID).Error; err != nil {
		return
	}

	now := time.Now()
	execution.StartedAt = &now
	execution.Status = "running"
	s.db.Save(&execution)

	// Get graph nodes in execution order
	nodes, err := s.getExecutionOrder(execution.GraphID)
	if err != nil {
		msg := err.Error()
		execution.Status = "failed"
		execution.ErrorMessage = &msg
		completed := time.Now()
		execution.CompletedAt = &completed
		s.db.Save(&execution)
		return
	}

	// Execute nodes in order, passing output to next input
	currentData := make(map[string]any)
	json.Unmarshal([]byte(execution.InputData), &currentData)

	for _, node := range nodes {
		output, err := s.executeNode(ctx, &node, &execution, currentData)
		if err != nil {
			msg := err.Error()
			execution.Status = "failed"
			execution.ErrorMessage = &msg
			completed := time.Now()
			execution.CompletedAt = &completed
			s.db.Save(&execution)
			return
		}
		currentData = output
	}

	// Success
	execution.OutputData = mustMarshal(currentData)
	execution.Status = "completed"
	completed := time.Now()
	execution.CompletedAt = &completed
	s.db.Save(&execution)
}

func (s *ExecutionService) getExecutionOrder(graphID uuid.UUID) ([]Node, error) {
	return s.graph.GetExecutionOrder(context.Background(), graphID)
}

func (s *ExecutionService) executeNode(ctx context.Context, node *Node, execution *Execution, input map[string]any) (map[string]any, error) {
	if s.executor == nil {
		return nil, fmt.Errorf("no executor configured for graph execution")
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	// Parse function ID from node
	funcID, err := uuid.Parse(node.FunctionID)
	if err != nil {
		return nil, fmt.Errorf("invalid function ID %s: %w", node.FunctionID, err)
	}

	// Load graph to get owner (tenant) ID
	var graph Graph
	if err := s.db.WithContext(ctx).First(&graph, execution.GraphID).Error; err != nil {
		return nil, fmt.Errorf("failed to load graph %s: %w", execution.GraphID, err)
	}
	tenantID, err := uuid.Parse(graph.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("invalid graph owner ID %s: %w", graph.OwnerID, err)
	}

	result, err := s.executor.Execute(ctx, tenantID, funcID, "wasm", inputJSON)
	if err != nil {
		return nil, fmt.Errorf("function execution failed: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("function returned error: %s", result.ErrorMessage)
	}

	var output map[string]any
	if err := json.Unmarshal(result.Output, &output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal output: %w", err)
	}

	return output, nil
}

func mustMarshal(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
