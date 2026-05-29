package ghost

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

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
		"tasks": toFrontendTasks(build),
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
		"ok":   true,
		"task": h.getTask(build, taskID),
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
		Artifacts  []Artifact `json:"artifacts,omitempty"`
		Output     string     `json:"output,omitempty"`
		Confidence float64    `json:"confidence,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

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

	h.recalculateProgress(build)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

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
		"ok":    true,
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

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
