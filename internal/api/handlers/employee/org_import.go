package employee

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type uploadOrgChartRequest struct {
	FileName string `json:"file_name"`
	FileType string `json:"file_type,omitempty"`
}

func (h *Handler) HandleUploadOrgChart(w http.ResponseWriter, r *http.Request) {
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

	var req uploadOrgChartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.FileName == "" {
		apierror.WriteError(w, apierror.NewBadRequest("file_name is required"))
		return
	}

	fileType := "csv"
	if req.FileType != "" {
		fileType = req.FileType
	}

	imp := &storage.OrgChartImport{
		TenantID:   claims.TenantID,
		UploadedBy: emp.ID,
		FileName:   req.FileName,
		FileType:   fileType,
		Status:     "pending",
	}

	created, err := h.repo.CreateOrgChartImport(r.Context(), imp)
	if err != nil {
		h.log.WithError(err).Error("Failed to upload org chart")
		apierror.WriteError(w, apierror.NewInternal("Failed to upload org chart"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"import": created,
	})
}

func (h *Handler) HandleGetImportStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid import ID"))
		return
	}

	imp, err := h.repo.GetOrgChartImportByID(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get import status")
		apierror.WriteError(w, apierror.NewInternal("Failed to get import status"))
		return
	}
	if imp == nil {
		apierror.WriteError(w, apierror.NewNotFound("Import not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"import": imp,
	})
}
