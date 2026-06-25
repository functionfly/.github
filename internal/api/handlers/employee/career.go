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

func (h *Handler) HandleListCareerPaths(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListCareerPathsOpts{
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
	if t := q.Get("track"); t != "" {
		opts.Track = &t
	}
	if l := q.Get("level"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			opts.Level = &n
		}
	}

	paths, total, err := h.repo.ListCareerPaths(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list career paths")
		apierror.WriteError(w, apierror.NewInternal("Failed to list career paths"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"paths":  paths,
		"total":  total,
		"limit":  opts.Limit,
		"offset": opts.Offset,
	})
}

func (h *Handler) HandleGetCareerPath(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid career path ID"))
		return
	}

	path, err := h.repo.GetCareerPathByID(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get career path")
		apierror.WriteError(w, apierror.NewInternal("Failed to get career path"))
		return
	}
	if path == nil {
		apierror.WriteError(w, apierror.NewNotFound("Career path not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"path": path,
	})
}

func (h *Handler) HandleGetMyCareerProgress(w http.ResponseWriter, r *http.Request) {
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

	progress, err := h.repo.GetEmployeeCareerProgressByEmployee(r.Context(), emp.ID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get career progress")
		apierror.WriteError(w, apierror.NewInternal("Failed to get career progress"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"progress": progress,
	})
}

type setCareerTargetRequest struct {
	CareerPathID string `json:"career_path_id"`
	TargetDate   string `json:"target_date,omitempty"`
}

func (h *Handler) HandleSetCareerTarget(w http.ResponseWriter, r *http.Request) {
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

	var req setCareerTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.CareerPathID == "" {
		apierror.WriteError(w, apierror.NewBadRequest("career_path_id is required"))
		return
	}

	pathID, err := uuid.Parse(req.CareerPathID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid career path ID"))
		return
	}

	path, err := h.repo.GetCareerPathByID(r.Context(), pathID)
	if err != nil || path == nil {
		apierror.WriteError(w, apierror.NewNotFound("Career path not found"))
		return
	}

	now := time.Now()
	prog := &storage.EmployeeCareerProgress{
		EmployeeID:   emp.ID,
		CareerPathID: pathID,
		Status:       "target",
		StartedAt:    &now,
	}
	if req.TargetDate != "" {
		if t, err := time.Parse("2006-01-02", req.TargetDate); err == nil {
			prog.TargetDate = &t
		}
	}

	created, err := h.repo.CreateEmployeeCareerProgress(r.Context(), prog)
	if err != nil {
		h.log.WithError(err).Error("Failed to set career target")
		apierror.WriteError(w, apierror.NewInternal("Failed to set career target"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"progress": created,
	})
}

func (h *Handler) HandleGetGapAnalysis(w http.ResponseWriter, r *http.Request) {
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

	pathID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid career path ID"))
		return
	}

	path, err := h.repo.GetCareerPathByID(r.Context(), pathID)
	if err != nil || path == nil {
		apierror.WriteError(w, apierror.NewNotFound("Career path not found"))
		return
	}

	skills, err := h.repo.GetEmployeeSkills(r.Context(), emp.ID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get employee skills")
		apierror.WriteError(w, apierror.NewInternal("Failed to get skills"))
		return
	}

	employeeSkills := map[string]bool{}
	for _, s := range skills {
		employeeSkills[s.SkillName] = true
	}

	var requiredSkills []string
	if path.Requirements != nil {
		if sk, ok := path.Requirements["skills"]; ok {
			if skillsList, ok := sk.([]interface{}); ok {
				for _, s := range skillsList {
					if skillName, ok := s.(string); ok {
						requiredSkills = append(requiredSkills, skillName)
					}
				}
			}
		}
	}

	var missingSkills []string
	for _, req := range requiredSkills {
		if !employeeSkills[req] {
			missingSkills = append(missingSkills, req)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"career_path":    path,
		"missing_skills": missingSkills,
		"total_required": len(requiredSkills),
		"total_missing":  len(missingSkills),
	})
}
