package employee

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleGetIdentityCard(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	employeeIDStr := mux.Vars(r)["employeeId"]
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid employee ID"))
		return
	}

	card, err := h.repo.GetIdentityCard(r.Context(), employeeID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get identity card")
		apierror.WriteError(w, apierror.NewInternal("Failed to get identity card"))
		return
	}
	if card == nil {
		apierror.WriteError(w, apierror.NewNotFound("Employee not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"identity_card": card,
	})
}
