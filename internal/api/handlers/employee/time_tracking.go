package employee

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListTimeEntries(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	emp, err := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	if err != nil || emp == nil {
		apierror.WriteError(w, apierror.NewNotFound("Employee profile not found"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListTimeEntriesOpts{
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
	if p := q.Get("project_id"); p != "" {
		if pid, err := uuid.Parse(p); err == nil {
			opts.ProjectID = &pid
		}
	}
	if t := q.Get("entry_type"); t != "" {
		opts.EntryType = &t
	}
	if s := q.Get("start_date"); s != "" {
		if d, err := time.Parse("2006-01-02", s); err == nil {
			opts.StartDate = &d
		}
	}
	if e := q.Get("end_date"); e != "" {
		if d, err := time.Parse("2006-01-02", e); err == nil {
			opts.EndDate = &d
		}
	}

	entries, total, err := h.repo.ListTimeEntries(r.Context(), emp.ID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list time entries")
		apierror.WriteError(w, apierror.NewInternal("Failed to list time entries"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
		"total":   total,
		"limit":   opts.Limit,
		"offset":  opts.Offset,
	})
}

type createTimeEntryRequest struct {
	ProjectID   *string `json:"project_id,omitempty"`
	TaskID      *string `json:"task_id,omitempty"`
	Date        string  `json:"date"`
	Hours       float64 `json:"hours"`
	Description string  `json:"description,omitempty"`
	EntryType   string  `json:"entry_type,omitempty"`
	IsBillable  *bool   `json:"is_billable,omitempty"`
}

func (h *Handler) HandleCreateTimeEntry(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	emp, err := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	if err != nil || emp == nil {
		apierror.WriteError(w, apierror.NewNotFound("Employee profile not found"))
		return
	}

	var req createTimeEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Date == "" || req.Hours <= 0 {
		apierror.WriteError(w, apierror.NewBadRequest("date and hours (> 0) are required"))
		return
	}

	entryDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid date format (use YYYY-MM-DD)"))
		return
	}

	entry := &storage.TimeEntry{
		EmployeeID: emp.ID,
		TenantID:   claims.TenantID,
		Date:       entryDate,
		Hours:      req.Hours,
		EntryType:  "work",
		IsBillable: true,
	}
	if req.ProjectID != nil {
		pid, err := uuid.Parse(*req.ProjectID)
		if err == nil {
			entry.ProjectID = &pid
		}
	}
	if req.TaskID != nil {
		tid, err := uuid.Parse(*req.TaskID)
		if err == nil {
			entry.TaskID = &tid
		}
	}
	if req.Description != "" {
		entry.Description = &req.Description
	}
	if req.EntryType != "" {
		entry.EntryType = req.EntryType
	}
	if req.IsBillable != nil {
		entry.IsBillable = *req.IsBillable
	}

	created, err := h.repo.CreateTimeEntry(r.Context(), entry)
	if err != nil {
		h.log.WithError(err).Error("Failed to create time entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to create time entry"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entry": created,
	})
}

type updateTimeEntryRequest struct {
	ProjectID   *string  `json:"project_id,omitempty"`
	TaskID      *string  `json:"task_id,omitempty"`
	Date        *string  `json:"date,omitempty"`
	Hours       *float64 `json:"hours,omitempty"`
	Description *string  `json:"description,omitempty"`
	EntryType   *string  `json:"entry_type,omitempty"`
	IsBillable  *bool    `json:"is_billable,omitempty"`
}

func (h *Handler) HandleUpdateTimeEntry(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid entry ID"))
		return
	}

	entry, err := h.repo.GetTimeEntryByID(r.Context(), id)
	if err != nil || entry == nil {
		apierror.WriteError(w, apierror.NewNotFound("Time entry not found"))
		return
	}

	var req updateTimeEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	updates := map[string]interface{}{}
	if req.ProjectID != nil {
		pid, err := uuid.Parse(*req.ProjectID)
		if err == nil {
			updates["project_id"] = pid
		}
	}
	if req.TaskID != nil {
		tid, err := uuid.Parse(*req.TaskID)
		if err == nil {
			updates["task_id"] = tid
		}
	}
	if req.Date != nil {
		if d, err := time.Parse("2006-01-02", *req.Date); err == nil {
			updates["date"] = d
		}
	}
	if req.Hours != nil {
		updates["hours"] = *req.Hours
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.EntryType != nil {
		updates["entry_type"] = *req.EntryType
	}
	if req.IsBillable != nil {
		updates["is_billable"] = *req.IsBillable
	}

	if err := h.repo.UpdateTimeEntry(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to update time entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to update time entry"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) HandleDeleteTimeEntry(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid entry ID"))
		return
	}

	entry, err := h.repo.GetTimeEntryByID(r.Context(), id)
	if err != nil || entry == nil {
		apierror.WriteError(w, apierror.NewNotFound("Time entry not found"))
		return
	}

	if err := h.repo.DeleteTimeEntry(r.Context(), id); err != nil {
		h.log.WithError(err).Error("Failed to delete time entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete time entry"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) HandleListPTO(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	emp, err := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	if err != nil || emp == nil {
		apierror.WriteError(w, apierror.NewNotFound("Employee profile not found"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListPTORequestsOpts{
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
	if t := q.Get("pto_type"); t != "" {
		opts.PTOType = &t
	}
	if s := q.Get("status"); s != "" {
		opts.Status = &s
	}

	requests, total, err := h.repo.ListPTORequests(r.Context(), emp.ID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list PTO requests")
		apierror.WriteError(w, apierror.NewInternal("Failed to list PTO requests"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": requests,
		"total":    total,
		"limit":    opts.Limit,
		"offset":   opts.Offset,
	})
}

type requestPTORequest struct {
	PTOType   string  `json:"pto_type"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	Days      float64 `json:"days"`
	Reason    string  `json:"reason,omitempty"`
}

func (h *Handler) HandleRequestPTO(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	emp, err := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	if err != nil || emp == nil {
		apierror.WriteError(w, apierror.NewNotFound("Employee profile not found"))
		return
	}

	var req requestPTORequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.PTOType == "" || req.StartDate == "" || req.EndDate == "" || req.Days <= 0 {
		apierror.WriteError(w, apierror.NewBadRequest("pto_type, start_date, end_date, and days (> 0) are required"))
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid start_date format"))
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid end_date format"))
		return
	}

	ptoReq := &storage.PTORequest{
		EmployeeID: emp.ID,
		TenantID:   claims.TenantID,
		PTOType:    req.PTOType,
		StartDate:  startDate,
		EndDate:    endDate,
		Days:       req.Days,
		Status:     "pending",
	}
	if req.Reason != "" {
		ptoReq.Reason = &req.Reason
	}

	created, err := h.repo.CreatePTORequest(r.Context(), ptoReq)
	if err != nil {
		h.log.WithError(err).Error("Failed to create PTO request")
		apierror.WriteError(w, apierror.NewInternal("Failed to create PTO request"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"request": created,
	})
}

type approvePTORequest struct {
	Notes string `json:"notes,omitempty"`
}

func (h *Handler) HandleApprovePTO(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request ID"))
		return
	}

	ptoReq, err := h.repo.GetPTORequestByID(r.Context(), id)
	if err != nil || ptoReq == nil {
		apierror.WriteError(w, apierror.NewNotFound("PTO request not found"))
		return
	}

	approver, err := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	if err != nil || approver == nil {
		apierror.WriteError(w, apierror.NewNotFound("Approver profile not found"))
		return
	}

	var req approvePTORequest
	json.NewDecoder(r.Body).Decode(&req)

	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}

	if err := h.repo.UpdatePTORequestStatus(r.Context(), id, "approved", &approver.ID, notes); err != nil {
		h.log.WithError(err).Error("Failed to approve PTO request")
		apierror.WriteError(w, apierror.NewInternal("Failed to approve PTO request"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

type rejectPTORequest struct {
	Notes string `json:"notes,omitempty"`
}

func (h *Handler) HandleRejectPTO(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request ID"))
		return
	}

	ptoReq, err := h.repo.GetPTORequestByID(r.Context(), id)
	if err != nil || ptoReq == nil {
		apierror.WriteError(w, apierror.NewNotFound("PTO request not found"))
		return
	}

	rejector, err := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	if err != nil || rejector == nil {
		apierror.WriteError(w, apierror.NewNotFound("Approver profile not found"))
		return
	}

	var req rejectPTORequest
	json.NewDecoder(r.Body).Decode(&req)

	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}

	if err := h.repo.UpdatePTORequestStatus(r.Context(), id, "rejected", &rejector.ID, notes); err != nil {
		h.log.WithError(err).Error("Failed to reject PTO request")
		apierror.WriteError(w, apierror.NewInternal("Failed to reject PTO request"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
