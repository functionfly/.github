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

func (h *Handler) HandleListEmployees(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListEmployeesOpts{
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
	if s := q.Get("status"); s != "" {
		opts.Status = &s
	}
	if s := q.Get("search"); s != "" {
		opts.Search = &s
	}

	employees, total, err := h.repo.ListEmployees(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list employees")
		apierror.WriteError(w, apierror.NewInternal("Failed to list employees"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"employees": employees,
		"total":     total,
		"limit":     opts.Limit,
		"offset":    opts.Offset,
	})
}

func (h *Handler) HandleGetEmployee(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee ID"))
		return
	}

	emp, err := h.repo.GetEmployeeByID(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get employee")
		apierror.WriteError(w, apierror.NewInternal("Failed to get employee"))
		return
	}
	if emp == nil {
		apierror.WriteError(w, apierror.NewNotFound("Employee not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"employee": emp,
	})
}

type createEmployeeRequest struct {
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	Email          string  `json:"email"`
	UserID         *string `json:"user_id"`
	EmployeeNumber *string `json:"employee_number"`
	FFID           *string `json:"ffid"`
	DepartmentID   *int64  `json:"department_id"`
	Title          *string `json:"title"`
	EmploymentType string  `json:"employment_type"`
	ClearanceLevel string  `json:"clearance_level"`
	WorkLocation   *string `json:"work_location"`
}

func (h *Handler) HandleCreateEmployee(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !claims.HasPermission("admin") && claims.Role != "admin" && claims.Role != "super_admin" {
		apierror.WriteError(w, apierror.NewForbidden("Admin access required"))
		return
	}

	var req createEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		apierror.WriteError(w, apierror.NewBadRequest("first_name, last_name, and email are required"))
		return
	}

	// Generate defaults
	userID := claims.UserID
	if req.UserID != nil {
		uid, err := uuid.Parse(*req.UserID)
		if err == nil {
			userID = uid
		}
	}

	empNumber := "EMP-" + uuid.New().String()[:8]
	if req.EmployeeNumber != nil {
		empNumber = *req.EmployeeNumber
	}

	ffid := "FF-26-" + uuid.New().String()[:4] + "-" + empNumber[4:]
	if req.FFID != nil {
		ffid = *req.FFID
	}

	empType := req.EmploymentType
	if empType == "" {
		empType = "full_time"
	}

	clearance := req.ClearanceLevel
	if clearance == "" {
		clearance = "standard"
	}

	emp := &types.Employee{
		ID:              uuid.New(),
		UserID:          userID,
		TenantID:        claims.TenantID,
		EmployeeNumber:  empNumber,
		FFID:            ffid,
		DepartmentID:    req.DepartmentID,
		EmploymentType:  empType,
		ClearanceLevel:  clearance,
		WorkLocation:    req.WorkLocation,
		Status:          "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	created, err := h.repo.CreateEmployee(r.Context(), emp)
	if err != nil {
		h.log.WithError(err).Error("Failed to create employee")
		apierror.WriteError(w, apierror.NewInternal("Failed to create employee"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"employee": created,
	})
}

type updateEmployeeRequest struct {
	FirstName      *string `json:"first_name,omitempty"`
	LastName       *string `json:"last_name,omitempty"`
	DepartmentID   *int64  `json:"department_id,omitempty"`
	Title          *string `json:"title,omitempty"`
	Status         *string `json:"status,omitempty"`
	WorkLocation   *string `json:"work_location,omitempty"`
	ClearanceLevel *string `json:"clearance_level,omitempty"`
}

func (h *Handler) HandleUpdateEmployee(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee ID"))
		return
	}

	isAdmin := claims.HasPermission("admin") || claims.Role == "admin" || claims.Role == "super_admin"
	_ = isAdmin

	var req updateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	updates := make(map[string]interface{})
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.WorkLocation != nil {
		updates["work_location"] = *req.WorkLocation
	}
	if req.ClearanceLevel != nil {
		updates["clearance_level"] = *req.ClearanceLevel
	}
	if req.DepartmentID != nil {
		updates["department_id"] = *req.DepartmentID
	}
	updates["updated_at"] = time.Now()

	if err := h.repo.UpdateEmployee(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to update employee")
		apierror.WriteError(w, apierror.NewInternal("Failed to update employee"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) HandleGenerateAccess(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !claims.HasPermission("admin") && claims.Role != "admin" && claims.Role != "super_admin" {
		apierror.WriteError(w, apierror.NewForbidden("Admin access required"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee ID"))
		return
	}

	emp, err := h.repo.GetEmployeeByID(r.Context(), id)
	if err != nil || emp == nil {
		apierror.WriteError(w, apierror.NewNotFound("Employee not found"))
		return
	}

	// Generate a short-lived access token for the employee portal
	accessToken := uuid.New().String()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":        accessToken,
		"employee_id":  emp.ID,
		"ffid":         emp.FFID,
		"expires_in":   3600,
		"portal_url":   "/login",
	})
}

type updateClearanceLevelRequest struct {
	Level int `json:"level"`
}

func (h *Handler) HandleUpdateClearanceLevel(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !claims.HasPermission("admin") && claims.Role != "admin" && claims.Role != "super_admin" {
		apierror.WriteError(w, apierror.NewForbidden("Admin access required"))
		return
	}

	employeeIDStr := mux.Vars(r)["id"]
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee ID"))
		return
	}

	var req updateClearanceLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Level < 0 || req.Level > 5 {
		apierror.WriteError(w, apierror.NewBadRequest("Level must be between 0 and 5"))
		return
	}

	if err := h.repo.UpdateClearanceLevel(r.Context(), employeeID, req.Level); err != nil {
		h.log.WithError(err).Error("Failed to update clearance level")
		apierror.WriteError(w, apierror.NewInternal("Failed to update clearance level"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"level":   req.Level,
	})
}
