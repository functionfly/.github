package employee

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListTasks(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListTasksOpts{
		Limit:  20,
		Offset: 0,
	}
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			opts.Limit = n
		}
	}
	if o := q.Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			opts.Offset = n
		}
	}
	if s := q.Get("project_id"); s != "" {
		if pid, err := uuid.Parse(s); err == nil {
			opts.ProjectID = &pid
		}
	}
	if s := q.Get("assignee_id"); s != "" {
		if aid, err := uuid.Parse(s); err == nil {
			opts.AssigneeID = &aid
		}
	}
	if s := q.Get("status"); s != "" {
		opts.Status = &s
	}
	if s := q.Get("priority"); s != "" {
		opts.Priority = &s
	}

	tasks, total, err := h.repo.ListTasks(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list tasks")
		apierror.WriteError(w, apierror.NewInternal("Failed to list tasks"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tasks":  tasks,
		"total":  total,
		"limit":  opts.Limit,
		"offset": opts.Offset,
	})
}

func (h *Handler) HandleGetTask(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid task ID"))
		return
	}

	task, err := h.repo.GetTaskByID(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get task")
		apierror.WriteError(w, apierror.NewInternal("Failed to get task"))
		return
	}
	if task == nil {
		apierror.WriteError(w, apierror.NewNotFound("Task not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task": task,
	})
}

type createTaskRequest struct {
	ProjectID   string `json:"project_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	AssigneeID  string `json:"assignee_id,omitempty"`
	Priority    string `json:"priority,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
}

func (h *Handler) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Title == "" || req.ProjectID == "" {
		apierror.WriteError(w, apierror.NewBadRequest("project_id and title are required"))
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid project ID"))
		return
	}

	now := time.Now()

	// Resolve employee ID from user ID
	emp, _ := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	reporterID := claims.UserID
	if emp != nil {
		reporterID = emp.ID
	}

	task := &types.Task{
		ID:         uuid.New(),
		ProjectID:  projectID,
		TenantID:   claims.TenantID,
		Title:      req.Title,
		ReporterID: reporterID,
		Status:     "todo",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if req.Description != "" {
		task.Description = &req.Description
	}
	if req.AssigneeID != "" {
		if aid, err := uuid.Parse(req.AssigneeID); err == nil {
			task.AssigneeID = &aid
		}
	}
	if req.Priority != "" {
		task.Priority = req.Priority
	} else {
		task.Priority = "medium"
	}
	if req.DueDate != "" {
		if t, err := time.Parse("2006-01-02", req.DueDate); err == nil {
			task.DueDate = &t
		}
	}

	created, err := h.repo.CreateTask(r.Context(), task)
	if err != nil {
		h.log.WithError(err).Error("Failed to create task")
		apierror.WriteError(w, apierror.NewInternal("Failed to create task"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"task":    created,
	})
}

type updateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	Priority    *string `json:"priority,omitempty"`
	AssigneeID  *string `json:"assignee_id,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
}

func (h *Handler) HandleUpdateTask(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid task ID"))
		return
	}

	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.AssigneeID != nil {
		if aid, err := uuid.Parse(*req.AssigneeID); err == nil {
			updates["assignee_id"] = aid
		}
	}
	if req.DueDate != nil {
		if t, err := time.Parse("2006-01-02", *req.DueDate); err == nil {
			updates["due_date"] = t
		}
	}
	updates["updated_at"] = time.Now()

	if len(updates) == 1 {
		apierror.WriteError(w, apierror.NewBadRequest("No fields to update"))
		return
	}

	if err := h.repo.UpdateTask(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to update task")
		apierror.WriteError(w, apierror.NewInternal("Failed to update task"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

type assignTaskRequest struct {
	AssigneeID string `json:"assignee_id"`
}

func (h *Handler) HandleAssignTask(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid task ID"))
		return
	}

	var req assignTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.AssigneeID == "" {
		apierror.WriteError(w, apierror.NewBadRequest("assignee_id is required"))
		return
	}

	assigneeID, err := uuid.Parse(req.AssigneeID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid assignee ID"))
		return
	}

	updates := map[string]interface{}{
		"assignee_id": assigneeID,
		"updated_at":  time.Now(),
	}

	if err := h.repo.UpdateTask(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to assign task")
		apierror.WriteError(w, apierror.NewInternal("Failed to assign task"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) HandleListTaskComments(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid task ID"))
		return
	}

	comments, err := h.repo.GetTaskComments(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get task comments")
		apierror.WriteError(w, apierror.NewInternal("Failed to get task comments"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"comments": comments,
	})
}

type createTaskCommentRequest struct {
	Body string `json:"body"`
}

func (h *Handler) HandleCreateTaskComment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid task ID"))
		return
	}

	var req createTaskCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Body == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Comment body is required"))
		return
	}

	// Resolve employee ID from user ID
	emp, _ := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	authorID := claims.UserID
	if emp != nil {
		authorID = emp.ID
	}

	comment := &types.TaskComment{
		TaskID:   id,
		AuthorID: authorID,
		Body:     req.Body,
	}

	created, err := h.repo.CreateTaskComment(r.Context(), comment)
	if err != nil {
		h.log.WithError(err).Error("Failed to create comment")
		apierror.WriteError(w, apierror.NewInternal("Failed to create comment"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"comment": created,
	})
}
