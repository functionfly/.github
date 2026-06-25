package employee

import (
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleGetTeamAnalytics(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	activeStatus := "active"
	openStatus := "open"
	pendingStatus := "pending"

	employees, totalEmployees, err := h.repo.ListEmployees(r.Context(), claims.TenantID, storage.ListEmployeesOpts{
		Status: &activeStatus,
		Limit:  1,
	})
	if err != nil {
		h.log.WithError(err).Warn("Failed to count employees")
	}
	_ = employees

	_, totalProjects, err := h.repo.ListProjects(r.Context(), claims.TenantID, storage.ListProjectsOpts{
		Status: &activeStatus,
		Limit:  1,
	})
	if err != nil {
		h.log.WithError(err).Warn("Failed to count projects")
	}

	_, totalTasks, err := h.repo.ListTasks(r.Context(), claims.TenantID, storage.ListTasksOpts{
		Status: &openStatus,
		Limit:  1,
	})
	if err != nil {
		h.log.WithError(err).Warn("Failed to count tasks")
	}

	goals, err := h.repo.GetGoalTree(r.Context(), claims.TenantID)
	if err != nil {
		h.log.WithError(err).Warn("Failed to get goals")
	}
	avgGoalProgress := 0
	if len(goals) > 0 {
		total := 0
		for _, g := range goals {
			total += g.ProgressPct
		}
		avgGoalProgress = total / len(goals)
	}

	var pendingPTO int
	var hoursThisMonth float64
	if totalEmployees > 0 {
		allEmps, _, err := h.repo.ListEmployees(r.Context(), claims.TenantID, storage.ListEmployeesOpts{
			Limit: 100,
		})
		if err == nil {
			firstOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
			for _, emp := range allEmps {
				ptos, _, err := h.repo.ListPTORequests(r.Context(), emp.ID, storage.ListPTORequestsOpts{
					Status: &pendingStatus,
					Limit:  100,
				})
				if err == nil {
					pendingPTO += len(ptos)
				}

				entries, _, err := h.repo.ListTimeEntries(r.Context(), emp.ID, storage.ListTimeEntriesOpts{
					StartDate: &firstOfMonth,
					Limit:     100,
				})
				if err == nil {
					for _, e := range entries {
						hoursThisMonth += e.Hours
					}
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"analytics": map[string]interface{}{
			"total_employees":   totalEmployees,
			"active_projects":   totalProjects,
			"open_tasks":        totalTasks,
			"avg_goal_progress": avgGoalProgress,
			"pending_pto":       pendingPTO,
			"hours_this_month":  hoursThisMonth,
		},
	})
}

func (h *Handler) HandleGetSkillGapAnalysis(w http.ResponseWriter, r *http.Request) {
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

	skills, err := h.repo.GetEmployeeSkills(r.Context(), employeeID)
	if err != nil {
		h.log.WithError(err).Error("Failed to get employee skills")
		apierror.WriteError(w, apierror.NewInternal("Failed to get skills"))
		return
	}
	if skills == nil {
		skills = []*storage.EmployeeSkill{}
	}

	employeeSkills := map[string]string{}
	for _, s := range skills {
		employeeSkills[s.SkillName] = s.Proficiency
	}

	goals, err := h.repo.GetGoalTree(r.Context(), claims.TenantID)
	if err != nil {
		h.log.WithError(err).Warn("Failed to get goals for gap analysis")
	}

	var gaps []map[string]interface{}
	var recommendations []map[string]interface{}
	for _, goal := range goals {
		if goal.Category == "skill" && goal.Status != "completed" {
			skillName := goal.Title
			if _, has := employeeSkills[skillName]; !has {
				gaps = append(gaps, map[string]interface{}{
					"skill":    skillName,
					"goal_id":  goal.ID,
					"priority": goal.Priority,
				})
				recommendations = append(recommendations, map[string]interface{}{
					"skill":       skillName,
					"suggestion":  "Consider enrolling in training for " + skillName,
					"goal_id":     goal.ID,
					"target_date": goal.TargetDate,
				})
			}
		}
	}

	if gaps == nil {
		gaps = []map[string]interface{}{}
	}
	if recommendations == nil {
		recommendations = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"employee_id":     employeeID,
		"skills":          skills,
		"gaps":            gaps,
		"recommendations": recommendations,
	})
}

func (h *Handler) HandleGetTimeReport(w http.ResponseWriter, r *http.Request) {
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
		Limit: 500,
	}
	if start := q.Get("start_date"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			opts.StartDate = &t
		}
	}
	if end := q.Get("end_date"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			opts.EndDate = &t
		}
	}
	if projectID := q.Get("project_id"); projectID != "" {
		if pid, err := uuid.Parse(projectID); err == nil {
			opts.ProjectID = &pid
		}
	}

	entries, _, err := h.repo.ListTimeEntries(r.Context(), emp.ID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list time entries")
		apierror.WriteError(w, apierror.NewInternal("Failed to get time report"))
		return
	}

	var totalHours, billableHours, nonBillableHours float64
	byProject := map[string]float64{}
	byType := map[string]float64{}
	byWeek := map[string]float64{}

	for _, e := range entries {
		totalHours += e.Hours
		if e.IsBillable {
			billableHours += e.Hours
		} else {
			nonBillableHours += e.Hours
		}

		if e.ProjectID != nil {
			byProject[e.ProjectID.String()] += e.Hours
		}
		byType[e.EntryType] += e.Hours

		year, week := e.Date.ISOWeek()
		weekKey := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, (week-1)*7).Format("2006-01-02")
		byWeek[weekKey] += e.Hours
	}

	type projectEntry struct {
		ProjectID string  `json:"project_id"`
		Hours     float64 `json:"hours"`
	}
	type typeEntry struct {
		Type  string  `json:"type"`
		Hours float64 `json:"hours"`
	}
	type weekEntry struct {
		Week  string  `json:"week"`
		Hours float64 `json:"hours"`
	}

	var projectList []projectEntry
	for pid, h := range byProject {
		projectList = append(projectList, projectEntry{ProjectID: pid, Hours: h})
	}
	var typeList []typeEntry
	for t, h := range byType {
		typeList = append(typeList, typeEntry{Type: t, Hours: h})
	}
	var weekList []weekEntry
	for w, h := range byWeek {
		weekList = append(weekList, weekEntry{Week: w, Hours: h})
	}

	if projectList == nil {
		projectList = []projectEntry{}
	}
	if typeList == nil {
		typeList = []typeEntry{}
	}
	if weekList == nil {
		weekList = []weekEntry{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"report": map[string]interface{}{
			"total_hours":       totalHours,
			"billable_hours":    billableHours,
			"non_billable_hours": nonBillableHours,
			"by_project":        projectList,
			"by_type":           typeList,
			"by_week":           weekList,
		},
	})
}
