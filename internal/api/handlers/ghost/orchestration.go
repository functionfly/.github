package ghost

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/functionfly/functionfly/internal/agent/deployment"
	"github.com/functionfly/functionfly/internal/agent/generation"
	"github.com/sirupsen/logrus"
)

func (h *Handler) runBuildOrchestration(buildID string) {
	if !h.secureContext {
		h.runBuildOrchestrationWithSimulation(buildID)
		return
	}

	h.runBuildOrchestrationSecure(buildID)
}

func (h *Handler) runBuildOrchestrationSecure(buildID string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(h.envContext.RuntimeLimits.TimeoutSeconds)*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			h.failBuild(buildID, "Build timed out")
			return
		default:
		}

		h.mu.Lock()
		build, ok := h.builds[buildID]
		if !ok || build.Phase == PhaseComplete || build.Phase == PhaseError || build.Phase == PhasePaused {
			h.mu.Unlock()
			return
		}

		nextTask := h.selectNextTask(build)
		if nextTask == nil {
			build.Phase = PhaseComplete
			build.Progress = 1.0
			build.UpdatedAt = time.Now()
			h.mu.Unlock()

			logrus.WithFields(logrus.Fields{
				"build_id": buildID,
				"goal":     build.Goal,
			}).Info("Ghost Mode build completed")

			return
		}

		now := time.Now()
		nextTask.Status = StatusInProgress
		nextTask.StartedAt = &now
		nextTask.AgentID = h.envContext.AgentID
		build.CurrentTaskID = nextTask.ID
		build.Phase = nextTask.Phase
		build.UpdatedAt = now
		h.mu.Unlock()

		if requiresApproval(nextTask.Title) {
			h.mu.Lock()
			build.HumanApprovalRequired = true
			build.ApprovalType = getApprovalType(nextTask.Title)
			build.UpdatedAt = time.Now()
			h.mu.Unlock()

			logrus.WithFields(logrus.Fields{
				"build_id":     buildID,
				"task":         nextTask.Title,
				"approvalType": build.ApprovalType,
			}).Info("Ghost Mode awaiting human approval")

			return
		}

		h.executeTaskSecure(ctx, buildID, nextTask)

		h.mu.RLock()
		build = h.builds[buildID]
		h.mu.RUnlock()
		if build == nil || build.Phase == PhasePaused || build.Phase == PhaseError {
			return
		}
	}
}

func (h *Handler) executeTaskSecure(ctx context.Context, buildID string, task *TaskState) {
	startTime := time.Now()

	h.addLog(buildID, task.ID, "info", "Starting task execution via Ghost Mode")

	output, err := h.executeTaskWithLLM(ctx, task)
	duration := time.Since(startTime)

	h.mu.Lock()
	defer h.mu.Unlock()

	build, ok := h.builds[buildID]
	if !ok {
		return
	}

	for i := range build.Tasks {
		if build.Tasks[i].ID == task.ID {
			if err != nil {
				build.Tasks[i].Status = StatusFailed
				build.Tasks[i].CompletedAt = timePtr(time.Now())
				build.Tasks[i].Logs = append(build.Tasks[i].Logs, LogEntry{
					Timestamp: time.Now(),
					Level:     "error",
					Message:   err.Error(),
				})
				build.Phase = PhaseError
				build.Error = err.Error()
				build.CurrentTaskID = ""
				build.UpdatedAt = time.Now()

				logrus.WithFields(logrus.Fields{
					"build_id": buildID,
					"task":     task.Title,
					"error":    err.Error(),
				}).Error("Ghost Mode task failed")

				return
			}

			build.Tasks[i].Status = StatusCompleted
			build.Tasks[i].CompletedAt = timePtr(time.Now())
			build.Tasks[i].DurationMs = int(duration.Milliseconds())
			build.Tasks[i].Confidence = calculateConfidence(output)
			build.Tasks[i].LLMOutput = output

			if artifacts := extractArtifacts(task.Title, output); len(artifacts) > 0 {
				build.Tasks[i].Artifacts = artifacts
				build.Artifacts = append(build.Artifacts, artifacts...)
			}

			build.CurrentTaskID = ""
			build.UpdatedAt = time.Now()

			logrus.WithFields(logrus.Fields{
				"build_id":  buildID,
				"task":      task.Title,
				"duration":  duration.Milliseconds(),
				"confidence": build.Tasks[i].Confidence,
			}).Info("Ghost Mode task completed")

			h.recalculateProgress(build)
			return
		}
	}
}

func (h *Handler) executeTaskWithLLM(ctx context.Context, task *TaskState) (string, error) {
	switch {
	case containsAny(task.Title, "architecture", "plan", "design"):
		return h.generateArchitectureTask(ctx, task)
	case containsAny(task.Title, "database", "schema", "sql"):
		return h.generateDatabaseTask(ctx, task)
	case containsAny(task.Title, "backend", "api", "server"):
		return h.generateBackendTask(ctx, task)
	case containsAny(task.Title, "frontend", "ui", "react"):
		return h.generateFrontendTask(ctx, task)
	case containsAny(task.Title, "test", "testing"):
		return h.generateTestsTask(ctx, task)
	case containsAny(task.Title, "deploy", "docker", "container"):
		return h.generateDeploymentTask(ctx, task)
	default:
		return h.generateGenericTask(ctx, task)
	}
}

func (h *Handler) generateArchitectureTask(ctx context.Context, task *TaskState) (string, error) {
	if h.genSvc == nil {
		return "", nil
	}

	req := &generation.GenerationRequest{
		AgentID:     h.envContext.AgentID,
		Name:        "ghost-architecture-task",
		Description: task.Description,
		Runtime:     "go",
		Prompt:      task.Description,
		Model:       "anthropic/claude-3-opus",
		Tags:        []string{"ghost-mode", "architecture"},
	}

	result, err := h.genSvc.GenerateFunction(ctx, req)
	if err != nil || !result.Success {
		return "", err
	}

	return result.Code, nil
}

func (h *Handler) generateDatabaseTask(ctx context.Context, task *TaskState) (string, error) {
	return generateSchemaFromDescription(task.Description), nil
}

func (h *Handler) generateBackendTask(ctx context.Context, task *TaskState) (string, error) {
	if h.genSvc != nil {
		req := &generation.GenerationRequest{
			AgentID:     h.envContext.AgentID,
			Name:        "ghost-backend-task",
			Description: task.Description,
			Runtime:     "go",
			Prompt:      buildBackendPrompt(task.Description, "go"),
			Model:       "inception/mercury-2",
			Tags:        []string{"ghost-mode", "backend"},
		}

		result, err := h.genSvc.GenerateFunction(ctx, req)
		if err == nil && result.Success {
			return result.Code, nil
		}
	}

	if h.deployGen != nil {
		deployReq := &deployment.GenerationRequest{
			AgentID: h.envContext.AgentID,
			FunctionSpec: deployment.FunctionSpec{
				Name:        "ghost-backend",
				Description: task.Description,
				Prompt:      task.Description,
			},
			Language: "go",
			Runtime:  "go1.21",
		}

		generated, err := h.deployGen.Generate(ctx, deployReq)
		if err == nil && generated.Status == deployment.GenerationStatusSuccess {
			return generated.GeneratedCode, nil
		}
	}

	return "", nil
}

func (h *Handler) generateFrontendTask(ctx context.Context, task *TaskState) (string, error) {
	if h.genSvc == nil {
		return "", nil
	}

	req := &generation.GenerationRequest{
		AgentID:     h.envContext.AgentID,
		Name:        "ghost-frontend-task",
		Description: task.Description,
		Runtime:     "typescript",
		Prompt:      buildFrontendPrompt(task.Description, "react"),
		Model:       "anthropic/claude-3-haiku",
		Tags:        []string{"ghost-mode", "frontend"},
	}

	result, err := h.genSvc.GenerateFunction(ctx, req)
	if err != nil || !result.Success {
		return "", err
	}

	return result.Code, nil
}

func (h *Handler) generateTestsTask(ctx context.Context, task *TaskState) (string, error) {
	if h.genSvc == nil {
		return "", nil
	}

	req := &generation.GenerationRequest{
		AgentID:     h.envContext.AgentID,
		Name:        "ghost-tests-task",
		Description: task.Description,
		Runtime:     "python3.11",
		Prompt:      task.Description,
		Model:       "inception/mercury-2",
		Tags:        []string{"ghost-mode", "testing"},
	}

	result, err := h.genSvc.GenerateFunction(ctx, req)
	if err != nil || !result.Success {
		return "", err
	}

	return result.Code, nil
}

func (h *Handler) generateDeploymentTask(ctx context.Context, task *TaskState) (string, error) {
	return generateDockerCompose(task.Description), nil
}

func (h *Handler) generateGenericTask(ctx context.Context, task *TaskState) (string, error) {
	if h.genSvc == nil {
		return "", nil
	}

	req := &generation.GenerationRequest{
		AgentID:     h.envContext.AgentID,
		Name:        "ghost-generic-task",
		Description: task.Description,
		Runtime:     "python3.11",
		Prompt:      task.Description,
		Model:       "inception/mercury-2",
		Tags:        []string{"ghost-mode"},
	}

	result, err := h.genSvc.GenerateFunction(ctx, req)
	if err != nil || !result.Success {
		return "", err
	}

	return result.Code, nil
}

func (h *Handler) runBuildOrchestrationWithSimulation(buildID string) {
	time.Sleep(2 * time.Second)

	for {
		h.mu.Lock()
		build, ok := h.builds[buildID]
		if !ok || build.Phase == PhaseComplete || build.Phase == PhaseError || build.Phase == PhasePaused {
			h.mu.Unlock()
			return
		}

		nextTask := h.selectNextTask(build)
		if nextTask == nil {
			build.Phase = PhaseComplete
			build.Progress = 1.0
			build.UpdatedAt = time.Now()
			h.mu.Unlock()
			return
		}

		now := time.Now()
		nextTask.Status = StatusInProgress
		nextTask.StartedAt = &now
		build.CurrentTaskID = nextTask.ID
		build.Phase = nextTask.Phase
		build.UpdatedAt = now

		if requiresApproval(nextTask.Title) {
			build.HumanApprovalRequired = true
			build.ApprovalType = getApprovalType(nextTask.Title)
			build.UpdatedAt = now
			h.mu.Unlock()
			return
		}

		h.mu.Unlock()

		h.simulateTaskWork(buildID, nextTask.ID)

		h.mu.RLock()
		build = h.builds[buildID]
		h.mu.RUnlock()
		if build == nil || build.Phase == PhasePaused || build.Phase == PhaseError {
			return
		}
	}
}

func (h *Handler) selectNextTask(build *BuildState) *TaskState {
	for i := range build.Tasks {
		if build.Tasks[i].Status == StatusPending {
			depsMet := true
			for _, dep := range build.Tasks[i].Dependencies {
				depTask := h.findTask(build, dep)
				if depTask == nil || depTask.Status != StatusCompleted {
					depsMet = false
					break
				}
			}
			if depsMet {
				return &build.Tasks[i]
			}
		}
	}
	return nil
}

func (h *Handler) simulateTaskWork(buildID, taskID string) {
	taskDurations := map[string]int{
		"task-1": 8,
		"task-2": 6,
		"task-3": 10,
		"task-4": 15,
		"task-5": 12,
		"task-6": 8,
		"task-7": 10,
		"task-8": 5,
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

func (h *Handler) resumeBuild(buildID string) {
	time.Sleep(500 * time.Millisecond)
	h.runBuildOrchestration(buildID)
}

func (h *Handler) failBuild(buildID, errorMsg string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	build, ok := h.builds[buildID]
	if !ok {
		return
	}

	build.Phase = PhaseError
	build.Error = errorMsg
	build.UpdatedAt = time.Now()

	logrus.WithFields(logrus.Fields{
		"build_id": buildID,
		"error":    errorMsg,
	}).Error("Ghost Mode build failed")
}

func requiresApproval(taskTitle string) bool {
	approvalTasks := map[string]bool{
		"Design database schema": true,
		"Deploy to staging":       true,
		"Deploy to production":    true,
	}
	return approvalTasks[taskTitle]
}

func getApprovalType(taskTitle string) string {
	switch taskTitle {
	case "Design database schema":
		return "schema"
	case "Deploy to staging", "Deploy to production":
		return "deployment"
	default:
		return "general"
	}
}

func calculateConfidence(output string) float64 {
	if output == "" {
		return 0.5
	}
	base := 0.75
	if len(output) > 100 {
		base += 0.05
	}
	if len(output) > 500 {
		base += 0.05
	}
	if containsString(output, "error") || containsString(output, "TODO") {
		base -= 0.1
	}
	if containsString(output, "test") || containsString(output, "Test") {
		base += 0.05
	}

	min := 0.6
	max := 0.98
	if base < min {
		base = min
	}
	if base > max {
		base = max
	}
	return base
}

func extractArtifacts(taskTitle, output string) []Artifact {
	var artifacts []Artifact

	switch {
	case containsAny(taskTitle, "database", "schema"):
		artifacts = append(artifacts, Artifact{
			Name: "schema.sql",
			Type: "schema",
			Path: "/ghost/artifacts/schema.sql",
			Size: len(output),
		})
	case containsAny(taskTitle, "backend", "api"):
		artifacts = append(artifacts, Artifact{
			Name: "handlers.go",
			Type: "code",
			Path: "/ghost/artifacts/handlers.go",
			Size: len(output),
		})
	case containsAny(taskTitle, "frontend", "ui"):
		artifacts = append(artifacts, Artifact{
			Name: "App.tsx",
			Type: "code",
			Path: "/ghost/artifacts/App.tsx",
			Size: len(output),
		})
	case containsAny(taskTitle, "test"):
		artifacts = append(artifacts, Artifact{
			Name: "test_suite.go",
			Type: "test",
			Path: "/ghost/artifacts/test_suite.go",
			Size: len(output),
		})
	}

	return artifacts
}

func generateSchemaFromDescription(description string) string {
	return fmt.Sprintf(`-- Ghost Mode Auto-generated Schema
-- Generated from: %s

CREATE TABLE IF NOT EXISTS resources (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  data JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_resources_name ON resources(name);
CREATE INDEX IF NOT EXISTS idx_resources_created_at ON resources(created_at);
`, description[:min(100, len(description))])
}

func generateDockerCompose(description string) string {
	return `version: '3.9'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - NODE_ENV=production
      - DATABASE_URL=${DATABASE_URL}
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:17-alpine
    environment:
      - POSTGRES_USER=ghost
      - POSTGRES_DB=ghostdb
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
`
}

func containsString(s, substr string) bool {
	if len(s) == 0 || len(substr) == 0 {
		return false
	}
	sLower := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		sLower[i] = c
	}
	substrLower := make([]byte, len(substr))
	for i := 0; i < len(substr); i++ {
		c := substr[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		substrLower[i] = c
	}

	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		match := true
		for j := 0; j < len(substrLower); j++ {
			if sLower[i+j] != substrLower[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if containsString(s, substr) {
			return true
		}
	}
	return false
}

func timePtr(t time.Time) *time.Time {
	return &t
}