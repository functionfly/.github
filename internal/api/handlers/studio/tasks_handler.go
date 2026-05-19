package studio

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// TasksHandler handles studio task HTTP requests
type TasksHandler struct {
	taskRepo *TaskRepository
}

// NewTasksHandler creates a new tasks handler
func NewTasksHandler(taskRepo *TaskRepository) *TasksHandler {
	return &TasksHandler{taskRepo: taskRepo}
}

// HandleListTasks handles GET /v1/studio/tasks
func (h *TasksHandler) HandleListTasks(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	status := r.URL.Query().Get("status")
	assigneeID := r.URL.Query().Get("assignee_id")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	var statusPtr *TaskStatus
	if status != "" {
		s := TaskStatus(status)
		if !isValidTaskStatus(s) {
			writeJSONError(w, http.StatusBadRequest, "Invalid status")
			return
		}
		statusPtr = &s
	}

	var assigneePtr *string
	if assigneeID != "" {
		assigneePtr = &assigneeID
	}

	params := ListTasksParams{
		TenantID:   tenantID,
		Status:     statusPtr,
		AssigneeID: assigneePtr,
		Limit:      limit,
		Offset:     offset,
	}

	tasks, err := h.taskRepo.ListTasks(r.Context(), params)
	if err != nil {
		logrus.WithError(err).Warn("studio tasks: failed to list tasks")
		writeJSON(w, http.StatusOK, map[string]interface{}{"tasks": []StudioTask{}})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})
}

// HandleCreateTask handles POST /v1/studio/tasks
func (h *TasksHandler) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Title       string            `json:"title"`
		Description string            `json:"description"`
		Priority    TaskPriority      `json:"priority"`
		AssigneeID  *string           `json:"assignee_id"`
		Metadata    map[string]any    `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		writeJSONError(w, http.StatusBadRequest, "title is required")
		return
	}

	if req.Priority == "" {
		req.Priority = TaskPriorityMedium
	}
	if !isValidTaskPriority(req.Priority) {
		writeJSONError(w, http.StatusBadRequest, "Invalid priority")
		return
	}

	task := &StudioTask{
		TenantID:    tenantID,
		Title:       req.Title,
		Description: req.Description,
		Status:      TaskStatusTodo,
		Priority:    req.Priority,
		AssigneeID:  req.AssigneeID,
		CreatedBy:   userID,
		Metadata:    req.Metadata,
	}

	if err := h.taskRepo.CreateTask(r.Context(), task); err != nil {
		logrus.WithError(err).Error("studio tasks: failed to create task")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create task")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"task": task})
}

// HandleGetTask handles GET /v1/studio/tasks/{id}
func (h *TasksHandler) HandleGetTask(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	taskID := mux.Vars(r)["id"]
	if taskID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	task, err := h.taskRepo.GetTask(r.Context(), tenantID, taskID)
	if err != nil {
		logrus.WithError(err).Warn("studio tasks: failed to get task")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get task")
		return
	}
	if task == nil {
		writeJSONError(w, http.StatusNotFound, "Task not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"task": task})
}

// HandleUpdateTask handles PATCH /v1/studio/tasks/{id}
func (h *TasksHandler) HandleUpdateTask(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	taskID := mux.Vars(r)["id"]
	if taskID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	task, err := h.taskRepo.UpdateTask(r.Context(), tenantID, taskID, updates)
	if err != nil {
		logrus.WithError(err).Error("studio tasks: failed to update task")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update task")
		return
	}
	if task == nil {
		writeJSONError(w, http.StatusNotFound, "Task not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"task": task})
}

// HandleDeleteTask handles DELETE /v1/studio/tasks/{id}
func (h *TasksHandler) HandleDeleteTask(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	taskID := mux.Vars(r)["id"]
	if taskID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	if err := h.taskRepo.DeleteTask(r.Context(), tenantID, taskID); err != nil {
		logrus.WithError(err).Error("studio tasks: failed to delete task")
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete task")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Task deleted"})
}

// HandleAssignTask handles POST /v1/studio/tasks/{id}/assign
func (h *TasksHandler) HandleAssignTask(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	taskID := mux.Vars(r)["id"]
	if taskID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req struct {
		AssigneeID string `json:"assignee_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AssigneeID == "" {
		writeJSONError(w, http.StatusBadRequest, "assignee_id is required")
		return
	}

	if err := h.taskRepo.AssignTask(r.Context(), tenantID, taskID, req.AssigneeID); err != nil {
		logrus.WithError(err).Error("studio tasks: failed to assign task")
		writeJSONError(w, http.StatusInternalServerError, "Failed to assign task")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Task assigned"})
}