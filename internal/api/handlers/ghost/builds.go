package ghost

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (h *Handler) HandleCreateBuild(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	envCtx := h.buildEnvironmentContext(claims)

	var req CreateBuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.Goal == "" {
		req.Goal = req.Name
	}
	if req.Goal == "" {
		writeError(w, http.StatusBadRequest, "MISSING_GOAL", "goal or name is required")
		return
	}

	if !IsGhostModeProduction() && IsGhostModeBeta() {
		logrus.WithFields(logrus.Fields{
			"tenant":  envCtx.TenantID,
			"user":    envCtx.UserID,
			"status":  "beta",
			"message": "Ghost Mode running in beta mode",
		}).Warn("Ghost Mode beta build created")
	}

	buildID := "build_" + uuid.New().String()[:8]
	tasks := h.generateInitialTasks(req.Goal, envCtx)

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

	logrus.WithFields(logrus.Fields{
		"build_id": buildID,
		"tenant":   envCtx.TenantID,
		"user":     envCtx.UserID,
		"goal":     req.Goal,
		"secure":   h.secureContext,
	}).Info("Ghost Mode build created")

	go h.runBuildOrchestration(buildID)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"ok":    true,
		"build": toFrontendBuild(build),
	})
}

func (h *Handler) generateInitialTasks(goal string, envCtx EnvironmentContext) []TaskState {
	domain := detectDomain(goal)

	tasks := []TaskState{
		{
			ID:           "task-1",
			Title:        "Analyze requirements and plan architecture",
			Description:  "Parse natural language goal, identify components, design system architecture",
			Status:       StatusPending,
			Phase:        PhasePlanning,
			Dependencies: []string{},
			AgentID:      envCtx.AgentID,
		},
		{
			ID:           "task-2",
			Title:        "Design database schema",
			Description:  "Generate PostgreSQL schema from data requirements using " + envCtx.Environment + " environment",
			Status:       StatusPending,
			Phase:        PhasePlanning,
			Dependencies: []string{"task-1"},
			AgentID:      envCtx.AgentID,
		},
		{
			ID:           "task-3",
			Title:        "Provision infrastructure",
			Description:  "Create Docker containers, networking, and cloud resources in " + envCtx.Region,
			Status:       StatusPending,
			Phase:        PhaseProvisioning,
			Dependencies: []string{"task-2"},
			AgentID:      envCtx.AgentID,
		},
		{
			ID:           "task-4",
			Title:        "Generate backend code",
			Description:  "Write API handlers, business logic, and database integration",
			Status:       StatusPending,
			Phase:        PhaseBuilding,
			Dependencies: []string{"task-3"},
			AgentID:      envCtx.AgentID,
		},
		{
			ID:           "task-5",
			Title:        "Generate frontend code",
			Description:  "Build React components, pages, and API integration",
			Status:       StatusPending,
			Phase:        PhaseBuilding,
			Dependencies: []string{"task-4"},
			AgentID:      envCtx.AgentID,
		},
		{
			ID:           "task-6",
			Title:        "Write unit tests",
			Description:  "Create test suites with 80% code coverage target",
			Status:       StatusPending,
			Phase:        PhaseBuilding,
			Dependencies: []string{"task-4", "task-5"},
			AgentID:      envCtx.AgentID,
		},
	}

	if domain == "e-commerce" || containsKeyword(goal, "shop", "store", "payment") {
		tasks = append(tasks, TaskState{
			ID:           "task-7",
			Title:        "Generate payment integration",
			Description:  "Integrate Stripe payment processing",
			Status:       StatusPending,
			Phase:        PhaseBuilding,
			Dependencies: []string{"task-5"},
			AgentID:      envCtx.AgentID,
		})
	}

	tasks = append(tasks,
		TaskState{
			ID:           "task-8",
			Title:        "Deploy to staging",
			Description:  "Blue-green deployment to staging environment",
			Status:       StatusPending,
			Phase:        PhaseDeploying,
			Dependencies: []string{"task-6"},
			AgentID:      envCtx.AgentID,
		},
		TaskState{
			ID:           "task-9",
			Title:        "Setup monitoring and alerts",
			Description:  "Configure dashboards, health checks, and alerting",
			Status:       StatusPending,
			Phase:        PhaseMonitoring,
			Dependencies: []string{"task-8"},
			AgentID:      envCtx.AgentID,
		},
	)

	return tasks
}

func detectDomain(goal string) string {
	if containsKeyword(goal, "e-commerce", "shop", "store", "cart", "payment", "stripe") {
		return "e-commerce"
	}
	if containsKeyword(goal, "saas", "dashboard", "analytics", "metrics") {
		return "saas"
	}
	if containsKeyword(goal, "api", "rest", "backend", "service") {
		return "api"
	}
	return "general"
}

func (h *Handler) HandleListBuilds(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	envCtx := h.buildEnvironmentContext(claims)

	h.mu.RLock()
	var builds []map[string]interface{}
	for _, b := range h.builds {
		builds = append(builds, toFrontendBuild(b))
	}
	h.mu.RUnlock()

	logrus.WithFields(logrus.Fields{
		"tenant": envCtx.TenantID,
		"user":   envCtx.UserID,
		"count":  len(builds),
	}).Debug("Ghost Mode builds listed")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
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

	if phase, ok := updates["phase"].(string); ok {
		validPhases := map[GhostPhase]bool{
			PhasePlanning:   true,
			PhaseProvisioning: true,
			PhaseBuilding:   true,
			PhaseDeploying:  true,
			PhaseMonitoring: true,
			PhaseComplete:   true,
			PhaseError:      true,
			PhasePaused:     true,
		}
		if validPhases[GhostPhase(phase)] {
			build.Phase = GhostPhase(phase)
		}
	}
	if progress, ok := updates["progress"].(float64); ok {
		if progress >= 0 && progress <= 1 {
			build.Progress = progress
		}
	}
	if errMsg, ok := updates["error"].(string); ok {
		build.Error = errMsg
	}
	if agentID, ok := updates["agent_id"].(string); ok {
		for i := range build.Tasks {
			if build.Tasks[i].AgentID == "" {
				build.Tasks[i].AgentID = agentID
			}
		}
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

	logrus.WithFields(logrus.Fields{
		"build_id": buildID,
	}).Info("Ghost Mode build deleted")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "build deleted",
	})
}

func (h *Handler) HandleApproval(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]
	if buildID == "" {
		buildID = r.URL.Query().Get("build_id")
	}

	var req ApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if buildID == "" {
		buildID = mux.Vars(r)["id"]
	}

	claims := middleware.GetUserFromContext(r)
	envCtx := h.buildEnvironmentContext(claims)

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

	validDecisions := map[string]bool{"approve": true, "reject": true, "revision": true}
	if !validDecisions[req.Decision] {
		writeError(w, http.StatusBadRequest, "INVALID_DECISION", "decision must be approve, reject, or revision")
		return
	}

	switch req.Decision {
	case "approve":
		build.HumanApprovalRequired = false
		build.ApprovalType = ""
		build.UpdatedAt = time.Now()

		logrus.WithFields(logrus.Fields{
			"build_id": buildID,
			"user":     envCtx.UserID,
			"tenant":   envCtx.TenantID,
		}).Info("Ghost Mode build approved")

		go h.resumeBuild(buildID)
	case "reject":
		build.Phase = PhaseError
		build.Error = "Approval rejected: " + req.Notes
		build.UpdatedAt = time.Now()

		logrus.WithFields(logrus.Fields{
			"build_id": buildID,
			"user":     envCtx.UserID,
			"tenant":   envCtx.TenantID,
			"notes":    req.Notes,
		}).Warn("Ghost Mode build rejected")
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

	logrus.WithFields(logrus.Fields{
		"build_id": buildID,
	}).Info("Ghost Mode build paused")

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

	if len(build.Tasks) > 0 {
		for i := range build.Tasks {
			if build.Tasks[i].Status == StatusInProgress {
				build.Tasks[i].Status = StatusPending
				build.Tasks[i].StartedAt = nil
				break
			}
		}
	}

	build.Phase = GhostPhase(build.Tasks[0].Phase)
	build.UpdatedAt = time.Now()

	logrus.WithFields(logrus.Fields{
		"build_id": buildID,
	}).Info("Ghost Mode build resumed")

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

	logrus.WithFields(logrus.Fields{
		"build_id": buildID,
	}).Info("Ghost Mode build cancelled")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"phase": build.Phase,
	})
}