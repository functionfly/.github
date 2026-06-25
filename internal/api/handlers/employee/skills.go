package employee

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListSkills(w http.ResponseWriter, r *http.Request) {
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

	skills, err := h.repo.GetEmployeeSkills(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get employee skills")
		apierror.WriteError(w, apierror.NewInternal("Failed to get employee skills"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"skills": skills,
	})
}

type addSkillRequest struct {
	SkillName   string  `json:"skill_name"`
	Category    string  `json:"category,omitempty"`
	Proficiency string  `json:"proficiency,omitempty"`
	YearsExp    float64 `json:"years_exp,omitempty"`
}

func (h *Handler) HandleAddSkill(w http.ResponseWriter, r *http.Request) {
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

	var req addSkillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.SkillName == "" {
		apierror.WriteError(w, apierror.NewBadRequest("skill_name is required"))
		return
	}

	proficiency := req.Proficiency
	if proficiency == "" {
		proficiency = "intermediate"
	}

	now := time.Now()
	skill := &types.EmployeeSkill{
		EmployeeID:  id,
		SkillName:   req.SkillName,
		Proficiency: proficiency,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if req.Category != "" {
		skill.Category = &req.Category
	}
	if req.YearsExp > 0 {
		skill.YearsExp = &req.YearsExp
	}

	created, err := h.repo.AddEmployeeSkill(r.Context(), skill)
	if err != nil {
		h.log.WithError(err).Error("Failed to add skill")
		apierror.WriteError(w, apierror.NewInternal("Failed to add skill"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"skill":   created,
	})
}

func (h *Handler) HandleListCertifications(w http.ResponseWriter, r *http.Request) {
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

	certs, err := h.repo.GetEmployeeCertifications(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get certifications")
		apierror.WriteError(w, apierror.NewInternal("Failed to get certifications"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"certifications": certs,
	})
}

type addCertificationRequest struct {
	Name         string `json:"name"`
	Issuer       string `json:"issuer"`
	IssuedDate   string `json:"issued_date"`
	ExpiryDate   string `json:"expiry_date,omitempty"`
	CredentialID string `json:"credential_id,omitempty"`
	CredentialURL string `json:"credential_url,omitempty"`
}

func (h *Handler) HandleAddCertification(w http.ResponseWriter, r *http.Request) {
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

	var req addCertificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Name == "" || req.Issuer == "" {
		apierror.WriteError(w, apierror.NewBadRequest("name and issuer are required"))
		return
	}

	cert := &types.EmployeeCertification{
		EmployeeID: id,
		Name:       req.Name,
		Issuer:     req.Issuer,
	}
	if req.IssuedDate != "" {
		if t, err := time.Parse("2006-01-02", req.IssuedDate); err == nil {
			cert.IssuedDate = &t
		}
	}
	if req.ExpiryDate != "" {
		if t, err := time.Parse("2006-01-02", req.ExpiryDate); err == nil {
			cert.ExpiryDate = &t
		}
	}
	if req.CredentialID != "" {
		cert.CredentialID = &req.CredentialID
	}
	if req.CredentialURL != "" {
		cert.CredentialURL = &req.CredentialURL
	}

	created, err := h.repo.AddEmployeeCertification(r.Context(), cert)
	if err != nil {
		h.log.WithError(err).Error("Failed to add certification")
		apierror.WriteError(w, apierror.NewInternal("Failed to add certification"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"certification": created,
	})
}

func (h *Handler) HandleListAchievements(w http.ResponseWriter, r *http.Request) {
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

	achievements, err := h.repo.GetEmployeeAchievements(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get achievements")
		apierror.WriteError(w, apierror.NewInternal("Failed to get achievements"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"achievements": achievements,
	})
}
