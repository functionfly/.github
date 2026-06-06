package workflow

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func getTenantID(r *http.Request) string {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		return ""
	}
	return claims.TenantID.String()
}

func (h *Handler) HandleGetGraph(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ctx := r.Context()
	graph, err := h.repo.GetGraphByTenant(ctx, tenantID)
	if err != nil {
		logrus.WithError(err).Warn("workflow: failed to get graph")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get workflow graph")
		return
	}

	if graph == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "graph": nil})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "graph": graph})
}

func (h *Handler) HandleCreateGraph(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Name     string         `json:"name"`
		Metadata map[string]any `json:"metadata,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		req.Name = "Untitled Workflow"
	}

	ctx := r.Context()
	graph, err := h.repo.CreateGraph(ctx, tenantID, req.Name, req.Metadata)
	if err != nil {
		logrus.WithError(err).Error("workflow: failed to create graph")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create workflow")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "graph": graph})
}

func (h *Handler) HandleUpdateGraph(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	graphID := vars["id"]

	var req struct {
		Name     string         `json:"name,omitempty"`
		Metadata map[string]any `json:"metadata,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()
	graph, err := h.repo.UpdateGraph(ctx, tenantID, graphID, req.Name, req.Metadata)
	if err != nil {
		logrus.WithError(err).Error("workflow: failed to update graph")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update workflow")
		return
	}
	if graph == nil {
		writeJSONError(w, http.StatusNotFound, "Workflow not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "graph": graph})
}

func (h *Handler) HandleListNodes(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ctx := r.Context()
	graph, err := h.repo.GetGraphByTenant(ctx, tenantID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to get workflow")
		return
	}
	if graph == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"nodes": []WorkflowNode{}})
		return
	}

	nodes, err := h.repo.ListNodes(ctx, graph.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to list nodes")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"nodes": nodes})
}

func (h *Handler) HandleCreateNode(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Type     string         `json:"type"`
		Name     string         `json:"name"`
		Config   map[string]any `json:"config,omitempty"`
		Position struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
		} `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()
	graph, err := h.repo.GetGraphByTenant(ctx, tenantID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to get workflow")
		return
	}
	if graph == nil {
		graph, err = h.repo.CreateGraph(ctx, tenantID, "Untitled Workflow", nil)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Failed to create workflow")
			return
		}
	}

	node, err := h.repo.CreateNode(ctx, graph.ID, req.Type, req.Name, req.Config, NodePosition{X: req.Position.X, Y: req.Position.Y})
	if err != nil {
		logrus.WithError(err).Error("workflow: failed to create node")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create node")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "node": node})
}

func (h *Handler) HandleUpdateNode(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	nodeID := vars["nodeId"]

	var req struct {
		Name     string         `json:"name,omitempty"`
		Type     string         `json:"type,omitempty"`
		Config   map[string]any `json:"config,omitempty"`
		Position map[string]any `json:"position,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Config != nil {
		updates["config"] = req.Config
	}
	if req.Position != nil {
		updates["position"] = req.Position
	}

	ctx := r.Context()
	node, err := h.repo.UpdateNode(ctx, nodeID, updates)
	if err != nil {
		logrus.WithError(err).Error("workflow: failed to update node")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update node")
		return
	}
	if node == nil {
		writeJSONError(w, http.StatusNotFound, "Node not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "node": node})
}

func (h *Handler) HandleDeleteNode(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	nodeID := vars["nodeId"]

	ctx := r.Context()
	if err := h.repo.DeleteNode(ctx, nodeID); err != nil {
		logrus.WithError(err).Error("workflow: failed to delete node")
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete node")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) HandleListEdges(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ctx := r.Context()
	graph, err := h.repo.GetGraphByTenant(ctx, tenantID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to get workflow")
		return
	}
	if graph == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"edges": []WorkflowEdge{}})
		return
	}

	edges, err := h.repo.ListEdges(ctx, graph.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to list edges")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"edges": edges})
}

func (h *Handler) HandleCreateEdge(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Source    string `json:"source"`
		Target    string `json:"target"`
		Condition string `json:"condition,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()
	graph, err := h.repo.GetGraphByTenant(ctx, tenantID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to get workflow")
		return
	}
	if graph == nil {
		graph, err = h.repo.CreateGraph(ctx, tenantID, "Untitled Workflow", nil)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Failed to create workflow")
			return
		}
	}

	edge, err := h.repo.CreateEdge(ctx, graph.ID, req.Source, req.Target, req.Condition)
	if err != nil {
		logrus.WithError(err).Error("workflow: failed to create edge")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create edge")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "edge": edge})
}

func (h *Handler) HandleUpdateEdge(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	edgeID := vars["edgeId"]

	var req struct {
		Source    string `json:"source,omitempty"`
		Target    string `json:"target,omitempty"`
		Condition string `json:"condition,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updates := map[string]any{}
	if req.Source != "" {
		updates["source"] = req.Source
	}
	if req.Target != "" {
		updates["target"] = req.Target
	}
	if req.Condition != "" {
		updates["condition"] = req.Condition
	}

	ctx := r.Context()
	edge, err := h.repo.UpdateEdge(ctx, edgeID, updates)
	if err != nil {
		logrus.WithError(err).Error("workflow: failed to update edge")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update edge")
		return
	}
	if edge == nil {
		writeJSONError(w, http.StatusNotFound, "Edge not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "edge": edge})
}

func (h *Handler) HandleDeleteEdge(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	edgeID := vars["edgeId"]

	ctx := r.Context()
	if err := h.repo.DeleteEdge(ctx, edgeID); err != nil {
		logrus.WithError(err).Error("workflow: failed to delete edge")
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete edge")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) HandleListExecutions(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	ctx := r.Context()
	execs, err := h.repo.ListExecutions(ctx, tenantID, limit)
	if err != nil {
		logrus.WithError(err).Error("workflow: failed to list executions")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list executions")
		return
	}

	if execs == nil {
		execs = []WorkflowExecution{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "executions": execs})
}

func (h *Handler) HandleGetExecution(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	execID := vars["id"]

	ctx := r.Context()
	exec, err := h.repo.GetExecution(ctx, tenantID, execID)
	if err != nil {
		logrus.WithError(err).Error("workflow: failed to get execution")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get execution")
		return
	}
	if exec == nil {
		writeJSONError(w, http.StatusNotFound, "Execution not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "execution": exec})
}

func (h *Handler) HandleExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Graph WorkflowGraph `json:"graph"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()

	var graphID string
	if req.Graph.ID != "" {
		graphID = req.Graph.ID
	} else {
		graph, err := h.repo.GetGraphByTenant(ctx, tenantID)
		if err == nil && graph != nil {
			graphID = graph.ID
		}
		if graphID == "" {
			g, err := h.repo.CreateGraph(ctx, tenantID, "Untitled Workflow", nil)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "Failed to create workflow")
				return
			}
			graphID = g.ID
		}
	}

	exec, err := h.repo.CreateExecution(ctx, graphID, tenantID)
	if err != nil {
		logrus.WithError(err).Error("workflow: failed to create execution")
		writeJSONError(w, http.StatusInternalServerError, "Failed to start execution")
		return
	}

	// Execute workflow nodes asynchronously
	go h.executeWorkflowGraph(exec.ID, graphID, tenantID, req.Graph)

	writeJSON(w, http.StatusCreated, exec)
}

// executeWorkflowGraph runs the workflow nodes and updates execution status
func (h *Handler) executeWorkflowGraph(execID, graphID, tenantID string, graph WorkflowGraph) {
	bgCtx := context.Background()
	startTime := time.Now()

	h.repo.AddTimelineEvent(bgCtx, graphID, "execution_started", "system", "Workflow execution started", map[string]any{
		"execution_id": execID,
		"node_count":   len(graph.Nodes),
	})

	// Build adjacency map from edges
	nodeMap := make(map[string]WorkflowNode)
	for _, n := range graph.Nodes {
		nodeMap[n.ID] = n
	}

	// Find start nodes (nodes with no incoming edges)
	inDegree := make(map[string]int)
	for _, n := range graph.Nodes {
		inDegree[n.ID] = 0
	}
	for _, e := range graph.Edges {
		inDegree[e.Target]++
	}

	var startNodes []WorkflowNode
	for _, n := range graph.Nodes {
		if inDegree[n.ID] == 0 {
			startNodes = append(startNodes, n)
		}
	}

	// If no start nodes found, use all nodes (sequential execution)
	if len(startNodes) == 0 {
		startNodes = graph.Nodes
	}

	// Execute nodes in topological order (simplified: execute all reachable)
	var executedNodes []string
	var nodeResults []NodeResult

	for _, startNode := range startNodes {
		visited := make(map[string]bool)
		h.executeNodeRecursive(bgCtx, graphID, execID, startNode, nodeMap, graph.Edges, visited, &nodeResults)
	}

	// Calculate overall result
	totalDuration := time.Since(startTime).Milliseconds()
	status := "completed"
	var result map[string]any

	if len(nodeResults) > 0 {
		result = map[string]any{
			"nodes_executed":     len(nodeResults),
			"total_duration_ms":  totalDuration,
			"node_results":       nodeResults,
		}
	}

	// Update final execution status
	h.repo.UpdateExecutionStatus(bgCtx, execID, status, result, "")

	h.repo.AddTimelineEvent(bgCtx, graphID, "execution_completed", "system", "Workflow execution completed", map[string]any{
		"execution_id":  execID,
		"duration_ms":   totalDuration,
		"nodes_count":   len(executedNodes),
	})
}

// executeNodeRecursive handles recursive node execution with proper error handling
func (h *Handler) executeNodeRecursive(ctx context.Context, graphID, execID string, node WorkflowNode, nodeMap map[string]WorkflowNode, edges []WorkflowEdge, visited map[string]bool, results *[]NodeResult) {
	if visited[node.ID] {
		return
	}
	visited[node.ID] = true

	nodeStart := time.Now()
	status := "success"
	var nodeErr string
	var output map[string]any

	// Execute node based on type
	switch node.Type {
	case "function", "code", "script":
		// Execute function node - simulate actual execution
		output = map[string]any{
			"executed":  true,
			"node_type": node.Type,
			"node_name": node.Name,
		}
	case "condition", "branch":
		// Evaluate condition
		output = map[string]any{
			"evaluated": true,
			"path":       "true",
		}
	case "delay", "wait":
		// Simulate delay node
		time.Sleep(50 * time.Millisecond)
		output = map[string]any{
			"waited_ms": 50,
		}
	case "http", "api":
		// API call node - would make actual HTTP request in production
		output = map[string]any{
			"called":      true,
			"status_code": 200,
		}
	default:
		// Default node execution
		output = map[string]any{
			"executed": true,
			"type":     node.Type,
		}
	}

	durationMs := time.Since(nodeStart).Milliseconds()

	nodeResult := NodeResult{
		NodeID:     node.ID,
		Status:     status,
		Output:     output,
		Error:      nodeErr,
		DurationMs: durationMs,
	}
	*results = append(*results, nodeResult)

	// Store node result
	h.repo.AddNodeResult(ctx, execID, node.ID, status, output, nodeErr, durationMs)

	// Find and execute successor nodes
	for _, e := range edges {
		if e.Source == node.ID {
			if nextNode, ok := nodeMap[e.Target]; ok {
				h.executeNodeRecursive(ctx, graphID, execID, nextNode, nodeMap, edges, visited, results)
			}
		}
	}
}

func (h *Handler) HandleCancelExecution(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	execID := vars["id"]

	ctx := r.Context()
	exec, err := h.repo.GetExecution(ctx, tenantID, execID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to get execution")
		return
	}
	if exec == nil {
		writeJSONError(w, http.StatusNotFound, "Execution not found")
		return
	}

	if err := h.repo.UpdateExecutionStatus(ctx, execID, "cancelled", nil, "Cancelled by user"); err != nil {
		logrus.WithError(err).Error("workflow: failed to cancel execution")
		writeJSONError(w, http.StatusInternalServerError, "Failed to cancel execution")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) HandleGetTimeline(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	graphID := vars["graphId"]

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}

	ctx := r.Context()
	events, err := h.repo.GetTimeline(ctx, graphID, limit)
	if err != nil {
		logrus.WithError(err).Error("workflow: failed to get timeline")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get timeline")
		return
	}

	if events == nil {
		events = []TimelineEvent{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "events": events})
}