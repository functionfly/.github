package employee

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleGetByFFID(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	ffid := mux.Vars(r)["ffid"]
	if ffid == "" {
		apierror.WriteError(w, apierror.NewBadRequest("FFID is required"))
		return
	}

	emp, err := h.repo.GetEmployeeByFFID(r.Context(), ffid)
	if err != nil {
		h.log.WithError(err).Error("Failed to get employee by FFID")
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
