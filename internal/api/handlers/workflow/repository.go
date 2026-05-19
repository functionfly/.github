package workflow

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type WorkflowGraph struct {
	ID        string           `json:"id"`
	TenantID  string           `json:"tenant_id"`
	Name      string           `json:"name"`
	Nodes     []WorkflowNode   `json:"nodes"`
	Edges     []WorkflowEdge   `json:"edges"`
	Metadata  map[string]any   `json:"metadata,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type WorkflowNode struct {
	ID       string         `json:"id"`
	GraphID  string         `json:"graph_id"`
	Type     string         `json:"type"`
	Name     string         `json:"name"`
	Config   map[string]any `json:"config,omitempty"`
	Position NodePosition   `json:"position"`
	CreatedAt time.Time     `json:"created_at"`
}

type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type WorkflowEdge struct {
	ID        string `json:"id"`
	GraphID   string `json:"graph_id"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Condition string `json:"condition,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type WorkflowExecution struct {
	ID           string          `json:"id"`
	GraphID      string          `json:"graph_id"`
	TenantID     string          `json:"tenant_id"`
	Status       string          `json:"status"`
	StartedAt    time.Time       `json:"started_at"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	Result       map[string]any   `json:"result,omitempty"`
	Error        string           `json:"error,omitempty"`
	NodeResults  []NodeResult     `json:"node_results,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
}

type NodeResult struct {
	NodeID     string         `json:"node_id"`
	Status     string         `json:"status"`
	Output     map[string]any `json:"output,omitempty"`
	Error      string         `json:"error,omitempty"`
	DurationMs int64          `json:"duration_ms"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetGraphByTenant(ctx context.Context, tenantID string) (*WorkflowGraph, error) {
	query := `
		SELECT id, tenant_id, name, metadata, created_at, updated_at
		FROM studio_workflows
		WHERE tenant_id = $1
		ORDER BY updated_at DESC
		LIMIT 1`
	var g WorkflowGraph
	var metaRaw []byte
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&g.ID, &g.TenantID, &g.Name, &metaRaw, &g.CreatedAt, &g.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get graph: %w", err)
	}
	if len(metaRaw) > 0 {
		_ = json.Unmarshal(metaRaw, &g.Metadata)
	}

	nodes, err := r.ListNodes(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	g.Nodes = nodes

	edges, err := r.ListEdges(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	g.Edges = edges

	return &g, nil
}

func (r *Repository) CreateGraph(ctx context.Context, tenantID, name string, metadata map[string]any) (*WorkflowGraph, error) {
	id := uuid.New().String()
	metaRaw, _ := json.Marshal(metadata)
	now := time.Now()

	query := `
		INSERT INTO studio_workflows (id, tenant_id, name, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, id, tenantID, name, metaRaw, now).Scan(&now, &now)
	if err != nil {
		return nil, fmt.Errorf("create graph: %w", err)
	}

	return &WorkflowGraph{
		ID:        id,
		TenantID:  tenantID,
		Name:      name,
		Nodes:     []WorkflowNode{},
		Edges:     []WorkflowEdge{},
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (r *Repository) UpdateGraph(ctx context.Context, tenantID, graphID string, name string, metadata map[string]any) (*WorkflowGraph, error) {
	metaRaw, _ := json.Marshal(metadata)
	query := `
		UPDATE studio_workflows
		SET name = $1, metadata = $2, updated_at = NOW()
		WHERE id = $3 AND tenant_id = $4
		RETURNING id, tenant_id, name, metadata, created_at, updated_at`
	var g WorkflowGraph
	var metaRawOut []byte
	err := r.db.QueryRowContext(ctx, query, name, metaRaw, graphID, tenantID).Scan(
		&g.ID, &g.TenantID, &g.Name, &metaRawOut, &g.CreatedAt, &g.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update graph: %w", err)
	}
	if len(metaRawOut) > 0 {
		_ = json.Unmarshal(metaRawOut, &g.Metadata)
	}
	return &g, nil
}

func (r *Repository) ListNodes(ctx context.Context, graphID string) ([]WorkflowNode, error) {
	query := `
		SELECT id, graph_id, type, name, config, position_x, position_y, created_at
		FROM studio_workflow_nodes
		WHERE graph_id = $1
		ORDER BY created_at`
	rows, err := r.db.QueryContext(ctx, query, graphID)
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	defer rows.Close()

	var nodes []WorkflowNode
	for rows.Next() {
		var n WorkflowNode
		var configRaw []byte
		err := rows.Scan(&n.ID, &n.GraphID, &n.Type, &n.Name, &configRaw, &n.Position.X, &n.Position.Y, &n.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan node: %w", err)
		}
		if len(configRaw) > 0 {
			_ = json.Unmarshal(configRaw, &n.Config)
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

func (r *Repository) CreateNode(ctx context.Context, graphID string, nodeType, name string, config map[string]any, pos NodePosition) (*WorkflowNode, error) {
	id := uuid.New().String()
	configRaw, _ := json.Marshal(config)

	query := `
		INSERT INTO studio_workflow_nodes (id, graph_id, type, name, config, position_x, position_y, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING created_at`
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, query, id, graphID, nodeType, name, configRaw, pos.X, pos.Y).Scan(&createdAt)
	if err != nil {
		return nil, fmt.Errorf("create node: %w", err)
	}

	return &WorkflowNode{
		ID:        id,
		GraphID:   graphID,
		Type:      nodeType,
		Name:      name,
		Config:    config,
		Position:  pos,
		CreatedAt: createdAt,
	}, nil
}

func (r *Repository) UpdateNode(ctx context.Context, nodeID string, updates map[string]any) (*WorkflowNode, error) {
	var n WorkflowNode
	var configRaw []byte

	query := `
		UPDATE studio_workflow_nodes
		SET name = COALESCE($1, name),
		    type = COALESCE($2, type),
		    config = COALESCE($3, config),
		    position_x = COALESCE($4, position_x),
		    position_y = COALESCE($5, position_y)
		WHERE id = $6
		RETURNING id, graph_id, type, name, config, position_x, position_y, created_at`

	var name, nodeType *string
	var config *[]byte
	var posX, posY *float64

	if n, ok := updates["name"].(string); ok {
		name = &n
	}
	if t, ok := updates["type"].(string); ok {
		nodeType = &t
	}
	if c, ok := updates["config"].(map[string]any); ok {
		raw, _ := json.Marshal(c)
		config = &raw
	}
	if p, ok := updates["position"].(map[string]any); ok {
		if x, ok := p["x"].(float64); ok {
			posX = &x
		}
		if y, ok := p["y"].(float64); ok {
			posY = &y
		}
	}

	err := r.db.QueryRowContext(ctx, query, name, nodeType, config, posX, posY, nodeID).Scan(
		&n.ID, &n.GraphID, &n.Type, &n.Name, &configRaw, &n.Position.X, &n.Position.Y, &n.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update node: %w", err)
	}
	if len(configRaw) > 0 {
		_ = json.Unmarshal(configRaw, &n.Config)
	}
	return &n, nil
}

func (r *Repository) DeleteNode(ctx context.Context, nodeID string) error {
	query := `DELETE FROM studio_workflow_nodes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, nodeID)
	return err
}

func (r *Repository) ListEdges(ctx context.Context, graphID string) ([]WorkflowEdge, error) {
	query := `
		SELECT id, graph_id, source, target, condition, created_at
		FROM studio_workflow_edges
		WHERE graph_id = $1
		ORDER BY created_at`
	rows, err := r.db.QueryContext(ctx, query, graphID)
	if err != nil {
		return nil, fmt.Errorf("list edges: %w", err)
	}
	defer rows.Close()

	var edges []WorkflowEdge
	for rows.Next() {
		var e WorkflowEdge
		err := rows.Scan(&e.ID, &e.GraphID, &e.Source, &e.Target, &e.Condition, &e.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan edge: %w", err)
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

func (r *Repository) CreateEdge(ctx context.Context, graphID, source, target, condition string) (*WorkflowEdge, error) {
	id := uuid.New().String()

	query := `
		INSERT INTO studio_workflow_edges (id, graph_id, source, target, condition, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING created_at`
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, query, id, graphID, source, target, condition).Scan(&createdAt)
	if err != nil {
		return nil, fmt.Errorf("create edge: %w", err)
	}

	return &WorkflowEdge{
		ID:        id,
		GraphID:   graphID,
		Source:    source,
		Target:    target,
		Condition: condition,
		CreatedAt: createdAt,
	}, nil
}

func (r *Repository) UpdateEdge(ctx context.Context, edgeID string, updates map[string]any) (*WorkflowEdge, error) {
	query := `
		UPDATE studio_workflow_edges
		SET source = COALESCE($1, source),
		    target = COALESCE($2, target),
		    condition = COALESCE($3, condition)
		WHERE id = $4
		RETURNING id, graph_id, source, target, condition, created_at`

	var source, target, condition *string
	if s, ok := updates["source"].(string); ok {
		source = &s
	}
	if t, ok := updates["target"].(string); ok {
		target = &t
	}
	if c, ok := updates["condition"].(string); ok {
		condition = &c
	}

	var e WorkflowEdge
	err := r.db.QueryRowContext(ctx, query, source, target, condition, edgeID).Scan(
		&e.ID, &e.GraphID, &e.Source, &e.Target, &e.Condition, &e.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update edge: %w", err)
	}
	return &e, nil
}

func (r *Repository) DeleteEdge(ctx context.Context, edgeID string) error {
	query := `DELETE FROM studio_workflow_edges WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, edgeID)
	return err
}

func (r *Repository) ListExecutions(ctx context.Context, tenantID string, limit int) ([]WorkflowExecution, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, graph_id, tenant_id, status, started_at, completed_at, result, error, created_at
		FROM studio_workflow_executions
		WHERE tenant_id = $1
		ORDER BY started_at DESC
		LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("list executions: %w", err)
	}
	defer rows.Close()

	var execs []WorkflowExecution
	for rows.Next() {
		var e WorkflowExecution
		var resultRaw []byte
		var completedAt sql.NullTime
		err := rows.Scan(&e.ID, &e.GraphID, &e.TenantID, &e.Status, &e.StartedAt, &completedAt, &resultRaw, &e.Error, &e.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan execution: %w", err)
		}
		if completedAt.Valid {
			e.CompletedAt = &completedAt.Time
		}
		if len(resultRaw) > 0 {
			_ = json.Unmarshal(resultRaw, &e.Result)
		}
		execs = append(execs, e)
	}
	return execs, rows.Err()
}

func (r *Repository) GetExecution(ctx context.Context, tenantID, execID string) (*WorkflowExecution, error) {
	query := `
		SELECT id, graph_id, tenant_id, status, started_at, completed_at, result, error, created_at
		FROM studio_workflow_executions
		WHERE id = $1 AND tenant_id = $2`
	var e WorkflowExecution
	var resultRaw []byte
	var completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, execID, tenantID).Scan(
		&e.ID, &e.GraphID, &e.TenantID, &e.Status, &e.StartedAt, &completedAt, &resultRaw, &e.Error, &e.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get execution: %w", err)
	}
	if completedAt.Valid {
		e.CompletedAt = &completedAt.Time
	}
	if len(resultRaw) > 0 {
		_ = json.Unmarshal(resultRaw, &e.Result)
	}

	nodeResults, err := r.GetNodeResults(ctx, execID)
	if err == nil {
		e.NodeResults = nodeResults
	}

	return &e, nil
}

func (r *Repository) CreateExecution(ctx context.Context, graphID, tenantID string) (*WorkflowExecution, error) {
	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO studio_workflow_executions (id, graph_id, tenant_id, status, started_at, created_at)
		VALUES ($1, $2, $3, 'running', $4, $4)
		RETURNING started_at, created_at`
	err := r.db.QueryRowContext(ctx, query, id, graphID, tenantID, now).Scan(&now, &now)
	if err != nil {
		return nil, fmt.Errorf("create execution: %w", err)
	}

	return &WorkflowExecution{
		ID:        id,
		GraphID:   graphID,
		TenantID:  tenantID,
		Status:    "running",
		StartedAt: now,
		CreatedAt: now,
	}, nil
}

func (r *Repository) UpdateExecutionStatus(ctx context.Context, execID string, status string, result map[string]any, errMsg string) error {
	resultRaw, _ := json.Marshal(result)
	var completedAt *time.Time
	if status == "completed" || status == "failed" || status == "cancelled" {
		now := time.Now()
		completedAt = &now
	}

	query := `
		UPDATE studio_workflow_executions
		SET status = $1, result = $2, error = $3, completed_at = $4
		WHERE id = $5`
	_, dbErr := r.db.ExecContext(ctx, query, status, resultRaw, errMsg, completedAt, execID)
	return dbErr
}

func (r *Repository) AddNodeResult(ctx context.Context, execID string, nodeID, status string, output map[string]any, nodeErr string, durationMs int64) error {
	query := `
		INSERT INTO studio_workflow_execution_nodes (id, execution_id, node_id, status, output, error, duration_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	id := uuid.New().String()
	outputRaw, _ := json.Marshal(output)
	_, err := r.db.ExecContext(ctx, query, id, execID, nodeID, status, outputRaw, nodeErr, durationMs)
	return err
}

func (r *Repository) GetNodeResults(ctx context.Context, execID string) ([]NodeResult, error) {
	query := `
		SELECT node_id, status, output, error, duration_ms
		FROM studio_workflow_execution_nodes
		WHERE execution_id = $1
		ORDER BY duration_ms`
	rows, err := r.db.QueryContext(ctx, query, execID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []NodeResult
	for rows.Next() {
		var r NodeResult
		var outputRaw []byte
		err := rows.Scan(&r.NodeID, &r.Status, &outputRaw, &r.Error, &r.DurationMs)
		if err != nil {
			return nil, err
		}
		if len(outputRaw) > 0 {
			_ = json.Unmarshal(outputRaw, &r.Output)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (r *Repository) GetTimeline(ctx context.Context, graphID string, limit int) ([]TimelineEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT id, graph_id, event_type, source, message, metadata, timestamp
		FROM studio_workflow_timeline
		WHERE graph_id = $1
		ORDER BY timestamp DESC
		LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, graphID, limit)
	if err != nil {
		return nil, fmt.Errorf("get timeline: %w", err)
	}
	defer rows.Close()

	var events []TimelineEvent
	for rows.Next() {
		var e TimelineEvent
		var metaRaw []byte
		err := rows.Scan(&e.ID, &e.GraphID, &e.Type, &e.Source, &e.Message, &metaRaw, &e.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("scan timeline event: %w", err)
		}
		if len(metaRaw) > 0 {
			_ = json.Unmarshal(metaRaw, &e.Metadata)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *Repository) AddTimelineEvent(ctx context.Context, graphID, eventType, source, message string, metadata map[string]any) (*TimelineEvent, error) {
	id := uuid.New().String()
	metaRaw, _ := json.Marshal(metadata)

	query := `
		INSERT INTO studio_workflow_timeline (id, graph_id, event_type, source, message, metadata, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING timestamp`
	var ts time.Time
	err := r.db.QueryRowContext(ctx, query, id, graphID, eventType, source, message, metaRaw).Scan(&ts)
	if err != nil {
		return nil, fmt.Errorf("add timeline event: %w", err)
	}

	return &TimelineEvent{
		ID:        id,
		GraphID:   graphID,
		Type:      eventType,
		Source:    source,
		Message:   message,
		Metadata:  metadata,
		Timestamp: ts,
	}, nil
}

type TimelineEvent struct {
	ID        string         `json:"id"`
	GraphID   string         `json:"graph_id"`
	Type      string         `json:"type"`
	Source    string         `json:"source"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}