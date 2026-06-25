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

func (h *Handler) HandleClassifyResource(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		ResourceType   string  `json:"resource_type"`
		ResourceID     string  `json:"resource_id"`
		Classification string  `json:"classification"`
		Reason         *string `json:"reason,omitempty"`
		ExpiresAt      *string `json:"expires_at,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.ResourceType == "" || req.ResourceID == "" || req.Classification == "" {
		apierror.WriteError(w, apierror.NewBadRequest("resource_type, resource_id, and classification are required"))
		return
	}

	validClasses := map[string]bool{
		"public": true, "internal": true, "confidential": true,
		"restricted": true, "founder": true,
	}
	if !validClasses[req.Classification] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid classification level"))
		return
	}

	resourceID, err := uuid.Parse(req.ResourceID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid resource_id"))
		return
	}

	dc := &storage.DataClassification{
		TenantID:       claims.TenantID,
		ResourceType:   req.ResourceType,
		ResourceID:     resourceID,
		Classification: req.Classification,
		ClassifiedBy:   &emp.ID,
		Reason:         req.Reason,
	}

	created, err := h.repo.CreateDataClassification(r.Context(), dc)
	if err != nil {
		h.log.WithError(err).Error("Failed to classify resource")
		apierror.WriteError(w, apierror.NewInternal("Failed to classify resource"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"classification": created,
	})
}

func (h *Handler) HandleGetClassification(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	resourceType := mux.Vars(r)["resourceType"]
	resourceIDStr := mux.Vars(r)["resourceId"]
	if resourceType == "" || resourceIDStr == "" {
		apierror.WriteError(w, apierror.NewBadRequest("resourceType and resourceId are required"))
		return
	}

	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid resource ID"))
		return
	}

	dc, err := h.repo.GetDataClassification(r.Context(), claims.TenantID, resourceType, resourceID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get classification")
		apierror.WriteError(w, apierror.NewInternal("Failed to get classification"))
		return
	}
	if dc == nil {
		apierror.WriteError(w, apierror.NewNotFound("Classification not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"classification": dc,
	})
}

func (h *Handler) HandleListClassifications(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListDataClassificationsOpts{
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
	if rt := q.Get("resource_type"); rt != "" {
		opts.ResourceType = &rt
	}
	if cl := q.Get("classification"); cl != "" {
		opts.Classification = &cl
	}

	dcs, total, err := h.repo.ListDataClassifications(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list classifications")
		apierror.WriteError(w, apierror.NewInternal("Failed to list classifications"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"classifications": dcs,
		"total":           total,
		"limit":           opts.Limit,
		"offset":          opts.Offset,
	})
}
