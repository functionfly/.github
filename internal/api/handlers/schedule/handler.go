package schedule

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/scheduler"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// Handler contains schedule management handlers
type Handler struct {
	scheduler *scheduler.FunctionScheduler
	repo      storage.Repository
	executor  ScheduleExecutor
}

// ScheduleExecutor defines the interface for triggering function executions
type ScheduleExecutor interface {
	// ExecuteFunction triggers a function execution and returns the result
	ExecuteFunction(ctx context.Context, functionID uuid.UUID, input []byte) ([]byte, error)
}

// NewHandler creates a new schedule handler
func NewHandler(sched *scheduler.FunctionScheduler, repo storage.Repository) *Handler {
	return &Handler{
		scheduler: sched,
		repo:      repo,
	}
}

// RegisterExecutor registers an executor for the scheduler
func (h *Handler) RegisterExecutor(executor ScheduleExecutor) {
	h.executor = executor
}

// HandleCreateSchedule handles POST /v1/functions/{id}/schedule
func (h *Handler) HandleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	// Check if function exists and belongs to user
	function, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if function.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	var req CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Validate cron expression
	if err := scheduler.ValidateCronExpression(req.Cron); err != nil {
		logrus.WithError(err).WithField("cron", req.Cron).Info("schedule: invalid cron expression")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid cron expression"))
		return
	}

	// Set defaults
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}

	config := &scheduler.ScheduleConfig{
		Cron:        req.Cron,
		Timezone:    req.Timezone,
		Enabled:     true,
		RunOnDeploy: req.RunOnDeploy,
	}

	// Add schedule to scheduler
	if err := h.scheduler.AddSchedule(r.Context(), functionID, config); err != nil {
		logrus.WithError(err).Error("Failed to create schedule")
		apierror.WriteError(w, apierror.NewInternal("Failed to create schedule"))
		return
	}

	// Update function in storage
	nextRun, _ := scheduler.GetNextRunTime(req.Cron, req.Timezone)
	updates := map[string]interface{}{
		"schedule": storage.ScheduleConfig{
			Cron:       req.Cron,
			Timezone:   req.Timezone,
			Enabled:    true,
			NextRun:    nextRun,
			RunOnDeploy: req.RunOnDeploy,
		},
	}

	_, err = h.repo.UpdateFunction(r.Context(), functionID, updates)
	if err != nil {
		logrus.WithError(err).Error("Failed to update function schedule")
		// Still return success since scheduler is running
	}

	// Get human-readable description
	description, _ := scheduler.GetHumanReadableSchedule(req.Cron)

	response := CreateScheduleResponse{
		FunctionID:  functionID,
		Cron:        req.Cron,
		Timezone:    req.Timezone,
		Enabled:    true,
		NextRun:    nextRun,
		Description: description,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// HandleGetSchedule handles GET /v1/functions/{id}/schedule
func (h *Handler) HandleGetSchedule(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	// Check if function exists and belongs to user
	function, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if function.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	schedule := function.Schedule
	if schedule == nil {
		apierror.WriteError(w, apierror.NewNotFound("No schedule found"))
		return
	}

	// Get human-readable description
	description, _ := scheduler.GetHumanReadableSchedule(schedule.Cron)

	response := GetScheduleResponse{
		FunctionID:  functionID,
		Cron:        schedule.Cron,
		Timezone:    schedule.Timezone,
		Enabled:     schedule.Enabled,
		LastRun:     schedule.LastRun,
		NextRun:     schedule.NextRun,
		Description: description,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleUpdateSchedule handles PUT /v1/functions/{id}/schedule
func (h *Handler) HandleUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	// Check if function exists and belongs to user
	function, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if function.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	var req UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	currentSchedule := function.Schedule
	if currentSchedule == nil {
		apierror.WriteError(w, apierror.NewNotFound("No schedule found"))
		return
	}

	// Update cron if provided
	if req.Cron != "" {
		if err := scheduler.ValidateCronExpression(req.Cron); err != nil {
			logrus.WithError(err).WithField("cron", req.Cron).Info("schedule: invalid cron expression on update")
			apierror.WriteError(w, apierror.NewBadRequest("Invalid cron expression"))
			return
		}
		currentSchedule.Cron = req.Cron
	}

	// Update timezone if provided
	if req.Timezone != "" {
		currentSchedule.Timezone = req.Timezone
	}

	// Update enabled status
	if req.Enabled != nil {
		currentSchedule.Enabled = *req.Enabled
	}

	// Update the scheduler
	if currentSchedule.Enabled {
		if err := h.scheduler.AddSchedule(r.Context(), functionID, &scheduler.ScheduleConfig{
			Cron:      currentSchedule.Cron,
			Timezone:  currentSchedule.Timezone,
			Enabled:   currentSchedule.Enabled,
		}); err != nil {
			logrus.WithError(err).Error("Failed to update schedule")
			apierror.WriteError(w, apierror.NewInternal("Failed to update schedule"))
			return
		}
	} else {
		if err := h.scheduler.RemoveSchedule(r.Context(), functionID); err != nil {
			logrus.WithError(err).Error("Failed to remove schedule")
		}
	}

	// Update function in storage
	nextRun, _ := scheduler.GetNextRunTime(currentSchedule.Cron, currentSchedule.Timezone)
	updates := map[string]interface{}{
		"schedule": storage.ScheduleConfig{
			Cron:      currentSchedule.Cron,
			Timezone:  currentSchedule.Timezone,
			Enabled:   currentSchedule.Enabled,
			NextRun:   nextRun,
		},
	}

	_, err = h.repo.UpdateFunction(r.Context(), functionID, updates)
	if err != nil {
		logrus.WithError(err).Error("Failed to update function schedule")
	}

	// Get human-readable description
	description, _ := scheduler.GetHumanReadableSchedule(currentSchedule.Cron)

	response := UpdateScheduleResponse{
		FunctionID:  functionID,
		Cron:        currentSchedule.Cron,
		Timezone:    currentSchedule.Timezone,
		Enabled:     currentSchedule.Enabled,
		NextRun:     nextRun,
		Description: description,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleDeleteSchedule handles DELETE /v1/functions/{id}/schedule
func (h *Handler) HandleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	// Check if function exists and belongs to user
	function, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if function.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	// Remove from scheduler
	if err := h.scheduler.RemoveSchedule(r.Context(), functionID); err != nil {
		logrus.WithError(err).Error("Failed to remove schedule")
	}

	// Update function in storage to remove schedule
	updates := map[string]interface{}{
		"schedule": nil,
	}

	_, err = h.repo.UpdateFunction(r.Context(), functionID, updates)
	if err != nil {
		logrus.WithError(err).Error("Failed to update function schedule")
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleListSchedules handles GET /v1/schedules
func (h *Handler) HandleListSchedules(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	// Get all functions for the tenant
	functions, err := h.repo.ListFunctionsByTenant(r.Context(), user.TenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list functions")
		apierror.WriteError(w, apierror.NewInternal("Failed to list functions"))
		return
	}

	// Filter functions with schedules
	schedules := make([]ScheduleInfo, 0)
	for _, fn := range functions {
		if fn.Schedule != nil && fn.Schedule.Enabled {
			description, _ := scheduler.GetHumanReadableSchedule(fn.Schedule.Cron)
			schedules = append(schedules, ScheduleInfo{
				FunctionID:  fn.ID,
				FunctionName: fn.Name,
				Cron:        fn.Schedule.Cron,
				Timezone:    fn.Schedule.Timezone,
				Enabled:     fn.Schedule.Enabled,
				LastRun:     fn.Schedule.LastRun,
				NextRun:     fn.Schedule.NextRun,
				Description: description,
			})
		}
	}

	response := ListSchedulesResponse{
		Schedules: schedules,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetPresets handles GET /v1/schedules/presets
func (h *Handler) HandleGetPresets(w http.ResponseWriter, r *http.Request) {
	presets := scheduler.GetSchedulePresets()

	response := GetPresetsResponse{
		Presets: presets,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleTriggerManual handles POST /v1/functions/{id}/schedule/trigger
func (h *Handler) HandleTriggerManual(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	// Check if function exists and belongs to user
	function, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if function.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	// Validate that we have an executor registered
	if h.executor == nil {
		logrus.Error("No executor registered for manual trigger")
		apierror.WriteError(w, apierror.NewServiceUnavailable("Execution service not available"))
		return
	}

	// Parse optional input from request body
	var triggerReq TriggerRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&triggerReq); err != nil {
			apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
			return
		}
	}

	// Build execution input with trigger metadata
	input := map[string]interface{}{
		"trigger":     "manual",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"request_id":   r.Header.Get("X-Request-ID"),
	}
	if triggerReq.Input != nil {
		input["data"] = triggerReq.Input
	}
	inputJSON, _ := json.Marshal(input)

	// Set timeout for execution
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Execute the function
	result, err := h.executor.ExecuteFunction(ctx, functionID, inputJSON)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Manual trigger execution failed")
		apierror.WriteError(w, apierror.NewInternal("Execution failed. Check server logs for details."))
		return
	}

	response := TriggerResponse{
		FunctionID: functionID,
		Status:     "triggered",
		Timestamp:  time.Now().UTC(),
		Result:     string(result),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Request/Response types

type CreateScheduleRequest struct {
	Cron        string `json:"cron"`
	Timezone    string `json:"timezone"`
	RunOnDeploy bool   `json:"run_on_deploy"`
}

type CreateScheduleResponse struct {
	FunctionID  uuid.UUID `json:"function_id"`
	Cron        string    `json:"cron"`
	Timezone    string    `json:"timezone"`
	Enabled     bool      `json:"enabled"`
	NextRun     time.Time `json:"next_run"`
	Description string    `json:"description"`
}

type GetScheduleResponse struct {
	FunctionID  uuid.UUID `json:"function_id"`
	Cron        string    `json:"cron"`
	Timezone    string    `json:"timezone"`
	Enabled     bool      `json:"enabled"`
	LastRun     time.Time `json:"last_run"`
	NextRun     time.Time `json:"next_run"`
	Description string    `json:"description"`
}

type UpdateScheduleRequest struct {
	Cron     string `json:"cron"`
	Timezone string `json:"timezone"`
	Enabled  *bool  `json:"enabled"`
}

type UpdateScheduleResponse struct {
	FunctionID  uuid.UUID `json:"function_id"`
	Cron        string    `json:"cron"`
	Timezone    string    `json:"timezone"`
	Enabled     bool      `json:"enabled"`
	NextRun     time.Time `json:"next_run"`
	Description string    `json:"description"`
}

type ScheduleInfo struct {
	FunctionID   uuid.UUID `json:"function_id"`
	FunctionName string    `json:"function_name"`
	Cron         string    `json:"cron"`
	Timezone     string    `json:"timezone"`
	Enabled      bool      `json:"enabled"`
	LastRun      time.Time `json:"last_run"`
	NextRun      time.Time `json:"next_run"`
	Description  string    `json:"description"`
}

type ListSchedulesResponse struct {
	Schedules []ScheduleInfo `json:"schedules"`
}

type GetPresetsResponse struct {
	Presets []scheduler.SchedulePreset `json:"presets"`
}

type TriggerResponse struct {
	FunctionID uuid.UUID `json:"function_id"`
	Status     string    `json:"status"`
	Timestamp  time.Time `json:"timestamp"`
	Result     string    `json:"result,omitempty"`
}

type TriggerRequest struct {
	Input interface{} `json:"input,omitempty"`
}
