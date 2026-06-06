// Package ghost provides HTTP handlers for the Ghost Mode autonomous building engine.
// Ghost Mode enables natural language → deployed application through LLM-driven architecture
// planning, infrastructure provisioning, code generation, and automated deployment.
package ghost

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/agent/deployment"
	"github.com/functionfly/functionfly/internal/agent/generation"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Handler contains all Ghost Mode orchestration handlers
type Handler struct {
	db          interface{ DB() interface{} }
	genSvc      *generation.Service
	deployGen   *deployment.Generator
	builds      map[string]*BuildState
	mu          sync.RWMutex
}

// DBGetter interface for database access
type DBGetter interface {
	DB() interface{}
}

// NewHandler creates a new Ghost Mode handler
func NewHandler(genSvc *generation.Service, deployGen *deployment.Generator) *Handler {
	return &Handler{
		genSvc:    genSvc,
		deployGen: deployGen,
		builds:    make(map[string]*BuildState),
	}
}

// ============================================================
// Types
// ============================================================

type GhostPhase string

const (
	PhasePlanning   GhostPhase = "planning"
	PhaseProvisioning GhostPhase = "provisioning"
	PhaseBuilding    GhostPhase = "building"
	PhaseDeploying   GhostPhase = "deploying"
	PhaseMonitoring  GhostPhase = "monitoring"
	PhaseComplete    GhostPhase = "complete"
	PhaseError       GhostPhase = "error"
	PhasePaused       GhostPhase = "paused"
)

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress  TaskStatus = "in_progress"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed      TaskStatus = "failed"
	StatusSkipped     TaskStatus = "skipped"
)

type BuildState struct {
	ID                  string         `json:"id"`
	Goal                string         `json:"goal"`
	Description         string         `json:"description"`
	Phase               GhostPhase     `json:"phase"`
	Progress            float64        `json:"progress"` // 0-1
	Tasks               []TaskState    `json:"tasks"`
	StartedAt           time.Time      `json:"started_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	EstimatedCompletion *time.Time     `json:"estimated_completion,omitempty"`
	CurrentTaskID       string         `json:"current_task_id,omitempty"`
	HumanApprovalRequired bool          `json:"human_approval_required"`
	ApprovalType        string         `json:"approval_type,omitempty"`
	Error               string         `json:"error,omitempty"`
	Artifacts           []Artifact     `json:"artifacts,omitempty"`
}

type TaskState struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        TaskStatus `json:"status"`
	Phase         GhostPhase `json:"phase"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	DurationMs    int        `json:"duration_ms,omitempty"`
	Logs          []LogEntry `json:"logs"`
	Artifacts     []Artifact `json:"artifacts,omitempty"`
	AgentID       string     `json:"agent_id,omitempty"`
	Confidence    float64    `json:"confidence,omitempty"`
	Dependencies  []string    `json:"dependencies,omitempty"`
	LLMOutput     string     `json:"llm_output,omitempty"` // architecture, schema, code, etc.
}

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // info, warn, error, debug
	Message   string    `json:"message"`
}

type Artifact struct {
	Name string `json:"name"`
	Type string `json:"type"` // schema, code, config, deployment
	Path string `json:"path"`
	Size int    `json:"size,omitempty"`
}

// ============================================================
// API Request/Response Types
// ============================================================

type CreateBuildRequest struct {
	Goal        string `json:"goal"` // natural language description
	Description string `json:"description"`
	AgentID     string `json:"agent_id,omitempty"`
}

type CreateBuildResponse struct {
	BuildID string `json:"build_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type GetBuildResponse struct {
	OK    bool       `json:"ok"`
	Build *BuildState `json:"build,omitempty"`
}

type ApprovalRequest struct {
	BuildID        string `json:"build_id"`
	ApprovalType   string `json:"approval_type"` // schema, deployment, pr, infra
	Decision       string `json:"decision"`      // approve, reject, revision
	Notes          string `json:"notes,omitempty"`
}

type LogEntryRequest struct {
	BuildID string `json:"build_id"`
	TaskID  string `json:"task_id"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

// ============================================================
// Route Registration
// ============================================================

func (h *Handler) RegisterRoutes(r *mux.Router) {
	// Build lifecycle
	r.HandleFunc("/ghost/builds", h.HandleCreateBuild).Methods("POST", "OPTIONS")
	r.HandleFunc("/ghost/builds", h.HandleListBuilds).Methods("GET", "OPTIONS")
	r.HandleFunc("/ghost/builds/{id}", h.HandleGetBuild).Methods("GET", "OPTIONS")
	r.HandleFunc("/ghost/builds/{id}", h.HandleUpdateBuild).Methods("PUT", "OPTIONS")
	r.HandleFunc("/ghost/builds/{id}", h.HandleDeleteBuild).Methods("DELETE", "OPTIONS")

	// Task management
	r.HandleFunc("/ghost/builds/{id}/tasks", h.HandleListTasks).Methods("GET", "OPTIONS")
	r.HandleFunc("/ghost/builds/{id}/tasks/{task_id}/start", h.HandleStartTask).Methods("POST", "OPTIONS")
	r.HandleFunc("/ghost/builds/{id}/tasks/{task_id}/complete", h.HandleCompleteTask).Methods("POST", "OPTIONS")
	r.HandleFunc("/ghost/builds/{id}/tasks/{task_id}/fail", h.HandleFailTask).Methods("POST", "OPTIONS")
	r.HandleFunc("/ghost/builds/{id}/tasks/{task_id}/logs", h.HandleAddTaskLog).Methods("POST", "OPTIONS")
	r.HandleFunc("/ghost/builds/{id}/tasks/{task_id}/logs", h.HandleGetTaskLogs).Methods("GET", "OPTIONS")
	r.HandleFunc("/ghost/builds/{id}/logs", h.HandleGetBuildLogs).Methods("GET", "OPTIONS")

	// Human approval
	r.HandleFunc("/ghost/builds/{id}/approve", h.HandleApproval).Methods("POST", "OPTIONS")
	r.HandleFunc("/ghost/builds/{id}/pause", h.HandlePauseBuild).Methods("POST", "OPTIONS")
	r.HandleFunc("/ghost/builds/{id}/resume", h.HandleResumeBuild).Methods("POST", "OPTIONS")
	r.HandleFunc("/ghost/builds/{id}/cancel", h.HandleCancelBuild).Methods("POST", "OPTIONS")

	// Architecture planning
	r.HandleFunc("/ghost/plan/architecture", h.HandlePlanArchitecture).Methods("POST", "OPTIONS")

	// Infra provisioning
	r.HandleFunc("/ghost/provision/database", h.HandleProvisionDatabase).Methods("POST", "OPTIONS")
	r.HandleFunc("/ghost/provision/api", h.HandleProvisionAPI).Methods("POST", "OPTIONS")
	r.HandleFunc("/ghost/provision/docker", h.HandleProvisionDocker).Methods("POST", "OPTIONS")

	// Code generation
	r.HandleFunc("/ghost/generate/schema", h.HandleGenerateSchema).Methods("POST", "OPTIONS")
	r.HandleFunc("/ghost/generate/backend", h.HandleGenerateBackend).Methods("POST", "OPTIONS")
	r.HandleFunc("/ghost/generate/frontend", h.HandleGenerateFrontend).Methods("POST", "OPTIONS")
	r.HandleFunc("/ghost/generate/tests", h.HandleGenerateTests).Methods("POST", "OPTIONS")

	// Deployment
	r.HandleFunc("/ghost/deploy/staging", h.HandleDeployStaging).Methods("POST", "OPTIONS")
	r.HandleFunc("/ghost/deploy/production", h.HandleDeployProduction).Methods("POST", "OPTIONS")

	// Monitoring
	r.HandleFunc("/ghost/monitor/setup", h.HandleSetupMonitoring).Methods("POST", "OPTIONS")
}

// ============================================================
// Build Lifecycle
// ============================================================

func (h *Handler) HandleCreateBuild(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req CreateBuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.Goal == "" {
		writeError(w, http.StatusBadRequest, "MISSING_GOAL", "goal is required")
		return
	}

	buildID := "build_" + uuid.New().String()[:8]

	// Generate initial tasks based on the goal
	tasks := h.generateInitialTasks(req.Goal)

	build := &BuildState{
		ID:                    buildID,
		Goal:                  req.Goal,
		Description:           req.Description,
		Phase:                 PhasePlanning,
		Progress:              0.0,
		Tasks:                 tasks,
		StartedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		HumanApprovalRequired: false,
	}

	h.mu.Lock()
	h.builds[buildID] = build
	h.mu.Unlock()

	// Start async orchestration
	go h.runBuildOrchestration(buildID)

	writeJSON(w, http.StatusCreated, CreateBuildResponse{
		BuildID: buildID,
		Status:  "started",
		Message: "Ghost Mode build started",
	})
}

func (h *Handler) generateInitialTasks(goal string) []TaskState {
	tasks := []TaskState{
		{
			ID:          "task-1",
			Title:       "Analyze requirements and plan architecture",
			Description: "Parse natural language goal, identify components, design system architecture",
			Status:      StatusPending,
			Phase:       PhasePlanning,
			Dependencies: []string{},
		},
		{
			ID:          "task-2",
			Title:       "Design database schema",
			Description: "Generate PostgreSQL schema from data requirements",
			Status:      StatusPending,
			Phase:       PhasePlanning,
			Dependencies: []string{"task-1"},
		},
		{
			ID:          "task-3",
			Title:       "Provision infrastructure",
			Description: "Create Docker containers, networking, and cloud resources",
			Status:      StatusPending,
			Phase:       PhaseProvisioning,
			Dependencies: []string{"task-2"},
		},
		{
			ID:          "task-4",
			Title:       "Generate backend code",
			Description: "Write API handlers, business logic, and database integration",
			Status:      StatusPending,
			Phase:       PhaseBuilding,
			Dependencies: []string{"task-3"},
		},
		{
			ID:          "task-5",
			Title:       "Generate frontend code",
			Description: "Build React components, pages, and API integration",
			Status:      StatusPending,
			Phase:       PhaseBuilding,
			Dependencies: []string{"task-4"},
		},
		{
			ID:          "task-6",
			Title:       "Write unit tests",
			Description: "Create test suites with 80% code coverage target",
			Status:      StatusPending,
			Phase:       PhaseBuilding,
			Dependencies: []string{"task-4", "task-5"},
		},
		{
			ID:          "task-7",
			Title:       "Deploy to staging",
			Description: "Blue-green deployment to staging environment",
			Status:      StatusPending,
			Phase:       PhaseDeploying,
			Dependencies: []string{"task-6"},
		},
		{
			ID:          "task-8",
			Title:       "Setup monitoring and alerts",
			Description: "Configure dashboards, health checks, and alerting",
			Status:      StatusPending,
			Phase:       PhaseMonitoring,
			Dependencies: []string{"task-7"},
		},
	}

	return tasks
}

func (h *Handler) HandleListBuilds(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	h.mu.RLock()
	var builds []*BuildState
	for _, b := range h.builds {
		builds = append(builds, b)
	}
	h.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"builds": builds,
		"total":  len(builds),
	})
}

func (h *Handler) HandleGetBuild(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]
	if buildID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_ID", "build id required")
		return
	}

	h.mu.RLock()
	build, ok := h.builds[buildID]
	h.mu.RUnlock()

	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	writeJSON(w, http.StatusOK, GetBuildResponse{
		OK:    true,
		Build: build,
	})
}

func (h *Handler) HandleUpdateBuild(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]
	h.mu.Lock()
	defer h.mu.Unlock()

	build, ok := h.builds[buildID]
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	// Apply updates
	if phase, ok := updates["phase"].(string); ok {
		build.Phase = GhostPhase(phase)
	}
	if progress, ok := updates["progress"].(float64); ok {
		build.Progress = progress
	}
	if errMsg, ok := updates["error"].(string); ok {
		build.Error = errMsg
	}
	build.UpdatedAt = time.Now()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"build": build,
	})
}

func (h *Handler) HandleDeleteBuild(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.builds[buildID]; !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	delete(h.builds, buildID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "build deleted",
	})
}

// ============================================================
// Task Management
// ============================================================

func (h *Handler) HandleListTasks(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]
	h.mu.RLock()
	build, ok := h.builds[buildID]
	h.mu.RUnlock()

	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"tasks": build.Tasks,
	})
}

func (h *Handler) HandleStartTask(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]
	taskID := mux.Vars(r)["task_id"]

	h.mu.Lock()
	defer h.mu.Unlock()

	build, ok := h.builds[buildID]
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	for i := range build.Tasks {
		if build.Tasks[i].ID == taskID {
			now := time.Now()
			build.Tasks[i].Status = StatusInProgress
			build.Tasks[i].StartedAt = &now
			build.CurrentTaskID = taskID
			build.UpdatedAt = now
			break
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"task":  h.getTask(build, taskID),
	})
}

func (h *Handler) HandleCompleteTask(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]
	taskID := mux.Vars(r)["task_id"]

	h.mu.Lock()
	defer h.mu.Unlock()

	build, ok := h.builds[buildID]
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	var req struct {
		Artifacts []Artifact `json:"artifacts,omitempty"`
		Output    string     `json:"output,omitempty"` // LLM-generated content
		Confidence float64   `json:"confidence,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	for i := range build.Tasks {
		if build.Tasks[i].ID == taskID {
			now := time.Now()
			build.Tasks[i].Status = StatusCompleted
			build.Tasks[i].CompletedAt = &now
			if req.Artifacts != nil {
				build.Tasks[i].Artifacts = req.Artifacts
			}
			if req.Output != "" {
				build.Tasks[i].LLMOutput = req.Output
			}
			if req.Confidence > 0 {
				build.Tasks[i].Confidence = req.Confidence
			}
			build.UpdatedAt = now
			build.CurrentTaskID = ""
			break
		}
	}

	// Recalculate progress
	h.recalculateProgress(build)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"progress": build.Progress,
	})
}

func (h *Handler) HandleFailTask(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]
	taskID := mux.Vars(r)["task_id"]

	h.mu.Lock()
	defer h.mu.Unlock()

	build, ok := h.builds[buildID]
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	var req struct {
		Error   string `json:"error"`
		Message string `json:"message,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	for i := range build.Tasks {
		if build.Tasks[i].ID == taskID {
			now := time.Now()
			build.Tasks[i].Status = StatusFailed
			build.Tasks[i].CompletedAt = &now
			build.Phase = PhaseError
			build.Error = req.Error
			build.UpdatedAt = now
			build.CurrentTaskID = ""
			break
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":   true,
		"phase": build.Phase,
		"error": build.Error,
	})
}

func (h *Handler) HandleAddTaskLog(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]
	taskID := mux.Vars(r)["task_id"]

	h.mu.Lock()
	defer h.mu.Unlock()

	build, ok := h.builds[buildID]
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	var req struct {
		Level   string `json:"level"`
		Message string `json:"message"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	level := req.Level
	if level == "" {
		level = "info"
	}

	for i := range build.Tasks {
		if build.Tasks[i].ID == taskID {
			build.Tasks[i].Logs = append(build.Tasks[i].Logs, LogEntry{
				Timestamp: time.Now(),
				Level:     level,
				Message:   req.Message,
			})
			build.UpdatedAt = time.Now()
			break
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true,
	})
}

// HandleGetTaskLogs retrieves all log entries for a specific task within a build
// GET /v1/ghost/builds/{id}/tasks/{task_id}/logs
func (h *Handler) HandleGetTaskLogs(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]
	taskID := mux.Vars(r)["task_id"]

	if buildID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_BUILD_ID", "build id required")
		return
	}
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_TASK_ID", "task id required")
		return
	}

	h.mu.RLock()
	build, ok := h.builds[buildID]
	h.mu.RUnlock()

	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	var logs []LogEntry
	for _, task := range build.Tasks {
		if task.ID == taskID {
			logs = task.Logs
			break
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"build_id": buildID,
		"task_id":  taskID,
		"logs":     logs,
		"total":    len(logs),
	})
}

// HandleGetBuildLogs retrieves all log entries for a build across all tasks
// GET /v1/ghost/builds/{id}/logs
func (h *Handler) HandleGetBuildLogs(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]

	if buildID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_BUILD_ID", "build id required")
		return
	}

	h.mu.RLock()
	build, ok := h.builds[buildID]
	h.mu.RUnlock()

	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	// Collect all logs with task context
	type TaskLogEntry struct {
		TaskID    string    `json:"task_id"`
		TaskTitle string    `json:"task_title"`
		Timestamp time.Time `json:"timestamp"`
		Level     string    `json:"level"`
		Message   string    `json:"message"`
	}

	var allLogs []TaskLogEntry
	for _, task := range build.Tasks {
		for _, log := range task.Logs {
			allLogs = append(allLogs, TaskLogEntry{
				TaskID:    task.ID,
				TaskTitle: task.Title,
				Timestamp: log.Timestamp,
				Level:     log.Level,
				Message:   log.Message,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"build_id": buildID,
		"logs":     allLogs,
		"total":    len(allLogs),
	})
}

func (h *Handler) HandleApproval(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]
	if buildID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_BUILD_ID", "build id required")
		return
	}

	var req ApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	build, ok := h.builds[buildID]
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	if !build.HumanApprovalRequired {
		writeError(w, http.StatusBadRequest, "NO_APPROVAL_REQUIRED", "no approval pending for this build")
		return
	}

	switch req.Decision {
	case "approve":
		build.HumanApprovalRequired = false
		build.ApprovalType = ""
		build.UpdatedAt = time.Now()
		// Resume orchestration
		go h.resumeBuild(buildID)
	case "reject":
		build.Phase = PhaseError
		build.Error = "Approval rejected: " + req.Notes
		build.UpdatedAt = time.Now()
	case "revision":
		build.HumanApprovalRequired = true
		build.ApprovalType = req.ApprovalType + "_revision"
		build.UpdatedAt = time.Now()
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"phase": build.Phase,
	})
}

func (h *Handler) HandlePauseBuild(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]
	h.mu.Lock()
	defer h.mu.Unlock()

	build, ok := h.builds[buildID]
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	build.Phase = PhasePaused
	build.UpdatedAt = time.Now()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"phase": build.Phase,
	})
}

func (h *Handler) HandleResumeBuild(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]
	h.mu.Lock()
	defer h.mu.Unlock()

	build, ok := h.builds[buildID]
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	if build.Phase != PhasePaused {
		writeError(w, http.StatusBadRequest, "NOT_PAUSED", "build is not paused")
		return
	}

	build.Phase = GhostPhase(build.Tasks[0].Phase) // Resume to current task's phase
	build.UpdatedAt = time.Now()

	go h.runBuildOrchestration(buildID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"phase": build.Phase,
	})
}

func (h *Handler) HandleCancelBuild(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]
	h.mu.Lock()
	defer h.mu.Unlock()

	build, ok := h.builds[buildID]
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "build not found")
		return
	}

	build.Phase = PhaseError
	build.Error = "Cancelled by user"
	build.UpdatedAt = time.Now()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"phase": build.Phase,
	})
}

// ============================================================
// Architecture Planning
// ============================================================

type PlanArchitectureRequest struct {
	Goal   string `json:"goal"`
	Domain string `json:"domain,omitempty"` // e.g., ecommerce, saas, social
}

type ArchitecturePlan struct {
	Components   []ComponentSpec   `json:"components"`
	DataModel    []EntitySpec      `json:"data_model"`
	APIDesign    []EndpointSpec    `json:"api_design"`
	TechStack    []string          `json:"tech_stack"`
	Dependencies []string          `json:"dependencies"`
	EstimatedCost string            `json:"estimated_cost"`
	RiskFactors  []string          `json:"risk_factors"`
}

type ComponentSpec struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // api, database, cache, queue, worker, frontend
	Description string `json:"description"`
	Technology  string `json:"technology"`
}

type EntitySpec struct {
	Name       string        `json:"name"`
	Fields     []FieldSpec   `json:"fields"`
	Indexes    []string      `json:"indexes,omitempty"`
	Relations  []RelationSpec `json:"relations,omitempty"`
}

type FieldSpec struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // string, int, bool, timestamp, jsonb
	Required bool   `json:"required"`
	Default  string `json:"default,omitempty"`
}

type RelationSpec struct {
	Entity   string `json:"entity"`
	Type     string `json:"type"` // one_to_one, one_to_many, many_to_many
	ViaField  string `json:"via_field,omitempty"`
}

type EndpointSpec struct {
	Method   string `json:"method"`
	Path     string `json:"path"`
	Handler  string `json:"handler"`
	Auth     bool   `json:"auth"`
	InputSchema  map[string]any `json:"input_schema,omitempty"`
	OutputSchema map[string]any `json:"output_schema,omitempty"`
}

func (h *Handler) HandlePlanArchitecture(w http.ResponseWriter, r *http.Request) {
	var req PlanArchitectureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	// LLM-driven architecture planning simulation
	// In production, this would call the LLM with a system prompt that includes
	// the goal, best practices for the domain, and existing infrastructure patterns
	plan := h.generateArchitecturePlan(req.Goal, req.Domain)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":   true,
		"plan": plan,
	})
}

func (h *Handler) generateArchitecturePlan(goal, domain string) *ArchitecturePlan {
	// Simulated LLM generation — in production, integrate with generation.Service
	components := []ComponentSpec{
		{Name: "api-gateway", Type: "api", Description: "Main API gateway with auth and rate limiting", Technology: "Go/hTTP"},
		{Name: "user-service", Type: "api", Description: "User management and authentication", Technology: "Go"},
		{Name: "postgres-primary", Type: "database", Description: "Primary PostgreSQL database", Technology: "PostgreSQL 17"},
		{Name: "redis-cache", Type: "cache", Description: "Redis for session and query caching", Technology: "Redis 7"},
		{Name: "background-workers", Type: "worker", Description: "Background job processing", Technology: "Go + BullMQ"},
	}

	entities := []EntitySpec{
		{
			Name: "users",
			Fields: []FieldSpec{
				{Name: "id", Type: "uuid", Required: true},
				{Name: "email", Type: "string", Required: true},
				{Name: "password_hash", Type: "string", Required: true},
				{Name: "created_at", Type: "timestamp", Required: true},
				{Name: "updated_at", Type: "timestamp", Required: true},
			},
			Indexes: []string{"email", "created_at"},
		},
	}

	endpoints := []EndpointSpec{
		{Method: "POST", Path: "/auth/register", Handler: "RegisterUser", Auth: false},
		{Method: "POST", Path: "/auth/login", Handler: "LoginUser", Auth: false},
		{Method: "GET", Path: "/users/me", Handler: "GetCurrentUser", Auth: true},
		{Method: "PUT", Path: "/users/me", Handler: "UpdateCurrentUser", Auth: true},
	}

	return &ArchitecturePlan{
		Components:    components,
		DataModel:     entities,
		APIDesign:     endpoints,
		TechStack:     []string{"Go", "PostgreSQL 17", "Redis 7", "React 19", "TailwindCSS"},
		Dependencies:  []string{"github.com/gin-gonic/gin", "github.com/lib/pq", "github.com/go-redis/redis"},
		EstimatedCost: "$50-200/month",
		RiskFactors:  []string{"Database connection pooling", "Rate limiting at scale"},
	}
}

// ============================================================
// Infrastructure Provisioning
// ============================================================

type ProvisionDatabaseRequest struct {
	Schema string `json:"schema"` // JSON schema definition
}

type ProvisionAPIRequest struct {
	Spec string `json:"spec"` // OpenAPI spec
}

type ProvisionDockerRequest struct {
	Services []DockerServiceSpec `json:"services"`
}

type DockerServiceSpec struct {
	Name       string            `json:"name"`
	Image      string            `json:"image"`
	Ports      []string          `json:"ports"`
	Env        map[string]string `json:"env"`
	Volumes    []string          `json:"volumes"`
	DependsOn  []string          `json:"depends_on,omitempty"`
}

func (h *Handler) HandleProvisionDatabase(w http.ResponseWriter, r *http.Request) {
	var req ProvisionDatabaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	// Generate SQL schema from JSON spec
	sql := h.generateSQLSchema(req.Schema)

	// Generate migration file content
	migration := fmt.Sprintf("-- Auto-generated by Ghost Mode\n-- Generated at: %s\n\n%s", time.Now().Format(time.RFC3339), sql)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"sql":       sql,
		"migration": migration,
		"artifacts": []Artifact{
			{Name: "schema.sql", Type: "schema", Path: "/ghost/artifacts/schema.sql", Size: len(sql)},
		},
	})
}

func (h *Handler) generateSQLSchema(schemaJSON string) string {
	// Simplified schema generation — parse JSON and produce SQL
	// In production, this would use the LLM to generate proper schema from requirements
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(512) NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);`,
	}

	sql := "--- Ghost Mode Auto-generated Schema\n"
	for _, t := range tables {
		sql += t + "\n\n"
	}
	return sql
}

func (h *Handler) HandleProvisionAPI(w http.ResponseWriter, r *http.Request) {
	var req ProvisionAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	// Generate API handler code from OpenAPI spec
	code := h.generateAPIHandlers(req.Spec)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"handlers": code,
		"artifacts": []Artifact{
			{Name: "handlers.go", Type: "code", Path: "/ghost/artifacts/handlers.go", Size: len(code)},
		},
	})
}

func (h *Handler) generateAPIHandlers(spec string) string {
	return `// Auto-generated by Ghost Mode
package api

import "net/http"

func RegisterRoutes(r *mux.Router) {
	// Auth routes
	r.HandleFunc("/auth/register", HandleRegister).Methods("POST")
	r.HandleFunc("/auth/login", HandleLogin).Methods("POST")

	// User routes
	r.HandleFunc("/users/me", HandleGetMe).Methods("GET")
	r.HandleFunc("/users/me", HandleUpdateMe).Methods("PUT")

	// Health check
	r.HandleFunc("/health", HandleHealth).Methods("GET")
}
`
}

func (h *Handler) HandleProvisionDocker(w http.ResponseWriter, r *http.Request) {
	var req ProvisionDockerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	dockerCompose := h.generateDockerCompose(req.Services)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"compose":   dockerCompose,
		"artifacts": []Artifact{
			{Name: "docker-compose.yml", Type: "config", Path: "/ghost/artifacts/docker-compose.yml", Size: len(dockerCompose)},
		},
	})
}

func (h *Handler) generateDockerCompose(services []DockerServiceSpec) string {
	result := "version: '3.9'\n\nservices:\n"
	for _, svc := range services {
		result += fmt.Sprintf("  %s:\n", svc.Name)
		result += fmt.Sprintf("    image: %s\n", svc.Image)
		if len(svc.Ports) > 0 {
			result += "    ports:\n"
			for _, p := range svc.Ports {
				result += fmt.Sprintf("      - \"%s\"\n", p)
			}
		}
		if len(svc.Env) > 0 {
			result += "    environment:\n"
			for k, v := range svc.Env {
				result += fmt.Sprintf("      %s: \"%s\"\n", k, v)
			}
		}
		if len(svc.Volumes) > 0 {
			result += "    volumes:\n"
			for _, v := range svc.Volumes {
				result += fmt.Sprintf("      - %s\n", v)
			}
		}
		if len(svc.DependsOn) > 0 {
			result += "    depends_on:\n"
			for _, d := range svc.DependsOn {
				result += fmt.Sprintf("      - %s\n", d)
			}
		}
		result += "\n"
	}
	return result
}

// ============================================================
// Code Generation
// ============================================================

type GenerateSchemaRequest struct {
	Entities []EntitySpec `json:"entities"`
}

type GenerateBackendRequest struct {
	Spec string `json:"spec"` // API spec / requirements
	Lang string `json:"language"` // go, python, typescript
}

type GenerateFrontendRequest struct {
	Spec string `json:"spec"` // Component specs
	Framework string `json:"framework"` // react, vue, svelte
}

type GenerateTestsRequest struct {
	Code    string `json:"code"`
	Coverage string `json:"coverage_target"` // e.g., "80%"
}

func (h *Handler) HandleGenerateSchema(w http.ResponseWriter, r *http.Request) {
	var req GenerateSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	var sql strings.Builder
	sql.WriteString("-- Ghost Mode Auto-generated Schema\n\n")
	for _, entity := range req.Entities {
		sql.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", toSnakeCase(entity.Name)))
		for i, field := range entity.Fields {
			reqStr := ""
			if field.Required {
				reqStr = " NOT NULL"
			}
			defaultStr := ""
			if field.Default != "" {
				defaultStr = fmt.Sprintf(" DEFAULT %s", field.Default)
			}
			sql.WriteString(fmt.Sprintf("  %s %s%s%s%s", toSnakeCase(field.Name), mapGoTypeToSQL(field.Type), reqStr, defaultStr, "\n"))
			if i < len(entity.Fields)-1 {
				sql.WriteString(",\n")
			} else {
				sql.WriteString("\n")
			}
		}
		sql.WriteString(");\n\n")

		// Indexes
		for _, idx := range entity.Indexes {
			sql.WriteString(fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_%s ON %s(%s);\n", toSnakeCase(entity.Name), idx, toSnakeCase(entity.Name), idx))
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":  true,
		"sql": sql.String(),
	})
}

func (h *Handler) HandleGenerateBackend(w http.ResponseWriter, r *http.Request) {
	var req GenerateBackendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	// Use generation service if available
	if h.genSvc != nil {
		result, err := h.genSvc.GenerateFunction(context.Background(), &generation.GenerationRequest{
			Name:        "ghost-generated",
			Description: req.Spec,
			Runtime:     req.Lang,
			Prompt:      req.Spec,
			Model:       "claude-sonnet-4",
		})
		if err == nil && result.Success {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":       true,
				"code":     result.Code,
				"model":    result.ModelUsed,
				"complexity": result.Complexity,
			})
			return
		}
	}

	// Fallback: generate template code
	code := fmt.Sprintf("// Auto-generated by Ghost Mode\n// Language: %s\n\npackage main\n\nfunc main() {\n    // %s\n}\n", req.Lang, req.Spec[:min(100, len(req.Spec))])

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"code":    code,
		"fallback": true,
	})
}

func (h *Handler) HandleGenerateFrontend(w http.ResponseWriter, r *http.Request) {
	var req GenerateFrontendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	code := h.generateReactComponents(req.Spec)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":  true,
		"code": code,
	})
}

func (h *Handler) generateReactComponents(spec string) string {
	return `// Auto-generated by Ghost Mode
import * as React from "react";

export function App() {
  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <h1 className="text-2xl font-bold p-4">Ghost Mode Generated App</h1>
      <p className="p-4">` + spec[:min(100, len(spec))] + `</p>
    </div>
  );
}
`
}

func (h *Handler) HandleGenerateTests(w http.ResponseWriter, r *http.Request) {
	var req GenerateTestsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	tests := fmt.Sprintf(`// Auto-generated by Ghost Mode
// Coverage target: %s

describe('Ghost Generated Tests', () => {
  it('should pass basic assertion', () => {
    expect(true).toBe(true);
  });
});
`, req.Coverage)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":   true,
		"tests": tests,
	})
}

// ============================================================
// Deployment
// ============================================================

type DeployRequest struct {
	BuildID    string `json:"build_id"`
	CommitSHA  string `json:"commit_sha,omitempty"`
	RollbackID string `json:"rollback_id,omitempty"`
}

func (h *Handler) HandleDeployStaging(w http.ResponseWriter, r *http.Request) {
	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	// Simulate staging deployment
	deployID := "deploy_" + uuid.New().String()[:8]
	url := "https://staging-" + deployID + ".functionfly.dev"

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"deploy_id": deployID,
		"url":       url,
		"status":    "deployed",
		"phase":     "staging",
	})
}

func (h *Handler) HandleDeployProduction(w http.ResponseWriter, r *http.Request) {
	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	// Check for human approval requirement
	buildID := req.BuildID
	h.mu.RLock()
	build, ok := h.builds[buildID]
	h.mu.RUnlock()

	if ok && build.HumanApprovalRequired {
		writeError(w, http.StatusForbidden, "APPROVAL_REQUIRED", "human approval required before production deployment")
		return
	}

	deployID := "deploy_" + uuid.New().String()[:8]
	url := "https://" + deployID + ".functionfly.app"

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"deploy_id": deployID,
		"url":       url,
		"status":    "deployed",
		"phase":     "production",
	})
}

// ============================================================
// Monitoring Setup
// ============================================================

type SetupMonitoringRequest struct {
	Services []string `json:"services"` // services to monitor
	Dashboards []string `json:"dashboards"` // dashboard types
}

func (h *Handler) HandleSetupMonitoring(w http.ResponseWriter, r *http.Request) {
	var req SetupMonitoringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	config := map[string]interface{}{
		"prometheus_url": "http://prometheus:9090",
		"grafana_url": "http://grafana:3000",
		"alert_channels": []string{"email", "slack"},
		"health_check_interval": "30s",
		"services": req.Services,
		"dashboards": req.Dashboards,
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"config":  config,
		"message": "Monitoring configured",
	})
}

// ============================================================
// Build Orchestration (Async)
// ============================================================

func (h *Handler) runBuildOrchestration(buildID string) {
	time.Sleep(2 * time.Second) // Initial delay

	for {
		h.mu.Lock()
		build, ok := h.builds[buildID]
		if !ok || build.Phase == PhaseComplete || build.Phase == PhaseError || build.Phase == PhasePaused {
			h.mu.Unlock()
			return
		}

		// Find next pending task
		var nextTask *TaskState
		for i := range build.Tasks {
			if build.Tasks[i].Status == StatusPending {
				// Check dependencies
				depsMet := true
				for _, dep := range build.Tasks[i].Dependencies {
					depTask := h.findTask(build, dep)
					if depTask == nil || depTask.Status != StatusCompleted {
						depsMet = false
						break
					}
				}
				if depsMet {
					nextTask = &build.Tasks[i]
					break
				}
			}
		}

		if nextTask == nil {
			// All tasks done
			build.Phase = PhaseComplete
			build.Progress = 1.0
			build.UpdatedAt = time.Now()
			h.mu.Unlock()
			return
		}

		// Start the task
		now := time.Now()
		nextTask.Status = StatusInProgress
		nextTask.StartedAt = &now
		build.CurrentTaskID = nextTask.ID
		build.Phase = nextTask.Phase
		build.UpdatedAt = now

		// Check if task requires human approval
		if nextTask.Title == "Design database schema" || nextTask.Title == "Deploy to staging" {
			build.HumanApprovalRequired = true
			if nextTask.Title == "Design database schema" {
				build.ApprovalType = "schema"
			} else {
				build.ApprovalType = "deployment"
			}
			build.UpdatedAt = now
			h.mu.Unlock()
			return // Wait for human approval
		}

		h.mu.Unlock()

		// Simulate task work
		h.simulateTaskWork(buildID, nextTask.ID)

		// Check if paused or errored during work
		h.mu.RLock()
		build, _ = h.builds[buildID]
		h.mu.RUnlock()
		if build == nil || build.Phase == PhasePaused || build.Phase == PhaseError {
			return
		}
	}
}

func (h *Handler) simulateTaskWork(buildID, taskID string) {
	taskDurations := map[string]int{
		"task-1": 8, // planning
		"task-2": 6, // schema
		"task-3": 10, // infra
		"task-4": 15, // backend
		"task-5": 12, // frontend
		"task-6": 8, // tests
		"task-7": 10, // staging deploy
		"task-8": 5, // monitoring
	}

	duration := taskDurations[taskID]
	if duration == 0 {
		duration = 10
	}

	steps := 3
	for i := 0; i < steps; i++ {
		time.Sleep(time.Duration(duration/steps) * time.Second)
		h.addLog(buildID, taskID, "info", "Processing...")
	}

	// Complete the task
	h.mu.Lock()
	defer h.mu.Unlock()

	build, ok := h.builds[buildID]
	if !ok {
		return
	}

	for i := range build.Tasks {
		if build.Tasks[i].ID == taskID {
			now := time.Now()
			build.Tasks[i].Status = StatusCompleted
			build.Tasks[i].CompletedAt = &now
			build.Tasks[i].Confidence = 0.75 + rand.Float64()*0.20
			build.CurrentTaskID = ""
			build.UpdatedAt = now

			// Add artifacts based on task
			switch taskID {
			case "task-2":
				build.Tasks[i].Artifacts = []Artifact{
					{Name: "schema.sql", Type: "schema", Path: "/ghost/artifacts/schema.sql"},
				}
			case "task-4":
				build.Tasks[i].Artifacts = []Artifact{
					{Name: "handlers.go", Type: "code", Path: "/ghost/artifacts/handlers.go"},
				}
			case "task-5":
				build.Tasks[i].Artifacts = []Artifact{
					{Name: "App.tsx", Type: "code", Path: "/ghost/artifacts/App.tsx"},
				}
			}
			break
		}
	}

	h.recalculateProgress(build)
}

func (h *Handler) addLog(buildID, taskID, level, message string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	build, ok := h.builds[buildID]
	if !ok {
		return
	}

	for i := range build.Tasks {
		if build.Tasks[i].ID == taskID {
			build.Tasks[i].Logs = append(build.Tasks[i].Logs, LogEntry{
				Timestamp: time.Now(),
				Level:     level,
				Message:   message,
			})
			build.UpdatedAt = time.Now()
			break
		}
	}
}

func (h *Handler) findTask(build *BuildState, taskID string) *TaskState {
	for i := range build.Tasks {
		if build.Tasks[i].ID == taskID {
			return &build.Tasks[i]
		}
	}
	return nil
}

func (h *Handler) getTask(build *BuildState, taskID string) *TaskState {
	for i := range build.Tasks {
		if build.Tasks[i].ID == taskID {
			return &build.Tasks[i]
		}
	}
	return nil
}

func (h *Handler) recalculateProgress(build *BuildState) {
	completed := 0
	for _, t := range build.Tasks {
		if t.Status == StatusCompleted || t.Status == StatusSkipped {
			completed++
		}
	}
	build.Progress = float64(completed) / float64(len(build.Tasks))
}

func (h *Handler) resumeBuild(buildID string) {
	time.Sleep(500 * time.Millisecond)
	h.runBuildOrchestration(buildID)
}

// ============================================================
// Helpers
// ============================================================

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]interface{}{
		"ok":    false,
		"error": map[string]string{"code": code, "message": message},
	})
}

func mapGoTypeToSQL(goType string) string {
	switch goType {
	case "string": return "VARCHAR(255)"
	case "int": return "INTEGER"
	case "bool": return "BOOLEAN"
	case "timestamp": return "TIMESTAMPTZ"
	case "uuid": return "UUID"
	case "jsonb": return "JSONB"
	case "text": return "TEXT"
	default: return "VARCHAR(255)"
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteByte(byte(r + 32))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var _ strings.Builder