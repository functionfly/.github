package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListDevices(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListDevicesOpts{
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
	if eid := q.Get("employee_id"); eid != "" {
		if id, err := uuid.Parse(eid); err == nil {
			opts.EmployeeID = &id
		}
	}
	if dt := q.Get("device_type"); dt != "" {
		opts.DeviceType = &dt
	}
	if cs := q.Get("compliance_status"); cs != "" {
		opts.ComplianceStatus = &cs
	}
	if s := q.Get("status"); s != "" {
		opts.Status = &s
	}

	devices, total, err := h.repo.ListDevices(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list devices")
		apierror.WriteError(w, apierror.NewInternal("Failed to list devices"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"devices": devices,
		"total":   total,
		"limit":   opts.Limit,
		"offset":  opts.Offset,
	})
}

type registerDeviceRequest struct {
	EmployeeID   *string                `json:"employee_id,omitempty"`
	DeviceName   string                 `json:"device_name"`
	DeviceType   string                 `json:"device_type,omitempty"`
	SerialNumber *string                `json:"serial_number,omitempty"`
	OS           *string                `json:"os,omitempty"`
	OSVersion    *string                `json:"os_version,omitempty"`
	Manufacturer *string                `json:"manufacturer,omitempty"`
	Model        *string                `json:"model,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

func (h *Handler) HandleRegisterDevice(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req registerDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.DeviceName == "" {
		apierror.WriteError(w, apierror.NewBadRequest("device_name is required"))
		return
	}

	deviceType := "laptop"
	if req.DeviceType != "" {
		deviceType = req.DeviceType
	}

	d := &storage.Device{
		TenantID:         claims.TenantID,
		DeviceName:       req.DeviceName,
		DeviceType:       deviceType,
		SerialNumber:     req.SerialNumber,
		OS:               req.OS,
		OSVersion:        req.OSVersion,
		Manufacturer:     req.Manufacturer,
		Model:            req.Model,
		ComplianceStatus: "unknown",
		Status:           "active",
	}
	if req.EmployeeID != nil {
		eid, err := uuid.Parse(*req.EmployeeID)
		if err == nil {
			d.EmployeeID = &eid
		}
	}
	if req.Metadata != nil {
		d.Metadata = storage.JSONMap(req.Metadata)
	}

	created, err := h.repo.CreateDevice(r.Context(), d)
	if err != nil {
		h.log.WithError(err).Error("Failed to register device")
		apierror.WriteError(w, apierror.NewInternal("Failed to register device"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"device": created,
	})
}

func (h *Handler) HandleUpdateDevice(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid device ID"))
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if err := h.repo.UpdateDevice(r.Context(), id, req); err != nil {
		h.log.WithError(err).Error("Failed to update device")
		apierror.WriteError(w, apierror.NewInternal("Failed to update device"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Handler) HandleGetDevice(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid device ID"))
		return
	}

	device, err := h.repo.GetDeviceByID(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get device")
		apierror.WriteError(w, apierror.NewInternal("Failed to get device"))
		return
	}
	if device == nil {
		apierror.WriteError(w, apierror.NewNotFound("Device not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"device": device,
	})
}
