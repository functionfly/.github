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

func (h *Handler) HandleCreateCourse(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req struct {
		Title       string  `json:"title"`
		Description *string `json:"description"`
		Category    *string `json:"category"`
		Difficulty  *string `json:"difficulty"`
		DurationMin *int    `json:"duration_min"`
		ContentURL  *string `json:"content_url"`
		IsMandatory bool    `json:"is_mandatory"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Title == "" {
		apierror.WriteError(w, apierror.NewBadRequest("title is required"))
		return
	}

	course, err := h.repo.CreateLearningCourse(r.Context(), &storage.LearningCourse{
		TenantID:    claims.TenantID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Difficulty:  req.Difficulty,
		DurationMin: req.DurationMin,
		ContentURL:  req.ContentURL,
		IsMandatory: req.IsMandatory,
		IsActive:    true,
	})
	if err != nil {
		h.log.WithError(err).Error("Failed to create learning course")
		apierror.WriteError(w, apierror.NewInternal("Failed to create course"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"course":  course,
		"success": true,
	})
}

func (h *Handler) HandleListCourses(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	courses, err := h.repo.ListLearningCourses(r.Context(), claims.TenantID)
	if err != nil {
		h.log.WithError(err).Error("Failed to list learning courses")
		apierror.WriteError(w, apierror.NewInternal("Failed to list courses"))
		return
	}
	if courses == nil {
		courses = []*storage.LearningCourse{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"courses": courses,
	})
}

func (h *Handler) HandleMyLearningProgress(w http.ResponseWriter, r *http.Request) {
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

	enrollments, err := h.repo.GetEmployeeLearning(r.Context(), emp.ID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get learning progress")
		apierror.WriteError(w, apierror.NewInternal("Failed to get learning progress"))
		return
	}
	if enrollments == nil {
		enrollments = []*storage.EmployeeLearning{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"enrollments": enrollments,
	})
}

func (h *Handler) HandleEnrollCourse(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	courseID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid course ID"))
		return
	}

	course, err := h.repo.GetLearningCourseByID(r.Context(), courseID)
	if err != nil || course == nil {
		apierror.WriteError(w, apierror.NewNotFound("Course not found"))
		return
	}

	emp, err := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	if err != nil || emp == nil {
		apierror.WriteError(w, apierror.NewNotFound("Employee profile not found"))
		return
	}

	now := time.Now()
	enrollment, err := h.repo.EnrollCourse(r.Context(), &storage.EmployeeLearning{
		EmployeeID: emp.ID,
		CourseID:   courseID,
		Status:     "in_progress",
		StartedAt:  &now,
	})
	if err != nil {
		h.log.WithError(err).Error("Failed to enroll in course")
		apierror.WriteError(w, apierror.NewInternal("Failed to enroll in course"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"enrollment": enrollment,
		"success":    true,
	})
}

type updateProgressRequest struct {
	Progress int    `json:"progress"`
	Status   string `json:"status,omitempty"`
}

func (h *Handler) HandleUpdateProgress(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	progressID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid progress ID"))
		return
	}

	var req updateProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Progress < 0 || req.Progress > 100 {
		apierror.WriteError(w, apierror.NewBadRequest("Progress must be between 0 and 100"))
		return
	}

	updates := map[string]interface{}{
		"progress_pct": req.Progress,
	}
	if req.Status != "" {
		updates["status"] = req.Status
	} else if req.Progress == 100 {
		now := time.Now()
		updates["status"] = "completed"
		updates["completed_at"] = &now
	}

	if err := h.repo.UpdateLearningProgress(r.Context(), progressID, updates); err != nil {
		h.log.WithError(err).Error("Failed to update learning progress")
		apierror.WriteError(w, apierror.NewInternal("Failed to update progress"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
