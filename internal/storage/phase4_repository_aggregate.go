package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CalculateTeamHealth aggregates team health metrics from tasks, time entries, and PTO requests.
func (r *Phase4Repository) CalculateTeamHealth(ctx context.Context, tenantID uuid.UUID, departmentID int64) (*TeamHealthMetric, error) {
	m := &TeamHealthMetric{
		TenantID:     tenantID,
		DepartmentID: &departmentID,
		MetricDate:   time.Now().Truncate(24 * time.Hour),
	}

	// Headcount
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM employees WHERE tenant_id = $1 AND department_id = $2 AND status = 'active'`,
		tenantID, departmentID).Scan(&m.Headcount)
	if err != nil {
		return nil, fmt.Errorf("failed to count headcount: %w", err)
	}

	// Workload: ratio of in-progress tasks to headcount
	var inProgressTasks int
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tasks t
		JOIN employees e ON t.assignee_id = e.id
		WHERE e.tenant_id = $1 AND e.department_id = $2 AND t.status = 'in_progress'`,
		tenantID, departmentID).Scan(&inProgressTasks)
	if err != nil {
		return nil, fmt.Errorf("failed to count in-progress tasks: %w", err)
	}

	if m.Headcount != nil && *m.Headcount > 0 {
		workload := float64(inProgressTasks) / float64(*m.Headcount) * 10
		if workload > 100 {
			workload = 100
		}
		m.WorkloadScore = &workload
	}

	// Velocity: tasks completed in last 30 days
	var completedTasks int
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tasks t
		JOIN employees e ON t.assignee_id = e.id
		WHERE e.tenant_id = $1 AND e.department_id = $2 AND t.status = 'done' AND t.updated_at >= $3`,
		tenantID, departmentID, thirtyDaysAgo).Scan(&completedTasks)
	if err != nil {
		return nil, fmt.Errorf("failed to count completed tasks: %w", err)
	}

	if m.Headcount != nil && *m.Headcount > 0 {
		velocity := float64(completedTasks) / float64(*m.Headcount)
		m.VelocityScore = &velocity
	}

	// PTO utilization: days used in last 90 days
	var ptoDays float64
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(p.days), 0) FROM pto_requests p
		JOIN employees e ON p.employee_id = e.id
		WHERE e.tenant_id = $1 AND e.department_id = $2 AND p.status = 'approved' AND p.start_date >= $3`,
		tenantID, departmentID, ninetyDaysAgo).Scan(&ptoDays)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate PTO days: %w", err)
	}

	if m.Headcount != nil && *m.Headcount > 0 {
		// Assume 10 PTO days per quarter as ideal
		ptoUtil := ptoDays / float64(*m.Headcount) / 10 * 100
		if ptoUtil > 100 {
			ptoUtil = 100
		}
		m.PTOUtilizationPct = &ptoUtil
	}

	// Avg overtime: hours above 40/week from time entries
	var avgOvertime float64
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(AVG(GREATEST(weekly_hours - 40, 0)), 0) FROM (
			SELECT te.employee_id, SUM(te.hours) as weekly_hours
			FROM time_entries te
			JOIN employees e ON te.employee_id = e.id
			WHERE e.tenant_id = $1 AND e.department_id = $2 AND te.date >= $3
			GROUP BY te.employee_id
		) weekly`, tenantID, departmentID, thirtyDaysAgo).Scan(&avgOvertime)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate avg overtime: %w", err)
	}
	m.AvgOvertimeHours = &avgOvertime

	// Burnout risk: composite of workload, overtime, and low PTO
	if m.WorkloadScore != nil && m.AvgOvertimeHours != nil {
		burnout := (*m.WorkloadScore * 0.4) + (*m.AvgOvertimeHours * 3 * 0.3)
		if m.PTOUtilizationPct != nil {
			burnout += (100 - *m.PTOUtilizationPct) * 0.3
		}
		if burnout > 100 {
			burnout = 100
		}
		m.BurnoutRisk = &burnout
	}

	return m, nil
}

// CalculateSkillsGraph aggregates skill data from employee_skills and projects.
func (r *Phase4Repository) CalculateSkillsGraph(ctx context.Context, tenantID uuid.UUID) ([]*SkillsGraph, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			s.skill_name,
			s.category,
			COUNT(DISTINCT s.employee_id) as total_employees,
			AVG(CASE s.proficiency
				WHEN 'beginner' THEN 1
				WHEN 'intermediate' THEN 2
				WHEN 'advanced' THEN 3
				WHEN 'expert' THEN 4
				ELSE 2
			END) as avg_proficiency
		FROM employee_skills s
		JOIN employees e ON s.employee_id = e.id
		WHERE e.tenant_id = $1 AND e.status = 'active'
		GROUP BY s.skill_name, s.category
		ORDER BY total_employees DESC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate skills graph: %w", err)
	}
	defer rows.Close()

	// Get total project count for demand calculation
	var totalProjects int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM projects WHERE tenant_id = $1 AND status = 'active'", tenantID).Scan(&totalProjects)
	if err != nil {
		return nil, fmt.Errorf("failed to count projects: %w", err)
	}

	var skills []*SkillsGraph
	now := time.Now()
	for rows.Next() {
		s := &SkillsGraph{
			TenantID:       tenantID,
			LastCalculated: &now,
		}
		if err := rows.Scan(&s.SkillName, &s.Category, &s.TotalEmployees, &s.AvgProficiency); err != nil {
			return nil, fmt.Errorf("failed to scan skill: %w", err)
		}

		// Simple supply score: percentage of employees with this skill
		supply := float64(s.TotalEmployees) * 100 / float64(max(totalProjects, 1))
		s.SupplyScore = &supply

		skills = append(skills, s)
	}
	return skills, nil
}

// CalculateReputation aggregates reputation metrics for an employee.
func (r *Phase4Repository) CalculateReputation(ctx context.Context, employeeID uuid.UUID, tenantID uuid.UUID) ([]*ReputationScore, error) {
	var scores []*ReputationScore
	now := time.Now()

	// Engineering: tasks completed
	var tasksCompleted int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tasks WHERE assignee_id = $1 AND status = 'done'`, employeeID).Scan(&tasksCompleted)
	if err != nil {
		return nil, fmt.Errorf("failed to count completed tasks: %w", err)
	}
	engScore := float64(tasksCompleted) * 10
	scores = append(scores, &ReputationScore{
		EmployeeID:     employeeID,
		TenantID:       tenantID,
		Category:       "engineering",
		Score:          engScore,
		LastCalculated: &now,
		Components:     JSONMap{"tasks_completed": tasksCompleted},
	})

	// Collaboration: task comments + peer feedback given
	var commentsGiven, feedbackGiven int
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM task_comments WHERE author_id = $1", employeeID).Scan(&commentsGiven)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM peer_feedback WHERE from_employee_id = $1", employeeID).Scan(&feedbackGiven)
	collabScore := float64(commentsGiven+feedbackGiven) * 5
	scores = append(scores, &ReputationScore{
		EmployeeID:     employeeID,
		TenantID:       tenantID,
		Category:       "collaboration",
		Score:          collabScore,
		LastCalculated: &now,
		Components:     JSONMap{"comments_given": commentsGiven, "feedback_given": feedbackGiven},
	})

	// Mentorship: mentoring sessions
	var mentorships int
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mentorship_matches WHERE mentor_id = $1 AND status = 'active'", employeeID).Scan(&mentorships)
	mentorScore := float64(mentorships) * 25
	scores = append(scores, &ReputationScore{
		EmployeeID:     employeeID,
		TenantID:       tenantID,
		Category:       "mentorship",
		Score:          mentorScore,
		LastCalculated: &now,
		Components:     JSONMap{"active_mentorships": mentorships},
	})

	// Innovation: grants submitted/funded
	var grantsSubmitted, grantsFunded int
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM innovation_grants WHERE proposer_id = $1", employeeID).Scan(&grantsSubmitted)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM innovation_grants WHERE proposer_id = $1 AND status = 'funded'", employeeID).Scan(&grantsFunded)
	innovScore := float64(grantsSubmitted)*10 + float64(grantsFunded)*50
	scores = append(scores, &ReputationScore{
		EmployeeID:     employeeID,
		TenantID:       tenantID,
		Category:       "innovation",
		Score:          innovScore,
		LastCalculated: &now,
		Components:     JSONMap{"grants_submitted": grantsSubmitted, "grants_funded": grantsFunded},
	})

	// Reliability: tasks completed on time
	var onTimeTasks int
	r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tasks WHERE assignee_id = $1 AND status = 'done' AND (due_date IS NULL OR updated_at <= due_date)`,
		employeeID).Scan(&onTimeTasks)
	reliabilityScore := float64(onTimeTasks) * 8
	scores = append(scores, &ReputationScore{
		EmployeeID:     employeeID,
		TenantID:       tenantID,
		Category:       "reliability",
		Score:          reliabilityScore,
		LastCalculated: &now,
		Components:     JSONMap{"on_time_tasks": onTimeTasks},
	})

	return scores, nil
}

// GenerateMissionControlSnapshot generates a comprehensive snapshot of tenant metrics.
func (r *Phase4Repository) GenerateMissionControlSnapshot(ctx context.Context, tenantID uuid.UUID) (*MissionControlSnapshot, error) {
	s := &MissionControlSnapshot{
		TenantID:     tenantID,
		SnapshotDate: time.Now().Truncate(24 * time.Hour),
	}
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	// Employee counts
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM employees WHERE tenant_id = $1", tenantID).Scan(&s.TotalEmployees)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM employees WHERE tenant_id = $1 AND status = 'active'", tenantID).Scan(&s.ActiveEmployees)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM employees WHERE tenant_id = $1 AND hire_date >= $2", tenantID, thirtyDaysAgo).Scan(&s.NewHires30d)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM employees WHERE tenant_id = $1 AND status = 'terminated' AND updated_at >= $2", tenantID, thirtyDaysAgo).Scan(&s.Departures30d)

	// Project counts
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM projects WHERE tenant_id = $1", tenantID).Scan(&s.TotalProjects)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM projects WHERE tenant_id = $1 AND status = 'active'", tenantID).Scan(&s.ActiveProjects)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM projects WHERE tenant_id = $1 AND status = 'completed' AND updated_at >= $2", tenantID, thirtyDaysAgo).Scan(&s.CompletedProjects30d)

	// Task counts
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE tenant_id = $1", tenantID).Scan(&s.TotalTasks)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE tenant_id = $1 AND status = 'done' AND updated_at >= $2", tenantID, thirtyDaysAgo).Scan(&s.CompletedTasks30d)

	// Avg task completion days
	r.db.QueryRowContext(ctx, `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 86400), 0)
		FROM tasks WHERE tenant_id = $1 AND status = 'done' AND updated_at >= $2`,
		tenantID, thirtyDaysAgo).Scan(&s.AvgTaskCompletionDays)

	// Learning hours
	r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(el.score), 0) FROM employee_learning el
		JOIN employees e ON el.employee_id = e.id
		WHERE e.tenant_id = $1 AND el.completed_at >= $2`,
		tenantID, thirtyDaysAgo).Scan(&s.TotalLearningHours)

	// Innovation grants
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM innovation_grants WHERE tenant_id = $1", tenantID).Scan(&s.InnovationGrantsSubmitted)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM innovation_grants WHERE tenant_id = $1 AND status = 'funded'", tenantID).Scan(&s.InnovationGrantsFunded)

	// PTO days used
	r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(p.days), 0) FROM pto_requests p
		JOIN employees e ON p.employee_id = e.id
		WHERE e.tenant_id = $1 AND p.status = 'approved' AND p.start_date >= $2`,
		tenantID, thirtyDaysAgo).Scan(&s.PTODaysUsed30d)

	return s, nil
}
